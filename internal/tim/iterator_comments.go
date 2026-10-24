// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorLongComment.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorShortComment.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/iterator/CodeIteratorInnerComment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// codeIteratorLongComment skips /' ... '/ comments spanning several lines.
type codeIteratorLongComment struct {
	abstractCodeIterator
	context *TContext
	logs    *[]text.StringLocated
}

func (it *codeIteratorLongComment) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		if it.context.lineType(*result) != lineCommentLongStart {
			return result, nil
		}
		var s *text.StringLocated
		for {
			if s, err = it.source.peek(); err != nil {
				return nil, err
			}
			if s == nil || strings.HasSuffix(s.Trimmed().S, "'/") {
				break
			}
			*it.logs = append(*it.logs, *s)
			if err := it.source.next(); err != nil {
				return nil, err
			}
		}
		if s != nil {
			*it.logs = append(*it.logs, *s)
			if err := it.source.next(); err != nil {
				return nil, err
			}
		}
	}
}

// codeIteratorShortComment skips ' comments and one-line /' ... '/ comments.
type codeIteratorShortComment struct {
	abstractCodeIterator
	context *TContext
	logs    *[]text.StringLocated
}

func (it *codeIteratorShortComment) peek() (*text.StringLocated, error) {
	for {
		result, err := it.source.peek()
		if err != nil || result == nil {
			return nil, err
		}
		if it.context.lineType(*result) == lineCommentSimple {
			*it.logs = append(*it.logs, *result)
			if err := it.next(); err != nil {
				return nil, err
			}
			continue
		}
		return result, nil
	}
}

// codeIteratorInnerComment removes comments inside lines.
type codeIteratorInnerComment struct {
	abstractCodeIterator
}

func (it *codeIteratorInnerComment) peek() (*text.StringLocated, error) {
	result, err := it.source.peek()
	if err != nil || result == nil {
		return nil, err
	}
	r := result.RemoveInnerComment()
	return &r, nil
}
