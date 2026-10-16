// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/PackageStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// PackageStyle is the shape of a package. Upstream defines it in svek; it
// lives here because Stereotype and ISkinParam return it. Upstream's
// toUSymbol belongs to the decoration.symbol package, which imports this
// one: it is a function of that package.
type PackageStyle int

// Package styles, in upstream's order.
const (
	PackageStyleFolder PackageStyle = iota
	PackageStyleRectangle
	PackageStyleNode
	PackageStyleFrame
	PackageStyleCloud
	PackageStyleDatabase
	PackageStyleAgent
	PackageStyleStorage
	PackageStyleComponent1
	PackageStyleComponent2
	PackageStyleArtifact
	PackageStyleCard
	packageStyleCount
)

var packageStyleNames = []string{"FOLDER", "RECTANGLE", "NODE", "FRAME", "CLOUD", "DATABASE", "AGENT", "STORAGE",
	"COMPONENT1", "COMPONENT2", "ARTIFACT", "CARD"}

// String returns upstream's constant name.
func (p PackageStyle) String() string { return enumName("PackageStyle", packageStyleNames, int(p)) }

// PackageStyleFromString returns the style named value ignoring case, or
// Rectangle for "rect"; ok is false when there is none (null upstream).
func PackageStyleFromString(value string) (PackageStyle, bool) {
	for i, name := range packageStyleNames {
		if javaEqualsIgnoreCase(name, value) {
			return PackageStyle(i), true
		}
	}
	if javaEqualsIgnoreCase("rect", value) {
		return PackageStyleRectangle, true
	}
	return 0, false
}

// DrawU draws the package border of a size; titleDim may be nil. Node and
// card panic, as upstream throws UnsupportedOperationException.
func (p PackageStyle) DrawU(ug UGraphic, d geom.XDimension2D, titleDim *geom.XDimension2D, shadowing bool) {
	if titleDim == nil {
		zero := geom.NewXDimension2D(0, 0)
		titleDim = &zero
	}
	width := d.Width()
	height := d.Height()
	switch p {
	case PackageStyleDatabase:
		drawPackageDatabase(ug, width, height, shadowing)
	case PackageStyleFolder:
		drawPackageFolder(ug, width, height, shadowing)
	case PackageStyleFrame:
		drawPackageFrame(ug, width, height, *titleDim, shadowing)
	case PackageStyleCloud:
		drawPackageCloud(ug, width, height, shadowing)
	case PackageStyleRectangle:
		drawPackageRect(ug, width, height, shadowing)
	case PackageStyleComponent1:
		drawPackageComponent1(ug, width, height, shadowing)
	case PackageStyleComponent2:
		drawPackageComponent2(ug, width, height, shadowing)
	case PackageStyleStorage:
		drawPackageStorage(ug, width, height, shadowing)
	case PackageStyleAgent:
		drawPackageRect(ug, width, height, shadowing)
	case PackageStyleArtifact:
		drawPackageArtifact(ug, width, height, shadowing)
	default:
		throw(unsupportedOperation, "")
	}
}

func drawPackageArtifact(ug UGraphic, width, height float64, shadowing bool) {
	polygon := NewUPolygon()
	polygon.AddPoint(0, 0)
	polygon.AddPoint(0, height)
	polygon.AddPoint(width, height)
	const cornersize = 10
	polygon.AddPoint(width, cornersize)
	polygon.AddPoint(width-cornersize, 0)
	polygon.AddPoint(0, 0)
	if shadowing {
		polygon.SetDeltaShadow(3.0)
	}
	ug.Draw(polygon)
	ug.Apply(UTranslateDx(width - cornersize)).Draw(ULineV(cornersize))
	ug.Apply(NewUTranslate(width, cornersize)).Draw(ULineH(-cornersize))
}

func drawPackageStorage(ug UGraphic, width, height float64, shadowing bool) {
	shape := URectangleBuild(width, height).Rounded(70)
	if shadowing {
		shape.SetDeltaShadow(3.0)
	}
	ug.Draw(shape)
}

func drawPackageComponent1(ug UGraphic, widthTotal, heightTotal float64, shadowing bool) {
	form := URectangleBuild(widthTotal, heightTotal)
	if shadowing {
		form.SetDeltaShadow(4)
	}
	small := URectangleBuild(10, 5)
	ug.Draw(form)
	// UML 1 Component Notation
	ug.Apply(NewUTranslate(-5, 5)).Draw(small)
	ug.Apply(NewUTranslate(-5, heightTotal-10)).Draw(small)
}

func drawPackageComponent2(ug UGraphic, widthTotal, heightTotal float64, shadowing bool) {
	form := URectangleBuild(widthTotal, heightTotal)
	if shadowing {
		form.SetDeltaShadow(4)
	}
	small := URectangleBuild(15, 10)
	tiny := URectangleBuild(4, 2)
	ug.Draw(form)
	// UML 2 Component Notation
	ug.Apply(NewUTranslate(widthTotal-20, 5)).Draw(small)
	ug.Apply(NewUTranslate(widthTotal-22, 7)).Draw(tiny)
	ug.Apply(NewUTranslate(widthTotal-22, 11)).Draw(tiny)
}

func drawPackageRect(ug UGraphic, width, height float64, shadowing bool) {
	shape := URectangleBuild(width, height)
	if shadowing {
		shape.SetDeltaShadow(3.0)
	}
	ug.Draw(shape)
}

func drawPackageCloud(ug UGraphic, width, height float64, shadowing bool) {
	shape := specificFrontierForCloud(width, height)
	if shadowing {
		shape.SetDeltaShadow(3.0)
	}
	ug.Apply(NewUTranslate(3, -3)).Draw(shape)
}

func specificFrontierForCloud(width, height float64) *UPath {
	path := UPathNone()
	path.MoveTo(0, 10)
	x := 0.0
	for i := 0; float64(i) < width-9; i += 10 {
		fi := float64(i)
		path.CubicTo(fi, -3+10, 2+fi, -5+10, 5+fi, -5+10)
		path.CubicTo(8+fi, -5+10, 10+fi, -3+10, 10+fi, 10)
		x = float64(i + 10)
	}
	y := 0.0
	for j := 10; float64(j) < height-9; j += 10 {
		fj := float64(j)
		path.CubicTo(x+3, fj, x+5, 2+fj, x+5, 5+fj)
		path.CubicTo(x+5, 8+fj, x+3, 10+fj, x, 10+fj)
		y = float64(j + 10)
	}
	for i := 0; float64(i) < width-9; i += 10 {
		fi := float64(i)
		path.CubicTo(x-fi, y+3, x-3-fi, y+5, x-5-fi, y+5)
		path.CubicTo(x-8-fi, y+5, x-10-fi, y+3, x-10-fi, y)
	}
	for j := 0; float64(j) < height-9-10; j += 10 {
		fj := float64(j)
		path.CubicTo(-3, y-fj, -5, y-2-fj, -5, y-5-fj)
		path.CubicTo(-5, y-8-fj, -3, y-10-fj, 0, y-10-fj)
	}
	return path
}

func drawPackageFrame(ug UGraphic, width, height float64, dimTitle geom.XDimension2D, shadowing bool) {
	shape := URectangleBuild(width, height)
	if shadowing {
		shape.SetDeltaShadow(3.0)
	}
	ug.Draw(shape)
	var textWidth, textHeight float64
	var cornersize int
	if dimTitle.Width() == 0 {
		textWidth = width / 3
		textHeight = 12
		cornersize = 7
	} else {
		textWidth = dimTitle.Width() + 10
		textHeight = dimTitle.Height() + 3
		cornersize = 10
	}
	polygon := UPathNone()
	polygon.MoveTo(textWidth, 1)
	polygon.LineTo(textWidth, textHeight-float64(cornersize))
	polygon.LineTo(textWidth-float64(cornersize), textHeight)
	polygon.LineTo(0, textHeight)
	ug.Draw(polygon)
}

func drawPackageFolder(ug UGraphic, width, height float64, shadowing bool) {
	wtitle := math.Max(30, width/4)
	shape := NewUPolygon()
	shape.AddPoint(0, 0)
	shape.AddPoint(wtitle, 0)
	const htitle = 10.0
	const marginTitleX3 = 7.0
	shape.AddPoint(wtitle+marginTitleX3, htitle)
	shape.AddPoint(width, htitle)
	shape.AddPoint(width, height)
	shape.AddPoint(0, height)
	shape.AddPoint(0, 0)
	if shadowing {
		shape.SetDeltaShadow(3.0)
	}
	ug.Draw(shape)
	ug.Apply(UTranslateDy(htitle)).Draw(ULineH(wtitle + marginTitleX3))
}

func drawPackageDatabase(ug UGraphic, width, height float64, shadowing bool) {
	shape := UPathNone()
	if shadowing {
		shape.SetDeltaShadow(3.0)
	}
	shape.MoveTo(0, 10)
	shape.CubicTo(10, 0, width/2-10, 0, width/2, 0)
	shape.CubicTo(width/2+10, 0, width-10, 0, width, 10)
	shape.LineTo(width, height-10)
	shape.CubicTo(width-10, height, width/2-10, height, width/2, height)
	shape.CubicTo(width/2+10, height, 10, height, 0, height-10)
	shape.LineTo(0, 10)
	ug.Draw(shape)
	closing := UPathNone()
	closing.MoveTo(0, 10)
	closing.CubicTo(10, 20, width/2-10, 20, width/2, 20)
	closing.CubicTo(width/2+10, 20, width-10, 20, width, 10)
	ug.Draw(closing)
}
