// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/Stereotype.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/StereotypeDecoration.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// Stereotype is the "<<...>>" label of an element, with an optional spot
// character or sprite ("<<(C,#color) label>>", "<<$sprite>>"). It is a
// Display element: its String is the label, with the spot character first.
type Stereotype struct {
	radius                float64
	circledFont           *font.UFont
	automaticPackageStyle bool
	decoration            *stereotypeDecoration
}

func checkStereotypeLabel(label string) {
	if !strings.HasPrefix(label, "<<") || !strings.HasSuffix(label, ">>") {
		throw(illegalArgument, label)
	}
}

// StereotypeBuild returns the stereotype of a label such as "<<foo>>", with
// automatic package style. A label not enclosed in "<<" and ">>" panics, as
// upstream throws IllegalArgumentException.
func StereotypeBuild(label string) *Stereotype { return StereotypeBuildAuto(label, true) }

// StereotypeBuildAuto returns the stereotype of a label.
func StereotypeBuildAuto(label string, automaticPackageStyle bool) *Stereotype {
	checkStereotypeLabel(label)
	decoration := buildSimpleDecoration(label)
	return &Stereotype{automaticPackageStyle: automaticPackageStyle, decoration: decoration}
}

// StereotypeBuildSpotted returns the stereotype of a label that may define
// a spot ("<<(C,#color) label>>") or a sprite, drawn with circledFont in a
// circle of radius. An unknown color is a *color.NoSuchColorError.
func StereotypeBuildSpotted(label string, radius float64, circledFont font.UFont,
	htmlColorSet *color.HColorSet) (*Stereotype, error) {
	checkStereotypeLabel(label)
	decoration, err := buildComplexDecoration(label, htmlColorSet)
	if err != nil {
		return nil, err
	}
	return &Stereotype{automaticPackageStyle: true, decoration: decoration, radius: radius,
		circledFont: &circledFont}, nil
}

// HtmlColor returns the color of the spot, nil for none.
func (s *Stereotype) HtmlColor() color.HColor { return s.decoration.htmlColor }

// Character returns the spot character, 0 for none.
func (s *Stereotype) Character() rune { return s.decoration.character }

// Sprite returns the sprite of the stereotype as a block, nil when there is
// none.
func (s *Stereotype) Sprite(container SpriteContainer) TextBlock {
	if s.decoration.spriteName == "" || container == nil {
		return nil
	}
	tmp := container.Sprite(s.decoration.spriteName)
	if tmp == nil {
		return nil
	}
	return tmp.AsTextBlock(s.HtmlColor(), nil, s.decoration.spriteScale, nil)
}

// IsWithOOSymbol reports whether the label is "<<O-O>>".
func (s *Stereotype) IsWithOOSymbol() bool {
	return javaEqualsIgnoreCase("<<O-O>>", s.decoration.label)
}

var stereotypeMultiplePattern = regexp.MustCompile(`<<[\t\n\v\f\r ]?((?:<&\w+>|[^<>])+?)[\t\n\v\f\r ]?>>`)

// MultipleLabels returns the names of the "<<name>>" parts of the label.
func (s *Stereotype) MultipleLabels() []string {
	var result []string
	for _, m := range stereotypeMultiplePattern.FindAllStringSubmatch(s.decoration.label, -1) {
		result = append(result, m[1])
	}
	return result
}

// IsSpotted reports whether the stereotype has a spot character.
func (s *Stereotype) IsSpotted() bool { return s.decoration.character != 0 }

// String is the label, with the spot character and a space first.
func (s *Stereotype) String() string {
	if s.decoration.character == 0 {
		return s.decoration.label
	}
	return string(s.decoration.character) + " " + s.decoration.label
}

// Radius returns the radius of the spot.
func (s *Stereotype) Radius() float64 { return s.radius }

// CircledFont returns the font of the spot, nil for none.
func (s *Stereotype) CircledFont() *font.UFont { return s.circledFont }

// Label returns the label written with guillemet; ok is false for "<<O-O>>"
// (null upstream). An archimate sprite gives its name.
func (s *Stereotype) Label(guillemet *Guillemet) (label string, ok bool) {
	if s.IsWithOOSymbol() {
		return "", false
	}
	if s.decoration.spriteName != "" && strings.HasPrefix(s.decoration.spriteName, "archimate/") {
		return guillemet.ManageGuillemet("<<" + s.decoration.spriteName[len("archimate/"):] + ">>"), true
	}
	return guillemet.ManageGuillemet(s.decoration.label), true
}

// labelOrNull is Label for the callers that concatenate it: Java gives
// "null" for "<<O-O>>".
func (s *Stereotype) labelOrNull(guillemet *Guillemet) string {
	l, ok := s.Label(guillemet)
	if !ok {
		return "null"
	}
	return l
}

// Labels returns the "<<...>>" parts of the label, each written with
// guillemet.
func (s *Stereotype) Labels(guillemet *Guillemet) []string {
	labelLocal, ok := s.Label(GuillemetDoubleComparator())
	if !ok {
		return nil
	}
	return cutStereotypeLabels(labelLocal, guillemet)
}

// Styles returns the style of each style name of the stereotype.
func (s *Stereotype) Styles(builder *StyleBuilder) []*Style {
	var result []*Style
	for _, name := range s.StyleNames() {
		result = append(result, builder.CreateStyleStereotype(name))
	}
	return result
}

// StyleNames returns the names of the stereotype parts, and the last part
// of the sprite name.
func (s *Stereotype) StyleNames() []string { return s.decoration.styleNames() }

// PackageStyle returns the package style named by the label ("<<Node>>",
// ...); ok is false when there is none or when the automatic package style
// is off.
func (s *Stereotype) PackageStyle() (PackageStyle, bool) {
	if !s.automaticPackageStyle {
		return 0, false
	}
	for p := PackageStyle(0); p < packageStyleCount; p++ {
		if javaEqualsIgnoreCase("<<"+p.String()+">>", s.decoration.label) {
			return p, true
		}
	}
	return 0, false
}

func (s *Stereotype) labelIs(names ...string) bool {
	for _, n := range names {
		if javaEqualsIgnoreCase(s.decoration.label, n) {
			return true
		}
	}
	return false
}

// IsBiddableOrUncertain is a test of the requirement diagrams.
func (s *Stereotype) IsBiddableOrUncertain() bool {
	return s.labelIs("<<B>>", "<<Biddable>>", "<<Uncertain>>")
}

// IsCausal is a test of the requirement diagrams.
func (s *Stereotype) IsCausal() bool { return s.labelIs("<<C>>", "<<Causal>>") }

// IsLexicalOrGiven is a test of the requirement diagrams.
func (s *Stereotype) IsLexicalOrGiven() bool {
	return s.labelIs("<<L>>", "<<Lexical>>", "<<X>>", "<<Given>>")
}

// IsDesignedOrSolved is a test of the requirement diagrams.
func (s *Stereotype) IsDesignedOrSolved() bool {
	return s.labelIs("<<D>>", "<<Designed>>", "<<Nested>>", "<<Solved>>")
}

// IsMachineOrSpecification is a test of the requirement diagrams.
func (s *Stereotype) IsMachineOrSpecification() bool {
	return s.labelIs("M", "<<Machine>>", "<<S>>", "<<Spec>>", "<<Specification>>")
}

// IsIcon reports whether the label is "<<icon>>".
func (s *Stereotype) IsIcon() bool { return s.labelIs("<<icon>>") }

// stereotypeDecoration is upstream's StereotypeDecoration: the label
// without spot definition, the spot and the sprite.
type stereotypeDecoration struct {
	label       string
	htmlColor   color.HColor
	character   rune
	spriteName  string // "" for none (null upstream)
	spriteScale float64
}

var (
	stereoCircleChar = regex.NewConcat(
		regex.NewLeaf(`\<\<`),
		regex.SpaceZeroOrMore(),
		regex.NewLeaf(`\(`),
		regex.NewNamedLeaf(1, "CHAR", `(\S)`),
		regex.Optional(regex.NewConcat(
			regex.SpaceZeroOrMore(),
			regex.NewLeaf(","),
			regex.SpaceZeroOrMore(),
			regex.NewNamedLeaf(1, "COLOR", `(#[0-9a-fA-F]{6}|\w+)`),
			regex.SpaceZeroOrMore())),
		regex.NewLeaf(`\)`),
		regex.Optional(regex.NewNamedLeaf(1, "LABEL", javaDots(`[,]?(.*?)`))),
		regex.NewLeaf(`\>\>`))
	stereoCircleSprite = regex.NewConcat(
		regex.NewLeaf(`\<\<`),
		regex.SpaceZeroOrMore(),
		regex.NewLeaf(`\(?\$`),
		regex.NewNamedLeaf(1, "NAME", "("+SpriteUtilsSpriteName+")"),
		regex.NewNamedLeaf(2, "SCALE", `((?:\{scale=|\*)([0-9.]+)\}?)?`),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(
			regex.NewLeaf(","),
			regex.SpaceZeroOrMore(),
			regex.NewNamedLeaf(1, "COLOR", `(#[0-9a-fA-F]{6}|\w+)`))),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewNamedLeaf(1, "LABEL", javaDots(`[),](.*?)`))),
		regex.NewLeaf(`\>\>`))
)

// String is upstream's toString.
func (d *stereotypeDecoration) String() string {
	spriteName := d.spriteName
	if spriteName == "" {
		spriteName = "null"
	}
	return "label='" + d.label + "' spriteName='" + spriteName + "'"
}

func (d *stereotypeDecoration) styleNames() []string {
	result := cutStereotypeLabels(d.label, GuillemetNone())
	if d.spriteName == "" {
		return result
	}
	if idx := strings.LastIndexByte(d.spriteName, '/'); idx != -1 {
		result = append(result, d.spriteName[idx+1:])
	}
	return result
}

func buildSimpleDecoration(name string) *stereotypeDecoration {
	d := &stereotypeDecoration{label: name}
	if strings.HasPrefix(name, "<<$") && strings.HasSuffix(name, ">>") {
		mCircleSprite := stereoCircleSprite.Matcher(name)
		if mCircleSprite == nil {
			// Upstream dereferences the missing match.
			throw(nullPointer, "")
		}
		d.spriteName, _ = mCircleSprite.Get("NAME", 0)
		scale, ok := mCircleSprite.Get("SCALE", 0)
		d.spriteScale = ParserScale(scale, ok, 1)
	}
	return d
}

func buildComplexDecoration(full string, htmlColorSet *color.HColorSet) (*stereotypeDecoration, error) {
	d := &stereotypeDecoration{}
	label := ""
	for _, name := range cutStereotypeLabels(full, GuillemetDoubleComparator()) {
		mCircleChar := stereoCircleChar.Matcher(name)
		mCircleSprite := stereoCircleSprite.Matcher(name)
		if mCircleSprite != nil {
			if l, ok := mCircleSprite.Get("LABEL", 0); ok && !isEmptyJava(l) {
				name = "<<" + l + ">>"
			} else {
				name = ""
			}
			var col color.HColor
			if colName, ok := mCircleSprite.Get("COLOR", 0); ok {
				c, err := htmlColorSet.GetColor(colName)
				if err != nil {
					return nil, err
				}
				col = c
			}
			if col == nil {
				d.htmlColor = color.Black
			} else {
				d.htmlColor = col
			}
			d.character = 0
			d.spriteName, _ = mCircleSprite.Get("NAME", 0)
			scale, ok := mCircleSprite.Get("SCALE", 0)
			d.spriteScale = ParserScale(scale, ok, 1)
		} else if mCircleChar != nil {
			if l, ok := mCircleChar.Get("LABEL", 0); ok && !isEmptyJava(l) {
				name = "<<" + l + ">>"
			} else {
				name = ""
			}
			d.htmlColor = nil
			if colName, ok := mCircleChar.Get("COLOR", 0); ok {
				c, err := htmlColorSet.GetColor(colName)
				if err != nil {
					return nil, err
				}
				d.htmlColor = c
			}
			ch, _ := mCircleChar.Get("CHAR", 0)
			d.character = javaCharAt(ch, 0)
		}
		label += name
	}
	d.label = label
	return d, nil
}

var stereotypeCutPattern = regex.MustCompile(javaDots(`\<{2,3}.*?\>{2,3}`))

// cutStereotypeLabels returns the "<<...>>" parts of label, written with
// guillemet; "<<<...>>>" parts (style names) are skipped.
func cutStereotypeLabels(label string, guillemet *Guillemet) []string {
	var result []string
	for _, group := range stereotypeCutPattern.FindAllString(label, -1) {
		if !strings.HasPrefix(group, "<<<") {
			result = append(result, guillemet.ManageGuillemetStrict(group))
		}
	}
	return result
}
