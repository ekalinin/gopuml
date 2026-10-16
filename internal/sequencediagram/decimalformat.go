// SPDX-License-Identifier: MIT
// Provenance: original

package sequencediagram

import (
	"strconv"
	"strings"
	"unicode/utf16"
)

// decimalFormat emulates the parts of java.text.DecimalFormat that
// autonumber formats use: the pattern syntax documented in DecimalFormat's
// API (prefix, integer and fraction digits, grouping, exponent, suffix,
// quotes, percent, per mille, currency and an optional negative
// subpattern) and the formatting of integers with the symbols of the
// en_US locale, the default locale of the reference image (the currency
// sign is "$", the doubled sign "USD"). It is written from the documentation and checked against the JVM
// with the differential tests of this package; it does not derive from the
// JDK sources.
type decimalFormat struct {
	posPrefix, posSuffix string
	negPrefix, negSuffix string

	minInt, maxInt         int
	minFrac, maxFrac       int
	groupingSize           int
	grouping               bool
	decimalAlwaysShown     bool
	multiplier             int64
	useExponent            bool
	minExponentDigits      int
	engineeringMultipleMax bool
}

const (
	dfDigit     = '#'
	dfZero      = '0'
	dfGrouping  = ','
	dfDecimal   = '.'
	dfPercent   = '%'
	dfPerMill   = '‰'
	dfCurrency  = '¤'
	dfMinus     = '-'
	dfSeparator = ';'
	dfExponent  = 'E'
	dfQuote     = '\''
)

// patternError is the IllegalArgumentException of an invalid pattern.
type patternError struct{ msg string }

func (e *patternError) Error() string { return e.msg }

// newDecimalFormat parses a pattern as DecimalFormat(String) does.
func newDecimalFormat(pattern string) (df *decimalFormat, err error) {
	defer func() {
		if r := recover(); r != nil {
			pe, ok := r.(*patternError)
			if !ok {
				panic(r)
			}
			df, err = nil, pe
		}
	}()
	return parseDecimalFormat(pattern), nil
}

func parseDecimalFormat(pattern string) *decimalFormat {
	p := utf16.Encode([]rune(pattern))
	fail := func(msg string) { panic(&patternError{msg}) }
	quoted := "\"" + pattern + "\""
	df := &decimalFormat{multiplier: 1}
	if len(p) == 0 {
		// DecimalFormat("") formats like "#,##0.###".
		df.minInt, df.maxInt, df.maxFrac, df.groupingSize, df.grouping = 1, maxIntDigits, 3, 3, true
		df.negPrefix = "-"
		return df
	}

	start := 0
	gotNegative := false
	for j := 0; j < 2 && start < len(p); j++ {
		var prefix, suffix strings.Builder
		affix := &prefix
		phase := 0
		inQuote := false
		digitLeft, zeroDigit, digitRight := 0, 0, 0
		groupingCount := -1
		decimalPos := -1
		multiplier := int64(1)
		useExp := false
		minExp := 0
		pos := start
		for ; pos < len(p); pos++ {
			ch := p[pos]
			switch phase {
			case 0, 2:
				if inQuote {
					if ch == dfQuote {
						if pos+1 < len(p) && p[pos+1] == dfQuote {
							pos++
							affix.WriteRune(dfQuote)
						} else {
							inQuote = false
						}
						continue
					}
					affix.WriteRune(rune(ch))
					continue
				}
				switch ch {
				case dfDigit, dfZero, dfGrouping, dfDecimal:
					phase = 1
					pos--
					continue
				case dfCurrency:
					if pos+1 < len(p) && p[pos+1] == dfCurrency {
						pos++
						affix.WriteString(currencyCode)
					} else {
						affix.WriteString(currencySymbol)
					}
					continue
				case dfQuote:
					if pos+1 < len(p) && p[pos+1] == dfQuote {
						pos++
						affix.WriteRune(dfQuote)
					} else {
						inQuote = true
					}
					continue
				case dfSeparator:
					if phase == 0 || j == 1 {
						fail("Unquoted special character ';' in pattern " + quoted)
					}
					start = pos + 1
					pos = len(p)
					gotNegative = true
					continue
				case dfPercent, dfPerMill:
					if multiplier != 1 {
						fail("Too many percent/per mille characters in pattern " + quoted)
					}
					if ch == dfPercent {
						multiplier = 100
					} else {
						multiplier = 1000
					}
					affix.WriteRune(rune(ch))
					continue
				}
				affix.WriteRune(rune(ch))
			case 1:
				if j == 1 {
					// The number part of the negative subpattern is ignored.
					switch ch {
					case dfDigit, dfZero, dfGrouping, dfDecimal:
						continue
					}
					phase = 2
					affix = &suffix
					pos--
					continue
				}
				switch ch {
				case dfDigit:
					if zeroDigit > 0 {
						digitRight++
					} else {
						digitLeft++
					}
					if groupingCount >= 0 && decimalPos < 0 {
						groupingCount++
					}
				case dfZero:
					if digitRight > 0 {
						fail("Unexpected '0' in pattern " + quoted)
					}
					zeroDigit++
					if groupingCount >= 0 && decimalPos < 0 {
						groupingCount++
					}
				case dfGrouping:
					groupingCount = 0
				case dfDecimal:
					if decimalPos >= 0 {
						fail("Multiple decimal separators in pattern " + quoted)
					}
					decimalPos = digitLeft + zeroDigit + digitRight
				default:
					if ch == dfExponent && !useExp {
						useExp = true
						minExp = 0
						pos++
						for pos < len(p) && p[pos] == dfZero {
							minExp++
							pos++
						}
						if digitLeft+zeroDigit < 1 || minExp < 1 {
							fail("Malformed exponential pattern " + quoted)
						}
						phase = 2
						affix = &suffix
						pos--
						continue
					}
					phase = 2
					affix = &suffix
					pos--
					continue
				}
			}
		}
		if j == 0 && pos >= len(p) && !gotNegative {
			start = len(p)
		}

		if j == 0 {
			if zeroDigit == 0 && digitLeft > 0 && decimalPos >= 0 {
				n := decimalPos
				if n == 0 {
					n++
				}
				digitRight = digitLeft - n
				digitLeft = n - 1
				zeroDigit = 1
			}
			if (decimalPos < 0 && digitRight > 0) ||
				(decimalPos >= 0 && (decimalPos < digitLeft || decimalPos > digitLeft+zeroDigit)) ||
				groupingCount == 0 || inQuote {
				fail("Malformed pattern " + quoted)
			}
			df.posPrefix, df.posSuffix = prefix.String(), suffix.String()
			total := digitLeft + zeroDigit + digitRight
			effectiveDecimalPos := total
			if decimalPos >= 0 {
				effectiveDecimalPos = decimalPos
			}
			df.minInt = effectiveDecimalPos - digitLeft
			if useExp {
				df.maxInt = digitLeft + df.minInt
			} else {
				df.maxInt = maxIntDigits
			}
			if decimalPos >= 0 {
				df.maxFrac = total - decimalPos
				df.minFrac = digitLeft + zeroDigit - decimalPos
			}
			df.grouping = groupingCount > 0
			if groupingCount > 0 {
				df.groupingSize = groupingCount
			}
			df.multiplier = multiplier
			df.decimalAlwaysShown = decimalPos == 0 || decimalPos == total
			df.useExponent = useExp
			df.minExponentDigits = minExp
			df.negPrefix, df.negSuffix = "-"+df.posPrefix, df.posSuffix
		} else {
			if inQuote {
				fail("Malformed pattern " + quoted)
			}
			df.negPrefix, df.negSuffix = prefix.String(), suffix.String()
		}
	}
	return df
}

// The currency of the en_US locale.
const (
	currencySymbol = "$"
	currencyCode   = "USD"
)

// maxIntDigits stands for Java's unbounded maximum of integer digits.
const maxIntDigits = 1 << 30

// format formats an integer.
func (df *decimalFormat) format(n int64) string {
	negative := n < 0
	var digits string
	if negative {
		digits = strconv.FormatUint(uint64(-n), 10)
	} else {
		digits = strconv.FormatInt(n, 10)
	}
	if df.multiplier != 1 && digits != "0" {
		digits += strings.Repeat("0", len(strconv.FormatInt(df.multiplier, 10))-1)
	}
	var body string
	if df.useExponent {
		body = df.formatExponent(digits)
	} else {
		body = df.formatPlain(digits)
	}
	if negative {
		return df.negPrefix + body + df.negSuffix
	}
	return df.posPrefix + body + df.posSuffix
}

func (df *decimalFormat) formatPlain(digits string) string {
	if digits == "0" {
		digits = ""
	}
	if len(digits) < df.minInt {
		digits = strings.Repeat("0", df.minInt-len(digits)) + digits
	}
	var sb strings.Builder
	if df.grouping && df.groupingSize > 0 {
		for i := 0; i < len(digits); i++ {
			if i > 0 && (len(digits)-i)%df.groupingSize == 0 {
				sb.WriteByte(dfGrouping)
			}
			sb.WriteByte(digits[i])
		}
	} else {
		sb.WriteString(digits)
	}
	if digits == "" && df.minFrac == 0 {
		// Java writes a zero when there would be no digit at all.
		sb.WriteByte('0')
	}
	if df.minFrac > 0 || df.decimalAlwaysShown {
		sb.WriteByte(dfDecimal)
	}
	sb.WriteString(strings.Repeat("0", df.minFrac))
	return sb.String()
}

// roundHalfEven rounds a string of decimal digits to n significant digits;
// it returns the kept digits and the change of magnitude (1 when rounding
// up adds a digit).
func roundHalfEven(digits string, n int) (string, int) {
	if n >= len(digits) {
		return digits, 0
	}
	if n < 0 {
		n = 0
	}
	kept := []byte(digits[:n])
	rest := digits[n:]
	up := false
	switch {
	case rest[0] > '5':
		up = true
	case rest[0] == '5':
		if strings.TrimRight(rest[1:], "0") != "" {
			up = true
		} else {
			up = n > 0 && (kept[n-1]-'0')%2 == 1
		}
	}
	if !up {
		return string(kept), 0
	}
	for i := n - 1; i >= 0; i-- {
		if kept[i] < '9' {
			kept[i]++
			return string(kept), 0
		}
		kept[i] = '0'
	}
	return "1" + string(kept), 1
}

func (df *decimalFormat) formatExponent(digits string) string {
	engineering := df.maxInt > df.minInt && df.maxInt > 1
	minInt := df.minInt
	if engineering {
		minInt = 1
	}
	var intPart, fracPart string
	exponent := 0
	if digits == "0" {
		intPart = strings.Repeat("0", minInt)
	} else {
		magnitude := len(digits)
		significant := df.minInt + df.maxFrac
		if engineering {
			significant = df.maxInt + df.maxFrac
		}
		significant = max(significant, 1)
		kept, carry := roundHalfEven(digits, significant)
		magnitude += carry
		kept = strings.TrimRight(kept, "0")
		intDigits := minInt
		if engineering {
			exponent = (magnitude - 1) / df.maxInt * df.maxInt
			intDigits = magnitude - exponent
		} else {
			exponent = magnitude - minInt
		}
		if len(kept) < intDigits {
			kept += strings.Repeat("0", intDigits-len(kept))
		}
		intPart, fracPart = kept[:intDigits], kept[intDigits:]
	}
	if len(fracPart) < df.minFrac {
		fracPart += strings.Repeat("0", df.minFrac-len(fracPart))
	}
	var sb strings.Builder
	sb.WriteString(intPart)
	if fracPart != "" || df.decimalAlwaysShown {
		sb.WriteByte(dfDecimal)
	}
	sb.WriteString(fracPart)
	sb.WriteByte(dfExponent)
	if exponent < 0 {
		sb.WriteByte(dfMinus)
		exponent = -exponent
	}
	e := strconv.Itoa(exponent)
	if len(e) < df.minExponentDigits {
		e = strings.Repeat("0", df.minExponentDigits-len(e)) + e
	}
	sb.WriteString(e)
	return sb.String()
}
