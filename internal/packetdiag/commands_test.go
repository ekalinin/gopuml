// SPDX-License-Identifier: MIT
// Provenance: original

package packetdiag

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// parse parses body (the lines between @startpacketdiag and @endpacketdiag)
// with Factory.
func parse(body ...string) core.Diagram {
	lines := []text.StringLocated{{S: "@startpacketdiag"}}
	for _, s := range body {
		lines = append(lines, text.StringLocated{S: s})
	}
	lines = append(lines, text.StringLocated{S: "@endpacketdiag"})
	return Factory.CreateSystem(core.NewRenderContext(), core.NewUmlSource(lines, lines, true))
}

// mustParse is parse for a body without error.
func mustParse(t *testing.T, body ...string) *PacketDiagram {
	t.Helper()
	result := parse(body...)
	d, ok := result.(*PacketDiagram)
	if !ok {
		if e, ok := result.(*core.ErrorDiagram); ok {
			t.Fatalf("%q: error at line %d: %s", body, e.Error.Position(), e.Error.Message())
		}
		t.Fatalf("%q: result %T", body, result)
	}
	return d
}

func TestCommandsStartAndEnd(t *testing.T) {
	for _, body := range [][]string{
		{"packetdiag {", "0: A", "}"},
		{"packetdiag", "0: A"},
		{"{", "0: A", "}"},
		{"", "0: A", "  }  "},
	} {
		if d := mustParse(t, body...); len(d.packetItems) != 1 {
			t.Errorf("%q: %d fields", body, len(d.packetItems))
		}
	}
}

func TestCommandNumRange(t *testing.T) {
	for _, tc := range []struct {
		line                       string
		start, end, height, rotate int32
		desc                       string
	}{
		{"3: text", 3, 3, 1, 0, "text"},
		{"3 text", 3, 3, 1, 0, "text"},
		{"0-7: Source Port", 0, 7, 1, 0, "Source Port"},
		{"* text", 0, 0, 1, 0, "text"},
		{"* A [len = 4]", 0, 3, 1, 0, "A"},
		{"2: A [len=3]", 2, 4, 1, 0, "A"},
		{"0-3: A [height = 3]", 0, 3, 3, 0, "A"},
		{"1: A [rotate=270]", 1, 1, 1, 270, "A"},
		{`1-10: server [color="palegreen", textcolor="blue", linecolor="red"]`, 1, 10, 1, 0, "server"},
		{"1: A [len = x, height = 2.5]", 1, 1, 1, 0, "A"},
		{"1: A [len = 99999999999]", 1, 1, 1, 0, "A"},
	} {
		d := mustParse(t, tc.line)
		p := d.packetItems[0]
		if p.bitStart != tc.start || p.bitEnd != tc.end || p.height != tc.height || p.textRotation != tc.rotate ||
			p.desc != tc.desc {
			t.Errorf("%q: %d-%d height %d rotate %d desc %q", tc.line, p.bitStart, p.bitEnd, p.height, p.textRotation, p.desc)
		}
	}
}

func TestCommandNumRangeAfterLastField(t *testing.T) {
	d := mustParse(t, "0-7: A", "* B [len = 2]", "* C")
	var got [][2]int32
	for _, p := range d.packetItems {
		got = append(got, [2]int32{p.bitStart, p.bitEnd})
	}
	if want := [][2]int32{{0, 7}, {8, 9}, {10, 10}}; len(got) != 3 || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("fields %v, want %v", got, want)
	}
}

func TestCommandDirectives(t *testing.T) {
	d := mustParse(t, "colwidth = 32;", "node_height=80", "scale_direction = rtl", "scale_interval = 4",
		"same_height = true", "0-31: A")
	if d.colWidth != 32 || d.bitHeight != 80 || d.scaleDirection != RTL || d.scaleInterval != 4 || !d.sameHeight {
		t.Errorf("colWidth %d, bitHeight %v, direction %v, interval %d, same height %v", d.colWidth, d.bitHeight,
			d.scaleDirection, d.scaleInterval, d.sameHeight)
	}
	d = mustParse(t, "COLWIDTH = 8", "scale_direction = ltr;", "same_height = false", "0-7: A")
	if d.colWidth != 8 || d.scaleDirection != LTR || d.sameHeight {
		t.Errorf("colWidth %d, direction %v, same height %v", d.colWidth, d.scaleDirection, d.sameHeight)
	}
}

func TestCommandsSyntaxError(t *testing.T) {
	result := parse("packetdiag {", "   foo: Source Port", "}")
	e, ok := result.(*core.ErrorDiagram)
	if !ok {
		t.Fatalf("result %T", result)
	}
	if got := e.Error.Message(); got != "Syntax Error? (Assumed diagram type: packet)" || e.Error.Line == nil ||
		e.Error.Line.S != "   foo: Source Port" {
		t.Errorf("error %q on %+v", got, e.Error.Line)
	}
}
