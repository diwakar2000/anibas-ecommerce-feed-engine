package exporter

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
)

func TestExportersWriteTargetSpecificCSV(t *testing.T) {
	product := domain.UniversalProduct{
		ID:                  "p-1",
		SKU:                 "sku-1",
		Title:               "Demo Shirt",
		Description:         "Soft cotton shirt",
		Price:               "19.99",
		Currency:            "USD",
		Quantity:            12,
		Condition:           "new",
		Brand:               "Anibas",
		GTIN:                "1234567890123",
		MPN:                 "ANB-1",
		Category:            "Apparel & Accessories",
		ImageURL:            "https://example.com/main.jpg",
		AdditionalImageURLs: []string{"https://example.com/side.jpg"},
		ProductURL:          "https://example.com/products/demo-shirt",
		Availability:        "in stock",
		VariantGroupID:      "vg-1",
		Option1Name:         "Size",
		Option1Value:        "M",
	}

	tests := []struct {
		target         string
		expectedFile   string
		expectedHeader string
		expectedCell   string
	}{
		{
			target:         "facebook_catalog_csv",
			expectedFile:   "facebook_catalog_preview.csv",
			expectedHeader: "id,title,description,availability,condition,price,link,image_link,brand,google_product_category,inventory",
			expectedCell:   "19.99 USD",
		},
		{
			target:         "instagram_shops",
			expectedFile:   "instagram_shops_preview.csv",
			expectedHeader: "id,title,description,availability,condition,price,link,image_link,brand,google_product_category,inventory",
			expectedCell:   "19.99 USD",
		},
		{
			target:         "google_merchant_center",
			expectedFile:   "google_merchant_center_preview.csv",
			expectedHeader: "id,title,description,link,image_link,additional_image_link,availability,price,condition,brand,gtin,mpn,google_product_category,item_group_id",
			expectedCell:   "1234567890123",
		},
		{
			target:         "tiktok_catalog",
			expectedFile:   "tiktok_catalog_preview.csv",
			expectedHeader: "sku_id,title,description,availability,condition,price,currency,link,image_link,additional_image_link,brand,product_type,item_group_id,inventory",
			expectedCell:   "USD",
		},
		{
			target:         "shopify_csv",
			expectedFile:   "shopify_products_preview.csv",
			expectedHeader: "Handle,Title,Body (HTML),Vendor,Product Category,Type,Published,Option1 Name,Option1 Value,Option2 Name,Option2 Value,Variant SKU,Variant Inventory Qty,Variant Price,Image Src,Status",
			expectedCell:   "p-1",
		},
		{
			target:         "woocommerce_csv",
			expectedFile:   "woocommerce_products_preview.csv",
			expectedHeader: "ID,Type,SKU,Name,Published,Visibility in catalog,Description,Regular price,Categories,Images,Stock,In stock?,Attribute 1 name,Attribute 1 value(s),Attribute 2 name,Attribute 2 value(s)",
			expectedCell:   "simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			targetExporter, err := New(tt.target)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if targetExporter.Filename() != tt.expectedFile {
				t.Fatalf("Filename() = %q, want %q", targetExporter.Filename(), tt.expectedFile)
			}

			var buffer bytes.Buffer
			if err := targetExporter.Export(context.Background(), &buffer, []domain.UniversalProduct{product}); err != nil {
				t.Fatalf("Export() error = %v", err)
			}

			reader := csv.NewReader(strings.NewReader(buffer.String()))
			rows, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("read exported csv: %v", err)
			}
			if len(rows) != 2 {
				t.Fatalf("row count = %d, want 2", len(rows))
			}
			if strings.Join(rows[0], ",") != tt.expectedHeader {
				t.Fatalf("header = %q, want %q", strings.Join(rows[0], ","), tt.expectedHeader)
			}
			if !contains(rows[1], tt.expectedCell) {
				t.Fatalf("export row %v does not contain %q", rows[1], tt.expectedCell)
			}
		})
	}
}

func TestNewRejectsUnknownTarget(t *testing.T) {
	_, err := New("unknown_target")
	if err == nil {
		t.Fatal("New() expected error for unknown target")
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
