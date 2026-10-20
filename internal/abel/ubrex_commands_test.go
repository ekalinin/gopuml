// SPDX-License-Identifier: MIT
// Provenance: original

package abel

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

func ubrexCommands() []command.Command[*hookedDiagram] {
	return []command.Command[*hookedDiagram]{
		UBrexCommandFootboxIgnored[*hookedDiagram](),
		UBrexCommandRankDir[*hookedDiagram](),
		UBrexCommandHideShow2[*hookedDiagram](),
	}
}

func parseUBrex(t *testing.T, body ...string) *hookedDiagram {
	t.Helper()
	lines := []text.StringLocated{{S: "@startuml"}}
	for _, s := range body {
		lines = append(lines, text.StringLocated{S: s})
	}
	lines = append(lines, text.StringLocated{S: "@enduml"})
	result := uml.NewFactory(core.Class, ubrexCommands, newHookedDiagram).
		CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
	d, ok := result.(*hookedDiagram)
	if !ok {
		if e, isError := result.(*core.ErrorDiagram); isError {
			t.Fatalf("%q: error %s", body, e.Error.Message())
		}
		t.Fatalf("%q: result %T", body, result)
	}
	return d
}

// TestUBrexSingleExactMatch checks that a UBrex command accepts one line
// that its pattern matches exactly once trimmed, and nothing else.
func TestUBrexSingleExactMatch(t *testing.T) {
	c := UBrexCommandFootboxIgnored[*hookedDiagram]()
	for line, want := range map[string]command.Control{
		"hide footbox":       command.OK,
		"  show   footbox  ": command.OK,
		"hidefootbox":        command.OK,
		"hide footboxes":     command.NotOK,
		"hide footbox now":   command.NotOK,
		"footbox":            command.NotOK,
	} {
		if got := c.IsValid(command.SingleString(line)); got != want {
			t.Errorf("%q: %v, want %v", line, got, want)
		}
	}
	two := command.SingleString("hide footbox").Add(text.StringLocated{S: "hide footbox"})
	if got := c.IsValid(two); got != command.NotOK {
		t.Errorf("two lines: %v", got)
	}
}

// TestUBrexRankDir checks "left to right direction" and "top to bottom
// direction".
func TestUBrexRankDir(t *testing.T) {
	if got := parseUBrex(t, "left to right direction").Skin().Rankdir(); got != geom.RankdirLeftToRight {
		t.Errorf("left to right: %v", got)
	}
	if got := parseUBrex(t, "left to right direction", "top to bottom direction").Skin().Rankdir(); got !=
		geom.RankdirTopToBottom {
		t.Errorf("top to bottom: %v", got)
	}
}

// TestUBrexHideShow2 checks the hide and show rules of a stereotype and a
// name.
func TestUBrexHideShow2(t *testing.T) {
	d := parseUBrex(t, "hide <<stereo>>", "show Name", "hide-class Other")
	rules := d.Hides2()
	want := []struct {
		what string
		show bool
	}{{"<<stereo>>", false}, {"Name", true}, {"Other", false}}
	if len(rules) != len(want) {
		t.Fatalf("rules %v", rules)
	}
	for i, w := range want {
		if rules[i].What() != w.what || rules[i].Show() != w.show {
			t.Errorf("rule %d: %q %v, want %q %v", i, rules[i].What(), rules[i].Show(), w.what, w.show)
		}
	}
}
