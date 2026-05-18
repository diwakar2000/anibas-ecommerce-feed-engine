package importer

import (
	"strings"
	"testing"
)

func TestCSVImporterParseReturnsColumnsAndPreview(t *testing.T) {
	input := "sku,title,price\nANI-1,Shirt,19.99\nANI-2,Hat,9.99\n"

	preview, err := NewCSVImporter().Parse(strings.NewReader(input), 1)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if preview.RowCount != 2 {
		t.Fatalf("RowCount = %d, want 2", preview.RowCount)
	}
	if len(preview.Rows) != 1 {
		t.Fatalf("preview rows = %d, want 1", len(preview.Rows))
	}
	if preview.Columns[0] != "sku" {
		t.Fatalf("first column = %q, want sku", preview.Columns[0])
	}
	if preview.Rows[0].Values["title"] != "Shirt" {
		t.Fatalf("preview title = %q, want Shirt", preview.Rows[0].Values["title"])
	}
}

func TestCSVImporterRejectsDuplicateColumns(t *testing.T) {
	input := "sku,SKU\nANI-1,ANI-1\n"

	_, err := NewCSVImporter().Parse(strings.NewReader(input), DefaultPreviewLimit)
	if err == nil {
		t.Fatal("Parse() error = nil, want duplicate column error")
	}
}
