// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/TimLoader.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/BlockUml.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// timLoader preprocesses the lines of one block.
type timLoader struct {
	context           *TContext
	global            *TMemoryGlobal
	preprocessorError *EaterException
	resultList        []text.StringLocated
}

func newTimLoader(pathSystem *PathSystem, defines *Defines, definitionsContainer DefinitionsContainer,
	location text.StringLocated) *timLoader {
	l := &timLoader{context: newTContext(pathSystem, defines, definitionsContainer), global: newTMemoryGlobal()}
	func() {
		// Errors are only logged upstream.
		defer func() { _ = recover() }()
		_ = defines.copyTo(l.global, location)
	}()
	return l
}

func (l *timLoader) load(list []text.StringLocated) {
	if err := l.execute(list); err != nil {
		e, _ := err.(*EaterException)
		l.context.resultList = append(l.context.resultList, withErrorPreprocessor(e.Location, e.Message))
		l.changeLastLine(e.Message)
		l.preprocessorError = e
	}
	l.resultList = l.context.resultList
}

// execute runs the lines. Upstream lets a runtime exception raised outside
// a line execution (for example a division by zero in "!$x = 1/0") escape
// and crash the whole render; here it is reported as "Fatal parsing error"
// at the directive being processed, the last line of the debug list.
func (l *timLoader) execute(list []text.StringLocated) (err error) {
	defer func() {
		if r := recover(); r != nil {
			location := list[0]
			if n := len(l.context.debug); n > 0 {
				location = l.context.debug[n-1]
			}
			err = newEaterException("Fatal parsing error", location)
		}
	}()
	_, err = l.context.executeLines(l.global, list, functionNone, false)
	return err
}

func withErrorPreprocessor(sl text.StringLocated, message string) text.StringLocated {
	sl.PreprocessorError = message
	return sl
}

func (l *timLoader) changeLastLine(message string) {
	if num := len(l.context.debug) - 1; num >= 0 {
		l.context.debug[num] = withErrorPreprocessor(l.context.debug[num], message)
	}
}

// BlockDefinitions is the DefinitionsContainer of one source: the
// preprocessed lines of its blocks, in order (upstream's BlockUmlBuilder).
// Add each block after preprocessing it, so that !includedef sees the
// @startdef blocks that precede it.
type BlockDefinitions struct {
	blocks [][]text.StringLocated
}

// Add records the preprocessed lines of a block.
func (d *BlockDefinitions) Add(lines []text.StringLocated) { d.blocks = append(d.blocks, lines) }

// GetDefinition returns the lines of the first block whose start line is
// "@startdef(id=<name>)", without its start and end lines.
func (d *BlockDefinitions) GetDefinition(name string) []string {
	signature := "@startdef(id=" + name + ")"
	for _, data := range d.blocks {
		if len(data) > 0 && strings.EqualFold(data[0].S, signature) {
			if len(data) < 2 {
				throwRuntime("IllegalArgumentException", "")
			}
			out := make([]string, 0, len(data)-2)
			for _, s := range data[1 : len(data)-1] {
				out = append(out, s.S)
			}
			return out
		}
	}
	return nil
}
