// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/SName.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "strings"

// SName is the name of a style selector element ("root", "element",
// "sequenceDiagram", "participant", ...). Names that are Java keywords end
// with "_" upstream (class_, package_, ...); String returns that name.
type SName int

// Style names, in upstream's order.
const (
	SNameAction SName = iota
	SNameActivationBox
	SNameActivity
	SNameActivityBar
	SNameActivityDiagram
	SNameActor
	SNameAgent
	SNameAnalog
	SNameArchimate
	SNameArrow
	SNameArtifact
	SNameBinary
	SNameBody
	SNameBoundary
	SNameBox
	SNameBoxless
	SNameBusiness
	SNameCaption
	SNameCard
	SNameCardinality
	SNameCircle
	SNameClassDiagram
	SNameClass
	SNameClickable
	SNameCloud
	SNameClosed
	SNameCollection
	SNameCollections
	SNameComponent
	SNameComposite
	SNameRobust
	SNameChenAttribute
	SNameChenEerDiagram
	SNameChenEntity
	SNameChenRelationship
	SNameConcise
	SNameClock
	SNameComponentDiagram
	SNameConstraintArrow
	SNameControl
	SNameDatabase
	SNameDay
	SNameDelay
	SNameDescription
	SNameDestroy
	SNameDiamond
	SNameDocument
	SNameEbnf
	SNameElement
	SNameEntity
	SNameEnd
	SNameStart
	SNameStop
	SNameFile
	SNameFilesDiagram
	SNameFolder
	SNameFooter
	SNameFrame
	SNameGanttDiagram
	SNameGeneric
	SNameGoto
	SNameGroup
	SNameGroupHeader
	SNameHeader
	SNameHexagon
	SNameHighlight
	SNameHnote
	SNameInterface
	SNameJson
	SNameJsonDiagram
	SNameGitDiagram
	SNameLabel
	SNameLeafNode
	SNameLegend
	SNameLifeLine
	SNameMainframe
	SNameMap
	SNameMilestone
	SNameMindmapDiagram
	SNameMonth
	SNameName
	SNameNetwork
	SNameNewpage
	SNameNode
	SNameNote
	SNameNwdiagDiagram
	SNamePacketdiagDiagram
	SNameObjectDiagram
	SNameObject
	SNamePackage
	SNameParticipant
	SNamePartition
	SNamePerson
	SNamePort
	SNameProcess
	SNameQualified
	SNameQueue
	SNameRectangle
	SNameReference
	SNameReferenceHeader
	SNameRegex
	SNameRequirement
	SNameRnote
	SNameRoot
	SNameRootNode
	SNameSaltDiagram
	SNameSeparator
	SNameSequenceDiagram
	SNameServer
	SNameStack
	SNameStateDiagram
	SNameState
	SNameStateBody
	SNameStereotype
	SNameStorage
	SNameSwimlane
	SNameTask
	SNameTimegrid
	SNameTimeline
	SNameTimingDiagram
	SNameTitle
	SNameUndone
	SNameUnstarted
	SNameUsecase
	SNameVerticalSeparator
	SNameYear
	SNameVisibilityIcon
	SNamePrivate
	SNameProtected
	SNamePublic
	SNameIEMandatory
	SNameSpot
	SNameSpotAnnotation
	SNameSpotInterface
	SNameSpotEnum
	SNameSpotProtocol
	SNameSpotStruct
	SNameSpotEntity
	SNameSpotException
	SNameSpotClass
	SNameSpotAbstractClass
	SNameSpotMetaClass
	SNameSpotStereotype
	SNameSpotDataClass
	SNameSpotRecord
	SNameWbsDiagram
	SNameYamlDiagram
	SNameChartDiagram
	SNameBar
	SNameLine
	SNameArea
	SNameScatter
	SNameAxis
	SNameHAxis
	SNameVAxis
	SNameGrid
	SNameAnnotation
	snameCount
)

var snameNames = [...]string{
	"action", "activationBox", "activity", "activityBar", "activityDiagram", "actor", "agent", "analog",
	"archimate", "arrow", "artifact", "binary", "body", "boundary", "box", "boxless", "business", "caption",
	"card", "cardinality", "circle", "classDiagram", "class_", "clickable", "cloud", "closed", "collection",
	"collections", "component", "composite", "robust", "chenAttribute", "chenEerDiagram", "chenEntity",
	"chenRelationship", "concise", "clock", "componentDiagram", "constraintArrow", "control", "database", "day",
	"delay", "description", "destroy", "diamond", "document", "ebnf", "element", "entity", "end", "start",
	"stop", "file", "filesDiagram", "folder", "footer", "frame", "ganttDiagram", "generic", "goto_", "group",
	"groupHeader", "header", "hexagon", "highlight", "hnote", "interface_", "json", "jsonDiagram", "gitDiagram",
	"label", "leafNode", "legend", "lifeLine", "mainframe", "map", "milestone", "mindmapDiagram", "month",
	"name", "network", "newpage", "node", "note", "nwdiagDiagram", "packetdiagDiagram", "objectDiagram",
	"object", "package_", "participant", "partition", "person", "port", "process", "qualified", "queue",
	"rectangle", "reference", "referenceHeader", "regex", "requirement", "rnote", "root", "rootNode",
	"saltDiagram", "separator", "sequenceDiagram", "server", "stack", "stateDiagram", "state", "stateBody",
	"stereotype", "storage", "swimlane", "task", "timegrid", "timeline", "timingDiagram", "title", "undone",
	"unstarted", "usecase", "verticalSeparator", "year", "visibilityIcon", "private_", "protected_", "public_",
	"IEMandatory", "spot", "spotAnnotation", "spotInterface", "spotEnum", "spotProtocol", "spotStruct",
	"spotEntity", "spotException", "spotClass", "spotAbstractClass", "spotMetaClass", "spotStereotype",
	"spotDataClass", "spotRecord", "wbsDiagram", "yamlDiagram", "chartDiagram", "bar", "line", "area",
	"scatter", "axis", "hAxis", "vAxis", "grid", "annotation",
}

// String returns upstream's constant name, for example "class_".
func (s SName) String() string { return enumName("SName", snameNames[:], int(s)) }

// snameByKey maps the lower-case names without "_" to the names: read-only
// data built at initialization, upstream's ALL map.
var snameByKey = func() map[string]SName {
	m := make(map[string]SName, snameCount)
	for i, n := range snameNames {
		m[strings.ToLower(strings.ReplaceAll(n, "_", ""))] = SName(i)
	}
	return m
}()

// SNameRetrieve returns the name written s, ignoring case and the "_" of
// Java keywords; ok is false when there is none (null upstream).
func SNameRetrieve(s string) (SName, bool) {
	n, ok := snameByKey[javaLower(s)]
	return n, ok
}
