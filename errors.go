package termlatex

import "errors"

var (
	// ErrNoBackend is returned when no supported TeX backend (pdflatex,
	// tectonic, latex+dvipng) is found in PATH.
	ErrNoBackend = errors.New("no LaTeX backend found in PATH")

	// ErrRenderFailed is returned when the TeX backend or image conversion tool
	// exits with an error. The wrapped error includes the tool's output.
	ErrRenderFailed = errors.New("LaTeX render failed")

	// ErrDisplay is returned when the rendered PNG cannot be sent to the
	// terminal (e.g. the writer is not a TTY or termimage returns an error).
	ErrDisplay = errors.New("terminal display failed")
)
