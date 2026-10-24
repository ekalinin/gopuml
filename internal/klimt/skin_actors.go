// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ActorStickMan.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ActorAwesome.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ActorHollow.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/CircleInterface.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/ActorStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// TextBlock returns the drawing of the actor style: upstream's
// ActorStyle.getTextBlock. A style outside the enum panics, as upstream
// throws IllegalStateException.
func (a ActorStyle) TextBlock(symbolContext *Fashion) TextBlock {
	switch a {
	case ActorStyleStickman:
		return NewActorStickMan(symbolContext, false)
	case ActorStyleStickmanBusiness:
		return NewActorStickMan(symbolContext, true)
	case ActorStyleAwesome:
		return NewActorAwesome(symbolContext)
	case ActorStyleHollow:
		return NewActorHollow(symbolContext)
	}
	throw(illegalState, "")
	return nil
}

// The dimensions of the stick man.
const (
	stickManArmsY       = 8
	stickManArmsLenght  = 13
	stickManBodyLenght  = 27
	stickManLegsX       = 13
	stickManLegsY       = 15
	stickManHeadDiam    = 16
	stickManHalfHead    = stickManHeadDiam / 2.0
	stickManArmsOrLegsX = max(stickManArmsLenght, stickManLegsX)
)

// ActorStickMan is the stick man actor, possibly the business actor (a
// stroke across the head).
type ActorStickMan struct {
	TextBlockDefaults
	fashion       *Fashion
	actorBusiness bool
}

// NewActorStickMan returns the stick man drawn with fashion.
func NewActorStickMan(fashion *Fashion, actorBusiness bool) *ActorStickMan {
	return &ActorStickMan{fashion: fashion, actorBusiness: actorBusiness}
}

// DrawU draws the head, the business stroke and the body.
func (a *ActorStickMan) DrawU(ug UGraphic) {
	startX := float64(stickManArmsOrLegsX) - stickManHalfHead + a.thickness()
	head := UEllipseBuild(stickManHeadDiam, stickManHeadDiam)
	centerX := startX + stickManHeadDiam/2
	path := UPathNone()
	path.MoveTo(0, 0)
	path.LineTo(0, stickManBodyLenght)
	path.MoveTo(-stickManArmsLenght, stickManArmsY)
	path.LineTo(stickManArmsLenght, stickManArmsY)
	path.MoveTo(0, stickManBodyLenght)
	path.LineTo(-stickManLegsX, stickManBodyLenght+stickManLegsY)
	path.MoveTo(0, stickManBodyLenght)
	path.LineTo(stickManLegsX, stickManBodyLenght+stickManLegsY)
	if a.fashion.DeltaShadow() != 0 {
		head.SetDeltaShadow(a.fashion.DeltaShadow())
		path.SetDeltaShadow(a.fashion.DeltaShadow())
	}
	ug = a.fashion.Apply(ug)
	ug.Apply(NewUTranslate(startX, a.thickness())).Draw(head)
	if a.actorBusiness {
		a.specialBusiness(ug.Apply(NewUTranslate(startX+stickManHeadDiam/2, a.thickness()+stickManHeadDiam/2)))
	}
	ug.Apply(NewUTranslate(centerX, stickManHeadDiam+a.thickness())).Apply(Bg(color.None())).Draw(path)
}

func (a *ActorStickMan) specialBusiness(ug UGraphic) {
	pi := math.Pi
	alpha := 21 * pi / 64
	p1 := stickManOnCircle(pi/4 + alpha)
	p2 := stickManOnCircle(pi/4 - alpha)
	ug = ug.Apply(UTranslatePoint(p1))
	ug.Draw(NewULine(p2.X-p1.X, p2.Y-p1.Y))
}

func stickManOnCircle(alpha float64) geom.XPoint2D {
	x := float64(stickManHeadDiam/2) * jmath.Cos(alpha)
	y := float64(stickManHeadDiam/2) * jmath.Sin(alpha)
	return geom.NewXPoint2D(x, y)
}

func (a *ActorStickMan) thickness() float64 { return a.fashion.Stroke().Thickness() }

// PreferredWidth returns the width of the arms or legs with the stroke.
func (a *ActorStickMan) PreferredWidth() float64 {
	return float64(stickManArmsOrLegsX*2) + 2*a.thickness()
}

// PreferredHeight returns the height of the head, body and legs with the
// stroke and the shadow.
func (a *ActorStickMan) PreferredHeight() float64 {
	return float64(stickManHeadDiam+stickManBodyLenght+stickManLegsY) + 2*a.thickness() + a.fashion.DeltaShadow() + 1
}

// CalculateDimension returns the preferred size.
func (a *ActorStickMan) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(a.PreferredWidth(), a.PreferredHeight())
}

// The dimensions of the awesome actor.
const (
	awesomeHeadDiam   = 32
	awesomeBodyWidth  = 54
	awesomeShoulder   = 16
	awesomeCollar     = 4
	awesomeRadius     = 8
	awesomeBodyHeight = 28
)

// ActorAwesome is the actor drawn as a bust.
type ActorAwesome struct {
	TextBlockDefaults
	fashion *Fashion
}

// NewActorAwesome returns the bust actor drawn with fashion.
func NewActorAwesome(fashion *Fashion) *ActorAwesome { return &ActorAwesome{fashion: fashion} }

// DrawU draws the head and the shoulders.
func (a *ActorAwesome) DrawU(ug UGraphic) {
	head := UEllipseBuild(awesomeHeadDiam, awesomeHeadDiam)
	centerX := a.PreferredWidth() / 2
	const (
		w = float64(awesomeBodyWidth)
		s = float64(awesomeShoulder)
		c = float64(awesomeCollar)
		r = float64(awesomeRadius)
		h = float64(awesomeBodyHeight)
	)
	path := UPathNone()
	path.MoveTo(0, c)
	path.CubicTo(c, c, w/2-s-c, c, w/2-s, 0)
	path.CubicTo(w/2-s/2, 0, w/2, s/2, w/2, s)
	path.LineTo(w/2, h-r)
	path.CubicTo(w/2, h-r/2, w/2-r/2, h, w/2-r, h)
	path.LineTo(-w/2+r, h)
	path.CubicTo(-w/2+r/2, h, -w/2, h-r/2, -w/2, h-r)
	path.LineTo(-w/2, s)
	path.CubicTo(-w/2, s/2, -w/2+s/2, 0, -w/2+s, 0)
	path.CubicTo(-w/2+s+c, c, -c, c, 0, c)
	path.ClosePath()
	if a.fashion.DeltaShadow() != 0 {
		head.SetDeltaShadow(a.fashion.DeltaShadow())
		path.SetDeltaShadow(a.fashion.DeltaShadow())
	}
	ug = a.fashion.Apply(ug)
	ug.Apply(NewUTranslate(centerX-head.Width()/2, a.thickness())).Draw(head)
	ug.Apply(NewUTranslate(centerX, head.Height()+a.thickness())).Draw(path)
}

func (a *ActorAwesome) thickness() float64 { return a.fashion.Stroke().Thickness() }

// PreferredWidth returns the width of the body with the stroke.
func (a *ActorAwesome) PreferredWidth() float64 { return awesomeBodyWidth + a.thickness()*2 }

// PreferredHeight returns the height of the head and body with the stroke.
func (a *ActorAwesome) PreferredHeight() float64 {
	return awesomeHeadDiam + awesomeBodyHeight + a.thickness()*2
}

// CalculateDimension returns the preferred size.
func (a *ActorAwesome) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(a.PreferredWidth(), a.PreferredHeight())
}

// The dimensions of the hollow actor.
const (
	hollowHeadDiam      = 9
	hollowBodyWidth     = 25
	hollowBodyHeight    = 21
	hollowNeckHeight    = 2
	hollowArmThickness  = 5
	hollowBodyThickness = 6
	hollowLegThickness  = 6
)

// ActorHollow is the actor drawn as an outline.
type ActorHollow struct {
	TextBlockDefaults
	fashion *Fashion
}

// NewActorHollow returns the outline actor drawn with fashion.
func NewActorHollow(fashion *Fashion) *ActorHollow { return &ActorHollow{fashion: fashion} }

// DrawU draws the head and the body outline.
func (a *ActorHollow) DrawU(ug UGraphic) {
	head := UEllipseBuild(hollowHeadDiam, hollowHeadDiam)
	centerX := a.PreferredWidth() / 2
	const (
		bw = float64(hollowBodyWidth)
		bh = float64(hollowBodyHeight)
		at = float64(hollowArmThickness)
		bt = float64(hollowBodyThickness)
		lt = float64(hollowLegThickness)
	)
	sqrt2 := math.Sqrt(2)
	path := UPathNone()
	path.MoveTo(-bw/2, 0)
	path.LineTo(-bw/2, at)
	path.LineTo(-bt/2, at)
	path.LineTo(-bt/2, bh-(bw+float64(lt*sqrt2)-bt)/2)
	path.LineTo(-bw/2, bh-float64(lt*sqrt2)/2)
	path.LineTo(-(bw/2 - float64(lt*sqrt2)/2), bh)
	path.LineTo(0, bh-(bw/2-float64(lt*sqrt2)/2))
	path.LineTo(+(bw/2 - float64(lt*sqrt2)/2), bh)
	path.LineTo(+bw/2, bh-float64(lt*sqrt2)/2)
	path.LineTo(+bt/2, bh-(bw+float64(lt*sqrt2)-bt)/2)
	path.LineTo(+bt/2, at)
	path.LineTo(+bw/2, at)
	path.LineTo(+bw/2, 0)
	path.LineTo(-bw/2, 0)
	path.ClosePath()
	if a.fashion.DeltaShadow() != 0 {
		head.SetDeltaShadow(a.fashion.DeltaShadow())
		path.SetDeltaShadow(a.fashion.DeltaShadow())
	}
	ug = a.fashion.Apply(ug)
	ug.Apply(NewUTranslate(centerX-head.Width()/2, a.thickness())).Draw(head)
	ug.Apply(NewUTranslate(centerX, head.Height()+a.thickness()+hollowNeckHeight)).Draw(path)
}

func (a *ActorHollow) thickness() float64 { return a.fashion.Stroke().Thickness() }

// PreferredWidth returns the width of the body with the stroke.
func (a *ActorHollow) PreferredWidth() float64 { return hollowBodyWidth + a.thickness()*2 }

// PreferredHeight returns the height of the head, neck and body with the
// stroke and the shadow.
func (a *ActorHollow) PreferredHeight() float64 {
	return hollowHeadDiam + hollowNeckHeight + hollowBodyHeight + a.thickness()*2 + a.fashion.DeltaShadow()
}

// CalculateDimension returns the preferred size.
func (a *ActorHollow) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(a.PreferredWidth(), a.PreferredHeight())
}

// CircleInterface is the circle of a provided interface.
type CircleInterface struct {
	thickness       float32
	headDiam        float64
	backgroundColor color.HColor
	foregroundColor color.HColor
}

// NewCircleInterface returns the circle of a diameter and a stroke
// thickness.
func NewCircleInterface(backgroundColor, foregroundColor color.HColor, headDiam float64,
	thickness float32) *CircleInterface {
	return &CircleInterface{backgroundColor: backgroundColor, foregroundColor: foregroundColor, headDiam: headDiam,
		thickness: thickness}
}

// DrawU draws the circle.
func (c *CircleInterface) DrawU(ug UGraphic) {
	head := UEllipseBuild(c.headDiam, c.headDiam)
	t := float64(c.thickness)
	ug.Apply(UStrokeWithThickness(t)).Apply(Bg(c.backgroundColor)).Apply(c.foregroundColor).
		Apply(NewUTranslate(t, t)).Draw(head)
}

// PreferredWidth returns the diameter with the stroke.
func (c *CircleInterface) PreferredWidth(font.StringBounder) float64 {
	return c.headDiam + float64(2*c.thickness)
}

// PreferredHeight returns the diameter with the stroke.
func (c *CircleInterface) PreferredHeight(font.StringBounder) float64 {
	return c.headDiam + float64(2*c.thickness)
}
