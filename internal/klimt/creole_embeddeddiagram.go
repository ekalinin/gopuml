// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/EmbeddedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"errors"
	"maps"
	"regexp"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// plantumlProcessingInstruction is the "<?plantuml ...?>" instruction that
// upstream removes from the SVG of an embedded diagram.
var plantumlProcessingInstruction = regexp.MustCompile(`<\?plantuml.+?\?>`)

// embeddedDiagram is upstream's EmbeddedDiagram atom: an embedded diagram
// ("{{" ... "}}"), rendered by the renderer of the render context once per
// output kind, and drawn as SVG in SVG output and as an image otherwise.
// When it cannot be rendered, it measures 42 x 42 and draws nothing, as
// upstream does after logging the exception.
type embeddedDiagram struct {
	AtomBase
	render   core.EmbeddedDiagramRenderer
	lines    []string
	previous map[string]string

	svg   *string
	image PortableImage
}

var _ Atom = (*embeddedDiagram)(nil)

func (e *embeddedDiagram) imageSvg() (string, error) {
	if e.svg == nil {
		data, err := e.render(e.lines, e.previous, true)
		if err != nil {
			return "", err
		}
		s := plantumlProcessingInstruction.ReplaceAllString(string(data), "")
		e.svg = &s
	}
	return *e.svg, nil
}

func (e *embeddedDiagram) portableImage() (PortableImage, error) {
	if e.image == nil {
		data, err := e.render(e.lines, e.previous, false)
		if err != nil {
			return nil, err
		}
		image, err := readRasterImage(data)
		if err != nil {
			return nil, err
		}
		if image == nil {
			return nil, errors.New("embedded diagram: unknown image format")
		}
		e.image = image
	}
	return e.image, nil
}

// CalculateDimension returns the size of the rendered diagram.
func (e *embeddedDiagram) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return e.Dimension(stringBounder, func(stringBounder font.StringBounder) geom.XDimension2D {
		if stringBounder.MatchesProperty("SVG") {
			if svg, err := e.imageSvg(); err == nil {
				image := NewUImageSvg(svg, 1)
				return geom.NewXDimension2D(image.Width(), image.Height())
			}
		} else if image, err := e.portableImage(); err == nil {
			return geom.NewXDimension2D(float64(image.Width()), float64(image.Height()))
		}
		return geom.NewXDimension2D(42, 42)
	})
}

// StartingAltitude returns 0.
func (e *embeddedDiagram) StartingAltitude(font.StringBounder) float64 { return 0 }

// Neutrons returns the atom as one neutron.
func (e *embeddedDiagram) Neutrons() []*Neutron { return atomNeutrons(e) }

// DrawU draws the rendered diagram.
func (e *embeddedDiagram) DrawU(ug UGraphic) {
	if ug.MatchesProperty("SVG") {
		if svg, err := e.imageSvg(); err == nil {
			ug.Draw(NewUImageSvg(svg, 1))
		}
		return
	}
	if image, err := e.portableImage(); err == nil {
		ug.Draw(NewUImage(NewPixelImage(image, AffineTransformBilinear)))
	}
}

// makeEmbeddedDiagram is upstream's EmbeddedDiagram constructor: the
// diagram is rendered with the renderer of ctx, starting with the
// skinparam values of the embedding diagram (upstream's
// Previous.createFrom(skinParam.values())).
func makeEmbeddedDiagram(ctx *core.RenderContext, values map[string]string, lines []string) Atom {
	if ctx == nil || ctx.EmbeddedDiagram == nil {
		notPorted("EmbeddedDiagram", "embedded diagram without a renderer in the render context")
	}
	return &embeddedDiagram{render: ctx.EmbeddedDiagram, lines: lines, previous: maps.Clone(values)}
}

// MakeEmbeddedDiagram implements EmbeddedDiagramMaker.
func (s *SkinParam) MakeEmbeddedDiagram(lines []string) Atom {
	return makeEmbeddedDiagram(s.pathSystem, s.Values(), lines)
}

// MakeEmbeddedDiagram implements EmbeddedDiagramMaker for the skin
// parameters it wraps.
func (d SkinParamDelegator) MakeEmbeddedDiagram(lines []string) Atom {
	maker, ok := d.ISkinParam.(EmbeddedDiagramMaker)
	if !ok {
		notPorted("EmbeddedDiagram", "embedded diagram {{")
	}
	return maker.MakeEmbeddedDiagram(lines)
}
