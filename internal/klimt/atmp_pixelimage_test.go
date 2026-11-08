// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"sync"
	"testing"
)

// A reused encoder writes the PNG data of an image as png.Encode does: in
// turn, so that buffers used for a large image encode a small one next, and
// in parallel; run with -race.
func TestToPngDataURLReusesEncoders(t *testing.T) {
	gradient := NewPortableImageARGB(16, 16)
	for y := range 16 {
		for x := range 16 {
			gradient.SetRGB(x, y, int32(0xFF000000|uint32(x*16)<<16|uint32(y*16)<<8|0x80))
		}
	}
	transparent := NewPortableImageARGB(1, 1)
	large := NewPortableImageARGB(200, 100)
	for y := range 100 {
		for x := range 200 {
			large.SetRGB(x, y, int32(uint32(x*y%256)<<24|uint32(x)<<16|uint32(y)))
		}
	}
	images := []*PortableImageARGB{gradient, large, transparent, large, gradient}
	want := make([]string, len(images))
	for i, p := range images {
		want[i] = freshPngDataURL(t, p)
	}
	check := func() {
		for i, p := range images {
			if got := p.ToPngDataURL(); got != want[i] {
				t.Errorf("image %d (%dx%d): %d characters, png.Encode gives %d", i, p.width, p.height, len(got), len(want[i]))
			}
		}
	}
	check()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(check)
	}
	wg.Wait()
}

// freshPngDataURL is the PNG data URL of an image encoded by png.Encode,
// with the pixels that ToPngDataURL encodes.
func freshPngDataURL(t *testing.T, p *PortableImageARGB) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, p.width, p.height))
	for y := range p.height {
		for x := range p.width {
			v := uint32(p.RGB(x, y))
			o := img.PixOffset(x, y)
			img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3] = byte(v>>16), byte(v>>8), byte(v), byte(v>>24)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
