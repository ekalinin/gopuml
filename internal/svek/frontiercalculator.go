// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/FrontierCalculator.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

const frontierDelta = 3 * abel.EntityPositionRadius

// frontierCalculator is upstream's FrontierCalculator: the frontier of a
// cluster with entry or exit points (ports) is the box around its inner
// nodes and clusters, extended to the points and pushed out where a point
// sits on a corner.
type frontierCalculator struct {
	core    geom.RectangleArea
	initial geom.RectangleArea
}

func newFrontierCalculator(initial geom.RectangleArea, insides []geom.RectangleArea, points []geom.XPoint2D,
	rankdir geom.Rankdir) *frontierCalculator {
	f := &frontierCalculator{initial: initial}
	hasCore := false
	for _, in := range insides {
		if !hasCore {
			f.core, hasCore = in, true
		} else {
			f.core = f.core.Merge(in)
		}
	}
	if !hasCore {
		center := initial.PointCenter()
		f.core = geom.NewRectangleArea(center.X-1, center.Y-1, center.X+1, center.Y+1)
	}
	for _, p := range points {
		f.core = f.core.MergePoint(p)
	}
	var touchMinX, touchMaxX, touchMinY, touchMaxY bool
	for _, p := range points {
		if p.X == f.core.MinX() {
			touchMinX = true
		}
		if p.X == f.core.MaxX() {
			touchMaxX = true
		}
		if p.Y == f.core.MinY() {
			touchMinY = true
		}
		if p.Y == f.core.MaxY() {
			touchMaxY = true
		}
	}
	if !touchMinX {
		f.core = f.core.WithMinX(initial.MinX())
	}
	if !touchMaxX {
		f.core = f.core.WithMaxX(initial.MaxX())
	}
	if !touchMinY {
		f.core = f.core.WithMinY(initial.MinY())
	}
	if !touchMaxY {
		f.core = f.core.WithMaxY(initial.MaxY())
	}
	var pushMinX, pushMaxX, pushMinY, pushMaxY bool
	for _, p := range points {
		if p.Y == f.core.MinY() || p.Y == f.core.MaxY() {
			if math.Abs(p.X-f.core.MaxX()) < frontierDelta {
				pushMaxX = true
			}
			if math.Abs(p.X-f.core.MinX()) < frontierDelta {
				pushMinX = true
			}
		}
		if p.X == f.core.MinX() || p.X == f.core.MaxX() {
			if math.Abs(p.Y-f.core.MaxY()) < frontierDelta {
				pushMaxY = true
			}
			if math.Abs(p.Y-f.core.MinY()) < frontierDelta {
				pushMinY = true
			}
		}
	}
	for _, p := range points {
		if rankdir == geom.RankdirLeftToRight {
			if p.X == f.core.MinX() && (p.Y == f.core.MinY() || p.Y == f.core.MaxY()) {
				pushMinX = false
			}
			if p.X == f.core.MaxX() && (p.Y == f.core.MinY() || p.Y == f.core.MaxY()) {
				pushMaxX = false
			}
		} else {
			if p.Y == f.core.MinY() && (p.X == f.core.MinX() || p.X == f.core.MaxX()) {
				pushMinY = false
			}
			if p.Y == f.core.MaxY() && (p.X == f.core.MinX() || p.X == f.core.MaxX()) {
				pushMaxY = false
			}
		}
	}
	if pushMaxX {
		f.core = f.core.AddMaxX(frontierDelta)
	}
	if pushMinX {
		f.core = f.core.AddMinX(-frontierDelta)
	}
	if pushMaxY {
		f.core = f.core.AddMaxY(frontierDelta)
	}
	if pushMinY {
		f.core = f.core.AddMinY(-frontierDelta)
	}
	return f
}

func (f *frontierCalculator) suggestedPosition() geom.RectangleArea { return f.core }

// ensureMinWidth widens the frontier around its center to minWidth,
// without going left of the initial box.
func (f *frontierCalculator) ensureMinWidth(minWidth float64) {
	delta := f.core.MaxX() - f.core.MinX() - minWidth
	if delta < 0 {
		newMinX := f.core.MinX() + delta/2
		newMaxX := f.core.MaxX() - delta/2
		if e := newMinX - f.initial.MinX(); e < 0 {
			newMinX -= e
			newMaxX -= e
		}
		f.core = f.core.WithMinX(newMinX)
		f.core = f.core.WithMaxX(newMaxX)
	}
}
