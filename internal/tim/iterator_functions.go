// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorReturnFunction.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorProcedure.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorLegacyDefine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// codeIteratorReturnFunction records !function definitions.
type codeIteratorReturnFunction struct {
	abstractCodeIterator
	functionsSet *FunctionsSet
	context      *TContext
	memory       TMemory
	logs         *[]text.StringLocated
}

func (it *codeIteratorReturnFunction) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		if p := it.functionsSet.pendingFunction(); p != nil && p.FunctionType() == functionReturn {
			*it.logs = append(*it.logs, *result)
			if it.context.lineType(*result) == lineEndFunction {
				if !p.doesContainReturn() {
					return nil, newEaterException("This function does not have any !return directive. Declare it as a procedure instead ?", *result)
				}
				it.functionsSet.executeEndfunction()
			} else if err := p.addBody(it.context, *result); err != nil {
				return nil, err
			}
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		if it.context.lineType(*result) == lineDeclareReturnFunction {
			*it.logs = append(*it.logs, *result)
			if err := it.functionsSet.executeDeclareReturnFunction(it.context, it.memory, *result); err != nil {
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

// codeIteratorProcedure records !procedure and !definelong definitions.
type codeIteratorProcedure struct {
	abstractCodeIterator
	functionsSet *FunctionsSet
	context      *TContext
	memory       TMemory
	logs         *[]text.StringLocated
}

func (it *codeIteratorProcedure) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		if p := it.functionsSet.pendingFunction(); p != nil &&
			(p.FunctionType() == functionProcedure || p.FunctionType() == functionLegacyDefineLong) {
			*it.logs = append(*it.logs, *result)
			if it.context.lineType(*result) == lineEndFunction {
				it.functionsSet.executeEndfunction()
			} else if err := p.addBody(it.context, *result); err != nil {
				return nil, err
			}
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		if it.context.lineType(*result) == lineDeclareProcedure {
			*it.logs = append(*it.logs, *result)
			if err := it.functionsSet.executeDeclareProcedure(it.context, it.memory, *result); err != nil {
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

// codeIteratorLegacyDefine handles !define NAME(args) and !definelong.
type codeIteratorLegacyDefine struct {
	abstractCodeIterator
	functionsSet *FunctionsSet
	context      *TContext
	memory       TMemory
	logs         *[]text.StringLocated
}

func (it *codeIteratorLegacyDefine) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		switch it.context.lineType(*result) {
		case lineLegacyDefine:
			*it.logs = append(*it.logs, *result)
			if err := it.functionsSet.executeLegacyDefine(it.context, it.memory, *result); err != nil {
				return nil, err
			}
		case lineLegacyDefinelong:
			*it.logs = append(*it.logs, *result)
			if err := it.functionsSet.executeLegacyDefineLong(it.context, it.memory, *result); err != nil {
				return nil, err
			}
		default:
			return result, nil
		}
		if err := it.next(); err != nil {
			return nil, err
		}
	}
}
