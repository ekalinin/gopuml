// SPDX-License-Identifier: MIT
// Provenance: original

package packetdiag

import (
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// texts returns the texts r recorded.
func texts(r *rec) []string {
	var result []string
	for _, l := range *r.log {
		if _, s, ok := strings.Cut(l, " text "); ok {
			result = append(result, s)
		}
	}
	return result
}

func TestPacketBlockShape(t *testing.T) {
	sp := newDiagram().SkinParam()
	sb := font.NewStringBounderSvg(nil)
	for _, tc := range []struct {
		label         string
		width, height int32
		text          string
	}{
		{"Checksum", 1, 1, "C..."},
		{"B", 1, 1, "B"},
		{"Source Port", 8, 3, "Source Port"},
	} {
		b := NewPacketBlock(tc.width, tc.height, tc.label, sp, false, false)
		labelHeight := b.labelTextBlock(tc.text).CalculateDimension(sb).Height()
		dim := b.ShapeTextBlock(sb, 42, 32).CalculateDimension(sb)
		if want := float64(tc.width) * 42; dim.Width() != want {
			t.Errorf("%s: width %v, want %v", tc.label, dim.Width(), want)
		}
		if want := max(float64(tc.height)*32, labelHeight+20); dim.Height() != want {
			t.Errorf("%s: height %v, want %v", tc.label, dim.Height(), want)
		}
		r := newRec()
		b.ShapeTextBlock(sb, 42, 32).DrawU(r)
		if got := texts(r); !slices.Equal(got, []string{tc.text}) {
			t.Errorf("%s: texts %q, want %q", tc.label, got, tc.text)
		}
	}
}

func TestJavaSubstring(t *testing.T) {
	if got := javaSubstring("Check😀sum", 7); got != "Check😀" {
		t.Errorf("javaSubstring = %q", got)
	}
}

func TestPacketIndicatorDimension(t *testing.T) {
	d := newDiagram()
	sb := font.NewStringBounderSvg(nil)
	number := createNumberTextBlock(d.Style().FontConfiguration(d.SkinParam().IHtmlColorSet(), nil), d.SkinParam(), "8").
		CalculateDimension(sb)
	for _, tc := range []struct {
		full   bool
		height float64
	}{{true, number.Height() + VLineFull}, {false, number.Height() + VLineShort}} {
		dim := NewPacketIndicator(tc.full, true, 8, d.Style(), d.SkinParam()).CalculateDimension(sb)
		if dim.Width() != number.Width() || dim.Height() != tc.height {
			t.Errorf("full %v: %vx%v, want %vx%v", tc.full, dim.Width(), dim.Height(), number.Width(), tc.height)
		}
	}
}
