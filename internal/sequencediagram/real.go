// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/Real.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealOrigin.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/AbstractReal.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealLine.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/PositiveForce.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealMoveable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealImpl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealDelta.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealDeltaLive.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealMax.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealMin.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealMiddle2.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/real/RealUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

// The constraint solver of the Teoz layout (upstream's real package): a
// Real is a coordinate on a line, defined from an origin by fixed offsets
// and by "at least" offsets; ensureBiggerThan adds a constraint. Compiling
// the line moves the moveable points until every constraint holds.
//
// Upstream's real package only serves Teoz; it is ported here, with the
// teoz files, so that its types stay unexported where upstream's are
// package-private.

import (
	"math"
	"strconv"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Real is a coordinate of the Teoz layout.
type Real interface {
	Name() string
	// CurrentValue returns the value; it is meaningful once the line is
	// compiled.
	CurrentValue() float64
	// AddFixed returns the coordinate at a fixed distance from this one.
	AddFixed(delta float64) Real
	// AddAtLeast returns a coordinate at least delta after this one.
	AddAtLeast(delta float64) Real
	// EnsureBiggerThan constrains this coordinate to be at least other.
	EnsureBiggerThan(other Real)
	realLine() *realLine
}

// RealOrigin is the origin of a line: compiling it solves the line.
type RealOrigin interface {
	Real
	CompileNow()
}

// realMoveable is a coordinate the solver can move.
type realMoveable interface {
	Real
	move(delta float64)
}

// realLine holds the constraints of a line. Upstream also keeps every
// Real of the line, to compute absolute bounds that nothing reads; this
// port does not.
type realLine struct {
	forces []*positiveForce
	// depth counts the nested RealMax evaluations, upstream's thread-local
	// recursion guard.
	depth int
}

func (l *realLine) addForce(force *positiveForce) { l.forces = append(l.forces, force) }

// compile applies the forces, in their order of creation, until none
// moves a point.
func (l *realLine) compile() {
	cpt := 0
	for {
		done := true
		for _, f := range l.forces {
			if f.apply() {
				done = false
			}
		}
		if done {
			return
		}
		cpt++
		if cpt > 99999 {
			uml.Throw(uml.IllegalStateException, "Infinite Loop?")
		}
	}
}

// positiveForce keeps movingPoint at least minimunDistance after
// fixedPoint.
type positiveForce struct {
	fixedPoint      Real
	movingPoint     realMoveable
	minimunDistance float64
}

func newPositiveForce(fixedPoint Real, movingPoint realMoveable, minimunDistance float64) *positiveForce {
	if fixedPoint == Real(movingPoint) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &positiveForce{fixedPoint: fixedPoint, movingPoint: movingPoint, minimunDistance: minimunDistance}
}

func (f *positiveForce) apply() bool {
	movingPointValue := f.movingPoint.CurrentValue()
	fixedPointValue := f.fixedPoint.CurrentValue()
	distance := movingPointValue - fixedPointValue
	diff := distance - f.minimunDistance
	epsilon := jmath.Max(0.000001, 1000*javaUlp(jmath.Max(math.Abs(movingPointValue), math.Abs(fixedPointValue))))
	if diff >= -epsilon {
		return false
	}
	f.movingPoint.move(-diff)
	return true
}

// javaUlp is Java's Math.ulp: the distance to the next double larger in
// magnitude.
func javaUlp(d float64) float64 {
	d = math.Abs(d)
	if math.IsInf(d, 0) {
		return math.Inf(1)
	}
	if d == math.MaxFloat64 {
		return math.Pow(2, 971)
	}
	return math.Nextafter(d, math.Inf(1)) - d
}

// realImpl is a moveable point with its own value: the origin, and the
// points created by AddAtLeast.
type realImpl struct {
	line         *realLine
	name         string
	currentValue float64
}

// RealUtilsCreateOrigin returns the origin of a new line, at 0.
func RealUtilsCreateOrigin() RealOrigin {
	return &realImpl{line: &realLine{}, name: "O"}
}

func (r *realImpl) realLine() *realLine         { return r.line }
func (r *realImpl) Name() string                { return r.name }
func (r *realImpl) CurrentValue() float64       { return r.currentValue }
func (r *realImpl) move(delta float64)          { r.currentValue += delta }
func (r *realImpl) AddFixed(delta float64) Real { return newRealDelta(r, delta) }

func (r *realImpl) AddAtLeast(delta float64) Real {
	result := &realImpl{line: r.line, name: "addAtLeast" + jfmt.DoubleToString(delta),
		currentValue: r.currentValue + delta}
	r.line.addForce(newPositiveForce(r, result, delta))
	return result
}

func (r *realImpl) EnsureBiggerThan(other Real) { r.line.addForce(newPositiveForce(other, r, 0)) }

func (r *realImpl) CompileNow() { r.line.compile() }

// realDelta is a point at a fixed distance from another one; moving it
// moves the other one.
type realDelta struct {
	delegated Real
	diff      float64
}

func newRealDelta(delegated Real, diff float64) *realDelta {
	return &realDelta{delegated: delegated, diff: diff}
}

func (r *realDelta) realLine() *realLine { return r.delegated.realLine() }
func (r *realDelta) Name() string        { return "[Delegated d=" + jfmt.DoubleToString(r.diff) + "]" }

func (r *realDelta) CurrentValue() float64 { return r.delegated.CurrentValue() + r.diff }

func (r *realDelta) AddFixed(delta float64) Real { return newRealDelta(r, delta) }

func (r *realDelta) AddAtLeast(delta float64) Real {
	return newRealDelta(r.delegated.AddAtLeast(delta), r.diff)
}

func (r *realDelta) EnsureBiggerThan(other Real) {
	r.delegated.EnsureBiggerThan(newRealDelta(other, -r.diff))
}

func (r *realDelta) move(delta float64) { r.delegated.(realMoveable).move(delta) }

// realDeltaLive is realDelta with an offset read when the value is: the
// offset may still grow after the point is created.
type realDeltaLive struct {
	delegated Real
	delta     func() float64
}

// RealUtilsWithLiveOffset is base plus an offset read on each evaluation.
func RealUtilsWithLiveOffset(base Real, offset func() float64) Real {
	return &realDeltaLive{delegated: base, delta: offset}
}

func (r *realDeltaLive) realLine() *realLine { return r.delegated.realLine() }
func (r *realDeltaLive) Name() string        { return "[DelegatedLive]" }

func (r *realDeltaLive) CurrentValue() float64 { return r.delegated.CurrentValue() + r.delta() }

func (r *realDeltaLive) AddFixed(delta float64) Real { return newRealDelta(r, delta) }

func (r *realDeltaLive) AddAtLeast(d float64) Real {
	return &realDeltaLive{delegated: r.delegated.AddAtLeast(d), delta: r.delta}
}

func (r *realDeltaLive) EnsureBiggerThan(other Real) {
	delta := r.delta
	r.delegated.EnsureBiggerThan(&realDeltaLive{delegated: other, delta: func() float64 { return -delta() }})
}

func (r *realDeltaLive) move(d float64) { r.delegated.(realMoveable).move(d) }

// realCacheUnset is upstream's Double.MAX_VALUE sentinel of the RealMax
// and RealMin caches.
const realCacheUnset = math.MaxFloat64

// realMax is the maximum of points. Its value is cached on the first
// evaluation, whatever the state of the line, as upstream: it must not be
// read before the line is compiled, nor be constrained.
type realMax struct {
	all   []Real
	cache float64
}

// RealUtilsMax returns the maximum of reals.
func RealUtilsMax(reals ...Real) Real {
	return &realMax{all: append([]Real(nil), reals...), cache: realCacheUnset}
}

func (r *realMax) realLine() *realLine { return r.all[0].realLine() }
func (r *realMax) Name() string        { return "max " + strconv.Itoa(len(r.all)) }

func (r *realMax) CurrentValue() float64 {
	if r.cache != realCacheUnset {
		return r.cache
	}
	line := r.realLine()
	line.depth++
	defer func() { line.depth-- }()
	if line.depth > 1000 {
		uml.Throw(uml.IllegalStateException, "Infinite recursion?")
	}
	result := r.all[0].CurrentValue()
	for _, x := range r.all[1:] {
		if v := x.CurrentValue(); v > result {
			result = v
		}
	}
	r.cache = result
	return result
}

func (r *realMax) AddFixed(delta float64) Real { return newRealDelta(r, delta) }

func (r *realMax) AddAtLeast(float64) Real {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (r *realMax) EnsureBiggerThan(Real) { uml.Throw(uml.UnsupportedOperationException, "") }

// realMin is the minimum of points, cached like realMax.
type realMin struct {
	all   []Real
	cache float64
}

// RealUtilsMin returns the minimum of reals.
func RealUtilsMin(reals ...Real) Real {
	return &realMin{all: append([]Real(nil), reals...), cache: realCacheUnset}
}

func (r *realMin) realLine() *realLine { return r.all[0].realLine() }
func (r *realMin) Name() string        { return "min " + strconv.Itoa(len(r.all)) }

func (r *realMin) CurrentValue() float64 {
	if r.cache != realCacheUnset {
		return r.cache
	}
	result := r.all[0].CurrentValue()
	for _, x := range r.all[1:] {
		if v := x.CurrentValue(); v < result {
			result = v
		}
	}
	r.cache = result
	return result
}

func (r *realMin) AddFixed(delta float64) Real { return newRealDelta(r, delta) }

func (r *realMin) AddAtLeast(float64) Real {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (r *realMin) EnsureBiggerThan(other Real) {
	for _, x := range r.all {
		x.EnsureBiggerThan(other)
	}
}

// realMiddle2 is the middle of two moveable points; moving it moves both.
type realMiddle2 struct {
	p1, p2 realMoveable
}

// RealUtilsMiddle returns the middle of two moveable points; other points
// panic, as upstream's cast fails.
func RealUtilsMiddle(r1, r2 Real) Real {
	p1, ok1 := r1.(realMoveable)
	p2, ok2 := r2.(realMoveable)
	if !ok1 || !ok2 {
		uml.Throw("java.lang.ClassCastException", "")
	}
	return &realMiddle2{p1: p1, p2: p2}
}

func (r *realMiddle2) realLine() *realLine { return r.p1.realLine() }
func (r *realMiddle2) Name() string        { return "middle" }

func (r *realMiddle2) CurrentValue() float64 { return (r.p1.CurrentValue() + r.p2.CurrentValue()) / 2 }

func (r *realMiddle2) AddFixed(delta float64) Real { return newRealDelta(r, delta) }

func (r *realMiddle2) AddAtLeast(float64) Real {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (r *realMiddle2) EnsureBiggerThan(other Real) {
	r.realLine().addForce(newPositiveForce(other, r, 0))
}

func (r *realMiddle2) move(delta float64) {
	r.p1.move(delta / 2)
	r.p2.move(delta / 2)
}
