// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/LinkType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/LinkStrategy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package decoration

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// LinkStrategy tells how link decorations are drawn with Graphviz. Upstream
// defines it in the abel package; it lives here because LinkType needs it
// (design D2).
type LinkStrategy int

// Link strategies, in upstream's order.
const (
	// LinkStrategyLegacyToBeRemoved draws the tail and head in the
	// generated DOT and reads their angle back from Graphviz' SVG.
	LinkStrategyLegacyToBeRemoved LinkStrategy = iota
	// LinkStrategySimplest leaves tail and head out of the DOT and computes
	// the decoration angle from the Bezier data.
	LinkStrategySimplest
)

// String returns upstream's constant name.
func (s LinkStrategy) String() string {
	if s == LinkStrategyLegacyToBeRemoved {
		return "LEGACY_toberemoved"
	}
	return "SIMPLEST"
}

// LinkType is the kind of a link: its two end decorations, its line style
// and its middle decoration. It is an immutable value.
type LinkType struct {
	decor1      LinkDecor
	linkStyle   LinkStyle
	decor2      LinkDecor
	middleDecor LinkMiddleDecor
}

// NewLinkType returns the plain link type with two end decorations.
func NewLinkType(decor1, decor2 LinkDecor) LinkType {
	return LinkType{decor1: decor1, decor2: decor2, middleDecor: LinkMiddleDecorNone, linkStyle: LinkStyleNormal()}
}

// IsDoubleDecorated reports whether both ends are decorated.
func (t LinkType) IsDoubleDecorated() bool {
	return t.decor1 != LinkDecorNone && t.decor2 != LinkDecorNone
}

// LooksLikeRevertedForSvg reports whether only the second end is
// decorated.
func (t LinkType) LooksLikeRevertedForSvg() bool {
	return t.decor1 == LinkDecorNone && t.decor2 != LinkDecorNone
}

// LooksLikeNoDecorAtAllSvg reports whether both ends or none are
// decorated.
func (t LinkType) LooksLikeNoDecorAtAllSvg() bool {
	if t.decor1 == LinkDecorNone && t.decor2 == LinkDecorNone {
		return true
	}
	return t.decor1 != LinkDecorNone && t.decor2 != LinkDecorNone
}

// WithoutDecors1 returns the type without its first decoration.
func (t LinkType) WithoutDecors1() LinkType {
	t.decor1 = LinkDecorNone
	return t
}

// WithoutDecors2 returns the type without its second decoration.
func (t LinkType) WithoutDecors2() LinkType {
	t.decor2 = LinkDecorNone
	return t
}

// String is upstream's toString, for example "NONE-DASHED(null)-EXTENDS".
func (t LinkType) String() string {
	return t.decor1.String() + "-" + t.linkStyle.String() + "-" + t.decor2.String()
}

// IsInvisible reports whether the line is hidden.
func (t LinkType) IsInvisible() bool { return t.linkStyle.IsInvisible() }

// GoDashed returns the type with a dashed line.
func (t LinkType) GoDashed() LinkType {
	t.linkStyle = LinkStyleDashed()
	return t
}

// GoDotted returns the type with a dotted line.
func (t LinkType) GoDotted() LinkType {
	t.linkStyle = LinkStyleDotted()
	return t
}

// GoThickness returns the type with a line thickness.
func (t LinkType) GoThickness(thickness float64) LinkType {
	t.linkStyle = t.linkStyle.GoThickness(thickness)
	return t
}

// GoBold returns the type with a bold line.
func (t LinkType) GoBold() LinkType {
	t.linkStyle = LinkStyleBold()
	return t
}

// Inversed returns the type seen from the other end.
func (t LinkType) Inversed() LinkType {
	return LinkType{decor1: t.decor2, decor2: t.decor1, middleDecor: t.middleDecor.Inversed(), linkStyle: t.linkStyle}
}

// WithMiddleCircle returns the type with a circle in the middle.
func (t LinkType) WithMiddleCircle() LinkType { return t.withMiddle(LinkMiddleDecorCircle) }

// WithMiddleCircleCircled returns the type with a circled circle in the
// middle.
func (t LinkType) WithMiddleCircleCircled() LinkType {
	return t.withMiddle(LinkMiddleDecorCircleCircled)
}

// WithMiddleCircleCircled1 returns the type with a circle in the middle,
// circled on the first side.
func (t LinkType) WithMiddleCircleCircled1() LinkType {
	return t.withMiddle(LinkMiddleDecorCircleCircled1)
}

// WithMiddleCircleCircled2 returns the type with a circle in the middle,
// circled on the second side.
func (t LinkType) WithMiddleCircleCircled2() LinkType {
	return t.withMiddle(LinkMiddleDecorCircleCircled2)
}

// WithMiddleSubset returns the type with a subset sign in the middle.
func (t LinkType) WithMiddleSubset() LinkType { return t.withMiddle(LinkMiddleDecorSubset) }

// WithMiddleSuperset returns the type with a superset sign in the middle.
func (t LinkType) WithMiddleSuperset() LinkType { return t.withMiddle(LinkMiddleDecorSuperset) }

func (t LinkType) withMiddle(middle LinkMiddleDecor) LinkType {
	t.middleDecor = middle
	return t
}

// Invisible returns the type with a hidden line, upstream's getInvisible.
func (t LinkType) Invisible() LinkType {
	t.linkStyle = LinkStyleInvisible()
	return t
}

// SpecificDecorationSvek returns the DOT attributes of the decorations.
func (t LinkType) SpecificDecorationSvek(linkStrategy LinkStrategy) string {
	if linkStrategy == LinkStrategySimplest {
		return "arrowtail=none,arrowhead=none"
	}
	sb := ""
	isEmpty1 := t.decor1 == LinkDecorNone
	isEmpty2 := t.decor2 == LinkDecorNone
	switch {
	case isEmpty1 && isEmpty2:
		sb = "arrowtail=none,arrowhead=none"
	case !isEmpty1 && !isEmpty2:
		sb = "dir=both,arrowtail=empty,arrowhead=empty"
	case isEmpty1 && !isEmpty2:
		sb = "arrowtail=empty,arrowhead=none,dir=back"
	}
	arrowsize := math.Max(t.decor1.ArrowSize(), t.decor2.ArrowSize())
	if arrowsize > 0 {
		if sb != "" {
			sb += ","
		}
		sb += "arrowsize=" + jfmt.DoubleToString(arrowsize)
	}
	return sb
}

// Decor1 returns the decoration of the first end.
func (t LinkType) Decor1() LinkDecor { return t.decor1 }

// Style returns the line style.
func (t LinkType) Style() LinkStyle { return t.linkStyle }

// Decor2 returns the decoration of the second end.
func (t LinkType) Decor2() LinkDecor { return t.decor2 }

// IsExtends reports whether an end is an EXTENDS triangle.
func (t LinkType) IsExtends() bool {
	return t.decor1 == LinkDecorExtends || t.decor2 == LinkDecorExtends
}

// Part1 returns the type with the first decoration only.
func (t LinkType) Part1() LinkType {
	t.decor2 = LinkDecorNone
	return t
}

// Part2 returns the type with the second decoration only.
func (t LinkType) Part2() LinkType {
	t.decor1 = LinkDecorNone
	return t
}

// Stroke3 is upstream's getStroke3: the stroke of the line, given the
// default stroke (nil for none).
func (t LinkType) Stroke3(defaultThickness *klimt.UStroke) klimt.UStroke {
	if t.linkStyle.IsThicknessOverrided() {
		return t.linkStyle.Stroke3()
	}
	if defaultThickness == nil {
		return t.linkStyle.Stroke3()
	}
	if defaultThickness.DashVisible() == 0 && defaultThickness.DashSpace() == 0 {
		return t.linkStyle.GoThickness(defaultThickness.Thickness()).Stroke3()
	}
	return *defaultThickness
}

// MiddleDecor returns the middle decoration.
func (t LinkType) MiddleDecor() LinkMiddleDecor { return t.middleDecor }

// WithLollipopInterfaceEye2 returns the type without its first decoration.
func (t LinkType) WithLollipopInterfaceEye2() LinkType { return t.WithoutDecors1() }

// WithLollipopInterfaceEye1 returns the type without its second decoration.
func (t LinkType) WithLollipopInterfaceEye1() LinkType { return t.WithoutDecors2() }

// LinkTypeName returns the semantic name of the type for the
// data-link-type SVG attribute; ok is false when there is none.
func (t LinkType) LinkTypeName() (name string, ok bool) {
	switch {
	case t.has(LinkDecorComposition):
		return "composition", true
	case t.has(LinkDecorAggregation):
		return "aggregation", true
	case t.has(LinkDecorExtends):
		return "extension", true
	case t.has(LinkDecorRedefines):
		return "redefines", true
	case t.has(LinkDecorDefinedBy):
		return "definedby", true
	case t.has(LinkDecorArrow) || t.has(LinkDecorArrowTriangle):
		return "dependency", true
	case t.has(LinkDecorNotNavigable):
		return "not_navigable", true
	case t.has(LinkDecorCrowfoot) || t.has(LinkDecorCircleCrowfoot) || t.has(LinkDecorLineCrowfoot):
		return "crowfoot", true
	case t.has(LinkDecorCircleLine) || t.has(LinkDecorDoubleLine) || t.bothNone():
		return "association", true
	case t.has(LinkDecorPlus):
		return "nested", true
	}
	return "", false
}

func (t LinkType) has(decor LinkDecor) bool { return t.decor1 == decor || t.decor2 == decor }

func (t LinkType) bothNone() bool { return t.decor1 == LinkDecorNone && t.decor2 == LinkDecorNone }
