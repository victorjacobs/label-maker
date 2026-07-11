package csvsource

import (
	"strings"
	"testing"
)

func TestParse_basic(t *testing.T) {
	input := "name,street,postcode,city\nJane Doe,12 Baker St,1000,Brussels\n"
	src, err := Parse(strings.NewReader(input), ',', false)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.Records) != 1 {
		t.Fatalf("want 1 record, got %d", len(src.Records))
	}
	if src.Records[0]["name"] != "Jane Doe" {
		t.Errorf("name = %q", src.Records[0]["name"])
	}
	if src.Records[0]["city"] != "Brussels" {
		t.Errorf("city = %q", src.Records[0]["city"])
	}
}

func TestParse_headerNormalization(t *testing.T) {
	input := "Post Code,First Name\n1000,Jane\n"
	src, err := Parse(strings.NewReader(input), ',', false)
	if err != nil {
		t.Fatal(err)
	}
	if src.Headers[0] != "post_code" {
		t.Errorf("header[0] = %q, want post_code", src.Headers[0])
	}
	if src.Headers[1] != "first_name" {
		t.Errorf("header[1] = %q, want first_name", src.Headers[1])
	}
	if src.Records[0]["post_code"] != "1000" {
		t.Errorf("post_code = %q", src.Records[0]["post_code"])
	}
}

func TestParse_noHeader(t *testing.T) {
	input := "Jane,Brussels\nBob,Liège\n"
	src, err := Parse(strings.NewReader(input), ',', true)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.Records) != 2 {
		t.Fatalf("want 2 records, got %d", len(src.Records))
	}
	if src.Headers[0] != "col0" || src.Headers[1] != "col1" {
		t.Errorf("headers = %v", src.Headers)
	}
	if src.Records[0]["col0"] != "Jane" {
		t.Errorf("col0 = %q", src.Records[0]["col0"])
	}
}

func TestParse_customDelimiter(t *testing.T) {
	input := "name;city\nJane;Brussels\n"
	src, err := Parse(strings.NewReader(input), ';', false)
	if err != nil {
		t.Fatal(err)
	}
	if src.Records[0]["name"] != "Jane" {
		t.Errorf("name = %q", src.Records[0]["name"])
	}
}

func TestParse_utf8(t *testing.T) {
	input := "name,city\nØmer Åberg,Liège\n"
	src, err := Parse(strings.NewReader(input), ',', false)
	if err != nil {
		t.Fatal(err)
	}
	if src.Records[0]["name"] != "Ømer Åberg" {
		t.Errorf("name = %q", src.Records[0]["name"])
	}
	if src.Records[0]["city"] != "Liège" {
		t.Errorf("city = %q", src.Records[0]["city"])
	}
}

func TestParse_quotedFieldWithNewline(t *testing.T) {
	// encoding/csv handles quoted fields with embedded newlines
	input := "name,address\nJane,\"12 Baker St\nLondon\"\n"
	src, err := Parse(strings.NewReader(input), ',', false)
	if err != nil {
		t.Fatal(err)
	}
	if src.Records[0]["address"] != "12 Baker St\nLondon" {
		t.Errorf("address = %q", src.Records[0]["address"])
	}
}

func TestParse_empty(t *testing.T) {
	src, err := Parse(strings.NewReader(""), ',', false)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.Records) != 0 {
		t.Errorf("want 0 records, got %d", len(src.Records))
	}
}

func TestNormalizeHeader(t *testing.T) {
	cases := [][2]string{
		{"Post Code", "post_code"},
		{"CITY", "city"},
		{"firstName", "firstname"},
		{"  name  ", "name"},
	}
	for _, tc := range cases {
		got := normalizeHeader(tc[0])
		if got != tc[1] {
			t.Errorf("normalizeHeader(%q) = %q, want %q", tc[0], got, tc[1])
		}
	}
}
