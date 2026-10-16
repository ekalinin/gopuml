// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/command/CommandCreateClassMultilines.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/NameAndCodeParser.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/command/SingleLineCommand2.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"errors"
	"regexp"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// This file holds what the entity commands share: the commands of
// upstream's command, command.note and descdiagram.command packages that
// work on any CucaDiagram live in this package (they import the entity
// model, which the command framework cannot), generic over Diagram.

// ColorResult turns a color error into upstream's result: an unknown color
// is CommandExecutionResult.badColor(); other errors panic.
func ColorResult(err error) command.Result {
	var noSuch *color.NoSuchColorError
	if errors.As(err, &noSuch) {
		return command.BadColor()
	}
	panic(err)
}

// GetColor is HColorSet.getColor; an unknown color gives an error.
func GetColor(s string) (color.HColor, error) { return color.Instance().GetColor(s) }

// URLBuilderOf is new UrlBuilder(skinParam.getValue("topurl"),
// UrlMode.STRICT).
func URLBuilderOf(d *CucaDiagram) *klimt.URLBuilder {
	topurl, ok := d.Skin().Value("topurl")
	return klimt.NewURLBuilder(topurl, ok, klimt.URLModeStrict)
}

// SpottedStereotype is upstream's Stereotype.build(label, radius, font,
// colorSet): a stereotype that may carry a spot ("<< (S,#FF7700) >>").
func SpottedStereotype(d *CucaDiagram, label string) (*klimt.Stereotype, error) {
	skin := d.Skin()
	font := skin.Font(nil, false, klimt.FontParamCircledCharacter)
	return klimt.StereotypeBuildSpotted(label, float64(skin.CircledCharacterRadius()), font, skin.IHtmlColorSet())
}

var tagsSplit = regexp.MustCompile(`[ ]+`)

// AddTags is upstream's CommandCreateClassMultilines.addTags: the "$tag"
// names of tags, separated by spaces, become tags of the entity. Null does
// nothing.
func AddTags(entity *Entity, tags OptString) {
	if !tags.OK {
		return
	}
	for _, tag := range javaSplit(tagsSplit, tags.S) {
		if tag == "" {
			uml.Throw(uml.StringIndexOutOfBoundsException, "begin 1, end 0, length 0")
		}
		entity.AddStereotag(klimt.NewStereotag(tag[1:]))
	}
}

// javaSplit is Java's String.split(regex): the pieces between the matches,
// without the trailing empty pieces, and without a leading empty piece for
// a zero-width match at the start.
func javaSplit(re *regexp.Regexp, s string) []string {
	if s == "" {
		return []string{""}
	}
	var result []string
	last := 0
	for _, m := range re.FindAllStringIndex(s, -1) {
		if m[1] == 0 {
			continue
		}
		result = append(result, s[last:m[0]])
		last = m[1]
	}
	result = append(result, s[last:])
	for len(result) > 1 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	if len(result) == 1 && result[0] == "" && s != "" {
		return []string{}
	}
	return result
}

// QuarkInContextOpt is quarkInContext with a name that may be null: with
// a namespace separator, null panics as upstream's NullPointerException.
func QuarkInContextOpt(d *CucaDiagram, reuseExistingChild bool, full string, ok bool) *Quark {
	if _, hasSep := d.NamespaceSeparator(); !ok && hasSep {
		uml.ThrowNPE(`Cannot invoke "String.endsWith(String)" because "full" is null`)
	}
	return d.QuarkInContext(reuseExistingChild, full)
}

// TagsLeaf is new RegexLeaf(4, name, Stereotag.pattern() + "?").
func TagsLeaf(name string) *regex.Leaf {
	return regex.NewNamedLeaf(4, name, klimt.StereotagPattern()+"?")
}

// NameAndCode regexes of upstream's NameAndCodeParser.
const (
	nameCodeDisplay      = `[%g]([^%g]*)[%g]`
	nameCodeCode         = `[^%s{}%g<>]+`
	nameCodeCodeNoDotDot = `[^%s{}%g<>:]+`
)

// NameAndCode is NameAndCodeParser.nameAndCode: `"display" as code`,
// `code as "display"`, `code` or `"code"`, captured as DISPLAY1/CODE1,
// CODE2/DISPLAY2, CODE3 and CODE4.
func NameAndCode() regex.Part {
	return regex.Or("",
		regex.NewConcat(
			uml.Leaf(1, "DISPLAY1", nameCodeDisplay),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "CODE1", "("+nameCodeCode+")")),
		regex.NewConcat(
			uml.Leaf(1, "CODE2", "("+nameCodeCode+")"),
			regex.SpaceOneOrMore(),
			uml.Lit("as"),
			regex.SpaceOneOrMore(),
			uml.Leaf(1, "DISPLAY2", nameCodeDisplay)),
		uml.Leaf(1, "CODE3", "("+nameCodeCode+")"),
		uml.Leaf(1, "CODE4", "[%g]([^%g]+)[%g]"))
}

// CodeForClass is NameAndCodeParser.codeForClass: CODE, a name or a quoted
// name.
func CodeForClass() *regex.Leaf {
	return uml.Leaf(1, "CODE", "("+nameCodeCode+"|[%g][^%g]+[%g])")
}

// CodeWithMemberForClass is NameAndCodeParser.codeWithMemberForClass: CODE
// as "class::member", the class and the member in two groups.
func CodeWithMemberForClass() *regex.Leaf {
	return uml.Leaf(2, "CODE", "("+nameCodeCodeNoDotDot+"|[%g][^%g]+[%g])::([%g][^%g]+[%g]|[^%s]+)")
}

// RemoveQuotesOpt is eventuallyRemoveStartingAndEndingDoubleQuote on a
// string that may be null.
func RemoveQuotesOpt(s string, ok bool) (string, bool) {
	if !ok {
		return "", false
	}
	return uml.RemoveQuotes(s), true
}
