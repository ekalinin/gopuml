// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import (
	"math"
	"testing"
)

var curveTests = []struct {
	c, left, right               XCubicCurve2D
	flatnessSq, flatness, length float64
	startingAngle, endingAngle   float64
}{
	{cv(fb(0x402d2a03487e9180), fb(0x40548d0e164e06ba), fb(0x4053e16734630a8a), fb(0x404072c4fed1ebd4), fb(0x4054c356a62ccfee), fb(0xc03b2bd9b2b04140), fb(0xbfb27555869b245c), fb(0xc04d5306a36c0590)), cv(fb(0x402d2a03487e9180), fb(0x40548d0e164e06ba), fb(0x404786a79d72dcba), fb(0x404cc67095b6fca4), fb(0x40500ad95e00adcc), fb(0x403e34dca873e23e), fb(0x404f63b728d76eb3), fb(0x40145d9df99d6980)), cv(fb(0x404f63b728d76eb3), fb(0x40145d9df99d6980), fb(0x404eb1bb95ad81ce), fb(0xc034060daba52d7e), fb(0x4044beb950cb2925), fb(0xc0457479be621318), fb(0xbfb27555869b245c), fb(0xc04d5306a36c0590)), fb(0x40b8a3a6c1b97001), fb(0x4053daed242a976d), fb(0x4061b39cabb43e89), fb(0xbfe4c7fbabde5621), fb(0xc0063c9f868e6ead)},
	{cv(fb(0xc03ef55daef01e7c), fb(0xc029b1e7a3756618), fb(0xc031f565c92ed078), fb(0x3fd5e3d12971e24d), fb(0xbfe3860dc6b56206), fb(0xc04fde78b06beaa6), fb(0x3fc1a91782ba9800), fb(0x3fe2c7b3521388c5)), cv(fb(0xc03ef55daef01e7c), fb(0xc029b1e7a3756618), fb(0xc0387561bc0f777a), fb(0xc01902c91a29d706), fb(0xc030df166be0da9f), fb(0xc032f9b1aa51be51), fb(0xc025a1b3898606f9), fb(0xc0394e5b91780802)), cv(fb(0xc025a1b3898606f9), fb(0xc0394e5b91780802), fb(0xc0130a747694b168), fb(0xc03fa305789e51b2), fb(0xbfce378fcc0d780c), fb(0xc03f9359e3239c83), fb(0x3fc1a91782ba9800), fb(0x3fe2c7b3521388c5)), fb(0x40aaf7e26073d235), fb(0x404d6069ded20107), fb(0x4040efe9e002934d), fb(0x3fe95d4b77cada11), fb(0x3ff8f25923b70dc7)},
	{cv(fb(0xc057ff79e4f2b8eb), fb(0xc035674232aa6dd8), fb(0x40483e6da5f90708), fb(0xbfecd2bbedbde798), fb(0xbfebcf3a6f548444), fb(0x40455725092336c4), fb(0x405221a9f9ce1a24), fb(0xc05184acdf81c666)), cv(fb(0xc057ff79e4f2b8eb), fb(0xc035674232aa6dd8), fb(0xc037c08623ec6ace), fb(0xc0264dd812185d15), fb(0x3f9d55309e945200), fb(0x401379dc20c02137), fb(0x402dd8f92f34f0aa), fb(0x4010ee9373ebf9ba)), cv(fb(0x402dd8f92f34f0aa), fb(0x4010ee9373ebf9ba), fb(0x403dd1a3e30d4b96), fb(0x400cc6958e2fa478), fb(0x4041ea0b84ef711b), fb(0xc02b64696bc0ac10), fb(0x405221a9f9ce1a24), fb(0xc05184acdf81c666)), fb(0x40be393d5a21dcc4), fb(0x4055fd8a893c4b71), fb(0x4065ecf79be322de), fb(0x3fc20b292a6dcbee), fb(0xbfefccef805e9566)},
	{cv(fb(0xc03d7b06519696bc), fb(0x3fdcfe6843c464f1), fb(0xbfb45dfa461c86ec), fb(0xc037931d4c480be0), fb(0x3fed9a11f2080dfb), fb(0x40313f2c99dffc7c), fb(0x405597919330f6dc), fb(0x40399c20abbe2a40)), cv(fb(0xc03d7b06519696bc), fb(0x3fdcfe6843c464f1), fb(0xc02d8f644bdcb343), fb(0xc0271f23ab38fa4c), fb(0xc01cb6f1b6928f5a), fb(0xc01d73145da109b0), fb(0x401db64bdd85d837), fb(0x3fec5120d65436a0)), cv(fb(0x401db64bdd85d837), fb(0x3fec5120d65436a0), fb(0x403608e25c678ff2), fb(0x402243ae499b0bac), fb(0x4045d2c5b71506f8), fb(0x40356da6a2cf135e), fb(0x405597919330f6dc), fb(0x40399c20abbe2a40)), fb(0x408b9a0323729f80), fb(0x403db8344bf915be), fb(0x405da31f7ca0cb5d), fb(0xbfe5ecdfff51645d), fb(0x3fb8fa3bc7a07eac)},
	{cv(fb(0x4054da9d3ad15e8c), fb(0xc04a495de512889c), fb(0x403f6907eb930708), fb(0xc053de01d6b1c4b0), fb(0xc04c99530bfc5748), fb(0x4047fce8ffb9dc3c), fb(0x4052d111806d9672), fb(0x40489a44c3ed58c0)), cv(fb(0x4054da9d3ad15e8c), fb(0xc04a495de512889c), fb(0x404cb4df35b6204e), fb(0xc0508158649d847f), fb(0x40364277aa9cb66c), fb(0xc044711f1007efc8), fb(0x4024547819f2b758), fb(0xc028763495111ba4)), cv(fb(0x4024547819f2b758), fb(0xc028763495111ba4), fb(0xbffedff90a9ff140), fb(0x40306c098afec3ec), fb(0x4022119fe9bdab38), fb(0x40484b96e1d39a7e), fb(0x4052d111806d9672), fb(0x40489a44c3ed58c0)), fb(0x40d1231767ef078f), fb(0x40608f0c3e8417d5), fb(0x405986a9eeef5821), fb(0xc005507eaf7ad56d), fb(0x3f83019e07a1d8ce)},
	{cv(fb(0x3fd4c2abd0ed8e61), fb(0x4043b55e71a17b60), fb(0x405871d0f3bce014), fb(0xc04948982c27b067), fb(0x4044cc0bd707f7bc), fb(0xc04ddc9b0ccbe618), fb(0x3fedca844cb4aab3), fb(0x3fa88d856e4fe9c3)), cv(fb(0x3fd4c2abd0ed8e61), fb(0x4043b55e71a17b60), fb(0x404886939f8dcda2), fb(0xc0164ce6ea18d41c), fb(0x404daf353f6754ca), fb(0xc03e5c3679bce5c4), fb(0x404a35f7148b3aae), fb(0xc0423682fafbb685)), cv(fb(0x404a35f7148b3aae), fb(0xc0423682fafbb685), fb(0x4046bcb8e9af2093), fb(0xc0453eeab918fa28), fb(0x40354335e83aca67), fb(0xc03dd677ab70521e), fb(0x3fedca844cb4aab3), fb(0x3fa88d856e4fe9c3)), fb(0x40c752a77c4d8a6c), fb(0x405b51aecdffdd71), fb(0x4043afd42eca5c6b), fb(0xbfe7dbae6dc95ce2), fb(0x400158718c253e80)},
	{cv(fb(0xc051d372ddbca2d6), fb(0xc052bc5b0f832591), fb(0xbfd474ffd65508bb), fb(0x40539ce3c1d5966a), fb(0x3fbdfa7d7b65e952), fb(0xbfd5d218d4ca3817), fb(0x405857fd8941a08a), fb(0xbfc2963440f5dd67)), cv(fb(0xc051d372ddbca2d6), fb(0xc052bc5b0f832591), fb(0xc041e7e7dd92f7df), fb(0x3ffc11164a4e1b20), fb(0xc031f4de3e0a736d), fb(0x4034679a5b533d0b), fb(0x4009769e287a2c24), fb(0x4033e7c768af620b)), cv(fb(0x4009769e287a2c24), fb(0x4033e7c768af620b), fb(0x40385285c828fe76), fb(0x403367f4760b870b), fb(0x40485f7c28a07a04), fb(0xbfcf1d32f54526ca), fb(0x405857fd8941a08a), fb(0xbfc2963440f5dd67)), fb(0x40c842b874fc05ff), fb(0x405bdce5af9f2d06), fb(0x4067109b0a8ea9b5), fb(0x3ff232bab35c87b5), fb(0x3f607cb0a49e50bc)},
	{cv(fb(0xc02bba3ff2e044c8), fb(0xc00fe9029ea8d700), fb(0xc0556caf0a8dee50), fb(0xc0505333edd3d058), fb(0x402585f99df38408), fb(0xbfff7a1517496700), fb(0xc04c72cca35fc5ba), fb(0x405388cb4d8adf78)), cv(fb(0xc02bba3ff2e044c8), fb(0xc00fe9029ea8d700), fb(0xc048e3f708e9f6e9), fb(0xc041527c02c91710), fb(0xc045cff36fdcba5c), fb(0xc04111cc227d0682), fb(0xc042791f751e9939), fb(0xc02fe9d18dfd491c)), cv(fb(0xc042791f751e9939), fb(0xc02fe9d18dfd491c), fb(0xc03e4496f4c0f02b), fb(0x4001ce35b7e61f40), fb(0xc037114e3be2e4b8), fb(0x40430ae2f92db9dc), fb(0xc04c72cca35fc5ba), fb(0x405388cb4d8adf78)), fb(0x40c16ba0a1bb7918), fb(0x40579c469173471a), fb(0x40572ded8bb7e453), fb(0xc0037b038b91468e), fb(0x40022d6372ba8e63)},
	{cv(fb(0x3fee1b380327657e), fb(0xc0538a9e777919ea), fb(0x3fe4bb308bea828d), fb(0xbfe3b796907d52ba), fb(0x4058e9b2666d3dfc), fb(0xc04f344cbcbab09f), fb(0x3fc5347b37396fae), fb(0x3feb7b1dde440edc)), cv(fb(0x3fee1b380327657e), fb(0xc0538a9e777919ea), fb(0x3fe96b344788f406), fb(0xc043b20da49a148f), fb(0x403978d598a336d1), fb(0xc041b9d1980c33c2), fb(0x4042e012c10c4921), fb(0xc040a61a378de2ab)), cv(fb(0x4042e012c10c4921), fb(0xc040a61a378de2ab), fb(0x404903bab5c6f6da), fb(0xc03f24c5ae1f2327), fb(0x4048f44ca408dab4), fb(0xc03ec6604541a064), fb(0x3fc5347b37396fae), fb(0x3feb7b1dde440edc)), fb(0x40c316b2ff80d66a), fb(0x4058b7126dc9af1d), fb(0x4053c1d2fcb0dc80), fb(0x3ff93174af98be93), fb(0x400499eed87bca65)},
	{cv(fb(0xc042f1298a4ce5ac), fb(0x4033c1e2d62b1b90), fb(0xc015a73fd39cfe30), fb(0xc058bd4c36806d3b), fb(0xbfe0e6e9efb52f18), fb(0xbfb127a2fe1076cd), fb(0x405822680c4472c4), fb(0x3fe60ab83a425e36)), cv(fb(0xc042f1298a4ce5ac), fb(0x4033c1e2d62b1b90), fb(0xc035a61184c08572), fb(0xc043ccd380f5a657), fb(0xc0289e9526f2f9f4), fb(0xc0464734d01acbd8), fb(0x40146a1ba7a4a256), fb(0xc0414a0d0df0221a)), cv(fb(0x40146a1ba7a4a256), fb(0xc0414a0d0df0221a), fb(0x40368458674bce25), fb(0xc03899ca978af0ba), fb(0x4048009a38650866), fb(0x3fd3e5c3da804f5c), fb(0x405822680c4472c4), fb(0x3fe60ab83a425e36)), fb(0x40c8ee0066af1c97), fb(0x405c3e95b6bf27f9), fb(0x4060f88ee1bf483f), fb(0xbff4dc660f19ca1d), fb(0x3f7fe4c10664dcc5)},
}

func TestXCubicCurve2D(t *testing.T) {
	for _, tt := range curveTests {
		left, right := tt.c.Subdivide()
		sameCurve(t, "Subdivide left", left, tt.left)
		sameCurve(t, "Subdivide right", right, tt.right)
		sameBits(t, "FlatnessSq", tt.c.FlatnessSq(), tt.flatnessSq)
		sameBits(t, "Flatness", tt.c.Flatness(), tt.flatness)
		sameBits(t, "Length", tt.c.Length(), tt.length)
	}
}

func TestXCubicCurve2DPoints(t *testing.T) {
	c := NewXCubicCurve2D(1, 2, 3, 4, 5, 6, 7, 8)
	if c.P1() != NewXPoint2D(1, 2) || c.CtrlP1() != NewXPoint2D(3, 4) || c.CtrlP2() != NewXPoint2D(5, 6) ||
		c.P2() != NewXPoint2D(7, 8) {
		t.Errorf("points of %v", c)
	}
	if XCubicCurve2DNone() != (XCubicCurve2D{}) {
		t.Error("XCubicCurve2DNone() is not the zero value")
	}
	// Upstream's setCurve(XCubicCurve2D) copies the second control point
	// into the end point.
	if got := setCurveFrom(c); got != NewXCubicCurve2D(1, 2, 3, 4, 5, 6, 5, 6) {
		t.Errorf("setCurveFrom = %v", got)
	}
	straight := NewXCubicCurve2D(0, 0, 1, 0, 2, 0, 3, 0)
	sameBits(t, "straight Flatness", straight.Flatness(), 0)
	sameBits(t, "Length", NewXCubicCurve2D(0, 0, 9, 9, 9, 9, 3, 4).Length(), 5)
	if math.IsNaN(straight.FlatnessSq()) {
		t.Error("FlatnessSq is NaN")
	}
}
