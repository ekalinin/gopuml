// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolSimpleAbstract.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolActor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolActorBusiness.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolBoundary.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolControl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolEntityDomain.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolInterface.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// simpleSmall is the anonymous text block of upstream's
// USymbolSimpleAbstract.asSmall: a drawing (a stick man, a circle...) with
// the stereotype above it and the label below it. The symbols built on it
// have no container form: their AsBig panics.
type simpleSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
	stickman          klimt.TextBlock
}

func newSimpleSmall(label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	drawing func(*klimt.Fashion) klimt.TextBlock) *simpleSmall {
	if stereotype == nil {
		throwNPE()
	}
	return &simpleSmall{label: label, stereotype: stereotype, symbolContext: symbolContext,
		stickman: drawing(symbolContext)}
}

func (b *simpleSmall) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	dimStickMan := b.stickman.CalculateDimension(stringBounder)
	dimTotal := b.CalculateDimension(stringBounder)
	stickmanX := (dimTotal.Width() - dimStickMan.Width()) / 2
	stickmanY := dimStereo.Height()
	ug = b.symbolContext.Apply(ug)
	b.stickman.DrawU(ug.Apply(klimt.NewUTranslate(stickmanX, stickmanY)))
	labelX := (dimTotal.Width() - dimLabel.Width()) / 2
	labelY := dimStickMan.Height() + dimStereo.Height()

	// Actor bug?
	ug2 := klimt.UGraphicStencilCreateDim(ug, dimLabel)
	b.label.DrawU(ug2.Apply(klimt.NewUTranslate(labelX, labelY)))

	stereoX := (dimTotal.Width() - dimStereo.Width()) / 2
	b.stereotype.DrawU(ug.Apply(klimt.UTranslateDx(stereoX)))
}

func (b *simpleSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	dimActor := b.stickman.CalculateDimension(stringBounder)
	return geom.XDimension2DMergeLayoutT12B3(dimStereo, dimActor, dimLabel)
}

// simpleBase holds the AsBig of upstream's USymbolSimpleAbstract.
type simpleBase struct{ usymbolBase }

// AsBig panics, as upstream throws UnsupportedOperationException.
func (simpleBase) AsBig(klimt.TextBlock, geom.HorizontalAlignment, klimt.TextBlock, float64, float64,
	*klimt.Fashion, geom.HorizontalAlignment) klimt.TextBlock {
	throwUnsupported()
	return nil
}

// usymbolActor is the actor symbol, drawn in an actor style.
type usymbolActor struct {
	simpleBase
	actorStyle klimt.ActorStyle
}

// SNames returns actor.
func (s *usymbolActor) SNames() []klimt.SName { return []klimt.SName{klimt.SNameActor} }

// AsSmall returns the actor with the stereotype and the label.
func (s *usymbolActor) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return newSimpleSmall(label, stereotype, symbolContext, s.actorStyle.TextBlock)
}

// usymbolActorBusiness is the business actor symbol.
type usymbolActorBusiness struct{ simpleBase }

// SNames returns business.
func (*usymbolActorBusiness) SNames() []klimt.SName { return []klimt.SName{klimt.SNameBusiness} }

// AsSmall returns the business actor with the stereotype and the label.
func (*usymbolActorBusiness) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return newSimpleSmall(label, stereotype, symbolContext, klimt.ActorStyleStickmanBusiness.TextBlock)
}

// simpleShadow is the shadow of the boundary, control and entity drawings:
// 4 when the fashion has a shadow.
func simpleShadow(symbolContext *klimt.Fashion) float64 {
	if symbolContext.IsShadowing() {
		return 4.0
	}
	return 0.0
}

// usymbolBoundary is the boundary symbol.
type usymbolBoundary struct{ simpleBase }

// SNames returns boundary.
func (*usymbolBoundary) SNames() []klimt.SName { return []klimt.SName{klimt.SNameBoundary} }

// AsSmall returns the boundary with the stereotype and the label.
func (*usymbolBoundary) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return newSimpleSmall(label, stereotype, symbolContext, func(f *klimt.Fashion) klimt.TextBlock {
		return klimt.NewBoundary(f.WithDeltaShadow(simpleShadow(f)))
	})
}

// usymbolControl is the control symbol.
type usymbolControl struct{ simpleBase }

// SNames returns control.
func (*usymbolControl) SNames() []klimt.SName { return []klimt.SName{klimt.SNameControl} }

// AsSmall returns the control with the stereotype and the label.
func (*usymbolControl) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return newSimpleSmall(label, stereotype, symbolContext, func(f *klimt.Fashion) klimt.TextBlock {
		return klimt.NewControl(f.WithDeltaShadow(simpleShadow(f)))
	})
}

// usymbolEntityDomain is the entity symbol.
type usymbolEntityDomain struct{ simpleBase }

// SNames returns entity.
func (*usymbolEntityDomain) SNames() []klimt.SName { return []klimt.SName{klimt.SNameEntity} }

// AsSmall returns the entity with the stereotype and the label.
func (*usymbolEntityDomain) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return newSimpleSmall(label, stereotype, symbolContext, func(f *klimt.Fashion) klimt.TextBlock {
		return klimt.NewEntityDomain(f.WithDeltaShadow(simpleShadow(f)))
	})
}

// usymbolInterface is the interface symbol: a small circle.
type usymbolInterface struct{ simpleBase }

// SNames returns interface.
func (*usymbolInterface) SNames() []klimt.SName { return []klimt.SName{klimt.SNameInterface} }

// AsSmall returns the circle with the stereotype and the label.
func (*usymbolInterface) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return newSimpleSmall(label, stereotype, symbolContext, func(f *klimt.Fashion) klimt.TextBlock {
		return klimt.NewCircleInterface2(f.BackColor(), f.ForeColor(), simpleShadow(f))
	})
}
