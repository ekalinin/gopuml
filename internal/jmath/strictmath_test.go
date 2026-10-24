// SPDX-License-Identifier: MIT
// Provenance: original

package jmath

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
)

// The expected values in strictmath_data_test.go were produced by Java's
// StrictMath (fdlibm functions and sqrt) and Math (the other functions).
// Results are compared as raw bits, except that every NaN is treated as
// the same value: NaN payloads differ between CPUs even within Java.

const canonicalNaN = 0x7ff8000000000000

func canon(x float64) uint64 {
	if x != x {
		return canonicalNaN
	}
	return math.Float64bits(x)
}

func f64(b uint64) float64 { return math.Float64frombits(b) }

var unaryFuncs = map[string]func(uint64) uint64{
	"sin":       func(x uint64) uint64 { return canon(Sin(f64(x))) },
	"cos":       func(x uint64) uint64 { return canon(Cos(f64(x))) },
	"tan":       func(x uint64) uint64 { return canon(Tan(f64(x))) },
	"asin":      func(x uint64) uint64 { return canon(Asin(f64(x))) },
	"acos":      func(x uint64) uint64 { return canon(Acos(f64(x))) },
	"atan":      func(x uint64) uint64 { return canon(Atan(f64(x))) },
	"exp":       func(x uint64) uint64 { return canon(Exp(f64(x))) },
	"log":       func(x uint64) uint64 { return canon(Log(f64(x))) },
	"cbrt":      func(x uint64) uint64 { return canon(Cbrt(f64(x))) },
	"sqrt":      func(x uint64) uint64 { return canon(Sqrt(f64(x))) },
	"toRadians": func(x uint64) uint64 { return canon(ToRadians(f64(x))) },
	"toDegrees": func(x uint64) uint64 { return canon(ToDegrees(f64(x))) },
	"round":     func(x uint64) uint64 { return uint64(Round(f64(x))) },
	"roundf":    func(x uint64) uint64 { return uint64(uint32(RoundF(math.Float32frombits(uint32(x))))) },
	"rint":      func(x uint64) uint64 { return canon(Rint(f64(x))) },
	"floor":     func(x uint64) uint64 { return canon(Floor(f64(x))) },
	"ceil":      func(x uint64) uint64 { return canon(Ceil(f64(x))) },
	"signum":    func(x uint64) uint64 { return canon(Signum(f64(x))) },
	"abs":       func(x uint64) uint64 { return canon(Abs(f64(x))) },
}

var binaryFuncs = map[string]func(a, b uint64) uint64{
	"atan2": func(y, x uint64) uint64 { return canon(Atan2(f64(y), f64(x))) },
	"pow":   func(x, y uint64) uint64 { return canon(Pow(f64(x), f64(y))) },
	"hypot": func(x, y uint64) uint64 { return canon(Hypot(f64(x), f64(y))) },
	"max":   func(a, b uint64) uint64 { return canon(Max(f64(a), f64(b))) },
	"min":   func(a, b uint64) uint64 { return canon(Min(f64(a), f64(b))) },
}

func TestUnaryCases(t *testing.T) {
	for name, cases := range unaryCases {
		fn := unaryFuncs[name]
		if fn == nil {
			t.Fatalf("no function for %q", name)
		}
		failed := 0
		for i := 0; i < len(cases); i += 2 {
			in, want := cases[i], cases[i+1]
			if got := fn(in); got != want {
				failed++
				if failed <= 10 {
					t.Errorf("%s(%#016x = %v) = %#016x (%v), Java gives %#016x (%v)",
						name, in, f64(in), got, f64(got), want, f64(want))
				}
			}
		}
		t.Logf("%-9s %d explicit inputs, %d mismatches", name, len(cases)/2, failed)
	}
}

func TestBinaryCases(t *testing.T) {
	for name, cases := range binaryCases {
		fn := binaryFuncs[name]
		if fn == nil {
			t.Fatalf("no function for %q", name)
		}
		failed := 0
		for i := 0; i < len(cases); i += 3 {
			a, b, want := cases[i], cases[i+1], cases[i+2]
			if got := fn(a, b); got != want {
				failed++
				if failed <= 10 {
					t.Errorf("%s(%v, %v) = %#016x (%v), Java gives %#016x (%v)",
						name, f64(a), f64(b), got, f64(got), want, f64(want))
				}
			}
		}
		t.Logf("%-9s %d explicit inputs, %d mismatches", name, len(cases)/3, failed)
	}
}

// splitMix64 is the SplitMix64 generator, mirrored by the Java program that
// produced strictmath_data_test.go.
type splitMix64 struct{ s uint64 }

func (r *splitMix64) next() uint64 {
	r.s += 0x9E3779B97F4A7C15
	z := r.s
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// gen draws input bit patterns from a splitMix64. The draws use only
// integer operations and exact conversions, so Java and Go produce the same
// inputs.
//
//   - bits: a random 64-bit pattern
//   - logu: exponent uniform in [a, b], random mantissa; sign random (0),
//     positive (1) or negative (-1)
//   - halves: n*0.5 for a random a-bit signed integer n
//   - ints: a random integer in [-a, a]
//   - near1: 1 plus or minus a random a-bit offset in units of the last place
//   - pick: a special value (zeros, NaN, infinities, ones, smallest
//     subnormals) in a quarter of the draws, a random pattern otherwise
//   - bits32, logu32, halves32: the float32 counterparts, as uint32 bits
type gen struct {
	kind       string
	a, b, sign int64
}

var pickBits = [...]uint64{
	0, 1 << 63, canonicalNaN, 0x7ff0000000000000, 0xfff0000000000000,
	0x3ff0000000000000, 0xbff0000000000000, 1, 0x8000000000000001,
}

func (g gen) draw(r *splitMix64) uint64 {
	switch g.kind {
	case "bits":
		return r.next()
	case "logu":
		r1, r2 := r.next(), r.next()
		e := g.a + int64(r1%uint64(g.b-g.a+1))
		return g.signBit(r2, 63) | uint64(e+1023)<<52 | r2&(1<<52-1)
	case "halves":
		n := int64(r.next()) >> (64 - g.a)
		return math.Float64bits(float64(n) * 0.5)
	case "ints":
		n := int64(r.next()%uint64(2*g.a+1)) - g.a
		return math.Float64bits(float64(n))
	case "pick":
		r1 := r.next()
		if r1%4 == 0 {
			return pickBits[(r1>>8)%uint64(len(pickBits))]
		}
		return r.next()
	case "near1":
		r1 := r.next()
		off := r1 & (1<<g.a - 1)
		if r1>>63 == 0 {
			return 0x3ff0000000000000 + off
		}
		return 0x3fefffffffffffff - off
	case "bits32":
		return r.next() >> 32
	case "logu32":
		r1, r2 := r.next(), r.next()
		e := g.a + int64(r1%uint64(g.b-g.a+1))
		return g.signBit(r2, 31) | uint64(e+127)<<23 | r2&(1<<23-1)
	case "halves32":
		n := int64(r.next()) >> (64 - g.a)
		return uint64(math.Float32bits(float32(n) * 0.5))
	}
	panic("unknown generator " + g.kind)
}

func (g gen) signBit(r2 uint64, pos uint) uint64 {
	switch {
	case g.sign > 0:
		return 0
	case g.sign < 0:
		return 1 << pos
	}
	return r2 >> 63 << pos
}

// bulkSpec is a set of n pseudo-random inputs for fn, drawn from a (and b
// for two-argument functions) with the given seed, and the SHA-256 of
// Java's results: each result as 8 little-endian bytes.
type bulkSpec struct {
	fn   string
	a, b gen
	n    int
	seed uint64
	sha  string
}

func TestBulk(t *testing.T) {
	if testing.Short() {
		t.Skip("evaluates millions of inputs")
	}
	total := map[string]int{}
	for _, s := range bulkSpecs {
		name := fmt.Sprintf("%s/%s", s.fn, s.a.kind)
		t.Run(name, func(t *testing.T) {
			unary, binary2 := unaryFuncs[s.fn], binaryFuncs[s.fn]
			if unary == nil && binary2 == nil {
				t.Fatalf("no function for %q", s.fn)
			}
			r := splitMix64{s.seed}
			h := sha256.New()
			var buf [8]byte
			for range s.n {
				a := s.a.draw(&r)
				var res uint64
				if unary != nil {
					res = unary(a)
				} else {
					res = binary2(a, s.b.draw(&r))
				}
				binary.LittleEndian.PutUint64(buf[:], res)
				h.Write(buf[:])
			}
			if got := hex.EncodeToString(h.Sum(nil)); got != s.sha {
				t.Errorf("%s over %d inputs (a=%v b=%v seed=%#x): results differ from Java", s.fn, s.n, s.a, s.b, s.seed)
			}
		})
		total[s.fn] += s.n
	}
	for fn, n := range total {
		t.Logf("%-9s %d bulk inputs", fn, n)
	}
}
