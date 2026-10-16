// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/BoxStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/Stereogroup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// BoxStyle is the shape of an action box, selected by a stereotype such as
// <<input>>. Only the identity of the shapes is ported here; their drawing
// belongs to the ftile layout step.
type BoxStyle int

// Box styles, in upstream's order.
const (
	BoxStylePlain BoxStyle = iota
	BoxStyleSdlInput
	BoxStyleSdlOutput
	BoxStyleSdlProcedure
	BoxStyleSdlLoad
	BoxStyleSdlSave
	BoxStyleSdlContinuous
	BoxStyleSdlTask
	BoxStyleUmlObject
	BoxStyleUmlObjectSignal
	BoxStyleUmlTrigger
	BoxStyleUmlSendSignal
	BoxStyleUmlAcceptEvent
	BoxStyleUmlTimeEvent
	numBoxStyles
)

// boxStyleStereotypes are the stereotypes of the styles; the plain style
// has none.
var boxStyleStereotypes = [numBoxStyles]string{"", "input", "output", "procedure", "load", "save",
	"continuous", "task", "object", "objectSignal", "trigger", "sendSignal", "acceptEvent", "timeEvent"}

var boxStyleShields = [numBoxStyles]float64{0, 10, 10, 0, 0, 0, 0, 0, 0, 10, 10, 10, 10, 10}

// Name is upstream's name(): the stereotype of the style; ok is false for
// the plain style (null upstream).
func (s BoxStyle) Name() (string, bool) {
	if s == BoxStylePlain {
		return "", false
	}
	return boxStyleStereotypes[s], true
}

// Shield returns the width of the pointed sides of the shape.
func (s BoxStyle) Shield() float64 { return boxStyleShields[s] }

// Stereotype returns the stereotype of the style, nil for the plain style.
func (s BoxStyle) Stereotype() *klimt.Stereotype {
	if s == BoxStylePlain {
		return nil
	}
	return klimt.StereotypeBuild("<<" + boxStyleStereotypes[s] + ">>")
}

// BoxStyleFromString returns the style of a stereotype name, ignoring case
// and dashes; the plain style for an unknown name.
func BoxStyleFromString(style string) BoxStyle {
	style = strings.ReplaceAll(style, "-", "")
	for s := BoxStyleSdlInput; s < numBoxStyles; s++ {
		if uml.EqualsIgnoreCase(style, boxStyleStereotypes[s]) {
			return s
		}
	}
	return BoxStylePlain
}

// boxStyleOf is upstream's Stereogroup.getBoxStyle: the first label of the
// group that names a box style.
func boxStyleOf(stereogroup *klimt.Stereogroup) BoxStyle {
	for _, label := range stereogroup.Labels() {
		if tmp := BoxStyleFromString(label); tmp != BoxStylePlain {
			return tmp
		}
	}
	return BoxStylePlain
}
