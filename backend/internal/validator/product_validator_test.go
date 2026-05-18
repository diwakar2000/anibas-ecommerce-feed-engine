package validator

import (
	"testing"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
)

func TestValidateProductsFlagsDuplicateSKU(t *testing.T) {
	products := []domain.UniversalProduct{
		validProduct("ANI-1"),
		validProduct("ANI-1"),
	}

	report := ValidateProducts(products)
	if report.Status != "invalid" {
		t.Fatalf("Status = %q, want invalid", report.Status)
	}
	if len(report.Errors) != 1 {
		t.Fatalf("errors = %d, want 1", len(report.Errors))
	}
	if report.Errors[0].Field != "sku" {
		t.Fatalf("field = %q, want sku", report.Errors[0].Field)
	}
}

func TestValidateProductsFlagsInvalidPriceAndImageURL(t *testing.T) {
	product := validProduct("ANI-1")
	product.Price = "-1"
	product.ImageURL = "ftp://example.test/image.jpg"

	report := ValidateProducts([]domain.UniversalProduct{product})
	if len(report.Errors) != 2 {
		t.Fatalf("errors = %d, want 2", len(report.Errors))
	}
}

func validProduct(sku string) domain.UniversalProduct {
	return domain.UniversalProduct{
		SKU:          sku,
		Title:        "Organic Cotton T-shirt",
		Description:  "Soft shirt",
		Price:        "19.99",
		Currency:     "USD",
		Quantity:     10,
		Condition:    "new",
		ImageURL:     "https://example.test/image.jpg",
		ProductURL:   "https://example.test/products/ani-1",
		Availability: "in stock",
	}
}
