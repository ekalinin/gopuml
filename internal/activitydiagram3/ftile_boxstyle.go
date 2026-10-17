// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/BoxStyle.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/stereo/Stereogroup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// BoxStyle is the shape of an action box, selected by a stereotype such as
// <<input>>.
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

const (
	boxStyleDeltaInputOutput = 10
	boxStyleDeltaContinuous  = 5.0
	boxStylePadding          = 5
)

// DrawMe draws the shape of size width x height (the shield included).
func (s BoxStyle) DrawMe(ug klimt.UGraphic, width, height, shadowing, roundCorner float64) {
	width -= s.Shield()
	if s == BoxStyleSdlProcedure {
		rect := klimt.URectangleBuild(width, height)
		rect.SetDeltaShadow(shadowing)
		ug.Draw(rect)
		vline := klimt.ULineV(height)
		ug.Apply(klimt.UTranslateDx(boxStylePadding)).Draw(vline)
		ug.Apply(klimt.UTranslateDx(width - boxStylePadding)).Draw(vline)
		return
	}
	shape := s.shape(width, height, roundCorner)
	shape.(klimt.Shadowable).SetDeltaShadow(shadowing)
	ug.Draw(shape)
}

// DrawMeDebug draws the bounds of the shape.
func (s BoxStyle) DrawMeDebug(ug klimt.UGraphic, width, height, shadowing, roundCorner float64) {
	rect := klimt.URectangleBuild(width, height)
	rect.SetDeltaShadow(shadowing)
	ug.Draw(rect)
}

// shape returns the outline of the shape of size width x height.
func (s BoxStyle) shape(width, height, roundCorner float64) klimt.UShape {
	switch s {
	case BoxStylePlain:
		return klimt.URectangleBuild(width, height).Rounded(roundCorner)
	case BoxStyleSdlInput, BoxStyleUmlTrigger:
		return boxStylePolygon(0, 0, width+boxStyleDeltaInputOutput, 0, width, height/2,
			width+boxStyleDeltaInputOutput, height, 0, height)
	case BoxStyleSdlOutput, BoxStyleUmlSendSignal:
		return boxStylePolygon(0, 0, width, 0, width+boxStyleDeltaInputOutput, height/2, width, height, 0, height)
	case BoxStyleSdlLoad:
		return boxStylePolygon(0, 0, width-boxStyleDeltaInputOutput, 0, width, height, boxStyleDeltaInputOutput,
			height)
	case BoxStyleSdlSave:
		return boxStylePolygon(boxStyleDeltaInputOutput, 0, width, 0, width-boxStyleDeltaInputOutput, height, 0,
			height)
	case BoxStyleSdlContinuous:
		result := klimt.UPathNone()
		result.Add([]float64{boxStyleDeltaContinuous, 0}, geom.SegMoveTo)
		result.Add([]float64{0, height / 2}, geom.SegLineTo)
		result.Add([]float64{boxStyleDeltaContinuous, height}, geom.SegLineTo)
		result.Add([]float64{width - boxStyleDeltaContinuous, 0}, geom.SegMoveTo)
		result.Add([]float64{width, height / 2}, geom.SegLineTo)
		result.Add([]float64{width - boxStyleDeltaContinuous, height}, geom.SegLineTo)
		return result
	case BoxStyleSdlTask, BoxStyleUmlObject:
		return klimt.URectangleBuild(width, height)
	case BoxStyleUmlObjectSignal:
		return boxStylePolygon(-boxStyleDeltaInputOutput, 0, width, 0, width+boxStyleDeltaInputOutput, height/2,
			width, height, -boxStyleDeltaInputOutput, height, 0, height/2)
	case BoxStyleUmlAcceptEvent:
		return boxStylePolygon(-boxStyleDeltaInputOutput, 0, width, 0, width, height, -boxStyleDeltaInputOutput,
			height, 0, height/2)
	case BoxStyleUmlTimeEvent:
		halfWidth := width / 2
		thirdHeight := height / 3
		return boxStylePolygon(halfWidth-thirdHeight, thirdHeight, halfWidth+thirdHeight, thirdHeight,
			halfWidth-thirdHeight, height, halfWidth+thirdHeight, height)
	}
	uml.Throw(uml.IllegalStateException, "")
	return nil
}

// boxStylePolygon returns the polygon of the coordinates x0, y0, x1, y1...
func boxStylePolygon(coords ...float64) *klimt.UPolygon {
	result := klimt.NewUPolygon()
	for i := 0; i < len(coords); i += 2 {
		result.AddPoint(coords[i], coords[i+1])
	}
	return result
}
