// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverRectangleG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverLineG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverPolygonG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverEllipseG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverPathG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverPixelG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverDotPathG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverImageG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverCenteredCharacterG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/DotPath.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package g2d

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/g2d/dejavu"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// argb returns a color as 0xAARRGGBB.
func argb(c color.XColor) uint32 { return uint32(c.RGB()) }

// isGradient reports whether a color is one of the gradients.
func isGradient(c color.HColor) bool {
	switch c.(type) {
	case *color.HColorGradient, *color.HColorLinearGradient:
		return true
	}
	return false
}

// driverRectangle draws rectangles.
type driverRectangle struct {
	dpiFactor float64
	visible   geom.EnsureVisible
}

func (d driverRectangle) Draw(rect *klimt.URectangle, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	g.setStroke(newBasicStroke(float64(float32(param.Stroke().Thickness()))))
	rx := rect.Rx()
	ry := rect.Ry()
	var s shape
	if rx == 0 && ry == 0 {
		s = rect2D{x, y, rect.Width(), rect.Height()}
	} else {
		s = roundRect2D{x, y, rect.Width(), rect.Height(), rx, ry}
	}

	d.visible.EnsureVisible(x, y)
	d.visible.EnsureVisible(x+rect.Width(), y+rect.Height())

	back := param.Backcolor()

	// Shadow
	if rect.DeltaShadow() != 0 {
		if back.IsTransparent() {
			drawOnlyLineShadowSpecial(g, s, rect.DeltaShadow(), d.dpiFactor)
		} else {
			drawShadow(g, s, rect.DeltaShadow(), d.dpiFactor)
		}
	}

	c := param.Color()
	if isGradient(back) {
		g.setPaint(paintGradient(x, y, mapper, rect.Width(), rect.Height(), back, g.tx))
		g.fill(s)
		drawBorder(param, c, mapper, rect, s, g, x, y)
	} else {
		if !back.IsTransparent() {
			g.setColor(argb(back.ToColor(mapper)))
			manageStroke(param.Stroke(), g)
			managePattern(param, g)
			g.fill(s)
		}
		if !c.Equals(back) {
			drawBorder(param, c, mapper, rect, s, g, x, y)
		}
	}
}

// sized is a shape with a width and a height (upstream's UShapeSized).
type sized interface {
	Width() float64
	Height() float64
}

// drawBorder strokes a shape with the color, unless it is transparent.
func drawBorder(param klimt.UParam, c color.HColor, mapper *color.ColorMapper, sz sized, s shape, g *graphics,
	x, y float64) {
	if c.IsTransparent() {
		return
	}
	if isGradient(c) {
		g.setPaint(paintGradient(x, y, mapper, sz.Width(), sz.Height(), c, g.tx))
	} else {
		g.setColor(argb(c.ToColor(mapper)))
	}
	manageStroke(param.Stroke(), g)
	g.draw(s)
}

// paintGradient returns the paint of a gradient over a box: a
// LinearGradientPaint for the SVG-like linear gradients, a two-color
// GradientPaint in the direction of the policy otherwise.
func paintGradient(x, y float64, mapper *color.ColorMapper, width, height float64, back color.HColor,
	tx affine) paint {
	if lg, ok := back.(*color.HColorLinearGradient); ok {
		return createLinearGradientPaint(lg, mapper, x, y, width, height, tx)
	}
	gr := back.(*color.HColorGradient)
	c1 := argb(gr.Color1().ToColor(mapper))
	c2 := argb(gr.Color2().ToColor(mapper))
	// As upstream, "(float) (y + height) / 2" divides the sum by 2.
	switch gr.Policy() {
	case '|':
		return newGradientPaint(float32(x), float32(y+height)/2, c1, float32(x+width), float32(y+height)/2, c2, tx)
	case '\\':
		return newGradientPaint(float32(x), float32(y+height), c1, float32(x+width), float32(y), c2, tx)
	case '-':
		return newGradientPaint(float32(x+width)/2, float32(y), c1, float32(x+width)/2, float32(y+height), c2, tx)
	}
	// for /
	return newGradientPaint(float32(x), float32(y), c1, float32(x+width), float32(y+height), c2, tx)
}

func createLinearGradientPaint(gr *color.HColorLinearGradient, mapper *color.ColorMapper, x, y, width, height float64,
	tx affine) paint {
	stops := gr.Stops()
	fractions := make([]float32, 0, len(stops))
	colors := make([]uint32, 0, len(stops))
	for _, stop := range stops {
		base := stop.Color().ToColor(mapper)
		alpha := int(math.Floor(float64(float64(base.Alpha())*stop.Opacity()) + 0.5))
		alpha = min(max(alpha, 0), 255)
		f := float32(stop.Offset())
		// Java throws IllegalArgumentException for fractions that do not
		// increase (upstream then fails); here a stop that does not move
		// forward replaces the previous one (design D16).
		if n := len(fractions); n > 0 && f <= fractions[n-1] {
			fractions = fractions[:n-1]
			colors = colors[:n-1]
		}
		fractions = append(fractions, f)
		colors = append(colors, uint32(alpha)<<24|argb(base)&0xFFFFFF)
	}
	var x1, y1, x2, y2 float32
	if gr.IsUserSpaceOnUse() {
		x1, y1, x2, y2 = float32(gr.X1()), float32(gr.Y1()), float32(gr.X2()), float32(gr.Y2())
	} else {
		x1 = float32(x + float64(gr.X1()*width))
		y1 = float32(y + float64(gr.Y1()*height))
		x2 = float32(x + float64(gr.X2()*width))
		y2 = float32(y + float64(gr.Y2()*height))
	}
	// Java rejects a single stop or a vector of length zero (upstream then
	// fails); paint the last color instead (design D16).
	if len(fractions) < 2 || x1 == x2 && y1 == y2 {
		return solidPaint(colors[len(colors)-1])
	}
	cycle := cycleNone
	switch gr.SpreadMethod() {
	case color.SpreadReflect:
		cycle = cycleReflect
	case color.SpreadRepeat:
		cycle = cycleRepeat
	}
	return newLinearGradientPaint(x1, y1, x2, y2, fractions, colors, cycle, tx)
}

// managePattern sets a texture paint for the stripe and circle patterns.
func managePattern(param klimt.UParam, g *graphics) {
	pattern := param.Pattern()
	if pattern == klimt.UPatternFull {
		return
	}
	rgb := argb(param.Backcolor().(*color.HColorSimple).AwtColor())
	t := &texturePaint{inv: g.tx.inverse()}
	switch pattern {
	case klimt.UPatternVerticalStripe:
		for i := range 4 {
			for j := range 4 {
				if i == 0 || i == 1 {
					t.tile[j*4+i] = rgb
				}
			}
		}
	case klimt.UPatternHorizontalStripe:
		for i := range 4 {
			for j := range 4 {
				if j == 0 || j == 1 {
					t.tile[j*4+i] = rgb
				}
			}
		}
	case klimt.UPatternSmallCircle:
		for _, p := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, 2}, {2, 1}} {
			t.tile[p[1]*4+p[0]] = rgb
		}
	default:
		return
	}
	g.setPaint(t)
}

// driverLine draws lines.
type driverLine struct{ dpiFactor float64 }

func (d driverLine) Draw(shape *klimt.ULine, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	line := line2D{x, y, x + shape.DX(), y + shape.DY()}
	manageStroke(param.Stroke(), g)
	// Shadow
	if shape.DeltaShadow() != 0 {
		drawShadow(g, line, shape.DeltaShadow(), d.dpiFactor)
	}
	drawBorder(param, param.Color(), mapper, shape, line, g, x, y)
}

// manageStroke sets the stroke: plain, or dashed with butt caps.
func manageStroke(stroke klimt.UStroke, g *graphics) {
	thickness := float64(float32(stroke.Thickness()))
	if stroke.DashVisible() == 0 {
		g.setStroke(newBasicStroke(thickness))
		return
	}
	dash1 := float64(float32(stroke.DashVisible()))
	dash2 := float64(float32(stroke.DashSpace()))
	g.setStroke(basicStroke{width: thickness, cap: capButt, join: joinMiter, miterLimit: 10,
		dash: []float64{dash1, dash2}})
}

// driverPolygon draws polygons.
type driverPolygon struct {
	dpiFactor float64
	visible   geom.EnsureVisible
}

func (d driverPolygon) Draw(shape *klimt.UPolygon, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	g.setStroke(newBasicStroke(float64(float32(param.Stroke().Thickness()))))

	p := &generalPath{}
	back := param.Backcolor()

	last := false
	for _, pt := range shape.Points() {
		xp := pt.X + x
		yp := pt.Y + y
		d.visible.EnsureVisible(xp, yp)
		if !last {
			p.moveTo(xp, yp)
		} else {
			p.lineTo(xp, yp)
		}
		last = true
	}
	if last {
		p.closePath()
	}

	if shape.DeltaShadow() != 0 {
		if back.IsTransparent() {
			drawOnlyLineShadowSpecial(g, p, shape.DeltaShadow(), d.dpiFactor)
		} else {
			drawShadow(g, p, shape.DeltaShadow(), d.dpiFactor)
		}
	}

	if isGradient(back) {
		g.setPaint(paintGradient(x, y, mapper, shape.Width(), shape.Height(), back, g.tx))
		g.fill(p)
	} else if !back.IsTransparent() {
		g.setColor(argb(back.ToColor(mapper)))
		managePattern(param, g)
		g.fill(p)
	}

	if !param.Color().IsTransparent() {
		g.setColor(argb(param.Color().ToColor(mapper)))
		manageStroke(param.Stroke(), g)
		g.draw(p)
	}
}

// driverEllipse draws ellipses and arcs.
type driverEllipse struct {
	dpiFactor float64
	visible   geom.EnsureVisible
}

func (d driverEllipse) Draw(ellipse *klimt.UEllipse, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	g.setStroke(newBasicStroke(float64(float32(param.Stroke().Thickness()))))
	d.visible.EnsureVisible(x, y)
	d.visible.EnsureVisible(x+ellipse.Width(), y+ellipse.Height())
	c := param.Color()
	if ellipse.Start() == 0 && ellipse.Extend() == 0 {
		s := ellipse2D{x, y, ellipse.Width(), ellipse.Height()}

		// Shadow
		if ellipse.DeltaShadow() != 0 {
			drawShadow(g, s, ellipse.DeltaShadow(), d.dpiFactor)
		}

		back := param.Backcolor()
		if isGradient(back) {
			g.setPaint(paintGradient(x, y, mapper, ellipse.Width(), ellipse.Height(), back, g.tx))
			g.fill(s)
			drawBorder(param, c, mapper, ellipse, s, g, x, y)
		} else {
			if !back.IsTransparent() {
				g.setColor(argb(param.Backcolor().ToColor(mapper)))
				managePattern(param, g)
				g.fill(s)
			}
			if !c.IsTransparent() && !c.Equals(param.Backcolor()) {
				drawBorder(param, c, mapper, ellipse, s, g, x, y)
			}
		}
	} else {
		arc := arc2D{x, y, ellipse.Width(), ellipse.Height(), ellipse.Start(), ellipse.Extend()}
		if !c.IsTransparent() {
			g.setColor(argb(c.ToColor(mapper)))
			g.draw(arc)
		}
	}
}

// driverPath draws paths.
type driverPath struct{ dpiFactor float64 }

func (d driverPath) Draw(shape *klimt.UPath, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	manageStroke(param.Stroke(), g)

	back := param.Backcolor()
	c := param.Color()

	p := &extendedGeneralPath{}
	minMax := geom.MinMaxEmpty(false)
	minMax = minMax.AddPointXY(x, y)
	for _, seg := range shape.Segments() {
		coord := seg.Coord()
		switch seg.SegmentType() {
		case geom.SegMoveTo:
			p.moveTo(x+coord[0], y+coord[1])
			minMax = minMax.AddPointXY(x+coord[0], y+coord[1])
		case geom.SegLineTo:
			p.lineTo(x+coord[0], y+coord[1])
			minMax = minMax.AddPointXY(x+coord[0], y+coord[1])
		case geom.SegCubicTo:
			p.curveTo(x+coord[0], y+coord[1], x+coord[2], y+coord[3], x+coord[4], y+coord[5])
			minMax = minMax.AddPointXY(x+coord[4], y+coord[5])
		case geom.SegArcTo:
			p.arcTo(coord[0], coord[1], coord[2], coord[3] != 0, coord[4] != 0, x+coord[5], y+coord[6])
		default:
			panic(&color.RuntimeError{Class: "java.lang.UnsupportedOperationException"})
		}
	}

	// Shadow
	if shape.DeltaShadow() != 0 {
		if back == nil || back.IsTransparent() {
			drawOnlyLineShadowSpecial(g, p, shape.DeltaShadow(), d.dpiFactor)
		} else {
			drawShadow(g, p, shape.DeltaShadow(), d.dpiFactor)
		}
	}

	if isGradient(back) {
		width := minMax.MaxX() - minMax.MinX()
		height := minMax.MaxY() - minMax.MinY()
		g.setPaint(paintGradient(minMax.MinX(), minMax.MinY(), mapper, width, height, back, g.tx))
		g.fill(p)
	} else if !back.IsTransparent() {
		g.setColor(argb(back.ToColor(mapper)))
		g.fill(p)
	}

	if !c.IsTransparent() && !c.Equals(back) {
		g.setColor(argb(c.ToColor(mapper)))
		g.draw(p)
	}
}

// drawPixel draws a pixel.
func drawPixel(x, y float64, mapper *color.ColorMapper, param klimt.UParam, g *graphics) {
	g.setColor(argb(param.Color().ToColor(mapper)))
	g.fillRect(int(x), int(y), 1, 1)
}

// driverDotPath draws the curves of Graphviz edges.
type driverDotPath struct{ visible geom.EnsureVisible }

func (d driverDotPath) Draw(shape *klimt.DotPath, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	manageStroke(param.Stroke(), g)

	if !param.Color().IsTransparent() {
		g.setColor(argb(param.Color().ToColor(mapper)))
		// DotPath.draw(Graphics2D): the curves appended to a GeneralPath.
		p := &generalPath{}
		for _, bez := range shape.Beziers() {
			c := &path{}
			c.moveTo(x+bez.X1, y+bez.Y1)
			c.cubicTo(x+bez.CtrlX1, y+bez.CtrlY1, x+bez.CtrlX2, y+bez.CtrlY2, x+bez.X2, y+bez.Y2)
			p.appendPath(c, true)
		}
		g.draw(p)
		shape.ManageEnsureVisible(x, y, d.visible)
	}
}

// driverImage draws raster images.
type driverImage struct {
	dpiFactor float64
	visible   geom.EnsureVisible
}

func (d driverImage) Draw(shape *klimt.UImage, x, y float64, mapper *color.ColorMapper, param klimt.UParam,
	g *graphics) {
	d.visible.EnsureVisible(x, y)
	d.visible.EnsureVisible(x+float64(shape.Width()), y+float64(shape.Height()))
	if d.dpiFactor == 1 {
		g.drawImage(fromPortable(shape.Image(1)), int(x), int(y))
	} else {
		back := g.tx
		g.tx = g.tx.scale(1/d.dpiFactor, 1/d.dpiFactor)
		g.drawImage(fromPortable(shape.Image(d.dpiFactor)), int(x*d.dpiFactor), int(y*d.dpiFactor))
		g.tx = back
	}
}

// fromPortable copies a portable image into an ARGB image.
func fromPortable(p klimt.PortableImage) *Image {
	im := NewImage(p.Width(), p.Height(), true)
	for y := range p.Height() {
		for x := range p.Width() {
			im.pix[y*im.width+x] = uint32(p.RGB(x, y))
		}
	}
	return im
}

// driverCenteredCharacter draws the centered characters of class headers
// and spots.
type driverCenteredCharacter struct{}

func (driverCenteredCharacter) Draw(characterCircled *klimt.UCenteredCharacter, x, y float64,
	mapper *color.ColorMapper, param klimt.UParam, g *graphics) {
	c := characterCircled.Char()
	f := characterCircled.Font()
	cx, cy, ok := dejavu.UnusedSpaceCenter(f, c)
	if !ok {
		// Outside of the table of upstream's results: the center of the
		// glyph's bounds approximates it.
		cx, cy = glyphCenter(f.UnderlyingFont(string(c)), c)
	}

	g.setColor(argb(param.Color().ToColor(mapper)))
	xpos := x - cx
	ypos := y - cy - 0.5

	text := string(c)
	g.drawString(text, f.UnderlyingFont(text), float64(float32(xpos)), float64(float32(ypos)))
}
