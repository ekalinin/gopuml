// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/ClassDiagramFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// Factory is the parser of class diagrams (and of object diagrams),
// upstream's ClassDiagramFactory. It is safe for concurrent use.
var Factory core.Factory = uml.NewFactory(core.Class, commands, NewClassDiagram)

// newEmptyPage is the empty diagram of a "newpage", upstream's
// createEmptyDiagram with the source and images of the diagram.
func newEmptyPage(d *ClassDiagram) *ClassDiagram {
	return NewClassDiagram(d.RenderContext(), d.Source(), d.MD5Map())
}

// commands returns the commands of class diagrams in upstream's order.
func commands() []cmd {
	cmds := []cmd{
		abel.CommandFootboxIgnored[*ClassDiagram](),
		abel.CommandRankDir[*ClassDiagram](),
		abel.CommandNewpage(newEmptyPage),
	}
	cmds = uml.AddCommonScaleCommands(cmds)
	cmds = append(cmds, commandAddMethod())
	cmds = uml.AddCommonHides(cmds)
	cmds = append(cmds,
		abel.CommandHideShow2[*ClassDiagram](),
		abel.CommandRemoveRestore[*ClassDiagram](),
		commandCreateClassMultilines(),
		commandCreateEntityObjectMultilines(),
		commandCreateMap(),
		commandCreateJSON(),
		commandCreateJSONSingleLine(),
		commandCreateClass(),
		commandCreateEntityObject(),

		commandAllowMixing(),
		commandCreateElementParenthesis(),
		commandLayoutNewLine(),

		abel.CommandPackage[*ClassDiagram](),
		abel.CommandEndPackage[*ClassDiagram](),
		abel.CommandPackageEmpty[*ClassDiagram](),
		abel.CommandPackageWithUSymbol[*ClassDiagram](),
		abel.CommandTogether[*ClassDiagram](),

		commandCreateElementFull2(modeNormalKeyword),
		commandCreateElementFull2(modeWithMixPrefix),
		abel.CommandNoteSingle[*ClassDiagram](),

		commandNamespace(),
		commandNamespace2(),
		commandNamespaceEmpty(),
		commandStereotype(),

		commandLinkClass(),
		commandLinkLollipop(),

		abel.CommandTipOnEntityMulti[*ClassDiagram](true),
		abel.CommandTipOnEntityMulti[*ClassDiagram](false),

		abel.CommandNoteOnEntitySingle[*ClassDiagram](abel.CodeForClass()),
		abel.CommandURL[*ClassDiagram](),

		abel.CommandNoteOnEntityMulti[*ClassDiagram](abel.CodeForClass(), true),
		abel.CommandNoteOnEntityMulti[*ClassDiagram](abel.CodeForClass(), false),
		abel.CommandNoteMulti[*ClassDiagram](),

		abel.CommandNoteOnLinkSingle[*ClassDiagram](),
		abel.CommandNoteOnLinkMulti[*ClassDiagram](),
		abel.CommandConstraintOnLinks[*ClassDiagram](),

		commandDiamondAssociation(),

		abel.CommandCreateElementMultilines[*ClassDiagram](0),
		abel.CommandCreateElementMultilines[*ClassDiagram](1),
	)
	cmds = uml.AddTitleCommands(cmds)
	return uml.AddCommonCommands2(cmds)
}
