// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/hcl/HclDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package hcl

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/jsondiagram"
)

// Factory parses @starthcl blocks, as upstream's HclDiagramFactory. It is
// safe for concurrent use.
var Factory core.Factory = hclFactory{}

type hclFactory struct{}

// DiagramType implements core.Factory.
func (hclFactory) DiagramType() core.DiagramType { return core.HCL }

// CreateSystem implements core.Factory: the lines between the start and the
// end lines, without the lines that StyleExtractor takes out, are the HCL
// text, drawn as a JSON diagram. The style blocks, the skin and the title
// are not applied, as upstream's code for them is commented out.
func (hclFactory) CreateSystem(ctx *core.RenderContext, src *core.UmlSource) core.Diagram {
	styleExtractor := jsondiagram.NewStyleExtractor(src.Lines())
	list := &hclSource{}
	if lines := styleExtractor.Lines(); len(lines) >= 2 {
		for _, line := range lines[1 : len(lines)-1] {
			list.add(line)
		}
	}
	return jsondiagram.NewJsonDiagram(ctx, src, core.HCL, hclData(list), nil, styleExtractor)
}

// hclData parses the source; HCL that upstream's parser rejects gives no
// data, as upstream's factory catches every exception.
func hclData(list *hclSource) json.Value {
	parser, err := newHclParser(list)
	if err != nil {
		return nil
	}
	data, err := parser.parseMe()
	if err != nil {
		return nil
	}
	return data
}
