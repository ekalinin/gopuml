// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/Value.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/ValueAbstract.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/ValueColor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/ValueImpl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/ValueNull.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/DarkString.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/value/Specificity.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/AutomaticCounter.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/AutomaticCounterBasic.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
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
	Specificity() Specificity
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

func (v valueAbstract) Specificity() Specificity {
	v.unsupported()
	return Specificity{}
}

// ValueColor is a color value set from a command (for example "#red" on an
// element), with the specificity of the value it replaces.
type ValueColor struct {
	valueAbstract
	color       color.HColor
	specificity Specificity
}

// NewValueColor returns a color value.
func NewValueColor(c color.HColor, specificity Specificity) *ValueColor {
	return &ValueColor{valueAbstract: valueAbstract{name: "net.sourceforge.plantuml.style.value.ValueColor"},
		color: c, specificity: specificity}
}

// String is upstream's toString, the color.
func (v *ValueColor) String() string { return v.color.String() }

// AsColor returns the color.
func (v *ValueColor) AsColor(*color.HColorSet) color.HColor { return v.color }

// Specificity returns the specificity.
func (v *ValueColor) Specificity() Specificity { return v.specificity }

// ValueNull is the value of a property that is not set.
type ValueNull struct {
	valueAbstract
}

var valueNull = &ValueNull{valueAbstract: valueAbstract{name: "net.sourceforge.plantuml.style.value.ValueNull"}}

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
func (v *ValueNull) String() string { return "net.sourceforge.plantuml.style.value.ValueNull" }

// Specificity is how specific a style declaration is, compared tier by tier
// from the most significant: a forced override, the ancestor cascade, the
// ancestor rank (0 for the element's own level, more negative farther up),
// the number of stereotypes of the selector, and the order of the
// declaration. It is a comparable value.
type Specificity struct {
	forced          bool
	ancestorCascade bool
	ancestorRank    int
	stereotypeCount int
	order           int
}

// SpecificityAtOrder returns the specificity of a plain declaration at its
// position.
func SpecificityAtOrder(order int) Specificity { return Specificity{order: order} }

// SpecificityForcedOverride returns the specificity of an explicit override,
// which wins over anything.
func SpecificityForcedOverride() Specificity { return Specificity{forced: true} }

// WithStereotypeCount returns the specificity of a selector with count
// stereotypes.
func (s Specificity) WithStereotypeCount(count int) Specificity {
	s.stereotypeCount = count
	return s
}

// WithAncestorRank returns the specificity applied through the ancestor
// cascade at rank.
func (s Specificity) WithAncestorRank(rank int) Specificity {
	s.ancestorCascade = true
	s.ancestorRank = rank
	return s
}

// HasStereotype reports whether the selector requires a stereotype.
func (s Specificity) HasStereotype() bool { return s.stereotypeCount > 0 }

// Compare compares the tiers in order: -1, 0 or 1.
func (s Specificity) Compare(other Specificity) int {
	if c := compareBool(s.forced, other.forced); c != 0 {
		return c
	}
	if c := compareBool(s.ancestorCascade, other.ancestorCascade); c != 0 {
		return c
	}
	for _, c := range [...][2]int{{s.ancestorRank, other.ancestorRank},
		{s.stereotypeCount, other.stereotypeCount}, {s.order, other.order}} {
		if c[0] != c[1] {
			if c[0] < c[1] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}

// IsBiggerThan reports whether s is more specific than other.
func (s Specificity) IsBiggerThan(other Specificity) bool { return s.Compare(other) > 0 }

// String is upstream's toString, for example "3" or "3,stereo=1".
func (s Specificity) String() string {
	switch {
	case s.forced:
		return "override"
	case s.ancestorCascade:
		return itoa(s.order) + ",stereo=" + itoa(s.stereotypeCount) + ",ancestor=" + itoa(s.ancestorRank)
	case s.stereotypeCount != 0:
		return itoa(s.order) + ",stereo=" + itoa(s.stereotypeCount)
	}
	return itoa(s.order)
}

// DarkString is the regular and dark mode texts of a value, with the
// specificity that decides merges. A missing text is "" with its flag false
// (null upstream).
type DarkString struct {
	value1, value2 string
	has1, has2     bool
	specificity    Specificity
}

// NewDarkString returns the texts; a nil pointer is a missing text.
func NewDarkString(value1, value2 *string, specificity Specificity) DarkString {
	d := DarkString{specificity: specificity}
	if value1 != nil {
		d.value1, d.has1 = *value1, true
	}
	if value2 != nil {
		d.value2, d.has2 = *value2, true
	}
	return d
}

// MergeWith returns the merge of two values: the more specific wins, except
// that a regular-only and a dark-only value combine.
func (d DarkString) MergeWith(other *DarkString) DarkString {
	if other == nil {
		return d
	}
	if (!d.has2 && !other.has2) || !d.has1 && !other.has1 {
		if d.specificity.IsBiggerThan(other.specificity) {
			return d
		}
		return *other
	}
	if !d.has2 && !other.has1 {
		return DarkString{value1: d.value1, has1: d.has1, value2: other.value2, has2: other.has2,
			specificity: d.specificity}
	}
	if !other.has2 && !d.has1 {
		return DarkString{value1: other.value1, has1: other.has1, value2: d.value2, has2: d.has2,
			specificity: other.specificity}
	}
	if d.specificity.IsBiggerThan(other.specificity) {
		return d
	}
	return *other
}

// WithAncestorRank returns the value applied through the ancestor cascade at
// rank.
func (d DarkString) WithAncestorRank(rank int) DarkString {
	d.specificity = d.specificity.WithAncestorRank(rank)
	return d
}

// WithStereotypeCount returns the value of a selector with count
// stereotypes.
func (d DarkString) WithStereotypeCount(count int) DarkString {
	d.specificity = d.specificity.WithStereotypeCount(count)
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
	return v1 + "/" + v2 + " (" + d.specificity.String() + ")"
}

// Value1 returns the regular text; ok is false when it is missing.
func (d DarkString) Value1() (string, bool) { return d.value1, d.has1 }

// Value2 returns the dark mode text; ok is false when it is missing.
func (d DarkString) Value2() (string, bool) { return d.value2, d.has2 }

// Specificity returns the specificity.
func (d DarkString) Specificity() Specificity { return d.specificity }

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

// ValueImplDark returns a dark mode value at the next order.
func ValueImplDark(value string, counter AutomaticCounter) *ValueImpl {
	return &ValueImpl{value: DarkString{value2: value, has2: true, specificity: SpecificityAtOrder(counter.NextInt())}}
}

// ValueImplRegular returns a regular value at the next order.
func ValueImplRegular(value string, counter AutomaticCounter) *ValueImpl {
	return &ValueImpl{value: DarkString{value1: value, has1: true, specificity: SpecificityAtOrder(counter.NextInt())}}
}

// ValueImplRegularSpecificity returns a regular value with a specificity.
func ValueImplRegularSpecificity(value string, specificity Specificity) *ValueImpl {
	return &ValueImpl{value: DarkString{value1: value, has1: true, specificity: specificity}}
}

// ValueImplOf returns a value with its regular and dark texts (one of them
// may be nil) and a specificity.
func ValueImplOf(light, dark *string, specificity Specificity) *ValueImpl {
	return &ValueImpl{value: NewDarkString(light, dark, specificity)}
}

// MergeWith returns the merge with other (which may be nil): DarkString
// merging with another ValueImpl, the more specific with a ValueColor.
func (v *ValueImpl) MergeWith(other Value) Value {
	if other == nil {
		return v
	}
	switch o := other.(type) {
	case *ValueImpl:
		return &ValueImpl{value: v.value.MergeWith(&o.value)}
	case *ValueColor:
		if o.Specificity().IsBiggerThan(v.Specificity()) {
			return o
		}
		return v
	}
	throw(unsupportedOperation, "")
	return nil
}

// WithAncestorRank returns the value applied through the ancestor cascade
// at rank.
func (v *ValueImpl) WithAncestorRank(rank int) *ValueImpl {
	return &ValueImpl{value: v.value.WithAncestorRank(rank)}
}

// WithStereotypeCount returns the value of a selector with count
// stereotypes.
func (v *ValueImpl) WithStereotypeCount(count int) *ValueImpl {
	return &ValueImpl{value: v.value.WithStereotypeCount(count)}
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

// Specificity returns the specificity.
func (v *ValueImpl) Specificity() Specificity { return v.value.Specificity() }
