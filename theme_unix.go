//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package termlatex

import (
	"image/color"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// queryTheme opens /dev/tty, puts it in raw mode, and asks for the OSC 10
// (foreground) and OSC 11 (background) colors. Returns ok=false on any
// failure so callers can fall back. The ioctl request constants differ across
// platforms and are defined in termios_*.go.
func queryTheme() (Theme, bool) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return Theme{}, false
	}
	defer tty.Close() //nolint:errcheck

	fd := int(tty.Fd())
	old, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil {
		return Theme{}, false
	}
	raw := *old
	// Disable canonical mode and echo so the response isn't line-buffered or
	// printed, and use VMIN=0/VTIME to bound the read with a ~200ms timeout.
	raw.Lflag &^= unix.ICANON | unix.ECHO
	raw.Cc[unix.VMIN] = 0
	raw.Cc[unix.VTIME] = 2 // tenths of a second
	if err := unix.IoctlSetTermios(fd, ioctlSetTermios, &raw); err != nil {
		return Theme{}, false
	}
	defer unix.IoctlSetTermios(fd, ioctlSetTermios, old) //nolint:errcheck

	fg, fgOK := queryOSC(tty, "10")
	bg, bgOK := queryOSC(tty, "11")
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
	// If only the background was reported, derive a legible glyph color from
	// its luminance.
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
func queryOSC(tty *os.File, code string) (color.Color, bool) {
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
		if err != nil || n == 0 {
			break
		}
	}
	return parseOSCColor(string(buf))
}
