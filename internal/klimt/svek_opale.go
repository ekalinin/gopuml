// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/Opale.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/utils/MathUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// The margins and folded corner of a note.
const (
	opaleCornersize = 10
	// OpaleMarginX1 is the left margin of the text of a note.
	OpaleMarginX1 = 6
	// OpaleMarginX2 is the right margin of the text of a note.
	OpaleMarginX2 = 15
	opaleMarginY  = 5
	opaleDelta    = 4
)

// Opale is the drawing of a note: a rectangle with a folded corner around
// a text, possibly with a pointer toward the noted element. Upstream
// defines it in svek.image; it lives here because the skin components
// (skin.rose) use its static shapes.
type Opale struct {
	TextBlockDefaults
	memo                TextBlockMemoized
	noteBackgroundColor color.HColor
	borderColor         color.HColor
	shadowing2          float64
	strategy            Direction
	hasStrategy         bool
	pp1                 geom.XPoint2D
	pp2                 geom.XPoint2D
	withLink            bool
	roundCorner         float64
	stroke              *UStroke
	textBlock           TextBlock
}

// NewOpale returns the note of a text; stroke may be nil.
func NewOpale(shadowing float64, borderColor, noteBackgroundColor color.HColor, textBlock TextBlock, withLink bool,
	stroke *UStroke) *Opale {
	return &Opale{noteBackgroundColor: noteBackgroundColor, withLink: withLink, shadowing2: shadowing,
		borderColor: borderColor, textBlock: textBlock, stroke: stroke}
}

// SetRoundCorner sets the radius of the corners.
func (o *Opale) SetRoundCorner(roundCorner float64) { o.roundCorner = roundCorner }

// SetOpale sets the side of the pointer and its two points.
func (o *Opale) SetOpale(strategy Direction, pp1, pp2 geom.XPoint2D) {
	o.strategy = strategy
	o.hasStrategy = true
	o.pp1 = pp1
	o.pp2 = pp2
}

func (o *Opale) width(stringBounder font.StringBounder) float64 {
	return o.textBlock.CalculateDimension(stringBounder).Width() + OpaleMarginX1 + OpaleMarginX2
}

func (o *Opale) height(stringBounder font.StringBounder) float64 {
	size := o.textBlock.CalculateDimension(stringBounder)
	return size.Height() + 2*opaleMarginY
}

// CalculateDimension returns the size of the note.
func (o *Opale) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return o.memo.Dimension(stringBounder, o.calculateDimensionSlow)
}

func (o *Opale) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	height := o.height(stringBounder)
	width := o.width(stringBounder)
	return geom.NewXDimension2D(width, height)
}

// DrawU draws the outline, the folded corner and the text. A pointer
// without side panics, as upstream throws IllegalArgumentException.
func (o *Opale) DrawU(ug UGraphic) {
	stringBounder := ug.StringBounder()
	ug = ug.Apply(Bg(o.noteBackgroundColor)).Apply(o.borderColor)
	var polygon *UPath
	switch {
	case !o.withLink:
		polygon = o.polygonNormal(stringBounder)
	case o.hasStrategy && o.strategy == DirectionLeft:
		polygon = o.polygonLeft(stringBounder, o.pp1, o.pp2)
	case o.hasStrategy && o.strategy == DirectionRight:
		polygon = o.polygonRight(stringBounder, o.pp1, o.pp2)
	case o.hasStrategy && o.strategy == DirectionUp:
		polygon = o.polygonUp(stringBounder, o.pp1, o.pp2)
	case o.hasStrategy && o.strategy == DirectionDown:
		polygon = o.polygonDown(stringBounder, o.pp1, o.pp2)
	default:
		throw(illegalArgument, "")
	}

	polygon.SetDeltaShadow(o.shadowing2)
	if o.stroke != nil {
		ug = ug.Apply(*o.stroke)
	}
	ug.Draw(polygon)
	ug.Draw(OpaleCorner(o.width(stringBounder), o.roundCorner))
	o.textBlock.DrawU(ug.Apply(NewUTranslate(OpaleMarginX1, opaleMarginY)))
}

func (o *Opale) polygonNormal(stringBounder font.StringBounder) *UPath {
	return OpalePolygonNormal(o.width(stringBounder), o.height(stringBounder), o.roundCorner)
}

// OpaleCorner returns the folded corner of a note of a width: upstream's
// static Opale.getCorner.
func OpaleCorner(width, roundCorner float64) *UPath {
	path := UPathNone()
	path.MoveTo(width-opaleCornersize, 0)
	if roundCorner == 0 {
		path.LineTo(width-opaleCornersize, opaleCornersize)
	} else {
		path.LineTo(width-opaleCornersize, opaleCornersize-roundCorner/4)
		path.ArcToPoint(geom.NewXPoint2D(width-opaleCornersize+roundCorner/4, opaleCornersize), roundCorner/4, 0, 0)
	}
	path.LineTo(width, opaleCornersize)
	path.LineTo(width-opaleCornersize, 0)
	path.ClosePath()
	return path
}

// OpalePolygonNormal returns the outline of a note without pointer:
// upstream's static Opale.getPolygonNormal.
func OpalePolygonNormal(width, height, roundCorner float64) *UPath {
	polygon := UPathNone()
	if roundCorner == 0 {
		polygon.MoveTo(0, 0)
		polygon.LineTo(0, height)
		polygon.LineTo(width, height)
		polygon.LineTo(width, opaleCornersize)
		polygon.LineTo(width-opaleCornersize, 0)
		polygon.LineTo(0, 0)
	} else {
		polygon.MoveTo(0, roundCorner/2)
		polygon.LineTo(0, height-roundCorner/2)
		polygon.ArcToPoint(geom.NewXPoint2D(roundCorner/2, height), roundCorner/2, 0, 0)
		polygon.LineTo(width-roundCorner/2, height)
		polygon.ArcToPoint(geom.NewXPoint2D(width, height-roundCorner/2), roundCorner/2, 0, 0)
		polygon.LineTo(width, opaleCornersize)
		polygon.LineTo(width-opaleCornersize, 0)
		polygon.LineTo(roundCorner/2, 0)
		polygon.ArcToPoint(geom.NewXPoint2D(0, roundCorner/2), roundCorner/2, 0, 0)
	}
	polygon.ClosePath()
	return polygon
}

// mathUtilsLimitation is upstream's MathUtils.limitation: v bounded by min
// and max, or v itself when the bounds are empty.
func mathUtilsLimitation(v, lo, hi float64) float64 {
	if lo >= hi {
		return v
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (o *Opale) polygonLeft(stringBounder font.StringBounder, pp1, pp2 geom.XPoint2D) *UPath {
	h := o.height(stringBounder)
	w := o.width(stringBounder)
	r := o.roundCorner
	polygon := UPathNone()
	polygon.MoveTo(0, r/2)

	y1 := pp1.Y - opaleDelta
	y1 = mathUtilsLimitation(y1, 0, h-2*opaleDelta)
	polygon.LineTo(0, y1)
	polygon.LineTo(pp2.X, pp2.Y)
	polygon.LineTo(0, y1+2*opaleDelta)

	polygon.LineTo(0, h-r/2)
	polygon.ArcToPoint(geom.NewXPoint2D(r/2, h), r/2, 0, 0)
	polygon.LineTo(w-r/2, h)
	polygon.ArcToPoint(geom.NewXPoint2D(w, h-r/2), r/2, 0, 0)
	polygon.LineTo(w, opaleCornersize)
	polygon.LineTo(w-opaleCornersize, 0)
	polygon.LineTo(r/2, 0)
	polygon.ArcToPoint(geom.NewXPoint2D(0, r/2), r/2, 0, 0)
	polygon.ClosePath()
	return polygon
}

func (o *Opale) polygonRight(stringBounder font.StringBounder, pp1, pp2 geom.XPoint2D) *UPath {
	h := o.height(stringBounder)
	w := o.width(stringBounder)
	r := o.roundCorner
	polygon := UPathNone()
	polygon.MoveTo(0, r/2)
	polygon.LineTo(0, h-r/2)
	polygon.ArcToPoint(geom.NewXPoint2D(r/2, h), r/2, 0, 0)
	polygon.LineTo(w-r/2, h)
	polygon.ArcToPoint(geom.NewXPoint2D(w, h-r/2), r/2, 0, 0)

	y1 := pp1.Y - opaleDelta
	y1 = mathUtilsLimitation(y1, opaleCornersize, h-2*opaleDelta)
	polygon.LineTo(w, y1+2*opaleDelta)
	polygon.LineTo(pp2.X, pp2.Y)
	polygon.LineTo(w, y1)

	polygon.LineTo(w, opaleCornersize)
	polygon.LineTo(w-opaleCornersize, 0)
	polygon.LineTo(r/2, 0)
	polygon.ArcToPoint(geom.NewXPoint2D(0, r/2), r/2, 0, 0)
	polygon.ClosePath()
	return polygon
}

func (o *Opale) polygonUp(stringBounder font.StringBounder, pp1, pp2 geom.XPoint2D) *UPath {
	h := o.height(stringBounder)
	w := o.width(stringBounder)
	r := o.roundCorner
	polygon := UPathNone()
	polygon.MoveTo(0, r/2)
	polygon.LineTo(0, h-r/2)
	polygon.ArcToPoint(geom.NewXPoint2D(r/2, h), r/2, 0, 0)
	polygon.LineTo(w-r/2, h)
	polygon.ArcToPoint(geom.NewXPoint2D(w, h-r/2), r/2, 0, 0)
	polygon.LineTo(w, opaleCornersize)
	polygon.LineTo(w-opaleCornersize, 0)

	x1 := pp1.X - opaleDelta
	x1 = mathUtilsLimitation(x1, 0, w-opaleCornersize)
	polygon.LineTo(x1+2*opaleDelta, 0)
	polygon.LineTo(pp2.X, pp2.Y)

	polygon.LineTo(x1, 0)
	polygon.LineTo(r/2, 0)
	polygon.ArcToPoint(geom.NewXPoint2D(0, r/2), r/2, 0, 0)
	polygon.ClosePath()
	return polygon
}

func (o *Opale) polygonDown(stringBounder font.StringBounder, pp1, pp2 geom.XPoint2D) *UPath {
	h := o.height(stringBounder)
	w := o.width(stringBounder)
	r := o.roundCorner
	polygon := UPathNone()
	polygon.MoveTo(0, r/2)
	polygon.LineTo(0, h-r/2)
	polygon.ArcToPoint(geom.NewXPoint2D(r/2, h), r/2, 0, 0)

	x1 := pp1.X - opaleDelta
	x1 = mathUtilsLimitation(x1, 0, w)
	polygon.LineTo(x1, h)
	polygon.LineTo(pp2.X, pp2.Y)
	polygon.LineTo(x1+2*opaleDelta, h)

	polygon.LineTo(w-r/2, h)
	polygon.ArcToPoint(geom.NewXPoint2D(w, h-r/2), r/2, 0, 0)
	polygon.LineTo(w, opaleCornersize)
	polygon.LineTo(w-opaleCornersize, 0)
	polygon.LineTo(r/2, 0)
	polygon.ArcToPoint(geom.NewXPoint2D(0, r/2), r/2, 0, 0)
	polygon.ClosePath()
	return polygon
}

// MarginX1 returns the left margin of the text.
func (o *Opale) MarginX1() int { return OpaleMarginX1 }
