package csvsource

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// Source holds parsed CSV data with normalized header names.
type Source struct {
	Headers []string            // normalized column names, in order
	Records []map[string]string // each row as field name → value
}

// Parse reads CSV from r and returns a Source. delimiter is the field separator
// (typically ','). When noHeader is true, the first row is treated as data and
// columns are named "col0", "col1", etc.
func Parse(r io.Reader, delimiter rune, noHeader bool) (*Source, error) {
	cr := csv.NewReader(r)
	cr.Comma = delimiter
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1 // allow variable number of fields

	rows, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}
	if len(rows) == 0 {
		return &Source{}, nil
	}

	var headers []string
	var dataRows [][]string

	if noHeader {
		for i := range rows[0] {
			headers = append(headers, fmt.Sprintf("col%d", i))
		}
		dataRows = rows
	} else {
		for _, h := range rows[0] {
			headers = append(headers, normalizeHeader(h))
		}
		dataRows = rows[1:]
	}

	records := make([]map[string]string, 0, len(dataRows))
	for _, row := range dataRows {
		rec := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				rec[h] = strings.TrimSpace(row[i])
			}
		}
		records = append(records, rec)
	}

	return &Source{Headers: headers, Records: records}, nil
}

// normalizeHeader lowercases s and replaces spaces with underscores.
func normalizeHeader(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return strings.Map(func(r rune) rune {
		if r == ' ' {
			return '_'
		}
		return r
	}, s)
}
