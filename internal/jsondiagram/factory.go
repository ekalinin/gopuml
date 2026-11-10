// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/JsonDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Factory parses @startjson blocks, as upstream's JsonDiagramFactory. It is
// safe for concurrent use.
var Factory core.Factory = jsonFactory{}

type jsonFactory struct{}

// DiagramType implements core.Factory.
func (jsonFactory) DiagramType() core.DiagramType { return core.JSON }

// CreateSystem implements core.Factory: the lines between the start and the
// end lines, without the lines that StyleExtractor takes out, are the JSON
// text; the highlight paths are read and the other lines that start with
// "#" are dropped. JSON that cannot be parsed gives a diagram without data.
func (jsonFactory) CreateSystem(ctx *core.RenderContext, src *core.UmlSource) core.Diagram {
	var highlighted []*Highlighted
	styleExtractor := NewStyleExtractor(src.Lines())
	var sb strings.Builder
	for _, line := range dataLines(styleExtractor.Lines()) {
		if strings.HasPrefix(line, "#") {
			if MatchesHighlightedDefinition(line) {
				highlighted = append(highlighted, BuildHighlighted(line))
			}
			continue
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	root, err := json.Parse(sb.String())
	if err != nil {
		root = nil
	}
	result := NewJsonDiagram(ctx, src, core.JSON, root, highlighted, styleExtractor)
	applyStylesAndTitle(result, styleExtractor)
	return result
}

// dataLines returns the lines between the start line and the end line.
func dataLines(lines []string) []string {
	if len(lines) < 2 {
		return nil
	}
	return lines[1 : len(lines)-1]
}

// applyStylesAndTitle applies the skin, the style blocks and the title of
// the source, as upstream's JSON and YAML factories do.
func applyStylesAndTitle(d *JsonDiagram, styleExtractor *StyleExtractor) {
	styleExtractor.applyStyles(&d.Base)
	if title := styleExtractor.Title(); title != "" {
		d.SetTitle(uml.DisplayPositionedSingle(nil, klimt.DisplayWithNewlines(d.Pragma(), title),
			geom.HorizontalAlignmentCenter, geom.VerticalAlignmentCenter))
	}
}
