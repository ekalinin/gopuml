// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverRectangleSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverLineSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverEllipseSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverPolygonSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverPathSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverDotPathSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverPixelSvg.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverImagePng.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/svg/DriverImageSvgSvg.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// DriverRectangleSvg draws rectangles, clipped to the clip of its
// container.
type DriverRectangleSvg struct{ clipContainer klimt.ClipContainer }

// NewDriverRectangleSvg returns the driver.
func NewDriverRectangleSvg(clipContainer klimt.ClipContainer) DriverRectangleSvg {
	return DriverRectangleSvg{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver.
func (d DriverRectangleSvg) Draw(rect *klimt.URectangle, x, y float64, mapper *color.ColorMapper,
	param klimt.UParam, svg *SvgGraphics) {
	rx := rect.Rx()
	ry := rect.Ry()
	width := rect.Width()
	height := rect.Height()

	ApplyFillColor(svg, mapper, param)
	ApplyStrokeColor(svg, mapper, param)

	svg.SetStrokeWidth(param.Stroke().Thickness(), param.Stroke().DasharraySvg())

	if clip := d.clipContainer.Clip(); clip != nil {
		r := clip.ClippedRectangle(geom.NewXRectangle2D(x, y, width, height))
		x = r.X
		y = r.Y
		width = r.Width
		height = r.Height
		if height <= 0 {
			return
		}
	}
	svg.SvgRectangle(x, y, width, height, rx/2, ry/2, rect.DeltaShadow())
}

// ApplyFillColor sets the fill from the background of param: a gradient
// reference or the color.
func ApplyFillColor(svg *SvgGraphics, mapper *color.ColorMapper, param klimt.UParam) {
	switch background := param.Backcolor().(type) {
	case *color.HColorLinearGradient:
		svg.SetFillColor("url(#" + svg.CreateSvgLinearGradient(background, mapper) + ")")
	case *color.HColorGradient:
		id := svg.CreateSvgGradient(background.Color1().ToRGB(mapper), background.Color2().ToRGB(mapper),
			background.Policy())
		svg.SetFillColor("url(#" + id + ")")
	default:
		svg.SetFillColorWith(background.ToSvg(mapper), background.TransparentFillBehavior())
	}
}

// ApplyStrokeColor sets the stroke from the color of param: a gradient
// reference or the color.
func ApplyStrokeColor(svg *SvgGraphics, mapper *color.ColorMapper, param klimt.UParam) {
	switch c := param.Color().(type) {
	case *color.HColorLinearGradient:
		svg.SetStrokeColor("url(#" + svg.CreateSvgLinearGradient(c, mapper) + ")")
	case *color.HColorGradient:
		id := svg.CreateSvgGradient(c.Color1().ToRGB(mapper), c.Color2().ToRGB(mapper), c.Policy())
		svg.SetStrokeColor("url(#" + id + ")")
	default:
		svg.SetStrokeColor(c.ToSvg(mapper))
	}
}

// DriverLineSvg draws lines, clipped to the clip of its container.
type DriverLineSvg struct{ clipContainer klimt.ClipContainer }

// NewDriverLineSvg returns the driver.
func NewDriverLineSvg(clipContainer klimt.ClipContainer) DriverLineSvg {
	return DriverLineSvg{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver.
func (d DriverLineSvg) Draw(shape *klimt.ULine, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	x2 := x + shape.DX()
	y2 := y + shape.DY()

	if clip := d.clipContainer.Clip(); clip != nil {
		line, ok := clip.ClippedLine(geom.NewXLine2D(x, y, x2, y2))
		if !ok {
			return
		}
		x = line.X1
		y = line.Y1
		x2 = line.X2
		y2 = line.Y2
	}

	if gr, ok := param.Color().(*color.HColorGradient); ok {
		svg.SetStrokeColor(gr.Color1().ToSvg(mapper))
	} else {
		svg.SetStrokeColor(param.Color().ToSvg(mapper))
	}
	svg.SetStrokeWidth(param.Stroke().Thickness(), param.Stroke().DasharraySvg())
	svg.SvgLine(x, y, x2, y2, shape.DeltaShadow())
}

// DriverEllipseSvg draws ellipses and elliptical arcs; nothing is drawn
// when a corner of the bounding box is outside the clip.
type DriverEllipseSvg struct{ clipContainer klimt.ClipContainer }

// NewDriverEllipseSvg returns the driver.
func NewDriverEllipseSvg(clipContainer klimt.ClipContainer) DriverEllipseSvg {
	return DriverEllipseSvg{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver.
func (d DriverEllipseSvg) Draw(shape *klimt.UEllipse, x, y float64, mapper *color.ColorMapper,
	param klimt.UParam, svg *SvgGraphics) {
	width := shape.Width()
	height := shape.Height()

	if clip := d.clipContainer.Clip(); clip != nil {
		if !clip.IsInside(x, y) {
			return
		}
		if !clip.IsInside(x+width, y+height) {
			return
		}
	}

	ApplyStrokeColor(svg, mapper, param)
	ApplyFillColor(svg, mapper, param)

	svg.SetStrokeWidth(param.Stroke().Thickness(), param.Stroke().DasharraySvg())

	start := shape.Start()
	extend := shape.Extend()
	cx := x + width/2
	cy := y + height/2
	if start == 0 && extend == 0 {
		svg.SvgEllipse(cx, cy, width/2, height/2, shape.DeltaShadow())
		return
	}
	start = start + 90
	rad := func(deg float64) float64 { return deg * math.Pi / 180. }
	var x1, y1, x2, y2 float64
	if extend > 0 {
		x1 = cx + float64(jmath.Sin(rad(start))*width/2)
		y1 = cy + float64(jmath.Cos(rad(start))*height/2)
		x2 = cx + float64(jmath.Sin(rad(start+extend))*width/2)
		y2 = cy + float64(jmath.Cos(rad(start+extend))*height/2)
	} else {
		x1 = cx + float64(jmath.Sin(rad(start+extend))*width/2)
		y1 = cy + float64(jmath.Cos(rad(start+extend))*height/2)
		x2 = cx + float64(jmath.Sin(rad(start))*width/2)
		y2 = cy + float64(jmath.Cos(rad(start))*height/2)
	}
	svg.SvgArcEllipse(width/2, height/2, x1, y1, x2, y2)
}

// DriverPolygonSvg draws polygons; nothing is drawn when a point is outside
// the clip.
type DriverPolygonSvg struct{ clipContainer klimt.ClipContainer }

// NewDriverPolygonSvg returns the driver.
func NewDriverPolygonSvg(clipContainer klimt.ClipContainer) DriverPolygonSvg {
	return DriverPolygonSvg{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver.
func (d DriverPolygonSvg) Draw(shape *klimt.UPolygon, x, y float64, mapper *color.ColorMapper,
	param klimt.UParam, svg *SvgGraphics) {
	points := shape.PointArray(x, y)
	if clip := d.clipContainer.Clip(); clip != nil {
		for j := 0; j+1 < len(points); j += 2 {
			if !clip.IsInside(points[j], points[j+1]) {
				return
			}
		}
	}

	ApplyFillColor(svg, mapper, param)
	ApplyStrokeColor(svg, mapper, param)

	svg.SetStrokeWidth(param.Stroke().Thickness(), param.Stroke().DasharraySvg())

	svg.SvgPolygon(shape.DeltaShadow(), points...)
}

// DriverPathSvg draws paths; nothing is drawn when the path is not inside
// the clip.
type DriverPathSvg struct{ clipContainer klimt.ClipContainer }

// NewDriverPathSvg returns the driver.
func NewDriverPathSvg(clipContainer klimt.ClipContainer) DriverPathSvg {
	return DriverPathSvg{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver. A path whose color and background are the
// same simple color is filled without stroke.
func (d DriverPathSvg) Draw(shape *klimt.UPath, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	if clip := d.clipContainer.Clip(); clip != nil && !clip.IsInsidePath(x, y, shape) {
		return
	}

	c := param.Color()
	back := param.Backcolor()

	if c != nil && back != nil && c.Equals(back) && !isGradient(c) {
		svg.SetFillColor(c.ToSvg(mapper))
		svg.setStrokeColorEmpty()
		svg.SetStrokeWidth(0, nil)
	} else {
		ApplyFillColor(svg, mapper, param)
		ApplyStrokeColor(svg, mapper, param)

		svg.SetStrokeWidth(param.Stroke().Thickness(), param.Stroke().DasharraySvg())
	}

	svg.SvgPath(x, y, shape, shape.DeltaShadow())
}

func isGradient(c color.HColor) bool {
	switch c.(type) {
	case *color.HColorGradient, *color.HColorLinearGradient:
		return true
	}
	return false
}

// DriverDotPathSvg draws Graphviz curves as unfilled paths; nothing is
// drawn with a transparent color.
type DriverDotPathSvg struct{}

// Draw implements klimt.UDriver.
func (DriverDotPathSvg) Draw(shape *klimt.DotPath, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	if !param.Color().IsTransparent() {
		ApplyStrokeColor(svg, mapper, param)

		svg.SetFillColor("")
		svg.SetStrokeWidth(param.Stroke().Thickness(), param.Stroke().DasharraySvg())

		svg.SvgPath(x, y, shape.ToUPath(), 0)
	}
}

// DriverPixelSvg draws a point as a half-pixel square.
type DriverPixelSvg struct{}

// Draw implements klimt.UDriver.
func (DriverPixelSvg) Draw(_ klimt.UPixel, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	drawPixel(x, y, mapper, param, svg)
}

func drawPixel(x, y float64, mapper *color.ColorMapper, param klimt.UParam, svg *SvgGraphics) {
	if gr, ok := param.Color().(*color.HColorGradient); ok {
		svg.SetStrokeColor(gr.Color1().ToSvg(mapper))
	} else {
		svg.SetStrokeColor(param.Color().ToSvg(mapper))
	}
	svg.SetStrokeWidth(0.5, nil)

	svg.SvgRectangle(x, y, 0.5, 0.5, 0, 0, 0)
}

// DriverImagePng draws raster images; nothing is drawn when a corner is
// outside the clip.
type DriverImagePng struct{ clipContainer klimt.ClipContainer }

// NewDriverImagePng returns the driver.
func NewDriverImagePng(clipContainer klimt.ClipContainer) DriverImagePng {
	return DriverImagePng{clipContainer: clipContainer}
}

// Draw implements klimt.UDriver.
func (d DriverImagePng) Draw(image *klimt.UImage, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	if clip := d.clipContainer.Clip(); clip != nil {
		if !clip.IsInside(x, y) {
			return
		}
		if !clip.IsInside(x+float64(image.Width()), y+float64(image.Height())) {
			return
		}
	}
	svg.SvgImagePng(image.Image(1), x, y)
}

// DriverImageSvgSvg draws SVG images.
type DriverImageSvgSvg struct{}

// Draw implements klimt.UDriver.
func (DriverImageSvgSvg) Draw(image *klimt.UImageSvg, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	svg *SvgGraphics) {
	svg.SvgImageSvg(image, x, y)
}
