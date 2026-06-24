package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"oryoo.com/dto"
	"oryoo.com/repository"
	"oryoo.com/service"
)

// ProductHandler serves product HTTP endpoints.
type ProductHandler struct {
	svc       *service.ProductService
	uploadDir string
	baseURL   string
}

// NewProductHandler creates a ProductHandler.
func NewProductHandler(svc *service.ProductService) *ProductHandler {
	dir := strings.TrimSpace(os.Getenv("UPLOAD_DIR"))
	if dir == "" {
		dir = "./uploads"
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UPLOAD_BASE_URL")), "/")
	if base == "" {
		base = "/uploads"
	}
	return &ProductHandler{svc: svc, uploadDir: dir, baseURL: base}
}

// Create handles POST /products.
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	product, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		status, msg := service.MapServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusCreated, product)
}

// List handles GET /products.
func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ProductListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		fail(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		fail(c, http.StatusInternalServerError, "could not list products")
		return
	}
	ok(c, http.StatusOK, resp)
}

// GetByID handles GET /products/:id.
func (h *ProductHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	product, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		status, msg := service.MapServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, product)
}

// Update handles PUT /products/:id.
func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	product, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		status, msg := service.MapServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, product)
}

// Delete handles DELETE /products/:id.
func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		status, msg := service.MapServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, gin.H{"deleted": true, "id": id})
}

// BulkUpload handles POST /products/bulk-upload.
func (h *ProductHandler) BulkUpload(c *gin.Context) {
	var req dto.BulkUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.svc.BulkUpload(c.Request.Context(), req)
	if err != nil {
		status, msg := service.MapServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusOK, resp)
}

// AddImages handles POST /products/:id/images (multipart or JSON urls).
func (h *ProductHandler) AddImages(c *gin.Context) {
	id := c.Param("id")
	var urls []string

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid multipart form")
			return
		}
		files := form.File["images"]
		if len(files) == 0 {
			files = form.File["files"]
		}
		for _, fh := range files {
			url, err := h.saveUploadedFile(c, fh, "images")
			if err != nil {
				fail(c, http.StatusInternalServerError, "could not save image")
				return
			}
			urls = append(urls, url)
		}
	} else {
		var req dto.AddImagesRequest
		if err := c.ShouldBindJSON(&req); err != nil || len(req.URLs) == 0 {
			fail(c, http.StatusBadRequest, "provide images files or urls[] in JSON")
			return
		}
		urls = req.URLs
	}

	images, err := h.svc.AddImages(c.Request.Context(), id, urls)
	if err != nil {
		status, msg := service.MapServiceError(err)
		if errors.Is(err, repository.ErrNotFound) {
			fail(c, status, msg)
			return
		}
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusCreated, gin.H{"images": images})
}

// AddDocuments handles POST /products/:id/documents.
func (h *ProductHandler) AddDocuments(c *gin.Context) {
	id := c.Param("id")
	var inputs []dto.DocumentInput

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid multipart form")
			return
		}
		files := form.File["documents"]
		if len(files) == 0 {
			files = form.File["files"]
		}
		docType := c.PostForm("doc_type")
		if docType == "" {
			docType = "other"
		}
		for _, fh := range files {
			url, err := h.saveUploadedFile(c, fh, "documents")
			if err != nil {
				fail(c, http.StatusInternalServerError, "could not save document")
				return
			}
			inputs = append(inputs, dto.DocumentInput{
				DocType:  docType,
				URL:      url,
				Filename: fh.Filename,
			})
		}
	} else {
		var req dto.AddDocumentsRequest
		if err := c.ShouldBindJSON(&req); err != nil || len(req.Documents) == 0 {
			fail(c, http.StatusBadRequest, "provide documents files or documents[] in JSON")
			return
		}
		inputs = req.Documents
	}

	docs, err := h.svc.AddDocuments(c.Request.Context(), id, inputs)
	if err != nil {
		status, msg := service.MapServiceError(err)
		fail(c, status, msg)
		return
	}
	ok(c, http.StatusCreated, gin.H{"documents": docs})
}

func (h *ProductHandler) saveUploadedFile(c *gin.Context, fh *multipart.FileHeader, subdir string) (string, error) {
	if err := os.MkdirAll(filepath.Join(h.uploadDir, subdir), 0o755); err != nil {
		return "", err
	}
	ext := filepath.Ext(fh.Filename)
	name := uuid.NewString() + ext
	dest := filepath.Join(h.uploadDir, subdir, name)

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return h.baseURL + "/" + subdir + "/" + name, nil
}

func ok(c *gin.Context, status int, data any) {
	c.JSON(status, JSONEnvelope{Success: true, Data: data})
}

func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, JSONEnvelope{Success: false, Error: msg})
}
