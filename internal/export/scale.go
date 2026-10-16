// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/Scale.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleProtected.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleWidth.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleHeight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleMaxWidth.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleMaxHeight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleWidthAndHeight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/ScaleMaxWidthAndHeight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package export

import "github.com/ekalinin/plantuml-in-go/internal/jmath"

// Scale is the "scale" of a diagram: the factor applied to an image of a
// given size.
type Scale interface {
	Scale(width, height float64) float64
}

// scaleProtected is upstream's ScaleProtected: the factor of internal,
// replaced by 1 when it is not positive and capped at 4.
type scaleProtected struct {
	internal func(width, height float64) float64
}

func (s scaleProtected) Scale(width, height float64) float64 {
	result := s.internal(width, height)
	if result <= 0 {
		return 1
	}
	if result > 4 {
		return 4
	}
	return result
}

// ScaleSimple is a fixed factor ("scale 1.5").
func ScaleSimple(scale float64) Scale {
	return scaleProtected{func(float64, float64) float64 { return scale }}
}

// ScaleWidth scales to a width ("scale 200 width").
func ScaleWidth(maxWidth float64) Scale {
	return scaleProtected{func(width, _ float64) float64 { return maxWidth / width }}
}

// ScaleHeight scales to a height ("scale 200 height").
func ScaleHeight(maxHeight float64) Scale {
	return scaleProtected{func(_, height float64) float64 { return maxHeight / height }}
}

// ScaleMaxWidth scales down to a maximum width ("scale max 200 width").
func ScaleMaxWidth(maxWidth float64) Scale {
	return scaleProtected{func(width, _ float64) float64 {
		result := maxWidth / width
		if result > 1 {
			return 1
		}
		return result
	}}
}

// ScaleMaxHeight scales down to a maximum height ("scale max 200 height").
func ScaleMaxHeight(maxHeight float64) Scale {
	return scaleProtected{func(_, height float64) float64 {
		result := maxHeight / height
		if result > 1 {
			return 1
		}
		return result
	}}
}

// ScaleWidthAndHeight scales to fit a box ("scale 200*100").
func ScaleWidthAndHeight(maxWidth, maxHeight float64) Scale {
	return scaleProtected{func(width, height float64) float64 {
		return jmath.Min(maxWidth/width, maxHeight/height)
	}}
}

// ScaleMaxWidthAndHeight scales down to fit a box ("scale max 200*100").
func ScaleMaxWidthAndHeight(maxWidth, maxHeight float64) Scale {
	return scaleProtected{func(width, height float64) float64 {
		m := jmath.Min(maxWidth/width, maxHeight/height)
		if m > 1 {
			return 1
		}
		return m
	}}
}
