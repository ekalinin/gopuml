// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/Sub.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// Sub is a !startsub ... !endsub block.
type Sub struct {
	name string
	body []text.StringLocated
}

func newSub(name string) *Sub { return &Sub{name: name} }

func (s *Sub) add(line text.StringLocated) { s.body = append(s.body, line) }

func (s *Sub) lines() []text.StringLocated { return s.body }

// subFromFile is Sub.fromFile: the lines of every block named blocname in
// the reader, or nil.
func subFromFile(reader core.ReadLine, blocname string, context *TContext, memory TMemory) (*Sub, error) {
	var result *Sub
	skip := false
	for {
		s, ok := reader.ReadLine()
		if !ok {
			return result, nil
		}
		typ := context.lineType(s.Trimmed())
		if typ == lineStartsub {
			eater := newEaterStartsub(s.Trimmed())
			if err := eater.analyze(context, memory); err != nil {
				return nil, err
			}
			if eater.getSubname() == blocname {
				skip = false
				if result == nil {
					result = newSub(blocname)
				}
			}
			continue
		}
		if typ == lineEndsub && result != nil {
			skip = true
		}
		if result != nil && !skip {
			result.add(s)
		}
	}
}
