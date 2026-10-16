// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

func TestXPoint2DArithmetic(t *testing.T) {
	tests := []struct {
		x, y, ox, oy                         float64
		distance, distanceSq, static, distXY float64
		hash                                 int32
		str                                  string
	}{
		{fb(0x402d2a03487e9180), fb(0x40548d0e164e06ba), fb(0x4053e16734630a8a), fb(0x404072c4fed1ebd4), fb(0x4054626a2060e093), fb(0x40b9f866a7bec174), fb(0x4054626a2060e093), fb(0x4054626a2060e093), 1584285495, "(14.582056298685302,82.20398481006387)"},
		{fb(0x4054c356a62ccfee), fb(0xc03b2bd9b2b04140), fb(0xbfb27555869b245c), fb(0xc04d5306a36c0590), fb(0x4056389cd5b7fe67), fb(0x40bedc779b93563b), fb(0x4056389cd5b7fe67), fb(0x4056389cd5b7fe67), 1493399377, "(83.05216364265627,-27.17129055789087)"},
		{fb(0xc03ef55daef01e7c), fb(0xc029b1e7a3756618), fb(0xc031f565c92ed078), fb(0x3fd5e3d12971e24d), fb(0x403284e86760cc77), fb(0x40756f5af07341cd), fb(0x403284e86760cc77), fb(0x403284e86760cc77), -768883936, "(-30.958460744498083,-12.847470386582287)"},
		{fb(0xbfe3860dc6b56206), fb(0xc04fde78b06beaa6), fb(0x3fc1a91782ba9800), fb(0x3fe2c7b3521388c5), fb(0x4050151303ae6ebf), fb(0x40b02a41c95d201f), fb(0x4050151303ae6ebf), fb(0x4050151303ae6ebf), -377808663, "(-0.6101139908815305,-63.73805814046254)"},
		{fb(0xc057ff79e4f2b8eb), fb(0xc035674232aa6dd8), fb(0x40483e6da5f90708), fb(0xbfecd2bbedbde798), fb(0x40623daa1e8b903f), fb(0x40d4cbac6cf3b60a), fb(0x40623daa1e8b903f), fb(0x40623daa1e8b903f), 390353452, "(-95.99181483939508,-21.403353849991475)"},
		{fb(0xbfebcf3a6f548444), fb(0x40455725092336c4), fb(0x405221a9f9ce1a24), fb(0xc05184acdf81c666), fb(0x4060d13108aa19ae), fb(0x40d1ad1122ecd451), fb(0x4060d13108aa19ae), fb(0x4060d13108aa19ae), 438676831, "(-0.869046418611227,42.680817739671994)"},
		{fb(0xc03d7b06519696bc), fb(0x3fdcfe6843c464f1), fb(0xbfb45dfa461c86ec), fb(0xc037931d4c480be0), fb(0x4042fc352e7bc820), fb(0x409686ff347f81e4), fb(0x4042fc352e7bc820), fb(0x4042fc352e7bc820), 230983763, "(-29.48056516577229,0.45302778833281737)"},
		{fb(0x3fed9a11f2080dfb), fb(0x40313f2c99dffc7c), fb(0x405597919330f6dc), fb(0x40399c20abbe2a40), fb(0x4055767f319874ed), fb(0x40bcca7b7a3edf98), fb(0x4055767f319874ed), fb(0x4055767f319874ed), -1479255238, "(0.9250573851754206,17.246774308380267)"},
		{fb(0x4054da9d3ad15e8c), fb(0xc04a495de512889c), fb(0x403f6907eb930708), fb(0xc053de01d6b1c4b0), fb(0x404d463db5590530), fb(0x40aac7ea0721be44), fb(0x404d463db5590530), fb(0x404d463db5590530), -1612823086, "(83.41584654279922,-52.573177942323326)"},
		{fb(0xc04c99530bfc5748), fb(0x4047fce8ffb9dc3c), fb(0x4052d111806d9672), fb(0x40489a44c3ed58c0), fb(0x40608f0c3e8417d5), fb(0x40d1231767ef078f), fb(0x40608f0c3e8417d5), fb(0x40608f0c3e8417d5), -1951469841, "(-57.19784688779197,47.97586056304178)"},
		{fb(0x3fd4c2abd0ed8e61), fb(0x4043b55e71a17b60), fb(0x405871d0f3bce014), fb(0xc04948982c27b067), fb(0x4060949b3e95d23b), fb(0x40d12e9abb13c5b3), fb(0x4060949b3e95d23b), fb(0x4060949b3e95d23b), 555490056, "(0.3243817844688462,39.416944698189354)"},
		{fb(0x4044cc0bd707f7bc), fb(0xc04ddc9b0ccbe618), fb(0x3fedca844cb4aab3), fb(0x3fa88d856e4fe9c3), fb(0x405212ae1fa335a5), fb(0x40b46a1d965d5849), fb(0x405212ae1fa335a5), fb(0x405212ae1fa335a5), 1674147386, "(41.594111327066884,-59.72348175006863)"},
	}
	for _, tt := range tests {
		p, o := NewXPoint2D(tt.x, tt.y), NewXPoint2D(tt.ox, tt.oy)
		sameBits(t, "Distance", p.Distance(o), tt.distance)
		sameBits(t, "DistanceSq", p.DistanceSq(o), tt.distanceSq)
		sameBits(t, "XPoint2DDistance", XPoint2DDistance(tt.x, tt.y, tt.ox, tt.oy), tt.static)
		sameBits(t, "DistanceXY", p.DistanceXY(tt.ox, tt.oy), tt.distXY)
		if got := p.HashCode(); got != tt.hash {
			t.Errorf("%v.HashCode() = %d, want %d", p, got, tt.hash)
		}
		if got := p.String(); got != tt.str {
			t.Errorf("String() = %q, want %q", got, tt.str)
		}
		samePoint(t, "Move", p.Move(tt.ox, tt.oy), XPoint2D{X: tt.x + tt.ox, Y: tt.y + tt.oy})
		samePoint(t, "MovePoint", p.MovePoint(o), p.Move(tt.ox, tt.oy))
	}
}

func TestXPoint2DHashCodeAndEquals(t *testing.T) {
	tests := []struct {
		p    XPoint2D
		str  string
		hash int32
	}{
		{NewXPoint2D(0, 0), "(0.0,0.0)", 0},
		{NewXPoint2D(math.Copysign(0, -1), 0), "(-0.0,0.0)", -2147483648},
		{NewXPoint2D(math.NaN(), 1), "(NaN,1.0)", -1075314688},
		{NewXPoint2D(1e300, -1e-300), "(1.0E300,-1.0E-300)", 966078182},
		{NewXPoint2D(math.Inf(1), 3), "(Infinity,3.0)", -1074266112},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.str {
			t.Errorf("String() = %q, want %q", got, tt.str)
		}
		if got := tt.p.HashCode(); got != tt.hash {
			t.Errorf("%v.HashCode() = %d, want %d", tt.p, got, tt.hash)
		}
	}
	if !NewXPoint2D(math.Copysign(0, -1), 1).Equals(NewXPoint2D(0, 1)) {
		t.Error("-0.0 must equal 0.0")
	}
	if nan := NewXPoint2D(math.NaN(), 1); nan.Equals(nan) {
		t.Error("NaN must not equal itself")
	}
	if NewXPoint2D(1, 2).Equals(NewXPoint2D(2, 1)) {
		t.Error("(1,2) must not equal (2,1)")
	}
}
