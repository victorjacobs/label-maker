package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/victorjacobs/label-maker/internal/config"
	"github.com/victorjacobs/label-maker/internal/csvsource"
	"github.com/victorjacobs/label-maker/internal/render"
)

var flags struct {
	input         string
	output        string
	labelWidth    float64
	labelHeight   float64
	page          string
	columns       int
	rows          int
	marginTop     float64
	marginLeft    float64
	gapX          float64
	gapY          float64
	padding       float64
	tmpl          string
	font          string
	fontSize      float64
	align         string
	valign        string
	copiesColumn  string
	copies        int
	skip          int
	drawBorder    bool
	delimiter     string
	noHeader      bool
}

var rootCmd = &cobra.Command{
	Use:   "label-maker",
	Short: "Turn a CSV of addresses into a print-ready PDF of labels",
	Long: `label-maker reads a CSV of addresses and renders one label per row
onto a PDF that lines up with standard label sheets.`,
	RunE: run,
}

func init() {
	f := rootCmd.Flags()

	f.StringVar(&flags.input, "input", "-", "CSV path; - for stdin")
	f.StringVar(&flags.output, "output", "-", "output PDF path; - for stdout")
	f.Float64Var(&flags.labelWidth, "label-width", 0, "label width in mm (required)")
	f.Float64Var(&flags.labelHeight, "label-height", 0, "label height in mm (required)")
	f.StringVar(&flags.page, "page", "a4", "page size: a4 | letter | WxH (mm)")
	f.IntVar(&flags.columns, "columns", 0, "number of columns (0 = auto-fit)")
	f.IntVar(&flags.rows, "rows", 0, "number of rows (0 = auto-fit)")
	f.Float64Var(&flags.marginTop, "margin-top", 0, "top margin in mm")
	f.Float64Var(&flags.marginLeft, "margin-left", 0, "left margin in mm")
	f.Float64Var(&flags.gapX, "gap-x", 0, "horizontal gap between labels in mm")
	f.Float64Var(&flags.gapY, "gap-y", 0, "vertical gap between labels in mm")
	f.Float64Var(&flags.padding, "padding", 2, "text inset inside each label in mm")
	f.StringVar(&flags.tmpl, "template", "", `label template (e.g. "{{.name}}\n{{.city}}"); empty = all columns`)
	f.StringVar(&flags.font, "font", "", "path to TTF font file; empty = embedded Go Regular")
	f.Float64Var(&flags.fontSize, "font-size", 0, "font size in pt (0 = auto-fit to label)")
	f.StringVar(&flags.align, "align", "left", "horizontal text alignment: left | center | right")
	f.StringVar(&flags.valign, "valign", "middle", "vertical text alignment: top | middle | bottom")
	f.StringVar(&flags.copiesColumn, "copies-column", "", "CSV column with per-row copy count")
	f.IntVar(&flags.copies, "copies", 1, "default copies per label row")
	f.IntVar(&flags.skip, "skip", 0, "skip N label slots (reuse partial sheets)")
	f.BoolVar(&flags.drawBorder, "draw-border", false, "draw label outlines (alignment/debug)")
	f.StringVar(&flags.delimiter, "delimiter", ",", "CSV field delimiter")
	f.BoolVar(&flags.noHeader, "no-header", false, "treat first CSV row as data, not a header")

	_ = rootCmd.MarkFlagRequired("label-width")
	_ = rootCmd.MarkFlagRequired("label-height")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	pageW, pageH, err := config.ParsePage(flags.page)
	if err != nil {
		return err
	}

	delimiter := ','
	if flags.delimiter != "" {
		runes := []rune(flags.delimiter)
		if len(runes) != 1 {
			return fmt.Errorf("--delimiter must be a single character (got %q)", flags.delimiter)
		}
		delimiter = runes[0]
	}

	cfg := &config.Config{
		InputPath:    flags.input,
		OutputPath:   flags.output,
		LabelW:       flags.labelWidth,
		LabelH:       flags.labelHeight,
		PageW:        pageW,
		PageH:        pageH,
		Columns:      flags.columns,
		Rows:         flags.rows,
		MarginTop:    flags.marginTop,
		MarginLeft:   flags.marginLeft,
		GapX:         flags.gapX,
		GapY:         flags.gapY,
		Padding:      flags.padding,
		Template:     flags.tmpl,
		FontPath:     flags.font,
		FontSize:     flags.fontSize,
		Align:        flags.align,
		VAlign:       flags.valign,
		CopiesColumn: flags.copiesColumn,
		Copies:       flags.copies,
		Skip:         flags.skip,
		DrawBorder:   flags.drawBorder,
		Delimiter:    delimiter,
		NoHeader:     flags.noHeader,
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	// Open input.
	var csvReader io.Reader
	if cfg.InputPath == "-" {
		csvReader = cmd.InOrStdin()
	} else {
		f, err := os.Open(cfg.InputPath)
		if err != nil {
			return fmt.Errorf("opening input: %w", err)
		}
		defer f.Close()
		csvReader = f
	}

	// Parse CSV.
	src, err := csvsource.Parse(csvReader, cfg.Delimiter, cfg.NoHeader)
	if err != nil {
		return fmt.Errorf("parsing CSV: %w", err)
	}

	// Build labels.
	labels, err := render.BuildLabels(src.Records, src.Headers, cfg)
	if err != nil {
		return err
	}

	// Open output.
	var out io.Writer
	if cfg.OutputPath == "-" {
		out = cmd.OutOrStdout()
	} else {
		f, err := os.Create(cfg.OutputPath)
		if err != nil {
			return fmt.Errorf("creating output: %w", err)
		}
		defer f.Close()
		out = f
	}

	// Render PDF.
	r := render.New(cfg)
	if err := r.Render(labels, out); err != nil {
		return fmt.Errorf("rendering PDF: %w", err)
	}

	return nil
}
