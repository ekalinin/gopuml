// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"fmt"
	"testing"
)

func TestEnumNames(t *testing.T) {
	tests := []struct {
		v    fmt.Stringer
		want string
	}{
		{HorizontalAlignmentLeft, "LEFT"},
		{HorizontalAlignmentCenter, "CENTER"},
		{HorizontalAlignmentRight, "RIGHT"},
		{VerticalAlignmentTop, "TOP"},
		{VerticalAlignmentCenter, "CENTER"},
		{VerticalAlignmentBottom, "BOTTOM"},
		{VerticalPositionTop, "TOP"},
		{VerticalPositionBottom, "BOTTOM"},
		{SideNorth, "NORTH"},
		{SideEast, "EAST"},
		{SideSouth, "SOUTH"},
		{SideWest, "WEST"},
		{RankdirLeftToRight, "LEFT_TO_RIGHT"},
		{RankdirTopToBottom, "TOP_TO_BOTTOM"},
		{GraphicPositionBottom, "BOTTOM"},
		{GraphicPositionBackgroundCornerBottomRight, "BACKGROUND_CORNER_BOTTOM_RIGHT"},
		{GraphicPositionBackgroundCornerTopRight, "BACKGROUND_CORNER_TOP_RIGHT"},
		{ImgValignTop, "TOP"},
		{ImgValignBottom, "BOTTOM"},
		{ImgValignMiddle, "MIDDLE"},
		{SegMoveTo, "SEG_MOVETO"},
		{SegLineTo, "SEG_LINETO"},
		{SegQuadTo, "SEG_QUADTO"},
		{SegCubicTo, "SEG_CUBICTO"},
		{SegClose, "SEG_CLOSE"},
		{SegArcTo, "SEG_ARCTO"},
		{Side(7), "Side(7)"},
		{HorizontalAlignment(-1), "HorizontalAlignment(-1)"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestEnumValueOf(t *testing.T) {
	if v, ok := RankdirValueOf("TOP_TO_BOTTOM"); !ok || v != RankdirTopToBottom {
		t.Errorf("RankdirValueOf = %v, %v", v, ok)
	}
	if _, ok := RankdirValueOf("top_to_bottom"); ok {
		t.Error("valueOf is case sensitive")
	}
	if v, ok := ImgValignValueOf("MIDDLE"); !ok || v != ImgValignMiddle {
		t.Errorf("ImgValignValueOf = %v, %v", v, ok)
	}
	if v, ok := SideValueOf("WEST"); !ok || v != SideWest {
		t.Errorf("SideValueOf = %v, %v", v, ok)
	}
	if v, ok := HorizontalAlignmentValueOf("CENTER"); !ok || v != HorizontalAlignmentCenter {
		t.Errorf("HorizontalAlignmentValueOf = %v, %v", v, ok)
	}
	if v, ok := VerticalAlignmentValueOf("CENTER"); !ok || v != VerticalAlignmentCenter {
		t.Errorf("VerticalAlignmentValueOf = %v, %v", v, ok)
	}
	if v, ok := VerticalPositionValueOf("BOTTOM"); !ok || v != VerticalPositionBottom {
		t.Errorf("VerticalPositionValueOf = %v, %v", v, ok)
	}
	if v, ok := GraphicPositionValueOf("BACKGROUND_CORNER_TOP_RIGHT"); !ok || v != GraphicPositionBackgroundCornerTopRight {
		t.Errorf("GraphicPositionValueOf = %v, %v", v, ok)
	}
	if v, ok := USegmentTypeValueOf("SEG_ARCTO"); !ok || v != SegArcTo {
		t.Errorf("USegmentTypeValueOf = %v, %v", v, ok)
	}
	if _, ok := SideValueOf(""); ok {
		t.Error(`SideValueOf("") succeeded`)
	}
}

func TestHorizontalAlignmentFromString(t *testing.T) {
	// Expected results computed by upstream's fromString(s) and
	// fromString(s, CENTER).
	tests := []struct {
		s     string
		want  HorizontalAlignment
		ok    bool
		withD HorizontalAlignment
	}{
		{"left", HorizontalAlignmentLeft, true, HorizontalAlignmentLeft},
		{"LEFT", HorizontalAlignmentLeft, true, HorizontalAlignmentLeft},
		{"Center", HorizontalAlignmentCenter, true, HorizontalAlignmentCenter},
		{"rıght", HorizontalAlignmentRight, true, HorizontalAlignmentRight},
		{"Rİght", HorizontalAlignmentRight, true, HorizontalAlignmentRight},
		{"riGht", HorizontalAlignmentRight, true, HorizontalAlignmentRight},
		{"ſ", 0, false, HorizontalAlignmentCenter},
		{"middle", 0, false, HorizontalAlignmentCenter},
		{"", 0, false, HorizontalAlignmentCenter},
		{"r", 0, false, HorizontalAlignmentCenter},
		{"RIGHT ", 0, false, HorizontalAlignmentCenter},
		{"LEFT\x00", 0, false, HorizontalAlignmentCenter},
	}
	for _, tt := range tests {
		got, ok := HorizontalAlignmentFromString(tt.s)
		if got != tt.want || ok != tt.ok {
			t.Errorf("HorizontalAlignmentFromString(%q) = %v, %v, want %v, %v", tt.s, got, ok, tt.want, tt.ok)
		}
		if got := HorizontalAlignmentFromStringDefault(tt.s, HorizontalAlignmentCenter); got != tt.withD {
			t.Errorf("HorizontalAlignmentFromStringDefault(%q) = %v, want %v", tt.s, got, tt.withD)
		}
	}
	if got := HorizontalAlignmentFromStringDefault("x", HorizontalAlignmentRight); got != HorizontalAlignmentRight {
		t.Errorf("default = %v", got)
	}
}

func TestHorizontalAlignmentHelpers(t *testing.T) {
	tests := []struct {
		a        HorizontalAlignment
		graphViz string
		position float64
	}{
		{HorizontalAlignmentLeft, "l", 0},
		{HorizontalAlignmentCenter, "c", 3.375},
		{HorizontalAlignmentRight, "r", 6.75},
	}
	for _, tt := range tests {
		if got := tt.a.GraphVizValue(); got != tt.graphViz {
			t.Errorf("%v.GraphVizValue() = %q, want %q", tt.a, got, tt.graphViz)
		}
		sameBits(t, tt.a.String()+".Position", tt.a.Position(3.5, 10.25), tt.position)
	}
}

func TestVerticalAlignmentFromString(t *testing.T) {
	tests := []struct {
		s    string
		want VerticalAlignment
	}{
		{"top", VerticalAlignmentTop},
		{"TOP", VerticalAlignmentTop},
		{"tOp", VerticalAlignmentTop},
		{"center", VerticalAlignmentBottom},
		{"bottom", VerticalAlignmentBottom},
		{"x", VerticalAlignmentBottom},
		{"ţop", VerticalAlignmentBottom},
		{"", VerticalAlignmentBottom},
	}
	for _, tt := range tests {
		if got := VerticalAlignmentFromString(tt.s); got != tt.want {
			t.Errorf("VerticalAlignmentFromString(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestUSegmentTypeNbPoints(t *testing.T) {
	tests := []struct {
		t    USegmentType
		want int
	}{
		{SegMoveTo, 1},
		{SegLineTo, 1},
		{SegCubicTo, 3},
		{SegClose, 0},
	}
	for _, tt := range tests {
		if got := tt.t.NbPoints(); got != tt.want {
			t.Errorf("%v.NbPoints() = %d, want %d", tt.t, got, tt.want)
		}
	}
	mustPanic(t, "SegQuadTo.NbPoints", func() { SegQuadTo.NbPoints() })
	mustPanic(t, "SegArcTo.NbPoints", func() { SegArcTo.NbPoints() })
	if SegArcToValue != 4321 {
		t.Errorf("SegArcToValue = %d", SegArcToValue)
	}
	if PathIteratorWindEvenOdd != 0 || PathIteratorWindNonZero != 1 || PathIteratorSegMoveTo != 0 ||
		PathIteratorSegLineTo != 1 || PathIteratorSegQuadTo != 2 || PathIteratorSegCubicTo != 3 || PathIteratorSegClose != 4 {
		t.Error("PathIterator constants differ from java.awt.geom.PathIterator")
	}
}
