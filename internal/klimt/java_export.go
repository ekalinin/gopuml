// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

// Exported forms of the Java emulation helpers of java.go, for the diagram
// packages built on klimt (svek).

// JavaD2I is Java's (int) cast of a double: truncation toward zero, NaN
// gives 0 and values out of range saturate.
func JavaD2I(x float64) int { return javaD2I(x) }

// JavaStringHashCode is Java's String.hashCode.
func JavaStringHashCode(s string) int32 { return javaStringHashCode(s) }

// JavaHashMapOrder returns keys, given in insertion order and without
// duplicates, in the iteration order of a java.util.HashMap or HashSet of
// strings created with the default capacity.
func JavaHashMapOrder(keys []string) []string { return javaHashMapOrder(keys) }

// JavaHashOrder returns the indexes of elements, given by their hashCode in
// insertion order and without duplicates, in the iteration order of a
// java.util.HashMap or HashSet created with the default capacity.
func JavaHashOrder(hashes []int32) []int { return javaHashOrder(hashes) }
