// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/AbstractDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/UgDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/TitledDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/WithSprite.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/DiagramDescription.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Diagram is a UML diagram as the common commands see it: a command
// diagram with the shared state of Base. Diagram types embed Base, which
// provides every method but the command.Diagram ones they override.
type Diagram interface {
	command.Diagram
	// Titled returns the shared state, upstream's TitledDiagram part.
	Titled() *Base
	// SkinParam returns the skin parameters, for the style commands.
	SkinParam() klimt.ISkinParam
}

// HideEmptyDescriptionSetter is implemented by the diagrams that honor
// "hide empty description" (upstream's TitledDiagram.setHideEmptyDescription
// does nothing).
type HideEmptyDescriptionSetter interface {
	SetHideEmptyDescription(hide bool)
}

// Base is the state that every UML diagram shares: upstream's
// AbstractDiagram, UgDiagram and TitledDiagram fields. It provides the
// default command.Diagram hooks through command.Base.
type Base struct {
	command.Base

	source      *core.UmlSource
	ctx         *core.RenderContext
	diagramType core.DiagramType
	md5map      map[string]string

	scale                *Scale
	splitPagesHorizontal int
	splitPagesVertical   int

	skinParam *klimt.SkinParam

	title, caption, legend, header, footer, mainFrame DisplayPositioned

	namespaceSeparator    string
	hasNamespaceSeparator bool
	skinParamUsed         bool
	useSmetana, useElk    bool
}

// NewBase returns the shared state of a diagram of type diagramType built
// from src. md5map holds the base64 images that PatchBase64 replaced by
// their MD5 digest; it may be nil.
func NewBase(ctx *core.RenderContext, src *core.UmlSource, diagramType core.DiagramType,
	md5map map[string]string) Base {
	if md5map == nil {
		md5map = map[string]string{}
	}
	// The !option values of the preprocessing artifact, which the engine
	// puts in the render context of the block.
	options := klimt.NewConfigurationStore[klimt.OptionKey]()
	for _, name := range slices.Sorted(maps.Keys(ctx.PreprocOptions)) {
		if key, ok := klimt.OptionKeyLazyFrom(name); ok {
			options.Define(key, ctx.PreprocOptions[name], true)
		}
	}
	return Base{
		source:               src,
		ctx:                  ctx,
		diagramType:          diagramType,
		md5map:               md5map,
		splitPagesHorizontal: 1,
		splitPagesVertical:   1,
		skinParam:            klimt.SkinParamCreate(md5map, ctx, diagramType, klimt.PragmaCreateEmpty(), options),
		title:                DisplayPositionedNone(geom.HorizontalAlignmentCenter, geom.VerticalAlignmentTop),
		caption:              DisplayPositionedNone(geom.HorizontalAlignmentCenter, geom.VerticalAlignmentBottom),
		legend:               DisplayPositionedNone(geom.HorizontalAlignmentCenter, geom.VerticalAlignmentBottom),
		header:               DisplayPositionedNone(geom.HorizontalAlignmentCenter, VerticalAlignmentNull),
		footer:               DisplayPositionedNone(geom.HorizontalAlignmentCenter, VerticalAlignmentNull),
		mainFrame:            DisplayPositionedNone(klimt.HorizontalAlignmentNull, VerticalAlignmentNull),
	}
}

// Titled implements Diagram.
func (b *Base) Titled() *Base { return b }

// Source returns the source of the diagram.
func (b *Base) Source() *core.UmlSource { return b.source }

// RenderContext returns the render context, upstream's PathSystem.
func (b *Base) RenderContext() *core.RenderContext { return b.ctx }

// DiagramType implements core.TypedDiagram.
func (b *Base) DiagramType() core.DiagramType { return b.diagramType }

// MD5Map returns the base64 images of the source by MD5 digest.
func (b *Base) MD5Map() map[string]string { return b.md5map }

// Pages is upstream's getNbImages: one image by default.
func (b *Base) Pages() int { return 1 }

// Scale returns the scale, nil when none is set.
func (b *Base) Scale() *Scale { return b.scale }

// SetScale sets the scale.
func (b *Base) SetScale(scale *Scale) { b.scale = scale }

// SplitPages returns the number of horizontal and vertical pages ("page"
// command).
func (b *Base) SplitPages() (horizontal, vertical int) {
	return b.splitPagesHorizontal, b.splitPagesVertical
}

// SetSplitPages sets the number of horizontal and vertical pages.
func (b *Base) SetSplitPages(horizontal, vertical int) {
	b.splitPagesHorizontal = horizontal
	b.splitPagesVertical = vertical
}

// SkinParam implements Diagram and klimt.StyleDiagram.
func (b *Base) SkinParam() klimt.ISkinParam { return b.skinParam }

// Skin returns the skin parameters.
func (b *Base) Skin() *klimt.SkinParam { return b.skinParam }

// CurrentStyleBuilder returns the style sheet of the diagram.
func (b *Base) CurrentStyleBuilder() *klimt.StyleBuilder { return b.skinParam.CurrentStyleBuilder() }

// Pragma returns the pragmas of the diagram.
func (b *Base) Pragma() *klimt.Pragma { return b.skinParam.Pragma() }

// AddWarning adds a warning, shown below the diagram. A nil warning is
// ignored.
func (b *Base) AddWarning(warning *klimt.Warning) {
	if warning != nil {
		b.Pragma().AddWarning(warning)
	}
}

// Warnings returns the warnings of the diagram: those of the preprocessor,
// then those of the pragmas, as TitledDiagram.getWarnings.
func (b *Base) Warnings() []*klimt.Warning {
	var out []*klimt.Warning
	for _, w := range b.ctx.PreprocWarnings {
		out = append(out, klimt.NewWarning(strings.Split(w, "\n")...))
	}
	return append(out, b.Pragma().Warnings()...)
}

// SetParam sets a skin parameter; the key is lower-cased.
func (b *Base) SetParam(key, value string) { b.skinParam.SetParam(LowerCase(key), value) }

// AddSprite adds a sprite.
func (b *Base) AddSprite(name string, sprite klimt.Sprite) { b.skinParam.AddSprite(name, sprite) }

// IsSkinParamUsed reports whether a skinparam command was used.
func (b *Base) IsSkinParamUsed() bool { return b.skinParamUsed }

// SetSkinParamUsed records that a skinparam command was used.
func (b *Base) SetSkinParamUsed(used bool) { b.skinParamUsed = used }

// NamespaceSeparator returns the namespace separator; ok is false when it is
// null (the default, or "none").
func (b *Base) NamespaceSeparator() (separator string, ok bool) {
	return b.namespaceSeparator, b.hasNamespaceSeparator
}

// SetNamespaceSeparator sets the namespace separator; ok false sets null.
func (b *Base) SetNamespaceSeparator(separator string, ok bool) {
	b.namespaceSeparator, b.hasNamespaceSeparator = separator, ok
}

// IsUseSmetana reports whether "!pragma layout smetana" was used.
func (b *Base) IsUseSmetana() bool { return b.useSmetana }

// SetUseSmetana records "!pragma layout smetana".
func (b *Base) SetUseSmetana(use bool) { b.useSmetana = use }

// IsUseElk reports whether "!pragma layout elk" was used.
func (b *Base) IsUseElk() bool { return b.useElk }

// SetUseElk records "!pragma layout elk".
func (b *Base) SetUseElk(use bool) { b.useElk = use }

// IsHandwritten reports whether the diagram is drawn handwritten. Upstream
// also reads "!option handwritten", which core.Factory does not carry yet.
func (b *Base) IsHandwritten() bool { return b.skinParam.Handwritten() }

// LoadSkin is the "skin" command: it replaces the style sheet by the skin
// file name + ".skin".
func (b *Base) LoadSkin(name string) command.Result {
	filename := name + ".skin"
	styleBuilder, err := klimt.StyleLoaderLoadSkin(b.ctx, filename)
	if err != nil {
		var noStyle *klimt.NoStyleAvailableError
		var parsing *klimt.StyleParsingError
		switch {
		case errors.As(err, &noStyle):
			return command.Error("Cannot find style " + name)
		case errors.As(err, &parsing):
			return command.Error("Cannot parse style " + name + ": " + parsing.Message)
		}
		return command.Error("Skin read error")
	}
	if missing := klimt.StyleLoaderMissingRootProperties(styleBuilder); len(missing) > 0 {
		names := make([]string, len(missing))
		for i, p := range missing {
			names[i] = p.String()
		}
		return command.Error("Incomplete style " + name + ": root does not define [" + strings.Join(names, ", ") + "]")
	}
	b.skinParam.SetDefaultSkin(filename)
	return command.Ok()
}

// SetTitle sets the title, unless the text is null or white.
func (b *Base) SetTitle(title DisplayPositioned) {
	if title.IsNull() || title.Display().IsWhite() {
		return
	}
	b.title = title
}

// Title returns the title.
func (b *Base) Title() DisplayPositioned { return b.title }

// TitleDisplay returns the text of the title.
func (b *Base) TitleDisplay() *klimt.Display { return b.title.Display() }

// SetMainFrame sets the main frame.
func (b *Base) SetMainFrame(mainFrame DisplayPositioned) { b.mainFrame = mainFrame }

// MainFrame returns the main frame.
func (b *Base) MainFrame() DisplayPositioned { return b.mainFrame }

// SetCaption sets the caption.
func (b *Base) SetCaption(caption DisplayPositioned) { b.caption = caption }

// Caption returns the caption.
func (b *Base) Caption() DisplayPositioned { return b.caption }

// SetLegend sets the legend.
func (b *Base) SetLegend(legend DisplayPositioned) { b.legend = legend }

// Legend returns the legend.
func (b *Base) Legend() DisplayPositioned { return b.legend }

// Header returns the header.
func (b *Base) Header() DisplayPositioned { return b.header }

// Footer returns the footer.
func (b *Base) Footer() DisplayPositioned { return b.footer }

// UpdateHeader sets the text, alignment and source line of the header.
func (b *Base) UpdateHeader(location *text.LineLocation, display *klimt.Display,
	horizontal geom.HorizontalAlignment) {
	b.header = b.header.WithDisplay(display).WithHorizontalAlignment(horizontal).WithLocation(location)
}

// UpdateFooter sets the text, alignment and source line of the footer.
func (b *Base) UpdateFooter(location *text.LineLocation, display *klimt.Display,
	horizontal geom.HorizontalAlignment) {
	b.footer = b.footer.WithDisplay(display).WithHorizontalAlignment(horizontal).WithLocation(location)
}

// DefaultMargins are the margins around the drawing, 10 for UML diagrams.
func (b *Base) DefaultMargins() klimt.ClockwiseTopRightBottomLeft { return klimt.ClockwiseSame(10) }
