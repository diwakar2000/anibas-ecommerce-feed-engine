package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/config"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/database"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/httpapi"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := database.Migrate(ctx, db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	productRepo := repository.NewProductRepository(db)
	if err := database.SeedProducts(ctx, productRepo); err != nil {
		log.Fatalf("seed products: %v", err)
	}

	router := gin.New()
	httpapi.RegisterRoutes(router, cfg, db, productRepo)

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown server: %v", err)
	}
}
