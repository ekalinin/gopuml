// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

func TestPlacementStrategies(t *testing.T) {
	sb := fakeBounder{}
	a := TextBlockUtilsEmpty(10, 5)
	b := TextBlockUtilsEmpty(20, 15)
	for _, c := range []struct {
		name string
		s    PlacementStrategy
		want []geom.XPoint2D
	}{
		{"Y1Y2", NewPlacementStrategyY1Y2(sb), []geom.XPoint2D{{X: 15, Y: 10}, {X: 10, Y: 25}}},
		{"Center", NewPlacementStrategyY1Y2Center(sb), []geom.XPoint2D{{X: 15, Y: 10}, {X: 10, Y: 25}}},
		{"Left", NewPlacementStrategyY1Y2Left(sb), []geom.XPoint2D{{X: 0, Y: 10}, {X: 0, Y: 25}}},
		{"Right", NewPlacementStrategyY1Y2Right(sb), []geom.XPoint2D{{X: 30, Y: 10}, {X: 20, Y: 25}}},
		{"Visibility", NewPlacementStrategyVisibility(sb, 7), []geom.XPoint2D{{X: 0, Y: 7}, {X: 7, Y: 0}}},
	} {
		c.s.Add(a)
		c.s.Add(b)
		got := c.s.Positions(40, 50)
		if len(got) != len(c.want) {
			t.Fatalf("%s: %d positions", c.name, len(got))
		}
		for i, p := range got {
			if p.Position != c.want[i] {
				t.Errorf("%s: block %d at %v, want %v", c.name, i, p.Position, c.want[i])
			}
		}
	}
	// The same block added twice keeps one place, as LinkedHashMap.put.
	s := NewPlacementStrategyY1Y2Left(sb)
	blk := NewTextBlockVertical([]TextBlock{a, b}, geom.HorizontalAlignmentLeft)
	s.Add(blk)
	s.Add(a)
	s.Add(blk)
	if n := len(s.Positions(10, 10)); n != 2 {
		t.Errorf("duplicate block: %d positions, want 2", n)
	}
	// An odd number of blocks panics in the visibility strategy.
	v := NewPlacementStrategyVisibility(sb, 3)
	v.Add(a)
	if class, _ := expectPanic(t, func() { v.Positions(1, 1) }); class != noSuchElement {
		t.Errorf("odd visibility blocks: %s", class)
	}
}
