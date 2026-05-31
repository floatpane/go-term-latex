<div align="center">

# go-term-latex

**Render LaTeX math equations in the terminal — Kitty, Sixel, or Unicode half-block fallback.**

[![Go Version](https://img.shields.io/github/go-mod/go-version/floatpane/go-term-latex)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/floatpane/go-term-latex.svg)](https://pkg.go.dev/github.com/floatpane/go-term-latex)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/floatpane/go-term-latex)](https://github.com/floatpane/go-term-latex/releases)
[![CI](https://github.com/floatpane/go-term-latex/actions/workflows/ci.yml/badge.svg)](https://github.com/floatpane/go-term-latex/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

`go-term-latex` turns a LaTeX math equation into terminal graphics in one call.
It shells out to a TeX backend installed on the host — `pdflatex`, `tectonic`,
or `latex`+`dvipng` — to produce a tight-cropped PNG, recolors it to match the
terminal theme, and writes it directly using the best protocol the terminal
supports (Kitty graphics, Sixel, or Unicode half-block). No external display
dependency — protocol detection and encoding are built in.

## Install

```bash
go get github.com/floatpane/go-term-latex
```

Requires Go 1.26+ and one of the following TeX backends:

| Backend | Packages needed |
|---------|----------------|
| **pdflatex** (recommended) | `texlive-latex-base` + `poppler-utils` (Linux) / MacTeX (macOS) |
| **tectonic** | `tectonic` (downloads TeX packages on first run) |
| **latex + dvipng** | `texlive-latex-base` + `dvipng` |

## Usage

```go
package main

import (
    "os"

    termlatex "github.com/floatpane/go-term-latex"
)

func main() {
    // Block (display) math — \[\displaystyle …\]
    _ = termlatex.Display(os.Stdout, `\int_0^\infty e^{-x^2}\,dx = \frac{\sqrt{\pi}}{2}`, termlatex.Options{})

    // Inline math — $…$
    _ = termlatex.Inline(os.Stdout, `E = mc^2`, termlatex.Options{})

    // Verbatim — you control all markup
    _ = termlatex.Render(os.Stdout, `\[\begin{pmatrix}a&b\\c&d\end{pmatrix}\]`, termlatex.Options{})
}
```

### Options

```go
termlatex.Options{
    Backend:    termlatex.PDFLaTeX, // pin a backend; default Auto-detects
    Protocol:   termlatex.Kitty,    // pin a protocol; default Auto-detects
    DPI:        200,                // render resolution (default 150)
    MaxWidth:   800,                // pixel cap before scaling (0 = fit terminal)
    MaxHeight:  400,
    Packages:   []string{"physics", "siunitx"}, // extra \usepackage entries
    Foreground: color.NRGBA{0, 200, 255, 255},  // glyph color; nil = auto-detect
    Background: color.NRGBA{20, 20, 20, 255},    // bg color; nil = auto-detect
    NoTheme:    false,              // true = raw black-on-white, no recolor
}
```

### Theme matching

The backend renders black glyphs on white. Before display the PNG is recolored
to match your terminal — glyphs take the terminal foreground, the paper takes
the background — so there's no white box on a dark terminal. Colors are detected
by querying the terminal (OSC 10 / OSC 11), falling back to `$COLORFGBG` and
then light-on-dark. Override with `Foreground`/`Background`, or set
`NoTheme: true` to keep the raw black-on-white render.

### Backend detection

```go
b, err := termlatex.Detect()
if err != nil {
    log.Fatal("no TeX backend found:", err)
}
fmt.Println("using", b) // "pdflatex", "tectonic", or "latex+dvipng"
```

## How it works

1. The equation is wrapped in a minimal `standalone` class document with
   `amsmath`, `amssymb`, and `amsfonts`.
2. The document is written to a temp dir and compiled by the chosen TeX backend
   (no `--shell-escape`, so the document cannot run arbitrary shell commands).
3. The PDF or DVI is converted to a tight-cropped PNG via `pdftoppm` or
   `dvipng`.
4. The black-on-white PNG is recolored to match the terminal theme (detected
   via OSC 10 / OSC 11), so glyphs use the terminal foreground and the
   background blends in.
5. The PNG is scaled to fit the terminal and written using the best available
   protocol (Kitty → Sixel → HalfBlock), detected from `$TERM` /
   `$TERM_PROGRAM`. Encoding is built in — no external display dependency.

## Documentation

Full API reference: [pkg.go.dev/github.com/floatpane/go-term-latex](https://pkg.go.dev/github.com/floatpane/go-term-latex)

## Contributing

PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

Do not pass untrusted LaTeX to this library without additional sanitization —
see [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
