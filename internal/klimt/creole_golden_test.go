// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The golden files compared here are written by
// testdata/javagen/CreoleGolden.java in the reference image.

// creoleRec is the harness's Rec: a UGraphic that records what is drawn,
// measuring with a given StringBounder.
type creoleRec struct {
	log     *[]string
	bounder font.StringBounder
	t       UTranslate
	color   color.HColor
	back    color.HColor
	stroke  UStroke
}

var _ UGraphic = (*creoleRec)(nil)

func newCreoleRec(log *[]string, bounder font.StringBounder) *creoleRec {
	return &creoleRec{log: log, bounder: bounder, color: color.Black, back: color.None(), stroke: UStrokeSimple()}
}

func (r *creoleRec) StringBounder() font.StringBounder { return r.bounder }

type creoleRecParam struct{ r *creoleRec }

func (p creoleRecParam) Color() color.HColor     { return p.r.color }
func (p creoleRecParam) Backcolor() color.HColor { return p.r.back }
func (p creoleRecParam) Stroke() UStroke         { return p.r.stroke }
func (p creoleRecParam) IsHidden() bool          { return false }
func (p creoleRecParam) Pattern() UPattern       { return UPatternFull }

func (r *creoleRec) Param() UParam { return creoleRecParam{r} }

func (r *creoleRec) Draw(shape UShape) {
	*r.log = append(*r.log, "@"+d2s(r.t.Dx())+","+d2s(r.t.Dy())+" c="+hc(r.color)+" b="+hc(r.back)+" s="+
		r.stroke.String()+" "+describeCreoleShape(shape))
}

func (r *creoleRec) Apply(change UChange) UGraphic {
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

func (r *creoleRec) ColorMapper() *color.ColorMapper { return color.IdentityMapper }
func (r *creoleRec) StartURL(url *URL) {
	*r.log = append(*r.log, "startUrl "+url.URL()+" tooltip="+url.Tooltip()+" label="+url.Label())
}
func (r *creoleRec) CloseURL() { *r.log = append(*r.log, "closeUrl") }
func (r *creoleRec) StartGroup(group *UGroup) {
	*r.log = append(*r.log, "startGroup "+groupString(group))
}
func (r *creoleRec) CloseGroup()                                { *r.log = append(*r.log, "closeGroup") }
func (r *creoleRec) FlushUg()                                   {}
func (r *creoleRec) MatchesProperty(string) bool                { return false }
func (r *creoleRec) DefaultBackground() color.HColor            { return color.White }
func (r *creoleRec) WriteToStream(io.Writer, string, int) error { return nil }

// creoleFCString is the harness's fcString: font, color, styles, extended color,
// position shift and SVG attributes.
func creoleFCString(fc *FontConfiguration) string {
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

func describeCreoleShape(shape UShape) string {
	switch s := shape.(type) {
	case *UText:
		return "text '" + s.Text() + "' " + creoleFCString(s.FontConfiguration())
	case *UImage:
		return "image " + strconv.Itoa(s.Width()) + "x" + strconv.Itoa(s.Height()) + " scale=" + d2s(s.ImageScale())
	case *UImageSvg:
		return "svg " + d2s(s.Width()) + "x" + d2s(s.Height())
	}
	return describeShape(shape)
}

// recoverJava records a panic as the harness records an exception.
func recoverJava(log *[]string) {
	switch r := recover().(type) {
	case nil:
	case *color.RuntimeError:
		*log = append(*log, "ERR "+r.Class[strings.LastIndexByte(r.Class, '.')+1:])
	case *color.NoSuchColorRuntimeError:
		*log = append(*log, "ERR NoSuchColorRuntimeException")
	default:
		panic(r)
	}
}

// creoleCase is a line of creole_inputs.txt.
type creoleCase struct {
	name     string
	size     int
	align    geom.HorizontalAlignment
	max      string
	hasMax   bool
	mode     CreoleMode
	nl       bool
	bordered bool
	text     string
}

func parseCreoleCase(t *testing.T, line string) creoleCase {
	parts := strings.Split(line, "\t")
	c := creoleCase{name: parts[0], size: 14, align: geom.HorizontalAlignmentLeft, mode: CreoleModeFull,
		text: strings.ReplaceAll(parts[2], "⇥", "\t")}
	for _, opt := range strings.Split(parts[1], ",") {
		if opt == "" {
			continue
		}
		k, v, _ := strings.Cut(opt, "=")
		switch k {
		case "size":
			c.size, _ = strconv.Atoi(v)
		case "align":
			if v == "NULL" {
				c.align = HorizontalAlignmentNull
			} else {
				c.align, _ = geom.HorizontalAlignmentValueOf(v)
			}
		case "max":
			c.max, c.hasMax = v, true
		case "mode":
			for m := CreoleModeFull; m <= CreoleModeFullButUnderscore; m++ {
				if m.String() == v {
					c.mode = m
				}
			}
		case "nl":
			c.nl = true
		case "bordered":
			c.bordered = true
		default:
			t.Fatalf("%s: unknown option %s", c.name, opt)
		}
	}
	return c
}

func (c creoleCase) display() *Display {
	if c.nl {
		return DisplayWithNewlines(PragmaCreateEmpty(), c.text)
	}
	d := DisplayEmpty()
	for _, line := range strings.Split(c.text, "⏎") {
		if label, ok := strings.CutPrefix(line, "STEREO:"); ok {
			d = d.Add(StereotypeBuild(label))
		} else {
			d = d.Add(Text(line))
		}
	}
	return d
}

func creoleFC(size int) *FontConfiguration {
	return NewFontConfiguration(font.SansSerif(size), color.Black, color.Blue, strokePtr(UStrokeSimple()))
}

// creoleBounders are the harness's bounders: the fixed one and the SVG one
// of the reference fonts.
var creoleBounders = []struct {
	name    string
	bounder font.StringBounder
}{{"fake", fakeBounder{}}, {"svg", font.NewStringBounderSvg(nil)}}

func drawCreoleBlock(out *[]string, name, bounderName string, bounder font.StringBounder, tb func() TextBlock) {
	var log []string
	func() {
		defer recoverJava(&log)
		block := tb()
		dim := block.CalculateDimension(bounder)
		log = append(log, "dim "+d2s(dim.Width())+"x"+d2s(dim.Height()))
		block.DrawU(newCreoleRec(&log, bounder))
	}()
	for _, l := range log {
		*out = append(*out, name+"\t"+bounderName+"\t"+l)
	}
}

// creoleDifferences are the cases where this port differs from upstream:
// fixes of upstream failures (design D16), the literal text the styling
// spec asks for, and a stub of another step. The golden lines of these
// cases are skipped and TestCreoleDifferences checks the behavior.
var creoleDifferences = map[string]string{
	"listnolinemax": "upstream throws NullPointerException for a list item without text and a maximum width",
	"imgnotpng":     "upstream throws NullPointerException when ImageIO cannot read the image data",
	"unicodebad":    "upstream throws IllegalArgumentException for a code point above U+10FFFF",
	// The specification asks for the literal text of icons, emoji and
	// sprites until they are ported; upstream draws them (or nothing for an
	// unknown sprite).
	"icon1":   "OpenIconic icons render as literal text",
	"icon2":   "OpenIconic icons render as literal text",
	"emoji1":  "emoji render as literal text",
	"sprite1": "unknown sprites render as literal text",
	"url5":    "icons, emoji and sprites in link labels render as literal text",
}

func creoleGoldenLines(t *testing.T) []string {
	var out []string
	for _, line := range readLines(t, "creole_inputs.txt") {
		c := parseCreoleCase(t, line)
		for _, b := range creoleBounders {
			drawCreoleBlock(&out, c.name, b.name, b.bounder, func() TextBlock {
				lineBreak := NewLineBreakStrategyOpt(c.max, c.hasMax)
				tb := c.display().Create0(creoleFC(c.size), c.align, newTestSkinParam(), lineBreak, c.mode, nil, nil)
				if c.bordered {
					tb = TextBlockUtilsBordered(tb, strokePtr(UStrokeWithThickness(1.5)), color.Red, color.MyYellow, 5,
						ClockwiseTopRightBottomLeftOf(2, 3, 4, 5), "id")
				}
				return tb
			})
		}
	}
	builder := mustLoadSkin(t, "plantuml.skin")
	note := StyleSignatureBasicOf(SNameRoot, SNameElement, SNameSequenceDiagram, SNameNote).MergedStyle(builder)
	title := StyleSignatureBasicOf(SNameRoot, SNameDocument, SNameTitle).MergedStyle(builder)
	noteText := DisplayCreate("**Note** with //style//", "----", "second line")
	titleText := DisplayCreate("The <color:red>title</color>", "line 2")
	for _, b := range creoleBounders {
		drawCreoleBlock(&out, "stylebordered", b.name, b.bounder, func() TextBlock {
			return note.CreateTextBlockBordered(noteText, color.Instance(), newTestSkinParam(), "nid",
				LineBreakStrategyNone)
		})
		drawCreoleBlock(&out, "title", b.name, b.bounder, func() TextBlock {
			return NewTextBlockTitle(title.FontConfiguration(color.Instance(), nil), titleText, newTestSkinParam())
		})
	}
	return out
}

func withoutCreoleDifferences(lines []string) []string {
	return slices.DeleteFunc(lines, func(l string) bool {
		name, _, _ := strings.Cut(l, "\t")
		_, skip := creoleDifferences[name]
		return skip
	})
}

func TestGoldenCreole(t *testing.T) {
	got := withoutCreoleDifferences(creoleGoldenLines(t))
	want := withoutCreoleDifferences(readLines(t, "creole.golden"))
	if len(got) != len(want) {
		t.Errorf("creole.golden: %d lines, want %d", len(got), len(want))
	}
	failures := 0
	for i := 0; i < min(len(got), len(want)); i++ {
		if got[i] != want[i] {
			failures++
			if failures <= 15 {
				t.Errorf("creole.golden line %d:\n got: %s\nwant: %s", i+1, got[i], want[i])
			}
		}
	}
	if failures > 15 {
		t.Errorf("creole.golden: %d more differences", failures-15)
	}
}

// TestCreoleDifferences checks the cases of creoleDifferences: what this
// port draws instead of upstream's exception or icon.
func TestCreoleDifferences(t *testing.T) {
	byName := map[string][]string{}
	for _, l := range creoleGoldenLines(t) {
		parts := strings.SplitN(l, "\t", 3)
		if parts[1] == "fake" {
			byName[parts[0]] = append(byName[parts[0]], parts[2])
		}
	}
	for name := range creoleDifferences {
		if _, ok := byName[name]; !ok {
			t.Errorf("%s: no such case", name)
		}
	}
	texts := func(name string) []string {
		var result []string
		for _, l := range byName[name] {
			if _, rest, ok := strings.Cut(l, " text '"); ok {
				result = append(result, rest[:strings.Index(rest, "' ")])
			}
		}
		return result
	}
	contains := func(name, want string) {
		t.Helper()
		if !slices.ContainsFunc(byName[name], func(l string) bool { return strings.Contains(l, want) }) {
			t.Errorf("%s: no %q in %q", name, want, byName[name])
		}
	}
	check := func(name string, want ...string) {
		t.Helper()
		if got := texts(name); !slices.Equal(got, want) {
			t.Errorf("%s: texts %q, want %q", name, got, want)
		}
	}
	// A list item without text: the bullet alone.
	contains("listnolinemax", "ellipse 5.0,5.0")
	check("listnolinemax")
	check("imgnotpng", "(Cannot decode: data:image/png;base64,aGVsbG8=)")
	check("unicodebad", "&#1114112;")
	check("icon1", "<&check>", " done")
	check("icon2", "<&check{scale=2}>", " ", "<#red&star>", " ", "<&x,color=blue>")
	check("emoji1", "<:smile:>", " emoji ", "<#red:smile:>")
	check("sprite1", "<$sprite>", " sprite ", "<$foo{scale=2}>", " ", "<#blue$foo>")
	check("url5", "<&check>", " label", " ", "<:smile:>", " x", " ", "<$foo>", " s")
	contains("url5", "startUrl http://a.com tooltip=http://a.com label=<&check> label")
}
