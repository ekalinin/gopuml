// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleQuery.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/AtomArray.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleAtom.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/LevelConstraint.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/Styleable.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/WithStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"slices"
	"strings"
)

// snameSet is a set of SName, upstream's EnumSet<SName>, iterated in enum
// order. It is comparable.
type snameSet [3]uint64

func (s snameSet) with(n SName) snameSet {
	s[n/64] |= 1 << (n % 64)
	return s
}

func (s snameSet) contains(n SName) bool { return s[n/64]&(1<<(n%64)) != 0 }

func (s snameSet) containsAll(other snameSet) bool {
	for i := range s {
		if other[i]&^s[i] != 0 {
			return false
		}
	}
	return true
}

func (s snameSet) union(other snameSet) snameSet {
	for i := range s {
		s[i] |= other[i]
	}
	return s
}

func (s snameSet) isEmpty() bool { return s == snameSet{} }

// names returns the names in enum order.
func (s snameSet) names() []SName {
	var result []SName
	for n := SName(0); n < snameCount; n++ {
		if s.contains(n) {
			result = append(result, n)
		}
	}
	return result
}

// String is EnumSet.toString, for example "[root, element]".
func (s snameSet) String() string {
	parts := make([]string, 0, 4)
	for _, n := range s.names() {
		parts = append(parts, n.String())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// levelNone is LevelConstraint.NO_LEVEL: no depth constraint.
const levelNone = -1

// checkLevel is LevelConstraint.of: a level outside 0..63 panics, as
// upstream throws IllegalArgumentException.
func checkLevel(level int) {
	if level != levelNone && (level < 0 || level >= 64) {
		throw(illegalArgument, "depth level out of range (must be 0..63): "+itoa(level))
	}
}

// levelConstraintMatches is LevelConstraint.matches: whether a stored
// declaration (declLevel, declStar) applies to the element of a query
// (queryLevel, queryStar). A declaration level is an exact match, or a
// minimum for a starred declaration; a starred query (an ancestor lookup)
// only matches starred declarations.
func levelConstraintMatches(declLevel int, declStar bool, queryLevel int, queryStar bool) bool {
	if declLevel != levelNone {
		if queryLevel == levelNone {
			return false
		}
		if declStar {
			if queryLevel < declLevel {
				return false
			}
		} else if queryLevel != declLevel {
			return false
		}
	}
	return !queryStar || declStar
}

// levelConstraintString is LevelConstraint.toString.
func levelConstraintString(level int, star bool) string {
	result := ""
	if level != levelNone {
		result = "depth(" + itoa(level) + ")"
	}
	if star {
		result += "*"
	}
	return result
}

// StyleSignatureStar is the star of a selector ("element*").
const StyleSignatureStar = "*"

// StyleSignature selects the style of an element.
type StyleSignature interface {
	MergedStyle(styleBuilder *StyleBuilder) *Style
	WithTOBECHANGED(stereotype *Stereotype) StyleSignature
	With(stereostyles *Stereostyles) StyleSignature
}

// StyleSignatureWithStereogroup is upstream's default
// withTOBECHANGED(Stereogroup): the signature with the stereotype of the
// group (unchanged for a nil group).
func StyleSignatureWithStereogroup(s StyleSignature, stereogroup *Stereogroup) StyleSignature {
	if stereogroup == nil {
		return s
	}
	return s.WithTOBECHANGED(stereogroup.BuildStereotype())
}

// StyleSignatureBasic is upstream's StyleQuery: the tags of an element (its
// style names and stereotypes, the atoms of a sorted AtomArray) and its
// LevelConstraint (depth level and star). It is an immutable, comparable
// value. The atoms are the names in enum order, then the stereotypes in
// Java's string order.
type StyleSignatureBasic struct {
	snames      snameSet
	stereotypes string // sorted names, each preceded by "\x00"
	level       int
	star        bool
}

var _ StyleSignature = StyleSignatureBasic{}

func joinStereotypes(names []string) string {
	slices.SortFunc(names, javaStringCompare)
	names = slices.Compact(names)
	var sb strings.Builder
	for _, n := range names {
		sb.WriteByte(0)
		sb.WriteString(n)
	}
	return sb.String()
}

// String is upstream's toString: the atoms, then the level constraint.
func (s StyleSignatureBasic) String() string {
	var parts []string
	for _, n := range s.snames.names() {
		parts = append(parts, n.String())
	}
	for _, st := range s.Stereotypes() {
		parts = append(parts, "."+st)
	}
	return "[" + strings.Join(parts, ", ") + "] " + levelConstraintString(s.level, s.star)
}

// StyleSignatureBasicEmpty returns the query without tag or level.
func StyleSignatureBasicEmpty() StyleSignatureBasic { return StyleSignatureBasic{level: levelNone} }

// AddClickable adds "clickable" when url is not nil.
func (s StyleSignatureBasic) AddClickable(url *URL) StyleSignatureBasic {
	if url == nil {
		return s
	}
	return s.AddSName(SNameClickable)
}

// AddLevel returns the query at a depth level, keeping its star. A level
// outside 0..63 panics, as upstream throws IllegalArgumentException.
func (s StyleSignatureBasic) AddLevel(level int) StyleSignatureBasic {
	checkLevel(level)
	s.level = level
	return s
}

func (s StyleSignatureBasic) withStereotypes(more ...string) StyleSignatureBasic {
	s.stereotypes = joinStereotypes(append(s.Stereotypes(), more...))
	return s
}

// AddStereotype returns the query with a stereotype, cleaned as
// StyleAtom.ofStereotype does: lower case, without "_" and ".".
func (s StyleSignatureBasic) AddStereotype(stereo string) StyleSignatureBasic {
	return s.withStereotypes(cleanStereotypeName(stereo))
}

// With returns the query with the style names of stereostyles.
func (s StyleSignatureBasic) With(stereostyles *Stereostyles) StyleSignature {
	if stereostyles.IsEmpty() {
		return s
	}
	var names []string
	for _, name := range stereostyles.StyleNames() {
		names = append(names, cleanStereotypeName(name))
	}
	return s.withStereotypes(names...)
}

// AddStereotypeOf is upstream's withStereotype(Stereotype): the query with
// every label of a stereotype that has style names; the query itself
// otherwise.
func (s StyleSignatureBasic) AddStereotypeOf(stereo *Stereotype) StyleSignatureBasic {
	if stereo == nil || len(stereo.StyleNames()) == 0 {
		return s
	}
	labels := stereo.Labels(GuillemetNone())
	if len(labels) == 0 {
		return s
	}
	var names []string
	for _, l := range labels {
		names = append(names, cleanStereotypeName(l))
	}
	return s.withStereotypes(names...)
}

// WithTOBECHANGED returns AddStereotypeOf: one query with all the labels.
func (s StyleSignatureBasic) WithTOBECHANGED(stereo *Stereotype) StyleSignature {
	return s.AddStereotypeOf(stereo)
}

// ForStereotypeItself returns, for a stereotype with style names, the query
// of the stereotype itself: the "stereotype" name and every label.
func (s StyleSignatureBasic) ForStereotypeItself(stereo *Stereotype) StyleSignature {
	if stereo == nil || len(stereo.StyleNames()) == 0 {
		return s
	}
	labels := stereo.Labels(GuillemetNone())
	if len(labels) == 0 {
		return s
	}
	var names []string
	for _, l := range labels {
		names = append(names, cleanStereotypeName(l))
	}
	return s.AddSName(SNameStereotype).withStereotypes(names...)
}

// AddSName returns the query with one more name.
func (s StyleSignatureBasic) AddSName(name SName) StyleSignatureBasic {
	s.snames = s.snames.with(name)
	return s
}

// AddStar returns the starred query, keeping its level.
func (s StyleSignatureBasic) AddStar() StyleSignatureBasic {
	s.star = true
	return s
}

// IsStarred reports whether the query is starred.
func (s StyleSignatureBasic) IsStarred() bool { return s.star }

// Level returns the depth level, -1 for none.
func (s StyleSignatureBasic) Level() int { return s.level }

// StyleSignatureBasicOf returns the query of names.
func StyleSignatureBasicOf(names ...SName) StyleSignatureBasic {
	result := StyleSignatureBasicEmpty()
	for _, n := range names {
		result.snames = result.snames.with(n)
	}
	return result
}

func cleanStereotypeName(name string) string {
	var sb strings.Builder
	for _, c := range name {
		if c != '_' && c != '.' {
			sb.WriteRune(javaCharLower(c))
		}
	}
	return sb.String()
}

// MergeWithStyles returns the query merged with the queries of others.
func (s StyleSignatureBasic) MergeWithStyles(others []*Style) StyleSignatureBasic {
	result := s
	for _, other := range others {
		result = result.MergeWith(other.Signature())
	}
	return result
}

// MergeWith returns the union of the two queries: every atom, the deeper
// level and either star.
func (s StyleSignatureBasic) MergeWith(other StyleSignatureBasic) StyleSignatureBasic {
	return StyleSignatureBasic{snames: s.snames.union(other.snames),
		stereotypes: joinStereotypes(append(s.Stereotypes(), other.Stereotypes()...)),
		level:       max(s.level, other.level), star: s.star || other.star}
}

// MergedStyle returns the merged style of the query, nil for a nil builder.
func (s StyleSignatureBasic) MergedStyle(styleBuilder *StyleBuilder) *Style {
	if styleBuilder == nil {
		return nil
	}
	return styleBuilder.MergedStyle(s)
}

// HasStereotype reports whether the query requires a stereotype.
func (s StyleSignatureBasic) HasStereotype() bool { return s.stereotypes != "" }

// StyleSignatureBasicActivity returns root.element.activityDiagram.activity.
func StyleSignatureBasicActivity() StyleSignatureBasic {
	return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameActivityDiagram, SNameActivity)
}

// StyleSignatureBasicActivityDiamond returns the query of activity
// diamonds.
func StyleSignatureBasicActivityDiamond() StyleSignatureBasic {
	return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameActivityDiagram, SNameActivity, SNameDiamond)
}

// StyleSignatureBasicActivityArrow returns the query of activity arrows.
func StyleSignatureBasicActivityArrow() StyleSignatureBasic {
	return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameActivityDiagram, SNameActivity, SNameArrow)
}

// StyleSignatureBasicOf5 is upstream's of(SName, SName, SName, SName,
// SName[]) for one or two extra names; more panic, as upstream throws.
func StyleSignatureBasicOf5(name0, name1, name2, name3 SName, names []SName) StyleSignatureBasic {
	switch len(names) {
	case 1:
		return StyleSignatureBasicOf(name0, name1, name2, name3, names[0])
	case 2:
		return StyleSignatureBasicOf(name0, name1, name2, name3, names[0], names[1])
	}
	throw(unsupportedOperation, "")
	return StyleSignatureBasic{}
}

// StyleSignatureBasicOf3 is upstream's of(SName, SName, SName, SName[],
// SName...).
func StyleSignatureBasicOf3(name0, name1, name2 SName, sNames []SName, other ...SName) StyleSignatureBasic {
	concat := append([]SName{name0, name1, name2}, sNames...)
	concat = append(concat, other...)
	return StyleSignatureBasicOf(concat...)
}

// SNames returns the names in enum order.
func (s StyleSignatureBasic) SNames() []SName { return s.snames.names() }

// Stereotypes returns the stereotypes, in Java's string order.
func (s StyleSignatureBasic) Stereotypes() []string {
	if s.stereotypes == "" {
		return nil
	}
	return strings.Split(s.stereotypes[1:], "\x00")
}

// styleAtom is upstream's StyleAtom: a style name, or a stereotype when
// isStereotype is set. It is comparable.
type styleAtom struct {
	sname        SName
	stereotype   string
	isStereotype bool
}

// atoms returns the atoms of the query in upstream's AtomArray order.
func (s StyleSignatureBasic) atoms() []styleAtom {
	var result []styleAtom
	for _, n := range s.snames.names() {
		result = append(result, styleAtom{sname: n})
	}
	for _, st := range s.Stereotypes() {
		result = append(result, styleAtom{stereotype: st, isStereotype: true})
	}
	return result
}

// Styleable is an element that has a style signature.
type Styleable interface {
	StyleSignature() StyleSignature
}

// WithStyle is an element that reports the styles it used.
type WithStyle interface {
	Styleable
	UsedStyles() []*Style
}
