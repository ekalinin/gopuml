// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/directdot/PSystemDot.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/directdot/CounterOutputStream.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/DirectOsDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package directdot

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/svek"
)

// PSystemDot is a dot diagram: a Graphviz graph that the layout engine of
// the render draws itself, as upstream's PSystemDot. It is not a
// core.TypedDiagram, as upstream's DirectOsDiagram is not a TitledDiagram.
type PSystemDot struct {
	data    strings.Builder
	seed    int64
	profile security.Profile
	layout  svek.Layout
}

func newPSystemDot(ctx *core.RenderContext, src *core.UmlSource) *PSystemDot {
	return &PSystemDot{seed: seed(src), profile: profileOf(ctx), layout: svek.LayoutFromContext(ctx)}
}

// Description implements core.Diagram.
func (*PSystemDot) Description() string { return "(Dot)" }

// Data returns the DOT source: the graph header and the lines that follow
// it, each ended by a newline.
func (d *PSystemDot) Data() string { return d.data.String() }

// profileOf is the security profile of a render, LEGACY by default.
func profileOf(ctx *core.RenderContext) security.Profile {
	if ctx == nil {
		return security.Legacy
	}
	if ctx.Security != nil {
		return ctx.Security.Profile
	}
	p, _ := security.ParseProfile(ctx.SecurityProfile)
	return p
}

// seed is UmlSource.seed: a hash of the lines of the source.
func seed(src *core.UmlSource) int64 {
	h := int64(1125899906842597)
	for _, l := range src.Lines() {
		h = 31*h + int64(klimt.JavaStringHashCode(l.S))
		h = 31*h + '\n'
	}
	return h
}
