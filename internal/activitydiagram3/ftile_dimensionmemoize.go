// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/cond/FtileDimensionMemoize.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import "github.com/ekalinin/plantuml-in-go/internal/klimt/font"

// ftileDimensionMemoizeHook is upstream's abstract
// calculateDimensionInternalSlow of FtileDimensionMemoize.
type ftileDimensionMemoizeHook interface {
	calculateDimensionInternalSlow(stringBounder font.StringBounder) FtileGeometry
}

// FtileDimensionMemoize caches the inner geometry of a tile; the tile
// embeds it, calls initFtile with itself and defines
// calculateDimensionInternalSlow.
type FtileDimensionMemoize struct {
	AbstractFtile
	dimensionInternal *FtileGeometry
}

func (f *FtileDimensionMemoize) calculateDimensionInternal(stringBounder font.StringBounder) FtileGeometry {
	if f.dimensionInternal == nil {
		g := f.self.(ftileDimensionMemoizeHook).calculateDimensionInternalSlow(stringBounder)
		f.dimensionInternal = &g
	}
	return *f.dimensionInternal
}

func (f *FtileDimensionMemoize) clearCacheDimensionInternal() { f.dimensionInternal = nil }
