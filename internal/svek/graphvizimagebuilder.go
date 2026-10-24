// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/GraphvizImageBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// graphvizImageBuilder builds the image of a diagram: the images of its
// entities, the DOT graph, the layout by the engine and the result.
type graphvizImageBuilder struct {
	dotData          *dotData
	dotMode          DotMode
	pragma           *klimt.Pragma
	styleName        klimt.SName
	dotStringFactory *DotStringFactory
	clusterManager   *clusterManager
}

func (b *graphvizImageBuilder) defaultStyleDefinitionArrow(stereotype *klimt.Stereotype) klimt.StyleSignature {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, b.styleName, klimt.SNameArrow).
		WithTOBECHANGED(stereotype)
}

func (b *graphvizImageBuilder) styleArrowCardinality(stereotype *klimt.Stereotype) klimt.StyleSignature {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, b.styleName, klimt.SNameArrow,
		klimt.SNameCardinality).WithTOBECHANGED(stereotype)
}

func (b *graphvizImageBuilder) isOpalisable(entity *abel.Entity) bool {
	if b.dotData.skinParam().StrictUmlStyle() {
		return false
	}
	if entity.IsGroup() {
		return false
	}
	if entity.LeafType() != abel.LeafTypeNote {
		return false
	}
	single := b.onlyOneLink(entity)
	if single == nil {
		return false
	}
	return single.Other(entity).LeafType() != abel.LeafTypeNote
}

// entityImageSimpleEmpty is the image of a diagram without entity.
type entityImageSimpleEmpty struct {
	klimt.TextBlockDefaults
	backColor color.HColor
}

func (e *entityImageSimpleEmpty) IsHidden() bool          { return false }
func (e *entityImageSimpleEmpty) Backcolor() color.HColor { return e.backColor }
func (e *entityImageSimpleEmpty) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(10, 10)
}
func (e *entityImageSimpleEmpty) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return geom.MinMaxFromDim(e.CalculateDimension(stringBounder))
}
func (e *entityImageSimpleEmpty) InnerPosition(string, font.StringBounder) *geom.XRectangle2D {
	return nil
}
func (e *entityImageSimpleEmpty) DrawU(klimt.UGraphic)       {}
func (e *entityImageSimpleEmpty) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }
func (e *entityImageSimpleEmpty) Shield(font.StringBounder) klimt.Margins {
	return klimt.MarginsNone()
}
func (e *entityImageSimpleEmpty) OverscanX(font.StringBounder) float64 { return 0 }
func (e *entityImageSimpleEmpty) IsCrash() bool                        { return false }

func (b *graphvizImageBuilder) backcolor() color.HColor {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument).
		MergedStyle(b.dotData.skinParam().CurrentStyleBuilder())
	return style.Value(klimt.PNameBackGroundColor).AsColor(b.dotData.skinParam().IHtmlColorSet())
}

// layoutEngine is the engine that lays the DOT graph out, and what is
// known about it before it runs.
type layoutEngine struct {
	ctx    context.Context
	engine dot.Engine
	// notFound is set when no usable dot executable was found: the
	// diagram then shows upstream's error instead of a layout.
	notFound *dot.NotFoundError
	// trace receives the DOT graph and the SVG output when it is not nil
	// (upstream's -debugsvek).
	trace func(dotString, svg string)
}

// buildImage builds the image of the diagram. The error is a layout error
// (dot missing, failing or timed out), returned together with an image
// that shows it, or the error of the context, without image.
func (b *graphvizImageBuilder) buildImage(stringBounder font.StringBounder, dotStrings []string,
	le *layoutEngine) (klimt.IEntityImage, error) {
	if b.dotData.isDegeneratedWithFewEntities(0) {
		return &entityImageSimpleEmpty{backColor: b.dotData.skinParam().BackgroundColor()}, nil
	}
	if b.dotData.isDegeneratedWithFewEntities(1) && b.dotData.diagramType() != core.State {
		single := b.dotData.leafs[0]
		group := single.ParentContainer()
		if group.IsRoot() && single.USymbol() != abel.USymbolHexagon {
			tmp := createEntityImageBlock(single, b.dotData.isHideEmptyDescriptionForState(),
				b.dotData.portionShower, nil, nil, b.dotData.links)
			return &EntityImageDegenerated{orig: tmp, backcolor: b.backcolor()}, nil
		}
	}
	b.dotData.removeIrrelevantSametail(b.dotStringFactory.bibliotekon)

	b.printGroups(stringBounder, b.dotData.rootGroup())
	b.printEntities(stringBounder, b.unpackagedEntities())

	bibliotekon := b.dotStringFactory.bibliotekon
	for _, link := range b.dotData.links {
		if link.IsRemoved() {
			continue
		}
		b.addLink(link, stringBounder, bibliotekon)
	}

	if le.notFound != nil {
		return graphvizError(le.notFound.Path, le.notFound.State), le.notFound
	}

	dotString := b.dotStringFactory.createDotString(stringBounder, b.dotMode, dotStrings)
	out, err := le.engine.Layout(le.ctx, dotString)
	if err != nil {
		if ctxErr := le.ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		var nf *dot.NotFoundError
		if errors.As(err, &nf) {
			return graphvizError(nf.Path, nf.State), err
		}
		return layoutCrash(err), err
	}
	svg := string(out)
	if le.trace != nil {
		le.trace(dotString, svg)
	}
	if len(svg) == 0 {
		return layoutCrash(errEmptySvg), &dot.FailureError{Reason: "empty output"}
	}
	b.dotStringFactory.solve(svg)
	return newSvekResult(b.dotData, b.dotStringFactory), nil
}

// addLink creates the edge of a link, and turns a note attached by its
// only link into an opale.
func (b *graphvizImageBuilder) addLink(link *abel.Link, stringBounder font.StringBounder, bibliotekon *Bibliotekon) {
	defer func() {
		// Upstream logs an IllegalStateException and goes on.
		if r := recover(); r != nil {
			if !isIllegalState(r) {
				panic(r)
			}
		}
	}()
	skinParam := b.dotData.skinParam()
	labelFont := b.defaultStyleDefinitionArrow(link.Stereotype()).MergedStyle(link.StyleBuilder()).
		FontConfiguration(skinParam.IHtmlColorSet(), nil)
	cardinalityFont := b.styleArrowCardinality(link.Stereotype()).MergedStyle(link.StyleBuilder()).
		FontConfiguration(skinParam.IHtmlColorSet(), nil)
	line := newSvekEdge(link, skinParam, stringBounder, labelFont, cardinalityFont, bibliotekon, b.pragma,
		b.dotStringFactory.graphvizVersion)
	bibliotekon.addLine(line)
	switch {
	case b.isOpalisable(link.Entity1()):
		node := bibliotekon.node(link.Entity1())
		if other := bibliotekon.node(link.Entity2()); other != nil {
			node.image.(*EntityImageNote).setOpaleLine(line, node, other)
			line.setOpale(true)
		}
	case b.isOpalisable(link.Entity2()):
		node := bibliotekon.node(link.Entity2())
		if other := bibliotekon.node(link.Entity1()); other != nil {
			node.image.(*EntityImageNote).setOpaleLine(line, node, other)
			line.setOpale(true)
		}
	}
}

func isIllegalState(r any) bool {
	e, ok := r.(*color.RuntimeError)
	return ok && e.Class == "java.lang.IllegalStateException"
}

var graphvizVersionPattern = regexp.MustCompile(`(?mi)!-- generated by graphviz(.*)`)

// extractGraphvizVersion returns the version that dot wrote in its output.
func extractGraphvizVersion(svg string) (string, bool) {
	if m := graphvizVersionPattern.FindStringSubmatch(svg); m != nil {
		return strings.TrimSpace(m[1]), true
	}
	return "", false
}

func (b *graphvizImageBuilder) onlyOneLink(ent *abel.Entity) *abel.Link {
	var single *abel.Link
	for _, link := range b.dotData.links {
		if link.IsInvis() {
			continue
		}
		if !link.Contains(ent) {
			continue
		}
		if single != nil {
			return nil
		}
		single = link
	}
	return single
}

// graphvizError is upstream's error(File): the text of the error diagram
// for a dot executable that cannot be used.
func graphvizError(dotExe string, state dot.ExeState) klimt.IEntityImage {
	return klimt.GraphicStringsCreateForError(dot.NotFoundLines(dotExe, state), false)
}

// layoutCrash is the image of a dot run that failed or timed out. Upstream
// draws a crash report without dot's message (or does not stop dot); per
// design D16 the image shows the message.
func layoutCrash(err error) klimt.IEntityImage {
	lines := []string{"An error has occured while laying the diagram out with Graphviz:", " "}
	for line := range strings.SplitSeq(err.Error(), "\n") {
		lines = append(lines, line)
	}
	return klimt.GraphicStringsCreateForError(lines, false)
}

func (b *graphvizImageBuilder) printEntities(stringBounder font.StringBounder, entities []*abel.Entity) {
	for _, ent := range entities {
		if ent.IsRemoved() {
			continue
		}
		b.printEntity(stringBounder, ent)
	}
}

func (b *graphvizImageBuilder) printEntity(stringBounder font.StringBounder, ent *abel.Entity) {
	if ent.IsRemoved() {
		throwIllegalState(ent.String())
	}
	image := b.printEntityInternal(stringBounder, ent)
	node := b.dotStringFactory.bibliotekon.createNode(ent, image, stringBounder)
	b.clusterManager.addNode(node)
}

func (b *graphvizImageBuilder) printEntityInternal(stringBounder font.StringBounder,
	ent *abel.Entity) klimt.IEntityImage {
	if ent.IsRemoved() {
		throwIllegalState("")
	}
	if ent.SvekImage() == nil {
		skinParam := b.dotData.skinParam()
		if skinParam.SameClassWidth() {
			width := b.maxWidth(stringBounder)
			skinParam.(interface{ SetParamSameClassWidth(float64) }).SetParamSameClassWidth(width)
		}
		return createEntityImageBlock(ent, b.dotData.isHideEmptyDescriptionForState(), b.dotData.portionShower,
			b.dotStringFactory.bibliotekon, b.dotStringFactory.graphvizVersion, b.dotData.links)
	}
	return ent.SvekImage()
}

func (b *graphvizImageBuilder) maxWidth(stringBounder font.StringBounder) float64 {
	result := 0.0
	for _, ent := range b.dotData.leafs {
		if !ent.LeafType().IsLikeClass() {
			continue
		}
		im := newEntityImageClass(ent, b.dotData.portionShower)
		if w := im.CalculateDimension(stringBounder).Width(); w > result {
			result = w
		}
	}
	return result
}

func (b *graphvizImageBuilder) unpackagedEntities() []*abel.Entity {
	var result []*abel.Entity
	for _, ent := range b.dotData.leafs {
		if b.dotData.topParent == ent.ParentContainer() {
			result = append(result, ent)
		}
	}
	return result
}

func (b *graphvizImageBuilder) printGroups(stringBounder font.StringBounder, parent *abel.Entity) {
	for _, g := range b.dotData.groupHierarchy.ChildrenGroups(parent) {
		if g.IsRemoved() {
			continue
		}
		if b.dotData.isEmpty(g) && g.GroupType() == abel.GroupTypePackage {
			g.MuteToType(abel.LeafTypeEmptyPackage)
			b.printEntity(stringBounder, g)
		} else {
			b.printGroup(stringBounder, g)
		}
	}
}

func (b *graphvizImageBuilder) printGroup(stringBounder font.StringBounder, g *abel.Entity) {
	if g.GroupType() == abel.GroupTypeConcurrentState {
		return
	}
	header := newClusterHeader(g, b.dotData.portionShower, stringBounder)
	b.clusterManager.openCluster(g, header)
	b.printEntities(stringBounder, g.Leafs())
	b.printGroups(stringBounder, g)
	b.clusterManager.closeCluster()
}
