// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import "testing"

func TestPositionableImpl(t *testing.T) {
	var p Positionable = PositionableImplCreate(NewXPoint2D(1, 2), NewXDimension2D(3, 4))
	p.MoveDelta(10, 20)
	if p.Position() != NewXPoint2D(11, 22) || p.Size() != NewXDimension2D(3, 4) {
		t.Errorf("after MoveDelta: %v %v", p.Position(), p.Size())
	}
}

func TestPositionableUtilsMoveAwayFrom(t *testing.T) {
	tests := []struct{ x, y, w, h, mx, my, wantX, wantY float64 }{
		{fb(0x402d2a03487e9180), fb(0x40548d0e164e06ba), fb(0x4042f9e170e09bb5), fb(0x40374a276629fbf7), fb(0x4036e3244d1e0252), fb(0x4053df28a5099eb2), fb(0x4034a6b2676a083e), fb(0x4052c3b3f7d8d664)},
		{fb(0xbfb27555869b245c), fb(0xc04d5306a36c0590), fb(0x403ae776a1b4c9c0), fb(0x4032b71ad16d547f), fb(0xbffde32b671affd3), fb(0xc04d4ea5dffd5530), fb(0xc0253b78a3fcf4fe), fb(0xc04fccdf12cc6558)},
		{fb(0xbfe3860dc6b56206), fb(0xc04fde78b06beaa6), fb(0x403e038837e6f21e), fb(0x40340f06290e7607), fb(0xc0246b28c6c71d43), fb(0xc0507837997a3812), fb(0xc0262c38685489a8), fb(0xc05090ff2b5e0b43)},
		{fb(0x40483e6da5f90708), fb(0xbfecd2bbedbde798), fb(0x403de9c09e0d55fd), fb(0x403844a101d3d7c1), fb(0x404bdec2d7ef0c42), fb(0xc01fa1d27cedc730), fb(0x404bd95cc7b69e7a), fb(0xc020597477e39831)},
	}
	for _, tt := range tests {
		fixe := NewPositionableImpl(tt.x, tt.y, NewXDimension2D(tt.w, tt.h))
		toMove := NewPositionableImpl(tt.mx, tt.my, NewXDimension2D(10.5, 7.25))
		got := PositionableUtilsMoveAwayFrom(fixe, toMove)
		samePoint(t, "PositionableUtilsMoveAwayFrom", got.Position(), XPoint2D{X: tt.wantX, Y: tt.wantY})
		if got.Size() != toMove.Size() {
			t.Errorf("size changed to %v", got.Size())
		}
		if toMove.Position() != NewXPoint2D(tt.mx, tt.my) {
			t.Error("toMove was moved")
		}
	}
	mustPanic(t, "PositionableUtilsMoveAwayFrom apart", func() {
		PositionableUtilsMoveAwayFrom(NewPositionableImpl(0, 0, NewXDimension2D(1, 1)),
			NewPositionableImpl(5, 5, NewXDimension2D(1, 1)))
	})
}

func TestPositionableUtilsAddMargin(t *testing.T) {
	inner := NewPositionableImpl(fb(0xc03e1f9adbb8f8da), fb(0xbfe4b367a0f9096c), NewXDimension2D(3.3, 4.4))
	m := PositionableUtilsAddMargin(inner, 0.7, 1.1)
	samePoint(t, "Position", m.Position(), XPoint2D{X: fb(0xc03ed2ce0eec2c0d), Y: fb(0xbffbf34d6a161e50)})
	sameBits(t, "Width", m.Size().Width(), fb(0x4012cccccccccccc))
	sameBits(t, "Height", m.Size().Height(), fb(0x401a666666666667))
	m.MoveDelta(1, 2)
	samePoint(t, "moved inner", inner.Position(), XPoint2D{X: fb(0xc03e1f9adbb8f8da) + 1, Y: fb(0xbfe4b367a0f9096c) + 2})

	a := NewPositionableImpl(0, 0, NewXDimension2D(10, 10))
	if PositionableUtilsIntersect(a, NewPositionableImpl(10, 0, NewXDimension2D(1, 1))) {
		t.Error("touching items must not intersect")
	}
	if !PositionableUtilsIntersect(a, PositionableUtilsAddMargin(NewPositionableImpl(10, 0, NewXDimension2D(1, 1)), 0.5, 0)) {
		t.Error("the margin must make items intersect")
	}
}
