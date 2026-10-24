// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram/command/UBrexCommandPartition.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram/command/UBrexCommandEndPartition.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram

import (
	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// commandPartition is upstream's UBrexCommandPartition: "partition Name
// #color <<stereo>> {" opens a partition.
func commandPartition() cmd {
	partitionColor := klimt.UBrexSimpleColor(color.TypeBack)
	return uml.UBrexTokens([]string{"partition"}, klimt.UBrexConcat(
		klimt.UBrexLeaf("partition"),
		klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexLeaf("【 〃 〶$NAME=〇+「〤〃」 〃 ┇ 〶$NAME=〇+〴S   】"),
		klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexOr(
			partitionColor.Regex(),
			// The legacy colors that upstream ignores.
			klimt.UBrexOptional(klimt.UBrexLeaf("【 # 〇{6}「0〜9a〜fA〜F」┇ 〇?# 〇+〴w 】"))),
		klimt.UBrexOptional(klimt.UBrexNamed("STEREOTYPE", klimt.UBrexLeaf("<<  〄+〴. ->〘 >>〙"))),
		klimt.UBrexSpaceZeroOrMore(),
		klimt.UBrexOptional(klimt.UBrexLeaf("{")),
		klimt.UBrexEnd()),
		func(d *ActivityDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			quark := d.QuarkInContext(true, d.CleanID(arg.Str("NAME", 0)))
			d.GotoGroup(location, quark, klimt.DisplayWithNewlines(d.Pragma(), quark.Name()), abel.GroupTypePackage,
				abel.USymbolNull)
			p := d.CurrentGroup()
			colors, err := partitionColor.Color(arg, d.Skin().IHtmlColorSet())
			if err != nil {
				return abel.ColorResult(err)
			}
			if !colors.IsEmpty() {
				p.SetColors(colors)
			}
			var stereotype *klimt.Stereotype
			if s, ok := arg.Get("STEREOTYPE", 0); ok {
				stereotype = klimt.StereotypeBuild(s)
			}
			p.SetStereotype(stereotype)
			return command.Ok()
		})
}

// commandEndPartition is upstream's UBrexCommandEndPartition: "end
// partition" or "}" closes the current partition. As upstream's pattern,
// the end of the text is a third alternative.
func commandEndPartition() cmd {
	return uml.UBrexTokens([]string{"end", "endpartition", "}"}, klimt.UBrexConcat(klimt.UBrexOr(
		klimt.UBrexConcat(klimt.UBrexLeaf("end"), klimt.UBrexSpaceZeroOrMore(), klimt.UBrexLeaf("partition")),
		klimt.UBrexLeaf("}"),
		klimt.UBrexEnd())),
		func(d *ActivityDiagram, _ *text.LineLocation, _ *regex.Result, _ command.ParserPass) command.Result {
			if !d.EndGroup() {
				return command.Error("No partition defined")
			}
			return command.Ok()
		})
}
