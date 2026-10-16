// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/GroupingTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/PartitionTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/Blotter.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/jmath"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// The margins of a fragment frame.
const (
	groupingExternalMarginX1 = 3
	groupingExternalMarginX2 = 9
	// groupingExternalMarginY is the vertical space around a frame, on each
	// side: reserved in PreferredHeight and applied at draw time (frameY),
	// never by offsetting the gauge min, which stays the chaining point of
	// parallel siblings.
	groupingExternalMarginY = 4
	groupingMarginX         = 16
	groupingMarginYMagic    = 20
)

// groupingTileLike is a GroupingTile or a PartitionTile (upstream's
// "instanceof GroupingTile").
type groupingTileLike interface {
	Tile
	grouping() *GroupingTile
}

// groupingTileHooks are the methods PartitionTile overrides.
type groupingTileHooks interface {
	groupingTileLike
	component(stringBounder font.StringBounder) klimt.Component
	area(stringBounder font.StringBounder) *klimt.Area
	minCurrentValueForDrawing() float64
	drawCompBackground(ug klimt.UGraphic, area *klimt.Area, back color.HColor, round float64)
}

// GroupingTile is a fragment ("alt", "loop", "group"...) and its children.
type GroupingTile struct {
	abstractTile
	hooks  groupingTileHooks
	tiles  []Tile
	min    Real
	max    Real
	start  *GroupingStart
	end    *GroupingLeaf
	yGauge *YGauge

	skin      *rose.Rose
	skinParam klimt.ISkinParam
	display   *klimt.Display

	bodyHeight    float64
	tileArguments *TileArguments
}

// Event implements Tile.
func (g *GroupingTile) Event() Event { return g.start }

// ContactPointRelative implements Tile.
func (g *GroupingTile) ContactPointRelative() float64 { return 0 }

// YGauge implements Tile.
func (g *GroupingTile) YGauge() *YGauge { return g.yGauge }

func (g *GroupingTile) grouping() *GroupingTile { return g }

func newGroupingTile(it *eventIterator, start *GroupingStart, tileArgumentsBackColorChanged,
	tileArgumentsOriginal *TileArguments, currentY *YGauge) *GroupingTile {
	g := &GroupingTile{}
	g.initGroupingTile(g, it, start, tileArgumentsBackColorChanged, tileArgumentsOriginal, currentY)
	return g
}

// initGroupingTile is upstream's constructor; hooks is the tile itself, a
// GroupingTile or a PartitionTile, whose component is used from the start.
func (g *GroupingTile) initGroupingTile(hooks groupingTileHooks, it *eventIterator, start *GroupingStart,
	tileArgumentsBackColorChanged, tileArgumentsOriginal *TileArguments, currentY *YGauge) {
	g.hooks = hooks
	g.initTile(hooks, tileArgumentsBackColorChanged.StringBounder())
	g.tileArguments = tileArgumentsOriginal
	stringBounder := tileArgumentsOriginal.StringBounder()
	g.start = start
	if start.Title() == "group" {
		g.display = displayCreateOpt(optString(start.Comment()))
	} else {
		g.display = displayCreateOpt(someString(start.Title()), optString(start.Comment()))
	}
	g.skin = tileArgumentsOriginal.Skin()
	g.skinParam = tileArgumentsBackColorChanged.SkinParam()

	var min2, max2 []Real
	var allElses []Tile
	dim1 := g.preferredDimensionIfEmpty(stringBounder)

	// A group following "& opt" starts at the previous tile's min (top
	// aligned); previousMax is kept so that the outer chain is pushed past
	// the taller of the two members
	previousMax := currentY.Max()
	parallel := start.IsParallel()
	// The gauge min is the chaining point, never the drawn frame's top (see
	// frameY)
	firstY := previousMax
	if parallel {
		firstY = currentY.Min()
	}

	// The body starts below the frame's top margin and the header
	h := dim1.Height() + groupingMarginYMagic/2 + groupingExternalMarginY
	currentY = YGaugeCreate(firstY.AddAtLeast(h), 0)

	for it.hasNext() {
		ev := it.next()
		if leaf, ok := ev.(*GroupingLeaf); ok && leaf.Type() == GroupingEnd {
			g.end = leaf
			break
		}
		for _, tile := range tileBuildOne(it, tileArgumentsOriginal, ev, hooks, currentY) {
			g.tiles = append(g.tiles, tile)
			currentY = tile.YGauge()
		}
	}

	// bodyHeight counts each parallel (&) cluster once
	g.bodyHeight = computeBodyHeight(g.tiles)

	for _, tile := range g.tiles {
		ev := tile.Event()
		if leaf, ok := ev.(*GroupingLeaf); ok && leaf.Type() == GroupingElse {
			allElses = append(allElses, tile)
			continue
		}
		// A "|||" has no horizontal position of its own (upstream's issue
		// #2789)
		if _, ok := ev.(*HSpace); ok {
			continue
		}
		// The drawn extent of a self message loop, not its layout
		// reservation (upstream's issue #2788)
		min2 = append(min2, tile.DrawnMinX().AddFixed(-groupingMarginX))
		m := tile.DrawnMaxX()
		max2 = append(max2, m.AddFixed(groupingMarginX))
	}
	width := dim1.Width()
	if len(min2) == 0 {
		min2 = append(min2, tileArgumentsOriginal.XOrigin())
	}

	g.min = RealUtilsMin(min2...)
	for _, anElse := range allElses {
		max2 = append(max2, anElse.MaxX())
	}

	max2 = append(max2, g.min.AddFixed(width+16))
	g.max = RealUtilsMax(max2...)
	if parallel {
		// Like YGaugeCreateParallel: the group's own max still dominates the
		// previous sibling's max
		parallelMax := firstY.AddAtLeast(g.PreferredHeight())
		parallelMax.EnsureBiggerThan(previousMax)
		g.yGauge = NewYGauge(firstY, parallelMax)
	} else {
		g.yGauge = NewYGauge(firstY, firstY.AddAtLeast(g.PreferredHeight()))
	}
}

// frameY returns the absolute y of the drawn frame's top edge.
func (g *GroupingTile) frameY() float64 {
	return g.YGauge().Min().CurrentValue() + groupingExternalMarginY
}

func (g *GroupingTile) component(font.StringBounder) klimt.Component {
	return g.skin.CreateComponent(g.start.UsedStyles(), klimt.ComponentTypeGroupingHeaderTeoz, nil, g.skinParam,
		g.display)
}

func (g *GroupingTile) preferredDimensionIfEmpty(stringBounder font.StringBounder) geom.XDimension2D {
	return g.hooks.component(stringBounder).PreferredDimension(stringBounder)
}

// DrawU draws the background, the frame, the notes of the end, then the
// children.
func (g *GroupingTile) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	area := g.hooks.area(stringBounder)
	comp := g.hooks.component(stringBounder)

	if context2D(ug).IsBackground() {
		g.drawBackground(ug, area)
	}

	// The frame is drawn groupingExternalMarginY below the gauge min
	comp.DrawU(ug.Apply(klimt.NewUTranslate(g.hooks.minCurrentValueForDrawing(), g.frameY())), area, context2D(ug))
	g.drawNotes(ug.Apply(klimt.UTranslateDy(g.frameY())))

	// Children draw themselves in absolute coordinates
	for _, tile := range g.tiles {
		tile.DrawU(ug)
	}
}

func (g *GroupingTile) area(stringBounder font.StringBounder) *klimt.Area {
	return klimt.AreaCreate(g.max.CurrentValue()-g.min.CurrentValue(), g.totalHeight(stringBounder))
}

func (g *GroupingTile) minCurrentValueForDrawing() float64 { return g.min.CurrentValue() }

// frameMinX returns the left of the frame itself.
func (g *GroupingTile) frameMinX() Real { return g.min }

// frameMaxX returns the right of the frame itself.
func (g *GroupingTile) frameMaxX() Real { return g.max }

func (g *GroupingTile) drawBackground(ug klimt.UGraphic, area *klimt.Area) {
	style := g.start.UsedStyles()[0]
	back := style.Value(klimt.PNameBackGroundColor).AsColor(g.skinParam.IHtmlColorSet())
	round := style.Value(klimt.PNameRoundCorner).AsDouble()
	// ug arrives untranslated: the Blotter draws its bands from a local
	// y=0, so it is translated to the frame top
	g.hooks.drawCompBackground(ug.Apply(klimt.UTranslateDy(g.frameY())), area, back, round)
}

func (g *GroupingTile) drawCompBackground(ug klimt.UGraphic, area *klimt.Area, back color.HColor, round float64) {
	dimensionToUse := area.DimensionToUse()
	blotter := NewBlotter(dimensionToUse, back, round)

	for _, tile := range g.tiles {
		if elseTile, ok := tile.(*ElseTile); ok {
			// The color boundary lands where ElseTile.DrawU draws its divider
			ypos := elseTile.YGauge().Min().CurrentValue() - g.frameY()
			backElse := elseTile.BackColorGeneral()
			// An else section without its own color inherits the group
			// background, which may come from the style
			if backElse == nil {
				backElse = back
			}
			blotter.AddChange(ypos+1, backElse)
		}
	}

	blotter.CloseChanges()
	blotter.DrawU(ug.Apply(klimt.UTranslateDx(g.min.CurrentValue())))
}

func (g *GroupingTile) totalHeight(stringBounder font.StringBounder) float64 {
	dimIfEmpty := g.preferredDimensionIfEmpty(stringBounder)
	return g.bodyHeight + dimIfEmpty.Height() + groupingMarginYMagic/2
}

// PreferredHeight implements Tile: the frame and its two margins.
func (g *GroupingTile) PreferredHeight() float64 {
	dim1 := g.preferredDimensionIfEmpty(g.StringBounder())
	return dim1.Height() + g.bodyHeight + groupingMarginYMagic + 2*groupingExternalMarginY
}

// AddConstraints registers the constraints of the children, then keeps the
// participants around the fragment clear of its frame.
func (g *GroupingTile) AddConstraints() {
	for _, tile := range g.tiles {
		tile.AddConstraints()
	}
	g.ensureFollowingParticipantClearsFrame()
	g.ensurePrecedingParticipantClearsFrame()
}

// ensureFollowingParticipantClearsFrame pushes the participant after the
// rightmost one touched by the fragment past the frame (upstream's issue
// #2788). It never constrains from a cached RealMax or RealMin: only plain
// living space positions and the whitelisted tile edges of stableMaxX.
func (g *GroupingTile) ensureFollowingParticipantClearsFrame() {
	livingSpaces := g.tileArguments.LivingSpaces()
	touched := map[*LivingSpace]bool{}
	collectTouchedLivingSpaces(g.tiles, livingSpaces, touched)

	rightmost := rightmostOf(livingSpaces, touched)
	if rightmost == nil {
		return
	}
	next := livingSpaces.Next(rightmost)
	if next == nil {
		return
	}

	nextPosA := next.PosA(g.StringBounder())
	frameMargin := float64(groupingMarginX + groupingExternalMarginX2)

	nextPosA.EnsureBiggerThan(rightmost.PosC(g.StringBounder()).AddFixed(frameMargin))

	g.ensureClearsFrameOf(g, nextPosA, frameMargin, livingSpaces)
}

// ensurePrecedingParticipantClearsFrame is the mirror for the left edge:
// the tiles' own left edges are pushed past the previous participant.
func (g *GroupingTile) ensurePrecedingParticipantClearsFrame() {
	livingSpaces := g.tileArguments.LivingSpaces()
	touched := map[*LivingSpace]bool{}
	collectTouchedLivingSpaces(g.tiles, livingSpaces, touched)

	leftmost := leftmostOf(livingSpaces, touched)
	if leftmost == nil {
		return
	}
	previous := livingSpaces.Previous(leftmost)
	if previous == nil {
		return
	}

	previousPosC := previous.PosC(g.StringBounder())
	frameMargin := float64(groupingMarginX + groupingExternalMarginX2)

	leftmost.PosC(g.StringBounder()).EnsureBiggerThan(previousPosC.AddFixed(frameMargin))

	g.ensureClearsFrameOfLeft(g, previousPosC, frameMargin)
}

func (g *GroupingTile) ensureClearsFrameOfLeft(group *GroupingTile, previousPosC Real, margin float64) {
	for _, tile := range group.tiles {
		if nested, ok := tile.(groupingTileLike); ok {
			notes := nested.grouping().notesWidth(g.StringBounder(), NoteLeft)
			g.ensureClearsFrameOfLeft(nested.grouping(), previousPosC,
				margin+groupingMarginX+groupingExternalMarginX2+notes)
			continue
		}
		for _, minX := range stableMinX(tile) {
			minX.EnsureBiggerThan(previousPosC.AddFixed(margin))
		}
	}
}

// stableMaxX returns the right edges of a tile that are safe to constrain
// (plain delta chains over living space positions, never a cached
// RealMax).
func stableMaxX(tile Tile) []Real {
	switch t := tile.(type) {
	case *NoteTile:
		return t.stableMaxX()
	case *CommunicationTileSelf, *CommunicationTileNoteRight, *CommunicationTileSelfNoteRight, *LifeEventTile,
		*ReferenceTile, *DividerTile:
		return []Real{tile.DrawnMaxX()}
	}
	return nil
}

// stableMinX is the left-side mirror of stableMaxX.
func stableMinX(tile Tile) []Real {
	switch t := tile.(type) {
	case *NoteTile:
		return t.stableMinX()
	case *CommunicationTileSelf, *CommunicationTileNoteLeft, *CommunicationTileSelfNoteLeft, *LifeEventTile,
		*ReferenceTile:
		return []Real{tile.DrawnMinX()}
	}
	return nil
}

// ensureClearsFrameOf pushes nextPosA past every right edge the frame of
// group is built on, margin beyond each.
func (g *GroupingTile) ensureClearsFrameOf(group *GroupingTile, nextPosA Real, margin float64,
	livingSpaces *LivingSpaces) {
	g.ensureClearsTitleOf(group, nextPosA, margin, livingSpaces)

	for _, tile := range group.tiles {
		if nested, ok := tile.(groupingTileLike); ok {
			notes := nested.grouping().notesWidth(g.StringBounder(), NoteRight)
			g.ensureClearsFrameOf(nested.grouping(), nextPosA, margin+groupingMarginX+groupingExternalMarginX2+notes,
				livingSpaces)
			continue
		}
		for _, maxX := range stableMaxX(tile) {
			nextPosA.EnsureBiggerThan(maxX.AddFixed(margin))
		}
	}
}

// ensureClearsTitleOf pushes nextPosA past the right of a frame stretched
// by its title, re-derived from the leftmost touched participant.
func (g *GroupingTile) ensureClearsTitleOf(group *GroupingTile, nextPosA Real, margin float64,
	livingSpaces *LivingSpaces) {
	touched := map[*LivingSpace]bool{}
	collectTouchedLivingSpaces(group.tiles, livingSpaces, touched)

	var leftmost *LivingSpace
	for _, candidate := range livingSpaces.Values() {
		if touched[candidate] {
			leftmost = candidate
			break
		}
	}
	if leftmost == nil {
		return
	}

	overhang := 0.0
	for _, tile := range group.tiles {
		if nt, ok := tile.(*NoteTile); ok && nt.leftAnchor() == leftmost {
			overhang = jmath.Max(overhang, nt.leftOverhang())
		}
	}

	width := group.preferredDimensionIfEmpty(g.StringBounder()).Width()
	nextPosA.EnsureBiggerThan(leftmost.PosC(g.StringBounder()).AddFixed(
		width + 16 - groupingMarginX - groupingMarginX - overhang + margin))
}

func addTouched(touched map[*LivingSpace]bool, livingSpaces *LivingSpaces, participant *Participant) {
	if participant == nil {
		return
	}
	if livingSpace := livingSpaces.Get(participant); livingSpace != nil {
		touched[livingSpace] = true
	}
}

// addNoteTouched marks the participants of a note; a "note across" (no
// participant) spans every participant.
func addNoteTouched(touched map[*LivingSpace]bool, livingSpaces *LivingSpaces, note *Note) {
	if note.Participant() == nil && note.Participant2() == nil {
		for _, ls := range livingSpaces.Values() {
			touched[ls] = true
		}
		return
	}
	addTouched(touched, livingSpaces, note.Participant())
	addTouched(touched, livingSpaces, note.Participant2())
}

func rightmostOf(livingSpaces *LivingSpaces, touched map[*LivingSpace]bool) *LivingSpace {
	var rightmost *LivingSpace
	for _, candidate := range livingSpaces.Values() {
		if touched[candidate] {
			rightmost = candidate
		}
	}
	return rightmost
}

func leftmostOf(livingSpaces *LivingSpaces, touched map[*LivingSpace]bool) *LivingSpace {
	for _, candidate := range livingSpaces.Values() {
		if touched[candidate] {
			return candidate
		}
	}
	return nil
}

// collectTouchedLivingSpaces finds every column touched by the tiles and
// their nested fragments.
func collectTouchedLivingSpaces(tiles []Tile, livingSpaces *LivingSpaces, touched map[*LivingSpace]bool) {
	for _, tile := range tiles {
		switch ev := tile.Event().(type) {
		case *Message:
			touched[livingSpaces.Get(ev.Participant1())] = true
			touched[livingSpaces.Get(ev.Participant2())] = true
		case *MessageExo:
			touched[livingSpaces.Get(ev.Participant())] = true
		case *LifeEvent:
			touched[livingSpaces.Get(ev.Participant())] = true
		case *Note:
			addNoteTouched(touched, livingSpaces, ev)
		case *Notes:
			for _, note := range ev.List() {
				addNoteTouched(touched, livingSpaces, note)
			}
		case *Reference:
			for _, participant := range ev.Participants() {
				addTouched(touched, livingSpaces, participant)
			}
		}
		if g, ok := tile.(groupingTileLike); ok {
			collectTouchedLivingSpaces(g.grouping().tiles, livingSpaces, touched)
		}
	}
}

// MinX implements Tile: the frame, its margin and the notes on its left.
func (g *GroupingTile) MinX() Real {
	return g.min.AddFixed(-groupingExternalMarginX1 - g.notesWidth(g.StringBounder(), NoteLeft))
}

// MaxX implements Tile: the frame, its margin and the notes on its right.
func (g *GroupingTile) MaxX() Real {
	return g.max.AddFixed(groupingExternalMarginX2 + g.notesWidth(g.StringBounder(), NoteRight))
}

// drawNotes draws the notes of the end at the top corners of the frame.
func (g *GroupingTile) drawNotes(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()

	xRight := g.max.CurrentValue()
	for _, note := range g.noteComponents(NoteRight) {
		dimNote := note.PreferredDimension(stringBounder)
		note.DrawU(ug.Apply(klimt.UTranslateDx(xRight)), klimt.AreaCreate(dimNote.Width(), dimNote.Height()),
			context2D(ug))
		xRight += dimNote.Width()
	}

	xLeft := g.min.CurrentValue()
	for _, note := range g.noteComponents(NoteLeft) {
		dimNote := note.PreferredDimension(stringBounder)
		note.DrawU(ug.Apply(klimt.UTranslateDx(xLeft-dimNote.Width())),
			klimt.AreaCreate(dimNote.Width(), dimNote.Height()), context2D(ug))
		xLeft -= dimNote.Width()
	}
}

func (g *GroupingTile) notesWidth(stringBounder font.StringBounder, position NotePosition) float64 {
	result := 0.0
	for _, note := range g.noteComponents(position) {
		result += note.PreferredDimension(stringBounder).Width()
	}
	return result
}

func (g *GroupingTile) noteComponents(position NotePosition) []klimt.Component {
	if g.end == nil {
		return nil
	}
	var result []klimt.Component
	for _, noteOnMessage := range g.end.NoteOnMessages() {
		// A "note left" is drawn on the left side of the frame, everything
		// else on the right
		isLeft := noteOnMessage.Position() == NoteLeft
		if isLeft != (position == NoteLeft) {
			continue
		}
		sk := noteOnMessage.SkinParamBackcolored(g.skinParam)
		result = append(result, g.skin.CreateComponentNote(noteOnMessage.UsedStyles(),
			noteOnMessage.NoteStyle().ComponentType(), sk, noteOnMessage.Display(), noteOnMessage.Colors()))
	}
	return result
}

// fillPositionelTiles runs OnGaugeResolved on the tiles and their nested
// tiles (livebox steps, creations and destructions), and lists the tiles.
func fillPositionelTiles(stringBounder font.StringBounder, tiles []Tile, local, full *[]Tile) {
	for _, tile := range tiles {
		tile.OnGaugeResolved()
		*local = append(*local, tile)
		*full = append(*full, tile)
		if g, ok := tile.(groupingTileLike); ok {
			var local2 []Tile
			fillPositionelTiles(stringBounder, g.grouping().tiles, &local2, full)
		}
	}
}

// computeBodyHeight sums the heights of the tiles, counting each run of
// parallel (&) tiles once: tallest before the contact line plus tallest
// after it. A life event is held with the pending run, so that a later
// "&" pulls it into the cluster; any other tile flushes the run.
func computeBodyHeight(tiles []Tile) float64 {
	total := 0.0
	var pending []Tile
	clusterOpen := false

	for _, tile := range tiles {
		if groupingIsParallel(tile) {
			clusterOpen = true
			pending = append(pending, tile)
		} else if _, ok := tile.(*LifeEventTile); ok {
			pending = append(pending, tile)
		} else {
			total += flushPending(&pending, clusterOpen)
			clusterOpen = false
			pending = append(pending, tile)
		}
	}
	total += flushPending(&pending, clusterOpen)
	return total
}

func flushPending(pending *[]Tile, wasCluster bool) float64 {
	if len(*pending) == 0 {
		return 0
	}
	result := 0.0
	if wasCluster {
		contact := 0.0
		after := 0.0
		for _, tile := range *pending {
			contact = jmath.Max(contact, tile.ContactPointRelative())
			after = jmath.Max(after, tile.ZZZ())
		}
		result = contact + after
	} else {
		for _, tile := range *pending {
			result += tile.PreferredHeight()
		}
	}
	*pending = (*pending)[:0]
	return result
}

// groupingIsParallel is upstream's GroupingTile.isParallel.
func groupingIsParallel(tile Tile) bool { return tile.Event().IsParallel() }

func (g *GroupingTile) addNewpageTiles(newpages *[]*NewpageTile) {
	for _, tile := range g.tiles {
		if nested, ok := tile.(groupingTileLike); ok {
			nested.grouping().addNewpageTiles(newpages)
		}
		if n, ok := tile.(*NewpageTile); ok {
			*newpages = append(*newpages, n)
		}
	}
}

// GroupingStart returns the start of the fragment.
func (g *GroupingTile) GroupingStart() *GroupingStart { return g.start }

// TileArguments returns the arguments of the children.
func (g *GroupingTile) TileArguments() *TileArguments { return g.tileArguments }

// SkinParam returns the skin parameters of the fragment.
func (g *GroupingTile) SkinParam() klimt.ISkinParam { return g.skinParam }

// partitionTitleVPad is the space above and below the title of a
// partition (upstream's issue #2679).
const partitionTitleVPad = 4

// PartitionTile is a "partition": a fragment across the whole diagram,
// with its title over the frame.
type PartitionTile struct {
	GroupingTile
}

func newPartitionTile(it *eventIterator, start *GroupingStart, tileArgumentsBackColorChanged,
	tileArgumentsOriginal *TileArguments, currentY *YGauge) *PartitionTile {
	p := &PartitionTile{}
	p.initGroupingTile(p, it, start, tileArgumentsBackColorChanged, tileArgumentsOriginal, currentY)
	return p
}

// titleBlock is the title, with the font of the group header style.
func (p *PartitionTile) titleBlock() klimt.TextBlock {
	styleHeader := p.GroupingStart().UsedStyles()[1]
	display := displayCreateOpt(optString(p.GroupingStart().Comment()))
	return display.Create(styleHeader.FontConfiguration(p.SkinParam().IHtmlColorSet(), nil),
		geom.HorizontalAlignmentLeft, p.SkinParam())
}

func (p *PartitionTile) component(font.StringBounder) klimt.Component {
	return &partitionComponent{p: p}
}

func (p *PartitionTile) area(stringBounder font.StringBounder) *klimt.Area {
	return klimt.AreaCreate(p.width(), p.totalHeight(stringBounder))
}

func (p *PartitionTile) width() float64 { return p.border2() - p.border1() }

func (p *PartitionTile) border2() float64 { return p.TileArguments().Border2() }

func (p *PartitionTile) border1() float64 { return p.TileArguments().Border1() }

func (p *PartitionTile) minCurrentValueForDrawing() float64 { return 0 }

func (p *PartitionTile) drawCompBackground(ug klimt.UGraphic, area *klimt.Area, back color.HColor, round float64) {
	rect := klimt.URectangleBuildDim(area.DimensionToUse()).Rounded(round)
	ug.Apply(klimt.UTranslateDx(p.border1())).Apply(klimt.Bg(back)).Draw(rect)
}

// partitionComponent is the anonymous component of upstream's
// PartitionTile.getComponent: the frame and the title.
type partitionComponent struct {
	p *PartitionTile
}

func (c *partitionComponent) StyleSignature() klimt.StyleSignature {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (c *partitionComponent) UsedStyles() []*klimt.Style {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (c *partitionComponent) PreferredWidth(font.StringBounder) float64 {
	uml.Throw(uml.UnsupportedOperationException, "")
	return 0
}

func (c *partitionComponent) PreferredHeight(font.StringBounder) float64 {
	uml.Throw(uml.UnsupportedOperationException, "")
	return 0
}

// PreferredDimension reserves the title and its padding.
func (c *partitionComponent) PreferredDimension(stringBounder font.StringBounder) geom.XDimension2D {
	titleDim := c.p.titleBlock().CalculateDimension(stringBounder)
	return geom.NewXDimension2D(titleDim.Width(), titleDim.Height()+2*partitionTitleVPad)
}

// DrawU draws the title centered over the frame, and the frame with the
// stroke of the group style.
func (c *partitionComponent) DrawU(ug klimt.UGraphic, area *klimt.Area, _ klimt.Context2D) {
	stringBounder := ug.StringBounder()
	styleFrame := c.p.GroupingStart().UsedStyles()[0]
	title := c.p.titleBlock()
	border1 := c.p.border1()
	delta := (c.p.width() - title.CalculateDimension(stringBounder).Width()) / 2

	title.DrawU(ug.Apply(klimt.NewUTranslate(border1+delta, partitionTitleVPad)))

	round := styleFrame.Value(klimt.PNameRoundCorner).AsDouble()
	rect := klimt.URectangleBuildDim(area.DimensionToUse()).Rounded(round)

	ug = styleFrame.ApplyStrokeAndLineColor(ug, c.p.SkinParam().IHtmlColorSet())
	ug.Apply(klimt.UTranslateDx(border1)).Draw(rect)
}

// Blotter paints the background bands of a fragment, one per section.
type Blotter struct {
	dim              geom.XDimension2D
	defaultBackcolor color.HColor
	round            float64
	last             color.HColor
	changes          sortedFloatMap[color.HColor]
}

// NewBlotter returns the bands of a frame; a nil color is transparent.
func NewBlotter(dim geom.XDimension2D, defaultBackcolor color.HColor, round float64) *Blotter {
	if defaultBackcolor == nil {
		defaultBackcolor = color.Transparent()
	}
	return &Blotter{round: round, dim: dim, defaultBackcolor: defaultBackcolor, last: defaultBackcolor}
}

// String is upstream's toString.
func (b *Blotter) String() string { return b.dim.String() + " " + b.defaultBackcolor.String() }

// DrawU draws the bands.
func (b *Blotter) DrawU(ug klimt.UGraphic) {
	current := b.defaultBackcolor
	y := 0.0
	for i, key := range b.changes.keys {
		if !current.IsTransparent() {
			rect := b.rectangleBackground(i, key-y)
			ug.Apply(current).Apply(klimt.Bg(current)).Apply(klimt.UTranslateDy(y)).Draw(rect)
		}
		y = key
		current = b.changes.values[i]
	}
}

func (b *Blotter) rectangleBackground(i int, height float64) klimt.UShape {
	width := b.dim.Width()
	if b.round == 0 {
		return klimt.URectangleBuild(width, height)
	}
	if b.changes.size() == 1 {
		return klimt.URectangleBuild(width, height).Rounded(b.round)
	}
	r := b.round / 2
	if i == 0 {
		result := klimt.UPathNone()
		result.MoveTo(r, 0)
		result.LineTo(width-r, 0)
		result.ArcTo(r, r, 0, 0, 1, width, r)
		result.LineTo(width, height)
		result.LineTo(0, height)
		result.LineTo(0, r)
		result.ArcTo(r, r, 0, 0, 1, r, 0)
		result.ClosePath()
		return result
	}
	if i == b.changes.size()-1 {
		result := klimt.UPathNone()
		result.MoveTo(0, 0)
		result.LineTo(width, 0)
		result.LineTo(width, height-r)
		result.ArcTo(r, r, 0, 0, 1, width-r, height)
		result.LineTo(r, height)
		result.ArcTo(r, r, 0, 0, 1, 0, height-r)
		result.LineTo(0, 0)
		result.ClosePath()
		return result
	}
	return klimt.URectangleBuild(width, height)
}

// CloseChanges ends the last band at the bottom of the frame.
func (b *Blotter) CloseChanges() { b.changes.put(b.dim.Height(), b.defaultBackcolor) }

// AddChange starts a band of another color at ypos.
func (b *Blotter) AddChange(ypos float64, c color.HColor) {
	if c == nil {
		c = color.Transparent()
	}
	if c.Equals(b.last) {
		return
	}
	b.changes.put(ypos, c)
	b.last = c
}
