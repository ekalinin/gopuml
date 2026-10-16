// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestRectangleAreaIntersection(t *testing.T) {
	tests := []struct {
		maxX, maxY  float64
		c           XCubicCurve2D
		x, y, angle float64
	}{
		{fb(0x404903ff2f6e85df), fb(0x4044011a85ed6e38), cv(fb(0x4042ad8e775198ab), fb(0x402caa1fc076bb64), fb(0x405a9787f598324c), fb(0x40226996b5ae773c), fb(0x405b3ab25acf0100), fb(0x404dc068d28c7b38), fb(0x4063593d465936d8), fb(0x40617ca6b9de6403)), fb(0x4048c5937d933802), fb(0x402c076551b1101c), fb(0x3f9c844df8934879)},
		{fb(0x4048f58d8feda938), fb(0x4045a70f0d13ff63), cv(fb(0x405415c75278e4fe), fb(0x4056465f1e47bdfc), fb(0x4063c947937e453f), fb(0xc004e55e28822b30), fb(0x403d42c0a322479d), fb(0x4022410374e3a7f7), fb(0x403923eca4342102), fb(0x4030bcd46925f019)), fb(0x40496ff8855dea1e), fb(0x402775a96a794f65), fb(0x4008e5709f2b3563)},
		{fb(0x40490cb0d0363915), fb(0x4048ee7dc636f304), cv(fb(0x403902136e1ca860), fb(0x403f684d6836f3ed), fb(0x4057e1a035fe77d8), fb(0xc055940d7684930e), fb(0x405597d584128645), fb(0xc043e908ffda13c2), fb(0x4057c56dd733fe94), fb(0x40519e1cd77530c4)), fb(0x404610b8ed2beddc), fb(0x3fe08107b425d40e), fb(0xbfefddc7264e735a)},
		{fb(0x404906da56baf637), fb(0x40466598a5576619), cv(fb(0x405e02c8d8e858b2), fb(0x4052307da08367a0), fb(0x4051589029911bcf), fb(0x404d93d16db93dc7), fb(0x4028b6552d14e970), fb(0xc05311cc9887def0), fb(0x4034540f6f0050a6), fb(0x4033e93552c8a03b)), fb(0x4049074b312cad17), fb(0x401da16440057a5d), fb(0xc002136354ab83a4)},
		{fb(0x40490af573115d3a), fb(0x4043f894ef3d1a68), cv(fb(0x401ca4b7db5d9a2c), fb(0x403076f98fde3eb2), fb(0x405d32026ff1506e), fb(0x40231b9e8eb32268), fb(0x4051464412d567ee), fb(0x403c9ef3c651f500), fb(0x4063441c34ee2bd9), fb(0x4051a4522f6c73ee)), fb(0x4048d8bb1d1e54b0), fb(0x402e4b70ee8a4e56), fb(0x3f97cc579d2d5087)},
		{fb(0x40475af9558b759d), fb(0x4044adfeee3e00dc), cv(fb(0x4059c7773ff3b15a), fb(0x40519470ea0872a2), fb(0x4064090dc46534fc), fb(0xc038a80f8b81a160), fb(0x404d30566b3ccbb8), fb(0x404d98ba66b3e26c), fb(0x4033081254c89b09), fb(0x4043b6c5902371cc)), fb(0x4038f4419e89be76), fb(0x4044c6d8661577b5), fb(0xc007424311dcb55f)},
	}
	for _, tt := range tests {
		got, ok := NewRectangleArea(0, 0, tt.maxX, tt.maxY).Intersection(tt.c)
		if !ok {
			t.Errorf("Intersection(%v) found nothing", tt.c)
			continue
		}
		samePoint(t, "Intersection", got.Point2D(), XPoint2D{X: tt.x, Y: tt.y})
		nearAngle(t, "Intersection angle", got.Angle(), tt.angle)
	}
	if _, ok := NewRectangleArea(0, 0, 10, 10).Intersection(NewXCubicCurve2D(1, 1, 50, 50, 60, 60, 2, 2)); ok {
		t.Error("a curve with both ends inside has no intersection")
	}
	// Close end points give the start point directly.
	got, ok := NewRectangleArea(0, 0, 10, 10).Intersection(NewXCubicCurve2D(9.5, 5, 9.5, 5, 10.5, 5, 10.5, 5))
	if !ok || got.Point2D() != NewXPoint2D(9.5, 5) || got.Angle() != 0 {
		t.Errorf("short curve = %v, %v", got, ok)
	}
}

func TestRectangleAreaClosestSide(t *testing.T) {
	r := NewRectangleArea(0, 0, 10, 10)
	tests := []struct {
		pt   XPoint2D
		want Side
		ok   bool
	}{
		{NewXPoint2D(5, 1), SideNorth, true},
		{NewXPoint2D(5, 9), SideSouth, true},
		{NewXPoint2D(9, 5), SideEast, true},
		{NewXPoint2D(1, 5), SideWest, true},
		{NewXPoint2D(0, 0), SideNorth, true},
		{NewXPoint2D(10, 10), SideSouth, true},
		{NewXPoint2D(5, 5), SideNorth, true},
		{NewXPoint2D(math.NaN(), 5), 0, false},
	}
	for _, tt := range tests {
		got, ok := r.ClosestSide(tt.pt)
		if got != tt.want || ok != tt.ok {
			t.Errorf("ClosestSide(%v) = %v, %v, want %v, %v", tt.pt, got, ok, tt.want, tt.ok)
		}
	}
}

func TestRectangleArea(t *testing.T) {
	if got, want := NewRectangleArea(1, 2, 3.5, 1e8).String(), "minX=1.0 maxX=3.5 minY=2.0 maxY=1.0E8"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	r := RectangleAreaBuild(NewXPoint2D(10, 2), NewXPoint2D(4, 8))
	if r != NewRectangleArea(4, 2, 10, 8) {
		t.Fatalf("RectangleAreaBuild = %v", r)
	}
	sameBits(t, "Width", r.Width(), 6)
	sameBits(t, "Height", r.Height(), 6)
	if !r.Contains(4, 2) || r.Contains(10, 5) || !r.ContainsPoint(NewXPoint2D(9.9, 7.9)) || r.ContainsPoint(NewXPoint2D(5, 8)) {
		t.Error("Contains")
	}
	for _, tt := range []struct {
		name      string
		got, want RectangleArea
	}{
		{"Move", r.Move(1, -1), NewRectangleArea(5, 1, 11, 7)},
		{"Merge", r.Merge(NewRectangleArea(0, 5, 6, 20)), NewRectangleArea(0, 2, 10, 20)},
		{"MergePoint", r.MergePoint(NewXPoint2D(12, 0)), NewRectangleArea(4, 0, 12, 8)},
		{"WithMinX", r.WithMinX(1), NewRectangleArea(1, 2, 10, 8)},
		{"WithMaxX", r.WithMaxX(1), NewRectangleArea(4, 2, 1, 8)},
		{"WithMinY", r.WithMinY(1), NewRectangleArea(4, 1, 10, 8)},
		{"WithMaxY", r.WithMaxY(1), NewRectangleArea(4, 2, 10, 1)},
		{"AddMinX", r.AddMinX(1), NewRectangleArea(5, 2, 10, 8)},
		{"AddMaxX", r.AddMaxX(1), NewRectangleArea(4, 2, 11, 8)},
		{"AddMinY", r.AddMinY(1), NewRectangleArea(4, 3, 10, 8)},
		{"AddMaxY", r.AddMaxY(1), NewRectangleArea(4, 2, 10, 9)},
		{"Delta", r.Delta(1, 2), NewRectangleArea(4, 2, 11, 10)},
	} {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
	if r.MinX() != 4 || r.MinY() != 2 || r.MaxX() != 10 || r.MaxY() != 8 {
		t.Error("getters")
	}
	if got := r.PointCenter(); got != NewXPoint2D(7, 5) {
		t.Errorf("PointCenter() = %v", got)
	}
	if got := r.Dimension(); got != NewXDimension2D(6, 6) {
		t.Errorf("Dimension() = %v", got)
	}
	mustPanic(t, "Dimension of an inverted rectangle", func() { r.WithMaxX(0).Dimension() })
}
