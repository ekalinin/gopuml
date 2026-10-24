// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/DescriptionDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package descdiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
)

// DescriptionDiagram is a parsed description diagram: a use case,
// component or deployment diagram.
type DescriptionDiagram struct {
	abel.AbstractEntityDiagram
}

var (
	_ core.TypedDiagram = (*DescriptionDiagram)(nil)
	_ abel.Diagram      = (*DescriptionDiagram)(nil)
)

// NewDescriptionDiagram returns an empty description diagram for src;
// md5map holds the base64 images of the source.
func NewDescriptionDiagram(ctx *core.RenderContext, src *core.UmlSource,
	md5map map[string]string) *DescriptionDiagram {
	d := &DescriptionDiagram{}
	d.InitCucaDiagram(ctx, src, core.Description, md5map, descriptionHooks{})
	return d
}

// descriptionHooks are the CucaDiagram methods DescriptionDiagram
// overrides.
type descriptionHooks struct{}

func (descriptionHooks) UpdateLasts(*abel.Entity)             {}
func (descriptionHooks) IsHideEmptyDescriptionForState() bool { return false }
func (descriptionHooks) CleanID(id string) string             { return cleanID(id) }

// cleanID is upstream's DescriptionDiagram.cleanId: it removes the "()" of
// an interface and the markers of the business actor (":name:/") and use
// case ("(name)/"), then the quotes.
func cleanID(id string) string {
	if strings.HasPrefix(id, "()") {
		id = text.Trin(id[2:])
	}
	if strings.HasPrefix(id, ":") && strings.HasSuffix(id, ":/") {
		return id[1 : len(id)-2]
	}
	if strings.HasPrefix(id, "(") && strings.HasSuffix(id, ")/") {
		return id[1 : len(id)-2]
	}
	return abel.DefaultCleanID(id)
}

// isUsecase reports whether the diagram has a use case or an actor drawn
// with the actor style of the skin.
func (d *DescriptionDiagram) isUsecase() bool {
	actor := abel.ActorStyleUSymbol(d.SkinParam().ActorStyle())
	for _, leaf := range d.Leafs() {
		if leaf.LeafType() == abel.LeafTypeUsecase || leaf.USymbol() == actor {
			return true
		}
	}
	return false
}

// MakeDiagramReady implements command.Diagram: the entities whose type is
// still unknown (named in links only) become description elements, drawn
// as actors in a use case diagram and as interfaces otherwise.
func (d *DescriptionDiagram) MakeDiagramReady() {
	d.AbstractEntityDiagram.MakeDiagramReady()
	defaultSymbol := abel.USymbolInterface
	if d.isUsecase() {
		defaultSymbol = abel.ActorStyleUSymbol(d.SkinParam().ActorStyle())
	}
	for _, leaf := range d.Leafs() {
		if leaf.LeafType() == abel.LeafTypeStillUnknown {
			leaf.MuteToTypeSymbol(abel.LeafTypeDescription, defaultSymbol)
		}
	}
}

// CheckFinalError implements command.Diagram: the intermediate packages
// are packed when "!pragma useIntermediatePackages false" is set, and the
// standalone entities are put in squares.
func (d *DescriptionDiagram) CheckFinalError() string {
	if d.Pragma().IsFalse(klimt.PragmaUseIntermediatePackages) {
		d.PackSomePackage()
	}
	d.ApplySingleStrategy()
	return ""
}
