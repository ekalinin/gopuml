// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml"
	"github.com/ekalinin/gopuml/internal/conformance"
)

const (
	casesDir     = "../../testdata/conformance/cases/http-server"
	referenceDir = "../../testdata/conformance/reference/http-server"
)

// TestRecordedReferences replays every HTTP conformance case against the
// handler. The renderer is a fake that answers with the diagram recorded
// in the reference (its body, size, description and error), so the test
// checks what the server adds: routing, status codes, headers and the
// decoded source and options handed to the renderer. It compares what the
// case lists in "compare", like the conformance runner, and Content-Type
// always.
func TestRecordedReferences(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(casesDir, "*.http.yaml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no HTTP cases: %v", err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		c, err := conformance.ParseHTTPCase("http-server/"+filepath.Base(file), data)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(c.ID, func(t *testing.T) { replay(t, c) })
	}
}

func replay(t *testing.T, c *conformance.HTTPCase) {
	refDir := filepath.Join(referenceDir, filepath.Base(c.ID))
	var ref conformance.HTTPRecord
	data, err := os.ReadFile(filepath.Join(refDir, "http.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatal(err)
	}
	var refBody []byte
	if ref.Body.Stored {
		if refBody, err = os.ReadFile(filepath.Join(refDir, ref.Body.Path)); err != nil {
			t.Fatal(err)
		}
	}

	srv, fake := newTestServer(t, Config{})
	fake.set(func(_ context.Context, _ string, _ *gopuml.Options) (*gopuml.Result, error) {
		return referenceDiagram(t, &ref, refBody)
	})

	req, err := http.NewRequest(ref.Method, srv.URL+ref.Path, strings.NewReader(c.Body))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)

	got, want := resp.Header.Get("Content-Type"), ref.Headers["Content-Type"]
	if got != want {
		t.Errorf("Content-Type %q, reference %q", got, want)
	}
	for _, what := range c.Compare {
		switch what {
		case "status":
			if resp.StatusCode != ref.Status {
				t.Errorf("status %d, reference %d (body %q)", resp.StatusCode, ref.Status, body)
			}
		case "headers":
			for _, k := range slices.Sorted(maps.Keys(ref.Headers)) {
				if got := resp.Header.Get(k); got != ref.Headers[k] {
					t.Errorf("header %s %q, reference %q", k, got, ref.Headers[k])
				}
			}
		case "body":
			if ref.Body.Stored && body != string(refBody) {
				t.Errorf("body differs from %s", ref.Body.Path)
			}
		}
	}

	// What the renderer received: the decoded URL text, or the source of
	// the request body.
	var wantSource string
	switch {
	case ref.Encoded != "":
		if wantSource, err = gopuml.Decode(ref.Encoded); err != nil {
			t.Fatal(err)
		}
	case strings.HasPrefix(ref.Path, "/svg/@@"):
		wantSource = "\n@@not-valid@@\n"
	case ref.Path == "/render" && ref.Status == http.StatusOK:
		wantSource = "@startuml\nAlice -> Bob\n@enduml"
	}
	if wantSource != "" {
		call := fake.last(t)
		if call.opts.SecurityProfile != "INTERNET" || !call.hasDeadline {
			t.Errorf("options %+v, deadline %v", call.opts, call.hasDeadline)
		}
		if strings.Contains(ref.Path, "/dsvg/") != call.opts.DarkMode {
			t.Errorf("dark mode %v for %s", call.opts.DarkMode, ref.Path)
		}
		if !strings.Contains(call.source, wantSource) {
			t.Errorf("rendered source %q, want %q", call.source, wantSource)
		}
	} else if fake.count() != 0 {
		t.Error("the renderer was called")
	}
	// The coder answer is the server's own encoding: it must decode to the
	// same source as the reference's.
	if ref.Path == "/coder" {
		mine, err1 := gopuml.Decode(body)
		theirs, err2 := gopuml.Decode(string(refBody))
		if err1 != nil || err2 != nil || mine != theirs {
			t.Errorf("coder: %q (%v), reference %q (%v)", mine, err1, theirs, err2)
		}
	}
}

// referenceDiagram is the diagram recorded in a reference, as the renderer
// would return it.
func referenceDiagram(t *testing.T, ref *conformance.HTTPRecord, body []byte) (*gopuml.Result, error) {
	h := ref.Headers
	res := &gopuml.Result{Data: body, Description: h["X-Plantuml-Diagram-Description"]}
	res.Width, _ = strconv.Atoi(h["X-Plantuml-Diagram-Width"])
	res.Height, _ = strconv.Atoi(h["X-Plantuml-Diagram-Height"])
	msg, ok := h["X-Plantuml-Diagram-Error"]
	if !ok {
		return res, nil
	}
	line, err := strconv.Atoi(h["X-Plantuml-Diagram-Error-Line"])
	if err != nil {
		t.Fatal(err)
	}
	if ref.Server == "plantuml-server" {
		line++ // plantuml-server sends the 0-based line
	}
	return res, &gopuml.SyntaxError{Line: line, Message: msg}
}

// TestRecordedReferenceHeaders keeps the header set of the svg-by-url
// reference visible in one place: everything the built-in server sends
// for a rendered diagram, as recorded.
func TestRecordedReferenceHeaders(t *testing.T) {
	fake := &fakeRenderer{}
	h, err := New(Config{Render: fake.render})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/svg/Syp9J4vLqBLJSCfFKh1Io4ZDoSa70000", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	want := map[string]string{
		"Access-Control-Allow-Origin":    "*",
		"Cache-Control":                  "no-cache",
		"Content-Type":                   "image/svg+xml",
		"X-PlantUML-Diagram-Description": "(2 participants)",
		"X-PlantUML-Diagram-Height":      "129",
		"X-PlantUML-Diagram-Width":       "119",
	}
	for k, v := range want {
		values := rec.Header()[k]
		if values == nil {
			values = rec.Header()[http.CanonicalHeaderKey(k)]
		}
		if len(values) != 1 || values[0] != v {
			t.Errorf("%s: %q, want %q", k, values, v)
		}
	}
}
