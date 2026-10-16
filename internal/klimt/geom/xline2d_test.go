// SPDX-License-Identifier: MIT
// Provenance: original

package geom

import "testing"

func TestXLine2DPtSegDistSq(t *testing.T) {
	tests := []struct{ x1, y1, x2, y2, px, py, want float64 }{
		{fb(0x402d2a03487e9180), fb(0x40548d0e164e06ba), fb(0x4053e16734630a8a), fb(0x404072c4fed1ebd4), fb(0x4054c356a62ccfee), fb(0xc03b2bd9b2b04140), fb(0x40ac493caad69c2f)},
		{fb(0xbfb27555869b245c), fb(0xc04d5306a36c0590), fb(0xc03ef55daef01e7c), fb(0xc029b1e7a3756618), fb(0xc02700073f22f8c2), fb(0xc044daa9d9d3a953), fb(0x3ee8672468000000)},
		{fb(0xc04fde78b06beaa6), fb(0x3fc1a91782ba9800), fb(0x3fe2c7b3521388c5), fb(0xc057ff79e4f2b8eb), fb(0xc035674232aa6dd8), fb(0x40483e6da5f90708), fb(0x40b021ece46bb612)},
		{fb(0xbfecd2bbedbde798), fb(0xbfebcf3a6f548444), fb(0x40455725092336c4), fb(0x405221a9f9ce1a24), fb(0xc05184acdf81c666), fb(0xc03d7b06519696bc), fb(0x40b5e36e08cd6c1b)},
		{fb(0x3fdcfe6843c464f1), fb(0xbfb45dfa461c86ec), fb(0xc037931d4c480be0), fb(0x3fed9a11f2080dfb), fb(0x40313f2c99dffc7c), fb(0x405597919330f6dc), fb(0x40be4b415f6e4169)},
		{fb(0x40399c20abbe2a40), fb(0x4054da9d3ad15e8c), fb(0xc04a495de512889c), fb(0x403f6907eb930708), fb(0xc00a8af7d60b7280), fb(0x405029d3ac1fdf3d), fb(0x3fc46cac88294000)},
		{fb(0x4052d111806d9672), fb(0x40489a44c3ed58c0), fb(0x3fd4c2abd0ed8e61), fb(0x4043b55e71a17b60), fb(0x405871d0f3bce014), fb(0xc04948982c27b067), fb(0x40c46ea690808a5b)},
		{fb(0x4044cc0bd707f7bc), fb(0xc04ddc9b0ccbe618), fb(0x3fedca844cb4aab3), fb(0x3fa88d856e4fe9c3), fb(0xc051d372ddbca2d6), fb(0xc052bc5b0f832591), fb(0x40c44861a1a20833)},
		{fb(0xbfd474ffd65508bb), fb(0x40539ce3c1d5966a), fb(0x3fbdfa7d7b65e952), fb(0xbfd5d218d4ca3817), fb(0x405857fd8941a08a), fb(0xbfc2963440f5dd67), fb(0x40c2798dc2d57550)},
		{fb(0xc02bba3ff2e044c8), fb(0xc00fe9029ea8d700), fb(0xc0556caf0a8dee50), fb(0xc0505333edd3d058), fb(0xc04438a3ec24abd3), fb(0xc03b3e2f8dd39379), fb(0x3fc7f919aef5c000)},
		{fb(0x405388cb4d8adf78), fb(0x3fee1b380327657e), fb(0xc0538a9e777919ea), fb(0x3fe4bb308bea828d), fb(0xbfe3b796907d52ba), fb(0x4058e9b2666d3dfc), fb(0x40c316771151f96f)},
		{fb(0xc04f344cbcbab09f), fb(0x3fc5347b37396fae), fb(0x3feb7b1dde440edc), fb(0xc042f1298a4ce5ac), fb(0x4033c1e2d62b1b90), fb(0xc015a73fd39cfe30), fb(0x40960cdc4500eb00)},
		{fb(0xc058bd4c36806d3b), fb(0xbfe0e6e9efb52f18), fb(0xbfb127a2fe1076cd), fb(0x405822680c4472c4), fb(0x3fe60ab83a425e36), fb(0x3fc003ebbd129329), fb(0x40b2c76219a42cea)},
		{fb(0xbfe443c8721e5bec), fb(0xbfed9d97e969e59a), fb(0xbfdf27d9c7521268), fb(0x403f37c679d689e8), fb(0xbfe287d5ee0c8cad), fb(0x4025e2ca3f617cdf), fb(0x3e4aff4800000000)},
		{fb(0x404d11746055bb00), fb(0x40379305e975eab4), fb(0xc051595ae8164ed2), fb(0x3fe4726a69291436), fb(0xc03f09af772fe948), fb(0x404145e9530ba284), fb(0x4086150240e14502)},
		{fb(0xc0537f91c26c9f80), fb(0x4058ac9ce92c1664), fb(0xbfe4bf89d2dd165a), fb(0xc039aa62e1ea09bc), fb(0xc052d39bbd686840), fb(0xc0527fddfbdec4b8), fb(0x40bee5d54b53eac8)},
	}
	for _, tt := range tests {
		sameBits(t, "XLine2DPtSegDistSq", XLine2DPtSegDistSq(tt.x1, tt.y1, tt.x2, tt.y2, tt.px, tt.py), tt.want)
	}
	// On the segment, before it and after it.
	sameBits(t, "on", XLine2DPtSegDistSq(0, 0, 10, 0, 5, 0), 0)
	sameBits(t, "before", XLine2DPtSegDistSq(0, 0, 10, 0, -3, 4), 25)
	sameBits(t, "after", XLine2DPtSegDistSq(0, 0, 10, 0, 13, -4), 25)
	sameBits(t, "beside", XLine2DPtSegDistSq(0, 0, 10, 0, 4, 3), 9)
}

func TestXLine2DIntersect(t *testing.T) {
	tests := []struct {
		ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 float64
		ok                                     bool
		x, y                                   float64
	}{
		{fb(0x4048d661a5be8ae4), fb(0xc03b9237431bf080), fb(0x40578074a7db5a7c), fb(0xbfdbded43a8f9ac4), fb(0x3fe33f89e528b952), fb(0xc04ffc8bcf5d9f33), fb(0x4022fe800330e8a0), fb(0x403c73b7f6397510), false, 0, 0},
		{fb(0xbfe3b7eb3772187c), fb(0xbfeb0ba491b08e43), fb(0x40440591b1d0b164), fb(0xc0060fb4106cf400), fb(0x3feb659fab691097), fb(0xbfe28ba9e73dfc32), fb(0xc0564c868b72bfd2), fb(0xc031ad2030a8c688), true, fb(0xbfe1d6dfb68672e6), fb(0xbfeb224503c77272)},
		{fb(0x403bbbee6b35c9e4), fb(0xc0531de607c936cc), fb(0x3fe8e22df5684aa1), fb(0x4013dd2ef7cb6890), fb(0x404deaad9160e608), fb(0x404b17edc8e569c0), fb(0x3fe8b318d2b21b45), fb(0xbfe8366eeb4a53b6), true, fb(0x4001cbf6e7045de0), fb(0x3fe3089107395700)},
		{fb(0x4034323cdecc4a20), fb(0xc02eeb660dbefb50), fb(0xc03cbb70de584528), fb(0x40587dc4aaa060f8), fb(0xbfeda83a07b6b0ba), fb(0x405261c20f442292), fb(0xc0301100a1ca6d1c), fb(0x404d1fbf6b296a98), true, fb(0xc029e88fc7320564), fb(0x404eb1c98c4b685e)},
		{fb(0xc0372d33b4006f24), fb(0xbfcf9ee04530e076), fb(0x404341c39ea023b8), fb(0x404e33d8f76b55ec), fb(0xbf461b826d886148), fb(0xbfcbb4646dbbcb3e), fb(0xc05216c704ace612), fb(0x3fcf2df9d24d0ee1), true, fb(0xc036ff2a767034a1), fb(0xbfb1fadbe4df04be)},
		{fb(0x40136cb4545fbc50), fb(0xc002085f4d73f220), fb(0x4057073ce8d1a63c), fb(0xbfd1ced54db976cd), fb(0x4047713e620fc4c4), fb(0x403e129dce8667b0), fb(0xc0429f634cf8385a), fb(0xc058d3587aa4a625), true, fb(0x403a2d5ce2c4f0e7), fb(0xbffc573a92c4083f)},
		{fb(0x405183895149332e), fb(0xc04e3eb057f50bf6), fb(0xc052eef9b3994bfc), fb(0xc052ffb757d91d6f), fb(0x404b5024970342c4), fb(0xbfebfa286694258c), fb(0x3fcb447365bcf72e), fb(0x40467fbfff644c3c), false, 0, 0},
		{fb(0x404c63f9ab26e36c), fb(0x4032de71ca24b4ac), fb(0xc0243abea47353d0), fb(0xc05672173f128ecc), fb(0xc050ab6a0831ef84), fb(0xc051787e054244ad), fb(0x4046f737002e6a94), fb(0xbfb43c477968edec), true, fb(0x40464e37cfbfd400), fb(0xbfecb7f2d4882840)},
		{fb(0xbfec7fefae2cc455), fb(0xc052e8a2dd13cd05), fb(0x40289c5c04654888), fb(0x4044e72d87d24874), fb(0xbfd6ea7fc9838908), fb(0x404247ea7110c2d0), fb(0x404e0e6839a06840), fb(0xbfeb56f67eb680aa), true, fb(0x4025dccc940cd0d2), fb(0x403d939aa727cb78)},
		{fb(0xbfe095523498fdc4), fb(0xc006899301d1dee0), fb(0x3feae1125ef74c7b), fb(0x4046ca2f23c17fa8), fb(0x40541ec49e0e334e), fb(0x40484aad3f9ddcb0), fb(0x40583dd85fd4f430), fb(0xbfcfbdc9c11e4c9f), false, 0, 0},
		{fb(0x4020d250a034cb40), fb(0xbfe1c7018ce00f12), fb(0x404ed7b08febc05c), fb(0xc054fed2e5af06ee), fb(0xc052f55dbf755126), fb(0xc046703aaa9197d0), fb(0x4053beae71a81238), fb(0xc0360dbc25bb65a4), true, fb(0x403b079a6f7e4192), fb(0xc03db6572c7b8b62)},
		{fb(0xc0412b9a5496ff8c), fb(0x404fac84bb5eaf84), fb(0xc0132b9c4e5f2a90), fb(0x3f7451c0876fcccd), fb(0x3fe2e842329f51c5), fb(0x4044c3a8271370b0), fb(0xc041ff4de89209a4), fb(0x3fcb4ae04b776110), true, fb(0xc02f3e60029af884), fb(0x403738479088b9d0)},
		{fb(0x3fc231b69f525100), fb(0xc052218f8009c351), fb(0xc04831bdd4ddce0d), fb(0xc032c1cd010e5554), fb(0xc00808ff7554ff20), fb(0xc04b23b6f497b653), fb(0xc0575f5bd54afe31), fb(0x40542bd6247947f2), false, 0, 0},
		{fb(0x404266ccdca4b2fc), fb(0x402cd1378ce16ce8), fb(0xc0564d26e8f8359d), fb(0xc03625dffa799210), fb(0xc035d5ce4ae355d4), fb(0xbfb4b895d7b32dae), fb(0xc0341451d296b7a8), fb(0xc038ac11f49c1050), true, fb(0xc035a8a44ad7beb0), fb(0xc0046a2fd4b438e8)},
		{fb(0x3fe0b1717db446ac), fb(0xc022b1b92f0a15d0), fb(0x4045ff834c928584), fb(0x405228249a5761c2), fb(0x403abca6788c2adc), fb(0x3f8fee5583fe0ae1), fb(0x3fc042b9c3acc857), fb(0x4033263c7dd8437c), true, fb(0x4026b4f92d4f72bc), fb(0x4026276fd47c64a2)},
		{fb(0x3fc9ec4124ad8f57), fb(0x3fc4537f9872b629), fb(0x3faad177459c5467), fb(0x3feeb21f24c9a785), fb(0x4052a2ae4fe3a608), fb(0xc020b6e7c752dd30), fb(0x404e3b9689678f14), fb(0xc057b93ffaa17d3f), false, 0, 0},
		{fb(0xc04cb6488fa24f90), fb(0x4053f07a3c10bc36), fb(0x3fdb7990467d6ccd), fb(0xbfeea25565dea21c), fb(0xbf9d9c8059ae69c3), fb(0xbf9d1188a392c290), fb(0xc04300c2b0c637cc), fb(0x4040773e1a76ca28), true, fb(0xbfe28114e19f4f00), fb(0x3fdcb24455dbcf00)},
		{fb(0x405779a58afb1f0c), fb(0xbfd33b8677497e48), fb(0xbfe06240acf0d90a), fb(0x402e870cac09aca0), fb(0x40387af746c2c028), fb(0x4046ce5ac3aaf28c), fb(0x3fec965fbdc22b33), fb(0xbfd59fccc5b8b7e9), true, fb(0x402055c0c57f5820), fb(0x402baa7527779397)},
		{fb(0xbfe2c3c68284cb72), fb(0xbfda60c4ebb36aac), fb(0xc04b7fa045e334de), fb(0xc04a75df8257f939), fb(0x40412e359b17a514), fb(0x3fa41130b770070a), fb(0xc04d6bf308521678), fb(0xc0447b46062febaa), true, fb(0xc03d00ed7b845974), fb(0xc03bd606ea3d2ce7)},
		{fb(0x401f50ac7a4f0900), fb(0xc047fd3afda7026a), fb(0xbfe0fbba604dd174), fb(0x3fec7147b764d1f3), fb(0xc047764783568f1c), fb(0x40452cbbb9f678bc), fb(0x404c768a11c4d6b0), fb(0xc050257d236979de), true, fb(0x3fe8f319213953c8), fb(0xc01b15e92b56bac0)},
	}
	for _, tt := range tests {
		a, b := NewXLine2D(tt.ax1, tt.ay1, tt.ax2, tt.ay2), NewXLine2D(tt.bx1, tt.by1, tt.bx2, tt.by2)
		got, ok := a.Intersect(b)
		if ok != tt.ok {
			t.Errorf("%v.Intersect(%v) ok = %v, want %v", a, b, ok, tt.ok)
			continue
		}
		if ok {
			samePoint(t, "Intersect", got, XPoint2D{X: tt.x, Y: tt.y})
		}
	}

	cross, ok := NewXLine2D(0, 0, 10, 10).Intersect(NewXLine2D(0, 10, 10, 0))
	if !ok || cross != (XPoint2D{X: 5, Y: 5}) {
		t.Errorf("crossing diagonals = %v, %v", cross, ok)
	}
	if _, ok := NewXLine2D(0, 0, 10, 0).Intersect(NewXLine2D(0, 1, 10, 1)); ok {
		t.Error("parallel segments must not intersect")
	}
	if end, ok := NewXLine2D(0, 0, 10, 0).Intersect(NewXLine2D(10, -5, 10, 5)); !ok || end != (XPoint2D{X: 10, Y: 0}) {
		t.Errorf("touching segments = %v, %v", end, ok)
	}
}

func TestXLine2DAngle(t *testing.T) {
	tests := []struct{ x1, y1, x2, y2, want float64 }{
		{fb(0x4051b7dfce08eb42), fb(0x3fc869e77426c538), fb(0xc055f9c996725614), fb(0x403ea01205ccea90), fb(0x40079e1ef994246e)},
		{fb(0xbfc0f20fa9eebd61), fb(0x3fd1ff5876af8e52), fb(0x4051f8fb730c3444), fb(0xbfe08643b1684398), fb(0xbf86addfa02f8375)},
		{fb(0x40507fe3d81b61de), fb(0x405392905bbd1b68), fb(0xbfd9b77c76bb9148), fb(0x40500108b09e7376), fb(0xc00770539bb13c9c)},
		{fb(0xbfd07481d3eaf1c5), fb(0xc031020afbdbffa8), fb(0x3fe015631b5dc28f), fb(0xc041de9e45c35990), fb(0xbff87bf152069f30)},
		{fb(0x40518ef3747fcc64), fb(0x3fb71a40abe3d23e), fb(0x4041b222a7618b9c), fb(0xbfd0642102815280), fb(0xc0090d9fac1d594a)},
		{fb(0xbfe4ed3cff665280), fb(0x3feeff2ca28673e9), fb(0xc04229eff5c0810a), fb(0x4054d697690d0428), fb(0x3fffabc56b269791)},
		{fb(0x3fab94c816ea031f), fb(0xc03cc35baf35cbbc), fb(0x4003557c4fa0ef40), fb(0x40543666d1d4b5d4), fb(0x3ff8c9b33db23512)},
		{fb(0xc05786816ff4d6bc), fb(0xbfe8e881d131b7b6), fb(0x3fd3f1f35d2f0d0a), fb(0x3fe97a736551ff40), fb(0x3f911381d7f6d8c1)},
	}
	for _, tt := range tests {
		l := NewXLine2D(tt.x1, tt.y1, tt.x2, tt.y2)
		nearAngle(t, "Angle", l.Angle(), tt.want)
	}
}

func TestXLine2DPoints(t *testing.T) {
	l := XLine2DLine(NewXPoint2D(1, 2), NewXPoint2D(4, 8))
	if l != NewXLine2D(1, 2, 4, 8) {
		t.Fatalf("XLine2DLine = %v", l)
	}
	if got := l.Middle(); got != NewXPoint2D(2.5, 5) {
		t.Errorf("Middle() = %v", got)
	}
	if l.P1() != NewXPoint2D(1, 2) || l.P2() != NewXPoint2D(4, 8) {
		t.Errorf("P1(), P2() = %v, %v", l.P1(), l.P2())
	}
	if got := l.WithPoint1(NewXPoint2D(0, 0)); got != NewXLine2D(0, 0, 4, 8) {
		t.Errorf("WithPoint1 = %v", got)
	}
	if got := l.WithPoint2(NewXPoint2D(0, 0)); got != NewXLine2D(1, 2, 0, 0) {
		t.Errorf("WithPoint2 = %v", got)
	}
}
