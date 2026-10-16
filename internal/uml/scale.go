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

package uml

import (
	"math"
	"strconv"
)

// ScaleKind is the class of a scale, upstream's Scale implementations.
type ScaleKind int

// Scale kinds.
const (
	ScaleSimple ScaleKind = iota
	ScaleWidth
	ScaleHeight
	ScaleMaxWidth
	ScaleMaxHeight
	ScaleWidthAndHeight
	ScaleMaxWidthAndHeight
)

var scaleKindNames = [...]string{"ScaleSimple", "ScaleWidth", "ScaleHeight", "ScaleMaxWidth", "ScaleMaxHeight",
	"ScaleWidthAndHeight", "ScaleMaxWidthAndHeight"}

// String returns the upstream class name, for example "ScaleSimple".
func (k ScaleKind) String() string { return scaleKindNames[k] }

// Scale is the "scale" setting of a diagram. A and B are the values of the
// command: the factor for ScaleSimple, the width or height for the
// one-value kinds, the width and height for the two-value kinds.
type Scale struct {
	Kind ScaleKind
	A, B float64
}

// String describes the scale, for example "ScaleSimple(1.5)".
func (s *Scale) String() string {
	f := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	switch s.Kind {
	case ScaleWidthAndHeight, ScaleMaxWidthAndHeight:
		return s.Kind.String() + "(" + f(s.A) + "," + f(s.B) + ")"
	}
	return s.Kind.String() + "(" + f(s.A) + ")"
}

// Factor returns the factor to apply to an image of the given size, as
// upstream's ScaleProtected.getScale: between 0 (excluded) and 4.
func (s *Scale) Factor(width, height float64) float64 {
	result := s.internal(width, height)
	if result <= 0 {
		return 1
	}
	if result > 4 {
		return 4
	}
	return result
}

func (s *Scale) internal(width, height float64) float64 {
	switch s.Kind {
	case ScaleWidth:
		return s.A / width
	case ScaleHeight:
		return s.A / height
	case ScaleMaxWidth:
		return min1(s.A / width)
	case ScaleMaxHeight:
		return min1(s.A / height)
	case ScaleWidthAndHeight:
		return math.Min(s.A/width, s.B/height)
	case ScaleMaxWidthAndHeight:
		return min1(math.Min(s.A/width, s.B/height))
	}
	return s.A
}

func min1(v float64) float64 {
	if v > 1 {
		return 1
	}
	return v
}
