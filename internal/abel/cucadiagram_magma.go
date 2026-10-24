// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/Magma.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/MagmaList.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/SquareMaker.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/SquareLinker.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/gopuml/internal/decoration"
	"github.com/ekalinin/gopuml/internal/jmath"
)

// magma is a set of standalone entities of one group, put in a square by
// invisible links.
type magma struct {
	diagram     *CucaDiagram
	standalones []*Entity
	linkType    decoration.LinkType
}

func newMagma(diagram *CucaDiagram, standalones []*Entity) *magma {
	return &magma{diagram: diagram, standalones: standalones,
		linkType: decoration.NewLinkType(decoration.LinkDecorNone, decoration.LinkDecorNone).Invisible()}
}

func (m *magma) putInSquare() {
	squareMakerPutInSquare(m.standalones, func(top, down *Entity) {
		m.diagram.AddLink(NewLink(nil, m.diagram, m.diagram.SkinParam().CurrentStyleBuilder(), top, down, m.linkType,
			LinkArgNoDisplay(2)))
	}, func(left, right *Entity) {
		m.diagram.AddLink(NewLink(nil, m.diagram, m.diagram.SkinParam().CurrentStyleBuilder(), left, right,
			m.linkType, LinkArgNoDisplay(1)))
	})
}

func (m *magma) container() *Entity {
	parent := m.standalones[0].ParentContainer()
	if parent == nil {
		return nil
	}
	return parent.ParentContainer()
}

func (m *magma) squareSize() int { return squareMakerComputeBranch(len(m.standalones)) }

func (m *magma) topLeft() *Entity { return m.standalones[0] }

func (m *magma) bottomLeft() *Entity {
	return m.standalones[squareMakerBottomLeft(len(m.standalones))]
}

func (m *magma) topRight() *Entity { return m.standalones[m.squareSize()-1] }

func (m *magma) linkToDown(down *magma) {
	m.diagram.AddLink(NewLink(nil, m.diagram, m.diagram.SkinParam().CurrentStyleBuilder(), m.bottomLeft(),
		down.topLeft(), m.linkType, LinkArgNoDisplay(2)))
}

func (m *magma) linkToRight(right *magma) {
	m.diagram.AddLink(NewLink(nil, m.diagram, m.diagram.SkinParam().CurrentStyleBuilder(), m.topRight(),
		right.topLeft(), m.linkType, LinkArgNoDisplay(1)))
}

// magmaList is upstream's MagmaList.
type magmaList struct {
	all []*magma
}

func (l *magmaList) magmas(group *Entity) *magmaList {
	result := &magmaList{}
	for _, m := range l.all {
		if m.container() == group {
			result.all = append(result.all, m)
		}
	}
	return result
}

func (l *magmaList) putInSquare() {
	squareMakerPutInSquare(l.all, func(top, down *magma) { top.linkToDown(down) },
		func(left, right *magma) { left.linkToRight(right) })
}

// squareMakerPutInSquare is upstream's SquareMaker.putInSquare: the
// elements in rows of computeBranch elements, each linked to the next of
// its row and the first of each row to the first of the next row.
func squareMakerPutInSquare[O any](data []O, topDown, leftRight func(a, b O)) {
	branch := squareMakerComputeBranch(len(data))
	headBranch := 0
	for i := 1; i < len(data); i++ {
		dist := i - headBranch
		ent2 := data[i]
		if dist == branch {
			topDown(data[headBranch], ent2)
			headBranch = i
		} else {
			leftRight(data[i-1], ent2)
		}
	}
}

// squareMakerComputeBranch returns the side of the smallest square holding
// size elements.
func squareMakerComputeBranch(size int) int {
	r := int(jmath.Sqrt(float64(size)))
	if r*r == size {
		return r
	}
	return r + 1
}

func squareMakerBottomLeft(size int) int {
	s := squareMakerComputeBranch(size)
	line := (size - 1) / s
	return line * s
}
