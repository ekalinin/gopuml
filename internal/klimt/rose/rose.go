// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/skin/rose/Rose.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package rose

import (
	"unicode"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// Participant is what Rose reads from a participant of a sequence diagram
// (upstream's sequencediagram.Participant, which imports this package):
// its two styles, the style of the participant (with its stereotype and
// colors) and the style of its stereotype label.
type Participant interface {
	UsedStyles() []*klimt.Style
}

// PaddingY is upstream's Rose.paddingY, the vertical padding of notes.
const PaddingY = 5

// rosePaddingX is upstream's Rose.paddingX, the horizontal padding of
// notes.
const rosePaddingX = 5

// Rose is the skin of sequence diagrams: it builds the components that
// draw their elements. It has no state.
type Rose struct{}

// NewRose returns the skin.
func NewRose() *Rose { return &Rose{} }

// HtmlColor returns the color of a parameter, or its default: upstream's
// two-argument getHtmlColor.
func (r *Rose) HtmlColor(skin klimt.ISkinParam, colorParam klimt.ColorParam) color.HColor {
	return r.HtmlColorStereo(skin, nil, colorParam)
}

// HtmlColorStereo returns the color of the first parameter that is set for
// the stereotype, or the default of the first parameter. It panics without
// parameter, as Java throws ArrayIndexOutOfBoundsException.
func (r *Rose) HtmlColorStereo(skin klimt.ISkinParam, stereotype *klimt.Stereotype,
	colorParams ...klimt.ColorParam) color.HColor {
	return klimt.SkinParamUtilsColor(skin, stereotype, colorParams...)
}

// Upstream's checkRose throws IllegalStateException when a subclass of Rose
// does not override the component factories; Go has no subclasses, so the
// check is not needed.

// CreateComponentNote returns the component of a note without position.
func (r *Rose) CreateComponentNote(styles []*klimt.Style, typ klimt.ComponentType, param klimt.ISkinParam,
	stringsToDisplay *klimt.Display, colors *color.Colors) klimt.Component {
	return r.createComponentNote(styles, typ, param, stringsToDisplay, colors, false)
}

// CreateComponentNoteAt returns the component of a note at a position: a
// note over several participants is centered unless noteTextAlignment is
// set.
func (r *Rose) CreateComponentNoteAt(styles []*klimt.Style, typ klimt.ComponentType, param klimt.ISkinParam,
	stringsToDisplay *klimt.Display, colors *color.Colors, notePosition NotePosition) klimt.Component {
	return r.createComponentNote(styles, typ, param, stringsToDisplay, colors,
		notePosition == NotePositionOverSeveral)
}

func (r *Rose) createComponentNote(styles []*klimt.Style, typ klimt.ComponentType, param klimt.ISkinParam,
	stringsToDisplay *klimt.Display, colors *color.Colors, overSeveral bool) klimt.Component {
	var textAlign, position geom.HorizontalAlignment
	if overSeveral {
		textAlign = param.HorizontalAlignment(klimt.AlignmentParamNoteTextAlignment, klimt.ArrowDirectionNull, false,
			geom.HorizontalAlignmentLeft)
		if textAlign == param.HorizontalAlignment(klimt.AlignmentParamNoteTextAlignment, klimt.ArrowDirectionNull,
			false, geom.HorizontalAlignmentCenter) {
			// Which means we use default
			position = textAlign
		} else {
			position = geom.HorizontalAlignmentCenter
		}
	} else {
		textAlign = param.HorizontalAlignment(klimt.AlignmentParamNoteTextAlignment, klimt.ArrowDirectionNull, false,
			klimt.HorizontalAlignmentNull)
		position = textAlign
	}

	switch typ {
	case klimt.ComponentTypeNote:
		return NewComponentRoseNote(style0(styles), stringsToDisplay, rosePaddingX, PaddingY, param, textAlign,
			position, colors)
	case klimt.ComponentTypeNoteHexagonal:
		return NewComponentRoseNoteHexagonal(style0(styles), stringsToDisplay, param, colors)
	case klimt.ComponentTypeNoteBox:
		return NewComponentRoseNoteBox(style0(styles), stringsToDisplay, param, colors)
	}
	throw(unsupportedOperation, typ.String())
	return nil
}

// style0 is upstream's styles[0], which throws NullPointerException for a
// null array.
func style0(styles []*klimt.Style) *klimt.Style {
	if styles == nil {
		throw(nullPointer, "")
	}
	return styles[0]
}

// CreateComponentParticipant returns the head or tail of a participant,
// drawn with its two styles. config is not used, as upstream.
func (r *Rose) CreateComponentParticipant(p Participant, typ klimt.ComponentType, _ *klimt.ArrowConfiguration,
	param klimt.ISkinParam, stringsToDisplay *klimt.Display) klimt.Component {
	styles := p.UsedStyles()

	padding := style0(styles).Padding()
	margin := styles[0].Margin()

	switch typ {
	case klimt.ComponentTypeParticipantHead, klimt.ComponentTypeParticipantTail:
		return NewComponentRoseParticipant(styles[0], styles[1], stringsToDisplay, param, minClassWidth(styles[0]),
			false, padding, margin)
	case klimt.ComponentTypeCollectionsHead, klimt.ComponentTypeCollectionsTail:
		return NewComponentRoseParticipant(styles[0], styles[1], stringsToDisplay, param, minClassWidth(styles[0]),
			true, padding, margin)
	case klimt.ComponentTypeActorHead:
		return NewComponentRoseActor(param.ActorStyle(), styles[0], styles[1], stringsToDisplay, true, param)
	case klimt.ComponentTypeActorTail:
		return NewComponentRoseActor(param.ActorStyle(), styles[0], styles[1], stringsToDisplay, false, param)
	case klimt.ComponentTypeBoundaryHead:
		return NewComponentRoseBoundary(styles[0], styles[1], stringsToDisplay, true, param)
	case klimt.ComponentTypeBoundaryTail:
		return NewComponentRoseBoundary(styles[0], styles[1], stringsToDisplay, false, param)
	case klimt.ComponentTypeControlHead:
		return NewComponentRoseControl(styles[0], styles[1], stringsToDisplay, true, param)
	case klimt.ComponentTypeControlTail:
		return NewComponentRoseControl(styles[0], styles[1], stringsToDisplay, false, param)
	case klimt.ComponentTypeEntityHead:
		return NewComponentRoseEntity(styles[0], styles[1], stringsToDisplay, true, param)
	case klimt.ComponentTypeEntityTail:
		return NewComponentRoseEntity(styles[0], styles[1], stringsToDisplay, false, param)
	case klimt.ComponentTypeQueueHead:
		return NewComponentRoseQueue(styles[0], styles[1], stringsToDisplay, true, param)
	case klimt.ComponentTypeQueueTail:
		return NewComponentRoseQueue(styles[0], styles[1], stringsToDisplay, false, param)
	case klimt.ComponentTypeDatabaseHead:
		return NewComponentRoseDatabase(styles[0], styles[1], stringsToDisplay, true, param)
	case klimt.ComponentTypeDatabaseTail:
		return NewComponentRoseDatabase(styles[0], styles[1], stringsToDisplay, false, param)
	}
	throw(unsupportedOperation, "")
	return nil
}

// CreateComponent returns the component of an element that is neither a
// note nor a participant head. Without styles it panics with
// UnsupportedOperationException, except for the delay line, the grouping
// space and the destroy cross, as upstream; an arrow type always panics
// with NullPointerException, as upstream passes null styles to
// createComponentArrow.
func (r *Rose) CreateComponent(styles []*klimt.Style, typ klimt.ComponentType, config *klimt.ArrowConfiguration,
	param klimt.ISkinParam, stringsToDisplay *klimt.Display) klimt.Component {
	if typ != klimt.ComponentTypeDelayLine && typ != klimt.ComponentTypeGroupingSpace &&
		typ != klimt.ComponentTypeDestroy && styles == nil {
		throw(unsupportedOperation, typ.String())
	}

	if typ.IsArrow() {
		return r.CreateComponentArrow(nil, config, param, stringsToDisplay)
	}

	switch typ {
	case klimt.ComponentTypeParticipantLine:
		return NewComponentRoseLine(style0(styles), false, stringsToDisplay, param)
	case klimt.ComponentTypeNote, klimt.ComponentTypeNoteHexagonal, klimt.ComponentTypeNoteBox:
		throw(unsupportedOperation, "")
	case klimt.ComponentTypeGroupingHeaderLegacy:
		return NewComponentRoseGroupingHeader(false, style0(styles), styles[1], stringsToDisplay, param)
	case klimt.ComponentTypeGroupingHeaderTeoz:
		return NewComponentRoseGroupingHeader(true, style0(styles), styles[1], stringsToDisplay, param)
	case klimt.ComponentTypeGroupingElseLegacy:
		return NewComponentRoseGroupingElse(false, style0(styles), stringsToDisplay.Get(0), param)
	case klimt.ComponentTypeGroupingElseTeoz:
		return NewComponentRoseGroupingElse(true, style0(styles), stringsToDisplay.Get(0), param)
	case klimt.ComponentTypeGroupingSpace:
		return NewComponentRoseGroupingSpace(7)
	case klimt.ComponentTypeActivationBoxCloseClose:
		return NewComponentRoseActiveLine(style0(styles), true, true, stringsToDisplay, param)
	case klimt.ComponentTypeActivationBoxCloseOpen:
		return NewComponentRoseActiveLine(style0(styles), true, false, stringsToDisplay, param)
	case klimt.ComponentTypeActivationBoxOpenClose:
		return NewComponentRoseActiveLine(style0(styles), false, true, stringsToDisplay, param)
	case klimt.ComponentTypeActivationBoxOpenOpen:
		return NewComponentRoseActiveLine(style0(styles), false, false, stringsToDisplay, param)
	case klimt.ComponentTypeDelayLine:
		return NewComponentRoseDelayLine(style0(styles), param)
	case klimt.ComponentTypeDelayText:
		return NewComponentRoseDelayText(style0(styles), stringsToDisplay, param)
	case klimt.ComponentTypeDestroy:
		return NewComponentRoseDestroy(style0(styles), param)
	case klimt.ComponentTypeNewpage:
		throw(unsupportedOperation, "")
	case klimt.ComponentTypeDivider:
		return NewComponentRoseDivider(style0(styles), stringsToDisplay, param)
	case klimt.ComponentTypeReference:
		return NewComponentRoseReference(style0(styles), styles[1], stringsToDisplay, param)
	case klimt.ComponentTypeEnglober:
		return NewComponentRoseEnglober(style0(styles), stringsToDisplay, param)
	}
	throw(unsupportedOperation, "")
	return nil
}

func minClassWidth(style *klimt.Style) float64 {
	return style.Value(klimt.PNameMinimumWidth).AsDouble()
}

// CreateComponentNewPage returns the line of a page break.
func (r *Rose) CreateComponentNewPage(styles []*klimt.Style, param klimt.ISkinParam) klimt.Component {
	return NewComponentRoseNewpage(style0(styles), param)
}

// CreateComponentArrow returns the arrow of a message: a self arrow, or an
// arrow whose label is aligned as the HorizontalAlignment of the arrow
// style says ("first", "direction", "reversedirection" or an alignment).
func (r *Rose) CreateComponentArrow(styles []*klimt.Style, config *klimt.ArrowConfiguration,
	param klimt.ISkinParam, stringsToDisplay *klimt.Display) klimt.ArrowComponent {
	styleWidth := style0(styles).WrapWidth()
	maxMessageSize := param.MaxMessageSize()
	if styleWidth.MaxWidth() > 0 {
		maxMessageSize = styleWidth
	}

	if config.ArrowDirection() == klimt.ArrowDirectionSelf {
		return NewComponentRoseSelfArrow(styles[0], stringsToDisplay, config, param, maxMessageSize,
			!param.StrictUmlStyle())
	}

	arrowDirection := config.ArrowDirection()

	signature := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameSequenceDiagram,
		klimt.SNameArrow)
	textStyle := signature.MergedStyle(param.CurrentStyleBuilder())
	value := textStyle.Value(klimt.PNameHorizontalAlignment).AsString()
	messageHorizontalAlignment := textStyle.HorizontalAlignment()

	switch {
	case equalsIgnoreCase("first", value):
		isReverseDefine := config.IsReverseDefine()
		if arrowDirection == klimt.ArrowDirectionRightToLeftReverse {
			if isReverseDefine {
				messageHorizontalAlignment = geom.HorizontalAlignmentLeft
			} else {
				messageHorizontalAlignment = geom.HorizontalAlignmentRight
			}
		} else {
			if isReverseDefine {
				messageHorizontalAlignment = geom.HorizontalAlignmentRight
			} else {
				messageHorizontalAlignment = geom.HorizontalAlignmentLeft
			}
		}
	case equalsIgnoreCase("direction", value):
		switch arrowDirection {
		case klimt.ArrowDirectionLeftToRightNormal:
			messageHorizontalAlignment = geom.HorizontalAlignmentLeft
		case klimt.ArrowDirectionRightToLeftReverse:
			messageHorizontalAlignment = geom.HorizontalAlignmentRight
		case klimt.ArrowDirectionBothDirection:
			messageHorizontalAlignment = geom.HorizontalAlignmentCenter
		}
	case equalsIgnoreCase("reversedirection", value):
		switch arrowDirection {
		case klimt.ArrowDirectionLeftToRightNormal:
			messageHorizontalAlignment = geom.HorizontalAlignmentRight
		case klimt.ArrowDirectionRightToLeftReverse:
			messageHorizontalAlignment = geom.HorizontalAlignmentLeft
		case klimt.ArrowDirectionBothDirection:
			messageHorizontalAlignment = geom.HorizontalAlignmentCenter
		}
	}

	return NewComponentRoseArrow(styles[0], stringsToDisplay, config, messageHorizontalAlignment, param,
		maxMessageSize, !param.StrictUmlStyle(), param.ResponseMessageBelowArrow())
}

// equalsIgnoreCase is Java's String.equalsIgnoreCase: characters equal
// once upper-cased, or once upper-cased and lower-cased.
func equalsIgnoreCase(a, b string) bool {
	if a == b {
		return true
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) != len(rb) {
		return false
	}
	for i, c1 := range ra {
		c2 := rb[i]
		if c1 == c2 {
			continue
		}
		u1, u2 := unicode.ToUpper(c1), unicode.ToUpper(c2)
		if u1 == u2 || unicode.ToLower(u1) == unicode.ToLower(u2) {
			continue
		}
		return false
	}
	return true
}

// Stroke is upstream's static Rose.getStroke: the thickness of a line
// parameter, or a stroke of the default thickness when it is not set.
func Stroke(param klimt.ISkinParam, lineParam klimt.LineParam, defaultValue float64) klimt.UStroke {
	result := param.Thickness(lineParam, nil)
	if result == nil {
		return klimt.UStrokeWithThickness(defaultValue)
	}
	return *result
}
