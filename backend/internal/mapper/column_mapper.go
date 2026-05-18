package mapper

import (
	"sort"
	"strings"
	"unicode"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
)

type FieldMappingSuggestion struct {
	Field        string  `json:"field"`
	SourceColumn string  `json:"source_column"`
	Confidence   float64 `json:"confidence"`
	Reason       string  `json:"reason"`
}

type alias struct {
	value  string
	reason string
}

var aliasesByField = map[string][]alias{
	"id": {
		{value: "id", reason: "exact product id header"},
		{value: "product id", reason: "common marketplace product id header"},
		{value: "item id", reason: "common marketplace item id header"},
	},
	"sku": {
		{value: "sku", reason: "exact SKU header"},
		{value: "seller sku", reason: "seller SKU header"},
		{value: "variant sku", reason: "Shopify variant SKU header"},
		{value: "product sku", reason: "product SKU header"},
		{value: "merchant sku", reason: "merchant SKU header"},
	},
	"title": {
		{value: "title", reason: "exact title header"},
		{value: "name", reason: "common product name header"},
		{value: "product title", reason: "product title header"},
		{value: "product name", reason: "product name header"},
		{value: "item title", reason: "marketplace item title header"},
	},
	"description": {
		{value: "description", reason: "exact description header"},
		{value: "product description", reason: "product description header"},
		{value: "body html", reason: "Shopify product body header"},
		{value: "body", reason: "product body header"},
		{value: "desc", reason: "short description header"},
	},
	"price": {
		{value: "price", reason: "exact price header"},
		{value: "regular price", reason: "WooCommerce regular price header"},
		{value: "sale price", reason: "WooCommerce sale price header"},
		{value: "variant price", reason: "Shopify variant price header"},
		{value: "retail price", reason: "retail price header"},
		{value: "amount", reason: "price amount header"},
	},
	"currency": {
		{value: "currency", reason: "exact currency header"},
		{value: "currency code", reason: "currency code header"},
		{value: "price currency", reason: "price currency header"},
	},
	"quantity": {
		{value: "quantity", reason: "exact quantity header"},
		{value: "qty", reason: "quantity abbreviation"},
		{value: "inventory", reason: "inventory header"},
		{value: "stock", reason: "stock header"},
		{value: "stock quantity", reason: "WooCommerce stock quantity header"},
		{value: "inventory quantity", reason: "Shopify inventory quantity header"},
	},
	"condition": {
		{value: "condition", reason: "exact condition header"},
		{value: "product condition", reason: "product condition header"},
	},
	"brand": {
		{value: "brand", reason: "exact brand header"},
		{value: "vendor", reason: "Shopify vendor header"},
		{value: "manufacturer", reason: "manufacturer header"},
		{value: "maker", reason: "maker header"},
	},
	"gtin": {
		{value: "gtin", reason: "exact GTIN header"},
		{value: "global trade item number", reason: "GTIN descriptive header"},
		{value: "upc", reason: "UPC product identifier header"},
		{value: "ean", reason: "EAN product identifier header"},
		{value: "barcode", reason: "barcode product identifier header"},
		{value: "variant barcode", reason: "Shopify variant barcode header"},
	},
	"mpn": {
		{value: "mpn", reason: "exact MPN header"},
		{value: "manufacturer part number", reason: "MPN descriptive header"},
		{value: "part number", reason: "manufacturer part number header"},
		{value: "model number", reason: "model number header"},
	},
	"category": {
		{value: "category", reason: "exact category header"},
		{value: "product category", reason: "product category header"},
		{value: "google product category", reason: "Google Merchant category header"},
		{value: "fb product category", reason: "Facebook category header"},
		{value: "product type", reason: "Shopify product type header"},
		{value: "type", reason: "commerce product type header"},
	},
	"image_url": {
		{value: "image url", reason: "image URL header"},
		{value: "image_url", reason: "image URL header"},
		{value: "image link", reason: "Facebook and Google image link header"},
		{value: "image_link", reason: "Facebook and Google image link header"},
		{value: "image src", reason: "Shopify image source header"},
		{value: "main image", reason: "primary image header"},
		{value: "primary image", reason: "primary image header"},
		{value: "product image", reason: "product image header"},
	},
	"additional_image_urls": {
		{value: "additional image urls", reason: "additional image URLs header"},
		{value: "additional_image_urls", reason: "additional image URLs header"},
		{value: "additional image link", reason: "Facebook and Google additional image link header"},
		{value: "additional_image_link", reason: "Facebook and Google additional image link header"},
		{value: "additional images", reason: "additional images header"},
		{value: "gallery", reason: "gallery image header"},
	},
	"product_url": {
		{value: "product url", reason: "product URL header"},
		{value: "product_url", reason: "product URL header"},
		{value: "link", reason: "Facebook and Google product link header"},
		{value: "url", reason: "URL header"},
		{value: "product link", reason: "product link header"},
		{value: "permalink", reason: "WooCommerce permalink header"},
	},
	"availability": {
		{value: "availability", reason: "exact availability header"},
		{value: "stock status", reason: "WooCommerce stock status header"},
		{value: "availability status", reason: "availability status header"},
	},
	"weight": {
		{value: "weight", reason: "exact weight header"},
		{value: "shipping weight", reason: "shipping weight header"},
		{value: "variant grams", reason: "Shopify variant weight header"},
	},
	"variant_group_id": {
		{value: "variant group id", reason: "variant group header"},
		{value: "variant_group_id", reason: "variant group header"},
		{value: "item group id", reason: "Facebook and Google variant group header"},
		{value: "item_group_id", reason: "Facebook and Google variant group header"},
		{value: "parent id", reason: "parent product header"},
	},
	"option_1_name": {
		{value: "option 1 name", reason: "Shopify option 1 name header"},
		{value: "option1 name", reason: "Shopify option 1 name header"},
		{value: "option_1_name", reason: "option 1 name header"},
	},
	"option_1_value": {
		{value: "option 1 value", reason: "Shopify option 1 value header"},
		{value: "option1 value", reason: "Shopify option 1 value header"},
		{value: "option_1_value", reason: "option 1 value header"},
	},
	"option_2_name": {
		{value: "option 2 name", reason: "Shopify option 2 name header"},
		{value: "option2 name", reason: "Shopify option 2 name header"},
		{value: "option_2_name", reason: "option 2 name header"},
	},
	"option_2_value": {
		{value: "option 2 value", reason: "Shopify option 2 value header"},
		{value: "option2 value", reason: "Shopify option 2 value header"},
		{value: "option_2_value", reason: "option 2 value header"},
	},
	"source_platform": {
		{value: "source platform", reason: "source platform header"},
		{value: "platform", reason: "platform header"},
	},
	"created_at": {
		{value: "created at", reason: "created timestamp header"},
		{value: "created_at", reason: "created timestamp header"},
		{value: "date created", reason: "created date header"},
	},
	"updated_at": {
		{value: "updated at", reason: "updated timestamp header"},
		{value: "updated_at", reason: "updated timestamp header"},
		{value: "date modified", reason: "modified date header"},
	},
}

func SuggestFieldMappings(columns []string, fields []domain.FieldDefinition) []FieldMappingSuggestion {
	usedColumns := make(map[string]struct{}, len(columns))
	suggestions := make([]FieldMappingSuggestion, 0, len(fields))

	for _, field := range fields {
		suggestion, ok := bestSuggestion(field.Name, columns, usedColumns)
		if !ok {
			continue
		}

		usedColumns[normalize(suggestion.SourceColumn)] = struct{}{}
		suggestions = append(suggestions, suggestion)
	}

	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].Field < suggestions[j].Field
	})

	return suggestions
}

func CanonicalizeFieldMapping(mapping map[string]string, columns []string) map[string]string {
	canonicalColumns := make(map[string]string, len(columns))
	for _, column := range columns {
		canonicalColumns[normalize(column)] = column
	}

	canonicalMapping := make(map[string]string, len(mapping))
	for field, sourceColumn := range mapping {
		normalizedColumn := normalize(sourceColumn)
		if normalizedColumn == "" {
			canonicalMapping[field] = ""
			continue
		}

		if canonicalColumn, ok := canonicalColumns[normalizedColumn]; ok {
			canonicalMapping[field] = canonicalColumn
			continue
		}

		canonicalMapping[field] = strings.TrimSpace(sourceColumn)
	}

	return canonicalMapping
}

func bestSuggestion(field string, columns []string, usedColumns map[string]struct{}) (FieldMappingSuggestion, bool) {
	best := FieldMappingSuggestion{}
	for _, column := range columns {
		normalizedColumn := normalize(column)
		if _, exists := usedColumns[normalizedColumn]; exists {
			continue
		}

		confidence, reason := scoreColumn(field, normalizedColumn)
		if confidence <= best.Confidence {
			continue
		}

		best = FieldMappingSuggestion{
			Field:        field,
			SourceColumn: column,
			Confidence:   confidence,
			Reason:       reason,
		}
	}

	if best.Confidence < 0.72 {
		return FieldMappingSuggestion{}, false
	}

	return best, true
}

func scoreColumn(field string, normalizedColumn string) (float64, string) {
	if normalizedColumn == normalize(field) {
		return 0.98, "matches Universal Product Schema field"
	}

	for _, alias := range aliasesByField[field] {
		normalizedAlias := normalize(alias.value)
		if normalizedColumn == normalizedAlias {
			return 0.95, alias.reason
		}
		if tokenCount(normalizedAlias) > 1 && containsTokens(normalizedColumn, normalizedAlias) {
			return 0.82, alias.reason
		}
	}

	return 0, ""
}

func containsTokens(value string, candidate string) bool {
	valueTokens := tokenSet(value)
	candidateTokens := tokenSet(candidate)
	if len(candidateTokens) == 0 || len(candidateTokens) > len(valueTokens) {
		return false
	}

	for token := range candidateTokens {
		if _, ok := valueTokens[token]; !ok {
			return false
		}
	}

	return true
}

func tokenSet(value string) map[string]struct{} {
	tokens := strings.Fields(value)
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		set[token] = struct{}{}
	}

	return set
}

func tokenCount(value string) int {
	return len(strings.Fields(value))
}

func normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	builder.Grow(len(value))

	lastWasSpace := true
	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			lastWasSpace = false
			continue
		}

		if !lastWasSpace {
			builder.WriteByte(' ')
			lastWasSpace = true
		}
	}

	return strings.TrimSpace(builder.String())
}
