//go:build !(linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package termlatex

// queryTheme is unsupported on platforms without termios; callers fall back to
// $COLORFGBG or the default dark theme.
func queryTheme() (Theme, bool) {
	return Theme{}, false
}
