package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Config holds all layout and rendering parameters.
type Config struct {
	InputPath, OutputPath string
	LabelW, LabelH        float64 // mm
	PageW, PageH          float64 // mm (resolved from preset)
	Columns, Rows         int     // 0 = auto-fit
	MarginTop, MarginLeft float64
	GapX, GapY            float64
	Padding               float64
	Template              string
	FontPath              string  // "" = embedded
	FontSize              float64 // 0 = auto-fit
	Align, VAlign         string
	CopiesColumn          string
	Copies                int
	Skip                  int
	DrawBorder            bool
	Delimiter             rune
	NoHeader              bool
}

var pagePresets = map[string][2]float64{
	"a4":     {210, 297},
	"letter": {215.9, 279.4},
}

// ParsePage resolves a page size string ("a4", "letter", or "WxH") to mm dimensions.
func ParsePage(s string) (w, h float64, err error) {
	if dims, ok := pagePresets[strings.ToLower(strings.TrimSpace(s))]; ok {
		return dims[0], dims[1], nil
	}
	parts := strings.SplitN(s, "x", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unknown page preset %q; use a4, letter, or WxH (mm, e.g. 210x297)", s)
	}
	w, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page width in %q: %w", s, err)
	}
	h, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page height in %q: %w", s, err)
	}
	return w, h, nil
}

// Validate returns an error if the configuration is logically invalid.
func (c *Config) Validate() error {
	if c.LabelW <= 0 || c.LabelH <= 0 {
		return fmt.Errorf("label dimensions must be positive (got %.2f x %.2f mm)", c.LabelW, c.LabelH)
	}
	if c.PageW <= 0 || c.PageH <= 0 {
		return fmt.Errorf("page dimensions must be positive")
	}
	printableW := c.PageW - c.MarginLeft
	printableH := c.PageH - c.MarginTop
	if c.LabelW > printableW {
		return fmt.Errorf("label width %.2f mm exceeds printable width %.2f mm (page %.2f - left margin %.2f)", c.LabelW, printableW, c.PageW, c.MarginLeft)
	}
	if c.LabelH > printableH {
		return fmt.Errorf("label height %.2f mm exceeds printable height %.2f mm (page %.2f - top margin %.2f)", c.LabelH, printableH, c.PageH, c.MarginTop)
	}
	if c.Copies < 1 {
		return fmt.Errorf("--copies must be at least 1 (got %d)", c.Copies)
	}
	if c.Skip < 0 {
		return fmt.Errorf("--skip must be non-negative (got %d)", c.Skip)
	}
	switch c.Align {
	case "left", "center", "right":
	default:
		return fmt.Errorf("--align must be left, center, or right (got %q)", c.Align)
	}
	switch c.VAlign {
	case "top", "middle", "bottom":
	default:
		return fmt.Errorf("--valign must be top, middle, or bottom (got %q)", c.VAlign)
	}
	return nil
}
