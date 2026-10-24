// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/asciiverse/ANote.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

// ANote is the note box around a text: a folded corner ("!." over "|_\")
// in ATXT, a double line box with a shaded fold in UTXT.
type ANote struct {
	width  int
	height int
	text   AsciiBlock
}

// NewANote returns the box around text; its size is computed once, from
// the text: max(5, text width + 2) by max(3, text height + 2).
func NewANote(text AsciiBlock) *ANote {
	textDim := text.AsciiDimension()
	return &ANote{text: text, width: max(5, textDim.Width()+2), height: max(3, textDim.Height()+2)}
}

// AsciiDimension returns the size of the box.
func (n *ANote) AsciiDimension() ADimension2D { return NewADimension2D(n.width, n.height) }

// AsciiDraw draws the box in the characters of the plan's format, then the
// text one column and one row inside.
func (n *ANote) AsciiDraw(plan *InfinitePlan) {
	if plan.IsUnicode() {
		n.asciiDrawUnicode(plan)
	} else {
		n.asciiDrawASCII(plan)
	}
	n.text.AsciiDraw(plan.Move(1, 1))
}

func (n *ANote) asciiDrawASCII(plan *InfinitePlan) {
	width, height := n.width, n.height
	if width > 2 {
		plan.DrawHLine(1, width-2, 0)
	}
	if height > 2 && width > 2 {
		plan.DrawHLine(1, width-2, height-1)
	}
	if height > 2 {
		plan.fillVLine('|', 1, height-2, 0)
	}
	if height > 2 && width > 1 {
		plan.fillVLine('|', 1, height-2, width-1)
	}
	plan.DrawChar(',')
	if height > 1 {
		plan.Move(0, height-1).DrawChar('`')
	}
	if height > 1 && width > 1 {
		plan.Move(width-1, height-1).DrawChar('\'')
	}
	// The folded corner, over the top right corner
	if width >= 4 {
		plan.Move(width-3, 0).DrawString("!.")
		plan.Move(width-1, 0).DrawChar(' ')
		if height > 1 {
			plan.Move(width-3, 1).DrawString("|_\\")
		}
	}
}

func (n *ANote) asciiDrawUnicode(plan *InfinitePlan) {
	width, height := n.width, n.height
	if width > 2 {
		plan.fillHLine('═', 1, width-2, 0)
	}
	if height > 2 && width > 2 {
		plan.fillHLine('═', 1, width-2, height-1)
	}
	if height > 2 {
		plan.fillVLine('║', 1, height-2, 0)
	}
	if height > 2 && width > 1 {
		plan.fillVLine('║', 1, height-2, width-1)
	}
	plan.DrawChar('╔')
	if width > 1 {
		plan.Move(width-1, 0).DrawChar('╗')
	}
	if height > 1 {
		plan.Move(0, height-1).DrawChar('╚')
	}
	if height > 1 && width > 1 {
		plan.Move(width-1, height-1).DrawChar('╝')
	}
	// The shaded fold, below the top border, left of the right border
	if width >= 2 && height >= 2 {
		plan.Move(width-2, 1).DrawChar('░')
	}
}
