// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/DotPath.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// DotPathTriPoints is the three points of a cubic segment after its start.
type DotPathTriPoints struct {
	X1, Y1, X2, Y2, X, Y float64
}

// NewDotPathTriPoints returns the points.
func NewDotPathTriPoints(p1, p2, p geom.XPoint2D) DotPathTriPoints {
	return DotPathTriPoints{X1: p1.X, Y1: p1.Y, X2: p2.X, Y2: p2.Y, X: p.X, Y: p.Y}
}

// DotPath is a sequence of cubic Bezier curves, as Graphviz returns edges.
//
// As upstream, whose curves are mutable objects, the paths built from
// another (AddCurve, AddBefore, AddAfter, SimulateCompound) share its curves:
// MoveStartPoint and MoveEndPoint change the shared first or last curve.
// Copy makes a deep copy.
type DotPath struct {
	beziers  []*geom.XCubicCurve2D
	comment  string
	codeLine string
}

var _ geom.Moveable = (*DotPath)(nil)

// NewDotPath returns an empty path.
func NewDotPath() *DotPath { return &DotPath{} }

// Copy returns a deep copy of the curves, without comment.
func (p *DotPath) Copy() *DotPath {
	result := &DotPath{}
	for _, c := range p.beziers {
		cc := *c
		result.beziers = append(result.beziers, &cc)
	}
	return result
}

// DotPathFromBeziers returns the path of curves.
func DotPathFromBeziers(beziers []geom.XCubicCurve2D) *DotPath {
	if beziers == nil {
		throw(nullPointer, "")
	}
	result := &DotPath{}
	for _, c := range beziers {
		cc := c
		result.beziers = append(result.beziers, &cc)
	}
	return result
}

// dotPathShared returns the path of curves, sharing them.
func dotPathShared(beziers []*geom.XCubicCurve2D) *DotPath { return &DotPath{beziers: beziers} }

func newCurve(c geom.XCubicCurve2D) *geom.XCubicCurve2D { return &c }

// AddCurve4 returns the path with one more curve.
func (p *DotPath) AddCurve4(pt1, pt2, pt3, pt4 geom.XPoint2D) *DotPath {
	beziersNew := slices.Clone(p.beziers)
	beziersNew = append(beziersNew, newCurve(geom.NewXCubicCurve2D(pt1.X, pt1.Y, pt2.X, pt2.Y, pt3.X, pt3.Y,
		pt4.X, pt4.Y)))
	return dotPathShared(beziersNew)
}

// AddCurve returns the path with one more curve starting at its end.
func (p *DotPath) AddCurve(pt2, pt3, pt4 geom.XPoint2D) *DotPath {
	last := p.beziers[len(p.beziers)-1]
	return p.AddCurve4(last.P2(), pt2, pt3, pt4)
}

// StartPoint returns the first point.
func (p *DotPath) StartPoint() geom.XPoint2D { return p.beziers[0].P1() }

// Sample returns the control points of the curves subdivided until flat,
// without duplicates, in the iteration order of upstream's HashSet.
func (p *DotPath) Sample() []geom.XPoint2D {
	seen := map[geom.XPoint2D]bool{}
	var found []geom.XPoint2D
	for _, bez := range p.beziers {
		sampleCurve(*bez, seen, &found)
	}
	hashes := make([]int32, len(found))
	for i, pt := range found {
		hashes[i] = pt.HashCode()
	}
	result := make([]geom.XPoint2D, len(found))
	for i, idx := range javaHashOrder(hashes) {
		result[i] = found[idx]
	}
	return result
}

func sampleCurve(bez geom.XCubicCurve2D, seen map[geom.XPoint2D]bool, result *[]geom.XPoint2D) {
	p1 := bez.CtrlP1()
	p2 := bez.CtrlP2()
	if bez.FlatnessSq() > 0.5 || p1.Distance(p2) > 4 {
		left, right := bez.Subdivide()
		sampleCurve(left, seen, result)
		sampleCurve(right, seen, result)
		return
	}
	for _, pt := range []geom.XPoint2D{p1, p2} {
		if !seen[pt] {
			seen[pt] = true
			*result = append(*result, pt)
		}
	}
}

// Middle returns the point of the curves (and the angle there) that is the
// closest to both ends.
func (p *DotPath) Middle() geom.PointAndAngle {
	var result *geom.XPoint2D
	angle := 0.0
	for _, bez := range p.beziers {
		left, right := bez.Subdivide()
		p1 := left.P1()
		p2 := left.P2()
		p3 := right.P1()
		p4 := right.P2()
		if result == nil || p.cost(p1) < p.cost(*result) {
			result = &p1
			angle = geom.BezierUtilsStartingAngle(left)
		}
		if p.cost(p2) < p.cost(*result) {
			result = &p2
			angle = geom.BezierUtilsEndingAngle(left)
		}
		if p.cost(p3) < p.cost(*result) {
			result = &p3
			angle = geom.BezierUtilsStartingAngle(right)
		}
		if p.cost(p4) < p.cost(*result) {
			result = &p4
			angle = geom.BezierUtilsEndingAngle(right)
		}
	}
	if result == nil {
		throw(nullPointer, "")
	}
	return geom.NewPointAndAngle(*result, angle)
}

func (p *DotPath) cost(pt geom.XPoint2D) float64 {
	start := p.StartPoint()
	end := p.EndPoint()
	return pt.DistanceSq(start) + pt.DistanceSq(end)
}

// MoveStartPointBy moves the start point by move.
func (p *DotPath) MoveStartPointBy(move UTranslate) { p.MoveStartPoint(move.Dx(), move.Dy()) }

// MoveEndPointBy moves the end point by move.
func (p *DotPath) MoveEndPointBy(move UTranslate) { p.MoveEndPoint(move.Dx(), move.Dy()) }

// MoveStartPoint moves the start point (and its control point) by (dx, dy),
// dropping the first curve when the move is longer than it.
func (p *DotPath) MoveStartPoint(dx, dy float64) {
	if len(p.beziers) > 1 && math.Sqrt(float64(dx*dx)+float64(dy*dy)) >= p.beziers[0].Length() {
		dx -= p.beziers[1].X1 - p.beziers[0].X1
		dy -= p.beziers[1].Y1 - p.beziers[0].Y1
		p.beziers = slices.Delete(p.beziers, 0, 1)
	}
	p.beziers[0].X1 += dx
	p.beziers[0].Y1 += dy
	p.beziers[0].CtrlX1 += dx
	p.beziers[0].CtrlY1 += dy
}

// EndPoint returns the last point.
func (p *DotPath) EndPoint() geom.XPoint2D { return p.beziers[len(p.beziers)-1].P2() }

// MoveEndPoint moves the end point (and its control point) by (dx, dy).
func (p *DotPath) MoveEndPoint(dx, dy float64) {
	last := len(p.beziers) - 1
	p.beziers[last].X2 += dx
	p.beziers[last].Y2 += dy
	p.beziers[last].CtrlX2 += dx
	p.beziers[last].CtrlY2 += dy
}

// MinFinder returns the smallest coordinates of all the points.
func (p *DotPath) MinFinder() *geom.MinFinder {
	result := geom.NewMinFinder()
	for _, c := range p.beziers {
		result.Manage(c.X1, c.Y1)
		result.Manage(c.X2, c.Y2)
		result.Manage(c.CtrlX1, c.CtrlY1)
		result.Manage(c.CtrlX2, c.CtrlY2)
	}
	return result
}

// MinMax returns the bounds of all the points, control points included.
func (p *DotPath) MinMax() geom.MinMax {
	result := geom.MinMaxEmpty(false)
	for _, c := range p.beziers {
		result = result.AddPointXY(c.X1, c.Y1)
		result = result.AddPointXY(c.X2, c.Y2)
		result = result.AddPointXY(c.CtrlX1, c.CtrlY1)
		result = result.AddPointXY(c.CtrlX2, c.CtrlY2)
	}
	return result
}

// MinDist returns the distance from ref to the closest point.
func (p *DotPath) MinDist(ref geom.XPoint2D) float64 {
	result := math.MaxFloat64
	for _, c := range p.beziers {
		for _, d := range []float64{ref.DistanceXY(c.X1, c.Y1), ref.DistanceXY(c.X2, c.Y2),
			ref.DistanceXY(c.CtrlX1, c.CtrlY1), ref.DistanceXY(c.CtrlX2, c.CtrlY2)} {
			if d < result {
				result = d
			}
		}
	}
	return result
}

func (p *DotPath) endTangeante() geom.XLine2D {
	last := p.beziers[len(p.beziers)-1]
	dx := last.X2 - last.CtrlX2
	dy := last.Y2 - last.CtrlY2
	if dx == 0 && dy == 0 {
		dx = last.X2 - last.X1
		dy = last.Y2 - last.Y1
	}
	return geom.NewXLine2D(last.X2, last.Y2, last.X2+dx, last.Y2+dy)
}

// EndAngle returns the angle of the path at its end.
func (p *DotPath) EndAngle() float64 {
	tan := p.endTangeante()
	return jmath.Atan2(tan.Y2-tan.Y1, tan.X2-tan.X1)
}

// StartAngle returns the angle of the path at its start.
func (p *DotPath) StartAngle() float64 {
	tan := p.startTangeante()
	return jmath.Atan2(tan.Y2-tan.Y1, tan.X2-tan.X1)
}

func (p *DotPath) startTangeante() geom.XLine2D {
	first := p.beziers[0]
	dx := first.CtrlX1 - first.X1
	dy := first.CtrlY1 - first.Y1
	if dx == 0 && dy == 0 {
		dx = first.X2 - first.X1
		dy = first.Y2 - first.Y1
	}
	return geom.NewXLine2D(first.X1, first.Y1, first.X1+dx, first.Y1+dy)
}

// AddBefore returns the path with a curve before its first.
func (p *DotPath) AddBefore(before geom.XCubicCurve2D) *DotPath {
	cp := append([]*geom.XCubicCurve2D{newCurve(before)}, p.beziers...)
	return dotPathShared(cp)
}

// AddAfter returns the path with a curve after its last.
func (p *DotPath) AddAfter(after geom.XCubicCurve2D) *DotPath {
	cp := append(slices.Clone(p.beziers), newCurve(after))
	return dotPathShared(cp)
}

// AddAfterPath returns the path followed by the curves of other.
func (p *DotPath) AddAfterPath(other *DotPath) *DotPath {
	cp := append(slices.Clone(p.beziers), other.beziers...)
	return dotPathShared(cp)
}

// ManageEnsureVisible reports the ends of every curve, moved by (x, y).
func (p *DotPath) ManageEnsureVisible(x, y float64, visible geom.EnsureVisible) {
	for _, bez := range p.beziers {
		visible.EnsureVisible(x+bez.X1, y+bez.Y1)
		visible.EnsureVisible(x+bez.X2, y+bez.Y2)
	}
}

// ToUPath returns the path as a UPath: a move then cubic segments.
func (p *DotPath) ToUPath() *UPath {
	result := NewUPath(p.comment, p.codeLine)
	start := true
	for _, bez := range p.beziers {
		if start {
			result.Add([]float64{bez.X1, bez.Y1}, geom.SegMoveTo)
			start = false
		}
		result.Add([]float64{bez.CtrlX1, bez.CtrlY1, bez.CtrlX2, bez.CtrlY2, bez.X2, bez.Y2}, geom.SegCubicTo)
	}
	return result
}

// DotPathCurveString is upstream's toString(XCubicCurve2D).
func DotPathCurveString(c geom.XCubicCurve2D) string {
	return "(" + jstr(c.X1) + "," + jstr(c.Y1) + ") " + "(" + jstr(c.CtrlX1) + "," + jstr(c.CtrlY1) + ") " +
		"(" + jstr(c.CtrlX2) + "," + jstr(c.CtrlY2) + ") " + "(" + jstr(c.X2) + "," + jstr(c.Y2) + ") "
}

// String is upstream's toString.
func (p *DotPath) String() string {
	var sb strings.Builder
	for _, c := range p.beziers {
		sb.WriteString(DotPathCurveString(*c))
		sb.WriteString(" - ")
	}
	return sb.String()
}

// DotPathReverseCurve returns the curve in the other direction.
func DotPathReverseCurve(curv geom.XCubicCurve2D) geom.XCubicCurve2D {
	return geom.NewXCubicCurve2D(curv.X2, curv.Y2, curv.CtrlX2, curv.CtrlY2, curv.CtrlX1, curv.CtrlY1, curv.X1,
		curv.Y1)
}

// Reverse returns the path in the other direction.
func (p *DotPath) Reverse() *DotPath {
	cp := make([]*geom.XCubicCurve2D, 0, len(p.beziers))
	for i := len(p.beziers) - 1; i >= 0; i-- {
		cp = append(cp, newCurve(DotPathReverseCurve(*p.beziers[i])))
	}
	return dotPathShared(cp)
}

// MoveDelta moves every point by (deltaX, deltaY), with new curves.
func (p *DotPath) MoveDelta(deltaX, deltaY float64) {
	for i, c := range p.beziers {
		p.beziers[i] = newCurve(geom.NewXCubicCurve2D(c.X1+deltaX, c.Y1+deltaY, c.CtrlX1+deltaX, c.CtrlY1+deltaY,
			c.CtrlX2+deltaX, c.CtrlY2+deltaY, c.X2+deltaX, c.Y2+deltaY))
	}
}

// Beziers returns a copy of the curves.
func (p *DotPath) Beziers() []geom.XCubicCurve2D {
	result := make([]geom.XCubicCurve2D, len(p.beziers))
	for i, c := range p.beziers {
		result[i] = *c
	}
	return result
}

// SimulateCompound cuts the parts of the path inside the tail and head
// boxes of a compound edge (nil for none).
func (p *DotPath) SimulateCompound(head, tail *geom.RectangleArea) *DotPath {
	if head == nil && tail == nil {
		return p
	}
	me := p
	if tail != nil && tail.ContainsPoint(p.StartPoint()) {
		result := NewDotPath()
		idx := 0
		for idx+1 < len(p.beziers) && tail.ContainsPoint(p.beziers[idx].P2()) {
			if !tail.ContainsPoint(p.beziers[idx].P1()) {
				throw(illegalState, "")
			}
			idx++
		}
		if !tail.ContainsPoint(p.beziers[idx].P2()) {
			current := *p.beziers[idx]
			for k := 0; k < 8; k++ {
				part1, part2 := current.Subdivide()
				if tail.ContainsPoint(part1.P2()) {
					current = part2
				} else {
					result.beziers = append([]*geom.XCubicCurve2D{newCurve(part2)}, result.beziers...)
					current = part1
				}
			}
			result.beziers = append(result.beziers, p.beziers[idx+1:]...)
			me = result
		}
	}
	if head != nil {
		result := NewDotPath()
		if head.ContainsPoint(p.EndPoint()) {
			for _, shared := range me.beziers {
				if !head.ContainsPoint(shared.P2()) {
					result.beziers = append(result.beziers, shared)
					continue
				}
				if head.ContainsPoint(shared.P1()) {
					return me
				}
				current := *shared
				for k := 0; k < 8; k++ {
					part1, part2 := current.Subdivide()
					if head.ContainsPoint(part1.P2()) {
						current = part1
					} else {
						result.beziers = append(result.beziers, newCurve(part1))
						current = part2
					}
				}
				return result
			}
		}
	}
	return me
}

// IsLine reports whether every curve is flat.
func (p *DotPath) IsLine() bool {
	for _, curve := range p.beziers {
		if curve.FlatnessSq() > 0.001 {
			return false
		}
	}
	return true
}

// SetCommentAndCodeLine sets the comment and code line written in the SVG.
func (p *DotPath) SetCommentAndCodeLine(comment, codeLine string) {
	p.comment = comment
	p.codeLine = codeLine
}

// MuteToRoundOrthogonalPaths replaces the right-angle corners of an
// orthogonal path by curves of the given radius.
func (p *DotPath) MuteToRoundOrthogonalPaths(radius float64) {
	if radius <= 0 || len(p.beziers) < 2 {
		return
	}
	if !p.isOrthogonalPathInternal() {
		return
	}
	var points []geom.XPoint2D
	if len(p.beziers) > 0 {
		points = append(points, p.beziers[0].P1())
		for _, bez := range p.beziers {
			points = append(points, bez.P2())
		}
	}
	if len(points) < 3 {
		return
	}
	corners := detectCorners(points)
	if len(corners) == 0 {
		return
	}
	newPoints := buildRoundedPath(points, corners, radius)
	p.beziers = nil
	for _, c := range buildBeziersFromPoints(newPoints, points, corners) {
		p.beziers = append(p.beziers, newCurve(c))
	}
}

func detectCorners(points []geom.XPoint2D) []int {
	var corners []int
	for i := 1; i < len(points)-1; i++ {
		prev := points[i-1]
		curr := points[i]
		next := points[i+1]
		dx1 := curr.X - prev.X
		dy1 := curr.Y - prev.Y
		dx2 := next.X - curr.X
		dy2 := next.Y - curr.Y
		len1 := math.Sqrt(float64(dx1*dx1) + float64(dy1*dy1))
		len2 := math.Sqrt(float64(dx2*dx2) + float64(dy2*dy2))
		if len1 < 0.01 || len2 < 0.01 {
			continue
		}
		dx1 /= len1
		dy1 /= len1
		dx2 /= len2
		dy2 /= len2
		dotProduct := float64(dx1*dx2) + float64(dy1*dy2)
		if math.Abs(dotProduct) < 0.1 {
			corners = append(corners, i)
		}
	}
	return corners
}

func buildRoundedPath(points []geom.XPoint2D, corners []int, radius float64) []geom.XPoint2D {
	newPoints := []geom.XPoint2D{points[0]}
	for i := 1; i < len(points)-1; i++ {
		if !slices.Contains(corners, i) {
			newPoints = append(newPoints, points[i])
			continue
		}
		prev := points[i-1]
		curr := points[i]
		next := points[i+1]
		dx1 := curr.X - prev.X
		dy1 := curr.Y - prev.Y
		dx2 := next.X - curr.X
		dy2 := next.Y - curr.Y
		len1 := math.Sqrt(float64(dx1*dx1) + float64(dy1*dy1))
		len2 := math.Sqrt(float64(dx2*dx2) + float64(dy2*dy2))
		dx1 /= len1
		dy1 /= len1
		dx2 /= len2
		dy2 /= len2
		truncDist1 := math.Min(radius, len1/2)
		truncDist2 := math.Min(radius, len2/2)
		newPoints = append(newPoints, geom.NewXPoint2D(curr.X-float64(dx1*truncDist1), curr.Y-float64(dy1*truncDist1)))
		newPoints = append(newPoints, curr)
		newPoints = append(newPoints, geom.NewXPoint2D(curr.X+float64(dx2*truncDist2), curr.Y+float64(dy2*truncDist2)))
	}
	newPoints = append(newPoints, points[len(points)-1])
	return newPoints
}

func buildBeziersFromPoints(newPoints, originalPoints []geom.XPoint2D, corners []int) []geom.XCubicCurve2D {
	var newBeziers []geom.XCubicCurve2D
	for i := 0; i < len(newPoints)-1; i++ {
		p1 := newPoints[i]
		p2 := newPoints[i+1]
		if i+2 < len(newPoints) {
			p3 := newPoints[i+2]
			isCornerControl := false
			for _, cornerIdx := range corners {
				if originalPoints[cornerIdx].Equals(p2) {
					isCornerControl = true
					break
				}
			}
			if isCornerControl {
				dx1 := p2.X - p1.X
				dy1 := p2.Y - p1.Y
				dx2 := p3.X - p2.X
				dy2 := p3.Y - p2.Y
				len1 := math.Sqrt(float64(dx1*dx1) + float64(dy1*dy1))
				len2 := math.Sqrt(float64(dx2*dx2) + float64(dy2*dy2))
				controlDist := math.Min(len1, len2) * 0.55228
				ctrl1X := p1.X + float64((dx1/len1)*controlDist)
				ctrl1Y := p1.Y + float64((dy1/len1)*controlDist)
				ctrl2X := p3.X - float64((dx2/len2)*controlDist)
				ctrl2Y := p3.Y - float64((dy2/len2)*controlDist)
				newBeziers = append(newBeziers, geom.NewXCubicCurve2D(p1.X, p1.Y, ctrl1X, ctrl1Y, ctrl2X, ctrl2Y,
					p3.X, p3.Y))
				i++ // Skip the control point and the next point
				continue
			}
		}
		newBeziers = append(newBeziers, geom.NewXCubicCurve2D(p1.X, p1.Y, p1.X, p1.Y, p2.X, p2.Y, p2.X, p2.Y))
	}
	return newBeziers
}

func (p *DotPath) isOrthogonalPathInternal() bool {
	for _, curve := range p.beziers {
		dx := math.Abs(curve.X2 - curve.X1)
		dy := math.Abs(curve.Y2 - curve.Y1)
		if dx > 0.1 && dy > 0.1 {
			return false
		}
	}
	return true
}
