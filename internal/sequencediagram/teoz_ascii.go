// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/Tile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/LivingSpace.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/LivingSpaces.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/PlayingSpaceWithParticipants.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/GroupingTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/PartitionTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/ElseTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/LifeEventTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/NewpageTile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/SequenceDiagramFileMakerTeoz.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"math"
	"reflect"
	"strings"

	"github.com/ekalinin/gopuml/internal/asciiverse"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The text output (ATXT, UTXT) of the Teoz layout. It has its own X line,
// with 1.0 for one character: each participant's box column is chained on
// an origin, the participants and the tiles add their constraints, and the
// line is solved once in PlayingSpaceWithParticipants.AsciiDraw. Rows are
// not solved: each tile is as high as its AsciiDimension, and the tiles
// are stacked (the members of a "&" run share their first row).

// teozClassPrefix is the Java package of the Teoz tiles, for the message of
// the UnsupportedOperationException of a tile without text drawing.
const teozClassPrefix = "net.sourceforge.plantuml.sequencediagram.teoz."

// AsciiAddConstraints adds the constraints of the tile on the text X line;
// none by default.
func (t *abstractTile) AsciiAddConstraints() {}

// AsciiMinX returns the left of the tile on the text X line, nil (none) by
// default.
func (t *abstractTile) AsciiMinX() Real { return nil }

// AsciiMaxX returns the right of the tile on the text X line, nil (none)
// by default.
func (t *abstractTile) AsciiMaxX() Real { return nil }

// AsciiDimension panics with UnsupportedOperationException naming the
// tile, as upstream's AsciiBlock default: the tiles without a text drawing
// (delays, dividers, notes over participants, references, "|||" spaces,
// notes above or below messages) make the text output fail.
func (t *abstractTile) AsciiDimension() asciiverse.ADimension2D {
	asciiverse.ThrowUnsupported(teozClassName(t.self))
	return asciiverse.ADimension2D{}
}

// AsciiDraw panics as AsciiDimension.
func (t *abstractTile) AsciiDraw(*asciiverse.InfinitePlan) {
	asciiverse.ThrowUnsupported(teozClassName(t.self))
}

// teozClassName returns the Java class name of a tile: the Go types of
// the tiles carry upstream's class names.
func teozClassName(tile Tile) string {
	typ := reflect.TypeOf(tile)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return teozClassPrefix + typ.Name()
}

// asciiColumn is Java's (int) cast of a position of the text X line.
func asciiColumn(r Real) int {
	x := r.CurrentValue()
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

// livingSpaceASCII holds the positions of a living space on the text X
// line: posB is the left column of the head box (as wide as the code of
// the participant plus 2), posC the lifeline, posD the column after the
// box.
type livingSpaceASCII struct {
	posB, posC, posD          Real
	marginBefore, marginAfter int
}

func (s *LivingSpace) asciiWidth() int { return s.p.AsciiDimension().Width() }

// SetAsciiPosB sets the left column of the head box.
func (s *LivingSpace) SetAsciiPosB(asciiPosB Real) { s.ascii.posB = asciiPosB }

// AsciiPosB returns the left column of the head box.
func (s *LivingSpace) AsciiPosB() Real { return s.ascii.posB }

// AsciiPosC returns the lifeline column, the middle of the box.
func (s *LivingSpace) AsciiPosC() Real {
	if s.ascii.posC == nil {
		s.ascii.posC = s.ascii.posB.AddFixed(float64(s.asciiWidth() / 2))
	}
	return s.ascii.posC
}

// AsciiPosD returns the column after the box.
func (s *LivingSpace) AsciiPosD() Real {
	if s.ascii.posD == nil {
		s.ascii.posD = s.ascii.posB.AddFixed(float64(s.asciiWidth()))
	}
	return s.ascii.posD
}

// EnsureAsciiMarginBefore widens the text space before the box.
func (s *LivingSpace) EnsureAsciiMarginBefore(margin int) {
	if margin < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	s.ascii.marginBefore = max(s.ascii.marginBefore, margin)
}

// EnsureAsciiMarginAfter widens the text space after the box.
func (s *LivingSpace) EnsureAsciiMarginAfter(margin int) {
	if margin < 0 {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	s.ascii.marginAfter = max(s.ascii.marginAfter, margin)
}

// AsciiPosA returns the left of the text space of the participant.
func (s *LivingSpace) AsciiPosA() Real { return s.AsciiPosB().AddFixed(float64(-s.ascii.marginBefore)) }

// AsciiPosE returns the right of the text space of the participant.
func (s *LivingSpace) AsciiPosE() Real { return s.AsciiPosD().AddFixed(float64(s.ascii.marginAfter)) }

// AsciiLifeColumn returns the lifeline column.
func (s *LivingSpace) AsciiLifeColumn() Real { return s.AsciiPosC() }

// AsciiLeftColumn returns the left column of the head box.
func (s *LivingSpace) AsciiLeftColumn() Real { return s.AsciiPosB() }

// asciiGap is the number of characters between neighbouring participants.
const asciiGap = 1

// AsciiAddConstraints keeps neighbouring participants asciiGap characters
// apart on the text X line.
func (l *LivingSpaces) AsciiAddConstraints() {
	var previous *LivingSpace
	for _, current := range l.Values() {
		if previous != nil {
			point1 := previous.AsciiPosE()
			point2 := current.AsciiPosA()
			point2.EnsureBiggerThan(point1.AddFixed(asciiGap))
		}
		previous = current
	}
}

// AsciiDimension is upstream's placeholder size of the text drawing of
// PlayingSpaceWithParticipants: AsciiBlock's default, which throws.
func (p *PlayingSpaceWithParticipants) AsciiDimension() asciiverse.ADimension2D {
	asciiverse.ThrowUnsupported(teozClassPrefix + "PlayingSpaceWithParticipants")
	return asciiverse.ADimension2D{}
}

// AsciiDraw draws the diagram as text: the heads, the tiles (which draw
// their messages, notes and frames), the footboxes, then the lifelines in
// the cells left empty. Every page is drawn: the text output has no pages.
func (p *PlayingSpaceWithParticipants) AsciiDraw(plan *asciiverse.InfinitePlan) {
	livingSpaces := p.playingSpace.LivingSpaces()
	participants := append([]*Participant(nil), livingSpaces.Participants()...)
	if len(participants) == 0 {
		return
	}

	// X layout: each box column is chained on the origin of a new line,
	// then the constraints are added and the line is solved once
	asciiXOrigin := RealUtilsCreateOrigin()
	asciiCurrent := asciiXOrigin.AddAtLeast(0)
	for _, participant := range participants {
		ls := livingSpaces.Get(participant)
		ls.SetAsciiPosB(asciiCurrent)
		asciiCurrent = asciiCurrent.AddAtLeast(float64(participant.AsciiDimension().Width()))
	}
	livingSpaces.AsciiAddConstraints()
	for _, tile := range p.playingSpace.Tiles() {
		tile.AsciiAddConstraints()
	}
	p.addAsciiParallelSiblingDisjointConstraints()
	asciiXOrigin.CompileNow()

	// Heads, at the left column of each box
	for _, participant := range participants {
		participant.AsciiDraw(plan.Move(asciiColumn(livingSpaces.Get(participant).AsciiLeftColumn()), 0))
	}

	const headBottom = 2
	y := headBottom + 1

	// Tiles, stacked as computeAsciiLayout says. A tile without a text
	// drawing panics here, as upstream throws
	topTiles := p.playingSpace.Tiles()
	layout := computeAsciiLayout(topTiles)
	for i, tile := range topTiles {
		tile.AsciiDraw(plan.Move(0, y+layout.rowOf[i]))
	}
	y += layout.totalHeight
	bottomY := y

	// Footboxes, right below the lifelines
	if p.playingSpace.IsShowFootbox() {
		for _, participant := range participants {
			participant.AsciiDraw(plan.Move(asciiColumn(livingSpaces.Get(participant).AsciiLeftColumn()), bottomY))
		}
	}

	// Lifelines from the heads to the last tile, only in empty cells
	vLine := plan.VLineChar()
	for _, participant := range participants {
		lx := asciiColumn(livingSpaces.Get(participant).AsciiLifeColumn())
		for yy := headBottom + 1; yy < bottomY; yy++ {
			if plan.CharAt(lx, yy) == ' ' {
				plan.Move(lx, yy).DrawChar(vLine)
			}
		}
		plan.Move(lx, headBottom).MuteLifelineBelowBox()
		plan.Move(lx, bottomY).MuteLifelineAboveBox()
	}
}

// addAsciiParallelSiblingDisjointConstraints keeps the members of a "&" run
// side by side on the text X line, as PlayingSpace does on the pixel one.
func (p *PlayingSpaceWithParticipants) addAsciiParallelSiblingDisjointConstraints() {
	var cluster []Tile
	for _, tile := range p.playingSpace.Tiles() {
		if groupingIsParallel(tile) {
			for _, other := range cluster {
				p.ensureAsciiDisjoint(other, tile)
			}
			cluster = append(cluster, tile)
		} else {
			cluster = append(cluster[:0], tile)
		}
	}
}

func (p *PlayingSpaceWithParticipants) ensureAsciiDisjoint(a, b Tile) {
	aMin := a.AsciiMinX()
	aMax := a.AsciiMaxX()
	bMin := b.AsciiMinX()
	bMax := b.AsciiMaxX()
	if aMin == nil || aMax == nil || bMin == nil || bMax == nil {
		return
	}
	// The direction comes from the box columns of the anchors (read before
	// the line is solved, as upstream); tiles sharing an anchor are skipped
	aAnchor := p.findAnchorLivingSpace(a)
	bAnchor := p.findAnchorLivingSpace(b)
	if aAnchor == nil || bAnchor == nil || aAnchor == bAnchor {
		return
	}
	if aAnchor.AsciiPosB().CurrentValue() <= bAnchor.AsciiPosB().CurrentValue() {
		bMin.EnsureBiggerThan(aMax)
	} else {
		aMin.EnsureBiggerThan(bMax)
	}
}

func (p *PlayingSpaceWithParticipants) findAnchorLivingSpace(tile Tile) *LivingSpace {
	livingSpaces := p.playingSpace.LivingSpaces()
	switch event := tile.Event().(type) {
	case *Message:
		return livingSpaces.Get(event.Participant1())
	case *MessageExo:
		return livingSpaces.Get(event.Participant())
	}
	if g, ok := tile.(groupingTileLike); ok {
		for _, child := range g.grouping().tiles {
			if found := p.findAnchorLivingSpace(child); found != nil {
				return found
			}
		}
	}
	return nil
}

// asciiLayout gives the row of each tile of a list, relative to the first
// one, and the rows they take.
type asciiLayout struct {
	rowOf       []int
	totalHeight int
}

// computeAsciiLayout stacks tiles: the members of a "&" run (with the life
// events around them) start on the same row and take the rows of the
// highest one; other tiles follow each other.
func computeAsciiLayout(tiles []Tile) asciiLayout {
	rowOf := make([]int, len(tiles))
	y := 0
	var pending []int
	clusterOpen := false
	for i, tile := range tiles {
		if groupingIsParallel(tile) {
			clusterOpen = true
			pending = append(pending, i)
		} else if _, ok := tile.(*LifeEventTile); ok {
			pending = append(pending, i)
		} else {
			y += flushAsciiPending(tiles, &pending, clusterOpen, rowOf, y)
			clusterOpen = false
			pending = append(pending, i)
		}
	}
	y += flushAsciiPending(tiles, &pending, clusterOpen, rowOf, y)
	return asciiLayout{rowOf: rowOf, totalHeight: y}
}

func flushAsciiPending(tiles []Tile, pending *[]int, wasCluster bool, rowOf []int, y int) int {
	if len(*pending) == 0 {
		return 0
	}
	var height int
	if wasCluster {
		maxHeight := 0
		for _, idx := range *pending {
			maxHeight = max(maxHeight, tiles[idx].AsciiDimension().Height())
		}
		for _, idx := range *pending {
			rowOf[idx] = y
		}
		height = maxHeight
	} else {
		localY := y
		for _, idx := range *pending {
			rowOf[idx] = localY
			localY += tiles[idx].AsciiDimension().Height()
		}
		height = localY - y
	}
	*pending = (*pending)[:0]
	return height
}

// The frame of a group is asciiFrameMargin characters outside the columns
// its children touch; its footer is one row.
const (
	asciiFrameMargin = 2
	asciiFooterRows  = 1
)

// asciiHeaderRows returns the rows of the frame header, as AGroupFrame
// computes them.
func (g *GroupingTile) asciiHeaderRows() int {
	cols := g.hooks.asciiFrameColumns()
	width := cols[1] - cols[0] + 1
	title, ok := g.hooks.asciiTitle()
	return asciiverse.NewAGroupFrame(asciiverse.NewADimension2D(width, 0), title, ok,
		g.hooks.asciiFrameHasTab()).BodyRowOffset()
}

// asciiFrameHasTab reports whether the frame has a title tab (every group
// but a partition).
func (g *GroupingTile) asciiFrameHasTab() bool { return true }

// AsciiAddConstraints adds the constraints of the children.
func (g *GroupingTile) AsciiAddConstraints() {
	for _, tile := range g.tiles {
		tile.AsciiAddConstraints()
	}
}

func (g *GroupingTile) asciiChildrenMin() Real {
	var mins []Real
	for _, tile := range g.tiles {
		if m := tile.AsciiMinX(); m != nil {
			mins = append(mins, m)
		}
	}
	if len(mins) == 0 {
		return nil
	}
	return RealUtilsMin(mins...)
}

func (g *GroupingTile) asciiChildrenMax() Real {
	var maxs []Real
	for _, tile := range g.tiles {
		if m := tile.AsciiMaxX(); m != nil {
			maxs = append(maxs, m)
		}
	}
	if len(maxs) == 0 {
		return nil
	}
	return RealUtilsMax(maxs...)
}

// AsciiMinX returns the left of the children with the frame margin, nil
// when no child has any.
func (g *GroupingTile) AsciiMinX() Real {
	childMin := g.asciiChildrenMin()
	if childMin == nil {
		return nil
	}
	return childMin.AddFixed(-asciiFrameMargin)
}

// AsciiMaxX returns the right of the children with the frame margin, nil
// when no child has any.
func (g *GroupingTile) AsciiMaxX() Real {
	childMax := g.asciiChildrenMax()
	if childMax == nil {
		return nil
	}
	return childMax.AddFixed(asciiFrameMargin)
}

// asciiFrameColumns returns the left and right columns of the frame; a
// group without positioned child spans every head box.
func (g *GroupingTile) asciiFrameColumns() [2]int {
	minX := g.hooks.AsciiMinX()
	maxX := g.hooks.AsciiMaxX()
	if minX != nil && maxX != nil {
		return [2]int{asciiColumn(minX), asciiColumn(maxX)}
	}
	lo := math.MaxInt32
	hi := math.MinInt32
	for _, ls := range g.tileArguments.LivingSpaces().Values() {
		// posD is the column after the box
		lo = min(lo, asciiColumn(ls.AsciiPosB()))
		hi = max(hi, asciiColumn(ls.AsciiPosD())-1)
	}
	if lo == math.MaxInt32 {
		return [2]int{0, 0}
	}
	return [2]int{lo - asciiFrameMargin, hi + asciiFrameMargin}
}

func (g *GroupingTile) asciiBodyHeight() int { return computeAsciiLayout(g.tiles).totalHeight }

// AsciiDimension returns the size of the frame: header, children, footer.
func (g *GroupingTile) AsciiDimension() asciiverse.ADimension2D {
	cols := g.hooks.asciiFrameColumns()
	width := cols[1] - cols[0] + 1
	height := g.asciiHeaderRows() + g.asciiBodyHeight() + asciiFooterRows
	return asciiverse.NewADimension2D(width, height)
}

// AsciiDraw draws the frame at its columns, then the children below the
// header; the group draws the separator of each "else" across the frame.
// The children place themselves on their own columns.
func (g *GroupingTile) AsciiDraw(plan *asciiverse.InfinitePlan) {
	cols := g.hooks.asciiFrameColumns()
	left := cols[0]
	width := cols[1] - cols[0] + 1

	title, ok := g.hooks.asciiTitle()
	frame := asciiverse.NewAGroupFrame(g.hooks.AsciiDimension(), title, ok, g.hooks.asciiFrameHasTab())
	frame.AsciiDraw(plan.Move(left, 0))

	layout := computeAsciiLayout(g.tiles)
	bodyStart := frame.BodyRowOffset()
	for i, tile := range g.tiles {
		y := bodyStart + layout.rowOf[i]
		if elseTile, isElse := tile.(*ElseTile); isElse {
			asciiverse.NewAElseSeparator(width, elseTile.AsciiLabel(), true).AsciiDraw(plan.Move(left, y))
		} else {
			tile.AsciiDraw(plan.Move(0, y))
		}
	}
}

// asciiTitle returns the lines of the group's display joined by spaces (a
// null line reads "null", as Java's StringBuilder.append).
func (g *GroupingTile) asciiTitle() (string, bool) {
	var sb strings.Builder
	for i := 0; i < g.display.Size(); i++ {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		if cs := g.display.Get(i); cs == nil {
			sb.WriteString("null")
		} else {
			sb.WriteString(cs.String())
		}
	}
	return sb.String(), true
}

// asciiTitle returns the comment of the partition only, without the
// "partition" keyword; it may be null.
func (p *PartitionTile) asciiTitle() (string, bool) { return p.GroupingStart().Comment() }

// asciiFrameHasTab reports false: a partition's title is on the top border.
func (p *PartitionTile) asciiFrameHasTab() bool { return false }

// asciiFrameColumns spans every head box with the frame margin, widened
// to the children (the notes attached to messages inside the partition).
func (p *PartitionTile) asciiFrameColumns() [2]int {
	left := math.MaxInt32
	right := math.MinInt32
	for _, ls := range p.TileArguments().LivingSpaces().Values() {
		// posD is the column after the box
		left = min(left, asciiColumn(ls.AsciiPosB()))
		right = max(right, asciiColumn(ls.AsciiPosD())-1)
	}
	left -= asciiFrameMargin
	right += asciiFrameMargin

	if childMin := p.AsciiMinX(); childMin != nil {
		left = min(left, asciiColumn(childMin))
	}
	if childMax := p.AsciiMaxX(); childMax != nil {
		right = max(right, asciiColumn(childMax))
	}
	return [2]int{left, right}
}

// AsciiDimension returns the row of the separator, as wide as the label.
func (t *ElseTile) AsciiDimension() asciiverse.ADimension2D {
	return asciiverse.NewADimension2D(uml.JavaLength(t.AsciiLabel()), 1)
}

// AsciiDraw draws nothing: the group draws the separator across its frame.
func (t *ElseTile) AsciiDraw(*asciiverse.InfinitePlan) {}

// AsciiLabel returns the comment of the "else", empty for none.
func (t *ElseTile) AsciiLabel() string {
	comment, _ := t.anElse.Comment()
	return comment
}

// isDestroy reports a "destroy": activations are not drawn in text, only
// destructions, as an "X" on the lifeline.
func (t *LifeEventTile) isDestroy() bool { return t.lifeEvent.Type() == LifeEventDestroy }

// AsciiDimension returns one row for a destruction, nothing otherwise.
func (t *LifeEventTile) AsciiDimension() asciiverse.ADimension2D {
	if t.isDestroy() {
		return asciiverse.NewADimension2D(0, 1)
	}
	return asciiverse.NewADimension2D(0, 0)
}

// AsciiDraw draws the "X" of a destruction on the lifeline.
func (t *LifeEventTile) AsciiDraw(plan *asciiverse.InfinitePlan) {
	if !t.isDestroy() {
		return
	}
	x := asciiColumn(t.livingSpace.AsciiPosC())
	plan.Move(x, 0).DrawString("X")
}

// AsciiMinX returns the lifeline column.
func (t *LifeEventTile) AsciiMinX() Real { return t.livingSpace.AsciiPosC() }

// AsciiMaxX returns the lifeline column.
func (t *LifeEventTile) AsciiMaxX() Real { return t.livingSpace.AsciiPosC() }

// newpageASCIIMargin is how far the page break line goes past the boxes.
const newpageASCIIMargin = 2

// asciiColumns returns the columns of the page break line: every head box
// with the margin.
func (t *NewpageTile) asciiColumns() [2]int {
	lo := math.MaxInt32
	hi := math.MinInt32
	for _, ls := range t.tileArguments.LivingSpaces().Values() {
		lo = min(lo, asciiColumn(ls.AsciiPosB()))
		hi = max(hi, asciiColumn(ls.AsciiPosD())-1)
	}
	if lo == math.MaxInt32 {
		return [2]int{0, 0}
	}
	return [2]int{lo - newpageASCIIMargin, hi + newpageASCIIMargin}
}

// AsciiDimension returns one row across the columns.
func (t *NewpageTile) AsciiDimension() asciiverse.ADimension2D {
	cols := t.asciiColumns()
	return asciiverse.NewADimension2D(cols[1]-cols[0]+1, 1)
}

// AsciiDraw draws a dotted line across the columns.
func (t *NewpageTile) AsciiDraw(plan *asciiverse.InfinitePlan) {
	cols := t.asciiColumns()
	width := cols[1] - cols[0] + 1
	if width > 0 {
		plan.Move(cols[0], 0).DrawHLineDotted(0, width-1, 0, true)
	}
}

// AsciiBlock returns the text drawing of the diagram. As upstream, num is
// ignored (every page is drawn) and the size is a fixed 20 by 20, which
// only the centering of the title reads.
func (m *SequenceDiagramFileMakerTeoz) AsciiBlock(int) asciiverse.AsciiBlock {
	return teozASCIIPage{body: m.body}
}

// teozASCIIPage is the anonymous AsciiBlock of upstream's getAsciiBlock.
type teozASCIIPage struct {
	body *PlayingSpaceWithParticipants
}

func (p teozASCIIPage) AsciiDimension() asciiverse.ADimension2D {
	return asciiverse.NewADimension2D(20, 20)
}

func (p teozASCIIPage) AsciiDraw(plan *asciiverse.InfinitePlan) { p.body.AsciiDraw(plan) }
