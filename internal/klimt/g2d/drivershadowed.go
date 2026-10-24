// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/drawing/g2d/DriverShadowedG2d.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package g2d

import (
	"math"
)

// The shadow colors.
const (
	shadowColor     = 0xFFAAAAAA // XColor(170, 170, 170)
	shadowColorLine = 0xFF1E1E1E // XColor(30, 30, 30)
	// Java's createGraphics draws in white by default.
	defaultForeground = 0xFFFFFFFF
)

// shadowImage draws a shape on a new transparent image, deltaShadow away
// from its top-left corner, without antialiasing (the image's Graphics2D
// has none), and blurs it. ok is false when the image would be empty.
func shadowImage(s shape, deltaShadow, dpiFactor float64, draw func(gg *graphics, s shape)) (img *Image,
	bx, by float64, ok bool) {
	bx, by, bw, bh := s.bounds2D()
	w := (bw + deltaShadow*2 + 6) * dpiFactor
	h := (bh + deltaShadow*2 + 6) * dpiFactor
	iw, ih := int(w), int(h)
	// Upstream catches the exception of an empty image and draws nothing.
	if iw <= 0 || ih <= 0 || float64(iw)*float64(ih) > 1<<28 {
		return nil, 0, 0, false
	}
	dest := NewImage(iw, ih, true)
	gg := newGraphics(dest)
	gg.tx = identity.scale(dpiFactor, dpiFactor).translate(deltaShadow-bx, deltaShadow-by)
	draw(gg, s)
	return convolve(dest, 6, dpiFactor), bx, by, true
}

// convolve is the blur of upstream's getConvolveOp: a ConvolveOp with a
// square kernel of side (int)(blurRadius * sqrt(dpiFactor)) whose
// (int)(blurRadius^2 * dpiFactor) elements are 1/count each (only the first
// side^2 are used), with EDGE_NO_OP. As the reference JVM computes it (its
// native imaging library): the four channels of the ARGB pixels are
// convolved as stored, not premultiplied; the pixel (x, y) gets the sum of
// the window [x - side/2, x - side/2 + side) x [y - side/2, ...) times the
// weight in 16-bit fixed point, truncated; the pixels less than side/2 from
// a border are copied.
func convolve(src *Image, blurRadius int, dpiFactor float64) *Image {
	side := int(float64(blurRadius) * math.Sqrt(dpiFactor))
	count := int(float64(blurRadius*blurRadius) * dpiFactor)
	if side <= 0 || count <= 0 {
		return src
	}
	weight := int64(float64(float32(1/float64(count))) * 65536)
	w, h := src.width, src.height
	dst := &Image{width: w, height: h, pix: make([]uint32, len(src.pix)), hasAlpha: src.hasAlpha}
	copy(dst.pix, src.pix)
	a := side / 2
	if w-a <= a || h-a <= a {
		return dst
	}
	// Horizontal sums of each channel over the window (sliding along the
	// row), kept for the side rows of the current window.
	hsum := func(y int, out [][4]int32) {
		row := src.pix[y*w : (y+1)*w]
		var acc [4]int32
		add := func(v uint32, sign int32) {
			acc[0] += sign * int32(v>>24)
			acc[1] += sign * int32(v>>16&0xFF)
			acc[2] += sign * int32(v>>8&0xFF)
			acc[3] += sign * int32(v&0xFF)
		}
		for i := range side {
			add(row[i], 1)
		}
		for x := a; x < w-a; x++ {
			out[x] = acc
			if x-a+side < w {
				add(row[x-a+side], 1)
				add(row[x-a], -1)
			}
		}
	}
	ring := make([][][4]int32, side)
	for j := range ring {
		ring[j] = make([][4]int32, w)
	}
	for j := range side - 1 {
		hsum(j, ring[j])
	}
	for y := a; y < h-a; y++ {
		last := y - a + side - 1
		hsum(last, ring[last%side])
		for x := a; x < w-a; x++ {
			var s [4]int64
			for j := range side {
				r := ring[(y-a+j)%side][x]
				s[0] += int64(r[0])
				s[1] += int64(r[1])
				s[2] += int64(r[2])
				s[3] += int64(r[3])
			}
			var c [4]uint32
			for k := range 4 {
				c[k] = uint32(min(255, s[k]*weight>>16))
			}
			dst.pix[y*w+x] = c[0]<<24 | c[1]<<16 | c[2]<<8 | c[3]
		}
	}
	return dst
}

// drawShadow draws the shadow of a shape: the shape filled in gray (a line
// stroked in dark gray), blurred, under the shape.
func drawShadow(g *graphics, s shape, deltaShadow, dpiFactor float64) {
	if dpiFactor < 1 {
		dpiFactor = 1
	}
	img, bx, by, ok := shadowImage(s, deltaShadow, dpiFactor, func(gg *graphics, s shape) {
		if _, isLine := s.(line2D); isLine {
			gg.setColor(shadowColorLine)
			gg.draw(s)
		} else {
			gg.setColor(shadowColor)
			gg.fill(s)
		}
	})
	if !ok {
		return
	}
	at := g.tx
	g.tx = g.tx.scale(1/dpiFactor, 1/dpiFactor)
	g.drawImage(img, int(bx*dpiFactor), int(by*dpiFactor))
	g.tx = at
}

// drawOnlyLineShadowSpecial draws the shadow of the outline of a shape with
// a transparent background, outside of the shape. As upstream, the outline
// is drawn in the default color of a new Graphics2D, white.
func drawOnlyLineShadowSpecial(g *graphics, s shape, deltaShadow, dpiFactor float64) {
	if dpiFactor < 1 {
		dpiFactor = 1
	}
	img, bx, by, ok := shadowImage(s, deltaShadow, dpiFactor, func(gg *graphics, s shape) {
		gg.setColor(defaultForeground)
		gg.draw(s)
	})
	if !ok {
		return
	}
	at := g.tx
	g.tx = g.tx.scale(1/dpiFactor, 1/dpiFactor)
	sav := g.saveClip()
	_, _, bw, bh := s.bounds2D()
	full := rect2D{0, 0, (bx + bw + deltaShadow*2 + 6) * dpiFactor, (by + bh + deltaShadow*2 + 6) * dpiFactor}
	minus := s.path()
	if dpiFactor != 1 {
		minus = minus.transformed(identity.scale(dpiFactor, dpiFactor))
	}
	g.setClipShapeMinus(full.transformedBy(g.tx), minus.transformed(g.tx))
	g.drawImage(img, int(bx*dpiFactor), int(by*dpiFactor))
	g.restoreClip(sav)
	g.tx = at
}

// transformedBy returns the device rectangle of a rectangle under a scale
// and translation transform.
func (r rect2D) transformedBy(t affine) rect2D {
	x0, y0 := t.apply(r.x, r.y)
	x1, y1 := t.apply(r.x+r.w, r.y+r.h)
	return rect2D{math.Min(x0, x1), math.Min(y0, y1), math.Abs(x1 - x0), math.Abs(y1 - y0)}
}
