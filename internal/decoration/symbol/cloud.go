// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbolCloud.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// usymbolCloud is the cloud symbol: a ring of bubbles, drawn with random
// numbers seeded by the size of the cloud. Upstream's older drawing (NEW
// false) is not ported: upstream never uses it.
type usymbolCloud struct{ usymbolBase }

// SNames returns cloud.
func (usymbolCloud) SNames() []klimt.SName { return []klimt.SName{klimt.SNameCloud} }

func drawCloud(ug klimt.UGraphic, width, height, shadowing float64) {
	shape := cloudFrontier(width, height)
	shape.SetDeltaShadow(shadowing)

	ug.Apply(klimt.UTranslateDy(0)).Draw(shape)
}

// cloudFrontier is upstream's getSpecificFrontierForCloudNew.
func cloudFrontier(width, height float64) *klimt.UPath {
	rnd := newJavaRandom(int64(width) + 7919*int64(height))
	var points []geom.XPoint2D

	bubbleSize := 11.0
	if math.Max(width, height)/bubbleSize > 16 {
		bubbleSize = math.Max(width, height) / 16
	}

	const margin1 = 8

	pointA := geom.NewXPoint2D(margin1, margin1)
	pointB := geom.NewXPoint2D(width-margin1, margin1)
	pointC := geom.NewXPoint2D(width-margin1, height-margin1)
	pointD := geom.NewXPoint2D(margin1, height-margin1)

	if width > 100 && height > 100 {
		points = cloudComplex(rnd, points, bubbleSize, pointA, pointB, pointC, pointD)
	} else {
		points = cloudSimple(rnd, points, bubbleSize, pointA, pointB, pointC, pointD)
	}

	points = append(points, points[0])

	result := klimt.UPathNone()
	result.MoveToPoint(points[0])
	for i := 0; i < len(points)-1; i++ {
		cloudAddCurve(rnd, result, points[i], points[i+1])
	}
	return result
}

func cloudComplex(rnd *javaRandom, points []geom.XPoint2D, bubbleSize float64, pointA, pointB, pointC,
	pointD geom.XPoint2D) []geom.XPoint2D {
	const margin2 = 7
	points = cloudSpecialLine(bubbleSize, rnd, points, cloudMvX(pointA, margin2), cloudMvX(pointB, -margin2))
	points = append(points, cloudMvY(pointB, margin2))
	points = cloudSpecialLine(bubbleSize, rnd, points, cloudMvY(pointB, margin2), cloudMvY(pointC, -margin2))
	points = append(points, cloudMvX(pointC, -margin2))
	points = cloudSpecialLine(bubbleSize, rnd, points, cloudMvX(pointC, -margin2), cloudMvX(pointD, margin2))
	points = append(points, cloudMvY(pointD, -margin2))
	points = cloudSpecialLine(bubbleSize, rnd, points, cloudMvY(pointD, -margin2), cloudMvY(pointA, margin2))
	points = append(points, cloudMvX(pointA, margin2))
	return points
}

func cloudSimple(rnd *javaRandom, points []geom.XPoint2D, bubbleSize float64, pointA, pointB, pointC,
	pointD geom.XPoint2D) []geom.XPoint2D {
	points = cloudSpecialLine(bubbleSize, rnd, points, pointA, pointB)
	points = cloudSpecialLine(bubbleSize, rnd, points, pointB, pointC)
	points = cloudSpecialLine(bubbleSize, rnd, points, pointC, pointD)
	points = cloudSpecialLine(bubbleSize, rnd, points, pointD, pointA)
	return points
}

func cloudMvX(pt geom.XPoint2D, dx float64) geom.XPoint2D { return geom.NewXPoint2D(pt.X+dx, pt.Y) }

func cloudMvY(pt geom.XPoint2D, dy float64) geom.XPoint2D { return geom.NewXPoint2D(pt.X, pt.Y+dy) }

func cloudSpecialLine(bubbleSize float64, rnd *javaRandom, points []geom.XPoint2D, p1,
	p2 geom.XPoint2D) []geom.XPoint2D {
	change := geom.CoordinateChangeCreate(p1, p2)
	length := change.Length()
	middle := change.TrueCoordinate(length/2, -cloudRnd(rnd, 1, 1+math.Min(12, bubbleSize*0.8)))
	points = cloudBubbleLine(rnd, points, p1, middle, bubbleSize)
	points = cloudBubbleLine(rnd, points, middle, p2, bubbleSize)
	return points
}

func cloudBubbleLine(rnd *javaRandom, points []geom.XPoint2D, p1, p2 geom.XPoint2D,
	bubbleSize float64) []geom.XPoint2D {
	change := geom.CoordinateChangeCreate(p1, p2)
	length := change.Length()
	nb := javaIntCast(length / bubbleSize)
	if nb == 0 {
		bubbleSize = length / 2
		nb = javaIntCast(length / bubbleSize)
	}
	for i := range nb {
		points = append(points, cloudRndPoint(rnd, change.TrueCoordinate(float64(i)*length/float64(nb), 0),
			bubbleSize*.2))
	}
	return points
}

func cloudAddCurve(rnd *javaRandom, path *klimt.UPath, p1, p2 geom.XPoint2D) {
	change := geom.CoordinateChangeCreate(p1, p2)
	length := change.Length()
	coef := cloudRnd(rnd, .25, .35)
	middle := change.TrueCoordinate(length*coef, -length*cloudRnd(rnd, .4, .55))
	middle2 := change.TrueCoordinate(length*(1-coef), -length*cloudRnd(rnd, .4, .55))
	path.CubicToPoints(middle, middle2, p2)
}

func cloudRnd(rnd *javaRandom, a, b float64) float64 {
	return float64(rnd.nextDouble()*(b-a)) + a
}

func cloudRndPoint(rnd *javaRandom, pt geom.XPoint2D, v float64) geom.XPoint2D {
	x := pt.X + float64(v*rnd.nextDouble())
	y := pt.Y + float64(v*rnd.nextDouble())
	return geom.NewXPoint2D(x, y)
}

func cloudMargin() margin { return newMargin(15, 15, 15, 15) }

// cloudSmall is the anonymous text block of upstream's asSmall.
type cloudSmall struct {
	klimt.TextBlockDefaults
	label, stereotype klimt.TextBlock
	symbolContext     *klimt.Fashion
}

func (b *cloudSmall) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = klimt.UGraphicStencilCreateDim(ug, dim)
	ug = b.symbolContext.Apply(ug)
	drawCloud(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	m := cloudMargin()
	tb := klimt.TextBlockUtilsMergeTB(b.stereotype, b.label, geom.HorizontalAlignmentCenter)
	tb.DrawU(ug.Apply(klimt.NewUTranslate(m.x1, m.y1)))
}

func (b *cloudSmall) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimLabel := b.label.CalculateDimension(stringBounder)
	dimStereo := b.stereotype.CalculateDimension(stringBounder)
	return cloudMargin().addDimension(dimStereo.MergeTB(dimLabel))
}

// AsSmall returns the cloud around the stereotype and the label.
func (usymbolCloud) AsSmall(_, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
	_ geom.HorizontalAlignment) klimt.TextBlock {
	return &cloudSmall{label: label, stereotype: stereotype, symbolContext: symbolContext}
}

// cloudBig is the anonymous text block of upstream's asBig.
type cloudBig struct {
	klimt.TextBlockDefaults
	title, stereotype klimt.TextBlock
	width, height     float64
	symbolContext     *klimt.Fashion
}

func (b *cloudBig) DrawU(ug klimt.UGraphic) {
	dim := b.CalculateDimension(ug.StringBounder())
	ug = b.symbolContext.Apply(ug)
	drawCloud(ug, dim.Width(), dim.Height(), b.symbolContext.DeltaShadow())
	drawBigCentered(ug, b.title, b.stereotype, b.width, 13, 13)
}

func (b *cloudBig) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

// drawBigCentered draws the stereotype centered at stereoY and the title
// centered below it, at titleY plus the height of the stereotype.
func drawBigCentered(ug klimt.UGraphic, title, stereotype klimt.TextBlock, width, stereoY, titleY float64) {
	dimStereo := stereotype.CalculateDimension(ug.StringBounder())
	posStereo := (width - dimStereo.Width()) / 2
	stereotype.DrawU(ug.Apply(klimt.NewUTranslate(posStereo, stereoY)))
	dimTitle := title.CalculateDimension(ug.StringBounder())
	posTitle := (width - dimTitle.Width()) / 2
	title.DrawU(ug.Apply(klimt.NewUTranslate(posTitle, titleY+dimStereo.Height())))
}

// AsBig returns the cloud of a size with the stereotype and the title.
func (usymbolCloud) AsBig(title klimt.TextBlock, _ geom.HorizontalAlignment, stereotype klimt.TextBlock,
	width, height float64, symbolContext *klimt.Fashion, _ geom.HorizontalAlignment) klimt.TextBlock {
	return &cloudBig{title: title, stereotype: stereotype, width: width, height: height,
		symbolContext: symbolContext}
}
