// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type tree struct {
	t    *testing.T
	root string
}

func newTree(t *testing.T) tree { return tree{t, t.TempDir()} }

func (tr tree) write(rel string, data any) {
	tr.t.Helper()
	var b []byte
	switch v := data.(type) {
	case string:
		b = []byte(v)
	default:
		var err error
		if b, err = json.Marshal(v); err != nil {
			tr.t.Fatal(err)
		}
	}
	p := filepath.Join(tr.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		tr.t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		tr.t.Fatal(err)
	}
}

const svgHello = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="50"><text>Alice</text><text>hello</text></svg>`

func sequenceMeta(id string, w, h int, extra ...RenderMeta) Meta {
	file := "block0-page0.svg"
	renders := append([]RenderMeta{{File: &file, Width: w, Height: h}}, extra...)
	return Meta{ID: id, Blocks: []BlockMeta{{Type: "SEQUENCE", Pages: 1, Renders: renders}}}
}

func newRunner(ref, act tree, known map[string]KnownDifference) *Runner {
	return &Runner{
		Reference: func(id string) string { return filepath.Join(ref.root, filepath.FromSlash(id)) },
		Actual:    act.root,
		Known:     known,
	}
}

func TestCompareDiagramPass(t *testing.T) {
	ref, act := newTree(t), newTree(t)
	txt := "block0-page0.atxt"
	ref.write("seq/a/meta.json", sequenceMeta("seq/a", 100, 50, RenderMeta{File: &txt}))
	ref.write("seq/a/block0-page0.svg", svgHello)
	ref.write("seq/a/block0-page0.atxt", "A -> B\n")
	act.write("seq/a/meta.json", sequenceMeta("seq/a", 102, 49, RenderMeta{File: &txt}))
	act.write("seq/a/block0-page0.svg", strings.Replace(svgHello, "<text>Alice</text><text>hello</text>", "<text>hello</text><text>Alice</text>", 1))
	act.write("seq/a/block0-page0.atxt", "A -> B   \n")

	res := newRunner(ref, act, nil).CompareDiagram(&Case{ID: "seq/a", Formats: []string{"svg", "txt"}})
	if res.Status() != StatusPass || res.Group != "sequence" {
		t.Fatalf("status %s group %s failures %v", res.Status(), res.Group, allFailures(&res))
	}
}

func TestCompareDiagramFailures(t *testing.T) {
	tests := []struct {
		name    string
		act     func(tr tree)
		wantErr string
	}{
		{"missing output tree", func(tree) {}, "reading output"},
		{"wrong text", func(tr tree) {
			tr.write("seq/a/meta.json", sequenceMeta("seq/a", 100, 50))
			tr.write("seq/a/block0-page0.svg", strings.Replace(svgHello, "hello", "bye", 1))
		}, `missing texts "hello"`},
		{"too wide", func(tr tree) {
			tr.write("seq/a/meta.json", sequenceMeta("seq/a", 100, 50))
			tr.write("seq/a/block0-page0.svg", strings.Replace(svgHello, `width="100"`, `width="120"`, 1))
		}, "width 120 is not within 5%"},
		{"metadata missing", func(tr tree) {
			tr.write("seq/a/meta.json", sequenceMeta("seq/a", 100, 50))
			tr.write("seq/a/block0-page0.svg", strings.Replace(svgHello, "</svg>", "<?plantuml-src abc?></svg>", 1))
		}, "plantuml-src metadata presence differs"},
		{"missing file", func(tr tree) {
			tr.write("seq/a/meta.json", sequenceMeta("seq/a", 100, 50))
		}, "reading output"},
		{"unexpected error", func(tr tree) {
			m := sequenceMeta("seq/a", 100, 50)
			m.Blocks[0].Error = &ErrorMeta{Line: 2}
			tr.write("seq/a/meta.json", m)
		}, "but the diagram failed at line 2"},
		{"block count", func(tr tree) {
			tr.write("seq/a/meta.json", Meta{ID: "seq/a"})
		}, "0 block(s), reference has 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, act := newTree(t), newTree(t)
			ref.write("seq/a/meta.json", sequenceMeta("seq/a", 100, 50))
			ref.write("seq/a/block0-page0.svg", svgHello)
			tt.act(act)
			res := newRunner(ref, act, nil).CompareDiagram(&Case{ID: "seq/a", Formats: []string{"svg"}})
			if res.Status() != StatusFail {
				t.Fatalf("status %s, want fail", res.Status())
			}
			if got := strings.Join(allFailures(&res), "; "); !strings.Contains(got, tt.wantErr) {
				t.Fatalf("failures %q do not contain %q", got, tt.wantErr)
			}
		})
	}
}

func TestCompareDiagramErrorLineAndCrashReport(t *testing.T) {
	ref, act := newTree(t), newTree(t)
	errMeta := Meta{ID: "e", Blocks: []BlockMeta{{Type: "ERROR", Pages: 1, Error: &ErrorMeta{Line: 3}}}}
	ref.write("e/meta.json", errMeta)
	act.write("e/meta.json", errMeta)
	if res := newRunner(ref, act, nil).CompareDiagram(&Case{ID: "e", Type: "sequence", Formats: []string{"svg"}}); res.Status() != StatusPass || res.Group != "sequence" {
		t.Fatalf("same error line: status %s group %s %v", res.Status(), res.Group, allFailures(&res))
	}

	crash := Meta{ID: "c", Blocks: []BlockMeta{{Type: "BPM", Pages: 1, Renders: []RenderMeta{{CrashReport: true, Status: 503}}}}}
	ref.write("c/meta.json", crash)
	act.write("c/meta.json", sequenceMeta("c", 10, 10))
	if res := newRunner(ref, act, nil).CompareDiagram(&Case{ID: "c", Formats: []string{"svg"}}); res.Status() != StatusFail {
		t.Fatal("a rendered diagram must not match a reference crash report")
	}
	act.write("c/meta.json", crash)
	if res := newRunner(ref, act, nil).CompareDiagram(&Case{ID: "c", Formats: []string{"svg"}}); res.Status() != StatusPass {
		t.Fatalf("crash report vs crash: %v", allFailures(&res))
	}
}

func TestCompareDiagramPreprocAndKnown(t *testing.T) {
	ref, act := newTree(t), newTree(t)
	m := Meta{ID: "p", Blocks: []BlockMeta{{Type: "SEQUENCE", Pages: 1}}}
	ref.write("p/meta.json", m)
	act.write("p/meta.json", m)
	ref.write("p/block0.preproc", "@startuml\nA -> B\n@enduml\n")
	act.write("p/block0.preproc", "@startuml\nA -> C\n@enduml\n")
	known := map[string]KnownDifference{"p": {ID: "p", Reason: "not yet", ResolvedBy: "later"}}
	res := newRunner(ref, act, known).CompareDiagram(&Case{ID: "p", Formats: []string{"preproc"}})
	if res.Status() != StatusKnown || res.FormatPassed("preproc") {
		t.Fatalf("status %s", res.Status())
	}
	act.write("p/block0.preproc", "@startuml\nA -> B\n@enduml\n")
	if res := newRunner(ref, act, known).CompareDiagram(&Case{ID: "p", Formats: []string{"preproc"}}); res.Status() != StatusKnownPass {
		t.Fatalf("status %s, want known-pass", res.Status())
	}
}

func TestCompareCLI(t *testing.T) {
	ref, act := newTree(t), newTree(t)
	rec := CLIRecord{Exit: 0, Files: []FileRecord{{Path: "a.svg", Stored: true}}}
	ref.write("cli/x/cli.json", rec)
	ref.write("cli/x/files/a.svg", svgHello)
	act.write("cli/x/cli.json", rec)
	act.write("cli/x/files/a.svg", svgHello)
	c := &CLICase{ID: "cli/x", Compare: []string{"exit", "files"}}
	if res := newRunner(ref, act, nil).CompareCLI(c); res.Status() != StatusPass {
		t.Fatalf("pass case: %v", allFailures(&res))
	}

	act.write("cli/x/cli.json", CLIRecord{Exit: 200, Files: []FileRecord{{Path: "b.svg"}}})
	res := newRunner(ref, act, nil).CompareCLI(c)
	got := strings.Join(allFailures(&res), "; ")
	for _, want := range []string{"exit status 200, reference 0", `created files ["b.svg"], reference ["a.svg"]`} {
		if !strings.Contains(got, want) {
			t.Errorf("failures %q do not contain %q", got, want)
		}
	}
}

func TestCompareHTTP(t *testing.T) {
	ref, act := newTree(t), newTree(t)
	rec := HTTPRecord{Status: 200, Headers: map[string]string{
		"Content-Type":             "image/svg+xml",
		"X-Plantuml-Diagram-Width": "100",
	}, Body: FileRecord{Path: "body.svg", Stored: true}}
	ref.write("h/x/http.json", rec)
	ref.write("h/x/body.svg", svgHello)
	rec.Headers = map[string]string{"Content-Type": "image/svg+xml", "X-Plantuml-Diagram-Width": "103"}
	act.write("h/x/http.json", rec)
	act.write("h/x/body.svg", svgHello)
	c := &HTTPCase{ID: "h/x", Compare: []string{"status", "headers", "body"}}
	if res := newRunner(ref, act, nil).CompareHTTP(c); res.Status() != StatusPass {
		t.Fatalf("pass case: %v", allFailures(&res))
	}
	rec.Status, rec.Headers["Content-Type"] = 400, "text/plain"
	act.write("h/x/http.json", rec)
	res := newRunner(ref, act, nil).CompareHTTP(c)
	got := strings.Join(allFailures(&res), "; ")
	for _, want := range []string{"status 400, reference 200", `header Content-Type "text/plain"`} {
		if !strings.Contains(got, want) {
			t.Errorf("failures %q do not contain %q", got, want)
		}
	}
}

func TestLoadKnownDifferences(t *testing.T) {
	tr := newTree(t)
	if m, err := LoadKnownDifferences(filepath.Join(tr.root, "missing.yaml")); err != nil || len(m) != 0 {
		t.Fatalf("missing file: %v %v", m, err)
	}
	tr.write("k.yaml", "- id: a\n  reason: puma engine\n  resolvedBy: add-puma\n")
	m, err := LoadKnownDifferences(filepath.Join(tr.root, "k.yaml"))
	if err != nil || m["a"].ResolvedBy != "add-puma" {
		t.Fatalf("%v %v", m, err)
	}
	tr.write("bad.yaml", "- id: a\n")
	if _, err := LoadKnownDifferences(filepath.Join(tr.root, "bad.yaml")); err == nil {
		t.Fatal("entry without reason must be rejected")
	}
	tr.write("dup.yaml", "- {id: a, reason: x}\n- {id: a, reason: y}\n")
	if _, err := LoadKnownDifferences(filepath.Join(tr.root, "dup.yaml")); err == nil {
		t.Fatal("duplicate entry must be rejected")
	}
}

func TestWriteReport(t *testing.T) {
	results := []CaseResult{
		{ID: "s1", Group: "sequence", Formats: []string{"svg"}},
		{ID: "s2", Group: "sequence", Formats: []string{"svg", "txt"}, FormatFailures: map[string][]string{"txt": {"line 1 differs"}}},
		{ID: "c1", Group: "class", Formats: []string{"svg"}, Failures: []string{"boom"}, Known: &KnownDifference{ID: "c1", Reason: "needs smetana", ResolvedBy: "add-smetana"}},
		{ID: "c2", Group: "class", Formats: []string{"svg"}, Known: &KnownDifference{ID: "c2", Reason: "old"}},
	}
	groups, formats := Summarize(results)
	if len(groups) != 2 || groups[1].Name != "sequence" || groups[1].Passed != 1 || groups[1].Failed != 1 {
		t.Fatalf("groups = %+v", groups)
	}
	if formats[0].Name != "svg" || formats[0].Cases != 4 || formats[0].Passed != 3 || formats[0].Known != 1 {
		t.Fatalf("formats = %+v", formats)
	}
	var b bytes.Buffer
	if err := WriteReport(&b, "java", results); err != nil {
		t.Fatal(err)
	}
	report := b.String()
	for _, want := range []string{
		"Implementation under test: java",
		"| sequence | 2 | 1 | 0 | 1 | 0 | 50.0% |",
		"- `s2`: txt: line 1 differs",
		"- `c1`: needs smetana (resolved by add-smetana)",
		"## Known differences that now pass (1)",
		"- `c2`",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}

func TestReportsUnsupported(t *testing.T) {
	notYet := func(msg string) *BlockMeta {
		return &BlockMeta{Type: "ERROR", Error: &ErrorMeta{Line: 1, Messages: []string{msg}}}
	}
	for _, tt := range []struct {
		name     string
		ref, act *BlockMeta
		want     bool
	}{
		{"state reported", &BlockMeta{Type: "STATE"}, notYet("state diagrams are not supported yet"), true},
		{"mindmap reported", &BlockMeta{Type: "MINDMAP"}, notYet("mindmap diagrams are not supported yet"), true},
		{"unknown directive", &BlockMeta{Type: "GANTT"}, &BlockMeta{Type: "OTHER"}, true},
		{"legacy activity", &BlockMeta{Type: "ACTIVITY"}, notYet("legacy activity diagrams are not supported yet"), true},
		{"activity v3 must render", &BlockMeta{Type: "ACTIVITY"}, notYet("state diagrams are not supported yet"), false},
		{"sequence must render", &BlockMeta{Type: "SEQUENCE"}, notYet("timing diagrams are not supported yet"), false},
		{"class must render", &BlockMeta{Type: "CLASS"}, &BlockMeta{Type: "OTHER"}, false},
		{"wrong output", &BlockMeta{Type: "STATE"}, &BlockMeta{Type: "SEQUENCE"}, false},
		{"syntax error is not unsupported", &BlockMeta{Type: "STATE"}, notYet("Syntax Error?"), false},
		{"reference error compares", &BlockMeta{Type: "ERROR", Error: &ErrorMeta{Line: 2}}, notYet("state diagrams are not supported yet"), false},
	} {
		if got := reportsUnsupported(tt.ref, tt.act, ""); got != tt.want {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestReportsUnsupportedReferenceError(t *testing.T) {
	ref := &BlockMeta{Type: "ERROR", Error: &ErrorMeta{Line: 3, Messages: []string{"Bad range"}}}
	act := &BlockMeta{Type: "ERROR", Error: &ErrorMeta{Line: 1, Messages: []string{"packetdiag diagrams are not supported yet"}}}
	if !reportsUnsupported(ref, act, "packet") {
		t.Error("an upstream packet error with a not-supported-yet error is not counted")
	}
	if reportsUnsupported(ref, act, "sequence") {
		t.Error("a sequence case must match the reference error")
	}
}
