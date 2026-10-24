// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/LinkDecor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/decoration/LinkMiddleDecor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package extremity

import (
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// The drawing methods of upstream's LinkDecor and LinkMiddleDecor enums.
// internal/decoration holds the enums but cannot depend on the drawing
// code, so they are functions of this package.

// FactoryComplete is LinkDecor.getExtremityFactoryComplete: the larger
// triangle for EXTENDS, else FactoryLegacy. It returns nil when upstream
// returns null.
func FactoryComplete(d decoration.LinkDecor, backgroundColor color.HColor) Factory {
	if d == decoration.LinkDecorExtends {
		return NewFactoryTriangle(nil, 18, 6, 18)
	}
	return FactoryLegacy(d, backgroundColor)
}

// FactoryLegacy is LinkDecor.getExtremityFactoryLegacy: the factory of the
// decoration, or nil when upstream returns null (NONE, EXTENDS and
// SQUARE_toberemoved).
func FactoryLegacy(d decoration.LinkDecor, backgroundColor color.HColor) Factory {
	switch d {
	case decoration.LinkDecorPlus:
		return NewFactoryPlus(backgroundColor)
	case decoration.LinkDecorRedefines:
		return NewFactoryExtendsLike(backgroundColor, false)
	case decoration.LinkDecorDefinedBy:
		return NewFactoryExtendsLike(backgroundColor, true)
	case decoration.LinkDecorHalfArrowUp:
		return NewFactoryHalfArrow(1)
	case decoration.LinkDecorHalfArrowDown:
		return NewFactoryHalfArrow(-1)
	case decoration.LinkDecorArrowTriangle:
		return NewFactoryTriangle(nil, 8, 3, 8)
	case decoration.LinkDecorCrowfoot:
		return NewFactoryCrowfoot()
	case decoration.LinkDecorCircleCrowfoot:
		return NewFactoryCircleCrowfoot()
	case decoration.LinkDecorLineCrowfoot:
		return NewFactoryLineCrowfoot()
	case decoration.LinkDecorCircleLine:
		return NewFactoryCircleLine()
	case decoration.LinkDecorDoubleLine:
		return NewFactoryDoubleLine()
	case decoration.LinkDecorCircleCross:
		return NewFactoryCircleCross(backgroundColor)
	case decoration.LinkDecorArrow:
		return NewFactoryArrow()
	case decoration.LinkDecorArrowAndCircle:
		return NewFactoryArrowAndCircle(backgroundColor)
	case decoration.LinkDecorNotNavigable:
		return NewFactoryNotNavigable()
	case decoration.LinkDecorAggregation:
		return NewFactoryDiamond(false)
	case decoration.LinkDecorComposition:
		return NewFactoryDiamond(true)
	case decoration.LinkDecorCircle:
		return NewFactoryCircle(false, backgroundColor)
	case decoration.LinkDecorCircleFill:
		return NewFactoryCircle(true, backgroundColor)
	case decoration.LinkDecorSquare:
		return NewFactorySquare(backgroundColor)
	case decoration.LinkDecorParenthesis:
		return NewFactoryParenthesis()
	case decoration.LinkDecorCircleConnect:
		return NewFactoryCircleConnect(backgroundColor)
	}
	return nil
}

// MiddleFactoryFor is LinkMiddleDecor.getMiddleFactory: backColor fills the
// circles, diagramBackColor hides the line under the arcs of CIRCLE_CIRCLED.
// NONE panics, as upstream throws UnsupportedOperationException.
func MiddleFactoryFor(d decoration.LinkMiddleDecor, backColor, diagramBackColor color.HColor) MiddleFactory {
	switch d {
	case decoration.LinkMiddleDecorCircle:
		return NewMiddleFactoryCircle(backColor)
	case decoration.LinkMiddleDecorCircleCircled:
		return NewMiddleFactoryCircleCircled(MiddleCircleCircledBoth, backColor, diagramBackColor)
	case decoration.LinkMiddleDecorCircleCircled1:
		return NewMiddleFactoryCircleCircled(MiddleCircleCircledMode1, backColor, diagramBackColor)
	case decoration.LinkMiddleDecorCircleCircled2:
		return NewMiddleFactoryCircleCircled(MiddleCircleCircledMode2, backColor, diagramBackColor)
	case decoration.LinkMiddleDecorSubset:
		return NewMiddleFactorySubset(false)
	case decoration.LinkMiddleDecorSuperset:
		return NewMiddleFactorySubset(true)
	}
	unsupported()
	return nil
}
