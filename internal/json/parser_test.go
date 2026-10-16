// SPDX-License-Identifier: MIT
// Provenance: original

package json

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
)

// parseCase is one row of parseCases: either the texts of the parsed value,
// or err and loc for a ParseError, or other for another error.
type parseCase struct {
	in                            string
	minimal, pretty, single, tabs string
	err                           string
	loc                           Location
	other                         string
}

func TestParseCases(t *testing.T) {
	for _, tc := range parseCases {
		t.Run(strconv.Quote(tc.in), func(t *testing.T) {
			v, err := Parse(tc.in)
			switch {
			case tc.err != "":
				pe, ok := errors.AsType[*ParseError](err)
				if !ok {
					t.Fatalf("Parse() = %v, %v; want ParseError %q", v, err, tc.err)
				}
				if pe.Error() != tc.err || pe.Location != tc.loc {
					t.Errorf("error = %q at %+v; want %q at %+v", pe.Error(), pe.Location, tc.err, tc.loc)
				}
			case tc.other != "":
				// other is "ExceptionClass: message".
				_, want, _ := strings.Cut(tc.other, ": ")
				if _, ok := errors.AsType[*IndexOutOfBoundsError](err); !ok || err.Error() != want {
					t.Errorf("Parse() error = %v; want IndexOutOfBoundsError %q", err, want)
				}
			default:
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}
				checkTexts(t, v, tc.minimal, tc.pretty, tc.single, tc.tabs)
			}
		})
	}
}

// checkTexts checks the four formattings of v.
func checkTexts(t *testing.T, v Value, minimal, pretty, single, tabs string) {
	t.Helper()
	if got := v.String(); got != minimal {
		t.Errorf("String() = %q; want %q", got, minimal)
	}
	if got := v.StringConfig(nil); got != minimal {
		t.Errorf("StringConfig(nil) = %q; want %q", got, minimal)
	}
	if got := v.StringConfig(Minimal); got != minimal {
		t.Errorf("StringConfig(Minimal) = %q; want %q", got, minimal)
	}
	if got := v.StringConfig(Pretty); got != pretty {
		t.Errorf("StringConfig(Pretty) = %q; want %q", got, pretty)
	}
	if got := v.StringConfig(SingleLine()); got != single {
		t.Errorf("StringConfig(SingleLine()) = %q; want %q", got, single)
	}
	if got := v.StringConfig(IndentWithTabs()); got != tabs {
		t.Errorf("StringConfig(IndentWithTabs()) = %q; want %q", got, tabs)
	}
}

func TestParseLongInputs(t *testing.T) {
	// Strings with escapes across several buffer fills.
	var in, minimal, pretty []string
	for i := range 300 {
		s := `"s` + strconv.Itoa(i) + `\n"`
		in = append(in, s)
		minimal = append(minimal, s)
		pretty = append(pretty, s)
	}
	v, err := Parse("[" + strings.Join(in, ",\n") + "]")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v.String(), "["+strings.Join(minimal, ",")+"]"; got != want {
		t.Errorf("String() = %q; want %q", got, want)
	}
	if got, want := v.StringConfig(Pretty), "[\n  "+strings.Join(pretty, ",\n  ")+"\n]"; got != want {
		t.Errorf("StringConfig(Pretty) = %q; want %q", got, want)
	}

	// A string longer than the buffer.
	var sb strings.Builder
	for i := range 1500 {
		sb.WriteByte(byte('a' + i%26))
	}
	long := sb.String()
	v, err = Parse(`{"k":"` + long + `\t", "n": 12345678901234567890.5e-3, ` + "\n\n" + `"bad": tru}`)
	pe, ok := errors.AsType[*ParseError](err)
	if !ok || pe.Error() != "Expected 'e' at 3:11" || pe.Location != (Location{1555, 3, 11}) {
		t.Errorf("Parse() = %v, %v; want Expected 'e' at 3:11 (offset 1555)", v, err)
	}
	v, err = Parse(`{"k":"` + long + `\t", "n": 12345678901234567890.5e-3}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.AsObject().Get("k").AsString(); got != long+"\t" {
		t.Errorf("k = %q; want the long string and a tab", got)
	}
	if got := v.AsObject().Get("n").String(); got != "12345678901234567890.5e-3" {
		t.Errorf("n = %q; want the number text verbatim", got)
	}
}

func TestParseNesting(t *testing.T) {
	nested := func(n int) string { return strings.Repeat("[", n) + strings.Repeat("]", n) }
	if _, err := Parse(nested(1000)); err != nil {
		t.Errorf("1000 levels: %v", err)
	}
	if _, err := Parse(nested(1001)); err == nil || err.Error() != "Nesting too deep at 1:1002" {
		t.Errorf("1001 levels: %v; want Nesting too deep at 1:1002", err)
	}
	if _, err := Parse(strings.Repeat(`{"a":`, 1001)); err == nil || err.Error() != "Nesting too deep at 1:5002" {
		t.Errorf("1001 object levels: %v; want Nesting too deep at 1:5002", err)
	}
}

func TestParseReader(t *testing.T) {
	v, err := ParseReader(iotest.OneByteReader(strings.NewReader(`{"a": [1, "x\u00e9"]}`)))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v.String(), "{\"a\":[1,\"x\xc3\xa9\"]}"; got != want {
		t.Errorf("String() = %q; want %q", got, want)
	}
	if _, err := ParseReader(strings.NewReader("[1,")); err == nil || err.Error() != "Unexpected end of input at 1:4" {
		t.Errorf("ParseReader error = %v", err)
	}
	readErr := errors.New("boom")
	if _, err := ParseReader(iotest.ErrReader(readErr)); !errors.Is(err, readErr) {
		t.Errorf("ParseReader error = %v; want %v", err, readErr)
	}
	if _, err := ParseReader(nil); err == nil {
		t.Error("ParseReader(nil) succeeded")
	}
	p := NewParser(&DefaultHandler{})
	if err := p.ParseReaderSize(strings.NewReader("1"), 0); err == nil || err.Error() != "buffersize is zero or negative" {
		t.Errorf("ParseReaderSize(0) error = %v", err)
	}
	// A tiny buffer gives the same result.
	h := &DefaultHandler{}
	if err := NewParser(h).ParseReaderSize(strings.NewReader("{\"a\" : \"b\\nc\",\n\"d\":[1.5e3]}"), 1); err != nil {
		t.Fatal(err)
	}
	if got, want := h.Value().String(), `{"a":"b\nc","d":[1.5e3]}`; got != want {
		t.Errorf("String() = %q; want %q", got, want)
	}
}

// TestParserReuse checks that a failed parse leaves its captured string
// content and nesting level behind, as upstream does.
func TestParserReuse(t *testing.T) {
	h := &DefaultHandler{}
	p := NewParser(h)
	if err := p.Parse(`["ab\ncd`); err == nil || err.Error() != "Unexpected end of input at 1:9" {
		t.Fatalf("first parse error = %v", err)
	}
	if err := p.Parse(`"xy"`); err != nil {
		t.Fatal(err)
	}
	if got, want := h.Value().String(), `"ab\ncdxy"`; got != want {
		t.Errorf("second parse = %s; want %s", got, want)
	}
	if err := p.Parse(`"zz"`); err != nil {
		t.Fatal(err)
	}
	if got, want := h.Value().String(), `"zz"`; got != want {
		t.Errorf("third parse = %s; want %s", got, want)
	}
	if err := p.Parse("[[[[x"); err == nil || err.Error() != "Expected value at 1:5" {
		t.Fatalf("fourth parse error = %v", err)
	}
	nested := strings.Repeat("[", 997) + strings.Repeat("]", 997)
	if err := p.Parse(nested); err == nil || err.Error() != "Nesting too deep at 1:997" {
		t.Errorf("fifth parse error = %v; want Nesting too deep at 1:997", err)
	}
	if err := NewParser(&DefaultHandler{}).Parse(nested); err != nil {
		t.Errorf("fresh parser: %v", err)
	}
}

// recorder logs the parser events with their locations.
type recorder struct {
	BaseHandler[string, string]
	log []string
}

func (r *recorder) l(event string) {
	loc := r.Location()
	r.log = append(r.log, fmt.Sprintf("%s@%d/%s", event, loc.Offset, loc))
}

func (r *recorder) StartNull()         { r.l("startNull") }
func (r *recorder) EndNull()           { r.l("endNull") }
func (r *recorder) StartBoolean()      { r.l("startBoolean") }
func (r *recorder) EndBoolean(v bool)  { r.l("endBoolean(" + strconv.FormatBool(v) + ")") }
func (r *recorder) StartString()       { r.l("startString") }
func (r *recorder) EndString(s string) { r.l("endString(" + s + ")") }
func (r *recorder) StartNumber()       { r.l("startNumber") }
func (r *recorder) EndNumber(s string) { r.l("endNumber(" + s + ")") }
func (r *recorder) StartArray() string { r.l("startArray"); return "A" }
func (r *recorder) EndArray(a string)  { r.l("endArray(" + a + ")") }
func (r *recorder) StartArrayValue(a string) {
	r.l("startArrayValue(" + a + ")")
}
func (r *recorder) EndArrayValue(a string) { r.l("endArrayValue(" + a + ")") }
func (r *recorder) StartObject() string    { r.l("startObject"); return "O" }
func (r *recorder) EndObject(o string)     { r.l("endObject(" + o + ")") }
func (r *recorder) StartObjectName(o string) {
	r.l("startObjectName(" + o + ")")
}
func (r *recorder) EndObjectName(o, n string) { r.l("endObjectName(" + o + "," + n + ")") }
func (r *recorder) StartObjectValue(o, n string) {
	r.l("startObjectValue(" + o + "," + n + ")")
}
func (r *recorder) EndObjectValue(o, n string) { r.l("endObjectValue(" + o + "," + n + ")") }

func TestHandlerEvents(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{
			"{\"a\" : [1, true, null],\n \"b\":\"x\\ty\" // c\n}",
			"startObject@0/1:1 | startObjectName(O)@1/1:2 | endObjectName(O,a)@4/1:5 | startObjectValue(O,a)@7/1:8 | " +
				"startArray@7/1:8 | startArrayValue(A)@8/1:9 | startNumber@8/1:9 | endNumber(1)@9/1:10 | endArrayValue(A)@9/1:10 | " +
				"startArrayValue(A)@11/1:12 | startBoolean@11/1:12 | endBoolean(true)@15/1:16 | endArrayValue(A)@15/1:16 | " +
				"startArrayValue(A)@17/1:18 | startNull@17/1:18 | endNull@21/1:22 | endArrayValue(A)@21/1:22 | endArray(A)@22/1:23 | " +
				"endObjectValue(O,a)@22/1:23 | startObjectName(O)@25/2:2 | endObjectName(O,b)@28/2:5 | startObjectValue(O,b)@29/2:6 | " +
				"startString@29/2:6 | endString(x\ty)@35/2:12 | endObjectValue(O,b)@35/2:12 | endObject(O)@42/3:2",
		},
		{
			"[ -1.5e3 , false ]",
			"startArray@0/1:1 | startArrayValue(A)@2/1:3 | startNumber@2/1:3 | endNumber(-1.5e3)@8/1:9 | endArrayValue(A)@8/1:9 | " +
				"startArrayValue(A)@11/1:12 | startBoolean@11/1:12 | endBoolean(false)@16/1:17 | endArrayValue(A)@16/1:17 | endArray(A)@18/1:19",
		},
		{
			// A surrogate pair escape gives one supplementary character.
			`"\uD83D\uDE00x"`,
			"startString@0/1:1 | endString(\xf0\x9f\x98\x80x)@15/1:16",
		},
		{
			"[1, x]",
			"startArray@0/1:1 | startArrayValue(A)@1/1:2 | startNumber@1/1:2 | endNumber(1)@2/1:3 | endArrayValue(A)@2/1:3 | " +
				"startArrayValue(A)@4/1:5 | error Expected value at 1:5",
		},
	}
	for _, tt := range tests {
		r := &recorder{}
		if err := NewParser(r).Parse(tt.in); err != nil {
			r.log = append(r.log, "error "+err.Error())
		}
		if got := strings.Join(r.log, " | "); got != tt.want {
			t.Errorf("events of %q:\n got %s\nwant %s", tt.in, got, tt.want)
		}
	}
}

func TestBaseHandlerWithoutParser(t *testing.T) {
	var h BaseHandler[int, int]
	if got := h.Location(); got != (Location{}) {
		t.Errorf("Location() = %v; want the zero Location", got)
	}
	if h.StartArray() != 0 || h.StartObject() != 0 {
		t.Error("StartArray or StartObject returned a non-zero handler value")
	}
}

func TestDefaultHandlerBeforeParse(t *testing.T) {
	if v := (&DefaultHandler{}).Value(); v != nil {
		t.Errorf("Value() = %v; want nil", v)
	}
}

func TestParseLoneSurrogate(t *testing.T) {
	// Java keeps the lone surrogate; a Go string holds U+FFFD instead.
	v, err := Parse(`"a\uD800b"`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v.AsString(), "a\xef\xbf\xbdb"; got != want {
		t.Errorf("AsString() = %q; want %q", got, want)
	}
}

func TestLocationString(t *testing.T) {
	if got := (Location{Offset: 12, Line: 3, Column: 4}).String(); got != "3:4" {
		t.Errorf("String() = %q; want 3:4", got)
	}
}

// panicking is a handler whose StartNull panics.
type panicking struct {
	BaseHandler[any, any]
}

func (*panicking) StartNull() { panic("handler failure") }

func TestHandlerPanicPropagates(t *testing.T) {
	defer func() {
		if r := recover(); r != "handler failure" {
			t.Errorf("recover() = %v; want the handler panic", r)
		}
	}()
	_ = NewParser(&panicking{}).Parse("[null]")
	t.Error("Parse returned")
}
