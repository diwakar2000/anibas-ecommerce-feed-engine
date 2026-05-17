package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/config"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/repository"
)

type productHandler struct {
	repo *repository.ProductRepository
}

func RegisterRoutes(router *gin.Engine, cfg config.Config, db *bun.DB, productRepo *repository.ProductRepository) {
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendOrigin},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := contextWithTimeout(c)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "degraded",
				"database": "unreachable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "ok",
		})
	})

	handler := productHandler{repo: productRepo}

	v1 := router.Group("/api/v1")
	v1.GET("/products", handler.listProducts)
	v1.POST("/products", handler.createProduct)
}

func (h productHandler) listProducts(c *gin.Context) {
	ctx, cancel := contextWithTimeout(c)
	defer cancel()

	products, err := h.repo.List(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h productHandler) createProduct(c *gin.Context) {
	var input repository.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := contextWithTimeout(c)
	defer cancel()

	product, err := h.repo.Create(ctx, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func contextWithTimeout(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 5*time.Second)
}
