// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/Value.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/ValueAbstract.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/ValueColor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/ValueImpl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/ValueNull.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/DarkString.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/AutomaticCounter.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/style/AutomaticCounterBasic.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Value is the value of a style property, read as the type the property
// needs.
type Value interface {
	// AsString returns the text, "" for a missing regular value.
	AsString() string
	AsColor(set *color.HColorSet) color.HColor
	AsInt() int
	AsIntButMinusOneIfError() int
	AsDouble() float64
	AsDoubleDefaultTo(defaultValue float64) float64
	AsBoolean() bool
	// AsFontFace returns the face (weight and italic axis) of the value.
	AsFontFace() font.UFontFace
	// AsHorizontalAlignment returns the alignment, HorizontalAlignmentNull
	// when the value is not one.
	AsHorizontalAlignment() geom.HorizontalAlignment
	Priority() int
	String() string
}

// HorizontalAlignmentNull stands for upstream's null HorizontalAlignment.
const HorizontalAlignmentNull geom.HorizontalAlignment = -1

// horizontalAlignmentFromString is upstream's HorizontalAlignment.fromString:
// HorizontalAlignmentNull when s is not an alignment.
func horizontalAlignmentFromString(s string) geom.HorizontalAlignment {
	a, ok := geom.HorizontalAlignmentFromString(s)
	if !ok {
		return HorizontalAlignmentNull
	}
	return a
}

// valueAbstract holds upstream's ValueAbstract defaults, which throw
// UnsupportedOperationException; name is the Java class.
type valueAbstract struct{ name string }

func (v valueAbstract) unsupported() { throw(unsupportedOperation, "Class=class "+v.name) }

func (v valueAbstract) AsString() string {
	v.unsupported()
	return ""
}

func (v valueAbstract) AsColor(*color.HColorSet) color.HColor {
	v.unsupported()
	return nil
}

func (v valueAbstract) AsInt() int {
	v.unsupported()
	return 0
}

func (v valueAbstract) AsIntButMinusOneIfError() int {
	v.unsupported()
	return 0
}

func (v valueAbstract) AsDouble() float64 {
	v.unsupported()
	return 0
}

func (v valueAbstract) AsDoubleDefaultTo(float64) float64 {
	v.unsupported()
	return 0
}

func (v valueAbstract) AsBoolean() bool {
	v.unsupported()
	return false
}

func (v valueAbstract) AsFontFace() font.UFontFace {
	v.unsupported()
	return font.UFontFace{}
}

func (v valueAbstract) AsHorizontalAlignment() geom.HorizontalAlignment {
	v.unsupported()
	return 0
}

func (v valueAbstract) Priority() int {
	v.unsupported()
	return 0
}

// ValueColor is a color value set from a command (for example "#red" on an
// element), with the priority of the value it replaces.
type ValueColor struct {
	valueAbstract
	color    color.HColor
	priority int
}

// NewValueColor returns a color value.
func NewValueColor(c color.HColor, priority int) *ValueColor {
	return &ValueColor{valueAbstract: valueAbstract{name: "net.sourceforge.plantuml.style.ValueColor"}, color: c,
		priority: priority}
}

// String is upstream's toString, the color.
func (v *ValueColor) String() string { return v.color.String() }

// AsColor returns the color.
func (v *ValueColor) AsColor(*color.HColorSet) color.HColor { return v.color }

// Priority returns the priority.
func (v *ValueColor) Priority() int { return v.priority }

// ValueNull is the value of a property that is not set.
type ValueNull struct {
	valueAbstract
}

var valueNull = &ValueNull{valueAbstract: valueAbstract{name: "net.sourceforge.plantuml.style.ValueNull"}}

// ValueNullValue returns upstream's ValueNull.NULL.
func ValueNullValue() Value { return valueNull }

// AsInt returns 0.
func (v *ValueNull) AsInt() int { return 0 }

// AsIntButMinusOneIfError returns 0.
func (v *ValueNull) AsIntButMinusOneIfError() int { return 0 }

// AsDouble returns 0.
func (v *ValueNull) AsDouble() float64 { return 0 }

// AsDoubleDefaultTo returns defaultValue.
func (v *ValueNull) AsDoubleDefaultTo(defaultValue float64) float64 { return defaultValue }

// AsBoolean returns false.
func (v *ValueNull) AsBoolean() bool { return false }

// AsString returns "".
func (v *ValueNull) AsString() string { return "" }

// AsFontFace returns the normal face.
func (v *ValueNull) AsFontFace() font.UFontFace { return font.NormalFace() }

// AsColor returns black.
func (v *ValueNull) AsColor(*color.HColorSet) color.HColor { return color.Black }

// AsHorizontalAlignment returns left.
func (v *ValueNull) AsHorizontalAlignment() geom.HorizontalAlignment {
	return geom.HorizontalAlignmentLeft
}

// String stands for Java's Object.toString.
func (v *ValueNull) String() string { return "net.sourceforge.plantuml.style.ValueNull" }

// DarkString is the regular and dark mode texts of a value, with the
// priority that decides merges (later declarations have higher
// priorities). A missing text is "" with its flag false (null upstream).
type DarkString struct {
	value1, value2 string
	has1, has2     bool
	priority       int
}

// NewDarkString returns the texts; a nil pointer is a missing text.
func NewDarkString(value1, value2 *string, priority int) DarkString {
	d := DarkString{priority: priority}
	if value1 != nil {
		d.value1, d.has1 = *value1, true
	}
	if value2 != nil {
		d.value2, d.has2 = *value2, true
	}
	return d
}

// MergeWith returns the merge of two values: the higher priority wins,
// except that a regular-only and a dark-only value combine.
func (d DarkString) MergeWith(other *DarkString) DarkString {
	if other == nil {
		return d
	}
	if (!d.has2 && !other.has2) || !d.has1 && !other.has1 {
		if d.priority > other.priority {
			return d
		}
		return *other
	}
	if !d.has2 && !other.has1 {
		return DarkString{value1: d.value1, has1: d.has1, value2: other.value2, has2: other.has2, priority: d.priority}
	}
	if !other.has2 && !d.has1 {
		return DarkString{value1: other.value1, has1: other.has1, value2: d.value2, has2: d.has2,
			priority: other.priority}
	}
	if d.priority > other.priority {
		return d
	}
	return *other
}

// AddPriority returns the value with its priority increased by delta.
func (d DarkString) AddPriority(delta int) DarkString {
	d.priority += delta
	return d
}

// String is upstream's toString, for example "red/null (3)".
func (d DarkString) String() string {
	v1, v2 := "null", "null"
	if d.has1 {
		v1 = d.value1
	}
	if d.has2 {
		v2 = d.value2
	}
	return v1 + "/" + v2 + " (" + itoa(d.priority) + ")"
}

// Value1 returns the regular text; ok is false when it is missing.
func (d DarkString) Value1() (string, bool) { return d.value1, d.has1 }

// Value2 returns the dark mode text; ok is false when it is missing.
func (d DarkString) Value2() (string, bool) { return d.value2, d.has2 }

// Priority returns the priority.
func (d DarkString) Priority() int { return d.priority }

// AutomaticCounter gives increasing priorities to the declared values.
type AutomaticCounter interface {
	NextInt() int
}

// AutomaticCounterBasic counts from 1.
type AutomaticCounterBasic struct {
	counter int
}

// NextInt returns the next number.
func (c *AutomaticCounterBasic) NextInt() int {
	c.counter++
	return c.counter
}

// ValueImpl is a value declared in a style sheet or converted from a
// skinparam.
//
// Upstream caches asInt, asIntButMinusOneIfError and asDouble in the value;
// values are shared by the read-only parsed skins here, so they are computed
// on each call instead (with the same results).
type ValueImpl struct {
	value DarkString
}

// ValueImplDark returns a dark mode value with the next priority.
func ValueImplDark(value string, counter AutomaticCounter) *ValueImpl {
	return &ValueImpl{value: DarkString{value2: value, has2: true, priority: counter.NextInt()}}
}

// ValueImplRegular returns a regular value with the next priority.
func ValueImplRegular(value string, counter AutomaticCounter) *ValueImpl {
	return &ValueImpl{value: DarkString{value1: value, has1: true, priority: counter.NextInt()}}
}

// ValueImplRegularPriority returns a regular value with a priority.
func ValueImplRegularPriority(value string, priority int) *ValueImpl {
	return &ValueImpl{value: DarkString{value1: value, has1: true, priority: priority}}
}

// MergeWith returns the merge with other (which may be nil): DarkString
// merging with another ValueImpl, the higher priority with a ValueColor.
func (v *ValueImpl) MergeWith(other Value) Value {
	if other == nil {
		return v
	}
	switch o := other.(type) {
	case *ValueImpl:
		return &ValueImpl{value: v.value.MergeWith(&o.value)}
	case *ValueColor:
		if o.Priority() > v.Priority() {
			return o
		}
		return v
	}
	throw(unsupportedOperation, "")
	return nil
}

// AddPriority returns the value with its priority increased by delta.
func (v *ValueImpl) AddPriority(delta int) *ValueImpl {
	return &ValueImpl{value: v.value.AddPriority(delta)}
}

// String is upstream's toString.
func (v *ValueImpl) String() string { return v.value.String() }

// AsString returns the regular text, "" when there is none (null
// upstream).
func (v *ValueImpl) AsString() string {
	s, _ := v.value.Value1()
	return s
}

// asStringOK returns the regular text and whether it is set.
func (v *ValueImpl) asStringOK() (string, bool) { return v.value.Value1() }

// valueString is upstream's Value.asString with its null: a value set only
// for the dark mode has no regular text (ok false).
func valueString(v Value) (string, bool) {
	if vi, ok := v.(*ValueImpl); ok {
		return vi.asStringOK()
	}
	return v.AsString(), true
}

// valueStringNonNull is valueString for the callers that dereference the
// text: a missing text panics, as upstream throws NullPointerException.
func valueStringNonNull(v Value) string {
	s, ok := valueString(v)
	if !ok {
		throw(nullPointer, "")
	}
	return s
}

// AsColor returns the color of the text: transparent for "none" and
// "transparent", white for an unknown color, with the dark mode variant.
// A dark-only value panics, as upstream throws IllegalArgumentException.
func (v *ValueImpl) AsColor(set *color.HColorSet) color.HColor {
	value1, ok1 := v.value.Value1()
	if ok1 && javaEqualsIgnoreCase("none", value1) {
		return color.Transparent()
	}
	if ok1 && javaEqualsIgnoreCase("transparent", value1) {
		return color.Transparent()
	}
	if !ok1 {
		throw(illegalArgument, v.value.String())
	}
	result := set.GetColorOrWhite(value1)
	if value2, ok2 := v.value.Value2(); ok2 {
		dark := set.GetColorOrWhite(value2)
		return result.WithDark(dark)
	}
	return result
}

// AsBoolean reports whether the text is "true", ignoring case.
func (v *ValueImpl) AsBoolean() bool {
	s, ok := v.value.Value1()
	return ok && javaEqualsIgnoreCase("true", s)
}

// AsInt returns the number made of the digits of the text, 0 without
// digits. Too many digits panic, as upstream throws NumberFormatException.
func (v *ValueImpl) AsInt() int {
	s := v.extractDigits()
	if s == "" {
		return 0
	}
	return javaParseInt(s)
}

// AsIntButMinusOneIfError returns the number made of the digits of the
// text, -1 without digits.
func (v *ValueImpl) AsIntButMinusOneIfError() int {
	s := v.extractDigits()
	if s == "" {
		return -1
	}
	return javaParseInt(s)
}

func (v *ValueImpl) extractDigits() string {
	s, ok := v.value.Value1()
	if !ok {
		throw(nullPointer, "")
	}
	var sb strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// AsDouble returns the number made of the digits and points of the text,
// NaN without them. Several points panic, as upstream throws
// NumberFormatException.
func (v *ValueImpl) AsDouble() float64 {
	s, ok := v.value.Value1()
	if !ok {
		throw(nullPointer, "")
	}
	var sb strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' || c == '.' {
			sb.WriteRune(c)
		}
	}
	if sb.Len() == 0 {
		return math.NaN()
	}
	return javaParseDouble(sb.String())
}

// AsDoubleDefaultTo returns AsDouble: upstream compares the result with NaN
// using ==, which is never true, so the default is never used.
func (v *ValueImpl) AsDoubleDefaultTo(defaultValue float64) float64 { return v.AsDouble() }

// AsFontFace returns the face of the text: bold, italic, plain or normal,
// or a CSS numeric weight; the normal face otherwise.
func (v *ValueImpl) AsFontFace() font.UFontFace {
	raw, ok := v.value.Value1()
	if !ok || raw == "" {
		return font.NormalFace()
	}
	s := javaLower(strings.TrimFunc(raw, func(r rune) bool { return r <= ' ' }))
	switch s {
	case "bold":
		return font.BoldFace()
	case "italic":
		return font.ItalicFace()
	case "plain", "normal":
		return font.NormalFace()
	}
	if fromCss, ok := font.UFontFaceFromCssWeight(s); ok {
		return fromCss
	}
	return font.NormalFace()
}

// AsHorizontalAlignment returns the alignment of the text,
// HorizontalAlignmentNull when it is not one.
func (v *ValueImpl) AsHorizontalAlignment() geom.HorizontalAlignment {
	return horizontalAlignmentFromString(v.AsString())
}

// Priority returns the priority.
func (v *ValueImpl) Priority() int { return v.value.Priority() }
