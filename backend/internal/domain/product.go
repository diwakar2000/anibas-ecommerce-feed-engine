package domain

import "time"

type UniversalProduct struct {
	ID                  string    `json:"id"`
	SKU                 string    `json:"sku"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Price               string    `json:"price"`
	Currency            string    `json:"currency"`
	Quantity            int       `json:"quantity"`
	Condition           string    `json:"condition"`
	Brand               string    `json:"brand"`
	GTIN                string    `json:"gtin"`
	MPN                 string    `json:"mpn"`
	Category            string    `json:"category"`
	ImageURL            string    `json:"image_url"`
	AdditionalImageURLs []string  `json:"additional_image_urls"`
	ProductURL          string    `json:"product_url"`
	Availability        string    `json:"availability"`
	Weight              string    `json:"weight"`
	VariantGroupID      string    `json:"variant_group_id"`
	Option1Name         string    `json:"option_1_name"`
	Option1Value        string    `json:"option_1_value"`
	Option2Name         string    `json:"option_2_name"`
	Option2Value        string    `json:"option_2_value"`
	SourcePlatform      string    `json:"source_platform"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type FieldDefinition struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

func UniversalProductSchema() []FieldDefinition {
	return []FieldDefinition{
		{Name: "id", Label: "ID"},
		{Name: "sku", Label: "SKU", Required: true},
		{Name: "title", Label: "Title", Required: true},
		{Name: "description", Label: "Description", Required: true},
		{Name: "price", Label: "Price", Required: true},
		{Name: "currency", Label: "Currency", Required: true},
		{Name: "quantity", Label: "Quantity", Required: true},
		{Name: "condition", Label: "Condition", Required: true},
		{Name: "brand", Label: "Brand"},
		{Name: "gtin", Label: "GTIN"},
		{Name: "mpn", Label: "MPN"},
		{Name: "category", Label: "Category"},
		{Name: "image_url", Label: "Image URL", Required: true},
		{Name: "additional_image_urls", Label: "Additional Image URLs"},
		{Name: "product_url", Label: "Product URL", Required: true},
		{Name: "availability", Label: "Availability", Required: true},
		{Name: "weight", Label: "Weight"},
		{Name: "variant_group_id", Label: "Variant Group ID"},
		{Name: "option_1_name", Label: "Option 1 Name"},
		{Name: "option_1_value", Label: "Option 1 Value"},
		{Name: "option_2_name", Label: "Option 2 Name"},
		{Name: "option_2_value", Label: "Option 2 Value"},
		{Name: "source_platform", Label: "Source Platform"},
		{Name: "created_at", Label: "Created At"},
		{Name: "updated_at", Label: "Updated At"},
	}
}
