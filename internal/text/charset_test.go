// SPDX-License-Identifier: MIT
// Provenance: original

package text

import (
	"io"
	"strings"
	"testing"
)

func TestDecodeUTF8(t *testing.T) {
	// Expected values follow Java's new String(bytes, UTF_8).
	for _, tt := range []struct {
		in   string
		want string
	}{
		{"caf\xc3\xa9", "café"},
		{"caf\xe9", "caf�"},
		{"caf\xe9!", "caf�!"},
		{"\xe2\x82A", "�A"},     // truncated 3-byte sequence
		{"\xe2\x82", "�"},       // truncated at the end
		{"\xff\xfe", "��"},      // two invalid bytes
		{"\xe0\x80\x80", "���"}, // overlong
		{"\xed\xa0\x80", "�"},   // encoded surrogate
		{"\xf0\x9f\x98", "�"},   // truncated 4-byte sequence
		{"\xf4\x90\x80\x80", "����"},
		{"\xc0\xaf", "��"},
	} {
		if got := DecodeUTF8([]byte(tt.in)); got != tt.want {
			t.Errorf("DecodeUTF8(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLookupCharset(t *testing.T) {
	for _, tt := range []struct {
		name, in, want string
	}{
		{"UTF-8", "caf\xc3\xa9", "café"},
		{"utf8", "caf\xe9", "caf�"},
		{"ISO-8859-1", "caf\xe9", "café"},
		{"latin1", "caf\xe9", "café"},
		{"ISO8859_1", "caf\xe9", "café"},
		{"windows-1252", "\x80", "€"},
		{"Cp1252", "\x80", "€"},
		{"UTF-16", "\xfe\xff\x00A", "A"},
		{"UTF-16LE", "A\x00", "A"},
	} {
		c, ok := LookupCharset(tt.name)
		if !ok {
			t.Errorf("LookupCharset(%q) unknown", tt.name)
			continue
		}
		if got := c.Decode([]byte(tt.in)); got != tt.want {
			t.Errorf("%s: Decode(%q) = %q, want %q", tt.name, tt.in, got, tt.want)
		}
		data, _ := io.ReadAll(c.NewReader(strings.NewReader(tt.in)))
		if got := DecodeUTF8(data); got != tt.want {
			t.Errorf("%s: NewReader(%q) = %q, want %q", tt.name, tt.in, got, tt.want)
		}
	}
	if _, ok := LookupCharset("no-such-charset"); ok {
		t.Error("unknown charset accepted")
	}
}

func TestRemoveQuotes(t *testing.T) {
	for in, want := range map[string]string{
		`"out"`: "out", "“out”": "out", "«out»": "out", "(out)": "out", "[out]": "out", ":out:": "out",
		`"`: `"`, `""`: "", "out": "out", `"out`: `"out`, "(out]": "(out]",
	} {
		if got := RemoveQuotes(in); got != want {
			t.Errorf("RemoveQuotes(%q) = %q, want %q", in, got, want)
		}
	}
	if got := RemoveQuotesFormat("(x)", `"`); got != "(x)" {
		t.Errorf("format without parentheses: %q", got)
	}
}
