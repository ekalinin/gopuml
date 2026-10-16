// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/Tile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommonTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/AbstractTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/AbstractCommunicationTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/Bordered.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/EventsHistoryMode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/HookType.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/ComponentAdapter.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/UGraphicInterceptorTile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/asciiverse"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Tile is an element of the vertical flow of the Teoz layout: a message,
// a note, a fragment... It knows its height, its horizontal bounds (as
// Reals of the X line) and its vertical gauge (Reals of the Y line).
//
// As upstream, a tile is also an asciiverse.AsciiBlock (its text drawing)
// with its constraints and bounds on the text X line; abstractTile gives
// the defaults (see teoz_ascii.go).
type Tile interface {
	klimt.UDrawable
	asciiverse.AsciiBlock
	// AsciiAddConstraints adds the constraints of the tile on the text X
	// line.
	AsciiAddConstraints()
	// AsciiMinX and AsciiMaxX are the bounds on the text X line, nil for
	// none.
	AsciiMinX() Real
	AsciiMaxX() Real
	PreferredHeight() float64
	// OnGaugeResolved runs once the Y line is solved (livebox steps,
	// creations and destructions).
	OnGaugeResolved()
	YGauge() *YGauge
	AddConstraints()
	MinX() Real
	MaxX() Real
	// DrawnMinX and DrawnMaxX are the bounds a group frame is sized around;
	// they default to MinX and MaxX.
	DrawnMinX() Real
	DrawnMaxX() Real
	MiddleX() float64
	Event() Event
	ContactPointRelative() float64
	ZZZ() float64
	MatchAnchor(anchor string) bool
}

// abstractTile is upstream's CommonTile and AbstractTile. A tile embeds it
// and calls initTile with itself, so that the default methods reach its
// overrides.
type abstractTile struct {
	self          Tile
	stringBounder font.StringBounder
}

func (t *abstractTile) initTile(self Tile, stringBounder font.StringBounder) {
	t.self = self
	t.stringBounder = stringBounder
}

// StringBounder returns the text measurer of the layout.
func (t *abstractTile) StringBounder() font.StringBounder { return t.stringBounder }

// OnGaugeResolved does nothing by default.
func (t *abstractTile) OnGaugeResolved() {}

// MiddleX returns the middle of the bounds.
func (t *abstractTile) MiddleX() float64 {
	maxv := t.self.MaxX().CurrentValue()
	minv := t.self.MinX().CurrentValue()
	return (minv + maxv) / 2
}

// ContactPointRelative returns -1: the tile has no contact line.
func (t *abstractTile) ContactPointRelative() float64 { return -1 }

// ZZZ returns the height below the contact line.
func (t *abstractTile) ZZZ() float64 { return t.self.PreferredHeight() - t.self.ContactPointRelative() }

// MatchAnchor reports whether the tile is the message of the anchor.
func (t *abstractTile) MatchAnchor(anchor string) bool {
	if msg, ok := t.self.Event().(AbstractMessage); ok {
		if a, has := msg.Base().Anchor(); has && a == anchor {
			return true
		}
	}
	return false
}

// DrawnMinX returns MinX.
func (t *abstractTile) DrawnMinX() Real { return t.self.MinX() }

// DrawnMaxX returns MaxX.
func (t *abstractTile) DrawnMaxX() Real { return t.self.MaxX() }

// abstractCommunicationTile is upstream's AbstractCommunicationTile.
type abstractCommunicationTile struct {
	abstractTile
	skinParamCommunication klimt.ISkinParam
}

func (t *abstractCommunicationTile) sequenceMessageSpan() bool {
	return t.skinParamCommunication.Pragma().IsTrue(klimt.PragmaSequenceMessageSpan)
}

// context2D is upstream's (Context2D) ug cast: tiles draw on a
// UGraphicInterceptorTile, which tells whether the background is drawn.
func context2D(ug klimt.UGraphic) klimt.Context2D { return ug.(klimt.Context2D) }

// Bordered gives the horizontal borders of the diagram.
type Bordered interface {
	Border1() float64
	Border2() float64
}

// EventsHistoryMode tells which activations after an event count for its
// activation level.
type EventsHistoryMode int

// Events history modes, in upstream's order.
const (
	IgnoreFutureDeactivate EventsHistoryMode = iota
	IgnoreFutureActivate
	ConsiderFutureDeactivate
)

// HookType is upstream's HookType.
type HookType int

// Hook types, in upstream's order.
const (
	HookNegative HookType = iota
	HookPositive
)

// ComponentAdapter draws a component as a text block of its preferred
// size.
type ComponentAdapter struct {
	klimt.TextBlockDefaults
	component klimt.Component
}

// NewComponentAdapter returns the adapter; a nil component draws nothing.
func NewComponentAdapter(component klimt.Component) *ComponentAdapter {
	return &ComponentAdapter{component: component}
}

// DrawU draws the component.
func (c *ComponentAdapter) DrawU(ug klimt.UGraphic) {
	if c.component == nil {
		return
	}
	c.component.DrawU(ug, klimt.NewArea(c.CalculateDimension(ug.StringBounder())), klimt.NewSimpleContext2D(false))
}

// CalculateDimension returns the preferred size of the component.
func (c *ComponentAdapter) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	if c.component == nil {
		return geom.NewXDimension2D(0, 0)
	}
	width := c.component.PreferredWidth(stringBounder)
	height := c.component.PreferredHeight(stringBounder)
	return geom.NewXDimension2D(width, height)
}

// Backcolor returns nil.
func (c *ComponentAdapter) Backcolor() color.HColor { return nil }

// UGraphicInterceptorTile is the graphic tiles draw on: it tells whether
// the background or the foreground is drawn, and draws tiles given as
// shapes.
type UGraphicInterceptorTile struct {
	klimt.UGraphicDelegator
	isBackground bool
}

var (
	_ klimt.UGraphic  = (*UGraphicInterceptorTile)(nil)
	_ klimt.Context2D = (*UGraphicInterceptorTile)(nil)
)

// NewUGraphicInterceptorTile returns the graphic over ug.
func NewUGraphicInterceptorTile(ug klimt.UGraphic, isBackground bool) *UGraphicInterceptorTile {
	return &UGraphicInterceptorTile{UGraphicDelegator: klimt.NewUGraphicDelegator(ug), isBackground: isBackground}
}

// Draw draws a tile on this graphic, and other shapes on the delegate.
func (u *UGraphicInterceptorTile) Draw(shape klimt.UShape) {
	if tile, ok := shape.(Tile); ok {
		tile.DrawU(u)
		return
	}
	u.Ug().Draw(shape)
}

// Apply returns the interceptor over the changed delegate.
func (u *UGraphicInterceptorTile) Apply(change klimt.UChange) klimt.UGraphic {
	return NewUGraphicInterceptorTile(u.Ug().Apply(change), u.isBackground)
}

// IsBackground implements klimt.Context2D.
func (u *UGraphicInterceptorTile) IsBackground() bool { return u.isBackground }

// optionalString is a Java String that may be null (ok false).
type optionalString struct {
	s  string
	ok bool
}

func optString(s string, ok bool) optionalString { return optionalString{s: s, ok: ok} }

func someString(s string) optionalString { return optionalString{s: s, ok: true} }

// displayCreateOpt is upstream's Display.create(CharSequence...) with
// lines that may be null.
func displayCreateOpt(lines ...optionalString) *klimt.Display {
	cs := make([]klimt.CharSequence, len(lines))
	for i, l := range lines {
		if l.ok {
			cs[i] = klimt.Text(l.s)
		}
	}
	return klimt.DisplayCreateSeq(cs)
}
