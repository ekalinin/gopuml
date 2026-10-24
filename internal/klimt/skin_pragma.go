// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/Pragma.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/PragmaKey.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/ConfigurationStore.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/OptionKey.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/warning/Warning.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/warning/WarningHandler.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/warning/JawsWarning.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"slices"
	"strings"
)

// PragmaKey is a known "!pragma" name.
type PragmaKey int

// Pragma keys, in upstream's order.
const (
	PragmaSequenceMessageSpan PragmaKey = iota
	PragmaEmulateNoGraphvizInstallation
	PragmaEmulateGraphvizCrash
	PragmaEmulateGraphviz244OnWindows
	PragmaAspect
	PragmaCompact
	PragmaDefaultLabelAngle
	PragmaDefaultLabelDistance
	PragmaEdgeCornerRadius
	PragmaGraphAttributes
	PragmaHorizontalLineBetweenDifferentPackageAllowed
	PragmaKermor
	PragmaLabelAngle
	PragmaLabelDistance
	PragmaRatio
	PragmaShowDeprecation
	PragmaSvgFont
	PragmaSvgInteractive
	PragmaSvgParser
	PragmaSvekTrace
	PragmaTeoz
	PragmaTexSystem
	PragmaTexPreamble
	PragmaUseIntermediatePackages
	PragmaUseVerticalIf
	pragmaKeyCount
)

var pragmaKeyNames = []string{"SEQUENCE_MESSAGE_SPAN", "EMULATE_NO_GRAPHVIZ_INSTALLATION",
	"EMULATE_GRAPHVIZ_CRASH", "EMULATE_GRAPHVIZ_244_ON_WINDOWS", "ASPECT", "COMPACT", "DEFAULT_LABEL_ANGLE",
	"DEFAULT_LABEL_DISTANCE", "EDGE_CORNER_RADIUS", "GRAPH_ATTRIBUTES",
	"HORIZONTAL_LINE_BETWEEN_DIFFERENT_PACKAGE_ALLOWED", "KERMOR", "LABEL_ANGLE", "LABEL_DISTANCE", "RATIO",
	"SHOW_DEPRECATION", "SVG_FONT", "SVG_INTERACTIVE", "SVG_PARSER", "SVEK_TRACE", "TEOZ", "TEX_SYSTEM",
	"TEX_PREAMBLE", "USE_INTERMEDIATE_PACKAGES", "USE_VERTICAL_IF"}

// String returns upstream's constant name.
func (k PragmaKey) String() string { return enumName("PragmaKey", pragmaKeyNames, int(k)) }

// DefaultValue returns the value of the pragma without value ("true" for
// svek_trace and teoz); ok is false when there is none.
func (k PragmaKey) DefaultValue() (string, bool) {
	if k == PragmaSvekTrace || k == PragmaTeoz {
		return "true", true
	}
	return "", false
}

// simplifyKey keeps the ASCII letters of s, lower-cased.
func simplifyKey(s string) string {
	var sb strings.Builder
	for _, c := range s {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			sb.WriteRune(c | 0x20)
		}
	}
	return sb.String()
}

// PragmaKeyLazyFrom returns the key whose name has the same letters as s,
// ignoring case, "_" and other characters; ok is false when there is none.
func PragmaKeyLazyFrom(s string) (PragmaKey, bool) {
	for i, name := range pragmaKeyNames {
		if simplifyKey(s) == simplifyKey(name) {
			return PragmaKey(i), true
		}
	}
	return 0, false
}

// Warning is a message shown with a diagram, for example about deprecated
// syntax. It compares by its lines.
type Warning struct {
	message []string
}

// NewWarning returns the warning of lines.
func NewWarning(lines ...string) *Warning { return &Warning{message: lines} }

// Message returns the lines.
func (w *Warning) Message() []string { return w.message }

// AsSingleLine returns the lines joined with newlines.
func (w *Warning) AsSingleLine() string { return strings.Join(w.message, "\n") }

// Equals reports whether the lines are equal.
func (w *Warning) Equals(other *Warning) bool { return slices.Equal(w.message, other.message) }

// WarningHandler collects warnings.
type WarningHandler interface {
	AddWarning(warning *Warning)
	Warnings() []*Warning
}

// JawsWarning is a deprecated backslash sequence of texts.
type JawsWarning int

// The deprecated sequences, in upstream's order.
const (
	JawsWarningBackslashNewline JawsWarning = iota
	JawsWarningBackslashLeft
	JawsWarningBackslashRight
	JawsWarningBackslashTabulation
	JawsWarningBackslashBackslash
)

var jawsWarningData = [...]struct{ ch, function string }{
	{"\\n", "%newline() or %n()"},
	{"\\l", "%left_align()"},
	{"\\r", "%right_align()"},
	{"\\t", "%tab()"},
	{"\\\\", "%backslash()"},
}

// ToWarning returns the warning about the sequence.
func (j JawsWarning) ToWarning() *Warning {
	d := jawsWarningData[j]
	return NewWarning("This diagram is using "+d.ch+" which is deprecated and will be removed in the future.",
		"You should use "+d.function+" instead in your diagram.")
}

// Pragma is the "!pragma" values of a diagram, and its warnings.
type Pragma struct {
	values   [pragmaKeyCount]string
	defined  [pragmaKeyCount]bool
	hasValue [pragmaKeyCount]bool // false for a defined pragma without value (null upstream)
	warnings []*Warning           // insertion order, without duplicates
}

var _ WarningHandler = (*Pragma)(nil)

// PragmaCreateEmpty returns pragmas without values.
func PragmaCreateEmpty() *Pragma { return &Pragma{} }

// Define sets a pragma by name; unknown names are ignored. A pragma without
// value (ok false) takes its default value, if any, and stays without value
// otherwise.
func (p *Pragma) Define(keyName string, value string, ok bool) {
	key, found := PragmaKeyLazyFrom(keyName)
	if !found {
		return
	}
	if !ok {
		value, ok = key.DefaultValue()
	}
	p.values[key] = value
	p.defined[key] = true
	p.hasValue[key] = ok
}

// IsDefine reports whether a pragma is set.
func (p *Pragma) IsDefine(key PragmaKey) bool { return p.defined[key] }

// Undefine removes a pragma.
func (p *Pragma) Undefine(key PragmaKey) {
	p.defined[key] = false
	p.hasValue[key] = false
	p.values[key] = ""
}

// Value returns the value of a pragma; ok is false when it is not set or
// set without value.
func (p *Pragma) Value(key PragmaKey) (string, bool) {
	if !p.defined[key] || !p.hasValue[key] {
		return "", false
	}
	return p.values[key], true
}

// IsTrue reports whether the value is "true" or "on", ignoring case.
func (p *Pragma) IsTrue(key PragmaKey) bool {
	v, ok := p.Value(key)
	return ok && (javaEqualsIgnoreCase("true", v) || javaEqualsIgnoreCase("on", v))
}

// IsFalse reports whether the value is "false" or "off", ignoring case.
func (p *Pragma) IsFalse(key PragmaKey) bool {
	v, ok := p.Value(key)
	return ok && (javaEqualsIgnoreCase("false", v) || javaEqualsIgnoreCase("off", v))
}

// PragmaLegacyReplaceBackslashNByNewline returns true: "\n" in texts is a
// newline.
func PragmaLegacyReplaceBackslashNByNewline() bool { return true }

// AddWarning records a warning (nil is ignored), once.
func (p *Pragma) AddWarning(warning *Warning) {
	if warning == nil {
		return
	}
	for _, w := range p.warnings {
		if w.Equals(warning) {
			return
		}
	}
	p.warnings = append(p.warnings, warning)
}

// Warnings returns the warnings in order.
func (p *Pragma) Warnings() []*Warning { return p.warnings }

// OptionKey is a known "!option" name.
type OptionKey int

// Option keys, in upstream's order.
const (
	OptionLanguage OptionKey = iota
	OptionUseDescriptiveNames
	OptionHandwritten
	OptionDebug
	OptionSvgDesc
	OptionSvgTitle
)

var optionKeyNames = []string{"LANGUAGE", "USE_DESCRIPTIVE_NAMES", "HANDWRITTEN", "DEBUG", "SVG_DESC", "SVG_TITLE"}

// String returns upstream's constant name.
func (k OptionKey) String() string { return enumName("OptionKey", optionKeyNames, int(k)) }

// DefaultValue returns the value of the option without value ("true" for
// handwritten and debug); ok is false when there is none.
func (k OptionKey) DefaultValue() (string, bool) {
	if k == OptionHandwritten || k == OptionDebug {
		return "true", true
	}
	return "", false
}

// OptionKeySimplify keeps the ASCII letters of s, lower-cased.
func OptionKeySimplify(s string) string { return simplifyKey(s) }

// OptionKeyLazyFrom returns the key whose name has the same letters as s;
// ok is false when there is none.
func OptionKeyLazyFrom(s string) (OptionKey, bool) {
	for i, name := range optionKeyNames {
		if simplifyKey(s) == simplifyKey(name) {
			return OptionKey(i), true
		}
	}
	return 0, false
}

// ConfigurationStore is a set of values by key, upstream's
// preproc.ConfigurationStore (the "!option" values). Upstream defines it in
// the preprocessor package; it lives here because ISkinSimple returns it.
type ConfigurationStore[K comparable] struct {
	values *orderedMap[K, string]
	nulls  map[K]bool
}

// NewConfigurationStore returns an empty store.
func NewConfigurationStore[K comparable]() *ConfigurationStore[K] {
	return &ConfigurationStore[K]{values: newOrderedMap[K, string](), nulls: map[K]bool{}}
}

// Define sets a value; ok false stores a key without value (null upstream).
func (c *ConfigurationStore[K]) Define(key K, value string, ok bool) {
	c.values.put(key, value)
	c.nulls[key] = !ok
}

// IsDefine reports whether a key is set.
func (c *ConfigurationStore[K]) IsDefine(key K) bool {
	_, ok := c.values.get(key)
	return ok
}

// IsTrue reports whether the value is "true", ignoring case.
func (c *ConfigurationStore[K]) IsTrue(key K) bool {
	v, ok := c.Value(key)
	return ok && javaEqualsIgnoreCase("true", v)
}

// Undefine removes a key.
func (c *ConfigurationStore[K]) Undefine(key K) {
	if _, ok := c.values.values[key]; !ok {
		return
	}
	delete(c.values.values, key)
	delete(c.nulls, key)
	c.values.keys = slices.DeleteFunc(c.values.keys, func(k K) bool { return k == key })
}

// Value returns the value of a key; ok is false when it is not set or set
// without value.
func (c *ConfigurationStore[K]) Value(key K) (string, bool) {
	v, ok := c.values.get(key)
	if !ok || c.nulls[key] {
		return "", false
	}
	return v, true
}
