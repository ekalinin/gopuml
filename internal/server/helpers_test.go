// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ekalinin/gopuml"
)

// renderCall records one call of the fake renderer.
type renderCall struct {
	source      string
	opts        gopuml.Options
	hasDeadline bool
}

// fakeRenderer stands in for gopuml.Render, which cannot draw diagrams
// yet. It records its calls and answers with fn.
type fakeRenderer struct {
	mu    sync.Mutex
	calls []renderCall
	fn    func(ctx context.Context, source string, opts *gopuml.Options) (*gopuml.Result, error)
}

func (f *fakeRenderer) render(ctx context.Context, source string, opts *gopuml.Options) (*gopuml.Result, error) {
	_, ok := ctx.Deadline()
	f.mu.Lock()
	f.calls = append(f.calls, renderCall{source: source, opts: *opts, hasDeadline: ok})
	fn := f.fn
	f.mu.Unlock()
	if fn == nil {
		return okResult(opts.Format), nil
	}
	return fn(ctx, source, opts)
}

// set replaces the answer of the renderer.
func (f *fakeRenderer) set(fn func(ctx context.Context, source string, opts *gopuml.Options) (*gopuml.Result, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fn = fn
}

func (f *fakeRenderer) last(t *testing.T) renderCall {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.calls) == 0 {
		t.Fatal("the renderer was not called")
	}
	return f.calls[len(f.calls)-1]
}

func (f *fakeRenderer) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// okResult is a canned rendering of "Alice -> Bob : hello", with the size
// and description of the svg-by-url reference.
func okResult(format gopuml.Format) *gopuml.Result {
	return &gopuml.Result{
		Data:        []byte("<" + format.String() + " data>"),
		Width:       119,
		Height:      129,
		Description: "(2 participants)",
		Pages:       1,
		Diagrams:    1,
	}
}

// errorResult is a canned error image, with the size of the
// syntax-error-by-url reference.
func errorResult() *gopuml.Result {
	return &gopuml.Result{Data: []byte("<error image>"), Width: 591, Height: 340, Description: "(Error)", Pages: 1, Diagrams: 1}
}

// newTestServer starts a server whose renderer is a fake. cfg.Render is
// replaced by the fake.
func newTestServer(t *testing.T, cfg Config) (*httptest.Server, *fakeRenderer) {
	t.Helper()
	fake := &fakeRenderer{}
	cfg.Render = fake.render
	h, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, fake
}

// noRedirect is a client that returns redirects instead of following them.
var noRedirect = &http.Client{
	Timeout:       30 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// response is a read HTTP response.
type response struct {
	status int
	header http.Header
	body   string
}

func do(t *testing.T, srv *httptest.Server, method, path, contentType, body string) response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response{status: resp.StatusCode, header: resp.Header, body: string(data)}
}

func get(t *testing.T, srv *httptest.Server, path string) response {
	t.Helper()
	return do(t, srv, http.MethodGet, path, "", "")
}

// checkHeaders compares the given headers exactly. An empty want value
// means the header must be absent.
func checkHeaders(t *testing.T, got http.Header, want map[string]string) {
	t.Helper()
	for k, v := range want {
		values, present := got[http.CanonicalHeaderKey(k)]
		switch {
		case v == "" && present:
			t.Errorf("header %s = %q, want none", k, values)
		case v != "" && (len(values) != 1 || values[0] != v):
			t.Errorf("header %s = %q, want %q", k, values, v)
		}
	}
}

func checkStatus(t *testing.T, r response, want int) {
	t.Helper()
	if r.status != want {
		t.Fatalf("status %d, want %d (body %q)", r.status, want, r.body)
	}
}
