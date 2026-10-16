// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Entity.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/EntityUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/SpecificBackcolorable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/LineConfigurable.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"fmt"
	"slices"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// Entity is an element of a graph diagram: a leaf (class, object, note,
// use case...) or a group (package, namespace...). It is the data of a
// quark of the namespace of the diagram.
type Entity struct {
	currentStyleBuilder *klimt.StyleBuilder
	diagram             *CucaDiagram
	quark               *Quark
	url                 *klimt.URL
	bodier              Bodier
	uid                 string
	display             *klimt.Display
	legend              *uml.DisplayPositioned

	leafType     LeafType
	stereotype   *klimt.Stereotype
	stereostyles *klimt.Stereostyles
	generic      OptString

	groupType GroupType

	margins             klimt.Margins
	portShortNames      []string
	xposition           int
	svekImage           klimt.IEntityImage
	symbol              USymbol
	rawLayout           int
	concurrentSeparator rune
	location            *text.LineLocation

	tags        []klimt.Stereotag
	notesTop    []*CucaNote
	notesBottom []*CucaNote

	together *Together

	packed        bool
	isStatic      bool
	visibility    klimt.VisibilityModifier
	hasVisibility bool
	colors        *color.Colors
	tipKeys       []string
	tips          map[string]*Tip
	kals          [4][]Kal
}

func (*Entity) isBag() {}

// newEntity is upstream's private constructor: the entity becomes the data
// of its quark and gets the next uid of the diagram ("entroot" for the
// root).
func newEntity(currentStyleBuilder *klimt.StyleBuilder, location *text.LineLocation, quark *Quark,
	diagram *CucaDiagram, bodier Bodier, rawLayout int) *Entity {
	if quark == nil {
		uml.ThrowNPE("")
	}
	e := &Entity{currentStyleBuilder: currentStyleBuilder, location: location, quark: quark, diagram: diagram,
		bodier: bodier, rawLayout: rawLayout, display: klimt.DisplayEmpty(), stereostyles: klimt.StereostylesNone(),
		margins: klimt.MarginsNone(), colors: color.EmptyColors(), tips: map[string]*Tip{},
		leafType: LeafTypeNull, groupType: GroupTypeNull}
	if quark.IsRoot() {
		e.uid = "entroot"
	} else {
		e.uid = fmt.Sprintf("ent%04d", diagram.UniqueSequenceValue())
	}
	quark.SetData(e)
	return e
}

// newLeafEntity returns a leaf of type leafType.
func newLeafEntity(currentStyleBuilder *klimt.StyleBuilder, location *text.LineLocation, quark *Quark,
	diagram *CucaDiagram, bodier Bodier, leafType LeafType, rawLayout int) *Entity {
	e := newEntity(currentStyleBuilder, location, quark, diagram, bodier, rawLayout)
	e.leafType = leafType
	return e
}

// newGroupEntity returns a group of type groupType.
func newGroupEntity(currentStyleBuilder *klimt.StyleBuilder, location *text.LineLocation, quark *Quark,
	diagram *CucaDiagram, bodier Bodier, groupType GroupType, rawLayout int) *Entity {
	e := newEntity(currentStyleBuilder, location, quark, diagram, bodier, rawLayout)
	e.groupType = groupType
	return e
}

// AddNote adds a note at the top or the bottom of a group; other positions
// are ignored.
func (e *Entity) AddNote(note *klimt.Display, position Position, colors *color.Colors) {
	switch position {
	case PositionTop:
		e.notesTop = append(e.notesTop, NewCucaNote(note, position, colors))
	case PositionBottom:
		e.notesBottom = append(e.notesBottom, NewCucaNote(note, position, colors))
	}
}

// Notes returns the notes at the top or the bottom; other positions panic,
// as upstream throws IllegalArgumentException.
func (e *Entity) Notes(position Position) []*CucaNote {
	switch position {
	case PositionTop:
		return e.notesTop
	case PositionBottom:
		return e.notesBottom
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return nil
}

// AddStereotag adds a "$tag", once.
func (e *Entity) AddStereotag(tag klimt.Stereotag) {
	if !slices.Contains(e.tags, tag) {
		e.tags = append(e.tags, tag)
	}
}

// Stereotags returns the tags in order.
func (e *Entity) Stereotags() []klimt.Stereotag { return e.tags }

// LeafType returns the type of a leaf, LeafTypeNull for a group.
func (e *Entity) LeafType() LeafType { return e.leafType }

// MuteToType turns the entity into a leaf of another type.
func (e *Entity) MuteToType(newType LeafType) {
	if e.leafType == LeafTypeClass && newType == LeafTypeObject {
		e.bodier.MuteClassToObject()
	}
	e.groupType = GroupTypeNull
	e.leafType = newType
}

// MuteToGroupType turns the entity into a group of another type.
func (e *Entity) MuteToGroupType(newType GroupType) {
	e.groupType = newType
	e.leafType = LeafTypeNull
}

func isClassLikeForMute(t LeafType) bool {
	switch t {
	case LeafTypeAnnotation, LeafTypeAbstractClass, LeafTypeClass, LeafTypeEnum, LeafTypeInterface,
		LeafTypeRecord, LeafTypeDataclass:
		return true
	}
	return false
}

// MuteToTypeSymbol is upstream's muteToType(LeafType, USymbol): it changes
// the type of a leaf declared again, and returns false when the change is
// not allowed (only class-like types may become another class-like type or
// an object).
func (e *Entity) MuteToTypeSymbol(newType LeafType, newSymbol USymbol) bool {
	if newType == LeafTypeNull {
		uml.ThrowNPE("")
	}
	if e.leafType != LeafTypeStillUnknown {
		if newType == e.leafType {
			return true
		}
		if !isClassLikeForMute(e.leafType) {
			return false
		}
		if !isClassLikeForMute(newType) && newType != LeafTypeObject {
			return false
		}
	}
	if e.leafType == LeafTypeClass && newType == LeafTypeObject {
		e.bodier.MuteClassToObject()
	}
	e.leafType = newType
	e.symbol = newSymbol
	return true
}

// Quark returns the quark of the entity.
func (e *Entity) Quark() *Quark { return e.quark }

// Display returns the text.
func (e *Entity) Display() *klimt.Display { return e.display }

// SetDisplay sets the text.
func (e *Entity) SetDisplay(display *klimt.Display) { e.display = display }

// UID returns the uid, for example "ent0002".
func (e *Entity) UID() string { return e.uid }

// Stereotype returns the stereotype, nil for none.
func (e *Entity) Stereotype() *klimt.Stereotype { return e.stereotype }

// SetStereotype sets the stereotype.
func (e *Entity) SetStereotype(stereotype *klimt.Stereotype) { e.stereotype = stereotype }

// String is upstream's toString.
func (e *Entity) String() string {
	return e.quark.String() + " " + e.display.String() + "(" + e.leafType.String() + ")[" + e.groupType.String() +
		"] " + e.uid
}

// URL99 returns the link of the entity, nil for none.
func (e *Entity) URL99() *klimt.URL { return e.url }

// HasURL reports whether the entity, its text or its members have a link.
func (e *Entity) HasURL() bool {
	if !klimt.DisplayIsNull(e.display) && e.display.HasURL() {
		return true
	}
	if e.bodier == nil {
		return false
	}
	if e.bodier.HasURL() {
		return true
	}
	return e.url != nil
}

// AddURL sets the link.
func (e *Entity) AddURL(url *klimt.URL) { e.url = url }

// Margins returns the margins of a leaf.
func (e *Entity) Margins() klimt.Margins {
	e.checkNotGroup()
	return e.margins
}

// EnsureMargins merges margins into the margins.
func (e *Entity) EnsureMargins(newMargins klimt.Margins) { e.margins = e.margins.Merge(newMargins) }

// Xposition returns the horizontal position of a leaf.
func (e *Entity) Xposition() int {
	e.checkNotGroup()
	return e.xposition
}

// SetXposition sets the horizontal position of a leaf.
func (e *Entity) SetXposition(pos int) {
	e.checkNotGroup()
	e.xposition = pos
}

// SvekImage returns the image of a leaf, nil for none.
func (e *Entity) SvekImage() klimt.IEntityImage {
	e.checkNotGroup()
	return e.svekImage
}

// SetSvekImage sets the image of a leaf.
func (e *Entity) SetSvekImage(svekImage klimt.IEntityImage) {
	e.checkNotGroup()
	e.svekImage = svekImage
}

// SetGeneric sets the generic parameters of a leaf.
func (e *Entity) SetGeneric(generic string) {
	e.checkNotGroup()
	e.generic = Some(generic)
}

// Generic returns the generic parameters of a leaf.
func (e *Entity) Generic() OptString {
	e.checkNotGroup()
	return e.generic
}

// GenericValue returns the generic parameters without the leaf check: a
// group keeps the parameters of the leaf it was.
func (e *Entity) GenericValue() OptString { return e.generic }

// Bodier returns the body, nil for the root.
func (e *Entity) Bodier() Bodier { return e.bodier }

// EntityPosition returns where the entity sits: ports, and states with an
// entry point or similar stereotype, are on the border of their container.
func (e *Entity) EntityPosition() EntityPosition {
	switch e.leafType {
	case LeafTypePortin:
		return EntityPositionPortin
	case LeafTypePortout:
		return EntityPositionPortout
	}
	if e.leafType != LeafTypeState {
		return EntityPositionNormal
	}
	if e.quark.IsRoot() {
		return EntityPositionNormal
	}
	if e.stereotype == nil {
		return EntityPositionNormal
	}
	return EntityPositionFromStereotype(e.stereotype.Label(klimt.GuillemetDoubleComparator()))
}

func (e *Entity) checkGroup() {
	if !e.IsGroup() {
		uml.Throw(uml.UnsupportedOperationException, "")
	}
}

func (e *Entity) checkNotGroup() {
	if e.IsGroup() {
		uml.Throw(uml.UnsupportedOperationException, "")
	}
}

// GroupType returns the type of a group; a leaf panics, as upstream throws
// UnsupportedOperationException.
func (e *Entity) GroupType() GroupType {
	e.checkGroup()
	return e.groupType
}

// PackageStyle returns the package style of the stereotype of a group; ok
// is false when there is none.
func (e *Entity) PackageStyle() (klimt.PackageStyle, bool) {
	e.checkGroup()
	if e.stereotype == nil {
		return 0, false
	}
	return e.stereotype.PackageStyle()
}

// IsGroup reports whether the entity is a group.
func (e *Entity) IsGroup() bool {
	if e.groupType != GroupTypeNull && e.leafType != LeafTypeNull {
		uml.Throw(uml.IllegalStateException, "")
	}
	if e.groupType != GroupTypeNull {
		return true
	}
	if e.leafType != LeafTypeNull {
		return false
	}
	uml.Throw(uml.IllegalStateException, "")
	return false
}

// OverrideImage turns a group into a leaf drawn with an image, removing the
// links inside the group.
func (e *Entity) OverrideImage(img klimt.IEntityImage, leafType LeafType) {
	e.checkGroup()
	e.svekImage = img
	e.url = nil
	for _, link := range slices.Clone(e.diagram.Links()) {
		if EntityUtilsIsPureInnerLink12(e, link) {
			e.diagram.RemoveLink(link)
		}
	}
	e.groupType = GroupTypeNull
	e.leafType = leafType
}

// USymbol returns the symbol: the use case and interface symbols for use
// cases and circles, the symbol given at creation otherwise.
func (e *Entity) USymbol() USymbol {
	switch e.leafType {
	case LeafTypeUsecase:
		return USymbolUsecase
	case LeafTypeUsecaseBusiness:
		return USymbolUsecaseBusiness
	case LeafTypeCircle:
		return USymbolInterface
	}
	return e.symbol
}

// SetUSymbol sets the symbol.
func (e *Entity) SetUSymbol(symbol USymbol) { e.symbol = symbol }

// IsHidden reports whether "hide" hides the entity or its container.
func (e *Entity) IsHidden() bool {
	if e.IsRoot() {
		return false
	}
	parentContainer := e.ParentContainer()
	if parentContainer == e {
		return false
	}
	if parentContainer != nil && parentContainer.IsHidden() {
		return true
	}
	return e.diagram.IsHidden(e)
}

// IsRemoved reports whether "remove" removes the entity or its container.
func (e *Entity) IsRemoved() bool {
	if e.IsRoot() {
		return false
	}
	parentContainer := e.ParentContainer()
	if parentContainer == e {
		return false
	}
	if parentContainer != nil && parentContainer.IsRemoved() {
		return true
	}
	return e.diagram.IsRemoved(e)
}

// IsAloneAndUnlinked reports whether the entity (every child of a group)
// has no visible link to an entity that is not removed.
func (e *Entity) IsAloneAndUnlinked() bool {
	if e.IsGroup() {
		for _, quarkChild := range e.quark.Children() {
			child := quarkChild.Data()
			if child == nil {
				uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.abel.Entity.isAloneAndUnlinked()" because "child" is null`)
			}
			if !child.IsAloneAndUnlinked() {
				return false
			}
		}
		return true
	}
	for _, link := range e.diagram.Links() {
		if link.Contains(e) {
			other := link.Other(e)
			removed := e.diagram.IsRemovedIgnoreUnlinked(other)
			if !removed && !link.Type().IsInvisible() {
				return false
			}
		}
	}
	return true
}

// RawLayout returns the layout row of "layout_new_line".
func (e *Entity) RawLayout() int { return e.rawLayout }

// ConcurrentSeparator returns the separator of concurrent states.
func (e *Entity) ConcurrentSeparator() rune { return e.concurrentSeparator }

// SetConcurrentSeparator sets the separator of concurrent states.
func (e *Entity) SetConcurrentSeparator(separator rune) { e.concurrentSeparator = separator }

// PutTip sets the tip of a member.
func (e *Entity) PutTip(member string, display *klimt.Display, colors *color.Colors,
	stereotype *klimt.Stereotype) {
	if _, ok := e.tips[member]; !ok {
		e.tipKeys = append(e.tipKeys, member)
	}
	e.tips[member] = NewTip(display, colors, stereotype)
}

// Tips returns the members with a tip, in order, and the tips by member.
func (e *Entity) Tips() (members []string, tips map[string]*Tip) { return e.tipKeys, e.tips }

// Colors returns the colors.
func (e *Entity) Colors() *color.Colors { return e.colors }

// SetColors sets the colors.
func (e *Entity) SetColors(colors *color.Colors) { e.colors = colors }

// SetSpecificColorTOBEREMOVED adds a color of a type; nil does nothing.
func (e *Entity) SetSpecificColorTOBEREMOVED(typ color.ColorType, c color.HColor) {
	if c != nil {
		e.colors = e.colors.Add(typ, c)
	}
}

// PortShortNames returns the names of the ports of a leaf.
func (e *Entity) PortShortNames() []string {
	e.checkNotGroup()
	return e.portShortNames
}

// PortShortNamesValue returns the names of the ports without the leaf
// check.
func (e *Entity) PortShortNamesValue() []string { return e.portShortNames }

// AddPortShortName adds the name of a port.
func (e *Entity) AddPortShortName(portShortName string) {
	if !slices.Contains(e.portShortNames, portShortName) {
		e.portShortNames = append(e.portShortNames, portShortName)
	}
}

// SetVisibilityModifier sets the visibility; ok false clears it.
func (e *Entity) SetVisibilityModifier(visibility klimt.VisibilityModifier, ok bool) {
	e.visibility, e.hasVisibility = visibility, ok
}

// VisibilityModifier returns the visibility; ok is false for none.
func (e *Entity) VisibilityModifier() (klimt.VisibilityModifier, bool) {
	return e.visibility, e.hasVisibility
}

// SetLegend sets the legend of a group.
func (e *Entity) SetLegend(legend uml.DisplayPositioned) {
	e.checkGroup()
	e.legend = &legend
}

// Legend returns the legend, nil for none.
func (e *Entity) Legend() *uml.DisplayPositioned { return e.legend }

// SetStereostyle sets the style names of a stereotype.
func (e *Entity) SetStereostyle(stereo string) { e.stereostyles = klimt.StereostylesBuild(stereo) }

// Stereostyles returns the style names of the stereotype.
func (e *Entity) Stereostyles() *klimt.Stereostyles { return e.stereostyles }

// Diagram returns the diagram.
func (e *Entity) Diagram() *CucaDiagram { return e.diagram }

// SetStatic marks a "static class".
func (e *Entity) SetStatic(isStatic bool) { e.isStatic = isStatic }

// IsStatic reports whether the entity is a "static class".
func (e *Entity) IsStatic() bool { return e.isStatic }

// SetTogether sets the together block, nil for none.
func (e *Entity) SetTogether(together *Together) { e.together = together }

// Together returns the together block, nil for none.
func (e *Entity) Together() *Together { return e.together }

// ParentContainer returns the entity of the parent quark, nil for the root
// (and for a parent quark without entity).
func (e *Entity) ParentContainer() *Entity {
	if e.quark.IsRoot() {
		return nil
	}
	return e.quark.Parent().Data()
}

// Leafs returns the leaves among the children, in order.
func (e *Entity) Leafs() []*Entity {
	var result []*Entity
	for _, child := range e.quark.Children() {
		if data := child.Data(); data != nil && !data.IsGroup() {
			result = append(result, data)
		}
	}
	return result
}

// Groups returns the groups among the children, in order.
func (e *Entity) Groups() []*Entity {
	var result []*Entity
	for _, child := range e.quark.Children() {
		if data := child.Data(); data != nil && data.IsGroup() {
			result = append(result, data)
		}
	}
	return result
}

// CountChildren returns the number of child quarks.
func (e *Entity) CountChildren() int { return e.quark.CountChildren() }

// IsRoot reports whether the entity is the root group.
func (e *Entity) IsRoot() bool { return e.quark.IsRoot() }

// IsEmpty reports whether every child of the group is removed.
func (e *Entity) IsEmpty() bool {
	for _, quarkChild := range e.quark.Children() {
		if !e.diagram.IsRemoved(quarkChild.Data()) {
			return false
		}
	}
	return true
}

// Name returns the name of the quark.
func (e *Entity) Name() string { return e.quark.Name() }

// IsAutarkic reports whether the group can be drawn on its own.
func (e *Entity) IsAutarkic() bool {
	switch e.GroupType() {
	case GroupTypePackage:
		return false
	case GroupTypeInnerActivity, GroupTypeConcurrentActivity, GroupTypeConcurrentState:
		return true
	}
	for _, link := range e.diagram.Links() {
		if !EntityUtilsIsPureInnerLink3(e, link) {
			return false
		}
	}
	for _, leaf := range e.Leafs() {
		if leaf.EntityPosition() != EntityPositionNormal {
			return false
		}
	}
	return true
}

// CanBePacked reports whether the group holds a single non-empty group and
// nothing else, without links: its title then joins the child's.
func (e *Entity) CanBePacked() bool {
	if e.IsPacked() {
		return false
	}
	if e.CountChildren() != 1 {
		return false
	}
	if len(e.Leafs()) != 0 {
		return false
	}
	for _, link := range e.diagram.Links() {
		if link.Contains(e) {
			return false
		}
	}
	groups := e.Groups()
	if len(groups) == 0 {
		uml.Throw(uml.NoSuchElementException, "")
	}
	return groups[0].CountChildren() != 0
}

// SetPacked marks the group as packed; as upstream, the argument is
// ignored.
func (e *Entity) SetPacked(bool) { e.packed = true }

// IsPacked reports whether the group is packed.
func (e *Entity) IsPacked() bool { return e.packed }

// SkinParam returns the skin parameters of the diagram.
func (e *Entity) SkinParam() klimt.ISkinParam { return e.diagram.SkinParam() }

// DiagramType returns the type of the diagram.
func (e *Entity) DiagramType() core.DiagramType { return e.diagram.DiagramType() }

// Location returns the source line, nil for none.
func (e *Entity) Location() *text.LineLocation { return e.location }

// CurrentStyleBuilder returns the style sheet of the entity: the latest one
// when a skinparam command was used, the one active at its creation
// otherwise.
func (e *Entity) CurrentStyleBuilder() *klimt.StyleBuilder {
	if e.diagram.IsSkinParamUsed() {
		return e.SkinParam().CurrentStyleBuilder()
	}
	return e.currentStyleBuilder
}

// BestMatch returns the body line that best matches a member name.
func (e *Entity) BestMatch(candidate string) (string, bool) { return e.bodier.BestMatch(candidate) }

// entityUtilsIsParent is upstream's EntityUtils.isParent.
func entityUtilsIsParent(groupToBeTested, parentGroup *Entity) bool {
	if !groupToBeTested.IsGroup() {
		return false
	}
	for !groupToBeTested.IsRoot() {
		if groupToBeTested == parentGroup {
			return true
		}
		groupToBeTested = groupToBeTested.ParentContainer()
		if !groupToBeTested.IsGroup() {
			return false
		}
	}
	return false
}

// EntityUtilsIsPureInnerLink12 reports whether both ends of a link are in
// a group.
func EntityUtilsIsPureInnerLink12(group *Entity, link *Link) bool {
	if !group.IsGroup() {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	group1 := link.Entity1().ParentContainer()
	group2 := link.Entity2().ParentContainer()
	return entityUtilsIsParent(group1, group) && entityUtilsIsParent(group2, group)
}

// EntityUtilsIsPureInnerLink3 reports whether both ends of a link are in a
// group, or both outside.
func EntityUtilsIsPureInnerLink3(group *Entity, link *Link) bool {
	if !group.IsGroup() {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	group1 := link.Entity1().ParentContainer()
	group2 := link.Entity2().ParentContainer()
	return entityUtilsIsParent(group2, group) == entityUtilsIsParent(group1, group)
}
