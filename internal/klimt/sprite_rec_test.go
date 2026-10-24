// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// The recorder and helpers of the SVG sprite tests, which compare with the
// golden files written by testdata/javagen/SpriteGolden.java
// (svgpath.golden, emoji.golden, svgsprite.golden).

// spriteRec is the harness's Rec: a UGraphic that records what is drawn,
// matching the "SVG" property when svg is true.
type spriteRec struct {
	log    *[]string
	t      UTranslate
	color  color.HColor
	back   color.HColor
	stroke UStroke
	svg    bool
	mapper *color.ColorMapper
}

var _ UGraphic = (*spriteRec)(nil)

func newSpriteRec(log *[]string, svg bool, mapper *color.ColorMapper) *spriteRec {
	return &spriteRec{log: log, color: color.Black, back: color.None(), stroke: UStrokeSimple(), svg: svg,
		mapper: mapper}
}

func (r *spriteRec) StringBounder() font.StringBounder { return fakeBounder{} }

type spriteRecParam struct{ r *spriteRec }

func (p spriteRecParam) Color() color.HColor     { return p.r.color }
func (p spriteRecParam) Backcolor() color.HColor { return p.r.back }
func (p spriteRecParam) Stroke() UStroke         { return p.r.stroke }
func (p spriteRecParam) IsHidden() bool          { return false }
func (p spriteRecParam) Pattern() UPattern       { return UPatternFull }

func (r *spriteRec) Param() UParam { return spriteRecParam{r} }

func (r *spriteRec) Draw(shape UShape) {
	*r.log = append(*r.log, "@"+d2s(r.t.Dx())+","+d2s(r.t.Dy())+" c="+spriteCol(r.color)+" b="+
		spriteCol(r.back)+" s="+r.stroke.String()+" "+describeSpriteShape(shape))
}

func (r *spriteRec) Apply(change UChange) UGraphic {
	c := *r
	switch ch := change.(type) {
	case UTranslate:
		c.t = ch.Compose(r.t)
	case color.HColor:
		c.color = ch
	case UBackground:
		c.back = ch.BackColor()
	case UStroke:
		c.stroke = ch
	default:
		*r.log = append(*r.log, "apply ?")
		return r
	}
	return &c
}

func (r *spriteRec) ColorMapper() *color.ColorMapper { return r.mapper }
func (r *spriteRec) StartURL(url *URL)               { *r.log = append(*r.log, "startUrl "+url.URL()) }
func (r *spriteRec) CloseURL()                       { *r.log = append(*r.log, "closeUrl") }
func (r *spriteRec) StartGroup(group *UGroup) {
	*r.log = append(*r.log, "startGroup "+groupString(group))
}
func (r *spriteRec) CloseGroup()                                { *r.log = append(*r.log, "closeGroup") }
func (r *spriteRec) FlushUg()                                   {}
func (r *spriteRec) MatchesProperty(name string) bool           { return r.svg && name == "SVG" }
func (r *spriteRec) DefaultBackground() color.HColor            { return color.White }
func (r *spriteRec) WriteToStream(io.Writer, string, int) error { return nil }

// spriteCol is the harness's col: linear gradients with their stops, other
// colors as AsString.
func spriteCol(c color.HColor) string {
	if c == nil {
		return "null"
	}
	if g, ok := c.(*color.HColorLinearGradient); ok {
		var stops []string
		for _, s := range g.Stops() {
			stops = append(stops, d2s(s.Offset())+":"+spriteCol(s.Color())+":"+d2s(s.Opacity()))
		}
		return "lg(" + d2s(g.X1()) + "," + d2s(g.Y1()) + "," + d2s(g.X2()) + "," + d2s(g.Y2()) + "," +
			boolStr(g.IsUserSpaceOnUse()) + "," + g.SpreadMethod().String() + "," + javaList(stops) + ")"
	}
	return c.AsString()
}

func spriteStyles(fc *FontConfiguration) string {
	var l []string
	for s := font.FontStylePlain; s <= font.FontStyleBackcolor; s++ {
		if fc.ContainsStyle(s) {
			l = append(l, s.String())
		}
	}
	return javaList(l)
}

func javaHex(v int32) string { return strconv.FormatUint(uint64(uint32(v)), 16) }

// spritePix is the harness's pix: the pixels of a small image, a hash of
// the others.
func spritePix(img PortableImage) string {
	prefix := itoa(img.Width()) + "x" + itoa(img.Height()) + ":"
	if img.Width()*img.Height() <= 16 {
		var l []string
		for y := 0; y < img.Height(); y++ {
			for x := 0; x < img.Width(); x++ {
				l = append(l, javaHex(img.RGB(x, y)))
			}
		}
		return prefix + strings.Join(l, ",")
	}
	var h int32
	for y := 0; y < img.Height(); y++ {
		for x := 0; x < img.Width(); x++ {
			h = 31*h + img.RGB(x, y)
		}
	}
	return prefix + "#" + javaHex(h)
}

func describeSpriteShape(shape UShape) string {
	switch s := shape.(type) {
	case *UPath:
		comment := s.Comment()
		if comment == "" {
			comment = "null"
		}
		return "path " + comment + " " + segmentsString(s)
	case *UEllipse:
		return "ellipse " + d2s(s.Width()) + "," + d2s(s.Height()) + " " + d2s(s.Start()) + "," + d2s(s.Extend())
	case *URectangle:
		return "rect " + d2s(s.Width()) + "," + d2s(s.Height()) + " r=" + d2s(s.Rx()) + "," + d2s(s.Ry())
	case *UText:
		fc := s.FontConfiguration()
		return "text '" + esc(s.Text()) + "' " + uf(fc.Font()) + " " + spriteCol(fc.Color()) + " " + spriteStyles(fc)
	case *UImage:
		return "image " + itoa(s.Width()) + "," + itoa(s.Height()) + " " + spritePix(s.Image(1))
	case *UImageSvg:
		return "imagesvg " + d2s(s.Width()) + "," + d2s(s.Height()) + " " + esc(s.Svg(true))
	}
	return "?"
}

// spriteErr is the harness's err.
func spriteErr(re *color.RuntimeError) string {
	name := re.Class[strings.LastIndexByte(re.Class, '.')+1:]
	if name == "NullPointerException" || name == "ClassCastException" || re.Message == "" {
		return "ERR " + name
	}
	return "ERR " + name + ": " + esc(re.Message)
}

// spriteTry runs f; it returns the harness's line of the exception f
// throws, "" when it returns.
func spriteTry(f func()) (result string) {
	defer func() {
		if r := recover(); r != nil {
			re, ok := r.(*color.RuntimeError)
			if !ok {
				panic(r)
			}
			result = spriteErr(re)
		}
	}()
	f()
	return ""
}

// spriteDrawing is the harness's drawing: the lines of a drawing on a
// recorder, then of an exception it throws.
func spriteDrawing(out *[]string, prefix string, svg bool, mapper *color.ColorMapper, d func(ug *spriteRec)) {
	var log []string
	if e := spriteTry(func() { d(newSpriteRec(&log, svg, mapper)) }); e != "" {
		log = append(log, e)
	}
	for _, l := range log {
		*out = append(*out, prefix+"\t"+l)
	}
}

func spriteInputs(t *testing.T, name string) []string {
	t.Helper()
	var result []string
	for _, line := range readLines(t, name) {
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result
}

// spriteUnesc undoes the \n, \t and \\ escapes of the input files.
func spriteUnesc(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			default:
				sb.WriteByte(s[i])
			}
			continue
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// spriteColor is the harness's color: nil for "null".
func spriteColor(s string) color.HColor {
	if s == "null" {
		return nil
	}
	return color.Instance().GetColorOrWhite(s)
}
