// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestCoordinateChange(t *testing.T) {
	tests := []struct{ x1, y1, x2, y2, a, b, length, x, y float64 }{
		{fb(0x3fd15f1f78c0b048), fb(0xbfd913014798619c), fb(0xc0581b1c2560abdf), fb(0x402aae31931f5528), fb(0xc054e4e25721f28f), fb(0x4051f84262adc118), fb(0x40586a930034e21e), fb(0x40523a4f92ff0435), fb(0xc054d3b405dcfbda)},
		{fb(0xc0538d7d81f40544), fb(0x4040a552b5fe4a40), fb(0x4041061ebedd88f4), fb(0x3fc05814f61927b8), fb(0xbfe17381f16927ba), fb(0xc01663d0ebf75d90), fb(0x405d43829bd95c8f), fb(0xc05414755d012559), fb(0x403c13f5ec2cc47d)},
		{fb(0xbfbc52887e131c71), fb(0x401e829e0dac7d30), fb(0x3fec5e4a5583a8e1), fb(0xbfea4fad65937055), fb(0x40576f49a59a798c), fb(0x3fee21623a2e6b0a), fb(0x4021044ee884e249), fb(0x40279ed0ebe36a20), fb(0xc05556b85f4ffa01)},
		{fb(0x3fe1777b49f0dc67), fb(0xbfeae0a111f16d80), fb(0x3fddd992895f2ff6), fb(0xbfe0d746d7d71e80), fb(0xc04399416d057158), fb(0xbfe5ddb6e30afff1), fb(0x3fd4b4ef37e38863), fb(0x4025a99b14da3e6f), fb(0xc04355beb0423349)},
		{fb(0xc0344ad90ec3c088), fb(0x404e62558b734350), fb(0xc03c727231d4fdc4), fb(0xbfeb7224543f3076), fb(0x3fe2691de6ac086e), fb(0x40293a4477bbe558), fb(0x404f14e0e893604b), fb(0xc01f73c30ae9eb82), fb(0x404d4586a984a3b0)},
		{fb(0xc032884fefd841dc), fb(0x3fecd825e0575c61), fb(0x3faca53d14eb5f5c), fb(0xc04910fd94c1ea5f), fb(0x3fea82f8fe154b64), fb(0x402c6a40c15d1010), fb(0x404b28314cb8a746), fb(0xc01398f156945358), fb(0x4013f0f4e67afbdb)},
		{fb(0xc052c8dff0bf4700), fb(0xc049ffecc075f9a4), fb(0xbfe2a032ecfd1564), fb(0x405304bb296d1cf2), fb(0xc03f1447c4a0c19c), fb(0xbfe12d2135fb17cc), fb(0x406286356bd515fc), fb(0xc05693e3e3067329), fb(0xc053c8410f00369e)},
		{fb(0x3fe8fa5c5ec47790), fb(0xc046267daf4187f6), fb(0xc0548edb8ce5f089), fb(0xc0438e96eab492fa), fb(0x3fcb41b03ab057f1), fb(0x3fde199e97765fec), fb(0x4054cb2c7db278f9), fb(0x3fe13d1306fe5801), fb(0xc04660dfcdcfef12)},
		{fb(0xc044df73c22de98b), fb(0x4049cff157519d60), fb(0xc04c6f00598bec8f), fb(0x4046a17f94d01d50), fb(0x40565596256bc5b4), fb(0xc0582e48d9b3d153), fb(0x403067d7ce2f8be9), fb(0xc0643351a2e2a59a), fb(0x405a882b9e4f1bfa)},
		{fb(0xbfe31924de0e4072), fb(0x4055ec6f0add27f6), fb(0x3fb7c6a518dd2215), fb(0x404613784bd4a4f0), fb(0xc03517c2155a9df0), fb(0x3f4915059dc39eb8), fb(0x4045c618c3da902e), fb(0xbfedc381748466c2), fb(0x405b32346ad1416e)},
	}
	for _, tt := range tests {
		c := NewCoordinateChange(tt.x1, tt.y1, tt.x2, tt.y2)
		sameBits(t, "Length", c.Length(), tt.length)
		samePoint(t, "TrueCoordinate", c.TrueCoordinate(tt.a, tt.b), XPoint2D{X: tt.x, Y: tt.y})
	}
	c := CoordinateChangeCreate(NewXPoint2D(1, 1), NewXPoint2D(1, 3))
	if got := c.TrueCoordinate(2, 1); got != NewXPoint2D(0, 3) {
		t.Errorf("TrueCoordinate(2, 1) = %v", got)
	}
	mustPanic(t, "NewCoordinateChange of equal points", func() { NewCoordinateChange(1, 2, 1, 2) })
}

func TestPointAndAngle(t *testing.T) {
	p := NewPointAndAngle(NewXPoint2D(1.5, -2), 0.25)
	if p.Pt() != NewXPoint2D(1.5, -2) || p.Angle() != 0.25 || p.X() != 1.5 || p.Y() != -2 {
		t.Errorf("PointAndAngle = %v", p)
	}
	d := NewPointDirected(NewXPoint2D(3, 4), math.Pi)
	if d.Point2D() != NewXPoint2D(3, 4) || d.Angle() != math.Pi {
		t.Errorf("PointDirected = %v", d)
	}
}

func TestBasicEnsureVisible(t *testing.T) {
	var ev EnsureVisible = NewBasicEnsureVisible()
	b := ev.(*BasicEnsureVisible)
	if got := b.Coords(2); got != "0,0,0,0" {
		t.Errorf("empty Coords = %q", got)
	}
	if b.Surface() != 0 || b.HasData() {
		t.Error("empty tracker has data")
	}
	ev.EnsureVisible(-3.7, 2.2)
	ev.EnsureVisible(10.9, -5.5)
	// Expected values computed by upstream.
	tests := []struct {
		scale float64
		want  string
	}{
		{1.5, "-5,-8,16,3"},
		{1e10, "-2147483648,-2147483648,2147483647,2147483647"},
		{math.NaN(), "0,0,0,0"},
	}
	for _, tt := range tests {
		if got := b.Coords(tt.scale); got != tt.want {
			t.Errorf("Coords(%v) = %q, want %q", tt.scale, got, tt.want)
		}
	}
	sameBits(t, "Surface", b.Surface(), fb(0x405c1ae147ae147c))
	if !b.HasData() {
		t.Error("HasData() = false")
	}
}
