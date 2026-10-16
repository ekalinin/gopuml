// SPDX-License-Identifier: MIT
// Provenance: original

// Package geom holds the geometry value types of the drawing model: points,
// dimensions, rectangles, lines, cubic curves, bounding boxes and the small
// alignment and position enums.
//
// It is a port of upstream's klimt/geom package. Floating point results are
// meant to be bit-identical to Java: formulas keep upstream's operation
// order, and every product that feeds an addition is written as
// float64(a*b) so that the compiler cannot fuse it into a multiply-add, which
// Java never does.
//
// Upstream exceptions become panics. Java methods that may return null return
// an extra boolean instead.
//
// Names follow upstream's: a Java constructor becomes NewT, a static method
// T.m becomes the function TM (for example XLine2D.line is XLine2DLine and
// BezierUtils.getAngle is BezierUtilsAngle), getters lose their "get" prefix,
// and public final fields become exported fields.
package geom
