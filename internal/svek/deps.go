// SPDX-License-Identifier: MIT
// Provenance: original

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/decoration/symbol"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/svek/extremity"
)

// The arrow heads and middle decorations of links
// (internal/svek/extremity).

type extremityFactory interface {
	CreateTBRDrawableLegacy(p0, p1, p2 geom.XPoint2D, side *geom.Side) klimt.UDrawable
	CreateUDrawable(p0 geom.XPoint2D, angle float64, side *geom.Side) klimt.UDrawable
}

type extremityDrawable interface {
	klimt.UDrawable
	SomePoint() *geom.XPoint2D
	DeltaForKal() klimt.UTranslate
	DecorationLength() float64
}

type middleFactory interface {
	CreateUDrawable(angle float64) klimt.UDrawable
}

func extremityFactoryLegacy(d decoration.LinkDecor, backgroundColor color.HColor) extremityFactory {
	if f := extremity.FactoryLegacy(d, backgroundColor); f != nil {
		return f
	}
	return nil
}

func extremityFactoryComplete(d decoration.LinkDecor, backgroundColor color.HColor) extremityFactory {
	if f := extremity.FactoryComplete(d, backgroundColor); f != nil {
		return f
	}
	return nil
}

func extremityFactoryExtends(backgroundColor color.HColor) extremityFactory {
	return extremity.NewFactoryExtends(backgroundColor)
}

func extremityOther(polygon *klimt.UPolygon) klimt.UDrawable { return extremity.NewOther(polygon) }

// extremityArrowLine returns the DrawLineIfTransparent method of an
// ExtremityArrow, nil for another drawable.
func extremityArrowLine(d klimt.UDrawable) func(ug klimt.UGraphic) {
	if a, ok := d.(*extremity.Arrow); ok {
		return a.DrawLineIfTransparent
	}
	return nil
}

func middleFactoryFor(d decoration.LinkMiddleDecor, backColor, diagramBackColor color.HColor) middleFactory {
	return extremity.MiddleFactoryFor(d, backColor, diagramBackColor)
}

// The symbols drawn around groups (internal/decoration/symbol).

// uSymbol is symbol.USymbol.
type uSymbol = symbol.USymbol

func usymbolOf(s abel.USymbol) uSymbol {
	if result := symbol.FromKey(string(s)); result != nil {
		return result
	}
	return nil
}

func packageStyleToUSymbol(p klimt.PackageStyle) uSymbol {
	if result := symbol.PackageStyleToUSymbol(p); result != nil {
		return result
	}
	return nil
}

// The bodies of entities (abel).

func entityBody(e *abel.Entity, skinParam klimt.ISkinParam, showMethods, showFields bool,
	stereotype *klimt.Stereotype, style *klimt.Style, fontConfiguration *klimt.FontConfiguration) klimt.TextBlock {
	return e.Bodier().Body(skinParam, showMethods, showFields, stereotype, style, fontConfiguration)
}
