//go:build js && wasm

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/exporter"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/importer"
	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/mapper"
)

type catalogImport struct {
	ID             int    `json:"id"`
	Filename       string `json:"filename"`
	SourcePlatform string `json:"source_platform"`
	RowCount       int    `json:"row_count"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

type uploadPreview struct {
	CatalogImport      catalogImport                   `json:"catalog_import"`
	Columns            []string                        `json:"columns"`
	PreviewRows        []importer.PreviewRow           `json:"preview_rows"`
	RowCount           int                             `json:"row_count"`
	MappingSuggestions []mapper.FieldMappingSuggestion `json:"mapping_suggestions"`
}

type schemaResponse struct {
	Fields             []domain.FieldDefinition   `json:"fields"`
	TargetRequirements []domain.TargetRequirement `json:"target_requirements"`
}

type mappingValue struct {
	Mode  string `json:"mode"`
	Value string `json:"value"`
}

type transformRequest struct {
	PreviewRows    []importer.PreviewRow   `json:"preview_rows"`
	Mapping        map[string]mappingValue `json:"mapping"`
	TargetPlatform string                  `json:"target_platform"`
	SourcePlatform string                  `json:"source_platform"`
}

type finding struct {
	Level     string `json:"level"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	RowNumber int    `json:"rowNumber,omitempty"`
}

type validationResponse struct {
	Findings []finding `json:"findings"`
}

type exportResponse struct {
	Filename string `json:"filename"`
	CSV      string `json:"csv"`
	RowCount int    `json:"row_count"`
}

type wasmResponse struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

var pricePattern = regexp.MustCompile(`[-+]?\d[\d\s,]*(\.\d+)?|[-+]?\d+(,\d{2})`)
var currencyCodePattern = regexp.MustCompile(`(?i)\b[A-Z]{3}\b`)
var currencySymbols = []string{"$", "€", "£", "¥", "₹", "₦", "₱", "₩", "₺", "₫", "₪", "₽"}

func main() {
	api := js.Global().Get("Object").New()
	api.Set("schema", js.FuncOf(schema))
	api.Set("parseCatalog", js.FuncOf(parseCatalog))
	api.Set("validateCatalog", js.FuncOf(validateCatalog))
	api.Set("exportCatalog", js.FuncOf(exportCatalog))
	api.Set("exportFacebookCatalog", js.FuncOf(exportFacebookCatalog))
	js.Global().Set("anibasWasm", api)

	select {}
}

func schema(_ js.Value, _ []js.Value) interface{} {
	return ok(schemaResponse{
		Fields:             domain.UniversalProductSchema(),
		TargetRequirements: domain.TargetRequirements(),
	})
}

func parseCatalog(_ js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return fail("CSV text is required")
	}

	filename := "catalog.csv"
	if len(args) > 1 && strings.TrimSpace(args[1].String()) != "" {
		filename = strings.TrimSpace(args[1].String())
	}

	sourcePlatform := "generic_csv"
	if len(args) > 2 && strings.TrimSpace(args[2].String()) != "" {
		sourcePlatform = strings.TrimSpace(args[2].String())
	}

	preview, err := importer.NewCSVImporter().Parse(strings.NewReader(args[0].String()), importer.DefaultPreviewLimit)
	if err != nil {
		return fail(err.Error())
	}

	now := time.Now()
	return ok(uploadPreview{
		CatalogImport: catalogImport{
			ID:             int(now.UnixMilli()),
			Filename:       filename,
			SourcePlatform: sourcePlatform,
			RowCount:       preview.RowCount,
			Status:         "parsed",
			CreatedAt:      now.Format(time.RFC3339),
		},
		Columns:            preview.Columns,
		PreviewRows:        preview.Rows,
		RowCount:           preview.RowCount,
		MappingSuggestions: mapper.SuggestFieldMappings(preview.Columns, domain.UniversalProductSchema()),
	})
}

func validateCatalog(_ js.Value, args []js.Value) interface{} {
	request, err := decodeTransformRequest(args)
	if err != nil {
		return fail(err.Error())
	}

	return ok(validationResponse{Findings: validateRows(request)})
}

func exportFacebookCatalog(_ js.Value, args []js.Value) interface{} {
	request, err := decodeTransformRequest(args)
	if err != nil {
		return fail(err.Error())
	}
	request.TargetPlatform = "facebook_catalog_csv"

	return exportRequest(request)
}

func exportCatalog(_ js.Value, args []js.Value) interface{} {
	request, err := decodeTransformRequest(args)
	if err != nil {
		return fail(err.Error())
	}
	if request.TargetPlatform == "" {
		return fail("target platform is required")
	}

	return exportRequest(request)
}

func exportRequest(request transformRequest) interface{} {
	products := make([]domain.UniversalProduct, 0, len(request.PreviewRows))
	for _, row := range request.PreviewRows {
		products = append(products, productFromRow(row.Values, request.Mapping, request.SourcePlatform))
	}

	targetExporter, err := exporter.New(request.TargetPlatform)
	if err != nil {
		return fail(err.Error())
	}

	var buffer bytes.Buffer
	if err := targetExporter.Export(context.Background(), &buffer, products); err != nil {
		return fail(err.Error())
	}

	return ok(exportResponse{
		Filename: targetExporter.Filename(),
		CSV:      buffer.String(),
		RowCount: len(products),
	})
}

func decodeTransformRequest(args []js.Value) (transformRequest, error) {
	if len(args) < 1 {
		return transformRequest{}, fmt.Errorf("catalog transformation payload is required")
	}

	var request transformRequest
	if err := json.Unmarshal([]byte(args[0].String()), &request); err != nil {
		return transformRequest{}, fmt.Errorf("read catalog transformation payload: %w", err)
	}
	if request.Mapping == nil {
		request.Mapping = map[string]mappingValue{}
	}

	return request, nil
}

func validateRows(request transformRequest) []finding {
	result := make([]finding, 0)
	requirement := requirementByID(request.TargetPlatform)
	if requirement.ID == "" {
		result = append(result, finding{
			Level:  "error",
			Title:  "Choose a target format",
			Detail: "A target format is required before validation can check channel requirements.",
		})
		return result
	}

	for _, item := range requirement.RequiredFields {
		if !mappingHasValue(request.Mapping, item.Field) && !canDeriveField(request.PreviewRows, request.Mapping, item.Field) {
			result = append(result, finding{
				Level:  "error",
				Title:  "Map " + labelForField(item.Field),
				Detail: firstNonEmpty(item.Note, fmt.Sprintf("%s is required for %s.", labelForField(item.Field), requirement.Label)),
			})
		}
	}

	for _, group := range requirement.RequirementGroups {
		if group.Level != "required" {
			continue
		}
		if mappedCount(request.Mapping, group.Fields) < group.Min {
			result = append(result, finding{
				Level:  "error",
				Title:  fmt.Sprintf("Map at least %d of %s", group.Min, joinedLabels(group.Fields)),
				Detail: group.Note,
			})
		}
	}

	seenSKU := map[string]int{}
	for _, row := range request.PreviewRows {
		for _, item := range requirement.RequiredFields {
			if !mappingHasValue(request.Mapping, item.Field) && !canDeriveField(request.PreviewRows, request.Mapping, item.Field) {
				continue
			}
			if resolveMappedValue(request.Mapping, item.Field, row.Values) == "" {
				result = append(result, finding{
					Level:     "error",
					Title:     labelForField(item.Field) + " is empty",
					Detail:    fmt.Sprintf("Preview row %d has no value for %s.", row.RowNumber, labelForField(item.Field)),
					RowNumber: row.RowNumber,
				})
			}
		}

		for _, group := range requirement.RequirementGroups {
			if group.Level != "required" || mappedCount(request.Mapping, group.Fields) < group.Min {
				continue
			}
			if valueCount(request.Mapping, group.Fields, row.Values) < group.Min {
				result = append(result, finding{
					Level:     "error",
					Title:     "Identifier value is missing",
					Detail:    fmt.Sprintf("Preview row %d needs at least %d of %s.", row.RowNumber, group.Min, joinedLabels(group.Fields)),
					RowNumber: row.RowNumber,
				})
			}
		}

		if mappingHasValue(request.Mapping, "sku") {
			sku := strings.ToLower(resolveMappedValue(request.Mapping, "sku", row.Values))
			if sku != "" {
				if firstRow, exists := seenSKU[sku]; exists {
					result = append(result, finding{
						Level:     "error",
						Title:     "Duplicate SKU in preview",
						Detail:    fmt.Sprintf("SKU %q appears in rows %d and %d.", resolveMappedValue(request.Mapping, "sku", row.Values), firstRow, row.RowNumber),
						RowNumber: row.RowNumber,
					})
				} else {
					seenSKU[sku] = row.RowNumber
				}
			}
		}

		result = append(result, formatFindings(row, request.Mapping)...)
	}

	return result
}

func formatFindings(row importer.PreviewRow, mapping map[string]mappingValue) []finding {
	result := make([]finding, 0)

	price := resolveMappedValue(mapping, "price", row.Values)
	if price != "" && !isPositiveNumber(price) {
		result = append(result, finding{Level: "error", Title: "Invalid price", Detail: fmt.Sprintf("Preview row %d: %q. Use a positive number such as 19.99.", row.RowNumber, price), RowNumber: row.RowNumber})
	}

	quantity := resolveMappedValue(mapping, "quantity", row.Values)
	if quantity != "" && !isWholeNumber(quantity) {
		result = append(result, finding{Level: "warning", Title: "Invalid quantity", Detail: fmt.Sprintf("Preview row %d: %q. Use a whole number greater than or equal to 0.", row.RowNumber, quantity), RowNumber: row.RowNumber})
	}

	currency := resolveMappedValue(mapping, "currency", row.Values)
	if currency != "" && !isValidCurrency(currency) {
		result = append(result, finding{Level: "warning", Title: "Invalid currency", Detail: fmt.Sprintf("Preview row %d: %q. Use a three-letter ISO code like USD or a currency symbol like $.", row.RowNumber, currency), RowNumber: row.RowNumber})
	}

	for _, field := range []string{"image_url", "product_url"} {
		value := resolveMappedValue(mapping, field, row.Values)
		if value != "" && !isHTTPURL(value) {
			result = append(result, finding{Level: "error", Title: "Invalid " + labelForField(field), Detail: fmt.Sprintf("Preview row %d: %q. Use a full http:// or https:// URL.", row.RowNumber, value), RowNumber: row.RowNumber})
		}
	}

	availability := strings.ToLower(resolveMappedValue(mapping, "availability", row.Values))
	if availability != "" && !oneOf(availability, "in stock", "out of stock", "preorder", "available for order", "discontinued") {
		result = append(result, finding{Level: "warning", Title: "Unusual availability", Detail: fmt.Sprintf("Preview row %d: %q. Use a marketplace-supported availability value.", row.RowNumber, availability), RowNumber: row.RowNumber})
	}

	condition := strings.ToLower(resolveMappedValue(mapping, "condition", row.Values))
	if condition != "" && !oneOf(condition, "new", "used", "refurbished") {
		result = append(result, finding{Level: "warning", Title: "Unusual condition", Detail: fmt.Sprintf("Preview row %d: %q. Use new, used, or refurbished.", row.RowNumber, condition), RowNumber: row.RowNumber})
	}

	return result
}

func productFromRow(values map[string]string, mapping map[string]mappingValue, sourcePlatform string) domain.UniversalProduct {
	quantity, _ := strconv.Atoi(resolveMappedValue(mapping, "quantity", values))
	return domain.UniversalProduct{
		ID:                  resolveMappedValue(mapping, "id", values),
		SKU:                 resolveMappedValue(mapping, "sku", values),
		Title:               resolveMappedValue(mapping, "title", values),
		Description:         resolveMappedValue(mapping, "description", values),
		Price:               resolveMappedValue(mapping, "price", values),
		Currency:            resolveMappedValue(mapping, "currency", values),
		Quantity:            quantity,
		Condition:           resolveMappedValue(mapping, "condition", values),
		Brand:               resolveMappedValue(mapping, "brand", values),
		GTIN:                resolveMappedValue(mapping, "gtin", values),
		MPN:                 resolveMappedValue(mapping, "mpn", values),
		Category:            resolveMappedValue(mapping, "category", values),
		ImageURL:            resolveMappedValue(mapping, "image_url", values),
		AdditionalImageURLs: splitList(resolveMappedValue(mapping, "additional_image_urls", values)),
		ProductURL:          resolveMappedValue(mapping, "product_url", values),
		Availability:        resolveMappedValue(mapping, "availability", values),
		Weight:              resolveMappedValue(mapping, "weight", values),
		VariantGroupID:      resolveMappedValue(mapping, "variant_group_id", values),
		Option1Name:         resolveMappedValue(mapping, "option_1_name", values),
		Option1Value:        resolveMappedValue(mapping, "option_1_value", values),
		Option2Name:         resolveMappedValue(mapping, "option_2_name", values),
		Option2Value:        resolveMappedValue(mapping, "option_2_value", values),
		SourcePlatform:      firstNonEmpty(resolveMappedValue(mapping, "source_platform", values), sourcePlatform),
	}
}

func requirementByID(id string) domain.TargetRequirement {
	for _, requirement := range domain.TargetRequirements() {
		if requirement.ID == id {
			return requirement
		}
	}
	return domain.TargetRequirement{}
}

func canDeriveField(rows []importer.PreviewRow, mapping map[string]mappingValue, field string) bool {
	if field != "currency" || !mappingHasValue(mapping, "price") {
		return false
	}
	for _, row := range rows {
		if resolveMappedValue(mapping, "currency", row.Values) != "" {
			return true
		}
	}
	return false
}

func resolveMappedValue(mapping map[string]mappingValue, field string, values map[string]string) string {
	if field == "currency" {
		if _, ok := mapping[field]; !ok {
			return extractCurrency(rawMappedValue(mapping, "price", values))
		}
	}

	rawValue := rawMappedValue(mapping, field, values)
	switch field {
	case "price":
		if price := extractPrice(rawValue); price != "" {
			return price
		}
	case "currency":
		if currency := extractCurrency(rawValue); currency != "" {
			return currency
		}
	}

	return rawValue
}

func rawMappedValue(mapping map[string]mappingValue, field string, values map[string]string) string {
	item, ok := mapping[field]
	if !ok {
		return ""
	}

	value := strings.TrimSpace(item.Value)
	if strings.EqualFold(item.Mode, "static") {
		return value
	}

	if exact, ok := values[value]; ok {
		return strings.TrimSpace(exact)
	}

	normalizedValue := normalizeColumn(value)
	for column, rowValue := range values {
		if normalizeColumn(column) == normalizedValue {
			return strings.TrimSpace(rowValue)
		}
	}

	return ""
}

func extractPrice(value string) string {
	match := pricePattern.FindString(value)
	if match == "" {
		return ""
	}

	normalized := strings.ReplaceAll(match, " ", "")
	if regexp.MustCompile(`^\d+,\d{2}$`).MatchString(normalized) {
		return strings.ReplaceAll(normalized, ",", ".")
	}

	return strings.ReplaceAll(normalized, ",", "")
}

func extractCurrency(value string) string {
	if code := currencyCodePattern.FindString(value); code != "" {
		return strings.ToUpper(code)
	}
	for _, symbol := range currencySymbols {
		if strings.Contains(value, symbol) {
			return symbol
		}
	}
	return ""
}

func normalizeColumn(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), "_", " "), "-", " "))), " ")
}

func mappingHasValue(mapping map[string]mappingValue, field string) bool {
	return strings.TrimSpace(mapping[field].Value) != ""
}

func mappedCount(mapping map[string]mappingValue, fields []string) int {
	count := 0
	for _, field := range fields {
		if mappingHasValue(mapping, field) {
			count++
		}
	}
	return count
}

func valueCount(mapping map[string]mappingValue, fields []string, values map[string]string) int {
	count := 0
	for _, field := range fields {
		if resolveMappedValue(mapping, field, values) != "" {
			count++
		}
	}
	return count
}

func labelForField(field string) string {
	for _, definition := range domain.UniversalProductSchema() {
		if definition.Name == field {
			return definition.Label
		}
	}
	return field
}

func joinedLabels(fields []string) string {
	labels := make([]string, 0, len(fields))
	for _, field := range fields {
		labels = append(labels, labelForField(field))
	}
	return strings.Join(labels, ", ")
}

func splitList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '|' || r == ';'
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func isPositiveNumber(value string) bool {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return err == nil && amount > 0
}

func isWholeNumber(value string) bool {
	amount, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil && amount >= 0
}

func isValidCurrency(value string) bool {
	trimmed := strings.TrimSpace(value)
	if regexp.MustCompile(`^[A-Za-z]{3}$`).MatchString(trimmed) {
		return true
	}
	for _, symbol := range currencySymbols {
		if trimmed == symbol {
			return true
		}
	}
	return false
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func oneOf(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func ok(data interface{}) string {
	return encode(wasmResponse{OK: true, Data: data})
}

func fail(message string) string {
	return encode(wasmResponse{OK: false, Error: message})
}

func encode(value interface{}) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf(`{"ok":false,"error":%q}`, err.Error())
	}

	return string(payload)
}
