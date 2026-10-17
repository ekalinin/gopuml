// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Margins.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ShapeType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/IEntityImage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ConditionStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/ConditionEndStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/DotSplines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tikz/TikzFontDistortion.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Hideable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"regexp"

	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// The small types below are defined by upstream in diagram-level packages
// (svek, dot, tikz, abel); they live here because classes of this package
// use them, and those packages import this one.

// Margins are the extra space around an entity image. Upstream defines it in
// svek.
type Margins struct {
	x1, x2, y1, y2 float64
}

// MarginsNone returns zero margins.
func MarginsNone() Margins { return Margins{} }

// MarginsUniform returns the same margin on every side.
func MarginsUniform(value float64) Margins { return Margins{value, value, value, value} }

// NewMargins returns the margins.
func NewMargins(x1, x2, y1, y2 float64) Margins { return Margins{x1: x1, x2: x2, y1: y1, y2: y2} }

// String is upstream's toString.
func (m Margins) String() string {
	return "MARGIN[" + jstr(m.x1) + "," + jstr(m.x2) + "," + jstr(m.y1) + "," + jstr(m.y2) + "]"
}

// Merge returns the larger margin of each side.
func (m Margins) Merge(other Margins) Margins {
	return Margins{math.Max(m.x1, other.x1), math.Max(m.x2, other.x2), math.Max(m.y1, other.y1),
		math.Max(m.y2, other.y2)}
}

// IsZero reports whether every margin is zero.
func (m Margins) IsZero() bool { return m.x1 == 0 && m.x2 == 0 && m.y1 == 0 && m.y2 == 0 }

// X1 returns the left margin.
func (m Margins) X1() float64 { return m.x1 }

// X2 returns the right margin.
func (m Margins) X2() float64 { return m.x2 }

// Y1 returns the top margin.
func (m Margins) Y1() float64 { return m.y1 }

// Y2 returns the bottom margin.
func (m Margins) Y2() float64 { return m.y2 }

// TotalWidth returns the sum of the left and right margins.
func (m Margins) TotalWidth() float64 { return m.x1 + m.x2 }

// TotalHeight returns the sum of the top and bottom margins.
func (m Margins) TotalHeight() float64 { return m.y1 + m.y2 }

// ShapeType is the Graphviz shape of an entity. Upstream defines it in svek.
type ShapeType int

// Shape types, in upstream's order.
const (
	ShapeTypeRectangle ShapeType = iota
	ShapeTypeRectanglePort
	ShapeTypeRectangleWithCircleInside
	ShapeTypeRectangleHTMLForPorts
	ShapeTypeRoundRectangle
	ShapeTypeCircle
	ShapeTypeOval
	ShapeTypeDiamond
	ShapeTypeOctagon
	ShapeTypeFolder
	ShapeTypeHexagon
	ShapeTypePort
)

// String returns upstream's constant name.
func (t ShapeType) String() string {
	return enumName("ShapeType", []string{"RECTANGLE", "RECTANGLE_PORT", "RECTANGLE_WITH_CIRCLE_INSIDE",
		"RECTANGLE_HTML_FOR_PORTS", "ROUND_RECTANGLE", "CIRCLE", "OVAL", "DIAMOND", "OCTAGON", "FOLDER", "HEXAGON",
		"PORT"}, int(t))
}

// Hideable is something that can be hidden. Upstream defines it in abel.
type Hideable interface {
	IsHidden() bool
}

// The constants of upstream's IEntityImage.
const (
	IEntityImageCorner     = 25
	IEntityImageMargin     = 5
	IEntityImageMarginLine = 5
)

// IEntityImage is the image of an entity laid out by Graphviz. Upstream
// defines it in svek; its default isCrash returns false.
type IEntityImage interface {
	Hideable
	TextBlock
	ShapeType() ShapeType
	Shield(stringBounder font.StringBounder) Margins
	OverscanX(stringBounder font.StringBounder) float64
	IsCrash() bool
}

// ConditionStyle is the shape of an activity condition. Upstream defines it
// in svek.
type ConditionStyle int

// Condition styles, in upstream's order.
const (
	ConditionStyleEmptyDiamond ConditionStyle = iota
	ConditionStyleInsideHexagon
	ConditionStyleInsideDiamond
)

var conditionStyleNames = []string{"EMPTY_DIAMOND", "INSIDE_HEXAGON", "INSIDE_DIAMOND"}

// String returns upstream's constant name.
func (c ConditionStyle) String() string {
	return enumName("ConditionStyle", conditionStyleNames, int(c))
}

// ConditionStyleFromString returns the style named value ignoring case,
// with the aliases InsideDiamond, Foo1, Diamond and Inside; ok is false when
// there is none (null upstream).
func ConditionStyleFromString(value string) (ConditionStyle, bool) {
	switch {
	case javaEqualsIgnoreCase("InsideDiamond", value), javaEqualsIgnoreCase("Foo1", value):
		return ConditionStyleInsideDiamond, true
	case javaEqualsIgnoreCase("Diamond", value):
		return ConditionStyleEmptyDiamond, true
	case javaEqualsIgnoreCase("Inside", value):
		return ConditionStyleInsideHexagon, true
	}
	for i, name := range conditionStyleNames {
		if javaEqualsIgnoreCase(name, value) {
			return ConditionStyle(i), true
		}
	}
	return 0, false
}

// ConditionEndStyle is the shape that ends an activity condition. Upstream
// defines it in svek.
type ConditionEndStyle int

// Condition end styles, in upstream's order.
const (
	ConditionEndStyleDiamond ConditionEndStyle = iota
	ConditionEndStyleHline
)

var conditionEndStyleNames = []string{"DIAMOND", "HLINE"}

// String returns upstream's constant name.
func (c ConditionEndStyle) String() string {
	return enumName("ConditionEndStyle", conditionEndStyleNames, int(c))
}

// ConditionEndStyleFromString returns the style named value ignoring case;
// ok is false when there is none (null upstream).
func ConditionEndStyleFromString(value string) (ConditionEndStyle, bool) {
	for i, name := range conditionEndStyleNames {
		if javaEqualsIgnoreCase(name, value) {
			return ConditionEndStyle(i), true
		}
	}
	return 0, false
}

// DotSplines is the Graphviz splines setting. Upstream defines it in dot.
type DotSplines int

// Splines settings, in upstream's order.
const (
	DotSplinesPolyline DotSplines = iota
	DotSplinesOrtho
	DotSplinesSplines
)

// String returns upstream's constant name.
func (d DotSplines) String() string {
	return enumName("DotSplines", []string{"POLYLINE", "ORTHO", "SPLINES"}, int(d))
}

// TikzFontDistortion is the font correction of the TikZ output: a
// magnification and a distortion. Upstream defines it in tikz.
type TikzFontDistortion struct {
	magnify     float64
	distortion  float64
	texPreamble string
}

var (
	tikzMagnifyPattern    = regexp.MustCompile(`^[0-9.]+$`)
	tikzDistortionPattern = regexp.MustCompile(`^[-0-9.]+$`)
)

// TikzFontDistortionFromValue parses "magnify;distortion"; anything else
// gives the default.
func TikzFontDistortionFromValue(value string, ok bool) *TikzFontDistortion {
	if !ok {
		return TikzFontDistortionDefault()
	}
	tokens := javaTokenize(value, ";")
	if len(tokens) < 2 {
		return TikzFontDistortionDefault()
	}
	v1, v2 := tokens[0], tokens[1]
	if tikzMagnifyPattern.MatchString(v1) && tikzDistortionPattern.MatchString(v2) {
		return &TikzFontDistortion{magnify: javaParseDouble(v1), distortion: javaParseDouble(v2)}
	}
	return TikzFontDistortionDefault()
}

// TikzFontDistortionDefault returns the default 1.20;4.0.
func TikzFontDistortionDefault() *TikzFontDistortion {
	return &TikzFontDistortion{magnify: 1.20, distortion: 4.0}
}

// String is upstream's toString.
func (t *TikzFontDistortion) String() string { return jstr(t.magnify) + ";" + jstr(t.distortion) }

// Magnify returns the magnification.
func (t *TikzFontDistortion) Magnify() float64 { return t.magnify }

// Distortion returns the distortion.
func (t *TikzFontDistortion) Distortion() float64 { return t.distortion }

// TexPreamble returns the TeX preamble, "" for none.
func (t *TikzFontDistortion) TexPreamble() string { return t.texPreamble }

// UpdateFromPragma takes the TeX preamble from the pragmas.
func (t *TikzFontDistortion) UpdateFromPragma(pragma *Pragma) {
	t.texPreamble, _ = pragma.Value(PragmaTexPreamble)
}
