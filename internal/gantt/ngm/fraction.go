// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ngm/math/Fraction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

// Package ngm is upstream's load model of gantt diagrams (packages
// gantt.ngm and gantt.ngm.math): exact fractions, piecewise constant
// functions of time that say how much of a task can be done at each
// instant (the days of the week, specific days, a time window, the
// working hours, and their combinations), and the integration of a load
// over them, which gives the end of a task from its start, or its start
// from its end.
package ngm

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// Java exception classes thrown by this package.
const (
	IllegalArgumentException = "java.lang.IllegalArgumentException"
	IllegalStateException    = "java.lang.IllegalStateException"
)

func throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }

// Fraction is upstream's Fraction: an exact rational number, reduced, with a
// positive denominator. The arithmetic wraps on overflow, as Java's longs.
type Fraction struct {
	num, den int64
}

// Constants of upstream's Fraction.
var (
	Zero = Fraction{0, 1}
	One  = Fraction{1, 1}
)

// The operations of upstream's Fraction.PRODUCT, SUM, MIN and MAX.
var (
	Product = Fraction.Multiply
	Sum     = Fraction.Add
	Min     = func(a, b Fraction) Fraction {
		if a.CompareTo(b) <= 0 {
			return a
		}
		return b
	}
	Max = func(a, b Fraction) Fraction {
		if a.CompareTo(b) >= 0 {
			return a
		}
		return b
	}
)

// NewFraction is new Fraction(numerator, denominator), which fails with
// Java's IllegalArgumentException when the denominator is zero.
func NewFraction(numerator, denominator int64) Fraction {
	if denominator == 0 {
		throw(IllegalArgumentException, "Denominator cannot be zero.")
	}
	if denominator < 0 {
		numerator = -numerator
		denominator = -denominator
	}
	abs := numerator
	if abs < 0 {
		abs = -abs
	}
	g := gcd(abs, denominator)
	return Fraction{numerator / g, denominator / g}
}

// FractionOf is Fraction.of(value).
func FractionOf(value int64) Fraction { return NewFraction(value, 1) }

// Numerator is getNumerator.
func (f Fraction) Numerator() int64 { return f.num }

// Denominator is getDenominator.
func (f Fraction) Denominator() int64 { return f.den }

// Reciprocal is reciprocal.
func (f Fraction) Reciprocal() Fraction { return NewFraction(f.den, f.num) }

// Negate is negate.
func (f Fraction) Negate() Fraction { return NewFraction(-f.num, f.den) }

// Add is add.
func (f Fraction) Add(other Fraction) Fraction {
	return NewFraction(f.num*other.den+other.num*f.den, f.den*other.den)
}

// Subtract is subtract.
func (f Fraction) Subtract(other Fraction) Fraction {
	return NewFraction(f.num*other.den-other.num*f.den, f.den*other.den)
}

// Multiply is multiply.
func (f Fraction) Multiply(other Fraction) Fraction {
	return NewFraction(f.num*other.num, f.den*other.den)
}

// MultiplyByLong is multiplyByLong.
func (f Fraction) MultiplyByLong(scalar int64) Fraction { return NewFraction(f.num*scalar, f.den) }

// Divide is divide.
func (f Fraction) Divide(other Fraction) Fraction { return f.Multiply(other.Reciprocal()) }

// WholePart is wholePart: the quotient rounded toward zero.
func (f Fraction) WholePart() int64 { return f.num / f.den }

// IsZero is isZero.
func (f Fraction) IsZero() bool { return f.num == 0 }

// CompareTo is compareTo, by cross products.
func (f Fraction) CompareTo(other Fraction) int {
	lhs, rhs := f.num*other.den, other.num*f.den
	switch {
	case lhs < rhs:
		return -1
	case lhs > rhs:
		return 1
	}
	return 0
}

// String is toString: "3" or "1/3".
func (f Fraction) String() string {
	if f.den == 1 {
		return strconv.FormatInt(f.num, 10)
	}
	return strconv.FormatInt(f.num, 10) + "/" + strconv.FormatInt(f.den, 10)
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
