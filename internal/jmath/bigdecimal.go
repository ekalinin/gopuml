// SPDX-License-Identifier: MIT
// Provenance: original

package jmath

import (
	"math"
	"math/big"
	"strings"
)

// Exception is the panic value of the Java exceptions that BigDecimal
// throws. Class is the qualified name of the exception and Message its
// message, empty when the JDK has none.
type Exception struct {
	Class   string
	Message string
}

// Error formats the exception as Java's Throwable.toString does.
func (e *Exception) Error() string {
	if e.Message == "" {
		return e.Class
	}
	return e.Class + ": " + e.Message
}

// Java exception classes thrown by BigDecimal.
const (
	NumberFormatException = "java.lang.NumberFormatException"
	ArithmeticException   = "java.lang.ArithmeticException"
)

func throw(class, message string) { panic(&Exception{Class: class, Message: message}) }

// BigDecimal is java.math.BigDecimal: the value unscaled × 10^-scale. Two
// values that differ only by their scale (1.5 and 1.50) are equal for
// CompareTo and print differently with ToPlainString. A BigDecimal is
// immutable; the zero value is 0 with a scale of 0.
type BigDecimal struct {
	unscaled *big.Int
	scale    int32
}

// BigDecimalZero is BigDecimal.ZERO.
var BigDecimalZero = BigDecimal{}

var (
	bigTen      = big.NewInt(10)
	bigFive     = big.NewInt(5)
	bigMinInt64 = big.NewInt(math.MinInt64)
	bigMaxInt64 = big.NewInt(math.MaxInt64)
)

func (b BigDecimal) intVal() *big.Int {
	if b.unscaled == nil {
		return new(big.Int)
	}
	return b.unscaled
}

// BigDecimalOfInt64 is new BigDecimal(long) and BigDecimal.valueOf(long): a
// scale of 0.
func BigDecimalOfInt64(v int64) BigDecimal { return BigDecimal{unscaled: big.NewInt(v)} }

// BigDecimalOfFloat64 is new BigDecimal(double): the exact binary value of
// v, so 0.1 has 55 fraction digits. Negative zero is 0. It panics with
// NumberFormatException on NaN and the infinities.
func BigDecimalOfFloat64(v float64) BigDecimal {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		throw(NumberFormatException, "Infinite or NaN")
	}
	if v == 0 {
		return BigDecimal{}
	}
	frac, exp := math.Frexp(v)
	// v = significand × 2^exponent with an integral significand of 53 bits,
	// reduced while even, as the JDK constructor does.
	significand := int64(math.Ldexp(frac, 53))
	exponent := exp - 53
	for significand&1 == 0 {
		significand >>= 1
		exponent++
	}
	u := big.NewInt(significand)
	if exponent >= 0 {
		return BigDecimal{unscaled: u.Lsh(u, uint(exponent))}
	}
	// significand / 2^k = significand × 5^k / 10^k.
	k := -exponent
	u.Mul(u, new(big.Int).Exp(bigFive, big.NewInt(int64(k)), nil))
	return BigDecimal{unscaled: u, scale: int32(k)}
}

// BigDecimalOf is new BigDecimal(String): an optional sign, digits with at
// most one decimal point, and an optional exponent (1.5e-3). It panics with
// the JDK's NumberFormatException messages, such as "Character array
// contains more than one decimal point." for 1.2.3 and "No digits found."
// for "+".
func BigDecimalOf(s string) BigDecimal {
	if s == "" {
		throw(NumberFormatException, "")
	}
	i := 0
	negative := false
	if s[0] == '+' || s[0] == '-' {
		negative = s[0] == '-'
		i++
	}
	var digits strings.Builder
	point := false
	scale := int64(0)
	for ; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digits.WriteByte(c)
			if point {
				scale++
			}
			continue
		case c == '.':
			if point {
				throw(NumberFormatException, "Character array contains more than one decimal point.")
			}
			point = true
			continue
		case c == 'e' || c == 'E':
		default:
			throw(NumberFormatException, "Character "+string(rune(c))+
				" is neither a decimal digit number, decimal point, nor \"e\" notation exponential mark.")
		}
		break
	}
	if digits.Len() == 0 {
		throw(NumberFormatException, "No digits found.")
	}
	if i < len(s) {
		scale -= parseExponent(s[i+1:])
	}
	if scale < math.MinInt32 || scale > math.MaxInt32 {
		throw(NumberFormatException, "Scale out of range.")
	}
	u, _ := new(big.Int).SetString(digits.String(), 10)
	if negative {
		u.Neg(u)
	}
	return BigDecimal{unscaled: u, scale: int32(scale)}
}

// parseExponent reads the digits after the exponent mark, with an optional
// sign. The JDK throws a NumberFormatException without a message when they
// are missing or not digits.
func parseExponent(s string) int64 {
	negative := false
	if s != "" && (s[0] == '+' || s[0] == '-') {
		negative = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		throw(NumberFormatException, "")
	}
	var e int64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			throw(NumberFormatException, "")
		}
		e = e*10 + int64(s[i]-'0')
		if e > math.MaxInt32 {
			throw(NumberFormatException, "Exponent overflow.")
		}
	}
	if negative {
		return -e
	}
	return e
}

// Scale is BigDecimal.scale.
func (b BigDecimal) Scale() int { return int(b.scale) }

// Signum is BigDecimal.signum: -1, 0 or 1.
func (b BigDecimal) Signum() int { return b.intVal().Sign() }

// unscaledAt returns the unscaled value of b at a scale at least b's.
func (b BigDecimal) unscaledAt(scale int32) *big.Int {
	u := new(big.Int).Set(b.intVal())
	if scale > b.scale {
		u.Mul(u, new(big.Int).Exp(bigTen, big.NewInt(int64(scale-b.scale)), nil))
	}
	return u
}

// Add is BigDecimal.add: the scale of the result is the larger scale.
func (b BigDecimal) Add(other BigDecimal) BigDecimal {
	scale := max(b.scale, other.scale)
	u := b.unscaledAt(scale)
	return BigDecimal{unscaled: u.Add(u, other.unscaledAt(scale)), scale: scale}
}

// Multiply is BigDecimal.multiply: the scale of the result is the sum of
// the scales.
func (b BigDecimal) Multiply(other BigDecimal) BigDecimal {
	return BigDecimal{unscaled: new(big.Int).Mul(b.intVal(), other.intVal()), scale: b.scale + other.scale}
}

// CompareTo is BigDecimal.compareTo: -1, 0 or 1, ignoring the scales, so
// 10 and 10.0 are equal.
func (b BigDecimal) CompareTo(other BigDecimal) int {
	scale := max(b.scale, other.scale)
	return b.unscaledAt(scale).Cmp(other.unscaledAt(scale))
}

// rat returns the exact value of b.
func (b BigDecimal) rat() *big.Rat {
	r := new(big.Rat).SetInt(b.intVal())
	if b.scale > 0 {
		return r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(bigTen, big.NewInt(int64(b.scale)), nil)))
	}
	if b.scale < 0 {
		return r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(bigTen, big.NewInt(int64(-b.scale)), nil)))
	}
	return r
}

// DoubleValue is BigDecimal.doubleValue: the float64 nearest to b.
func (b BigDecimal) DoubleValue() float64 {
	f, _ := b.rat().Float64()
	return f
}

// integer returns b without its fraction, truncated toward zero, as
// BigDecimal.toBigInteger, and whether the fraction was zero.
func (b BigDecimal) integer() (*big.Int, bool) {
	if b.scale <= 0 {
		return b.unscaledAt(0), true
	}
	q, r := new(big.Int).QuoRem(b.intVal(), new(big.Int).Exp(bigTen, big.NewInt(int64(b.scale)), nil), new(big.Int))
	return q, r.Sign() == 0
}

// LongValue is BigDecimal.longValue: the integral part of b, truncated
// toward zero, in the low 64 bits of its two's complement.
func (b BigDecimal) LongValue() int64 {
	q, _ := b.integer()
	// The low 64 bits of the two's complement, as BigInteger.longValue.
	m := new(big.Int).And(q, new(big.Int).SetUint64(math.MaxUint64))
	return int64(m.Uint64())
}

// LongValueExact is BigDecimal.longValueExact. It panics with the JDK's
// ArithmeticException when b has a fraction ("Rounding necessary") or does
// not fit in an int64 ("Overflow").
func (b BigDecimal) LongValueExact() int64 {
	q, exact := b.integer()
	if !exact {
		throw(ArithmeticException, "Rounding necessary")
	}
	if q.Cmp(bigMinInt64) < 0 || q.Cmp(bigMaxInt64) > 0 {
		throw(ArithmeticException, "Overflow")
	}
	return q.Int64()
}

// ToPlainString is BigDecimal.toPlainString: the value without an
// exponent, with scale fraction digits (1.50), or followed by -scale zeros
// for a negative scale.
func (b BigDecimal) ToPlainString() string {
	u := b.intVal()
	if b.scale == 0 {
		return u.String()
	}
	if b.scale < 0 {
		if u.Sign() == 0 {
			return "0"
		}
		return u.String() + strings.Repeat("0", int(-b.scale))
	}
	digits := new(big.Int).Abs(u).String()
	sign := ""
	if u.Sign() < 0 {
		sign = "-"
	}
	scale := int(b.scale)
	if len(digits) <= scale {
		return sign + "0." + strings.Repeat("0", scale-len(digits)) + digits
	}
	return sign + digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
}
