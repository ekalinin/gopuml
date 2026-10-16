// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/iterator/CodeIteratorAffectation.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/plantuml-in-go/internal/json"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// codeIteratorAffectation handles variable assignments. A JSON value may
// span several lines: while it does not parse, the next line is appended.
type codeIteratorAffectation struct {
	abstractCodeIterator
	context *TContext
	memory  TMemory
	logs    *[]text.StringLocated
}

func (it *codeIteratorAffectation) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		if it.context.lineType(*result) == lineAffectation {
			*it.logs = append(*it.logs, *result)
			if err := it.doAffectation(*result); err != nil {
				return nil, err
			}
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		return result, nil
	}
}

func (it *codeIteratorAffectation) doAffectation(result text.StringLocated) error {
	lastLocation := -1
	for range 9999 {
		parseErr, err := it.executeAffectation(result)
		if parseErr == nil {
			return err
		}
		if parseErr.Location.Column <= lastLocation {
			return newEaterException("Error in JSON format", result)
		}
		lastLocation = parseErr.Location.Column
		if err := it.next(); err != nil {
			return err
		}
		forward, err := it.source.peek()
		if err != nil {
			return err
		}
		if forward == nil {
			throwRuntime("NullPointerException", "")
		}
		result = text.StringLocated{S: result.S + forward.S, Location: result.Location, PreprocessorError: result.PreprocessorError}
	}
	return nil
}

// executeAffectation runs the assignment; it returns the JSON parse error
// (upstream's ParseException) instead of panicking with it.
func (it *codeIteratorAffectation) executeAffectation(s text.StringLocated) (parseErr *json.ParseError, err error) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*json.ParseError); ok {
				parseErr = pe
				return
			}
			panic(r)
		}
	}()
	return nil, newEaterAffectation(s).analyze(it.context, it.memory)
}
