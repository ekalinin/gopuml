// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import "testing"

func TestXRectangle2DBounds(t *testing.T) {
	r := NewXRectangle2D(1, 2, 10, 20)
	for _, tt := range []struct {
		name      string
		got, want float64
	}{
		{"CenterX", r.CenterX(), 6},
		{"CenterY", r.CenterY(), 12},
		{"MinX", r.MinX(), 1},
		{"MaxX", r.MaxX(), 11},
		{"MinY", r.MinY(), 2},
		{"MaxY", r.MaxY(), 22},
	} {
		sameBits(t, tt.name, tt.got, tt.want)
	}
}

func TestXRectangle2DIntersects(t *testing.T) {
	r := NewXRectangle2D(0, 0, 10, 10)
	tests := []struct {
		other XRectangle2D
		want  bool
	}{
		{NewXRectangle2D(5, 5, 10, 10), true},
		{NewXRectangle2D(2, 2, 1, 1), true},
		{NewXRectangle2D(-5, -5, 30, 30), true},
		{NewXRectangle2D(10, 0, 5, 5), false}, // touching sides
		{NewXRectangle2D(0, 10, 5, 5), false},
		{NewXRectangle2D(-5, 0, 5, 5), false},
		{NewXRectangle2D(20, 20, 5, 5), false},
		{NewXRectangle2D(2, 2, 0, 5), false}, // empty
		{NewXRectangle2D(2, 2, 5, -1), false},
	}
	for _, tt := range tests {
		if got := r.Intersects(tt.other); got != tt.want {
			t.Errorf("Intersects(%v) = %v, want %v", tt.other, got, tt.want)
		}
	}
	if NewXRectangle2D(0, 0, 0, 10).Intersects(r) {
		t.Error("an empty rectangle must not intersect")
	}
}

func TestXRectangle2DContains(t *testing.T) {
	r := NewXRectangle2D(0, 0, 10, 10)
	tests := []struct {
		x, y float64
		want bool
	}{
		{0, 0, true},
		{5, 9.999, true},
		{10, 5, false},
		{5, 10, false},
		{-0.001, 5, false},
	}
	for _, tt := range tests {
		if got := r.Contains(tt.x, tt.y); got != tt.want {
			t.Errorf("Contains(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
	mustPanic(t, "Contains on an empty rectangle", func() { NewXRectangle2D(0, 0, 0, 10).Contains(0, 0) })
	mustPanic(t, "ContainsPoint", func() { r.ContainsPoint(NewXPoint2D(1, 1)) })
}

func TestXRectangle2DIntersect(t *testing.T) {
	tests := []struct {
		x, y, w, h, x1, y1, x2, y2 float64
		ok                         bool
		px, py                     float64
	}{
		{fb(0xc019d6a1bbc80c30), fb(0x3f913b0e68e67000), fb(0x404a3a74b000c19a), fb(0x4057e6ba99f8328d), fb(0x4033b0804ba711e4), fb(0x4046505154833ae2), fb(0xc05d9c8412a25749), fb(0x406a1412deeae189), true, fb(0xc019d6a1bbc80c2c), fb(0x4052eb0e9cbd82b9)},
		{fb(0x3fe7691f6a2ceb73), fb(0xc0570ff3f83718c7), fb(0x40250bee5ad748f0), fb(0x4050861048b21a70), fb(0x40183160c62e1141), fb(0xc04e41fe9f512fae), fb(0xbff3f45bfa4ad725), fb(0xbff2276e1dcc2f3c), true, fb(0x3ffd3a3b23721e68), fb(0xc03a278ebe13f95c)},
		{fb(0xbfe6b45791666e08), fb(0xbfeb6ae63741343c), fb(0x4052a8ea4709827e), fb(0x404fd95460ecb870), fb(0x403e16f67300b017), fb(0x4038fcc98cae428c), fb(0x403f113368657cc4), fb(0xc0705626bcdbaef6), true, fb(0x403e2d8bc4b6c626), fb(0xbfeb6ae637413440)},
		{fb(0x3fc6b50073e270b8), fb(0xbfe19b989302da9a), fb(0x404a9168e8df7288), fb(0x4050bed5f60d16a4), fb(0x402aec3ceaaefc4b), fb(0x403031f93174ffcf), fb(0x403abed2e9c73769), fb(0x4048d7d28ec7968c), false, 0, 0},
		{fb(0xc0408c43ab159ba0), fb(0xc026b297c1534cf8), fb(0x40474aa5fb4eae38), fb(0x40568c4ada872418), fb(0xc02bbcb2424dfb18), fb(0x4040ec6396f76a31), fb(0x40031ab8b7f07861), fb(0x400124b295749a40), false, 0, 0},
		{fb(0xc01ae2cfad51a500), fb(0xc054f05ce1892633), fb(0x403cc176aca501e0), fb(0x403b98fdcbe09d2c), fb(0x402d58f9a2265912), fb(0xc05265d8eeb95b0c), fb(0x406230d36592bd82), fb(0xc049cd358ca00eba), true, fb(0x403608c2c15098a0), fb(0xc05216b097c3cf73)},
		{fb(0xbfea34acab698344), fb(0xbfd3db3231eee1ec), fb(0x3ffa38432f22f7c0), fb(0x404ea00ec5a2be5f), fb(0x40027f64973612f9), fb(0x403e59b54c594114), fb(0xbfc4796af93c6cfb), fb(0x3fe825d6a30a13f4), true, fb(0x3fea3bd9b2dc6c3c), fb(0x4028f7cb1e712014)},
		{fb(0xc042c6debf726c0c), fb(0xbfe2ce3538ba8e60), fb(0x3ff20b5787372e8c), fb(0x3ff9adf93b64fddc), fb(0xc042a2c81063fdaf), fb(0xbfc7dce26c203dc8), fb(0xc0427eb161558f52), fb(0x3fe3b6c0a05cee6a), false, 0, 0},
		{fb(0x3fe88f1bfb07dda4), fb(0x3fd040c1dd986600), fb(0x3ff74b58d0261ef5), fb(0x4056f849070a8744), fb(0x400fd06966d49880), fb(0x40428c73b0be7400), fb(0x404745b0f483d4c6), fb(0xbff436c896023b5e), false, 0, 0},
		{fb(0x4058fc6399780ed8), fb(0xbfd048f0476b5597), fb(0x4033506edb30bd9c), fb(0x4054cac379a2531a), fb(0x405b644478b99e42), fb(0x4045f2ef4178ab2f), fb(0x405c3ec2eac0cd60), fb(0x406f77d3b4c9d47b), true, fb(0x405b8d482b78fc9c), fb(0x4054ba7a895ae7c4)},
	}
	for _, tt := range tests {
		r, l := NewXRectangle2D(tt.x, tt.y, tt.w, tt.h), NewXLine2D(tt.x1, tt.y1, tt.x2, tt.y2)
		got, ok := r.Intersect(l)
		if ok != tt.ok {
			t.Errorf("%v.Intersect(%v) ok = %v, want %v", r, l, ok, tt.ok)
			continue
		}
		if ok {
			samePoint(t, "Intersect", got, XPoint2D{X: tt.px, Y: tt.py})
		}
	}
	// The top side comes first.
	got, ok := NewXRectangle2D(0, 0, 10, 10).Intersect(NewXLine2D(5, -5, 5, 15))
	if !ok || got != NewXPoint2D(5, 0) {
		t.Errorf("vertical line = %v, %v", got, ok)
	}
}
