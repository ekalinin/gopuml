// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// parseModel parses a sequence diagram body and fails on errors.
func parseModel(t *testing.T, body string) *SequenceDiagram {
	t.Helper()
	d, _, err := parseCase("@startuml\n" + body + "\n@enduml\n")
	if err != nil {
		t.Fatal(err)
	}
	switch d := d.(type) {
	case *SequenceDiagram:
		return d
	case *core.ErrorDiagram:
		t.Fatalf("error %q at line %d", d.Error.Message(), d.Error.Position())
	}
	t.Fatalf("result %T", d)
	return nil
}

func codes(ps []*Participant) string {
	var s []string
	for _, p := range ps {
		s = append(s, p.Code()+":"+p.Type().String())
	}
	return strings.Join(s, " ")
}

func messageAt(t *testing.T, d *SequenceDiagram, i int) *Message {
	t.Helper()
	m, ok := d.Events()[i].(*Message)
	if !ok {
		t.Fatalf("event %d is %T", i, d.Events()[i])
	}
	return m
}

// The tests below follow the scenarios of the sequence-diagram spec.

func TestScenarioDeclaredParticipantKinds(t *testing.T) {
	d := parseModel(t, "actor User\ndatabase DB\nUser -> DB : query")
	if got := codes(d.Participants()); got != "User:ACTOR DB:DATABASE" {
		t.Errorf("participants %s", got)
	}
	m := messageAt(t, d, 0)
	if m.Participant1().Code() != "User" || m.Participant2().Code() != "DB" || m.Label().String() != "[query]" {
		t.Errorf("message %s -> %s %s", m.Participant1(), m.Participant2(), m.Label())
	}
	if d.Description() != "(2 participants)" || d.DiagramType() != core.Sequence {
		t.Errorf("description %s type %v", d.Description(), d.DiagramType())
	}
}

func TestScenarioImplicitParticipant(t *testing.T) {
	d := parseModel(t, "Alice -> Bob : hi")
	if got := codes(d.Participants()); got != "Alice:PARTICIPANT Bob:PARTICIPANT" {
		t.Errorf("participants %s", got)
	}
}

func TestScenarioDottedReply(t *testing.T) {
	d := parseModel(t, "Alice -> Bob : request\nBob --> Alice : response")
	if messageAt(t, d, 0).ArrowConfiguration().IsDotted() || !messageAt(t, d, 1).ArrowConfiguration().IsDotted() {
		t.Error("the request is solid and the response dotted")
	}
}

func TestScenarioSelfMessage(t *testing.T) {
	d := parseModel(t, "Alice -> Alice : think")
	m := messageAt(t, d, 0)
	if !m.IsSelfMessage() || m.Label().String() != "[think]" {
		t.Errorf("self message %v %s", m.IsSelfMessage(), m.Label())
	}
}

func TestScenarioAutonumberCustomStartAndFormat(t *testing.T) {
	d := parseModel(t, "autonumber 10 5 \"<b>[000]\"\nA -> B : one\nB -> A : two")
	for i, want := range []string{"<b>[010]", "<b>[015]"} {
		if got, ok := messageAt(t, d, i).MessageNumber(); !ok || got != want {
			t.Errorf("message %d number %q, want %q", i, got, want)
		}
	}
	if got := messageAt(t, d, 0).LabelNumbered().String(); got != "[<b>[010], one]" {
		t.Errorf("numbered label %s", got)
	}
}

func TestScenarioActivationWithReturn(t *testing.T) {
	d := parseModel(t, "Alice -> Bob ++ : call\nreturn done")
	ev := d.Events()
	if len(ev) != 4 {
		t.Fatalf("%d events", len(ev))
	}
	act, ok := ev[1].(*LifeEvent)
	if !ok || act.Type() != LifeEventActivate || act.Participant().Code() != "Bob" || act.Message() != ev[0] {
		t.Errorf("activation %v", ev[1])
	}
	ret := messageAt(t, d, 2)
	if !ret.ArrowConfiguration().IsDotted() || ret.Participant1().Code() != "Bob" ||
		ret.Participant2().Code() != "Alice" || ret.Label().String() != "[done]" {
		t.Errorf("return %s -> %s %s", ret.Participant1(), ret.Participant2(), ret.Label())
	}
	if deact, ok := ev[3].(*LifeEvent); !ok || deact.Type() != LifeEventDeactivate || deact.Message() != ev[2] {
		t.Errorf("deactivation %v", ev[3])
	}
	if d.ActivatingMessage() != nil {
		t.Error("the activation is closed")
	}
}

func TestScenarioAlternativeFragment(t *testing.T) {
	d := parseModel(t, "alt success\nA -> B\nelse failure\nB -> A\nend")
	start, ok := d.Events()[0].(*GroupingStart)
	if !ok || start.Title() != "alt" || start.Type() != GroupingStartType {
		t.Fatalf("start %v", d.Events()[0])
	}
	if c, _ := start.Comment(); c != "success" {
		t.Errorf("first section %q", c)
	}
	els, ok := d.Events()[2].(*GroupingLeaf)
	if c, _ := els.Comment(); !ok || els.Type() != GroupingElse || c != "failure" || els.GroupingStart() != start {
		t.Errorf("else %v", d.Events()[2])
	}
	if end, ok := d.Events()[4].(*GroupingLeaf); !ok || end.Type() != GroupingEnd || len(start.Children()) != 2 {
		t.Errorf("end %v", d.Events()[4])
	}
}

func TestScenarioNoteOverTwoParticipants(t *testing.T) {
	d := parseModel(t, "participant Alice\nparticipant Bob\nnote over Alice, Bob : shared")
	n, ok := d.Events()[0].(*Note)
	if !ok || n.Position() != NoteOverSeveral || n.Participant().Code() != "Alice" ||
		n.Participant2().Code() != "Bob" || n.Display().String() != "[shared]" {
		t.Errorf("note %v", d.Events()[0])
	}
}

func TestScenarioDivider(t *testing.T) {
	d := parseModel(t, "A -> B\n== Initialization ==\nB -> A")
	if div, ok := d.Events()[1].(*Divider); !ok || div.Text().String() != "[Initialization ]" {
		t.Errorf("divider %v", d.Events()[1])
	}
}

func TestScenarioDelaysSpacingReferences(t *testing.T) {
	d := parseModel(t, "A -> B\n...\n... later ...\n|||\n||45||\nref over A, B : text")
	ev := d.Events()
	if delay, ok := ev[1].(*Delay); !ok || delay.Text().Size() != 0 {
		t.Errorf("delay %v", ev[1])
	}
	if delay, ok := ev[2].(*Delay); !ok || delay.Text().String() != "[ later ]" {
		t.Errorf("labeled delay %v", ev[2])
	}
	if s, ok := ev[3].(*HSpace); !ok || s.Pixel() != 25 {
		t.Errorf("space %v", ev[3])
	}
	if s, ok := ev[4].(*HSpace); !ok || s.Pixel() != 45 {
		t.Errorf("sized space %v", ev[4])
	}
	if r, ok := ev[5].(*Reference); !ok || r.String() != "A-B" || r.Strings().String() != "[text]" {
		t.Errorf("reference %v", ev[5])
	}
}

func TestScenarioFootboxAndTwoPages(t *testing.T) {
	d := parseModel(t, "hide footbox\nAlice -> Bob : one\nnewpage Part 2\nAlice -> Bob : two")
	if d.IsShowFootbox() {
		t.Error("hide footbox")
	}
	if d.Pages() != 2 {
		t.Fatalf("%d pages", d.Pages())
	}
	if got := d.TitleOfPage(1).Display().String(); got != "[Part 2]" {
		t.Errorf("title of the second page %s", got)
	}
	if _, ok := d.Events()[1].(*Newpage); !ok {
		t.Errorf("event 1 is %T", d.Events()[1])
	}
}

func TestScenarioTeozPragmaHasNoEffect(t *testing.T) {
	d := parseModel(t, "!pragma teoz false\nAlice -> Bob : hi")
	if v, _ := d.Pragma().Value(klimt.PragmaTeoz); v != "false" {
		t.Errorf("pragma teoz %q", v)
	}
	if d.DefaultMargins() != klimt.ClockwiseSame(5) {
		t.Error("Teoz margins whatever the pragma")
	}
}

func TestScenarioBoxesOrderAndColors(t *testing.T) {
	d := parseModel(t, "box \"Internal\" #LightBlue\nparticipant B order 20\nparticipant A <<S>> order 10 #red\nend box\nparticipant C")
	if got := codes(d.Participants()); got != "C:PARTICIPANT A:PARTICIPANT B:PARTICIPANT" {
		t.Errorf("order %s", got)
	}
	a := d.Participants()[1]
	if d.Englober(a) == nil || d.Englober(a).Title().String() != "[Internal]" || d.Englober(d.Participants()[0]) != nil {
		t.Error("box membership")
	}
	if a.Stereotype() == nil || a.Stereotype().String() != "<<S>>" || a.Colors().String() == "{} null" {
		t.Errorf("stereotype %v colors %s", a.Stereotype(), a.Colors())
	}
}

func TestSyntaxErrorLine(t *testing.T) {
	d, _, err := parseCase("@startuml\nA -> B\nthis is wrong\n@enduml\n")
	if err != nil {
		t.Fatal(err)
	}
	e, ok := d.(*core.ErrorDiagram)
	if !ok || e.Error.Message() != "Syntax Error? (Assumed diagram type: sequence)" || e.Error.Position() != 2 {
		t.Fatalf("got %#v", d)
	}
}

func TestExportNotDrawable(t *testing.T) {
	d := parseModel(t, "A -> B")
	if err := d.Export(0); !errors.Is(err, ErrNotDrawable) {
		t.Errorf("Export = %v", err)
	}
}

func TestParallelParses(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := strings.Repeat("P", i+1)
			d, _, err := parseCase("@startuml\nautonumber\n" + name + " -> B : x\nskinparam ArrowColor red\n@enduml\n")
			sd, ok := d.(*SequenceDiagram)
			if err != nil || !ok || sd.Participants()[0].Code() != name {
				t.Errorf("parallel parse %d", i)
			}
		}()
	}
	wg.Wait()
}
