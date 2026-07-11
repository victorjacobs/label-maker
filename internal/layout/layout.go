package layout

import "math"

// Config holds the geometric parameters needed to build a grid.
type Config struct {
	PageW, PageH          float64
	LabelW, LabelH        float64
	MarginTop, MarginLeft float64
	GapX, GapY            float64
	Columns, Rows         int // 0 = auto-fit
}

// Grid describes the label grid on a single page.
type Grid struct {
	Cols, Rows            int
	labelW, labelH        float64
	marginTop, marginLeft float64
	gapX, gapY            float64
}

// Slot is a single label position with its page and top-left mm coordinates.
type Slot struct {
	Page, Col, Row int
	X, Y           float64
}

// NewGrid computes the column/row counts (auto-fitting when Columns or Rows is
// zero) and returns a Grid ready for slot iteration.
func NewGrid(cfg Config) Grid {
	cols := cfg.Columns
	rows := cfg.Rows

	if cols == 0 {
		cols = int(math.Floor((cfg.PageW - cfg.MarginLeft + cfg.GapX) / (cfg.LabelW + cfg.GapX)))
	}
	if rows == 0 {
		rows = int(math.Floor((cfg.PageH - cfg.MarginTop + cfg.GapY) / (cfg.LabelH + cfg.GapY)))
	}

	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}

	return Grid{
		Cols: cols, Rows: rows,
		labelW: cfg.LabelW, labelH: cfg.LabelH,
		marginTop: cfg.MarginTop, marginLeft: cfg.MarginLeft,
		gapX: cfg.GapX, gapY: cfg.GapY,
	}
}

// LabelsPerPage returns the number of label slots on a single page.
func (g Grid) LabelsPerPage() int { return g.Cols * g.Rows }

// Slot returns the position for the i-th label (0-indexed, row-major).
// The Page field indicates which page (0-indexed) the slot lands on.
func (g Grid) Slot(i int) Slot {
	perPage := g.Cols * g.Rows
	page := i / perPage
	pos := i % perPage
	row := pos / g.Cols
	col := pos % g.Cols

	x := g.marginLeft + float64(col)*(g.labelW+g.gapX)
	y := g.marginTop + float64(row)*(g.labelH+g.gapY)

	return Slot{Page: page, Col: col, Row: row, X: x, Y: y}
}
