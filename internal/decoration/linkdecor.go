// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/LinkDecor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/LinkMiddleDecor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package decoration

import (
	"cmp"
	"regexp"
	"slices"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// LinkDecor is the decoration at one end of a link: an arrow head, a
// diamond, a crow's foot...
type LinkDecor int

// Link decorations, in upstream's order.
const (
	LinkDecorNone LinkDecor = iota
	LinkDecorExtends
	LinkDecorComposition
	LinkDecorAggregation
	LinkDecorNotNavigable
	LinkDecorRedefines
	LinkDecorDefinedBy
	LinkDecorCrowfoot
	LinkDecorCircleCrowfoot
	LinkDecorCircleLine
	LinkDecorDoubleLine
	LinkDecorLineCrowfoot
	LinkDecorArrow
	LinkDecorArrowTriangle
	LinkDecorArrowAndCircle
	LinkDecorCircle
	LinkDecorCircleFill
	LinkDecorCircleConnect
	LinkDecorParenthesis
	LinkDecorSquare
	LinkDecorCircleCross
	LinkDecorPlus
	LinkDecorHalfArrowUp
	LinkDecorHalfArrowDown
	LinkDecorSquareToBeRemoved
)

type linkDecorData struct {
	name      string
	decors1   []string
	decors2   []string
	margin    int
	fill      bool
	arrowSize float64
}

// linkDecors holds upstream's constructor arguments. PARENTHESIS has the
// arrow size 1.0 as GlobalConfig.USE_INTERFACE_EYE2 is false.
var linkDecors = [...]linkDecorData{
	{"NONE", nil, nil, 2, false, 0},
	{"EXTENDS", []string{"<|", "^"}, []string{"|>", "^"}, 30, false, 2},
	{"COMPOSITION", []string{"*"}, []string{"*"}, 15, true, 1.3},
	{"AGGREGATION", []string{"o"}, []string{"o"}, 15, false, 1.3},
	{"NOT_NAVIGABLE", []string{"x"}, []string{"x"}, 1, false, 0.5},
	{"REDEFINES", []string{"<||"}, []string{"||>"}, 30, false, 2},
	{"DEFINEDBY", []string{"<|:"}, []string{":|>"}, 30, false, 2},
	{"CROWFOOT", []string{"}"}, []string{"{"}, 10, true, 0.8},
	{"CIRCLE_CROWFOOT", []string{"}o"}, []string{"o{"}, 14, false, 0.8},
	{"CIRCLE_LINE", []string{"|o"}, []string{"o|"}, 10, false, 0.8},
	{"DOUBLE_LINE", []string{"||"}, []string{"||"}, 7, false, 0.7},
	{"LINE_CROWFOOT", []string{"}|"}, []string{"|{"}, 10, false, 0.8},
	{"ARROW", []string{"<", "<_"}, []string{">", "_>"}, 10, true, 0.5},
	{"ARROW_TRIANGLE", []string{"<<"}, []string{">>"}, 10, true, 0.8},
	{"ARROW_AND_CIRCLE", nil, nil, 10, false, 0.5},
	{"CIRCLE", []string{"0"}, []string{"0"}, 0, false, 0.5},
	{"CIRCLE_FILL", []string{"@"}, []string{"@"}, 0, false, 0.5},
	{"CIRCLE_CONNECT", []string{"0)"}, []string{"(0"}, 0, false, 0.5},
	{"PARENTHESIS", []string{")"}, []string{"("}, 0, false, 1.0},
	{"SQUARE", []string{"#"}, []string{"#"}, 0, false, 0.5},
	{"CIRCLE_CROSS", nil, nil, 0, false, 0.5},
	{"PLUS", []string{"+"}, []string{"+"}, 0, false, 1.5},
	{"HALF_ARROW_UP", nil, []string{`\\`}, 0, false, 1.5},
	{"HALF_ARROW_DOWN", nil, []string{"//"}, 0, false, 1.5},
	{"SQUARE_toberemoved", nil, nil, 30, false, 0},
}

// decors1 and decors2 are upstream's DECORS1 and DECORS2: the decoration of
// each head text, on the first and on the second end. They are read-only
// after initialization.
var decors1, decors2 = func() (map[string]LinkDecor, map[string]LinkDecor) {
	d1, d2 := map[string]LinkDecor{}, map[string]LinkDecor{}
	for i, d := range linkDecors {
		for _, s := range d.decors1 {
			d1[s] = LinkDecor(i)
		}
		for _, s := range d.decors2 {
			d2[s] = LinkDecor(i)
		}
	}
	return d1, d2
}()

// String returns upstream's constant name, for example "EXTENDS".
func (d LinkDecor) String() string { return linkDecors[d].name }

// Margin returns the space the decoration needs.
func (d LinkDecor) Margin() int { return linkDecors[d].margin }

// IsFill reports whether the decoration is filled.
func (d LinkDecor) IsFill() bool { return linkDecors[d].fill }

// ArrowSize returns the size of the decoration for Graphviz.
func (d LinkDecor) ArrowSize() float64 { return linkDecors[d].arrowSize }

// IsExtendsLike reports whether the decoration is a triangle: EXTENDS,
// REDEFINES or DEFINEDBY.
func (d LinkDecor) IsExtendsLike() bool {
	return d == LinkDecorExtends || d == LinkDecorRedefines || d == LinkDecorDefinedBy
}

// LookupDecors1 returns the decoration of a head text on the first end of a
// link, NONE for an unknown text; ok false (null) gives NONE.
func LookupDecors1(s string, ok bool) LinkDecor {
	if !ok {
		return LinkDecorNone
	}
	if d, found := decors1[text.Trin(s)]; found {
		return d
	}
	return LinkDecorNone
}

// LookupDecors2 returns the decoration of a head text on the second end of
// a link, NONE for an unknown text; ok false (null) gives NONE.
func LookupDecors2(s string, ok bool) LinkDecor {
	if !ok {
		return LinkDecorNone
	}
	if d, found := decors2[text.Trin(s)]; found {
		return d
	}
	return LinkDecorNone
}

// RegexDecors1 is upstream's getRegexDecors1: an optional group matching
// the head texts of the first end.
func RegexDecors1() string { return buildRegexFromDecorKeys(decors1) }

// RegexDecors2 is upstream's getRegexDecors2: an optional group matching
// the head texts of the second end.
func RegexDecors2() string { return buildRegexFromDecorKeys(decors2) }

// buildRegexFromDecorKeys is upstream's buildRegexFromDecorKeys: the keys
// by decreasing length (upstream sorts the keys of a HashMap; keys of the
// same length cannot match at the same place, so their order does not
// matter), quoted, with a word boundary next to a leading or trailing "o".
// Java 19 and later use ASCII word characters for \b, as RE2.
func buildRegexFromDecorKeys(keys map[string]LinkDecor) string {
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	slices.SortFunc(sorted, func(a, b string) int {
		if c := cmp.Compare(len(b), len(a)); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	})
	parts := make([]string, len(sorted))
	for i, key := range sorted {
		quoted := regexp.QuoteMeta(key)
		startsWithO := strings.HasPrefix(key, "o")
		endsWithO := strings.HasSuffix(key, "o")
		switch {
		case startsWithO && endsWithO:
			quoted = `\b` + quoted + `\b`
		case startsWithO:
			quoted = `\b` + quoted
		case endsWithO:
			quoted = quoted + `\b`
		}
		parts[i] = quoted
	}
	return "(" + strings.Join(parts, "|") + ")?"
}

// LinkMiddleDecor is the decoration in the middle of a link.
type LinkMiddleDecor int

// Middle decorations, in upstream's order.
const (
	LinkMiddleDecorNone LinkMiddleDecor = iota
	LinkMiddleDecorCircle
	LinkMiddleDecorCircleCircled
	LinkMiddleDecorCircleCircled1
	LinkMiddleDecorCircleCircled2
	LinkMiddleDecorSubset
	LinkMiddleDecorSuperset
)

var linkMiddleDecorNames = [...]string{"NONE", "CIRCLE", "CIRCLE_CIRCLED", "CIRCLE_CIRCLED1", "CIRCLE_CIRCLED2",
	"SUBSET", "SUPERSET"}

// String returns upstream's constant name.
func (d LinkMiddleDecor) String() string { return linkMiddleDecorNames[d] }

// Inversed returns the decoration seen from the other end.
func (d LinkMiddleDecor) Inversed() LinkMiddleDecor {
	switch d {
	case LinkMiddleDecorCircleCircled1:
		return LinkMiddleDecorCircleCircled2
	case LinkMiddleDecorCircleCircled2:
		return LinkMiddleDecorCircleCircled1
	}
	return d
}
