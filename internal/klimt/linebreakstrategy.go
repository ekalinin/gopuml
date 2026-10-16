// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/LineBreakStrategy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "regexp"

// LineBreakStrategy is the maximum width of a text ("maxMessageSize",
// "MaximumWidth"): a number of pixels, "auto", or nothing. It is a
// comparable value.
type LineBreakStrategy struct {
	value string
	set   bool // false for upstream's null value
}

// LineBreakStrategyNone is the strategy without a value.
var LineBreakStrategyNone = LineBreakStrategy{}

// NewLineBreakStrategy returns the strategy of a value.
func NewLineBreakStrategy(value string) LineBreakStrategy {
	return LineBreakStrategy{value: value, set: true}
}

// NewLineBreakStrategyOpt returns the strategy of a value that may be
// missing (null upstream).
func NewLineBreakStrategyOpt(value string, ok bool) LineBreakStrategy {
	if !ok {
		return LineBreakStrategyNone
	}
	return NewLineBreakStrategy(value)
}

// String is upstream's toString, "null" without value.
func (l LineBreakStrategy) String() string {
	if !l.set {
		return "null"
	}
	return l.value
}

// IsAuto reports whether the value is "auto", ignoring case.
func (l LineBreakStrategy) IsAuto() bool { return l.set && javaEqualsIgnoreCase("auto", l.value) }

var signedInteger = regexp.MustCompile(`^-?[0-9]+$`)

// MaxWidth returns the width when the value is an integer, 0 otherwise.
func (l LineBreakStrategy) MaxWidth() float64 {
	if l.set && signedInteger.MatchString(l.value) {
		return javaParseDouble(l.value)
	}
	return 0
}
