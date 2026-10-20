// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageBranch.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/image/EntityImageSynchroBar.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// branchSize is the half size of upstream's EntityImageBranch.
const branchSize = 12

// EntityImageBranch is upstream's EntityImageBranch: the diamond of a
// branch or choice.
type EntityImageBranch struct {
	abstractEntityImage
}

var _ klimt.IEntityImage = (*EntityImageBranch)(nil)

func newEntityImageBranch(entity *abel.Entity) *EntityImageBranch {
	return &EntityImageBranch{abstractEntityImage: newAbstractEntityImage(entity)}
}

// CalculateDimension returns 24 x 24.
func (e *EntityImageBranch) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(branchSize*2, branchSize*2)
}

// DrawU draws the diamond.
func (e *EntityImageBranch) DrawU(ug klimt.UGraphic) {
	diams := klimt.NewUPolygon()
	diams.AddPoint(branchSize, 0)
	diams.AddPoint(branchSize*2, branchSize)
	diams.AddPoint(branchSize, branchSize*2)
	diams.AddPoint(0, branchSize)
	diams.AddPoint(branchSize, 0)
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivity, klimt.SNameDiamond).MergedStyle(e.skinParam.CurrentStyleBuilder())
	border := style.Value(klimt.PNameLineColor).AsColor(e.skinParam.IHtmlColorSet())
	back := style.Value(klimt.PNameBackGroundColor).AsColor(e.skinParam.IHtmlColorSet())
	diams.SetDeltaShadow(style.Shadowing())
	// Upstream builds this group without the source line.
	group := klimt.NewUGroup()
	group.Put(klimt.UGroupClass, "entity")
	group.Put(klimt.UGroupID, "entity_"+e.entity.Name())
	group.Put(klimt.UGroupDataEntity, e.entity.Name())
	group.Put(klimt.UGroupDataUID, e.entity.UID())
	group.Put(klimt.UGroupDataQualifiedName, e.entity.Quark().QualifiedName())
	ug.StartGroup(group)
	ug.Apply(border).Apply(klimt.Bg(back)).Apply(style.Stroke()).Draw(diams)
	ug.CloseGroup()
}

// ShapeType returns DIAMOND.
func (e *EntityImageBranch) ShapeType() klimt.ShapeType { return klimt.ShapeTypeDiamond }

// EntityImageSynchroBar is upstream's EntityImageSynchroBar: the black bar
// of a synchronization, a fork or a join.
type EntityImageSynchroBar struct {
	abstractEntityImage
}

var _ klimt.IEntityImage = (*EntityImageSynchroBar)(nil)

func newEntityImageSynchroBar(entity *abel.Entity) *EntityImageSynchroBar {
	return &EntityImageSynchroBar{abstractEntityImage: newAbstractEntityImage(entity)}
}

// CalculateDimension returns 80 x 8, or 8 x 80 from left to right.
func (e *EntityImageSynchroBar) CalculateDimension(font.StringBounder) geom.XDimension2D {
	if e.skinParam.Rankdir() == geom.RankdirLeftToRight {
		return geom.NewXDimension2D(8, 80)
	}
	return geom.NewXDimension2D(80, 8)
}

// DrawU draws the bar.
func (e *EntityImageSynchroBar) DrawU(ug klimt.UGraphic) {
	dim := e.CalculateDimension(ug.StringBounder())
	rect := klimt.URectangleBuild(dim.Width(), dim.Height())
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameActivityBar).WithTOBECHANGED(e.entity.Stereotype()).MergedStyle(e.skinParam.CurrentStyleBuilder())
	back := style.Value(klimt.PNameBackGroundColor).AsColor(e.skinParam.IHtmlColorSet())
	rect.SetDeltaShadow(style.Shadowing())
	ug.Apply(color.None()).Apply(klimt.Bg(back)).Draw(rect)
}

// ShapeType returns RECTANGLE.
func (e *EntityImageSynchroBar) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }
