// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/PreprocessingArtifact.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/ConfigurationStore.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/OptionKey.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"slices"
	"strings"
)

// OptionKey is a key of !option.
type OptionKey struct {
	Name         string
	defaultValue *string
}

// The option keys of upstream, in order.
var optionKeys = []OptionKey{
	{Name: "LANGUAGE"},
	{Name: "USE_DESCRIPTIVE_NAMES"},
	{Name: "HANDWRITTEN", defaultValue: new("true")},
	{Name: "DEBUG", defaultValue: new("true")},
	{Name: "SVG_DESC"},
	{Name: "SVG_TITLE"},
}

// simplifyOptionKey keeps the ASCII letters, in lower case.
func simplifyOptionKey(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch {
		case c >= 'A' && c <= 'Z':
			b.WriteRune(c + 'a' - 'A')
		case c >= 'a' && c <= 'z':
			b.WriteRune(c)
		}
	}
	return b.String()
}

// lazyFromOptionKey is OptionKey.lazyFrom.
func lazyFromOptionKey(s string) (OptionKey, bool) {
	for _, key := range optionKeys {
		if simplifyOptionKey(s) == simplifyOptionKey(key.Name) {
			return key, true
		}
	}
	return OptionKey{}, false
}

// configurationStore holds the option values, in definition order.
type configurationStore struct {
	keys   []string
	values map[string]string
}

func (c *configurationStore) define(key OptionKey, value string) {
	if c.values == nil {
		c.values = map[string]string{}
	}
	if _, ok := c.values[key.Name]; !ok {
		c.keys = append(c.keys, key.Name)
	}
	c.values[key.Name] = value
}

// PreprocessingArtifact holds what the preprocessor passes to the diagram
// besides its lines: the !option values and the warnings.
type PreprocessingArtifact struct {
	option   configurationStore
	warnings []string
}

func (a *PreprocessingArtifact) addWarning(warning string) {
	if !slices.Contains(a.warnings, warning) {
		a.warnings = append(a.warnings, warning)
	}
}

// Options returns the !option values by upstream key name (for example
// "HANDWRITTEN").
func (a *PreprocessingArtifact) Options() map[string]string {
	out := map[string]string{}
	for _, k := range a.option.keys {
		out[k] = a.option.values[k]
	}
	return out
}

// OptionValue returns the value of an option.
func (a *PreprocessingArtifact) OptionValue(name string) (string, bool) {
	v, ok := a.option.values[name]
	return v, ok
}

// OptionIsTrue is ConfigurationStore.isTrue.
func (a *PreprocessingArtifact) OptionIsTrue(name string) bool {
	return strings.EqualFold(a.option.values[name], "true")
}

// Warnings returns the warnings, in order.
func (a *PreprocessingArtifact) Warnings() []string { return slices.Clone(a.warnings) }
