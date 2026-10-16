// SPDX-License-Identifier: MIT
// Provenance: original

// Package jmath is a deterministic, Java-compatible math library.
//
// The elementary functions (Sin, Cos, Tan, Asin, Acos, Atan, Atan2, Exp,
// Log, Pow, Cbrt) are a port of fdlibm 5.3, the library behind Java's
// StrictMath, and return results that are bit-identical to StrictMath on
// every operating system and architecture. The remaining functions follow
// the semantics of java.lang.Math (Round, Rint, Signum, Max, Min, ToRadians,
// ToDegrees, ...).
//
// Go may fuse a*b+c into a single fused multiply-add on some architectures
// (arm64, amd64 v3), which changes the last bits of a result. Every product
// that feeds an addition or a subtraction is therefore written float64(a*b):
// the Go specification guarantees that an explicit conversion rounds the
// product and prevents fusion. From package math only exact functions are
// used: bit access, Abs, Sqrt, Floor, Ceil, Copysign and IsNaN.
package jmath
