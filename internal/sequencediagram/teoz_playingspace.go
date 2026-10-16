// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/PlayingSpace.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/PlayingSpaceWithParticipants.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"math"

	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// PlayingSpace holds the tiles of the diagram and its horizontal bounds.
type PlayingSpace struct {
	startingY     float64
	min           Real
	max           Real
	isShowFootbox bool

	tiles         []Tile
	livingSpaces  *LivingSpaces
	linkAnchors   []*LinkAnchor
	skinParam     klimt.ISkinParam
	diagram       *SequenceDiagram
	stringBounder font.StringBounder
}

// NewPlayingSpace builds the tiles of the events of the diagram.
func NewPlayingSpace(diagram *SequenceDiagram, dolls *Dolls, tileArguments *TileArguments) *PlayingSpace {
	s := &PlayingSpace{startingY: 8, diagram: diagram, livingSpaces: tileArguments.LivingSpaces(),
		linkAnchors: diagram.LinkAnchors(), skinParam: diagram.SkinParam(),
		stringBounder: tileArguments.StringBounder()}

	min2 := []Real{tileArguments.XOrigin()}
	max2 := []Real{tileArguments.XOrigin()}

	if dolls.Size() > 0 {
		min2 = append(min2, dolls.MinX(tileArguments.StringBounder()))
		max2 = append(max2, dolls.MaxX(tileArguments.StringBounder()))
	}

	// Gauge chain origin: startingY of headroom above the first tile
	ycurrent := YGaugeCreate(tileArguments.YOrigin().AddFixed(s.startingY), 0)

	s.tiles = append(s.tiles, tileBuildSeveral(&eventIterator{events: diagram.Events()}, tileArguments, nil,
		ycurrent)...)

	for _, tile := range s.tiles {
		min2 = append(min2, tile.MinX())
		max2 = append(max2, tile.MaxX())
	}

	for _, livingSpace := range s.livingSpaces.Values() {
		max2 = append(max2, livingSpace.PosD(tileArguments.StringBounder()))
		max2 = append(max2, livingSpace.PosC2(tileArguments.StringBounder()))
	}

	s.min = RealUtilsMin(min2...)
	s.max = RealUtilsMax(max2...)

	s.isShowFootbox = diagram.IsShowFootbox()
	return s
}

// DrawBackground draws the backgrounds of the tiles.
func (s *PlayingSpace) DrawBackground(ug klimt.UGraphic) {
	s.drawUInternal(NewUGraphicInterceptorTile(ug, true))
}

// DrawForeground draws the tiles.
func (s *PlayingSpace) DrawForeground(ug klimt.UGraphic) {
	s.drawUInternal(NewUGraphicInterceptorTile(ug, false))
}

func (s *PlayingSpace) drawUInternal(ug klimt.UGraphic) float64 {
	stringBounder := ug.StringBounder()
	var local, full []Tile
	fillPositionelTiles(stringBounder, s.tiles, &local, &full)
	// Each tile draws itself in absolute coordinates, translated by its own
	// gauge min
	for _, tile := range local {
		tile.DrawU(ug)
	}
	for _, linkAnchor := range s.linkAnchors {
		ytile1 := fromAnchor(full, linkAnchor.Anchor1)
		ytile2 := fromAnchor(full, linkAnchor.Anchor2)
		if ytile1 != nil && ytile2 != nil {
			linkAnchor.drawAnchor(ug, ytile1, ytile2, s.skinParam)
		}
	}
	// The last top-level tile's gauge dominates everything before it
	if len(s.tiles) == 0 {
		return s.startingY
	}
	return s.tiles[len(s.tiles)-1].YGauge().Max().CurrentValue()
}

func fromAnchor(positionedTiles []Tile, anchor string) Tile {
	for _, ytile := range positionedTiles {
		if ytile.MatchAnchor(anchor) {
			return ytile
		}
	}
	return nil
}

// PreferredHeight returns the height of the tiles; it runs the layout pass
// that positions them.
func (s *PlayingSpace) PreferredHeight(stringBounder font.StringBounder) float64 {
	limitFinder := klimt.LimitFinderCreate(stringBounder, true)
	interceptor := NewUGraphicInterceptorTile(limitFinder, false)
	finalY := s.drawUInternal(interceptor)
	return jmath.Max(limitFinder.MinMax().Dimension().Height(), finalY) + 10
}

// AddConstraints registers the constraints of the tiles, then keeps the
// footprints of parallel (&) siblings apart.
func (s *PlayingSpace) AddConstraints() {
	for _, tile := range s.tiles {
		tile.AddConstraints()
	}
	s.addParallelSiblingDisjointConstraints()
}

func (s *PlayingSpace) addParallelSiblingDisjointConstraints() {
	var cluster []Tile
	for _, tile := range s.tiles {
		if groupingIsParallel(tile) {
			for _, other := range cluster {
				s.ensureDisjoint(other, tile)
			}
			cluster = append(cluster, tile)
		} else {
			cluster = cluster[:0]
			cluster = append(cluster, tile)
		}
	}
}

// ensureDisjoint pushes whichever footprint starts further right so that
// it clears the other. The direction is read from plain living space
// positions, never from the cached minimum and maximum of a group.
func (s *PlayingSpace) ensureDisjoint(a, b Tile) {
	aMin := a.MinX()
	aMax := a.MaxX()
	bMin := b.MinX()
	bMax := b.MaxX()
	if aMin == nil || aMax == nil || bMin == nil || bMax == nil {
		return
	}
	aAnchor := s.findAnchorLivingSpace(a)
	bAnchor := s.findAnchorLivingSpace(b)
	if aAnchor == nil || bAnchor == nil {
		return
	}
	// Two parallel tiles sharing a participant are not independent
	if aAnchor == bAnchor {
		return
	}
	if aAnchor.PosB(s.stringBounder).CurrentValue() <= bAnchor.PosB(s.stringBounder).CurrentValue() {
		bMin.EnsureBiggerThan(aMax)
	} else {
		aMin.EnsureBiggerThan(bMax)
	}
}

func (s *PlayingSpace) findAnchorLivingSpace(tile Tile) *LivingSpace {
	switch event := tile.Event().(type) {
	case *Message:
		return s.livingSpaces.Get(event.Participant1())
	case *MessageExo:
		return s.livingSpaces.Get(event.Participant())
	}
	if g, ok := tile.(groupingTileLike); ok {
		for _, child := range g.grouping().tiles {
			if found := s.findAnchorLivingSpace(child); found != nil {
				return found
			}
		}
	}
	return nil
}

// MinX returns the left of the diagram.
func (s *PlayingSpace) MinX(font.StringBounder) Real { return s.min }

// MaxX returns the right of the diagram.
func (s *PlayingSpace) MaxX(font.StringBounder) Real { return s.max }

// IsShowFootbox reports whether the footboxes are drawn.
func (s *PlayingSpace) IsShowFootbox() bool { return s.isShowFootbox }

// LivingSpaces returns the columns.
func (s *PlayingSpace) LivingSpaces() *LivingSpaces { return s.livingSpaces }

// Tiles returns the top-level tiles.
func (s *PlayingSpace) Tiles() []Tile { return s.tiles }

// Diagram returns the diagram.
func (s *PlayingSpace) Diagram() *SequenceDiagram { return s.diagram }

// Border1 implements Bordered.
func (s *PlayingSpace) Border1() float64 { return s.min.CurrentValue() }

// Border2 implements Bordered.
func (s *PlayingSpace) Border2() float64 { return s.max.CurrentValue() }

// NewpageTiles returns the page breaks, nested ones included.
func (s *PlayingSpace) NewpageTiles() []*NewpageTile {
	var result []*NewpageTile
	for _, tile := range s.tiles {
		if g, ok := tile.(groupingTileLike); ok {
			g.grouping().addNewpageTiles(&result)
		}
		if n, ok := tile.(*NewpageTile); ok {
			result = append(result, n)
		}
	}
	return result
}

// YNewPages returns the y of the page breaks, between 0 and
// Double.MAX_VALUE.
func (s *PlayingSpace) YNewPages() []float64 {
	yNewPages := []float64{0.0}
	for _, newpageTile := range s.NewpageTiles() {
		yNewPages = append(yNewPages, newpageTile.YGauge().Min().CurrentValue())
	}
	return append(yNewPages, math.MaxFloat64)
}

// NbPages returns the number of pages.
func (s *PlayingSpace) NbPages() int { return len(s.YNewPages()) - 1 }

// PlayingSpaceWithParticipants draws a page of the diagram: the tiles, the
// lifelines and the participant heads and footboxes.
type PlayingSpaceWithParticipants struct {
	klimt.TextBlockDefaults
	klimt.TextBlockMemoized
	playingSpace *PlayingSpace
	pageIndex    int
}

// NewPlayingSpaceWithParticipants returns the drawing of playingSpace.
func NewPlayingSpaceWithParticipants(playingSpace *PlayingSpace) *PlayingSpaceWithParticipants {
	if playingSpace == nil {
		uml.ThrowNPE("")
	}
	return &PlayingSpaceWithParticipants{playingSpace: playingSpace}
}

// CalculateDimension returns the size of the current page.
func (p *PlayingSpaceWithParticipants) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return p.Dimension(stringBounder, p.calculateDimensionSlow)
}

func (p *PlayingSpaceWithParticipants) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	width := p.playingSpace.MaxX(stringBounder).CurrentValue() - p.playingSpace.MinX(stringBounder).CurrentValue()

	// PreferredHeight runs the layout pass that positions the tiles, so it
	// must be called before yMin and yMax
	fullHeight := p.playingSpace.PreferredHeight(stringBounder)
	pageHeight := p.yMax(fullHeight) - p.yMin()

	factor := 1.0
	if p.playingSpace.IsShowFootbox() {
		factor = 2
	}
	height := pageHeight + float64(factor*p.playingSpace.LivingSpaces().HeadHeight(stringBounder))
	return geom.NewXDimension2D(width, height)
}

// yMin and yMax rely on the solved gauges of the tiles: consecutive pages
// overlap on the page break, so that its dashed line is visible at the
// bottom of one page and the top of the next.
func (p *PlayingSpaceWithParticipants) yMin() float64 {
	if p.pageIndex == 0 {
		return 0
	}
	return p.playingSpace.NewpageTiles()[p.pageIndex-1].YGauge().Min().CurrentValue()
}

func (p *PlayingSpaceWithParticipants) yMax(fullHeight float64) float64 {
	newpages := p.playingSpace.NewpageTiles()
	if p.pageIndex >= len(newpages) {
		return fullHeight
	}
	newpage := newpages[p.pageIndex]
	return jmath.Min(newpage.YGauge().Min().CurrentValue()+newpage.PreferredHeight(), fullHeight)
}

// DrawU draws the current page.
func (p *PlayingSpaceWithParticipants) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()

	context := klimt.NewSimpleContext2D(false)
	livingSpaces := p.playingSpace.LivingSpaces()
	headHeight := livingSpaces.HeadHeight(stringBounder)

	// PreferredHeight runs the layout pass that positions the tiles
	fullHeight := p.playingSpace.PreferredHeight(stringBounder)
	ymin := p.yMin()
	ymax := p.yMax(fullHeight)
	pageHeight := ymax - ymin

	// The body is translated so that the current page starts right below the
	// heads, and clipped so that the content of the other pages stays hidden
	ugBody := ug.Apply(klimt.UTranslateDy(headHeight - ymin))
	if p.playingSpace.NbPages() > 1 {
		clip := klimt.NewUClip(-1000, ymin, math.MaxFloat64, pageHeight+1)
		ugBody = ugBody.Apply(clip)
	}

	p.playingSpace.DrawBackground(ugBody)
	// Lifelines and liveboxes use absolute positions over the whole diagram:
	// they are drawn with the full height, the clip trims them to the page
	livingSpaces.DrawLifeLines(ugBody, fullHeight, context)

	livingSpaces.DrawHeads(ug, context, geom.VerticalAlignmentBottom)
	if p.playingSpace.IsShowFootbox() {
		livingSpaces.DrawHeads(ug.Apply(klimt.UTranslateDy(pageHeight+headHeight)), context, geom.VerticalAlignmentTop)
	}

	p.playingSpace.DrawForeground(ugBody)
}

// Backcolor returns nil.
func (p *PlayingSpaceWithParticipants) Backcolor() color.HColor { return nil }

// MinX returns the left of the diagram.
func (p *PlayingSpaceWithParticipants) MinX(stringBounder font.StringBounder) Real {
	return p.playingSpace.MinX(stringBounder)
}

// NbPages returns the number of pages.
func (p *PlayingSpaceWithParticipants) NbPages() int { return p.playingSpace.NbPages() }

// SetIndex selects the page to draw.
func (p *PlayingSpaceWithParticipants) SetIndex(index int) {
	if index != p.pageIndex {
		p.pageIndex = index
		p.InvalidateDimensionCache()
	}
}
