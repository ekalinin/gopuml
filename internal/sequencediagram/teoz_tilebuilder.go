// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/teoz/TileBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

// eventIterator is upstream's Iterator<Event> over the events of the
// diagram: GroupingTile reads its children from the same iterator.
type eventIterator struct {
	events []Event
	pos    int
}

func (it *eventIterator) hasNext() bool { return it.pos < len(it.events) }

func (it *eventIterator) next() Event {
	ev := it.events[it.pos]
	it.pos++
	return ev
}

// tileBuildSeveral builds the tiles of the remaining events, each chained
// after the previous one on the Y line.
func tileBuildSeveral(it *eventIterator, tileArguments *TileArguments, parent Tile, currentY *YGauge) []Tile {
	var tiles []Tile
	for it.hasNext() {
		ev := it.next()
		for _, tile := range tileBuildOne(it, tileArguments, ev, parent, currentY) {
			tiles = append(tiles, tile)
			// Y chaining: the next tile's gauge is built on this one's
			currentY = tile.YGauge()
		}
	}
	return tiles
}

// tileBuildOne builds the tiles of an event; a fragment reads its children
// from it.
func tileBuildOne(it *eventIterator, tileArguments *TileArguments, ev Event, parent Tile,
	currentY *YGauge) []Tile {
	stringBounder := tileArguments.StringBounder()
	skin := tileArguments.Skin()
	skinParam := tileArguments.SkinParam()
	livingSpaces := tileArguments.LivingSpaces()

	var tiles []Tile
	switch ev := ev.(type) {
	case *Message:
		livingSpace1 := livingSpaces.Get(ev.Participant1())
		livingSpace2 := livingSpaces.Get(ev.Participant2())
		reverse := false
		var result Tile
		if ev.IsSelfMessage() {
			result = newCommunicationTileSelf(stringBounder, livingSpace1, ev, skin, skinParam, livingSpaces,
				currentY)
		} else {
			tile := newCommunicationTile(stringBounder, livingSpaces, ev, skin, skinParam, currentY)
			reverse = tile.IsReverse(stringBounder)
			result = tile
		}
		for _, noteOnMessage := range ev.NoteOnMessages() {
			notePosition := noteOnMessage.Position()
			switch {
			case notePosition == NoteLeft && ev.IsSelfMessage():
				result = newCommunicationTileSelfNoteLeft(result.(*CommunicationTileSelf), ev, skin, skinParam,
					noteOnMessage, currentY)
			case notePosition == NoteLeft:
				ls := livingSpace1
				if reverse {
					ls = livingSpace2
				}
				result = newCommunicationTileNoteLeft(result, ev, skin, skinParam, ls, noteOnMessage, currentY)
			case notePosition == NoteRight && ev.IsSelfMessage():
				result = newCommunicationTileSelfNoteRight(result.(*CommunicationTileSelf), ev, skin, skinParam,
					noteOnMessage, currentY)
			case notePosition == NoteRight:
				ls := livingSpace2
				if reverse {
					ls = livingSpace1
				}
				result = newCommunicationTileNoteRight(result, ev, skin, skinParam, ls, noteOnMessage, currentY)
			case notePosition == NoteBottom:
				result = newCommunicationTileNoteBottom(result, ev, skin, skinParam, noteOnMessage, currentY)
			case notePosition == NoteTop:
				result = newCommunicationTileNoteTop(result, ev, skin, skinParam, noteOnMessage, currentY)
			}
		}
		tiles = append(tiles, result)
	case *MessageExo:
		livingSpace1 := livingSpaces.Get(ev.Participant())
		var result Tile = newCommunicationExoTile(livingSpace1, ev, skin, skinParam, tileArguments, currentY)
		for _, noteOnMessage := range ev.NoteOnMessages() {
			notePosition := ev.NoteOnMessages()[0].Position()
			if notePosition == NoteLeft {
				result = newCommunicationTileNoteLeft(result, ev, skin, skinParam, livingSpace1, noteOnMessage,
					currentY)
			} else if notePosition == NoteRight {
				result = newCommunicationTileNoteRight(result, ev, skin, skinParam, livingSpace1, noteOnMessage,
					currentY)
			}
		}
		tiles = append(tiles, result)
	case *Note:
		livingSpace1 := livingSpaces.Get(ev.Participant())
		var livingSpace2 *LivingSpace
		if ev.Participant2() != nil {
			livingSpace2 = livingSpaces.Get(ev.Participant2())
		}
		if livingSpace1 == nil && livingSpace2 == nil {
			livingSpace1 = tileArguments.FirstLivingSpace()
			livingSpace2 = tileArguments.LastLivingSpace()
		}
		tiles = append(tiles, newNoteTile(stringBounder, livingSpace1, livingSpace2, ev, skin, skinParam, currentY))
	case *Notes:
		tiles = append(tiles, newNotesTile(stringBounder, livingSpaces, ev, skin, skinParam, currentY))
	case *Divider:
		tiles = append(tiles, newDividerTile(ev, tileArguments, currentY))
	case *GroupingStart:
		backColored := tileArguments.WithBackColorGeneral(ev.BackColorElement(), ev.BackColorGeneral())
		if ev.Type() == GroupingStartPartition {
			tiles = append(tiles, newPartitionTile(it, ev, backColored, tileArguments, currentY))
		} else {
			tiles = append(tiles, newGroupingTile(it, ev, backColored, tileArguments, currentY))
		}
	case *GroupingLeaf:
		if ev.Type() == GroupingElse {
			tiles = append(tiles, newElseTile(ev, skin, skinParam, parent, currentY))
		}
	case *Reference:
		tiles = append(tiles, newReferenceTile(ev, tileArguments.WithBackColor(ev), currentY))
	case *Delay:
		tiles = append(tiles, newDelayTile(ev, tileArguments, currentY))
	case *HSpace:
		tiles = append(tiles, newHSpaceTile(ev, tileArguments, currentY))
	case *LifeEvent:
		livingSpace := livingSpaces.Get(ev.Participant())
		tiles = append(tiles, newLifeEventTile(ev, tileArguments, livingSpace, skin, skinParam, currentY))
	case *Newpage:
		tiles = append(tiles, newNewpageTile(ev, tileArguments, currentY))
	}
	// Upstream prints "TileBuilder::Ignoring" on standard error for the
	// other events (an "end" outside its fragment cannot reach here).
	return tiles
}
