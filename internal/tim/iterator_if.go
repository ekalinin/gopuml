// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorIf.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// codeIteratorIf handles !if, !ifdef, !ifndef, !elseif, !else and !endif,
// and skips the lines of false branches.
type codeIteratorIf struct {
	abstractCodeIterator
	context *TContext
	memory  TMemory
	logs    *[]text.StringLocated
}

func (it *codeIteratorIf) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		typ := it.context.lineType(*result)
		if isConditionalDirective(typ) {
			*it.logs = append(*it.logs, *result)
			if err := it.processConditionalDirective(*result, typ); err != nil {
				return nil, err
			}
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		if it.shouldSkipLine() {
			*it.logs = append(*it.logs, *result)
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		return result, nil
	}
}

func isConditionalDirective(typ TLineType) bool {
	switch typ {
	case lineIf, lineIfdef, lineIfndef, lineElse, lineElseif, lineEndif:
		return true
	}
	return false
}

func (it *codeIteratorIf) processConditionalDirective(line text.StringLocated, typ TLineType) error {
	switch typ {
	case lineIf:
		return it.executeIf(line)
	case lineIfdef:
		return it.executeIfdef(line)
	case lineIfndef:
		return it.executeIfndef(line)
	case lineElse:
		return it.executeElse(line)
	case lineElseif:
		return it.executeElseIf(line)
	case lineEndif:
		return it.executeEndif(line)
	}
	return nil
}

func (it *codeIteratorIf) shouldSkipLine() bool {
	return it.memory.peekIf() != nil && !it.memory.areAllIfOk()
}

func (it *codeIteratorIf) executeIf(s text.StringLocated) error {
	isTrue := false
	if it.memory.areAllIfOk() {
		condition := newEaterIf(s)
		if err := condition.analyze(it.context, it.memory); err != nil {
			return err
		}
		isTrue = condition.isTrue()
	}
	it.memory.addIf(executionContextIfFromValue(isTrue))
	return nil
}

func (it *codeIteratorIf) executeIfdef(s text.StringLocated) error {
	condition := newEaterIfdef(s)
	if err := condition.analyze(it.context, it.memory); err != nil {
		return err
	}
	it.memory.addIf(executionContextIfFromValue(condition.isTrue(it.context, it.memory)))
	return nil
}

func (it *codeIteratorIf) executeIfndef(s text.StringLocated) error {
	condition := newEaterIfndef(s)
	if err := condition.analyze(it.context, it.memory); err != nil {
		return err
	}
	it.memory.addIf(executionContextIfFromValue(condition.isTrue(it.context, it.memory)))
	return nil
}

func (it *codeIteratorIf) executeElseIf(s text.StringLocated) error {
	poll, err := it.getRequiredIfContext(s, "elseif")
	if err != nil {
		return err
	}
	poll.enteringElseIf()
	if !poll.hasBeenBurn {
		condition := newEaterElseIf(s)
		if err := condition.analyze(it.context, it.memory); err != nil {
			return err
		}
		if condition.isTrue() {
			poll.nowInSomeElseIf()
		}
	}
	return nil
}

func (it *codeIteratorIf) executeElse(s text.StringLocated) error {
	poll, err := it.getRequiredIfContext(s, "else")
	if err != nil {
		return err
	}
	poll.nowInElse()
	return nil
}

func (it *codeIteratorIf) executeEndif(s text.StringLocated) error {
	if it.memory.pollIf() == nil {
		return newEaterException("No if related to this endif", s)
	}
	return nil
}

func (it *codeIteratorIf) getRequiredIfContext(s text.StringLocated, directive string) (*ExecutionContextIf, error) {
	poll := it.memory.peekIf()
	if poll == nil {
		return nil, newEaterException("No if related to this "+directive, s)
	}
	return poll, nil
}
