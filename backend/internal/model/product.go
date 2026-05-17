package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Product struct {
	bun.BaseModel `bun:"table:products,alias:p"`

	ID          int64     `bun:",pk,autoincrement" json:"id"`
	SKU         string    `bun:",unique,notnull" json:"sku"`
	Title       string    `bun:",notnull" json:"title"`
	Description string    `json:"description"`
	PriceCents  int64     `bun:",notnull" json:"price_cents"`
	Currency    string    `bun:",notnull" json:"currency"`
	Inventory   int       `bun:",notnull" json:"inventory"`
	CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}
