// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/svg/UGraphicSvg.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svg

import (
	"io"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// UGraphicSvg is the SVG output graphic: it dispatches shapes to the SVG
// drivers, which write them into its SvgGraphics.
type UGraphicSvg struct {
	klimt.AbstractUGraphic[*SvgGraphics]
	textAsPath bool
	fileFormat font.FileFormat
	option     *SvgOption
}

var (
	_ klimt.CommonUGraphic = (*UGraphicSvg)(nil)
	_ klimt.ClipContainer  = (*UGraphicSvg)(nil)
)

func newUGraphicSvg(stringBounder font.StringBounder, textAsPath bool, fileFormat font.FileFormat) *UGraphicSvg {
	g := &UGraphicSvg{AbstractUGraphic: klimt.NewAbstractUGraphic[*SvgGraphics](stringBounder),
		textAsPath: textAsPath, fileFormat: fileFormat}
	g.SetSelf(g)
	g.register()
	return g
}

// Build returns the SVG graphic of an option. textAsPath draws texts as
// glyph outlines (not ported yet: such texts panic); seed makes the ids of
// filters and gradients; fileFormat is SVG or SVG_DETERMINISTIC. The
// option's background must not be nil, as upstream.
func Build(option *SvgOption, textAsPath bool, seed int64, stringBounder font.StringBounder,
	fileFormat font.FileFormat) *UGraphicSvg {
	g := newUGraphicSvg(stringBounder, textAsPath, fileFormat)
	g.Copy(option.Backcolor(), option.ColorMapper(), NewSvgGraphics(seed, option))
	g.option = option
	return g
}

// CopyUGraphic returns a copy sharing the document and the bounds.
func (g *UGraphicSvg) CopyUGraphic() klimt.CommonUGraphic {
	result := newUGraphicSvg(g.StringBounder(), g.textAsPath, g.fileFormat)
	result.CopyFrom(&g.AbstractUGraphic)
	result.option = g.option
	return result
}

// ManageHiddenAutomatically returns false: hidden shapes reach the drivers,
// which count them in the canvas size without writing them.
func (g *UGraphicSvg) ManageHiddenAutomatically() bool { return false }

// BeforeDraw hides the next shape when the graphic is hidden.
func (g *UGraphicSvg) BeforeDraw() { g.GraphicObject().SetHidden(g.Param().IsHidden()) }

// AfterDraw shows the next shapes again.
func (g *UGraphicSvg) AfterDraw() { g.GraphicObject().SetHidden(false) }

func (g *UGraphicSvg) register() {
	a := &g.AbstractUGraphic
	klimt.RegisterDriver[*klimt.URectangle, *SvgGraphics](a, NewDriverRectangleSvg(g))
	if g.textAsPath {
		klimt.RegisterDriver[*klimt.UText, *SvgGraphics](a, NewDriverTextAsPathSvg(g))
	} else {
		klimt.RegisterDriver[*klimt.UText, *SvgGraphics](a, NewDriverTextSvg(g.StringBounder(), g))
	}
	klimt.RegisterDriver[*klimt.ULine, *SvgGraphics](a, NewDriverLineSvg(g))
	klimt.RegisterDriver[*klimt.UPixel, *SvgGraphics](a, klimt.UDriverFunc[*klimt.UPixel, *SvgGraphics](
		func(_ *klimt.UPixel, x, y float64, mapper *color.ColorMapper, param klimt.UParam, svg *SvgGraphics) {
			drawPixel(x, y, mapper, param, svg)
		}))
	klimt.RegisterDriver[klimt.UPixel, *SvgGraphics](a, DriverPixelSvg{})
	klimt.RegisterDriver[*klimt.UPolygon, *SvgGraphics](a, NewDriverPolygonSvg(g))
	klimt.RegisterDriver[*klimt.UEllipse, *SvgGraphics](a, NewDriverEllipseSvg(g))
	klimt.RegisterDriver[*klimt.UImage, *SvgGraphics](a, NewDriverImagePng(g))
	klimt.RegisterDriver[*klimt.UImageSvg, *SvgGraphics](a, DriverImageSvgSvg{})
	klimt.IgnoreShape[*klimt.UImageTikz, *SvgGraphics](a)
	klimt.RegisterDriver[*klimt.UPath, *SvgGraphics](a, NewDriverPathSvg(g))
	klimt.RegisterDriver[*klimt.DotPath, *SvgGraphics](a, DriverDotPathSvg{})
	klimt.RegisterDriver[*klimt.UCenteredCharacter, *SvgGraphics](a, NewDriverCenteredCharacterSvg(g.fileFormat))
}

// SvgGraphics returns the document being built.
func (g *UGraphicSvg) SvgGraphics() *SvgGraphics { return g.GraphicObject() }

// WriteToStream writes the document, with the <?plantuml-src ...?>
// instruction when metadata is not empty (upstream's non-null metadata;
// diagrams never have an empty one).
func (g *UGraphicSvg) WriteToStream(w io.Writer, metadata string, dpi int) error {
	if metadata != "" {
		g.GraphicObject().AddCommentMetadata(metadata)
	}
	return g.GraphicObject().CreateXML(w)
}

// StartGroup opens a <g> with the attributes of group.
func (g *UGraphicSvg) StartGroup(group *klimt.UGroup) { g.GraphicObject().StartGroup(group.Entries()) }

// CloseGroup closes the innermost group.
func (g *UGraphicSvg) CloseGroup() { g.GraphicObject().CloseGroup() }

// StartURL opens a link to url.
func (g *UGraphicSvg) StartURL(url *klimt.URL) {
	g.GraphicObject().OpenLink(url.URL(), url.Tooltip(), g.option.LinkTarget())
}

// CloseURL closes the innermost link.
func (g *UGraphicSvg) CloseURL() { g.GraphicObject().CloseLink() }

// DrawComment writes a comment.
func (g *UGraphicSvg) DrawComment(comment *klimt.UComment) {
	g.GraphicObject().AddComment(comment.Comment())
}

// MatchesProperty reports whether propertyName is "SVG", in any case.
func (g *UGraphicSvg) MatchesProperty(propertyName string) bool {
	if javaEqualsIgnoreCase(propertyName, "SVG") {
		return true
	}
	return g.AbstractUGraphic.MatchesProperty(propertyName)
}
