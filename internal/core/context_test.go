// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"fmt"
	"sync"
	"testing"
)

// configFactory accepts every block and records the second line, where the
// first config line of the render context lands.
type configFactory struct{}

func (configFactory) DiagramType() DiagramType { return Sequence }

func (configFactory) CreateSystem(ctx *RenderContext, src *UmlSource) Diagram {
	return okDiagram{by: src.Lines()[1].S + "|" + ctx.Pragmas["layout"]}
}

func TestParseWithConcurrentContexts(t *testing.T) {
	const renders = 50
	var wg sync.WaitGroup
	results := make([]string, renders)
	for i := range renders {
		wg.Go(func() {
			ctx := NewRenderContext()
			ctx.Config = []string{fmt.Sprintf("skinparam id %d", i)}
			ctx.Pragmas["layout"] = fmt.Sprint(i % 3)
			d := Parse(ctx, "@startuml\nA -> B\n@enduml\n", []Factory{configFactory{}})
			results[i] = d[0].(okDiagram).by
		})
	}
	wg.Wait()
	for i, got := range results {
		if want := fmt.Sprintf("skinparam id %d|%d", i, i%3); got != want {
			t.Errorf("render %d saw %q, want %q", i, got, want)
		}
	}
}

func TestRenderContextClone(t *testing.T) {
	a := NewRenderContext()
	a.Config = []string{"x"}
	a.Defines["A"] = "1"
	b := a.Clone()
	b.Config[0] = "y"
	b.Defines["A"] = "2"
	b.Pragmas["p"] = "q"
	if a.Config[0] != "x" || a.Defines["A"] != "1" || len(a.Pragmas) != 0 {
		t.Errorf("clone shares state with the original: %+v", a)
	}
	if a.SecurityProfile != "LEGACY" || a.DotTimeout != DefaultDotTimeout {
		t.Errorf("defaults: %+v", a)
	}
}
