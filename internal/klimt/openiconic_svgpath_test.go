// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"math"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The expected values are the output of testdata/javagen/SpriteGolden.java
// (svgpath.golden), which TestGoldenSvgPath compares in full.

func svgPathMovements(p *SvgPath) string {
	var moves []string
	for _, m := range p.movements {
		moves = append(moves, m.ToSvg())
	}
	return strings.Join(moves, "|")
}

func svgPathCommands(p *SvgPath) string {
	var cmds []string
	for _, c := range p.commands {
		cmds = append(cmds, c.ToSvg())
	}
	return strings.Join(cmds, " ")
}

// svgPathDrawn returns the segments that DrawMe draws, "" when it draws
// nothing.
func svgPathDrawn(p *SvgPath, at *XAffineTransform) string {
	var log []string
	ug := newSpriteRec(&log, false, color.IdentityMapper)
	p.DrawMe(ug, at)
	if len(log) == 0 {
		return ""
	}
	_, segs, _ := strings.Cut(log[0], " path null ")
	return segs
}

func TestSvgPath(t *testing.T) {
	tests := []struct {
		name, data, commands, movements, path string
	}{
		{"absolute lines", "M10 20 L30 40", "M 10 20 L 30 40", "M 10 20 |L 30 40 ",
			"[SEG_MOVETO [10.0, 20.0], SEG_LINETO [30.0, 40.0]]"},
		{"relative lines", "m10 20 l30 40 l-5 -5", "m 10 20 l 30 40 l -5 -5", "M 10 20 |L 40 60 |L 35 55 ",
			"[SEG_MOVETO [10.0, 20.0], SEG_LINETO [40.0, 60.0], SEG_LINETO [35.0, 55.0]]"},
		{"implicit line after move", "M10,20 30,40 50,60", "M 10 20 L 30 40 L 50 60",
			"M 10 20 |L 30 40 |L 50 60 ",
			"[SEG_MOVETO [10.0, 20.0], SEG_LINETO [30.0, 40.0], SEG_LINETO [50.0, 60.0]]"},
		{"horizontal and vertical, closed", "M0 0 H10 V20 H5 Z", "M 0 0 H 10 V 20 H 5 Z",
			"M 0 0 |L 10 0 |L 10 20 |L 5 20 |Z ",
			"[SEG_MOVETO [0.0, 0.0], SEG_LINETO [10.0, 0.0], SEG_LINETO [10.0, 20.0], SEG_LINETO [5.0, 20.0]]"},
		{"relative cubic curves", "m1 2 c3 4 5 6 7 8 c1 1 2 2 3 3", "m 1 2 c 3 4 5 6 7 8 c 1 1 2 2 3 3",
			"M 1 2 |C 4 6 6 8 8 10 |C 9 11 10 12 11 13 ",
			"[SEG_MOVETO [1.0, 2.0], SEG_CUBICTO [4.0, 6.0, 6.0, 8.0, 8.0, 10.0], " +
				"SEG_CUBICTO [9.0, 11.0, 10.0, 12.0, 11.0, 13.0]]"},
		{"smooth curves", "m1 1 s5 5 6 6 s1 1 2 2", "m 1 1 s 5 5 6 6 s 1 1 2 2",
			"M 1 1 |C 6 6 6 6 7 7 |C 8 8 8 8 9 9 ",
			"[SEG_MOVETO [1.0, 1.0], SEG_CUBICTO [6.0, 6.0, 6.0, 6.0, 7.0, 7.0], " +
				"SEG_CUBICTO [8.0, 8.0, 8.0, 8.0, 9.0, 9.0]]"},
		{"quadratic curves", "m0 0 q5 5 10 0 t10 0 t10 0", "m 0 0 q 5 5 10 0 t 10 0 t 10 0",
			"M 0 0 |Q 5 5 10 0 |T 20 0 |T 30 0 ",
			"[SEG_MOVETO [0.0, 0.0], SEG_CUBICTO [5.0, 5.0, 5.0, 5.0, 10.0, 0.0], " +
				"SEG_CUBICTO [15.0, -5.0, 15.0, -5.0, 20.0, 0.0], SEG_CUBICTO [20.0, 0.0, 20.0, 0.0, 30.0, 0.0]]"},
		{"relative arcs", "m10 10 a5 5 30 1 0 10 10 a2 3 0 0 0 -4 -4",
			"m 10 10 a 5 5 30 1 0 10 10 a 2 3 0 0 0 -4 -4", "M 10 10 |A 5 5 30 1 0 20 20 |A 2 3 0 0 0 16 16 ",
			"[SEG_MOVETO [10.0, 10.0], SEG_ARCTO [5.0, 5.0, 30.0, 1.0, 0.0, 20.0, 20.0], " +
				"SEG_ARCTO [2.0, 3.0, 0.0, 0.0, 0.0, 16.0, 16.0]]"},
		{"closed paths", "M2 2 L5 5 Z L7 7 z", "M 2 2 L 5 5 Z L 7 7 z", "M 2 2 |L 5 5 |Z |L 7 7 |Z ",
			"[SEG_MOVETO [2.0, 2.0], SEG_LINETO [5.0, 5.0], SEG_LINETO [7.0, 7.0]]"},
		{"relative move after close", "M10 10 Z m5 5 l1 1", "M 10 10 Z m 5 5 l 1 1", "M 10 10 |Z |M 15 15 |L 16 16 ",
			"[SEG_MOVETO [10.0, 10.0], SEG_MOVETO [15.0, 15.0], SEG_LINETO [16.0, 16.0]]"},
		{"compact numbers", "M1e1-5.5.5 1.5.5-.5", "M 10 -5.5 L 0.5 1.5 L 0.5 -0.5", "M 10 -5.5 |L 0.5 1.5 |L 0.5 -0.5 ",
			"[SEG_MOVETO [10.0, -5.5], SEG_LINETO [0.5, 1.5], SEG_LINETO [0.5, -0.5]]"},
		{"moves only", "M1 1 M2 2", "M 1 1 M 2 2", "M 1 1 |M 2 2 ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSvgPath(tt.data, UTranslateNone())
			if got := svgPathCommands(p); got != tt.commands {
				t.Errorf("commands = %q, want %q", got, tt.commands)
			}
			if got := svgPathMovements(p); got != tt.movements {
				t.Errorf("movements = %q, want %q", got, tt.movements)
			}
			if got := svgPathDrawn(p, NewXAffineTransformIdentity()); got != tt.path {
				t.Errorf("path = %s, want %s", got, tt.path)
			}
		})
	}
}

func TestSvgPathTransforms(t *testing.T) {
	p := NewSvgPath("m10 20 l30 40 l-5 -5", UTranslateNone())
	if got, want := svgPathDrawn(p, XAffineTransformScaleInstance(2, 2)),
		"[SEG_MOVETO [20.0, 40.0], SEG_LINETO [80.0, 120.0], SEG_LINETO [70.0, 110.0]]"; got != want {
		t.Errorf("scaled path = %s, want %s", got, want)
	}
	at := XAffineTransformTranslateInstance(3, -2)
	pi := math.Pi
	at.Rotate(pi/6, 5, 5)
	if got, want := svgPathDrawn(p, at), "[SEG_MOVETO [4.8301270189221945, 18.49038105676658], "+
		"SEG_LINETO [10.810889132455358, 68.13139720814412], SEG_LINETO [8.980762113533164, 61.30127018922193]]"; got != want {
		t.Errorf("rotated path = %s, want %s", got, want)
	}
	translated := NewSvgPath("m10 20 l30 40 l-5 -5", NewUTranslate(3, 4))
	if got, want := svgPathDrawn(translated, XAffineTransformScaleInstance(2, 3)),
		"[SEG_MOVETO [26.0, 72.0], SEG_LINETO [86.0, 192.0], SEG_LINETO [76.0, 177.0]]"; got != want {
		t.Errorf("translated path = %s, want %s", got, want)
	}
	var log []string
	p.DrawMeFactor(newSpriteRec(&log, false, color.IdentityMapper), 1.5)
	if want := "path null [SEG_MOVETO [15.0, 30.0], SEG_LINETO [60.0, 90.0], SEG_LINETO [52.5, 82.5]]"; len(log) != 1 ||
		!strings.HasSuffix(log[0], want) {
		t.Errorf("DrawMeFactor = %q, want %q", log, want)
	}
}

func TestSvgPathErrors(t *testing.T) {
	tests := []struct{ data, want string }{
		{"1 2 L 3 4", "ERR NullPointerException"},
		{"M 10", "ERR NoSuchElementException"},
		{"M 10 10 L z", "ERR ClassCastException"},
		{"M 1 1 X 2", "ERR UnsupportedOperationException: X"},
		{"M0 0 z 5 5", "ERR ClassCastException"},
	}
	for _, tt := range tests {
		if got := spriteTry(func() { NewSvgPath(tt.data, UTranslateNone()) }); got != tt.want {
			t.Errorf("NewSvgPath(%q) = %q, want %q", tt.data, got, tt.want)
		}
	}
	p := NewSvgPath("M0 0 L5 5 T 7 7", UTranslateNone())
	if got, want := spriteTry(func() { svgPathDrawn(p, NewXAffineTransformIdentity()) }),
		"ERR IllegalArgumentException: L"; got != want {
		t.Errorf("T after L = %q, want %q", got, want)
	}
}

func TestGoldenSvgPath(t *testing.T) {
	var out []string
	for n, line := range spriteInputs(t, "svgpath_inputs.txt") {
		data := spriteUnesc(line)
		p := itoa(n + 1)
		out = append(out, p+"\tinput\t"+esc(data))
		var path *SvgPath
		if e := spriteTry(func() { path = NewSvgPath(data, UTranslateNone()) }); e != "" {
			out = append(out, p+"\t"+e)
			continue
		}
		var cmds []string
		for _, c := range path.commands {
			cmds = append(cmds, c.ToSvg())
		}
		out = append(out, p+"\tcommands\t"+strings.Join(cmds, " "))
		var moves []string
		for _, m := range path.movements {
			moves = append(moves, m.ToSvg())
		}
		out = append(out, p+"\tmovements\t"+strings.Join(moves, "|"))
		out = append(out, p+"\ttoSvg\t"+path.ToSvg())
		identity := color.IdentityMapper
		spriteDrawing(&out, p+"\tid", false, identity, func(ug *spriteRec) {
			path.DrawMe(ug, NewXAffineTransformIdentity())
		})
		spriteDrawing(&out, p+"\ts2", false, identity, func(ug *spriteRec) {
			path.DrawMe(ug, XAffineTransformScaleInstance(2, 2))
		})
		spriteDrawing(&out, p+"\trot", false, identity, func(ug *spriteRec) {
			at := XAffineTransformTranslateInstance(3, -2)
			pi := math.Pi
			at.Rotate(pi/6, 5, 5)
			path.DrawMe(ug, at)
		})
		spriteDrawing(&out, p+"\tmat", false, identity, func(ug *spriteRec) {
			at := XAffineTransformScaleInstance(1.5, 1.5)
			at.Concatenate(NewXAffineTransformFlat([]float64{1, 0.5, -0.5, 1, 10, 20}))
			path.DrawMe(ug, at)
		})
		spriteDrawing(&out, p+"\tf15", false, identity, func(ug *spriteRec) { path.DrawMeFactor(ug, 1.5) })
		spriteDrawing(&out, p+"\ttr", false, identity, func(ug *spriteRec) {
			NewSvgPath(data, NewUTranslate(3, 4)).DrawMe(ug, XAffineTransformScaleInstance(2, 3))
		})
	}
	compareGolden(t, "svgpath.golden", out)
}
