// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/MutableImage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/awt/PortableImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "github.com/ekalinin/plantuml-in-go/internal/klimt/color"

// PortableImage is a raster image, upstream's klimt.awt.PortableImage
// without the java.awt accessors (getBufferedImage, createGraphics,
// getGraphics), which belong to the raster driver.
type PortableImage interface {
	Width() int
	Height() int
	// RGB returns the ARGB value of a pixel, as BufferedImage.getRGB.
	RGB(x, y int) int32
	SetRGB(x, y int, rgb int32)
	// ToPngDataURL returns the image as a "data:image/png;base64," URL.
	ToPngDataURL() string
	// Transparency returns java.awt.Transparency's OPAQUE (1), BITMASK (2)
	// or TRANSLUCENT (3).
	Transparency() int
	Subimage(x, y, width, height int) PortableImage
	Scale(scaleFactor float64, interpolationType AffineTransformType) PortableImage
}

// The image types of upstream's PortableImage.
const (
	PortableImageTypeIntRGB     = 1
	PortableImageTypeIntARGB    = 2
	PortableImageTypeIntARGBPre = 3
)

// MutableImage is an image that can be scaled and recolored, each change
// giving a new image.
type MutableImage interface {
	Image() PortableImage
	WithScale(scale float64) MutableImage
	MuteColor(newColor color.XColor) MutableImage
	MuteTransparentColor(newColor color.XColor) MutableImage
	Monochrome() MutableImage
	Scale() float64
}
