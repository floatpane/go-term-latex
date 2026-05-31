package termlatex

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"path/filepath"
)

// Options configure rendering and display.
type Options struct {
	// Backend selects the TeX pipeline. Auto (default) tries PDFLaTeX →
	// Tectonic → DVIPng in order.
	Backend Backend

	// Protocol selects the terminal graphics protocol. AutoProtocol (default)
	// detects from $TERM / $TERM_PROGRAM.
	Protocol Protocol

	// DPI is the render resolution for the PNG stage. Default 150.
	// Higher values produce sharper output in Kitty/Sixel at the cost of
	// a larger transfer; 96–200 is the useful range for most terminals.
	DPI int

	// MaxWidth / MaxHeight are pixel bounds for scaling before display.
	// 0 = fit to terminal.
	MaxWidth, MaxHeight int

	// Packages is a list of extra LaTeX packages added to the standalone
	// document preamble, e.g. []string{"physics", "siunitx"}.
	Packages []string

	// Foreground and Background override the glyph and background colors used
	// when recoloring the rendered equation. When either is nil, the missing
	// color is auto-detected from the terminal (OSC 10/11), falling back to a
	// dark theme. See NoTheme to disable recoloring entirely.
	Foreground color.Color
	Background color.Color

	// NoTheme disables terminal color detection and recoloring; the raw
	// black-on-white render is displayed as-is.
	NoTheme bool
}

// theme resolves the colors to recolor with, honoring explicit overrides and
// otherwise detecting from the terminal.
func (o Options) theme() Theme {
	if o.Foreground != nil && o.Background != nil {
		return Theme{Fg: o.Foreground, Bg: o.Background}
	}
	t := DetectTheme()
	if o.Foreground != nil {
		t.Fg = o.Foreground
	}
	if o.Background != nil {
		t.Bg = o.Background
	}
	return t
}

func (o Options) dpi() int {
	if o.DPI > 0 {
		return o.DPI
	}
	return 150
}

// Render writes equation to w as terminal graphics. equation is passed to the
// LaTeX document verbatim — callers control all markup.
func Render(w io.Writer, equation string, opts Options) error {
	pngBytes, err := renderPNG(equation, opts)
	if err != nil {
		return err
	}

	var img *image.NRGBA
	if opts.NoTheme {
		img, err = decodePNG(pngBytes)
	} else {
		img, err = recolor(pngBytes, opts.theme())
	}
	if err != nil {
		return err
	}

	return displayImage(w, img, opts)
}

// Display renders equation as block (display) math — equivalent to wrapping
// the equation in \[\displaystyle … \].
func Display(w io.Writer, equation string, opts Options) error {
	return Render(w, `\[\displaystyle `+equation+`\]`, opts)
}

// Inline renders equation as inline math — equivalent to wrapping in $…$.
func Inline(w io.Writer, equation string, opts Options) error {
	return Render(w, `$`+equation+`$`, opts)
}

// readFirstPNG reads the PNG that pdftoppm wrote at outBase. pdftoppm with
// -singlefile writes either outBase+".png" or outBase+"-1.png" depending on
// version; this function handles both.
func readFirstPNG(outBase string) ([]byte, error) {
	matches, err := filepath.Glob(outBase + "*.png")
	if err != nil {
		return nil, fmt.Errorf("%w: glob: %w", ErrRenderFailed, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: pdftoppm produced no PNG at %s", ErrRenderFailed, outBase)
	}
	// filepath.Glob returns sorted results; take the first (and normally only).
	data, err := readFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}
	return data, nil
}
