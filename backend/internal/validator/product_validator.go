package validator

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
)

type Level string

const (
	ErrorLevel   Level = "error"
	WarningLevel Level = "warning"
)

type Issue struct {
	Level     Level  `json:"level"`
	RowNumber int    `json:"row_number,omitempty"`
	Field     string `json:"field,omitempty"`
	Message   string `json:"message"`
}

type Report struct {
	Status   string  `json:"status"`
	Errors   []Issue `json:"errors"`
	Warnings []Issue `json:"warnings"`
}

func ValidateProducts(products []domain.UniversalProduct) Report {
	report := Report{
		Status:   "valid",
		Errors:   make([]Issue, 0),
		Warnings: make([]Issue, 0),
	}
	seenSKUs := make(map[string]int, len(products))

	for index, product := range products {
		rowNumber := index + 1
		report.Errors = append(report.Errors, requiredFieldErrors(rowNumber, product)...)
		report.Errors = append(report.Errors, priceErrors(rowNumber, product.Price)...)
		report.Errors = append(report.Errors, quantityErrors(rowNumber, product.Quantity)...)
		report.Errors = append(report.Errors, imageURLErrors(rowNumber, product.ImageURL)...)

		sku := strings.ToLower(strings.TrimSpace(product.SKU))
		if sku == "" {
			continue
		}

		if firstRow, exists := seenSKUs[sku]; exists {
			report.Errors = append(report.Errors, Issue{
				Level:     ErrorLevel,
				RowNumber: rowNumber,
				Field:     "sku",
				Message:   fmt.Sprintf("SKU duplicates row %d", firstRow),
			})
			continue
		}

		seenSKUs[sku] = rowNumber
	}

	if len(report.Errors) > 0 {
		report.Status = "invalid"
	}

	return report
}

func requiredFieldErrors(rowNumber int, product domain.UniversalProduct) []Issue {
	values := map[string]string{
		"sku":          product.SKU,
		"title":        product.Title,
		"description":  product.Description,
		"price":        product.Price,
		"currency":     product.Currency,
		"condition":    product.Condition,
		"image_url":    product.ImageURL,
		"product_url":  product.ProductURL,
		"availability": product.Availability,
	}

	issues := make([]Issue, 0)
	for field, value := range values {
		if strings.TrimSpace(value) == "" {
			issues = append(issues, Issue{
				Level:     ErrorLevel,
				RowNumber: rowNumber,
				Field:     field,
				Message:   fmt.Sprintf("%s is required", field),
			})
		}
	}

	return issues
}

func priceErrors(rowNumber int, price string) []Issue {
	if strings.TrimSpace(price) == "" {
		return nil
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(price), 64)
	if err != nil || value <= 0 {
		return []Issue{{
			Level:     ErrorLevel,
			RowNumber: rowNumber,
			Field:     "price",
			Message:   "price must be a positive decimal number",
		}}
	}

	return nil
}

func quantityErrors(rowNumber int, quantity int) []Issue {
	if quantity < 0 {
		return []Issue{{
			Level:     ErrorLevel,
			RowNumber: rowNumber,
			Field:     "quantity",
			Message:   "quantity cannot be negative",
		}}
	}

	return nil
}

func imageURLErrors(rowNumber int, imageURL string) []Issue {
	if strings.TrimSpace(imageURL) == "" {
		return nil
	}

	parsed, err := url.ParseRequestURI(imageURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return []Issue{{
			Level:     ErrorLevel,
			RowNumber: rowNumber,
			Field:     "image_url",
			Message:   "image_url must be a valid absolute URL",
		}}
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return []Issue{{
			Level:     ErrorLevel,
			RowNumber: rowNumber,
			Field:     "image_url",
			Message:   "image_url must use http or https",
		}}
	}

	return nil
}
