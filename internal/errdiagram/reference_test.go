// SPDX-License-Identifier: MIT
// Provenance: original

package errdiagram_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/conformance"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/errdiagram"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/svg/svgtest"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/text"
)

const repoRoot = "../.."

// notParsedYet are the error cases whose diagram type is not parsed yet,
// so that this implementation draws its "not supported yet" image instead.
var notParsedYet = map[string]bool{
	"gantt-diagram/no-matching-verb-phrase": true,
	"gantt-diagram/syntax-error":            true,
	"gantt-diagram/no-starting-date":        true,
	"gantt-diagram/late-project-start":      true,
	"gantt-diagram/unknown-task":            true,
	"gantt-diagram/no-group-to-close":       true,
	"gantt-diagram/twenty-years":            true,
	"gantt-diagram/invalid-date":            true,
}

// handled reports whether a reference block is an image of this package.
func handled(b conformance.BlockMeta) bool {
	return b.Type == "ERROR" || b.Description == "(Unsupported)" || b.Description == "(Empty)"
}

// diagramOf is the diagram of a block as the engine builds it once
// WelcomeFactory goes first in its factory list.
func diagramOf(rc *core.RenderContext, b *engine.Block) core.Diagram {
	if d, ok := b.Diagram.(*core.ErrorDiagram); ok {
		if n := len(d.Trace); n > 0 && d.Trace[n-1].PreprocessorError != "" {
			return d
		}
	}
	factories := append([]core.Factory{errdiagram.WelcomeFactory{}}, engine.Factories...)
	return core.CreateDiagram(rc, b.Data, b.Raw, factories)
}

// result counts the compared images.
type result struct {
	// identical images are byte-identical; exceptMetadata images are
	// byte-identical but for the encoding of the plantuml-src metadata,
	// which decodes to the same text (Go's compress/flate differs from
	// Java's zlib, design D10).
	identical, exceptMetadata int
}

func (r result) String() string {
	return fmt.Sprintf("%d byte-identical, %d byte-identical but for the metadata encoding", r.identical,
		r.exceptMetadata)
}

var metadataPI = regexp.MustCompile(`<\?plantuml-src [^?]*\?>`)

// compare checks an image against its reference: byte for byte, or byte
// for byte but for the metadata, which must decode to the same text
// (svgtest.Normalize).
func (r *result) compare(t *testing.T, name string, got []byte, data export.ImageData, refFile string,
	ref conformance.RenderMeta) {
	t.Helper()
	want, err := os.ReadFile(refFile)
	if err != nil {
		t.Errorf("%s: %v", name, err)
		return
	}
	if data.Width() != ref.Width || data.Height() != ref.Height || data.Status() != ref.Status {
		t.Errorf("%s: image data %dx%d status %d, reference %dx%d status %d", name, data.Width(), data.Height(),
			data.Status(), ref.Width, ref.Height, ref.Status)
	}
	if bytes.Equal(got, want) {
		r.identical++
		return
	}
	g, w := svgtest.Normalize(string(got)), svgtest.Normalize(string(want))
	if g != w {
		i := firstDiff(g, w)
		t.Errorf("%s: SVG differs at byte %d\n got: %s\nwant: %s", name, i, around(g, i), around(w, i))
		return
	}
	if !bytes.Equal(metadataPI.ReplaceAll(got, nil), metadataPI.ReplaceAll(want, nil)) {
		t.Errorf("%s: SVG differs outside the metadata, but not once normalized (PNG data)", name)
		return
	}
	r.exceptMetadata++
}

func firstDiff(a, b string) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func around(s string, i int) string { return s[max(0, i-200):min(len(s), i+200)] }

func readMeta(t *testing.T, file string) conformance.Meta {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var meta conformance.Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatal(err)
	}
	return meta
}

func exportSVG(t *testing.T, d core.Diagram, b *engine.Block) ([]byte, export.ImageData) {
	t.Helper()
	var buf bytes.Buffer
	data, err := errdiagram.Export(&buf, d, errdiagram.Source{Data: b.Data, Raw: b.Raw},
		export.NewFileFormatOption(font.FormatSVG))
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.Bytes(), data
}

func parse(t *testing.T, rc *core.RenderContext, src engine.Source) []*engine.Block {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, src)
	if err != nil {
		t.Fatal(err)
	}
	return blocks
}

// TestDiagramReferences draws the error, unsupported and welcome images of
// the diagram cases (testdata/conformance/cases/**/*.puml) and compares them
// with the images of the reference jar. The jar reads them with a
// SourceStringReader, whose line locations are described as "string", in the
// directory of the case.
func TestDiagramReferences(t *testing.T) {
	cases, err := conformance.LoadCases(os.DirFS(filepath.Join(repoRoot, "testdata/conformance/cases")), ".")
	if err != nil {
		t.Fatal(err)
	}
	var r result
	for _, c := range cases {
		if notParsedYet[c.ID] {
			continue
		}
		refDir := filepath.Join(repoRoot, "testdata/conformance/reference", c.ID)
		if _, err := os.Stat(filepath.Join(refDir, "meta.json")); err != nil {
			continue
		}
		meta := readMeta(t, filepath.Join(refDir, "meta.json"))
		var blocks []*engine.Block
		rc := core.NewRenderContext()
		rc.CurrentDir = filepath.Join(repoRoot, "testdata/conformance/cases", filepath.Dir(c.Path))
		// The fixtures of a case are staged next to it by refgen
		fixtures := filepath.Join(repoRoot, "testdata/conformance/cases", filepath.FromSlash(c.ID)+".files")
		if st, err := os.Stat(fixtures); err == nil && st.IsDir() {
			rc.CurrentDir = fixtures
		}
		for _, bm := range meta.Blocks {
			if !handled(bm) {
				continue
			}
			if blocks == nil {
				blocks = parse(t, rc, engine.Source{Text: c.Source})
			}
			if bm.Index >= len(blocks) {
				t.Errorf("%s: block %d missing", c.ID, bm.Index)
				continue
			}
			b := blocks[bm.Index]
			d := diagramOf(rc, b)
			if got := d.Description(); got != bm.Description {
				t.Errorf("%s block %d: description %q, reference %q", c.ID, bm.Index, got, bm.Description)
				continue
			}
			for _, rm := range bm.Renders {
				if rm.File == nil || !strings.HasSuffix(*rm.File, ".svg") {
					continue
				}
				got, data := exportSVG(t, d, b)
				r.compare(t, c.ID+"/"+*rm.File, got, data, filepath.Join(refDir, *rm.File), rm)
			}
		}
	}
	t.Logf("diagram references: %v", r)
	if r.identical+r.exceptMetadata < 8 {
		t.Errorf("only %d images compared", r.identical+r.exceptMetadata)
	}
}

// TestCLIReferences draws the error images that the command line cases write
// (testdata/conformance/reference/cli and the CLI cases of the other
// capabilities), as the command line reads a file: line locations described
// by the file name, the security profile of the environment.
func TestCLIReferences(t *testing.T) {
	cases, err := conformance.LoadCLICases(os.DirFS(filepath.Join(repoRoot, "testdata/conformance/cases")), ".")
	if err != nil {
		t.Fatal(err)
	}
	var r result
	for _, c := range cases {
		refDir := filepath.Join(repoRoot, "testdata/conformance/reference", c.ID, "files")
		for name, content := range c.Files {
			if !strings.HasSuffix(name, ".puml") {
				continue
			}
			refFile := filepath.Join(refDir, strings.TrimSuffix(name, ".puml")+".svg")
			want, err := os.ReadFile(refFile)
			if err != nil || !bytes.Contains(want, []byte(`fill="#33FF02"`)) {
				continue
			}
			dir := t.TempDir()
			for n, s := range c.Files {
				if err := os.WriteFile(filepath.Join(dir, n), []byte(s), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			rc := core.NewRenderContext()
			cfg := security.ConfigFromEnv(func(k string) string { return c.Env[k] }, security.Legacy)
			rc.SecurityProfile, rc.Security = cfg.Profile.String(), &cfg
			file := filepath.Join(dir, name)
			rc.Desc = name
			rc.CurrentDir = dir
			blocks := parse(t, rc, engine.Source{Text: content, File: file})
			if len(blocks) != 1 {
				t.Errorf("%s: %d blocks", c.ID, len(blocks))
				continue
			}
			d := diagramOf(rc, blocks[0])
			got, data := exportSVG(t, d, blocks[0])
			// The CLI records no image data: check the status only.
			ref := conformance.RenderMeta{Width: data.Width(), Height: data.Height(), Status: errdiagram.StatusError}
			r.compare(t, c.ID+"/"+name, got, data, refFile, ref)
		}
	}
	t.Logf("CLI references: %v", r)
	if r.identical+r.exceptMetadata < 6 {
		t.Errorf("only %d images compared", r.identical+r.exceptMetadata)
	}
}

// TestHTTPReferences draws the error bodies of the builtin server cases: the
// source of the URL, described as "string", with the INTERNET profile. The
// X-PlantUML-Diagram-Width and Height headers are upstream's image data.
func TestHTTPReferences(t *testing.T) {
	cases, err := conformance.LoadHTTPCases(os.DirFS(filepath.Join(repoRoot, "testdata/conformance/cases")), ".")
	if err != nil {
		t.Fatal(err)
	}
	var r result
	for _, c := range cases {
		refDir := filepath.Join(repoRoot, "testdata/conformance/reference", c.ID)
		var rec conformance.HTTPRecord
		data, err := os.ReadFile(filepath.Join(refDir, "http.json"))
		if err != nil || json.Unmarshal(data, &rec) != nil {
			continue
		}
		if rec.Server != "builtin" || rec.Headers["X-Plantuml-Diagram-Description"] != "(Error)" || c.Source == "" {
			continue
		}
		rc := core.NewRenderContext()
		rc.SecurityProfile = "INTERNET"
		blocks := parse(t, rc, engine.Source{Text: c.Source})
		d := diagramOf(rc, blocks[0])
		got, img := exportSVG(t, d, blocks[0])
		var ref conformance.RenderMeta
		fmt.Sscan(rec.Headers["X-Plantuml-Diagram-Width"], &ref.Width)
		fmt.Sscan(rec.Headers["X-Plantuml-Diagram-Height"], &ref.Height)
		ref.Status = rec.Status
		r.compare(t, c.ID, got, img, filepath.Join(refDir, rec.Body.Path), ref)
	}
	t.Logf("HTTP references: %v", r)
	if r.identical+r.exceptMetadata < 2 {
		t.Errorf("only %d images compared", r.identical+r.exceptMetadata)
	}
}

// TestVegaReferences draws the error images of upstream's Vega suite
// (.cache/vega and .cache/reference/vega, skipped when absent). Their diagram types are not
// parsed by this implementation, so each error is built as upstream reports
// it: the lines up to the reference error line, and the reference message.
func TestVegaReferences(t *testing.T) {
	vegaRoot := filepath.Join(repoRoot, ".cache/vega")
	refRoot := filepath.Join(repoRoot, ".cache/reference/vega")
	for _, dir := range []string{vegaRoot, refRoot} {
		if _, err := os.Stat(dir); err != nil {
			t.Skip("no Vega suite or references")
		}
	}
	cases, err := conformance.LoadVegaCases(os.DirFS(vegaRoot), ".")
	if err != nil {
		t.Fatal(err)
	}
	var r result
	for _, c := range cases {
		id := strings.TrimPrefix(c.ID, conformance.VegaPrefix)
		refDir := filepath.Join(refRoot, id)
		if _, err := os.Stat(filepath.Join(refDir, "meta.json")); err != nil {
			continue
		}
		meta := readMeta(t, filepath.Join(refDir, "meta.json"))
		for _, bm := range meta.Blocks {
			if bm.Type != "ERROR" || len(bm.Renders) == 0 {
				continue
			}
			rc := core.NewRenderContext()
			// The includes of a case are relative to its directory
			rc.CurrentDir = filepath.Join(vegaRoot, filepath.Dir(id))
			blocks := parse(t, rc, engine.Source{Text: c.Source})
			b := blocks[bm.Index]
			var trace []text.StringLocated
			for _, l := range b.Data {
				trace = append(trace, l)
				if l.Location.Position == bm.Error.Line-1 {
					break
				}
			}
			last := trace[len(trace)-1]
			d := &core.ErrorDiagram{Source: b.Data, Trace: trace,
				Error: core.ErrorUml{Type: core.SyntaxError, Error: bm.Error.Messages[0], Line: &last}}
			// An error of the preprocessor is drawn from its own trace,
			// which keeps the directive lines that Data drops.
			if e, ok := b.Diagram.(*core.ErrorDiagram); ok && e.Error.Error == bm.Error.Messages[0] {
				d = e
			}
			for _, rm := range bm.Renders {
				if rm.File == nil {
					continue
				}
				got, data := exportSVG(t, d, b)
				r.compare(t, c.ID+"/"+*rm.File, got, data, filepath.Join(refDir, *rm.File), rm)
			}
		}
	}
	t.Logf("Vega references: %v", r)
	if r.identical+r.exceptMetadata < 6 {
		t.Errorf("only %d images compared", r.identical+r.exceptMetadata)
	}
}
