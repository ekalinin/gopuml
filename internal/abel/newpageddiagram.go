// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/NewpagedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"errors"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// ErrNotDrawable is returned by Export: the graph layout and drawing of the
// diagrams built on CucaDiagram are not ported yet.
var ErrNotDrawable = errors.New("drawing of class diagrams is not implemented yet")

// NewpagedDiagram is a diagram split by "newpage" into several diagrams of
// the same type, each starting empty. The commands run on the last page.
type NewpagedDiagram struct {
	command.Base
	source   *core.UmlSource
	diagrams []command.Diagram
}

var _ command.Delegator = (*NewpagedDiagram)(nil)

// NewNewpagedDiagram returns the diagram of two pages. A page that is
// itself a NewpagedDiagram panics, as upstream throws
// IllegalArgumentException.
func NewNewpagedDiagram(source *core.UmlSource, diag1, diag2 command.Diagram) *NewpagedDiagram {
	if _, ok := diag1.(*NewpagedDiagram); ok {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if _, ok := diag2.(*NewpagedDiagram); ok {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &NewpagedDiagram{source: source, diagrams: []command.Diagram{diag1, diag2}}
}

// LastDiagram returns the last page.
func (n *NewpagedDiagram) LastDiagram() command.Diagram { return n.diagrams[len(n.diagrams)-1] }

// Delegate implements command.Delegator: the command runs on the last page;
// a "newpage" there adds a page.
func (n *NewpagedDiagram) Delegate(exec func(target command.Diagram) command.Result) command.Result {
	nb := len(n.diagrams)
	tmp := exec(n.diagrams[nb-1])
	if new1, ok := tmp.NewDiagram.(*NewpagedDiagram); ok {
		if len(new1.diagrams) != 2 {
			uml.Throw(uml.IllegalStateException, "")
		}
		if new1.diagrams[0] != n.diagrams[nb-1] {
			uml.Throw(uml.IllegalStateException, "")
		}
		n.diagrams = append(n.diagrams, new1.diagrams[1])
		return command.NewDiagramResult(n)
	}
	return tmp
}

// Description implements core.Diagram: the descriptions of the pages,
// separated by spaces.
func (n *NewpagedDiagram) Description() string {
	parts := make([]string, len(n.diagrams))
	for i, d := range n.diagrams {
		parts[i] = d.Description()
	}
	return strings.Join(parts, " ")
}

// Pages is upstream's getNbImages, which NewpagedDiagram does not
// override: one image.
func (n *NewpagedDiagram) Pages() int { return 1 }

// MakeDiagramReady implements command.Diagram.
func (n *NewpagedDiagram) MakeDiagramReady() {
	for _, d := range n.diagrams {
		d.MakeDiagramReady()
	}
}

// CheckFinalError implements command.Diagram: the first error of a page.
func (n *NewpagedDiagram) CheckFinalError() string {
	for _, p := range n.diagrams {
		if check := p.CheckFinalError(); check != "" {
			return check
		}
	}
	return ""
}

// Diagrams returns the pages.
func (n *NewpagedDiagram) Diagrams() []command.Diagram { return n.diagrams }

// Source returns the source of the diagram.
func (n *NewpagedDiagram) Source() *core.UmlSource { return n.source }

// Export draws a page; the drawing is not ported yet.
func (n *NewpagedDiagram) Export(int) error { return ErrNotDrawable }
