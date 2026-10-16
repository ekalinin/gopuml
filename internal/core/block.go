// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/BlockUmlBuilder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/BlockUml.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/ReadLineWithYamlHeader.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Block is one @start...@end block of a source.
type Block struct {
	// Lines run from the start line to the end line, without a YAML header
	// that directly follows the start line (upstream's raw source).
	Lines []text.StringLocated
}

// BuildBlocks extracts the blocks of src in order. desc describes the source
// in line locations. config lines are inserted after every start line.
func BuildBlocks(src, desc string, config []string) []*Block {
	reader := &uncommentReader{raw: newLineReader(src, desc, nil)}
	includer := &mergeLinesReader{source: &addConfigReader{raw: reader, config: config}}

	var blocks []*Block
	var current []text.StringLocated
	inBlock, paused := false, false
	for {
		line, ok := includer.ReadLine()
		if !ok {
			break
		}
		s := line.S
		if IsStartDirective(s) {
			current, inBlock, paused = nil, true, false
		}
		if IsPauseDirective(s) || IsExit(s) {
			paused = true
			reader.paused = true
		}
		switch {
		case inBlock && !paused:
			current = append(current, line)
		case paused && inBlock:
			if appended, ok := PossibleAppend(line); ok {
				current = append(current, appended)
			}
		}
		if IsUnpauseDirective(s) {
			paused = false
			reader.paused = false
		}
		if IsEndDirective(s) && inBlock {
			if paused {
				current = append(current, line)
			}
			blocks = append(blocks, &Block{Lines: removeYamlHeader(current)})
			current, inBlock = nil, false
			reader.paused = false
		}
	}
	return blocks
}

// removeYamlHeader drops a "---" delimited header that directly follows the
// start line. An unterminated header is kept.
func removeYamlHeader(lines []text.StringLocated) []text.StringLocated {
	if len(lines) < 2 || lines[1].S != "---" {
		return lines
	}
	for i := 2; i < len(lines); i++ {
		if lines[i].S == "---" {
			return append([]text.StringLocated{lines[0]}, lines[i+1:]...)
		}
	}
	return lines
}

var threeLetterExtension = regexp.MustCompile(`\.\w\w\w$`)

// Name returns the diagram name written after the start tag, for example
// "login" for "@startuml login". ok is false when the block has no name.
//
// Upstream reads the name from the preprocessed start line; the raw start
// line is used here, which differs only when the preprocessor rewrites it.
func (b *Block) Name() (name string, ok bool) {
	if len(b.Lines) == 0 {
		return "", false
	}
	m := patternFilename.FindStringSubmatch(text.Trin(b.Lines[0].S))
	if m == nil {
		return "", false
	}
	name, _, _ = strings.Cut(m[1], ",")
	if strings.ContainsAny(name, "<>|") {
		return "", false
	}
	name = strings.TrimPrefix(name, "file://")
	return threeLetterExtension.ReplaceAllString(name, ""), true
}
