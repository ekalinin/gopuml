// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestBezierUtilsAngles(t *testing.T) {
	for _, tt := range curveTests {
		c := tt.c
		nearAngle(t, "BezierUtilsStartingAngle", BezierUtilsStartingAngle(c), tt.startingAngle)
		nearAngle(t, "BezierUtilsEndingAngle", BezierUtilsEndingAngle(c), tt.endingAngle)
	}
	// With a control point on its end point, the chord gives the angle.
	c := NewXCubicCurve2D(0, 0, 0, 0, 10, 10, 10, 10)
	sameBits(t, "degenerate StartingAngle", BezierUtilsStartingAngle(c), math.Pi/4)
	sameBits(t, "degenerate EndingAngle", BezierUtilsEndingAngle(c), math.Pi/4)
	sameBits(t, "BezierUtilsAngle", BezierUtilsAngle(NewXPoint2D(1, 1), NewXPoint2D(1, 5)), math.Pi/2)
	mustPanic(t, "BezierUtilsAngle of equal points", func() { BezierUtilsAngle(NewXPoint2D(1, 1), NewXPoint2D(1, 1)) })
}

func TestBezierUtilsSubdivide(t *testing.T) {
	tests := []struct {
		c           XCubicCurve2D
		coef        float64
		left, right XCubicCurve2D
	}{
		{cv(fb(0x3fc003ebbd129329), fb(0xbfe443c8721e5bec), fb(0xbfed9d97e969e59a), fb(0xbfdf27d9c7521268), fb(0x403f37c679d689e8), fb(0x3fe931acf4b8acee), fb(0xc05737e2010316a1), fb(0xc003af613b787660)), fb(0x3fe0000000000000), cv(fb(0x3fc003ebbd129329), fb(0xbfe443c8721e5bec), fb(0xbfd99c9cfa2540d0), fb(0xbfe1ebdaaae3b290), fb(0x401d7df4d2ba10b4), fb(0xbfca39f544b7c166), fb(0xbfcdef2fcf0852d0), fb(0xbfd18a034b345290)), cv(fb(0xbfcdef2fcf0852d0), fb(0xbfd18a034b345290), fb(0xc01f5ce7cfaa95e1), fb(0xbfd5f70bf40cc46c), fb(0xc03ed3e0c51ae84e), fb(0xbfeac5ebfc949649), fb(0xc05737e2010316a1), fb(0xc003af613b787660))},
		{cv(fb(0x404d11746055bb00), fb(0x40379305e975eab4), fb(0xc051595ae8164ed2), fb(0x3fe4726a69291436), fb(0xc03f09af772fe948), fb(0x404145e9530ba284), fb(0xc0537f91c26c9f80), fb(0x4058ac9ce92c1664)), fb(0x3fd3333333333333), cv(fb(0x404d11746055bb00), fb(0x40379305e975eab4), fb(0xc00b0606e5a10ca8), fb(0x401d0e517c18a401), fb(0xc02b8dd0c91a80cc), fb(0x4022fb98d052fafc), fb(0xc025cfe1361a6c8e), fb(0x402146fd18e36516)), cv(fb(0xc025cfe1361a6c8e), fb(0x402146fd18e36516), fb(0xc036938ef59ec9dc), fb(0x40334dd96bfc2afc), fb(0xc0405acb6021f5e4), fb(0x4043fc8a8b378afd), fb(0xc0537f91c26c9f80), fb(0x4058ac9ce92c1664))},
		{cv(fb(0xbfe4bf89d2dd165a), fb(0xc039aa62e1ea09bc), fb(0xc052d39bbd686840), fb(0xc0527fddfbdec4b8), fb(0x4048d661a5be8ae4), fb(0xc03b9237431bf080), fb(0x40578074a7db5a7c), fb(0xbfdbded43a8f9ac4)), fb(0x3fe6b851eb851eb8), cv(fb(0xbfe4bf89d2dd165a), fb(0xc039aa62e1ea09bc), fb(0xc04af6c4ccb2ca7b), fb(0xc051b0b58f6858eb), fb(0xc0508173bd8600be), fb(0xc0574e0e4a8d5045), fb(0xc02e489bec235154), fb(0xc05aae67c5bba002)), cv(fb(0xc02e489bec235154), fb(0xc05aae67c5bba002), fb(0x4046591cbf561aee), fb(0xc04c8c7782419cc6), fb(0x405980e8b0bcd2d4), fb(0xc033e27c3b0463df), fb(0x40578074a7db5a7c), fb(0xbfdbded43a8f9ac4))},
	}
	for _, tt := range tests {
		left, right := bezierUtilsSubdivide(tt.c, tt.coef)
		sameCurve(t, "bezierUtilsSubdivide left", left, tt.left)
		sameCurve(t, "bezierUtilsSubdivide right", right, tt.right)
	}
}

func TestBezierUtilsShorten(t *testing.T) {
	shape := NewXRectangle2D(0, 0, 50, 40)
	tests := []struct {
		c, want XCubicCurve2D
		panics  bool
	}{
		{cv(fb(0x40391ecc0fd50df5), fb(0x401cd2538103ce14), fb(0x4043bfa000cc3a28), fb(0x404339dbfb1cba88), fb(0x4051a1c918959c6c), fb(0x404dd90c011595dd), fb(0x405435119de255dd), fb(0x4051bdd0fa74fe60)), cv(fb(0x40391ecc0fd50df5), fb(0x401cd2538103ce14), fb(0x40402783045b6091), fb(0x4036d4266b3d344a), fb(0x4045d48e0eab8d08), fb(0x4041f9c399dbe12c), fb(0x4045d48e0eab8d08), fb(0x4041f9c399dbe12c)), false},
		{cv(fb(0x40540d303fda9c30), fb(0x4051bbe10b75994c), fb(0x4051881ac102e061), fb(0x404e17990e5e40b1), fb(0xc0474ab7481f5d8e), fb(0xc052e19d427ca25b), fb(0x402dffd4584bd0ff), fb(0x4033f449e2c54634)), cv(fb(0x4034bab410ded58e), fb(0x40166f02a47cc9a4), fb(0x4045bc050ee85715), fb(0x403cf1dcd75737a2), fb(0x4052caa5806ebe48), fb(0x405063d6c9525cd2), fb(0x4052caa5806ebe48), fb(0x405063d6c9525cd2)), false},
		{cv(fb(0x404497d456df1661), fb(0x401233673c4e5e72), fb(0x403db59c3470c393), fb(0x4051664355487af2), fb(0x40208da85b4835ac), fb(0x404dc82b616b4df1), fb(0x4061e5cf3023e7bb), fb(0x4051883ab51dfa28)), cv(fb(0x404497d456df1661), fb(0x401233673c4e5e72), fb(0x4041b951388bbc15), fb(0x40428979c90d60d9), fb(0x403b386d5110f3af), fb(0x404969e96785c162), fb(0x403b386d5110f3af), fb(0x404969e96785c162)), false},
		{cv(fb(0x405a55d3503c6e24), fb(0x406297c73ac95926), fb(0x4054823bfa333e67), fb(0x405ec1639105b036), fb(0x405848425f83f170), fb(0x4058eb299e994760), fb(0x4039236c9681ec7e), fb(0x401a2f568f2bfb32)), cv(fb(0x4039236c9681ec7e), fb(0x401a2f568f2bfb32), fb(0x404e911d85246c90), fb(0x404a8e1f078c0713), fb(0x4052d6e6f7b6e71a), fb(0x40548eab0dcabfaa), fb(0x4052d6e6f7b6e71a), fb(0x40548eab0dcabfaa)), false},
	}
	for _, tt := range tests {
		if tt.panics {
			mustPanic(t, "bezierUtilsShorten", func() { bezierUtilsShorten(tt.c, shape) })
			continue
		}
		sameCurve(t, "bezierUtilsShorten", bezierUtilsShorten(tt.c, shape), tt.want)
	}
	mustPanic(t, "bezierUtilsShorten inside", func() { bezierUtilsShorten(NewXCubicCurve2D(1, 1, 2, 2, 3, 3, 4, 4), shape) })
}

func TestBezierUtilsIntersect(t *testing.T) {
	tests := []struct{ x, y, w, h, x1, y1, x2, y2, px, py float64 }{
		{fb(0xbfde2bad867ace31), fb(0x4051d61d1ddf4260), fb(0x40561753d433ff86), fb(0x40584e9d25713f8c), fb(0x4045dafc792709ea), fb(0x405dfd6bb097e226), fb(0x40736fc132f05ad6), fb(0x4079a6cf37e1bab6), fb(0x4055fcece286bdf8), fb(0x40650145985e599c)},
		{fb(0x404f2b05f7aff0dc), fb(0x3fe654e736c90373), fb(0x4049facfb48fa92d), fb(0x4040fe15cb968c5a), fb(0x4075f9ac6a72b840), fb(0xc06f8c341dcb264d), fb(0x40561436e8fbe2b9), fb(0x4031b0bd054cd476), fb(0x405a3998e460ec7e), fb(0x3fe5ba9b4a7fe8f4)},
		{fb(0xbfd950f4fe185c78), fb(0x405016a6ccd1c99c), fb(0x403824bd55885f84), fb(0x405900790b624389), fb(0x40275a35ad979ca0), fb(0x405c96e35282eb60), fb(0x407373563985be80), fb(0x40431ddab94e5bbe), fb(0x4037d0c0c9a4055b), fb(0x405bd170baac2392)},
		{fb(0xc04b2644e00654ab), fb(0xc043f5d41587033e), fb(0x4057271a9d70437c), fb(0x40507983b7b022ac), fb(0x407027cc69490a84), fb(0x4070114fedc3e924), fb(0xc01ff95214b08978), fb(0xc01be282eeb70490), fb(0x4039510ede0b76f5), fb(0x403a0913374fc949)},
		{fb(0xbfe54772aa905e80), fb(0x404fdea3f057160c), fb(0x4053746789ee607c), fb(0x3ffa937316e5bd76), fb(0x40431f49bf441f02), fb(0x40502478de595681), fb(0x407528ef559d9cff), fb(0x40443d93747d3522), fb(0x404853ee718158fa), fb(0x404fde0cd7f1d27b)},
		{fb(0x3fee20e88c3e3d3b), fb(0x3fc2df7386b3fab8), fb(0x40434717a44adb60), fb(0x3ff1c9d583b5633e), fb(0x40740c50f59729d0), fb(0xbfdbc60ca0253508), fb(0x4034381ee8accd4a), fb(0x3fe681b2656261ec), fb(0x4043c235cc754068), fb(0x3fe42b7ac93baef6)},
		{fb(0x4049ae3c7e53d234), fb(0xc028d54cd539ef70), fb(0x3ff55420c39e2d63), fb(0x401d0fb2be56a5e0), fb(0x404a038d01624aea), fb(0xc021916025a445f8), fb(0x407ad45864cb152c), fb(0xc070c9390f024ba8), fb(0x404a5bf7a2c39e66), fb(0xc02284ee5873c366)},
		{fb(0xc058e42f17c596b2), fb(0xc0562507935317d3), fb(0x40480e1d92823789), fb(0x40277c788f8af418), fb(0x4071302d59bc53ed), fb(0xc0719989154a1f63), fb(0xc052e0a7b32508d0), fb(0xc054ad400a5a6892), fb(0xc05049f3630fbfd2), fb(0xc056254648d01f11)},
	}
	for _, tt := range tests {
		got := BezierUtilsIntersect(NewXLine2D(tt.x1, tt.y1, tt.x2, tt.y2), NewXRectangle2D(tt.x, tt.y, tt.w, tt.h))
		samePoint(t, "BezierUtilsIntersect", got, XPoint2D{X: tt.px, Y: tt.py})
	}
	mustPanic(t, "BezierUtilsIntersect outside", func() {
		BezierUtilsIntersect(NewXLine2D(20, 20, 30, 30), NewXRectangle2D(0, 0, 10, 10))
	})
}

func TestBezierUtilsHelpers(t *testing.T) {
	if got := BezierUtilsMiddle(NewXLine2D(1, 2, 4, 7)); got != NewXPoint2D(2.5, 4.5) {
		t.Errorf("BezierUtilsMiddle = %v", got)
	}
	if got := BezierUtilsMiddlePoints(NewXPoint2D(1, 2), NewXPoint2D(4, 7)); got != NewXPoint2D(2.5, 4.5) {
		t.Errorf("BezierUtilsMiddlePoints = %v", got)
	}
	sameBits(t, "bezierUtilsDistLine", bezierUtilsDistLine(NewXLine2D(0, 0, 3, 4)), 5)
	sameBits(t, "bezierUtilsDistCurve", bezierUtilsDistCurve(NewXCubicCurve2D(0, 0, 7, 7, 8, 8, 3, 4)), 5)

	a := NewPositionableImpl(0, 0, NewXDimension2D(10, 10))
	b := NewPositionableImpl(5, 5, NewXDimension2D(10, 10))
	c := NewPositionableImpl(10, 0, NewXDimension2D(10, 10))
	if !BezierUtilsIntersectPositionables(a, b) || BezierUtilsIntersectPositionables(a, c) {
		t.Error("BezierUtilsIntersectPositionables")
	}
	if got := BezierUtilsCenter(b); got != NewXPoint2D(10, 10) {
		t.Errorf("BezierUtilsCenter = %v", got)
	}
}
