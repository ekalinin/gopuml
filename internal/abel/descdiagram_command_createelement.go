// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandCreateElementMultilines.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/descdiagram/command/CommandCreateElementFull.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// The element patterns of upstream's CommandCreateElementFull.
const (
	// ElementAllTypes are the keywords of the description elements.
	ElementAllTypes = "person|artifact|actor/|actor|folder|card|file|package|rectangle|hexagon|label|node|frame|" +
		"cloud|action|process|database|queue|stack|storage|agent|usecase/|usecase|component|boundary|control|" +
		"entity|interface|circle|collections|port|portin|portout"
	elementCodeCore = `[%pLN_.]+|\(\)[%s]*[%pLN_.]+|\(\)[%s]*[%g][^%g]+[%g]|:[^:]+:/?|\([^()]+\)/?|\[[^\[\]]+\]`
	// ElementCode is an element code: a name, "() name", ":actor:",
	// "(use case)" or "[component]".
	ElementCode = "(" + elementCodeCore + ")"
	// ElementCodeWithQuote is ElementCode or a quoted name.
	ElementCodeWithQuote = "(" + elementCodeCore + "|[%g].+?[%g])"
	elementDisplayCore   = `[%g].+?[%g]|:[^:]+:/?|\([^()]+\)/?|\[[^\[\]]+\]`
	// ElementDisplay is an element display: quoted, ":actor:", "(use
	// case)" or "[component]".
	ElementDisplay = "(" + elementDisplayCore + ")"
	// ElementDisplayWithoutQuote is ElementDisplay or a name.
	ElementDisplayWithoutQuote = "(" + elementDisplayCore + "|[%pLN_.]+)"
)

// ExistsWithBadType3 is upstream's CommandCreateElementFull.existsWithBadType3:
// the quark has an entity of another type, or of another symbol.
func ExistsWithBadType3(code *Quark, typ LeafType, usymbol USymbol) bool {
	if code.Data() == nil {
		return false
	}
	other := code.Data()
	if other.LeafType() != typ {
		return true
	}
	return usymbol != USymbolNull && other.USymbol() != usymbol
}

// CommandCreateElementMultilines is upstream's CommandCreateElementMultilines:
// `node code as "first line` ... `last line"` (typ 0) or `node code [`
// ... `]` (typ 1).
func CommandCreateElementMultilines[D Diagram](typ int) command.Command[D] {
	parts := []regex.Part{regex.Start(),
		uml.Leaf(1, "TYPE", "("+ElementAllTypes+")[%s]+"),
		uml.Leaf(1, "CODE", "([%pLN_.]+)"),
		klimt.StereotypePatternOptional("STEREO"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		color.Exp1(),
		regex.SpaceZeroOrMore(),
	}
	var end = uml.Compile(`^(.*)[%g]$`)
	if typ == 0 {
		parts = append(parts, uml.Lit("as"), regex.SpaceZeroOrMore(), uml.Lit("[%g]"),
			uml.Leaf(1, "DESC", "([^%g]*)"), regex.End())
	} else {
		parts = append(parts, uml.Lit(`\[`), uml.Leaf(1, "DESC", "(.*)"), regex.End())
		end = uml.Compile(`^([^\[\]]*)\]$`)
	}
	return Multilines(&uml.Multilines2[D]{
		Start:    regex.NewConcat(parts...),
		Strategy: uml.RemoveStartingQuote,
		TrimEnd:  uml.TrimBoth,
		End:      end,
	}, func(cmd *uml.Multilines2[D], d D, lines command.BlocLines) command.Result {
		c := d.Cuca()
		location := uml.Location(lines)
		lines = uml.TrimSmart(lines, 1)
		line0 := cmd.StartArg(lines)
		symbol := uml.UpperCase(line0.Str("TYPE", 0))
		var leafType LeafType
		var usymbol USymbol
		switch {
		case uml.EqualsIgnoreCase(symbol, "usecase"):
			leafType, usymbol = LeafTypeUsecase, USymbolUsecase
		case uml.EqualsIgnoreCase(symbol, "usecase/"):
			leafType, usymbol = LeafTypeUsecaseBusiness, USymbolUsecaseBusiness
		default:
			usymbol = skinUSymbol(c, symbol, true)
			if usymbol == USymbolNull {
				uml.Throw(uml.IllegalStateException, "")
			}
			leafType = LeafTypeDescription
		}
		idShort := line0.Str("CODE", 0)
		lineLast := end.FindStringSubmatch(lines.Last().Trimmed().S)
		lines = uml.SubExtract(lines, 1, 1)
		display := klimt.DisplayCreateFoo(lines.Lines())
		if descStart, ok := line0.Get("DESC", 0); !uml.IsEmpty(descStart, ok) {
			display = display.AddFirst(klimt.Text(descStart))
		}
		if lineLast == nil {
			uml.ThrowNPE(`Cannot invoke "java.util.List.get(int)" because "lineLast" is null`)
		}
		if !uml.IsEmpty(lineLast[1], true) {
			display = display.Add(klimt.Text(lineLast[1]))
		}
		quark := c.QuarkInContext(true, c.CleanID(idShort))
		result := quark.Data()
		if result == nil {
			result = c.ReallyCreateLeaf(location, quark, display, leafType, usymbol)
		}
		if ExistsWithBadType3(quark, leafType, usymbol) {
			return command.Error("This element (" + quark.Name() + ") is already defined")
		}
		if stereotype, ok := line0.Get("STEREO", 0); ok {
			st, err := SpottedStereotype(c, stereotype)
			if err != nil {
				return ColorResult(err)
			}
			result.SetStereotype(st)
		}
		if urlString, ok := line0.Get(klimt.URLBuilderKey, 0); ok {
			result.AddURL(URLBuilderOf(c).URL(urlString))
		}
		colors, err := color.SimpleColor(color.TypeBack).Color(line0, c.Skin().IHtmlColorSet())
		if err != nil {
			return ColorResult(err)
		}
		result.SetColors(colors)
		return command.Ok()
	})
}
