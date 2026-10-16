// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"errors"
	"sync"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

func parse(t *testing.T, source string) core.Diagram {
	t.Helper()
	d, preproc, err := parseCase(source)
	if err != nil || preproc {
		t.Fatalf("parse: %v %v", err, preproc)
	}
	return d
}

func TestDiagramIdentity(t *testing.T) {
	d, ok := parse(t, "start\n:a;\nstop\n").(*ActivityDiagram3)
	if !ok {
		t.Fatalf("got %T", d)
	}
	if d.Description() != "activity3" || d.DiagramType() != core.Activity || d.Pages() != 1 {
		t.Errorf("description %q type %v pages %d", d.Description(), d.DiagramType(), d.Pages())
	}
	if Factory.DiagramType() != core.Activity {
		t.Errorf("factory type %v", Factory.DiagramType())
	}
	if err := d.Export(0); !errors.Is(err, ErrNotDrawable) {
		t.Errorf("Export = %v", err)
	}
}

func TestInstructionTree(t *testing.T) {
	d := parse(t, "|A|\nstart\nif (c) then (yes)\n:a;\nelse (no)\n:b;\nendif\nstop\n").(*ActivityDiagram3)
	root := d.Swimlanes().Root().(*InstructionList)
	if len(root.All()) != 3 || d.Swimlanes().Current() != Instruction(root) {
		t.Fatalf("root %d instructions", len(root.All()))
	}
	i := root.All()[1].(*InstructionIf)
	if !i.IsEndifCalled() || i.Parent() != Instruction(root) || len(i.Thens()) != 1 || i.ElseBranch() == nil {
		t.Fatalf("if %+v", i)
	}
	if got := i.Thens()[0].LabelPositive().Display().String(); got != "[yes]" {
		t.Errorf("then label %s", got)
	}
	a := i.Thens()[0].Last().(*InstructionSimple)
	if a.Label().String() != "[a]" || a.SwimlaneIn().Name() != "A" {
		t.Errorf("action %s in %v", a.Label(), a.SwimlaneIn())
	}
	if lanes := i.Swimlanes(); len(lanes) != 1 || lanes[0].Name() != "A" {
		t.Errorf("lanes %v", lanes)
	}
}

func TestBoxStyleFromString(t *testing.T) {
	tests := map[string]BoxStyle{
		"input": BoxStyleSdlInput, "INPUT": BoxStyleSdlInput, "send-signal": BoxStyleUmlSendSignal,
		"objectSignal": BoxStyleUmlObjectSignal, "timeEvent": BoxStyleUmlTimeEvent, "nothing": BoxStylePlain, "": BoxStylePlain,
	}
	for s, want := range tests {
		if got := BoxStyleFromString(s); got != want {
			t.Errorf("BoxStyleFromString(%q) = %v, want %v", s, got, want)
		}
	}
	if _, ok := BoxStylePlain.Name(); ok || BoxStylePlain.Stereotype() != nil {
		t.Error("plain style has a name")
	}
	if name, _ := BoxStyleSdlOutput.Name(); name != "output" || BoxStyleSdlOutput.Shield() != 10 {
		t.Errorf("output style %q %v", name, BoxStyleSdlOutput.Shield())
	}
	if got := boxStyleOf(klimt.StereogroupBuild("<<#red>> <<task>> <<input>>")); got != BoxStyleSdlTask {
		t.Errorf("boxStyleOf = %v", got)
	}
}

func TestRainbow(t *testing.T) {
	d := parse(t, ":a;\n").(*ActivityDiagram3)
	r, err := RainbowBuild(d.Skin(), "#red,dashed;#blue", true, 3)
	if err != nil {
		t.Fatal(err)
	}
	if r.Size() != 2 || r.ColorArrowSeparationSpace() != 3 || r.Colors()[0].Style().String() != "DASHED(null)" {
		t.Errorf("rainbow %v", r)
	}
	if _, err := RainbowBuild(d.Skin(), "thickness=2", true, 0); !errors.As(err, new(*color.NoSuchColorError)) {
		t.Errorf("thickness error = %v", err)
	}
	if r, _ := RainbowBuild(d.Skin(), "", false, 0); r.Size() != 0 {
		t.Errorf("none %v", r)
	}
	if RainbowFromColor(nil, nil).Size() != 0 {
		t.Error("nil color")
	}
	if !LinkRenderingNone().IsNone() || LinkRenderingNone().WithDisplay(klimt.DisplayCreate("x")).IsNone() {
		t.Error("IsNone")
	}
}

func TestSwimlaneOrder(t *testing.T) {
	a, b := NewSwimlane("a", 0, klimt.PragmaCreateEmpty()), NewSwimlane("b", 1, klimt.PragmaCreateEmpty())
	if a.CompareTo(b) >= 0 || b.CompareTo(a) <= 0 || a.CompareTo(a) != 0 {
		t.Error("CompareTo")
	}
	if !a.IsSmallerThanAllOthers(SwimlaneSet{a, b}) || b.IsSmallerThanAllOthers(SwimlaneSet{a, b}) ||
		a.IsSmallerThanAllOthers(SwimlaneSet{a}) {
		t.Error("IsSmallerThanAllOthers")
	}
	set := SwimlaneSet(nil).add(a).add(a).add(nil).add(b)
	if len(set) != 3 || !set.Contains(nil) {
		t.Errorf("set %v", set)
	}
}

// TestConcurrentParse parses in parallel: the factory and its commands are
// shared, the diagrams are not (go test -race).
func TestConcurrentParse(t *testing.T) {
	sources := []string{
		"start\n:a;\nif (c) then\n:b;\nendif\nstop\n",
		"|A|\nfork\n:a;\nfork again\n|B|\n:b;\nend fork\n",
		"repeat\n:a;\nbackward :b;\nrepeat while (c)\n",
		"switch (x)\ncase (a)\n:a;\nendswitch\nnote right: n\n",
	}
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			want := dumpCase(testCase{name: "x", source: src})
			for range 5 {
				got := dumpCase(testCase{name: "x", source: src})
				if len(got) != len(want) {
					t.Errorf("unstable dump of %q", src)
					return
				}
			}
		}(sources[i%len(sources)])
	}
	wg.Wait()
}
