// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/ExtendedGeneralPath.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package g2d

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
)

// extendedGeneralPath is a general path with SVG elliptical arcs (upstream's
// ExtendedGeneralPath, from Apache Batik): lines and curves go to an
// enclosed GeneralPath, arcs become the curves of a rotated Arc2D. A move is
// only written when a segment follows it. Only the operations
// DriverPathG2d uses are ported.
type extendedGeneralPath struct {
	gp generalPath
	// lastMove is set when the last segment is a move, not written yet.
	lastMove     bool
	moveX, moveY float64
	// The current point.
	cx, cy float64
}

// arcTo adds an elliptical arc to (x, y): radii rx and ry, the x axis of
// the ellipse rotated by angle degrees, and the SVG arc flags.
func (p *extendedGeneralPath) arcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	// Ensure radii are valid.
	if rx == 0 || ry == 0 {
		p.lineTo(x, y)
		return
	}
	p.checkMoveTo()
	x0, y0 := p.cx, p.cy
	if x0 == x && y0 == y {
		// The same end points: the arc is omitted.
		return
	}
	arc, ok := computeArc(x0, y0, rx, ry, angle, largeArcFlag, sweepFlag, x, y)
	if !ok {
		return
	}
	cx, cy := arc.x+float64(arc.w/2), arc.y+float64(arc.h/2)
	t := identity.translate(cx, cy).rotate(jmath.ToRadians(angle)).translate(-cx, -cy)
	p.gp.appendPath(arcPath(arc.x, arc.y, arc.w, arc.h, arc.start, arc.extent, t), true)
	p.lastMove = false
	p.cx, p.cy = x, y
}

// computeArc builds the unrotated arc of an SVG elliptical arc, following
// the implementation notes of the SVG specification.
func computeArc(x0, y0, rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) (arc2D, bool) {
	// Compute the half distance between the current and the final point.
	dx2 := float64((x0 - x) / 2.0)
	dy2 := float64((y0 - y) / 2.0)
	// Convert angle from degrees to radians.
	angle = jmath.ToRadians(math.Mod(angle, 360.0))
	cosAngle := jmath.Cos(angle)
	sinAngle := jmath.Sin(angle)

	// Step 1: compute (x1, y1).
	x1 := float64(cosAngle*dx2) + float64(sinAngle*dy2)
	y1 := float64(-sinAngle*dx2) + float64(cosAngle*dy2)
	// Ensure radii are large enough.
	rx = math.Abs(rx)
	ry = math.Abs(ry)
	prx := rx * rx
	pry := ry * ry
	px1 := x1 * x1
	py1 := y1 * y1
	radiiCheck := float64(px1/prx) + float64(py1/pry)
	if radiiCheck > 1 {
		rx = math.Sqrt(radiiCheck) * rx
		ry = math.Sqrt(radiiCheck) * ry
		prx = rx * rx
		pry = ry * ry
	}

	// Step 2: compute (cx1, cy1).
	sign := 1.0
	if largeArcFlag == sweepFlag {
		sign = -1
	}
	sq := (float64(prx*pry) - float64(prx*py1) - float64(pry*px1)) / (float64(prx*py1) + float64(pry*px1))
	if sq < 0 {
		sq = 0
	}
	coef := sign * math.Sqrt(sq)
	cx1 := float64(coef * ((rx * y1) / ry))
	cy1 := float64(coef * -((ry * x1) / rx))

	// Step 3: compute (cx, cy) from (cx1, cy1).
	sx2 := float64((x0 + x) / 2.0)
	sy2 := float64((y0 + y) / 2.0)
	cx := sx2 + (float64(cosAngle*cx1) - float64(sinAngle*cy1))
	cy := sy2 + (float64(sinAngle*cx1) + float64(cosAngle*cy1))

	// Step 4: compute the angle start and the angle extent.
	ux := (x1 - cx1) / rx
	uy := (y1 - cy1) / ry
	vx := (-x1 - cx1) / rx
	vy := (-y1 - cy1) / ry
	// Compute the angle start.
	n := math.Sqrt(float64(ux*ux) + float64(uy*uy))
	p := ux
	sign = 1
	if uy < 0 {
		sign = -1
	}
	angleStart := jmath.ToDegrees(sign * jmath.Acos(p/n))

	// Compute the angle extent.
	n = math.Sqrt((float64(ux*ux) + float64(uy*uy)) * (float64(vx*vx) + float64(vy*vy)))
	p = float64(ux*vx) + float64(uy*vy)
	sign = 1
	if float64(ux*vy)-float64(uy*vx) < 0 {
		sign = -1
	}
	angleExtent := jmath.ToDegrees(sign * jmath.Acos(p/n))
	if !sweepFlag && angleExtent > 0 {
		angleExtent -= 360
	} else if sweepFlag && angleExtent < 0 {
		angleExtent += 360
	}
	angleExtent = math.Mod(angleExtent, 360)
	angleStart = math.Mod(angleStart, 360)

	return arc2D{x: cx - rx, y: cy - ry, w: float64(rx * 2.0), h: float64(ry * 2.0), start: -angleStart, extent: -angleExtent}, true
}

// moveTo records a move; it is written when a segment follows.
func (p *extendedGeneralPath) moveTo(x, y float64) {
	p.lastMove = true
	p.moveX, p.moveY = x, y
	p.cx, p.cy = x, y
}

func (p *extendedGeneralPath) lineTo(x, y float64) {
	p.checkMoveTo()
	p.gp.lineTo(x, y)
	p.lastMove = false
	p.cx, p.cy = x, y
}

func (p *extendedGeneralPath) curveTo(x1, y1, x2, y2, x3, y3 float64) {
	p.checkMoveTo()
	p.gp.cubicTo(x1, y1, x2, y2, x3, y3)
	p.lastMove = false
	p.cx, p.cy = x3, y3
}

// checkMoveTo writes a pending move.
func (p *extendedGeneralPath) checkMoveTo() {
	if p.lastMove {
		p.gp.moveTo(p.moveX, p.moveY)
	}
}

func (p *extendedGeneralPath) path() *path { return p.gp.path() }

func (p *extendedGeneralPath) bounds2D() (x, y, w, h float64) { return p.gp.bounds2D() }
