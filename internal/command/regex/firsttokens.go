// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/regex/FirstTokens.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package regex

import (
	"math"
	"slices"
	"strings"
)

// firstTokensMax is the number of tokens past which a pattern is not
// described token by token.
const firstTokensMax = 64

// FirstTokens reads, from a command pattern in upstream's syntax, the first
// tokens a line must start with for the pattern to match it, as
// text.StringLocated.FirstToken computes them: the leading run of ASCII
// letters, lower-cased; the first character alone for anything else; ""
// for a blank line. It returns nil when they cannot be determined (the
// command may then match any line). The tokens are sorted.
func FirstTokens(pattern string) []string {
	// Commands are matched with find(): a pattern that does not anchor its
	// start constrains the first token not at all.
	if !strings.HasPrefix(pattern, "^") {
		return nil
	}
	r := &firstTokensReader{pattern: []rune(pattern)}
	state := r.alternation(newTokenStateAtLineStart())
	if r.unknown || r.pos != len(r.pattern) {
		return nil
	}
	return state.tokens()
}

// tokenState is what is known once a prefix of the pattern has been read:
// whether the line may still be in its leading blanks, the tokens a letter
// may still extend, and the tokens a non-letter has closed.
type tokenState struct {
	leading bool
	open    map[string]bool
	done    map[string]bool
}

func newTokenStateAtLineStart() *tokenState {
	return &tokenState{leading: true, open: map[string]bool{}, done: map[string]bool{}}
}

func (s *tokenState) copy() *tokenState {
	result := &tokenState{leading: s.leading, open: map[string]bool{}, done: map[string]bool{}}
	for k := range s.open {
		result.open[k] = true
	}
	for k := range s.done {
		result.done[k] = true
	}
	return result
}

// isSettled reports that nothing that follows can change the answer.
func (s *tokenState) isSettled() bool { return !s.leading && len(s.open) == 0 }

func (s *tokenState) letters(l string) {
	extended := map[string]bool{}
	for prefix := range s.open {
		extended[prefix+l] = true
	}
	if s.leading {
		extended[l] = true
	}
	s.open = extended
	s.leading = false
}

func (s *tokenState) blank() {
	for k := range s.open {
		s.done[k] = true
	}
	s.open = map[string]bool{}
}

func (s *tokenState) separator(chars []rune) {
	for k := range s.open {
		s.done[k] = true
	}
	s.open = map[string]bool{}
	if s.leading {
		for _, ch := range chars {
			s.done[string(ch)] = true
		}
	}
	s.leading = false
}

func (s *tokenState) union(other *tokenState) {
	s.leading = s.leading || other.leading
	for k := range other.open {
		s.open[k] = true
	}
	for k := range other.done {
		s.done[k] = true
	}
}

func (s *tokenState) size() int { return len(s.open) + len(s.done) }

func (s *tokenState) containsAllOpen(other *tokenState) bool {
	for k := range other.open {
		if !s.open[k] {
			return false
		}
	}
	return true
}

func (s *tokenState) tokens() []string {
	set := map[string]bool{}
	for k := range s.done {
		set[k] = true
	}
	for k := range s.open {
		set[k] = true
	}
	if s.leading {
		set[""] = true
	}
	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	slices.Sort(result)
	return result
}

func unionStates(a, b *tokenState) *tokenState {
	if a == nil {
		return b
	}
	a.union(b)
	return a
}

type firstTokensReader struct {
	pattern []rune
	pos     int
	unknown bool
}

func (r *firstTokensReader) at(i int) rune { return r.pattern[i] }

func (r *firstTokensReader) startsWith(prefix string, from int) bool {
	p := []rune(prefix)
	return from+len(p) <= len(r.pattern) && string(r.pattern[from:from+len(p)]) == prefix
}

func (r *firstTokensReader) alternation(in *tokenState) *tokenState {
	result := r.sequence(in.copy())
	for !r.unknown && r.pos < len(r.pattern) && r.at(r.pos) == '|' {
		r.pos++
		result.union(r.sequence(in.copy()))
	}
	return result
}

func (r *firstTokensReader) sequence(state *tokenState) *tokenState {
	for !r.unknown && r.pos < len(r.pattern) {
		ch := r.at(r.pos)
		if ch == '|' || ch == ')' {
			break
		}
		if state.isSettled() {
			r.skipElement()
			continue
		}
		state = r.element(state)
		if state.size() > firstTokensMax {
			r.unknown = true
		}
	}
	return state
}

func (r *firstTokensReader) element(state *tokenState) *tokenState {
	ch := r.at(r.pos)
	switch {
	case ch == '^' || ch == '$':
		r.pos++
		return state
	case ch == '(':
		if r.isLookaround() {
			r.skipGroup()
			if !r.unknown {
				r.skipQuantifier()
			}
			return state
		}
		r.pos++
		if r.startsWith("?:", r.pos) {
			r.pos += 2
		} else if r.pos < len(r.pattern) && r.at(r.pos) == '?' {
			r.unknown = true
			return state
		}
		inner := r.alternation(state.copy())
		if r.unknown {
			return state
		}
		if r.pos >= len(r.pattern) || r.at(r.pos) != ')' {
			r.unknown = true
			return state
		}
		r.pos++
		return r.quantified(state, inner)
	case ch == '[':
		end := r.endOfCharClass()
		if end < 0 {
			r.unknown = true
			return state
		}
		charClass := readCharClass(r.pattern[r.pos+1 : end])
		if charClass == nil {
			r.unknown = true
			return state
		}
		r.pos = end + 1
		var after *tokenState
		if charClass.blank {
			after = state.copy()
			after.blank()
		}
		if len(charClass.chars) > 0 {
			asSeparator := state.copy()
			asSeparator.separator(charClass.chars)
			after = unionStates(after, asSeparator)
		}
		for _, letter := range charClass.letters {
			asLetter := state.copy()
			asLetter.letters(string(letter))
			after = unionStates(after, asLetter)
		}
		return r.quantified(state, after)
	case ch == '\\':
		r.pos++
		if r.pos >= len(r.pattern) {
			r.unknown = true
			return state
		}
		escaped := r.at(r.pos)
		r.pos++
		if isASCIILetter(escaped) && escaped != 's' && escaped != 'd' && escaped != 'n' && escaped != 't' &&
			escaped != 'r' && escaped != 'f' {
			r.unknown = true
			return state
		}
		after := state.copy()
		switch {
		case escaped == 'd':
			after.separator(tokenDigits)
		case isASCIILetter(escaped):
			after.blank()
		default:
			after.separator([]rune{escaped})
		}
		return r.quantified(state, after)
	case isASCIILetter(ch):
		r.pos++
		after := state.copy()
		after.letters(string(asciiLower(ch)))
		return r.quantified(state, after)
	case ch == '.' || ch == '%' || ch == '*' || ch == '+' || ch == '?' || ch == '{':
		r.unknown = true
		return state
	}
	r.pos++
	after := state.copy()
	if isTrimmableRune(ch) {
		after.blank()
	} else {
		after.separator([]rune{ch})
	}
	return r.quantified(state, after)
}

func (r *firstTokensReader) isLookaround() bool {
	return r.startsWith("(?=", r.pos) || r.startsWith("(?!", r.pos) || r.startsWith("(?<=", r.pos) ||
		r.startsWith("(?<!", r.pos)
}

func (r *firstTokensReader) quantified(before, after *tokenState) *tokenState {
	if r.pos >= len(r.pattern) {
		return after
	}
	ch := r.at(r.pos)
	minCount, maxCount := 1, 1
	switch ch {
	case '?':
		r.pos++
		minCount = 0
	case '*':
		r.pos++
		minCount, maxCount = 0, math.MaxInt32
	case '+':
		r.pos++
		maxCount = math.MaxInt32
	case '{':
		end := indexRune(r.pattern, '}', r.pos)
		if end < 0 {
			r.unknown = true
			return after
		}
		bounds := string(r.pattern[r.pos+1 : end])
		r.pos = end + 1
		if comma := strings.IndexByte(bounds, ','); comma < 0 {
			minCount = parseCount(bounds)
			maxCount = minCount
		} else {
			minCount = parseCount(bounds[:comma])
			if upper := bounds[comma+1:]; upper == "" {
				maxCount = math.MaxInt32
			} else {
				maxCount = parseCount(upper)
			}
		}
		if minCount < 0 || maxCount < 0 {
			r.unknown = true
			return after
		}
	default:
		return after
	}
	if r.pos < len(r.pattern) && (r.at(r.pos) == '?' || r.at(r.pos) == '+') {
		r.pos++
	}
	if maxCount > 1 && !before.containsAllOpen(after) {
		r.unknown = true
		return after
	}
	if minCount == 0 {
		after.union(before)
	}
	return after
}

func parseCount(s string) int {
	if s == "" {
		return -1
	}
	result := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return -1
		}
		result = result*10 + int(ch-'0')
	}
	return result
}

var tokenDigits = []rune("0123456789")

const (
	tokenQuotes       = "'\u2018\u2019"
	tokenDoubleQuotes = "\"\u201C\u201D" + invisibleQuote
)

// tokenCharClass is a character class read as the characters it holds:
// whether it holds a blank, its ASCII letters (lower-cased) and its other
// characters.
type tokenCharClass struct {
	blank   bool
	letters []rune
	chars   []rune
}

func (c *tokenCharClass) add(ch rune) {
	switch {
	case isASCIILetter(ch):
		if l := asciiLower(ch); !slices.Contains(c.letters, l) {
			c.letters = append(c.letters, l)
		}
	case isTrimmableRune(ch):
		c.blank = true
	default:
		if !slices.Contains(c.chars, ch) {
			c.chars = append(c.chars, ch)
		}
	}
}

// readCharClass reads the content of [...], nil when it cannot be
// enumerated.
func readCharClass(content []rune) *tokenCharClass {
	if len(content) > 0 && content[0] == '^' {
		return nil
	}
	result := &tokenCharClass{}
	i := 0
	for i < len(content) {
		ch := content[i]
		if ch == '%' {
			rest := string(content[i:])
			switch {
			case strings.HasPrefix(rest, "%s"):
				result.blank = true
			case strings.HasPrefix(rest, "%q"):
				for _, c := range tokenQuotes {
					result.chars = appendRune(result.chars, c)
				}
			case strings.HasPrefix(rest, "%g"):
				for _, c := range tokenDoubleQuotes {
					result.chars = appendRune(result.chars, c)
				}
			default:
				return nil
			}
			i += 2
			continue
		}
		if ch == '[' || strings.HasPrefix(string(content[i:]), "&&") {
			return nil
		}
		var first rune
		if ch == '\\' {
			if i+1 >= len(content) {
				return nil
			}
			escaped := content[i+1]
			i += 2
			switch {
			case escaped == 's' || escaped == 'n' || escaped == 't' || escaped == 'r' || escaped == 'f' ||
				escaped == '0':
				result.blank = true
				continue
			case escaped == 'd':
				for _, d := range tokenDigits {
					result.chars = appendRune(result.chars, d)
				}
				continue
			case isASCIILetter(escaped):
				return nil
			}
			first = escaped
		} else {
			first = ch
			i++
		}
		if i+1 < len(content) && content[i] == '-' {
			last := content[i+1]
			i += 2
			if last == '\\' {
				if i >= len(content) {
					return nil
				}
				last = content[i]
				i++
				if isASCIILetter(last) {
					return nil
				}
			}
			if last < first || last-first > firstTokensMax {
				return nil
			}
			for c := first; c <= last; c++ {
				result.add(c)
			}
			continue
		}
		result.add(first)
	}
	if !result.blank && len(result.chars) == 0 && len(result.letters) == 0 {
		return nil
	}
	slices.Sort(result.letters)
	slices.Sort(result.chars)
	return result
}

func appendRune(list []rune, r rune) []rune {
	if slices.Contains(list, r) {
		return list
	}
	return append(list, r)
}

func (r *firstTokensReader) endOfCharClass() int {
	i := r.pos + 1
	if i < len(r.pattern) && r.at(i) == '^' {
		i++
	}
	if i < len(r.pattern) && r.at(i) == ']' {
		i++
	}
	for i < len(r.pattern) {
		switch r.at(i) {
		case '\\':
			i += 2
		case ']':
			return i
		default:
			i++
		}
	}
	return -1
}

func (r *firstTokensReader) skipElement() {
	switch r.at(r.pos) {
	case '(':
		r.skipGroup()
		if r.unknown {
			return
		}
	case '[':
		end := r.endOfCharClass()
		if end < 0 {
			r.unknown = true
			return
		}
		r.pos = end + 1
	case '\\':
		r.pos += 2
	default:
		r.pos++
	}
	if r.pos > len(r.pattern) {
		r.unknown = true
		return
	}
	r.skipQuantifier()
}

func (r *firstTokensReader) skipGroup() {
	depth := 0
	for r.pos < len(r.pattern) {
		ch := r.at(r.pos)
		if ch == '\\' {
			r.pos += 2
			continue
		}
		if ch == '[' {
			end := r.endOfCharClass()
			if end < 0 {
				r.unknown = true
				return
			}
			r.pos = end + 1
			continue
		}
		r.pos++
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return
			}
		}
	}
	r.unknown = true
}

func (r *firstTokensReader) skipQuantifier() {
	if r.pos >= len(r.pattern) {
		return
	}
	switch r.at(r.pos) {
	case '?', '*', '+':
		r.pos++
	case '{':
		end := indexRune(r.pattern, '}', r.pos)
		if end < 0 {
			r.unknown = true
			return
		}
		r.pos = end + 1
	default:
		return
	}
	if r.pos < len(r.pattern) && (r.at(r.pos) == '?' || r.at(r.pos) == '+') {
		r.pos++
	}
}

func indexRune(s []rune, r rune, from int) int {
	for i := from; i < len(s); i++ {
		if s[i] == r {
			return i
		}
	}
	return -1
}

func isASCIILetter(ch rune) bool { return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' }

func asciiLower(ch rune) rune { return ch | 0x20 }

// isTrimmableRune is upstream's StringUtils.isTrimmable.
func isTrimmableRune(ch rune) bool { return ch <= ' ' || ch == '\u00A0' }
