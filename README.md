# label-maker

Turn a CSV of addresses into a print-ready PDF that lines up with standard label
sheets. Give it the physical label size in millimetres and it places labels on an
exact mm grid — no guessing, no printer-margin surprises.

```sh
label-maker \
  --input addresses.csv \
  --output labels.pdf \
  --label-width 63.5 --label-height 38.1 \
  --margin-top 15 --margin-left 7 --gap-x 2.5 \
  --template "{{.name}}\n{{.street}}\n{{.postcode}} {{.city}}"
```

## Installation

With Nix:
```sh
nix run github:victorjacobs/label-maker -- --help
```

From source:
```sh
git clone https://github.com/victorjacobs/label-maker
cd label-maker
nix build
./result/bin/label-maker --help
```

## Quick examples

**Minimal — auto-fit columns and rows, default template:**
```sh
label-maker \
  --input addresses.csv \
  --output labels.pdf \
  --label-width 63.5 --label-height 38.1
```

**Common Avery L7160 sheet (A4, 21 labels, 3×7):**
```sh
label-maker \
  --input addresses.csv \
  --output labels.pdf \
  --label-width 63.5 --label-height 38.1 \
  --margin-top 15 --margin-left 7 \
  --gap-x 2.5 --gap-y 0 \
  --template "{{.name}}\n{{.street}}\n{{.postcode}} {{.city}}"
```

**Reuse a partial sheet — skip the first 6 used slots:**
```sh
label-maker --input addresses.csv --output labels.pdf \
  --label-width 63.5 --label-height 38.1 \
  --margin-top 15 --margin-left 7 --gap-x 2.5 \
  --skip 6
```

**Print borders to check alignment before a real run:**
```sh
label-maker --input addresses.csv --output check.pdf \
  --label-width 63.5 --label-height 38.1 \
  --margin-top 15 --margin-left 7 --gap-x 2.5 \
  --draw-border
```

**Print 2 copies of each address:**
```sh
label-maker --input addresses.csv --output labels.pdf \
  --label-width 63.5 --label-height 38.1 --copies 2
```

**Per-row copy count from a CSV column:**
```sh
# addresses.csv has a "qty" column
label-maker --input addresses.csv --output labels.pdf \
  --label-width 63.5 --label-height 38.1 --copies-column qty
```

## CSV format

The first row is treated as a header by default. Column names are normalised to
lowercase with spaces replaced by underscores (`Post Code` → `post_code`), and
used as template variables:

```csv
name,street,postcode,city
Jane Doe,12 Baker St,1000,Brussels
Ømer Åberg,5 Rue Neuve,4000,Liège
```

With no `--template`, all non-empty columns are joined with newlines in column
order.

Use `--no-header` when there is no header row; columns are named `col0`, `col1`, etc.

## Template syntax

`--template` is a Go [`text/template`](https://pkg.go.dev/text/template) string.
Use `\n` for line breaks (the shell won't expand it, the tool handles it):

```
--template "{{.name}}\n{{.street}}\n{{.postcode}} {{.city}}"
```

## Common label dimensions

| Sheet | Labels | Size (mm) | Margins | Gap X |
|-------|--------|-----------|---------|-------|
| Avery L7160 (A4) | 21 (3×7) | 63.5 × 38.1 | top 15, left 7 | 2.5 |
| Avery L7163 (A4) | 14 (2×7) | 99.1 × 38.1 | top 15, left 7 | 2.5 |
| Avery 5160 (Letter) | 30 (3×10) | 66.675 × 25.4 | top 12.7, left 7.25 | 3.175 |

Measure your own sheets with a ruler if the above don't match — label
manufacturers are not always precise.

## All flags

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | `-` (stdin) | CSV file path |
| `--output` | `-` (stdout) | Output PDF path |
| `--label-width` | **required** | Label width in mm |
| `--label-height` | **required** | Label height in mm |
| `--page` | `a4` | Page size: `a4`, `letter`, or `WxH` in mm (e.g. `210x297`) |
| `--columns` | `0` | Columns per page (0 = auto-fit) |
| `--rows` | `0` | Rows per page (0 = auto-fit) |
| `--margin-top` | `0` | Distance from page top to first label row, mm |
| `--margin-left` | `0` | Distance from page left to first label column, mm |
| `--gap-x` | `0` | Horizontal gap between labels, mm |
| `--gap-y` | `0` | Vertical gap between labels, mm |
| `--padding` | `2` | Text inset inside each label, mm |
| `--template` | `""` | Label body template; empty = all columns joined |
| `--font` | `""` | Path to a TTF font file; empty = embedded Go Regular |
| `--font-size` | `0` | Font size in pt; `0` = auto-fit down to 5pt floor |
| `--align` | `left` | Horizontal alignment: `left`, `center`, `right` |
| `--valign` | `middle` | Vertical alignment: `top`, `middle`, `bottom` |
| `--copies-column` | `""` | CSV column name containing per-row copy count |
| `--copies` | `1` | Default copy count per row |
| `--skip` | `0` | Leave the first N label slots blank (row-major) |
| `--draw-border` | `false` | Draw a rectangle around each label for alignment checks |
| `--delimiter` | `,` | CSV field delimiter character |
| `--no-header` | `false` | Treat the first CSV row as data, not a header |

## Development

```sh
nix develop          # shell with Go, gopls, golangci-lint, gotools
go test ./...
go run . -- --help
```

See `CLAUDE.md` for architecture and contributor notes.
