// SPDX-License-Identifier: MIT
// Provenance: original

package timing

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// parseSource parses src with the timing factory; the diagram is nil when
// the factory does not accept the source.
func parseSource(t *testing.T, src string) (*TimingDiagram, core.Diagram) {
	t.Helper()
	rc := core.NewRenderContext()
	blocks := core.BuildBlocks(src, rc.Desc, rc.Config)
	result := Factory.CreateSystem(rc, core.NewUmlSource(blocks[0].Lines, blocks[0].Lines, true))
	d, _ := result.(*TimingDiagram)
	return d, result
}

func mustParse(t *testing.T, src string) *TimingDiagram {
	t.Helper()
	d, result := parseSource(t, src)
	if d == nil {
		t.Fatalf("not a timing diagram: %#v", result)
	}
	return d
}

func source(lines ...string) string {
	return "@startuml\n" + strings.Join(lines, "\n") + "\n@enduml\n"
}

// errorOf returns the error message of a source the timing factory rejects,
// without the assumed diagram type, and "" when it accepts it.
func errorOf(t *testing.T, lines ...string) string {
	t.Helper()
	_, result := parseSource(t, source(lines...))
	e, ok := result.(*core.ErrorDiagram)
	if !ok {
		return ""
	}
	return strings.TrimSuffix(e.Error.Message(), " (Assumed diagram type: timing)")
}

func TestCommandErrors(t *testing.T) {
	const npe = "You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.NullPointerException: "
	tests := []struct {
		lines []string
		want  string
	}{
		{[]string{`concise "U" as U #notacolor`}, "No such color"},
		{[]string{`concise "U" as U`, `@0`, `X is Idle`}, `Unknown "X"`},
		{[]string{`concise "U" as U`, `X has A`}, "Unknown X"},
		{[]string{`concise "U" as U`, `@:x`}, "What time?"},
		{[]string{`concise "U" as U`, `@:x+5`},
			npe + `Cannot invoke "net.sourceforge.plantuml.timingdiagram.TimeTick.getTime()" because "result" is null`},
		{[]string{`concise "U" as U`, `@X`}, "No such participant X"},
		{[]string{`binary "B" as B`, `0 is high`}, "Missing @ line before this"},
		{[]string{`concise "U" as U`, `@U`, `:x is A`},
			npe + `Cannot invoke "java.lang.Comparable.compareTo(Object)" because "k1" is null`},
		{[]string{`concise "U" as U`, `@0`, `@U`, `:x is A`},
			"You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.NullPointerException"},
		{[]string{`concise "U" as U`, `concise "W" as W`, `U -> X`}, "No such element: X"},
		{[]string{`binary "B" as B`, `@0 <-> @50`}, "You have to provide a participant"},
		{[]string{`concise "W" as W`, `W@:x <-> @50`}, "Unknown time label"},
		{[]string{`concise "U" as U`, `note top of X : x`}, `Unknown "X"`},
		{[]string{`concise "U" as U`, `scale 0 as 50 pixels`}, "Bad value"},
		{[]string{`concise "U" as U`, `use date format "yyyy-qq"`}, "Bad date format"},
		{[]string{`clock clk with period 10`, `clk is high`},
			"You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.UnsupportedOperationException"},
		{[]string{`concise "U" as U`, `@0 as :s`, `@:s+1.2.3`},
			"You should send a mail to plantuml@gmail.com with this log (V1.2026.9) java.lang.NumberFormatException: " +
				"Character array contains more than one decimal point."},
	}
	for _, tt := range tests {
		if got := errorOf(t, tt.lines...); got != tt.want {
			t.Errorf("%q: error %q, want %q", tt.lines, got, tt.want)
		}
	}
	// "analog" needs a quoted title.
	if _, result := parseSource(t, source(`analog V`)); result == nil {
		t.Error("analog V: no result")
	} else if _, ok := result.(*TimingDiagram); ok {
		t.Error("analog V is read")
	}
}

func TestPlayerDeclarations(t *testing.T) {
	d := mustParse(t, source(
		`robust "Web Browser" as WB`,
		`robust NT`,
		`concise "U" as U #lightblue`,
		`compact concise "C" as C`,
		`rectangle "Bus" as B`,
		`clock "Clock" as CL with period 12.5 pulse 2 offset 0.1`,
		`binary "Bin" as BI`,
		`analog "V" between 0 and 5.25 as V`))
	if p, ok := d.player("WB").(*playerRobust); !ok || p.title.IsWhite() {
		t.Error("robust player with a title")
	}
	if p := d.player("NT").base(); !p.title.IsWhite() {
		t.Error("robust player without a title has one")
	}
	if p := d.player("U").base(); p.generalBackgroundColor == nil || p.compact {
		t.Error("concise player with a background color")
	}
	if p := d.player("C").base(); !p.compact {
		t.Error("compact concise player is not compact")
	}
	if _, ok := d.player("B").(*playerRectangle); !ok {
		t.Error("rectangle player")
	}
	clock, ok := d.player("CL").(*playerClock)
	if !ok || clock.period.ToPlainString() != "12.5" || clock.pulse.ToPlainString() != "2" ||
		clock.offset.ToPlainString() != "0.1" {
		t.Errorf("clock %+v", clock)
	}
	if _, ok := d.player("BI").(*playerBinary); !ok {
		t.Error("binary player")
	}
	analog, ok := d.player("V").(*playerAnalog)
	if !ok || *analog.timeSeries.minOverride != 0 || *analog.timeSeries.maxOverride != 5.25 {
		t.Errorf("analog player %+v", analog)
	}
}

func TestStateCommands(t *testing.T) {
	d := mustParse(t, source(
		`robust "W" as WB`,
		`WB has A,B,C`,
		`WB has "Long" as L`,
		`@0`,
		`WB is Idle #pink : slow`,
		`@10`,
		`WB is {A,B}`,
		`@20`,
		`WB is L`))
	p := d.player("WB").(*playerRobust)
	if got := strings.Join(p.statesLabels(), ","); got != "A,B,C,Long" {
		t.Errorf("declared states %s", got)
	}
	first := p.changes[0]
	if first.state() != "Idle" || first.comment == nil || *first.comment != "slow" ||
		first.colors.Color(color.TypeBack) == nil {
		t.Errorf("first change %+v", first)
	}
	if got := strings.Join(p.changes[1].states, ","); got != "A,B" {
		t.Errorf("two states %s", got)
	}
	if got := p.changes[2].state(); got != "Long" {
		t.Errorf("labeled state %s", got)
	}
}

func TestTimeCommands(t *testing.T) {
	d := mustParse(t, source(
		`clock clk with period 10`,
		`concise "U" as U`,
		`@+50`,
		`U is A`,
		`@+5`,
		`U is B`,
		`@2.5`,
		`U is C`,
		`@0 as :start`,
		`@:start+70`,
		`U is D`,
		`@clk*3`,
		`U is E`))
	p := d.player("U").(*playerConcise)
	var times []string
	for _, change := range p.changes {
		times = append(times, change.when.time.ToPlainString())
	}
	if got := strings.Join(times, ","); got != "2.5,30,50,55,70" {
		t.Errorf("times %s, want +50 absolute without a current time and +5 relative to it", got)
	}
	hour := mustParse(t, source(`concise "U" as U`, `@0:15:00`, `U is A`, `@2024/07/15`, `U is B`))
	changes := hour.player("U").(*playerConcise).changes
	if changes[0].when.format != formatHour || changes[0].when.time.ToPlainString() != "900" {
		t.Errorf("0:15:00 = %v", changes[0].when)
	}
	if changes[1].when.format != formatDate || changes[1].when.time.ToPlainString() != "1721001600" {
		t.Errorf("2024/07/15 = %v", changes[1].when)
	}
	block := mustParse(t, source(`concise "U" as U`, `concise "W" as W`, `@W`, `0 is A`, `+20 is B`))
	if block.getLastPlayer() != block.player("W") || len(block.player("W").(*playerConcise).changes) != 2 {
		t.Error("@W does not start the block of W")
	}
}

func TestMessageAndConstraintCommands(t *testing.T) {
	d := mustParse(t, source(
		`concise "U" as WU`,
		`robust "W" as WB`,
		`@100`,
		`WU -> WB : URL`,
		`WU@0 -[#red]> WB@+50 : URL`,
		`WB@0 <-> @+50 : lag`,
		`highlight 0 to 50 #Gold;line:DimGrey : Peak`))
	if len(d.messages) != 2 || d.messages[0].tickInPlayer1.tick.time.ToPlainString() != "100" {
		t.Fatalf("messages %+v", d.messages)
	}
	second := d.messages[1]
	if second.tickInPlayer2.tick.time.ToPlainString() != "150" {
		t.Errorf("WB@+50 = %v, want 150, relative to the current time", second.tickInPlayer2.tick)
	}
	if c := second.SpecificColor(); c == nil || c.ToRGB(color.IdentityMapper) != "#FF0000" {
		t.Errorf("message color %v", c)
	}
	constraint := d.player("WB").(*playerRobust).constraints[0]
	if constraint.tick2.time.ToPlainString() != "50" {
		t.Errorf("@+50 of a constraint = %v, want 50, relative to its first time", constraint.tick2)
	}
	if len(d.highlights) != 1 || d.highlights[0].colors.Color(color.TypeBack) == nil ||
		d.highlights[0].colors.Color(color.TypeLine) == nil {
		t.Errorf("highlight %+v", d.highlights)
	}
}

func TestNoteCommands(t *testing.T) {
	d := mustParse(t, source(
		`concise "U" as WU`,
		`@10`,
		`note top of WU : first`,
		`note bottom of WU`,
		`line one`,
		`line two`,
		`end note`))
	notes := d.player("WU").base().notes
	if len(notes) != 2 || notes[0].position != abel.PositionTop || notes[1].position != abel.PositionBottom ||
		notes[1].note.Size() != 2 || notes[0].when.time.ToPlainString() != "10" {
		t.Errorf("notes %+v", notes)
	}
}

func TestOptionCommands(t *testing.T) {
	d := mustParse(t, source(
		`hide footbox`,
		`title Plan`,
		`mode compact`,
		`manual time-axis`,
		`scale 1 m as 60 pixels`,
		`use date format "MMM yyyy"`,
		`analog "V" as V`,
		`V ticks num on multiple 1`,
		`V is 200 pixels height`))
	if !d.compactByDefault || d.timeAxisStategy != TimeAxisManual || d.sdf == nil {
		t.Errorf("compact %v, axis %v, date format %v", d.compactByDefault, d.timeAxisStategy, d.sdf)
	}
	if d.ruler.forcedTickUnitary != 60 || d.ruler.tickIntervalInPixels != 60 {
		t.Errorf("scale %d as %d", d.ruler.forcedTickUnitary, d.ruler.tickIntervalInPixels)
	}
	v := d.player("V").(*playerAnalog)
	if v.ticksEvery == nil || *v.ticksEvery != 1 || v.suggestedHeight != 200 {
		t.Errorf("analog ticks %v, height %d", v.ticksEvery, v.suggestedHeight)
	}
	if d.Title().Display().IsWhite() {
		t.Error("no title")
	}
	if hidden := mustParse(t, source(`hide time-axis`)); hidden.timeAxisStategy != TimeAxisHidden {
		t.Error("hide time-axis")
	}
}
