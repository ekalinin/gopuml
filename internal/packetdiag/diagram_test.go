// SPDX-License-Identifier: MIT
// Provenance: original

package packetdiag

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
)

func newDiagram() *PacketDiagram {
	lines := []text.StringLocated{{S: "@startpacketdiag"}, {S: "@endpacketdiag"}}
	return NewPacketDiagram(core.NewRenderContext(), core.NewUmlSource(lines, lines, true), nil)
}

// grid describes the rows of d: the width and height of each block, with
// "<" and ">" for open edges.
func grid(d *PacketDiagram) string {
	var rows []string
	for _, row := range d.packetGrid {
		var blocks []string
		for _, b := range row {
			s := fmt.Sprintf("%s%dx%d", b.label, b.width, b.height)
			if b.leftOpen {
				s = "<" + s
			}
			if b.rightOpen {
				s += ">"
			}
			blocks = append(blocks, s)
		}
		rows = append(rows, strings.Join(blocks, " "))
	}
	return strings.Join(rows, " | ")
}

// indicators describes the ticks of d: the bit number, "F" for a full tick
// and "#" for a numbered one.
func indicators(d *PacketDiagram) string {
	var ticks []string
	for _, i := range d.packetIndicators {
		s := i.bitNumber
		if i.full {
			s += "F"
		}
		if i.numbered {
			s += "#"
		}
		ticks = append(ticks, s)
	}
	return strings.Join(ticks, " ")
}

func TestPacketDiagramDefaults(t *testing.T) {
	d := newDiagram()
	d.SetColWidth(0)
	d.AddPacketItemRange(0, 15, 1, "A", 0)
	d.build()
	if d.ColWidth() != 16 || d.scaleInterval != 8 {
		t.Errorf("colWidth %d, scaleInterval %d", d.ColWidth(), d.scaleInterval)
	}
	if got, want := indicators(d), "0F# 1 2 3 4F 5 6 7 8F# 9 10 11 12F 13 14 15 16F#"; got != want {
		t.Errorf("indicators %q, want %q", got, want)
	}
	if d.bitWidth != 42 || d.bitHeight != 32 {
		t.Errorf("bit %vx%v", d.bitWidth, d.bitHeight)
	}
}

func TestPacketDiagramScaleInterval(t *testing.T) {
	d := newDiagram()
	d.SetColWidth(8)
	d.UpdateScaleInterval(20)
	d.UpdateScaleInterval(0)
	d.AddPacketItemRange(0, 7, 1, "A", 0)
	d.build()
	if d.scaleInterval != 8 {
		t.Errorf("scaleInterval %d, want 8", d.scaleInterval)
	}
}

func TestPacketDiagramLayout(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(d *PacketDiagram)
		grid  string
		ticks string
	}{
		{"field after an 8-bit field", func(d *PacketDiagram) {
			d.AddPacketItemRange(0, 7, 1, "A", 0)
			d.AddPacketItemRange(8, 23, 1, "B", 0)
		}, "A8x1 B8x1> | <B8x1", ""},
		{"17 bits on 16-bit rows", func(d *PacketDiagram) {
			d.AddPacketItemRange(1, 17, 1, "AB", 0)
		}, "AB16x1> | <AB1x1", ""},
		{"16 bits on 8-bit rows", func(d *PacketDiagram) {
			d.SetColWidth(8)
			d.AddPacketItemRange(0, 15, 1, "Long", 0)
		}, "Long8x1> | <Long8x1", "0F# 1 2F 3 4F# 5 6F 7 8F#"},
		{"narrow first row", func(d *PacketDiagram) {
			d.AddPacketItemRange(0, 3, 1, "A", 0)
			d.AddPacketItemRange(4, 4, 1, "B", 0)
		}, "A4x1 B1x1", "0F# 1F 2F 3F 4F 5F"},
		{"same height", func(d *PacketDiagram) {
			d.SetSameHeight(true)
			d.AddPacketItemRange(0, 3, 3, "A", 0)
			d.AddPacketItemRange(4, 7, 1, "B", 0)
			d.AddPacketItemRange(8, 15, 1, "C", 0)
		}, "A4x3 B4x3 C8x3", ""},
		{"different heights", func(d *PacketDiagram) {
			d.AddPacketItemRange(0, 3, 3, "A", 0)
			d.AddPacketItemRange(4, 7, 1, "B", 0)
		}, "A4x3 B4x1", ""},
		{"right to left", func(d *PacketDiagram) {
			d.SetScaleDirection("rtl")
			d.SetColWidth(8)
			d.AddPacketItemRange(0, 3, 1, "A", 0)
			d.AddPacketItemRange(4, 7, 1, "B", 0)
		}, "B4x1 A4x1", "8F# 7 6F 5 4F# 3 2F 1 0F#"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := newDiagram()
			tc.setup(d)
			d.build()
			if got := grid(d); got != tc.grid {
				t.Errorf("grid %q, want %q", got, tc.grid)
			}
			if got := indicators(d); tc.ticks != "" && got != tc.ticks {
				t.Errorf("indicators %q, want %q", got, tc.ticks)
			}
		})
	}
}

// javaException returns the Java exception build panics with.
func javaException(t *testing.T, d *PacketDiagram) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		var e *color.RuntimeError
		if err, ok := r.(error); !ok || !errors.As(err, &e) {
			t.Fatalf("panic %v, want a Java exception", r)
		}
		msg = e.Error()
	}()
	d.build()
	return ""
}

func TestPacketDiagramJavaExceptions(t *testing.T) {
	for _, tc := range []struct {
		name       string
		start, end int32
		want       string
	}{
		{"4-3", 4, 3, "java.lang.ArithmeticException: / by zero"},
		{"* A [len = 0]", 0, -1, "java.lang.ArithmeticException: / by zero"},
		{"5-2", 5, 2, "java.lang.IllegalArgumentException: -1"},
	} {
		d := newDiagram()
		d.AddPacketItemRange(tc.start, tc.end, 1, "x", 0)
		if got := javaException(t, d); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestPacketDiagramLastPacketEnd(t *testing.T) {
	d := newDiagram()
	if _, ok := d.LastPacketEnd(); ok {
		t.Error("an empty diagram has a last packet")
	}
	d.AddPacketItemRange(0, 3, 1, "A", 0)
	d.AddPacketItemRange(9, 12, 1, "B", 0)
	if end, ok := d.LastPacketEnd(); !ok || end != 12 {
		t.Errorf("last end %d, %v", end, ok)
	}
	var widths []int32
	for _, p := range d.packetItems {
		widths = append(widths, p.width)
	}
	if !slices.Equal(widths, []int32{4, 4}) {
		t.Errorf("widths %v", widths)
	}
}
