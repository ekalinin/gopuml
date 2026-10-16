// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileSelf.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationExoTile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteLeft.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileNoteRight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileSelfNoteLeft.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/CommunicationTileSelfNoteRight.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/asciiverse"
)

// The text drawing of the messages and of the notes attached to them. The
// tiles read their own columns from the text X line: the plan given to
// AsciiDraw only sets their first row.

// asciiLabel is the label of the message with one blank column on each
// side (none for an empty label).
func (t *CommunicationTile) asciiLabel() asciiverse.AsciiBlock {
	return asciiverse.MarginLR(asciiverse.DisplayBlock(t.message.Label()), 1, 1)
}

// asciiLabelRows returns the rows of the label, at least 1.
func (t *CommunicationTile) asciiLabelRows() int {
	return max(1, t.asciiLabel().AsciiDimension().Height())
}

// AsciiDraw draws the label centered between the lifelines, then the arrow
// on the row below it, between the lifelines, its head next to the target.
func (t *CommunicationTile) AsciiDraw(plan *asciiverse.InfinitePlan) {
	xa := asciiColumn(t.livingSpace1.AsciiLifeColumn())
	xb := asciiColumn(t.livingSpace2.AsciiLifeColumn())

	label := t.asciiLabel()
	labelDim := label.AsciiDimension()
	if labelDim.Width() > 0 {
		mid := (xa + xb) / 2
		label.AsciiDraw(plan.Move(mid-labelDim.Width()/2, 0))
	}

	arrowRow := t.asciiLabelRows()
	from := min(xa, xb)
	to := max(xa, xb)
	plan.DrawHLineDotted(from+1, to-1, arrowRow, t.message.ArrowConfiguration().IsDotted())

	// The head stays ASCII in UTXT
	if xb >= xa {
		plan.Move(xb-1, arrowRow).DrawChar('>')
	} else {
		plan.Move(xb+1, arrowRow).DrawChar('<')
	}
}

// AsciiDimension returns the width of the message by the label rows, the
// arrow row and a blank row.
func (t *CommunicationTile) AsciiDimension() asciiverse.ADimension2D {
	return asciiverse.NewADimension2D(t.asciiMessageWidth(), t.asciiLabelRows()+2)
}

// asciiMessageWidth is the padded label and one free column next to each
// lifeline.
func (t *CommunicationTile) asciiMessageWidth() int {
	return t.asciiLabel().AsciiDimension().Width() + 2
}

// AsciiAddConstraints keeps the lifelines far enough apart for the label.
func (t *CommunicationTile) AsciiAddConstraints() {
	width := float64(t.asciiMessageWidth())
	point1 := t.livingSpace1.AsciiPosC()
	point2 := t.livingSpace2.AsciiPosC()
	if point1.CurrentValue() > point2.CurrentValue() {
		point1.EnsureBiggerThan(point2.AddFixed(width))
	} else {
		point2.EnsureBiggerThan(point1.AddFixed(width))
	}
}

// AsciiMinX returns the leftmost lifeline.
func (t *CommunicationTile) AsciiMinX() Real {
	return RealUtilsMin(t.livingSpace1.AsciiLifeColumn(), t.livingSpace2.AsciiLifeColumn())
}

// AsciiMaxX returns the rightmost lifeline.
func (t *CommunicationTile) AsciiMaxX() Real {
	return RealUtilsMax(t.livingSpace1.AsciiLifeColumn(), t.livingSpace2.AsciiLifeColumn())
}

// The loop of a self message: 5 columns ("----." / "|" / "<---'"), then
// one blank column before the label.
const (
	asciiSelfBoxWidth = 5
	asciiSelfLabelGap = 1
)

func (t *CommunicationTileSelf) asciiLabel() asciiverse.AsciiBlock {
	return asciiverse.MarginLR(asciiverse.DisplayBlock(t.message.Label()), 1, 1)
}

func (t *CommunicationTileSelf) asciiLabelRows() int {
	return max(1, t.asciiLabel().AsciiDimension().Height())
}

// asciiSelfWidth is how far the loop and its label reach from the
// lifeline.
func (t *CommunicationTileSelf) asciiSelfWidth() int {
	labelWidth := t.asciiLabel().AsciiDimension().Width()
	if labelWidth == 0 {
		return asciiSelfBoxWidth
	}
	return asciiSelfBoxWidth + asciiSelfLabelGap + labelWidth
}

// AsciiDraw draws the loop on the right of the lifeline (on its left for a
// message defined in reverse), the label after it.
func (t *CommunicationTileSelf) AsciiDraw(plan *asciiverse.InfinitePlan) {
	xa := asciiColumn(t.livingSpace1.AsciiLifeColumn())
	dotted := t.message.ArrowConfiguration().IsDotted()
	rows := t.asciiLabelRows()
	bottomRow := rows + 1
	label := t.asciiLabel()
	labelWidth := label.AsciiDimension().Width()

	if t.isReverseDefine() {
		nearCol := xa - 1
		farCol := xa - asciiSelfBoxWidth

		plan.DrawHLineDotted(farCol+1, nearCol, 0, dotted)
		plan.Move(farCol, 0).DrawChar(plan.TopLeftChar())
		for r := 1; r <= rows; r++ {
			plan.Move(farCol, r).DrawChar(plan.VLineChar())
		}
		plan.DrawHLineDotted(farCol+1, nearCol, bottomRow, dotted)
		plan.Move(farCol, bottomRow).DrawChar(plan.BottomLeftChar())
		plan.Move(nearCol, bottomRow).DrawChar('>')

		if labelWidth > 0 {
			label.AsciiDraw(plan.Move(farCol-asciiSelfLabelGap-labelWidth, 1))
		}
	} else {
		nearCol := xa + 1
		farCol := xa + asciiSelfBoxWidth

		plan.DrawHLineDotted(nearCol, farCol-1, 0, dotted)
		plan.Move(farCol, 0).DrawChar(plan.TopRightChar())
		for r := 1; r <= rows; r++ {
			plan.Move(farCol, r).DrawChar(plan.VLineChar())
		}
		plan.DrawHLineDotted(nearCol, farCol-1, bottomRow, dotted)
		plan.Move(farCol, bottomRow).DrawChar(plan.BottomRightChar())
		plan.Move(nearCol, bottomRow).DrawChar('<')

		if labelWidth > 0 {
			label.AsciiDraw(plan.Move(farCol+asciiSelfLabelGap, 1))
		}
	}
}

// AsciiDimension returns the reach of the loop by the label rows and two
// rows.
func (t *CommunicationTileSelf) AsciiDimension() asciiverse.ADimension2D {
	return asciiverse.NewADimension2D(t.asciiSelfWidth(), t.asciiLabelRows()+2)
}

// AsciiAddConstraints pushes the neighbour on the side of the loop far
// enough away.
func (t *CommunicationTileSelf) AsciiAddConstraints() {
	width := float64(t.asciiSelfWidth())
	if t.isReverseDefine() {
		if previous := t.previous(); previous != nil {
			t.livingSpace1.AsciiPosC().EnsureBiggerThan(previous.AsciiPosC().AddFixed(width))
		}
	} else {
		if next := t.next(); next != nil {
			next.AsciiPosC().EnsureBiggerThan(t.livingSpace1.AsciiPosC().AddFixed(width))
		}
	}
}

// AsciiMinX returns the left of the loop.
func (t *CommunicationTileSelf) AsciiMinX() Real {
	if t.isReverseDefine() {
		return t.livingSpace1.AsciiPosC().AddFixed(float64(-t.asciiSelfWidth()))
	}
	return t.livingSpace1.AsciiPosC()
}

// AsciiMaxX returns the right of the loop.
func (t *CommunicationTileSelf) AsciiMaxX() Real {
	if t.isReverseDefine() {
		return t.livingSpace1.AsciiPosC()
	}
	return t.livingSpace1.AsciiPosC().AddFixed(float64(t.asciiSelfWidth()))
}

// asciiExoMargin is how far past the outermost box the border of the
// diagram is, for a message from or to the border.
const asciiExoMargin = 2

// asciiExoShortMinWidth is the smallest stub of a short exo arrow.
const asciiExoShortMinWidth = 3

func (t *CommunicationExoTile) asciiLabel() asciiverse.AsciiBlock {
	return asciiverse.MarginLR(asciiverse.DisplayBlock(t.message.Label()), 1, 1)
}

func (t *CommunicationExoTile) asciiLabelRows() int {
	return max(1, t.asciiLabel().AsciiDimension().Height())
}

func (t *CommunicationExoTile) asciiShortWidth() int {
	return max(asciiExoShortMinWidth, t.asciiLabel().AsciiDimension().Width())
}

// asciiEdge returns the column the arrow reaches: the border of the
// diagram, or a stub next to the participant for a short arrow.
func (t *CommunicationExoTile) asciiEdge() Real {
	leftBorder := t.message.Type().IsLeftBorder()
	if t.isShortArrow() {
		width := float64(t.asciiShortWidth())
		if leftBorder {
			return t.livingSpace.AsciiPosC().AddFixed(-width)
		}
		return t.livingSpace.AsciiPosC().AddFixed(width)
	}
	if leftBorder {
		return t.tileArguments.FirstLivingSpace().AsciiPosB().AddFixed(-asciiExoMargin)
	}
	return t.tileArguments.LastLivingSpace().AsciiPosD().AddFixed(asciiExoMargin - 1)
}

// AsciiDraw draws the arrow between the lifeline and the edge, its head
// next to the lifeline or on the edge, then the label centered over it.
func (t *CommunicationExoTile) AsciiDraw(plan *asciiverse.InfinitePlan) {
	p := asciiColumn(t.livingSpace.AsciiLifeColumn())
	edge := asciiColumn(t.asciiEdge())
	dotted := t.message.ArrowConfiguration().IsDotted()
	targetIsParticipant := t.message.Type() == ExoFromLeft || t.message.Type() == ExoFromRight
	arrowRow := t.asciiLabelRows()

	var dashFrom, dashTo, arrowCol int
	var arrowChar uint16
	if edge < p {
		if targetIsParticipant {
			dashFrom, dashTo, arrowCol, arrowChar = edge, p-2, p-1, '>'
		} else {
			dashFrom, dashTo, arrowCol, arrowChar = edge+1, p-1, edge, '<'
		}
	} else {
		if targetIsParticipant {
			dashFrom, dashTo, arrowCol, arrowChar = p+2, edge, p+1, '<'
		} else {
			dashFrom, dashTo, arrowCol, arrowChar = p+1, edge-1, edge, '>'
		}
	}
	if dashFrom <= dashTo {
		plan.DrawHLineDotted(dashFrom, dashTo, arrowRow, dotted)
	}
	plan.Move(arrowCol, arrowRow).DrawChar(arrowChar)

	label := t.asciiLabel()
	labelWidth := label.AsciiDimension().Width()
	if labelWidth > 0 {
		a := min(edge, p)
		b := max(edge, p)
		mid := (a + b) / 2
		label.AsciiDraw(plan.Move(mid-labelWidth/2, 0))
	}
}

// AsciiDimension returns the length of the arrow by the label rows, the
// arrow row and a blank row.
func (t *CommunicationExoTile) AsciiDimension() asciiverse.ADimension2D {
	p := asciiColumn(t.livingSpace.AsciiLifeColumn())
	edge := asciiColumn(t.asciiEdge())
	return asciiverse.NewADimension2D(abs(edge-p), t.asciiLabelRows()+2)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AsciiAddConstraints pushes the neighbour on the side of a short arrow far
// enough away; an arrow to the border needs nothing.
func (t *CommunicationExoTile) AsciiAddConstraints() {
	if !t.isShortArrow() {
		return
	}
	width := float64(t.asciiShortWidth())
	if t.message.Type().IsLeftBorder() {
		if previous := t.tileArguments.LivingSpaces().Previous(t.livingSpace); previous != nil {
			t.livingSpace.AsciiPosC().EnsureBiggerThan(previous.AsciiPosC().AddFixed(width))
		}
	} else {
		if next := t.tileArguments.LivingSpaces().Next(t.livingSpace); next != nil {
			next.AsciiPosC().EnsureBiggerThan(t.livingSpace.AsciiPosC().AddFixed(width))
		}
	}
}

// AsciiMinX returns the left of the lifeline and the edge.
func (t *CommunicationExoTile) AsciiMinX() Real {
	return RealUtilsMin(t.asciiEdge(), t.livingSpace.AsciiLifeColumn())
}

// AsciiMaxX returns the right of the lifeline and the edge.
func (t *CommunicationExoTile) AsciiMaxX() Real {
	return RealUtilsMax(t.asciiEdge(), t.livingSpace.AsciiLifeColumn())
}

// asciiNoteText is the text of a note attached to a message, with one
// blank column on its left and three on its right.
func asciiNoteText(noteOnMessage *Note) asciiverse.AsciiBlock {
	return asciiverse.MarginLR(asciiverse.DisplayBlock(noteOnMessage.Display()), 1, 3)
}

// asciiNoteBoxWidth is the width of the note box with two blank columns on
// each side.
func asciiNoteBoxWidth(noteOnMessage *Note) int {
	return asciiverse.MarginLR(asciiverse.NewANote(asciiNoteText(noteOnMessage)), 2, 2).AsciiDimension().Width()
}

// asciiNoteHeight is the height of the note box.
func asciiNoteHeight(noteOnMessage *Note) int {
	return asciiverse.NewANote(asciiNoteText(noteOnMessage)).AsciiDimension().Height()
}

// asciiNoteDimension is the width of the message by the highest of the
// message and the note.
func asciiNoteDimension(tile Tile, noteOnMessage *Note) asciiverse.ADimension2D {
	tileDim := tile.AsciiDimension()
	noteHeight := asciiNoteHeight(noteOnMessage)
	return asciiverse.NewADimension2D(tileDim.Width(), max(tileDim.Height(), noteHeight))
}

// AsciiAddConstraints adds the constraints of the message; the note is not
// reserved on the text X line.
func (t *CommunicationTileNoteLeft) AsciiAddConstraints() { t.tile.AsciiAddConstraints() }

// AsciiMinX returns the left of the message and of the note box, so that a
// frame around the message contains the note.
func (t *CommunicationTileNoteLeft) AsciiMinX() Real {
	boxWidth := asciiNoteBoxWidth(t.noteOnMessage)
	noteMin := t.livingSpace.AsciiLifeColumn().AddFixed(float64(-(boxWidth + 1)))
	tileMin := t.tile.AsciiMinX()
	if tileMin == nil {
		return noteMin
	}
	return RealUtilsMin(tileMin, noteMin)
}

// AsciiMaxX returns the right of the message.
func (t *CommunicationTileNoteLeft) AsciiMaxX() Real { return t.tile.AsciiMaxX() }

// AsciiDimension returns the width of the message by the highest of the
// message and the note.
func (t *CommunicationTileNoteLeft) AsciiDimension() asciiverse.ADimension2D {
	return asciiNoteDimension(t.tile, t.noteOnMessage)
}

// AsciiDraw draws the message, then the note box ending before the
// lifeline.
func (t *CommunicationTileNoteLeft) AsciiDraw(plan *asciiverse.InfinitePlan) {
	t.tile.AsciiDraw(plan)

	noteBox := asciiverse.MarginLR(plan.CreateNoteBox(asciiNoteText(t.noteOnMessage)), 2, 2)
	sourceColumn := asciiColumn(t.livingSpace.AsciiLifeColumn())
	width := noteBox.AsciiDimension().Width()
	left := sourceColumn - width - 1
	noteBox.AsciiDraw(plan.Move(left, 0))
}

// AsciiAddConstraints adds the constraints of the message.
func (t *CommunicationTileNoteRight) AsciiAddConstraints() { t.tile.AsciiAddConstraints() }

// AsciiMinX returns the left of the message.
func (t *CommunicationTileNoteRight) AsciiMinX() Real { return t.tile.AsciiMinX() }

// AsciiMaxX returns the right of the message and of the note box.
func (t *CommunicationTileNoteRight) AsciiMaxX() Real {
	boxWidth := asciiNoteBoxWidth(t.noteOnMessage)
	noteMax := t.livingSpace.AsciiLifeColumn().AddFixed(float64(boxWidth + 1))
	tileMax := t.tile.AsciiMaxX()
	if tileMax == nil {
		return noteMax
	}
	return RealUtilsMax(tileMax, noteMax)
}

// AsciiDimension returns the width of the message by the highest of the
// message and the note.
func (t *CommunicationTileNoteRight) AsciiDimension() asciiverse.ADimension2D {
	return asciiNoteDimension(t.tile, t.noteOnMessage)
}

// AsciiDraw draws the message, then the note box after the lifeline.
func (t *CommunicationTileNoteRight) AsciiDraw(plan *asciiverse.InfinitePlan) {
	t.tile.AsciiDraw(plan)

	noteBox := asciiverse.MarginLR(plan.CreateNoteBox(asciiNoteText(t.noteOnMessage)), 2, 2)
	targetColumn := asciiColumn(t.livingSpace.AsciiLifeColumn())
	left := targetColumn + 2
	noteBox.AsciiDraw(plan.Move(left, 0))
}

// AsciiAddConstraints adds the constraints of the self message.
func (t *CommunicationTileSelfNoteLeft) AsciiAddConstraints() { t.tile.AsciiAddConstraints() }

// AsciiMinX returns the left of the loop and of the note box.
func (t *CommunicationTileSelfNoteLeft) AsciiMinX() Real {
	boxWidth := asciiNoteBoxWidth(t.noteOnMessage)
	tileMin := t.tile.AsciiMinX()
	noteMin := tileMin.AddFixed(float64(-(boxWidth + 1)))
	return RealUtilsMin(tileMin, noteMin)
}

// AsciiMaxX returns the right of the loop.
func (t *CommunicationTileSelfNoteLeft) AsciiMaxX() Real { return t.tile.AsciiMaxX() }

// AsciiDimension returns the width of the loop by the highest of the loop
// and the note.
func (t *CommunicationTileSelfNoteLeft) AsciiDimension() asciiverse.ADimension2D {
	return asciiNoteDimension(t.tile, t.noteOnMessage)
}

// AsciiDraw draws the loop, then the note box ending before the left of
// the loop.
func (t *CommunicationTileSelfNoteLeft) AsciiDraw(plan *asciiverse.InfinitePlan) {
	t.tile.AsciiDraw(plan)

	noteBox := asciiverse.MarginLR(plan.CreateNoteBox(asciiNoteText(t.noteOnMessage)), 2, 2)
	anchorColumn := asciiColumn(t.tile.AsciiMinX())
	width := noteBox.AsciiDimension().Width()
	left := anchorColumn - width - 1
	noteBox.AsciiDraw(plan.Move(left, 0))
}

// AsciiAddConstraints adds the constraints of the self message.
func (t *CommunicationTileSelfNoteRight) AsciiAddConstraints() { t.tile.AsciiAddConstraints() }

// AsciiMinX returns the left of the loop.
func (t *CommunicationTileSelfNoteRight) AsciiMinX() Real { return t.tile.AsciiMinX() }

// AsciiMaxX returns the right of the loop and of the note box.
func (t *CommunicationTileSelfNoteRight) AsciiMaxX() Real {
	boxWidth := asciiNoteBoxWidth(t.noteOnMessage)
	tileMax := t.tile.AsciiMaxX()
	noteMax := tileMax.AddFixed(float64(boxWidth + 1))
	return RealUtilsMax(tileMax, noteMax)
}

// AsciiDimension returns the width of the loop by the highest of the loop
// and the note.
func (t *CommunicationTileSelfNoteRight) AsciiDimension() asciiverse.ADimension2D {
	return asciiNoteDimension(t.tile, t.noteOnMessage)
}

// AsciiDraw draws the loop, then the note box after the right of the loop.
func (t *CommunicationTileSelfNoteRight) AsciiDraw(plan *asciiverse.InfinitePlan) {
	t.tile.AsciiDraw(plan)

	noteBox := asciiverse.MarginLR(plan.CreateNoteBox(asciiNoteText(t.noteOnMessage)), 2, 2)
	anchorColumn := asciiColumn(t.tile.AsciiMaxX())
	left := anchorColumn + 2
	noteBox.AsciiDraw(plan.Move(left, 0))
}
