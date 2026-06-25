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
		log.Println("warning: product APIs require PostgreSQL (set DATABASE_URL or PG* env vars)")
	}

	var productHandler *handler.ProductHandler
	if helper.IsPostgres {
		repo := repository.NewPostgresProductRepository(helper.DB)
		svc := service.NewProductService(repo)
		var err error
		productHandler, err = handler.NewProductHandler(svc)
		if err != nil {
			log.Fatalf("product handler: %v", err)
		}
	}

	engine := router.Setup(productHandler)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + strings.TrimPrefix(strings.TrimSpace(p), ":")
	}

	log.Printf("listening on %s", addr)
	log.Fatal(engine.Run(addr))
}
