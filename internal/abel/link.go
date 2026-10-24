// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/Link.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Link is a relation between two entities, with its type, label, end
// texts, note and drawing options.
type Link struct {
	decoration.WithLinkType

	cl1, cl2     *Entity
	port1, port2 OptString
	linkArg      *LinkArg
	uid          string
	note         *CucaNote

	invis              bool
	weight             float64
	constraint         bool
	inverted           bool
	linkArrow          LinkArrow
	opale              bool
	horizontalSolitary bool
	sametail           OptString
	styleBuilder       *klimt.StyleBuilder
	stereotype         *klimt.Stereotype
	cucaDiagram        *CucaDiagram
	location           *text.LineLocation
	url                *klimt.URL
	linkConstraint     *LinkConstraint
	codeLine           *text.LineLocation
}

// NewLink returns a link from cl1 to cl2. A length below 1 panics, as
// upstream throws IllegalArgumentException; the link gets the next uid of
// the diagram.
func NewLink(location *text.LineLocation, cucaDiagram *CucaDiagram, styleBuilder *klimt.StyleBuilder,
	cl1, cl2 *Entity, typ decoration.LinkType, linkArg *LinkArg) *Link {
	if linkArg.Length() < 1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if cl1 == nil || cl2 == nil {
		uml.ThrowNPE("")
	}
	l := &Link{location: location, cucaDiagram: cucaDiagram, styleBuilder: styleBuilder, cl1: cl1, cl2: cl2,
		linkArg: linkArg, weight: 1.0, constraint: true, linkArrow: LinkArrowNoneOrSeveral}
	l.WithLinkType = decoration.InitWithLinkType(typ, l.GoNorank)
	l.uid = cucaDiagram.UniqueSequence("lnk")
	if cucaDiagram.Pragma().IsTrue(klimt.PragmaKermor) {
		if !cl1.EntityPosition().IsNormal() != !cl2.EntityPosition().IsNormal() {
			l.SetConstraint(false)
		}
	}
	return l
}

// StyleBuilder returns the style sheet active when the link was created.
func (l *Link) StyleBuilder() *klimt.StyleBuilder { return l.styleBuilder }

// LinkStrategy returns how the decorations are drawn: SIMPLEST.
func (l *Link) LinkStrategy() decoration.LinkStrategy { return decoration.LinkStrategySimplest }

// IDCommentForSvg returns the id of the link in SVG comments.
func (l *Link) IDCommentForSvg() string {
	if l.LinkType.LooksLikeRevertedForSvg() {
		return l.cl1.Name() + "-backto-" + l.cl2.Name()
	}
	if l.LinkType.LooksLikeNoDecorAtAllSvg() {
		return l.cl1.Name() + "-" + l.cl2.Name()
	}
	return l.cl1.Name() + "-to-" + l.cl2.Name()
}

// CommentForSvg returns the text of the SVG comment of the link.
func (l *Link) CommentForSvg() string {
	if l.LinkType.LooksLikeRevertedForSvg() {
		return "reverse link " + l.cl1.Name() + " to " + l.cl2.Name()
	}
	return "link " + l.cl1.Name() + " to " + l.cl2.Name()
}

// Inv returns the link in the other direction, with the next uid.
func (l *Link) Inv() *Link {
	result := NewLink(l.location, l.cucaDiagram, l.styleBuilder, l.cl2, l.cl1, l.Type().Inversed(), l.linkArg.Inv())
	result.inverted = !l.inverted
	result.port1 = l.port2
	result.port2 = l.port1
	result.url = l.url
	result.linkConstraint = l.linkConstraint
	result.stereotype = l.stereotype
	result.linkArg.SetVisibilityModifier(l.linkArg.VisibilityModifier())
	result.linkArrow = l.linkArrow
	return result
}

// GoNorank is upstream's goNorank: the link does not rank its ends.
func (l *Link) GoNorank() { l.SetConstraint(false) }

// Labeldistance returns the label distance.
func (l *Link) Labeldistance() OptString { return l.linkArg.Labeldistance() }

// Labelangle returns the label angle.
func (l *Link) Labelangle() OptString { return l.linkArg.Labelangle() }

// UID returns the uid, for example "lnk4".
func (l *Link) UID() string { return l.uid }

// IsInvis reports whether the line is invisible.
func (l *Link) IsInvis() bool {
	if l.LinkType.IsInvisible() {
		return true
	}
	return l.invis
}

// SetInvis makes the line invisible.
func (l *Link) SetInvis(invis bool) { l.invis = invis }

// IsBetween reports whether the link joins two entities, in any order.
func (l *Link) IsBetween(cl1, cl2 *Entity) bool {
	return (cl1 == l.cl1 && cl2 == l.cl2) || (cl1 == l.cl2 && cl2 == l.cl1)
}

// Entity1 returns the first end.
func (l *Link) Entity1() *Entity { return l.cl1 }

// Entity2 returns the second end.
func (l *Link) Entity2() *Entity { return l.cl2 }

// PortName1 returns the port of the first end.
func (l *Link) PortName1() OptString { return l.port1 }

// PortName2 returns the port of the second end.
func (l *Link) PortName2() OptString { return l.port2 }

// Type returns the type: without decorations for an opale link or a link
// with a same tail.
func (l *Link) Type() decoration.LinkType {
	if l.opale {
		return decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone)
	}
	if l.sametail.OK {
		return decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone)
	}
	return l.LinkType
}

func isReallyGroup(ent *Entity) bool {
	if !ent.IsGroup() {
		return false
	}
	return len(ent.Groups())+len(ent.Leafs()) > 0
}

// TypePatchCluster returns the type without the decorations at the ends
// that are non-empty groups.
func (l *Link) TypePatchCluster() decoration.LinkType {
	result := l.Type()
	if isReallyGroup(l.cl1) {
		result = result.WithoutDecors2()
	}
	if isReallyGroup(l.cl2) {
		result = result.WithoutDecors1()
	}
	return result
}

// Label returns the label.
func (l *Link) Label() *klimt.Display { return l.linkArg.Label() }

// Length returns the length.
func (l *Link) Length() int { return l.linkArg.Length() }

// SetLength sets the length.
func (l *Link) SetLength(length int) { l.linkArg.SetLength(length) }

// Quantifier1 returns the cardinality of the first end.
func (l *Link) Quantifier1() OptString { return l.linkArg.Quantifier1() }

// Quantifier2 returns the cardinality of the second end.
func (l *Link) Quantifier2() OptString { return l.linkArg.Quantifier2() }

// Role1 returns the role of the first end.
func (l *Link) Role1() OptString { return l.linkArg.Role1() }

// Role2 returns the role of the second end.
func (l *Link) Role2() OptString { return l.linkArg.Role2() }

// Weight returns the Graphviz weight.
func (l *Link) Weight() float64 { return l.weight }

// SetWeight sets the Graphviz weight.
func (l *Link) SetWeight(weight float64) { l.weight = weight }

// Note returns the note, nil for none.
func (l *Link) Note() *CucaNote { return l.note }

// AddNote sets the note.
func (l *Link) AddNote(note *CucaNote) { l.note = note }

// AddNoteFrom copies the note of another link with a strategy.
func (l *Link) AddNoteFrom(other *Link, strategy NoteLinkStrategy) {
	if other.note != nil {
		l.note = other.note.WithStrategy(strategy)
	}
}

// IsAutoLinkOfAGroup reports whether the link joins a group to itself.
func (l *Link) IsAutoLinkOfAGroup() bool {
	if !l.cl1.IsGroup() || !l.cl2.IsGroup() {
		return false
	}
	return l.cl1 == l.cl2
}

// ContainsType reports whether an end is of a type.
func (l *Link) ContainsType(typ LeafType) bool {
	return l.cl1.LeafType() == typ || l.cl2.LeafType() == typ
}

// Contains reports whether the entity is an end.
func (l *Link) Contains(entity *Entity) bool { return l.cl1 == entity || l.cl2 == entity }

// Other returns the other end; an entity that is not an end panics, as
// upstream throws IllegalArgumentException.
func (l *Link) Other(entity *Entity) *Entity {
	if l.cl1 == entity {
		return l.cl2
	}
	if l.cl2 == entity {
		return l.cl1
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return nil
}

// IsConstraint reports whether the link ranks its ends.
func (l *Link) IsConstraint() bool { return l.constraint }

// SetConstraint sets whether the link ranks its ends.
func (l *Link) SetConstraint(constraint bool) { l.constraint = constraint }

// SetOpale marks the link of an opale note.
func (l *Link) SetOpale(opale bool) { l.opale = opale }

// SetHorizontalSolitary marks the horizontal link of a note.
func (l *Link) SetHorizontalSolitary(horizontalSolitary bool) {
	l.horizontalSolitary = horizontalSolitary
}

// IsHorizontalSolitary reports whether the link is the horizontal link of
// a note.
func (l *Link) IsHorizontalSolitary() bool { return l.horizontalSolitary }

// LinkArrow returns the label arrow, reversed for an inverted link.
func (l *Link) LinkArrow() LinkArrow {
	if l.inverted {
		return l.linkArrow.Reverse()
	}
	return l.linkArrow
}

// SetLinkArrow sets the label arrow.
func (l *Link) SetLinkArrow(linkArrow LinkArrow) { l.linkArrow = linkArrow }

// IsInverted reports whether the link was inverted (written right to left
// or bottom to top).
func (l *Link) IsInverted() bool { return l.inverted }

// HasEntryPoint reports whether an end is a leaf on the border of its
// container.
func (l *Link) HasEntryPoint() bool {
	return (!l.cl1.IsGroup() && l.cl1.EntityPosition() != EntityPositionNormal) ||
		(!l.cl2.IsGroup() && l.cl2.EntityPosition() != EntityPositionNormal)
}

// HasTwoEntryPointsSameContainer reports whether both ends are on the
// border of the same container.
func (l *Link) HasTwoEntryPointsSameContainer() bool {
	return !l.cl1.IsGroup() && !l.cl2.IsGroup() && l.cl1.EntityPosition() != EntityPositionNormal &&
		l.cl2.EntityPosition() != EntityPositionNormal && l.cl1.ParentContainer() == l.cl2.ParentContainer()
}

// URL returns the link target, nil for none.
func (l *Link) URL() *klimt.URL { return l.url }

// SetURL sets the link target.
func (l *Link) SetURL(url *klimt.URL) { l.url = url }

// IsHidden reports whether the link or an end is hidden.
func (l *Link) IsHidden() bool { return l.Hidden || l.cl1.IsHidden() || l.cl2.IsHidden() }

// SameConnections reports whether two links join the same entities.
func (l *Link) SameConnections(other *Link) bool {
	return (l.cl1 == other.cl1 && l.cl2 == other.cl2) || (l.cl1 == other.cl2 && l.cl2 == other.cl1)
}

// DoesTouch reports whether two links share an end.
func (l *Link) DoesTouch(other *Link) bool {
	return l.cl1 == other.cl1 || l.cl1 == other.cl2 || l.cl2 == other.cl1 || l.cl2 == other.cl2
}

// IsAutolink reports whether the link joins an entity to itself.
func (l *Link) IsAutolink() bool { return l.cl1 == l.cl2 }

// IsRemoved reports whether the stereotype of the link or an end is
// removed.
func (l *Link) IsRemoved() bool {
	if l.stereotype != nil && l.cucaDiagram.IsStereotypeRemoved(l.stereotype) {
		return true
	}
	return l.cl1.IsRemoved() || l.cl2.IsRemoved()
}

// HasURL reports whether the label or the link has a target.
func (l *Link) HasURL() bool {
	if !klimt.DisplayIsNull(l.linkArg.Label()) && l.linkArg.Label().HasURL() {
		return true
	}
	return l.url != nil
}

// Sametail returns the same tail group, null for none.
func (l *Link) Sametail() OptString { return l.sametail }

// SetSametail sets the same tail group.
func (l *Link) SetSametail(sametail OptString) { l.sametail = sametail }

// SetPortMembers sets the ports of both ends and records them on the
// entities.
func (l *Link) SetPortMembers(port1, port2 OptString) {
	l.port1 = port1
	l.port2 = port2
	if port1.OK {
		l.cl1.AddPortShortName(port1.S)
	}
	if port2.OK {
		l.cl2.AddPortShortName(port2.S)
	}
}

// SetLinkConstraint sets the constraint between two links.
func (l *Link) SetLinkConstraint(linkConstraint *LinkConstraint) { l.linkConstraint = linkConstraint }

// LinkConstraint returns the constraint, nil for none.
func (l *Link) LinkConstraint() *LinkConstraint { return l.linkConstraint }

// CodeLine returns the source line number set by SetCodeLine; ok is false
// when none was set.
func (l *Link) CodeLine() (string, bool) {
	if l.codeLine == nil {
		return "", false
	}
	return itoa(l.codeLine.Position), true
}

// SetCodeLine sets the source line of the link.
func (l *Link) SetCodeLine(location *text.LineLocation) { l.codeLine = location }

// SetStereotype sets the stereotype.
func (l *Link) SetStereotype(stereotype *klimt.Stereotype) { l.stereotype = stereotype }

// Stereotype returns the stereotype, nil for none.
func (l *Link) Stereotype() *klimt.Stereotype { return l.stereotype }

// LinkArg returns the label and end texts.
func (l *Link) LinkArg() *LinkArg { return l.linkArg }

// VisibilityModifier returns the visibility of the label; ok is false for
// none.
func (l *Link) VisibilityModifier() (klimt.VisibilityModifier, bool) {
	return l.linkArg.VisibilityModifier()
}

// IsOpale reports whether the link is the link of an opale note.
func (l *Link) IsOpale() bool { return l.opale }

// HasKal1 reports whether the first end has a qualifier.
func (l *Link) HasKal1() bool { return l.linkArg.Kal1().OK && l.linkArg.Kal1().S != "" }

// HasKal2 reports whether the second end has a qualifier.
func (l *Link) HasKal2() bool { return l.linkArg.Kal2().OK && l.linkArg.Kal2().S != "" }

// Location returns the source line, nil for none.
func (l *Link) Location() *text.LineLocation { return l.location }

// IsSingle reports whether the link is single (promoted from WithLinkType,
// listed for the users of CucaDiagram.AddLink).
func (l *Link) IsSingle() bool { return l.WithLinkType.IsSingle() }
