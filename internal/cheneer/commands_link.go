// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/command/CommandAssociate.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/command/CommandSimpleSubclass.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cheneer/command/CommandMultiSubclass.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cheneer

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// noDecor is upstream's new LinkType(LinkDecor.NONE, LinkDecor.NONE).
func noDecor() decoration.LinkType {
	return decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone)
}

// entityNamed returns the entity of a name, nil when there is none.
func entityNamed(d *ChenEerDiagram, name string) *abel.Entity {
	return d.QuarkInContext(true, name).Data()
}

// commandAssociate is upstream's CommandAssociate: a link between two
// entities or relationships with an optional cardinality, `A -N- B`, bold
// when the first character is "=".
func commandAssociate() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "NAME1", "([%pLN_.-]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "PARTICIPATION", "([-=])"),
		regex.Optional(uml.Leaf(1, "CARDINALITY", `([%pLN]+|\([%pLN]+,[%s]*[%pLN]+\))`)),
		uml.Leaf(1, "PARTICIPATION2", "([-=])"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NAME2", "([%pLN_.-]+)"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeLine).Regex(),
		regex.End()),
		func(d *ChenEerDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			name1 := d.CleanID(arg.Str("NAME1", 0))
			name2 := d.CleanID(arg.Str("NAME2", 0))
			isDouble := arg.Str("PARTICIPATION", 0) == "="
			cardinality, hasCardinality := arg.Get("CARDINALITY", 0)
			entity1 := entityNamed(d, name1)
			if entity1 == nil {
				return command.Error("No such entity: " + name1)
			}
			entity2 := entityNamed(d, name2)
			if entity2 == nil {
				return command.Error("No such entity: " + name2)
			}
			linkType := noDecor()
			if isDouble {
				linkType = linkType.GoBold()
			}
			link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), entity1, entity2,
				linkType, abel.LinkArgBuild(klimt.DisplayWithNewlinesString(d.Pragma(), cardinality, hasCardinality), 3))
			link.SetPortMembers(d.PortID(entity1.Name()), d.PortID(entity2.Name()))
			colors, err := color.SimpleColor(color.TypeLine).Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			link.SetColors(colors)
			d.AddLink(link)
			return command.Ok()
		})
}

// commandSimpleSubclass is upstream's CommandSimpleSubclass: a subclass
// link between two entities, `A ->- B` with the superset sign in its
// middle or `A -<- B` with the subset sign, bold when the first character
// is "=".
func commandSimpleSubclass() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "NAME1", "([%pLN_.-]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "PARTICIPATION", "([-=])"),
		uml.Leaf(1, "DIRECTION", "([<>])"),
		uml.Leaf(1, "PARTICIPATION2", "([-=])"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "NAME2", "([%pLN_.-]+)"),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeLine).Regex(),
		regex.End()),
		func(d *ChenEerDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			name1 := d.CleanID(arg.Str("NAME1", 0))
			name2 := d.CleanID(arg.Str("NAME2", 0))
			isDouble := arg.Str("PARTICIPATION", 0) == "="
			isSuperset := arg.Str("DIRECTION", 0) == ">"
			entity1 := entityNamed(d, name1)
			if entity1 == nil {
				return command.Error("No such entity: " + name1)
			}
			entity2 := entityNamed(d, name2)
			if entity2 == nil {
				return command.Error("No such entity: " + name2)
			}
			linkType := noDecor()
			if isDouble {
				linkType = linkType.GoBold()
			}
			if isSuperset {
				linkType = linkType.WithMiddleSuperset()
			} else {
				linkType = linkType.WithMiddleSubset()
			}
			link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), entity1, entity2,
				linkType, abel.LinkArgBuild(klimt.DisplayNULL(), 3))
			link.SetPortMembers(d.PortID(entity1.Name()), d.PortID(entity2.Name()))
			colors, err := color.SimpleColor(color.TypeLine).Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			link.SetColors(colors)
			d.AddLink(link)
			return command.Ok()
		})
}

// commandMultiSubclass is upstream's CommandMultiSubclass: a group of
// subclasses linked to their superclass through a circle with a symbol,
// `A ->- d { B, C }`. With "d" and "o" the links to the subclasses have
// the superset sign, with "U" the link to the superclass has it; only the
// link to the superclass is bold when the first character is "=".
func commandMultiSubclass() cmd {
	return uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "SUPERCLASS", "([%pLN_.-]+)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "PARTICIPATION", "([-=])"),
		uml.Leaf(1, "", "(>)"),
		uml.Leaf(1, "PARTICIPATION2", "([-=])"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "SYMBOL", "([doU])"),
		regex.SpaceZeroOrMore(),
		uml.Lit(`\{`),
		uml.Leaf(1, "SUBCLASSES", "(.+)"),
		uml.Lit(`\}`),
		regex.SpaceZeroOrMore(),
		color.SimpleColor(color.TypeBack).Regex(),
		regex.End()),
		func(d *ChenEerDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			superclass := d.CleanID(arg.Str("SUPERCLASS", 0))
			subclasses := arg.Str("SUBCLASSES", 0)
			var subclassIDs []string
			for _, subclass := range klimt.JavaSplit(subclasses, ",") {
				subclassIDs = append(subclassIDs, d.CleanID(text.JavaTrim(subclass)))
			}
			isDouble := arg.Str("PARTICIPATION", 0) == "="
			symbol := arg.Str("SYMBOL", 0)
			colors, err := color.SimpleColor(color.TypeBack).Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			// The circle is created before the superclass is looked up, as
			// upstream does: the uids of the entities and links follow.
			centerQuark := d.QuarkInContext(false, superclass+"/"+symbol+subclasses+"/center")
			centerEntity := d.ReallyCreateLeaf(location, centerQuark, klimt.DisplayCreate(symbol),
				abel.LeafTypeChenCircle, abel.USymbolNull)
			centerEntity.SetColors(colors)
			superclassEntity := entityNamed(d, superclass)
			if superclassEntity == nil {
				return command.Error("No such entity: " + superclass)
			}
			linkType := noDecor()
			if isDouble {
				linkType = linkType.GoBold()
			}
			if symbol == "U" {
				linkType = linkType.WithMiddleSuperset()
			}
			link := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(), superclassEntity,
				centerEntity, linkType, abel.LinkArgBuild(klimt.DisplayNULL(), 2))
			link.SetPortMembers(d.PortID(superclassEntity.Name()), d.PortID(centerEntity.Name()))
			link.SetColors(colors)
			d.AddLink(link)
			for _, subclass := range subclassIDs {
				subclassEntity := entityNamed(d, subclass)
				if subclassEntity == nil {
					return command.Error("No such entity: " + subclass)
				}
				subclassLinkType := noDecor()
				if symbol != "U" {
					subclassLinkType = subclassLinkType.WithMiddleSuperset()
				}
				subclassLink := abel.NewLink(location, &d.CucaDiagram, d.SkinParam().CurrentStyleBuilder(),
					centerEntity, subclassEntity, subclassLinkType, abel.LinkArgBuild(klimt.DisplayNULL(), 3))
				subclassLink.SetColors(colors)
				d.AddLink(subclassLink)
			}
			return command.Ok()
		})
}
