// SPDX-License-Identifier: MIT
// Provenance: original

package svek_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/classdiagram"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/svek"
)

// The golden files of testdata/golden are written by
// testdata/javagen/SvekGolden.java with the upstream jar and Graphviz 14.0.1
// in the reference image (instructions in its header): the jar's SVG, the
// DOT graph svek gave to dot, and dot's SVG output. The NAME.dot15.svg files
// are the output of Graphviz 15.0.0 for the same DOT graphs, and
// NAME.go15.svg the drawing of this package with them when it differs from
// the jar's drawing with dot 14.0.1 (-update rewrites these two).

var update = flag.Bool("update", false, "rewrite the dot 15 outputs and drawings")

// savedEngine is a layout engine that returns a saved dot output, after
// checking that it is given the DOT graph the output was made for.
type savedEngine struct {
	t        *testing.T
	wantDot  string
	svg      []byte
	gotDot   string
	called   bool
	checkDot bool
}

func (e *savedEngine) Layout(_ context.Context, dot string) ([]byte, error) {
	e.called = true
	e.gotDot = dot
	if e.checkDot && dot != e.wantDot {
		return nil, errors.New("unexpected DOT graph")
	}
	return e.svg, nil
}

var plantumlSrc = regexp.MustCompile(`<\?plantuml-src [^?]*\?>`)

func parseClassDiagram(t *testing.T, file string) *classdiagram.ClassDiagram {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	blocks, err := engine.Parse(context.Background(), core.NewRenderContext(), engine.Source{Text: string(data)})
	if err != nil {
		t.Fatal(err)
	}
	d, ok := blocks[0].Diagram.(*classdiagram.ClassDiagram)
	if !ok {
		t.Fatalf("%s: not a class diagram: %T", file, blocks[0].Diagram)
	}
	return d
}

func renderWith(t *testing.T, file string, eng *savedEngine) string {
	t.Helper()
	d := parseClassDiagram(t, file)
	var buf bytes.Buffer
	fo := export.NewFileFormatOptionMetadata(font.FormatSVG, true)
	if _, err := d.ExportDiagramLayout(context.Background(), &buf, 0, fo, svek.Layout{Engine: eng}); err != nil {
		t.Fatalf("export: %v", err)
	}
	return plantumlSrc.ReplaceAllString(buf.String(), "")
}

func goldenCases(t *testing.T) []string {
	t.Helper()
	cases, err := filepath.Glob("testdata/cases/*.puml")
	if err != nil || len(cases) == 0 {
		t.Fatal("no cases", err)
	}
	return cases
}

// TestGoldenDotAndDrawing checks, for each case, the DOT graph against the
// one upstream gave to dot, and the drawing made with dot 14.0.1's saved
// output against the jar's SVG (task 12.2 and 12.3).
func TestGoldenDotAndDrawing(t *testing.T) {
	for _, file := range goldenCases(t) {
		name := strings.TrimSuffix(filepath.Base(file), ".puml")
		t.Run(name, func(t *testing.T) {
			golden := filepath.Join("testdata", "golden", name)
			want, err := os.ReadFile(golden + ".svg")
			if err != nil {
				t.Fatal(err)
			}
			wantDot, errDot := os.ReadFile(golden + ".dot")
			dotSvg, _ := os.ReadFile(golden + ".dot14.svg")
			eng := &savedEngine{t: t, wantDot: string(wantDot), svg: dotSvg}
			if errDot != nil {
				if name == "smetana" {
					t.Skip("upstream lays this case out with Smetana (design D16)")
				}
			}
			got := renderWith(t, file, eng)
			if errDot == nil {
				if !eng.called {
					t.Fatal("the layout engine was not called")
				}
				if eng.gotDot != string(wantDot) {
					t.Errorf("DOT graph differs:\n%s", firstDiff(string(wantDot), eng.gotDot))
				}
			} else if eng.called {
				t.Errorf("upstream draws this case without dot, the engine was called")
			}
			if w := plantumlSrc.ReplaceAllString(string(want), ""); got != w {
				t.Errorf("SVG differs:\n%s", firstDiff(w, got))
			}
		})
	}
}

// firstDiff shows the first difference of two texts, split at tags.
func firstDiff(want, got string) string {
	w := strings.Split(strings.ReplaceAll(want, "><", ">\n<"), "\n")
	g := strings.Split(strings.ReplaceAll(got, "><", ">\n<"), "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return "line " + itoa(i+1) + ":\nwant: " + wl + "\ngot:  " + gl
		}
	}
	return "(no difference)"
}

func itoa(i int) string {
	var b []byte
	if i == 0 {
		return "0"
	}
	for ; i > 0; i /= 10 {
		b = append([]byte{byte('0' + i%10)}, b...)
	}
	return string(b)
}

var textElement = regexp.MustCompile(`<text[^>]*>[^<]*</text>`)

// texts returns the sorted contents of the <text> elements.
func texts(svg string) []string {
	var result []string
	for _, m := range textElement.FindAllString(svg, -1) {
		result = append(result, m[strings.Index(m, ">")+1:strings.LastIndex(m, "<")])
	}
	sort.Strings(result)
	return result
}

// TestGoldenDot15 lays the cases out with the saved output of Graphviz
// 15.0.0 (task 12.3: the parser against the latest release): the drawing
// must match its golden file, and show the same texts as upstream's.
func TestGoldenDot15(t *testing.T) {
	for _, file := range goldenCases(t) {
		name := strings.TrimSuffix(filepath.Base(file), ".puml")
		golden := filepath.Join("testdata", "golden", name)
		wantDot, err := os.ReadFile(golden + ".dot")
		if err != nil {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if *update {
				out, err := exec.Command("dot", "-V").CombinedOutput()
				if err != nil || !strings.Contains(string(out), "version 15.0.0") {
					t.Fatalf("-update needs Graphviz 15.0.0 on PATH: %s %v", out, err)
				}
				cmd := exec.Command("dot", "-Tsvg")
				cmd.Stdin = bytes.NewReader(wantDot)
				svg, err := cmd.Output()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden+".dot15.svg", svg, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			dotSvg, err := os.ReadFile(golden + ".dot15.svg")
			if err != nil {
				t.Fatal(err)
			}
			eng := &savedEngine{t: t, wantDot: string(wantDot), svg: dotSvg, checkDot: true}
			got := renderWith(t, file, eng)
			jar, err := os.ReadFile(golden + ".svg")
			if err != nil {
				t.Fatal(err)
			}
			jarSvg := plantumlSrc.ReplaceAllString(string(jar), "")
			// NAME.go15.svg exists only when dot 15 lays the case out
			// differently from dot 14, whose drawing is the jar's.
			if *update {
				if got != jarSvg {
					if err := os.WriteFile(golden+".go15.svg", []byte(got), 0o644); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Remove(golden + ".go15.svg"); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			want := jarSvg
			if w, err := os.ReadFile(golden + ".go15.svg"); err == nil {
				want = string(w)
			}
			if got != want {
				t.Errorf("SVG differs:\n%s", firstDiff(want, got))
			}
			if g, j := texts(got), texts(string(jar)); strings.Join(g, "\n") != strings.Join(j, "\n") {
				t.Errorf("texts differ from upstream's: %q, want %q", g, j)
			}
		})
	}
}
