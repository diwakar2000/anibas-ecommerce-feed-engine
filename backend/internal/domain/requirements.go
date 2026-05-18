package domain

type FieldRequirement struct {
	Field string `json:"field"`
	Level string `json:"level"`
	Note  string `json:"note,omitempty"`
}

type FieldRequirementGroup struct {
	Fields []string `json:"fields"`
	Min    int      `json:"min"`
	Level  string   `json:"level"`
	Note   string   `json:"note"`
}

type TargetRequirement struct {
	ID                 string                  `json:"id"`
	Label              string                  `json:"label"`
	RequiredFields     []FieldRequirement      `json:"required_fields"`
	ConditionalFields  []FieldRequirement      `json:"conditional_fields"`
	RecommendedFields  []FieldRequirement      `json:"recommended_fields"`
	RequirementGroups  []FieldRequirementGroup `json:"requirement_groups"`
	UnsupportedColumns []string                `json:"unsupported_columns,omitempty"`
	Notes              []string                `json:"notes,omitempty"`
}

func TargetRequirements() []TargetRequirement {
	requirements := []TargetRequirement{
		{
			ID:             "facebook_catalog_csv",
			Label:          "Facebook Catalog CSV",
			RequiredFields: requirements("title", "description", "availability", "condition", "price", "currency", "product_url", "image_url"),
			RequirementGroups: []FieldRequirementGroup{
				{Fields: []string{"id", "sku"}, Min: 1, Level: "required", Note: "Meta requires a stable product id; SKU can be exported as id when product id is missing."},
				{Fields: []string{"brand", "gtin", "mpn"}, Min: 1, Level: "required", Note: "Meta requires at least one product identifier: brand, GTIN, or MPN."},
			},
			ConditionalFields: []FieldRequirement{
				{Field: "quantity", Level: "conditional", Note: "Required for some Meta commerce surfaces such as checkout, Marketplace, and some Page shop use cases."},
				{Field: "category", Level: "conditional", Note: "Required for some Meta checkout and Marketplace use cases as google_product_category."},
			},
		},
		{
			ID:             "instagram_shops",
			Label:          "Instagram Shops",
			RequiredFields: requirements("title", "description", "availability", "condition", "price", "currency", "product_url", "image_url"),
			RequirementGroups: []FieldRequirementGroup{
				{Fields: []string{"id", "sku"}, Min: 1, Level: "required", Note: "Instagram Shops uses Meta catalog item ids; SKU can be exported as id when product id is missing."},
				{Fields: []string{"brand", "gtin", "mpn"}, Min: 1, Level: "required", Note: "Meta requires at least one product identifier: brand, GTIN, or MPN."},
			},
			ConditionalFields: []FieldRequirement{
				{Field: "quantity", Level: "conditional", Note: "Required for checkout-enabled commerce surfaces."},
				{Field: "category", Level: "conditional", Note: "Required for some checkout use cases as google_product_category."},
			},
		},
		{
			ID:             "google_merchant_center",
			Label:          "Google Merchant Center",
			RequiredFields: requirements("title", "description", "availability", "price", "currency", "product_url", "image_url"),
			RequirementGroups: []FieldRequirementGroup{
				{Fields: []string{"id", "sku"}, Min: 1, Level: "required", Note: "Google requires a stable id; SKU can be exported as id when product id is missing."},
			},
			ConditionalFields: []FieldRequirement{
				{Field: "condition", Level: "conditional", Note: "Required when the product is used or refurbished; optional for new products."},
				{Field: "brand", Level: "conditional", Note: "Required for most new products except some media, books, and recording brands."},
				{Field: "gtin", Level: "conditional", Note: "Required when the product has a known manufacturer-assigned GTIN."},
				{Field: "mpn", Level: "conditional", Note: "Required when the product has no manufacturer-assigned GTIN."},
				{Field: "category", Level: "conditional", Note: "Required for certain product categories and strongly recommended otherwise."},
				{Field: "variant_group_id", Level: "conditional", Note: "Required for variants in several target countries and for free listings variants."},
			},
		},
		{
			ID:             "tiktok_catalog",
			Label:          "TikTok Catalog",
			RequiredFields: requirements("title", "description", "availability", "condition", "price", "currency", "product_url", "image_url", "brand"),
			RequirementGroups: []FieldRequirementGroup{
				{Fields: []string{"id", "sku"}, Min: 1, Level: "required", Note: "TikTok requires sku_id; SKU can be exported as sku_id when product id is missing."},
			},
			RecommendedFields: []FieldRequirement{
				{Field: "category", Level: "recommended", Note: "Useful as product_type or google_product_category for catalog delivery quality."},
				{Field: "variant_group_id", Level: "recommended", Note: "Useful as item_group_id for product variants."},
				{Field: "additional_image_urls", Level: "recommended", Note: "TikTok supports additional image links for richer ads."},
			},
		},
		{
			ID:             "shopify_csv",
			Label:          "Shopify Product CSV",
			RequiredFields: requirements("title"),
			ConditionalFields: []FieldRequirement{
				{Field: "variant_group_id", Level: "conditional", Note: "Shopify requires URL handle when adding variants or updating products. Add a dedicated handle field before production Shopify export."},
				{Field: "option_1_name", Level: "conditional", Note: "Required with option values when variant columns are present."},
				{Field: "option_1_value", Level: "conditional", Note: "Required with option names when variant columns are present."},
			},
			RecommendedFields: []FieldRequirement{
				{Field: "sku", Level: "recommended", Note: "Useful for variant identity and future updates."},
				{Field: "price", Level: "recommended", Note: "Useful as Variant Price."},
				{Field: "image_url", Level: "recommended", Note: "Useful as image source."},
			},
			UnsupportedColumns: []string{"handle"},
			Notes:              []string{"Shopify's URL handle is not yet represented in the Universal Product Schema."},
		},
		{
			ID:    "woocommerce_csv",
			Label: "WooCommerce Product CSV",
			RequiredFields: []FieldRequirement{
				{Field: "title", Level: "required", Note: "Maps to WooCommerce Name."},
			},
			RecommendedFields: []FieldRequirement{
				{Field: "sku", Level: "recommended", Note: "WooCommerce marks SKU as required but can auto-generate it if missing."},
				{Field: "price", Level: "recommended", Note: "Maps to regular price or sale price depending export settings."},
				{Field: "quantity", Level: "recommended", Note: "Maps to stock only when stock management is enabled."},
				{Field: "image_url", Level: "recommended", Note: "Maps to Images; external URLs are supported when directly accessible."},
			},
			ConditionalFields: []FieldRequirement{
				{Field: "variant_group_id", Level: "conditional", Note: "Required when exporting variations that need a parent product reference."},
				{Field: "option_1_name", Level: "conditional", Note: "Required for variable product attributes."},
				{Field: "option_1_value", Level: "conditional", Note: "Required for variable product attributes."},
			},
		},
	}

	for i := range requirements {
		requirements[i].RequiredFields = nonNilFieldRequirements(requirements[i].RequiredFields)
		requirements[i].ConditionalFields = nonNilFieldRequirements(requirements[i].ConditionalFields)
		requirements[i].RecommendedFields = nonNilFieldRequirements(requirements[i].RecommendedFields)
		requirements[i].RequirementGroups = nonNilRequirementGroups(requirements[i].RequirementGroups)
	}

	return requirements
}

func requirements(fields ...string) []FieldRequirement {
	result := make([]FieldRequirement, 0, len(fields))
	for _, field := range fields {
		result = append(result, FieldRequirement{Field: field, Level: "required"})
	}

	return result
}

func nonNilFieldRequirements(items []FieldRequirement) []FieldRequirement {
	if items == nil {
		return []FieldRequirement{}
	}

	return items
}

func nonNilRequirementGroups(items []FieldRequirementGroup) []FieldRequirementGroup {
	if items == nil {
		return []FieldRequirementGroup{}
	}

	return items
}
