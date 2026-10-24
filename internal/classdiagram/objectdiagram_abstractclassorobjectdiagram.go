// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/objectdiagram/AbstractClassOrObjectDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// AbstractClassOrObjectDiagram is upstream's base of class and object
// diagrams: the association classes, lollipops and legends of packages.
type AbstractClassOrObjectDiagram struct {
	abel.AbstractEntityDiagram

	associations []*association
}

func (d *AbstractClassOrObjectDiagram) initAbstractClassOrObjectDiagram(ctx *core.RenderContext,
	src *core.UmlSource, typ core.DiagramType, md5map map[string]string) {
	d.InitCucaDiagram(ctx, src, typ, md5map, nil)
	d.SetNamespaceSeparator(".", true)
}

func (d *AbstractClassOrObjectDiagram) newLink(location *text.LineLocation, cl1, cl2 *abel.Entity,
	typ decoration.LinkType, linkArg *abel.LinkArg) *abel.Link {
	return abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), cl1, cl2, typ, linkArg)
}

// InsertBetween replaces the last link between two entities by two links
// through node; it returns false when there is no such link.
func (d *AbstractClassOrObjectDiagram) InsertBetween(location *text.LineLocation, entity1, entity2,
	node *abel.Entity) bool {
	link := d.foundLink(entity1, entity2)
	if link == nil {
		return false
	}
	iconSize := d.SkinParam().ClassAttributeIconSize() > 0
	l1 := d.newLink(location, entity1, node, link.Type(),
		abel.LinkArgBuildManage(link.Label(), link.Length(), iconSize).
			WithQuantifier(link.Quantifier1(), abel.OptString{}).
			WithDistanceAngle(link.Labeldistance(), link.Labelangle()))
	l2 := d.newLink(location, node, entity2, link.Type(),
		abel.LinkArgBuildManage(link.Label(), link.Length(), iconSize).
			WithQuantifier(abel.OptString{}, link.Quantifier2()).
			WithDistanceAngle(link.Labeldistance(), link.Labelangle()))
	d.AddLink(l1)
	d.AddLink(l2)
	d.RemoveLink(link)
	return true
}

func (d *AbstractClassOrObjectDiagram) foundLink(entity1, entity2 *abel.Entity) *abel.Link {
	links := d.Links()
	for i := len(links) - 1; i >= 0; i-- {
		if links[i].IsBetween(entity1, entity2) {
			return links[i]
		}
	}
	return nil
}

// NbOfHozizontalLollipop returns the number of horizontal links from the
// entity to a lollipop. A lollipop panics, as upstream throws
// IllegalArgumentException.
func (d *AbstractClassOrObjectDiagram) NbOfHozizontalLollipop(entity *abel.Entity) int {
	if entity.LeafType() == abel.LeafTypeLollipopFull || entity.LeafType() == abel.LeafTypeLollipopHalf {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	result := 0
	for _, link := range d.Links() {
		if link.Length() == 1 && link.Contains(entity) &&
			(link.ContainsType(abel.LeafTypeLollipopFull) || link.ContainsType(abel.LeafTypeLollipopHalf)) {
			result++
		}
	}
	return result
}

// AssociationClass4 is the association between two association links,
// "(A, B) .. (C, D)": a link between two new points on the links.
func (d *AbstractClassOrObjectDiagram) AssociationClass4(location *text.LineLocation, entity1A, entity1B,
	entity2A, entity2B *abel.Entity, linkType decoration.LinkType, label *klimt.Display) command.Result {
	same1 := d.existingAssociatedPoints(entity1A, entity1B)
	same2 := d.existingAssociatedPoints(entity2A, entity2B)
	if len(same1) != 0 || len(same2) != 0 {
		return command.Error("Cannot link two associations points")
	}
	tmp1 := d.UniqueSequence("apoint")
	tmp2 := d.UniqueSequence("apoint")
	code1 := d.CurrentGroup().Quark().Child(tmp1)
	point1 := d.ReallyCreateLeaf(location, code1, klimt.DisplayWithNewlines(d.Pragma(), ""),
		abel.LeafTypePointForAssociation, abel.USymbolNull)
	code2 := d.CurrentGroup().Quark().Child(tmp2)
	point2 := d.ReallyCreateLeaf(location, code2, klimt.DisplayWithNewlines(d.Pragma(), ""),
		abel.LeafTypePointForAssociation, abel.USymbolNull)
	d.insertPointBetween(location, entity1A, entity1B, point1)
	d.insertPointBetween(location, entity2A, entity2B, point2)
	d.AddLink(d.newLink(location, point1, point2, linkType, abel.LinkArgBuild(label, 1)))
	return command.Ok()
}

func (d *AbstractClassOrObjectDiagram) insertPointBetween(location *text.LineLocation, entity1A, entity1B,
	point1 *abel.Entity) {
	existingLink1 := d.foundLink(entity1A, entity1B)
	if existingLink1 == nil {
		existingLink1 = d.newLink(location, entity1A, entity1B,
			decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone), abel.LinkArgNoDisplay(2))
	} else {
		d.RemoveLink(existingLink1)
	}
	entity1real, entity2real := existingLink1.Entity1(), existingLink1.Entity2()
	if existingLink1.IsInverted() {
		entity1real, entity2real = existingLink1.Entity2(), existingLink1.Entity1()
	}
	entity1ToPoint := d.newLink(location, entity1real, point1, existingLink1.Type().Part2(),
		abel.LinkArgBuild(existingLink1.Label(), existingLink1.Length()).
			WithQuantifier(existingLink1.Quantifier1(), abel.OptString{}).
			WithDistanceAngle(existingLink1.Labeldistance(), existingLink1.Labelangle()))
	entity1ToPoint.SetLinkArrow(existingLink1.LinkArrow())
	pointToEntity2 := d.newLink(location, point1, entity2real, existingLink1.Type().Part1(),
		abel.LinkArgNoDisplay(existingLink1.Length()).
			WithQuantifier(abel.OptString{}, existingLink1.Quantifier2()).
			WithDistanceAngle(existingLink1.Labeldistance(), existingLink1.Labelangle()))
	d.AddLink(entity1ToPoint)
	d.AddLink(pointToEntity2)
}

// AssociationClass is the association class "(A, B) .. C" (mode 1) or
// "C .. (A, B)" (mode 2): a point on the link between A and B, linked to
// C. It returns false when A and B already have two association classes.
func (d *AbstractClassOrObjectDiagram) AssociationClass(location *text.LineLocation, mode int, entity1,
	entity2, associed *abel.Entity, linkType decoration.LinkType, label *klimt.Display) bool {
	same := d.existingAssociatedPoints(entity1, entity2)
	switch {
	case len(same) > 1:
		return false
	case len(same) == 0:
		a := d.newAssociation(location, mode, entity1, entity2, associed)
		a.createNew(location, mode, linkType, label)
		d.associations = append(d.associations, a)
		return true
	}
	a := same[0].createSecondAssociation(location, mode, associed, label)
	a.createInSecond(location, linkType, label)
	d.associations = append(d.associations, a)
	return true
}

func (d *AbstractClassOrObjectDiagram) existingAssociatedPoints(entity1, entity2 *abel.Entity) []*association {
	var same []*association
	for _, existing := range d.associations {
		if existing.sameCouple(entity1, entity2) {
			same = append(same, existing)
		}
	}
	return same
}

// association is upstream's AbstractClassOrObjectDiagram.Association: an
// association class on the link between two entities.
type association struct {
	diagram  *AbstractClassOrObjectDiagram
	entity1  *abel.Entity
	entity2  *abel.Entity
	associed *abel.Entity
	point    *abel.Entity

	existingLink *abel.Link

	entity1ToPoint  *abel.Link
	pointToEntity2  *abel.Link
	pointToAssocied *abel.Link

	other *association
}

func (d *AbstractClassOrObjectDiagram) newAssociation(location *text.LineLocation, _ int, entity1, entity2,
	associed *abel.Entity) *association {
	a := &association{diagram: d, entity1: entity1, entity2: entity2, associed: associed}
	idShort := d.UniqueSequence("apoint")
	var quark *abel.Quark
	if entity1.Quark().Parent() == entity2.Quark().Parent() {
		quark = entity1.Quark().Parent().Child(idShort)
	} else {
		quark = d.QuarkInContext(true, d.CleanID(idShort))
	}
	a.point = d.ReallyCreateLeaf(location, quark, klimt.DisplayWithNewlines(d.Pragma(), ""),
		abel.LeafTypePointForAssociation, abel.USymbolNull)
	return a
}

func (a *association) createSecondAssociation(location *text.LineLocation, mode2 int, associed2 *abel.Entity,
	_ *klimt.Display) *association {
	result := a.diagram.newAssociation(location, mode2, a.entity1, a.entity2, associed2)
	result.existingLink = a.existingLink
	result.other = a
	if a.existingLink.Length() == 1 {
		a.entity1ToPoint.SetLength(2)
		a.pointToEntity2.SetLength(2)
		a.pointToAssocied.SetLength(1)
	}
	return result
}

func (a *association) createNew(location *text.LineLocation, mode int, linkType decoration.LinkType,
	label *klimt.Display) {
	d := a.diagram
	a.existingLink = d.foundLink(a.entity1, a.entity2)
	if a.existingLink == nil {
		a.existingLink = d.newLink(location, a.entity1, a.entity2,
			decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone), abel.LinkArgNoDisplay(2))
	} else {
		d.RemoveLink(a.existingLink)
	}
	existing := a.existingLink
	entity1real, entity2real := existing.Entity1(), existing.Entity2()
	a.entity1ToPoint = d.newLink(location, entity1real, a.point, existing.Type().Part2(),
		abel.LinkArgBuild(existing.Label(), existing.Length()).
			WithQuantifier(existing.Quantifier1(), abel.OptString{}).
			WithDistanceAngle(existing.Labeldistance(), existing.Labelangle()))
	a.entity1ToPoint.SetLinkArrow(existing.LinkArrow())
	a.pointToEntity2 = d.newLink(location, a.point, entity2real, existing.Type().Part1(),
		abel.LinkArgNoDisplay(existing.Length()).
			WithQuantifier(abel.OptString{}, existing.Quantifier2()).
			WithDistanceAngle(existing.Labeldistance(), existing.Labelangle()))
	length := 1
	if existing.Length() == 1 && a.entity1 != a.entity2 {
		length = 2
	}
	if existing.Length() == 2 && a.entity1 == a.entity2 {
		length = 2
	}
	if length == 1 {
		a.entity1ToPoint.AddNoteFrom(existing, abel.NoteLinkStrategyNormal)
	} else {
		a.entity1ToPoint.AddNoteFrom(existing, abel.NoteLinkStrategyHalfPrintedFull)
		a.pointToEntity2.AddNoteFrom(existing, abel.NoteLinkStrategyHalfNotPrinted)
	}
	d.AddLink(a.entity1ToPoint)
	d.AddLink(a.pointToEntity2)
	if mode == 1 {
		a.pointToAssocied = d.newLink(location, a.point, a.associed, linkType, abel.LinkArgBuild(label, length))
	} else {
		a.pointToAssocied = d.newLink(location, a.associed, a.point, linkType, abel.LinkArgBuild(label, length))
	}
	d.AddLink(a.pointToAssocied)
}

func (a *association) createInSecond(location *text.LineLocation, linkType decoration.LinkType,
	label *klimt.Display) {
	d := a.diagram
	a.existingLink = d.foundLink(a.entity1, a.entity2)
	if a.existingLink == nil {
		a.existingLink = d.newLink(location, a.entity1, a.entity2,
			decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone), abel.LinkArgNoDisplay(2))
	} else {
		d.RemoveLink(a.existingLink)
	}
	existing := a.existingLink
	a.entity1ToPoint = d.newLink(location, a.entity1, a.point, existing.Type().Part2(),
		abel.LinkArgBuild(existing.Label(), 2).WithQuantifier(existing.Quantifier1(), abel.OptString{}).
			WithDistanceAngle(existing.Labeldistance(), existing.Labelangle()))
	a.pointToEntity2 = d.newLink(location, a.point, a.entity2, existing.Type().Part1(),
		abel.LinkArgNoDisplay(2).WithQuantifier(abel.OptString{}, existing.Quantifier2()).
			WithDistanceAngle(existing.Labeldistance(), existing.Labelangle()))
	d.AddLink(a.entity1ToPoint)
	d.AddLink(a.pointToEntity2)
	if a.other.pointToAssocied.Entity1().LeafType() == abel.LeafTypePointForAssociation {
		d.RemoveLink(a.other.pointToAssocied)
		a.other.pointToAssocied = a.other.pointToAssocied.Inv()
		d.AddLink(a.other.pointToAssocied)
	}
	a.pointToAssocied = d.newLink(location, a.point, a.associed, linkType, abel.LinkArgBuild(label, 1))
	d.AddLink(a.pointToAssocied)
	lnode := d.newLink(location, a.other.point, a.point,
		decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone), abel.LinkArgNoDisplay(1))
	lnode.SetInvis(true)
	d.AddLink(lnode)
}

func (a *association) sameCouple(entity1, entity2 *abel.Entity) bool {
	return (a.entity1 == entity1 && a.entity2 == entity2) || (a.entity1 == entity2 && a.entity2 == entity1)
}

// SetLegend sets the legend of the current package, or of the diagram at
// the top level.
func (d *AbstractClassOrObjectDiagram) SetLegend(legend uml.DisplayPositioned) {
	currentGroup := d.CurrentGroup()
	if currentGroup.IsRoot() {
		d.Base.SetLegend(legend)
		return
	}
	currentGroup.SetLegend(legend)
}
