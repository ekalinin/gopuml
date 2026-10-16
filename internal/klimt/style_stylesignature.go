// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/StyleKey.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/StyleSignature.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/StyleSignatureBasic.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/StyleSignatures.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/Styleable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/WithStyle.java
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

// StyleKey is the selector part of a style signature: the names, the depth
// level (-1 for none) and the star. It is a comparable value.
type StyleKey struct {
	snames   snameSet
	isStared bool
	level    int
}

// StyleKeyEmpty returns the key without names, level or star.
func StyleKeyEmpty() StyleKey { return StyleKey{level: -1} }

// String is upstream's toString.
func (k StyleKey) String() string {
	var sb strings.Builder
	sb.WriteString(k.snames.String() + " ")
	if k.level != -1 {
		sb.WriteString(" " + itoa(k.level))
	}
	if k.isStared {
		sb.WriteString(" (*)")
	}
	return sb.String()
}

// AddClickable adds the "clickable" name when url is not nil.
func (k StyleKey) AddClickable(url *URL) StyleKey {
	if url == nil {
		return k
	}
	return k.AddSName(SNameClickable)
}

// AddLevel returns the key with a depth level.
func (k StyleKey) AddLevel(level int) StyleKey {
	k.level = level
	return k
}

// AddSName returns the key with one more name.
func (k StyleKey) AddSName(name SName) StyleKey {
	k.snames = k.snames.with(name)
	return k
}

// AddStar returns the starred key.
func (k StyleKey) AddStar() StyleKey {
	k.isStared = true
	return k
}

// StyleKeyOf returns the key of names.
func StyleKeyOf(names ...SName) StyleKey {
	k := StyleKeyEmpty()
	for _, n := range names {
		k.snames = k.snames.with(n)
	}
	return k
}

// MergeWith returns the union of the names, the larger level and either
// star.
func (k StyleKey) MergeWith(other StyleKey) StyleKey {
	return StyleKey{snames: k.snames.union(other.snames), level: max(k.level, other.level),
		isStared: k.isStared || other.isStared}
}

// SNames returns the names in enum order.
func (k StyleKey) SNames() []SName { return k.snames.names() }

// IsStared reports whether the key is starred.
func (k StyleKey) IsStared() bool { return k.isStared }

// Level returns the depth level, -1 for none.
func (k StyleKey) Level() int { return k.level }

// StyleSignatureStar is the star of a selector ("element*").
const StyleSignatureStar = "*"

// StyleSignature selects the style of an element: a single signature or the
// union of several (one per stereotype).
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

// StyleSignatureBasic is a style selector: style names, depth level, star
// and stereotypes. It is an immutable, comparable value; the stereotypes are
// kept sorted (upstream's HashSet only shows its order in toString).
type StyleSignatureBasic struct {
	key         StyleKey
	stereotypes string // sorted names, each preceded by "\x00"
}

var _ StyleSignature = StyleSignatureBasic{}

func joinStereotypes(names []string) string {
	slices.Sort(names)
	names = slices.Compact(names)
	var sb strings.Builder
	for _, n := range names {
		sb.WriteByte(0)
		sb.WriteString(n)
	}
	return sb.String()
}

// StyleSignatureBasicCreateStereotype returns the signature of a stereotype
// style.
func StyleSignatureBasicCreateStereotype(s string) StyleSignatureBasic {
	return StyleSignatureBasicEmpty().AddStereotype(s)
}

// String is upstream's toString, the stereotypes sorted.
func (s StyleSignatureBasic) String() string {
	return s.key.String() + " [" + strings.Join(s.Stereotypes(), ", ") + "]"
}

// StyleSignatureBasicEmpty returns the empty signature.
func StyleSignatureBasicEmpty() StyleSignatureBasic { return StyleSignatureBasic{key: StyleKeyEmpty()} }

// AddClickable adds "clickable" when url is not nil.
func (s StyleSignatureBasic) AddClickable(url *URL) StyleSignatureBasic {
	if url == nil {
		return s
	}
	return StyleSignatureBasic{key: s.key.AddClickable(url), stereotypes: s.stereotypes}
}

// AddLevel returns the signature with a depth level.
func (s StyleSignatureBasic) AddLevel(level int) StyleSignatureBasic {
	return StyleSignatureBasic{key: s.key.AddLevel(level), stereotypes: s.stereotypes}
}

func (s StyleSignatureBasic) withStereotypes(more ...string) StyleSignatureBasic {
	names := append(s.Stereotypes(), more...)
	return StyleSignatureBasic{key: s.key, stereotypes: joinStereotypes(names)}
}

// AddStereotype returns the signature with a stereotype, cleaned: lower
// case, without "_" and ".".
func (s StyleSignatureBasic) AddStereotype(stereo string) StyleSignatureBasic {
	return s.withStereotypes(cleanStereotypeName(stereo))
}

// With returns the signature with the style names of stereostyles.
func (s StyleSignatureBasic) With(stereostyles *Stereostyles) StyleSignature {
	if stereostyles.IsEmpty() {
		return s
	}
	return s.withStereotypes(stereostyles.StyleNames()...)
}

// AddStereotypeOf returns the signature with the labels of stereo, cleaned;
// the signature itself for a nil stereotype.
func (s StyleSignatureBasic) AddStereotypeOf(stereo *Stereotype) StyleSignatureBasic {
	if stereo == nil {
		return s
	}
	var names []string
	for _, l := range stereo.Labels(GuillemetNone()) {
		names = append(names, cleanStereotypeName(l))
	}
	return s.withStereotypes(names...)
}

// WithTOBECHANGED returns, for a stereotype with style names, the union of
// the signature with each of its labels.
func (s StyleSignatureBasic) WithTOBECHANGED(stereo *Stereotype) StyleSignature {
	if stereo == nil || len(stereo.StyleNames()) == 0 {
		return s
	}
	labels := stereo.Labels(GuillemetNone())
	if len(labels) == 0 {
		return s
	}
	result := &StyleSignatures{}
	for _, name := range labels {
		result.Add(s.AddStereotype(name))
	}
	return result
}

// ForStereotypeItself returns, for a stereotype with style names, the union
// of the stereotype styles of each label.
func (s StyleSignatureBasic) ForStereotypeItself(stereo *Stereotype) StyleSignature {
	if stereo == nil || len(stereo.StyleNames()) == 0 {
		return s
	}
	labels := stereo.Labels(GuillemetNone())
	if len(labels) == 0 {
		return s
	}
	result := &StyleSignatures{}
	for _, name := range labels {
		result.Add(s.AddStereotype(name).AddSName(SNameStereotype))
	}
	return result
}

// AddSName returns the signature with one more name.
func (s StyleSignatureBasic) AddSName(name SName) StyleSignatureBasic {
	return StyleSignatureBasic{key: s.key.AddSName(name), stereotypes: s.stereotypes}
}

// AddStar returns the starred signature.
func (s StyleSignatureBasic) AddStar() StyleSignatureBasic {
	return StyleSignatureBasic{key: s.key.AddStar(), stereotypes: s.stereotypes}
}

// IsStarred reports whether the signature is starred.
func (s StyleSignatureBasic) IsStarred() bool { return s.key.isStared }

// MatchAll reports whether the declaration s applies to the element: its
// level matches (exactly, or at least for a starred declaration), its
// names and stereotypes are all in the element's, and the element is not
// starred unless the declaration is.
func (s StyleSignatureBasic) MatchAll(element StyleSignatureBasic) bool {
	declaration := s
	if declaration.key.level != -1 {
		if declaration.key.isStared {
			if element.key.level == -1 {
				return false
			}
			if element.key.level < declaration.key.level {
				return false
			}
		} else {
			if element.key.level == -1 {
				return false
			}
			if element.key.level != declaration.key.level {
				return false
			}
		}
	}
	if element.IsStarred() && !declaration.IsStarred() {
		return false
	}
	if !element.key.snames.containsAll(declaration.key.snames) {
		return false
	}
	if declaration.stereotypes != "" {
		elementStereotypes := element.Stereotypes()
		for _, st := range declaration.Stereotypes() {
			if !slices.Contains(elementStereotypes, st) {
				return false
			}
		}
	}
	return true
}

// StyleSignatureBasicOf returns the signature of names.
func StyleSignatureBasicOf(names ...SName) StyleSignatureBasic {
	return StyleSignatureBasic{key: StyleKeyOf(names...)}
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

// MergeWithStyles returns the signature merged with the signatures of
// others.
func (s StyleSignatureBasic) MergeWithStyles(others []*Style) StyleSignatureBasic {
	result := s
	for _, other := range others {
		result = result.MergeWith(other.Signature())
	}
	return result
}

// MergeWith returns the union of the two signatures.
func (s StyleSignatureBasic) MergeWith(other StyleSignatureBasic) StyleSignatureBasic {
	names := append(s.Stereotypes(), other.Stereotypes()...)
	return StyleSignatureBasic{key: s.key.MergeWith(other.key), stereotypes: joinStereotypes(names)}
}

// MergedStyle returns the merged style of the signature, nil for a nil
// builder.
func (s StyleSignatureBasic) MergedStyle(styleBuilder *StyleBuilder) *Style {
	if styleBuilder == nil {
		return nil
	}
	return styleBuilder.MergedStyle(s)
}

// IsWithDot reports whether the signature has stereotypes.
func (s StyleSignatureBasic) IsWithDot() bool { return s.stereotypes != "" }

// StyleSignatureBasicActivity returns root.element.activityDiagram.activity.
func StyleSignatureBasicActivity() StyleSignatureBasic {
	return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameActivityDiagram, SNameActivity)
}

// StyleSignatureBasicActivityDiamond returns the signature of activity
// diamonds.
func StyleSignatureBasicActivityDiamond() StyleSignatureBasic {
	return StyleSignatureBasicOf(SNameRoot, SNameElement, SNameActivityDiagram, SNameActivity, SNameDiamond)
}

// StyleSignatureBasicActivityArrow returns the signature of activity
// arrows.
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

// IsEmpty reports whether the signature has no names and no stereotypes.
func (s StyleSignatureBasic) IsEmpty() bool { return s.key.snames.isEmpty() && s.stereotypes == "" }

// Key returns the selector key.
func (s StyleSignatureBasic) Key() StyleKey { return s.key }

// Stereotypes returns the stereotypes, sorted.
func (s StyleSignatureBasic) Stereotypes() []string {
	if s.stereotypes == "" {
		return nil
	}
	return strings.Split(s.stereotypes[1:], "\x00")
}

// StyleSignatures is the union of several signatures; its merged style
// merges theirs, keeping the stereotype values of the first.
type StyleSignatures struct {
	all []StyleSignature
}

var _ StyleSignature = (*StyleSignatures)(nil)

// Add adds a signature.
func (s *StyleSignatures) Add(signature StyleSignature) { s.all = append(s.all, signature) }

// String is upstream's toString.
func (s *StyleSignatures) String() string {
	parts := make([]string, len(s.all))
	for i, sig := range s.all {
		if st, ok := sig.(interface{ String() string }); ok {
			parts[i] = st.String()
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// MergedStyle merges the styles of the signatures. An empty union panics,
// as upstream throws UnsupportedOperationException.
func (s *StyleSignatures) MergedStyle(currentStyleBuilder *StyleBuilder) *Style {
	if len(s.all) == 0 {
		throw(unsupportedOperation, "")
	}
	var result *Style
	for _, basic := range s.all {
		tmp := basic.MergedStyle(currentStyleBuilder)
		if result == nil {
			result = tmp
		} else {
			result = result.MergeWith(tmp, MergeKeepExistingValueOfStereotype)
		}
	}
	return result
}

// WithTOBECHANGED panics, as upstream throws UnsupportedOperationException.
func (s *StyleSignatures) WithTOBECHANGED(*Stereotype) StyleSignature {
	throw(unsupportedOperation, "")
	return nil
}

// With returns the union of each signature with the style names.
func (s *StyleSignatures) With(stereostyles *Stereostyles) StyleSignature {
	if len(s.all) == 0 {
		throw(unsupportedOperation, "")
	}
	result := &StyleSignatures{}
	for _, basic := range s.all {
		result.Add(basic.With(stereostyles))
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
