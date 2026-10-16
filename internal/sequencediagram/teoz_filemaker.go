// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/teoz/SequenceDiagramFileMakerTeoz.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/graphic/FileMaker.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/rose"
)

// FileMaker draws the pages of a sequence diagram. Upstream's FileMaker
// also returns the ASCII block of a page (getAsciiBlock), which the text
// output step adds.
type FileMaker interface {
	NbPages() int
	TextBlock(num int) klimt.TextBlock
}

// SequenceDiagramFileMakerTeoz lays out a sequence diagram with Teoz: it
// builds the living spaces and the tiles, registers their constraints and
// solves the X and Y lines.
type SequenceDiagramFileMakerTeoz struct {
	diagram       *SequenceDiagram
	skin          *rose.Rose
	dolls         *Dolls
	stringBounder font.StringBounder

	body *PlayingSpaceWithParticipants

	min1 Real

	livingSpaces    *LivingSpaces
	heightEnglober1 float64
	heightEnglober2 float64
}

// NewSequenceDiagramFileMakerTeoz lays out the diagram with the text
// measurer of the output format.
func NewSequenceDiagramFileMakerTeoz(diagram *SequenceDiagram, skin *rose.Rose,
	stringBounder font.StringBounder) *SequenceDiagramFileMakerTeoz {
	m := &SequenceDiagramFileMakerTeoz{diagram: diagram, skin: skin, stringBounder: stringBounder,
		livingSpaces: NewLivingSpaces()}
	m.body = NewPlayingSpaceWithParticipants(m.createMainTile())

	m.min1 = m.body.MinX(stringBounder)

	m.heightEnglober1 = m.dolls.OffsetForEnglobers(stringBounder)
	if m.heightEnglober1 != 0 {
		m.heightEnglober2 = 10
	}
	return m
}

func (m *SequenceDiagramFileMakerTeoz) createMainTile() *PlayingSpace {
	xorigin := RealUtilsCreateOrigin()
	xcurrent := xorigin.AddAtLeast(0)
	yorigin := RealUtilsCreateOrigin()
	for _, p := range m.diagram.Participants() {
		livingSpace := NewLivingSpace(p, m.diagram.Englober(p), m.skin, m.skinParam(), xcurrent,
			m.diagram.Events())
		m.livingSpaces.Put(p, livingSpace)
		xcurrent = livingSpace.PosD(m.stringBounder).AddAtLeast(0)
	}

	tileArguments := NewTileArguments(m.stringBounder, m.livingSpaces, m.skin, m.diagram.SkinParam(), xorigin,
		yorigin)

	m.dolls = NewDolls(tileArguments)
	// The dolls constraints are set before the PlayingSpace is built: it
	// captures their bounds, which freeze the margins needed for the box
	// titles
	m.dolls.AddConstraints(m.stringBounder)
	mainTile := NewPlayingSpace(m.diagram, m.dolls, tileArguments)
	m.livingSpaces.AddConstraints(m.stringBounder)
	mainTile.AddConstraints()
	xorigin.CompileNow()
	// The Y line is solved once, before anything reads a gauge
	yorigin.CompileNow()
	tileArguments.SetBordered(mainTile)
	return mainTile
}

func (m *SequenceDiagramFileMakerTeoz) skinParam() klimt.ISkinParam { return m.diagram.SkinParam() }

// NbPages returns the number of pages.
func (m *SequenceDiagramFileMakerTeoz) NbPages() int { return m.body.NbPages() }

// TextBlock returns the drawing of page num.
func (m *SequenceDiagramFileMakerTeoz) TextBlock(num int) klimt.TextBlock {
	return &teozPage{maker: m, num: num}
}

// teozPage is the anonymous text block of upstream's getTextBlock.
type teozPage struct {
	klimt.TextBlockDefaults
	maker *SequenceDiagramFileMakerTeoz
	num   int
}

func (t *teozPage) DrawU(ug klimt.UGraphic) {
	m := t.maker
	ug = ug.Apply(klimt.NewUTranslate(5, 5))

	m.body.SetIndex(t.num)
	min1translate := klimt.UTranslateDx(-m.min1.CurrentValue())
	ug = ug.Apply(min1translate)

	m.dolls.DrawEnglobers(ug,
		m.body.CalculateDimension(m.stringBounder).Height()+m.heightEnglober1+m.heightEnglober2/2,
		klimt.NewSimpleContext2D(true))

	ug = ug.Apply(klimt.UTranslateDy(m.heightEnglober1))
	m.body.DrawU(ug)
}

func (t *teozPage) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	m := t.maker
	// The exporter calculates the dimension before drawing, so the page
	// index is set here too
	m.body.SetIndex(t.num)
	totalWidth := m.body.CalculateDimension(stringBounder).Width()
	totalHeight := m.body.CalculateDimension(stringBounder).Height() + m.heightEnglober1 + m.heightEnglober2
	return geom.NewXDimension2D(totalWidth+10, totalHeight+10)
}

func (t *teozPage) Backcolor() color.HColor { return nil }
