// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/EntityPosition.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/LinkArrow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/NoteLinkStrategy.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Bag.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Together.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/CucaNote.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Tip.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Removeable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/Position.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// EntityPosition is where an entity sits relative to its container: inside
// (NORMAL) or on its border (entry and exit points, pins, ports).
type EntityPosition int

// Entity positions, in upstream's order.
const (
	EntityPositionNormal EntityPosition = iota
	EntityPositionEntryPoint
	EntityPositionExitPoint
	EntityPositionInputPin
	EntityPositionOutputPin
	EntityPositionExpansionInput
	EntityPositionExpansionOutput
	EntityPositionPortin
	EntityPositionPortout
)

// EntityPositionRadius is upstream's EntityPosition.RADIUS.
const EntityPositionRadius = 6.0

var entityPositionNames = [...]string{"NORMAL", "ENTRY_POINT", "EXIT_POINT", "INPUT_PIN", "OUTPUT_PIN",
	"EXPANSION_INPUT", "EXPANSION_OUTPUT", "PORTIN", "PORTOUT"}

// String returns upstream's constant name.
func (p EntityPosition) String() string { return entityPositionNames[p] }

// IsNormal reports whether the position is NORMAL.
func (p EntityPosition) IsNormal() bool { return p == EntityPositionNormal }

// IsInput reports whether the position is an input on the border.
func (p EntityPosition) IsInput() bool {
	switch p {
	case EntityPositionEntryPoint, EntityPositionInputPin, EntityPositionExpansionInput, EntityPositionPortin:
		return true
	}
	return false
}

// IsOutput reports whether the position is an output on the border.
func (p EntityPosition) IsOutput() bool {
	switch p {
	case EntityPositionExitPoint, EntityPositionOutputPin, EntityPositionExpansionOutput, EntityPositionPortout:
		return true
	}
	return false
}

// IsPort reports whether the position is a port.
func (p EntityPosition) IsPort() bool { return p == EntityPositionPortin || p == EntityPositionPortout }

// UsePortP reports whether links attach to the entity as to a port.
func (p EntityPosition) UsePortP() bool {
	return p.IsPort() || p == EntityPositionExitPoint || p == EntityPositionEntryPoint
}

// Dimension returns the size of the symbol of the position.
func (p EntityPosition) Dimension(rankdir geom.Rankdir) geom.XDimension2D {
	if p == EntityPositionExpansionInput || p == EntityPositionExpansionOutput {
		if rankdir == geom.RankdirTopToBottom {
			return geom.NewXDimension2D(EntityPositionRadius*2*4, EntityPositionRadius*2)
		}
		return geom.NewXDimension2D(EntityPositionRadius*2, EntityPositionRadius*2*4)
	}
	return geom.NewXDimension2D(EntityPositionRadius*2, EntityPositionRadius*2)
}

// ShapeType returns the shape of the position; NORMAL panics as upstream
// throws IllegalStateException.
func (p EntityPosition) ShapeType() klimt.ShapeType {
	if p == EntityPositionNormal {
		uml.Throw(uml.IllegalStateException, "")
	}
	if p == EntityPositionEntryPoint || p == EntityPositionExitPoint {
		return klimt.ShapeTypeRectanglePort
	}
	return klimt.ShapeTypeRectangle
}

// EntityPositionFromStereotype returns the position named by a stereotype
// label such as "<<entrypoint>>", NORMAL for other labels. "<<port>>"
// panics, as upstream throws UnsupportedOperationException.
func EntityPositionFromStereotype(label string, ok bool) EntityPosition {
	is := func(s string) bool { return ok && uml.EqualsIgnoreCase(s, label) }
	switch {
	case is("<<port>>"):
		uml.Throw(uml.UnsupportedOperationException, "")
	case is("<<entrypoint>>"):
		return EntityPositionEntryPoint
	case is("<<exitpoint>>"):
		return EntityPositionExitPoint
	case is("<<inputpin>>"):
		return EntityPositionInputPin
	case is("<<outputpin>>"):
		return EntityPositionOutputPin
	case is("<<expansioninput>>"):
		return EntityPositionExpansionInput
	case is("<<expansionoutput>>"):
		return EntityPositionExpansionOutput
	}
	return EntityPositionNormal
}

// LinkArrow is the small arrow drawn next to a link label (": label >").
type LinkArrow int

// Link arrows, in upstream's order.
const (
	LinkArrowNoneOrSeveral LinkArrow = iota
	LinkArrowDirectNormal
	LinkArrowBackward
)

var linkArrowNames = [...]string{"NONE_OR_SEVERAL", "DIRECT_NORMAL", "BACKWARD"}

// String returns upstream's constant name.
func (a LinkArrow) String() string { return linkArrowNames[a] }

// Reverse returns the arrow in the other direction.
func (a LinkArrow) Reverse() LinkArrow {
	switch a {
	case LinkArrowDirectNormal:
		return LinkArrowBackward
	case LinkArrowBackward:
		return LinkArrowDirectNormal
	}
	return LinkArrowNoneOrSeveral
}

// NoteLinkStrategy tells how the note of a link split by an association
// class is drawn.
type NoteLinkStrategy int

// Note link strategies, in upstream's order.
const (
	NoteLinkStrategyNormal NoteLinkStrategy = iota
	NoteLinkStrategyHalfPrintedFull
	NoteLinkStrategyHalfNotPrinted
)

var noteLinkStrategyNames = [...]string{"NORMAL", "HALF_PRINTED_FULL", "HALF_NOT_PRINTED"}

// String returns upstream's constant name.
func (s NoteLinkStrategy) String() string { return noteLinkStrategyNames[s] }

// ComputeDimension returns the size of a note of the given size.
func (s NoteLinkStrategy) ComputeDimension(width, height float64) geom.XDimension2D {
	switch s {
	case NoteLinkStrategyHalfPrintedFull:
		return geom.NewXDimension2D(width/2, height)
	case NoteLinkStrategyHalfNotPrinted:
		return geom.NewXDimension2D(0, 0)
	}
	return geom.NewXDimension2D(width, height)
}

// Position is a side of an element: upstream's utils.Position, moved here
// for the notes (design D2).
type Position int

// Positions, in upstream's order.
const (
	PositionRight Position = iota
	PositionLeft
	PositionBottom
	PositionTop
)

var positionNames = [...]string{"RIGHT", "LEFT", "BOTTOM", "TOP"}

// String returns upstream's constant name.
func (p Position) String() string { return positionNames[p] }

// PositionValueOf is upstream's Position.valueOf; an unknown name panics as
// upstream throws IllegalArgumentException.
func PositionValueOf(name string) Position {
	for i, n := range positionNames {
		if n == name {
			return Position(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException, "No enum constant net.sourceforge.plantuml.utils.Position."+name)
	return PositionRight
}

// WithRankdir returns the position for a layout direction: left to right
// turns RIGHT into BOTTOM, LEFT into TOP and back.
func (p Position) WithRankdir(rankdir geom.Rankdir) Position {
	if rankdir == geom.RankdirTopToBottom {
		return p
	}
	switch p {
	case PositionRight:
		return PositionBottom
	case PositionLeft:
		return PositionTop
	case PositionBottom:
		return PositionRight
	}
	return PositionLeft
}

// ReverseDirection returns the direction opposite to LEFT or RIGHT; other
// positions panic as upstream throws UnsupportedOperationException.
func (p Position) ReverseDirection() klimt.Direction {
	switch p {
	case PositionLeft:
		return klimt.DirectionRight
	case PositionRight:
		return klimt.DirectionLeft
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return klimt.DirectionRight
}

// Bag is an element of the stack of a diagram: a group entity or a
// together block.
type Bag interface {
	isBag()
}

// Together is a "together { }" block, which keeps its elements close.
type Together struct {
	parent *Together
}

func (*Together) isBag() {}

// Parent returns the enclosing together block, nil for none.
func (t *Together) Parent() *Together { return t.parent }

// Removeable is an element that "remove" can take out of the diagram.
type Removeable interface {
	IsRemoved() bool
}

// CucaNote is a note attached to a link or to the top or bottom of a
// group.
type CucaNote struct {
	display  *klimt.Display
	position Position
	colors   *color.Colors
	strategy NoteLinkStrategy
}

// NewCucaNote is upstream's CucaNote.build: a note with the NORMAL
// strategy.
func NewCucaNote(display *klimt.Display, position Position, colors *color.Colors) *CucaNote {
	return &CucaNote{display: display, position: position, colors: colors, strategy: NoteLinkStrategyNormal}
}

// WithStrategy returns the note with another strategy.
func (n *CucaNote) WithStrategy(strategy NoteLinkStrategy) *CucaNote {
	return &CucaNote{display: n.display, position: n.position, colors: n.colors, strategy: strategy}
}

// Display returns the text.
func (n *CucaNote) Display() *klimt.Display { return n.display }

// Strategy returns the strategy.
func (n *CucaNote) Strategy() NoteLinkStrategy { return n.strategy }

// Colors returns the colors.
func (n *CucaNote) Colors() *color.Colors { return n.colors }

// Position returns the position.
func (n *CucaNote) Position() Position { return n.position }

// Tip is one "note right of Class::member" tip; the tips of one side of a
// class share one TIPS entity, each with its own text, colors and
// stereotype.
type Tip struct {
	display    *klimt.Display
	colors     *color.Colors
	stereotype *klimt.Stereotype
}

// NewTip returns a tip.
func NewTip(display *klimt.Display, colors *color.Colors, stereotype *klimt.Stereotype) *Tip {
	return &Tip{display: display, colors: colors, stereotype: stereotype}
}

// Display returns the text.
func (t *Tip) Display() *klimt.Display { return t.display }

// Colors returns the colors.
func (t *Tip) Colors() *color.Colors { return t.colors }

// Stereotype returns the stereotype, nil for none.
func (t *Tip) Stereotype() *klimt.Stereotype { return t.stereotype }
