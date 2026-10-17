// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/atmp/CucaDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/EntityFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/EntityPortion.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Diagram is a diagram built on CucaDiagram, as the shared entity commands
// see it.
type Diagram interface {
	uml.Diagram
	// Cuca returns the entity model of the diagram.
	Cuca() *CucaDiagram
}

// CucaHooks are the methods of CucaDiagram that diagram types override
// (upstream's protected updateLasts and isHideEmptyDescriptionForState).
type CucaHooks interface {
	UpdateLasts(result *Entity)
	IsHideEmptyDescriptionForState() bool
}

type noHooks struct{}

func (noHooks) UpdateLasts(*Entity)                  {}
func (noHooks) IsHideEmptyDescriptionForState() bool { return false }

// EntityHideOrShow is a "hide" or "show" rule on a portion of the entities
// of a gender.
type EntityHideOrShow struct {
	Gender  EntityGender
	Portion uml.EntityPortion
	Show    bool
}

// CucaDiagram is the entity model of graph diagrams, upstream's
// net.atmp.CucaDiagram: the namespace of entities, the links, the stack of
// open groups and together blocks, and the hide, show, remove and restore
// rules. Diagram types embed it (through AbstractEntityDiagram for class,
// description and state diagrams).
type CucaDiagram struct {
	uml.Base

	hooks CucaHooks

	hideOrShows            []EntityHideOrShow
	hideVisibilityModifier VisibilitySet

	hides2  []*HideOrShow
	removed []*HideOrShow

	cpt1, cpt2 int

	stacks []Bag

	visibilityModifierPresent bool

	links []*Link

	namespace *Plasma
	root      *Quark

	rawLayout  int
	lastEntity *Entity
}

// InitCucaDiagram initializes the model of a diagram of type diagramType
// built from src, with its root group. It must be called on the final
// location of the CucaDiagram (the entities keep a pointer to it); hooks
// may be nil.
func (d *CucaDiagram) InitCucaDiagram(ctx *core.RenderContext, src *core.UmlSource, diagramType core.DiagramType,
	md5map map[string]string, hooks CucaHooks) {
	d.Base = uml.NewBase(ctx, src, diagramType, md5map)
	if hooks == nil {
		hooks = noHooks{}
	}
	d.hooks = hooks
	d.namespace = newPlasma(&d.Base)
	d.root = d.namespace.Root()
	newGroupEntity(nil, nil, d.root, d, nil, GroupTypeRoot, 0)
	d.stacks = []Bag{d.root.Data()}
}

// Cuca implements Diagram.
func (d *CucaDiagram) Cuca() *CucaDiagram { return d }

// SetNamespaceSeparator sets the namespace separator; ok false sets null.
// Upstream also sets the separator of the namespace, which reads it from
// the diagram here.
func (d *CucaDiagram) SetNamespaceSeparator(separator string, ok bool) {
	d.Base.SetNamespaceSeparator(separator, ok)
}

// StartingPass implements command.Diagram.
func (d *CucaDiagram) StartingPass(command.ParserPass) {
	d.SetLastEntity(nil)
	d.cpt2 = 0
	if len(d.stacks) > 1 {
		d.stacks = d.stacks[:1]
	}
}

// PortFor returns the port part of a "entity::port" name of the entity of
// a quark, null otherwise.
func (d *CucaDiagram) PortFor(entString string, ident *Quark) OptString {
	x := strings.LastIndex(entString, "::")
	if x == -1 {
		return OptString{}
	}
	if strings.HasPrefix(entString, ident.Name()) {
		return Some(entString[x+2:])
	}
	return OptString{}
}

// CurrentGroup returns the innermost open group (the root at the top
// level).
func (d *CucaDiagram) CurrentGroup() *Entity {
	for pos := len(d.stacks) - 1; pos >= 0; pos-- {
		if e, ok := d.stacks[pos].(*Entity); ok {
			return e
		}
	}
	uml.Throw(uml.IllegalStateException, "")
	return nil
}

// CurrentTogether returns the open together block, nil when the top of the
// stack is a group. An empty stack panics as upstream's List.get.
func (d *CucaDiagram) CurrentTogether() *Together {
	pos := len(d.stacks) - 1
	if pos < 0 {
		uml.ThrowIndexOutOfBounds(pos, 0)
	}
	if t, ok := d.stacks[pos].(*Together); ok {
		return t
	}
	return nil
}

// CleanID removes the quotes around an id; null stays null.
func (d *CucaDiagram) CleanID(id string) string { return uml.RemoveQuotes(id) }

// HasURL reports whether an entity has a link.
func (d *CucaDiagram) HasURL() bool {
	for _, quark := range d.Quarks() {
		if ent := quark.Data(); ent != nil && ent.HasURL() {
			return true
		}
	}
	return false
}

// SetLastEntity sets the entity a note without "of" attaches to.
func (d *CucaDiagram) SetLastEntity(last *Entity) { d.lastEntity = last }

// ReallyCreateLeaf creates the leaf of a quark, in the current together
// block. A quark with an entity panics (IllegalStateException), as does
// the null display (IllegalArgumentException). Creating a class-like leaf
// creates the missing groups of the namespace.
func (d *CucaDiagram) ReallyCreateLeaf(location *text.LineLocation, ident *Quark, display *klimt.Display,
	typ LeafType, symbol USymbol) *Entity {
	if typ == LeafTypeNull {
		uml.ThrowNPE("")
	}
	if ident.Data() != nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	if klimt.DisplayIsNull(display) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	result := d.CreateLeaf(location, ident, d, typ, &d.hideVisibilityModifier)
	result.SetUSymbol(symbol)
	d.lastEntity = result
	result.SetTogether(d.CurrentTogether())
	d.hooks.UpdateLasts(result)
	result.SetDisplay(display)
	if typ.IsLikeClass() {
		d.eventuallyBuildPhantomGroups(location)
	}
	return result
}

// QuarkInContext is quarkInContextSafe for names that are valid; a bad
// name panics as upstream's Failable.get.
func (d *CucaDiagram) QuarkInContext(reuseExistingChild bool, full string) *Quark {
	q, err, _ := d.QuarkInContextSafe(reuseExistingChild, full)
	if err != "" {
		uml.Throw(uml.IllegalStateException, "")
	}
	return q
}

// QuarkInContextSafe returns the quark of a name in the current group:
// without namespace separator, the first quark of that name or a child of
// the current group; with one, an absolute name ("::a"), a name whose
// first part is a top-level package, a unique existing name (when
// reuseExistingChild) or a child of the current group. A bad name gives
// an error message and a score.
func (d *CucaDiagram) QuarkInContextSafe(reuseExistingChild bool, full string) (*Quark, string, int) {
	sep, ok := d.NamespaceSeparator()
	if !ok {
		if result := d.FirstWithName(full); result != nil {
			return result, "", 0
		}
		return d.CurrentGroup().Quark().Child(full), "", 0
	}
	if badName(sep, full) {
		return nil, "Bad name since " + sep + " is a separator", 3
	}
	currentQuark := d.CurrentGroup().Quark()
	if rest, ok := strings.CutPrefix(full, sep); ok {
		return d.root.Child(rest), "", 0
	}
	x := strings.Index(full, sep)
	if x == -1 {
		if reuseExistingChild && d.CountByName(full) == 1 {
			byName := d.FirstWithName(full)
			if byName != currentQuark {
				return byName, "", 0
			}
		}
		return currentQuark.Child(full), "", 0
	}
	first := d.root.ChildIfExists(full[:x])
	if first != nil && first.Data() != nil && !first.Data().IsGroup() {
		return nil, "Not a package: " + full[:x], 0
	}
	if first != nil {
		return d.root.Child(full), "", 0
	}
	return currentQuark.Child(full), "", 0
}

func badName(separator, full string) bool {
	return strings.HasSuffix(full, separator) || strings.Contains(full, separator+separator)
}

// RemovePortID removes the "::port" part of an id, unless "::" is the
// namespace separator.
func (d *CucaDiagram) RemovePortID(id string) string {
	if sep, ok := d.NamespaceSeparator(); ok && sep == "::" {
		return id
	}
	x := strings.LastIndex(id, "::")
	if x == -1 {
		return id
	}
	return id[:x]
}

// PortID returns the port part of an id, null when there is none or when
// "::" is the namespace separator.
func (d *CucaDiagram) PortID(id string) OptString {
	if sep, ok := d.NamespaceSeparator(); ok && sep == "::" {
		return OptString{}
	}
	x := strings.LastIndex(id, "::")
	if x == -1 {
		return OptString{}
	}
	return Some(id[x+2:])
}

// ChildrenGroups implements GroupHierarchy.
func (d *CucaDiagram) ChildrenGroups(entity *Entity) []*Entity { return entity.Groups() }

// eventuallyBuildPhantomGroups creates a package for every quark without
// entity that has children.
func (d *CucaDiagram) eventuallyBuildPhantomGroups(location *text.LineLocation) {
	for _, quark := range d.Quarks() {
		if quark.Data() != nil {
			continue
		}
		if quark.CountChildren() > 0 {
			display := klimt.DisplayWithNewlines(d.Pragma(), quark.Name())
			result := d.CreateGroup(location, quark, GroupTypePackage)
			result.SetDisplay(display)
		}
	}
}

// EventuallyBuildPhantomGroups creates the missing packages, as upstream
// does before drawing.
func (d *CucaDiagram) EventuallyBuildPhantomGroups() { d.eventuallyBuildPhantomGroups(nil) }

// GotoTogether opens a together block.
func (d *CucaDiagram) GotoTogether() command.Result {
	d.stacks = append(d.stacks, &Together{parent: d.CurrentTogether()})
	return command.Ok()
}

// GotoGroup opens the group of a quark, creating it with display when
// needed; symbol, when not null, becomes its symbol.
func (d *CucaDiagram) GotoGroup(location *text.LineLocation, quark *Quark, display *klimt.Display,
	typ GroupType, symbol USymbol) command.Result {
	if quark.Data() == nil {
		result := d.CreateGroup(location, quark, typ)
		result.SetTogether(d.CurrentTogether())
		result.SetDisplay(display)
	}
	ent := quark.Data()
	ent.MuteToGroupType(typ)
	if symbol != USymbolNull {
		ent.SetUSymbol(symbol)
	}
	d.stacks = append(d.stacks, quark.Data())
	return command.Ok()
}

// EndGroup closes the innermost group or together block; it returns false
// when the stack is empty.
func (d *CucaDiagram) EndGroup() bool {
	if len(d.stacks) > 0 {
		d.stacks = d.stacks[:len(d.stacks)-1]
		return true
	}
	return false
}

// Group returns the entity of the first quark of a name, nil for none.
func (d *CucaDiagram) Group(code string) *Entity {
	if quark := d.FirstWithName(code); quark != nil {
		return quark.Data()
	}
	return nil
}

// IsGroupNamed reports whether the first quark of a name is a group.
func (d *CucaDiagram) IsGroupNamed(code string) bool {
	quark := d.FirstWithName(code)
	if quark == nil {
		return false
	}
	return d.IsGroupQuark(quark)
}

// IsGroupQuark reports whether the entity of a quark is a group.
func (d *CucaDiagram) IsGroupQuark(quark *Quark) bool {
	ent := quark.Data()
	if ent == nil {
		return false
	}
	return ent.IsGroup()
}

var cucaNumberPattern = regexp.MustCompile(`^[+-]?(\.?\d+|\d+\.\d*)$`)

func isNumber(s string) bool { return cucaNumberPattern.MatchString(s) }

// ResetPragmaLabel undefines the label distance and angle pragmas.
func (d *CucaDiagram) ResetPragmaLabel() {
	d.Pragma().Undefine(klimt.PragmaLabelDistance)
	d.Pragma().Undefine(klimt.PragmaLabelAngle)
}

func (d *CucaDiagram) pragmaNumber(key, defaultKey klimt.PragmaKey, defaultValue string) string {
	pragma := d.Pragma()
	if pragma.IsDefine(key) {
		if s, ok := pragma.Value(key); ok && isNumber(s) {
			return s
		} else if !ok {
			uml.ThrowNPE(`Cannot invoke "java.lang.CharSequence.length()" because "this.text" is null`)
		}
	}
	if pragma.IsDefine(defaultKey) {
		if s, ok := pragma.Value(defaultKey); ok && isNumber(s) {
			return s
		} else if !ok {
			uml.ThrowNPE(`Cannot invoke "java.lang.CharSequence.length()" because "this.text" is null`)
		}
	}
	return defaultValue
}

// Labeldistance returns the label distance of the links: the pragma, or
// 1.7.
func (d *CucaDiagram) Labeldistance() OptString {
	return Some(d.pragmaNumber(klimt.PragmaLabelDistance, klimt.PragmaDefaultLabelDistance, "1.7"))
}

// Labelangle returns the label angle of the links: the pragma, or 25.
func (d *CucaDiagram) Labelangle() OptString {
	return Some(d.pragmaNumber(klimt.PragmaLabelAngle, klimt.PragmaDefaultLabelAngle, "25"))
}

// IsEmpty implements GroupHierarchy.
func (d *CucaDiagram) IsEmpty(entity *Entity) bool { return entity.IsEmpty() }

// IsVisibilityModifierPresent reports whether a member has a visibility
// character.
func (d *CucaDiagram) IsVisibilityModifierPresent() bool { return d.visibilityModifierPresent }

// SetVisibilityModifierPresent records that a member has a visibility
// character.
func (d *CucaDiagram) SetVisibilityModifierPresent(present bool) {
	d.visibilityModifierPresent = present
}

// ShowPortion implements PortionShower: the last rule on the portion that
// selects the entity decides; circles are hidden in strict UML style.
func (d *CucaDiagram) ShowPortion(portion uml.EntityPortion, entity *Entity) bool {
	if d.SkinParam().StrictUmlStyle() && portion == uml.PortionCircledCharacter {
		return false
	}
	result := true
	for _, cmd := range d.hideOrShows {
		if cmd.Portion == portion && cmd.Gender.Contains(entity) {
			result = cmd.Show
		}
	}
	return result
}

// HideOrShowVisibilityModifier hides or shows the members of some
// visibilities.
func (d *CucaDiagram) HideOrShowVisibilityModifier(visibilities VisibilitySet, show bool) {
	if show {
		d.hideVisibilityModifier &^= visibilities
	} else {
		d.hideVisibilityModifier |= visibilities
	}
}

// VisibleStereotypeLabels implements PortionShower.
func (d *CucaDiagram) VisibleStereotypeLabels(entity *Entity) []string {
	stereotype := entity.Stereotype()
	if stereotype == nil {
		return nil
	}
	visible := []string{}
	for _, label := range stereotype.Labels(klimt.GuillemetDoubleComparator()) {
		if d.isStereotypeLabelShown(label) {
			visible = append(visible, label)
		}
	}
	return visible
}

func (d *CucaDiagram) isStereotypeLabelShown(label string) bool {
	result := true
	for _, cmd := range d.hideOrShows {
		if cmd.Portion != uml.PortionStereotype {
			continue
		}
		gender, ok := cmd.Gender.Gender()
		if !ok || gender == label {
			result = cmd.Show
		}
	}
	return result
}

// HideOrShow adds a rule on the portions of the entities of a gender.
func (d *CucaDiagram) HideOrShow(gender EntityGender, portions uml.EntityPortion, show bool) {
	for _, portion := range portions.AsSet().List() {
		d.hideOrShows = append(d.hideOrShows, EntityHideOrShow{Gender: gender, Portion: portion, Show: show})
	}
}

// HideOrShows returns the rules on portions.
func (d *CucaDiagram) HideOrShows() []EntityHideOrShow { return d.hideOrShows }

// HideOrShow2 adds a "hide" or "show" rule on elements.
func (d *CucaDiagram) HideOrShow2(what string, show bool) {
	d.hides2 = append(d.hides2, NewHideOrShow(d.fixWhat(what), show))
}

// RemoveOrRestore adds a "remove" or "restore" rule on elements.
func (d *CucaDiagram) RemoveOrRestore(what string, show bool) {
	d.removed = append(d.removed, NewHideOrShow(d.fixWhat(what), show))
}

// Hides2 returns the "hide" and "show" rules on elements.
func (d *CucaDiagram) Hides2() []*HideOrShow { return d.hides2 }

// Removed returns the "remove" and "restore" rules.
func (d *CucaDiagram) Removed() []*HideOrShow { return d.removed }

func (d *CucaDiagram) fixWhat(what string) string {
	if sep, ok := d.NamespaceSeparator(); ok {
		currentQuark := d.CurrentGroup().Quark()
		if currentQuark.QualifiedName() != "" {
			what = currentQuark.QualifiedName() + sep + what
		}
	}
	return what
}

// HidesVisibilityModifier returns the hidden visibilities.
func (d *CucaDiagram) HidesVisibilityModifier() VisibilitySet { return d.hideVisibilityModifier }

// IsStandalone reports whether no link touches the entity.
func (d *CucaDiagram) IsStandalone(ent *Entity) bool {
	for _, link := range d.links {
		if link.Entity1() == ent || link.Entity2() == ent {
			return false
		}
	}
	return true
}

// IsStandaloneForArgo reports whether no visible link touches the entity.
func (d *CucaDiagram) IsStandaloneForArgo(ent *Entity) bool {
	for _, link := range d.links {
		if link.IsHidden() || link.IsInvis() {
			continue
		}
		if link.Entity1() == ent || link.Entity2() == ent {
			return false
		}
	}
	return true
}

// LastLink returns the last link between entities that are not notes, nil
// for none.
func (d *CucaDiagram) LastLink() *Link {
	for i := len(d.links) - 1; i >= 0; i-- {
		link := d.links[i]
		if link.Entity1().LeafType() != LeafTypeNote && link.Entity2().LeafType() != LeafTypeNote {
			return link
		}
	}
	return nil
}

// TwoLastLinks returns the two last links between entities that are not
// notes, the last first; nil when there are fewer.
func (d *CucaDiagram) TwoLastLinks() []*Link {
	var result []*Link
	for i := len(d.links) - 1; i >= 0; i-- {
		link := d.links[i]
		if link.Entity1().LeafType() != LeafTypeNote && link.Entity2().LeafType() != LeafTypeNote {
			result = append(result, link)
			if len(result) == 2 {
				return result
			}
		}
	}
	return nil
}

// LastEntity returns the last created entity, nil for none.
func (d *CucaDiagram) LastEntity() *Entity { return d.lastEntity }

// ApplySingleStrategy puts the standalone entities of each group, when
// there are three or more, in squares of invisible links, then the squares
// of each group when there are three or more.
func (d *CucaDiagram) ApplySingleStrategy() {
	magmas := &magmaList{}
	groups := d.GroupsAndRoot()
	for _, g := range groups {
		var standalones []*Entity
		for _, ent := range g.Leafs() {
			if d.IsStandalone(ent) {
				standalones = append(standalones, ent)
			}
		}
		if len(standalones) < 3 {
			continue
		}
		m := newMagma(d, standalones)
		m.putInSquare()
		magmas.all = append(magmas.all, m)
	}
	for _, g := range groups {
		list := magmas.magmas(g)
		if len(list.all) < 3 {
			continue
		}
		list.putInSquare()
	}
}

// IsHideEmptyDescriptionForState reports whether state diagrams hide empty
// descriptions.
func (d *CucaDiagram) IsHideEmptyDescriptionForState() bool {
	return d.hooks.IsHideEmptyDescriptionForState()
}

// ConstraintOnLinks sets a constraint between two links.
func (d *CucaDiagram) ConstraintOnLinks(link1, link2 *Link, display *klimt.Display) command.Result {
	linkConstraint := NewLinkConstraint(link1, link2, display)
	link1.SetLinkConstraint(linkConstraint)
	link2.SetLinkConstraint(linkConstraint)
	return command.Ok()
}

// DefaultMargins returns the margins around the drawing: 0, 5, 5, 0.
func (d *CucaDiagram) DefaultMargins() klimt.ClockwiseTopRightBottomLeft {
	return klimt.ClockwiseTopRightBottomLeftOf(0, 5, 5, 0)
}

// UniqueSequenceValue returns the next value of the diagram counter.
func (d *CucaDiagram) UniqueSequenceValue() int {
	d.cpt1++
	return d.cpt1
}

// UniqueSequence returns prefix and the next value of the diagram counter.
func (d *CucaDiagram) UniqueSequence(prefix string) string {
	d.cpt1++
	return prefix + strconv.Itoa(d.cpt1)
}

// UniqueSequence2 returns prefix and the next value of the pass counter.
func (d *CucaDiagram) UniqueSequence2(prefix string) string {
	d.cpt2++
	return prefix + strconv.Itoa(d.cpt2)
}

// IsStereotypeRemoved reports whether "remove <<stereotype>>" removes a
// stereotype.
func (d *CucaDiagram) IsStereotypeRemoved(stereotype *klimt.Stereotype) bool {
	result := false
	for _, hide := range d.removed {
		result = hide.ApplyStereotype(result, stereotype)
	}
	return result
}

// IsHidden reports whether the "hide" rules hide an entity; a note linked
// to a single entity follows that entity.
func (d *CucaDiagram) IsHidden(leaf *Entity) bool {
	if leaf.IsRoot() {
		return false
	}
	if other := d.isNoteWithSingleLinkAttachedTo(leaf); other != nil && other != leaf {
		return d.IsHidden(other)
	}
	hidden := false
	for _, hide := range d.hides2 {
		hidden = hide.Apply(hidden, leaf)
	}
	return hidden
}

// IsRemoved reports whether the "remove" rules remove an entity; a note
// linked to a single entity follows that entity.
func (d *CucaDiagram) IsRemoved(leaf *Entity) bool {
	if leaf == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.abel.Entity.isRoot()" because "leaf" is null`)
	}
	if leaf.IsRoot() {
		return false
	}
	if other := d.isNoteWithSingleLinkAttachedTo(leaf); other != nil {
		return d.IsRemoved(other)
	}
	result := false
	for _, hide := range d.removed {
		result = hide.Apply(result, leaf)
	}
	return result
}

func (d *CucaDiagram) isNoteWithSingleLinkAttachedTo(note *Entity) *Entity {
	if note.LeafType() != LeafTypeNote {
		return nil
	}
	var other *Entity
	for _, link := range d.links {
		if link.Type().IsInvisible() {
			continue
		}
		if !link.Contains(note) {
			continue
		}
		if other != nil {
			return nil
		}
		other = link.Other(note)
		if other.LeafType() == LeafTypeNote {
			return nil
		}
	}
	return other
}

// IsRemovedIgnoreUnlinked is IsRemoved without the "@unlinked" rules and
// the containers.
func (d *CucaDiagram) IsRemovedIgnoreUnlinked(leaf *Entity) bool {
	result := false
	for _, hide := range d.removed {
		if !hide.IsAboutUnlinked() {
			result = hide.Apply(result, leaf)
		}
	}
	return result
}

// CreateLeaf creates the entity of a quark as a leaf of a type, with the
// body of the type.
func (d *CucaDiagram) CreateLeaf(location *text.LineLocation, quark *Quark, diagram *CucaDiagram,
	entityType LeafType, hideVisibilityModifier *VisibilitySet) *Entity {
	var bodier Bodier
	switch entityType {
	case LeafTypeNull:
		uml.ThrowNPE("")
	case LeafTypeMap:
		bodier = newBodierMap()
	case LeafTypeJSON:
		bodier = &BodierJSon{}
	default:
		bodier = BodyFactoryCreateLeaf(d.SkinParam(), entityType, hideVisibilityModifier)
	}
	result := newLeafEntity(d.SkinParam().CurrentStyleBuilder(), location, quark, d, bodier, entityType,
		diagram.rawLayout)
	bodier.SetLeaf(result)
	return result
}

// CreateGroup creates the entity of a quark as a group of a type; it
// returns the existing entity of the quark if any.
func (d *CucaDiagram) CreateGroup(location *text.LineLocation, quark *Quark, groupType GroupType) *Entity {
	if groupType == GroupTypeNull {
		uml.ThrowNPE("")
	}
	if quark.Data() != nil {
		return quark.Data()
	}
	bodier := BodyFactoryCreateGroup(d.SkinParam())
	return newGroupEntity(d.SkinParam().CurrentStyleBuilder(), location, quark, d, bodier, groupType, d.rawLayout)
}

// Leafs returns the leaves of the diagram, in quark order.
func (d *CucaDiagram) Leafs() []*Entity {
	var result []*Entity
	for _, quark := range d.Quarks() {
		if quark.IsRoot() {
			continue
		}
		if data := quark.Data(); data != nil && !data.IsGroup() {
			result = append(result, data)
		}
	}
	return result
}

// Groups returns the groups of the diagram but the root, in quark order.
func (d *CucaDiagram) Groups() []*Entity {
	var result []*Entity
	for _, quark := range d.Quarks() {
		if quark.IsRoot() {
			continue
		}
		if data := quark.Data(); data != nil && data.IsGroup() {
			result = append(result, data)
		}
	}
	return result
}

// GroupsAndRoot returns the groups of the diagram, the root first.
func (d *CucaDiagram) GroupsAndRoot() []*Entity {
	var result []*Entity
	for _, quark := range d.Quarks() {
		if data := quark.Data(); data != nil && data.IsGroup() {
			result = append(result, data)
		}
	}
	return result
}

// RawLayout returns the current layout row.
func (d *CucaDiagram) RawLayout() int { return d.rawLayout }

// IncRawLayout starts a new layout row ("layout_new_line").
func (d *CucaDiagram) IncRawLayout() { d.rawLayout++ }

// Links returns the links in order.
func (d *CucaDiagram) Links() []*Link { return d.links }

// AddLink adds a link, unless it is single and a link joins the same
// entities.
func (d *CucaDiagram) AddLink(link *Link) {
	if link.IsSingle() && d.containsSimilarLink(link) {
		return
	}
	d.links = append(d.links, link)
}

func (d *CucaDiagram) containsSimilarLink(other *Link) bool {
	for _, link := range d.links {
		if other.SameConnections(link) {
			return true
		}
	}
	return false
}

// RemoveLink removes a link; a link that is not in the diagram panics, as
// upstream throws IllegalArgumentException.
func (d *CucaDiagram) RemoveLink(link *Link) {
	idx := slices.Index(d.links, link)
	if idx == -1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	d.links = slices.Delete(d.links, idx, idx+1)
}

// Quarks returns every quark, the root first.
func (d *CucaDiagram) Quarks() []*Quark { return slices.Clone(d.namespace.Quarks()) }

// RootGroup implements GroupHierarchy.
func (d *CucaDiagram) RootGroup() *Entity { return d.root.Data() }

// FirstWithName returns the first quark of a name, nil for none.
func (d *CucaDiagram) FirstWithName(full string) *Quark { return d.namespace.FirstWithName(full) }

// CountByName returns the number of quarks of a name.
func (d *CucaDiagram) CountByName(full string) int { return d.namespace.CountByName(full) }

// Namespace returns the namespace of the diagram.
func (d *CucaDiagram) Namespace() *Plasma { return d.namespace }

// Stacks returns the stack of open groups and together blocks.
func (d *CucaDiagram) Stacks() []Bag { return d.stacks }
