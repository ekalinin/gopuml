// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/DiagramSourceCompressor.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/TranscoderImpl.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/TranscoderSmart.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/TranscoderSmartProtected.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package code

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

const (
	startUml = "@startuml"
	endUml   = "@enduml"
)

// compressSource keeps the lines between @startuml and @enduml, as
// upstream's DiagramSourceCompressor does before encoding.
func compressSource(data string) string {
	var sb, full strings.Builder
	appendLine := func(b *strings.Builder, s string) {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s)
	}
	startDone := false
	for _, l := range core.ReadUncommented(data, "COMPRESS") {
		appendLine(&full, l.S)
		switch {
		case strings.HasPrefix(l.S, "@startuml"):
			startDone = true
		case strings.HasPrefix(l.S, "@enduml"):
			return sb.String()
		case startDone:
			appendLine(&sb, l.S)
		}
	}
	if !startDone {
		return cleanSource(stripStartEnd(full.String(), true))
	}
	return sb.String()
}

func cleanSource(s string) string {
	s = text.Trin(s)
	s = stripStartEnd(s, false)
	s = removeFromToken(s, endUml)
	s = removeFromToken(s, startUml)
	return text.Trin(s)
}

// removeFromToken removes each token with the rest of its line.
func removeFromToken(s, token string) string {
	pos := strings.Index(s, token)
	if pos == -1 {
		return s
	}
	var sb strings.Builder
	from := 0
	for pos != -1 {
		sb.WriteString(s[from:pos])
		from = pos + len(token)
		for from < len(s) && s[from] != '\n' && s[from] != '\r' {
			from++
		}
		pos = strings.Index(s[from:], token)
		if pos != -1 {
			pos += from
		}
	}
	sb.WriteString(s[from:])
	return sb.String()
}

func isSourceBlank(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\u000B' || c == '\f' || c == '\r' || c == '\u00A0'
}

func isOtherLineTerminator(c rune) bool { return c == '\u0085' || c == '\u2028' || c == '\u2029' }

// stripStartEnd removes the blanks around s, a leading "@startuml" line and
// a trailing "@enduml" (in lower case only). With
// dollarBeforeFinalTerminator, one U+0085, U+2028 or U+2029 at the end is
// dropped first, as the "$" of the regex upstream used to apply.
func stripStartEnd(s string, dollarBeforeFinalTerminator bool) string {
	r := []rune(s)
	hasAt := func(i int, token string) bool {
		return i >= 0 && i+len(token) <= len(r) && string(r[i:i+len(token)]) == token
	}
	start := 0
	for start < len(r) && isSourceBlank(r[start]) {
		start++
	}
	if hasAt(start, startUml) {
		start += len(startUml)
		for start < len(r) && r[start] != '\n' && r[start] != '\r' {
			start++
		}
		for start < len(r) && isSourceBlank(r[start]) {
			start++
		}
	}
	end := len(r)
	if dollarBeforeFinalTerminator && end > start && isOtherLineTerminator(r[end-1]) {
		end--
	}
	for end > start && isSourceBlank(r[end-1]) {
		end--
	}
	if e := end - len(endUml); e >= start && hasAt(e, endUml) {
		end = e
		for end > start && isSourceBlank(r[end-1]) {
			end--
		}
	}
	return string(r[start:end])
}

// decompressSource adds @startuml/@enduml around decoded text that does not
// start with @start.
func decompressSource(s string) string {
	result := cleanSource(s)
	if strings.HasPrefix(result, "@start") {
		return result
	}
	result = "@startuml\n" + result
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result + "@enduml"
}

// Encode returns the URL text of a diagram source: the lines between
// @startuml and @enduml, deflated and written with PlantUML's alphabet.
func Encode(source string) string {
	return encodeASCII(deflateRaw([]byte(compressSource(source))))
}

func decodeWith(code string, decode func(string) ([]byte, error), inflate func([]byte) ([]byte, error)) (string, error) {
	data, err := decode(code)
	if err != nil {
		return "", err
	}
	if inflate != nil {
		if data, err = inflate(data); err != nil {
			return "", err
		}
	}
	return decompressSource(text.DecodeUTF8(data)), nil
}

// Decode returns the diagram source of URL text in any form upstream reads:
// ~0 (deflate), ~1 (legacy Huffman, zlib), ~h (hexadecimal), ~g (gzip),
// ~zip~ (zip), and unprefixed text (deflate, then legacy Huffman).
func Decode(code string) (string, error) {
	switch {
	case strings.HasPrefix(code, "~0"):
		return decodeWith(code[2:], decodeASCII, inflateRaw)
	case strings.HasPrefix(code, "~1"):
		return decodeWith(code[2:], decodeASCII, inflateZlib)
	case strings.HasPrefix(code, "~h"):
		return decodeWith(code[2:], decodeHex, nil)
	case strings.HasPrefix(code, "~g"):
		return decodeWith(code[2:], decodeASCII, gunzip)
	case strings.HasPrefix(code, "~zip~"):
		return decodeWith(code[5:], decodeASCII, unzip)
	}
	if s, err := decodeWith(code, decodeASCII, inflateRaw); err == nil {
		return s, nil
	}
	return decodeWith(code, decodeASCII, inflateZlib)
}

// DecodeProtected is the server's decoding: instead of failing on text that
// is not deflate (or legacy Huffman for ~1) data, it returns a diagram
// explaining that the URL is bad, as upstream's built-in server does.
func DecodeProtected(code string) (string, error) {
	switch {
	case strings.HasPrefix(code, "~1"):
		if s, err := decodeWith(code[2:], decodeASCII, inflateZlib); err == nil {
			return s, nil
		}
		return badHuffmanURL(code[2:]), nil
	case strings.HasPrefix(code, "~h"), strings.HasPrefix(code, "~g"), strings.HasPrefix(code, "~zip~"):
		return Decode(code)
	}
	data := strings.TrimPrefix(code, "~0")
	if s, err := decodeWith(data, decodeASCII, inflateRaw); err == nil {
		return s, nil
	}
	return badDeflateURL(data), nil
}

func appendURL(lines []string, url string) []string {
	for len(url) > 80 {
		lines = append(lines, url[:80])
		url = url[80:]
	}
	if url != "" {
		lines = append(lines, url)
	}
	return lines
}

func badHuffmanURL(code string) string {
	lines := []string{
		"@startuml",
		"legend",
		"The plugin you are using seems to generated a bad URL.",
		"This URL does not look like HUFFMAN data.",
		"",
		"See https://plantuml.com/text-encoding",
		"",
		"You may contact the PlantUML team at plantuml@gmail.com",
		"But you should also probably contact the plugin authors you are currently using and send them this image",
		"",
		"For the record, here is your data:",
		"",
	}
	lines = appendURL(lines, code)
	lines = append(lines, "endlegend", "@enduml")
	return strings.Join(lines, "\n") + "\n"
}

func badDeflateURL(code string) string {
	short := code
	if len(short) > 30 {
		short = short[:30] + "..."
	}
	lines := []string{
		"@startuml",
		"legend",
		"The plugin you are using seems to generated a bad URL.",
		"This URL does not look like DEFLATE data.",
		"It looks like your plugin is using HUFFMAN encoding.",
		"",
		"This means you have now to add an header ~1 to your data. For example, you have to change:",
		"http://www.plantuml.com/plantuml/png/" + short,
		"to",
		"http://www.plantuml.com/plantuml/png/~1" + short,
		"",
		"It will work this way",
		"You may contact the PlantUML team at plantuml@gmail.com",
		"But you should also probably contact the plugin authors you are currently using and send them this image",
		"",
		"For the record, here is your data:",
		"",
	}
	lines = appendURL(lines, code)
	lines = append(lines, "endlegend", "@enduml")
	return strings.Join(lines, "\n") + "\n"
}
