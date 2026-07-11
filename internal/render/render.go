package render

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/go-pdf/fpdf"
	"github.com/victorjacobs/label-maker/internal/config"
	"github.com/victorjacobs/label-maker/internal/layout"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	fontFamily  = "goregular"
	maxFontSize = 12.0
	minFontSize = 5.0
	// 1 pt = 25.4/72 mm; multiply by line spacing factor 1.3
	ptToMmFactor = (25.4 / 72.0) * 1.3
)

// Label is a single rendered label: the lines of text to draw.
type Label struct {
	Lines []string
}

// Renderer builds the PDF.
type Renderer struct {
	cfg *config.Config
	pdf *fpdf.Fpdf
}

// New creates a Renderer. Call Render to produce the PDF.
func New(cfg *config.Config) *Renderer {
	return &Renderer{cfg: cfg}
}

// Render processes records and writes a PDF to w.
func (r *Renderer) Render(labels []Label, w io.Writer) error {
	cfg := r.cfg

	pdf := fpdf.NewCustom(&fpdf.InitType{
		UnitStr: "mm",
		Size:    fpdf.SizeType{Wd: cfg.PageW, Ht: cfg.PageH},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	r.pdf = pdf

	if err := r.loadFont(); err != nil {
		return fmt.Errorf("loading font: %w", err)
	}
	pdf.SetFont(fontFamily, "", maxFontSize)

	grid := layout.NewGrid(layout.Config{
		PageW: cfg.PageW, PageH: cfg.PageH,
		LabelW: cfg.LabelW, LabelH: cfg.LabelH,
		MarginTop: cfg.MarginTop, MarginLeft: cfg.MarginLeft,
		GapX: cfg.GapX, GapY: cfg.GapY,
		Columns: cfg.Columns, Rows: cfg.Rows,
	})

	currentPage := -1

	for i, label := range labels {
		slot := grid.Slot(i + cfg.Skip)
		if slot.Page != currentPage {
			pdf.AddPage()
			currentPage = slot.Page
			if cfg.DrawBorder {
				r.drawPageGrid(grid)
			}
		}
		r.renderLabel(slot, label.Lines)
	}

	if currentPage == -1 {
		// No labels — add a blank page so the PDF is still valid.
		pdf.AddPage()
		if cfg.DrawBorder {
			r.drawPageGrid(grid)
		}
	}

	return pdf.Output(w)
}

// BuildLabels expands CSV records into a flat list of labels, applying the
// template, copies column, and default copies count.
func BuildLabels(records []map[string]string, headers []string, cfg *config.Config) ([]Label, error) {
	tmpl, err := buildTemplate(cfg.Template, headers)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var labels []Label
	for _, rec := range records {
		lines, err := executeTemplate(tmpl, rec, headers, cfg.Template == "")
		if err != nil {
			return nil, fmt.Errorf("executing template: %w", err)
		}

		copies := cfg.Copies
		if cfg.CopiesColumn != "" {
			if n, err := parseCopies(rec[cfg.CopiesColumn]); err == nil && n > 0 {
				copies = n
			}
		}

		lbl := Label{Lines: lines}
		for range copies {
			labels = append(labels, lbl)
		}
	}
	return labels, nil
}

// buildTemplate parses and returns the label text/template. When tmplStr is
// empty (default template), it returns nil to signal "use default".
func buildTemplate(tmplStr string, headers []string) (*template.Template, error) {
	if tmplStr == "" {
		return nil, nil
	}
	// Interpret literal \n as newline (shell doesn't expand \n in --template flags).
	expanded := expandEscapes(tmplStr)
	return template.New("label").Parse(expanded)
}

// executeTemplate renders a record into label lines.
// When tmpl is nil the default template (all non-empty columns joined) is used.
func executeTemplate(tmpl *template.Template, rec map[string]string, headers []string, useDefault bool) ([]string, error) {
	var text string
	if useDefault {
		buf := &bytes.Buffer{}
		for _, h := range headers {
			if v := rec[h]; v != "" {
				if buf.Len() > 0 {
					buf.WriteByte('\n')
				}
				buf.WriteString(v)
			}
		}
		text = buf.String()
	} else {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, rec); err != nil {
			return nil, err
		}
		text = buf.String()
	}

	return splitLines(text), nil
}

func splitLines(s string) []string {
	var out []string
	for _, line := range splitNewlines(s) {
		trimmed := trimWhitespace(line)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// drawPageGrid draws the label grid for the current page. Called once per page
// when --draw-border is set.
//
// With gap=0 it draws a single grid line at each boundary so adjacent labels
// don't produce doubled strokes. With a non-zero gap it draws a rectangle for
// each label slot (showing the gap as blank space between outlines).
func (r *Renderer) drawPageGrid(grid layout.Grid) {
	cfg := r.cfg
	r.pdf.SetLineWidth(0.25)

	if cfg.GapX == 0 && cfg.GapY == 0 {
		totalW := float64(grid.Cols) * cfg.LabelW
		totalH := float64(grid.Rows) * cfg.LabelH
		for col := 0; col <= grid.Cols; col++ {
			x := cfg.MarginLeft + float64(col)*cfg.LabelW
			r.pdf.Line(x, cfg.MarginTop, x, cfg.MarginTop+totalH)
		}
		for row := 0; row <= grid.Rows; row++ {
			y := cfg.MarginTop + float64(row)*cfg.LabelH
			r.pdf.Line(cfg.MarginLeft, y, cfg.MarginLeft+totalW, y)
		}
	} else {
		for i := 0; i < grid.LabelsPerPage(); i++ {
			s := grid.Slot(i)
			r.pdf.Rect(s.X, s.Y, cfg.LabelW, cfg.LabelH, "D")
		}
	}
}

// renderLabel draws a single label at the given slot position.
func (r *Renderer) renderLabel(slot layout.Slot, lines []string) {
	cfg := r.cfg
	x, y := slot.X, slot.Y
	w, h := cfg.LabelW, cfg.LabelH
	pad := cfg.Padding

	if len(lines) == 0 {
		return
	}

	innerW := w - 2*pad
	innerH := h - 2*pad

	fontSize := cfg.FontSize
	if fontSize == 0 {
		fontSize = r.autoFitFontSize(lines, innerW)
	}
	r.pdf.SetFontSize(fontSize)

	lineH := fontSize * ptToMmFactor
	totalH := float64(len(lines)) * lineH

	var startY float64
	switch cfg.VAlign {
	case "top":
		startY = y + pad
	case "bottom":
		offset := innerH - totalH
		if offset < 0 {
			offset = 0
		}
		startY = y + pad + offset
	default: // middle
		offset := (innerH - totalH) / 2
		if offset < 0 {
			offset = 0
		}
		startY = y + pad + offset
	}

	alignStr := r.fpdfAlign()
	maxY := y + h - pad

	for _, line := range lines {
		if startY+lineH > maxY+lineH/2 { // stop before clipping outside the label
			break
		}
		r.pdf.SetXY(x+pad, startY)
		r.pdf.CellFormat(innerW, lineH, line, "", 0, alignStr, false, 0, "")
		startY += lineH
	}
}

// autoFitFontSize returns the largest font size (down to minFontSize) at which
// every line fits within maxWidth mm.
func (r *Renderer) autoFitFontSize(lines []string, maxWidth float64) float64 {
	for size := maxFontSize; size >= minFontSize; size -= 0.5 {
		r.pdf.SetFontSize(size)
		fits := true
		for _, line := range lines {
			if r.pdf.GetStringWidth(line) > maxWidth {
				fits = false
				break
			}
		}
		if fits {
			return size
		}
	}
	return minFontSize
}

func (r *Renderer) fpdfAlign() string {
	switch r.cfg.Align {
	case "center":
		return "C"
	case "right":
		return "R"
	default:
		return "L"
	}
}

// loadFont loads the embedded Go Regular font, or the user-supplied TTF when
// cfg.FontPath is set. fpdf stores errors internally; they surface on Output().
func (r *Renderer) loadFont() error {
	if r.cfg.FontPath != "" {
		r.pdf.AddUTF8Font(fontFamily, "", r.cfg.FontPath)
	} else {
		r.pdf.AddUTF8FontFromBytes(fontFamily, "", goregular.TTF)
	}
	return r.pdf.Error()
}

func parseCopies(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// expandEscapes converts literal \n sequences to newlines in template strings.
func expandEscapes(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && s[i+1] == 'n' {
			out = append(out, '\n')
			i++
		} else {
			out = append(out, s[i])
		}
	}
	return string(out)
}

func splitNewlines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func trimWhitespace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
