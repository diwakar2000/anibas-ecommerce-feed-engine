package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/model"
)

type ProductRepository struct {
	db *bun.DB
}

type CreateProductInput struct {
	SKU         string `json:"sku" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents" binding:"required,min=1"`
	Currency    string `json:"currency" binding:"required,len=3"`
	Inventory   int    `json:"inventory" binding:"min=0"`
}

func NewProductRepository(db *bun.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context) ([]model.Product, error) {
	products := make([]model.Product, 0)
	err := r.db.NewSelect().
		Model(&products).
		Order("created_at DESC").
		Scan(ctx)

	return products, err
}

func (r *ProductRepository) Create(ctx context.Context, input CreateProductInput) (model.Product, error) {
	product := model.Product{
		SKU:         strings.TrimSpace(input.SKU),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		PriceCents:  input.PriceCents,
		Currency:    strings.ToUpper(strings.TrimSpace(input.Currency)),
		Inventory:   input.Inventory,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if product.SKU == "" || product.Title == "" || product.Currency == "" {
		return model.Product{}, errors.New("missing required product fields")
	}

	_, err := r.db.NewInsert().
		Model(&product).
		Returning("*").
		Exec(ctx)

	return product, err
}
