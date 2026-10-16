// SPDX-License-Identifier: MIT
// Provenance: original

package extremity

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// The helpers of the harness testdata/javagen/ExtremityGolden.java.

const goldenDir = "testdata/golden"

func d2s(d float64) string { return jfmt.DoubleToString(d) }

func hc(c color.HColor) string {
	if c == nil {
		return "null"
	}
	return c.AsString()
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
	}
	return "?"
}

// rec is the harness's Rec: a UGraphic that records what is drawn.
type rec struct {
	log    *[]string
	t      klimt.UTranslate
	color  color.HColor
	back   color.HColor
	stroke klimt.UStroke
}

var _ klimt.UGraphic = (*rec)(nil)

func (r *rec) StringBounder() font.StringBounder { return nil }

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
	if change == nil {
		// Java's Rec calls getClass() on a null change.
		panic(&color.RuntimeError{Class: "java.lang.NullPointerException"})
	}
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
	default:
		*r.log = append(*r.log, "apply ?")
		return r
	}
	return &c
}

func (r *rec) ColorMapper() *color.ColorMapper            { return color.IdentityMapper }
func (r *rec) StartURL(*klimt.URL)                        { *r.log = append(*r.log, "startUrl") }
func (r *rec) CloseURL()                                  { *r.log = append(*r.log, "closeUrl") }
func (r *rec) StartGroup(*klimt.UGroup)                   { *r.log = append(*r.log, "startGroup") }
func (r *rec) CloseGroup()                                { *r.log = append(*r.log, "closeGroup") }
func (r *rec) FlushUg()                                   {}
func (r *rec) MatchesProperty(string) bool                { return false }
func (r *rec) DefaultBackground() color.HColor            { return color.White }
func (r *rec) WriteToStream(io.Writer, string, int) error { return nil }

// step runs f, recording a panic as the harness records an exception: the
// simple class name and the message ("null" when there is none), without
// the message of a NullPointerException.
func step(log *[]string, f func()) {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case *color.RuntimeError:
			name := r.Class[strings.LastIndexByte(r.Class, '.')+1:]
			if name == "NullPointerException" {
				*log = append(*log, "ERR "+name)
				return
			}
			msg := r.Message
			if msg == "" {
				msg = "null"
			}
			*log = append(*log, "ERR "+name+" "+msg)
		default:
			panic(r)
		}
	}()
	f()
}

func readGolden(t *testing.T, name string) []string {
	t.Helper()
	f, err := os.Open(filepath.Join(goldenDir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(zr)
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
	want := readGolden(t, name)
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
