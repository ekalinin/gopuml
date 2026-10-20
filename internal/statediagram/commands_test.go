// SPDX-License-Identifier: MIT
// Provenance: original

package statediagram

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// parseWith parses body (the lines between @startuml and @enduml) with a
// factory of the given commands only.
func parseWith(commands func() []cmd, body ...string) core.Diagram {
	lines := []text.StringLocated{{S: "@startuml"}}
	for _, s := range body {
		lines = append(lines, text.StringLocated{S: s})
	}
	lines = append(lines, text.StringLocated{S: "@enduml"})
	return uml.NewFactory(core.State, commands, NewStateDiagram).
		CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
}

// mustParse is parseWith for a body without error.
func mustParse(t *testing.T, commands func() []cmd, body ...string) *StateDiagram {
	t.Helper()
	result := parseWith(commands, body...)
	d, ok := result.(*StateDiagram)
	if !ok {
		if e, ok := result.(*core.ErrorDiagram); ok {
			t.Fatalf("error at line %d: %s", e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", result)
	}
	return d
}

// parseError is parseWith for a body with an error; it returns the message.
func parseError(t *testing.T, commands func() []cmd, body ...string) string {
	t.Helper()
	e, ok := parseWith(commands, body...).(*core.ErrorDiagram)
	if !ok {
		t.Fatalf("no error for %q", body)
	}
	return e.Error.Message()
}

// entity returns the entity of the first quark named code, at any level.
func entity(t *testing.T, d *StateDiagram, code string) *abel.Entity {
	t.Helper()
	for _, q := range d.Quarks() {
		if q.Name() == code && q.Data() != nil {
			return q.Data()
		}
	}
	t.Fatalf("no entity %q", code)
	return nil
}

func displayOf(e *abel.Entity) string {
	var lines []string
	for i := range e.Display().Size() {
		lines = append(lines, e.Display().Get(i).String())
	}
	return strings.Join(lines, "\n")
}

func createStateCommands() []cmd { return []cmd{commandCreateState(), commandAddField()} }

// TestCreateStateNames checks the four forms of the name: code, quoted
// code, code as "display" and "display" as code.
func TestCreateStateNames(t *testing.T) {
	d := mustParse(t, createStateCommands,
		`state Idle`,
		`state "Quoted"`,
		`state Wait as "Waiting for input"`,
		`state "Long name" as Long`)
	for code, display := range map[string]string{
		"Idle": "Idle", "Quoted": "Quoted", "Wait": "Waiting for input", "Long": "Long name",
	} {
		e := entity(t, d, code)
		if e.LeafType() != abel.LeafTypeState || displayOf(e) != display {
			t.Errorf("%s: %v %q", code, e.LeafType(), displayOf(e))
		}
	}
}

// TestCreateStateStereotypes checks the stereotypes that make pseudo
// states, and a stereotype that does not.
func TestCreateStateStereotypes(t *testing.T) {
	d := mustParse(t, createStateCommands,
		`state c <<choice>>`,
		`state f <<fork>>`,
		`state j <<join>>`,
		`state s <<start>>`,
		`state e <<end>>`,
		`state h <<history>>`,
		`state hh <<history*>>`,
		`state other <<custom>>`)
	for code, want := range map[string]abel.LeafType{
		"c": abel.LeafTypeStateChoice, "f": abel.LeafTypeStateForkJoin, "j": abel.LeafTypeStateForkJoin,
		"s": abel.LeafTypeCircleStart, "e": abel.LeafTypeCircleEnd, "h": abel.LeafTypePseudoState,
		"hh": abel.LeafTypeDeepHistory, "other": abel.LeafTypeState,
	} {
		if got := entity(t, d, code).LeafType(); got != want {
			t.Errorf("%s: %v, want %v", code, got, want)
		}
	}
	if st := entity(t, d, "other").Stereotype(); st == nil {
		t.Error("no stereotype")
	} else if label, _ := st.Label(klimt.GuillemetDoubleComparator()); label != "<<custom>>" {
		t.Errorf("stereotype %q", label)
	}
}

// TestCreateStateOptions checks tags, URLs, colors, the line color and
// style, and the description line.
func TestCreateStateOptions(t *testing.T) {
	d := mustParse(t, createStateCommands,
		`state Error $alert [[http://example.com]] #pink ##[dashed]red : failed`,
		`Error : retry`)
	e := entity(t, d, "Error")
	if tags := e.Stereotags(); len(tags) != 1 || tags[0].Name() != "alert" {
		t.Errorf("tags %v", tags)
	}
	if e.URL99() == nil || e.URL99().URL() != "http://example.com" {
		t.Errorf("url %v", e.URL99())
	}
	colors := e.Colors()
	if colors.Color(color.TypeBack) == nil || colors.Color(color.TypeLine) == nil ||
		colors.LineStyle() != color.LineStyleDashed {
		t.Errorf("colors %v", colors)
	}
	body := e.Bodier().RawBody()
	if len(body) != 2 || body[0].String() != "failed" || body[1].String() != "retry" {
		t.Errorf("body %v", body)
	}
}

// TestAddFieldCreatesState checks that "Code : text" creates the state.
func TestAddFieldCreatesState(t *testing.T) {
	d := mustParse(t, createStateCommands, `"Quoted" : first`, `Plain : second`)
	for _, code := range []string{"Quoted", "Plain"} {
		if e := entity(t, d, code); e.LeafType() != abel.LeafTypeState || len(e.Bodier().RawBody()) != 1 {
			t.Errorf("%s: %v", code, e.LeafType())
		}
	}
}

// TestCreateStateBorderStereotypeAtTopLevel checks upstream's error for an
// entry point outside of a composite state.
func TestCreateStateBorderStereotypeAtTopLevel(t *testing.T) {
	if msg := parseError(t, createStateCommands, `state entry1 <<entryPoint>>`); !strings.HasPrefix(msg,
		"You cannot use this stereotype here") {
		t.Errorf("error %q", msg)
	}
}

// TestCreateStateBadColor checks that an unknown line color is an error.
func TestCreateStateBadColor(t *testing.T) {
	if msg := parseError(t, createStateCommands, `state A ##nosuchcolor`); msg == "" {
		t.Error("no error message")
	}
}
