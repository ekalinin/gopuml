// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSLink.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// wbsLink is a link from one node to another, drawn over the tree.
type wbsLink struct {
	element1 *wElement
	element2 *wElement
	color    color.HColor
}

// newWBSLink returns the link; a nil color throws IllegalArgumentException,
// as upstream.
func newWBSLink(element1, element2 *wElement, c color.HColor) *wbsLink {
	if c == nil {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return &wbsLink{element1: element1, element2: element2, color: c}
}
