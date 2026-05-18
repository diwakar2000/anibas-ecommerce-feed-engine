package importer

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

const DefaultPreviewLimit = 20

type PreviewRow struct {
	RowNumber int               `json:"row_number"`
	Values    map[string]string `json:"values"`
}

type Preview struct {
	Columns  []string     `json:"columns"`
	Rows     []PreviewRow `json:"preview_rows"`
	RowCount int          `json:"row_count"`
}

type CSVImporter struct{}

func NewCSVImporter() CSVImporter {
	return CSVImporter{}
}

func (CSVImporter) Parse(reader io.Reader, previewLimit int) (Preview, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Preview{}, errors.New("CSV file is empty")
		}
		return Preview{}, fmt.Errorf("read CSV header: %w", err)
	}

	columns, err := normalizeHeader(header)
	if err != nil {
		return Preview{}, err
	}

	preview := Preview{
		Columns: columns,
		Rows:    make([]PreviewRow, 0, previewLimit),
	}

	for {
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return Preview{}, fmt.Errorf("read CSV row %d: %w", preview.RowCount+2, err)
		}

		preview.RowCount++
		if len(preview.Rows) < previewLimit {
			preview.Rows = append(preview.Rows, PreviewRow{
				RowNumber: preview.RowCount + 1,
				Values:    rowValues(columns, record),
			})
		}
	}

	return preview, nil
}

func normalizeHeader(header []string) ([]string, error) {
	columns := make([]string, len(header))
	seen := make(map[string]struct{}, len(header))

	for index, column := range header {
		value := strings.TrimSpace(strings.TrimPrefix(column, "\ufeff"))
		if value == "" {
			return nil, fmt.Errorf("CSV header column %d is empty", index+1)
		}

		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("CSV header contains duplicate column %q", value)
		}

		seen[key] = struct{}{}
		columns[index] = value
	}

	return columns, nil
}

func rowValues(columns []string, record []string) map[string]string {
	values := make(map[string]string, len(columns))
	for index, column := range columns {
		if index >= len(record) {
			values[column] = ""
			continue
		}
		values[column] = strings.TrimSpace(record[index])
	}

	return values
}
