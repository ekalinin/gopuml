// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram

import (
	"os"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// parseSource parses a whole @startuml block with the factory.
func parseSource(t *testing.T, src string) *ActivityDiagram {
	t.Helper()
	var lines []text.StringLocated
	for s := range strings.SplitSeq(strings.TrimRight(src, "\n"), "\n") {
		lines = append(lines, text.StringLocated{S: s})
	}
	result := Factory.CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
	d, ok := result.(*ActivityDiagram)
	if !ok {
		if e, isError := result.(*core.ErrorDiagram); isError {
			t.Fatalf("error at line %d: %s", e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", result)
	}
	return d
}

// branches is a diagram with two branches, bars, labels and ends; upstream
// describes it as "(10 activities)".
const branches = `@startuml
(*) --> Receive
if "in stock?" then
  -->[yes] "Pack the order"
  --> ===Ship===
else
  -->[no] "Order from supplier"
  --> (*)
endif
if "" then
  -->[gift] ===Ship===
else
  -->[normal] Wrap
  --> ===Ship===
endif
===Ship=== --> "Send"
--> (*)
@enduml`

// TestFactoryParsesBranches checks that the factory parses a diagram with
// branches, bars and labels as upstream does.
func TestFactoryParsesBranches(t *testing.T) {
	if got := parseSource(t, branches).Description(); got != "(10 activities)" {
		t.Errorf("description %q", got)
	}
}

// TestFactoryParsesVegaCase checks the corpus case A0004, when the Vega
// suite is fetched in .cache (it is not committed).
func TestFactoryParsesVegaCase(t *testing.T) {
	data, err := os.ReadFile("../../.cache/vega/nonreg/simple/A0004.puml")
	if err != nil {
		t.Skip("the Vega suite is not fetched")
	}
	src := string(data)
	src = src[strings.Index(src, "@startuml"):]
	if got := parseSource(t, src).Description(); got != "(12 activities)" {
		t.Errorf("description %q", got)
	}
}
