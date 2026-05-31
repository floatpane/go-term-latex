package termlatex

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"sort"
)

// renderHalfBlock renders img using Unicode half-block characters (▀ U+2580).
// Each cell encodes two vertically stacked pixels via 24-bit foreground (top)
// and background (bottom) ANSI color. Works on any UTF-8 truecolor terminal.
func renderHalfBlock(w io.Writer, img image.Image) error {
	bw := bufio.NewWriter(w)
	b := img.Bounds()
	height := b.Dy()

	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		for x := b.Min.X; x < b.Max.X; x++ {
			top := toRGB(img.At(x, y))

			var bot [3]uint8
			if y+1 < height+b.Min.Y {
				bot = toRGB(img.At(x, y+1))
			}

			if _, err := fmt.Fprintf(bw, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
				top[0], top[1], top[2],
				bot[0], bot[1], bot[2],
			); err != nil {
				return err
			}
		}
		if _, err := bw.WriteString("\x1b[0m\n"); err != nil {
			return err
		}
	}
	return bw.Flush()
}

func toRGB(c color.Color) [3]uint8 {
	r, g, b, _ := c.RGBA()
	return [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

const kittyChunk = 4096

// renderKitty encodes img as a Kitty graphics protocol sequence (PNG, f=100)
// and writes it chunked to w.
func renderKitty(w io.Writer, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("kitty: png encode: %w", err)
	}
	enc := base64.StdEncoding.EncodeToString(buf.Bytes())

	for i := 0; i < len(enc); i += kittyChunk {
		end := i + kittyChunk
		if end > len(enc) {
			end = len(enc)
		}
		chunk := enc[i:end]
		more := 1
		if end == len(enc) {
			more = 0
		}

		var ctrl string
		if i == 0 {
			// a=T: transmit & display. f=100: PNG. q=2: suppress ACK.
			ctrl = fmt.Sprintf("a=T,f=100,q=2,m=%d", more)
		} else {
			ctrl = fmt.Sprintf("q=2,m=%d", more)
		}
		if _, err := fmt.Fprintf(w, "\x1b_G%s;%s\x1b\\", ctrl, chunk); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

// renderSixel encodes img as a DEC Sixel sequence, quantizing to 256 colors
// via median cut.
func renderSixel(w io.Writer, img image.Image) error {
	bw := bufio.NewWriter(w)
	b := img.Bounds()
	width, height := b.Dx(), b.Dy()

	palette := medianCut(img, 256)

	if _, err := bw.WriteString("\x1bPq"); err != nil {
		return err
	}
	for i, c := range palette {
		r, g, bl, _ := c.RGBA()
		// Sixel color values are percentages (0-100).
		if _, err := fmt.Fprintf(bw, "#%d;2;%d;%d;%d", i,
			int(r>>8)*100/255, int(g>>8)*100/255, int(bl>>8)*100/255,
		); err != nil {
			return err
		}
	}

	for bandY := 0; bandY < height; bandY += 6 {
		bands := make([][]byte, len(palette))
		for i := range bands {
			bands[i] = make([]byte, width)
		}
		for x := 0; x < width; x++ {
			for bit := 0; bit < 6; bit++ {
				py := bandY + bit
				if py >= height {
					break
				}
				idx := nearestColor(palette, img.At(b.Min.X+x, b.Min.Y+py))
				bands[idx][x] |= 1 << uint(bit)
			}
		}
		for i, band := range bands {
			if allZero(band) {
				continue
			}
			if _, err := fmt.Fprintf(bw, "#%d", i); err != nil {
				return err
			}
			for _, v := range band {
				if err := bw.WriteByte(v + 63); err != nil {
					return err
				}
			}
			if err := bw.WriteByte('$'); err != nil { // carriage return within row
				return err
			}
		}
		if err := bw.WriteByte('-'); err != nil { // next sixel row
			return err
		}
	}
	if _, err := bw.WriteString("\x1b\\\n"); err != nil {
		return err
	}
	return bw.Flush()
}

func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// medianCut quantizes img to at most n colors using the median cut algorithm.
func medianCut(img image.Image, n int) []color.Color {
	b := img.Bounds()
	pixels := make([]color.RGBA, 0, b.Dx()*b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x8000 {
				continue
			}
			pixels = append(pixels, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: 255})
		}
	}

	buckets := [][]color.RGBA{pixels}
	for len(buckets) < n && anyBucketSplittable(buckets) {
		buckets = splitLargest(buckets)
	}

	palette := make([]color.Color, len(buckets))
	for i, bucket := range buckets {
		palette[i] = average(bucket)
	}
	return palette
}

func anyBucketSplittable(bs [][]color.RGBA) bool {
	for _, b := range bs {
		if len(b) > 1 {
			return true
		}
	}
	return false
}

func splitLargest(buckets [][]color.RGBA) [][]color.RGBA {
	idx := 0
	for i, b := range buckets {
		if len(b) > len(buckets[idx]) {
			idx = i
		}
	}
	bucket := buckets[idx]
	ch := dominantChannel(bucket)
	sort.Slice(bucket, func(i, j int) bool {
		switch ch {
		case 0:
			return bucket[i].R < bucket[j].R
		case 1:
			return bucket[i].G < bucket[j].G
		default:
			return bucket[i].B < bucket[j].B
		}
	})
	mid := len(bucket) / 2
	result := make([][]color.RGBA, 0, len(buckets)+1)
	result = append(result, buckets[:idx]...)
	result = append(result, bucket[:mid], bucket[mid:])
	result = append(result, buckets[idx+1:]...)
	return result
}

func dominantChannel(pixels []color.RGBA) int {
	var minR, minG, minB uint8 = 255, 255, 255
	var maxR, maxG, maxB uint8
	for _, p := range pixels {
		if p.R < minR {
			minR = p.R
		}
		if p.R > maxR {
			maxR = p.R
		}
		if p.G < minG {
			minG = p.G
		}
		if p.G > maxG {
			maxG = p.G
		}
		if p.B < minB {
			minB = p.B
		}
		if p.B > maxB {
			maxB = p.B
		}
	}
	rr := int(maxR) - int(minR)
	gg := int(maxG) - int(minG)
	bb := int(maxB) - int(minB)
	if rr >= gg && rr >= bb {
		return 0
	}
	if gg >= bb {
		return 1
	}
	return 2
}

func average(pixels []color.RGBA) color.Color {
	if len(pixels) == 0 {
		return color.RGBA{}
	}
	var r, g, b int64
	for _, p := range pixels {
		r += int64(p.R)
		g += int64(p.G)
		b += int64(p.B)
	}
	n := int64(len(pixels))
	return color.RGBA{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n), A: 255}
}

func nearestColor(palette []color.Color, c color.Color) int {
	r0, g0, b0, _ := c.RGBA()
	best := 0
	bestDist := math.MaxFloat64
	for i, p := range palette {
		r1, g1, b1, _ := p.RGBA()
		dr := float64(int(r0>>8) - int(r1>>8))
		dg := float64(int(g0>>8) - int(g1>>8))
		db := float64(int(b0>>8) - int(b1>>8))
		d := dr*dr + dg*dg + db*db
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	return best
}
