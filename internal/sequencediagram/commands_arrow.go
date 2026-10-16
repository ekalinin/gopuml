// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandArrow.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandExoArrowAny.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandExoArrowLeft.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/sequencediagram/command/CommandExoArrowRight.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/descdiagram/command/CommandLinkElement.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package sequencediagram

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/command/regex"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// anchorPattern is CommandArrow.ANCHOR: "{name} " before a participant.
const anchorPattern = `(\{([%pLN_]+)\}[%s]+)?`

// lineStyle is CommandLinkElement.LINE_STYLE.
const lineStyle = `(?:#\w+|dotted|dashed|plain|bold|hidden|norank|single|node|thickness=\d+)` +
	`(?:,#\w+|,dotted|,dashed|,plain|,bold|,hidden|,norank|,single|,node|,thickness=\d+)*`

// colorOrStylePattern is CommandArrow.getColorOrStylePattern: "[#red,dashed]".
const colorOrStylePattern = `(?:\[(` + lineStyle + `)\])?`

func participantOr(prefix string) regex.Part {
	return regex.Or(prefix,
		uml.Leaf(1, prefix+"CODE", "([%pLN_.@]+)"),
		uml.Leaf(1, prefix+"LONG", "[%g]([^%g]+)[%g]"),
		uml.Leaf(2, prefix+"LONGCODE", "[%g]([^%g]+)[%g][%s]*as[%s]+([%pLN_.@]+)"),
		uml.Leaf(2, prefix+"CODELONG", "([%pLN_.@]+)[%s]+as[%s]*[%g]([^%g]+)[%g]"))
}

// commandArrow is the message command "A -> B : label" with all its arrow
// forms.
func commandArrow() cmd {
	single := uml.Single(regex.NewConcat(regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		uml.Leaf(2, "ANCHOR", anchorPattern),
		participantOr("PART1"),
		uml.Leaf(2, "PART1ANCHOR", anchorPattern),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.Or("ARROW_DRESSING1",
			uml.Lit("[%s][ox]"),
			uml.Lit(`(?:[%s][ox]|\(\d+\))?<<?_?`),
			uml.Lit("(?:[%s][ox])?//?"),
			uml.Lit(`(?:[%s][ox])?\\\\?`))),
		regex.Or("",
			regex.NewConcat(
				uml.Leaf(1, "ARROW_BODYA1", "(-+)"),
				uml.Leaf(1, "ARROW_STYLE1", colorOrStylePattern),
				uml.Leaf(1, "ARROW_BODYB1", "(-*)")),
			regex.NewConcat(
				uml.Leaf(1, "ARROW_BODYA2", "(-*)"),
				uml.Leaf(1, "ARROW_STYLE2", colorOrStylePattern),
				uml.Leaf(1, "ARROW_BODYB2", "(-+)"))),
		regex.Optional(regex.Or("ARROW_DRESSING2",
			uml.Lit(`_?>>?(?:[ox][%s]|\(\d+\))?`),
			uml.Lit("//?(?:[ox][%s])?"),
			uml.Lit(`\\\\?(?:[ox][%s])?`),
			uml.Lit("[ox][%s]"))),
		regex.SpaceZeroOrMore(),
		participantOr("PART2"),
		uml.Leaf(1, "MULTICAST", `((?:\s&\s[%pLN_.@]+)*)`),
		uml.Leaf(2, "PART2ANCHOR", anchorPattern),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "ACTIVATION", `(?:(\+\+|\*\*|!!|--|--\+\+|\+\+--)?)`),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "LIFECOLOR", `(?:(#\w+)?)`),
		klimt.StereotypePatternOptional("STEREOTYPE"),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "MESSAGE", "(?::[%s]*(.*))?"),
		regex.End()), executeArrow)
	return &uml.Guarded[*SequenceDiagram]{SingleLine: single, MaxSize: 2000}
}

func containsAny(s string, tests ...string) bool {
	for _, t := range tests {
		if strings.Contains(s, t) {
			return true
		}
	}
	return false
}

// dressing is CommandArrow.getDressing: the dressing without '_', in lower
// case.
func dressing(arg *regex.Result, key string) string {
	return uml.LowerCase(strings.ReplaceAll(arg.Str(key, 0), "_", ""))
}

// inclination is CommandArrow.getInclination: the number in "(n)".
func inclination(key string, ok bool) int {
	if !ok {
		return 0
	}
	x1 := strings.IndexByte(key, '(')
	if x1 == -1 {
		return 0
	}
	x2 := strings.IndexByte(key, ')')
	if x2 == -1 {
		return 0
	}
	return uml.JavaParseInt(key[x1+1 : x2])
}

// arrowParticipant is CommandArrow.getOrCreateParticipant.
func arrowParticipant(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, n string) *Participant {
	if code, ok := arg.Get(n+"CODE", 0); ok {
		return d.OrCreateParticipantDisplay(location, code, klimt.DisplayWithNewlines(d.Pragma(), code))
	}
	if code, ok := arg.Get(n+"LONG", 0); ok {
		return d.OrCreateParticipantDisplay(location, code, klimt.DisplayWithNewlines(d.Pragma(), code))
	}
	if long, ok := arg.Get(n+"LONGCODE", 0); ok {
		return d.OrCreateParticipantDisplay(location, arg.Str(n+"LONGCODE", 1), klimt.DisplayWithNewlines(d.Pragma(), long))
	}
	if code, ok := arg.Get(n+"CODELONG", 0); ok {
		return d.OrCreateParticipantDisplay(location, code, klimt.DisplayWithNewlines(d.Pragma(), arg.Str(n+"CODELONG", 1)))
	}
	uml.Throw(uml.IllegalStateException, "")
	return nil
}

// applyStyle is CommandArrow.applyStyle: "dashed", "dotted", "hidden" and
// colors, comma separated.
func applyStyle(arrowStyle string, ok bool, config ArrowConfiguration) (ArrowConfiguration, error) {
	if !ok {
		return config, nil
	}
	for _, s := range strings.Split(arrowStyle, ",") {
		if s == "" {
			continue
		}
		switch {
		case uml.EqualsIgnoreCase(s, "dashed"), uml.EqualsIgnoreCase(s, "dotted"):
			config = config.WithBody(ArrowBodyDotted)
		case uml.EqualsIgnoreCase(s, "bold"):
		case uml.EqualsIgnoreCase(s, "hidden"):
			config = config.WithBody(ArrowBodyHidden)
		default:
			c, err := getColor(s)
			if err != nil {
				return config, err
			}
			config = config.WithColor(c)
		}
	}
	return config, nil
}

func lifeColor(arg *regex.Result) (color.HColor, error) {
	s, ok := arg.Get("LIFECOLOR", 0)
	if !ok {
		return nil, nil
	}
	return getColor(s)
}

func executeArrow(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
	dressing1 := dressing(arg, "ARROW_DRESSING1")
	dressing2 := dressing(arg, "ARROW_DRESSING2")
	inclination1 := inclination(arg.Get("ARROW_DRESSING1", 0))
	inclination2 := inclination(arg.Get("ARROW_DRESSING2", 0))

	hasDressing1butx := containsAny(dressing1, "<", "\\", "/")
	xInDressing1 := strings.Contains(dressing1, "x")
	hasDressing2butx := containsAny(dressing2, ">", "\\", "/")
	xInDressing2 := strings.Contains(dressing2, "x")
	var reverseDefine bool
	switch {
	case hasDressing2butx || (xInDressing1 && xInDressing2):
		reverseDefine = false
	case hasDressing1butx:
		reverseDefine = true
	case xInDressing1 || xInDressing2:
		reverseDefine = false
	default:
		return command.Error("Illegal sequence arrow")
	}

	var p1, p2 *Participant
	var circleAtStart, circleAtEnd, sync1, sync2 bool
	if reverseDefine {
		p2 = arrowParticipant(d, location, arg, "PART1")
		p1 = arrowParticipant(d, location, arg, "PART2")
		circleAtStart = strings.Contains(dressing2, "o")
		circleAtEnd = strings.Contains(dressing1, "o")
		sync2 = containsAny(dressing1, "<<", "\\\\", "//")
		sync1 = containsAny(dressing2, ">>", "\\\\", "//")
	} else {
		p1 = arrowParticipant(d, location, arg, "PART1")
		p2 = arrowParticipant(d, location, arg, "PART2")
		circleAtStart = strings.Contains(dressing1, "o")
		circleAtEnd = strings.Contains(dressing2, "o")
		sync1 = containsAny(dressing1, "<<", "\\\\", "//")
		sync2 = containsAny(dressing2, ">>", "\\\\", "//")
	}

	bodyA, _ := arg.Lazy("ARROW_BODYA", 0)
	bodyB, _ := arg.Lazy("ARROW_BODYB", 0)
	dotted := len(bodyA)+len(bodyB) > 1

	var labels *klimt.Display
	if message, ok := arg.Get("MESSAGE", 0); ok {
		labels = klimt.DisplayWithNewlines(d.Pragma(), message)
	} else {
		labels = klimt.DisplayCreate("")
	}

	config := ArrowConfigurationNormal()
	if hasDressing1butx && hasDressing2butx {
		config = ArrowConfigurationBoth()
	}
	if dotted {
		config = config.WithBody(ArrowBodyDotted)
	}
	if sync1 {
		config = config.WithHead1(ArrowHeadAsync)
	}
	if sync2 {
		config = config.WithHead2(ArrowHeadAsync)
	}
	if strings.Contains(dressing2, "\\") || strings.Contains(dressing1, "/") {
		config = config.WithPart(ArrowPartTopPart)
	}
	if strings.Contains(dressing2, "/") || strings.Contains(dressing1, "\\") {
		config = config.WithPart(ArrowPartBottomPart)
	}
	if circleAtEnd {
		config = config.WithDecoration2(ArrowDecorationCircle)
	}
	if circleAtStart {
		config = config.WithDecoration1(ArrowDecorationCircle)
	}
	if reverseDefine {
		if xInDressing1 {
			config = config.WithHead2(ArrowHeadCrossX)
		}
		if xInDressing2 {
			config = config.WithHead1(ArrowHeadCrossX)
		}
	} else {
		if xInDressing1 {
			config = config.WithHead1(ArrowHeadCrossX)
		}
		if xInDressing2 {
			config = config.WithHead2(ArrowHeadCrossX)
		}
	}
	if reverseDefine {
		config = config.ReverseDefine()
	}
	style, hasStyle := arg.Lazy("ARROW_STYLE", 0)
	config, err := applyStyle(style, hasStyle, config)
	if err != nil {
		return colorResult(err)
	}
	config = config.WithInclination(inclination1 + inclination2)

	activationSpec, hasActivation := arg.Get("ACTIVATION", 0)
	if hasActivation && activationSpec[0] == '*' {
		d.Activate(p2, LifeEventCreate, nil, nil)
	}

	number, hasNumber := d.NextMessageNumber()
	msg := NewMessage(d.CurrentStyleBuilder(), p1, p2, d.ManageVariable(labels), config, number, hasNumber, location)
	msg.SetMulticast(multicasts(d, location, arg))
	if url, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		msg.SetURL(urlBuilder(d).URL(url))
	}
	if stereotype, ok := arg.Get("STEREOTYPE", 0); ok {
		msg.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	if _, ok := arg.Get("PARALLEL", 0); ok {
		msg.GoParallel()
	}
	msg.SetAnchor(arg.Get("ANCHOR", 1))
	msg.SetPart1Anchor(arg.Get("PART1ANCHOR", 1))
	msg.SetPart2Anchor(arg.Get("PART2ANCHOR", 1))

	if status := d.AddMessage(msg); !status.IsOK() {
		return status
	}
	activationColor, err := lifeColor(arg)
	if err != nil {
		return colorResult(err)
	}
	if hasActivation {
		return manageActivations(activationSpec, d, p1, p2, activationColor)
	}
	if d.IsAutoactivate() && (config.Head() == ArrowHeadNormal || config.Head() == ArrowHeadAsync) {
		if config.IsDotted() {
			d.Activate(p1, LifeEventDeactivate, nil, nil)
		} else {
			d.Activate(p2, LifeEventActivate, activationColor, nil)
		}
	}
	return command.Ok()
}

// multicasts is CommandArrow.getMulticasts: the participants after "&".
func multicasts(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result) []*Participant {
	multicast, ok := arg.Get("MULTICAST", 0)
	if !ok {
		return nil
	}
	var result []*Participant
	for _, s := range uml.JavaSplit(multicast, "&") {
		s = text.Trin(s)
		if s == "" {
			continue
		}
		result = append(result, d.OrCreateParticipant(location, s))
	}
	return result
}

// manageActivations is CommandArrow.manageActivations: "++", "--", "**",
// "!!" and the "--++" and "++--" pairs. Like upstream, it ignores the
// errors of Activate.
func manageActivations(spec string, d *SequenceDiagram, p1, p2 *Participant, activationColor color.HColor) command.Result {
	switch spec[0] {
	case '+':
		d.Activate(p2, LifeEventActivate, activationColor, nil)
	case '-':
		d.Activate(p1, LifeEventDeactivate, nil, nil)
	case '!':
		d.Activate(p2, LifeEventDestroy, nil, nil)
	}
	if len(spec) == 4 {
		switch spec[2] {
		case '+':
			d.Activate(p2, LifeEventActivate, activationColor, nil)
		case '-':
			d.Activate(p1, LifeEventDeactivate, nil, nil)
		}
	}
	return command.Ok()
}

const (
	arrowSuppCircle1 = "ARROW_SUPPCIRCLE1"
	arrowSuppCircle2 = "ARROW_SUPPCIRCLE2"
)

// exoTail is the end shared by the border message patterns.
func exoTail() []regex.Part {
	return []regex.Part{
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "ACTIVATION", "(?:([+*!-]+)?)"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "LIFECOLOR", `(?:(#\w+)?)`),
		regex.SpaceZeroOrMore(),
		klimt.URLBuilderOptional(),
		regex.SpaceZeroOrMore(),
		regex.Optional(regex.NewConcat(uml.Lit(":"), regex.SpaceZeroOrMore(), uml.Leaf(1, "LABEL", "(.*)"))),
		regex.End(),
	}
}

func exoArrowBody() regex.Part {
	return regex.Or("",
		regex.NewConcat(
			uml.Leaf(1, "ARROW_BOTHDRESSING", `(<<?|//?|\\\\?)?`),
			uml.Leaf(1, "ARROW_BODYA1", "(-+)"),
			uml.Leaf(1, "ARROW_STYLE1", colorOrStylePattern),
			uml.Leaf(1, "ARROW_BODYB1", "(-*)"),
			uml.Leaf(1, "ARROW_DRESSING1", `(>>?|//?|\\\\?)`)),
		regex.NewConcat(
			uml.Leaf(1, "ARROW_DRESSING2", `(<<?|//?|\\\\?)`),
			uml.Leaf(1, "ARROW_BODYB2", "(-*)"),
			uml.Leaf(1, "ARROW_STYLE2", colorOrStylePattern),
			uml.Leaf(1, "ARROW_BODYA2", "(-+)")))
}

// commandExoArrowLeft is "[-> A" and "?-> A".
func commandExoArrowLeft() cmd {
	parts := []regex.Part{regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		uml.Leaf(2, "ANCHOR", anchorPattern),
		uml.Leaf(1, arrowSuppCircle2, `([?\[\]][ox]?)?`),
		exoArrowBody(),
		uml.Leaf(1, arrowSuppCircle1, "([ox][%s]+)?"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, "PARTICIPANT", "([%pLN_.@]+|[%g][^%g]+[%g])"),
	}
	return uml.Single(regex.NewConcat(append(parts, exoTail()...)...),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeExoArrow(d, location, arg, exoTypeLeft(arg))
		})
}

// commandExoArrowRight is "A ->]" and "A ->?".
func commandExoArrowRight() cmd {
	parts := []regex.Part{regex.Start(),
		uml.Leaf(1, "PARALLEL", "(&[%s]*)?"),
		uml.Leaf(2, "ANCHOR", anchorPattern),
		uml.Leaf(1, "PARTICIPANT", "([%pLN_.@]+|[%g][^%g]+[%g])"),
		regex.SpaceZeroOrMore(),
		uml.Leaf(1, arrowSuppCircle1, "([%s]+[ox])?"),
		exoArrowBody(),
		uml.Leaf(1, arrowSuppCircle2, `([ox]?[?\]\[])?`),
	}
	return uml.Single(regex.NewConcat(append(parts, exoTail()...)...),
		func(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result, _ command.ParserPass) command.Result {
			return executeExoArrow(d, location, arg, exoTypeRight(arg))
		})
}

// exoTypeLeft is CommandExoArrowLeft.getMessageExoType.
func exoTypeLeft(arg *regex.Result) MessageExoType {
	start, hasStart := arg.Get(arrowSuppCircle2, 0)
	_, d1 := arg.Get("ARROW_DRESSING1", 0)
	_, d2 := arg.Get("ARROW_DRESSING2", 0)
	if hasStart && strings.Contains(start, "]") {
		if d1 {
			return ExoFromRight
		}
		if d2 {
			return ExoToRight
		}
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if d1 {
		return ExoFromLeft
	}
	if d2 {
		return ExoToLeft
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return 0
}

// exoTypeRight is CommandExoArrowRight.getMessageExoType.
func exoTypeRight(arg *regex.Result) MessageExoType {
	start, hasStart := arg.Get(arrowSuppCircle2, 0)
	_, d1 := arg.Get("ARROW_DRESSING1", 0)
	_, d2 := arg.Get("ARROW_DRESSING2", 0)
	if hasStart && strings.Contains(start, "[") {
		if d1 {
			return ExoToLeft
		}
		if d2 {
			return ExoFromLeft
		}
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if d1 {
		return ExoToRight
	}
	if d2 {
		return ExoFromRight
	}
	uml.Throw(uml.IllegalArgumentException, "")
	return 0
}

// exoArrowPart is CommandExoArrowAny.getArrowPart.
func exoArrowPart(dressing string, typ MessageExoType) ArrowPart {
	if strings.Contains(dressing, "/") {
		if typ.Direction() == 1 {
			return ArrowPartBottomPart
		}
		return ArrowPartTopPart
	}
	if strings.Contains(dressing, "\\") {
		if typ.Direction() == 1 {
			return ArrowPartTopPart
		}
		return ArrowPartBottomPart
	}
	return ArrowPartFull
}

func containsSymbol(arg *regex.Result, suppCircle, symbol string) bool {
	s, ok := arg.Get(suppCircle, 0)
	return ok && strings.Contains(s, symbol)
}

// javaString is Java's string conversion of a nullable string.
func javaString(s string, ok bool) string {
	if !ok {
		return "null"
	}
	return s
}

// executeExoArrow is CommandExoArrowAny.executeArg.
func executeExoArrow(d *SequenceDiagram, location *text.LineLocation, arg *regex.Result,
	messageExoType MessageExoType) command.Result {
	body := javaString(arg.Lazy("ARROW_BODYA", 0)) + javaString(arg.Lazy("ARROW_BODYB", 0))
	dressingValue, hasDressing := arg.Lazy("ARROW_DRESSING", 0)
	if !hasDressing {
		uml.ThrowNPE(`Cannot invoke "String.length()" because "dressing" is null`)
	}
	p := d.OrCreateParticipant(location, uml.RemoveQuotes(arg.Str("PARTICIPANT", 0)))

	sync := uml.JavaLength(dressingValue) == 2
	dotted := strings.Contains(body, "--")

	var labels *klimt.Display
	if label, ok := arg.Get("LABEL", 0); ok {
		labels = klimt.DisplayWithNewlines(d.Pragma(), label)
	} else {
		labels = klimt.DisplayCreate("")
	}
	config := ArrowConfigurationNormal()
	if _, both := arg.Get("ARROW_BOTHDRESSING", 0); both {
		config = ArrowConfigurationBoth()
	}
	if dotted {
		config = config.WithBody(ArrowBodyDotted)
	}
	if sync {
		config = config.WithHead(ArrowHeadAsync)
	}
	config = config.WithPart(exoArrowPart(dressingValue, messageExoType))
	style, hasStyle := arg.Lazy("ARROW_STYLE", 0)
	config, err := applyStyle(style, hasStyle, config)
	if err != nil {
		return colorResult(err)
	}

	activationSpec, hasActivation := arg.Get("ACTIVATION", 0)
	if hasActivation && activationSpec[0] == '*' {
		d.Activate(p, LifeEventCreate, nil, nil)
	}

	if messageExoType == ExoToRight || messageExoType == ExoToLeft {
		if containsSymbol(arg, arrowSuppCircle1, "o") {
			config = config.WithDecoration1(ArrowDecorationCircle)
		}
		if containsSymbol(arg, arrowSuppCircle1, "x") {
			config = config.WithHead1(ArrowHeadCrossX)
		}
		if containsSymbol(arg, arrowSuppCircle2, "o") {
			config = config.WithDecoration2(ArrowDecorationCircle)
		}
		if containsSymbol(arg, arrowSuppCircle2, "x") {
			config = config.WithHead2(ArrowHeadCrossX)
		}
	} else {
		if containsSymbol(arg, arrowSuppCircle2, "o") {
			config = config.WithDecoration1(ArrowDecorationCircle)
		}
		if containsSymbol(arg, arrowSuppCircle2, "x") {
			config = config.WithHead1(ArrowHeadCrossX)
		}
		if containsSymbol(arg, arrowSuppCircle1, "o") {
			config = config.WithDecoration2(ArrowDecorationCircle)
		}
		if containsSymbol(arg, arrowSuppCircle1, "x") {
			config = config.WithHead2(ArrowHeadCrossX)
		}
	}

	number, hasNumber := d.NextMessageNumber()
	msg := NewMessageExo(d.CurrentStyleBuilder(), p, messageExoType, labels, config, number, hasNumber,
		containsSymbol(arg, arrowSuppCircle2, "?"), location)
	if url, ok := arg.Get(klimt.URLBuilderKey, 0); ok {
		msg.SetURL(urlBuilder(d).URL(url))
	}
	if _, ok := arg.Get("PARALLEL", 0); ok {
		msg.GoParallel()
	}
	msg.SetAnchor(arg.Get("ANCHOR", 1))
	msg.SetPart1Anchor(arg.Get("PART1ANCHOR", 1))
	msg.SetPart2Anchor(arg.Get("PART2ANCHOR", 1))

	if status := d.AddMessage(msg); !status.IsOK() {
		return status
	}
	activationColor, err := lifeColor(arg)
	if err != nil {
		return colorResult(err)
	}
	if hasActivation {
		switch activationSpec[0] {
		case '+':
			d.Activate(p, LifeEventActivate, activationColor, nil)
		case '-':
			d.Activate(p, LifeEventDeactivate, nil, nil)
		case '!':
			d.Activate(p, LifeEventDestroy, nil, nil)
		}
	} else if d.IsAutoactivate() && (config.Head() == ArrowHeadNormal || config.Head() == ArrowHeadAsync) {
		if config.IsDotted() {
			d.Activate(p, LifeEventDeactivate, nil, nil)
		} else {
			d.Activate(p, LifeEventActivate, activationColor, nil)
		}
	}
	return command.Ok()
}
