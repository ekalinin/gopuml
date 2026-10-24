// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/SvgPath.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/openiconic/StringDecipher.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"
	"unicode"
	"unicode/utf16"
)

// SvgPath is SVG path data ("M10 10 l5 0 z") parsed into absolute
// movements, drawn as a UPath.
type SvgPath struct {
	// http://www.w3.org/TR/SVG11/paths.html#PathDataEllipticalArcCommands
	// https://developer.mozilla.org/en-US/docs/Web/SVG/Tutorial/Paths
	// http://tutorials.jenkov.com/svg/path-element.html
	// https://www.w3.org/TR/SVG/paths.html#PathDataQuadraticBezierCommands
	movements []*Movement
	commands  []SvgCommand
	translate UTranslate
}

// NewSvgPath parses path data. Malformed data panics as upstream throws: a
// number before any letter (NullPointerException), a missing argument
// (NoSuchElementException), a letter where a number is expected
// (ClassCastException), an unknown letter (UnsupportedOperationException).
func NewSvgPath(path string, translate UTranslate) *SvgPath {
	p := &SvgPath{translate: translate}
	for _, token := range stringDecipherDecipher(path) {
		if len(token) == 1 && javaIsLetter(token[0]) {
			p.commands = append(p.commands, NewSvgCommandLetter(rune(token[0])))
		} else {
			p.commands = append(p.commands, SvgCommandNumberParse(string(utf16.Decode(token))))
		}
	}
	p.commands = svgPathInsertMissingLetter(p.commands)
	svgPathCheckArguments(p.commands)
	last := NewSvgPosition(0, 0)
	lastMove := NewSvgPosition(0, 0)
	var mirrorControlPoint *SvgPosition
	it := &svgCommandIterator{commands: p.commands}
	for it.hasNext() {
		movement := newMovementFrom(it)
		movement = movement.ToAbsoluteUpperCase(last)

		if movement.Letter() == 'Z' {
			last = lastMove
		}

		if movement.Is('S') {
			movement = movement.MutoToC(mirrorControlPoint)
		}

		p.movements = append(p.movements, movement)

		if movement.Letter() == 'M' {
			lastMove = *movement.LastPosition()
		}

		if lp := movement.LastPosition(); lp != nil {
			last = *lp
		}

		mirrorControlPoint = movement.MirrorControlPoint()
	}
	return p
}

// svgPathInsertMissingLetter repeats the last letter before numbers that
// follow a complete command ("M1 2 3 4" is "M1 2 L3 4").
func svgPathInsertMissingLetter(commands []SvgCommand) []SvgCommand {
	var result []SvgCommand
	it := &svgCommandIterator{commands: commands}
	var lastLetter *SvgCommandLetter
	for it.hasNext() {
		cmd := it.next()
		var nb int
		if _, isNumber := cmd.(SvgCommandNumber); isNumber {
			if lastLetter == nil {
				// upstream adds null, then calls argumentNumber on it
				throw(nullPointer, "")
			}
			result = append(result, *lastLetter, cmd)
			nb = lastLetter.ArgumentNumber() - 1
		} else {
			result = append(result, cmd)
			implicit := cmd.(SvgCommandLetter).Implicit()
			lastLetter = &implicit
			nb = lastLetter.ArgumentNumber()
		}
		for range max(nb, 0) {
			result = append(result, it.nextNumber())
		}
	}
	return result
}

// svgPathCheckArguments checks that each letter is followed by its numbers.
func svgPathCheckArguments(commands []SvgCommand) {
	it := &svgCommandIterator{commands: commands}
	for it.hasNext() {
		cmd := it.nextLetter()
		nb := cmd.ArgumentNumber()
		for range nb {
			it.nextNumber()
		}
	}
}

// ToSvg returns the path as an SVG path element of the absolute movements.
func (p *SvgPath) ToSvg() string {
	var sb strings.Builder
	sb.WriteString("<path d=\"")
	for _, move := range p.movements {
		sb.WriteString(move.ToSvg())
		sb.WriteByte(' ')
	}
	sb.WriteString("\"/>")
	return sb.String()
}

// toUPathFactor is upstream's toUPath(double, double): the path scaled. As
// upstream, a T movement panics (NullPointerException): the previous
// movement is never set.
func (p *SvgPath) toUPathFactor(factorx, factory float64) *UPath {
	result := UPathNone()
	var previous *Movement
	for _, move := range p.movements {
		letter := move.Letter()
		position := move.LastPosition()
		switch letter {
		case 'M':
			result.MoveTo(float64(position.XDouble()*factorx), float64(position.YDouble()*factory))
		case 'C':
			ctl1 := move.SvgPosition(0)
			ctl2 := move.SvgPosition(2)
			result.CubicTo(float64(ctl1.XDouble()*factorx), float64(ctl1.YDouble()*factory),
				float64(ctl2.XDouble()*factorx), float64(ctl2.YDouble()*factory),
				float64(position.XDouble()*factorx), float64(position.YDouble()*factory))
		case 'Q':
			ctl := move.SvgPosition(0)
			result.CubicTo(float64(ctl.XDouble()*factorx), float64(ctl.YDouble()*factory),
				float64(ctl.XDouble()*factorx), float64(ctl.YDouble()*factory),
				float64(position.XDouble()*factorx), float64(position.YDouble()*factory))
		case 'T':
			if previous == nil {
				throw(nullPointer, "")
			}
			if previous.Letter() != 'Q' {
				throw(illegalArgument, "")
			}
			// https://stackoverflow.com/questions/5287559/calculating-control-points-for-a-shorthand-smooth-svg-path-bezier-curve
			lastCtl := previous.SvgPosition(0)
			lastP := previous.LastPosition()
			ctl := lastP.Mirror(lastCtl)
			result.CubicTo(float64(ctl.XDouble()*factorx), float64(ctl.YDouble()*factory),
				float64(ctl.XDouble()*factorx), float64(ctl.YDouble()*factory),
				float64(position.XDouble()*factorx), float64(position.YDouble()*factory))
		case 'L':
			result.LineTo(float64(position.XDouble()*factorx), float64(position.YDouble()*factory))
		case 'A':
			rx := move.Argument(0)
			ry := move.Argument(1)
			xAxisRotation := move.Argument(2)
			largeArcFlag := move.Argument(3)
			sweepFlag := move.Argument(4)
			result.ArcTo(float64(rx*factorx), float64(ry*factory), xAxisRotation, largeArcFlag, sweepFlag,
				float64(position.XDouble()*factorx), float64(position.YDouble()*factory))
		case 'Z':
			result.ClosePath()
		default:
			throw(unsupportedOperation, "letter "+string(letter))
		}
	}
	return result.Translate(float64(p.translate.Dx()*factorx), float64(p.translate.Dy()*factory))
}

// toUPathAffine is upstream's toUPath(XAffineTransform): the path
// transformed by at, arc radii scaled by its scale factors.
func (p *SvgPath) toUPathAffine(at *XAffineTransform) *UPath {
	result := UPathNone()
	var previous *Movement
	for _, move := range p.movements {
		letter := move.Letter()
		position := move.LastPosition()
		switch letter {
		case 'M':
			result.MoveToPoint(position.Affine(at))
		case 'C':
			ctl1 := move.SvgPosition(0)
			ctl2 := move.SvgPosition(2)
			result.CubicToPoints(ctl1.Affine(at), ctl2.Affine(at), position.Affine(at))
		case 'Q':
			ctl := move.SvgPosition(0)
			result.CubicToPoints(ctl.Affine(at), ctl.Affine(at), position.Affine(at))
		case 'T':
			if previous == nil {
				throw(nullPointer, "")
			}
			switch previous.Letter() {
			case 'Q':
				// https://stackoverflow.com/questions/5287559/calculating-control-points-for-a-shorthand-smooth-svg-path-bezier-curve
				lastCtl := previous.SvgPosition(0)
				lastP := previous.LastPosition()
				ctl := lastP.Mirror(lastCtl)
				result.CubicToPoints(ctl.Affine(at), ctl.Affine(at), position.Affine(at))
			case 'T':
				// Same logic as for Q followed by T: reflect the previous control point
				lastCtl := previous.SvgPosition(0)
				lastP := previous.LastPosition()
				ctl := lastP.Mirror(lastCtl)
				result.CubicToPoints(ctl.Affine(at), ctl.Affine(at), position.Affine(at))
			default:
				throw(illegalArgument, string(previous.Letter()))
			}
		case 'L':
			result.LineToPoint(position.Affine(at))
		case 'A':
			rx := move.Argument(0)
			ry := move.Argument(1)
			xAxisRotation := move.Argument(2)
			largeArcFlag := move.Argument(3)
			sweepFlag := move.Argument(4)
			tmp := position.Affine(at)
			result.ArcTo(float64(rx*at.ScaleX()), float64(ry*at.ScaleY()), xAxisRotation, largeArcFlag, sweepFlag,
				tmp.X, tmp.Y)
		case 'Z':
			result.ClosePath()
		default:
			throw(unsupportedOperation, "letter "+string(letter))
		}
		previous = move
	}
	return result.Translate(float64(p.translate.Dx()*at.ScaleX()), float64(p.translate.Dy()*at.ScaleY()))
}

// DrawMe draws the path transformed by at, unless it only moves.
func (p *SvgPath) DrawMe(ug UGraphic, at *XAffineTransform) {
	path := p.toUPathAffine(at)
	if !path.IsInvisible() {
		ug.Draw(path)
	}
}

// DrawMeFactor draws the path scaled by factor (upstream's drawMe(UGraphic,
// double)).
func (p *SvgPath) DrawMeFactor(ug UGraphic, factor float64) {
	ug.Draw(p.toUPathFactor(factor, factor))
}

// stringDecipherDecipher is upstream's StringDecipher.decipher: the tokens of
// path data, each a command letter or the text of a number, as UTF-16 code
// units. Whitespace and commas separate tokens; a character that starts
// neither a letter nor a number ends the parsing.
func stringDecipherDecipher(path string) [][]uint16 {
	input := utf16.Encode([]rune(path))
	var numbers [][]uint16

	for len(input) > 0 {
		// Step 1: Skip any separator characters (whitespace or commas)
		for len(input) > 0 && (javaIsWhitespace(rune(input[0])) || input[0] == ',') {
			input = input[1:]
		}

		if len(input) == 0 {
			break
		}

		// Step 2: If the next character is a command letter (e.g. M, L, C...),
		// isolate it
		if javaIsLetter(input[0]) {
			numbers = append(numbers, input[:1])
			input = input[1:]
			continue
		}

		i := 0

		// Step 3: Handle optional sign for the mantissa
		c := input[i]
		if c == '+' || c == '-' {
			i++
		}

		// Step 4: Read mantissa (digits with optional dot, only one allowed)
		seenDot := false
		seenMantissaDigit := false
		for i < len(input) {
			c = input[i]
			if javaIsDigit(c) {
				seenMantissaDigit = true
				i++
			} else if c == '.' && !seenDot {
				seenDot = true
				i++
			} else {
				break
			}
		}

		// If no digit has been read, it's not a valid number
		if !seenMantissaDigit {
			break
		}

		// Step 5: Optionally parse scientific notation (e.g. 1.2e-3)
		if i < len(input) && (input[i] == 'e' || input[i] == 'E') {
			expStart := i
			i++ // consume 'e' or 'E'

			// Optional sign in exponent
			if i < len(input) && (input[i] == '+' || input[i] == '-') {
				i++
			}

			seenExpDigit := false
			for i < len(input) && javaIsDigit(input[i]) {
				seenExpDigit = true
				i++
			}

			// If exponent is not valid (no digit), backtrack
			if !seenExpDigit {
				i = expStart
			}
		}

		// Step 6: Store the parsed numeric token
		numbers = append(numbers, input[:i])

		// Step 7: Move forward in the input
		input = input[i:]
	}

	return numbers
}

// javaIsDigit is Java's Character.isDigit(char): a decimal digit of any
// script (category Nd).
func javaIsDigit(ch uint16) bool { return !utf16.IsSurrogate(rune(ch)) && unicode.IsDigit(rune(ch)) }
