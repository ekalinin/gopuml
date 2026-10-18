// SPDX-License-Identifier: MIT
// Provenance: original

package descdiagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// parseWith parses body (the lines between @startuml and @enduml) with a
// factory of the given commands only.
func parseWith(t *testing.T, commands func() []cmd, body ...string) core.Diagram {
	t.Helper()
	lines := []text.StringLocated{{S: "@startuml"}}
	for _, s := range body {
		lines = append(lines, text.StringLocated{S: s})
	}
	lines = append(lines, text.StringLocated{S: "@enduml"})
	return uml.NewFactory(core.Description, commands, NewDescriptionDiagram).
		CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
}

func elementCommands() []cmd { return []cmd{commandCreateElementFull()} }

func parseElements(t *testing.T, body ...string) *DescriptionDiagram {
	t.Helper()
	result := parseWith(t, elementCommands, body...)
	d, ok := result.(*DescriptionDiagram)
	if !ok {
		if e, ok := result.(*core.ErrorDiagram); ok {
			t.Fatalf("error at line %d: %s", e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", result)
	}
	return d
}

func entity(t *testing.T, d *DescriptionDiagram, code string) *abel.Entity {
	t.Helper()
	e := d.QuarkInContext(false, code).Data()
	if e == nil {
		t.Fatalf("no entity %q", code)
	}
	return e
}

func TestCreateElementKeywords(t *testing.T) {
	for _, kw := range strings.Split(abel.ElementAllTypes, "|") {
		if strings.HasPrefix(kw, "port") {
			continue // TestCreateElementPorts
		}
		d := parseElements(t, kw+" E")
		e := entity(t, d, "E")
		switch kw {
		case "usecase":
			if e.LeafType() != abel.LeafTypeUsecase {
				t.Errorf("%s: leaf type %v", kw, e.LeafType())
			}
		case "usecase/":
			if e.LeafType() != abel.LeafTypeUsecaseBusiness {
				t.Errorf("%s: leaf type %v", kw, e.LeafType())
			}
		case "circle":
			if e.LeafType() != abel.LeafTypeCircle || e.USymbol() != abel.USymbolInterface {
				t.Errorf("%s: leaf type %v, symbol %v", kw, e.LeafType(), e.USymbol())
			}
		default:
			if e.LeafType() != abel.LeafTypeDescription || e.USymbol() == abel.USymbolNull {
				t.Errorf("%s: leaf type %v, symbol %v", kw, e.LeafType(), e.USymbol())
			}
		}
	}
	for kw, want := range map[string]abel.USymbol{"actor": abel.USymbolActorStickman,
		"actor/": abel.USymbolActorStickmanBusiness, "node": abel.USymbolNode, "database": abel.USymbolDatabase,
		"component": abel.USymbolComponent2, "interface": abel.USymbolInterface, "cloud": abel.USymbolCloud} {
		if got := entity(t, parseElements(t, kw+" E"), "E").USymbol(); got != want {
			t.Errorf("%s: symbol %v, want %v", kw, got, want)
		}
	}
}

func TestCreateElementShortForms(t *testing.T) {
	for _, c := range []struct {
		line, code, display string
		typ                 abel.LeafType
		symbol              abel.USymbol
	}{
		{":Admin:", "Admin", "Admin", abel.LeafTypeDescription, abel.USymbolActorStickman},
		{":Boss:/", "Boss", "Boss", abel.LeafTypeDescription, abel.USymbolActorStickmanBusiness},
		{"(Reset password)", "Reset password", "Reset password", abel.LeafTypeUsecase, abel.USymbolUsecase},
		{"(Pay)/", "Pay", "Pay", abel.LeafTypeUsecaseBusiness, abel.USymbolUsecaseBusiness},
		{"[Web server]", "Web server", "Web server", abel.LeafTypeDescription, abel.USymbolComponent2},
		{"() REST", "REST", "REST", abel.LeafTypeDescription, abel.USymbolInterface},
		{`actor "Head of sales" as Boss`, "Boss", "Head of sales", abel.LeafTypeDescription,
			abel.USymbolActorStickman},
		{"usecase U as (Check out)", "U", "Check out", abel.LeafTypeUsecase, abel.USymbolUsecase},
		{"(Check out) as U", "U", "Check out", abel.LeafTypeUsecase, abel.USymbolUsecase},
		{`"Display" as N`, "N", "Display", abel.LeafTypeDescription, abel.USymbolActorStickman},
	} {
		d := parseElements(t, c.line)
		e := entity(t, d, c.code)
		if e.LeafType() != c.typ || e.USymbol() != c.symbol {
			t.Errorf("%s: leaf type %v, symbol %v", c.line, e.LeafType(), e.USymbol())
		}
		if got := e.Display().Get(0).String(); got != c.display {
			t.Errorf("%s: display %q, want %q", c.line, got, c.display)
		}
	}
}

func TestCreateElementStereotypeAndColor(t *testing.T) {
	d := parseElements(t, "node Server <<linux>> #red")
	e := entity(t, d, "Server")
	if st := e.Stereotype(); st == nil || !strings.Contains(st.String(), "linux") {
		t.Errorf("stereotype %v", st)
	}
	if e.Colors().Color(color.TypeBack) == nil {
		t.Error("no back color")
	}
}

func TestCreateElementPorts(t *testing.T) {
	result := parseWith(t, elementCommands, "portin p1")
	e, ok := result.(*core.ErrorDiagram)
	if !ok || !strings.Contains(e.Error.Message(), "Port can only be used inside an element") {
		t.Errorf("port at root level: %T %v", result, result)
	}
}

func TestCreateElementForbidden(t *testing.T) {
	if _, ok := parseWith(t, elementCommands, "Server").(*core.ErrorDiagram); !ok {
		t.Error("a line made of a single name was accepted")
	}
}
