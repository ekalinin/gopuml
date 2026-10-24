// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/LeafType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/GroupType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/uml"
)

// LeafType is the kind of an entity that is not a group.
type LeafType int

// Leaf types, in upstream's order. LeafTypeNull stands for upstream's null
// (the leaf type of a group).
const (
	LeafTypeNull LeafType = iota - 1
	LeafTypeEmptyPackage
	LeafTypeAbstractClass
	LeafTypeClass
	LeafTypeInterface
	LeafTypeAnnotation
	LeafTypeProtocol
	LeafTypeStruct
	LeafTypeException
	LeafTypeMetaclass
	LeafTypeStereotype
	LeafTypeLollipopFull
	LeafTypeLollipopHalf
	LeafTypeNote
	LeafTypeTips
	LeafTypeObject
	LeafTypeMap
	LeafTypeJSON
	LeafTypeAssociation
	LeafTypeEnum
	LeafTypeCircle
	LeafTypeDataclass
	LeafTypeRecord
	LeafTypeUsecase
	LeafTypeUsecaseBusiness
	LeafTypeDescription
	LeafTypeArcCircle
	LeafTypeActivity
	LeafTypeBranch
	LeafTypeSynchroBar
	LeafTypeCircleStart
	LeafTypeCircleEnd
	LeafTypePointForAssociation
	LeafTypeActivityConcurrent
	LeafTypeState
	LeafTypeStateConcurrent
	LeafTypePseudoState
	LeafTypeDeepHistory
	LeafTypeStateChoice
	LeafTypeStateForkJoin
	LeafTypeStateTransitionLabel
	LeafTypeJunction
	LeafTypeBlock
	LeafTypeEntity
	LeafTypeDomain
	LeafTypeRequirement
	LeafTypePortin
	LeafTypePortout
	LeafTypeChenEntity
	LeafTypeChenRelationship
	LeafTypeChenAttribute
	LeafTypeChenCircle
	LeafTypeStillUnknown
)

var leafTypeNames = [...]string{
	"EMPTY_PACKAGE", "ABSTRACT_CLASS", "CLASS", "INTERFACE", "ANNOTATION", "PROTOCOL", "STRUCT", "EXCEPTION",
	"METACLASS", "STEREOTYPE", "LOLLIPOP_FULL", "LOLLIPOP_HALF", "NOTE", "TIPS", "OBJECT", "MAP", "JSON",
	"ASSOCIATION", "ENUM", "CIRCLE", "DATACLASS", "RECORD", "USECASE", "USECASE_BUSINESS", "DESCRIPTION",
	"ARC_CIRCLE", "ACTIVITY", "BRANCH", "SYNCHRO_BAR", "CIRCLE_START", "CIRCLE_END", "POINT_FOR_ASSOCIATION",
	"ACTIVITY_CONCURRENT", "STATE", "STATE_CONCURRENT", "PSEUDO_STATE", "DEEP_HISTORY", "STATE_CHOICE",
	"STATE_FORK_JOIN", "STATE_TRANSITION_LABEL", "JUNCTION", "BLOCK", "ENTITY", "DOMAIN", "REQUIREMENT", "PORTIN", "PORTOUT",
	"CHEN_ENTITY", "CHEN_RELATIONSHIP", "CHEN_ATTRIBUTE", "CHEN_CIRCLE", "STILL_UNKNOWN",
}

// String returns upstream's constant name, "null" for LeafTypeNull.
func (t LeafType) String() string {
	if t == LeafTypeNull {
		return "null"
	}
	return leafTypeNames[t]
}

// LeafTypeValueOf is upstream's LeafType.valueOf; an unknown name panics as
// upstream throws IllegalArgumentException.
func LeafTypeValueOf(name string) LeafType {
	for i, n := range leafTypeNames {
		if n == name {
			return LeafType(i)
		}
	}
	uml.Throw(uml.IllegalArgumentException,
		"No enum constant net.sourceforge.plantuml.abel.LeafType."+name)
	return LeafTypeNull
}

// LeafTypeOf is upstream's getLeafType: the type of a keyword such as
// "abstract class", "static class" or "diamond".
func LeafTypeOf(typ string) LeafType {
	typ = uml.UpperCase(typ)
	switch {
	case strings.HasPrefix(typ, "ABSTRACT"):
		return LeafTypeAbstractClass
	case strings.HasPrefix(typ, "DIAMOND"):
		return LeafTypeStateChoice
	case strings.HasPrefix(typ, "STATIC"):
		return LeafTypeClass
	}
	return LeafTypeValueOf(typ)
}

// IsLikeClass reports whether the type is drawn as a class.
func (t LeafType) IsLikeClass() bool {
	switch t {
	case LeafTypeAnnotation, LeafTypeAbstractClass, LeafTypeClass, LeafTypeInterface, LeafTypeEnum, LeafTypeEntity,
		LeafTypeProtocol, LeafTypeStruct, LeafTypeException, LeafTypeMetaclass, LeafTypeStereotype,
		LeafTypeDataclass, LeafTypeRecord:
		return true
	}
	return false
}

// ToHTML returns the name in lower case with spaces, capitalized, for
// example "Abstract class".
func (t LeafType) ToHTML() string {
	html := uml.LowerCase(strings.ReplaceAll(t.String(), "_", " "))
	if html == "" {
		return html
	}
	return uml.UpperCase(html[:1]) + html[1:]
}

// GroupType is the kind of a group entity.
type GroupType int

// Group types, in upstream's order. GroupTypeNull stands for upstream's
// null (the group type of a leaf).
const (
	GroupTypeNull GroupType = iota - 1
	GroupTypeRoot
	GroupTypePackage
	GroupTypeState
	GroupTypeConcurrentState
	GroupTypeInnerActivity
	GroupTypeConcurrentActivity
	GroupTypeDomain
	GroupTypeRequirement
)

var groupTypeNames = [...]string{"ROOT", "PACKAGE", "STATE", "CONCURRENT_STATE", "INNER_ACTIVITY",
	"CONCURRENT_ACTIVITY", "DOMAIN", "REQUIREMENT"}

// String returns upstream's constant name, "null" for GroupTypeNull.
func (t GroupType) String() string {
	if t == GroupTypeNull {
		return "null"
	}
	return groupTypeNames[t]
}
