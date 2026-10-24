// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/ULayoutGroup.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PlacementStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/AbstractPlacementStrategy.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PlacementStrategyVisibility.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PlacementStrategyY1Y2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PlacementStrategyY1Y2Center.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PlacementStrategyY1Y2Left.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/geom/PlacementStrategyY1Y2Right.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"reflect"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// The layout helpers of upstream's klimt.geom package (ULayoutGroup and the
// placement strategies). They live here because they place text blocks,
// and klimt.geom cannot import this package.

// PlacedBlock is a block and its position, an entry of the LinkedHashMap
// that upstream's placement strategies return.
type PlacedBlock struct {
	Block    TextBlock
	Position geom.XPoint2D
}

// PlacementStrategy places the blocks added to it in an area. Upstream
// defines it in klimt.geom.
type PlacementStrategy interface {
	// Add adds a block.
	Add(block TextBlock)
	// Positions returns the position of each block in an area of the given
	// size, in the order the blocks were added.
	Positions(width, height float64) []PlacedBlock
}

// sameTextBlock tells whether two blocks are the same Java object: blocks
// of pointer types are compared by identity, blocks of other types (Java
// objects without a pointer counterpart) are always distinct.
func sameTextBlock(a, b TextBlock) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ta, tb := reflect.TypeOf(a), reflect.TypeOf(b)
	if ta.Kind() != reflect.Pointer || tb.Kind() != reflect.Pointer {
		return false
	}
	// Pointers to distinct zero-size values may be equal in Go.
	if ta.Elem().Size() == 0 {
		return false
	}
	return a == b
}

// placedDimension is an entry of AbstractPlacementStrategy's dimensions.
type placedDimension struct {
	block TextBlock
	dim   geom.XDimension2D
}

// AbstractPlacementStrategy is upstream's AbstractPlacementStrategy: the
// blocks with their dimensions, in insertion order (a LinkedHashMap keyed
// by the block).
type AbstractPlacementStrategy struct {
	stringBounder font.StringBounder
	dimensions    []placedDimension
}

// NewAbstractPlacementStrategy returns an empty strategy measuring with
// stringBounder.
func NewAbstractPlacementStrategy(stringBounder font.StringBounder) AbstractPlacementStrategy {
	return AbstractPlacementStrategy{stringBounder: stringBounder}
}

// Add records a block and its dimension; a block added again keeps its
// place and gets its new dimension, as LinkedHashMap.put does.
func (s *AbstractPlacementStrategy) Add(block TextBlock) {
	d := block.CalculateDimension(s.stringBounder)
	for i := range s.dimensions {
		if sameTextBlock(s.dimensions[i].block, block) {
			s.dimensions[i].dim = d
			return
		}
	}
	s.dimensions = append(s.dimensions, placedDimension{block: block, dim: d})
}

// Size returns the number of blocks.
func (s *AbstractPlacementStrategy) Size() int { return len(s.dimensions) }

// SumWidth returns the sum of the widths.
func (s *AbstractPlacementStrategy) SumWidth() float64 {
	result := 0.0
	for _, d := range s.dimensions {
		result += d.dim.Width()
	}
	return result
}

// SumHeight returns the sum of the heights.
func (s *AbstractPlacementStrategy) SumHeight() float64 {
	result := 0.0
	for _, d := range s.dimensions {
		result += d.dim.Height()
	}
	return result
}

// MaxWidth returns the largest width.
func (s *AbstractPlacementStrategy) MaxWidth() float64 {
	result := 0.0
	for _, d := range s.dimensions {
		result = math.Max(result, d.dim.Width())
	}
	return result
}

// MaxHeight returns the largest height.
func (s *AbstractPlacementStrategy) MaxHeight() float64 {
	result := 0.0
	for _, d := range s.dimensions {
		result = math.Max(result, d.dim.Height())
	}
	return result
}

// StringBounder returns the bounder of the strategy.
func (s *AbstractPlacementStrategy) StringBounder() font.StringBounder { return s.stringBounder }

// PlacementStrategyVisibility is upstream's PlacementStrategyVisibility:
// the blocks go by pairs, an icon at x=0 and a text at col2, each pair on
// its own row, centered vertically in the row.
type PlacementStrategyVisibility struct {
	AbstractPlacementStrategy
	col2 int
}

// NewPlacementStrategyVisibility returns the strategy with the texts at
// col2.
func NewPlacementStrategyVisibility(stringBounder font.StringBounder, col2 int) *PlacementStrategyVisibility {
	return &PlacementStrategyVisibility{AbstractPlacementStrategy: NewAbstractPlacementStrategy(stringBounder),
		col2: col2}
}

// Positions implements PlacementStrategy. An odd number of blocks panics,
// as upstream's iterator throws NoSuchElementException.
func (s *PlacementStrategyVisibility) Positions(width, height float64) []PlacedBlock {
	var result []PlacedBlock
	y := 0.0
	for i := 0; i < len(s.dimensions); i += 2 {
		if i+1 >= len(s.dimensions) {
			throw(noSuchElement, "")
		}
		ent1, ent2 := s.dimensions[i], s.dimensions[i+1]
		height1 := ent1.dim.Height()
		height2 := ent2.dim.Height()
		maxHeight12 := math.Max(height1, height2)
		result = putPlaced(result, ent1.block, geom.NewXPoint2D(0, 2+y+(maxHeight12-height1)/2))
		result = putPlaced(result, ent2.block, geom.NewXPoint2D(float64(s.col2), y+(maxHeight12-height2)/2))
		y += maxHeight12
	}
	return result
}

// putPlaced is LinkedHashMap.put on a list of placed blocks.
func putPlaced(result []PlacedBlock, block TextBlock, pos geom.XPoint2D) []PlacedBlock {
	for i := range result {
		if sameTextBlock(result[i].Block, block) {
			result[i].Position = pos
			return result
		}
	}
	return append(result, PlacedBlock{Block: block, Position: pos})
}

// placementY1Y2 is the shared layout of the Y1Y2 strategies: the blocks
// stacked with equal spaces around them, x given by xOf.
func (s *AbstractPlacementStrategy) placementY1Y2(width, height float64,
	xOf func(width float64, dim geom.XDimension2D) float64) []PlacedBlock {
	usedHeight := s.SumHeight()
	space := (height - usedHeight) / float64(len(s.dimensions)+1)
	var result []PlacedBlock
	y := space
	for _, ent := range s.dimensions {
		x := xOf(width, ent.dim)
		result = putPlaced(result, ent.block, geom.NewXPoint2D(x, y))
		y += ent.dim.Height() + space
	}
	return result
}

func centeredX(width float64, dim geom.XDimension2D) float64 { return (width - dim.Width()) / 2 }

// PlacementStrategyY1Y2 is upstream's PlacementStrategyY1Y2: the blocks
// stacked and centered horizontally.
type PlacementStrategyY1Y2 struct{ AbstractPlacementStrategy }

// NewPlacementStrategyY1Y2 returns the strategy.
func NewPlacementStrategyY1Y2(stringBounder font.StringBounder) *PlacementStrategyY1Y2 {
	return &PlacementStrategyY1Y2{NewAbstractPlacementStrategy(stringBounder)}
}

// Positions implements PlacementStrategy.
func (s *PlacementStrategyY1Y2) Positions(width, height float64) []PlacedBlock {
	return s.placementY1Y2(width, height, centeredX)
}

// PlacementStrategyY1Y2Center is upstream's PlacementStrategyY1Y2Center:
// the blocks stacked and centered horizontally.
type PlacementStrategyY1Y2Center struct{ AbstractPlacementStrategy }

// NewPlacementStrategyY1Y2Center returns the strategy.
func NewPlacementStrategyY1Y2Center(stringBounder font.StringBounder) *PlacementStrategyY1Y2Center {
	return &PlacementStrategyY1Y2Center{NewAbstractPlacementStrategy(stringBounder)}
}

// Positions implements PlacementStrategy.
func (s *PlacementStrategyY1Y2Center) Positions(width, height float64) []PlacedBlock {
	return s.placementY1Y2(width, height, centeredX)
}

// PlacementStrategyY1Y2Left is upstream's PlacementStrategyY1Y2Left: the
// blocks stacked at x=0.
type PlacementStrategyY1Y2Left struct{ AbstractPlacementStrategy }

// NewPlacementStrategyY1Y2Left returns the strategy.
func NewPlacementStrategyY1Y2Left(stringBounder font.StringBounder) *PlacementStrategyY1Y2Left {
	return &PlacementStrategyY1Y2Left{NewAbstractPlacementStrategy(stringBounder)}
}

// Positions implements PlacementStrategy.
func (s *PlacementStrategyY1Y2Left) Positions(width, height float64) []PlacedBlock {
	return s.placementY1Y2(width, height, func(float64, geom.XDimension2D) float64 { return 0 })
}

// PlacementStrategyY1Y2Right is upstream's PlacementStrategyY1Y2Right: the
// blocks stacked and aligned on the right.
type PlacementStrategyY1Y2Right struct{ AbstractPlacementStrategy }

// NewPlacementStrategyY1Y2Right returns the strategy.
func NewPlacementStrategyY1Y2Right(stringBounder font.StringBounder) *PlacementStrategyY1Y2Right {
	return &PlacementStrategyY1Y2Right{NewAbstractPlacementStrategy(stringBounder)}
}

// Positions implements PlacementStrategy.
func (s *PlacementStrategyY1Y2Right) Positions(width, height float64) []PlacedBlock {
	return s.placementY1Y2(width, height, func(width float64, dim geom.XDimension2D) float64 {
		return width - dim.Width()
	})
}

// ULayoutGroup is upstream's ULayoutGroup: blocks drawn where a placement
// strategy puts them.
type ULayoutGroup struct {
	placementStrategy PlacementStrategy
}

// NewULayoutGroup returns an empty group placed by placementStrategy.
func NewULayoutGroup(placementStrategy PlacementStrategy) *ULayoutGroup {
	return &ULayoutGroup{placementStrategy: placementStrategy}
}

// DrawU draws the blocks placed in an area of the given size.
func (g *ULayoutGroup) DrawU(ug UGraphic, width, height float64) {
	for _, ent := range g.placementStrategy.Positions(width, height) {
		ent.Block.DrawU(ug.Apply(UTranslatePoint(ent.Position)))
	}
}

// Add adds a block.
func (g *ULayoutGroup) Add(block TextBlock) { g.placementStrategy.Add(block) }

// InnerPosition returns the position of a member in the first block that
// has it, in an area of the given size; nil when no block has it.
func (g *ULayoutGroup) InnerPosition(member string, width, height float64,
	stringBounder font.StringBounder) *geom.XRectangle2D {
	for _, ent := range g.placementStrategy.Positions(width, height) {
		if result := ent.Block.InnerPosition(member, stringBounder); result != nil {
			r := UTranslatePoint(ent.Position).Apply(*result)
			return &r
		}
	}
	return nil
}
