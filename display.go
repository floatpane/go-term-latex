package termlatex

import (
	"fmt"
	"image"
	"io"

	xdraw "golang.org/x/image/draw"
)

// displayImage scales img to fit the terminal and writes it to w using the
// resolved protocol.
func displayImage(w io.Writer, img *image.NRGBA, opts Options) error {
	proto := opts.Protocol
	if proto == AutoProtocol {
		proto = bestProtocol()
	}

	maxW, maxH := effectiveDimensions(opts, proto)
	img = fit(img, maxW, maxH)

	var err error
	switch proto {
	case Kitty:
		err = renderKitty(w, img)
	case Sixel:
		err = renderSixel(w, img)
	case HalfBlock, AutoProtocol:
		err = renderHalfBlock(w, img)
	}
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDisplay, err)
	}
	return nil
}

// effectiveDimensions returns the pixel bounds for scaling. Explicit
// MaxWidth/MaxHeight win; otherwise they are derived from the terminal size,
// reserving two rows of headroom for the prompt.
//
// For HalfBlock: 1 col = 1 px wide, 1 row = 2 px tall.
// For Kitty/Sixel: cols/rows are multiplied by the cell pixel size.
func effectiveDimensions(opts Options, proto Protocol) (int, int) {
	w, h := opts.MaxWidth, opts.MaxHeight
	if w > 0 && h > 0 {
		return w, h
	}

	cols, rows := detectTermChars()
	cw, ch := detectCellPixels()

	const headroom = 2
	effRows := rows - headroom
	if effRows < 1 {
		effRows = 1
	}

	var tw, th int
	if proto == HalfBlock {
		tw, th = cols, effRows*2
	} else {
		tw, th = cols*cw, effRows*ch
	}
	if w <= 0 {
		w = tw
	}
	if h <= 0 {
		h = th
	}
	return w, h
}

// fit scales src to fit within maxW×maxH preserving aspect ratio. Returns src
// unchanged when it already fits.
func fit(src *image.NRGBA, maxW, maxH int) *image.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if maxW <= 0 || maxH <= 0 || (sw <= maxW && sh <= maxH) {
		return src
	}

	scale := float64(maxW) / float64(sw)
	if scaleH := float64(maxH) / float64(sh); scaleH < scale {
		scale = scaleH
	}
	dw := max(1, int(float64(sw)*scale))
	dh := max(1, int(float64(sh)*scale))

	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, b, xdraw.Src, nil)
	return dst
}
