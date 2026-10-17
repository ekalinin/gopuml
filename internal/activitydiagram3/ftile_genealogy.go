// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/Genealogy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Genealogy records the parent of each tile below a root, to find the
// position of a descendant.
type Genealogy struct {
	myFatherIs map[Ftile]Ftile
	root       Ftile
}

// NewGenealogy returns the genealogy of the tiles below root.
func NewGenealogy(root Ftile) *Genealogy {
	g := &Genealogy{myFatherIs: map[Ftile]Ftile{}, root: root}
	g.process(root)
	return g
}

func (g *Genealogy) process(current Ftile) {
	for _, child := range current.MyChildren() {
		g.setMyFather(child, current)
		g.process(child)
	}
}

// MyFather returns the parent of me, nil for none.
func (g *Genealogy) MyFather(me Ftile) Ftile { return g.myFatherIs[me] }

func (g *Genealogy) setMyFather(child, father Ftile) {
	if _, ok := g.myFatherIs[child]; ok {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	g.myFatherIs[child] = father
}

// Translate returns the position of child in the root.
func (g *Genealogy) Translate(child Ftile, stringBounder font.StringBounder) klimt.UTranslate {
	current := child
	result := klimt.UTranslateNone()
	for current != g.root {
		father := g.MyFather(current)
		if father == nil {
			uml.ThrowNPE("")
		}
		tr := father.TranslateFor(current, stringBounder)
		if tr == nil {
			uml.ThrowNPE("")
		}
		result = tr.Compose(result)
		current = father
	}
	return result
}
