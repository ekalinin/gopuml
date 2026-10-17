// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/UGraphicG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/drawing/g2d/DriverPixelG2d.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/SecurityUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package g2d

import (
	"io"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/png"
)

// UGraphicG2d is the raster output graphic: it dispatches shapes to the g2d
// drivers, which draw them on an image with Java2D's rendering rules.
type UGraphicG2d struct {
	klimt.AbstractUGraphic[*graphics]
	image        *Image
	dpiFactor    float64
	format       font.FileFormat
	antiAliasing klimt.UAntiAliasing
	links        *linkState
}

// linkState holds the links of a drawing, shared by the graphics derived
// from it: the open links (nil for an ignored link) and every link met.
type linkState struct {
	open            []*klimt.URL
	all             []*klimt.URL
	seen            map[*klimt.URL]bool
	allowJavascript bool
}

var _ klimt.CommonUGraphic = (*UGraphicG2d)(nil)

// New returns the graphic of an empty image: defaultBackground is the
// background of the drawing, dpiFactor scales everything drawn (the
// diagram scale times dpi / 96), format is PNG, PNG_EMPTY (which writes
// nothing) or RAW.
func New(defaultBackground color.HColor, colorMapper *color.ColorMapper, stringBounder font.StringBounder,
	builder *EmptyImageBuilder, dpiFactor float64, format font.FileFormat) *UGraphicG2d {
	g := &UGraphicG2d{AbstractUGraphic: klimt.NewAbstractUGraphic[*graphics](stringBounder),
		format: format, dpiFactor: dpiFactor, antiAliasing: klimt.AntiAliasingOn,
		links: &linkState{seen: map[*klimt.URL]bool{}}}
	g.SetSelf(g)
	gr := builder.g
	g.Copy(defaultBackground, colorMapper, gr)
	if dpiFactor != 1.0 {
		gr.tx = gr.tx.scale(dpiFactor, dpiFactor)
	}
	g.image = builder.im
	g.register()
	return g
}

// AllowJavascriptInLink keeps "javascript:" links in the image map, upstream's
// PLANTUML_ALLOW_JAVASCRIPT_IN_LINK (false by default).
func (g *UGraphicG2d) AllowJavascriptInLink(allow bool) { g.links.allowJavascript = allow }

// CopyUGraphic returns a copy sharing the image, the bounds and the links.
func (g *UGraphicG2d) CopyUGraphic() klimt.CommonUGraphic {
	result := &UGraphicG2d{AbstractUGraphic: klimt.NewAbstractUGraphic[*graphics](g.StringBounder()),
		image: g.image, dpiFactor: g.dpiFactor, format: g.format, antiAliasing: g.antiAliasing, links: g.links}
	result.SetSelf(result)
	result.CopyFrom(&g.AbstractUGraphic)
	result.register()
	return result
}

// Apply returns a copy with a change applied; UAntiAliasing switches the
// antialiasing of the copy.
func (g *UGraphicG2d) Apply(change klimt.UChange) klimt.UGraphic {
	cp := g.AbstractUGraphic.Apply(change).(*UGraphicG2d)
	if aa, ok := change.(klimt.UAntiAliasing); ok {
		cp.antiAliasing = aa
	}
	return cp
}

func (g *UGraphicG2d) register() {
	a := &g.AbstractUGraphic
	klimt.RegisterDriver[*klimt.URectangle, *graphics](a, driverRectangle{dpiFactor: g.dpiFactor, visible: g})
	if g.dpiFactor != 1.0 {
		klimt.RegisterDriver[*klimt.UText, *graphics](a, driverTextAsPath{visible: g, stringBounder: g.StringBounder()})
	} else {
		klimt.RegisterDriver[*klimt.UText, *graphics](a, driverText{visible: g, stringBounder: g.StringBounder()})
	}
	klimt.RegisterDriver[*klimt.ULine, *graphics](a, driverLine{dpiFactor: g.dpiFactor})
	pixel := klimt.UDriverFunc[*klimt.UPixel, *graphics](
		func(_ *klimt.UPixel, x, y float64, mapper *color.ColorMapper, param klimt.UParam, gr *graphics) {
			drawPixel(x, y, mapper, param, gr)
		})
	klimt.RegisterDriver[*klimt.UPixel, *graphics](a, pixel)
	klimt.RegisterDriver[klimt.UPixel, *graphics](a, klimt.UDriverFunc[klimt.UPixel, *graphics](
		func(_ klimt.UPixel, x, y float64, mapper *color.ColorMapper, param klimt.UParam, gr *graphics) {
			drawPixel(x, y, mapper, param, gr)
		}))
	klimt.RegisterDriver[*klimt.UPolygon, *graphics](a, driverPolygon{dpiFactor: g.dpiFactor, visible: g})
	klimt.RegisterDriver[*klimt.UEllipse, *graphics](a, driverEllipse{dpiFactor: g.dpiFactor, visible: g})
	klimt.IgnoreShape[*klimt.UImageSvg, *graphics](a)
	klimt.IgnoreShape[*klimt.UImageTikz, *graphics](a)
	klimt.RegisterDriver[*klimt.UImage, *graphics](a, driverImage{dpiFactor: g.dpiFactor, visible: g})
	klimt.RegisterDriver[*klimt.DotPath, *graphics](a, driverDotPath{visible: g})
	klimt.RegisterDriver[*klimt.UPath, *graphics](a, driverPath{dpiFactor: g.dpiFactor})
	klimt.RegisterDriver[*klimt.UCenteredCharacter, *graphics](a, driverCenteredCharacter{})
}

// BeforeDraw applies the clip and the antialiasing of the graphic.
func (g *UGraphicG2d) BeforeDraw() {
	g.applyClip()
	g.GraphicObject().aa = g.antiAliasing == klimt.AntiAliasingOn
}

func (g *UGraphicG2d) applyClip() {
	uclip := g.Clip()
	if uclip == nil {
		g.GraphicObject().setClip(nil)
		return
	}
	g.GraphicObject().setClip(&rect2D{uclip.X(), uclip.Y(), uclip.Width(), uclip.Height()})
}

// DpiFactor returns the factor that scales the drawing.
func (g *UGraphicG2d) DpiFactor() float64 { return g.dpiFactor }

// StartURL opens a link. "javascript:" links are ignored (a security
// issue), unless allowed.
func (g *UGraphicG2d) StartURL(url *klimt.URL) {
	if url == nil {
		panic(&color.RuntimeError{Class: "java.lang.NullPointerException"})
	}
	if !g.links.allowJavascript && isJavascriptLink(url.URL()) {
		g.links.open = append(g.links.open, nil)
		return
	}
	g.links.open = append(g.links.open, url)
	if !g.links.seen[url] {
		g.links.seen[url] = true
		g.links.all = append(g.links.all, url)
	}
}

// CloseURL closes the innermost link.
func (g *UGraphicG2d) CloseURL() { g.links.open = g.links.open[:len(g.links.open)-1] }

// EnsureVisible adds a drawn point to the areas of the open links.
func (g *UGraphicG2d) EnsureVisible(x, y float64) {
	for _, u := range g.links.open {
		if u != nil && (g.Clip() == nil || g.Clip().IsInside(x, y)) {
			u.EnsureVisible(x, y)
		}
	}
}

// AllUrlsEncountered returns the links met, in the order they were opened
// (upstream keeps them in a HashSet).
func (g *UGraphicG2d) AllUrlsEncountered() []*klimt.URL { return g.links.all }

// Image returns the image drawn on.
func (g *UGraphicG2d) Image() *Image { return g.image }

// WriteToStream writes the image as PNG with the plantuml text chunk when
// metadata is not empty; PNG_EMPTY writes nothing.
func (g *UGraphicG2d) WriteToStream(w io.Writer, metadata string, dpi int) error {
	if g.format == font.FormatPNGEmpty {
		return nil
	}
	im := g.image
	return png.Write(w, png.RasterImage{Width: im.width, Height: im.height, Pix: im.pix, HasAlpha: im.hasAlpha},
		metadata, dpi)
}

// isJavascriptLink is upstream's SecurityUtils.isJavascriptLink: the
// letters of the link, lower-cased, start with "javascript".
func isJavascriptLink(url string) bool {
	var letters strings.Builder
	for _, r := range strings.ToLower(url) {
		if r >= 'a' && r <= 'z' {
			letters.WriteRune(r)
		}
	}
	return strings.HasPrefix(letters.String(), "javascript")
}
