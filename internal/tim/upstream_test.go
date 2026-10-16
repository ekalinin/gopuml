// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"encoding/json"
	"io/fs"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/security"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// upstreamCase is one case of testdata/upstream/upstream.json, produced by
// testdata/upstream/Harness.java with the reference jar.
type upstreamCase struct {
	Name        string            `json:"name"`
	Profile     string            `json:"profile"`
	IncludePath string            `json:"includePath"`
	Dir         string            `json:"dir"`
	Defines     map[string]string `json:"defines"`
	Files       map[string]string `json:"files"`
	Main        string            `json:"main"`
	Crash       string            `json:"crash"`
	Stderr      string            `json:"stderr"`
	Blocks      []upstreamBlock   `json:"blocks"`
}

type upstreamBlock struct {
	Error    bool              `json:"error"`
	Lines    []upstreamLine    `json:"lines"`
	Debug    []upstreamLine    `json:"debug"`
	Options  map[string]string `json:"options"`
	Warnings []string          `json:"warnings"`
}

type upstreamLine struct {
	S   string `json:"s"`
	Loc string `json:"loc"`
	Err string `json:"err"`
}

type upstreamDoc struct {
	LineTypes []struct {
		S    string `json:"s"`
		Type string `json:"type"`
	} `json:"lineTypes"`
	Cases []upstreamCase `json:"cases"`
}

func loadUpstreamDoc(t *testing.T) upstreamDoc {
	t.Helper()
	data, err := os.ReadFile("testdata/upstream/upstream.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc upstreamDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

// TestUpstreamLineTypes compares the line classification with
// TLineType.getFromLineInternal of the reference jar.
func TestUpstreamLineTypes(t *testing.T) {
	doc := loadUpstreamDoc(t)
	if len(doc.LineTypes) == 0 {
		t.Fatal("no line types")
	}
	for _, l := range doc.LineTypes {
		if got := getTLineType(l.S).String(); got != l.Type {
			t.Errorf("%q: %s, want %s", l.S, got, l.Type)
		}
	}
}

// knownDifferences are the cases where this implementation deliberately
// differs from the reference jar, with the reason.
var knownDifferences = map[string]string{
	"bi-splitstr-regex-diff-surrogate-empty": "Java's regex matches between the two chars of a surrogate pair and " +
		"splits into lone surrogates; Go strings hold no lone surrogate",
	"bi-splitstr-regex-diff-block":             "Unicode blocks (\\p{InX}) are not supported in regular expressions",
	"bi-splitstr-regex-diff-block-property":    "Unicode blocks (\\p{block=X}) are not supported in regular expressions",
	"bi-splitstr-regex-diff-grapheme-boundary": "\\b{g} (grapheme boundary) is not supported in regular expressions",
	"bi-splitstr-regex-diff-emoji":             "the Emoji properties are not supported in regular expressions (no Go table)",
	"bi-splitstr-regex-diff-script-alias":      "four-letter script aliases are not supported in regular expressions",
	"bi-splitstr-regex-diff-mirrored":          "\\p{javaMirrored} is not supported in regular expressions (no Go table)",
	"bi-splitstr-regex-diff-named-char":        "\\N{name} is not supported in regular expressions (no Unicode name table)",
	"edge-include-tilde": "the test reads from an fs.FS, where ~/ (the user's home directory) does not exist: " +
		"the include is not found instead of failing to read the file",
}

func loadUpstreamCases(t *testing.T) []upstreamCase { return loadUpstreamDoc(t).Cases }

// locationString formats a location like the harness: "desc:pos < parent".
func locationString(loc *text.LineLocation) string {
	if loc == nil {
		return "null"
	}
	s := loc.Desc + ":" + strconv.Itoa(loc.Position)
	if loc.Parent != nil {
		s += " < " + locationString(loc.Parent)
	}
	return s
}

func toUpstreamLines(lines []text.StringLocated) []upstreamLine {
	out := make([]upstreamLine, len(lines))
	for i, l := range lines {
		out[i] = upstreamLine{S: l.S, Loc: locationString(l.Location), Err: l.PreprocessorError}
	}
	return out
}

// runUpstreamCase preprocesses the case like upstream's BlockUmlBuilder: the
// fixture files are in an FS under /work.
func runUpstreamCase(c upstreamCase) (blocks []upstreamBlock, log []string) {
	// The case directory and the current directory exist, as in the harness.
	fsys := fstest.MapFS{"work": &fstest.MapFile{Mode: fs.ModeDir}}
	if c.Dir != "" {
		fsys["work/"+c.Dir] = &fstest.MapFile{Mode: fs.ModeDir}
	}
	for name, content := range c.Files {
		fsys["work/"+name] = &fstest.MapFile{Data: []byte(content)}
	}
	profile := security.Legacy
	if c.Profile != "" {
		profile, _ = security.ParseProfile(c.Profile)
	}
	config := security.Config{Profile: profile}
	if c.IncludePath != "" {
		config.IncludePaths = []string{"/work/" + c.IncludePath}
	}
	access := &security.Access{Config: config, FS: fsys}
	ctx := core.NewRenderContext()
	ctx.CurrentDir = "/work/" + c.Dir
	ctx.Defines = c.Defines
	defs := &BlockDefinitions{}
	for _, b := range core.BuildBlocks(c.Main, "string", nil) {
		res, _ := PreprocessWith(ctx, access, b.Lines, Options{
			Definitions: defs,
			// The reference image runs in UTC.
			Location: time.UTC,
			Getenv:   harnessGetenv(c),
		})
		defs.Add(res.Lines)
		blocks = append(blocks, upstreamBlock{
			Error:    res.Artifact != nil && len(res.Lines) > 0 && res.Lines[len(res.Lines)-1].PreprocessorError != "",
			Lines:    toUpstreamLines(res.Lines),
			Debug:    toUpstreamLines(res.Debug),
			Options:  res.Artifact.Options(),
			Warnings: res.Artifact.Warnings(),
		})
		log = append(log, res.Log...)
	}
	return blocks, log
}

// harnessGetenv looks names up like %getenv in the harness: the Java system
// properties that the harness sets or that depend on the platform, and no
// environment variable of interest.
func harnessGetenv(c upstreamCase) func(string) (string, bool) {
	return func(name string) (string, bool) {
		switch name {
		case "line.separator":
			return "\n", true
		case "path.separator":
			return ":", true
		case "PLANTUML_SECURITY_PROFILE":
			return c.Profile, c.Profile != ""
		}
		return "", false
	}
}

func TestUpstreamCases(t *testing.T) {
	for _, c := range loadUpstreamCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			if reason, ok := knownDifferences[c.Name]; ok {
				t.Skip(reason)
			}
			blocks, log := runUpstreamCase(c)
			if c.Crash != "" {
				// Upstream lets the exception escape and the render crashes;
				// this implementation reports "Fatal parsing error".
				if len(blocks) == 0 || !blocks[0].Error {
					t.Fatalf("upstream crashed with %s; want an error, got %+v", c.Crash, blocks)
				}
				return
			}
			if len(blocks) != len(c.Blocks) {
				t.Fatalf("got %d blocks, want %d", len(blocks), len(c.Blocks))
			}
			for i, want := range c.Blocks {
				got := blocks[i]
				compareLines(t, "lines", got.Lines, want.Lines)
				compareLines(t, "debug", got.Debug, want.Debug)
				if got.Error != want.Error {
					t.Errorf("block %d: error %v, want %v", i, got.Error, want.Error)
				}
				if len(want.Options) == 0 {
					want.Options = map[string]string{}
				}
				if !mapsEqual(got.Options, want.Options) {
					t.Errorf("block %d: options %v, want %v", i, got.Options, want.Options)
				}
				if !slices.Equal(got.Warnings, want.Warnings) && (len(got.Warnings) > 0 || len(want.Warnings) > 0) {
					t.Errorf("block %d: warnings %q, want %q", i, got.Warnings, want.Warnings)
				}
			}
			var wantLog []string
			for l := range strings.SplitSeq(c.Stderr, "\n") {
				if l != "" {
					wantLog = append(wantLog, l)
				}
			}
			if !slices.Equal(log, wantLog) && (len(log) > 0 || len(wantLog) > 0) {
				t.Errorf("log %q, want %q", log, wantLog)
			}
		})
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

func compareLines(t *testing.T, what string, got, want []upstreamLine) {
	t.Helper()
	if slices.Equal(got, want) {
		return
	}
	var sb strings.Builder
	for i := range max(len(got), len(want)) {
		var g, w upstreamLine
		if i < len(got) {
			g = got[i]
		}
		if i < len(want) {
			w = want[i]
		}
		mark := "  "
		if g != w {
			mark = "! "
		}
		sb.WriteString(mark + strconv.Quote(g.S) + " @" + g.Loc + " " + strconv.Quote(g.Err) + "\n")
		if g != w {
			sb.WriteString("  want " + strconv.Quote(w.S) + " @" + w.Loc + " " + strconv.Quote(w.Err) + "\n")
		}
	}
	t.Errorf("%s differ:\n%s", what, sb.String())
}
