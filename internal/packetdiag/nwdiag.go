// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/nwdiag/NwDiagram.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/nwdiag/VerticalLine.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package packetdiag

import (
	"math"
	"regexp"

	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// attributePattern is NwDiagram's pattern of a name = value pair, with
// Java's \s.
var attributePattern = func() *regexp.Regexp {
	p, err := regex.Transform(`\s*(\w+)\s*=\s*("([^"]*)"|[^\s,]+)`)
	if err != nil {
		panic(err)
	}
	return regexp.MustCompile(p)
}()

// toSet is upstream's NwDiagram.toSet: the name = value pairs of an
// attribute list, without the quotes of quoted values. A name given twice
// keeps its last value. ok is false for upstream's null definition.
func toSet(definition string, ok bool) map[string]string {
	result := map[string]string{}
	if !ok {
		return result
	}
	for _, m := range attributePattern.FindAllStringSubmatchIndex(definition, -1) {
		name := definition[m[2]:m[3]]
		value := definition[m[4]:m[5]]
		if m[6] >= 0 {
			value = definition[m[6]:m[7]]
		}
		result[name] = value
	}
	return result
}

// verticalLine is upstream's nwdiag VerticalLine: a vertical line from y1 to
// y2 at x 0, interrupted by a small arc at each of the skipped ordinates.
type verticalLine struct {
	y1, y2 float64
	skip   []float64
}

func newVerticalLine(y1, y2 float64, skip []float64) *verticalLine {
	return &verticalLine{y1: math.Min(y1, y2), y2: math.Max(y1, y2), skip: skip}
}

func (v *verticalLine) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(klimt.Bg(color.None()))
	drawn := false
	current := v.y1
	path := klimt.UPathNone()
	path.MoveTo(0, current)
	for _, step := range v.skip {
		if step < v.y1 {
			continue
		}
		drawn = true
		if step == v.y2 {
			path.LineTo(0, v.y2)
		} else {
			path.LineTo(0, math.Min(v.y2, step-3))
			if v.y2 > step {
				path.ArcTo(4, 4, 0, 0, 1, 0, step+9)
				continue
			}
		}
		ug.Draw(path)
		path = klimt.UPathNone()
		current = step + 9
		path.MoveTo(0, current)
		if current >= v.y2 {
			break
		}
	}
	if !drawn {
		path.LineTo(0, v.y2)
		ug.Draw(path)
	}
}
