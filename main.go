package main

import (
	"log"
	"os"
	"strings"

	"oryoo.com/handler"
	"oryoo.com/helper"
	"oryoo.com/repository"
	"oryoo.com/router"
	"oryoo.com/service"
)

func main() {
	if err := helper.InitDB(); err != nil {
		log.Fatalf("database: %v", err)
	}

	if !helper.IsPostgres {
		log.Println("warning: product and category APIs require PostgreSQL (set DATABASE_URL or PG* env vars)")
	}

	var productHandler *handler.ProductHandler
	var categoryHandler *handler.CategoryHandler
	if helper.IsPostgres {
		productRepo := repository.NewPostgresProductRepository(helper.DB)
		productSvc := service.NewProductService(productRepo)
		var err error
		productHandler, err = handler.NewProductHandler(productSvc)
		if err != nil {
			log.Fatalf("product handler: %v", err)
		}

		categoryRepo := repository.NewPostgresCategoryRepository(helper.DB)
		categorySvc := service.NewCategoryService(categoryRepo)
		categoryHandler, err = handler.NewCategoryHandler(categorySvc)
		if err != nil {
			log.Fatalf("category handler: %v", err)
		}
	}

	engine := router.Setup(productHandler, categoryHandler)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + strings.TrimPrefix(strings.TrimSpace(p), ":")
	}

	log.Printf("listening on %s", addr)
	log.Fatal(engine.Run(addr))
}
