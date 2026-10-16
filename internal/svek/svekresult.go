// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/SvekResult.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Untranslated.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// untranslated is upstream's Untranslated: an image that also draws in the
// coordinates of the whole diagram.
type untranslated interface {
	drawUntranslated(ug klimt.UGraphic, minX, minY float64)
}

// SvekResult is the drawing of a laid out graph: the clusters, then the
// nodes, then the edges.
type SvekResult struct {
	klimt.TextBlockDefaults
	dotData        *dotData
	clusterManager *DotStringFactory
	minMax         *geom.MinMax
}

var _ klimt.IEntityImage = (*SvekResult)(nil)

func newSvekResult(dotData *dotData, clusterManager *DotStringFactory) *SvekResult {
	return &SvekResult{dotData: dotData, clusterManager: clusterManager}
}

// DrawU draws the graph.
func (r *SvekResult) DrawU(ug klimt.UGraphic) {
	bibliotekon := r.clusterManager.bibliotekon
	for _, cluster := range bibliotekon.allCluster {
		if !cluster.group.IsPacked() {
			cluster.DrawU(ug)
		}
	}
	style2 := r.defaultStyleDefinition(nil).MergedStyle(r.dotData.skinParam().CurrentStyleBuilder())
	borderColor := color.NoGradient(style2.Value(klimt.PNameLineColor).AsColor(r.dotData.skinParam().IHtmlColorSet()))

	for _, node := range bibliotekon.allNodes() {
		minX, minY := node.minX, node.minY
		ug2 := ug
		if node.IsHidden() {
			ug2 = ug.Apply(klimt.UHiddenHidden)
		}
		image := node.image
		image.DrawU(ug2.Apply(klimt.NewUTranslate(minX, minY)))
		if u, ok := image.(untranslated); ok {
			u.drawUntranslated(ug.Apply(borderColor), minX, minY)
		}
	}

	ids := map[string]bool{}
	r.computeKal()
	for _, svekEdge := range bibliotekon.allLines {
		ug2 := ug
		if svekEdge.IsHidden() {
			ug2 = ug.Apply(klimt.UHiddenHidden)
		}
		svekEdge.ids = ids
		svekEdge.DrawU(ug2)
	}
}

func (r *SvekResult) computeKal() {
	for _, line := range r.clusterManager.bibliotekon.allLines {
		line.computeKal()
	}
	for _, node := range r.clusterManager.bibliotekon.allNodes() {
		node.fixOverlap()
	}
}

func (r *SvekResult) defaultStyleDefinition(stereotype *klimt.Stereotype) klimt.StyleSignature {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement,
		uml.DiagramTypeStyleName(r.dotData.diagramType()), klimt.SNameArrow).WithTOBECHANGED(stereotype)
}

// Backcolor returns the background color of the document.
func (r *SvekResult) Backcolor() color.HColor {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument).
		MergedStyle(r.dotData.skinParam().CurrentStyleBuilder())
	return style.Value(klimt.PNameBackGroundColor).AsColor(r.dotData.skinParam().IHtmlColorSet())
}

// CalculateDimension returns the size of the drawing; the first call moves
// the graph so that it starts at (6, 6).
func (r *SvekResult) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	if r.minMax == nil {
		mm := klimt.TextBlockUtilsMinMax(r, stringBounder, false)
		r.minMax = &mm
		r.clusterManager.MoveDelta(6-mm.MinX(), 6-mm.MinY())
	}
	return r.minMax.Dimension().DeltaWH(15, 15)
}

// ShapeType returns RECTANGLE.
func (r *SvekResult) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// Shield returns no margin.
func (r *SvekResult) Shield(font.StringBounder) klimt.Margins { return klimt.MarginsNone() }

// IsHidden returns false.
func (r *SvekResult) IsHidden() bool { return false }

// OverscanX returns 0.
func (r *SvekResult) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash returns false.
func (r *SvekResult) IsCrash() bool { return false }
