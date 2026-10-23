// SPDX-License-Identifier: MIT
// Provenance: original

//go:build !js

package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runNode runs an ES module script in Node.js from dir and decodes the JSON
// it prints into v. The script imports the playground of the landing page
// as PLAYGROUND.
func runNode(t *testing.T, dir, script string, v any) {
	t.Helper()
	playground, err := filepath.Abs(filepath.Join("..", "..", "site", "playground.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	script = strings.ReplaceAll(script, "PLAYGROUND", "file://"+filepath.ToSlash(playground))
	name := filepath.Join(dir, "playground-run.mjs")
	if err := os.WriteFile(name, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", name)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("node: %v\n%s", err, stderr)
	}
	if err := json.Unmarshal(out, v); err != nil {
		t.Fatalf("output %q: %v", out, err)
	}
}

// playgroundResult is a render result of the playground, with the start of
// its SVG.
type playgroundResult struct {
	SVG   string `json:"svg"`
	Type  string `json:"type"`
	Error *struct {
		Kind    string `json:"kind"`
		Message string `json:"message"`
		Line    int    `json:"line"`
		Text    string `json:"text"`
	} `json:"error"`
}

func TestPlayground(t *testing.T) {
	// Spec scenarios "Sequence diagram", "Syntax error", "Graph-based
	// diagram", "Example diagram" and "Download" of landing-page, with the
	// module of this build. The DOM part is checked by hand in a browser.
	dir := build(t)
	script := `import { load } from './plantuml.mjs';
import { createPlayground, examples, download } from 'PLAYGROUND';
const pg = createPlayground(load);
const show = async (r) => ({
  svg: r.svg ? (await r.svg.text()).slice(0, 5) : '',
  type: r.svg ? r.svg.type : '',
  error: r.error,
});
const out = { examples: {} };
for (const e of examples) {
  out.examples[e.name] = await show(await pg.render(e.source));
}
const sequence = await pg.render('@startuml\nAlice -> Bob : hello\n@enduml\n');
out.sequence = await show(sequence);
const syntax = await pg.render('@startuml\nAlice -> Bob : hello\nthis is not a valid line ->>>\n@enduml\n');
out.syntax = await show(syntax);
out.graph = await show(await pg.render('@startuml\nclass A\nclass B\nA --> B\n@enduml\n'));
const file = download(sequence);
out.download = { filename: file.filename, type: file.blob.type, svg: (await file.blob.text()).slice(0, 5) };
out.downloadError = download(syntax);
out.status = pg.status;
console.log(JSON.stringify(out));
process.exit(0);
`
	var out struct {
		Examples      map[string]playgroundResult `json:"examples"`
		Sequence      playgroundResult            `json:"sequence"`
		Syntax        playgroundResult            `json:"syntax"`
		Graph         playgroundResult            `json:"graph"`
		Download      struct{ Filename, Type, SVG string }
		DownloadError any    `json:"downloadError"`
		Status        string `json:"status"`
	}
	runNode(t, dir, script, &out)

	if len(out.Examples) < 2 {
		t.Errorf("%d examples, want sequence and activity examples", len(out.Examples))
	}
	for name, r := range out.Examples {
		if r.Error != nil || r.SVG != "<svg " {
			t.Errorf("example %q: error %+v, svg %q", name, r.Error, r.SVG)
		}
	}
	if r := out.Sequence; r.Error != nil || r.SVG != "<svg " || r.Type != "image/svg+xml" {
		t.Errorf("sequence: error %+v, svg %q, type %q", r.Error, r.SVG, r.Type)
	}
	if r := out.Syntax; r.Error == nil || r.Error.Line != 3 || !strings.Contains(r.Error.Text, "line 3") || r.SVG != "<svg " {
		t.Errorf("syntax error: error %+v, error image %q", r.Error, r.SVG)
	}
	if r := out.Graph; r.Error == nil || r.Error.Kind != "graph-layout-unavailable" || r.Error.Text != "graph layout is not available in this build" || r.SVG != "<svg " {
		t.Errorf("class diagram: error %+v, error image %q", r.Error, r.SVG)
	}
	if d := out.Download; d.Filename != "diagram.svg" || d.Type != "image/svg+xml" || d.SVG != "<svg " {
		t.Errorf("download: %+v", d)
	}
	if out.DownloadError != nil {
		t.Errorf("download of an error: %v, want nothing", out.DownloadError)
	}
	if out.Status != "ready" {
		t.Errorf("status %q, want ready", out.Status)
	}
}

func TestPlaygroundLoading(t *testing.T) {
	// Spec scenarios "Reading the top of the page" and "Module not
	// available" of landing-page, with fake loaders: nothing is loaded
	// before the playground is activated, a failed load is reported, and
	// only the latest render is kept.
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	script := `import { createPlayground, describe } from 'PLAYGROUND';
const out = {};
let loads = 0;
const fake = {
  render: (source) => new Promise((resolve) => setTimeout(() => resolve({
    data: new TextEncoder().encode('<svg>' + source + '</svg>'), error: null, width: 1, height: 1,
  }), source === 'slow' ? 50 : 0)),
};
const pg = createPlayground(async () => { loads++; return fake; });
out.before = { loads, status: pg.status };
const first = pg.activate();
out.loading = pg.status;
await Promise.all([first, pg.activate()]);
out.after = { loads, status: pg.status };
const [slow, fast] = await Promise.all([pg.render('slow'), pg.render('fast')]);
out.slow = slow;
out.fast = fast && await fast.svg.text();

const failing = createPlayground(async () => { throw new Error('404 plantuml.wasm'); });
out.failed = await failing.activate().then(() => 'loaded', (e) => e.message);
out.failedStatus = failing.status;
out.failedRender = await failing.render('x').then(() => 'rendered', (e) => e.message);

out.describe = [
  describe({ message: 'plantuml: line 3: Syntax Error?', line: 3 }),
  describe({ message: 'Syntax Error?', line: 2 }),
  describe({ message: 'graph layout is not available in this build' }),
];
console.log(JSON.stringify(out));
process.exit(0);
`
	var out struct {
		Before struct {
			Loads  int
			Status string
		}
		Loading string
		After   struct {
			Loads  int
			Status string
		}
		Slow         any
		Fast         string
		Failed       string
		FailedStatus string   `json:"failedStatus"`
		FailedRender string   `json:"failedRender"`
		Describe     []string `json:"describe"`
	}
	runNode(t, t.TempDir(), script, &out)

	if out.Before.Loads != 0 || out.Before.Status != "idle" {
		t.Errorf("before activation: %+v, want no load", out.Before)
	}
	if out.Loading != "loading" {
		t.Errorf("status while loading: %q", out.Loading)
	}
	if out.After.Loads != 1 || out.After.Status != "ready" {
		t.Errorf("after two activations: %+v, want one load", out.After)
	}
	if out.Slow != nil || out.Fast != "<svg>fast</svg>" {
		t.Errorf("out-of-order renders: slow %v, fast %q; want only the latest", out.Slow, out.Fast)
	}
	if out.Failed != "404 plantuml.wasm" || out.FailedStatus != "failed" || out.FailedRender != "404 plantuml.wasm" {
		t.Errorf("failed load: %q, status %q, render %q", out.Failed, out.FailedStatus, out.FailedRender)
	}
	want := []string{"plantuml: line 3: Syntax Error?", "Line 2: Syntax Error?", "graph layout is not available in this build"}
	if strings.Join(out.Describe, "\n") != strings.Join(want, "\n") {
		t.Errorf("describe: %q, want %q", out.Describe, want)
	}
}
