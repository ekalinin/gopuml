// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/SvekNode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cruise/XNode.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/jfmt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// SvekNode is a node of the DOT graph: the image of an entity, its dot
// name ("sh0002") and color, and its position once dot has laid it out.
type SvekNode struct {
	typ      klimt.ShapeType
	dimImage *geom.XDimension2D

	uid   string
	color int32

	minX, minY float64
	shieldVal  *klimt.Margins

	entityPosition abel.EntityPosition
	image          klimt.IEntityImage
	stringBounder  font.StringBounder

	cluster *Cluster
	leaf    *abel.Entity

	polygon *klimt.UPolygon
}

var _ geom.Positionable = (*SvekNode)(nil)

func newSvekNode(ent *abel.Entity, image klimt.IEntityImage, colors *colorSequence,
	stringBounder font.StringBounder) *SvekNode {
	n := &SvekNode{stringBounder: stringBounder, entityPosition: ent.EntityPosition(), image: image,
		typ: image.ShapeType(), leaf: ent}
	n.color = colors.value()
	n.uid = fmt.Sprintf("sh%04d", n.color)
	return n
}

// together returns the together block of the entity, nil for none.
func (n *SvekNode) together() *abel.Together {
	if n.leaf == nil {
		return nil
	}
	return n.leaf.Together()
}

func (n *SvekNode) dim() geom.XDimension2D {
	if n.dimImage == nil {
		d := n.image.CalculateDimension(n.stringBounder)
		n.dimImage = &d
	}
	return *n.dimImage
}

// Width returns the width of the image.
func (n *SvekNode) Width() float64 { return n.dim().Width() }

// Height returns the height of the image.
func (n *SvekNode) Height() float64 { return n.dim().Height() }

func (n *SvekNode) appendShape(sb *strings.Builder, stringBounder font.StringBounder) {
	switch {
	case n.typ == klimt.ShapeTypeRectangleHTMLForPorts:
		n.appendLabelHTMLSpecialForLink(sb, stringBounder)
		println(sb)
		return
	case n.typ == klimt.ShapeTypeRectanglePort:
		n.appendLabelHTMLSpecialForPort(sb, stringBounder)
		println(sb)
		return
	case n.typ == klimt.ShapeTypeRectangleWithCircleInside:
		n.appendHTML(sb)
		println(sb)
		return
	case n.typ == klimt.ShapeTypeRectangle && n.isShielded():
		n.appendHTML(sb)
		println(sb)
		return
	}
	sb.WriteString(n.uid)
	sb.WriteString(" [")
	n.appendShapeInternal(sb)
	sb.WriteString(",")
	sb.WriteString("label=\"\"")
	sb.WriteString(",")
	sb.WriteString("width=" + PixelToInches(n.Width()))
	sb.WriteString(",")
	sb.WriteString("height=" + PixelToInches(n.Height()))
	sb.WriteString(",")
	sb.WriteString("color=\"" + hexColor(n.color) + "\"")
	sb.WriteString("];")
	println(sb)
}

func (n *SvekNode) maxWidthFromLabelForEntryExit(stringBounder font.StringBounder) float64 {
	if im, ok := n.image.(interface {
		MaxWidthFromLabelForEntryExit(font.StringBounder) float64
	}); ok {
		return im.MaxWidthFromLabelForEntryExit(stringBounder)
	}
	throwUnsupported("")
	return 0
}

func (n *SvekNode) appendLabelHTMLSpecialForPort(sb *strings.Builder, stringBounder font.StringBounder) {
	width2 := klimt.JavaD2I(n.maxWidthFromLabelForEntryExit(stringBounder))
	if width2 > 40 {
		n.appendLabelHTMLSpecialForPortHTML(sb, width2-40)
	} else {
		n.appendLabelHTMLSpecialForPortBasic(sb)
	}
}

func (n *SvekNode) appendLabelHTMLSpecialForPortHTML(sb *strings.Builder, fullWidth int) {
	if fullWidth < 10 {
		fullWidth = 10
	}
	fw := strconv.Itoa(fullWidth)
	sb.WriteString(n.uid)
	sb.WriteString(" [")
	sb.WriteString("shape=plaintext")
	sb.WriteString(",")
	sb.WriteString("label=<")
	sb.WriteString("<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\" CELLPADDING=\"0\">")
	sb.WriteString("<TR><TD WIDTH=\"" + fw + "\" HEIGHT=\"1\" COLSPAN=\"3\"></TD></TR>")
	sb.WriteString("<TR><TD></TD><TD FIXEDSIZE=\"TRUE\" PORT=\"P\"  BORDER=\"1\" COLOR=\"" + hexColor(n.color) +
		"\" WIDTH=\"" + strconv.Itoa(klimt.JavaD2I(n.Width())) + "\" HEIGHT=\"" +
		strconv.Itoa(klimt.JavaD2I(n.Height())) + "\"></TD><TD></TD></TR>")
	sb.WriteString("<TR><TD WIDTH=\"" + fw + "\" HEIGHT=\"1\" COLSPAN=\"3\"></TD></TR>")
	sb.WriteString("</TABLE>")
	sb.WriteString(">];")
}

func (n *SvekNode) appendLabelHTMLSpecialForPortBasic(sb *strings.Builder) {
	sb.WriteString(n.uid)
	sb.WriteString(" [")
	sb.WriteString("shape=rect")
	sb.WriteString(",")
	sb.WriteString("label=\"\"")
	sb.WriteString(",")
	sb.WriteString("width=" + PixelToInches(n.Width()))
	sb.WriteString(",")
	sb.WriteString("height=" + PixelToInches(n.Height()))
	sb.WriteString(",")
	sb.WriteString("color=\"" + hexColor(n.color) + "\"")
	sb.WriteString("];")
}

func (n *SvekNode) shield() klimt.Margins {
	if n.shieldVal == nil {
		s := n.image.Shield(n.stringBounder)
		n.shieldVal = &s
		if !s.IsZero() && n.typ != klimt.ShapeTypeRectangle && n.typ != klimt.ShapeTypeRectangleHTMLForPorts &&
			n.typ != klimt.ShapeTypeRectangleWithCircleInside {
			throwIllegalState("")
		}
	}
	return *n.shieldVal
}

func (n *SvekNode) appendHTML(sb *strings.Builder) {
	sb.WriteString(n.uid)
	sb.WriteString(" [")
	sb.WriteString("shape=plaintext,")
	sb.WriteString("label=<")
	n.appendLabelHTML(sb)
	sb.WriteString(">")
	sb.WriteString("];")
	println(sb)
}

func (n *SvekNode) appendLabelHTML(sb *strings.Builder) {
	sb.WriteString("<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\" CELLPADDING=\"0\">")
	sb.WriteString("<TR>")
	appendTdEmpty(sb)
	appendTd(sb, 1, n.shield().Y1())
	appendTdEmpty(sb)
	sb.WriteString("</TR>")
	sb.WriteString("<TR>")
	appendTd(sb, n.shield().X1(), 1)
	sb.WriteString("<TD BGCOLOR=\"" + hexColor(n.color) + "\"")
	sb.WriteString(" FIXEDSIZE=\"TRUE\" WIDTH=\"" + jfmt.DoubleToString(n.Width()) + "\" HEIGHT=\"" +
		jfmt.DoubleToString(n.Height()) + "\"")
	sb.WriteString(" PORT=\"h\">")
	sb.WriteString("</TD>")
	appendTd(sb, n.shield().X2(), 1)
	sb.WriteString("</TR>")
	sb.WriteString("<TR>")
	appendTdEmpty(sb)
	appendTd(sb, 1, n.shield().Y2())
	appendTdEmpty(sb)
	sb.WriteString("</TR>")
	sb.WriteString("</TABLE>")
}

func (n *SvekNode) appendLabelHTMLSpecialForLink(sb *strings.Builder, stringBounder font.StringBounder) {
	ports := n.image.(klimt.WithPorts).Ports(stringBounder)

	sb.WriteString(n.uid)
	sb.WriteString(" [")
	sb.WriteString("shape=plaintext,")
	sb.WriteString("label=<")
	sb.WriteString("<TABLE BGCOLOR=\"" + hexColor(n.color) +
		"\" BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\" CELLPADDING=\"0\">")
	sum := 0
	for _, g := range ports.AllPortGeometry() {
		portID := g.ID()
		missing := klimt.JavaD2I(g.Position() - float64(sum))
		sum += missing
		n.appendTr(sb, "", missing)
		intHeight := klimt.JavaD2I(g.Height())
		n.appendTr(sb, portID, intHeight)
		sum += intHeight
	}
	diff := n.Height() - float64(sum)
	n.appendTr(sb, "", klimt.JavaD2I(diff))
	sb.WriteString("</TABLE>")
	sb.WriteString(">")
	sb.WriteString("];")
	println(sb)
}

func (n *SvekNode) appendTr(sb *strings.Builder, portID string, height int) {
	if height <= 0 {
		return
	}
	sb.WriteString("<TR>")
	sb.WriteString("<TD ")
	sb.WriteString(" FIXEDSIZE=\"TRUE\" WIDTH=\"" + jfmt.DoubleToString(n.Width()) + "\" HEIGHT=\"" +
		strconv.Itoa(height) + "\"")
	if portID != "" {
		sb.WriteString(" PORT=\"" + portID + "\"")
	}
	sb.WriteString(">")
	sb.WriteString("</TD>")
	sb.WriteString("</TR>")
}

func appendTd(sb *strings.Builder, w, h float64) {
	sb.WriteString("<TD")
	sb.WriteString(" FIXEDSIZE=\"TRUE\" WIDTH=\"" + jfmt.DoubleToString(w) + "\" HEIGHT=\"" +
		jfmt.DoubleToString(h) + "\"")
	sb.WriteString(">")
	sb.WriteString("</TD>")
}

func appendTdEmpty(sb *strings.Builder) {
	sb.WriteString("<TD>")
	sb.WriteString("</TD>")
}

func (n *SvekNode) appendShapeInternal(sb *strings.Builder) {
	switch n.typ {
	case klimt.ShapeTypeRectangle:
		if n.isShielded() {
			throwUnsupported("")
		}
		sb.WriteString("shape=rect")
	case klimt.ShapeTypeRectangleWithCircleInside, klimt.ShapeTypeFolder:
		sb.WriteString("shape=rect")
	case klimt.ShapeTypeRectangleHTMLForPorts:
		throwUnsupported("")
	case klimt.ShapeTypeOctagon:
		sb.WriteString("shape=octagon")
	case klimt.ShapeTypeHexagon:
		sb.WriteString("shape=hexagon")
	case klimt.ShapeTypeDiamond:
		sb.WriteString("shape=diamond")
	case klimt.ShapeTypeCircle:
		sb.WriteString("shape=circle")
	case klimt.ShapeTypeOval:
		sb.WriteString("shape=ellipse")
	case klimt.ShapeTypeRoundRectangle:
		sb.WriteString("shape=rect,style=rounded")
	default:
		throwIllegalState(n.typ.String())
	}
}

// UID returns the dot name of the node.
func (n *SvekNode) UID() string { return n.uid }

// MinX returns the left of the node once laid out.
func (n *SvekNode) MinX() float64 { return n.minX }

// MinY returns the top of the node once laid out.
func (n *SvekNode) MinY() float64 { return n.minY }

// Image returns the image of the entity.
func (n *SvekNode) Image() klimt.IEntityImage { return n.image }

// Position implements geom.Positionable.
func (n *SvekNode) Position() geom.XPoint2D { return geom.NewXPoint2D(n.minX, n.minY) }

// Size implements geom.Positionable.
func (n *SvekNode) Size() geom.XDimension2D { return n.dim() }

func (n *SvekNode) rectangleArea() geom.RectangleArea {
	return geom.NewRectangleArea(n.minX, n.minY, n.minX+n.Width(), n.minY+n.Height())
}

// isShielded reports whether the node gets extra room around it in the
// DOT graph: the target of a qualified association, or an image with a
// shield.
func (n *SvekNode) isShielded() bool {
	// Upstream avoids computing the shield before every edge is created
	// (issue 1467): a qualified association is enough.
	for _, link := range n.leaf.Diagram().Links() {
		if (n.leaf == link.Entity1() && link.HasKal1()) || (n.leaf == link.Entity2() && link.HasKal2()) {
			return true
		}
	}
	return !n.shield().IsZero()
}

// MoveDelta implements geom.Moveable.
func (n *SvekNode) MoveDelta(deltaX, deltaY float64) {
	n.minX += deltaX
	n.minY += deltaY
}

// IsHidden reports whether the image is hidden.
func (n *SvekNode) IsHidden() bool { return n.image.IsHidden() }

func (n *SvekNode) setPolygon(minX, minY float64, points []geom.XPoint2D) {
	n.polygon = klimt.NewUPolygonPoints(points).Translate(-minX, -minY)
}

func (n *SvekNode) point2D(x, y float64) geom.XPoint2D { return geom.NewXPoint2D(n.minX+x, n.minY+y) }

func (n *SvekNode) overscanX(stringBounder font.StringBounder) float64 {
	return n.image.OverscanX(stringBounder)
}

func (n *SvekNode) addImpact(angle float64) {
	n.image.(interface{ AddImpact(float64) }).AddImpact(angle)
}

func (n *SvekNode) drawKals(ug klimt.UGraphic) {
	for _, d := range []klimt.Direction{klimt.DirectionDown, klimt.DirectionUp, klimt.DirectionLeft,
		klimt.DirectionRight} {
		for _, kal := range n.leaf.Kals(d) {
			kal.(*Kal).DrawU(ug)
		}
	}
}

func (n *SvekNode) fixOverlap() {
	n.fixHoverlap(n.leaf.Kals(klimt.DirectionDown))
	n.fixHoverlap(n.leaf.Kals(klimt.DirectionUp))
}

func (n *SvekNode) fixHoverlap(list []abel.Kal) {
	var los lineOfSegments
	for _, k := range list {
		kal := k.(*Kal)
		los.addSegment(kal.x1(), kal.x2())
	}
	res := los.solveOverlaps()
	for i, k := range list {
		kal := k.(*Kal)
		kal.moveX(res[i] - kal.x1())
	}
}

// magneticBorder is the magnetic border of the image, at the position of
// the node.
func (n *SvekNode) magneticBorder() klimt.MagneticBorder {
	return klimt.MagneticBorderFunc(func(stringBounder font.StringBounder, position geom.XPoint2D) klimt.UTranslate {
		orig := n.image.MagneticBorder()
		return orig.ForceAt(stringBounder, position.Move(-n.minX, -n.minY))
	})
}
