// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/Movement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "strings"

// Movement is a path command with its arguments.
type Movement struct {
	letter    SvgCommandLetter
	arguments []SvgCommandNumber
}

// svgCommandIterator walks the commands of a path, as upstream's
// Iterator<SvgCommand>.
type svgCommandIterator struct {
	commands []SvgCommand
	pos      int
}

func (it *svgCommandIterator) hasNext() bool { return it.pos < len(it.commands) }

// next returns the next command; it panics past the end, as Java throws
// NoSuchElementException.
func (it *svgCommandIterator) next() SvgCommand {
	if it.pos >= len(it.commands) {
		throw(noSuchElement, "")
	}
	c := it.commands[it.pos]
	it.pos++
	return c
}

// nextLetter is next with upstream's cast to SvgCommandLetter.
func (it *svgCommandIterator) nextLetter() SvgCommandLetter {
	c := it.next()
	l, ok := c.(SvgCommandLetter)
	if !ok {
		throwSvgCommandCast(c)
	}
	return l
}

// nextNumber is next with upstream's cast to SvgCommandNumber.
func (it *svgCommandIterator) nextNumber() SvgCommandNumber {
	c := it.next()
	n, ok := c.(SvgCommandNumber)
	if !ok {
		throwSvgCommandCast(c)
	}
	return n
}

// throwSvgCommandCast panics as a failed Java cast of a path token throws
// ClassCastException.
func throwSvgCommandCast(c SvgCommand) {
	from, to := "SvgCommandNumber", "SvgCommandLetter"
	if _, ok := c.(SvgCommandLetter); ok {
		from, to = to, from
	}
	throw("java.lang.ClassCastException", "class net.sourceforge.plantuml.openiconic."+from+
		" cannot be cast to class net.sourceforge.plantuml.openiconic."+to)
}

// newMovementFrom reads a letter and its arguments from it.
func newMovementFrom(it *svgCommandIterator) *Movement {
	letter := it.nextLetter()
	nb := letter.ArgumentNumber()
	args := make([]SvgCommandNumber, nb)
	for i := range nb {
		args[i] = it.nextNumber()
	}
	return &Movement{letter: letter, arguments: args}
}

// newMovementArgs returns the movement of a letter and arguments.
func newMovementArgs(letter SvgCommandLetter, args ...SvgCommandNumber) *Movement {
	return &Movement{letter: letter, arguments: args}
}

// newMovementPositions returns the movement of a letter and points.
func newMovementPositions(letter SvgCommandLetter, pos ...SvgPosition) *Movement {
	args := make([]SvgCommandNumber, 2*len(pos))
	for i, p := range pos {
		args[2*i] = p.X()
		args[2*i+1] = p.Y()
	}
	return &Movement{letter: letter, arguments: args}
}

// Letter returns the command letter.
func (m *Movement) Letter() rune { return m.letter.Letter() }

// MutoToC turns an "S" movement into a "C" one: its first control point is
// mirrorControlPoint, or the second control point when there is none (as
// upstream does). Other letters panic, as upstream throws
// UnsupportedOperationException.
func (m *Movement) MutoToC(mirrorControlPoint *SvgPosition) *Movement {
	if !m.Is('S') {
		throw(unsupportedOperation, "")
	}
	if mirrorControlPoint == nil {
		return newMovementPositions(NewSvgCommandLetter('C'), m.SvgPosition(0), m.SvgPosition(0),
			*m.LastPosition())
	}
	return newMovementPositions(NewSvgCommandLetter('C'), *mirrorControlPoint, m.SvgPosition(0),
		*m.LastPosition())
}

// ToSvg returns the letter and the arguments, each followed by a space.
func (m *Movement) ToSvg() string {
	var sb strings.Builder
	sb.WriteString(m.letter.ToSvg())
	sb.WriteByte(' ')
	for _, arg := range m.arguments {
		sb.WriteString(arg.ToSvg())
		sb.WriteByte(' ')
	}
	return sb.String()
}

// SvgPosition returns the point of the arguments index and index+1. It
// panics past the arguments, as Java throws IndexOutOfBoundsException.
func (m *Movement) SvgPosition(index int) SvgPosition {
	return NewSvgPositionOf(m.argument(index), m.argument(index+1))
}

func (m *Movement) argument(index int) SvgCommandNumber {
	if index < 0 || index >= len(m.arguments) {
		throw("java.lang.ArrayIndexOutOfBoundsException",
			"Index "+itoa(index)+" out of bounds for length "+itoa(len(m.arguments)))
	}
	return m.arguments[index]
}

// Argument returns the argument index.
func (m *Movement) Argument(index int) float64 { return m.argument(index).Double() }

// LastPosition returns the end point: the last two arguments, nil for a
// command without arguments (Z).
func (m *Movement) LastPosition() *SvgPosition {
	if m.letter.ArgumentNumber() == 0 {
		return nil
	}
	p := m.SvgPosition(len(m.arguments) - 2)
	return &p
}

// ToAbsoluteUpperCase returns the movement in absolute coordinates, delta
// being the current point: H and V become L, relative letters become upper
// case.
func (m *Movement) ToAbsoluteUpperCase(delta SvgPosition) *Movement {
	switch {
	case m.letter.Is('H'):
		tmp := m.argument(0)
		return newMovementPositions(NewSvgCommandLetter('L'), NewSvgPositionOf(tmp, delta.Y()))
	case m.letter.Is('V'):
		tmp := m.argument(0)
		return newMovementPositions(NewSvgCommandLetter('L'), NewSvgPositionOf(delta.X(), tmp))
	case m.letter.Is('h'):
		tmp := m.argument(0)
		return newMovementPositions(NewSvgCommandLetter('L'), NewSvgPositionOf(delta.X().Add(tmp), delta.Y()))
	case m.letter.Is('v'):
		tmp := m.argument(0)
		return newMovementPositions(NewSvgCommandLetter('L'), NewSvgPositionOf(delta.X(), delta.Y().Add(tmp)))
	}

	if m.letter.IsUpperCase() {
		return m
	}

	switch {
	case m.letter.Is('m'):
		return newMovementPositions(NewSvgCommandLetter('M'), delta.Add(m.SvgPosition(0)))
	case m.letter.Is('l'):
		return newMovementPositions(NewSvgCommandLetter('L'), delta.Add(m.SvgPosition(0)))
	case m.letter.Is('t'):
		return newMovementPositions(NewSvgCommandLetter('T'), delta.Add(m.SvgPosition(0)))
	case m.letter.Is('z'):
		return newMovementArgs(NewSvgCommandLetter('Z'))
	case m.letter.Is('c'):
		return newMovementPositions(NewSvgCommandLetter('C'), delta.Add(m.SvgPosition(0)),
			delta.Add(m.SvgPosition(2)), delta.Add(m.SvgPosition(4)))
	case m.letter.Is('q'):
		return newMovementPositions(NewSvgCommandLetter('Q'), delta.Add(m.SvgPosition(0)),
			delta.Add(m.SvgPosition(2)))
	case m.letter.Is('s'):
		return newMovementPositions(NewSvgCommandLetter('S'), delta.Add(m.SvgPosition(0)),
			delta.Add(m.SvgPosition(2)))
	case m.letter.Is('a'):
		last := delta.Add(*m.LastPosition())
		return newMovementArgs(NewSvgCommandLetter('A'), m.argument(0), m.argument(1), m.argument(2),
			m.argument(3), m.argument(4), last.X(), last.Y())
	}
	throw(unsupportedOperation, "Movement::goUpperCase "+m.letter.String())
	return nil
}

// MirrorControlPoint returns the reflection of the last control point
// through the end point for C and S, nil for other letters. Relative c and s
// panic, as upstream throws IllegalStateException.
func (m *Movement) MirrorControlPoint() *SvgPosition {
	if m.letter.Is('c') || m.letter.Is('s') {
		throw(illegalState, "")
	}
	if m.letter.Is('C') {
		center := m.LastPosition()
		controlPoint := m.SvgPosition(2)
		p := center.Mirror(controlPoint)
		return &p
	}
	if m.letter.Is('S') {
		center := m.LastPosition()
		controlPoint := m.SvgPosition(0)
		p := center.Mirror(controlPoint)
		return &p
	}
	return nil
}

// Is reports whether the movement has the letter r.
func (m *Movement) Is(r rune) bool { return m.letter.Is(r) }
