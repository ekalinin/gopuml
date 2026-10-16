// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/AbstractEntityDiagram.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/UmlSource.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"regexp"
	"strconv"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// AbstractEntityDiagram is the base of the class, description and state
// diagrams: CucaDiagram with their description, DOT settings and package
// packing.
type AbstractEntityDiagram struct {
	CucaDiagram
}

// DotStrings returns the DOT settings of the diagram, with the
// "graphattributes" pragma.
func (d *AbstractEntityDiagram) DotStrings() []string {
	def := []string{"nodesep=.35;", "ranksep=0.8;", "edge [fontsize=11,labelfontsize=11];",
		"node [fontsize=11,height=.35,width=.55];"}
	if !d.Pragma().IsDefine(klimt.PragmaGraphAttributes) {
		return def
	}
	attribute, _ := d.Pragma().Value(klimt.PragmaGraphAttributes)
	return append(def, attribute)
}

// sourceID is upstream's "id=([\w]+)\b" pattern of UmlSource.getId.
var sourceID = regexp.MustCompile(`id=([a-zA-Z0-9_]+)\b`)

// Description implements core.Diagram: "(N entities)", with the id of the
// start line ("@startuml id=foo") when there is one.
func (d *AbstractEntityDiagram) Description() string {
	result := "(" + strconv.Itoa(len(d.Leafs())) + " entities"
	if src := d.Source(); src != nil {
		if m := sourceID.FindStringSubmatch(src.Lines()[0].S); m != nil {
			result += ", " + m[1]
		}
	}
	return result + ")"
}

// PackSomePackage joins the title of each group holding a single group to
// its child's ("a.b"), as long as some group can be packed.
func (d *AbstractEntityDiagram) PackSomePackage() {
	separator, ok := d.NamespaceSeparator()
	if !ok {
		separator = "."
	}
	for {
		changed := false
		for _, group := range d.Groups() {
			if group.CanBePacked() {
				child := group.Groups()[0]
				appended := group.Display().Get(0).String() + separator
				child.SetDisplay(child.Display().AppendFirstLine(appended))
				group.SetPacked(true)
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}
