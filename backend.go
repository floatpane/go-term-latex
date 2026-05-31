package termlatex

import (
	"context"
	"fmt"
	"os/exec"
)

// Backend selects the tool used to convert a LaTeX document to a PNG image.
type Backend int

const (
	// Auto tries PDFLaTeX, then Tectonic, then DVIPng in order.
	Auto Backend = iota
	// PDFLaTeX uses pdflatex (from any TeX distribution) and pdftoppm
	// (from poppler-utils) to convert PDF→PNG.
	PDFLaTeX
	// Tectonic uses the self-contained tectonic engine (downloads TeX packages
	// on demand) and pdftoppm for PDF→PNG.
	Tectonic
	// DVIPng uses latex (DVI mode) and dvipng for direct DVI→PNG conversion.
	// Produces the sharpest output for math at low DPI.
	DVIPng
)

func (b Backend) String() string {
	switch b {
	case Auto:
		return autoName
	case PDFLaTeX:
		return "pdflatex"
	case Tectonic:
		return "tectonic"
	case DVIPng:
		return "latex+dvipng"
	}
	return autoName
}

// Detect returns the first available backend in preference order:
// PDFLaTeX → Tectonic → DVIPng. Returns ErrNoBackend if none found.
func Detect() (Backend, error) {
	for _, candidate := range []struct {
		b    Backend
		bins []string
	}{
		{PDFLaTeX, []string{"pdflatex", "pdftoppm"}},
		{Tectonic, []string{"tectonic", "pdftoppm"}},
		{DVIPng, []string{"latex", "dvipng"}},
	} {
		if allInPath(candidate.bins...) {
			return candidate.b, nil
		}
	}
	return Auto, ErrNoBackend
}

func allInPath(bins ...string) bool {
	for _, b := range bins {
		if _, err := exec.LookPath(b); err != nil {
			return false
		}
	}
	return true
}

// renderPNG dispatches to the appropriate backend renderer.
func renderPNG(ctx context.Context, equation string, opts Options) ([]byte, error) {
	b := opts.Backend
	if b == Auto {
		var err error
		b, err = Detect()
		if err != nil {
			return nil, fmt.Errorf("%w: install pdflatex+pdftoppm, tectonic+pdftoppm, or latex+dvipng", ErrNoBackend)
		}
	}
	switch b {
	case PDFLaTeX:
		return renderPDFLaTeX(ctx, equation, opts)
	case Tectonic:
		return renderTectonic(ctx, equation, opts)
	case DVIPng:
		return renderDVIPng(ctx, equation, opts)
	case Auto:
		// Unreachable: Auto is resolved above.
		return nil, fmt.Errorf("%w: unresolved auto backend", ErrNoBackend)
	}
	return nil, fmt.Errorf("%w: unknown backend %d", ErrNoBackend, b)
}
