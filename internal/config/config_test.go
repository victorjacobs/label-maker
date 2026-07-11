package config

import (
	"testing"
)

func TestParsePage(t *testing.T) {
	cases := []struct {
		input   string
		wantW   float64
		wantH   float64
		wantErr bool
	}{
		{"a4", 210, 297, false},
		{"A4", 210, 297, false},
		{"letter", 215.9, 279.4, false},
		{"210x297", 210, 297, false},
		{"100.5x200.5", 100.5, 200.5, false},
		{"bogus", 0, 0, true},
		{"axb", 0, 0, true},
	}
	for _, tc := range cases {
		w, h, err := ParsePage(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParsePage(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && (w != tc.wantW || h != tc.wantH) {
			t.Errorf("ParsePage(%q) = %.1fx%.1f, want %.1fx%.1f", tc.input, w, h, tc.wantW, tc.wantH)
		}
	}
}

func TestValidate(t *testing.T) {
	good := &Config{
		LabelW: 63.5, LabelH: 38.1,
		PageW: 210, PageH: 297,
		Copies: 1, Align: "left", VAlign: "middle",
	}
	if err := good.Validate(); err != nil {
		t.Errorf("valid config failed: %v", err)
	}

	// Label wider than printable area
	wide := *good
	wide.LabelW = 300
	if err := wide.Validate(); err == nil {
		t.Error("expected error for label wider than page")
	}

	// Bad align
	badAlign := *good
	badAlign.Align = "justify"
	if err := badAlign.Validate(); err == nil {
		t.Error("expected error for bad align")
	}

	// Bad valign
	badVAlign := *good
	badVAlign.VAlign = "center"
	if err := badVAlign.Validate(); err == nil {
		t.Error("expected error for bad valign")
	}

	// Copies < 1
	noCopies := *good
	noCopies.Copies = 0
	if err := noCopies.Validate(); err == nil {
		t.Error("expected error for copies=0")
	}
}
