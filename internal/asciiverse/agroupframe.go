// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/asciiverse/AGroupFrame.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

// AGroupFrame is the frame around a group (alt, loop, group, partition...):
// its size is given by the group. With a tab (every group but a partition)
// the title is in a cut corner tab at the top left; without (a partition)
// it is centered on the top border.
type AGroupFrame struct {
	dimension ADimension2D
	title     []uint16
	useTab    bool
}

// NewAGroupFrame returns the frame of a size with its title; a null title
// (ok false) is empty.
func NewAGroupFrame(dimension ADimension2D, title string, ok bool, useTab bool) *AGroupFrame {
	if !ok {
		title = ""
	}
	return &AGroupFrame{dimension: dimension, title: javaChars(title), useTab: useTab}
}

// AsciiDimension returns the size given to NewAGroupFrame.
func (f *AGroupFrame) AsciiDimension() ADimension2D { return f.dimension }

// BodyRowOffset returns the rows of the header: 3 with a tab (top border,
// title row, tab closing row), 1 otherwise.
func (f *AGroupFrame) BodyRowOffset() int {
	if f.hasTab() {
		return 3
	}
	return 1
}

func (f *AGroupFrame) hasTab() bool {
	return f.useTab && len(f.title) > 0 && f.dimension.Width() > f.tabWidth()+2
}

func (f *AGroupFrame) hasTitleOnBorder() bool {
	return !f.useTab && len(f.title) > 0 && f.dimension.Width() >= len(f.title)+2+2
}

// tabWidth is the width of " " + title + " ".
func (f *AGroupFrame) tabWidth() int { return len(f.title) + 2 }

// AsciiDraw draws the frame in the shape of the plan's format.
func (f *AGroupFrame) AsciiDraw(plan *InfinitePlan) {
	if plan.IsUnicode() {
		f.asciiDrawUnicode(plan)
	} else {
		f.asciiDrawASCII(plan)
	}
}

// asciiDrawASCII draws a dashed top border, '!' sides, a tilde bottom
// border (a plain bottom border without tab) and the cut corner tab.
func (f *AGroupFrame) asciiDrawASCII(plan *InfinitePlan) {
	width := f.dimension.Width()
	height := f.dimension.Height()
	if width >= 1 {
		plan.Move(0, 0).DrawChar(plan.TopLeftChar())
		if width > 2 {
			plan.DrawHLine(1, width-2, 0)
		}
		if width > 1 {
			plan.Move(width-1, 0).DrawChar(plan.TopRightChar())
		}
	}
	for y := 1; y <= height-2; y++ {
		plan.Move(0, y).DrawChar('!')
		if width > 1 {
			plan.Move(width-1, y).DrawChar('!')
		}
	}
	if height >= 2 {
		if f.useTab {
			for x := 0; x < width; x++ {
				plan.Move(x, height-1).DrawChar('~')
			}
		} else {
			plan.Move(0, height-1).DrawChar(plan.BottomLeftChar())
			if width > 2 {
				plan.DrawHLine(1, width-2, height-1)
			}
			if width > 1 {
				plan.Move(width-1, height-1).DrawChar(plan.BottomRightChar())
			}
		}
	}
	if f.hasTab() {
		tabWidth := f.tabWidth()
		row := spaced(f.title)
		row = append(row, '/')
		plan.Move(1, 1).drawChars(row)
		for x := 0; x < tabWidth-1; x++ {
			plan.Move(1+x, 2).DrawChar('_')
		}
		plan.Move(1+tabWidth-1, 2).DrawChar('/')
	}
	f.drawTitleOnBorderIfNeeded(plan, width)
}

// drawTitleOnBorderIfNeeded centers " title " on the top border of a frame
// without tab.
func (f *AGroupFrame) drawTitleOnBorderIfNeeded(plan *InfinitePlan, width int) {
	if !f.hasTitleOnBorder() {
		return
	}
	labeled := spaced(f.title)
	start := (width - len(labeled)) / 2
	plan.Move(start, 0).drawChars(labeled)
}

// asciiDrawUnicode draws a double line frame; the tab shares its top and
// left borders and its right edge continues as a single line divider down
// to the bottom border.
func (f *AGroupFrame) asciiDrawUnicode(plan *InfinitePlan) {
	width := f.dimension.Width()
	height := f.dimension.Height()
	tab := f.hasTab()
	dividerColumn := 1 + f.tabWidth()
	if width >= 1 {
		plan.Move(0, 0).DrawChar('╔')
		if width > 2 {
			for x := 1; x <= width-2; x++ {
				c := uint16('═')
				if x == dividerColumn && tab {
					c = '╤'
				}
				plan.Move(x, 0).DrawChar(c)
			}
		}
		if width > 1 {
			plan.Move(width-1, 0).DrawChar('╗')
		}
	}
	for y := 1; y <= height-2; y++ {
		plan.Move(0, y).DrawChar('║')
		if width > 1 {
			plan.Move(width-1, y).DrawChar('║')
		}
	}
	if height >= 2 {
		plan.Move(0, height-1).DrawChar('╚')
		for x := 1; x <= width-2; x++ {
			c := uint16('═')
			if x == dividerColumn && tab {
				c = '╧'
			}
			plan.Move(x, height-1).DrawChar(c)
		}
		if width > 1 {
			plan.Move(width-1, height-1).DrawChar('╝')
		}
	}
	if tab {
		tabWidth := f.tabWidth()
		plan.Move(1, 1).drawChars(spaced(f.title))
		plan.Move(dividerColumn, 1).DrawChar('│')
		plan.Move(0, 2).DrawChar('╟')
		for x := 1; x < tabWidth+1; x++ {
			plan.Move(x, 2).DrawChar('─')
		}
		plan.Move(dividerColumn, 2).DrawChar('┘')
		for y := 3; y <= height-2; y++ {
			plan.Move(dividerColumn, y).DrawChar('│')
		}
	}
	f.drawTitleOnBorderIfNeeded(plan, width)
}
