package termlatex

import (
	"image/color"
	"os"
	"strconv"
	"strings"
)

// Theme holds the foreground (glyph) and background colors used to recolor a
// rendered equation so it blends with the terminal.
type Theme struct {
	Fg color.Color
	Bg color.Color
}

// default themes used when the terminal does not answer an OSC color query.
var (
	darkTheme  = Theme{Fg: color.White, Bg: color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}}
	lightTheme = Theme{Fg: color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}, Bg: color.White}
)

// DetectTheme queries the controlling terminal for its foreground and
// background colors via OSC 10 / OSC 11. If the terminal does not respond
// (not a TTY, query unsupported, or timeout) it falls back to a dark theme,
// or to $COLORFGBG when that env var is set.
func DetectTheme() Theme {
	t, ok := queryTheme()
	if ok {
		return t
	}
	return fallbackTheme()
}

// fallbackTheme picks a theme without querying the terminal: $COLORFGBG if
// present (e.g. "15;0" = light fg on dark bg), otherwise dark.
func fallbackTheme() Theme {
	if v := os.Getenv("COLORFGBG"); v != "" {
		parts := strings.Split(v, ";")
		if len(parts) >= 2 {
			if bg, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				// ANSI colors 0-7 are dark, 8-15 light. A light bg index
				// means a light theme.
				if bg >= 7 {
					return lightTheme
				}
				return darkTheme
			}
		}
	}
	return darkTheme
}

// parseOSCColor extracts an RGB color from an OSC reply such as
// "\x1b]11;rgb:1c1c/1c1c/1c1c\x07". Each component may be 1-4 hex digits;
// the value is scaled to 8 bits.
func parseOSCColor(s string) (color.Color, bool) {
	i := strings.Index(s, "rgb:")
	if i < 0 {
		return nil, false
	}
	spec := s[i+len("rgb:"):]
	// Trim trailing terminator (BEL or ST).
	if j := strings.IndexAny(spec, "\x07\x1b"); j >= 0 {
		spec = spec[:j]
	}
	parts := strings.Split(spec, "/")
	if len(parts) != 3 {
		return nil, false
	}
	var rgb [3]uint8
	for k, p := range parts {
		v, err := strconv.ParseUint(p, 16, 32)
		if err != nil || len(p) == 0 || len(p) > 4 {
			return nil, false
		}
		// Scale the parsed value to 8 bits based on its hex-digit width.
		shift := uint(4 * len(p))
		rgb[k] = uint8((v * 0xff) / ((1 << shift) - 1))
	}
	return color.NRGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 0xff}, true
}

// isDark reports whether c has low perceived luminance.
func isDark(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	// RGBA returns 16-bit values; scale to 8-bit.
	y := (299*(r>>8) + 587*(g>>8) + 114*(b>>8)) / 1000
	return y < 128
}
