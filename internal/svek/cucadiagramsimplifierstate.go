// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/CucaDiagramSimplifierState.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// simplifyState is upstream's CucaDiagramSimplifierState.simplify: each
// autarkic group (a composite state with no transition crossing its
// border, or a concurrent region), deepest first, is laid out on its own
// and becomes a leaf drawn with that image, until none is left. A layout
// error stops it: the error is returned with the image that shows it, or
// the error of the context, without image.
func simplifyState(d *abel.CucaDiagram, stringBounder font.StringBounder, dotMode DotMode, le *layoutEngine,
	version dot.GraphvizVersion) (klimt.IEntityImage, error) {
	for changed := true; changed; {
		changed = false
		for _, g := range orderedGroups(d.RootGroup()) {
			if !g.IsAutarkic() {
				continue
			}
			img, err := newGroupMakerState(d, g, stringBounder, dotMode, le, version).image()
			if img == nil || err != nil {
				return img, err
			}
			leafType := abel.LeafTypeState
			if g.GroupType() == abel.GroupTypeConcurrentState {
				leafType = abel.LeafTypeStateConcurrent
			}
			g.OverrideImage(img, leafType)
			changed = true
		}
	}
	return nil, nil
}

// orderedGroups is upstream's getOrdered: the groups below root, level by
// level with the children of each group in reverse order, then the whole
// list reversed, so that the deepest groups come first.
func orderedGroups(root *abel.Entity) []*abel.Entity {
	ordered := []*abel.Entity{root}
	seen := map[*abel.Entity]bool{root: true}
	for {
		size := len(ordered)
		for _, g := range slices.Clone(ordered) {
			children := g.Groups()
			for i := len(children) - 1; i >= 0; i-- {
				if child := children[i]; !seen[child] {
					seen[child] = true
					ordered = append(ordered, child)
				}
			}
		}
		if size == len(ordered) {
			break
		}
	}
	var result []*abel.Entity
	for _, g := range slices.Backward(ordered) {
		if !g.IsRoot() {
			result = append(result, g)
		}
	}
	return result
}
