// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"bytes"
	"errors"
	"image/png"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/svg/svgtest"
)

func newTestGraphic(t *testing.T, w, h int, format font.FileFormat) *UGraphicG2d {
	t.Helper()
	b := NewEmptyImageBuilder("", w, h, color.NewXColor(255, 255, 255), DefaultImageLimit)
	return New(color.White, color.IdentityMapper, font.StringBounderAwt{}, b, 1, format)
}

func expectPanic(t *testing.T, f func()) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("no panic")
		}
		if e, ok := r.(error); ok {
			msg = e.Error()
		}
	}()
	f()
	return ""
}

func TestImageLimit(t *testing.T) {
	for v, want := range map[string]int{"": 4096, "1000": 1000, "abc": 4096, "-5": 4096, "12a": 4096} {
		if got, err := ImageLimit(v); got != want || err != nil {
			t.Errorf("%q: %d %v", v, got, err)
		}
	}
	var re *color.RuntimeError
	if _, err := ImageLimit("99999999999"); !errors.As(err, &re) || re.Class != "java.lang.NumberFormatException" {
		t.Errorf("overflow: %v", err)
	}
}

func TestEmptyImageBuilder(t *testing.T) {
	if msg := expectPanic(t, func() { NewEmptyImageBuilder("", 0, 10, color.NewXColor(0, 0, 0), 4096) }); !strings.Contains(msg,
		"IllegalArgumentException") {
		t.Errorf("panic %q", msg)
	}
	b := NewEmptyImageBuilder("", 5000, 20, color.NewXColorAlpha(0, 0, 0, 0), 4096)
	im := b.Image()
	if im.Width() != 4096 || im.Height() != 20 || !im.HasAlpha() || im.ARGB(0, 0) != 0 {
		t.Errorf("image %dx%d %v %x", im.Width(), im.Height(), im.HasAlpha(), im.ARGB(0, 0))
	}
	b = NewEmptyImageBuilder("", 3, 2, color.NewXColor(10, 20, 30), 4096)
	if im := b.Image(); im.HasAlpha() || im.ARGB(2, 1) != 0xFF0A141E || len(im.Pix()) != 6 {
		t.Errorf("opaque image %v %x", im.HasAlpha(), im.ARGB(2, 1))
	}
	nrgba := b.Image().ToNRGBA()
	if c := nrgba.NRGBAAt(1, 1); c.R != 10 || c.G != 20 || c.B != 30 || c.A != 255 {
		t.Errorf("NRGBA %v", c)
	}
	if r, _, _, a := b.Image().At(-1, 0).RGBA(); r != 0 || a != 0 || b.Image().Bounds().Dx() != 3 ||
		b.Image().ColorModel() == nil {
		t.Error("image.Image methods")
	}
	// A long watermark is broken into lines.
	b = NewEmptyImageBuilder("a very long watermark text that does not fit", 60, 40, color.NewXColor(255, 255, 255),
		4096)
	gray := 0
	for _, v := range b.Image().Pix() {
		if v != 0xFFFFFFFF {
			gray++
		}
	}
	if gray == 0 {
		t.Error("no watermark")
	}
}

func TestStyledStrings(t *testing.T) {
	const b, e = "\uEEF2", "\uEEF1"
	for in, want := range map[string]string{
		"plain":                         "plain",
		"a" + b + "b" + e + "c":         "a|*b|c",
		b + "bold":                      "*bold",
		"xy" + b + "z" + e:              "xy|*z",
		b + "a" + e + b + "b" + e + "t": "*a|*b|t",
	} {
		var parts []string
		for _, s := range buildStyledStrings(in) {
			if s.bold {
				parts = append(parts, "*"+s.text)
			} else {
				parts = append(parts, s.text)
			}
		}
		if got := strings.Join(parts, "|"); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
	// An end marker before the start marker: upstream fails, here the bold
	// part is empty.
	if got := buildStyledStrings("a" + e + b + "c"); len(got) != 3 || got[1].text != "" || !got[1].bold ||
		got[2].text != "c" || !got[2].bold {
		t.Errorf("%+v", got)
	}
}

func TestLinks(t *testing.T) {
	g := newTestGraphic(t, 50, 50, font.FormatPNG)
	u1 := klimt.NewURL("https://example.com", "", false, "", false)
	u2 := klimt.NewURL("java\tscript:alert(1)", "", false, "", false)
	g.StartURL(u1)
	g.StartURL(u2)
	g.Apply(color.Black).Apply(klimt.NewUTranslate(5, 5)).Draw(klimt.URectangleBuild(10, 10))
	g.CloseURL()
	g.CloseURL()
	g.StartURL(u1)
	g.CloseURL()
	if all := g.AllUrlsEncountered(); len(all) != 1 || all[0] != u1 || !u1.HasData() {
		t.Errorf("links %v", all)
	}
	if !isJavascriptLink(" JavaScript:x") || isJavascriptLink("https://java.com") {
		t.Error("javascript detection")
	}
	g.AllowJavascriptInLink(true)
	g.StartURL(u2)
	g.CloseURL()
	if len(g.AllUrlsEncountered()) != 2 {
		t.Error("javascript link not kept when allowed")
	}
	expectPanic(t, func() { g.StartURL(nil) })
}

func TestGraphicState(t *testing.T) {
	g := newTestGraphic(t, 20, 20, font.FormatPNG)
	off := g.Apply(klimt.AntiAliasingOff).(*UGraphicG2d)
	if off.antiAliasing != klimt.AntiAliasingOff || g.antiAliasing != klimt.AntiAliasingOn {
		t.Error("antialiasing is per graphic")
	}
	if g.DpiFactor() != 1 || g.Image() == nil || off.Image() != g.Image() {
		t.Error("state")
	}
	var buf bytes.Buffer
	if err := newTestGraphic(t, 20, 20, font.FormatPNGEmpty).WriteToStream(&buf, "x", 96); err != nil ||
		buf.Len() != 0 {
		t.Errorf("PNG_EMPTY wrote %d bytes, %v", buf.Len(), err)
	}
	if err := newTestGraphic(t, 20, 20, font.FormatRAW).WriteToStream(&buf, "", 96); err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(&buf); err != nil {
		t.Errorf("RAW is written as PNG: %v", err)
	}
	// A quadratic segment is not supported, as upstream (UPath.QuadTo
	// stores a cubic curve).
	p := klimt.NewUPath("", "")
	p.MoveTo(0, 0)
	p.Add([]float64{5, 5, 10, 0}, geom.SegQuadTo)
	if msg := expectPanic(t, func() { g.Apply(color.Black).Draw(p) }); !strings.Contains(msg,
		"UnsupportedOperationException") {
		t.Errorf("panic %q", msg)
	}
	// Degenerate shapes draw nothing.
	ug := g.Apply(color.Black).Apply(klimt.Bg(color.Red))
	ug.Draw(klimt.URectangleBuild(-5, 10))
	ug.Draw(klimt.UEllipseBuild(-1, -1))
	ug.Draw(klimt.NewULine(0, 0))
	ug.Draw(klimt.NewUPolygon())
	ug.Apply(klimt.NewUStroke(2, 2, 1)).Draw(klimt.NewULine(0, 0))
}

func TestBlendTables(t *testing.T) {
	for i := range 256 {
		if mul8[255][i] != uint8(i) || mul8[i][255] != uint8(i) || mul8[0][i] != 0 {
			t.Fatalf("mul8 %d", i)
		}
		if i > 0 && div8[i][i] != 255 {
			t.Fatalf("div8 %d", i)
		}
	}
	if mul8[128][128] != 64 || div8[128][64] != 128 {
		t.Errorf("mul8 %d div8 %d", mul8[128][128], div8[128][64])
	}
	im := NewImage(1, 1, true)
	im.blend(0, 0x80FF0000, 255)
	if im.pix[0] != 0x80FF0000 {
		t.Errorf("blend on transparent %x", im.pix[0])
	}
	im = NewImage(1, 1, false)
	im.blend(0, 0xFFFFFFFF, 128)
	if im.pix[0] != 0xFF808080 {
		t.Errorf("half white on black %x", im.pix[0])
	}
}

func TestCenteredCharacterOutsideTable(t *testing.T) {
	g := newTestGraphic(t, 60, 60, font.FormatPNG)
	// Size 40 is outside of upstream's UnusedSpace table: the glyph bounds
	// center the character.
	g.Apply(color.Black).Apply(klimt.NewUTranslate(30, 30)).Draw(klimt.NewUCenteredCharacter('é',
		font.Build("SansSerif", font.NormalFace(), 40)))
	dark := 0
	for _, v := range g.Image().Pix() {
		if v&0xFF < 128 {
			dark++
		}
	}
	if dark < 20 {
		t.Errorf("%d dark pixels", dark)
	}
	if cx, cy := glyphCenter(font.NewAwtFont("SansSerif", font.AwtPlain, 12), ' '); cx != 0 || cy != 0 {
		t.Errorf("space center %v %v", cx, cy)
	}
}

// TestConcurrentRenders draws the test cases in parallel (for the race
// detector): renders share only read-only data.
func TestConcurrentRenders(t *testing.T) {
	cases, err := svgtest.ParseFile("testdata/cases.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]byte{}
	for _, c := range cases[:8] {
		out, err := runCase(c)
		if err != nil {
			t.Fatal(err)
		}
		want[c.Name] = out
	}
	var wg sync.WaitGroup
	for range 4 {
		for _, c := range cases[:8] {
			wg.Go(func() {
				out, err := runCase(c)
				if err != nil || !bytes.Equal(out, want[c.Name]) {
					t.Errorf("%s: concurrent render differs (%v)", c.Name, err)
				}
			})
		}
	}
	wg.Wait()
}

// BenchmarkSequence draws the sequence-like test case.
func BenchmarkSequence(b *testing.B) {
	cases, err := svgtest.ParseFile("testdata/cases.txt")
	if err != nil {
		b.Fatal(err)
	}
	for _, c := range cases {
		if c.Name != "sequence" {
			continue
		}
		for b.Loop() {
			if _, err := runCase(c); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// TestDegenerateCurveStroke draws the cylinder of a database participant,
// whose curves have control points on their end points: the stroke must
// not fill the area between the curve and its chord.
func TestDegenerateCurveStroke(t *testing.T) {
	g := newTestGraphic(t, 130, 90, font.FormatPNG)
	w, h := 36.0, 46.0
	shape := klimt.UPathNone()
	shape.MoveTo(0, 10)
	shape.CubicTo(0, 0, w/2, 0, w/2, 0)
	shape.CubicTo(w/2, 0, w, 0, w, 10)
	shape.LineTo(w, h-10)
	shape.CubicTo(w, h, w/2, h, w/2, h)
	shape.CubicTo(w/2, h, 0, h, 0, h-10)
	shape.LineTo(0, 10)
	g.Apply(svgtest.Color("#181818")).Apply(klimt.Bg(svgtest.Color("#E2E2F0"))).
		Apply(klimt.UStrokeWithThickness(0.5)).Apply(klimt.NewUTranslate(72.482, 24)).Draw(shape)
	if got := g.Image().ARGB(78, 28); got != 0xFFE2E2F0 {
		t.Errorf("pixel inside the cylinder %x", got)
	}
}
