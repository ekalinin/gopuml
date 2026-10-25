// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/Something.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/Subject.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/Verbs.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/Words.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/DayPattern.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/TimeResolution.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/PairOfSomething.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/lang/TwoNames.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ulang/VerbPhraseAction.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/ulang/VerbPhraseMatcher.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/gantt/command/NaturalGanttCommand.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package gantt

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/jtime"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// captureLookup is upstream's CaptureLookup: the named values of a match.
type captureLookup interface {
	FindFirstValueByKey(key string) (string, bool)
	FindValuesByKey(key string) []string
}

// something is upstream's Something: the complement of a verb.
type something interface {
	me(gantt *GanttDiagram, arg captureLookup) Failable[any]
	complementUbrex() klimt.UBrexPart
}

// subject is upstream's Subject: the start of a sentence and the verb
// phrases that may follow it, in the order they are tried.
type subject interface {
	subjectUbrex() klimt.UBrexPart
	me(gantt *GanttDiagram, arg *klimt.UBrexMatcher) Failable[any]
	verbPhrases() []*verbPhraseAction
}

// The words of upstream's Words.
const (
	wordAfter      = "after"
	wordAnd        = "and"
	wordAs         = "as"
	wordAt         = "at"
	wordBefore     = "before"
	wordCompletion = "completion"
	wordEnd        = "end"
	wordFor        = "for"
	wordFrom       = "from"
	wordOn         = "on"
	wordStart      = "start"
	wordThe        = "the"
	wordTo         = "to"
	wordWith       = "with"
	wordThen       = "then"
	wordIt         = "it"
)

func leaves(words []string) []klimt.UBrexPart {
	result := make([]klimt.UBrexPart, len(words))
	for i, w := range words {
		result[i] = klimt.UBrexLeaf(w)
	}
	return result
}

// wordsZeroOrMore is Words.uzeroOrMore: any of the words, each followed by
// spaces, any number of times.
func wordsZeroOrMore(words ...string) klimt.UBrexPart {
	return klimt.UBrexZeroOrMore(klimt.UBrexConcat(klimt.UBrexOr(leaves(words)...), klimt.UBrexSpaceOneOrMore()))
}

// wordsOneOf is Words.uoneOf.
func wordsOneOf(words ...string) klimt.UBrexPart { return klimt.UBrexOr(leaves(words)...) }

// wordsExactly is Words.uexactly: each word after spaces.
func wordsExactly(words ...string) klimt.UBrexPart {
	var parts []klimt.UBrexPart
	for _, w := range words {
		parts = append(parts, klimt.UBrexSpaceOneOrMore(), klimt.UBrexLeaf(w))
	}
	return klimt.UBrexConcat(parts...)
}

// wordsExactly2 is Words.uexactly2: the words separated by spaces.
func wordsExactly2(words ...string) klimt.UBrexPart {
	var parts []klimt.UBrexPart
	for i, w := range words {
		if i > 0 {
			parts = append(parts, klimt.UBrexSpaceOneOrMore())
		}
		parts = append(parts, klimt.UBrexLeaf(w))
	}
	return klimt.UBrexConcat(parts...)
}

// The verbs of upstream's Verbs.
var (
	verbAre                = klimt.UBrexLeaf("are")
	verbDisplayOnSameRowAs = klimt.UBrexLeaf("display〇?s 〇+〴s on 〇+〴s same 〇+〴s row 〇+〴s as")
	verbEnds               = klimt.UBrexLeaf("ends")
	verbHappens            = klimt.UBrexLeaf("happens")
	verbIs                 = klimt.UBrexLeaf("is")
	verbIsColored          = klimt.UBrexLeaf("is 〇+〴s colo〇?ured")
	verbIsDisplayedAs      = klimt.UBrexLeaf("is 〇+〴s displayed 〇+〴s as")
	verbIsOff              = klimt.UBrexLeaf("is 〇+〴s off")
	verbIsOn               = klimt.UBrexLeaf("is 〇+〴s on")
	verbIsOrAre            = klimt.UBrexLeaf("【is┇are】")
	verbIsOrAreNamed       = klimt.UBrexLeaf("【is┇are】〇+〴s named")
	verbJust               = klimt.UBrexLeaf("just")
	verbLinksTo            = klimt.UBrexLeaf("links 〇+〴s to")
	verbOccurs             = klimt.UBrexLeaf("occurs")
	verbPauses             = klimt.UBrexLeaf("pauses")
	verbRequires           = klimt.UBrexLeaf("【lasts┇requires】")
	verbStarts             = klimt.UBrexLeaf("starts")
	verbWorksOn            = klimt.UBrexLeaf("works 〇+〴s on")
)

// dayPattern is upstream's DayPattern: the three forms of an absolute date
// (5 dec 2018, 2018/12/05, dec 5, 2018), with keys suffixed by id.
type dayPattern struct {
	yearA, yearB, yearC, monthA, monthB, monthC, dayA, dayB, dayC string
}

func newDayPattern(id string) dayPattern {
	return dayPattern{"AYEAR" + id, "BYEAR" + id, "CYEAR" + id, "AMONTH" + id, "BMONTH" + id, "CMONTH" + id,
		"ADAY" + id, "BDAY" + id, "CDAY" + id}
}

func (p dayPattern) ubrex() klimt.UBrexPart {
	return klimt.UBrexOr(ubrexDDMonthYYYY(p.yearA, p.monthA, p.dayA), ubrexYYYYMMDD(p.yearB, p.monthB, p.dayB),
		ubrexMonthDDYYYY(p.yearC, p.monthC, p.dayC))
}

// day is getDay: the date, ok false when none of the forms matched. The
// day, the month and the year are parsed in upstream's order.
func (p dayPattern) day(arg captureLookup) (jtime.LocalDate, bool) {
	if d, ok := arg.FindFirstValueByKey(p.dayA); ok {
		day := parseInt(d)
		month := MonthFromString(firstValue(arg, p.monthA))
		year := parseInt(firstValue(arg, p.yearA))
		return jtime.DateOf(year, month.Value(), day), true
	}
	if d, ok := arg.FindFirstValueByKey(p.dayB); ok {
		day := parseInt(d)
		month := parseInt(firstValue(arg, p.monthB))
		year := parseInt(firstValue(arg, p.yearB))
		return jtime.DateOf(year, month, day), true
	}
	if d, ok := arg.FindFirstValueByKey(p.dayC); ok {
		day := parseInt(d)
		month := MonthFromString(firstValue(arg, p.monthC))
		year := parseInt(firstValue(arg, p.yearC))
		return jtime.DateOf(year, month.Value(), day), true
	}
	return jtime.LocalDate{}, false
}

// ubrexDDMonthYYYY is TimeResolution.toUbrexA_DD_MONTH_YYYY.
func ubrexDDMonthYYYY(year, month, day string) klimt.UBrexPart {
	return klimt.UBrexConcat(
		klimt.UBrexNamed(day, klimt.UBrexLeaf("〇{1-2}〴d")),
		klimt.UBrexUpto(klimt.UBrexLeaf("〴D"), klimt.UBrexNamed(month, MonthUbrex())),
		klimt.UBrexLeaf("〇+〴D"),
		klimt.UBrexNamed(year, klimt.UBrexLeaf("〇{1-4}〴d")))
}

// ubrexYYYYMMDD is TimeResolution.toUbrexB_YYYY_MM_DD.
func ubrexYYYYMMDD(year, month, day string) klimt.UBrexPart {
	nonDigit := klimt.UBrexLeaf("〇+〴D")
	return klimt.UBrexConcat(
		klimt.UBrexNamed(year, klimt.UBrexLeaf("〇{1-4}〴d")), nonDigit,
		klimt.UBrexNamed(month, klimt.UBrexLeaf("〇{1-2}〴d")), nonDigit,
		klimt.UBrexNamed(day, klimt.UBrexLeaf("〇{1-2}〴d")))
}

// ubrexMonthDDYYYY is TimeResolution.toUbrexC_MONTH_DD_YYYY.
func ubrexMonthDDYYYY(year, month, day string) klimt.UBrexPart {
	nonDigit := klimt.UBrexLeaf("〇+〴D")
	return klimt.UBrexConcat(
		klimt.UBrexNamed(month, MonthUbrex()), nonDigit,
		klimt.UBrexNamed(day, klimt.UBrexLeaf("〇{1-2}〴d")), nonDigit,
		klimt.UBrexNamed(year, klimt.UBrexLeaf("〇{1-4}〴d")))
}

// pairOfSomething is upstream's PairOfSomething: two complements after one
// another.
type pairOfSomething struct {
	complement1, complement2 something
}

func (p pairOfSomething) me(gantt *GanttDiagram, arg captureLookup) Failable[any] {
	r1 := p.complement1.me(gantt, arg)
	r2 := p.complement2.me(gantt, arg)
	if r1.IsFail() {
		return r1
	}
	if r2.IsFail() {
		return r2
	}
	return FailableOk[any]([2]any{r1.Get(), r2.Get()})
}

func (p pairOfSomething) complementUbrex() klimt.UBrexPart {
	return klimt.UBrexConcat(p.complement1.complementUbrex(), klimt.UBrexSpaceOneOrMore(),
		p.complement2.complementUbrex())
}

// twoNames is upstream's TwoNames.
type twoNames struct {
	name1, name2 string
}

// verbPhraseAction is upstream's VerbPhraseAction: a verb, the words that
// may follow it, a complement, and what the phrase does.
type verbPhraseAction struct {
	verb klimt.UBrexPart
	// ignored are the words skipped after the verb, nil for none.
	ignored    *klimt.UBrexPart
	complement something
	execute    func(gantt *GanttDiagram, subject, complement any) command.Result
	// complementPart caches the expression of the complement.
	complementPart klimt.UBrexPart
}

func phrase(verb klimt.UBrexPart, ignored *klimt.UBrexPart, complement something,
	execute func(gantt *GanttDiagram, subject, complement any) command.Result) *verbPhraseAction {
	return &verbPhraseAction{verb: verb, ignored: ignored, complement: complement, execute: execute,
		complementPart: complement.complementUbrex()}
}

func ignore(part klimt.UBrexPart) *klimt.UBrexPart { return &part }

var spacesPart = klimt.UBrexLeaf("〇+〴s")

// skipSpaces is VerbPhraseAction.skipSpaces: the spaces at the start are
// skipped, and only spaces, not tabs.
func skipSpaces(tn *klimt.UBrexNavigator) { tn.Jump(spacesPart.MatchNavigator(tn).AcceptedLength()) }

// parse is parseVerbPhrase: the verb, the ignored words and the complement
// read from tn, which moves after them. It returns the match of the
// complement.
func (v *verbPhraseAction) parse(tn *klimt.UBrexNavigator) *klimt.UBrexMatcher {
	skipSpaces(tn)
	verbMatch := v.verb.MatchNavigator(tn)
	if verbMatch.AcceptedLength() == 0 {
		uml.Throw(uml.IllegalStateException, "")
	}
	tn.Jump(verbMatch.AcceptedLength())
	skipSpaces(tn)
	if v.ignored != nil {
		tn.Jump(v.ignored.MatchNavigator(tn).AcceptedLength())
		skipSpaces(tn)
	}
	complementMatcher := v.complementPart.MatchNavigator(tn)
	if complementMatcher.AcceptedLength() == 0 {
		uml.Throw(uml.IllegalStateException, "")
	}
	tn.Jump(complementMatcher.AcceptedLength())
	return complementMatcher
}

// check is check: whether the phrase is at the start of tn, which does not
// move.
func (v *verbPhraseAction) check(tn *klimt.UBrexNavigator) bool {
	tn = tn.Copy()
	verbMatch := v.verb.MatchNavigator(tn)
	if verbMatch.AcceptedLength() == 0 {
		return false
	}
	tn.Jump(verbMatch.AcceptedLength())
	skipSpaces(tn)
	if v.ignored != nil {
		tn.Jump(v.ignored.MatchNavigator(tn).AcceptedLength())
		skipSpaces(tn)
	}
	return v.complementPart.MatchNavigator(tn).AcceptedLength() > 0
}

var andPart = klimt.UBrexLeaf("〇+〴s and 〇+〴s")

// naturalGanttCommand is upstream's NaturalGanttCommand: a sentence of a
// subject, on lines read as they are, not trimmed.
type naturalGanttCommand struct {
	subject     subject
	subjectPart klimt.UBrexPart
	phrases     []*verbPhraseAction
}

func newNaturalGanttCommand(s subject) *naturalGanttCommand {
	return &naturalGanttCommand{subject: s, subjectPart: s.subjectUbrex(), phrases: s.verbPhrases()}
}

// IsValid implements command.Command: the subject and one of its verb
// phrases.
func (c *naturalGanttCommand) IsValid(lines command.BlocLines) command.Control {
	tn := klimt.NewUBrexNavigator(lines.At(0).S)
	skipSpaces(tn)
	v1 := c.subjectPart.MatchNavigator(tn).AcceptedLength()
	if v1 == 0 {
		return command.NotOK
	}
	tn.Jump(v1)
	skipSpaces(tn)
	for _, verbPhrase := range c.phrases {
		if verbPhrase.check(tn) {
			return command.OK
		}
	}
	return command.NotOK
}

// Execute implements command.Command: the subject, then the first verb
// phrase that matches, then the next ones after " and ".
func (c *naturalGanttCommand) Execute(gantt *GanttDiagram, lines command.BlocLines, _ command.ParserPass) command.Result {
	tn := klimt.NewUBrexNavigator(lines.At(0).S)
	skipSpaces(tn)
	matcherSubject := c.subjectPart.MatchNavigator(tn)
	tn.Jump(matcherSubject.AcceptedLength())
	skipSpaces(tn)
	candidates := slices.Clone(c.phrases)
	subjectResult := c.subject.me(gantt, matcherSubject)
	if subjectResult.IsFail() {
		return command.Error(subjectResult.Error())
	}
	for len(candidates) > 0 {
		verbPhrase := candidates[0]
		candidates = candidates[1:]
		if !verbPhrase.check(tn) {
			continue
		}
		complementMatcher := verbPhrase.parse(tn)
		complement := verbPhrase.complement.me(gantt, complementMatcher)
		if complement.IsFail() {
			return command.Error(complement.Error())
		}
		result := verbPhrase.execute(gantt, subjectResult.Get(), complement.Get())
		if !result.IsOK() {
			return result
		}
		andLength := andPart.MatchNavigator(tn).AcceptedLength()
		if andLength == 0 {
			return result
		}
		tn.Jump(andLength)
		candidates = slices.Clone(c.phrases)
	}
	return command.Error("No matching verb phrase")
}

// EligibleFor implements command.Command.
func (c *naturalGanttCommand) EligibleFor(pass command.ParserPass) bool { return pass == command.One }

// firstValue is findFirstValueByKey for a key that has a value.
func firstValue(arg captureLookup, key string) string {
	v, _ := arg.FindFirstValueByKey(key)
	return v
}

// parseInt is Integer.parseInt.
func parseInt(s string) int { return uml.JavaParseInt(s) }
