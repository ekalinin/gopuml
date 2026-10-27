// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/Stripe.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/StripeFrontier.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/SymetricalTee.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/SymetricalTeePositioned.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/mindmap/Tetris.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package mindmap

import (
	"cmp"
	"math"
	"slices"

	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The toString methods of these classes, which only debugging uses, and the
// methods that nothing calls (StripeFrontier.contains, Stripe.contains,
// SymetricalTee.getFullElongation and getFullThickness) are not ported.

// stripe is the height value of the frontier between start and end.
type stripe struct {
	x1, x2, value float64
}

func newStripe(x1, x2, value float64) stripe {
	if x2 <= x1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return stripe{x1: x1, x2: x2, value: value}
}

// stripeFrontier is the lower edge of the tees placed so far, as stripes
// that cover the whole axis without overlapping. Upstream keeps them in a
// TreeSet ordered by their start; no two stripes share a start, so a slice
// sorted by start iterates in the same order.
type stripeFrontier struct {
	stripes []stripe
}

func newStripeFrontier() *stripeFrontier {
	return &stripeFrontier{stripes: []stripe{newStripe(-math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64)}}
}

func (f *stripeFrontier) isEmpty() bool { return len(f.stripes) == 1 }

// getContact is the highest value of the stripes under x1 to x2.
func (f *stripeFrontier) getContact(x1, x2 float64) float64 {
	result := -math.MaxFloat64
	for _, s := range f.collisionning(x1, x2) {
		result = max(result, s.value)
	}
	return result
}

func (f *stripeFrontier) addSegment(x1, x2, value float64) {
	if x2 <= x1 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	collisions := f.collisionning(x1, x2)
	if len(collisions) > 1 {
		x := x1
		for _, tmp := range collisions[1:] {
			f.addSegment(x, tmp.x1, value)
			x = tmp.x1
		}
		f.addSegment(x, x2, value)
	} else {
		f.addSingleInternal(x1, x2, value, collisions[0])
	}
}

func (f *stripeFrontier) addSingleInternal(x1, x2, value float64, touch stripe) {
	if value <= touch.value {
		return
	}
	f.remove(touch)
	if touch.x1 != x1 {
		f.add(newStripe(touch.x1, x1, touch.value))
	}
	f.add(newStripe(x1, x2, value))
	if x2 != touch.x2 {
		f.add(newStripe(x2, touch.x2, touch.value))
	}
}

// add and remove are TreeSet's, which compare the stripes by their start.
func (f *stripeFrontier) add(s stripe) {
	i, found := slices.BinarySearchFunc(f.stripes, s.x1, func(e stripe, x1 float64) int {
		return cmp.Compare(e.x1, x1)
	})
	if !found {
		f.stripes = slices.Insert(f.stripes, i, s)
	}
}

func (f *stripeFrontier) remove(s stripe) {
	f.stripes = slices.DeleteFunc(f.stripes, func(e stripe) bool { return e.x1 == s.x1 })
}

// collisionning returns the stripes under x1 to x2, in order.
func (f *stripeFrontier) collisionning(x1, x2 float64) []stripe {
	var result []stripe
	for _, s := range f.stripes {
		if x1 >= s.x2 {
			continue
		}
		result = append(result, s)
		if x2 <= s.x2 {
			return result
		}
	}
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// symetricalTee is the shape of a finger: its phalanx (thickness1 by
// elongation1) followed by its nail (thickness2 by elongation2), both
// centered on the same axis.
type symetricalTee struct {
	thickness1, elongation1, thickness2, elongation2 float64
}

func newSymetricalTee(thickness1, elongation1, thickness2, elongation2 float64) *symetricalTee {
	return &symetricalTee{thickness1: thickness1, elongation1: elongation1, thickness2: thickness2,
		elongation2: elongation2}
}

// symetricalTeePositioned is a tee with the position of its axis.
type symetricalTeePositioned struct {
	tee *symetricalTee
	y   float64
}

func newSymetricalTeePositioned(tee *symetricalTee) *symetricalTeePositioned {
	return &symetricalTeePositioned{tee: tee}
}

func (p *symetricalTeePositioned) moveSoThatSegmentA1isOn(newY float64) {
	current := p.segmentA1().Y1
	p.y += newY - current
}

func (p *symetricalTeePositioned) moveSoThatSegmentA2isOn(newY float64) {
	current := p.segmentA2().Y1
	p.y += newY - current
}

func (p *symetricalTeePositioned) move(delta float64) { p.y += delta }

func (p *symetricalTeePositioned) segmentA1() geom.XLine2D {
	return geom.NewXLine2D(0, p.y-p.tee.thickness1/2, p.tee.elongation1, p.y-p.tee.thickness1/2)
}

func (p *symetricalTeePositioned) segmentB1() geom.XLine2D {
	return geom.NewXLine2D(0, p.y+p.tee.thickness1/2, p.tee.elongation1, p.y+p.tee.thickness1/2)
}

func (p *symetricalTeePositioned) segmentA2() geom.XLine2D {
	return geom.NewXLine2D(p.tee.elongation1, p.y-p.tee.thickness2/2, p.tee.elongation1+p.tee.elongation2,
		p.y-p.tee.thickness2/2)
}

func (p *symetricalTeePositioned) segmentB2() geom.XLine2D {
	return geom.NewXLine2D(p.tee.elongation1, p.y+p.tee.thickness2/2, p.tee.elongation1+p.tee.elongation2,
		p.y+p.tee.thickness2/2)
}

func (p *symetricalTeePositioned) maxX() float64 { return p.tee.elongation1 + p.tee.elongation2 }

func (p *symetricalTeePositioned) maxY() float64 {
	return p.y + max(p.tee.thickness1/2, p.tee.thickness2/2)
}

func (p *symetricalTeePositioned) minY() float64 {
	return p.y - max(p.tee.thickness1/2, p.tee.thickness2/2)
}

// getMax returns the lower of two placements of the same tee.
func (p *symetricalTeePositioned) getMax(other *symetricalTeePositioned) *symetricalTeePositioned {
	if p.tee != other.tee {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if other.y > p.y {
		return other
	}
	return p
}

// tetris packs tees one below the other, each as high as the frontier of
// the previous ones allows, then centers them on 0.
type tetris struct {
	frontier   *stripeFrontier
	elements   []*symetricalTeePositioned
	minY, maxY float64
}

// newTetris is upstream's constructor without the name, which only
// toString reads.
func newTetris() *tetris {
	return &tetris{frontier: newStripeFrontier(), minY: math.MaxFloat64, maxY: -math.MaxFloat64}
}

func (t *tetris) balance() {
	if len(t.elements) == 0 {
		return
	}
	if t.minY != math.MaxFloat64 {
		uml.Throw(uml.IllegalStateException, "")
	}
	for _, element := range t.elements {
		t.minY = min(t.minY, element.minY())
		t.maxY = max(t.maxY, element.maxY())
	}
	mean := (t.minY + t.maxY) / 2
	for _, stp := range t.elements {
		stp.move(-mean)
	}
}

func (t *tetris) height() float64 {
	if len(t.elements) == 0 {
		return 0
	}
	return t.maxY - t.minY
}

func (t *tetris) width() float64 {
	result := 0.0
	for _, tee := range t.elements {
		result = max(result, tee.maxX())
	}
	return result
}

func (t *tetris) add(tee *symetricalTee) {
	if t.frontier.isEmpty() {
		t.addInternal(newSymetricalTeePositioned(tee))
		return
	}
	c1 := t.frontier.getContact(0, tee.elongation1)
	c2 := t.frontier.getContact(tee.elongation1, tee.elongation1+tee.elongation2)
	p1 := newSymetricalTeePositioned(tee)
	p1.moveSoThatSegmentA1isOn(c1)
	p2 := newSymetricalTeePositioned(tee)
	p2.moveSoThatSegmentA2isOn(c2)
	t.addInternal(p1.getMax(p2))
}

func (t *tetris) addInternal(result *symetricalTeePositioned) {
	t.elements = append(t.elements, result)
	b1 := result.segmentB1()
	t.frontier.addSegment(b1.X1, b1.X2, b1.Y1)
	b2 := result.segmentB2()
	if b2.X1 != b2.X2 {
		t.frontier.addSegment(b2.X1, b2.X2, b2.Y1)
	}
}
