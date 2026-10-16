// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"cmp"
	"slices"
)

// Upstream iterates some java.util.HashMap instances where the order changes
// the result (the candidates of FunctionsSet.getFunctionSmart and the
// parentheses of TokenStack.guessFunctions). javaHashOrder reproduces
// HashMap's iteration order: entries are visited bucket by bucket, and in
// insertion order inside a bucket. The table starts with 16 buckets and
// doubles when it holds more than 3/4 of its capacity, or when a bucket of a
// table smaller than 64 grows beyond 8 entries. Tree bins (8 or more entries
// in one bucket of a table of 64 or more) are not reproduced.
type javaHashOrder struct {
	capacity int
	hashes   []int32 // spread hashes, in insertion order
}

func newJavaHashOrder() *javaHashOrder { return &javaHashOrder{capacity: 16} }

// javaSpread is HashMap.hash: the high half of the hash code folded into the
// low half.
func javaSpread(h int32) int32 { return h ^ int32(uint32(h)>>16) }

// add records a new key with the given hashCode. Keys must be distinct.
func (o *javaHashOrder) add(hashCode int32) {
	h := javaSpread(hashCode)
	mask := int32(o.capacity - 1)
	inBucket := 0
	for _, other := range o.hashes {
		if other&mask == h&mask {
			inBucket++
		}
	}
	o.hashes = append(o.hashes, h)
	if inBucket >= 8 && o.capacity < 64 {
		o.capacity *= 2
	}
	if len(o.hashes) > o.capacity*3/4 {
		o.capacity *= 2
	}
}

// order returns the insertion indexes in iteration order.
func (o *javaHashOrder) order() []int {
	idx := make([]int, len(o.hashes))
	for i := range idx {
		idx[i] = i
	}
	mask := int32(o.capacity - 1)
	slices.SortStableFunc(idx, func(a, b int) int {
		return cmp.Compare(o.hashes[a]&mask, o.hashes[b]&mask)
	})
	return idx
}
