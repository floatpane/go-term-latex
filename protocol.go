package termlatex

import (
	"os"
	"strings"
)

// Protocol selects the terminal graphics protocol used to display an equation.
type Protocol int

const (
	// AutoProtocol detects the best protocol from $TERM / $TERM_PROGRAM.
	AutoProtocol Protocol = iota
	// HalfBlock renders with Unicode half-block characters (▀). Works on any
	// terminal with UTF-8 and 24-bit color.
	HalfBlock
	// Sixel uses the DEC Sixel protocol (xterm, foot, mlterm, WezTerm…).
	Sixel
	// Kitty uses the Kitty graphics protocol (kitty, Ghostty, WezTerm…).
	Kitty
)

func (p Protocol) String() string {
	switch p {
	case Kitty:
		return "kitty"
	case Sixel:
		return "sixel"
	case HalfBlock:
		return "halfblock"
	default:
		return "auto"
	}
}

// bestProtocol returns the best protocol the current terminal supports based on
// environment variables. It sends no ANSI queries, so it needs no TTY.
func bestProtocol() Protocol {
	// KITTY_WINDOW_ID is set by kitty itself.
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return Kitty
	}

	term := os.Getenv("TERM")
	termProg := strings.ToLower(os.Getenv("TERM_PROGRAM"))

	if term == "ghostty" || strings.HasPrefix(term, "xterm-ghostty") || termProg == "ghostty" {
		return Kitty
	}
	if termProg == "wezterm" {
		return Kitty
	}

	switch termProg {
	case "foot", "mlterm", "contour":
		return Sixel
	}
	if strings.Contains(term, "sixel") {
		return Sixel
	}

	return HalfBlock
}
