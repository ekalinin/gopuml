// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorWhile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorForeach.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// codeIteratorWhile handles !while and !endwhile.
type codeIteratorWhile struct {
	abstractCodeIterator
	context *TContext
	memory  TMemory
	logs    *[]text.StringLocated
}

func (it *codeIteratorWhile) peek() (*text.StringLocated, error) {
	level := 0
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		typ := it.context.lineType(*result)
		currentWhile := it.memory.peekWhile()
		if currentWhile != nil && currentWhile.skipMe {
			switch typ {
			case lineWhile:
				level++
			case lineEndwhile:
				level--
				if level == -1 {
					it.memory.pollWhile()
					level = 0
				}
			}
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		switch typ {
		case lineWhile:
			*it.logs = append(*it.logs, *result)
			if err := it.executeWhile(result.Trimmed()); err != nil {
				return nil, err
			}
		case lineEndwhile:
			*it.logs = append(*it.logs, *result)
			if currentWhile == nil {
				return nil, newEaterException("No while related to this endwhile", *result)
			}
			value, err := currentWhile.conditionValue(*result, it.context, it.memory)
			if err != nil {
				return nil, err
			}
			if value.ToBoolean() {
				if err := it.source.jumpToCodePosition(currentWhile.codePosition, *result); err != nil {
					return nil, err
				}
			} else {
				it.memory.pollWhile()
			}
		default:
			return result, nil
		}
		if err := it.next(); err != nil {
			return nil, err
		}
	}
}

func (it *codeIteratorWhile) executeWhile(s text.StringLocated) error {
	condition := newEaterWhile(s)
	if err := condition.analyze(it.context, it.memory); err != nil {
		return err
	}
	theWhile := executionContextWhileFromValue(condition.getWhileExpression(), it.source.getCodePosition())
	value, err := theWhile.conditionValue(s, it.context, it.memory)
	if err != nil {
		return err
	}
	if !value.ToBoolean() {
		theWhile.skipMe = true
	}
	it.memory.addWhile(theWhile)
	return nil
}

// codeIteratorForeach handles !foreach and !endfor.
type codeIteratorForeach struct {
	abstractCodeIterator
	context *TContext
	memory  TMemory
	logs    *[]text.StringLocated
}

func (it *codeIteratorForeach) peek() (*text.StringLocated, error) {
	level := 0
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		typ := it.context.lineType(*result)
		foreach := it.memory.peekForeach()
		if foreach != nil && foreach.skipMe {
			switch typ {
			case lineForeach:
				level++
			case lineEndforeach:
				level--
				if level == -1 {
					it.memory.pollForeach()
					level = 0
				}
			}
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		switch typ {
		case lineForeach:
			*it.logs = append(*it.logs, *result)
			if err := it.executeForeach(result.Trimmed()); err != nil {
				return nil, err
			}
		case lineEndforeach:
			*it.logs = append(*it.logs, *result)
			if foreach == nil {
				return nil, newEaterException("No foreach related to this endforeach", *result)
			}
			foreach.inc()
			if foreach.skipMe {
				it.memory.pollForeach()
			} else {
				if err := it.setLoopVariable(foreach, *result); err != nil {
					return nil, err
				}
				if err := it.source.jumpToCodePosition(foreach.codePosition, *result); err != nil {
					return nil, err
				}
			}
		default:
			return result, nil
		}
		if err := it.next(); err != nil {
			return nil, err
		}
	}
}

func (it *codeIteratorForeach) executeForeach(s text.StringLocated) error {
	condition := newEaterForeach(s)
	if err := condition.analyze(it.context, it.memory); err != nil {
		return err
	}
	foreach := executionContextForeachFromValue(condition.getVarname(), condition.getJSONValue(), it.source.getCodePosition())
	if condition.isSkip() {
		foreach.skipMe = true
	} else if err := it.setLoopVariable(foreach, s); err != nil {
		return err
	}
	it.memory.addForeach(foreach)
	return nil
}

func (it *codeIteratorForeach) setLoopVariable(foreach *ExecutionContextForeach, position text.StringLocated) error {
	first := foreach.currentValue()
	return it.memory.PutVariable(foreach.varname, TValueFromJSON(first), scopeGlobal, position)
}
