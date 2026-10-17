// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseActor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseBoundary.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseControl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseEntity.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseDatabase.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/ComponentRoseQueue.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"github.com/ekalinin/gopuml/internal/decoration/symbol"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// stickmanHead is the code that upstream repeats in ComponentRoseActor,
// ComponentRoseBoundary, ComponentRoseControl, ComponentRoseEntity and
// ComponentRoseDatabase: a drawing (the "stickman") with the name below it
// at the head of the lifeline, or above it at the tail.
type stickmanHead struct {
	klimt.AbstractTextualComponent
	stickman klimt.TextBlock
	head     bool
}

func (c *stickmanHead) initStickmanHead(self klimt.AbstractComponentHooks, style, stereo *klimt.Style,
	stringsToDisplay *klimt.Display, head bool, skinParam klimt.ISkinParam) {
	c.InitAbstractTextualComponent(self, style, stereo, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(0, 3, 0, 3), skinParam, stringsToDisplay, false)
	c.head = head
}

// DrawInternalU draws the name and the drawing.
func (c *stickmanHead) DrawInternalU(ug klimt.UGraphic, _ *klimt.Area) {
	textBlock := c.TextBlock()
	stringBounder := ug.StringBounder()
	dimStickman := c.stickman.CalculateDimension(stringBounder)
	delta := (c.PreferredWidth(stringBounder) - dimStickman.Width()) / 2

	if c.head {
		textBlock.DrawU(ug.Apply(klimt.NewUTranslate(c.textMiddlePostion(stringBounder), dimStickman.Height())))
		ug = ug.Apply(klimt.UTranslateDx(delta))
	} else {
		textBlock.DrawU(ug.Apply(klimt.UTranslateDx(c.textMiddlePostion(stringBounder))))
		ug = ug.Apply(klimt.NewUTranslate(delta, c.TextHeight(stringBounder)))
	}
	c.stickman.DrawU(ug)
}

func (c *stickmanHead) textMiddlePostion(stringBounder font.StringBounder) float64 {
	return (c.PreferredWidth(stringBounder) - c.TextWidth(stringBounder)) / 2.0
}

// PreferredHeight returns the height of the drawing and the name.
func (c *stickmanHead) PreferredHeight(stringBounder font.StringBounder) float64 {
	dimStickman := c.stickman.CalculateDimension(stringBounder)
	return dimStickman.Height() + c.TextHeight(stringBounder)
}

// PreferredWidth returns the width of the drawing or of the name.
func (c *stickmanHead) PreferredWidth(stringBounder font.StringBounder) float64 {
	dimStickman := c.stickman.CalculateDimension(stringBounder)
	return jmath.Max(dimStickman.Width(), c.TextWidth(stringBounder))
}

// ComponentRoseActor is the head or tail of an actor: the drawing of the
// actor style and the name.
type ComponentRoseActor struct{ stickmanHead }

// NewComponentRoseActor returns the head (head true) or tail of an actor.
func NewComponentRoseActor(actorStyle klimt.ActorStyle, style, stereo *klimt.Style, stringsToDisplay *klimt.Display,
	head bool, skinParam klimt.ISkinParam) *ComponentRoseActor {
	c := &ComponentRoseActor{}
	c.initStickmanHead(c, style, stereo, stringsToDisplay, head, skinParam)
	biColor := c.SymbolContext()
	c.stickman = actorStyle.TextBlock(biColor)
	return c
}

// ComponentRoseBoundary is the head or tail of a boundary.
type ComponentRoseBoundary struct{ stickmanHead }

// NewComponentRoseBoundary returns the head (head true) or tail of a
// boundary.
func NewComponentRoseBoundary(style, stereo *klimt.Style, stringsToDisplay *klimt.Display, head bool,
	skinParam klimt.ISkinParam) *ComponentRoseBoundary {
	c := &ComponentRoseBoundary{}
	c.initStickmanHead(c, style, stereo, stringsToDisplay, head, skinParam)
	biColor := c.SymbolContext()
	c.stickman = klimt.NewBoundary(biColor)
	return c
}

// ComponentRoseControl is the head or tail of a control.
type ComponentRoseControl struct{ stickmanHead }

// NewComponentRoseControl returns the head (head true) or tail of a
// control.
func NewComponentRoseControl(style, stereo *klimt.Style, stringsToDisplay *klimt.Display, head bool,
	skinParam klimt.ISkinParam) *ComponentRoseControl {
	c := &ComponentRoseControl{}
	c.initStickmanHead(c, style, stereo, stringsToDisplay, head, skinParam)
	biColor := c.SymbolContext()
	c.stickman = klimt.NewControl(biColor)
	return c
}

// ComponentRoseEntity is the head or tail of an entity.
type ComponentRoseEntity struct{ stickmanHead }

// NewComponentRoseEntity returns the head (head true) or tail of an entity.
func NewComponentRoseEntity(style, stereo *klimt.Style, stringsToDisplay *klimt.Display, head bool,
	skinParam klimt.ISkinParam) *ComponentRoseEntity {
	c := &ComponentRoseEntity{}
	c.initStickmanHead(c, style, stereo, stringsToDisplay, head, skinParam)
	biColor := c.SymbolContext()
	c.stickman = klimt.NewEntityDomain(biColor)
	return c
}

// ComponentRoseDatabase is the head or tail of a database: a small
// database symbol with the name.
type ComponentRoseDatabase struct{ stickmanHead }

// NewComponentRoseDatabase returns the head (head true) or tail of a
// database.
func NewComponentRoseDatabase(style, stereo *klimt.Style, stringsToDisplay *klimt.Display, head bool,
	skinParam klimt.ISkinParam) *ComponentRoseDatabase {
	c := &ComponentRoseDatabase{}
	c.initStickmanHead(c, style, stereo, stringsToDisplay, head, skinParam)
	biColor := c.SymbolContext()
	symbolContext := klimt.NewFashion(biColor.BackColor(), biColor.ForeColor()).WithStroke(c.Stroke()).
		WithShadow(biColor.DeltaShadow())
	c.stickman = symbol.Database.AsSmall(nil, klimt.TextBlockUtilsEmpty(16, 17), klimt.TextBlockUtilsEmpty(0, 0),
		symbolContext, geom.HorizontalAlignmentCenter)
	return c
}

// ComponentRoseQueue is the head or tail of a queue: a queue symbol around
// the name.
type ComponentRoseQueue struct {
	klimt.AbstractTextualComponent
	stickman klimt.TextBlock
	head     bool
}

// NewComponentRoseQueue returns the head (head true) or tail of a queue.
func NewComponentRoseQueue(style, stereo *klimt.Style, stringsToDisplay *klimt.Display, head bool,
	skinParam klimt.ISkinParam) *ComponentRoseQueue {
	c := &ComponentRoseQueue{}
	c.InitAbstractTextualComponent(c, style, stereo, klimt.LineBreakStrategyNone,
		klimt.ClockwiseTopRightBottomLeftOf(0, 3, 0, 3), skinParam, stringsToDisplay, false)
	biColor := c.SymbolContext()
	c.head = head
	c.stickman = symbol.Queue.AsSmall(klimt.TextBlockUtilsEmpty(0, 0), c.TextBlock(), klimt.TextBlockUtilsEmpty(0, 0),
		biColor, geom.HorizontalAlignmentCenter)
	return c
}

// DrawInternalU draws the symbol, centered.
func (c *ComponentRoseQueue) DrawInternalU(ug klimt.UGraphic, _ *klimt.Area) {
	stringBounder := ug.StringBounder()
	dimStickman := c.stickman.CalculateDimension(stringBounder)
	delta := (c.PreferredWidth(stringBounder) - dimStickman.Width()) / 2

	ug = ug.Apply(klimt.UTranslateDx(delta))
	c.stickman.DrawU(ug)
}

// PreferredHeight returns the height of the symbol.
func (c *ComponentRoseQueue) PreferredHeight(stringBounder font.StringBounder) float64 {
	return c.stickman.CalculateDimension(stringBounder).Height()
}

// PreferredWidth returns the width of the symbol.
func (c *ComponentRoseQueue) PreferredWidth(stringBounder font.StringBounder) float64 {
	return c.stickman.CalculateDimension(stringBounder).Width()
}
