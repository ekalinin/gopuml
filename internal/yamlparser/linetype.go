// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/yaml/parser/YamlLineType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

// YamlLineType is the kind of a YAML line.
type YamlLineType int

// The kinds of YAML lines.
const (
	EmptyLine          YamlLineType = iota
	NoKeyOnlyText                   // text without a key
	KeyOnly                         // a key without a value
	KeyAndValue                     // a key and a scalar value
	KeyAndFlowSequence              // a key and an inline list, such as [a, b, c]
	KeyAndBlockStyle                // a key and "|"
	KeyAndFoldedStyle               // a key and ">"
	PlainElementList                // "- value"
	PlainDash                       // "-" alone
)

var yamlLineTypeNames = [...]string{"EMPTY_LINE", "NO_KEY_ONLY_TEXT", "KEY_ONLY", "KEY_AND_VALUE",
	"KEY_AND_FLOW_SEQUENCE", "KEY_AND_BLOCK_STYLE", "KEY_AND_FOLDED_STYLE", "PLAIN_ELEMENT_LIST", "PLAIN_DASH"}

// String is the name of upstream's constant.
func (t YamlLineType) String() string { return yamlLineTypeNames[t] }
