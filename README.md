# label-maker

A Go CLI that turns a CSV of addresses into a print-ready PDF of labels laid
out on an exact mm grid, ready to print on standard label sheets.

## Quick start

```sh
nix run github:victorjacobs/label-maker -- \
  --input addresses.csv \
  --output labels.pdf \
  --label-width 63.5 \
  --label-height 38.1
```

Or in a local checkout:

```sh
nix run . -- --input addresses.csv --output labels.pdf \
  --label-width 63.5 --label-height 38.1
```

## Development

```sh
nix develop       # drops you into a shell with Go, gopls, golangci-lint
go test ./...
go run . --help
```

## Build

```sh
nix build         # produces ./result/bin/label-maker
```

After adding new Go dependencies:

```sh
nix develop --command go mod tidy
nix develop --command go mod vendor
git add vendor
nix build
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | `-` (stdin) | CSV path |
| `--output` | `-` (stdout) | Output PDF path |
| `--label-width` | **required** | Label width in mm |
| `--label-height` | **required** | Label height in mm |
| `--page` | `a4` | `a4` \| `letter` \| `WxH` (mm) |
| `--columns` | `0` | Columns (0 = auto-fit) |
| `--rows` | `0` | Rows (0 = auto-fit) |
| `--margin-top` | `0` | Top margin in mm |
| `--margin-left` | `0` | Left margin in mm |
| `--gap-x` | `0` | Horizontal gap between labels in mm |
| `--gap-y` | `0` | Vertical gap between labels in mm |
| `--padding` | `2` | Text inset inside each label in mm |
| `--template` | `""` | Go template for label body (e.g. `{{.name}}\n{{.city}}`) |
| `--font` | `""` | Path to TTF font (empty = embedded Go Regular) |
| `--font-size` | `0` | Font size in pt (0 = auto-fit to label) |
| `--align` | `left` | `left` \| `center` \| `right` |
| `--valign` | `middle` | `top` \| `middle` \| `bottom` |
| `--copies-column` | `""` | CSV column with per-row copy count |
| `--copies` | `1` | Default copies per row |
| `--skip` | `0` | Skip N label slots (reuse partial sheets) |
| `--draw-border` | `false` | Draw label outlines for alignment checking |
| `--delimiter` | `,` | CSV field delimiter |
| `--no-header` | `false` | Treat first CSV row as data |

## CSV format

With a header row (default), template variables are the column names
lowercased with spaces replaced by underscores:

```csv
name,street,postcode,city
Jane Doe,12 Baker St,1000,Brussels
Ømer Åberg,5 Rue Neuve,4000,Liège
```

```sh
label-maker --input addresses.csv --output labels.pdf \
  --label-width 63.5 --label-height 38.1 \
  --margin-top 15 --margin-left 7 --gap-x 2.5 \
  --template "{{.name}}\n{{.street}}\n{{.postcode}} {{.city}}"
```

The default template (no `--template`) joins all non-empty columns
with newlines in column order.

## Page presets

| Preset | mm (W×H) |
|--------|----------|
| `a4` | 210 × 297 |
| `letter` | 215.9 × 279.4 |

Custom: `--page 210x297`
