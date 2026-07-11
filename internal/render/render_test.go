package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/victorjacobs/label-maker/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		LabelW: 63.5, LabelH: 38.1,
		PageW: 210, PageH: 297,
		MarginTop: 15, MarginLeft: 7,
		GapX: 2.5, GapY: 0,
		Padding:  2,
		Align:    "left",
		VAlign:   "middle",
		Copies:   1,
		FontSize: 10,
	}
}

func TestRender_smokeTest(t *testing.T) {
	cfg := testConfig()
	r := New(cfg)

	labels := []Label{
		{Lines: []string{"Jane Doe", "12 Baker St", "1000 Brussels"}},
		{Lines: []string{"Ømer Åberg", "5 Rue Neuve", "4000 Liège"}},
	}

	var buf bytes.Buffer
	if err := r.Render(labels, &buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	got := buf.Bytes()
	if len(got) < 100 {
		t.Fatalf("PDF too small (%d bytes)", len(got))
	}
	if !bytes.HasPrefix(got, []byte("%PDF-")) {
		t.Errorf("output does not start with %%PDF-; got: %q", got[:min(20, len(got))])
	}
}

func TestRender_emptyLabels(t *testing.T) {
	cfg := testConfig()
	r := New(cfg)

	var buf bytes.Buffer
	if err := r.Render(nil, &buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("empty labels should still produce a valid PDF")
	}
}

func TestBuildLabels_defaultTemplate(t *testing.T) {
	cfg := testConfig()
	records := []map[string]string{
		{"name": "Jane", "city": "Brussels"},
	}
	headers := []string{"name", "city"}

	labels, err := BuildLabels(records, headers, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 1 {
		t.Fatalf("want 1 label, got %d", len(labels))
	}
	if labels[0].Lines[0] != "Jane" || labels[0].Lines[1] != "Brussels" {
		t.Errorf("lines = %v", labels[0].Lines)
	}
}

func TestBuildLabels_customTemplate(t *testing.T) {
	cfg := testConfig()
	cfg.Template = `{{.name}}\n{{.city}}`
	records := []map[string]string{
		{"name": "Jane", "city": "Brussels"},
	}
	labels, err := BuildLabels(records, []string{"name", "city"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels[0].Lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %v", len(labels[0].Lines), labels[0].Lines)
	}
}

func TestBuildLabels_copies(t *testing.T) {
	cfg := testConfig()
	cfg.Copies = 3
	records := []map[string]string{{"name": "Jane"}}
	labels, err := BuildLabels(records, []string{"name"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 3 {
		t.Errorf("want 3 copies, got %d", len(labels))
	}
}

func TestBuildLabels_copiesColumn(t *testing.T) {
	cfg := testConfig()
	cfg.CopiesColumn = "qty"
	records := []map[string]string{
		{"name": "Jane", "qty": "2"},
		{"name": "Bob", "qty": "1"},
	}
	labels, err := BuildLabels(records, []string{"name", "qty"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 3 { // 2 + 1
		t.Errorf("want 3 labels (2+1), got %d", len(labels))
	}
}

func TestExpandEscapes(t *testing.T) {
	got := expandEscapes(`hello\nworld`)
	if !strings.Contains(got, "\n") {
		t.Errorf("expandEscapes did not expand \\n: %q", got)
	}
}

func TestSplitLines_trims(t *testing.T) {
	lines := splitLines("  hello  \n\n  world  \n")
	if len(lines) != 2 || lines[0] != "hello" || lines[1] != "world" {
		t.Errorf("splitLines = %v", lines)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
