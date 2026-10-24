// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/Bodier.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodierAbstract.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodierLikeClassOrObject.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodierMap.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodierJSon.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodierSimple.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cucadiagram/BodyFactory.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Bodier holds the body of an entity: the members of a class or an object,
// the entries of a map, the data of a JSON element or the description lines
// of other elements. Body (upstream's getBody) draws it; it is in
// cucadiagram_body.go.
type Bodier interface {
	// Body is upstream's getBody: the drawing of the body (nil where
	// upstream returns null). fontConfiguration may be nil; maps and JSON
	// elements use it for their entries.
	Body(skinParam klimt.ISkinParam, showMethods, showFields bool, stereotype *klimt.Stereotype,
		style *klimt.Style, fontConfiguration *klimt.FontConfiguration) klimt.TextBlock
	// SetLeaf sets the entity of the body.
	SetLeaf(leaf *Entity)
	// FieldsToDisplay returns the fields, as members.
	FieldsToDisplay() *klimt.Display
	// MethodsToDisplay returns the methods, as members.
	MethodsToDisplay() *klimt.Display
	// AddFieldOrMethod adds a body line; it returns false when the line is
	// not valid for the body.
	AddFieldOrMethod(s string) bool
	// RawBody returns the lines as added.
	RawBody() []klimt.CharSequence
	// MuteClassToObject turns the body of a class into the body of an
	// object.
	MuteClassToObject()
	// HasURL reports whether a member has a link.
	HasURL() bool
	// BestMatch returns the line that best matches a member name; ok is
	// false when there is none.
	BestMatch(candidate string) (string, bool)
}

// VisibilitySet is a set of visibilities, upstream's
// Set<VisibilityModifier>.
type VisibilitySet uint16

// Contains reports whether v is in the set.
func (s VisibilitySet) Contains(v klimt.VisibilityModifier) bool { return s&(1<<v) != 0 }

// ContainsOpt is contains for a visibility that may be null (ok false):
// null is never in the set.
func (s VisibilitySet) ContainsOpt(v klimt.VisibilityModifier, ok bool) bool {
	return ok && s.Contains(v)
}

// With returns the set with v added.
func (s VisibilitySet) With(v klimt.VisibilityModifier) VisibilitySet { return s | 1<<v }

// List returns the visibilities of the set in enum order.
func (s VisibilitySet) List() []klimt.VisibilityModifier {
	var out []klimt.VisibilityModifier
	for v := klimt.VisibilityPrivateField; v <= klimt.VisibilityIEMandatory; v++ {
		if s.Contains(v) {
			out = append(out, v)
		}
	}
	return out
}

// bodierAbstract is upstream's BodierAbstract: the raw lines and the
// entity.
type bodierAbstract struct {
	rawBody []klimt.CharSequence
	leaf    *Entity
}

// SetLeaf implements Bodier; a nil entity panics as upstream's
// requireNonNull.
func (b *bodierAbstract) SetLeaf(leaf *Entity) {
	if leaf == nil {
		uml.ThrowNPE("")
	}
	b.leaf = leaf
}

// RawBody implements Bodier.
func (b *bodierAbstract) RawBody() []klimt.CharSequence { return b.rawBody }

// BestMatch implements Bodier: the line with the lowest match score. An
// empty candidate panics, as upstream throws IllegalArgumentException.
func (b *bodierAbstract) BestMatch(candidate string) (string, bool) {
	if candidate == "" {
		uml.Throw(uml.IllegalArgumentException, "candidate must not be empty")
	}
	best, found := "", false
	bestScore := int64(math.MaxInt64)
	for _, line := range b.rawBody {
		score := matchScore(line.String(), candidate)
		if score < bestScore {
			best, found = line.String(), true
			bestScore = score
			if bestScore == 0 {
				return best, true
			}
		}
	}
	return best, found
}

const (
	weightBeforeMatchStep       = int64(1)
	weightAfterSeparator        = int64(1_000)
	weightTrailingLetters       = int64(1_000_000)
	weightBeforeMatchLetterStep = int64(1000_000_000)
)

// matchScore is upstream's BodierAbstract.matchScore, on UTF-16 units: the
// cost of finding candidate in fullString, MaxInt64 when it is absent.
func matchScore(fullString, candidate string) int64 {
	full := utf16.Encode([]rune(fullString))
	cand := utf16.Encode([]rune(candidate))
	if len(cand) == 0 {
		uml.Throw(uml.IllegalArgumentException, "candidate must not be empty")
	}
	score := int64(0)
	for i := 0; i <= len(full)-len(cand); i++ {
		if startsWithUnits(full, i, cand) {
			separatorSeen := false
			for j := i + len(cand); j < len(full); j++ {
				ch := rune(full[j])
				if !separatorSeen && isAlphanum(ch) {
					score += weightTrailingLetters
				} else {
					separatorSeen = true
					score += weightAfterSeparator
				}
			}
			return score
		}
		if unicode.IsLetter(rune(full[i])) || isOtherAlphabetic(rune(full[i])) {
			score += weightBeforeMatchLetterStep
		} else {
			score += weightBeforeMatchStep
		}
	}
	return math.MaxInt64
}

func startsWithUnits(full []uint16, start int, cand []uint16) bool {
	if start < 0 || start > len(full) || start+len(cand) > len(full) {
		return false
	}
	for i, c := range cand {
		if full[start+i] != c {
			return false
		}
	}
	return true
}

// isAlphanum is Character.isAlphabetic or isDigit, or '_'.
func isAlphanum(c rune) bool {
	return unicode.IsLetter(c) || isOtherAlphabetic(c) || unicode.IsDigit(c) || c == '_'
}

// isOtherAlphabetic completes unicode.IsLetter for Java's
// Character.isAlphabetic: letter numbers and the Other_Alphabetic property.
func isOtherAlphabetic(c rune) bool {
	return unicode.Is(unicode.Nl, c) || unicode.Is(unicode.Other_Alphabetic, c)
}

// BodierLikeClassOrObject is the body of a class-like entity or an object:
// its lines split into fields and methods.
type BodierLikeClassOrObject struct {
	bodierAbstract
	hideVisibilityModifier *VisibilitySet
	typ                    LeafType
	methodsToDisplay       []*Member
	fieldsToDisplay        []*Member
	methodsReady           bool
	fieldsReady            bool
}

// newBodierLikeClassOrObject returns the body of an entity of type typ;
// hideVisibilityModifier is the live set of hidden visibilities of the
// diagram. A map panics, as upstream throws IllegalArgumentException.
func newBodierLikeClassOrObject(typ LeafType, hideVisibilityModifier *VisibilitySet) *BodierLikeClassOrObject {
	if typ == LeafTypeMap {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	if typ == LeafTypeNull {
		uml.ThrowNPE("")
	}
	return &BodierLikeClassOrObject{typ: typ, hideVisibilityModifier: hideVisibilityModifier}
}

// MuteClassToObject implements Bodier.
func (b *BodierLikeClassOrObject) MuteClassToObject() {
	b.methodsReady, b.fieldsReady = false, false
	b.typ = LeafTypeObject
}

// AddFieldOrMethod implements Bodier.
func (b *BodierLikeClassOrObject) AddFieldOrMethod(s string) bool {
	b.methodsReady, b.fieldsReady = false, false
	b.rawBody = append(b.rawBody, klimt.Text(s))
	return true
}

// bodierURLPattern is upstream's BodierLikeClassOrObject.URL_PATTERN:
// UrlBuilder.getRegexp() compiled as a plain java.util.regex pattern, so
// that the %s and %g macros stand for the characters '%', 's' and 'g', and
// the matching is case-sensitive.
var bodierURLPattern = regexp.MustCompile(uml.JavaDots(func() string {
	const startPart = `\[\[[%s]*`
	const endPart = `[%s]*\]\]`
	const javaSpace = `\t\n\x0B\f\r `
	quoted := startPart + `[%g]([^%g]+)[%g]` + `(?:[%s]*\{([^{}]*)\})?` +
		`(?:[%s]([^%s\{\}\[\]][^\[\]]*))?` + endPart
	onlyTooltip := startPart + `\{(.*)\}` + endPart
	onlyTooltipAndLabel := startPart + `\{([^{}]*)\}` + `[%s]*` + `([^\[%s\{\}\[\]][^\[\]]*)` + endPart
	linkTooltipNoLabel := startPart + `([^` + javaSpace + `%g{}\[\]]+?)` + `[%s]*` + `\{(.+)\}` + endPart
	linkWithOptional := startPart + `([^%s%g\[\]]+?)` + `(?:[%s]*\{([^{}]*)\})?` +
		`(?:[%s]([^%s\{\}\[\]][^\[\]]*))?` + endPart
	return quoted + "|" + onlyTooltip + "|" + onlyTooltipAndLabel + "|" + linkTooltipNoLabel + "|" + linkWithOptional
}()))

func bodierIsMethod(s string) bool {
	purged := bodierURLPattern.ReplaceAllString(s, "")
	if strings.Contains(purged, "{method}") {
		return true
	}
	if strings.Contains(purged, "{field}") {
		return false
	}
	return strings.Contains(purged, "(") || strings.Contains(purged, ")")
}

// MethodsToDisplay implements Bodier: the method lines, without the hidden
// visibilities and the final empty members.
func (b *BodierLikeClassOrObject) MethodsToDisplay() *klimt.Display {
	if !b.methodsReady {
		b.methodsToDisplay = nil
		for i, cs := range b.rawBody {
			s := cs.String()
			if !b.isMethodAt(i) {
				continue
			}
			if s == "" && len(b.methodsToDisplay) == 0 {
				continue
			}
			m := MemberMethod(s)
			if b.hideVisibilityModifier == nil || !b.hideVisibilityModifier.ContainsOpt(m.VisibilityModifier()) {
				b.methodsToDisplay = append(b.methodsToDisplay, m)
			}
		}
		b.methodsToDisplay = removeFinalEmptyMembers(b.methodsToDisplay)
		b.methodsReady = true
	}
	return membersDisplay(b.methodsToDisplay)
}

func (b *BodierLikeClassOrObject) isMethodAt(i int) bool {
	if i > 0 && i < len(b.rawBody)-1 && b.rawBody[i].String() == "" && bodierIsMethod(b.rawBody[i-1].String()) &&
		bodierIsMethod(b.rawBody[i+1].String()) {
		return true
	}
	return bodierIsMethod(b.rawBody[i].String())
}

// FieldsToDisplay implements Bodier: the field lines (every line for an
// object), without the hidden visibilities and the final empty members.
func (b *BodierLikeClassOrObject) FieldsToDisplay() *klimt.Display {
	if !b.fieldsReady {
		b.fieldsToDisplay = nil
		for _, cs := range b.rawBody {
			s := cs.String()
			if b.typ != LeafTypeObject && bodierIsMethod(s) {
				continue
			}
			if s == "" && len(b.fieldsToDisplay) == 0 {
				continue
			}
			m := MemberField(s)
			if b.hideVisibilityModifier == nil || !b.hideVisibilityModifier.ContainsOpt(m.VisibilityModifier()) {
				b.fieldsToDisplay = append(b.fieldsToDisplay, m)
			}
		}
		b.fieldsToDisplay = removeFinalEmptyMembers(b.fieldsToDisplay)
		b.fieldsReady = true
	}
	return membersDisplay(b.fieldsToDisplay)
}

func removeFinalEmptyMembers(result []*Member) []*Member {
	for len(result) > 0 && text.Trin(result[len(result)-1].Display(false)) == "" {
		result = result[:len(result)-1]
	}
	return result
}

func membersDisplay(members []*Member) *klimt.Display {
	cs := make([]klimt.CharSequence, len(members))
	for i, m := range members {
		cs[i] = m
	}
	return klimt.DisplayCreateSeq(cs)
}

// HasURL implements Bodier.
func (b *BodierLikeClassOrObject) HasURL() bool {
	for _, cs := range b.FieldsToDisplay().AsList() {
		if m, ok := cs.(*Member); ok && m.HasURL() {
			return true
		}
	}
	for _, cs := range b.MethodsToDisplay().AsList() {
		if m, ok := cs.(*Member); ok && m.HasURL() {
			return true
		}
	}
	return false
}

// RawBodyWithoutHidden returns the lines as members, without the hidden
// visibilities (upstream's rawBodyWithoutHidden, used to draw enhanced
// bodies and objects).
func (b *BodierLikeClassOrObject) RawBodyWithoutHidden() []klimt.CharSequence {
	var result []klimt.CharSequence
	for _, cs := range b.rawBody {
		s := cs.String()
		var m *Member
		if bodierIsMethod(s) {
			m = MemberMethod(s)
		} else {
			m = MemberField(s)
		}
		if !b.hideVisibilityModifier.ContainsOpt(m.VisibilityModifier()) {
			result = append(result, m)
		}
	}
	return result
}

// Type returns the leaf type of the body.
func (b *BodierLikeClassOrObject) Type() LeafType { return b.typ }

// BodierMap is the body of a map: its "key => value" entries in order. A
// "key *--> target" line links the entry to another entity, its value is
// "\x00".
type BodierMap struct {
	bodierAbstract
	keys   []string
	values map[string]string
}

func newBodierMap() *BodierMap { return &BodierMap{values: map[string]string{}} }

// MuteClassToObject implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierMap) MuteClassToObject() { uml.Throw(uml.UnsupportedOperationException, "") }

var bodierMapLinkedEntry = regexp.MustCompile(`(\*-+_?>)`)

// BodierMapLinkedEntry is upstream's getLinkedEntry: the "*-->" arrow of a
// map line; ok is false when there is none.
func BodierMapLinkedEntry(s string) (string, bool) {
	m := bodierMapLinkedEntry.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}
	return m[1], true
}

func (b *BodierMap) put(key, value string) {
	if _, ok := b.values[key]; !ok {
		b.keys = append(b.keys, key)
	}
	b.values[key] = value
}

// AddFieldOrMethod implements Bodier: "key => value" or a linked entry;
// other lines are refused.
func (b *BodierMap) AddFieldOrMethod(s string) bool {
	if x := strings.Index(s, "=>"); x != -1 {
		b.put(text.Trin(s[:x]), text.Trin(s[x+2:]))
		return true
	}
	if link, ok := BodierMapLinkedEntry(s); ok {
		pos := strings.Index(s, link)
		b.put(text.Trin(s[:pos]), "\x00")
		return true
	}
	return false
}

// MethodsToDisplay implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierMap) MethodsToDisplay() *klimt.Display {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// FieldsToDisplay implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierMap) FieldsToDisplay() *klimt.Display {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// HasURL implements Bodier.
func (b *BodierMap) HasURL() bool { return false }

// Entries returns the keys in order and the values by key.
func (b *BodierMap) Entries() (keys []string, values map[string]string) { return b.keys, b.values }

// BodierJSon is the body of a JSON element: its data.
type BodierJSon struct {
	json json.Value
}

// SetLeaf implements Bodier.
func (b *BodierJSon) SetLeaf(*Entity) {}

// MethodsToDisplay implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierJSon) MethodsToDisplay() *klimt.Display {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// FieldsToDisplay implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierJSon) FieldsToDisplay() *klimt.Display {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// MuteClassToObject implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierJSon) MuteClassToObject() { uml.Throw(uml.UnsupportedOperationException, "") }

// HasURL implements Bodier.
func (b *BodierJSon) HasURL() bool { return false }

// RawBody implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierJSon) RawBody() []klimt.CharSequence {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// AddFieldOrMethod implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierJSon) AddFieldOrMethod(string) bool {
	uml.Throw(uml.UnsupportedOperationException, "")
	return false
}

// BestMatch implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierJSon) BestMatch(string) (string, bool) {
	uml.Throw(uml.UnsupportedOperationException, "")
	return "", false
}

// SetJSON sets the data.
func (b *BodierJSon) SetJSON(value json.Value) { b.json = value }

// JSON returns the data, nil for none.
func (b *BodierJSon) JSON() json.Value { return b.json }

// BodierSimple is the body of the other elements and of groups: its lines
// split at the line breaks.
type BodierSimple struct {
	bodierAbstract
	skinParam klimt.ISkinParam
}

// MuteClassToObject implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierSimple) MuteClassToObject() { uml.Throw(uml.UnsupportedOperationException, "") }

// AddFieldOrMethod implements Bodier: the lines of s, split at "\n".
func (b *BodierSimple) AddFieldOrMethod(s string) bool {
	display := klimt.DisplayWithNewlines(b.skinParam.Pragma(), s)
	b.rawBody = append(b.rawBody, display.AsList()...)
	return true
}

// MethodsToDisplay implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierSimple) MethodsToDisplay() *klimt.Display {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// FieldsToDisplay implements Bodier; it panics as upstream throws
// UnsupportedOperationException.
func (b *BodierSimple) FieldsToDisplay() *klimt.Display {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// HasURL implements Bodier.
func (b *BodierSimple) HasURL() bool { return false }

// BodyFactoryCreateLeaf is upstream's BodyFactory.createLeaf: members for
// class-like entities and objects, plain lines otherwise.
func BodyFactoryCreateLeaf(skinParam klimt.ISkinParam, typ LeafType, hideVisibilityModifier *VisibilitySet) Bodier {
	if typ.IsLikeClass() || typ == LeafTypeObject {
		return newBodierLikeClassOrObject(typ, hideVisibilityModifier)
	}
	return &BodierSimple{skinParam: skinParam}
}

// BodyFactoryCreateGroup is upstream's BodyFactory.createGroup.
func BodyFactoryCreateGroup(skinParam klimt.ISkinParam) Bodier {
	return &BodierSimple{skinParam: skinParam}
}
