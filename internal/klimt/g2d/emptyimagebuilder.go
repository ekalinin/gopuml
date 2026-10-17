// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/EmptyImageBuilder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/GraphvizUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package g2d

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// DefaultImageLimit is the largest width and height of an image when
// PLANTUML_LIMIT_SIZE is not set.
const DefaultImageLimit = 4096

var digits = regexp.MustCompile(`^[0-9]+$`)

// ImageLimit is upstream's GraphvizUtils.getenvImageLimit for the value of
// PLANTUML_LIMIT_SIZE: the value when it is a number, else 4096. As Java's
// Integer.parseInt, a number too large for an int is an error, which
// upstream reports as a crash.
func ImageLimit(value string) (int, error) {
	if value == "" || !digits.MatchString(value) {
		return DefaultImageLimit, nil
	}
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, &color.RuntimeError{Class: "java.lang.NumberFormatException",
			Message: "For input string: \"" + value + "\""}
	}
	return int(v), nil
}

// EmptyImageBuilder creates the image of a drawing, filled with its
// background, and its Graphics2D.
type EmptyImageBuilder struct {
	im *Image
	g  *graphics
}

// NewEmptyImageBuilder returns an image of width x height pixels, each
// limited to limit (upstream's PLANTUML_LIMIT_SIZE, logged and clipped),
// filled with background. The image is opaque (TYPE_INT_RGB) when the
// background is opaque, else it has alpha. watermark, when not empty, is
// written in light gray over the image. Width and height must be positive,
// else it panics as upstream throws IllegalArgumentException.
func NewEmptyImageBuilder(watermark string, width, height int, background color.XColor, limit int) *EmptyImageBuilder {
	if width <= 0 || height <= 0 {
		panic(&color.RuntimeError{Class: "java.lang.IllegalArgumentException",
			Message: "width and height must be positive"})
	}
	width = min(width, limit)
	height = min(height, limit)
	im := NewImage(width, height, background.Alpha() != 255)
	g := newGraphics(im)
	g.aa = true
	g.setColor(argb(background))
	g.fillRect(0, 0, width, height)
	b := &EmptyImageBuilder{im: im, g: g}
	if watermark != "" {
		g.setColor(0xFFC8C8C8) // XColor(200, 200, 200)
		b.printWatermark(watermark, width, height)
	}
	return b
}

// Image returns the image.
func (b *EmptyImageBuilder) Image() *Image { return b.im }

// The default font of a Graphics2D: Dialog, plain, 12.
var watermarkFont = font.NewAwtFont("Dialog", font.AwtPlain, 12)

func (b *EmptyImageBuilder) printWatermark(watermark string, maxWidth, maxHeight int) {
	height := int(watermarkFont.StringBounds(watermark).Height)
	width := int(b.g.stringWidth(watermark, watermarkFont))
	if height < 2 || width < 2 {
		return
	}
	if width <= maxWidth {
		for y := height; y < maxHeight; y += height + 1 {
			for x := 0; x < maxWidth; x += width + 10 {
				b.g.drawString(watermark, watermarkFont, float64(x), float64(y))
			}
		}
		return
	}
	withBreaks := b.withBreaks(watermark, maxWidth)
	y := 0
	for y < maxHeight {
		for _, s := range withBreaks {
			b.g.drawString(s, watermarkFont, 0, float64(y))
			y += int(watermarkFont.StringBounds(s).Height)
		}
		y += 10
	}
}

func (b *EmptyImageBuilder) withBreaks(watermark string, maxWidth int) []string {
	var result []string
	pending := ""
	for _, word := range strings.Fields(watermark) {
		candidate := word
		if pending != "" {
			candidate = pending + " " + word
		}
		if int(b.g.stringWidth(candidate, watermarkFont)) < maxWidth {
			pending = candidate
		} else {
			result = append(result, pending)
			pending = word
		}
	}
	if pending != "" {
		result = append(result, pending)
	}
	return result
}
