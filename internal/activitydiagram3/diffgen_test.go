// SPDX-License-Identifier: MIT
// Provenance: original

//go:build activityref

package activitydiagram3

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// The generated drawing cases mix the constructs of activity diagrams in
// random nested combinations (fixed seed, so the cases and their cached
// references are stable): if/elseif/else, switch, while and repeat with
// backward actions, fork/split/merge, partitions and groups, notes, arrows
// with labels and colors, kill/detach/break/stop, connectors and lanes.

const generatedDrawingCases = 700

type diagramGen struct {
	rnd    *rand.Rand
	sb     strings.Builder
	n      int
	lanes  []string
	inLoop int
}

func (g *diagramGen) line(depth int, format string, args ...any) {
	g.sb.WriteString(strings.Repeat("  ", depth))
	fmt.Fprintf(&g.sb, format, args...)
	g.sb.WriteString("\n")
}

func (g *diagramGen) pick(items ...string) string { return items[g.rnd.IntN(len(items))] }

func (g *diagramGen) label() string {
	g.n++
	switch g.rnd.IntN(6) {
	case 0:
		return fmt.Sprintf("step %d with a longer text", g.n)
	case 1:
		return fmt.Sprintf("multi\\nline %d", g.n)
	case 2:
		return fmt.Sprintf("**bold** %d", g.n)
	}
	return fmt.Sprintf("A%d", g.n)
}

func (g *diagramGen) action(depth int) {
	switch g.rnd.IntN(10) {
	case 0:
		// a colored action directly before a construct makes upstream lose
		// the construct, so a plain action follows it
		g.line(depth, "#%s:%s;", g.pick("pink", "lightblue", "palegreen"), g.label())
		g.line(depth, ":%s;", g.label())
	case 1:
		g.line(depth, ":%s; <<%s>>", g.label(), g.pick("input", "output", "procedure", "load", "save",
			"continuous", "task", "objectSignal", "acceptEvent", "sendSignal"))
	default:
		g.line(depth, ":%s;", g.label())
	}
}

func (g *diagramGen) arrow(depth int) {
	switch g.rnd.IntN(4) {
	case 0:
		g.line(depth, "-> %s;", g.label())
	case 1:
		g.line(depth, "-[#%s]-> %s;", g.pick("red", "blue", "green"), g.label())
	case 2:
		g.line(depth, "-[#%s,%s]-> %s;", g.pick("red", "blue"), g.pick("dashed", "dotted", "bold"), g.label())
	default:
		g.line(depth, "-[#%s]->;", g.pick("red", "blue", "green"))
	}
}

func (g *diagramGen) note(depth int) {
	side := g.pick("left", "right")
	switch g.rnd.IntN(3) {
	case 0:
		g.line(depth, "note %s: %s", side, g.label())
	case 1:
		g.line(depth, "note %s", side)
		g.line(depth+1, "%s", g.label())
		g.line(depth+1, "* item")
		g.line(depth, "end note")
	default:
		g.line(depth, "floating note %s: %s", side, g.label())
	}
}

// block writes count statements; it returns false when the flow stopped.
func (g *diagramGen) block(depth, count int) bool {
	for range count {
		if !g.statement(depth) {
			return false
		}
	}
	return true
}

func (g *diagramGen) statement(depth int) bool {
	r := g.rnd.IntN(100)
	if depth > 2 && r >= 45 {
		r = g.rnd.IntN(45)
	}
	switch {
	case r < 25:
		g.action(depth)
	case r < 32:
		g.arrow(depth)
		g.action(depth)
	case r < 37:
		g.action(depth)
		g.note(depth)
	case r < 40:
		g.line(depth, "(%s)", g.pick("A", "B", "C"))
	case r < 43 && len(g.lanes) > 1 && depth == 0:
		g.line(depth, "|%s|", g.lanes[g.rnd.IntN(len(g.lanes))])
		g.action(depth)
	case r < 45:
		return g.terminal(depth)
	case r < 58:
		g.ifStatement(depth)
	case r < 64:
		g.switchStatement(depth)
	case r < 72:
		g.whileStatement(depth)
	case r < 79:
		g.repeatStatement(depth)
	case r < 87:
		g.forkStatement(depth)
	case r < 91:
		g.splitStatement(depth)
	default:
		g.groupStatement(depth)
	}
	return true
}

func (g *diagramGen) terminal(depth int) bool {
	switch g.rnd.IntN(4) {
	case 0:
		g.action(depth)
		g.line(depth, "%s", g.pick("kill", "detach"))
		return false
	case 1:
		if g.inLoop > 0 {
			g.action(depth)
			g.line(depth, "break")
			return false
		}
		g.action(depth)
	case 2:
		g.line(depth, "%s", g.pick("stop", "end"))
		return false
	default:
		g.action(depth)
	}
	return true
}

func (g *diagramGen) branch(depth int) {
	g.block(depth, 1+g.rnd.IntN(2))
}

func (g *diagramGen) ifStatement(depth int) {
	g.line(depth, "if (%s?) then (%s)", g.label(), g.pick("yes", "y", "ok"))
	g.branch(depth + 1)
	for range g.rnd.IntN(3) {
		g.line(depth, "elseif (%s?) then (%s)", g.label(), g.pick("yes", "maybe"))
		g.branch(depth + 1)
	}
	if g.rnd.IntN(3) > 0 {
		g.line(depth, "else (%s)", g.pick("no", "n", "otherwise"))
		if g.rnd.IntN(4) > 0 {
			g.branch(depth + 1)
		}
	}
	g.line(depth, "endif")
}

func (g *diagramGen) switchStatement(depth int) {
	g.line(depth, "switch (%s?)", g.label())
	for i := range 1 + g.rnd.IntN(4) {
		g.line(depth, "case (%c)", 'A'+i)
		g.branch(depth + 1)
	}
	g.line(depth, "endswitch")
}

func (g *diagramGen) whileStatement(depth int) {
	g.inLoop++
	defer func() { g.inLoop-- }()
	g.line(depth, "while (%s?) is (%s)", g.label(), g.pick("yes", "more"))
	g.branch(depth + 1)
	g.line(depth, "endwhile (%s)", g.pick("no", "done"))
}

func (g *diagramGen) repeatStatement(depth int) {
	g.inLoop++
	defer func() { g.inLoop-- }()
	if g.rnd.IntN(3) == 0 {
		g.line(depth, "repeat :%s;", g.label())
	} else {
		g.line(depth, "repeat")
	}
	g.branch(depth + 1)
	if g.rnd.IntN(3) == 0 {
		g.line(depth+1, "backward:%s;", g.label())
	}
	g.line(depth, "repeat while (%s?) is (%s) not (%s)", g.label(), g.pick("yes", "again"), g.pick("no", "stop"))
}

func (g *diagramGen) forkStatement(depth int) {
	g.line(depth, "fork")
	g.branch(depth + 1)
	for range 1 + g.rnd.IntN(2) {
		g.line(depth, "fork again")
		g.branch(depth + 1)
	}
	g.line(depth, "%s", g.pick("end fork", "end merge", "end fork {or}", "end fork"))
}

func (g *diagramGen) splitStatement(depth int) {
	g.line(depth, "split")
	g.branch(depth + 1)
	for range 1 + g.rnd.IntN(2) {
		g.line(depth, "split again")
		g.branch(depth + 1)
	}
	g.line(depth, "end split")
}

func (g *diagramGen) groupStatement(depth int) {
	g.line(depth, "%s %s {", g.pick("partition", "group", "package", "rectangle", "card"), g.label())
	g.branch(depth + 1)
	g.line(depth, "}")
}

// generatedCases returns the random drawing cases.
func generatedDrawingCasesList() []testCase {
	var result []testCase
	for i := range generatedDrawingCases {
		g := &diagramGen{rnd: rand.New(rand.NewPCG(42, uint64(i)))}
		if g.rnd.IntN(4) == 0 {
			g.lanes = []string{"Lane A", "Lane B", "Lane C"}[:2+g.rnd.IntN(2)]
			g.line(0, "|%s|", g.lanes[0])
		}
		if g.rnd.IntN(2) == 0 {
			g.line(0, "start")
		}
		if g.block(0, 2+g.rnd.IntN(4)) && g.rnd.IntN(2) == 0 {
			g.line(0, "stop")
		}
		result = append(result, testCase{name: fmt.Sprintf("gen/%03d", i), source: g.sb.String()})
	}
	return result
}
