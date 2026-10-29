// SPDX-License-Identifier: MIT
// Provenance: original

package jfmt

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// ParseDouble is Java's Double.parseDouble: the string without the
// characters up to the space at both ends, an optional sign, then NaN,
// Infinity, a decimal number with an optional exponent or a hexadecimal
// number with a binary exponent (0x1.8p1), optionally followed by f, F, d
// or D. It returns the JDK's NumberFormatException message for the other
// strings, such as Go's "inf" or "1_0".
func ParseDouble(s string) (float64, error) {
	in := strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
	if in == "" {
		return 0, errors.New("empty String")
	}
	bad := errors.New(`For input string: "` + in + `"`)
	body := in
	negative := false
	if body[0] == '+' || body[0] == '-' {
		negative = body[0] == '-'
		body = body[1:]
	}
	switch body {
	case "NaN":
		return math.NaN(), nil
	case "Infinity":
		if negative {
			return math.Inf(-1), nil
		}
		return math.Inf(1), nil
	}
	if n := len(body); n > 0 && strings.ContainsRune("fFdD", rune(body[n-1])) {
		body = body[:n-1]
	}
	if !javaDecimal(body) && !javaHex(body) {
		return 0, bad
	}
	v, err := strconv.ParseFloat(body, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, bad
	}
	if negative {
		v = -v
	}
	return v, nil
}

// digitsOf returns the length of the prefix of s made of the digits that
// accept says are digits.
func digitsOf(s string, accept func(byte) bool) int {
	i := 0
	for i < len(s) && accept(s[i]) {
		i++
	}
	return i
}

func isDecimalDigit(c byte) bool { return c >= '0' && c <= '9' }

func isHexDigit(c byte) bool {
	return isDecimalDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// exponent reports whether s is an exponent: a sign and at least one digit.
func exponent(s string) bool {
	if s != "" && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}
	return s != "" && digitsOf(s, isDecimalDigit) == len(s)
}

// javaDecimal reports whether s is digits with an optional point, with at
// least one digit, and an optional exponent (5., .5, 1e3).
func javaDecimal(s string) bool {
	n := digitsOf(s, isDecimalDigit)
	rest := s[n:]
	m := 0
	if strings.HasPrefix(rest, ".") {
		rest = rest[1:]
		m = digitsOf(rest, isDecimalDigit)
		rest = rest[m:]
	}
	if n+m == 0 {
		return false
	}
	if rest == "" {
		return true
	}
	return (rest[0] == 'e' || rest[0] == 'E') && exponent(rest[1:])
}

// javaHex reports whether s is a hexadecimal float with its required binary
// exponent (0x1p3, 0x1.8p1).
func javaHex(s string) bool {
	if len(s) < 2 || s[0] != '0' || (s[1] != 'x' && s[1] != 'X') {
		return false
	}
	rest := s[2:]
	n := digitsOf(rest, isHexDigit)
	rest = rest[n:]
	m := 0
	if strings.HasPrefix(rest, ".") {
		rest = rest[1:]
		m = digitsOf(rest, isHexDigit)
		rest = rest[m:]
	}
	if n+m == 0 || rest == "" {
		return false
	}
	return (rest[0] == 'p' || rest[0] == 'P') && exponent(rest[1:])
}
