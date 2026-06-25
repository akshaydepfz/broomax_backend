package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"oryoo.com/handler"
	"oryoo.com/helper"
)

// Setup builds the Gin engine with all routes and middleware.
func Setup(productHandler *handler.ProductHandler, categoryHandler *handler.CategoryHandler) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), corsMiddleware())

	// Admin routes (legacy net/http handlers bridged via Gin)
	r.POST("/admin/login", gin.WrapF(handler.AdminLoginHandler))
	r.POST("/admin/create-admin", gin.WrapF(handler.CreateAdminHandler))

	admin := r.Group("/admin", adminAuthMiddleware())
	{
		_ = admin // future admin-only routes
	}

	// Product catalog API (PostgreSQL)
	if productHandler != nil {
		products := r.Group("/products")
		{
			products.GET("", productHandler.List)
			products.GET("/:id", productHandler.GetByID)

			protected := products.Group("", adminAuthMiddleware())
			{
				protected.POST("", productHandler.Create)
				protected.POST("/bulk-upload", productHandler.BulkUpload)
				protected.PUT("/:id", productHandler.Update)
				protected.DELETE("/:id", productHandler.Delete)
				protected.POST("/:id/images", productHandler.AddImages)
				protected.POST("/:id/documents", productHandler.AddDocuments)
			}
		}
	}

	if categoryHandler != nil {
		categories := r.Group("/categories")
		{
			categories.GET("", categoryHandler.List)
			categories.GET("/:id", categoryHandler.GetByID)

			protected := categories.Group("", adminAuthMiddleware())
			{
				protected.POST("", categoryHandler.Create)
				protected.PUT("/:id", categoryHandler.Update)
				protected.DELETE("/:id", categoryHandler.Delete)
				protected.POST("/:id/image", categoryHandler.AddImage)
			}
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, handler.JSONEnvelope{Success: false, Error: "not found"})
	})

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

var adminPublicPaths = map[string]struct{}{
	"/admin/login":        {},
	"/admin/create-admin": {},
}

func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, ok := adminPublicPaths[path]; ok {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.JSONEnvelope{
				Success: false, Error: "missing or invalid authorization",
			})
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.JSONEnvelope{
				Success: false, Error: "missing or invalid authorization",
			})
			return
		}

		static := os.Getenv("ADMIN_BEARER_SECRET")
		if static != "" && raw == static {
			c.Next()
			return
		}

		if _, err := helper.ParseAndVerifyAdminJWT(raw); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.JSONEnvelope{
				Success: false, Error: "invalid or expired token",
			})
			return
		}
		c.Next()
	}
}
