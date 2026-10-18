// SPDX-License-Identifier: MIT
// Provenance: original

package abel

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// prefixHooks is a cleanId override that also removes a leading "x_".
type prefixHooks struct{ noHooks }

func (prefixHooks) CleanID(id string) string { return strings.TrimPrefix(DefaultCleanID(id), "x_") }

type hookedDiagram struct {
	AbstractEntityDiagram
}

func newHookedDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *hookedDiagram {
	d := &hookedDiagram{}
	d.InitCucaDiagram(ctx, src, core.Class, md5map, prefixHooks{})
	return d
}

// TestCleanIDHook checks that the shared commands clean ids through the
// CleanID hook of the diagram, as upstream's commands call the overridable
// cleanId.
func TestCleanIDHook(t *testing.T) {
	factory := uml.NewFactory(core.Class, func() []command.Command[*hookedDiagram] {
		return []command.Command[*hookedDiagram]{
			CommandPackage[*hookedDiagram](),
			CommandEndPackage[*hookedDiagram](),
			CommandNoteSingle[*hookedDiagram](),
		}
	}, newHookedDiagram)
	var lines []text.StringLocated
	for _, s := range []string{"@startuml", "package x_P {", "}", `note "hello" as x_N`, "@enduml"} {
		lines = append(lines, text.StringLocated{S: s})
	}
	result := factory.CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
	d, ok := result.(*hookedDiagram)
	if !ok {
		t.Fatalf("result %T", result)
	}
	for name, typ := range map[string]LeafType{"P": LeafTypeStillUnknown, "N": LeafTypeNote} {
		q := d.QuarkInContext(false, name)
		if q.Data() == nil {
			t.Errorf("%s was not created under its cleaned id", name)
			continue
		}
		if typ == LeafTypeNote && q.Data().LeafType() != LeafTypeNote {
			t.Errorf("%s: leaf type %v, want a note", name, q.Data().LeafType())
		}
		if typ != LeafTypeNote && !q.Data().IsGroup() {
			t.Errorf("%s: not a group", name)
		}
	}
	if q := d.QuarkInContext(false, "x_P"); q.Data() != nil {
		t.Error("x_P was created with the raw id")
	}
}
