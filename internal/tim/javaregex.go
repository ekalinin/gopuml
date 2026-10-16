// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dlclark/regexp2"
)

// User regular expressions (%splitstr_regex) have Java's syntax and
// semantics. They run on github.com/dlclark/regexp2, a backtracking engine
// with the .NET syntax and a match timeout (design D4). translateJavaRegex
// parses the Java pattern like java.util.regex.Pattern and writes an
// equivalent .NET pattern:
//
//   - \Q...\E quoting, possessive quantifiers (as atomic groups), \x{h..h},
//     octal \0nnn, \cX, \e, \a and \uXXXX surrogate pairs are rewritten;
//   - \d, \w, \s, \b and the POSIX classes (\p{Alpha}...) are ASCII-only as
//     in Java, unless the U flag is set; \h, \v, \R and the java* classes
//     (\p{javaLowerCase}...) get their Java definition;
//   - "." excludes all Java line terminators (\n, \r, \u0085, \u2028,
//     \u2029), and ^, $ and \Z follow Java's rules for them, with the m, s
//     and d flags;
//   - the i flag folds the case of ASCII letters only, unless the u flag is
//     set; the x flag ignores white space and comments, also inside
//     character classes;
//   - character classes support Java's unions ([a[bc]]) and intersections
//     ([a-z&&[^aeiou]]);
//   - Unicode categories, binary properties and scripts are supported
//     with Java's names (\p{Lu}, \p{IsAlphabetic}, \p{IsLatin},
//     \p{sc=Greek}...);
//   - named groups are numbered in order with the other groups, as in
//     Java, and a reference to a group that does not exist never matches;
//   - Java's syntax errors (a dangling quantifier, "{" that is not a
//     quantifier, an unknown escape letter...) are errors.
//
// Remaining differences: Unicode blocks (\p{InGreek}), four-letter script
// aliases, the Emoji properties, \p{javaMirrored}, \N{name} and \b{g} are
// not supported and are errors; \X matches a character followed by its
// combining marks, not a full grapheme cluster; case folding with the u flag
// and of back references uses .NET's rules; .NET accepts some patterns that
// Java rejects, such as lookbehinds without a maximum length; the CANON_EQ
// flag (c) is ignored; and since Go strings hold no lone surrogate, a
// pattern never matches inside a surrogate pair.

// regexMatchTimeout bounds each match of a user regular expression.
const regexMatchTimeout = 2 * time.Second

// Flags of java.util.regex.Pattern, as inline letters.
const (
	reCaseInsensitive = 1 << iota // i
	reUnixLines                   // d
	reMultiline                   // m
	reDotAll                      // s
	reUnicodeCase                 // u
	reComments                    // x
	reUnicodeClass                // U
)

// errRegexTimeout reports a match that ran longer than regexMatchTimeout.
var errRegexTimeout = errors.New("regular expression timeout")

// compileJavaRegex compiles a Java regular expression; the error is the
// PatternSyntaxException message.
func compileJavaRegex(pattern string) (*regexp2.Regexp, error) {
	translated, err := translateJavaRegex(pattern)
	if err != nil {
		return nil, err
	}
	re, err := regexp2.Compile(translated, regexp2.None)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = regexMatchTimeout
	return re, nil
}

// javaSplit is String.split(regex) (Pattern.split with limit 0): the
// pieces between the matches, without the trailing empty pieces and
// without a leading empty piece for a zero-width match at the start.
func javaSplit(s string, re *regexp2.Regexp) ([]string, error) {
	input := []rune(s)
	var list []string
	index := 0
	m, err := re.FindRunesMatch(input)
	for ; err == nil && m != nil; m, err = re.FindNextMatch(m) {
		if index == 0 && m.Index == 0 && m.Length == 0 {
			// no empty leading substring included for zero-width match
			// at the beginning of the input char sequence.
			continue
		}
		list = append(list, string(input[index:m.Index]))
		index = m.Index + m.Length
	}
	if err != nil {
		return nil, errRegexTimeout
	}
	// If no match was found, return this
	if index == 0 {
		return []string{s}, nil
	}
	list = append(list, string(input[index:]))
	for len(list) > 0 && list[len(list)-1] == "" {
		list = list[:len(list)-1]
	}
	return list, nil
}

// javaRegexError is a PatternSyntaxException.
type javaRegexError struct {
	desc string
}

func (e *javaRegexError) Error() string { return e.desc }

// translateJavaRegex returns the .NET form of a Java regular expression.
func translateJavaRegex(pattern string) (string, error) {
	src := removeQEQuoting([]rune(pattern))
	// A first pass counts the capturing groups, which back references need.
	first := &javaRegexTranslator{src: src}
	if _, err := first.translate(); err != nil {
		return "", err
	}
	second := &javaRegexTranslator{src: src, totalGroups: first.groups}
	return second.translate()
}

// removeQEQuoting is Pattern.RemoveQEQuoting: the text between \Q and \E
// (or the end of the pattern) becomes literal, with its ASCII characters
// other than letters escaped. Digits are written \x3N so that they cannot
// extend an escape before them.
func removeQEQuoting(src []rune) []rune {
	var out []rune
	for i := 0; i < len(src); i++ {
		c := src[i]
		if c != '\\' || i+1 >= len(src) {
			out = append(out, c)
			continue
		}
		if src[i+1] != 'Q' {
			out = append(out, c, src[i+1])
			i++
			continue
		}
		for i += 2; i < len(src); i++ {
			c := src[i]
			if c == '\\' && i+1 < len(src) && src[i+1] == 'E' {
				i++
				break
			}
			switch {
			case c >= 0x80 || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
				out = append(out, c)
			case c >= '0' && c <= '9':
				out = append(out, '\\', 'x', '3', c)
			default:
				out = append(out, '\\', c)
			}
		}
	}
	return out
}

// javaRegexTranslator parses a Java pattern (after removeQEQuoting) and
// writes the .NET pattern.
type javaRegexTranslator struct {
	src         []rune
	pos         int
	flags       int
	groups      int // capturing groups opened so far
	totalGroups int // capturing groups of the whole pattern
	named       map[string]int
	out         []rune
}

func (t *javaRegexTranslator) translate() (s string, err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(*javaRegexError)
			if !ok {
				panic(r)
			}
			s, err = "", e
		}
	}()
	t.named = map[string]int{}
	t.expr()
	if t.pos < len(t.src) {
		if t.src[t.pos] == ')' {
			t.fail("Unmatched closing ')'")
		}
		t.fail("Unexpected internal error")
	}
	return string(t.out), nil
}

func (t *javaRegexTranslator) fail(desc string) {
	panic(&javaRegexError{desc: fmt.Sprintf("%s near index %d", desc, t.pos)})
}

func (t *javaRegexTranslator) has(flag int) bool { return t.flags&flag != 0 }

func (t *javaRegexTranslator) emit(s string) { t.out = append(t.out, []rune(s)...) }

// peek returns the next character, -1 at the end; with the x flag it first
// skips white space and comments.
func (t *javaRegexTranslator) peek() rune {
	if t.has(reComments) {
		t.skipWhitespace()
	}
	if t.pos >= len(t.src) {
		return -1
	}
	return t.src[t.pos]
}

// next consumes the next character and returns the following one, as
// Pattern.next.
func (t *javaRegexTranslator) next() rune {
	t.pos++
	return t.peek()
}

// read consumes and returns the next character.
func (t *javaRegexTranslator) read() rune {
	ch := t.peek()
	if ch >= 0 {
		t.pos++
	}
	return ch
}

// raw returns the character at pos+offset without skipping white space,
// -1 outside the pattern.
func (t *javaRegexTranslator) raw(offset int) rune {
	if i := t.pos + offset; i >= 0 && i < len(t.src) {
		return t.src[i]
	}
	return -1
}

func (t *javaRegexTranslator) skipWhitespace() {
	for t.pos < len(t.src) {
		c := t.src[t.pos]
		switch {
		case c == ' ' || (c >= '\t' && c <= '\r'):
			t.pos++
		case c == '#':
			for t.pos < len(t.src) && !t.isLineSeparator(t.src[t.pos]) {
				t.pos++
			}
		default:
			return
		}
	}
}

func (t *javaRegexTranslator) isLineSeparator(c rune) bool {
	if t.has(reUnixLines) {
		return c == '\n'
	}
	return c == '\n' || c == '\r' || c == 0x85 || c == 0x2028 || c == 0x2029
}

// expr is Pattern.expr: alternatives separated by '|'.
func (t *javaRegexTranslator) expr() {
	for {
		t.sequence()
		if t.peek() != '|' {
			return
		}
		t.pos++
		t.emit("|")
	}
}

// sequence is Pattern.sequence: atoms with their quantifiers, up to '|',
// ')' or the end.
func (t *javaRegexTranslator) sequence() {
	for {
		ch := t.peek()
		start := len(t.out)
		switch ch {
		case -1, '|', ')':
			return
		case '(':
			if !t.group() {
				// an inline modifier only: no atom
				continue
			}
		case '[':
			t.pos++
			t.emit(t.clazz(true).String())
		case '\\':
			t.escapeOutsideClass()
		case '^':
			t.pos++
			t.emitCaret()
		case '$':
			t.pos++
			t.emitDollar(t.has(reMultiline))
		case '.':
			t.pos++
			switch {
			case t.has(reDotAll):
				t.emit(`[\s\S]`)
			case t.has(reUnixLines):
				t.emit(`[^\n]`)
			default:
				t.emit(`[^\n\r\u0085\u2028\u2029]`)
			}
		case '*', '+', '?':
			t.fail("Dangling meta character '" + string(ch) + "'")
		case '{':
			// as in Pattern.atom, an empty atom that the quantifier applies to
			t.emit("(?:)")
		default:
			t.pos++
			t.emitLiteral(ch)
		}
		t.closure(start)
	}
}

// closure is Pattern.closure: the quantifier after the atom written from
// start, if any, with its lazy (?) or possessive (+) form.
func (t *javaRegexTranslator) closure(start int) {
	var q string
	switch ch := t.peek(); ch {
	case '?', '*', '+':
		t.pos++
		q = string(ch)
	case '{':
		t.pos++
		q = t.curly()
	default:
		return
	}
	switch t.peek() {
	case '?':
		t.pos++
		t.emit(q + "?")
	case '+':
		t.pos++
		atom := string(t.out[start:])
		t.out = append(t.out[:start], []rune("(?>"+atom+q+")")...)
	default:
		t.emit(q)
	}
}

// curly parses "{n}", "{n,}" or "{n,m}" after the '{'.
func (t *javaRegexTranslator) curly() string {
	readInt := func() int {
		n := 0
		for t.pos < len(t.src) && t.src[t.pos] >= '0' && t.src[t.pos] <= '9' {
			n = n*10 + int(t.src[t.pos]-'0')
			if n > 1<<31-1 {
				t.fail("Illegal repetition range")
			}
			t.pos++
		}
		return n
	}
	if c := t.raw(0); c < '0' || c > '9' {
		t.fail("Illegal repetition")
	}
	cmin := readInt()
	cmax := cmin
	q := "{" + strconv.Itoa(cmin)
	if t.raw(0) == ',' {
		t.pos++
		q += ","
		if c := t.raw(0); c >= '0' && c <= '9' {
			cmax = readInt()
			q += strconv.Itoa(cmax)
		} else {
			cmax = 1<<31 - 1
		}
	}
	if t.raw(0) != '}' {
		t.fail("Unclosed counted closure")
	}
	t.pos++
	if cmax < cmin {
		t.fail("Illegal repetition range")
	}
	return q + "}"
}

// group parses a group after '('; it returns false for an inline modifier
// only, such as "(?i)".
func (t *javaRegexTranslator) group() bool {
	save := t.flags
	t.pos++
	if t.raw(0) != '?' {
		t.groups++
		t.emit("(")
		t.groupBody(save)
		return true
	}
	t.pos++
	switch ch := t.raw(0); ch {
	case ':', '=', '!', '>':
		t.pos++
		t.emit("(?" + string(ch))
	case '<':
		t.pos++
		switch c := t.raw(0); {
		case c == '=' || c == '!':
			t.pos++
			t.emit("(?<" + string(c))
		case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
			start := t.pos
			for c := t.raw(0); c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'; c = t.raw(0) {
				t.pos++
			}
			name := string(t.src[start:t.pos])
			if t.raw(0) != '>' {
				t.fail("named capturing group is missing trailing '>'")
			}
			t.pos++
			if _, ok := t.named[name]; ok {
				t.fail("Named capturing group <" + name + "> is already defined")
			}
			t.groups++
			t.named[name] = t.groups
			t.emit("(")
		default:
			t.fail("capturing group name does not start with a Latin letter")
		}
	default:
		t.addFlags()
		switch t.read() {
		case ')':
			// the flags apply to the rest of the enclosing group
			return false
		case ':':
			t.emit("(?:")
		default:
			t.fail("Unknown inline modifier")
		}
	}
	t.groupBody(save)
	return true
}

// groupBody parses the expression of a group and its ')', then restores
// the flags of the enclosing group.
func (t *javaRegexTranslator) groupBody(save int) {
	t.expr()
	if t.peek() != ')' {
		t.fail("Unclosed group")
	}
	t.pos++
	t.emit(")")
	t.flags = save
}

// addFlags parses the letters of an inline modifier, such as "i-s".
func (t *javaRegexTranslator) addFlags() {
	on := true
	for {
		var flag int
		switch t.raw(0) {
		case 'i':
			flag = reCaseInsensitive
		case 'd':
			flag = reUnixLines
		case 'm':
			flag = reMultiline
		case 's':
			flag = reDotAll
		case 'u':
			flag = reUnicodeCase
		case 'x':
			flag = reComments
		case 'U':
			flag = reUnicodeClass | reUnicodeCase
		case 'c':
			// CANON_EQ is not supported
		case '-':
			if !on {
				return
			}
			on = false
		default:
			return
		}
		t.pos++
		if on {
			t.flags |= flag
		} else {
			t.flags &^= flag
		}
	}
}

// emitCaret writes the .NET form of '^'.
func (t *javaRegexTranslator) emitCaret() {
	switch {
	case !t.has(reMultiline):
		t.emit(`(?:\A)`)
	case t.has(reUnixLines):
		t.emit(`(?:(?:\A|(?<=\n))(?=[\s\S]))`)
	default:
		// not between \r and \n, and not at the end of the input
		t.emit(`(?:(?:\A|(?<=[\n\u0085\u2028\u2029])|(?<=\r)(?!\n))(?=[\s\S]))`)
	}
}

// emitDollar writes the .NET form of '$' (and of \Z when multiline is
// false).
func (t *javaRegexTranslator) emitDollar(multiline bool) {
	switch {
	case t.has(reUnixLines) && multiline:
		t.emit(`(?=\n|\z)`)
	case t.has(reUnixLines):
		t.emit(`(?=\n?\z)`)
	case multiline:
		t.emit(`(?:(?=[\n\r\u0085\u2028\u2029]|\z)(?!(?<=\r)\n))`)
	default:
		t.emit(`(?:(?=(?:\r\n|[\n\r\u0085\u2028\u2029])?\z)(?!(?<=\r)\n))`)
	}
}

// emitLiteral writes a character that matches itself, with the case
// folding of the i flag.
func (t *javaRegexTranslator) emitLiteral(c rune) {
	if t.has(reCaseInsensitive) {
		if t.has(reUnicodeCase) {
			if unicode.ToLower(c) != c || unicode.ToUpper(c) != c {
				t.emit("(?i:" + escapeRegexChar(c) + ")")
				return
			}
		} else if other := asciiOtherCase(c); other != c {
			t.emit("[" + escapeRegexChar(c) + escapeRegexChar(other) + "]")
			return
		}
	}
	t.emit(escapeRegexChar(c))
}

// asciiOtherCase returns the other case of an ASCII letter, else c.
func asciiOtherCase(c rune) rune {
	switch {
	case c >= 'a' && c <= 'z':
		return c - 'a' + 'A'
	case c >= 'A' && c <= 'Z':
		return c - 'A' + 'a'
	}
	return c
}

// escapeRegexChar writes a literal character, inside or outside a class:
// ASCII punctuation is escaped, control characters and surrogates become
// \uXXXX.
func escapeRegexChar(c rune) string {
	switch {
	case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_':
		// .NET rejects the escape of a word character
		return string(c)
	case c > ' ' && c < 0x7f:
		return `\` + string(c)
	case c < 0x80 || unicode.Is(unicode.Cs, c):
		return fmt.Sprintf(`\u%04X`, c)
	}
	return string(c)
}

// escapeOutsideClass parses an escape after '\' outside a class.
func (t *javaRegexTranslator) escapeOutsideClass() {
	t.pos++
	ch := t.raw(0)
	switch ch {
	case 'p', 'P':
		t.pos++
		t.emit(t.family(ch == 'P').String())
		return
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		t.pos++
		t.emitBackReference(t.ref(int(ch - '0')))
		return
	case 'k':
		t.pos++
		if t.raw(0) != '<' {
			t.fail(`\k is not followed by '<' for named capturing group`)
		}
		t.pos++
		start := t.pos
		for c := t.raw(0); c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'; c = t.raw(0) {
			t.pos++
		}
		name := string(t.src[start:t.pos])
		if t.raw(0) != '>' {
			t.fail("named capturing group is missing trailing '>'")
		}
		t.pos++
		n, ok := t.named[name]
		if !ok {
			t.fail("named capturing group <" + name + "> does not exist")
		}
		t.emitBackReference(n)
		return
	case 'A':
		t.pos++
		t.emit(`(?:\A)`)
		return
	case 'G':
		t.pos++
		t.emit(`(?:\G)`)
		return
	case 'z':
		t.pos++
		t.emit(`(?:\z)`)
		return
	case 'Z':
		t.pos++
		t.emitDollar(false)
		return
	case 'b', 'B':
		t.pos++
		if ch == 'b' && t.raw(0) == '{' {
			t.fail("Unsupported grapheme boundary \\b{g}")
		}
		w := t.wordClass().String()
		boundary := `(?:(?<=` + w + `)(?!` + w + `)|(?<!` + w + `)(?=` + w + `))`
		if ch == 'B' {
			boundary = `(?!` + boundary + `)`
		}
		t.emit(boundary)
		return
	case 'R':
		t.pos++
		t.emit(`(?>\r\n|[\n\u000B\f\r\u0085\u2028\u2029])`)
		return
	case 'X':
		t.pos++
		t.emit(`(?>\P{M}\p{M}*)`)
		return
	}
	c, class := t.escape(false)
	if class != nil {
		t.emit(class.String())
		return
	}
	t.emitLiteral(c)
}

// ref is Pattern.ref: more digits extend the group number while the group
// exists.
func (t *javaRegexTranslator) ref(n int) int {
	for {
		c := t.peek()
		if c < '0' || c > '9' {
			return n
		}
		next := n*10 + int(c-'0')
		if t.groups < next {
			return n
		}
		n = next
		t.pos++
	}
}

// emitBackReference writes a reference to group n, which never matches
// when the group does not exist.
func (t *javaRegexTranslator) emitBackReference(n int) {
	if n > t.totalGroups {
		t.emit(`(?!)`)
		return
	}
	if t.has(reCaseInsensitive) {
		t.emit(`(?i:\` + strconv.Itoa(n) + `)`)
		return
	}
	t.emit(`(?:\` + strconv.Itoa(n) + `)`)
}

// escape is Pattern.escape for the escapes that are characters or
// predicates: it returns the character, or a class for a predicate. A
// letter that is not an escape is an error; any other character is
// itself.
func (t *javaRegexTranslator) escape(isRange bool) (rune, *charClass) {
	ch := t.raw(0)
	if ch < 0 {
		t.fail("Unexpected internal error")
	}
	t.pos++
	switch ch {
	case '0':
		return t.octal(), nil
	case 'a':
		return 0x07, nil
	case 'e':
		return 0x1b, nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'c':
		if t.pos >= len(t.src) {
			t.fail("Illegal control escape sequence")
		}
		c := t.src[t.pos] ^ 64
		t.pos++
		return c, nil
	case 'u':
		return t.unicodeEscape(), nil
	case 'x':
		return t.hexEscape(), nil
	case 'd', 'D':
		c := simpleClass("0-9")
		if t.has(reUnicodeClass) {
			c = simpleClass(`\p{Nd}`)
		}
		return 0, negateIf(c, ch == 'D')
	case 'w', 'W':
		return 0, negateIf(t.wordClass(), ch == 'W')
	case 's', 'S':
		c := simpleClass(`\t\n\u000B\f\r\u0020`)
		if t.has(reUnicodeClass) {
			c = simpleClass(`\p{White_Space}`)
		}
		return 0, negateIf(c, ch == 'S')
	case 'h', 'H':
		c := simpleClass(`\t\u0020\u00A0\u1680\u180E\u2000-\u200A\u202F\u205F\u3000`)
		return 0, negateIf(c, ch == 'H')
	case 'v', 'V':
		if ch == 'v' && isRange {
			return 0x0b, nil
		}
		c := simpleClass(`\n\u000B\f\r\u0085\u2028\u2029`)
		return 0, negateIf(c, ch == 'V')
	case 'N':
		t.fail("Unsupported named character \\N")
	}
	if ch < 0x80 && unicode.IsLetter(ch) {
		t.fail("Illegal/unsupported escape sequence")
	}
	return ch, nil
}

// wordClass is \w: ASCII, or Unicode with the U flag.
func (t *javaRegexTranslator) wordClass() charClass {
	if t.has(reUnicodeClass) {
		return simpleClass(unicodeWord)
	}
	return simpleClass("a-zA-Z_0-9")
}

const unicodeWord = `\p{L}\p{Nl}\p{Other_Alphabetic}\p{Mn}\p{Me}\p{Mc}\p{Nd}\p{Pc}\p{Join_Control}`

func negateIf(c charClass, negate bool) *charClass {
	if negate {
		c = c.negate()
	}
	return &c
}

// octal is Pattern.o: one to three octal digits after \0 (three only when
// the first is at most 3).
func (t *javaRegexTranslator) octal() rune {
	isOctal := func(c rune) bool { return c >= '0' && c <= '7' }
	n := t.raw(0)
	if !isOctal(n) {
		t.fail("Illegal octal escape sequence")
	}
	t.pos++
	m := t.raw(0)
	if !isOctal(m) {
		return n - '0'
	}
	t.pos++
	o := t.raw(0)
	if isOctal(o) && n <= '3' {
		t.pos++
		return (n-'0')*64 + (m-'0')*8 + (o - '0')
	}
	return (n-'0')*8 + (m - '0')
}

// hexEscape is Pattern.x: \xhh or \x{h...h}.
func (t *javaRegexTranslator) hexEscape() rune {
	if t.raw(0) == '{' {
		t.pos++
		start := t.pos
		for isHexDigit(t.raw(0)) {
			t.pos++
		}
		if t.pos == start || t.raw(0) != '}' {
			t.fail("Unclosed hexadecimal escape sequence")
		}
		v, err := strconv.ParseInt(string(t.src[start:t.pos]), 16, 64)
		if err != nil || v > unicode.MaxRune {
			t.fail("Hexadecimal codepoint is too big")
		}
		t.pos++
		return rune(v)
	}
	if !isHexDigit(t.raw(0)) || !isHexDigit(t.raw(1)) {
		t.fail("Illegal hexadecimal escape sequence")
	}
	v, _ := strconv.ParseInt(string(t.src[t.pos:t.pos+2]), 16, 32)
	t.pos += 2
	return rune(v)
}

// unicodeEscape is Pattern.u: \uXXXX, a surrogate pair of two escapes
// giving one code point.
func (t *javaRegexTranslator) unicodeEscape() rune {
	uxxxx := func() rune {
		for i := range 4 {
			if !isHexDigit(t.raw(i)) {
				t.fail("Illegal Unicode escape sequence")
			}
		}
		v, _ := strconv.ParseInt(string(t.src[t.pos:t.pos+4]), 16, 32)
		t.pos += 4
		return rune(v)
	}
	n := uxxxx()
	if n >= 0xd800 && n <= 0xdbff && t.raw(0) == '\\' && t.raw(1) == 'u' {
		save := t.pos
		t.pos += 2
		ok := true
		for i := range 4 {
			ok = ok && isHexDigit(t.raw(i))
		}
		if ok {
			if n2 := uxxxx(); n2 >= 0xdc00 && n2 <= 0xdfff {
				return (n-0xd800)<<10 + (n2 - 0xdc00) + 0x10000
			}
		}
		t.pos = save
	}
	return n
}

func isHexDigit(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// charClass is a matcher of one character: a .NET class body (simple), or
// any .NET pattern that matches exactly one character.
type charClass struct {
	body string
	neg  bool
	pat  string
}

func simpleClass(body string) charClass { return charClass{body: body} }

func (c charClass) String() string {
	switch {
	case c.pat != "":
		return c.pat
	case c.neg:
		return "[^" + c.body + "]"
	}
	return "[" + c.body + "]"
}

func (c charClass) union(o charClass) charClass {
	if c.pat == "" && o.pat == "" && !c.neg && !o.neg {
		return charClass{body: c.body + o.body}
	}
	return charClass{pat: "(?:" + c.String() + "|" + o.String() + ")"}
}

func (c charClass) negate() charClass {
	if c.pat == "" {
		c.neg = !c.neg
		return c
	}
	return charClass{pat: `(?:(?!` + c.pat + `)[\s\S])`}
}

func (c charClass) and(o charClass) charClass {
	return charClass{pat: "(?:(?=" + o.String() + ")" + c.String() + ")"}
}

// clazz is Pattern.clazz, after the '[': a class with its unions and
// intersections, up to the ']' (consumed when consume is true).
func (t *javaRegexTranslator) clazz(consume bool) charClass {
	var prev *charClass
	var bits *charClass
	isNeg := false
	ch := t.peek()
	if ch == '^' && t.raw(-1) == '[' {
		ch = t.next()
		isNeg = true
	}
	addBits := func(c charClass) {
		if bits == nil {
			bits = &c
		} else {
			u := bits.union(c)
			bits = &u
		}
	}
	unionPrev := func(c charClass) {
		if prev == nil {
			prev = &c
		} else {
			u := prev.union(c)
			prev = &u
		}
	}
	for {
		switch ch {
		case '[':
			t.pos++
			unionPrev(t.clazz(true))
			ch = t.peek()
			continue
		case '&':
			if t.raw(1) == '&' {
				t.pos += 2
				ch = t.peek()
				var right *charClass
				for ch != ']' && ch != '&' {
					var c charClass
					if ch == '[' {
						t.pos++
						c = t.clazz(true)
					} else {
						c = t.clazz(false)
					}
					if right == nil {
						right = &c
					} else {
						u := right.union(c)
						right = &u
					}
					ch = t.peek()
				}
				if bits != nil {
					// bits used, union has high precedence
					unionPrev(*bits)
					bits = nil
				}
				if prev == nil {
					if right == nil {
						t.fail("Bad class syntax")
					}
					prev = right
				} else if right != nil {
					a := prev.and(*right)
					prev = &a
				}
				continue
			}
		case -1:
			t.fail("Unclosed character class")
		case ']':
			if prev != nil || bits != nil {
				if consume {
					t.pos++
				}
				result := charClass{}
				switch {
				case prev == nil:
					result = *bits
				case bits != nil:
					result = prev.union(*bits)
				default:
					result = *prev
				}
				if isNeg {
					return result.negate()
				}
				return result
			}
		}
		c, isBits := t.rangeItem()
		if isBits {
			addBits(c)
		} else {
			unionPrev(c)
		}
		ch = t.peek()
	}
}

// rangeItem is Pattern.range: a character, a range or a predicate escape.
// isBits is true for characters and ranges, which upstream collects in a
// set that has precedence over intersections.
func (t *javaRegexTranslator) rangeItem() (c charClass, isBits bool) {
	ch := t.peek()
	if ch == '\\' {
		t.pos++
		switch e := t.raw(0); e {
		case 'p', 'P':
			t.pos++
			return t.family(e == 'P'), false
		case 'Q', 'E':
			t.fail("Illegal/unsupported escape sequence")
		}
		if e := t.raw(0); e >= '1' && e <= '9' || strings.ContainsRune("ABGRXZbkz", e) {
			t.fail("Illegal/unsupported escape sequence")
		}
		isRange := t.raw(1) == '-'
		r, class := t.escape(isRange)
		if class != nil {
			return *class, false
		}
		ch = r
	} else {
		t.pos++
	}
	if t.peek() == '-' {
		endRange := t.raw(1)
		if endRange == '[' {
			return t.single(ch), true
		}
		if endRange != ']' {
			t.pos++
			var m rune
			if t.peek() == '\\' {
				t.pos++
				r, class := t.escape(true)
				if class != nil {
					m = -1
				} else {
					m = r
				}
			} else {
				m = t.read()
			}
			if m < ch {
				t.fail("Illegal character range")
			}
			return t.rangeClass(ch, m), true
		}
	}
	return t.single(ch), true
}

// single is a class of one character, with the case folding of the i flag.
func (t *javaRegexTranslator) single(c rune) charClass {
	body := escapeRegexChar(c)
	if t.has(reCaseInsensitive) {
		if t.has(reUnicodeCase) {
			for _, v := range []rune{unicode.ToLower(c), unicode.ToUpper(c), unicode.ToTitle(c)} {
				if v != c {
					body += escapeRegexChar(v)
				}
			}
		} else if other := asciiOtherCase(c); other != c {
			body += escapeRegexChar(other)
		}
	}
	return simpleClass(body)
}

// rangeClass is a class of the characters from lo to hi, with the case
// folding of the i flag.
func (t *javaRegexTranslator) rangeClass(lo, hi rune) charClass {
	body := escapeRegexChar(lo) + "-" + escapeRegexChar(hi)
	if !t.has(reCaseInsensitive) {
		return simpleClass(body)
	}
	if t.has(reUnicodeCase) {
		return charClass{pat: "(?i:[" + body + "])"}
	}
	// the ASCII letters of the range in the other case
	for _, r := range [][2]rune{{'A', 'Z'}, {'a', 'z'}} {
		a, b := max(lo, r[0]), min(hi, r[1])
		if a <= b {
			body += escapeRegexChar(asciiOtherCase(a)) + "-" + escapeRegexChar(asciiOtherCase(b))
		}
	}
	return simpleClass(body)
}

// family is Pattern.family, after \p or \P: a property name of one letter
// or between braces.
func (t *javaRegexTranslator) family(complement bool) charClass {
	var name string
	if t.raw(0) == '{' {
		t.pos++
		start := t.pos
		for t.pos < len(t.src) && t.src[t.pos] != '}' {
			t.pos++
		}
		if t.pos >= len(t.src) {
			t.fail("Unclosed character family")
		}
		name = string(t.src[start:t.pos])
		t.pos++
		if name == "" {
			t.fail("Empty character family")
		}
	} else {
		if t.pos >= len(t.src) {
			t.fail("Illegal/unsupported escape sequence")
		}
		name = string(t.src[t.pos])
		t.pos++
	}
	c, ok := t.property(name)
	if !ok {
		t.fail("Unknown character property name {" + name + "}")
	}
	if complement {
		c = c.negate()
	}
	return c
}

// property returns the class of a property name, as Pattern.family looks
// it up.
func (t *javaRegexTranslator) property(name string) (charClass, bool) {
	if key, value, ok := strings.Cut(name, "="); ok {
		switch strings.ToLower(key) {
		case "sc", "script":
			return scriptClass(value)
		case "gc", "general_category":
			return t.namedProperty(value)
		}
		return charClass{}, false
	}
	if strings.HasPrefix(name, "In") {
		// Unicode blocks are not supported
		return charClass{}, false
	}
	if short, ok := strings.CutPrefix(name, "Is"); ok {
		if c, ok := binaryProperty(short); ok {
			return c, true
		}
		if c, ok := posixUnicodeClass(short); ok {
			return c, true
		}
		if c, ok := t.namedProperty(short); ok {
			return c, true
		}
		return scriptClass(short)
	}
	if t.has(reUnicodeClass) {
		if c, ok := posixUnicodeClass(name); ok {
			return c, true
		}
	}
	return t.namedProperty(name)
}

// namedProperty is CharPredicates.forProperty: a general category, a POSIX
// class (ASCII) or a java.lang.Character class.
func (t *javaRegexTranslator) namedProperty(name string) (charClass, bool) {
	if t.has(reCaseInsensitive) {
		switch name {
		case "Lu", "Ll", "Lt":
			return simpleClass(`\p{Lu}\p{Ll}\p{Lt}`), true
		case "Lower", "Upper":
			return simpleClass("a-zA-Z"), true
		case "javaLowerCase", "javaUpperCase":
			return simpleClass(`\p{Lu}\p{Ll}\p{Lt}\p{Other_Lowercase}\p{Other_Uppercase}`), true
		}
	}
	switch name {
	case "Cn":
		return simpleClass(`\p{C}\p{L}\p{M}\p{N}\p{P}\p{S}\p{Z}`).negate(), true
	case "C":
		// Go's C has no unassigned code points; Java's has
		return simpleClass(`\p{L}\p{M}\p{N}\p{P}\p{S}\p{Z}`).negate(), true
	case "LC":
		return simpleClass(`\p{Lu}\p{Ll}\p{Lt}`), true
	case "LD":
		return simpleClass(`\p{L}\p{Nd}`), true
	case "L1":
		return simpleClass(`\u0000-\u00FF`), true
	case "all":
		return simpleClass(`\s\S`), true
	case "ASCII":
		return simpleClass(`\u0000-\u007F`), true
	case "Alnum":
		return simpleClass("0-9A-Za-z"), true
	case "Alpha":
		return simpleClass("A-Za-z"), true
	case "Blank":
		return simpleClass(`\u0020\t`), true
	case "Cntrl":
		return simpleClass(`\u0000-\u001F\u007F`), true
	case "Digit":
		return simpleClass("0-9"), true
	case "Graph":
		return simpleClass(`\!-\~`), true
	case "Lower":
		return simpleClass("a-z"), true
	case "Print":
		return simpleClass(`\u0020-\~`), true
	case "Punct":
		return simpleClass(`\!-\/\:-\@\[-\` + "`" + `\{-\~`), true
	case "Space":
		return simpleClass(`\t\n\u000B\f\r\u0020`), true
	case "Upper":
		return simpleClass("A-Z"), true
	case "XDigit":
		return simpleClass("0-9a-fA-F"), true
	case "javaLowerCase":
		return simpleClass(`\p{Ll}\p{Other_Lowercase}`), true
	case "javaUpperCase":
		return simpleClass(`\p{Lu}\p{Other_Uppercase}`), true
	case "javaAlphabetic":
		return simpleClass(`\p{L}\p{Nl}\p{Other_Alphabetic}`), true
	case "javaIdeographic":
		return simpleClass(`\p{Ideographic}`), true
	case "javaTitleCase":
		return simpleClass(`\p{Lt}`), true
	case "javaDigit":
		return simpleClass(`\p{Nd}`), true
	case "javaDefined":
		return simpleClass(`\p{C}\p{L}\p{M}\p{N}\p{P}\p{S}\p{Z}`), true
	case "javaLetter":
		return simpleClass(`\p{L}`), true
	case "javaLetterOrDigit":
		return simpleClass(`\p{L}\p{Nd}`), true
	case "javaJavaIdentifierStart":
		return simpleClass(`\p{L}\p{Nl}\p{Sc}\p{Pc}`), true
	case "javaJavaIdentifierPart":
		return simpleClass(`\p{L}\p{Nl}\p{Sc}\p{Pc}\p{Nd}\p{Mn}\p{Mc}` + identifierIgnorable), true
	case "javaUnicodeIdentifierStart":
		return simpleClass(`\p{L}\p{Nl}\p{Other_ID_Start}`), true
	case "javaUnicodeIdentifierPart":
		return simpleClass(`\p{L}\p{Nl}\p{Pc}\p{Nd}\p{Mn}\p{Mc}\p{Other_ID_Start}\p{Other_ID_Continue}` +
			identifierIgnorable), true
	case "javaIdentifierIgnorable":
		return simpleClass(identifierIgnorable), true
	case "javaSpaceChar":
		return simpleClass(`\p{Zs}\p{Zl}\p{Zp}`), true
	case "javaWhitespace":
		return charClass{pat: `(?:(?![\u00A0\u2007\u202F])[\t-\r\u001C-\u001F\p{Zs}\p{Zl}\p{Zp}])`}, true
	case "javaISOControl":
		return simpleClass(`\u0000-\u001F\u007F-\u009F`), true
	}
	if _, ok := unicode.Categories[name]; ok {
		return simpleClass(`\p{` + name + `}`), true
	}
	return charClass{}, false
}

const identifierIgnorable = `\u0000-\u0008\u000E-\u001B\u007F-\u009F\p{Cf}`

// binaryProperty is the binary properties of
// CharPredicates.forUnicodeProperty, without the Emoji properties; the
// POSIX names that it tries next are posixUnicodeClass.
func binaryProperty(name string) (charClass, bool) {
	switch strings.ToUpper(name) {
	case "ALPHABETIC":
		return simpleClass(`\p{L}\p{Nl}\p{Other_Alphabetic}`), true
	case "ASSIGNED":
		return simpleClass(`\p{C}\p{L}\p{M}\p{N}\p{P}\p{S}\p{Z}`), true
	case "CONTROL":
		return simpleClass(`\p{Cc}`), true
	case "HEXDIGIT", "HEX_DIGIT":
		return simpleClass(`\p{Nd}\p{Hex_Digit}`), true
	case "IDEOGRAPHIC":
		return simpleClass(`\p{Ideographic}`), true
	case "JOINCONTROL", "JOIN_CONTROL":
		return simpleClass(`\p{Join_Control}`), true
	case "LETTER":
		return simpleClass(`\p{L}`), true
	case "LOWERCASE":
		return simpleClass(`\p{Ll}\p{Other_Lowercase}`), true
	case "NONCHARACTERCODEPOINT", "NONCHARACTER_CODE_POINT":
		return simpleClass(`\p{Noncharacter_Code_Point}`), true
	case "TITLECASE":
		return simpleClass(`\p{Lt}`), true
	case "PUNCTUATION":
		return simpleClass(`\p{P}`), true
	case "UPPERCASE":
		return simpleClass(`\p{Lu}\p{Other_Uppercase}`), true
	case "WHITESPACE", "WHITE_SPACE":
		return simpleClass(`\p{White_Space}`), true
	case "WORD":
		return simpleClass(unicodeWord), true
	}
	return charClass{}, false
}

// posixUnicodeClass is CharPredicates.forPOSIXName, used with the U flag.
func posixUnicodeClass(name string) (charClass, bool) {
	switch strings.ToUpper(name) {
	case "ALPHA":
		return simpleClass(`\p{L}\p{Nl}\p{Other_Alphabetic}`), true
	case "LOWER":
		return simpleClass(`\p{Ll}\p{Other_Lowercase}`), true
	case "UPPER":
		return simpleClass(`\p{Lu}\p{Other_Uppercase}`), true
	case "SPACE":
		return simpleClass(`\p{White_Space}`), true
	case "PUNCT":
		return simpleClass(`\p{P}`), true
	case "XDIGIT":
		return simpleClass(`\p{Nd}\p{Hex_Digit}`), true
	case "ALNUM":
		return simpleClass(`\p{L}\p{Nl}\p{Other_Alphabetic}\p{Nd}`), true
	case "CNTRL":
		return simpleClass(`\p{Cc}`), true
	case "DIGIT":
		return simpleClass(`\p{Nd}`), true
	case "BLANK":
		return simpleClass(`\p{Zs}\t`), true
	case "GRAPH":
		return simpleClass(`\p{Zs}\p{Zl}\p{Zp}\p{Cc}\p{Cs}`).union(
			simpleClass(`\p{C}\p{L}\p{M}\p{N}\p{P}\p{S}\p{Z}`).negate()).negate(), true
	case "PRINT":
		graph, _ := posixUnicodeClass("GRAPH")
		return graph.union(simpleClass(`\p{Zs}\t`)).and(simpleClass(`\p{Cc}`).negate()), true
	}
	return charClass{}, false
}

// scriptClass is CharPredicates.forUnicodeScript: a script name, ignoring
// case.
func scriptClass(name string) (charClass, bool) {
	for script := range unicode.Scripts {
		if strings.EqualFold(script, name) {
			return simpleClass(`\p{` + script + `}`), true
		}
	}
	return charClass{}, false
}
