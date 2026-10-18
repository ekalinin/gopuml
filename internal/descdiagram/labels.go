// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/Labels.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/StringWithArrow.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package descdiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/command/regex"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Labels is upstream's Labels: the cardinalities, roles and label of a
// link, the cardinalities possibly written in the label
// (`: "1" label "many"`).
type Labels struct {
	FirstLabel, SecondLabel abel.OptString
	FirstRole, SecondRole   abel.OptString
	Label                   abel.OptString
	LinkArrow               abel.LinkArrow
}

var (
	labelsBoth       = uml.Compile(`^[%g]([^%g]+)[%g]([^%g]+)[%g]([^%g]+)[%g]$`)
	labelsFirstOnly  = uml.Compile(`^[%g]([^%g]+)[%g]([^%g]+)$`)
	labelsSecondOnly = uml.Compile(`^([^%g]+)[%g]([^%g]+)[%g]$`)
)

func cleanRole(role string, ok bool) abel.OptString {
	if !ok {
		return abel.OptString{}
	}
	return abel.Some(text.RemoveQuotesFormat(role, `"`))
}

// NewLabels reads the labels of a link command.
func NewLabels(arg *regex.Result) *Labels {
	l := &Labels{
		FirstLabel:  abel.Opt(arg.Get("FIRST_LABEL", 0)),
		SecondLabel: abel.Opt(arg.Get("SECOND_LABEL", 0)),
		FirstRole:   cleanRole(arg.Get("FIRST_ROLE", 0)),
		SecondRole:  cleanRole(arg.Get("SECOND_ROLE", 0)),
	}
	labelLink, ok := arg.Get("LABEL_LINK", 0)
	if ok {
		labelLink = l.init(labelLink)
	}
	l.Label, l.LinkArrow = StringWithArrow(labelLink, ok)
	return l
}

func (l *Labels) init(labelLink string) string {
	if !l.FirstLabel.OK && !l.SecondLabel.OK {
		if m := labelsBoth.FindStringSubmatch(labelLink); m != nil {
			l.FirstLabel = abel.Some(m[1])
			l.SecondLabel = abel.Some(m[3])
			return text.Trin(uml.RemoveQuotes(text.Trin(m[2])))
		}
		if m := labelsFirstOnly.FindStringSubmatch(labelLink); m != nil {
			l.FirstLabel = abel.Some(m[1])
			l.SecondLabel = abel.OptString{}
			return text.Trin(uml.RemoveQuotes(text.Trin(m[2])))
		}
		if m := labelsSecondOnly.FindStringSubmatch(labelLink); m != nil {
			l.FirstLabel = abel.OptString{}
			l.SecondLabel = abel.Some(m[2])
			return text.Trin(uml.RemoveQuotes(text.Trin(m[1])))
		}
	}
	return text.RemoveQuotesFormat(labelLink, `"`)
}

// Display is the display of the label.
func (l *Labels) Display(pragma *klimt.Pragma) *klimt.Display {
	return klimt.DisplayWithNewlinesString(pragma, l.Label.S, l.Label.OK)
}

// StringWithArrow is upstream's StringWithArrow: a label with a leading or
// trailing "<" or ">" gives the label arrow.
func StringWithArrow(completeLabel string, ok bool) (abel.OptString, abel.LinkArrow) {
	if !ok {
		return abel.OptString{}, abel.LinkArrowNoneOrSeveral
	}
	completeLabel = text.RemoveQuotesFormat(completeLabel, `"`)
	if klimt.DisplayHasSeveralGuideLines(completeLabel) {
		return abel.Some(completeLabel), abel.LinkArrowNoneOrSeveral
	}
	switch {
	case completeLabel == "<":
		return abel.OptString{}, abel.LinkArrowBackward
	case completeLabel == ">":
		return abel.OptString{}, abel.LinkArrowDirectNormal
	case strings.HasPrefix(completeLabel, "< "):
		return abel.Some(text.Trin(completeLabel[2:])), abel.LinkArrowBackward
	case strings.HasPrefix(completeLabel, "> "):
		return abel.Some(text.Trin(completeLabel[2:])), abel.LinkArrowDirectNormal
	case strings.HasSuffix(completeLabel, " >"):
		return abel.Some(text.Trin(completeLabel[:len(completeLabel)-2])), abel.LinkArrowDirectNormal
	case strings.HasSuffix(completeLabel, " <"):
		return abel.Some(text.Trin(completeLabel[:len(completeLabel)-2])), abel.LinkArrowBackward
	}
	return abel.Some(completeLabel), abel.LinkArrowNoneOrSeveral
}
