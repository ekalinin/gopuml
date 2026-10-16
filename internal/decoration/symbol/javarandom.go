// SPDX-License-Identifier: MIT
// Provenance: original

package symbol

// javaRandom is java.util.Random: the 48-bit linear congruential generator
// that USymbolCloud seeds with the size of the cloud, so that a cloud of a
// given size always has the same bubbles.
type javaRandom struct {
	seed int64
}

const (
	javaRandomMultiplier = 0x5DEECE66D
	javaRandomAddend     = 0xB
	javaRandomMask       = (1 << 48) - 1
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

// nextDouble is Random.nextDouble: 53 random bits scaled to [0, 1).
func (r *javaRandom) nextDouble() float64 {
	return float64(int64(r.next(26))<<27+int64(r.next(27))) * 0x1.0p-53
}
