// SPDX-License-Identifier: MIT
// Provenance: original

package tim

// javaRandom is java.util.Random: a 48-bit linear congruential generator,
// so that a seeded generator gives the same numbers as in Java.
type javaRandom struct {
	seed int64
}

const (
	javaRandomMultiplier = 0x5DEECE66D
	javaRandomAddend     = 0xB
	javaRandomMask       = 1<<48 - 1
)

// newJavaRandom is new Random(seed).
func newJavaRandom(seed int64) *javaRandom {
	return &javaRandom{seed: (seed ^ javaRandomMultiplier) & javaRandomMask}
}

// next is Random.next(bits).
func (r *javaRandom) next(bits uint) int32 {
	r.seed = (r.seed*javaRandomMultiplier + javaRandomAddend) & javaRandomMask
	return int32(uint64(r.seed) >> (48 - bits))
}

// nextInt is Random.nextInt(bound); ok is false where Java throws
// IllegalArgumentException (a bound that is not positive).
func (r *javaRandom) nextInt(bound int32) (v int32, ok bool) {
	if bound <= 0 {
		return 0, false
	}
	v = r.next(31)
	m := bound - 1
	if bound&m == 0 {
		// bound is a power of 2
		return int32((int64(bound) * int64(v)) >> 31), true
	}
	for u := v; ; u = r.next(31) {
		v = u % bound
		if u-v+m >= 0 {
			return v, true
		}
	}
}
