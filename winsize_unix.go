//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package termlatex

import (
	"os"

	"golang.org/x/sys/unix"
)

func detectTermChars() (cols, rows int) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 220, 50
	}
	defer f.Close() //nolint:errcheck
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 {
		return 220, 50
	}
	return int(ws.Col), int(ws.Row)
}

func detectCellPixels() (cellW, cellH int) {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 8, 16
	}
	defer f.Close() //nolint:errcheck
	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 || ws.Xpixel == 0 || ws.Ypixel == 0 {
		return 8, 16
	}
	return int(ws.Xpixel) / int(ws.Col), int(ws.Ypixel) / int(ws.Row)
}
