// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"image"
	"image/color"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestCompareOutcome(t *testing.T) {
	tests := []struct {
		name     string
		ref, got Outcome
		wantErr  string
	}{
		{"both render", Outcome{}, Outcome{}, ""},
		{"same error line", Outcome{true, 3}, Outcome{true, 3}, ""},
		{"different error line", Outcome{true, 3}, Outcome{true, 4}, "error line 4, reference reports line 3"},
		{"reference fails", Outcome{true, 2}, Outcome{}, "but the diagram rendered"},
		{"implementation fails", Outcome{}, Outcome{true, 5}, "but the diagram failed at line 5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { checkErr(t, CompareOutcome(tt.ref, tt.got), tt.wantErr) })
	}
}

func TestSVGTexts(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><g>
<text x="1" y="2">Alice</text><rect/>
<text x="3" y="4" font-weight="bold">  hello
   &lt;world&gt;  </text>
<text x="5"><tspan>a</tspan> <tspan>b</tspan></text>
<text x="6">   </text>
</g></svg>`
	texts, err := SVGTexts([]byte(svg))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Alice", "hello <world>", "a b"}
	if !slices.Equal(texts, want) {
		t.Fatalf("texts = %q, want %q", texts, want)
	}
	if _, err := SVGTexts([]byte("<svg><text>open")); err == nil {
		t.Fatal("expected an error for truncated SVG")
	}
}

func TestCompareTexts(t *testing.T) {
	checkErr(t, CompareTexts([]string{"Alice", "Bob", "hi"}, []string{"hi", "Bob", "Alice"}), "")
	checkErr(t, CompareTexts([]string{"Alice", "Bob", "hi"}, []string{"Alice", "Bob"}), `missing texts "hi"`)
	checkErr(t, CompareTexts([]string{"Alice"}, []string{"Alice", "Alice"}), `unexpected texts "Alice"`)
	checkErr(t, CompareTexts([]string{"a", "b"}, []string{"a", "c"}), `missing texts "b"; unexpected texts "c"`)
}

func TestCompareSize(t *testing.T) {
	checkErr(t, CompareSize(100, 200, 100, 200, SizeTolerance), "")
	checkErr(t, CompareSize(100, 200, 105, 190, SizeTolerance), "")
	checkErr(t, CompareSize(100, 200, 106, 200, SizeTolerance), "width 106 is not within 5% of the reference 100")
	checkErr(t, CompareSize(100, 200, 100, 189, SizeTolerance), "height 189 is not within 5%")
}

func TestCompareText(t *testing.T) {
	checkErr(t, CompareText("a\n b\n", "a  \n b\t\n"), "")
	checkErr(t, CompareText("a\r\nb\r\n", "a\nb\n"), "")
	checkErr(t, CompareText("a\nb\n", "a\nc\n"), `line 2 differs: got "c", reference "b"`)
	checkErr(t, CompareText("a\n b\n", "a\nb\n"), "line 2 differs") // leading whitespace counts
	checkErr(t, CompareText("a\nb\n", "a\n"), `line 2 differs: got "", reference "b"`)
	checkErr(t, CompareText("a\nb", "a"), "output has 1 lines, reference has 2")
	checkErr(t, CompareText("a", "a\nb"), `first extra line "b"`)
}

// boxes draws dark rectangles on a white image, a stand-in for diagrams.
func boxes(w, h int, rects ...image.Rectangle) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.White)
		}
	}
	for _, r := range rects {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				if x == r.Min.X || x == r.Max.X-1 || y == r.Min.Y || y == r.Max.Y-1 {
					img.Set(x, y, color.Black)
				}
			}
		}
	}
	return img
}

func TestCompareImages(t *testing.T) {
	ref := boxes(120, 80, image.Rect(10, 10, 50, 40), image.Rect(60, 30, 110, 70))

	score, err := CompareImages(ref, ref, MinSSIM)
	if err != nil || score < 0.9999 {
		t.Fatalf("identical images: score %f, err %v", score, err)
	}

	// A one-pixel shift of one box, as from slightly different text metrics.
	shifted := boxes(120, 80, image.Rect(11, 10, 51, 40), image.Rect(60, 30, 110, 70))
	if score, err := CompareImages(ref, shifted, MinSSIM); err != nil {
		t.Fatalf("one-pixel shift should pass: score %f, err %v", score, err)
	}

	// A slightly larger output is resampled to the reference size.
	larger := boxes(124, 83, image.Rect(10, 10, 52, 41), image.Rect(62, 31, 114, 72))
	if score, err := CompareImages(ref, larger, MinSSIM); err != nil {
		t.Fatalf("resampled output should pass: score %f, err %v", score, err)
	}

	// A different layout must fail.
	different := boxes(120, 80, image.Rect(5, 45, 40, 75), image.Rect(70, 5, 115, 25))
	if score, err := CompareImages(ref, different, MinSSIM); err == nil {
		t.Fatalf("different layout should fail, score %f", score)
	}

	// Transparent pixels count as white.
	transparent := image.NewRGBA(image.Rect(0, 0, 120, 80))
	white := boxes(120, 80)
	if score, err := CompareImages(white, transparent, MinSSIM); err != nil {
		t.Fatalf("transparent should equal white: score %f, err %v", score, err)
	}
}

func TestResvg(t *testing.T) {
	r := Resvg{Path: "../../.cache/tools/resvg", FontsDir: "../../.cache/fonts/dejavu"}
	if _, err := exec.LookPath(r.Path); err != nil {
		t.Skip("resvg not fetched; run tools/reference/fetch-tools.sh")
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="20"><rect x="5" y="5" width="30" height="10" fill="#000"/><text x="2" y="18" font-family="sans-serif" font-size="8">Ag</text></svg>`
	img, err := r.Rasterize([]byte(svg), 80, 40)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 80 || b.Dy() != 40 {
		t.Fatalf("size %v, want 80x40", b)
	}
}

func checkErr(t *testing.T, err error, want string) {
	t.Helper()
	switch {
	case want == "" && err != nil:
		t.Fatalf("unexpected error: %v", err)
	case want != "" && err == nil:
		t.Fatalf("expected an error containing %q", want)
	case want != "" && !strings.Contains(err.Error(), want):
		t.Fatalf("error %q does not contain %q", err, want)
	}
}

func TestSVGSizeAndMetadata(t *testing.T) {
	svg := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" style="width:240px" width="240px" height="259.5px" viewBox="0 0 120 130"><g><?plantuml-src SoWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiX8pSd9vt98pKi1IW80?></g></svg>`)
	w, h, ok := SVGSize(svg)
	if !ok || w != 240 || h != 260 {
		t.Errorf("SVGSize = %d, %d, %v", w, h, ok)
	}
	if _, _, ok := SVGSize([]byte("<svg/>")); ok {
		t.Error("an SVG without size must report ok false")
	}
	if !HasSVGSourceMetadata(svg) || HasSVGSourceMetadata([]byte("<svg/>")) {
		t.Error("HasSVGSourceMetadata")
	}
}

func TestHasPNGSourceMetadata(t *testing.T) {
	chunk := func(kind, body string) string {
		n := len(body)
		return string([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}) + kind + body + "CRC!"
	}
	png := "\x89PNG\r\n\x1a\n" + chunk("IHDR", "0123456789abc")
	if HasPNGSourceMetadata([]byte(png + chunk("IEND", ""))) {
		t.Error("PNG without text chunk")
	}
	if !HasPNGSourceMetadata([]byte(png + chunk("iTXt", "plantuml\x00\x01\x00\x00\x00data") + chunk("IEND", ""))) {
		t.Error("PNG with plantuml iTXt chunk")
	}
	if HasPNGSourceMetadata([]byte(png + chunk("tEXt", "copyleft\x00x") + chunk("IEND", ""))) {
		t.Error("other keywords do not count")
	}
	if HasPNGSourceMetadata([]byte("not a png")) {
		t.Error("not a PNG")
	}
}
