// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/awt/PortableImageAwt.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"

	"github.com/ekalinin/gopuml/internal/jmath"
)

// scaleARGB is upstream's PortableImageAwt.scale: an image of
// Math.round(width * scaleFactor) x Math.round(height * scaleFactor)
// pixels, the source drawn scaled with an AffineTransformOp. As Java's
// imaging library does, each destination pixel center is mapped back to
// the source; pixels that fall outside of it stay transparent. Nearest
// neighbor takes the source pixel under the center; bilinear interpolates
// the four source pixels around it (edges extended), channel by channel
// without premultiplication. The bilinear rounding approximates Java's
// fixed point arithmetic: values may differ by one.
func scaleARGB(src PortableImage, scaleFactor float64, interpolationType AffineTransformType) PortableImage {
	if scaleFactor == 1 {
		return src
	}
	sw, sh := src.Width(), src.Height()
	w := int(jmath.Round(float64(sw) * scaleFactor))
	h := int(jmath.Round(float64(sh) * scaleFactor))
	result := NewPortableImageARGB(max(w, 0), max(h, 0))
	at := func(x, y, shift int) float64 {
		x = min(max(x, 0), sw-1)
		y = min(max(y, 0), sh-1)
		return float64(uint32(src.RGB(x, y)) >> shift & 0xFF)
	}
	for y := range h {
		cy := (float64(y) + 0.5) / scaleFactor
		if cy >= float64(sh) {
			continue
		}
		for x := range w {
			cx := (float64(x) + 0.5) / scaleFactor
			if cx >= float64(sw) {
				continue
			}
			if interpolationType != AffineTransformBilinear {
				result.SetRGB(x, y, src.RGB(int(cx), int(cy)))
				continue
			}
			sx := math.Max(cx-0.5, 0)
			sy := math.Max(cy-0.5, 0)
			x0, y0 := int(sx), int(sy)
			fx, fy := sx-float64(x0), sy-float64(y0)
			var v uint32
			for _, shift := range []int{24, 16, 8, 0} {
				top := at(x0, y0, shift) + float64(fx*(at(x0+1, y0, shift)-at(x0, y0, shift)))
				bottom := at(x0, y0+1, shift) + float64(fx*(at(x0+1, y0+1, shift)-at(x0, y0+1, shift)))
				c := math.Floor(top + float64(fy*(bottom-top)) + 0.5)
				v |= uint32(min(255, max(0, c))) << shift
			}
			result.SetRGB(x, y, int32(v))
		}
	}
	return result
}
