// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/GroupMakerActivity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/InnerActivity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/CucaDiagramSimplifierActivity.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// simplifyActivity is upstream's CucaDiagramSimplifierActivity.simplify:
// each autarkic group (an inner activity) is laid out on its own and
// becomes an activity drawn with that image, until none is left. A layout
// error stops it, as simplifyState.
func simplifyActivity(d *abel.CucaDiagram, stringBounder font.StringBounder, dotMode DotMode, le *layoutEngine,
	version dot.GraphvizVersion) (klimt.IEntityImage, error) {
	for changed := true; changed; {
		changed = false
		for _, g := range slices.Clone(d.Groups()) {
			if !g.IsAutarkic() {
				continue
			}
			img, err := newGroupMakerActivity(d, g, stringBounder, dotMode, le, version).image()
			if img == nil || err != nil {
				return img, err
			}
			g.OverrideImage(img, abel.LeafTypeActivity)
			changed = true
		}
	}
	return nil, nil
}

// groupMakerActivity is upstream's GroupMakerActivity: the image of an inner
// activity laid out on its own, with its own Graphviz run.
type groupMakerActivity struct {
	diagram       *abel.CucaDiagram
	group         *abel.Entity
	stringBounder font.StringBounder
	dotMode       DotMode
	le            *layoutEngine
	version       dot.GraphvizVersion
}

func newGroupMakerActivity(diagram *abel.CucaDiagram, group *abel.Entity, stringBounder font.StringBounder,
	dotMode DotMode, le *layoutEngine, version dot.GraphvizVersion) *groupMakerActivity {
	return &groupMakerActivity{diagram: diagram, group: group, stringBounder: stringBounder, dotMode: dotMode, le: le,
		version: version}
}

// innerActivityGroupHierarchy is GroupMakerActivity.InnerGroupHierarchy:
// the groups of the inner activity are the top groups of its layout.
type innerActivityGroupHierarchy struct {
	diagram *abel.CucaDiagram
	group   *abel.Entity
}

func (h innerActivityGroupHierarchy) ChildrenGroups(parent *abel.Entity) []*abel.Entity {
	if parent.IsRoot() {
		return h.diagram.ChildrenGroups(h.group)
	}
	return h.diagram.ChildrenGroups(parent)
}

func (h innerActivityGroupHierarchy) IsEmpty(g *abel.Entity) bool { return h.diagram.IsEmpty(g) }

// pureInnerLinks are the links between leaves of the group.
func (m *groupMakerActivity) pureInnerLinks() []*abel.Link {
	var result []*abel.Link
	for _, link := range m.diagram.Links() {
		e1, e2 := link.Entity1(), link.Entity2()
		if e1.ParentContainer() == m.group && !e1.IsGroup() && e2.ParentContainer() == m.group && !e2.IsGroup() {
			result = append(result, link)
		}
	}
	return result
}

// image returns the image of the inner activity. The error is a layout
// error, returned with the image that shows it, or the error of the
// context, without image.
func (m *groupMakerActivity) image() (klimt.IEntityImage, error) {
	if m.group.CountChildren() == 0 {
		return newEntityImageState(m.group), nil
	}
	links := m.pureInnerLinks()
	skinParam := m.diagram.SkinParam()
	bibliotekon := newBibliotekon(links)
	data := newDotData(m.diagram, m.group, links, m.group.Leafs(), innerActivityGroupHierarchy{m.diagram, m.group},
		portionShowerAll{})
	root := newRootCluster(m.diagram, bibliotekon.colorSequence, data.rootGroup())
	cm := &clusterManager{bibliotekon: bibliotekon, current: root}
	dotStringFactory := newDotStringFactory(bibliotekon, root, m.diagram.DiagramType(), skinParam, m.version)
	builder := &graphvizImageBuilder{dotData: data, pragma: m.diagram.Pragma(), styleName: klimt.SNameActivityDiagram,
		dotMode: m.dotMode, dotStringFactory: dotStringFactory, clusterManager: cm}
	if m.group.GroupType() != abel.GroupTypeInnerActivity {
		throwUnsupported(m.group.GroupType().String())
	}
	stereo := m.group.Stereotype()
	r := rose.NewRose()
	borderColor := r.HtmlColorStereo(skinParam, stereo, klimt.ColorParamActivityBorder)
	backColor := m.group.Colors().Color(color.TypeBack)
	if backColor == nil {
		backColor = r.HtmlColorStereo(skinParam, stereo, klimt.ColorParamBackground)
	}
	shadowing := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameGroup).MergedStyle(skinParam.CurrentStyleBuilder()).Shadowing()
	image, err := builder.buildImage(m.stringBounder, []string{}, m.le)
	if image == nil || err != nil {
		return image, err
	}
	return &InnerActivity{im: image, borderColor: borderColor, backColor: backColor, shadowing: shadowing}, nil
}

// innerActivityThicknessBorder is InnerActivity.THICKNESS_BORDER.
const innerActivityThicknessBorder = 1.5

// InnerActivity is upstream's InnerActivity: an inner activity laid out on
// its own, drawn in a rounded box.
type InnerActivity struct {
	klimt.TextBlockDefaults
	im                     klimt.IEntityImage
	borderColor, backColor color.HColor
	shadowing              float64
}

var _ klimt.IEntityImage = (*InnerActivity)(nil)

// DrawU draws the box and the inside.
func (a *InnerActivity) DrawU(ug klimt.UGraphic) {
	total := a.CalculateDimension(ug.StringBounder())
	ug = ug.Apply(klimt.Bg(a.backColor)).Apply(a.borderColor).Apply(klimt.UStrokeWithThickness(innerActivityThicknessBorder))
	rect := klimt.URectangleBuild(total.Width(), total.Height()).Rounded(klimt.IEntityImageCorner)
	rect.SetDeltaShadow(a.shadowing)
	ug.Draw(rect)
	ug = ug.Apply(klimt.UStrokeSimple())
	a.im.DrawU(ug)
}

// CalculateDimension returns the size of the inside.
func (a *InnerActivity) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.im.CalculateDimension(stringBounder)
}

// ShapeType returns ROUND_RECTANGLE.
func (a *InnerActivity) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRoundRectangle }

// Shield returns no margin.
func (a *InnerActivity) Shield(font.StringBounder) klimt.Margins { return klimt.MarginsNone() }

// IsHidden reports whether the inside is hidden.
func (a *InnerActivity) IsHidden() bool { return a.im.IsHidden() }

// OverscanX returns 0.
func (a *InnerActivity) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash returns false.
func (a *InnerActivity) IsCrash() bool { return false }
