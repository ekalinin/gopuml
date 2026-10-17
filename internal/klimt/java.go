// SPDX-License-Identifier: MIT
// Provenance: original

package klimt

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// Upstream's unchecked exceptions are panics with a *color.RuntimeError
// value, the convention of the klimt packages; checked exceptions are errors.
const (
	unsupportedOperation = "java.lang.UnsupportedOperationException"
	illegalArgument      = "java.lang.IllegalArgumentException"
	illegalState         = "java.lang.IllegalStateException"
	nullPointer          = "java.lang.NullPointerException"
	numberFormat         = "java.lang.NumberFormatException"
	indexOutOfBounds     = "java.lang.IndexOutOfBoundsException"
	noSuchElement        = "java.util.NoSuchElementException"
)

func throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }

// notPorted panics for a code path whose upstream class is not ported yet.
// The message names the missing class, so that the gap is never silent.
func notPorted(upstreamClass, what string) {
	panic(&color.RuntimeError{Class: unsupportedOperation,
		Message: "not ported yet: " + upstreamClass + " (" + what + "); see internal/klimt/HANDOFF.md"})
}

// jstr is Java's Double.toString, used where upstream concatenates a double
// to a string.
func jstr(d float64) string { return jfmt.DoubleToString(d) }

// javaD2I is Java's (int) cast of a double: truncation toward zero, NaN to
// zero and saturation at the int range.
func javaD2I(x float64) int {
	switch {
	case math.IsNaN(x):
		return 0
	case x >= math.MaxInt32:
		return math.MaxInt32
	case x <= math.MinInt32:
		return math.MinInt32
	}
	return int(x)
}

// javaEqualsIgnoreCase is Java's String.equalsIgnoreCase.
func javaEqualsIgnoreCase(a, b string) bool {
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

// javaLower is Java's String.toLowerCase(Locale.ENGLISH), upstream's
// StringUtils.goLowerCase. U+0130 becomes "i" followed by U+0307 as in Java,
// and a capital sigma ending a word becomes the final sigma. Java finds word
// ends with its BreakIterator; here a word is a run of letters, marks and
// digits, an approximation that only matters for Greek text.
func javaLower(s string) string {
	if strings.ContainsRune(s, 'İ') {
		s = strings.ReplaceAll(s, "İ", "i̇")
	}
	if !strings.ContainsRune(s, 'Σ') {
		return strings.ToLower(s)
	}
	runes := []rune(s)
	inWord := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsDigit(r) }
	cased := func(r rune) bool { return unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsTitle(r) }
	var sb strings.Builder
	for i, r := range runes {
		if r != 'Σ' {
			sb.WriteString(strings.ToLower(string(r)))
			continue
		}
		before := false
		for j := i - 1; j >= 0 && inWord(runes[j]); j-- {
			if cased(runes[j]) {
				before = true
				break
			}
		}
		after := false
		for j := i + 1; j < len(runes) && inWord(runes[j]); j++ {
			if cased(runes[j]) {
				after = true
				break
			}
		}
		if before && !after {
			sb.WriteRune('ς')
		} else {
			sb.WriteRune('σ')
		}
	}
	return sb.String()
}

// javaCharLower is Java's Character.toLowerCase(char): characters outside
// the BMP are two UTF-16 units that do not change.
func javaCharLower(r rune) rune {
	if r >= 0x10000 {
		return r
	}
	return unicode.ToLower(r)
}

// javaStringCompare is Java's String.compareTo: UTF-16 code unit order.
func javaStringCompare(a, b string) int {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			return int(ua[i]) - int(ub[i])
		}
	}
	return len(ua) - len(ub)
}

// javaDot is the character class of Java's regex ".": any character but
// the line terminators.
const javaDot = `[^\n\r\x{85}\x{2028}\x{2029}]`

// javaUpper is Java's String.toUpperCase(Locale.ENGLISH), upstream's
// StringUtils.goUpperCase, for "ß" and the simple case mappings. Java also
// applies the other unconditional full mappings of SpecialCasing.txt
// (ligatures such as U+FB01, U+0149, the Greek letters with iota subscript),
// which strings.ToUpper does not; the callers only pass ASCII names.
func javaUpper(s string) string {
	if !strings.ContainsRune(s, 'ß') {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(strings.ReplaceAll(s, "ß", "SS"))
}

// javaLen is Java's String.length(): the number of UTF-16 code units.
func javaLen(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// javaCharAt is Java's String.charAt(i): the UTF-16 code unit at index i.
// It panics as Java throws StringIndexOutOfBoundsException.
func javaCharAt(s string, i int) rune {
	u := utf16.Encode([]rune(s))
	if i < 0 || i >= len(u) {
		throw("java.lang.StringIndexOutOfBoundsException", "Index "+strconv.Itoa(i)+" out of bounds for length "+strconv.Itoa(len(u)))
	}
	return rune(u[i])
}

// javaParseInt is Java's Integer.parseInt(s) for a decimal string: an
// optional sign followed by ASCII digits within the int range. It panics with
// a NumberFormatException otherwise.
func javaParseInt(s string) int {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil || s == "" || strings.HasPrefix(s, "+") && len(s) == 1 {
		throw(numberFormat, "For input string: \""+s+"\"")
	}
	return int(v)
}

// javaParseDouble is Java's Double.parseDouble for the decimal forms used by
// PlantUML (digits, sign, point, exponent, surrounding whitespace). It panics
// with a NumberFormatException when s is not a number.
func javaParseDouble(s string) float64 {
	v, ok := javaTryParseDouble(s)
	if !ok {
		throw(numberFormat, "For input string: \""+s+"\"")
	}
	return v
}

// javaHexFloat is Java's hexadecimal floating-point literal.
var javaHexFloat = regexp.MustCompile(`^[+-]?0[xX]([0-9a-fA-F]+\.?[0-9a-fA-F]*|\.[0-9a-fA-F]+)[pP][+-]?[0-9]+$`)

// javaTryParseDouble is javaParseDouble reporting failure instead of
// panicking. It accepts what Java's Double.parseDouble accepts: surrounding
// characters up to the space are trimmed; "NaN" and "Infinity" with an
// optional sign; decimal and hexadecimal ("0x1.8p1") literals, with an
// optional type suffix (d, D, f, F), which does not change the result.
func javaTryParseDouble(s string) (float64, bool) {
	t := strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
	switch t {
	case "NaN", "+NaN", "-NaN":
		return math.NaN(), true
	case "Infinity", "+Infinity":
		return math.Inf(1), true
	case "-Infinity":
		return math.Inf(-1), true
	}
	body := t
	if body != "" {
		if last := body[len(body)-1]; last == 'd' || last == 'D' || last == 'f' || last == 'F' {
			body = body[:len(body)-1]
		}
	}
	if body == "" {
		return 0, false
	}
	if javaHexFloat.MatchString(body) {
		v, err := strconv.ParseFloat(body, 64)
		if err != nil {
			if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
				return v, true
			}
			return 0, false
		}
		return v, true
	}
	for _, r := range body {
		if !(r >= '0' && r <= '9' || r == '.' || r == '-' || r == '+' || r == 'e' || r == 'E') {
			return 0, false
		}
	}
	v, err := strconv.ParseFloat(body, 64)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			return v, true
		}
		return 0, false
	}
	return v, true
}

// isJavaWhitespaceRegex reports whether r is in Java's regex \s class.
func isJavaWhitespaceRegex(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == 0x0B || r == '\f' || r == '\r'
}

// javaStringHashCode is Java's String.hashCode over the UTF-16 code units.
func javaStringHashCode(s string) int32 {
	var h int32
	for _, u := range utf16.Encode([]rune(s)) {
		h = 31*h + int32(u)
	}
	return h
}

// javaHashMapOrder returns keys, given in insertion order and without
// duplicates, in the iteration order of a java.util.HashMap (or HashSet)
// created with the default capacity and filled by put in that order.
func javaHashMapOrder(keys []string) []string {
	hashes := make([]int32, len(keys))
	for i, k := range keys {
		hashes[i] = javaStringHashCode(k)
	}
	result := make([]string, len(keys))
	for i, idx := range javaHashOrder(hashes) {
		result[i] = keys[idx]
	}
	return result
}

// javaHashOrder returns the indexes of elements, given by their hashCode in
// insertion order and without duplicates, in the iteration order of a
// java.util.HashMap (or HashSet) created with the default capacity. It
// replays HashMap.putVal: an element is appended to bucket
// (h ^ h>>>16) & (capacity - 1); the 16-slot table doubles when the size
// exceeds three quarters of it, and also when a bucket that already holds 8
// elements receives one more while the table has fewer than 64 slots
// (treeifyBin). A resize splits each bucket without reordering it. Buckets
// that HashMap turns into trees (a 9th element in a table of 64 slots or
// more) move their tree root first; that order is not emulated.
func javaHashOrder(hashes []int32) []int {
	spread := func(h32 int32) uint32 {
		h := uint32(h32)
		return h ^ h>>16
	}
	capacity := 16
	buckets := make([][]int, capacity)
	resize := func() {
		capacity *= 2
		next := make([][]int, capacity)
		for _, b := range buckets {
			for _, idx := range b {
				nb := int(spread(hashes[idx]) & uint32(capacity-1))
				next[nb] = append(next[nb], idx)
			}
		}
		buckets = next
	}
	size := 0
	for i, h := range hashes {
		b := int(spread(h) & uint32(capacity-1))
		buckets[b] = append(buckets[b], i)
		if len(buckets[b]) > 8 && capacity < 64 {
			resize()
		}
		size++
		if size > capacity*3/4 {
			resize()
		}
	}
	result := make([]int, 0, len(hashes))
	for _, b := range buckets {
		result = append(result, b...)
	}
	return result
}

// javaTokenize is java.util.StringTokenizer(s, delims) without returning the
// delimiters: the maximal runs of characters that are not delimiters.
func javaTokenize(s, delims string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(delims, r) })
}

// javaTokenizeWithDelims is java.util.StringTokenizer(s, delims, true): the
// runs of non-delimiters and each delimiter as its own token.
func javaTokenizeWithDelims(s, delims string) []string {
	var result []string
	start := -1
	for i, r := range s {
		if strings.ContainsRune(delims, r) {
			if start >= 0 {
				result = append(result, s[start:i])
				start = -1
			}
			result = append(result, string(r))
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		result = append(result, s[start:])
	}
	return result
}

// javaSplit is Java's String.split(regex) for a literal, non-empty
// separator: trailing empty strings are removed.
func javaSplit(s, sep string) []string {
	parts := strings.Split(s, sep)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		// Java returns the input itself when nothing remains and the input
		// was empty, and an empty array when it was only separators.
		if s == "" {
			return []string{""}
		}
	}
	return parts
}

// itoa is Java's int to string conversion.
func itoa(i int) string { return strconv.Itoa(i) }

// javaDoubleHashCode is Java's Double.hashCode.
func javaDoubleHashCode(d float64) int32 {
	bits := uint64(0x7ff8000000000000)
	if !math.IsNaN(d) {
		bits = math.Float64bits(d)
	}
	return int32(bits ^ bits>>32)
}

// javaDoubleCompare is Java's Double.compare: -0.0 is less than 0.0 and NaN
// is greater than every other value and equal to itself.
func javaDoubleCompare(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	bits := func(x float64) int64 {
		if math.IsNaN(x) {
			return 0x7ff8000000000000
		}
		return int64(math.Float64bits(x))
	}
	ba, bb := bits(a), bits(b)
	switch {
	case ba == bb:
		return 0
	case ba < bb:
		return -1
	}
	return 1
}

// javaDots rewrites the unescaped "." outside character classes of a
// regular expression as javaDot, so that RE2 matches like Java's
// java.util.regex, whose "." also excludes '\r', U+0085, U+2028 and U+2029.
func javaDots(pattern string) string {
	var sb strings.Builder
	inClass := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '\\' && i+1 < len(pattern):
			sb.WriteByte(c)
			i++
			sb.WriteByte(pattern[i])
			continue
		case c == '[' && !inClass:
			inClass = true
		case c == ']' && inClass:
			inClass = false
		case c == '.' && !inClass:
			sb.WriteString(javaDot)
			continue
		}
		sb.WriteByte(c)
	}
	return sb.String()
}
