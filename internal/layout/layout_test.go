package layout

import (
	"testing"
)

// A4 with standard Avery-style label dimensions.
func a4Cfg() Config {
	return Config{
		PageW: 210, PageH: 297,
		LabelW: 63.5, LabelH: 38.1,
		MarginTop: 15, MarginLeft: 7,
		GapX: 2.5, GapY: 0,
	}
}

func TestAutoFit_A4(t *testing.T) {
	g := NewGrid(a4Cfg())
	// cols = floor((210 - 7 + 2.5) / (63.5 + 2.5)) = floor(205.5/66) = 3
	if g.Cols != 3 {
		t.Errorf("Cols = %d, want 3", g.Cols)
	}
	// rows = floor((297 - 15 + 0) / (38.1 + 0)) = floor(282/38.1) = 7
	if g.Rows != 7 {
		t.Errorf("Rows = %d, want 7", g.Rows)
	}
}

func TestAutoFit_Letter(t *testing.T) {
	cfg := Config{
		PageW: 215.9, PageH: 279.4,
		LabelW: 66.675, LabelH: 25.4,
		MarginTop: 12.7, MarginLeft: 7.25,
		GapX: 3.175, GapY: 0,
	}
	g := NewGrid(cfg)
	if g.Cols < 1 || g.Rows < 1 {
		t.Errorf("got degenerate grid: %dx%d", g.Cols, g.Rows)
	}
}

func TestAutoFit_NoGap(t *testing.T) {
	cfg := Config{
		PageW: 100, PageH: 100,
		LabelW: 25, LabelH: 25,
		MarginTop: 0, MarginLeft: 0,
	}
	g := NewGrid(cfg)
	// cols = floor((100 - 0 + 0) / (25 + 0)) = 4
	if g.Cols != 4 {
		t.Errorf("Cols = %d, want 4", g.Cols)
	}
	if g.Rows != 4 {
		t.Errorf("Rows = %d, want 4", g.Rows)
	}
}

func TestExplicitColsRows(t *testing.T) {
	cfg := a4Cfg()
	cfg.Columns = 2
	cfg.Rows = 5
	g := NewGrid(cfg)
	if g.Cols != 2 {
		t.Errorf("Cols = %d, want 2", g.Cols)
	}
	if g.Rows != 5 {
		t.Errorf("Rows = %d, want 5", g.Rows)
	}
}

func TestSlotCoordinates(t *testing.T) {
	g := NewGrid(a4Cfg())
	// Slot 0: page=0, col=0, row=0
	s0 := g.Slot(0)
	if s0.Page != 0 || s0.Col != 0 || s0.Row != 0 {
		t.Errorf("slot 0: page=%d col=%d row=%d", s0.Page, s0.Col, s0.Row)
	}
	if s0.X != 7 || s0.Y != 15 {
		t.Errorf("slot 0 coords: X=%.2f Y=%.2f, want X=7 Y=15", s0.X, s0.Y)
	}

	// Slot 1: col=1, row=0
	s1 := g.Slot(1)
	wantX := 7.0 + 1*(63.5+2.5) // 7 + 66 = 73
	if s1.X != wantX {
		t.Errorf("slot 1 X = %.2f, want %.2f", s1.X, wantX)
	}

	// Slot 3: col=0, row=1
	s3 := g.Slot(3)
	if s3.Col != 0 || s3.Row != 1 {
		t.Errorf("slot 3: col=%d row=%d", s3.Col, s3.Row)
	}
	if s3.Y != 15+38.1 {
		t.Errorf("slot 3 Y = %.2f, want %.2f", s3.Y, 15+38.1)
	}
}

func TestPageBreak(t *testing.T) {
	g := NewGrid(a4Cfg()) // 3 cols × 7 rows = 21 per page
	perPage := g.LabelsPerPage()
	if perPage != 21 {
		t.Fatalf("labelsPerPage = %d, want 21", perPage)
	}

	// Slot 20 is the last on page 0
	s20 := g.Slot(20)
	if s20.Page != 0 {
		t.Errorf("slot 20 should be on page 0, got page %d", s20.Page)
	}

	// Slot 21 starts page 1
	s21 := g.Slot(21)
	if s21.Page != 1 {
		t.Errorf("slot 21 should be on page 1, got page %d", s21.Page)
	}
	// And its coordinates should be the same as slot 0
	s0 := g.Slot(0)
	if s21.X != s0.X || s21.Y != s0.Y {
		t.Errorf("slot 21 coords (%.2f,%.2f) should equal slot 0 (%.2f,%.2f)", s21.X, s21.Y, s0.X, s0.Y)
	}
}

func TestSkipOffset(t *testing.T) {
	// Skip 5 slots: slots 0-4 are empty, first label goes to slot 5
	// Verify that slot 5 is at (col=2, row=1) for a 3-col grid
	g := NewGrid(a4Cfg()) // 3 cols
	s5 := g.Slot(5)
	if s5.Col != 2 || s5.Row != 1 {
		t.Errorf("slot 5: col=%d row=%d, want col=2 row=1", s5.Col, s5.Row)
	}
}

func TestMinimumGrid(t *testing.T) {
	// Margins that leave barely one label
	cfg := Config{
		PageW: 70, PageH: 45,
		LabelW: 63.5, LabelH: 38.1,
		MarginTop: 1, MarginLeft: 1,
	}
	g := NewGrid(cfg)
	if g.Cols < 1 || g.Rows < 1 {
		t.Errorf("expected at least 1x1 grid, got %dx%d", g.Cols, g.Rows)
	}
}
