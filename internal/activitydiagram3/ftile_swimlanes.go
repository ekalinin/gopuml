// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Swimlanes.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// Swimlanes holds the lanes and the instruction tree of a diagram while it
// is parsed: the root list, the instruction that receives the next ones,
// the current lane and the arrow that leads to the next instruction.
//
// Upstream's Swimlanes is also the text block that lays out and draws the
// diagram; that part (the ftile factories, lane widths and dividers) belongs
// to the ftile layout step.
type Swimlanes struct {
	skinParam *klimt.SkinParam
	pragma    *klimt.Pragma

	swimlanesRaw    []*Swimlane
	currentSwimlane *Swimlane

	root               Instruction
	currentInstruction Instruction

	nextLinkRenderer *LinkRendering
}

// NewSwimlanes returns the lanes of a diagram, with an empty root list.
func NewSwimlanes(skinParam *klimt.SkinParam, pragma *klimt.Pragma) *Swimlanes {
	root := InstructionListEmpty()
	return &Swimlanes{skinParam: skinParam, pragma: pragma, root: root, currentInstruction: root,
		nextLinkRenderer: LinkRenderingNone()}
}

// StyleSignature returns the style signature of the lanes.
func (s *Swimlanes) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameSwimlane)
}

// List returns the lanes in order of creation.
func (s *Swimlanes) List() []*Swimlane { return s.swimlanesRaw }

// Swimlane makes the lane name current, creating it when it does not
// exist; color, when not nil, becomes its background and label, when not
// null, its title.
func (s *Swimlanes) Swimlane(name string, c color.HColor, label *klimt.Display) {
	s.currentSwimlane = s.getOrCreate(name)
	if c != nil {
		s.currentSwimlane.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	if !klimt.DisplayIsNull(label) {
		s.currentSwimlane.SetDisplay(label)
	}
}

func (s *Swimlanes) getOrCreate(name string) *Swimlane {
	for _, sw := range s.swimlanesRaw {
		if sw.Name() == name {
			return sw
		}
	}
	result := NewSwimlane(name, len(s.swimlanesRaw), s.pragma)
	s.swimlanesRaw = append(s.swimlanesRaw, result)
	return result
}

// Root returns the root instruction list.
func (s *Swimlanes) Root() Instruction { return s.root }

// Current returns the instruction that receives the next instructions.
func (s *Swimlanes) Current() Instruction { return s.currentInstruction }

// SetCurrent sets the instruction that receives the next instructions.
func (s *Swimlanes) SetCurrent(current Instruction) { s.currentInstruction = current }

// NextLinkRenderer returns the arrow that leads to the next instruction.
func (s *Swimlanes) NextLinkRenderer() *LinkRendering { return s.nextLinkRenderer }

// SetNextLinkRenderer sets the arrow that leads to the next instruction; a
// nil link panics, as upstream's requireNonNull.
func (s *Swimlanes) SetNextLinkRenderer(link *LinkRendering) {
	if link == nil {
		uml.ThrowNPE("")
	}
	s.nextLinkRenderer = link
}

// CurrentSwimlane returns the current lane, nil before the first one.
func (s *Swimlanes) CurrentSwimlane() *Swimlane { return s.currentSwimlane }

// CurrentStyleBuilder returns the style builder of the diagram.
func (s *Swimlanes) CurrentStyleBuilder() *klimt.StyleBuilder {
	return s.skinParam.CurrentStyleBuilder()
}
