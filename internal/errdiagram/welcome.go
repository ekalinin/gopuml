// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/eggs/PSystemWelcome.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/eggs/PSystemWelcomeFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package errdiagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Welcome is the welcome screen of an empty @startuml block, upstream's
// PSystemWelcome.
type Welcome struct {
	source *core.UmlSource
}

// Description implements core.Diagram.
func (*Welcome) Description() string { return "(Empty)" }

// WelcomeFactory is upstream's PSystemWelcomeFactory: it accepts the
// @startuml blocks of two lines (the start and end lines), and goes first
// in the factory list.
type WelcomeFactory struct{}

var _ core.Factory = WelcomeFactory{}

// DiagramType returns core.Sequence, upstream's hack to be tried for
// "@startuml".
func (WelcomeFactory) DiagramType() core.DiagramType { return core.Sequence }

// CreateSystem returns the welcome screen of a block of two lines, nil
// otherwise.
func (WelcomeFactory) CreateSystem(_ *core.RenderContext, src *core.UmlSource) core.Diagram {
	if src.TotalLineCount() == 2 {
		return &Welcome{source: src}
	}
	return nil
}

// welcomeLines are the strings of PSystemWelcome; position
// BACKGROUND_CORNER_BOTTOM_RIGHT adds four blank lines, where the logo goes.
func welcomeLines(position geom.GraphicPosition) []string {
	lines := []string{
		"<b>Welcome to PlantUML!",
		" ",
		"You can start with a simple UML Diagram like:",
		" ",
		`""Bob->Alice: Hello""`,
		" ",
		"Or",
		" ",
		`""class Example""`,
		" ",
		"You will find more information about PlantUML syntax on <u>https://plantuml.com</u>",
		" ",
		`(Details by typing ""license"" keyword)`,
		" ",
	}
	if position == geom.GraphicPositionBackgroundCornerBottomRight {
		lines = append(lines, " ", " ", " ", " ")
	}
	return lines
}

// welcomeTextBlock is PSystemWelcome.getTextBlock with a position: the
// strings in black on white with the logo.
func welcomeTextBlock(position geom.GraphicPosition) klimt.TextBlock {
	return klimt.GraphicStringsCreateBlackOnWhiteImage(welcomeLines(position), plantumlImage(), position)
}
