package termlatex

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

// recolor remaps a black-on-white render to the given theme: white background
// pixels become bg, black glyph pixels become fg, and antialiased edges are
// linearly blended between the two. The result is fully opaque so it displays
// correctly in every terminal protocol.
func recolor(pngBytes []byte, t Theme) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("%w: decode png: %w", ErrRenderFailed, err)
	}

	fr, fg, fb := rgb8(t.Fg)
	br, bg, bb := rgb8(t.Bg)

	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := src.At(x, y).RGBA()
			// Perceived luminance, 0 (black glyph) .. 255 (white paper).
			lum := (299*(r>>8) + 587*(g>>8) + 114*(bl>>8)) / 1000
			dst.SetNRGBA(x, y, color.NRGBA{
				R: lerp(fr, br, lum),
				G: lerp(fg, bg, lum),
				B: lerp(fb, bb, lum),
				A: 0xff,
			})
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, dst); err != nil {
		return nil, fmt.Errorf("%w: encode png: %w", ErrRenderFailed, err)
	}
	return out.Bytes(), nil
}

// lerp blends fg→bg by t/255 (t=0 → fg, t=255 → bg).
func lerp(fg, bg uint8, t uint32) uint8 {
	if t > 255 {
		t = 255
	}
	return uint8((uint32(fg)*(255-t) + uint32(bg)*t) / 255)
}

func rgb8(c color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := c.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
}
