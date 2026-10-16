// SPDX-License-Identifier: MIT
// Provenance: original

package json

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestEscapeAll(t *testing.T) {
	var sb strings.Builder
	for c := range 0x80 {
		sb.WriteByte(byte(c))
	}
	sb.WriteString("\u0080\u00ff\u2027\u2028\u2029\u202a\U0001f600\ufeff")
	if got := ValueOfString(sb.String()).String(); got != escapeAllWant {
		t.Errorf("String() =\n%q\nwant\n%q", got, escapeAllWant)
	}
}

func TestEscapeMemberName(t *testing.T) {
	o := NewObject().Add("k\"\u2028\x01", ValueOfString("v"))
	if got := o.String(); got != escapeMemberNameWant {
		t.Errorf("String() = %q; want %q", got, escapeMemberNameWant)
	}
}

func TestEscapeControlAndSeparators(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", `""`},
		{"plain", `"plain"`},
		{"\b\f\n\r\t", `"\u0008\u000c\n\r\t"`},
		{"\x00\x1f\x7f", `"\u0000\u001f` + "\x7f" + `"`},
		{"a\u2028b\u2029c", `"a\u2028b\u2029c"`},
		{"/", `"/"`},
		{`\"`, `"\\\""`},
		// Characters near the separators are written as is.
		{"\u2027\u202a", "\"\u2027\u202a\""},
		// Invalid UTF-8 bytes are written as is.
		{"\xe2\x80", "\"\xe2\x80\""},
		{"\xff", "\"\xff\""},
	}
	for _, tt := range tests {
		if got := ValueOfString(tt.in).String(); got != tt.want {
			t.Errorf("ValueOfString(%q).String() = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestPrettyPrintConfigs(t *testing.T) {
	v, err := Parse(`{"a":[1,{}],"b":{"c":"d"}}`)
	if err != nil {
		t.Fatal(err)
	}
	indent0, err := IndentWithSpaces(0)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.StringConfig(indent0); got != indent0Want {
		t.Errorf("IndentWithSpaces(0) = %q; want %q", got, indent0Want)
	}
	indent4, err := IndentWithSpaces(4)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.StringConfig(indent4); got != indent4Want {
		t.Errorf("IndentWithSpaces(4) = %q; want %q", got, indent4Want)
	}
	if _, err := IndentWithSpaces(-1); err == nil || err.Error() != "number is negative" {
		t.Errorf("IndentWithSpaces(-1) error = %v; want number is negative", err)
	}
	// Scalars are not affected by the configuration.
	for _, s := range []Value{Null, ValueOfInt(3), ValueOfString("a\nb")} {
		if got := s.StringConfig(Pretty); got != s.String() {
			t.Errorf("StringConfig(Pretty) = %q; want %q", got, s.String())
		}
	}
}

// failingWriter fails after accepting limit bytes.
type failingWriter struct {
	limit int
	buf   bytes.Buffer
}

var errWrite = errors.New("write failed")

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.buf.Len()+len(p) > w.limit {
		return 0, errWrite
	}
	return w.buf.Write(p)
}

func TestWriteJSON(t *testing.T) {
	long := strings.Repeat("x", 300)
	o := NewObject().
		Add("long", ValueOfString(long)).
		Add("items", NewArray(ValueOfInt(1), ValueOfString(strings.Repeat("y", 100)), True))
	for range 20 {
		o.Add("k", ValueOfInt(7))
	}
	for _, config := range []WriterConfig{nil, Minimal, Pretty, SingleLine(), IndentWithTabs()} {
		var buf bytes.Buffer
		if err := o.WriteJSON(&buf, config); err != nil {
			t.Fatal(err)
		}
		if got, want := buf.String(), o.StringConfig(config); got != want {
			t.Errorf("WriteJSON() = %q; want %q", got, want)
		}
	}
	if !strings.HasPrefix(o.String(), `{"long":"`+long+`","items":[1,"yyyy`) {
		t.Errorf("String() = %q", o.String())
	}

	if err := o.WriteJSON(&failingWriter{limit: 50}, Minimal); !errors.Is(err, errWrite) {
		t.Errorf("WriteJSON to a failing writer: %v; want %v", err, errWrite)
	}
	if err := ValueOfInt(1).WriteJSON(&failingWriter{limit: 0}, Minimal); !errors.Is(err, errWrite) {
		t.Errorf("WriteJSON to a failing writer: %v; want %v", err, errWrite)
	}
	if err := o.WriteJSON(nil, Minimal); err == nil {
		t.Error("WriteJSON(nil) succeeded")
	}
}

func TestWritingBuffer(t *testing.T) {
	var buf bytes.Buffer
	b := newWritingBuffer(&buf, 4)
	b.writeString("ab")
	b.writeByte('c')
	b.writeString("de") // flushes "abc" first
	b.writeString("123456")
	b.writeByte('z')
	if got := buf.String(); got != "abcde123456" {
		t.Errorf("before flush: %q; want abcde123456", got)
	}
	if err := b.flush(); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "abcde123456z" {
		t.Errorf("after flush: %q; want abcde123456z", got)
	}
}

func TestWriteJSONScalars(t *testing.T) {
	for _, v := range []Value{Null, ValueOfInt(-3), ValueOfString("a\tb"), NewArray(), NewObject()} {
		var buf bytes.Buffer
		if err := v.WriteJSON(&buf, Pretty); err != nil {
			t.Fatal(err)
		}
		if got, want := buf.String(), v.StringConfig(Pretty); got != want {
			t.Errorf("WriteJSON() = %q; want %q", got, want)
		}
	}
}
