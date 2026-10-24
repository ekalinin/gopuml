// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/RoleLabels.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
	"github.com/ekalinin/gopuml/internal/text"
)

// The role labels of an association (A "1"/"-owner" --> B):
// roleLabelsCreate builds the block, with a leading visibility character
// turned into the icon of the class attributes, and roleLabelsPosition
// places it on the other side of the line than the quantifier, never on the
// line.
const (
	roleLabelsGap                     = 2
	roleLabelsMinDistanceForDirection = 15
	roleLabelsSearchStep              = 2
	roleLabelsSearchRange             = 30
)

// roleLabelsCreate is RoleLabels.create.
func roleLabelsCreate(role string, pragma *klimt.Pragma, font *klimt.FontConfiguration,
	skinParam klimt.ISkinParam) klimt.TextBlock {
	// A role is a property of the class, so it gets the icons of the fields
	// (hollow), as the attributes do, and not the ones of the methods (filled).
	var visibilityModifier klimt.VisibilityModifier
	hasVisibility := false
	txt := role
	if skinParam.ClassAttributeIconSize() > 0 {
		visibilityModifier, hasVisibility = klimt.VisibilityOf(role, true)
		if hasVisibility {
			txt = text.JavaTrim(role[1:])
		}
	}

	result := klimt.DisplayWithNewlines(pragma, txt).Create(font,
		skinParam.DefaultTextAlignment(geom.HorizontalAlignmentCenter), skinParam)

	if hasVisibility {
		fore := rose.NewRose().HtmlColor(skinParam, visibilityModifier.Foreground())
		icon := visibilityModifier.UBlock(skinParam.ClassAttributeIconSize(), fore, nil, false)
		icon = klimt.TextBlockUtilsWithMargin4(icon, 0, 1, 2, 0)
		result = klimt.TextBlockUtilsMergeLR(icon, result, geom.VerticalAlignmentCenter)
	}
	return result
}

// roleLabelsPosition is RoleLabels.getPosition: the top-left corner of the
// role. The role is preferably drawn on the other side of the line than the
// quantifier, at the same height (or at the same abscissa for a horizontal
// line). When this collides with one of the obstacles (typically the node
// the line is attached to: a line leaving a node diagonally can have its
// "other side" inside the node), the role is moved to the closest place
// with some room. qDim is the size of the quantifier, drawn at
// quantifierPos, rDim the size of the role; pathSamples are the points of
// the line, nil for the straight segment between the two endpoints.
func roleLabelsPosition(qDim, rDim geom.XDimension2D, quantifierPos, thisEndpoint, otherEndpoint geom.XPoint2D,
	pathSamples []geom.XPoint2D, obstacles []geom.RectangleArea) geom.XPoint2D {
	samples := pathSamples
	if samples == nil {
		samples = roleLabelsSampleSegment(thisEndpoint, otherEndpoint)
	}

	// Graphviz places the quantifier relatively to the direction of the line
	// where it leaves the node, not relatively to the direction from one end
	// to the other: a diagonal line can leave a node vertically and then bend.
	towards := roleLabelsPointAwayFrom(thisEndpoint, otherEndpoint, samples)
	dirX := towards.X - thisEndpoint.X
	dirY := towards.Y - thisEndpoint.Y

	if math.Abs(dirX)+math.Abs(dirY) < 0.001 {
		return geom.NewXPoint2D(quantifierPos.X, quantifierPos.Y+qDim.Height())
	}

	preferred := roleLabelsMirroredPosition(qDim, rDim, quantifierPos, thisEndpoint, dirX, dirY, samples)

	forbidden := append([]geom.RectangleArea{}, obstacles...)
	forbidden = append(forbidden, geom.NewRectangleArea(quantifierPos.X, quantifierPos.Y,
		quantifierPos.X+qDim.Width(), quantifierPos.Y+qDim.Height()))
	if roleLabelsIsFree(preferred, rDim, forbidden, nil) {
		return preferred
	}

	if alternative, ok := roleLabelsSearchFreePosition(qDim, rDim, quantifierPos, thisEndpoint, dirX, dirY,
		samples, forbidden); ok {
		return alternative
	}
	return preferred
}

func roleLabelsMirroredPosition(qDim, rDim geom.XDimension2D, quantifierPos, thisEndpoint geom.XPoint2D,
	dirX, dirY float64, samples []geom.XPoint2D) geom.XPoint2D {
	if math.Abs(dirY) >= math.Abs(dirX) {
		// Mostly vertical: mirror across the line X
		roleY := quantifierPos.Y
		qCenterX := quantifierPos.X + qDim.Width()/2
		lineX := thisEndpoint.X
		var roleX float64
		if qCenterX < lineX {
			roleX = lineX + roleLabelsGap
			// The line is not vertical: at the height of the label, it may be
			// further to the right than at its end.
			for _, pt := range samples {
				if pt.Y >= roleY-roleLabelsGap && pt.Y <= roleY+rDim.Height()+roleLabelsGap {
					roleX = jmath.Max(roleX, pt.X+roleLabelsGap)
				}
			}
		} else {
			roleX = lineX - rDim.Width() - roleLabelsGap
			for _, pt := range samples {
				if pt.Y >= roleY-roleLabelsGap && pt.Y <= roleY+rDim.Height()+roleLabelsGap {
					roleX = jmath.Min(roleX, pt.X-rDim.Width()-roleLabelsGap)
				}
			}
		}
		return geom.NewXPoint2D(roleX, roleY)
	}

	// Mostly horizontal: mirror across the line Y
	roleX := quantifierPos.X
	qCenterY := quantifierPos.Y + qDim.Height()/2
	lineY := thisEndpoint.Y
	var roleY float64
	if qCenterY < lineY {
		roleY = lineY + roleLabelsGap
		for _, pt := range samples {
			if pt.X >= roleX-roleLabelsGap && pt.X <= roleX+rDim.Width()+roleLabelsGap {
				roleY = jmath.Max(roleY, pt.Y+roleLabelsGap)
			}
		}
	} else {
		roleY = lineY - rDim.Height() - roleLabelsGap
		for _, pt := range samples {
			if pt.X >= roleX-roleLabelsGap && pt.X <= roleX+rDim.Width()+roleLabelsGap {
				roleY = jmath.Min(roleY, pt.Y-rDim.Height()-roleLabelsGap)
			}
		}
	}
	return geom.NewXPoint2D(roleX, roleY)
}

// roleLabelsSearchFreePosition looks for the free place the closest to the
// mirror image of the quantifier across the line, in the neighbourhood of
// the end of the line; ok is false when there is none.
func roleLabelsSearchFreePosition(qDim, rDim geom.XDimension2D, quantifierPos, endpoint geom.XPoint2D,
	dirX, dirY float64, samples []geom.XPoint2D, forbidden []geom.RectangleArea) (geom.XPoint2D, bool) {
	length := jmath.Hypot(dirX, dirY)
	tx := dirX / length
	ty := dirY / length

	qx := quantifierPos.X + qDim.Width()/2
	qy := quantifierPos.Y + qDim.Height()/2
	along := float64((qx-endpoint.X)*tx) + float64((qy-endpoint.Y)*ty)
	idealX := float64(2*(endpoint.X+float64(tx*along))) - qx
	idealY := float64(2*(endpoint.Y+float64(ty*along))) - qy

	var best geom.XPoint2D
	found := false
	bestDistance := math.MaxFloat64
	for x := endpoint.X - rDim.Width() - roleLabelsSearchRange; x <= endpoint.X+roleLabelsSearchRange; x += roleLabelsSearchStep {
		for y := endpoint.Y - rDim.Height() - roleLabelsSearchRange; y <= endpoint.Y+roleLabelsSearchRange; y += roleLabelsSearchStep {
			// Sliding along the line is preferred to moving away from it
			deltaX := x + rDim.Width()/2 - idealX
			deltaY := y + rDim.Height()/2 - idealY
			distance := math.Abs(float64(deltaX*tx)+float64(deltaY*ty)) +
				float64(2*math.Abs(float64(-deltaX*ty)+float64(deltaY*tx)))
			if distance >= bestDistance {
				continue
			}

			candidate := geom.NewXPoint2D(x, y)
			if roleLabelsIsFree(candidate, rDim, forbidden, samples) {
				best = candidate
				found = true
				bestDistance = distance
			}
		}
	}
	return best, found
}

func roleLabelsIsFree(pos geom.XPoint2D, dim geom.XDimension2D, forbidden []geom.RectangleArea,
	samples []geom.XPoint2D) bool {
	x1 := pos.X
	y1 := pos.Y
	x2 := x1 + dim.Width()
	y2 := y1 + dim.Height()
	for _, area := range forbidden {
		if x1 < area.MaxX() && x2 > area.MinX() && y1 < area.MaxY() && y2 > area.MinY() {
			return false
		}
	}
	for _, pt := range samples {
		if pt.X >= x1-roleLabelsGap && pt.X <= x2+roleLabelsGap && pt.Y >= y1-roleLabelsGap &&
			pt.Y <= y2+roleLabelsGap {
			return false
		}
	}
	return true
}

// roleLabelsPointAwayFrom returns the point of the line the closest to the
// endpoint, among the ones far enough from it to give a meaningful
// direction.
func roleLabelsPointAwayFrom(endpoint, otherEndpoint geom.XPoint2D, samples []geom.XPoint2D) geom.XPoint2D {
	result := otherEndpoint
	best := math.MaxFloat64
	for _, pt := range samples {
		distance := pt.Distance(endpoint)
		if distance >= roleLabelsMinDistanceForDirection && distance < best {
			best = distance
			result = pt
		}
	}
	return result
}

func roleLabelsSampleSegment(p1, p2 geom.XPoint2D) []geom.XPoint2D {
	var result []geom.XPoint2D
	steps := max(1, int(math.Ceil(p1.Distance(p2)/2)))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		result = append(result, geom.NewXPoint2D(p1.X+float64((p2.X-p1.X)*t), p1.Y+float64((p2.Y-p1.Y)*t)))
	}
	return result
}
