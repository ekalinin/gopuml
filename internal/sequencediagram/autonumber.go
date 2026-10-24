// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/AutoNumber.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/sequencediagram/DottedNumber.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/uml"
)

// AutoNumber is the "autonumber" state of a sequence diagram.
type AutoNumber struct {
	running   bool
	current   *DottedNumber
	increment int
	format    *decimalFormat
	last      string
}

// Go starts numbering at startingNumber.
func (a *AutoNumber) Go(startingNumber *DottedNumber, increment int, format *decimalFormat) {
	a.running = true
	a.current = startingNumber
	a.increment = increment
	a.format = format
}

// Stop suspends numbering.
func (a *AutoNumber) Stop() { a.running = false }

// Resume resumes numbering, with a new format when format is not nil.
func (a *AutoNumber) Resume(format *decimalFormat) {
	a.running = true
	if format != nil {
		a.format = format
	}
}

// ResumeIncrement resumes numbering with a new increment, and a new format
// when format is not nil.
func (a *AutoNumber) ResumeIncrement(increment int, format *decimalFormat) {
	a.running = true
	a.increment = increment
	if format != nil {
		a.format = format
	}
}

// IncrementIntermediate increments the second to last level ("autonumber
// inc").
func (a *AutoNumber) IncrementIntermediate() {
	if a.current == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.sequencediagram.DottedNumber.incrementIntermediate()" because "this.current" is null`)
	}
	a.current.IncrementIntermediate()
}

// IncrementIntermediateAt increments level position ("autonumber inc A").
func (a *AutoNumber) IncrementIntermediateAt(position int) {
	if a.current == nil {
		uml.ThrowNPE(`Cannot invoke "net.sourceforge.plantuml.sequencediagram.DottedNumber.incrementIntermediate(int)" because "this.current" is null`)
	}
	a.current.IncrementIntermediateAt(position)
}

// NextMessageNumber returns the number of the next message and advances;
// ok is false when numbering is off.
func (a *AutoNumber) NextMessageNumber() (string, bool) {
	if !a.running {
		return "", false
	}
	a.last = a.current.Format(a.format)
	a.current.IncrementMinor(a.increment)
	return a.last, true
}

// CurrentMessageNumber returns the last number given, without bold tags
// unless formatted.
func (a *AutoNumber) CurrentMessageNumber(formatted bool) string {
	if formatted {
		return a.last
	}
	return strings.ReplaceAll(strings.ReplaceAll(a.last, "<b>", ""), "</b>", "")
}

// IsRunning reports whether numbering is on.
func (a *AutoNumber) IsRunning() bool { return a.running }

// Current returns the next number, nil before the first "autonumber".
func (a *AutoNumber) Current() *DottedNumber { return a.current }

// Increment returns the increment.
func (a *AutoNumber) Increment() int { return a.increment }

// DottedNumber is a message number with levels, such as "1.2.3".
type DottedNumber struct {
	nums       []int32
	separators []string
}

var dottedNumberParts = regexp.MustCompile(`(\d+)|(\D+)`)

// NewDottedNumber parses a number such as "10" or "1.2.1".
func NewDottedNumber(value string) *DottedNumber {
	d := &DottedNumber{}
	for _, part := range dottedNumberParts.FindAllString(value, -1) {
		if part[0] >= '0' && part[0] <= '9' {
			d.nums = append(d.nums, int32(uml.JavaParseInt(part)))
		} else {
			d.separators = append(d.separators, part)
		}
	}
	return d
}

// String is upstream's toString: the levels and their separators.
func (d *DottedNumber) String() string {
	var sb strings.Builder
	for i, n := range d.nums {
		sb.WriteString(strconv.Itoa(int(n)))
		if i < len(d.separators) {
			sb.WriteString(d.separators[i])
		}
	}
	return sb.String()
}

func (d *DottedNumber) at(i int) int32 {
	if i < 0 || i >= len(d.nums) {
		uml.ThrowIndexOutOfBounds(i, len(d.nums))
	}
	return d.nums[i]
}

// IncrementMinor adds step to the last level, with Java's int overflow.
func (d *DottedNumber) IncrementMinor(step int) {
	last := len(d.nums) - 1
	d.nums[last] = d.at(last) + int32(step)
}

// IncrementIntermediate increments the second to last level (the only one
// for single numbers) and resets the levels after it to 1.
func (d *DottedNumber) IncrementIntermediate() {
	intermediate := 0
	if len(d.nums) != 1 {
		intermediate = len(d.nums) - 2
	}
	d.IncrementIntermediateAt(intermediate)
}

// IncrementIntermediateAt increments level position and resets the levels
// after it to 1.
func (d *DottedNumber) IncrementIntermediateAt(position int) {
	newValue := d.at(position) + 1
	for i := position + 1; i < len(d.nums); i++ {
		d.nums[i] = 1
	}
	d.nums[position] = newValue
}

// Format formats the number: single numbers with the decimal format,
// dotted numbers in bold.
func (d *DottedNumber) Format(format *decimalFormat) string {
	if len(d.nums) == 1 && len(d.separators) == 0 {
		return format.format(int64(d.nums[0]))
	}
	return "<b>" + d.String() + "</b>"
}
