// SPDX-License-Identifier: MIT
// Provenance: original

// Package svgtest runs the drawing scripts of the SVG golden tests of the
// svg and export packages. The same scripts are run with the upstream jar
// by internal/klimt/svg/testdata/javagen/SvgGolden.java, which writes the
// golden files.
//
// A script is a list of cases. Each line holds space separated tokens; in a
// token \s is a space, \t a tab, \n a newline, \\ a backslash, \_ nothing
// and \uXXXX a UTF-16 unit. Lines starting with # are comments.
//
//	case NAME                     starts a case
//	write [METADATA]              writes the SVG (with metadata) and ends it
//	opt KEY VALUE...              an SvgOption setting, before drawing:
//	                              scale, decimal, lengthadjust, preserve,
//	                              hover, dimstyle, mindim W H, back COLOR,
//	                              linktarget, title, svgtitle, svgdesc,
//	                              interactive, rootattr K V, mapper NAME,
//	                              seed, format, insecure, allowjs
//	build                         builds the graphic (else the first drawing
//	                              command does)
//
// Drawing commands change the current graphic or draw on it:
//
//	reset | push | pop            back to the root, save, restore
//	color C | bg C                foreground and background colors
//	stroke T [DASH SPACE]         stroke
//	translate DX DY | clip X Y W H | hidden | shown | fg2bg
//	rect W H [ROUND [SHADOW]]     line DX DY [SHADOW]
//	ellipse W H [START EXTEND [SHADOW]]
//	polygon SHADOW X Y X Y...
//	path SHADOW COMMENT CODELINE  (- for none) then M x y, L x y,
//	                              Q 4 values, C 6 values, A 7 values, Z
//	dotpath X1 Y1 X2 Y2 X3 Y3 X4 Y4 [X2 Y2 X3 Y3 X4 Y4]...
//	pixel | empty W H | comment TEXT
//	centered CHAR SIZE [FAMILY [FACE]]   Monospaced bold by default
//	text FAMILY SIZE FACE STYLES COLOR EXTCOLOR ATTRS ORIENTATION TEXT
//	                              FACE: plain, bold, italic, bolditalic,
//	                              wNNN, iwNNN; STYLES: - or a comma list of
//	                              font styles, thinunderlinestroke,
//	                              nounderlinestroke; EXTCOLOR, ATTRS: - or a
//	                              value
//	image W H                     a raster image with the pixels of Pixel
//	imagesvg SCALE SVG
//	group KEY=VALUE...            UGroupType names | endgroup
//	url URL [TOOLTIP] | endurl
//
// Colors are HColorSet names and values ("#FF0000", "red/blue"...),
// "null", dark(C1,C2), opacity(C) for the WITH_FILL_OPACITY behavior and
// lg(x1,y1,x2,y2,userSpace,SPREAD,offset:color:opacity,...) for a linear
// gradient.
package svgtest

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/plantuml-in-go/internal/code"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/svg"
	"github.com/ekalinin/plantuml-in-go/internal/security"
)

// Case is one test case: its commands between "case" and "write".
type Case struct {
	Name     string
	Lines    [][]string
	Metadata string
}

// Unescape decodes the escapes of a token.
func Unescape(s string) string {
	var units []uint16
	var b strings.Builder
	flush := func() {
		if len(units) > 0 {
			b.WriteString(string(utf16.Decode(units)))
			units = nil
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' || i+1 >= len(s) {
			flush()
			b.WriteByte(c)
			continue
		}
		i++
		if s[i] == 'u' {
			v, _ := strconv.ParseUint(s[i+1:i+5], 16, 16)
			units = append(units, uint16(v))
			i += 4
			continue
		}
		flush()
		switch s[i] {
		case 's':
			b.WriteByte(' ')
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case '_':
		default:
			b.WriteByte(s[i])
		}
	}
	flush()
	return b.String()
}

// ParseFile reads the cases of a script file.
func ParseFile(path string) ([]Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []Case
	var current *Case
	for n, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		tokens := make([]string, len(fields))
		for i, f := range fields {
			tokens[i] = Unescape(f)
		}
		switch tokens[0] {
		case "case":
			current = &Case{Name: tokens[1]}
		case "write":
			if current == nil {
				return nil, fmt.Errorf("line %d: write outside of a case", n+1)
			}
			if len(tokens) > 1 {
				current.Metadata = tokens[1]
			}
			cases = append(cases, *current)
			current = nil
		default:
			if current == nil {
				return nil, fmt.Errorf("line %d: command outside of a case", n+1)
			}
			current.Lines = append(current.Lines, tokens)
		}
	}
	return cases, nil
}

func num(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("svgtest: bad number " + s)
	}
	return v
}

// Color parses a color token; "null" gives nil.
func Color(s string) color.HColor {
	switch {
	case s == "null":
		return nil
	case strings.HasPrefix(s, "dark("):
		p := strings.Split(s[5:len(s)-1], ",")
		return Color(p[0]).WithDark(Color(p[1]))
	case strings.HasPrefix(s, "opacity("):
		return Color(s[8 : len(s)-1]).WithTransparentFillBehavior(color.WithFillOpacity)
	case strings.HasPrefix(s, "lg("):
		p := strings.Split(s[3:len(s)-1], ",")
		var stops []color.Stop
		for _, st := range p[6:] {
			q := strings.Split(st, ":")
			stops = append(stops, color.NewStop(num(q[0]), Color(q[1]), num(q[2])))
		}
		spread := map[string]color.SpreadMethod{"PAD": color.SpreadPad, "REFLECT": color.SpreadReflect,
			"REPEAT": color.SpreadRepeat}[p[5]]
		return color.NewHColorLinearGradient(num(p[0]), num(p[1]), num(p[2]), num(p[3]), p[4] == "true", spread,
			stops)
	}
	c, err := color.Instance().GetColor(s)
	if err != nil {
		panic("svgtest: bad color " + s)
	}
	return c
}

// Mapper returns a color mapper by name.
func Mapper(s string) *color.ColorMapper {
	switch s {
	case "identity":
		return color.IdentityMapper
	case "dark":
		return color.DarkModeMapper
	case "mono":
		return color.MonochromeMapper
	case "monoreverse":
		return color.MonochromeReverseMapper
	case "reverse":
		return color.ReverseMapper(color.OrderRGB)
	case "lightness":
		return color.LigthnessInverseMapper
	}
	panic("svgtest: bad mapper " + s)
}

func face(s string) font.UFontFace {
	switch s {
	case "plain":
		return font.NormalFace()
	case "bold":
		return font.BoldFace()
	case "italic":
		return font.ItalicFace()
	case "bolditalic":
		return font.BoldItalicFace()
	}
	if w, ok := strings.CutPrefix(s, "iw"); ok {
		v, _ := strconv.Atoi(w)
		return font.NewUFontFaceCss(font.StyleItalic, v)
	}
	if w, ok := strings.CutPrefix(s, "w"); ok {
		v, _ := strconv.Atoi(w)
		return font.NewUFontFaceCss(font.StyleNormal, v)
	}
	panic("svgtest: bad face " + s)
}

var fontStyles = map[string]font.FontStyle{"plain": font.FontStylePlain, "italic": font.FontStyleItalic,
	"bold": font.FontStyleBold, "underline": font.FontStyleUnderline, "strike": font.FontStyleStrike,
	"wave": font.FontStyleWave, "backcolor": font.FontStyleBackcolor}

func groupType(name string) klimt.UGroupType {
	for k := klimt.UGroupType(0); k < 64; k++ {
		if k.String() == name {
			return k
		}
	}
	panic("svgtest: bad group type " + name)
}

// FileFormat returns a format by upstream name.
func FileFormat(name string) font.FileFormat {
	for f := font.FileFormat(0); f <= font.FormatSVG; f++ {
		if f.String() == name {
			return f
		}
	}
	panic("svgtest: bad format " + name)
}

// Pixel is the ARGB value of the test images at (x, y).
func Pixel(x, y int) int32 {
	a := 0xFF
	if (x+y)%3 == 0 {
		a = 0x80
	}
	return int32(uint32(a)<<24 | uint32((x*37)&0xFF)<<16 | uint32((y*59)&0xFF)<<8 | uint32((x*y*13)&0xFF))
}

func nul(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

func pt(t []string, i int) geom.XPoint2D { return geom.NewXPoint2D(num(t[i]), num(t[i+1])) }

// Drawer runs drawing commands.
type Drawer struct {
	Root  klimt.UGraphic
	stack []klimt.UGraphic
}

// Draw runs one drawing command on ug and returns the current graphic.
func (d *Drawer) Draw(ug klimt.UGraphic, t []string) klimt.UGraphic {
	switch t[0] {
	case "reset":
		return d.Root
	case "push":
		d.stack = append(d.stack, ug)
		return ug
	case "pop":
		top := d.stack[len(d.stack)-1]
		d.stack = d.stack[:len(d.stack)-1]
		return top
	case "color":
		return ug.Apply(Color(t[1]))
	case "bg":
		return ug.Apply(klimt.Bg(Color(t[1])))
	case "stroke":
		if len(t) > 2 {
			return ug.Apply(klimt.NewUStroke(num(t[2]), num(t[3]), num(t[1])))
		}
		return ug.Apply(klimt.UStrokeWithThickness(num(t[1])))
	case "translate":
		return ug.Apply(klimt.NewUTranslate(num(t[1]), num(t[2])))
	case "clip":
		return ug.Apply(klimt.NewUClip(num(t[1]), num(t[2]), num(t[3]), num(t[4])))
	case "hidden":
		return ug.Apply(klimt.UHiddenHidden)
	case "shown":
		return ug.Apply(klimt.UHiddenNormalShown)
	case "fg2bg":
		return ug.Apply(klimt.CopyForegroundColorToBackgroundColor{})
	case "rect":
		r := klimt.URectangleBuild(num(t[1]), num(t[2]))
		if len(t) > 3 {
			r = r.Rounded(num(t[3]))
		}
		if len(t) > 4 {
			r.SetDeltaShadow(num(t[4]))
		}
		ug.Draw(r)
	case "line":
		l := klimt.NewULine(num(t[1]), num(t[2]))
		if len(t) > 3 {
			l.SetDeltaShadow(num(t[3]))
		}
		ug.Draw(l)
	case "ellipse":
		var e *klimt.UEllipse
		if len(t) > 3 {
			e = klimt.NewUEllipse(num(t[1]), num(t[2]), num(t[3]), num(t[4]))
		} else {
			e = klimt.UEllipseBuild(num(t[1]), num(t[2]))
		}
		if len(t) > 5 {
			e.SetDeltaShadow(num(t[5]))
		}
		ug.Draw(e)
	case "polygon":
		p := klimt.NewUPolygon()
		for i := 2; i+1 < len(t); i += 2 {
			p.AddPoint(num(t[i]), num(t[i+1]))
		}
		p.SetDeltaShadow(num(t[1]))
		ug.Draw(p)
	case "path":
		ug.Draw(path(t))
	case "dotpath":
		p := klimt.NewDotPath().AddCurve4(pt(t, 1), pt(t, 3), pt(t, 5), pt(t, 7))
		for i := 9; i+5 < len(t); i += 6 {
			p = p.AddCurve(pt(t, i), pt(t, i+2), pt(t, i+4))
		}
		ug.Draw(p)
	case "pixel":
		ug.Draw(&klimt.UPixel{})
	case "text":
		ug.Draw(text(t))
	case "image":
		w, _ := strconv.Atoi(t[1])
		h, _ := strconv.Atoi(t[2])
		im := klimt.NewPortableImageARGB(w, h)
		for x := range w {
			for y := range h {
				im.SetRGB(x, y, Pixel(x, y))
			}
		}
		ug.Draw(klimt.NewUImage(klimt.NewPixelImage(im, klimt.AffineTransformBilinear)))
	case "imagesvg":
		ug.Draw(klimt.NewUImageSvg(t[2], num(t[1])))
	case "comment":
		ug.Draw(klimt.NewUComment(t[1]))
	case "centered":
		size, _ := strconv.Atoi(t[2])
		family, f := "Monospaced", font.BoldFace()
		if len(t) > 3 {
			family = t[3]
		}
		if len(t) > 4 {
			f = face(t[4])
		}
		ug.Draw(klimt.NewUCenteredCharacter([]rune(t[1])[0], font.Build(family, f, size)))
	case "empty":
		ug.Draw(klimt.NewUEmpty(num(t[1]), num(t[2])))
	case "group":
		g := klimt.NewUGroup()
		for _, kv := range t[1:] {
			k, v, _ := strings.Cut(kv, "=")
			g.Put(groupType(k), v)
		}
		ug.StartGroup(g)
	case "endgroup":
		ug.CloseGroup()
	case "url":
		tooltip, ok := "", false
		if len(t) > 2 {
			tooltip, ok = t[2], true
		}
		ug.StartURL(klimt.NewURL(t[1], tooltip, ok, "", false))
	case "endurl":
		ug.CloseURL()
	default:
		panic("svgtest: bad command " + t[0])
	}
	return ug
}

func path(t []string) *klimt.UPath {
	p := klimt.NewUPath(nul(t[2]), nul(t[3]))
	p.SetDeltaShadow(num(t[1]))
	for i := 4; i < len(t); {
		s := t[i]
		i++
		switch s {
		case "M":
			p.MoveTo(num(t[i]), num(t[i+1]))
			i += 2
		case "L":
			p.LineTo(num(t[i]), num(t[i+1]))
			i += 2
		case "Q":
			p.QuadTo(num(t[i]), num(t[i+1]), num(t[i+2]), num(t[i+3]))
			i += 4
		case "C":
			p.CubicTo(num(t[i]), num(t[i+1]), num(t[i+2]), num(t[i+3]), num(t[i+4]), num(t[i+5]))
			i += 6
		case "A":
			p.ArcTo(num(t[i]), num(t[i+1]), num(t[i+2]), num(t[i+3]), num(t[i+4]), num(t[i+5]), num(t[i+6]))
			i += 7
		case "Z":
			p.ClosePath()
		default:
			panic("svgtest: bad path segment " + s)
		}
	}
	return p
}

func text(t []string) *klimt.UText {
	size, _ := strconv.Atoi(t[2])
	f := font.Build(t[1], face(t[3]), size)
	styles := t[4]
	underline := klimt.UStrokeSimple()
	underlinePtr := &underline
	switch {
	case strings.Contains(styles, "nounderlinestroke"):
		underlinePtr = nil
	case strings.Contains(styles, "thinunderlinestroke"):
		thin := klimt.UStrokeWithThickness(0)
		underlinePtr = &thin
	}
	fc := klimt.NewFontConfiguration(f, Color(t[5]), Color("#0000FF"), underlinePtr)
	if styles != "-" {
		for _, st := range strings.Split(styles, ",") {
			if !strings.HasSuffix(st, "underlinestroke") {
				fc = fc.Add(fontStyles[st])
			}
		}
	}
	if t[6] != "-" {
		fc = fc.ChangeExtendedColor(Color(t[6]))
	}
	if t[7] != "-" {
		fc = fc.ChangeAttributes(klimt.SvgAttributesBuild(t[7]))
	}
	txt := klimt.UTextBuild(t[9], fc)
	if orientation, _ := strconv.Atoi(t[8]); orientation != 0 {
		txt = txt.WithOrientation(orientation)
	}
	return txt
}

// SvgCase holds the settings of an svg package case.
type SvgCase struct {
	Option *svg.SvgOption
	Seed   int64
	Format font.FileFormat
	opts   *klimt.ConfigurationStore[klimt.OptionKey]
}

// NewSvgCase returns the default settings: a white background, seed 42,
// SVG.
func NewSvgCase() *SvgCase {
	return &SvgCase{Option: svg.SvgOptionBasic().WithBackcolor(color.White), Seed: 42, Format: font.FormatSVG}
}

// SetOption applies an "opt" line.
func (s *SvgCase) SetOption(t []string) {
	o := s.Option
	switch t[1] {
	case "scale":
		o.WithScale(num(t[2]))
	case "decimal":
		v, _ := strconv.Atoi(t[2])
		o.WithDecimal(v)
	case "lengthadjust":
		o.WithLengthAdjust(map[string]klimt.LengthAdjust{"NONE": klimt.LengthAdjustNone,
			"SPACING": klimt.LengthAdjustSpacing, "SPACING_AND_GLYPHS": klimt.LengthAdjustSpacingAndGlyphs}[t[2]])
	case "preserve":
		o.WithPreserveAspectRatio(t[2])
	case "hover":
		o.WithHoverPathColorRGB(t[2])
	case "dimstyle":
		o.WithSvgDimensionStyle(t[2] == "true")
	case "mindim":
		o.WithMinDim(geom.NewXDimension2D(num(t[2]), num(t[3])))
	case "back":
		o.WithBackcolor(Color(t[2]))
	case "linktarget":
		o.WithLinkTarget(t[2])
	case "title":
		o.WithTitleText(t[2])
	case "svgtitle", "svgdesc":
		if s.opts == nil {
			s.opts = klimt.NewConfigurationStore[klimt.OptionKey]()
		}
		key := klimt.OptionSvgTitle
		if t[1] == "svgdesc" {
			key = klimt.OptionSvgDesc
		}
		s.opts.Define(key, t[2], true)
	case "interactive":
		o.WithInteractive(t[2])
	case "rootattr":
		o.WithRootAttribute(t[2], t[3])
	case "mapper":
		o.WithColorMapper(Mapper(t[2]))
	case "seed":
		v, _ := strconv.ParseInt(t[2], 10, 64)
		s.Seed = v
	case "format":
		s.Format = FileFormat(t[2])
	case "insecure":
		o.WithSecurityProfile(security.Insecure)
	case "allowjs":
		o.WithAllowJavascriptInLink(true)
	default:
		panic("svgtest: bad option " + t[1])
	}
}

// Build returns the graphic of the settings.
func (s *SvgCase) Build() *svg.UGraphicSvg {
	if s.opts != nil {
		s.Option.WithConfigurationStore(s.opts)
	}
	return svg.Build(s.Option, false, s.Seed, font.NewStringBounderSvg(nil), s.Format)
}

// RunSvg runs a case of the svg package and returns the SVG text.
func RunSvg(c Case) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("case %s: panic: %v", c.Name, r)
		}
	}()
	s := NewSvgCase()
	var root *svg.UGraphicSvg
	var d Drawer
	var ug klimt.UGraphic
	build := func() {
		root = s.Build()
		d.Root = root
		ug = root
	}
	for _, t := range c.Lines {
		switch t[0] {
		case "opt":
			s.SetOption(t)
			continue
		case "build":
			build()
			continue
		}
		if root == nil {
			build()
		}
		ug = d.Draw(ug, t)
	}
	if root == nil {
		build()
	}
	var buf bytes.Buffer
	if err := root.WriteToStream(&buf, c.Metadata, 96); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var (
	pngDataURL  = regexp.MustCompile(`data:image/png;base64,([A-Za-z0-9+/=]*)`)
	metadataPI  = regexp.MustCompile(`<\?plantuml-src ([^?]*)\?>`)
	normalizeKO = "undecodable"
)

// Normalize replaces what may differ between upstream's and this output
// without changing the drawing: the PNG data URLs become a hash of their
// decoded pixels, and the plantuml-src text its decoded source.
func Normalize(svgText string) string {
	svgText = pngDataURL.ReplaceAllStringFunc(svgText, func(m string) string {
		data, err := base64.StdEncoding.DecodeString(pngDataURL.FindStringSubmatch(m)[1])
		if err != nil {
			return normalizeKO
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return normalizeKO
		}
		h := sha256.New()
		b := img.Bounds()
		fmt.Fprintf(h, "%dx%d;", b.Dx(), b.Dy())
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				r, g, bl, a := img.At(x, y).RGBA()
				// Compare non-premultiplied 8-bit channels.
				if a != 0 {
					r, g, bl = r*0xffff/a, g*0xffff/a, bl*0xffff/a
				}
				fmt.Fprintf(h, "%02x%02x%02x%02x", a>>8, r>>8, g>>8, bl>>8)
			}
		}
		return fmt.Sprintf("data:image/png;pixels=%x", h.Sum(nil))
	})
	return metadataPI.ReplaceAllStringFunc(svgText, func(m string) string {
		src, err := code.Decode(metadataPI.FindStringSubmatch(m)[1])
		if err != nil {
			return "<?plantuml-src " + normalizeKO + "?>"
		}
		return "<?plantuml-src decoded:" + strconv.Quote(src) + "?>"
	})
}
