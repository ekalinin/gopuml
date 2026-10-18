// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/EmbeddedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package engine

import (
	"bytes"
	"context"
	"errors"
	"maps"
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// withEmbeddedDiagrams returns rc, or a copy of it that renders embedded
// diagrams when rc has no renderer yet.
func withEmbeddedDiagrams(rc *core.RenderContext) *core.RenderContext {
	if rc.EmbeddedDiagram != nil {
		return rc
	}
	out := rc.Clone()
	out.EmbeddedDiagram = func(lines []string, previous map[string]string, svg bool) ([]byte, error) {
		return renderEmbedded(out, lines, previous, svg)
	}
	return out
}

// renderEmbedded is the construction and export of upstream's
// EmbeddedDiagram: the lines are one block, read without the definitions
// and configuration of the command line (a BlockUml with
// Defines.createEmpty()), drawn with the skinparam values of the
// embedding diagram (Previous), as SVG or PNG. A layout error comes with
// the image that shows it, which is returned.
func renderEmbedded(rc *core.RenderContext, lines []string, previous map[string]string, svg bool) ([]byte,
	error) {
	nested := rc.Clone()
	nested.Defines = map[string]string{}
	nested.Config = nil
	nested.Pragmas = map[string]string{}
	nested.Previous = maps.Clone(previous)
	blocks, err := Parse(context.Background(), nested, Source{Text: strings.Join(lines, "\n")})
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, errors.New("embedded diagram without a block")
	}
	format := font.FormatPNG
	if svg {
		format = font.FormatSVG
	}
	var buf bytes.Buffer
	_, err = blocks[0].Render(context.Background(), &buf, 0, Options{Format: format})
	if buf.Len() > 0 {
		return buf.Bytes(), nil
	}
	if err == nil {
		err = errors.New("embedded diagram without an image")
	}
	return nil, err
}
