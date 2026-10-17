// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/UGroup.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/UGroupType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// UGroupType is the kind of an attribute of a drawing group (an SVG <g>).
type UGroupType int

// Group attributes, in upstream's order.
const (
	UGroupID UGroupType = iota
	UGroupClass
	UGroupTitle
	UGroupDataEntity
	UGroupDataQualifiedName
	UGroupDataEntity1
	UGroupDataEntity2
	UGroupDataEntityUID
	UGroupDataEntity1UID
	UGroupDataEntity2UID
	UGroupDataParticipant
	UGroupDataParticipant1
	UGroupDataParticipant2
	UGroupDataUID
	UGroupDataSourceLine
	UGroupDataVisibilityModifier
	UGroupDataLinkType
	ugroupTypeCount
)

var ugroupTypeNames = []string{"ID", "CLASS", "TITLE", "DATA_ENTITY", "DATA_QUALIFIED_NAME", "DATA_ENTITY_1",
	"DATA_ENTITY_2", "DATA_ENTITY_UID", "DATA_ENTITY_1_UID", "DATA_ENTITY_2_UID", "DATA_PARTICIPANT",
	"DATA_PARTICIPANT_1", "DATA_PARTICIPANT_2", "DATA_UID", "DATA_SOURCE_LINE", "DATA_VISIBILITY_MODIFIER",
	"DATA_LINK_TYPE"}

// String returns upstream's constant name, for example "DATA_SOURCE_LINE".
func (t UGroupType) String() string { return enumName("UGroupType", ugroupTypeNames, int(t)) }

// SvgKeyAttributeName returns the SVG attribute name, for example
// "data-source-line".
func (t UGroupType) SvgKeyAttributeName() string {
	return strings.ReplaceAll(strings.ToLower(t.String()), "_", "-")
}

// UGroup is the set of attributes of a drawing group, kept in UGroupType
// order (upstream's EnumMap).
type UGroup struct {
	values  [ugroupTypeCount]string
	present [ugroupTypeCount]bool
}

// NewUGroup returns an empty group.
func NewUGroup() *UGroup { return &UGroup{} }

// NewUGroupAt returns a group with the source line of location; a nil
// location gives an empty group.
func NewUGroupAt(location *text.LineLocation) *UGroup {
	g := &UGroup{}
	if location != nil {
		g.values[UGroupDataSourceLine] = itoa(location.Position)
		g.present[UGroupDataSourceLine] = true
	}
	return g
}

// Put sets an attribute; characters other than word characters, '-' and
// spaces become '.'.
func (g *UGroup) Put(key UGroupType, value string) {
	g.values[key] = fixUGroupValue(value)
	g.present[key] = true
}

// UGroupSingleton returns a group with one attribute.
func UGroupSingleton(key UGroupType, value string) *UGroup {
	g := NewUGroup()
	g.Put(key, value)
	return g
}

var nonWord = regexp.MustCompile(`[^-\w ]`)

// fixUGroupValue replaces each character that is not [-\w ] by '.'. Java's
// \w is ASCII only, as RE2's, and both match code points.
func fixUGroupValue(value string) string { return nonWord.ReplaceAllString(value, ".") }

// UGroupEntry is one attribute of a group.
type UGroupEntry struct {
	Key   UGroupType
	Value string
}

// Entries returns the attributes in UGroupType order.
func (g *UGroup) Entries() []UGroupEntry {
	var result []UGroupEntry
	for k := UGroupType(0); k < ugroupTypeCount; k++ {
		if g.present[k] {
			result = append(result, UGroupEntry{Key: k, Value: g.values[k]})
		}
	}
	return result
}

// Get returns the value of an attribute; ok is false when it is not set.
func (g *UGroup) Get(key UGroupType) (value string, ok bool) {
	return g.values[key], g.present[key]
}
