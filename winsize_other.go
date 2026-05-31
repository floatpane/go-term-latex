//go:build !(linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package termlatex

// Fallback terminal metrics for platforms without TIOCGWINSZ.
func detectTermChars() (cols, rows int)    { return 220, 50 }
func detectCellPixels() (cellW, cellH int) { return 8, 16 }
