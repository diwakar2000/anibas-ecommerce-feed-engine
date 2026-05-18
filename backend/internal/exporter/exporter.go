package exporter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/diwakar/anibas-ecommerce-feed-engine/backend/internal/domain"
)

type Exporter interface {
	TargetPlatform() string
	Filename() string
	Export(ctx context.Context, writer io.Writer, products []domain.UniversalProduct) error
}

type csvExporter struct {
	target   string
	filename string
	header   []string
	row      func(domain.UniversalProduct) []string
}

func New(targetPlatform string) (Exporter, error) {
	switch targetPlatform {
	case "facebook_catalog_csv":
		return NewFacebookCatalogCSVExporter(), nil
	case "instagram_shops":
		return csvExporter{
			target:   "instagram_shops",
			filename: "instagram_shops_preview.csv",
			header:   metaHeader(),
			row:      metaRow,
		}, nil
	case "google_merchant_center":
		return csvExporter{
			target:   "google_merchant_center",
			filename: "google_merchant_center_preview.csv",
			header: []string{
				"id",
				"title",
				"description",
				"link",
				"image_link",
				"additional_image_link",
				"availability",
				"price",
				"condition",
				"brand",
				"gtin",
				"mpn",
				"google_product_category",
				"item_group_id",
			},
			row: googleMerchantRow,
		}, nil
	case "tiktok_catalog":
		return csvExporter{
			target:   "tiktok_catalog",
			filename: "tiktok_catalog_preview.csv",
			header: []string{
				"sku_id",
				"title",
				"description",
				"availability",
				"condition",
				"price",
				"currency",
				"link",
				"image_link",
				"additional_image_link",
				"brand",
				"product_type",
				"item_group_id",
				"inventory",
			},
			row: tiktokRow,
		}, nil
	case "shopify_csv":
		return csvExporter{
			target:   "shopify_csv",
			filename: "shopify_products_preview.csv",
			header: []string{
				"Handle",
				"Title",
				"Body (HTML)",
				"Vendor",
				"Product Category",
				"Type",
				"Published",
				"Option1 Name",
				"Option1 Value",
				"Option2 Name",
				"Option2 Value",
				"Variant SKU",
				"Variant Inventory Qty",
				"Variant Price",
				"Image Src",
				"Status",
			},
			row: shopifyRow,
		}, nil
	case "woocommerce_csv":
		return csvExporter{
			target:   "woocommerce_csv",
			filename: "woocommerce_products_preview.csv",
			header: []string{
				"ID",
				"Type",
				"SKU",
				"Name",
				"Published",
				"Visibility in catalog",
				"Description",
				"Regular price",
				"Categories",
				"Images",
				"Stock",
				"In stock?",
				"Attribute 1 name",
				"Attribute 1 value(s)",
				"Attribute 2 name",
				"Attribute 2 value(s)",
			},
			row: woocommerceRow,
		}, nil
	default:
		return nil, fmt.Errorf("target export %q is not implemented", targetPlatform)
	}
}

type FacebookCatalogCSVExporter struct{}

func NewFacebookCatalogCSVExporter() FacebookCatalogCSVExporter {
	return FacebookCatalogCSVExporter{}
}

func (FacebookCatalogCSVExporter) TargetPlatform() string {
	return "facebook_catalog_csv"
}

func (FacebookCatalogCSVExporter) Filename() string {
	return "facebook_catalog_preview.csv"
}

func (FacebookCatalogCSVExporter) Export(ctx context.Context, writer io.Writer, products []domain.UniversalProduct) error {
	return writeCSV(ctx, writer, metaHeader(), products, metaRow)
}

func (e csvExporter) TargetPlatform() string {
	return e.target
}

func (e csvExporter) Filename() string {
	return e.filename
}

func (e csvExporter) Export(ctx context.Context, writer io.Writer, products []domain.UniversalProduct) error {
	return writeCSV(ctx, writer, e.header, products, e.row)
}

func writeCSV(
	ctx context.Context,
	writer io.Writer,
	header []string,
	products []domain.UniversalProduct,
	row func(domain.UniversalProduct) []string,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, product := range products {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := csvWriter.Write(row(product)); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func metaHeader() []string {
	return []string{
		"id",
		"title",
		"description",
		"availability",
		"condition",
		"price",
		"link",
		"image_link",
		"brand",
		"google_product_category",
		"inventory",
	}
}

func metaRow(product domain.UniversalProduct) []string {
	return []string{
		exportID(product),
		product.Title,
		product.Description,
		product.Availability,
		product.Condition,
		priceWithCurrency(product),
		product.ProductURL,
		product.ImageURL,
		product.Brand,
		product.Category,
		quantityValue(product),
	}
}

func googleMerchantRow(product domain.UniversalProduct) []string {
	return []string{
		exportID(product),
		product.Title,
		product.Description,
		product.ProductURL,
		product.ImageURL,
		strings.Join(product.AdditionalImageURLs, ","),
		product.Availability,
		priceWithCurrency(product),
		product.Condition,
		product.Brand,
		product.GTIN,
		product.MPN,
		product.Category,
		product.VariantGroupID,
	}
}

func tiktokRow(product domain.UniversalProduct) []string {
	return []string{
		exportID(product),
		product.Title,
		product.Description,
		product.Availability,
		product.Condition,
		product.Price,
		product.Currency,
		product.ProductURL,
		product.ImageURL,
		strings.Join(product.AdditionalImageURLs, ","),
		product.Brand,
		product.Category,
		product.VariantGroupID,
		quantityValue(product),
	}
}

func shopifyRow(product domain.UniversalProduct) []string {
	return []string{
		shopifyHandle(product),
		product.Title,
		product.Description,
		product.Brand,
		product.Category,
		product.Category,
		"TRUE",
		firstNonEmpty(product.Option1Name, "Title"),
		firstNonEmpty(product.Option1Value, "Default Title"),
		product.Option2Name,
		product.Option2Value,
		product.SKU,
		quantityValue(product),
		product.Price,
		product.ImageURL,
		"active",
	}
}

func woocommerceRow(product domain.UniversalProduct) []string {
	images := append([]string{product.ImageURL}, product.AdditionalImageURLs...)
	return []string{
		product.ID,
		"simple",
		product.SKU,
		product.Title,
		"1",
		"visible",
		product.Description,
		product.Price,
		product.Category,
		joinNonEmpty(images, ","),
		quantityValue(product),
		stockStatus(product),
		product.Option1Name,
		product.Option1Value,
		product.Option2Name,
		product.Option2Value,
	}
}

func exportID(product domain.UniversalProduct) string {
	if product.ID != "" {
		return product.ID
	}

	return product.SKU
}

func priceWithCurrency(product domain.UniversalProduct) string {
	if product.Price == "" || product.Currency == "" {
		return product.Price
	}

	return fmt.Sprintf("%s %s", product.Price, product.Currency)
}

func quantityValue(product domain.UniversalProduct) string {
	if product.Quantity == 0 {
		return ""
	}

	return strconv.Itoa(product.Quantity)
}

func stockStatus(product domain.UniversalProduct) string {
	if strings.EqualFold(product.Availability, "out of stock") || product.Quantity < 0 {
		return "0"
	}

	return "1"
}

func shopifyHandle(product domain.UniversalProduct) string {
	return slug(firstNonEmpty(product.ID, product.SKU, product.Title))
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	builder.Grow(len(value))
	lastDash := false

	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func joinNonEmpty(values []string, separator string) string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return strings.Join(result, separator)
}
