// SPDX-License-Identifier: MIT
// Provenance: original

// Package g2d is the raster output driver: upstream's klimt.drawing.g2d
// (UGraphicG2d and its Driver*G2d) with EmptyImageBuilder, drawing on an
// ARGB image that internal/png writes as PNG.
//
// Upstream draws with Java2D. This package emulates the parts of a
// Graphics2D the drivers use, with the rules the reference JVM (OpenJDK 21,
// Marlin renderer, FreeType) applies, so that images are close to
// upstream's pixel for pixel:
//
//   - geometry: the paths of Java's shapes (rectangles, rounded
//     rectangles, ellipses, arcs, lines, GeneralPath with float32 points,
//     Batik's elliptical arcs), the parallelogram shortcuts for rectangle
//     fills and outlines and for lines, the snapping of strokes to pixel
//     centers, square caps, miter joins with a limit of 10, dashes;
//   - coverage: Marlin's sampling (8 sub-scanlines, 256 sub-columns) for
//     shapes and strokes, exact areas for parallelograms and glyphs, one
//     sample per pixel without antialiasing (graphics.go, raster.go);
//   - compositing: the 8-bit SrcOver arithmetic of Java2D's loops on
//     TYPE_INT_RGB and TYPE_INT_ARGB images (image.go);
//   - paints: GradientPaint, LinearGradientPaint (with Java's lookup
//     tables and tile by tile evaluation) and the stripe textures
//     (paint.go);
//   - text: glyph outlines of the embedded DejaVu fonts (subpackage
//     dejavu) as FreeType loads them unhinted, placed at whole pixels with
//     fractional advances, or filled as paths when the dpi scales the
//     drawing (text.go);
//   - shadows: the shape drawn without antialiasing and blurred by
//     ConvolveOp as Java's imaging library computes it (raw channels,
//     fixed point, copied borders);
//   - strokes: flattened subpaths, dashed (the pattern restarts with each
//     subpath, the last dash of a closed one continues into the first),
//     covered by a quadrilateral per segment, joins and caps, filled with
//     the non-zero rule (stroke.go).
//
// The package uses no third-party code. Products that feed additions are written float64(a*b) (design D17;
// fma_test.go checks that no fused multiply-add is emitted), so the images
// do not depend on the architecture.
//
// Not reproduced: TrueType hinting, which Java applies to text drawn before
// the first DriverTextG2d text and to the watermark (fractional metrics
// off); the exact rounding of Java's bilinear image scaling; Java's text
// layout for complex scripts (bidirectional text, shaping).
package g2d
