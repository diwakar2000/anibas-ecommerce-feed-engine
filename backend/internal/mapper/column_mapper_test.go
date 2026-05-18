package mapper

import (
	"testing"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
)

func TestSuggestFieldMappingsRecognizesCommonMarketplaceHeaders(t *testing.T) {
	columns := []string{
		"Variant SKU",
		"Product Title",
		"Body HTML",
		"Variant Price",
		"Currency Code",
		"Inventory Quantity",
		"Image Src",
		"Product URL",
		"Variant Barcode",
		"Manufacturer Part Number",
	}

	suggestions := SuggestFieldMappings(columns, domain.UniversalProductSchema())
	mapping := toMap(suggestions)

	expected := map[string]string{
		"sku":         "Variant SKU",
		"title":       "Product Title",
		"description": "Body HTML",
		"price":       "Variant Price",
		"currency":    "Currency Code",
		"quantity":    "Inventory Quantity",
		"image_url":   "Image Src",
		"product_url": "Product URL",
		"gtin":        "Variant Barcode",
		"mpn":         "Manufacturer Part Number",
	}

	for field, column := range expected {
		if mapping[field] != column {
			t.Fatalf("mapping[%q] = %q, want %q", field, mapping[field], column)
		}
	}
}

func TestSuggestFieldMappingsDoesNotReuseColumns(t *testing.T) {
	columns := []string{"ID", "Product ID", "Product Title"}

	suggestions := SuggestFieldMappings(columns, domain.UniversalProductSchema())
	mapping := toMap(suggestions)

	if mapping["id"] != "ID" {
		t.Fatalf("id mapped to %q, want ID", mapping["id"])
	}
	if mapping["title"] != "Product Title" {
		t.Fatalf("title mapped to %q, want Product Title", mapping["title"])
	}
}

func TestCanonicalizeFieldMappingMatchesColumnsCaseInsensitively(t *testing.T) {
	columns := []string{"SKU", "Product Title", "Image_Link"}
	mapping := map[string]string{
		"sku":       "sku",
		"title":     " product   title ",
		"image_url": "image-link",
		"brand":     "Vendor Name",
	}

	canonical := CanonicalizeFieldMapping(mapping, columns)

	if canonical["sku"] != "SKU" {
		t.Fatalf("sku = %q, want SKU", canonical["sku"])
	}
	if canonical["title"] != "Product Title" {
		t.Fatalf("title = %q, want Product Title", canonical["title"])
	}
	if canonical["image_url"] != "Image_Link" {
		t.Fatalf("image_url = %q, want Image_Link", canonical["image_url"])
	}
	if canonical["brand"] != "Vendor Name" {
		t.Fatalf("brand = %q, want unmatched trimmed value", canonical["brand"])
	}
}

func toMap(suggestions []FieldMappingSuggestion) map[string]string {
	result := make(map[string]string, len(suggestions))
	for _, suggestion := range suggestions {
		result[suggestion.Field] = suggestion.SourceColumn
	}

	return result
}
