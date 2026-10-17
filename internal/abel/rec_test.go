// SPDX-License-Identifier: MIT
// Provenance: original

package abel_test

import (
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// The helpers of the harness internal/abel/testdata/javagen/BodyGolden.java,
// which are those of internal/klimt/rose/testdata/javagen/RoseGolden.java
// (copied from internal/decoration/symbol/rec_test.go).

const goldenDir = "testdata/golden"

func d2s(d float64) string { return jfmt.DoubleToString(d) }

func hc(c color.HColor) string {
	if c == nil {
		return "null"
	}
	return c.AsString()
}

// fakeBounder is the harness's StringBounder: 0.6 em per UTF-16 unit and a
// height of 1.25 em.
type fakeBounder struct{ font.StringBounderDefaults }

func (fakeBounder) CalculateDimension(f font.UFont, text string) font.Dimension {
	n := len(utf16.Encode([]rune(text)))
	return font.Dimension{Width: float64(n) * f.Size2D() * 0.6, Height: f.Size2D() * 1.25}
}

func (fakeBounder) FileFormat() font.FileFormat { return font.FormatPNG }

// bounders are the harness's bounders: the fixed one and the SVG one of the
// reference fonts.
var bounders = []struct {
	name    string
	bounder font.StringBounder
}{{"fake", fakeBounder{}}, {"svg", font.NewStringBounderSvg(nil)}}

func uf(f font.UFont) string {
	return f.Family("", font.ContextSVG) + "/" + f.FontFace().String() + "/" + strconv.Itoa(f.Size())
}

// fcString is the harness's fcString: font, color, styles, extended color,
// position shift and SVG attributes.
func fcString(fc *klimt.FontConfiguration) string {
	var styles []string
	for s := font.FontStylePlain; s <= font.FontStyleBackcolor; s++ {
		if fc.ContainsStyle(s) {
			styles = append(styles, s.String())
		}
	}
	result := uf(fc.Font()) + " " + hc(fc.Color()) + " [" + strings.Join(styles, ",") + "] ext=" +
		hc(fc.ExtendedColor()) + " space=" + strconv.Itoa(fc.Space())
	if attr := fc.Attributes(); attr != nil && len(attr.Attributes()) > 0 {
		var parts []string
		for _, k := range slices.Sorted(maps.Keys(attr.Attributes())) {
			parts = append(parts, k+"="+attr.Attributes()[k])
		}
		result += " attr={" + strings.Join(parts, ", ") + "}"
	}
	return result
}

func javaList(items []string) string { return "[" + strings.Join(items, ", ") + "]" }

func describe(shape klimt.UShape) string {
	switch s := shape.(type) {
	case *klimt.URectangle:
		return "rect " + d2s(s.Width()) + "," + d2s(s.Height()) + " r=" + d2s(s.Rx()) + "," + d2s(s.Ry()) + " sh=" +
			d2s(s.DeltaShadow())
	case *klimt.UEllipse:
		return "ellipse " + d2s(s.Width()) + "," + d2s(s.Height()) + " " + d2s(s.Start()) + "," + d2s(s.Extend()) +
			" sh=" + d2s(s.DeltaShadow())
	case *klimt.ULine:
		return "line " + d2s(s.DX()) + "," + d2s(s.DY())
	case *klimt.UPolygon:
		var pts []string
		for _, pt := range s.Points() {
			pts = append(pts, d2s(pt.X)+":"+d2s(pt.Y))
		}
		return "polygon " + javaList(pts) + " sh=" + d2s(s.DeltaShadow())
	case *klimt.UPath:
		var segs []string
		for _, seg := range s.Segments() {
			segs = append(segs, seg.String())
		}
		return "path " + javaList(segs) + " sh=" + d2s(s.DeltaShadow())
	case *klimt.UText:
		return "text '" + s.Text() + "' " + fcString(s.FontConfiguration())
	case *klimt.UImage:
		return "image " + strconv.Itoa(s.Width()) + "x" + strconv.Itoa(s.Height()) + " scale=" + d2s(s.ImageScale())
	case *klimt.UImageSvg:
		return "svg " + d2s(s.Width()) + "x" + d2s(s.Height())
	case *klimt.UEmpty:
		return "empty " + d2s(s.Width()) + "," + d2s(s.Height())
	case *klimt.UHorizontalLine:
		return "hline"
	}
	return "?"
}

// rec is the harness's Rec: a UGraphic that records what is drawn.
type rec struct {
	log     *[]string
	bounder font.StringBounder
	t       klimt.UTranslate
	color   color.HColor
	back    color.HColor
	stroke  klimt.UStroke
}

var _ klimt.UGraphic = (*rec)(nil)

func newRec(log *[]string, bounder font.StringBounder) *rec {
	return &rec{log: log, bounder: bounder, color: color.Black, back: color.None(), stroke: klimt.UStrokeSimple()}
}

func (r *rec) StringBounder() font.StringBounder { return r.bounder }

type recParam struct{ r *rec }

func (p recParam) Color() color.HColor     { return p.r.color }
func (p recParam) Backcolor() color.HColor { return p.r.back }
func (p recParam) Stroke() klimt.UStroke   { return p.r.stroke }
func (p recParam) IsHidden() bool          { return false }
func (p recParam) Pattern() klimt.UPattern { return klimt.UPatternFull }

func (r *rec) Param() klimt.UParam { return recParam{r} }

func (r *rec) Draw(shape klimt.UShape) {
	*r.log = append(*r.log, "@"+d2s(r.t.Dx())+","+d2s(r.t.Dy())+" c="+hc(r.color)+" b="+hc(r.back)+" s="+
		r.stroke.String()+" "+describe(shape))
}

func (r *rec) Apply(change klimt.UChange) klimt.UGraphic {
	c := *r
	switch ch := change.(type) {
	case klimt.UTranslate:
		c.t = ch.Compose(r.t)
	case color.HColor:
		c.color = ch
	case klimt.UBackground:
		c.back = ch.BackColor()
	case klimt.UStroke:
		c.stroke = ch
	case klimt.UAntiAliasing:
		*r.log = append(*r.log, "apply UAntiAliasing")
		return r
	default:
		*r.log = append(*r.log, "apply ?")
		return r
	}
	return &c
}

func (r *rec) ColorMapper() *color.ColorMapper { return color.IdentityMapper }
func (r *rec) StartURL(url *klimt.URL) {
	*r.log = append(*r.log, "startUrl "+url.URL()+" tooltip="+url.Tooltip()+" label="+url.Label())
}
func (r *rec) CloseURL() { *r.log = append(*r.log, "closeUrl") }
func (r *rec) StartGroup(group *klimt.UGroup) {
	var parts []string
	for _, e := range group.Entries() {
		parts = append(parts, e.Key.String()+"="+e.Value)
	}
	*r.log = append(*r.log, "startGroup {"+strings.Join(parts, ", ")+"}")
}
func (r *rec) CloseGroup()                                { *r.log = append(*r.log, "closeGroup") }
func (r *rec) FlushUg()                                   {}
func (r *rec) MatchesProperty(string) bool                { return false }
func (r *rec) DefaultBackground() color.HColor            { return color.White }
func (r *rec) WriteToStream(io.Writer, string, int) error { return nil }

// step runs f, recording a panic as the harness records an exception.
func step(log *[]string, f func()) {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case *color.RuntimeError:
			*log = append(*log, "ERR "+r.Class[strings.LastIndexByte(r.Class, '.')+1:])
		case *color.NoSuchColorRuntimeError:
			*log = append(*log, "ERR NoSuchColorRuntimeException")
		default:
			panic(r)
		}
	}()
	f()
}

func readLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(goldenDir, name))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func compareGolden(t *testing.T, name string, got []string) {
	t.Helper()
	want := readLines(t, name)
	if len(got) != len(want) {
		t.Errorf("%s: %d lines, want %d", name, len(got), len(want))
	}
	failures := 0
	for i := 0; i < min(len(got), len(want)); i++ {
		if got[i] != want[i] {
			failures++
			if failures <= 10 {
				t.Errorf("%s line %d:\n got: %s\nwant: %s", name, i+1, got[i], want[i])
			}
		}
	}
	if failures > 10 {
		t.Errorf("%s: %d more differences", name, failures-10)
	}
}
