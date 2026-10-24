// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// The layout tests check the positions computed by Teoz (the lifeline x
// of each participant, the y of each message arrow) against the
// coordinates of the reference SVGs rendered by the upstream jar.

var svgLine = regexp.MustCompile(`<line x1="([-0-9.]+)" y1="([-0-9.]+)" x2="([-0-9.]+)" y2="([-0-9.]+)"`)

type svgSegment struct{ x1, y1, x2, y2 float64 }

func parseSVGLines(t *testing.T, svg string) []svgSegment {
	t.Helper()
	var result []svgSegment
	for _, m := range svgLine.FindAllStringSubmatch(svg, -1) {
		var v [4]float64
		for i := range v {
			f, err := strconv.ParseFloat(m[i+1], 64)
			if err != nil {
				t.Fatal(err)
			}
			v[i] = f
		}
		result = append(result, svgSegment{v[0], v[1], v[2], v[3]})
	}
	return result
}

// near compares a computed coordinate with an SVG one, written with three
// decimals.
func near(a, b float64) bool { return math.Abs(a-b) < 0.0006 }

func TestTeozLayoutPositions(t *testing.T) {
	cases := []string{
		"sequence-diagram/implicit-participant",
		"sequence-diagram/declared-participant-kinds",
		"sequence-diagram/default-engine",
		"sequence-diagram/alt-fragment",
		"sequence-diagram/activation-return",
		"sequence-diagram/dotted-reply",
		"sequence-diagram/note-over-two",
		"sequence-diagram/divider",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(conformanceRoot, name+".puml"))
			if err != nil {
				t.Fatal(err)
			}
			svg, err := os.ReadFile(filepath.Join(curatedReferenceRoot, name, "block0-page0.svg"))
			if err != nil {
				t.Fatal(err)
			}
			lines := parseSVGLines(t, string(svg))
			parsed, _, err := parseCase(frontMatterBody(string(src)))
			if err != nil {
				t.Fatal(err)
			}
			d, ok := parsed.(*SequenceDiagram)
			if !ok {
				t.Fatalf("not a sequence diagram: %T", parsed)
			}
			stringBounder := font.NewStringBounderSvg(d.Skin())
			maker := NewSequenceDiagramFileMakerTeoz(d, rose.NewRose(), stringBounder)
			// The layout pass resolves the gauges and the livebox steps
			maker.TextBlock(0).CalculateDimension(stringBounder)

			// The page is drawn at (5, 5) inside the 5 pixel margins, the
			// body below the heads (and the box titles)
			livingSpaces := maker.body.playingSpace.LivingSpaces()
			dx := 10 - maker.min1.CurrentValue()
			dy := 10 + maker.heightEnglober1 + livingSpaces.HeadHeight(stringBounder)

			for _, ls := range livingSpaces.Values() {
				x := ls.PosC(stringBounder).CurrentValue() + dx
				found := false
				for _, l := range lines {
					if l.x1 == l.x2 && near(l.x1, x) {
						found = true
					}
				}
				if !found {
					t.Errorf("no lifeline of %s at x=%v", ls.Participant().Code(), x)
				}
			}

			messages := 0
			var check func(tiles []Tile)
			check = func(tiles []Tile) {
				for _, tile := range tiles {
					if g, ok := tile.(groupingTileLike); ok {
						check(g.grouping().tiles)
						continue
					}
					ct, ok := tile.(*CommunicationTile)
					if !ok {
						continue
					}
					messages++
					comp := ct.component(stringBounder)
					arrowY := comp.StartPoint(stringBounder, comp.PreferredDimension(stringBounder)).Y
					y := ct.YGauge().Min().CurrentValue() + arrowY + dy
					found := false
					for _, l := range lines {
						if l.y1 == l.y2 && near(l.y1, y) {
							found = true
						}
					}
					if !found {
						t.Errorf("no arrow of %v at y=%v", ct.message.Label(), y)
					}
				}
			}
			check(maker.body.playingSpace.Tiles())
			if messages == 0 {
				t.Error("no message checked")
			}
		})
	}
}

func TestRealSolver(t *testing.T) {
	origin := RealUtilsCreateOrigin()
	a := origin.AddAtLeast(10)
	b := a.AddFixed(5)
	c := origin.AddAtLeast(0)
	c.EnsureBiggerThan(b.AddFixed(20))
	m := RealUtilsMax(a, c)
	mid := RealUtilsMiddle(a, c.AddFixed(0))
	origin.CompileNow()
	if a.CurrentValue() != 10 || b.CurrentValue() != 15 || c.CurrentValue() != 35 {
		t.Errorf("a=%v b=%v c=%v", a.CurrentValue(), b.CurrentValue(), c.CurrentValue())
	}
	if m.CurrentValue() != 35 || mid.CurrentValue() != 22.5 {
		t.Errorf("max=%v middle=%v", m.CurrentValue(), mid.CurrentValue())
	}

	// A maximum is cached on its first evaluation, as upstream
	o2 := RealUtilsCreateOrigin()
	p := o2.AddAtLeast(0)
	early := RealUtilsMax(p)
	if early.CurrentValue() != 0 {
		t.Fatal(early.CurrentValue())
	}
	p.EnsureBiggerThan(o2.AddFixed(7))
	o2.CompileNow()
	if p.CurrentValue() != 7 || early.CurrentValue() != 0 || RealUtilsMax(p).CurrentValue() != 7 {
		t.Errorf("p=%v early=%v", p.CurrentValue(), early.CurrentValue())
	}

	// Pushing a middle moves both of its points by half of the push, so
	// the solver converges to the bound, within its tolerance
	o3 := RealUtilsCreateOrigin()
	x1 := o3.AddAtLeast(0)
	x2 := o3.AddAtLeast(10)
	middle := RealUtilsMiddle(x1, x2)
	middle.EnsureBiggerThan(o3.AddFixed(15))
	o3.CompileNow()
	if math.Abs(x1.CurrentValue()-10) > 1e-5 || math.Abs(x2.CurrentValue()-20) > 1e-5 {
		t.Errorf("x1=%v x2=%v", x1.CurrentValue(), x2.CurrentValue())
	}

	// A cycle of constraints panics after 99999 passes
	o4 := RealUtilsCreateOrigin()
	y1 := o4.AddAtLeast(0)
	y2 := y1.AddAtLeast(1)
	y1.EnsureBiggerThan(y2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("no panic for a cycle")
		}
	}()
	o4.CompileNow()
}

func TestSegmentCut(t *testing.T) {
	full := NewSegment(0, 100)
	got := full.CutSegmentIfNeed([]Segment{NewSegment(60, 70), NewSegment(20, 30), NewSegment(150, 160)})
	want := []Segment{NewSegment(0, 20), NewSegment(30, 60), NewSegment(70, 100)}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("segment %d = %v, want %v", i, got[i], want[i])
		}
	}
}

// A participant name sits in the room that MinimumWidth adds according to
// the HorizontalAlignment of the style: left, centered (also without an
// alignment) or right.
func TestParticipantMinimumWidthAlignment(t *testing.T) {
	nameX := func(alignment string) float64 {
		t.Helper()
		src := "@startuml\n<style>\nsequenceDiagram {\n  participant {\n    MinimumWidth 100\n" + alignment +
			"  }\n}\n</style>\nparticipant \"A\" as A\nA -> B : Message\n@enduml\n"
		svg, err := renderSVG(t, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		m := regexp.MustCompile(`<text x="([-\d.]+)"[^>]*>A</text>`).FindStringSubmatch(svg)
		if m == nil {
			t.Fatalf("no name A in %s", svg)
		}
		x, _ := strconv.ParseFloat(m[1], 64)
		return x
	}
	left := nameX("    HorizontalAlignment left\n")
	center := nameX("    HorizontalAlignment center\n")
	right := nameX("    HorizontalAlignment right\n")
	if def := nameX(""); def != center {
		t.Errorf("default %v, center %v", def, center)
	}
	if !(left < center && center < right) || math.Abs((right-left)-2*(center-left)) > 0.002 {
		t.Errorf("left %v, center %v, right %v", left, center, right)
	}
}

// A line that starts and ends with U+00A0 is parsed as the same line without
// them, as upstream trims U+00A0 with the other blanks.
func TestNonBreakingSpaceTrimmed(t *testing.T) {
	plain, err := renderSVG(t, "@startuml\nAlice -> Bob : hi\n@enduml\n", 0)
	if err != nil {
		t.Fatal(err)
	}
	nbsp, err := renderSVG(t, "@startuml\n Alice -> Bob : hi \n@enduml\n", 0)
	if err != nil {
		t.Fatal(err)
	}
	strip := regexp.MustCompile(`<\?plantuml-src [^?]*\?>`)
	if strip.ReplaceAllString(plain, "") != strip.ReplaceAllString(nbsp, "") {
		t.Errorf("the lines with U+00A0 render differently:\n%s\n%s", plain, nbsp)
	}
}
