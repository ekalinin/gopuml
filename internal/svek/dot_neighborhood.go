// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/Neighborhood.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/EntityImageProtected.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"math"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// neighborhood is upstream's dot.Neighborhood: the links of a parent with
// many children ("skinparam groupInheritance"), drawn as one extends
// triangle per contact point of the shared tail, and the other links of the
// entity drawn up to its border.
type neighborhood struct {
	leaf            *abel.Entity
	sametailLinks   []*abel.Link
	allButSametails []*abel.Link
}

func newNeighborhood(leaf *abel.Entity, sametailLinks, all []*abel.Link) *neighborhood {
	n := &neighborhood{leaf: leaf, sametailLinks: sametailLinks}
	for _, l := range all {
		keep := true
		for _, s := range sametailLinks {
			if l == s {
				keep = false
				break
			}
		}
		if keep {
			n.allButSametails = append(n.allButSametails, l)
		}
	}
	return n
}

func (n *neighborhood) drawU(ug klimt.UGraphic, minX, minY float64, bibliotekon *Bibliotekon,
	shapeDim geom.XDimension2D) {
	// contactPoints is upstream's HashSet<XPoint2D>.
	var contactPoints []geom.XPoint2D
	var hashes []int32
	hasNull := false
	for _, link := range n.sametailLinks {
		contact, ok := bibliotekon.line(link).startContactPoint()
		if !ok {
			hasNull = true
			continue
		}
		dup := false
		for _, p := range contactPoints {
			if p.Equals(contact) {
				dup = true
				break
			}
		}
		if !dup {
			contactPoints = append(contactPoints, contact)
			hashes = append(hashes, contact.HashCode())
		}
	}
	if hasNull {
		// A null point in the HashSet: its intersection throws.
		throwNPE()
	}
	rect := geom.NewXRectangle2D(minX, minY, shapeDim.Width(), shapeDim.Height())
	center := geom.NewXPoint2D(rect.CenterX(), rect.CenterY())

	for _, i := range klimt.JavaHashOrder(hashes) {
		pt := contactPoints[i]
		inter, ok := rectIntersection(rect, center, pt)
		if !ok {
			continue
		}
		theta := jmath.Atan2(center.X-pt.X, -(center.Y - pt.Y))
		middle := drawExtends(ug, inter, theta)
		drawLine(ug, middle, pt)
	}
	for _, link := range n.allButSametails {
		line := bibliotekon.line(link)
		var contact geom.XPoint2D
		var ok bool
		if link.Entity1() == n.leaf {
			contact, ok = line.startContactPoint()
		} else {
			contact, ok = line.endContactPoint()
		}
		if !ok {
			continue
		}
		inter, ok := rectIntersection(rect, center, contact)
		if !ok {
			continue
		}
		drawLine(ug, inter, contact)
	}
}

func drawExtends(ug klimt.UGraphic, contact geom.XPoint2D, theta float64) geom.XPoint2D {
	poly := klimt.NewUPolygon()
	poly.AddPoint(0, 0)
	poly.AddPoint(7, 20)
	poly.AddPoint(-7, 20)
	poly.Rotate(theta)
	translate := klimt.UTranslatePoint(contact)
	ug.Apply(translate).Draw(poly)
	p1 := translate.Translated(poly.Points()[1])
	p2 := translate.Translated(poly.Points()[2])
	return geom.NewXPoint2D((p1.X+p2.X)/2, (p1.Y+p2.Y)/2)
}

func rectIntersection(rect geom.XRectangle2D, pt1, pt2 geom.XPoint2D) (geom.XPoint2D, bool) {
	sides := [][4]float64{
		{rect.MinX(), rect.MinY(), rect.MaxX(), rect.MinY()},
		{rect.MinX(), rect.MaxY(), rect.MaxX(), rect.MaxY()},
		{rect.MinX(), rect.MinY(), rect.MinX(), rect.MaxY()},
		{rect.MaxX(), rect.MinY(), rect.MaxX(), rect.MaxY()},
	}
	for _, s := range sides {
		if p, ok := segmentIntersection(s[0], s[1], s[2], s[3], pt1.X, pt1.Y, pt2.X, pt2.Y); ok {
			return p, true
		}
	}
	return geom.XPoint2D{}, false
}

const neighborhoodEpsilon = .001

func segmentIntersection(x1, y1, x2, y2, x3, y3, x4, y4 float64) (geom.XPoint2D, bool) {
	d := float64((x1-x2)*(y3-y4)) - float64((y1-y2)*(x3-x4))
	if d == 0 {
		return geom.XPoint2D{}, false
	}
	a := float64(x1*y2) - float64(y1*x2)
	b := float64(x3*y4) - float64(y3*x4)
	xi := (float64((x3-x4)*a) - float64((x1-x2)*b)) / d
	yi := (float64((y3-y4)*a) - float64((y1-y2)*b)) / d
	if xi+neighborhoodEpsilon < math.Min(x1, x2) || xi-neighborhoodEpsilon > math.Max(x1, x2) {
		return geom.XPoint2D{}, false
	}
	if xi+neighborhoodEpsilon < math.Min(x3, x4) || xi-neighborhoodEpsilon > math.Max(x3, x4) {
		return geom.XPoint2D{}, false
	}
	if yi+neighborhoodEpsilon < math.Min(y1, y2) || yi-neighborhoodEpsilon > math.Max(y1, y2) {
		return geom.XPoint2D{}, false
	}
	if yi+neighborhoodEpsilon < math.Min(y3, y4) || yi-neighborhoodEpsilon > math.Max(y3, y4) {
		return geom.XPoint2D{}, false
	}
	return geom.NewXPoint2D(xi, yi), true
}

func drawLine(ug klimt.UGraphic, pt1, pt2 geom.XPoint2D) {
	ug.Apply(klimt.NewUTranslate(pt1.X, pt1.Y)).Draw(klimt.NewULine(pt2.X-pt1.X, pt2.Y-pt1.Y))
}

// EntityImageProtected is the image of a parent with many children: the
// image with a border, and its neighborhood drawn over the links.
type EntityImageProtected struct {
	orig         klimt.IEntityImage
	border       float64
	bibliotekon  *Bibliotekon
	neighborhood *neighborhood
}

var _ klimt.IEntityImage = (*EntityImageProtected)(nil)

// InnerPosition returns the position of a member, inside the border.
func (p *EntityImageProtected) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	result := p.orig.InnerPosition(member, stringBounder)
	if result == nil {
		throwNPE()
	}
	r := geom.NewXRectangle2D(result.MinX()+p.border, result.MinY()+p.border, result.Width, result.Height)
	return &r
}

// IsHidden reports whether the entity is hidden.
func (p *EntityImageProtected) IsHidden() bool { return p.orig.IsHidden() }

// Backcolor returns the background of the image.
func (p *EntityImageProtected) Backcolor() color.HColor { return p.orig.Backcolor() }

// CalculateDimension returns the size of the image with its border.
func (p *EntityImageProtected) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return p.orig.CalculateDimension(stringBounder).Delta(2 * p.border)
}

// MinMax panics, as upstream's default.
func (p *EntityImageProtected) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return klimt.TextBlockDefaults{}.MinMax(stringBounder)
}

// MagneticBorder returns no force.
func (p *EntityImageProtected) MagneticBorder() klimt.MagneticBorder {
	return klimt.MagneticBorderNone{}
}

// DrawU draws the image inside the border.
func (p *EntityImageProtected) DrawU(ug klimt.UGraphic) {
	p.orig.DrawU(ug.Apply(klimt.NewUTranslate(p.border, p.border)))
}

func (p *EntityImageProtected) drawUntranslated(ug klimt.UGraphic, minX, minY float64) {
	dim := p.orig.CalculateDimension(ug.StringBounder())
	p.neighborhood.drawU(ug, minX+p.border, minY+p.border, p.bibliotekon, dim)
}

// ShapeType returns the shape of the image.
func (p *EntityImageProtected) ShapeType() klimt.ShapeType { return p.orig.ShapeType() }

// Shield returns the shield of the image.
func (p *EntityImageProtected) Shield(stringBounder font.StringBounder) klimt.Margins {
	return p.orig.Shield(stringBounder)
}

// OverscanX returns the overscan of the image.
func (p *EntityImageProtected) OverscanX(stringBounder font.StringBounder) float64 {
	return p.orig.OverscanX(stringBounder)
}

// IsCrash returns false.
func (p *EntityImageProtected) IsCrash() bool { return false }

// Ports returns the ports of the image.
func (p *EntityImageProtected) Ports(stringBounder font.StringBounder) *klimt.Ports {
	return p.orig.(klimt.WithPorts).Ports(stringBounder)
}
