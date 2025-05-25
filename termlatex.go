package termlatex

import (
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"

	termimage "github.com/floatpane/termimage"
)

// Protocol is re-exported from termimage for convenience.
type Protocol = termimage.Protocol

// Terminal graphics protocol constants. AutoProtocol detects from $TERM.
const (
	AutoProtocol Protocol = termimage.Auto
	HalfBlock             = termimage.HalfBlock
	Sixel                 = termimage.Sixel
	Kitty                 = termimage.Kitty
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
	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
	proto := opts.Protocol
	if proto == 0 {
		proto = AutoProtocol
	}
	if err := termimage.Display(w, uri, termimage.Options{
		Protocol:  proto,
		MaxWidth:  opts.MaxWidth,
		MaxHeight: opts.MaxHeight,
	}); err != nil {
		return fmt.Errorf("%w: %w", ErrDisplay, err)
	}
	return nil
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
