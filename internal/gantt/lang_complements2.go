// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementFromTo.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementInColors.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementInColors2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementInColorsFromTo.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementNamed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementTask.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementUrl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementWithColorLink.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/ComplementWorkingHours.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// complementFromTo is upstream's ComplementFromTo: "from [A] to [B]".
type complementFromTo struct{}

func (complementFromTo) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("from"), klimt.UBrexSpaceOneOrMore(), taskCodeUbrex("COMPLEMENT1"),
		klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("to"), klimt.UBrexSpaceOneOrMore(), taskCodeUbrex("COMPLEMENT2"))
}

func (complementFromTo) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](twoNames{firstValue(arg, "COMPLEMENT1"), firstValue(arg, "COMPLEMENT2")})
}

// colorOrNil is getColorOrWhite of a captured color, nil when the key has
// no value.
func colorOrNil(gantt *GanttDiagram, arg captureLookup, key string) color.HColor {
	if v, ok := arg.FindFirstValueByKey(key); ok {
		return gantt.ColorSet().GetColorOrWhite(v)
	}
	return nil
}

// complementInColors is upstream's ComplementInColors: "in red/blue".
type complementInColors struct{}

func (complementInColors) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("in"), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexNamed("COMPLEMENT1", klimt.UBrexLeaf("〇?# 〇+〴w")),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexLeaf("/"),
			klimt.UBrexNamed("COMPLEMENT2", klimt.UBrexLeaf("〇?# 〇+〴w")))))
}

func (complementInColors) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](NewCenterBorderColor(colorOrNil(gantt, arg, "COMPLEMENT1"),
		colorOrNil(gantt, arg, "COMPLEMENT2"), ""))
}

// complementInColors2 is upstream's ComplementInColors2: "colored in red".
type complementInColors2 struct{}

func (complementInColors2) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("colo〇?ured"), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexOptional(klimt.UBrexLeaf("in〇+〴s")), klimt.UBrexNamed("COMPLEMENT1", klimt.UBrexLeaf("〇?#〇+〴w")),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexLeaf("/"),
			klimt.UBrexNamed("COMPLEMENT2", klimt.UBrexLeaf("〇?#〇+〴w")))))
}

func (complementInColors2) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](NewCenterBorderColor(colorOrNil(gantt, arg, "COMPLEMENT1"),
		colorOrNil(gantt, arg, "COMPLEMENT2"), ""))
}

// complementInColorsFromTo is upstream's ComplementInColorsFromTo: "from
// red to green".
type complementInColorsFromTo struct{}

func (complementInColorsFromTo) complementUbrex() klimt.UBrexPart {
	c := func(key string) klimt.UBrexPart { return klimt.UBrexNamed(key, klimt.UBrexLeaf("〇?# 〇+〴w")) }
	return klimt.UBrexConcat(klimt.UBrexLeaf("from"), klimt.UBrexSpaceOneOrMore(), c("FROM1"),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexLeaf("/"), c("FROM2"))), klimt.UBrexSpaceOneOrMore(),
		klimt.UBrexLeaf("to"), klimt.UBrexSpaceOneOrMore(), c("TO1"),
		klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexLeaf("/"), c("TO2"))))
}

func (complementInColorsFromTo) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any]([]*CenterBorderColor{
		NewCenterBorderColor(colorOrNil(gantt, arg, "FROM1"), colorOrNil(gantt, arg, "FROM2"), ""),
		NewCenterBorderColor(colorOrNil(gantt, arg, "TO1"), colorOrNil(gantt, arg, "TO2"), "")})
}

// complementNamed is upstream's ComplementNamed: a name between brackets.
type complementNamed struct{}

func (complementNamed) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("["), klimt.UBrexNamed("COMPLEMENT", klimt.UBrexLeaf("〇+「〤[]」")),
		klimt.UBrexLeaf("]"))
}

func (complementNamed) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	return FailableOk[any](firstValue(arg, "COMPLEMENT"))
}

// complementTask is upstream's ComplementTask, whose capture keeps the
// brackets, so that a task is never found, as upstream.
type complementTask struct{}

func (complementTask) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexNamed("COMPLEMENT", taskCodeUbrex("FOO"))
}

func (complementTask) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	code := firstValue(arg, "COMPLEMENT")
	task := gantt.ExistingTask(code)
	if task == nil {
		return FailableError[any]("No such task " + code)
	}
	return FailableOk[any](task)
}

// complementURL is upstream's ComplementUrl: [[link]].
type complementURL struct{}

func (complementURL) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexNamed("COMPLEMENT", klimt.UBrexLeaf("[[〄>]]"))
}

func (complementURL) me(_ *GanttDiagram, arg captureLookup) Failable[any] {
	url := klimt.NewUbrexURLBuilder("", klimt.URLModeStrict).URL(firstValue(arg, "COMPLEMENT"))
	return FailableOk[any](url)
}

// complementWithColorLink is upstream's ComplementWithColorLink: "with
// red dotted link".
type complementWithColorLink struct{}

func (complementWithColorLink) complementUbrex() klimt.UBrexPart {
	optionalStyle := klimt.UBrexOptional(klimt.UBrexConcat(klimt.UBrexNamed("STYLE",
		klimt.UBrexOr(klimt.UBrexLeaf("dotted"), klimt.UBrexLeaf("bold"), klimt.UBrexLeaf("dashed"))),
		klimt.UBrexSpaceOneOrMore()))
	return klimt.UBrexConcat(klimt.UBrexLeaf("with"), klimt.UBrexSpaceOneOrMore(), optionalStyle,
		klimt.UBrexNamed("COLOR", klimt.UBrexLeaf("〇?# 〇+〴w")), klimt.UBrexSpaceOneOrMore(), optionalStyle,
		klimt.UBrexLeaf("link"))
}

func (complementWithColorLink) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	col1 := colorOrNil(gantt, arg, "COLOR")
	style, _ := arg.FindFirstValueByKey("STYLE")
	return FailableOk[any](NewCenterBorderColor(col1, col1, style))
}

// complementWorkingHours is upstream's ComplementWorkingHours.
type complementWorkingHours struct{}

func (complementWorkingHours) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(klimt.UBrexLeaf("working"), klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf("hours"))
}

func (complementWorkingHours) me(*GanttDiagram, captureLookup) Failable[any] {
	return FailableOk[any]("")
}
