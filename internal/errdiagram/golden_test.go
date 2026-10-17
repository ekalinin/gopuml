// SPDX-License-Identifier: MIT
// Provenance: original

package errdiagram_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/conformance"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/errdiagram"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/text"
)

// goldenCase is a case of testdata/cases.txt (the format is described
// there).
type goldenCase struct {
	name, kind             string
	dark, red, noMetadata  bool
	noRaw                  bool
	lines, raw             []text.StringLocated
	trace                  int
	errorMessage, assumed  string
	location, mainLocation *text.LineLocation
}

func (c *goldenCase) addLine(s string) {
	next := c.location.OneLineRead()
	if c.location == c.mainLocation {
		c.mainLocation = next
	}
	c.location = next
	c.lines = append(c.lines, text.NewStringLocated(s, next))
}

// unescape reads \t, \\ and \uXXXX (UTF-16 units) as ErrGolden.java does.
func unescape(s string) string {
	var units []uint16
	in := utf16.Encode([]rune(s))
	for i := 0; i < len(in); i++ {
		if in[i] != '\\' || i+1 >= len(in) {
			units = append(units, in[i])
			continue
		}
		i++
		switch in[i] {
		case 't':
			units = append(units, '\t')
		case 'u':
			v, _ := strconv.ParseUint(string(utf16.Decode(in[i+1:i+5])), 16, 16)
			units = append(units, uint16(v))
			i += 4
		default:
			units = append(units, in[i])
		}
	}
	return string(utf16.Decode(units))
}

func parseGoldenCases(t *testing.T, file string) []*goldenCase {
	t.Helper()
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var cases []*goldenCase
	var c *goldenCase
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, arg, _ := strings.Cut(line, " ")
		arg = unescape(arg)
		switch key {
		case "case":
			loc := text.NewLineLocation("string", nil)
			c = &goldenCase{name: arg, kind: "error", trace: -1, location: loc, mainLocation: loc}
		case "kind":
			c.kind = arg
		case "option":
			switch arg {
			case "dark":
				c.dark = true
			case "red":
				c.red = true
			case "nometadata":
				c.noMetadata = true
			default:
				t.Fatalf("unknown option %q", arg)
			}
		case "line":
			c.addLine(arg)
		case "repeat":
			n, rest, _ := strings.Cut(arg, " ")
			count, _ := strconv.Atoi(n)
			for i := 1; i <= count; i++ {
				c.addLine(strings.ReplaceAll(rest, "%d", strconv.Itoa(i)))
			}
		case "include":
			desc, index, _ := strings.Cut(arg, " ")
			i, _ := strconv.Atoi(index)
			c.location = text.NewLineLocation(desc, c.lines[i].Location)
		case "endinclude":
			c.location = c.mainLocation
		case "raw":
			c.raw = append(c.raw, text.NewStringLocated(arg, nil))
		case "noraw":
			c.noRaw = true
		case "trace":
			c.trace, _ = strconv.Atoi(arg)
		case "error":
			c.errorMessage = arg
		case "assumed":
			c.assumed = arg
		case "end":
			cases = append(cases, c)
			c = nil
		default:
			t.Fatalf("unknown line %q", line)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return cases
}

// diagramType is DiagramType.valueOf.
func diagramType(t *testing.T, name string) *core.DiagramType {
	t.Helper()
	for dt := core.Sequence; dt <= core.Unknown; dt++ {
		if dt.String() == name {
			return &dt
		}
	}
	t.Fatalf("unknown diagram type %q", name)
	return nil
}

// render draws a golden case as upstream's ErrGolden does.
func (c *goldenCase) render(t *testing.T) ([]byte, export.ImageData) {
	t.Helper()
	opt := export.NewFileFormatOptionMetadata(font.FormatSVG, !c.noMetadata)
	if c.red {
		opt = opt.WithUseRedForError()
	}
	if c.dark {
		opt = opt.WithColorMapper(color.DarkModeMapper)
	}
	var buf bytes.Buffer
	if c.kind == "novalid" {
		data, err := errdiagram.ExportNoValidStart(&buf, opt)
		if err != nil {
			t.Fatal(err)
		}
		return buf.Bytes(), data
	}
	src := errdiagram.Source{Data: c.lines, Raw: c.raw}
	var d core.Diagram
	switch c.kind {
	case "error":
		n := len(c.lines)
		if c.trace >= 0 {
			n = c.trace
		}
		trace := c.lines[:n]
		last := trace[n-1]
		e := core.ErrorUml{Type: core.SyntaxError, Error: c.errorMessage, Line: &last}
		if c.assumed != "" {
			e.AssumedType = diagramType(t, c.assumed)
		}
		d = &core.ErrorDiagram{Source: c.lines, Trace: trace, Error: e}
	case "preprocerror":
		input := append([]text.StringLocated(nil), c.lines...)
		input[len(input)-1].PreprocessorError = c.errorMessage
		last := input[len(input)-1]
		d = &core.ErrorDiagram{Source: input, Trace: input,
			Error: core.ErrorUml{Type: core.SyntaxError, Error: c.errorMessage, Line: &last}}
		src = errdiagram.Source{Data: input}
	case "unsupported":
		d = &core.UnsupportedDiagram{Directive: c.lines[0].S}
	case "welcome":
		raw := c.raw
		if raw == nil {
			raw = c.lines
		}
		d = errdiagram.WelcomeFactory{}.CreateSystem(core.NewRenderContext(), core.NewUmlSource(c.lines, raw, true))
		if d == nil {
			t.Fatal("WelcomeFactory refused a block of two lines")
		}
	default:
		t.Fatalf("unknown kind %q", c.kind)
	}
	data, err := errdiagram.Export(&buf, d, src, opt)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), data
}

// TestGolden compares the cases of testdata/cases.txt, which have no
// conformance reference, with the images of the upstream jar
// (testdata/golden, from testdata/javagen/ErrGolden.java).
func TestGolden(t *testing.T) {
	refData := map[string]conformance.RenderMeta{}
	data, err := os.ReadFile("testdata/golden/imagedata.txt")
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		var name string
		var m conformance.RenderMeta
		if _, err := fmt.Sscan(line, &name, &m.Width, &m.Height, &m.Status); err != nil {
			t.Fatalf("imagedata.txt: %q: %v", line, err)
		}
		refData[name] = m
	}
	cases := parseGoldenCases(t, "testdata/cases.txt")
	if len(cases) != len(refData) {
		t.Errorf("%d cases, %d golden image data", len(cases), len(refData))
	}
	var r result
	for _, c := range cases {
		got, img := c.render(t)
		r.compare(t, c.name, got, img, filepath.Join("testdata", "golden", c.name+".svg"), refData[c.name])
	}
	t.Logf("golden images: %v", r)
	if r.identical+r.exceptMetadata != len(cases) {
		t.Errorf("%d of %d images match", r.identical+r.exceptMetadata, len(cases))
	}
}
