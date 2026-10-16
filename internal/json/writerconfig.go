// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/WriterConfig.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

// WriterConfig controls the formatting of the JSON text. Use [Minimal],
// [Pretty] or a [*PrettyPrint].
type WriterConfig interface {
	createWriter(buffer *writingBuffer) jsonWriter
}

// minimalConfig is the configuration of [Minimal].
type minimalConfig struct{}

func (minimalConfig) createWriter(buffer *writingBuffer) jsonWriter {
	return &minimalWriter{writer: buffer}
}

var (
	// Minimal writes JSON in its minimal form, without any additional
	// whitespace. This is the default.
	Minimal WriterConfig = minimalConfig{}

	// Pretty writes JSON pretty-printed, with each value on a separate line
	// and an indentation of two spaces.
	Pretty WriterConfig = newPrettyPrint("  ", true)
)
