// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/StringUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
)

// Java exception classes thrown by the ported code. A command that panics
// with *color.RuntimeError is reported as upstream's ProtectedCommand
// reports the exception: "You should send a mail ... " followed by the
// exception as Throwable.toString prints it.
const (
	NullPointerException            = "java.lang.NullPointerException"
	IllegalArgumentException        = "java.lang.IllegalArgumentException"
	IllegalStateException           = "java.lang.IllegalStateException"
	NumberFormatException           = "java.lang.NumberFormatException"
	IndexOutOfBoundsException       = "java.lang.IndexOutOfBoundsException"
	ArrayIndexOutOfBoundsException  = "java.lang.ArrayIndexOutOfBoundsException"
	StringIndexOutOfBoundsException = "java.lang.StringIndexOutOfBoundsException"
	UnsupportedOperationException   = "java.lang.UnsupportedOperationException"
	ClassCastException              = "java.lang.ClassCastException"
	NegativeArraySizeException      = "java.lang.NegativeArraySizeException"
	OutOfMemoryError                = "java.lang.OutOfMemoryError"
	NoSuchElementException          = "java.util.NoSuchElementException"
	EmptyStackException             = "java.util.EmptyStackException"
	ArithmeticException             = "java.lang.ArithmeticException"
)

// Throw panics with a Java exception of class with message ("" for none).
func Throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }

// ThrowNPE panics with the helpful NullPointerException message of Java 14
// and later, for example `Cannot invoke "String.length()" because "s" is
// null`.
func ThrowNPE(message string) { Throw(NullPointerException, message) }

// ThrowIndexOutOfBounds panics as Java's List.get with a bad index.
func ThrowIndexOutOfBounds(index, length int) {
	Throw(IndexOutOfBoundsException, "Index "+strconv.Itoa(index)+" out of bounds for length "+strconv.Itoa(length))
}

// JavaParseInt is Java's Integer.parseInt for the decimal strings the
// command patterns capture: an optional sign and ASCII digits. Anything
// else, and values outside the int range, panic with Java's
// NumberFormatException.
func JavaParseInt(s string) int {
	fail := func() { Throw(NumberFormatException, `For input string: "`+s+`"`) }
	digits := s
	if digits != "" && (digits[0] == '-' || digits[0] == '+') {
		digits = digits[1:]
	}
	if digits == "" {
		fail()
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			fail()
		}
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		fail()
	}
	return int(v)
}

// JavaParseDouble is Java's Double.parseDouble for the strings of digits
// and dots that the scale commands capture.
func JavaParseDouble(s string) float64 {
	if strings.Count(s, ".") > 1 {
		Throw(NumberFormatException, "multiple points")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			return v
		}
		Throw(NumberFormatException, `For input string: "`+s+`"`)
	}
	return v
}

// LowerCase is upstream's StringUtils.goLowerCase (toLowerCase with the
// English locale) for the ASCII keywords the commands capture.
func LowerCase(s string) string { return strings.ToLower(s) }

// UpperCase is upstream's StringUtils.goUpperCase for the ASCII keywords the
// commands capture.
func UpperCase(s string) string { return strings.ToUpper(s) }

// EqualsIgnoreCase is Java's String.equalsIgnoreCase.
func EqualsIgnoreCase(a, b string) bool {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	if len(ua) != len(ub) {
		return false
	}
	for i := range ua {
		ca, cb := rune(ua[i]), rune(ub[i])
		if ca == cb {
			continue
		}
		la, lb := []rune(strings.ToUpper(string(ca))), []rune(strings.ToUpper(string(cb)))
		if len(la) == 1 && len(lb) == 1 && la[0] == lb[0] {
			continue
		}
		la, lb = []rune(strings.ToLower(string(la[0]))), []rune(strings.ToLower(string(lb[0])))
		if len(la) == 1 && len(lb) == 1 && la[0] == lb[0] {
			continue
		}
		return false
	}
	return true
}

// IsEmpty is upstream's StringUtils.isEmpty: s is null (ok false), empty, or
// made of spaces, tabulations, line ends and NUL characters.
func IsEmpty(s string, ok bool) bool {
	if !ok {
		return true
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\r', '\n', 0:
		default:
			return false
		}
	}
	return true
}

// RemoveQuotes is upstream's
// StringUtils.eventuallyRemoveStartingAndEndingDoubleQuote.
func RemoveQuotes(s string) string { return text.RemoveQuotes(s) }

var splitComma = Compile(`([%pLN_.]+|[%g][^%g]+[%g])`)

// SplitComma is upstream's StringUtils.splitComma: the names and quoted
// names of s, without their quotes. As upstream, characters that cannot be
// part of a name, such as '@', separate names.
func SplitComma(s string) []string {
	s = text.Trin(s)
	var result []string
	for _, m := range splitComma.FindAllString(s, -1) {
		result = append(result, RemoveQuotes(m))
	}
	return result
}

// JavaLength is Java's String.length: the number of UTF-16 code units.
func JavaLength(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// ArrowDirection is upstream's StringUtils.getArrowDirection: the
// direction of an arrow ending with '>' (named in its body), or starting
// with '<' (LEFT for two characters, UP otherwise).
func ArrowDirection(s string) klimt.Direction {
	if strings.HasSuffix(s, ">") {
		return QueueDirection(s[:len(s)-1])
	}
	if strings.HasPrefix(s, "<") {
		if JavaLength(s) == 2 {
			return klimt.DirectionLeft
		}
		return klimt.DirectionUp
	}
	Throw(IllegalArgumentException, s)
	return klimt.DirectionDown
}

var (
	arrowWordChars = regexp.MustCompile(`\w*`)
	arrowDashes    = regexp.MustCompile(`-+`)
)

// ManageArrowForCuca is upstream's StringUtils.manageArrowForCuca: the
// arrow without its direction names, one dash long from left to right, and
// at least two dashes long from top to bottom.
func ManageArrowForCuca(s string) string {
	dir := ArrowDirection(s)
	s = strings.ReplaceAll(s, "=", "-")
	s = arrowWordChars.ReplaceAllString(s, "")
	if dir == klimt.DirectionLeft || dir == klimt.DirectionRight {
		s = arrowDashes.ReplaceAllString(s, "-")
	}
	if JavaLength(s) == 2 && (dir == klimt.DirectionUp || dir == klimt.DirectionDown) {
		s = strings.Replace(s, "-", "--", 1)
	}
	return s
}

// QueueDirection is upstream's StringUtils.getQueueDirection: the
// direction named in an arrow body, DOWN by default (RIGHT for a single
// character).
func QueueDirection(s string) klimt.Direction {
	if strings.ContainsAny(s, "<>") {
		Throw(IllegalArgumentException, s)
	}
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "left"):
		return klimt.DirectionLeft
	case strings.Contains(s, "right"):
		return klimt.DirectionRight
	case strings.Contains(s, "up"):
		return klimt.DirectionUp
	case strings.Contains(s, "down"):
		return klimt.DirectionDown
	case strings.Contains(s, "l"):
		return klimt.DirectionLeft
	case strings.Contains(s, "r"):
		return klimt.DirectionRight
	case strings.Contains(s, "u"):
		return klimt.DirectionUp
	case strings.Contains(s, "d"):
		return klimt.DirectionDown
	case JavaLength(s) == 1:
		return klimt.DirectionRight
	}
	return klimt.DirectionDown
}
