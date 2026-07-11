# label-maker

A Go CLI that turns a CSV of addresses into a print-ready PDF of labels laid
out on a sheet. You give it a CSV, the physical label size in millimetres, and
it emits a PDF whose labels line up with the label paper in your printer.

> Status: **greenfield**. This file is the spec. Implement it from scratch.
> Existing label tools are clunky; the goal here is a small, correct,
> scriptable tool with sane defaults and precise control when needed.

---

## Goals & non-goals

**Goals**
- Read a CSV of addresses and render one label per row (with optional copies).
- Accept the physical label size in mm and the page size, and place labels on
  an exact mm grid so they align with real label sheets.
- Produce a deterministic, print-ready PDF (no printer margins surprises).
- Unicode-correct text (accents, etc.) — this is addresses.
- Be fully scriptable: every layout parameter is a flag; no interactive prompts.
- Reproducible toolchain and build via Nix flake.

**Non-goals (for v1)**
- No barcodes/QR, no images/logos on labels.
- No GUI, no live preview.
- No address validation or geocoding.
- No per-label styling from the CSV (one style for the whole run).

---

## Language, libraries, toolchain

- **Go** (latest stable; pin in `go.mod` and via Nix).
- **Module path**: `github.com/victorjacobs/label-maker` (use for `go.mod` and
  all internal import paths, e.g. `github.com/victorjacobs/label-maker/internal/layout`).
- **PDF**: `github.com/go-pdf/fpdf` (maintained fork of gofpdf). Use unit `mm`.
- **CLI**: `github.com/spf13/cobra`. Single root command that does the work;
  structured so subcommands can be added later.
- **Templating**: stdlib `text/template` for the label body.
- **CSV**: stdlib `encoding/csv`.
- **Fonts**: bundle a Unicode TTF (DejaVu Sans + DejaVu Sans Bold, or Go's
  `golang.org/x/image/font` Go fonts) embedded via `go:embed` so the binary is
  self-contained. Allow override with `--font`.

Keep the dependency tree small. Prefer stdlib where reasonable.

---

## Units & coordinate model

- All external measurements are **millimetres**. Internally use mm everywhere
  and let fpdf handle mm→pt.
- Origin is top-left of the page. `x` grows right, `y` grows down.
- A label at grid position `(col, row)` (both 0-indexed) has its top-left at:
  ```
  x = marginLeft + col * (labelWidth  + gapX)
  y = marginTop  + row * (labelHeight + gapY)
  ```
- Text is drawn inside the label with `padding` inset on all sides.

---

## CLI

```
label-maker \
  --input addresses.csv \       # CSV path; "-" for stdin
  --output labels.pdf \         # output path; "-" for stdout
  --label-width 63.5 \          # mm, required
  --label-height 38.1 \         # mm, required
  --page a4 \                   # a4 | letter | WxH (e.g. 210x297), default a4
  --columns 0 \                 # 0 = auto-fit from geometry
  --rows 0 \                    # 0 = auto-fit from geometry
  --margin-top 15 --margin-left 7 \   # mm to first label's top-left
  --gap-x 2.5 --gap-y 0 \       # mm between labels
  --padding 2 \                 # mm text inset inside each label
  --template "{{.name}}\n{{.street}}\n{{.postcode}} {{.city}}" \
  --font "" \                   # path to TTF; empty = embedded DejaVu Sans
  --font-size 0 \               # pt; 0 = auto-fit to label
  --align left \                # left | center | right
  --valign middle \             # top | middle | bottom
  --copies-column "" \          # CSV column giving per-row copy count
  --copies 1 \                  # default copies per row
  --skip 0 \                    # skip N label slots (reuse partial sheets)
  --draw-border \               # draw label outlines (alignment/debug)
  --delimiter "," \             # CSV delimiter
  --no-header                   # treat first row as data, not header
```

Flags map 1:1 to a `Config` struct. Validate early with clear errors
(e.g. label bigger than printable area, unknown page preset).

### Behaviours
- `--columns/--rows = 0` → auto-fit the maximum whole labels that fit given
  page size, margins, label size, and gaps.
- `--skip N` leaves the first N slots blank (top-left to bottom-right,
  row-major), so a partially used sheet can be reused.
- When labels overflow the current page, start a new page.
- `--font-size 0` → auto-shrink: pick the largest size (down to a floor, e.g.
  5pt) at which every line of the label fits within `width - 2*padding`; wrap
  long lines only if still overflowing.
- `--draw-border` draws a thin rectangle at each label's bounds for aligning
  the printer / checking the grid. Off by default (real label sheets need no
  border).

---

## CSV → label text

- With a header row (default), columns are addressable by name in the template
  (lower-cased, spaces→underscores for the key, e.g. `Post Code` → `.post_code`).
- Default template if `--template` is empty: join all non-empty columns of the
  row with newlines, in column order.
- `\n` in the template string means a line break in the label.
- Trim whitespace on each rendered line; drop empty lines.
- Copies: if `--copies-column` is set and the cell parses as a positive int,
  render that many copies of the row; otherwise fall back to `--copies`.

Example CSV (`testdata/addresses.csv`):
```csv
name,street,postcode,city
Jane Doe,12 Baker St,1000,Brussels
Ømer Åberg,5 Rue Neuve,4000,Liège
```

---

## Page presets

| preset | mm (W×H)      |
|--------|---------------|
| a4     | 210 × 297     |
| letter | 215.9 × 279.4 |

Also accept `WxH` (mm, floats), e.g. `--page 210x297`. Add more presets easily
via a lookup map. (Common Avery-style label presets are a nice-to-have follow-up
but out of scope for v1 — the mm-driven grid already covers them.)

---

## Project layout

```
.
├── CLAUDE.md
├── README.md              # user-facing usage
├── flake.nix
├── flake.lock
├── go.mod
├── go.sum
├── main.go                # thin: calls cmd.Execute()
├── internal/
│   ├── cli/               # cobra command, flag→Config, validation
│   ├── config/            # Config struct + defaults + validation
│   ├── csvsource/         # CSV parse → []Record (map[string]string)
│   ├── layout/            # page/label geometry, auto-fit, slot iteration
│   ├── render/            # fpdf rendering: fonts, text fit, page breaks
│   └── assets/            # embedded fonts (go:embed)
└── testdata/
    └── addresses.csv
```

Keep `internal/layout` pure (no fpdf, no IO) so the geometry math is unit
testable. `internal/render` depends on fpdf and consumes layout output.

---

## Core types (sketch)

```go
type Config struct {
    InputPath, OutputPath          string
    LabelW, LabelH                 float64 // mm
    PageW, PageH                   float64 // mm (resolved from preset)
    Columns, Rows                  int     // 0 = auto
    MarginTop, MarginLeft          float64
    GapX, GapY                     float64
    Padding                        float64
    Template                       string
    FontPath                       string  // "" = embedded
    FontSize                       float64 // 0 = auto
    Align, VAlign                  string
    CopiesColumn                   string
    Copies                         int
    Skip                           int
    DrawBorder                     bool
    Delimiter                      rune
    NoHeader                       bool
}

// layout
type Grid struct {
    Cols, Rows int
    // slot(i) -> (x, y) top-left in mm; handles page index via i / (Cols*Rows)
}

type Slot struct { Page, Col, Row int; X, Y float64 }
```

---

## Rendering rules

1. Set PDF unit mm, page size from config, margins 0 (we place absolutely).
2. Load font (embedded or `--font`) with `AddUTF8Font`; enable UTF-8 text.
3. Build the ordered list of labels: for each CSV record, render template →
   lines, expand by copies. Prepend `Skip` empty slots.
4. Iterate slots row-major; on slot index crossing `Cols*Rows`, `AddPage()`.
5. For each label: compute inner box `(x+pad, y+pad, w-2pad, h-2pad)`,
   determine font size (auto or fixed), draw each line honoring `--align`
   and `--valign`. Clip/wrap so text never escapes the label box.
6. If `--draw-border`, stroke the label rectangle.
7. Output to file or stdout (`-`).

---

## Nix

Provide a `flake.nix` with:
- `devShells.default`: `go`, `gopls`, `golangci-lint`, `gotools`.
- `packages.default`: `pkgs.buildGoModule` building the CLI (set `vendorHash`;
  document `nix build` and how to update the hash).
- `apps.default`: runnable via `nix run . -- --help`.
- Use `flake-utils` (or `forEachSystem`) for multi-system.

Dev flow: `nix develop` → `go run . --help`. CI/build: `nix build`.

---

## Testing

- `internal/layout`: table tests for auto-fit counts and slot coordinates
  (a4, letter, custom sizes, with/without gaps, skip offset, page breaks).
- `internal/csvsource`: header normalization, delimiter, no-header, copies col,
  UTF-8, quoted fields with embedded newlines.
- `internal/config`: validation errors (label > printable area, bad preset).
- `internal/render`: smoke test that a small CSV produces a non-empty valid PDF
  (check `%PDF` header and page count); optionally a golden-ish size assertion.
- Provide `testdata/addresses.csv` and use it in an end-to-end test.

---

## Acceptance criteria

- `label-maker --input testdata/addresses.csv --output out.pdf \
   --label-width 63.5 --label-height 38.1 --page a4` produces a valid PDF.
- Labels land on an exact mm grid; `--draw-border` outlines align to the
  configured geometry (verifiable by measuring the PDF).
- Auto-fit chooses the correct column/row counts for the geometry.
- Unicode addresses render correctly.
- `--skip`, `--copies`, `--copies-column`, and multi-page overflow all work.
- `nix build` and `nix run . -- --help` succeed; `nix develop` gives a working
  Go toolchain.

---

## Implementation order (suggested)

1. `go.mod`, flake, `main.go` + cobra skeleton printing parsed `Config`.
2. `config` (presets, validation) + `csvsource`.
3. `layout` geometry + tests.
4. `render` (fonts, text fit, page breaks, border).
5. End-to-end wiring, `testdata`, tests, `README.md`.
