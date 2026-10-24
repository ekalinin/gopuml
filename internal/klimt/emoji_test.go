// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The expected values are the output of testdata/javagen/SpriteGolden.java
// (emoji.golden), which TestGoldenEmoji compares in full.

func TestColorResolver(t *testing.T) {
	tests := []struct {
		font, forced string
		min, max     int
		code, want   string
	}{
		{"null", "null", 0, 255, "", "#000000"},
		{"#3366CC", "null", 0, 255, "", "#3366CC"},
		{"null", "#808080", 0, 255, "", "#808080"},
		{"null", "null", 0, 255, "red", "#FF0000"},
		{"null", "null", 0, 255, "#123456", "#123456"},
		{"null", "null", 0, 255, "123456", "#123456"},
		{"null", "null", 0, 255, "currentColor", "#FFFFFF"},
		{"null", "null", 0, 255, "none", "transparent"},
		{"null", "null", 0, 255, "NONE", "transparent"},
		{"null", "null", 0, 255, "foo", "#FFFFFF"},
		{"null", "null", 0, 255, "red-blue", "ERR ClassCastException"},
		{"null", "#808080", 0, 255, "red", "#4C4C4C"},
		{"null", "#808080", 0, 255, "#123456", "#2D2D2D"},
		{"null", "#FF8000", 0, 255, "red", "#FEA882"},
		{"null", "#FF8000", 0, 255, "#123456", "#FE9863"},
		{"null", "#FF8000", 0, 255, "currentColor", "#FEFEFE"},
		{"null", "#FF8000", 50, 200, "red", "#FE8D49"},
		{"null", "#FF8000", 50, 200, "#123456", "#FB7E00"},
		{"null", "#FF8000", 50, 200, "none", "transparent"},
	}
	for _, tt := range tests {
		cr := NewColorResolver(spriteColor(tt.font), spriteColor(tt.forced), spriteRange{tt.min, tt.max})
		var got string
		if tt.code == "" {
			got = spriteCol(cr.DefaultColor())
		} else if e := spriteTry(func() { got = spriteCol(cr.TrueColor(tt.code)) }); e != "" {
			got = e
		}
		if got != tt.want {
			t.Errorf("font %s, forced %s, range %d-%d, %q: got %s, want %s", tt.font, tt.forced, tt.min, tt.max,
				tt.code, got, tt.want)
		}
	}
	for _, bad := range []string{"red-blue", "?red:blue", "automatic"} {
		if got := spriteTry(func() { NewColorResolver(spriteColor(bad), nil, spriteRange{0, 255}) }); got !=
			"ERR ClassCastException" {
			t.Errorf("font color %s: got %q, want a ClassCastException", bad, got)
		}
	}
}

func TestUGraphicWithScale(t *testing.T) {
	var log []string
	cr := NewColorResolver(spriteColor("#3366CC"), nil, spriteRange{0, 255})
	u := NewUGraphicWithScale(newSpriteRec(&log, false, color.IdentityMapper), cr, 1.5)
	check := func(name string, u *UGraphicWithScale, want string) {
		t.Helper()
		got := u.AffineTransform().String() + " angle=" + d2s(u.Angle()) + " scale=" + d2s(u.InitialScale())
		if got != want {
			t.Errorf("%s:\n got %s\nwant %s", name, got, want)
		}
	}
	check("init", u, "XAffineTransform[[1.5, 0.0, 0.0], [0.0, 1.5, 0.0]] angle=0.0 scale=1.5")
	u = u.ApplyTranslate(3, -4)
	check("translate", u, "XAffineTransform[[1.5, 0.0, 4.5], [0.0, 1.5, -6.0]] angle=0.0 scale=1.5")
	u = u.ApplyScale(2, 2)
	check("scale", u, "XAffineTransform[[3.0, 0.0, 9.0], [0.0, 3.0, -12.0]] angle=0.0 scale=2.0")
	u = u.ApplyRotate(30, 5, 6)
	check("rotate", u, "XAffineTransform[[2.598076211353316, -1.4999999999999998, 20.00961894323342], "+
		"[1.4999999999999998, 2.598076211353316, -17.088457268119896]] angle=30.0 scale=2.0")
	u = u.ApplyMatrix(1, 0.5, -0.5, 1, 10, 20)
	check("matrix", u, "XAffineTransform[[1.848076211353316, -2.799038105676658, 15.990381056766584], "+
		"[2.799038105676658, 1.848076211353316, 49.873066958946424]] angle=30.0 scale=2.0")
	if got := spriteTry(func() { u.ApplyScale(2, 3) }); got != "ERR IllegalArgumentException" {
		t.Errorf("ApplyScale(2, 3) = %q, want an IllegalArgumentException", got)
	}

	u.Draw(UEllipseBuild(3, 4))
	if want := "@0.0,0.0 c=#3366CC b=#3366CC s=0.0-0.0-1.0 ellipse 3.0,4.0 0.0,0.0"; len(log) != 1 || log[0] != want {
		t.Errorf("drawing = %q, want %q", log, want)
	}
}

// spriteRange is a GrayLevelRange of fixed levels.
type spriteRange struct{ min, max int }

func (r spriteRange) MinGrayLevel() int { return r.min }
func (r spriteRange) MaxGrayLevel() int { return r.max }

func ugsStep(out *[]string, name string, u *UGraphicWithScale) {
	*out = append(*out, "ugs\t"+name+"\t"+u.AffineTransform().String()+" angle="+d2s(u.Angle())+" scale="+
		d2s(u.InitialScale()))
}

func TestGoldenEmoji(t *testing.T) {
	var out []string
	fonts := []string{"null", "#3366CC"}
	forceds := []string{"null", "#808080", "#FF8000", "#0000FF"}
	ranges := [][2]int{{0, 255}, {50, 200}, {999, -1}}
	codes := []string{"red", "#123456", "123456", "#FFF", "currentColor", "none", "NONE", "transparent", "foo",
		"#80FF8000", "black", "white", "#FF0000", "red-blue", "automatic", "?red:blue"}
	for _, fnt := range fonts {
		for _, forced := range forceds {
			for _, r := range ranges {
				p := "resolver\t" + fnt + "\t" + forced + "\t" + itoa(r[0]) + "," + itoa(r[1])
				cr := NewColorResolver(spriteColor(fnt), spriteColor(forced), spriteRange{r[0], r[1]})
				out = append(out, p+"\tdefault\t"+spriteCol(cr.DefaultColor()))
				for _, code := range codes {
					var v string
					if e := spriteTry(func() { v = spriteCol(cr.TrueColor(code)) }); e != "" {
						v = e
					}
					out = append(out, p+"\t"+code+"\t"+v)
				}
			}
		}
	}
	for _, bad := range []string{"red-blue", "?red:blue", "automatic"} {
		var v string
		if e := spriteTry(func() {
			v = spriteCol(NewColorResolver(spriteColor(bad), nil, spriteRange{0, 255}).DefaultColor())
		}); e != "" {
			v = e
		}
		out = append(out, "resolver-font\t"+bad+"\t"+v)
		if e := spriteTry(func() {
			v = spriteCol(NewColorResolver(nil, spriteColor(bad), spriteRange{0, 255}).DefaultColor())
		}); e != "" {
			v = e
		}
		out = append(out, "resolver-forced\t"+bad+"\t"+v)
	}

	cr := NewColorResolver(spriteColor("#3366CC"), nil, spriteRange{0, 255})
	identity := color.IdentityMapper
	spriteDrawing(&out, "ugs\tdraw", false, identity, func(ug *spriteRec) {
		u := NewUGraphicWithScale(ug, cr, 1.5)
		ugsStep(&out, "init", u)
		u = u.ApplyTranslate(3, -4)
		ugsStep(&out, "translate", u)
		u = u.ApplyScale(2, 2)
		ugsStep(&out, "scale", u)
		u = u.ApplyRotate(30, 5, 6)
		ugsStep(&out, "rotate", u)
		u = u.ApplyMatrix(1, 0.5, -0.5, 1, 10, 20)
		ugsStep(&out, "matrix", u)
		u = u.ApplyRotate(-45, 0, 0)
		ugsStep(&out, "rotate0", u)
		u = u.ApplyScale(0.5, 0.5)
		ugsStep(&out, "scale-half", u)
		if e := spriteTry(func() { u.ApplyScale(2, 3) }); e != "" {
			out = append(out, "ugs\tscale-xy\t"+e)
		} else {
			out = append(out, "ugs\tscale-xy\tok")
		}
		out = append(out, "ugs\ttruecolor\t"+spriteCol(u.TrueColor("red"))+" "+spriteCol(u.DefaultColor()))
		u.Draw(UEllipseBuild(3, 4))
		u.Apply(NewUTranslate(1, 2)).Apply(UStrokeWithThickness(2)).Draw(URectangleBuild(5, 6))
		u.Apply(color.Red).Apply(Bg(color.Blue)).Draw(URectangleBuild(1, 1))
	})
	forced := NewColorResolver(spriteColor("#3366CC"), spriteColor("#FF8000"), spriteRange{10, 100})
	spriteDrawing(&out, "ugs\tforced", false, identity, func(ug *spriteRec) {
		u := NewUGraphicWithScale(ug, forced, 2)
		ugsStep(&out, "forced-init", u)
		out = append(out, "ugs\tforced-truecolor\t"+spriteCol(u.TrueColor("red"))+" "+
			spriteCol(u.TrueColor("#FFFFFF")))
		u.Draw(UEllipseBuild(1, 1))
	})
	none := NewColorResolver(nil, nil, spriteRange{0, 255})
	spriteDrawing(&out, "ugs\tnone", false, identity, func(ug *spriteRec) {
		NewUGraphicWithScale(ug, none, 1).Draw(UEllipseBuild(1, 1))
	})
	compareGolden(t, "emoji.golden", out)
}
