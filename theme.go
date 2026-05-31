package termlatex

import (
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
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

// queryTheme opens /dev/tty, puts it in raw mode, and asks for the OSC 10
// (foreground) and OSC 11 (background) colors. Returns ok=false on any
// failure so callers can fall back.
func queryTheme() (Theme, bool) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return Theme{}, false
	}
	defer tty.Close() //nolint:errcheck

	fd := int(tty.Fd())
	old, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return Theme{}, false
	}
	raw := *old
	// Disable canonical mode and echo so the response isn't line-buffered or
	// printed, and use VMIN=0/VTIME to bound the read with a ~200ms timeout.
	raw.Lflag &^= unix.ICANON | unix.ECHO
	raw.Cc[unix.VMIN] = 0
	raw.Cc[unix.VTIME] = 2 // tenths of a second
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &raw); err != nil {
		return Theme{}, false
	}
	defer unix.IoctlSetTermios(fd, unix.TCSETS, old) //nolint:errcheck

	fg, fgOK := queryOSC(tty, fd, "10")
	bg, bgOK := queryOSC(tty, fd, "11")
	if !fgOK && !bgOK {
		return Theme{}, false
	}

	t := fallbackTheme()
	if fgOK {
		t.Fg = fg
	}
	if bgOK {
		t.Bg = bg
	}
	// If only one color was reported, derive the other from the background's
	// luminance so glyphs stay legible.
	if bgOK && !fgOK {
		if isDark(bg) {
			t.Fg = color.White
		} else {
			t.Fg = color.NRGBA{A: 0xff}
		}
	}
	return t, true
}

// queryOSC writes an OSC color query ("\x1b]<code>;?\x07") and parses the
// "rgb:RRRR/GGGG/BBBB" reply.
func queryOSC(tty *os.File, fd int, code string) (color.Color, bool) {
	if _, err := tty.WriteString("\x1b]" + code + ";?\x07"); err != nil {
		return nil, false
	}

	var buf []byte
	tmp := make([]byte, 64)
	deadline := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(deadline) {
		n, err := tty.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if strings.ContainsAny(string(buf), "\x07\\") {
				break
			}
		}
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
	}
	return parseOSCColor(string(buf))
}

// parseOSCColor extracts an RGB color from an OSC reply such as
// "\x1b]11;rgb:1c1c/1c1c/1c1c\x07". Each component may be 1-4 hex digits;
// the most-significant byte is used.
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
