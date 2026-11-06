// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/directdot/PSystemDotFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package directdot

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
)

// Factory parses @startdot blocks, as upstream's PSystemDotFactory.
var Factory = &command.BasicFactory{
	Type:        core.Dot,
	InitDiagram: func(*core.RenderContext, *core.UmlSource, string) core.Diagram { return nil },
	ExecuteLine: executeLine,
}

// executeLine starts the diagram at a Graphviz graph header and adds every
// following line to its DOT source; a first line that is not a header is a
// syntax error. Upstream keeps the source in a field of the factory and
// creates a diagram for each line; here the diagram holds it, as concurrent
// renders share the factory.
func executeLine(ctx *core.RenderContext, src *core.UmlSource, system core.Diagram, line string) core.Diagram {
	if system == nil {
		if !core.IsGraphvizDotHeader(line) {
			return nil
		}
		system = newPSystemDot(ctx, src)
	}
	d := system.(*PSystemDot)
	d.data.WriteString(line)
	d.data.WriteByte('\n')
	return d
}
