// SPDX-License-Identifier: MIT
// Provenance: original

package timing

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

func emptyDiagram(t *testing.T) *TimingDiagram {
	t.Helper()
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks("@startuml\n@enduml\n", rc.Desc, rc.Config)
	return NewTimingDiagram(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true), nil)
}

func TestDiagramDescription(t *testing.T) {
	if got := emptyDiagram(t).Description(); got != "(Timing Diagram)" {
		t.Errorf("description %q", got)
	}
}

func TestLastPlayer(t *testing.T) {
	d := emptyDiagram(t)
	robust := d.createPlayerRobust("R", "", false, nil, nil)
	d.createPlayerClock("C", "", jmath.BigDecimalOfInt64(10), jmath.BigDecimalZero, jmath.BigDecimalZero, nil)
	d.createPlayerBinary("B", "", nil)
	d.createPlayerAnalog("A", "a", nil)
	if d.getLastPlayer() != robust {
		t.Error("a clock, binary or analog player became the last player")
	}
	concise := d.createPlayerConcise("C2", "", false, nil, nil)
	if d.getLastPlayer() != concise {
		t.Error("a concise player is not the last player")
	}
	rectangle := d.createPlayerRectangle("R2", "", false, nil, nil)
	if d.getLastPlayer() != rectangle {
		t.Error("a rectangle player is not the last player")
	}
	d.setLastPlayer(robust)
	if d.getLastPlayer() != robust {
		t.Error("setLastPlayer")
	}
}

func TestRedeclaredPlayerKeepsItsPlace(t *testing.T) {
	d := emptyDiagram(t)
	d.createPlayerRobust("A", "", false, nil, nil)
	d.createPlayerConcise("B", "", false, nil, nil)
	second := d.createPlayerConcise("A", "", false, nil, nil)
	players := d.playersInOrder()
	if len(players) != 2 || players[0] != second {
		t.Errorf("players %v, want the second A first", players)
	}
}

func tickAt(s string) *TimeTick { return newTimeTick(jmath.BigDecimalOf(s), formatDecimal) }

func TestChangesAtTheSameTime(t *testing.T) {
	d := emptyDiagram(t)
	concise := d.createPlayerConcise("U", "", false, nil, nil).(*playerConcise)
	concise.setState(tickAt("10"), nil, color.EmptyColors(), "A")
	concise.setState(tickAt("10.0"), nil, color.EmptyColors(), "B")
	if len(concise.changes) != 1 || concise.changes[0].state() != "A" {
		t.Errorf("concise keeps %v, want the first change A", concise.changes)
	}
	binary := d.createPlayerBinary("B", "", nil).(*playerBinary)
	first := tickAt("10")
	binary.setState(first, nil, color.EmptyColors(), "high")
	binary.setState(tickAt("10.0"), nil, color.EmptyColors(), "low")
	if len(binary.values) != 1 || binary.values[0].tick != first || binary.values[0].value.state() != lowString {
		t.Errorf("binary keeps %+v, want the first tick with the last value", binary.values)
	}
	analog := d.createPlayerAnalog("A", "a", nil)
	analog.setState(tickAt("10"), nil, nil, "1")
	analog.setState(tickAt("10"), nil, nil, "2")
	if len(analog.timeSeries.values) != 1 || analog.timeSeries.values[0].value != 2 {
		t.Errorf("analog keeps %+v, want the last value 2", analog.timeSeries.values)
	}
}

func TestBinaryValues(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"high", highString}, {"HIGH", highString}, {"1", highString},
		{"low", lowString}, {"foo", lowString}, {"{hidden}", lowString}, {"0", lowString},
	} {
		if got := convertBinaryValue(tt.in); got != tt.want {
			t.Errorf("%s = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestAnalogValues(t *testing.T) {
	d := emptyDiagram(t)
	analog := d.createPlayerAnalog("A", "a", nil)
	analog.setState(tickAt("0"), nil, nil, "3")
	analog.setState(tickAt("10"), nil, nil, "abc")
	if analog.initialState == nil || *analog.initialState != 3 {
		t.Errorf("initial value %v, want the first value 3", analog.initialState)
	}
	if v, _ := analog.timeSeries.get(tickAt("10")); v != 0 {
		t.Errorf("abc = %v, want 0", v)
	}
	analog.setState(tickAt("20"), nil, nil, "4")
	if got := analog.timeSeries.valueAt(tickAt("15")); got != 2 {
		t.Errorf("value at 15 = %v, want 2 between 0 and 4", got)
	}
}

func TestStateLabels(t *testing.T) {
	d := emptyDiagram(t)
	p := d.createPlayerConcise("U", "", false, nil, nil).(*playerConcise)
	p.setState(tickAt("0"), nil, color.EmptyColors(), "W")
	p.defineState("W", "Long waiting state")
	p.setState(tickAt("10"), nil, color.EmptyColors(), "W")
	if p.changes[0].state() != "W" || p.changes[1].state() != "Long waiting state" {
		t.Errorf("states %q and %q, want the label applied from its declaration", p.changes[0].state(),
			p.changes[1].state())
	}
}

func TestClock(t *testing.T) {
	d := emptyDiagram(t)
	clock := d.createPlayerClock("clk", "", jmath.BigDecimalOfInt64(10), jmath.BigDecimalZero, jmath.BigDecimalZero,
		nil)
	tick := d.clockValue("clk", 3)
	if tick == nil || tick.time.ToPlainString() != "30" || tick.format != formatDecimal {
		t.Errorf("clk*3 = %v", tick)
	}
	if d.clockValue("other", 1) != nil {
		t.Error("an unknown clock has a value")
	}
	const want = "java.lang.UnsupportedOperationException"
	if got := panicMessage(func() { clock.setState(tickAt("0"), nil, nil, "high") }); got != want {
		t.Errorf("state of a clock: %q", got)
	}
	if got := panicMessage(func() { clock.defineState("a", "a") }); got != want {
		t.Errorf("state declaration of a clock: %q", got)
	}
	if got := panicMessage(func() {
		clock.createConstraint(tickAt("0"), tickAt("1"), "", false, klimt.ArrowConfigurationWithDirectionBoth())
	}); got != want {
		t.Errorf("constraint on a clock: %q", got)
	}
}
