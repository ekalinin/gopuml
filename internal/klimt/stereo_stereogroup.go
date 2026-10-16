// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/Stereogroup.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/Stereostyles.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/Stereotag.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/StereotypePattern.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
)

const stereogroupKey = "STEREOGROUP"

// Stereogroup is the raw "<<a>> <<b>>" text of a command, from which
// stereotypes, colors and styles are taken.
//
// Upstream's getBoxStyle (activitydiagram3.ftile.BoxStyle) and getLeafType
// (abel.LeafType) return diagram-level types; those packages compute them
// from Labels.
type Stereogroup struct {
	definition string
	set        bool // false for upstream's NONE (null definition)
}

// StereogroupNone returns the empty group.
func StereogroupNone() *Stereogroup { return &Stereogroup{} }

var stereogroupPattern = regexp.MustCompile(`<<([^<>]+)>>`)

// StereogroupOptional returns the optional leaf capturing a group as
// "STEREOGROUP".
func StereogroupOptional() regex.Part {
	return regex.Optional(regex.NewNamedLeaf(1, stereogroupKey, `(<<[^<>]+>>(?:[%s]*<<[^<>]+>>)*)`))
}

// StereogroupBuildFrom returns the group captured by StereogroupOptional.
func StereogroupBuildFrom(arg *regex.Result) *Stereogroup {
	full, ok := arg.Get(stereogroupKey, 0)
	return &Stereogroup{definition: full, set: ok}
}

// StereogroupBuild returns the group of a text.
func StereogroupBuild(full string) *Stereogroup { return &Stereogroup{definition: full, set: true} }

// IsEmpty reports whether there is no definition.
func (g *Stereogroup) IsEmpty() bool { return !g.set }

// BuildStereotype returns the stereotype of the definition, nil when there
// is none.
func (g *Stereogroup) BuildStereotype() *Stereotype {
	if !g.set {
		return nil
	}
	return StereotypeBuild(g.definition)
}

// Labels returns the trimmed names of the "<<name>>" parts.
func (g *Stereogroup) Labels() []string {
	if !g.set {
		return nil
	}
	var result []string
	for _, m := range stereogroupPattern.FindAllStringSubmatch(g.definition, -1) {
		result = append(result, strings.TrimFunc(m[1], func(r rune) bool { return r <= ' ' }))
	}
	return result
}

// Mute returns style with the inner colors of the group.
func (g *Stereogroup) Mute(style *Style, colorSet *color.HColorSet) (*Style, error) {
	colors, err := g.InnerColors(colorSet)
	if err != nil {
		return nil, err
	}
	return style.EventuallyOverrideColors(colors), nil
}

// HColor returns the color of pname for the group: from its inner colors,
// else from the style of the signature with the group's stereotype.
func (g *Stereogroup) HColor(styleSignature StyleSignature, pname PName, styleBuilder *StyleBuilder,
	colorSet *color.HColorSet) (color.HColor, error) {
	style := StyleSignatureWithStereogroup(styleSignature, g).MergedStyle(styleBuilder)
	colors, err := g.InnerColors(colorSet)
	if err != nil {
		return nil, err
	}
	return ColorsColorOf(colors, style, pname, colorSet), nil
}

// InnerColors returns the colors given by the labels: "###color" (text),
// "##[style]color" (line), "#..." (colors) and "prop:value;..." (a style).
func (g *Stereogroup) InnerColors(colorSet *color.HColorSet) (*color.Colors, error) {
	colors := color.EmptyColors()
	for _, label := range g.Labels() {
		switch {
		case strings.HasPrefix(label, "###"):
			label = label[3:]
			if textColor := colorSet.GetColorOrNull(label); textColor != nil {
				colors = colors.Add(color.TypeText, textColor)
			}
		case strings.HasPrefix(label, "##"):
			label = label[2:]
			if label == "" {
				throw("java.lang.StringIndexOutOfBoundsException", "Index 0 out of bounds for length 0")
			}
			if label[0] == '[' {
				if x := strings.IndexByte(label, ']'); x != -1 {
					colors = colors.AddLegacyStroke(label[1:x])
					label = label[x+1:]
				} else {
					label = ""
				}
			}
			if label != "" {
				if lineColor := colorSet.GetColorOrNull(label); lineColor != nil {
					colors = colors.Add(color.TypeLine, lineColor)
				}
			}
		case strings.HasPrefix(label, "#"):
			c, err := color.NewColors(label, colorSet, color.TypeBack)
			if err != nil {
				return nil, err
			}
			colors = colors.MergeWith(c)
		case strings.Contains(label, ":") && strings.Contains(label, ";"):
			style, err := NewStyleParser().ParseSingleLine(label)
			if err == nil {
				colors = ColorsApplyStyle(colors, style, colorSet)
			}
		}
	}
	return colors, nil
}

// Stereostyles are the "<<<name>>>" style names of a label.
type Stereostyles struct {
	names []string // insertion order, without duplicates
}

// StereostylesNone returns no style names.
func StereostylesNone() *Stereostyles { return &Stereostyles{} }

// IsEmpty reports whether there is no style name.
func (s *Stereostyles) IsEmpty() bool { return len(s.names) == 0 }

var stereostylesPattern = regex.MustCompile(javaDots(`\<{3}(.*?)\>{3}`))

// StereostylesBuild returns the "<<<name>>>" names of label.
func StereostylesBuild(label string) *Stereostyles {
	result := &Stereostyles{}
	for _, m := range stereostylesPattern.FindAllStringSubmatch(label, -1) {
		name := m[1]
		dup := false
		for _, n := range result.names {
			if n == name {
				dup = true
				break
			}
		}
		if !dup {
			result.names = append(result.names, name)
		}
	}
	return result
}

// StyleNames returns the names in order.
func (s *Stereostyles) StyleNames() []string { return s.names }

const stereotagSingle = `(\$[^%s{}%g<>$]+)`

// StereotagPattern returns the pattern of "$tag $tag2" tags.
func StereotagPattern() string {
	return "(" + stereotagSingle + "([%s]+" + stereotagSingle + ")*)"
}

// Stereotag is a "$name" tag of an element; it compares by name.
type Stereotag struct {
	name string
}

// NewStereotag returns the tag of a name without "$". A name starting with
// "$" panics, as upstream throws IllegalArgumentException.
func NewStereotag(name string) Stereotag {
	if strings.HasPrefix(name, "$") {
		throw(illegalArgument, name)
	}
	return Stereotag{name: name}
}

// Name returns the name.
func (t Stereotag) Name() string { return t.name }

// String returns "$" and the name.
func (t Stereotag) String() string { return "$" + t.name }

// StereotypePatternOptional returns the optional "<<...>>" leaf captured as
// param, with the surrounding spaces. Upstream deprecates it for
// StereogroupOptional.
func StereotypePatternOptional(param string) regex.Part {
	return regex.NewConcat(regex.SpaceZeroOrMore(), regex.Optional(StereotypePatternMandatory(param)),
		regex.SpaceZeroOrMore())
}

// StereotypePatternOptionalIGNORED is StereotypePatternOptional("IGNORED").
func StereotypePatternOptionalIGNORED() regex.Part { return StereotypePatternOptional("IGNORED") }

// StereotypePatternMandatory returns the "<<...>>" leaf captured as param.
func StereotypePatternMandatory(param string) *regex.Leaf {
	return regex.NewNamedLeaf(1, param, javaDots(`(\<\<.+?\>\>)`))
}

// StereotypePatternOptionalArchimate returns the optional "<<name>>" leaf of
// archimate elements, captured as param.
func StereotypePatternOptionalArchimate(param string) regex.Part {
	return regex.NewConcat(regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewNamedLeaf(1, param, `(\<\<[-\w]+?\>\>)`)), regex.SpaceZeroOrMore())
}

// StereotypePatternRemoveChevronBrackets removes the "<<" and ">>" around a
// stereotype.
func StereotypePatternRemoveChevronBrackets(stereo string) string {
	if strings.HasPrefix(stereo, "<<") && strings.HasSuffix(stereo, ">>") {
		return stereo[2 : len(stereo)-2]
	}
	return stereo
}
