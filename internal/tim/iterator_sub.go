// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorSub.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// codeIteratorSub records !startsub ... !endsub blocks and then lets their
// lines through.
type codeIteratorSub struct {
	abstractCodeIterator
	subs              map[string]*Sub
	readingInProgress codeIterator
	memory            TMemory
	context           *TContext
}

func (it *codeIteratorSub) peek() (*text.StringLocated, error) {
	if it.readingInProgress != nil {
		return it.readingInProgress.peek()
	}
	result, err := it.source.peek()
	if err != nil || result == nil {
		return nil, err
	}
	if it.context.lineType(*result) == lineStartsub {
		eater := newEaterStartsub(result.Trimmed())
		if err := eater.analyze(it.context, it.memory); err != nil {
			return nil, err
		}
		created := newSub(eater.getSubname())
		it.subs[eater.getSubname()] = created
		if err := it.source.next(); err != nil {
			return nil, err
		}
		for {
			s, err := it.source.peek()
			if err != nil {
				return nil, err
			}
			if s == nil {
				break
			}
			if t := it.context.lineType(*s); t == lineStartsub {
				return nil, newEaterException("Cannot nest sub", *result)
			} else if t == lineEndsub {
				if err := it.source.next(); err != nil {
					return nil, err
				}
				it.readingInProgress = newCodeIteratorImpl(created.lines())
				break
			}
			created.add(*s)
			if err := it.source.next(); err != nil {
				return nil, err
			}
		}
	}
	if it.readingInProgress != nil {
		return it.readingInProgress.peek()
	}
	return result, nil
}

func (it *codeIteratorSub) next() error {
	if it.readingInProgress == nil {
		return it.source.next()
	}
	if err := it.readingInProgress.next(); err != nil {
		return err
	}
	s, err := it.readingInProgress.peek()
	if err != nil {
		return err
	}
	if s == nil {
		it.readingInProgress = nil
	}
	return nil
}
