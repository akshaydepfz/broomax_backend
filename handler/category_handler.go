package handler

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"oryoo.com/dto"
	"oryoo.com/helper"
	"oryoo.com/service"
)

// CategoryHandler serves category HTTP endpoints.
type CategoryHandler struct {
	svc      *service.CategoryService
	uploader *helper.S3Uploader
}

// NewCategoryHandler creates a CategoryHandler.
func NewCategoryHandler(svc *service.CategoryService) (*CategoryHandler, error) {
	uploader, err := helper.NewS3UploaderFromEnv()
	if err != nil {
		return nil, err
	}
	return &CategoryHandler{svc: svc, uploader: uploader}, nil
}

// Create handles POST /categories (JSON or multipart with image).
func (h *CategoryHandler) Create(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req, imageURL, err := h.parseCategoryMultipart(c)
		if err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		if imageURL != "" {
			req.ImageURL = imageURL
		}
		category, err := h.svc.Create(c.Request.Context(), req)
		if err != nil {
			status, msg := service.MapCategoryServiceError(err)
			fail(c, status, msg)
			return
		}
		ok(c, http.StatusCreated, category)
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	category, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		status, msg := service.MapCategoryServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusCreated, category)
}

// List handles GET /categories.
func (h *CategoryHandler) List(c *gin.Context) {
	var q dto.CategoryListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		fail(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	if raw := c.Query("is_active"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			fail(c, http.StatusBadRequest, "is_active must be true or false")
			return
		}
		q.IsActive = &parsed
	}
	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		fail(c, http.StatusInternalServerError, "could not list categories")
		return
	}
	ok(c, http.StatusOK, resp)
}

// GetByID handles GET /categories/:id.
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	category, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		status, msg := service.MapCategoryServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, category)
}

// Update handles PUT /categories/:id (JSON or multipart with image).
func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req, imageURL, err := h.parseCategoryUpdateMultipart(c)
		if err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		if imageURL != "" {
			req.ImageURL = &imageURL
		}
		category, err := h.svc.Update(c.Request.Context(), id, req)
		if err != nil {
			status, msg := service.MapCategoryServiceError(err)
			fail(c, status, msg)
			return
		}
		ok(c, http.StatusOK, category)
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	category, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		status, msg := service.MapCategoryServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, category)
}

// Delete handles DELETE /categories/:id.
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		status, msg := service.MapCategoryServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, gin.H{"deleted": true, "id": id})
}

// AddImage handles POST /categories/:id/image (multipart or JSON image_url).
func (h *CategoryHandler) AddImage(c *gin.Context) {
	id := c.Param("id")
	var imageURL string

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid multipart form")
			return
		}
		files := form.File["image"]
		if len(files) == 0 {
			files = form.File["images"]
		}
		if len(files) == 0 {
			files = form.File["files"]
		}
		if len(files) == 0 {
			fail(c, http.StatusBadRequest, "provide image file")
			return
		}
		imageURL, err = h.saveCategoryFile(c, files[0], "categories")
		if err != nil {
			fail(c, http.StatusInternalServerError, "could not save image")
			return
		}
	} else {
		var req dto.AddCategoryImageRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.ImageURL) == "" {
			fail(c, http.StatusBadRequest, "provide image file or image_url in JSON")
			return
		}
		imageURL = req.ImageURL
	}

	category, err := h.svc.SetImageURL(c.Request.Context(), id, imageURL)
	if err != nil {
		status, msg := service.MapCategoryServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, category)
}

func (h *CategoryHandler) parseCategoryMultipart(c *gin.Context) (dto.CreateCategoryRequest, string, error) {
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		title = strings.TrimSpace(c.PostForm("name"))
	}
	if title == "" {
		return dto.CreateCategoryRequest{}, "", errText("title is required")
	}

	req := dto.CreateCategoryRequest{
		Title:       title,
		Description: strings.TrimSpace(c.PostForm("description")),
	}
	if raw := c.PostForm("sort_order"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return dto.CreateCategoryRequest{}, "", errText("sort_order must be an integer")
		}
		req.SortOrder = &n
	}
	if raw := c.PostForm("is_active"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return dto.CreateCategoryRequest{}, "", errText("is_active must be true or false")
		}
		req.IsActive = &parsed
	}
	if raw := strings.TrimSpace(c.PostForm("image_url")); raw != "" {
		req.ImageURL = raw
	}

	imageURL, err := h.uploadCategoryImageFromForm(c)
	if err != nil {
		return dto.CreateCategoryRequest{}, "", err
	}
	return req, imageURL, nil
}

func (h *CategoryHandler) parseCategoryUpdateMultipart(c *gin.Context) (dto.UpdateCategoryRequest, string, error) {
	var req dto.UpdateCategoryRequest
	if raw := strings.TrimSpace(c.PostForm("title")); raw != "" {
		req.Title = &raw
	} else if raw := strings.TrimSpace(c.PostForm("name")); raw != "" {
		req.Title = &raw
	}
	if raw := c.PostForm("description"); raw != "" {
		v := strings.TrimSpace(raw)
		req.Description = &v
	}
	if raw := c.PostForm("sort_order"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return dto.UpdateCategoryRequest{}, "", errText("sort_order must be an integer")
		}
		req.SortOrder = &n
	}
	if raw := c.PostForm("is_active"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return dto.UpdateCategoryRequest{}, "", errText("is_active must be true or false")
		}
		req.IsActive = &parsed
	}
	if raw := strings.TrimSpace(c.PostForm("image_url")); raw != "" {
		req.ImageURL = &raw
	}

	imageURL, err := h.uploadCategoryImageFromForm(c)
	if err != nil {
		return dto.UpdateCategoryRequest{}, "", err
	}
	return req, imageURL, nil
}

func (h *CategoryHandler) uploadCategoryImageFromForm(c *gin.Context) (string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return "", errText("invalid multipart form")
	}
	files := form.File["image"]
	if len(files) == 0 {
		files = form.File["images"]
	}
	if len(files) == 0 {
		return "", nil
	}
	return h.saveCategoryFile(c, files[0], "categories")
}

func (h *CategoryHandler) saveCategoryFile(c *gin.Context, fh *multipart.FileHeader, subdir string) (string, error) {
	ext := filepath.Ext(fh.Filename)
	name := uuid.NewString() + ext
	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = helper.ContentTypeForFilename(fh.Filename)
	}

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	return h.uploader.Upload(c.Request.Context(), src, contentType, subdir, name)
}

type textError string

func (e textError) Error() string { return string(e) }

func errText(msg string) error { return textError(msg) }
