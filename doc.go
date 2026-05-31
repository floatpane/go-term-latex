// Package termlatex renders LaTeX math equations in the terminal using Kitty
// graphics, Sixel, or Unicode half-block characters as a fallback.
//
// Rendering is delegated to a TeX backend installed on the host — pdflatex,
// tectonic, or latex+dvipng, tried in that order. The backend produces a PNG,
// which is displayed through [github.com/floatpane/termimage].
//
// # Quick start
//
//	err := termlatex.Display(os.Stdout, `\int_0^\infty e^{-x^2}\,dx`, termlatex.Options{})
//	// ⇒ Rendered equation in the terminal, protocol auto-detected.
//
// # Equation helpers
//
// [Display] wraps the equation in \[\displaystyle…\] (block math).
// [Inline] wraps it in $…$ (inline math, tighter bounding box).
// [Render] passes the equation verbatim — callers control all markup.
//
// # Backend selection
//
// Backends are tried in preference order: PDFLaTeX → Tectonic → DVIPng.
// [Detect] returns the first available one. Set [Options.Backend] to pin a
// specific backend.
//
// # Theme matching
//
// The TeX backend renders black glyphs on white. Before display the PNG is
// recolored to match the terminal: [DetectTheme] queries the terminal for its
// foreground and background colors (OSC 10 / OSC 11), and glyphs are remapped
// to the foreground while the paper becomes the background. The result is fully
// opaque, so it blends correctly in every protocol — no white box on a dark
// terminal.
//
// Detection falls back to $COLORFGBG, then to light-on-dark, when the terminal
// does not answer. Override the colors with [Options.Foreground] and
// [Options.Background], or set [Options.NoTheme] to display the raw
// black-on-white render unchanged.
//
// # Shell escape
//
// pdflatex and latex are invoked with -interaction=nonstopmode and
// -halt-on-error but WITHOUT -shell-escape, so the document cannot execute
// arbitrary shell commands. Do not pass untrusted LaTeX to this library without
// additional sanitization.
package termlatex
