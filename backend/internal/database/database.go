package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/model"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/repository"
)

func Connect(databaseURL string) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))
	db := bun.NewDB(sqldb, pgdialect.New())

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func Migrate(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*model.Product)(nil)).
		IfNotExists().
		Exec(ctx)

	return err
}

func SeedProducts(ctx context.Context, repo *repository.ProductRepository) error {
	products, err := repo.List(ctx)
	if err != nil {
		return err
	}

	if len(products) > 0 {
		return nil
	}

	_, err = repo.Create(ctx, repository.CreateProductInput{
		SKU:         "ANI-STARTER-001",
		Title:       "Starter Feed Product",
		Description: "A seeded product proving the Svelte app, Go API, and Postgres database are talking.",
		PriceCents:  2499,
		Currency:    "USD",
		Inventory:   25,
	})

	return err
}
