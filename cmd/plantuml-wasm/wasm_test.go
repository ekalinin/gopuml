// SPDX-License-Identifier: MIT
// Provenance: original

//go:build !js

package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/gopuml"
)

var (
	buildOnce sync.Once
	buildDir  string
	buildErr  error
)

// build compiles the module and copies the loader files into a temporary
// directory, as make wasm does in dist/wasm.
func build(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	buildOnce.Do(func() {
		buildDir, buildErr = os.MkdirTemp("", "plantuml-wasm")
		if buildErr != nil {
			return
		}
		cmd := exec.Command("go", "build", "-o", filepath.Join(buildDir, "plantuml.wasm"), ".")
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = errors.New(string(out))
			return
		}
		for src, dst := range map[string]string{
			filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js"): "wasm_exec.js",
			"plantuml.mjs": "plantuml.mjs",
		} {
			data, err := os.ReadFile(src)
			if err != nil {
				buildErr = err
				return
			}
			if err := os.WriteFile(filepath.Join(buildDir, dst), data, 0o644); err != nil {
				buildErr = err
				return
			}
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return buildDir
}

func TestMain(m *testing.M) {
	code := m.Run()
	if buildDir != "" {
		os.RemoveAll(buildDir)
	}
	os.Exit(code)
}

// jsResult is the result object of plantuml.render, with data as a
// string.
type jsResult struct {
	Data        *string `json:"data"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Description string  `json:"description"`
	Error       *struct {
		Kind    string `json:"kind"`
		Message string `json:"message"`
		Line    int    `json:"line"`
	} `json:"error"`
}

// renderInNode loads the module with the ES module loader in Node.js and
// renders source with options.
func renderInNode(t *testing.T, source string, options map[string]any) jsResult {
	t.Helper()
	dir := build(t)
	input, _ := json.Marshal(map[string]any{"source": source, "options": options})
	script := `import { load } from './plantuml.mjs';
const { source, options } = JSON.parse(process.argv[2]);
// A file given as {"base64": ...} is passed as its bytes.
for (const [name, value] of Object.entries(options?.files ?? {}))
  if (typeof value === 'object') options.files[name] = new Uint8Array(Buffer.from(value.base64, 'base64'));
const plantuml = await load();
const r = await plantuml.render(source, options);
r.data = r.data === null ? null : new TextDecoder().decode(r.data);
console.log(JSON.stringify(r));
process.exit(0);
`
	if err := os.WriteFile(filepath.Join(dir, "run.mjs"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", "run.mjs", string(input))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("node: %v\n%s", err, stderr)
	}
	var r jsResult
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("result %q: %v", out, err)
	}
	return r
}

func TestWasmBuildProducesModule(t *testing.T) {
	// Spec scenario "Build produces the module".
	dir := build(t)
	for _, f := range []string{"plantuml.wasm", "wasm_exec.js", "plantuml.mjs"} {
		if st, err := os.Stat(filepath.Join(dir, f)); err != nil || st.Size() == 0 {
			t.Errorf("%s: %v", f, err)
		}
	}
	// The loader runs the module and the API answers.
	r := renderInNode(t, "no diagram here", nil)
	if r.Error == nil || r.Error.Kind != "no-diagram" {
		t.Errorf("render without a diagram: %+v", r.Error)
	}
}

func TestWasmRenderMatchesNative(t *testing.T) {
	// Spec scenario "Render in a browser".
	src := "@startuml\nAlice -> Bob : hello\n@enduml\n"
	native, err := gopuml.Render(context.Background(), src, &gopuml.Options{SecurityProfile: "SANDBOX"})
	if errors.Is(err, gopuml.ErrUnsupportedDiagram) {
		t.Skip("sequence diagrams are not drawn yet (task 8.8)")
	}
	if err != nil {
		t.Fatal(err)
	}
	r := renderInNode(t, src, map[string]any{"format": "svg"})
	if r.Error != nil || r.Data == nil || *r.Data != string(native.Data) || r.Width != native.Width || r.Height != native.Height {
		t.Errorf("WASM result differs from native: error %+v, %dx%d", r.Error, r.Width, r.Height)
	}
}

func TestWasmStdlibInclude(t *testing.T) {
	// Spec scenario "Standard library include": the include resolves from
	// the embedded standard library, without host files, in the SANDBOX
	// profile.
	src := "@startuml\n!include <tupadr3/font-awesome/server>\nAlice -> Bob : <$server> hi\n@enduml\n"
	native, err := gopuml.Render(context.Background(), src, &gopuml.Options{SecurityProfile: "SANDBOX"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(native.Data), "<image ") {
		t.Fatalf("native render draws no sprite: %s", native.Data)
	}
	r := renderInNode(t, src, map[string]any{"format": "svg"})
	if r.Error != nil || r.Data == nil || *r.Data != string(native.Data) || r.Width != native.Width || r.Height != native.Height {
		t.Errorf("WASM result differs from native: error %+v, %dx%d", r.Error, r.Width, r.Height)
	}
}

func TestWasmHostSuppliedInclude(t *testing.T) {
	// Spec scenario "Include supplied by the host": the include resolves from
	// the files of the options, in the SANDBOX profile.
	src := "@startuml\n!include style.iuml\nAlice -> Bob : $GREETING\n@enduml\n"
	files := map[string]any{"style.iuml": "!$GREETING = \"hello\"\n"}
	r := renderInNode(t, src, map[string]any{"format": "svg", "files": files})
	if r.Error != nil && r.Error.Kind != "unsupported-diagram" {
		t.Fatalf("with the file: %+v", r.Error)
	}
	if r.Error == nil && !strings.Contains(*r.Data, "hello") {
		t.Errorf("the included variable is not used: %s", *r.Data)
	}
	r = renderInNode(t, src, map[string]any{"format": "svg"})
	if r.Error == nil || r.Error.Kind != "syntax" || !strings.Contains(r.Error.Message, "style.iuml") {
		t.Errorf("without the file: %+v", r.Error)
	}
}

func TestWasmHostSuppliedArchive(t *testing.T) {
	// An archive given as a Uint8Array in the files of the options is
	// imported, and its entries are included.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("lib/part.iuml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("!$GREETING = \"hello\"\n")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	src := "@startuml\n!import archive.zip\n!include lib/part.iuml\nAlice -> Bob : $GREETING\n@enduml\n"
	files := map[string]any{"archive.zip": map[string]any{"base64": base64.StdEncoding.EncodeToString(buf.Bytes())}}
	r := renderInNode(t, src, map[string]any{"format": "svg", "files": files})
	if r.Error != nil {
		t.Fatalf("with the archive: %+v", r.Error)
	}
	if !strings.Contains(*r.Data, "hello") {
		t.Errorf("the variable of the archive entry is not used: %s", *r.Data)
	}
}

func TestWasmClassDiagramError(t *testing.T) {
	// Spec scenario "Class diagram in WASM".
	r := renderInNode(t, "@startuml\nclass A\nclass B\nA --> B\n@enduml\n", map[string]any{"format": "svg"})
	if r.Error == nil {
		t.Fatal("class diagram rendered without graph layout")
	}
	if r.Error.Kind == "unsupported-diagram" || r.Error.Kind == "syntax" {
		// Until the class factory exists, the sequence parser reports a
		// syntax error for class syntax.
		t.Skip("class diagrams are not parsed yet (task 12.5)")
	}
	if r.Error.Kind != "graph-layout-unavailable" || r.Error.Message != "graph layout is not available in this build" || r.Data == nil {
		t.Errorf("error %+v, image present %v", r.Error, r.Data != nil)
	}
}

func TestWasmDescriptionDiagram(t *testing.T) {
	// Spec scenario "Description diagram in WASM".
	r := renderInNode(t, "@startuml\nnode Server\ndatabase Orders\nServer --> Orders\n@enduml\n",
		map[string]any{"format": "svg"})
	if r.Error == nil {
		t.Fatal("description diagram rendered without graph layout")
	}
	if r.Error.Kind != "graph-layout-unavailable" || r.Error.Message != "graph layout is not available in this build" || r.Data == nil {
		t.Errorf("error %+v, image present %v", r.Error, r.Data != nil)
	}
}

func TestWasmStateDiagram(t *testing.T) {
	// Spec scenario "State diagram in WASM".
	r := renderInNode(t, "@startuml\n[*] --> Idle\nstate Active {\n[*] --> Running\n}\nIdle --> Active\n@enduml\n",
		map[string]any{"format": "svg"})
	if r.Error == nil {
		t.Fatal("state diagram rendered without graph layout")
	}
	if r.Error.Kind != "graph-layout-unavailable" || r.Error.Message != "graph layout is not available in this build" || r.Data == nil {
		t.Errorf("error %+v, image present %v", r.Error, r.Data != nil)
	}
}

func TestWasmLegacyActivityDiagram(t *testing.T) {
	// Spec scenario "Legacy activity diagram in WASM".
	r := renderInNode(t, "@startuml\n(*) --> \"First step\"\n\"First step\" --> (*)\n@enduml\n",
		map[string]any{"format": "svg"})
	if r.Error == nil {
		t.Fatal("legacy activity diagram rendered without graph layout")
	}
	if r.Error.Kind != "graph-layout-unavailable" || r.Error.Message != "graph layout is not available in this build" || r.Data == nil {
		t.Errorf("error %+v, image present %v", r.Error, r.Data != nil)
	}
}

func TestWasmPacketDiagram(t *testing.T) {
	// Spec scenario "Packet diagram in WASM": packet diagrams have a
	// built-in layout, so the WASM build draws them as the native build.
	src := "@startpacketdiag\n0-7: Source Port\n8-15: Destination Port\n@endpacketdiag\n"
	native, err := gopuml.Render(context.Background(), src, &gopuml.Options{SecurityProfile: "SANDBOX"})
	if err != nil {
		t.Fatal(err)
	}
	r := renderInNode(t, src, map[string]any{"format": "svg"})
	if r.Error != nil || r.Data == nil || *r.Data != string(native.Data) || r.Width != native.Width || r.Height != native.Height {
		t.Errorf("WASM result differs from native: error %+v, %dx%d", r.Error, r.Width, r.Height)
	}
}

func TestWasmGanttDiagram(t *testing.T) {
	// Spec scenario "Gantt diagram in WASM": gantt diagrams have a built-in
	// layout, so the WASM build draws them as the native build.
	src := "@startgantt\nProject starts 2020-01-06\n[Prototype design] lasts 13 days\n" +
		"[Test prototype] lasts 4 days\n[Test prototype] starts at [Prototype design]'s end\n@endgantt\n"
	native, err := gopuml.Render(context.Background(), src, &gopuml.Options{SecurityProfile: "SANDBOX"})
	if err != nil {
		t.Fatal(err)
	}
	r := renderInNode(t, src, map[string]any{"format": "svg"})
	if r.Error != nil || r.Data == nil || *r.Data != string(native.Data) || r.Width != native.Width || r.Height != native.Height {
		t.Errorf("WASM result differs from native: error %+v, %dx%d", r.Error, r.Width, r.Height)
	}
}

func TestWasmMindMapDiagram(t *testing.T) {
	// Spec scenario "Mindmap diagram in WASM".
	t.Skip("mindmap diagrams are registered by task 5.2 of add-mindmap-diagrams")
}
