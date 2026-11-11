// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/YamlDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/yamlparser"
)

// YamlFactory parses @startyaml blocks, as upstream's YamlDiagramFactory.
// It is safe for concurrent use.
var YamlFactory core.Factory = yamlFactory{}

type yamlFactory struct{}

// DiagramType implements core.Factory.
func (yamlFactory) DiagramType() core.DiagramType { return core.YAML }

// CreateSystem implements core.Factory: the lines between the start and the
// end lines, without the lines that StyleExtractor takes out and the
// highlight paths, are the YAML text, drawn as a JSON diagram.
func (yamlFactory) CreateSystem(ctx *core.RenderContext, src *core.UmlSource) core.Diagram {
	styleExtractor := NewStyleExtractor(src.Lines())
	highlighted, root := yamlData(dataLines(styleExtractor.Lines()))
	result := NewJsonDiagram(ctx, src, core.YAML, root, highlighted, styleExtractor)
	applyStylesAndTitle(result, styleExtractor)
	return result
}

// yamlData reads the highlight paths and the YAML of the lines. A highlight
// path that cannot be read, and YAML that upstream's parser rejects, give
// no data, as upstream's factory catches every exception.
func yamlData(lines []string) (highlighted []*Highlighted, root json.Value) {
	var list []string
	for _, line := range lines {
		if MatchesHighlightedDefinition(line) {
			h := tryBuildHighlighted(line)
			if h == nil {
				return highlighted, nil
			}
			highlighted = append(highlighted, h)
			continue
		}
		list = append(list, line)
	}
	m, err := yamlparser.Parse(list)
	if err != nil {
		return highlighted, nil
	}
	root, err = yamlparser.MonomorphToJSON(m)
	if err != nil {
		return highlighted, nil
	}
	return highlighted, root
}

// tryBuildHighlighted is BuildHighlighted, nil when it throws.
func tryBuildHighlighted(line string) (h *Highlighted) {
	defer func() {
		if v := recover(); v != nil {
			if _, ok := v.(*color.RuntimeError); !ok {
				panic(v)
			}
			h = nil
		}
	}()
	return BuildHighlighted(line)
}
