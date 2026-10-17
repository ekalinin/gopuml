// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/symbol/USymbols.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ActorStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/ComponentStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/PackageStyle.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"regexp"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// USymbol names one of upstream's USymbols constants, the shape drawn
// around an element (node, cloud, database...), by its registry key, for
// example "NODE". The empty value stands for null.
//
// The model only needs the identity of the symbol; the drawing step maps
// the names to the shapes of internal/decoration/symbol.
type USymbol string

// The symbols of upstream's USymbols registry.
const (
	USymbolNull                  USymbol = ""
	USymbolAction                USymbol = "ACTION"
	USymbolActorAwesome          USymbol = "ACTOR_AWESOME"
	USymbolActorHollow           USymbol = "ACTOR_HOLLOW"
	USymbolActorStickman         USymbol = "ACTOR_STICKMAN"
	USymbolActorStickmanBusiness USymbol = "ACTOR_STICKMAN_BUSINESS"
	USymbolAgent                 USymbol = "AGENT"
	USymbolArchimate             USymbol = "ARCHIMATE"
	USymbolArtifact              USymbol = "ARTIFACT"
	USymbolBoundary              USymbol = "BOUNDARY"
	USymbolCard                  USymbol = "CARD"
	USymbolCloud                 USymbol = "CLOUD"
	USymbolCollections           USymbol = "COLLECTIONS"
	USymbolComponentRectangle    USymbol = "COMPONENT_RECTANGLE"
	USymbolComponent1            USymbol = "COMPONENT1"
	USymbolComponent2            USymbol = "COMPONENT2"
	USymbolControl               USymbol = "CONTROL"
	USymbolDatabase              USymbol = "DATABASE"
	USymbolEntityDomain          USymbol = "ENTITY_DOMAIN"
	USymbolFile                  USymbol = "FILE"
	USymbolFolder                USymbol = "FOLDER"
	USymbolFrame                 USymbol = "FRAME"
	USymbolGroup                 USymbol = "GROUP"
	USymbolHexagon               USymbol = "HEXAGON"
	USymbolInterface             USymbol = "INTERFACE"
	USymbolLabel                 USymbol = "LABEL"
	USymbolNode                  USymbol = "NODE"
	USymbolPackage               USymbol = "PACKAGE"
	USymbolPartition             USymbol = "PARTITION"
	USymbolPerson                USymbol = "PERSON"
	USymbolProcess               USymbol = "PROCESS"
	USymbolQueue                 USymbol = "QUEUE"
	USymbolRectangle             USymbol = "RECTANGLE"
	USymbolStack                 USymbol = "STACK"
	USymbolStorage               USymbol = "STORAGE"
	USymbolUsecase               USymbol = "USECASE"
	USymbolUsecaseBusiness       USymbol = "USECASE_BUSINESS"
)

// usymbolRegistry is upstream's USymbols.all, by key.
var usymbolRegistry = func() map[string]USymbol {
	result := map[string]USymbol{}
	for _, s := range []USymbol{USymbolAction, USymbolActorAwesome, USymbolActorHollow, USymbolActorStickman,
		USymbolActorStickmanBusiness, USymbolAgent, USymbolArchimate, USymbolArtifact, USymbolBoundary,
		USymbolCard, USymbolCloud, USymbolCollections, USymbolComponentRectangle, USymbolComponent1,
		USymbolComponent2, USymbolControl, USymbolDatabase, USymbolEntityDomain, USymbolFile, USymbolFolder,
		USymbolFrame, USymbolGroup, USymbolHexagon, USymbolInterface, USymbolLabel, USymbolNode, USymbolPackage,
		USymbolPartition, USymbolPerson, USymbolProcess, USymbolQueue, USymbolRectangle, USymbolStack,
		USymbolStorage, USymbolUsecase, USymbolUsecaseBusiness} {
		result[string(s)] = s
	}
	return result
}()

// String returns the registry key, "null" for USymbolNull.
func (s USymbol) String() string {
	if s == USymbolNull {
		return "null"
	}
	return string(s)
}

// ActorStyleUSymbol is upstream's ActorStyle.toUSymbol.
func ActorStyleUSymbol(style klimt.ActorStyle) USymbol {
	switch style {
	case klimt.ActorStyleStickman:
		return USymbolActorStickman
	case klimt.ActorStyleAwesome:
		return USymbolActorAwesome
	case klimt.ActorStyleHollow:
		return USymbolActorHollow
	case klimt.ActorStyleStickmanBusiness:
		return USymbolActorStickmanBusiness
	}
	uml.Throw(uml.IllegalStateException, "")
	return USymbolNull
}

// ComponentStyleUSymbol is upstream's ComponentStyle.toUSymbol.
func ComponentStyleUSymbol(style klimt.ComponentStyle) USymbol {
	switch style {
	case klimt.ComponentStyleUML1:
		return USymbolComponent1
	case klimt.ComponentStyleUML2:
		return USymbolComponent2
	}
	return USymbolComponentRectangle
}

// PackageStyleUSymbol is upstream's PackageStyle.toUSymbol: null for the
// styles without a symbol.
func PackageStyleUSymbol(style klimt.PackageStyle) USymbol {
	switch style {
	case klimt.PackageStyleNode:
		return USymbolNode
	case klimt.PackageStyleCard:
		return USymbolCard
	case klimt.PackageStyleDatabase:
		return USymbolDatabase
	case klimt.PackageStyleCloud:
		return USymbolCloud
	case klimt.PackageStyleFrame:
		return USymbolFrame
	case klimt.PackageStyleRectangle:
		return USymbolRectangle
	case klimt.PackageStyleFolder:
		return USymbolPackage
	}
	return USymbolNull
}

var nonWord = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// USymbolFromStyles is upstream's USymbols.fromString(String, ActorStyle,
// ComponentStyle, PackageStyle), used for the stereotypes and keywords of
// containers: "package", "actor" and "component" follow the styles, other
// names are looked up without their non-word characters. ok false (null)
// and unknown names give null.
func USymbolFromStyles(s string, ok bool, actorStyle klimt.ActorStyle, componentStyle klimt.ComponentStyle,
	packageStyle klimt.PackageStyle) USymbol {
	if !ok {
		return USymbolNull
	}
	switch {
	case uml.EqualsIgnoreCase(s, "package"):
		return PackageStyleUSymbol(packageStyle)
	case uml.EqualsIgnoreCase(s, "actor"):
		return ActorStyleUSymbol(actorStyle)
	case uml.EqualsIgnoreCase(s, "component"):
		return ComponentStyleUSymbol(componentStyle)
	case uml.EqualsIgnoreCase(s, "entity"):
		return USymbolEntityDomain
	case uml.EqualsIgnoreCase(s, "circle"):
		return USymbolInterface
	}
	return usymbolRegistry[uml.UpperCase(nonWord.ReplaceAllString(s, ""))]
}

// USymbolFromSkin is upstream's USymbols.fromString(String, ISkinParam),
// used for the keywords of elements: null for an unknown keyword.
func USymbolFromSkin(symbol string, skinParam klimt.ISkinParam) USymbol {
	is := func(s string) bool { return uml.EqualsIgnoreCase(symbol, s) }
	switch {
	case is("artifact"):
		return USymbolArtifact
	case is("folder"):
		return USymbolFolder
	case is("file"):
		return USymbolFile
	case is("package"):
		return USymbolPackage
	case is("rectangle"):
		return USymbolRectangle
	case is("person"):
		return USymbolPerson
	case is("hexagon"):
		return USymbolHexagon
	case is("label"):
		return USymbolLabel
	case is("collections"):
		return USymbolCollections
	case is("node"):
		return USymbolNode
	case is("frame"):
		return USymbolFrame
	case is("cloud"):
		return USymbolCloud
	case is("action"):
		return USymbolAction
	case is("process"):
		return USymbolProcess
	case is("database"):
		return USymbolDatabase
	case is("queue"):
		return USymbolQueue
	case is("stack"):
		return USymbolStack
	case is("storage"):
		return USymbolStorage
	case is("agent"):
		return USymbolAgent
	case is("actor/"):
		return USymbolActorStickmanBusiness
	case is("actor"):
		return ActorStyleUSymbol(skinParam.ActorStyle())
	case is("component"):
		return ComponentStyleUSymbol(skinParam.ComponentStyle())
	case is("boundary"):
		return USymbolBoundary
	case is("control"):
		return USymbolControl
	case is("entity"):
		return USymbolEntityDomain
	case is("card"):
		return USymbolCard
	case is("interface"):
		return USymbolInterface
	case is("()"):
		return USymbolInterface
	}
	return USymbolNull
}
