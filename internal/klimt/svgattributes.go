// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/SvgAttributes.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"maps"
	"slices"

	"github.com/ekalinin/gopuml/internal/command/regex"
)

// SvgAttributes are extra SVG attributes of a text, sorted by name
// (upstream's TreeMap). It is immutable; Add returns a new value.
type SvgAttributes struct {
	attributes map[string]string
}

// SvgAttributesEmpty returns no attributes.
func SvgAttributesEmpty() *SvgAttributes { return &SvgAttributes{attributes: map[string]string{}} }

func (a *SvgAttributes) cloneMe() *SvgAttributes {
	return &SvgAttributes{attributes: maps.Clone(a.attributes)}
}

var svgAttributePattern = regex.MustCompile(`(\w+)\s*=\s*([%g][^%g]*[%g]|(?:\w+))`)

// SvgAttributesBuild parses "name=value" pairs, values being words or
// quoted strings.
func SvgAttributesBuild(args string) *SvgAttributes {
	result := SvgAttributesEmpty()
	for _, m := range svgAttributePattern.FindAllStringSubmatch(args, -1) {
		result.attributes[m[1]] = eventuallyRemoveStartingAndEndingDoubleQuote(m[2])
	}
	return result
}

// Attributes returns the attributes; the map must not be modified.
func (a *SvgAttributes) Attributes() map[string]string { return a.attributes }

// Keys returns the attribute names in upstream's order (sorted).
func (a *SvgAttributes) Keys() []string { return slices.Sorted(maps.Keys(a.attributes)) }

// Add returns the attributes with one more.
func (a *SvgAttributes) Add(key, value string) *SvgAttributes {
	result := a.cloneMe()
	result.attributes[key] = value
	return result
}

// AddAll returns the attributes with those of toBeAdded.
func (a *SvgAttributes) AddAll(toBeAdded *SvgAttributes) *SvgAttributes {
	result := a.cloneMe()
	maps.Copy(result.attributes, toBeAdded.attributes)
	return result
}
