// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/asciiverse/InfinitePlan.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package asciiverse

import (
	"bufio"
	"io"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ekalinin/gopuml/internal/asciiart"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// InfinitePlan is the character grid of the text formats, unbounded in
// every direction, seen from a position (dx, dy): Move returns a plan over
// the same grid at another position, and the drawing methods write
// relative to the position.
//
// The plan owns every character choice that depends on the format: ASCII
// for ATXT, box drawing characters for UTXT.
type InfinitePlan struct {
	stringBounder *asciiart.TextStringBounder
	plan          *[]*InfiniteString
	dx, dy        int
}

// NewInfinitePlan returns an empty grid for a text format.
func NewInfinitePlan(fileFormat font.FileFormat) *InfinitePlan {
	return &InfinitePlan{stringBounder: asciiart.NewTextStringBounder(fileFormat), plan: &[]*InfiniteString{}}
}

// Move returns the plan over the same grid at a position moved by (movex,
// movey).
func (p *InfinitePlan) Move(movex, movey int) *InfinitePlan {
	return &InfinitePlan{stringBounder: p.stringBounder, plan: p.plan, dx: p.dx + movex, dy: p.dy + movey}
}

// DrawChar writes a character at the position. As upstream, a negative row
// panics (Java's List.get).
func (p *InfinitePlan) DrawChar(c uint16) {
	p.line(p.dy).SetCharAt(p.dx, c)
}

// MuteChar replaces the character at the position by the one following it
// in values (pairs: old, new), for the first pair that matches; an odd
// number of values panics with IllegalArgumentException.
func (p *InfinitePlan) MuteChar(values ...uint16) {
	if len(values)%2 != 0 {
		throw(illegalArgument, "")
	}
	line := p.line(p.dy)
	existing := line.CharAt(p.dx)
	for i := 0; i < len(values); i += 2 {
		if existing == values[i] {
			line.SetCharAt(p.dx, values[i+1])
			return
		}
	}
}

// DrawString writes the UTF-16 code units of s from the position on.
func (p *InfinitePlan) DrawString(s string) {
	p.line(p.dy).SetStringAt(p.dx, s)
}

// drawChars writes UTF-16 code units from the position on.
func (p *InfinitePlan) drawChars(chars []uint16) {
	p.line(p.dy).setCharsAt(p.dx, chars)
}

// javaChars returns the UTF-16 code units of s, the chars of a Java string.
func javaChars(s string) []uint16 { return utf16.Encode([]rune(s)) }

// line returns the line at an absolute row, adding lines as needed;
// a negative row panics as Java's List.get.
func (p *InfinitePlan) line(yAbsolute int) *InfiniteString {
	for yAbsolute >= len(*p.plan) {
		*p.plan = append(*p.plan, &InfiniteString{})
	}
	if yAbsolute < 0 {
		throw("java.lang.IndexOutOfBoundsException", "")
	}
	return (*p.plan)[yAbsolute]
}

// CharAt returns the character at (x, y) relative to the position, a space
// outside the written lines.
func (p *InfinitePlan) CharAt(x, y int) uint16 {
	x += p.dx
	y += p.dy
	if y < 0 || y >= len(*p.plan) {
		return ' '
	}
	return (*p.plan)[y].CharAt(x)
}

// ExportTxt writes every line, from the leftmost column any line used, in
// UTF-8 with a "\n" after each line. A lone surrogate is written as '?',
// as Java's encoder does.
func (p *InfinitePlan) ExportTxt(w io.Writer) error {
	bw := bufio.NewWriter(w)
	startingPosition := 0
	for _, line := range *p.plan {
		startingPosition = min(startingPosition, line.LeftmostPosition())
	}
	var buf []byte
	for _, line := range *p.plan {
		buf = appendJavaUTF8(buf[:0], line.Chars(startingPosition))
		buf = append(buf, '\n')
		if _, err := bw.Write(buf); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// appendJavaUTF8 appends UTF-16 code units encoded as Java's UTF-8 encoder
// does: surrogate pairs as one code point, lone surrogates as '?'.
func appendJavaUTF8(buf []byte, chars []uint16) []byte {
	for i := 0; i < len(chars); i++ {
		c := rune(chars[i])
		switch {
		case utf16.IsSurrogate(c):
			if c < 0xdc00 && i+1 < len(chars) && chars[i+1] >= 0xdc00 && chars[i+1] <= 0xdfff {
				buf = utf8.AppendRune(buf, utf16.DecodeRune(c, rune(chars[i+1])))
				i++
			} else {
				buf = append(buf, '?')
			}
		default:
			buf = utf8.AppendRune(buf, c)
		}
	}
	return buf
}

// StringBounder returns the text measurer of the format.
func (p *InfinitePlan) StringBounder() *asciiart.TextStringBounder { return p.stringBounder }

// IsUnicode reports whether the grid is drawn with box drawing characters
// (UTXT).
func (p *InfinitePlan) IsUnicode() bool { return p.stringBounder.FileFormat() == font.FormatUTXT }

// HLineChar returns the horizontal line character: '─' or '-'.
func (p *InfinitePlan) HLineChar() uint16 {
	if p.IsUnicode() {
		return '─'
	}
	return '-'
}

// VLineChar returns the vertical line character: '│' or '|'.
func (p *InfinitePlan) VLineChar() uint16 {
	if p.IsUnicode() {
		return '│'
	}
	return '|'
}

// tDownChar is a T pointing down (a box bottom with a lifeline below).
func (p *InfinitePlan) tDownChar() uint16 {
	if p.IsUnicode() {
		return '┬'
	}
	return '+'
}

// tUpChar is a T pointing up (a lifeline arriving into a box top).
func (p *InfinitePlan) tUpChar() uint16 {
	if p.IsUnicode() {
		return '┴'
	}
	return '+'
}

// MuteLifelineBelowBox turns a box border at the position into a T
// pointing down, and a blank into a lifeline segment.
func (p *InfinitePlan) MuteLifelineBelowBox() {
	p.MuteChar(p.HLineChar(), p.tDownChar(), ' ', p.VLineChar())
}

// MuteLifelineAboveBox turns a box border at the position into a T
// pointing up, and a blank into a lifeline segment.
func (p *InfinitePlan) MuteLifelineAboveBox() {
	p.MuteChar(p.HLineChar(), p.tUpChar(), ' ', p.VLineChar())
}

// DrawHLine draws a horizontal line from xFrom to xTo (inclusive) on row y.
func (p *InfinitePlan) DrawHLine(xFrom, xTo, y int) {
	p.fillHLine(p.HLineChar(), xFrom, xTo, y)
}

// DrawVLine draws a vertical line from yFrom to yTo (inclusive) on column
// x.
func (p *InfinitePlan) DrawVLine(yFrom, yTo, x int) {
	p.fillVLine(p.VLineChar(), yFrom, yTo, x)
}

// DrawHLineDotted draws a horizontal line, every other cell only when
// dotted (upstream's drawHLine(int, int, int, boolean)).
func (p *InfinitePlan) DrawHLineDotted(xFrom, xTo, y int, dotted bool) {
	if !dotted {
		p.DrawHLine(xFrom, xTo, y)
		return
	}
	c := p.HLineChar()
	idx := 0
	for i := xFrom; i <= xTo; i, idx = i+1, idx+1 {
		if idx%2 == 0 {
			p.Move(i, y).DrawChar(c)
		}
	}
}

func (p *InfinitePlan) fillHLine(c uint16, xFrom, xTo, y int) {
	for i := xFrom; i <= xTo; i++ {
		p.Move(i, y).DrawChar(c)
	}
}

func (p *InfinitePlan) fillVLine(c uint16, yFrom, yTo, x int) {
	for j := yFrom; j <= yTo; j++ {
		p.Move(x, j).DrawChar(c)
	}
}

// TopLeftChar returns '┌' or ','.
func (p *InfinitePlan) TopLeftChar() uint16 {
	if p.IsUnicode() {
		return '┌'
	}
	return ','
}

// TopRightChar returns '┐' or '.'.
func (p *InfinitePlan) TopRightChar() uint16 {
	if p.IsUnicode() {
		return '┐'
	}
	return '.'
}

// BottomLeftChar returns '└' or '`'.
func (p *InfinitePlan) BottomLeftChar() uint16 {
	if p.IsUnicode() {
		return '└'
	}
	return '`'
}

// BottomRightChar returns '┘' or an apostrophe.
func (p *InfinitePlan) BottomRightChar() uint16 {
	if p.IsUnicode() {
		return '┘'
	}
	return '\''
}

// DrawBox draws a box of a size at the position; a size of 0 or less panics
// with IllegalArgumentException.
func (p *InfinitePlan) DrawBox(width, height int) {
	if width <= 0 || height <= 0 {
		throw(illegalArgument, "")
	}
	// top border
	p.DrawChar(p.TopLeftChar())
	if width > 2 {
		p.DrawHLine(1, width-2, 0)
	}
	if width > 1 {
		p.Move(width-1, 0).DrawChar(p.TopRightChar())
	}
	// sides
	if height > 2 {
		p.DrawVLine(1, height-2, 0)
		if width > 1 {
			p.DrawVLine(1, height-2, width-1)
		}
	}
	// bottom border
	if height > 1 {
		p.Move(0, height-1).DrawChar(p.BottomLeftChar())
		if width > 2 {
			p.DrawHLine(1, width-2, height-1)
		}
		if width > 1 {
			p.Move(width-1, height-1).DrawChar(p.BottomRightChar())
		}
	}
}

// CreateNoteBox returns the note box around text (an ANote), to draw later.
func (p *InfinitePlan) CreateNoteBox(text AsciiBlock) AsciiBlock { return NewANote(text) }
