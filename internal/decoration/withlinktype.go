// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/decoration/WithLinkType.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package decoration

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// WithLinkType is the state a link shares with its line options: its type,
// hidden, single and node style flags, and its colors. Links embed it and
// give the norank hook (upstream's abstract goNorank) with SetNorank.
type WithLinkType struct {
	LinkType      LinkType
	Hidden        bool
	single        bool
	useNodeStyle  bool
	colors        *color.Colors
	supplementary []*color.Colors
	norank        func()
}

// InitWithLinkType returns the state of a link of type typ; norank is
// upstream's goNorank.
func InitWithLinkType(typ LinkType, norank func()) WithLinkType {
	return WithLinkType{LinkType: typ, colors: color.EmptyColors(), norank: norank}
}

// SpecificColor returns the line color, nil for none.
func (w *WithLinkType) SpecificColor() color.HColor { return w.colors.Color(color.TypeLine) }

// SetSpecificColor sets the line color of the first line (i is 0) or adds
// the colors of a supplementary line.
func (w *WithLinkType) SetSpecificColor(specificColor color.HColor, i int) {
	if i == 0 {
		w.colors = w.colors.Add(color.TypeLine, specificColor)
	} else {
		w.supplementary = append(w.supplementary, w.colors.Add(color.TypeLine, specificColor))
	}
}

// SupplementaryColors returns the colors of the supplementary lines.
func (w *WithLinkType) SupplementaryColors() []*color.Colors { return w.supplementary }

// SetColors sets the colors.
func (w *WithLinkType) SetColors(colors *color.Colors) { w.colors = colors }

// Colors returns the colors.
func (w *WithLinkType) Colors() *color.Colors { return w.colors }

// GoDashed makes the line dashed.
func (w *WithLinkType) GoDashed() { w.LinkType = w.LinkType.GoDashed() }

// GoDotted makes the line dotted.
func (w *WithLinkType) GoDotted() { w.LinkType = w.LinkType.GoDotted() }

// GoThickness sets the line thickness.
func (w *WithLinkType) GoThickness(thickness float64) { w.LinkType = w.LinkType.GoThickness(thickness) }

// GoHidden hides the link.
func (w *WithLinkType) GoHidden() { w.Hidden = true }

// GoBold makes the line bold.
func (w *WithLinkType) GoBold() { w.LinkType = w.LinkType.GoBold() }

// GoSingle marks the link as single: a similar link is not added twice.
func (w *WithLinkType) GoSingle() { w.single = true }

// IsSingle reports whether the link is single.
func (w *WithLinkType) IsSingle() bool { return w.single }

// GoNodeStyle draws the link with the style of its nodes.
func (w *WithLinkType) GoNodeStyle() { w.useNodeStyle = true }

// UseNodeStyle reports whether the link uses the style of its nodes.
func (w *WithLinkType) UseNodeStyle() bool { return w.useNodeStyle }

// ApplyStyle applies the options of a "-[...]->" style, such as
// "#red,dashed;#blue": each ';' part is one line, each ',' item a color
// or an option. ok false (null) does nothing.
func (w *WithLinkType) ApplyStyle(arrowStyle string, ok bool) {
	if !ok {
		return
	}
	for i, s := range tokens(arrowStyle, ';') {
		w.applyOneStyle(s, i)
	}
}

func (w *WithLinkType) applyOneStyle(arrowStyle string, i int) {
	for _, s := range tokens(arrowStyle, ',') {
		switch {
		case uml.EqualsIgnoreCase(s, "dashed"):
			w.GoDashed()
		case uml.EqualsIgnoreCase(s, "bold"):
			w.GoBold()
		case uml.EqualsIgnoreCase(s, "dotted"):
			w.GoDotted()
		case uml.EqualsIgnoreCase(s, "hidden"):
			w.GoHidden()
		case uml.EqualsIgnoreCase(s, "single"):
			w.GoSingle()
		case uml.EqualsIgnoreCase(s, "plain"):
			// Do nothing
		case uml.EqualsIgnoreCase(s, "node"):
			w.GoNodeStyle()
		case uml.EqualsIgnoreCase(s, "norank"):
			w.norank()
		case strings.HasPrefix(s, "thickness="):
			w.GoThickness(uml.JavaParseDouble(s[len("thickness="):]))
		default:
			w.SetSpecificColor(color.Instance().GetColorOrWhite(s), i)
		}
	}
}

// tokens is java.util.StringTokenizer with one delimiter: the non-empty
// parts of s.
func tokens(s string, delim rune) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == delim })
}
