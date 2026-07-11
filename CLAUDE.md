# label-maker — Claude instructions

A Go CLI that turns a CSV of addresses into a print-ready PDF of labels on an
exact mm grid. This file is the authoritative guide for working in this repo.

---

## Toolchain

All Go commands go through `nix develop`. Never invoke the system `go` directly.

```sh
nix develop --command go test ./...
nix develop --command go build ./...
nix develop --command go run . -- --help
```

`nix build` produces `./result/bin/label-maker`. `nix run . -- <flags>` runs it
without a local install. `.envrc` wires this up via direnv.

After changing dependencies:
```sh
nix develop --command go get <pkg>
nix develop --command go mod tidy
git add go.mod go.sum
```
Then update `vendorHash` in `flake.nix`: set it to `pkgs.lib.fakeHash`, run
`nix build`, copy the printed sha256, paste it in.

---

## Project layout

```
main.go                 thin entry point — calls cli.Execute()
internal/
  cli/root.go           cobra command, flag → Config, wires all packages
  config/config.go      Config struct, ParsePage, Validate
  csvsource/            CSV → []map[string]string with header normalisation
  layout/               pure geometry: Grid, Slot, auto-fit (no fpdf, no IO)
  render/               fpdf PDF renderer, font loading, template execution
testdata/addresses.csv  Unicode sample data used in tests
vendor/                 committed Go module vendor directory
flake.nix               devShell + buildGoModule + app
```

---

## Architecture rules

- **`internal/layout` is pure.** No fpdf, no file I/O. Geometry math only.
  This keeps it fast to unit-test without any PDF overhead.
- **`internal/render` owns the PDF.** It imports layout for slot coords and
  drives fpdf. Font loading, template execution, and copies expansion all live here.
- **Config flows one way:** CLI flags → `config.Config` → validated once →
  passed read-only to render and layout. Nothing mutates Config after validation.
- All measurements are **millimetres** end-to-end. fpdf is initialised with
  `UnitStr: "mm"`; never do manual pt↔mm conversions except the line-height
  constant (`ptToMmFactor = (25.4/72.0) * 1.3`).

---

## Key design decisions

- **Embedded font**: Go Regular from `golang.org/x/image/font/gofont/goregular`.
  Loaded via `fpdf.AddUTF8FontFromBytes` — no temp file needed.
- **Font size auto-fit**: iterates 12pt → 5pt in 0.5pt steps; picks the largest
  size where every line fits within `labelWidth - 2*padding`. Floor is 5pt.
- **Slot indexing**: `grid.Slot(i)` is the single source of truth for position.
  `i` includes the `--skip` offset — the caller adds `cfg.Skip` before passing
  `i` to `grid.Slot`.
- **`\n` in `--template`**: the shell doesn't expand `\n` in flag values, so
  `render.expandEscapes` converts literal `\n` to newlines before template parse.
- **fpdf error model**: `AddUTF8Font*` returns void; errors accumulate in
  `pdf.Error()`. Check it after font loading; `pdf.Output()` also surfaces them.

---

## Coordinate model

```
x = marginLeft + col * (labelWidth  + gapX)
y = marginTop  + row * (labelHeight + gapY)
```

Auto-fit:
```
cols = floor((pageWidth  - marginLeft + gapX) / (labelWidth  + gapX))
rows = floor((pageHeight - marginTop  + gapY) / (labelHeight + gapY))
```

---

## Testing

Run all tests: `nix develop --command go test ./...`

- `internal/config`: validation errors, `ParsePage` presets and custom WxH.
- `internal/csvsource`: header normalisation, delimiter, no-header, UTF-8,
  quoted fields with embedded newlines.
- `internal/layout`: auto-fit counts (A4 + standard label = 3×7), slot coords,
  page-break boundaries, skip offset.
- `internal/render`: smoke test (PDF starts with `%PDF-`), `BuildLabels` with
  default template / custom template / copies / copies-column.

No golden-file PDF tests — the smoke test checking `%PDF-` is sufficient.

---

## Adding a page preset

Edit the `pagePresets` map in `internal/config/config.go`. No other changes needed.

---

## Non-goals (v1)

No barcodes, no images/logos, no GUI, no address validation, no per-label
styling from CSV data.
