// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/ClusterHeader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/abel/Entity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageStateCommon.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/PortionShower.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
	"github.com/ekalinin/gopuml/internal/uml"
)

// PortionShower tells which portions of an entity are shown (upstream's
// cucadiagram.PortionShower); the CucaDiagram implements it.
type PortionShower interface {
	ShowPortion(portion uml.EntityPortion, entity *abel.Entity) bool
	VisibleStereotypeLabels(entity *abel.Entity) []string
}

// ClusterHeader is the title of a group: its name, its stereotype and the
// room they take, which the DOT graph reserves as the label of the
// cluster.
type ClusterHeader struct {
	titleAndAttributeWidth  int
	titleAndAttributeHeight int
	title                   klimt.TextBlock
	stereo                  klimt.TextBlock
	g                       *abel.Entity
}

func newClusterHeader(g *abel.Entity, portionShower PortionShower, stringBounder font.StringBounder) *ClusterHeader {
	skinParam := g.SkinParam()
	h := &ClusterHeader{g: g}
	h.title = h.titleBlock()
	h.stereo = h.stereoBlock(portionShower)
	stereoAndTitle := klimt.TextBlockUtilsMergeTB(h.stereo, h.title, h.titleHorizontalAlignment())
	dimLabel := stereoAndTitle.CalculateDimension(stringBounder)
	if dimLabel.Width() > 0 {
		dimAttribute := stateDescription(g, skinParam).CalculateDimension(stringBounder)
		attributeHeight := dimAttribute.Height()
		attributeWidth := dimAttribute.Width()
		marginForFields := 0.0
		if attributeHeight > 0 {
			marginForFields = klimt.IEntityImageMargin
		}
		suppHeightBecauseOfShape, suppWidthBecauseOfShape := 0, 0
		if sym := usymbolOf(g.USymbol()); sym != nil {
			suppHeightBecauseOfShape = sym.SuppHeightBecauseOfShape()
			suppWidthBecauseOfShape = sym.SuppWidthBecauseOfShape()
		}
		h.titleAndAttributeWidth = klimt.JavaD2I(max(dimLabel.Width(), attributeWidth)) + suppWidthBecauseOfShape
		h.titleAndAttributeHeight = klimt.JavaD2I(dimLabel.Height() + attributeHeight + marginForFields +
			float64(suppHeightBecauseOfShape))
	}
	return h
}

// Title returns the name of the group.
func (h *ClusterHeader) Title() klimt.TextBlock { return h.title }

// Stereo returns the stereotype of the group.
func (h *ClusterHeader) Stereo() klimt.TextBlock { return h.stereo }

func (h *ClusterHeader) titleBlock() klimt.TextBlock {
	label := h.g.Display()
	if label == nil {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	style := h.style()
	fontConfiguration := style.FontConfiguration(h.g.SkinParam().IHtmlColorSet(), h.g.Colors())
	alignment := style.HorizontalAlignment()
	result := label.Create(fontConfiguration, alignment, h.g.SkinParam())
	if modifier, ok := h.g.VisibilityModifier(); ok {
		r := rose.NewRose()
		back := roseColorOpt(r, h.g.SkinParam(), modifier)
		fore := r.HtmlColor(h.g.SkinParam(), modifier.Foreground())
		uBlock := klimt.TextBlockUtilsWithMargin4(
			modifier.UBlock(h.g.SkinParam().ClassAttributeIconSize(), fore, back, false), 0, 0, 4, 0)
		result = klimt.TextBlockUtilsMergeLR(uBlock, result, geom.VerticalAlignmentCenter)
	}
	return result
}

func (h *ClusterHeader) style() *klimt.Style {
	return h.signature().WithTOBECHANGED(h.g.Stereotype()).With(h.g.Stereostyles()).
		MergedStyle(h.g.SkinParam().CurrentStyleBuilder())
}

func (h *ClusterHeader) signature() klimt.StyleSignatureBasic {
	sname := uml.DiagramTypeStyleName(h.g.SkinParam().DiagramType())
	uSym := usymbolOf(h.g.USymbol())
	switch {
	case h.g.GroupType() == abel.GroupTypeState:
		return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameStateDiagram,
			klimt.SNameState, klimt.SNameName)
	case uSym != nil:
		return klimt.StyleSignatureBasicOf3(klimt.SNameRoot, klimt.SNameElement, sname, uSym.SNames(),
			klimt.SNameComposite, klimt.SNameTitle)
	case h.g.GroupType() == abel.GroupTypePackage:
		return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, sname, klimt.SNamePackage,
			klimt.SNameTitle)
	}
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, sname, klimt.SNameComposite,
		klimt.SNameTitle)
}

func (h *ClusterHeader) titleHorizontalAlignment() geom.HorizontalAlignment {
	return h.style().HorizontalAlignment()
}

func (h *ClusterHeader) stereoBlock(portionShower PortionShower) klimt.TextBlock {
	stereo := h.stereoBlockWithoutLegend(portionShower)
	legend := h.g.Legend()
	if legend == nil || legend.IsNull() {
		return stereo
	}
	legendBlock := entityImageLegend(legend.Display(), h.g.SkinParam())
	return decorateEntityImageAdd(nil, legendBlock, stereo, legend.HorizontalAlignment(),
		legend.VerticalAlignment())
}

func (h *ClusterHeader) stereoBlockWithoutLegend(portionShower PortionShower) klimt.TextBlock {
	stereotype := h.g.Stereotype()
	if stereotype == nil {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	skinParam := h.g.SkinParam()
	// A composite state without a symbol is drawn by Cluster.drawUState,
	// which never draws this block (upstream issue 2783).
	if skinParam.DiagramType() == core.State && h.g.USymbol() == abel.USymbolNull {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	if tmp := stereotype.Sprite(skinParam); tmp != nil {
		return tmp
	}
	stereos := stereotype.Labels(skinParam.Guillemet())
	if stereos == nil {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	visibleStereotypes := portionShower.VisibleStereotypeLabels(h.g)
	if len(visibleStereotypes) == 0 {
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	style := clusterDefaultStyleDefinition(uml.DiagramTypeStyleName(skinParam.DiagramType()),
		usymbolOf(h.g.USymbol()), h.g.GroupType()).
		ForStereotypeItself(h.g.Stereotype()).MergedStyle(skinParam.CurrentStyleBuilder())
	fontConfiguration := style.FontConfiguration(skinParam.IHtmlColorSet(), nil)
	horizontalAlignment := h.titleHorizontalAlignment()
	return klimt.DisplayCreate(visibleStereotypes...).Create(fontConfiguration, horizontalAlignment, skinParam)
}

// stateDescription is Entity.getStateDescription: the description lines
// of a group, drawn with the state description style.
func stateDescription(g *abel.Entity, skinParam klimt.ISkinParam) klimt.TextBlock {
	if !g.IsGroup() {
		throwUnsupported("")
	}
	style := stateStyle().AddSName(klimt.SNameDescription).WithTOBECHANGED(g.Stereotype()).
		MergedStyle(skinParam.CurrentStyleBuilder())
	details := g.Bodier().RawBody()
	if len(details) == 0 {
		return klimt.NewTextBlockEmpty(0, 0)
	}
	fontConfiguration := klimt.FontConfigurationFromStyle(skinParam, style, nil)
	var display *klimt.Display
	for _, s := range details {
		d := klimt.DisplayWithNewlines(skinParam.Pragma(), s.String())
		if display == nil {
			display = d
		} else {
			display = display.AddAll(d)
		}
	}
	return display.Create(fontConfiguration, style.HorizontalAlignment(), skinParam)
}

// stateStyle is EntityImageStateCommon.STYLE.
func stateStyle() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameStateDiagram, klimt.SNameState)
}

// roseColorOpt is Rose.getHtmlColor of the background of a visibility
// modifier, which may be null.
func roseColorOpt(r *rose.Rose, skinParam klimt.ISkinParam, modifier klimt.VisibilityModifier) color.HColor {
	param, ok := modifier.Background()
	if !ok {
		return nil
	}
	return r.HtmlColor(skinParam, param)
}
