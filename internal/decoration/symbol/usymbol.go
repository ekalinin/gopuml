// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbol.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbols.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/PackageStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ComponentStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ActorStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package symbol

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// USymbol is a shape drawn around a text.
type USymbol interface {
	// SNames returns the style names of the symbol.
	SNames() []klimt.SName
	// AsSmall returns the symbol around a label and a stereotype, sized to
	// fit them; name is the title of the symbols that show one.
	AsSmall(name, label, stereotype klimt.TextBlock, symbolContext *klimt.Fashion,
		stereoAlignment geom.HorizontalAlignment) klimt.TextBlock
	// AsBig returns the symbol of a given size with a title and a
	// stereotype at its top, as a container.
	AsBig(label klimt.TextBlock, labelAlignment geom.HorizontalAlignment, stereotype klimt.TextBlock,
		width, height float64, symbolContext *klimt.Fashion, stereoAlignment geom.HorizontalAlignment) klimt.TextBlock
	// SuppHeightBecauseOfShape returns the extra height of the shape.
	SuppHeightBecauseOfShape() int
	// SuppWidthBecauseOfShape returns the extra width of the shape.
	SuppWidthBecauseOfShape() int
}

// usymbolBase holds upstream's default methods of USymbol.
type usymbolBase struct{}

// SuppHeightBecauseOfShape returns 0, the default.
func (usymbolBase) SuppHeightBecauseOfShape() int { return 0 }

// SuppWidthBecauseOfShape returns 0, the default.
func (usymbolBase) SuppWidthBecauseOfShape() int { return 0 }

// margin is upstream's USymbol.Margin: the space around the text of a
// symbol.
type margin struct {
	x1, x2, y1, y2 float64
}

func newMargin(x1, x2, y1, y2 float64) margin { return margin{x1: x1, x2: x2, y1: y1, y2: y2} }

func (m margin) addDimension(dim geom.XDimension2D) geom.XDimension2D {
	return geom.NewXDimension2D(dim.Width()+m.x1+m.x2, dim.Height()+m.y1+m.y2)
}

// The symbols of upstream's USymbols registry. They are stateless values,
// shared by every render; each is a distinct value, as upstream's
// instances are, so they can be compared with ==.
var (
	// Action is upstream's USymbols.ACTION.
	Action USymbol = &usymbolAction{sname: klimt.SNameAction}
	// ActorAwesome is upstream's USymbols.ACTOR_AWESOME.
	ActorAwesome USymbol = &usymbolActor{actorStyle: klimt.ActorStyleAwesome}
	// ActorHollow is upstream's USymbols.ACTOR_HOLLOW.
	ActorHollow USymbol = &usymbolActor{actorStyle: klimt.ActorStyleHollow}
	// ActorStickman is upstream's USymbols.ACTOR_STICKMAN.
	ActorStickman USymbol = &usymbolActor{actorStyle: klimt.ActorStyleStickman}
	// ActorStickmanBusiness is upstream's USymbols.ACTOR_STICKMAN_BUSINESS.
	ActorStickmanBusiness USymbol = &usymbolActorBusiness{}
	// Agent is upstream's USymbols.AGENT.
	Agent USymbol = &usymbolRectangle{sname: klimt.SNameAgent}
	// Archimate is upstream's USymbols.ARCHIMATE.
	Archimate USymbol = &usymbolRectangle{sname: klimt.SNameArchimate}
	// Artifact is upstream's USymbols.ARTIFACT.
	Artifact USymbol = usymbolArtifact{}
	// Boundary is upstream's USymbols.BOUNDARY.
	Boundary USymbol = &usymbolBoundary{}
	// Card is upstream's USymbols.CARD.
	Card USymbol = usymbolCard{}
	// Cloud is upstream's USymbols.CLOUD.
	Cloud USymbol = usymbolCloud{}
	// Collections is upstream's USymbols.COLLECTIONS.
	Collections USymbol = usymbolCollections{}
	// ComponentRectangle is upstream's USymbols.COMPONENT_RECTANGLE.
	ComponentRectangle USymbol = &usymbolRectangle{sname: klimt.SNameComponent}
	// Component1 is upstream's USymbols.COMPONENT1.
	Component1 USymbol = usymbolComponent1{}
	// Component2 is upstream's USymbols.COMPONENT2.
	Component2 USymbol = usymbolComponent2{}
	// Control is upstream's USymbols.CONTROL.
	Control USymbol = &usymbolControl{}
	// Database is upstream's USymbols.DATABASE.
	Database USymbol = usymbolDatabase{}
	// EntityDomain is upstream's USymbols.ENTITY_DOMAIN.
	EntityDomain USymbol = &usymbolEntityDomain{}
	// File is upstream's USymbols.FILE.
	File USymbol = usymbolFile{}
	// Folder is upstream's USymbols.FOLDER.
	Folder USymbol = &usymbolFolder{sname: klimt.SNameFolder, showTitle: false}
	// Frame is upstream's USymbols.FRAME.
	Frame USymbol = &usymbolFrame{sname: klimt.SNameFrame}
	// Group is upstream's USymbols.GROUP.
	Group USymbol = &usymbolFrame{sname: klimt.SNameGroup}
	// Hexagon is upstream's USymbols.HEXAGON.
	Hexagon USymbol = usymbolHexagon{}
	// Interface is upstream's USymbols.INTERFACE.
	Interface USymbol = &usymbolInterface{}
	// Label is upstream's USymbols.LABEL.
	Label USymbol = usymbolLabel{}
	// Node is upstream's USymbols.NODE.
	Node USymbol = &usymbolNode{}
	// Package is upstream's USymbols.PACKAGE: a folder that shows the title
	// in its tab.
	Package USymbol = &usymbolFolder{sname: klimt.SNamePackage, showTitle: true}
	// Partition is upstream's USymbols.PARTITION.
	Partition USymbol = &usymbolFrame{sname: klimt.SNamePartition}
	// Person is upstream's USymbols.PERSON.
	Person USymbol = usymbolPerson{}
	// Process is upstream's USymbols.PROCESS.
	Process USymbol = &usymbolProcess{sname: klimt.SNameProcess}
	// Queue is upstream's USymbols.QUEUE.
	Queue USymbol = usymbolQueue{}
	// Rectangle is upstream's USymbols.RECTANGLE.
	Rectangle USymbol = &usymbolRectangle{sname: klimt.SNameRectangle}
	// Stack is upstream's USymbols.STACK.
	Stack USymbol = usymbolStack{}
	// Storage is upstream's USymbols.STORAGE.
	Storage USymbol = usymbolStorage{}
	// Usecase is upstream's USymbols.USECASE.
	Usecase USymbol = &usymbolUsecase{isBusiness: false}
	// UsecaseBusiness is upstream's USymbols.USECASE_BUSINESS.
	UsecaseBusiness USymbol = &usymbolUsecase{isBusiness: true}
)

// registry is upstream's USymbols.all, by registry key.
var registry = map[string]USymbol{
	"ACTION": Action, "ACTOR_AWESOME": ActorAwesome, "ACTOR_HOLLOW": ActorHollow, "ACTOR_STICKMAN": ActorStickman,
	"ACTOR_STICKMAN_BUSINESS": ActorStickmanBusiness, "AGENT": Agent, "ARCHIMATE": Archimate, "ARTIFACT": Artifact,
	"BOUNDARY": Boundary, "CARD": Card, "CLOUD": Cloud, "COLLECTIONS": Collections,
	"COMPONENT_RECTANGLE": ComponentRectangle, "COMPONENT1": Component1, "COMPONENT2": Component2,
	"CONTROL": Control, "DATABASE": Database, "ENTITY_DOMAIN": EntityDomain, "FILE": File, "FOLDER": Folder,
	"FRAME": Frame, "GROUP": Group, "HEXAGON": Hexagon, "INTERFACE": Interface, "LABEL": Label, "NODE": Node,
	"PACKAGE": Package, "PARTITION": Partition, "PERSON": Person, "PROCESS": Process, "QUEUE": Queue,
	"RECTANGLE": Rectangle, "STACK": Stack, "STORAGE": Storage, "USECASE": Usecase,
	"USECASE_BUSINESS": UsecaseBusiness,
}

// FromKey returns the symbol registered under key, the registry key of
// upstream's USymbols (for example "NODE"), as the entity model
// (internal/abel's USymbol) names symbols. It returns nil for "" (null) and
// for unknown keys.
func FromKey(key string) USymbol { return registry[key] }

// PackageStyleToUSymbol is upstream's PackageStyle.toUSymbol: nil (null)
// for the styles without a symbol.
func PackageStyleToUSymbol(style klimt.PackageStyle) USymbol {
	switch style {
	case klimt.PackageStyleNode:
		return Node
	case klimt.PackageStyleCard:
		return Card
	case klimt.PackageStyleDatabase:
		return Database
	case klimt.PackageStyleCloud:
		return Cloud
	case klimt.PackageStyleFrame:
		return Frame
	case klimt.PackageStyleRectangle:
		return Rectangle
	case klimt.PackageStyleFolder:
		return Package
	}
	return nil
}

// ComponentStyleToUSymbol is upstream's ComponentStyle.toUSymbol.
func ComponentStyleToUSymbol(style klimt.ComponentStyle) USymbol {
	switch style {
	case klimt.ComponentStyleUML1:
		return Component1
	case klimt.ComponentStyleUML2:
		return Component2
	case klimt.ComponentStyleRectangle:
		return ComponentRectangle
	}
	panic(&color.RuntimeError{Class: "java.lang.AssertionError"})
}

// ActorStyleToUSymbol is upstream's ActorStyle.toUSymbol.
func ActorStyleToUSymbol(style klimt.ActorStyle) USymbol {
	switch style {
	case klimt.ActorStyleStickman:
		return ActorStickman
	case klimt.ActorStyleAwesome:
		return ActorAwesome
	case klimt.ActorStyleHollow:
		return ActorHollow
	case klimt.ActorStyleStickmanBusiness:
		return ActorStickmanBusiness
	}
	panic(&color.RuntimeError{Class: "java.lang.IllegalStateException"})
}

// throwNPE panics as upstream's Objects.requireNonNull.
func throwNPE() { panic(&color.RuntimeError{Class: "java.lang.NullPointerException"}) }

// throwUnsupported panics as upstream's UnsupportedOperationException.
func throwUnsupported() { panic(&color.RuntimeError{Class: "java.lang.UnsupportedOperationException"}) }

// illegalArgument is upstream's IllegalArgumentException.
func illegalArgument() *color.RuntimeError {
	return &color.RuntimeError{Class: "java.lang.IllegalArgumentException"}
}

// javaIntCast is Java's (int) cast of a double: truncation toward zero, 0
// for NaN, saturated outside the int range.
func javaIntCast(f float64) int {
	switch {
	case math.IsNaN(f):
		return 0
	case f >= math.MaxInt32:
		return math.MaxInt32
	case f <= math.MinInt32:
		return math.MinInt32
	}
	return int(f)
}
