// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/AbstractEntityImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// abstractEntityImage is upstream's AbstractEntityImage: the entity of an
// image and the defaults of IEntityImage. Images embed it and keep their
// own TextBlockMemoized.
type abstractEntityImage struct {
	klimt.TextBlockDefaults
	entity    *abel.Entity
	skinParam klimt.ISkinParam
}

func newAbstractEntityImage(entity *abel.Entity) abstractEntityImage {
	if entity == nil {
		throwNPE()
	}
	return abstractEntityImage{entity: entity, skinParam: entity.SkinParam()}
}

// IsHidden reports whether the entity is hidden.
func (a *abstractEntityImage) IsHidden() bool { return a.entity.IsHidden() }

// Backcolor returns the background color of the diagram.
func (a *abstractEntityImage) Backcolor() color.HColor { return a.skinParam.BackgroundColor() }

func (a *abstractEntityImage) stereo() *klimt.Stereotype { return a.entity.Stereotype() }

// Shield returns no margin.
func (a *abstractEntityImage) Shield(font.StringBounder) klimt.Margins { return klimt.MarginsNone() }

// OverscanX returns 0.
func (a *abstractEntityImage) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash returns false.
func (a *abstractEntityImage) IsCrash() bool { return false }

func (a *abstractEntityImage) styleName() klimt.SName {
	return uml.DiagramTypeStyleName(a.entity.DiagramType())
}

// entityGroup is the group of an entity image: its class, id and data
// attributes.
func entityGroup(e *abel.Entity) *klimt.UGroup {
	group := klimt.NewUGroupAt(e.Location())
	group.Put(klimt.UGroupClass, "entity")
	group.Put(klimt.UGroupID, "entity_"+e.Name())
	group.Put(klimt.UGroupDataEntity, e.Name())
	group.Put(klimt.UGroupDataUID, e.UID())
	group.Put(klimt.UGroupDataQualifiedName, e.Quark().QualifiedName())
	return group
}
