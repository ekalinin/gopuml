// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/ImageData.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/api/ImageDataAbstract.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/api/ImageDataSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/api/ImageDataComplex.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package export

import (
	"math"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ImageData describes a generated image: its size in pixels, its status
// and, for PNG images with links, an HTML image map.
type ImageData interface {
	Width() int
	Height() int
	// ContainsCMapData reports whether the image has an image map.
	ContainsCMapData() bool
	// CMapData returns the image map with nameID as id and name; it panics
	// when there is none, as upstream throws UnsupportedOperationException.
	CMapData(nameID string) string
	// WarningOrError returns the warning or error of the diagram, "" for
	// none.
	WarningOrError() string
	Status() int
	// RootCause returns the error of a failed render, nil for none.
	RootCause() error
}

// imageDataAbstract is the size and status of an image.
type imageDataAbstract struct {
	width, height int
	status        int
}

func newImageDataAbstractDim(dim geom.XDimension2D) imageDataAbstract {
	return imageDataAbstract{width: javaD2I(dim.Width()), height: javaD2I(dim.Height())}
}

func (d *imageDataAbstract) Width() int  { return d.width }
func (d *imageDataAbstract) Height() int { return d.height }
func (d *imageDataAbstract) Status() int { return d.status }

// SetStatus sets the status.
func (d *imageDataAbstract) SetStatus(status int) { d.status = status }

// ImageDataSimple is the data of an image without image map.
type ImageDataSimple struct {
	imageDataAbstract
	rootCause error
}

var _ ImageData = (*ImageDataSimple)(nil)

// NewImageDataSimple returns the data of an image of a size.
func NewImageDataSimple(width, height int) *ImageDataSimple {
	return &ImageDataSimple{imageDataAbstract: imageDataAbstract{width: width, height: height}}
}

// NewImageDataSimpleDim returns the data of an image of a dimension,
// truncated to integers, with a status.
func NewImageDataSimpleDim(dim geom.XDimension2D, status int) *ImageDataSimple {
	d := &ImageDataSimple{imageDataAbstract: newImageDataAbstractDim(dim)}
	d.status = status
	return d
}

// ImageDataOK returns an empty image with status 0.
func ImageDataOK() *ImageDataSimple { return NewImageDataSimple(0, 0) }

// ImageDataError returns an empty image with status 503 and the error that
// caused it (nil for none).
func ImageDataError(rootCause error) *ImageDataSimple {
	d := NewImageDataSimple(0, 0)
	d.status = 503
	d.rootCause = rootCause
	return d
}

// ContainsCMapData returns false.
func (d *ImageDataSimple) ContainsCMapData() bool { return false }

// CMapData panics: there is no image map.
func (d *ImageDataSimple) CMapData(string) string {
	panic(unsupportedOperationError())
}

// WarningOrError returns "".
func (d *ImageDataSimple) WarningOrError() string { return "" }

// RootCause returns the error of a failed render.
func (d *ImageDataSimple) RootCause() error { return d.rootCause }

// ImageDataComplex is the data of a PNG image with an image map of its
// links.
type ImageDataComplex struct {
	imageDataAbstract
	cmap           *klimt.CMapData
	warningOrError string
}

var _ ImageData = (*ImageDataComplex)(nil)

// NewImageDataComplex returns the data of an image with an image map (nil
// for none), a warning or error ("" for none) and a status.
func NewImageDataComplex(dim geom.XDimension2D, cmap *klimt.CMapData, warningOrError string,
	status int) *ImageDataComplex {
	d := &ImageDataComplex{imageDataAbstract: newImageDataAbstractDim(dim), cmap: cmap,
		warningOrError: warningOrError}
	d.status = status
	return d
}

// ContainsCMapData reports whether the image map has links.
func (d *ImageDataComplex) ContainsCMapData() bool { return d.cmap != nil && d.cmap.ContainsData() }

// CMapData returns the image map.
func (d *ImageDataComplex) CMapData(nameID string) string { return d.cmap.AsString(nameID) }

// WarningOrError returns the warning or error of the diagram.
func (d *ImageDataComplex) WarningOrError() string { return d.warningOrError }

// RootCause returns nil.
func (d *ImageDataComplex) RootCause() error { return nil }

// javaD2I is Java's (int) cast of a double: truncation toward zero, NaN to
// zero and saturation at the int range.
func javaD2I(x float64) int {
	switch {
	case math.IsNaN(x):
		return 0
	case x >= math.MaxInt32:
		return math.MaxInt32
	case x <= math.MinInt32:
		return math.MinInt32
	}
	return int(x)
}
