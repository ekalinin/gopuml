// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/gopuml"
)

const helloSource = "@startuml\nAlice -> Bob : hello\n@enduml\n"

func TestGETRoutes(t *testing.T) {
	builtin := map[string]string{"Cache-Control": "no-cache", "Server": "PlantUML PicoWebServer 1.2026.8"}
	pserver := map[string]string{"Cache-Control": "public, max-age=432000", "Server": ""}
	tests := []struct {
		route  string
		format gopuml.Format
		mime   string
		dark   bool
		style  map[string]string
	}{
		{"png", gopuml.PNG, "image/png", false, builtin},
		{"svg", gopuml.SVG, "image/svg+xml", false, builtin},
		{"txt", gopuml.TXT, "text/plain", false, builtin},
		{"utxt", gopuml.UTXT, "text/plain;charset=UTF-8", false, builtin},
		{"img", gopuml.PNG, "image/png", false, pserver},
		{"dpng", gopuml.PNG, "image/png", true, pserver},
		{"dimg", gopuml.PNG, "image/png", true, pserver},
		{"dsvg", gopuml.SVG, "image/svg+xml", true, pserver},
	}
	srv, fake := newTestServer(t, Config{})
	encoded := gopuml.Encode(helloSource)
	want, err := gopuml.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		for _, prefix := range []string{"", "/plantuml"} {
			path := prefix + "/" + tt.route + "/" + encoded
			t.Run(path, func(t *testing.T) {
				r := get(t, srv, path)
				checkStatus(t, r, http.StatusOK)
				headers := map[string]string{
					"Content-Type":                   tt.mime,
					"Access-Control-Allow-Origin":    "*",
					"X-PlantUML-Diagram-Width":       "119",
					"X-PlantUML-Diagram-Height":      "129",
					"X-PlantUML-Diagram-Description": "(2 participants)",
					"X-PlantUML-Diagram-Error":       "",
					"X-PlantUML-Diagram-Error-Line":  "",
				}
				for k, v := range tt.style {
					headers[k] = v
				}
				checkHeaders(t, r.header, headers)
				if r.body != string(okResult(tt.format).Data) {
					t.Errorf("body %q", r.body)
				}
				c := fake.last(t)
				if c.source != want {
					t.Errorf("rendered source %q, want %q", c.source, want)
				}
				if c.opts.Format != tt.format || c.opts.DarkMode != tt.dark || c.opts.SecurityProfile != "INTERNET" {
					t.Errorf("options %+v", c.opts)
				}
				if !c.hasDeadline {
					t.Error("the render context has no deadline")
				}
			})
		}
	}
	t.Run("HEAD", func(t *testing.T) {
		r := do(t, srv, http.MethodHead, "/svg/"+encoded, "", "")
		checkStatus(t, r, http.StatusOK)
		checkHeaders(t, r.header, map[string]string{"Content-Type": "image/svg+xml", "X-PlantUML-Diagram-Width": "119"})
		if r.body != "" {
			t.Errorf("HEAD body %q", r.body)
		}
	})
	t.Run("last segment", func(t *testing.T) {
		// Upstream reads the text after the last '/'.
		checkStatus(t, get(t, srv, "/svg/ignored/"+encoded), http.StatusOK)
		if c := fake.last(t); c.source != want {
			t.Errorf("rendered source %q", c.source)
		}
	})
}

// TestHeaderSpelling checks that the X-PlantUML-* headers go out with
// upstream's spelling.
func TestHeaderSpelling(t *testing.T) {
	fake := &fakeRenderer{fn: func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
		return errorResult(), &gopuml.SyntaxError{Line: 3, Message: "Syntax Error? (Assumed diagram type: sequence)"}
	}}
	h, err := New(Config{Render: fake.render})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/svg/"+gopuml.Encode(helloSource), nil))
	for _, k := range []string{
		"X-PlantUML-Diagram-Width", "X-PlantUML-Diagram-Height", "X-PlantUML-Diagram-Description",
		"X-PlantUML-Diagram-Error", "X-PlantUML-Diagram-Error-Line",
	} {
		if _, ok := rec.Header()[k]; !ok {
			t.Errorf("header %s is not sent with this spelling: %v", k, rec.Header())
		}
	}
}

func TestSyntaxErrorByURL(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
		return errorResult(), &gopuml.SyntaxError{Line: 3, Message: "Syntax Error? (Assumed diagram type: sequence)"}
	})
	encoded := gopuml.Encode("@startuml\nAlice -> Bob\nfoo bar baz\n@enduml\n")
	for _, tt := range []struct {
		path  string
		cache string
	}{
		{"/svg/", "no-cache"},
		{"/plantuml/svg/", "no-cache"},
		{"/dsvg/", "public, max-age=432000"},
	} {
		t.Run(tt.path, func(t *testing.T) {
			r := get(t, srv, tt.path+encoded)
			checkStatus(t, r, http.StatusBadRequest)
			checkHeaders(t, r.header, map[string]string{
				"Access-Control-Allow-Origin":    "*",
				"Cache-Control":                  tt.cache,
				"Content-Type":                   "image/svg+xml",
				"X-PlantUML-Diagram-Description": "(Error)",
				"X-PlantUML-Diagram-Error":       "Syntax Error? (Assumed diagram type: sequence)",
				"X-PlantUML-Diagram-Error-Line":  "3",
				"X-PlantUML-Diagram-Height":      "340",
				"X-PlantUML-Diagram-Width":       "591",
			})
			if r.body != "<error image>" {
				t.Errorf("body %q", r.body)
			}
		})
	}
	t.Run("without error image", func(t *testing.T) {
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
			return nil, &gopuml.SyntaxError{Line: 2, Message: "cannot include /etc/passwd"}
		})
		r := get(t, srv, "/svg/"+encoded)
		checkStatus(t, r, http.StatusBadRequest)
		checkHeaders(t, r.header, map[string]string{
			"X-PlantUML-Diagram-Description": "(Error)",
			"X-PlantUML-Diagram-Error":       "cannot include /etc/passwd",
			"X-PlantUML-Diagram-Error-Line":  "2",
			"X-PlantUML-Diagram-Width":       "",
		})
	})
}

func TestUnsupportedDiagram(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
		return errorResult(), wrapf(gopuml.ErrUnsupportedDiagram, "state diagrams are not supported yet")
	})
	r := get(t, srv, "/svg/"+gopuml.Encode("@startuml\n[*] --> Idle\n@enduml\n"))
	checkStatus(t, r, http.StatusBadRequest)
	checkHeaders(t, r.header, map[string]string{
		"X-PlantUML-Diagram-Error":      "state diagrams are not supported yet",
		"X-PlantUML-Diagram-Error-Line": "",
	})
}

// wrapf wraps a sentinel error like the gopuml package does.
func wrapf(sentinel error, msg string) error {
	return &wrapped{sentinel, msg}
}

type wrapped struct {
	sentinel error
	msg      string
}

func (w *wrapped) Error() string { return w.sentinel.Error() + ": " + w.msg }
func (w *wrapped) Unwrap() error { return w.sentinel }

func TestUndecodableURL(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	fake.set(func(_ context.Context, source string, _ *gopuml.Options) (*gopuml.Result, error) {
		return &gopuml.Result{Data: []byte("<bad url image>"), Width: 789, Height: 310, Description: "(0 entities)"}, nil
	})
	r := get(t, srv, "/svg/@@not-valid@@")
	checkStatus(t, r, http.StatusOK)
	checkHeaders(t, r.header, map[string]string{
		"Access-Control-Allow-Origin":    "*",
		"Cache-Control":                  "no-cache",
		"Content-Type":                   "image/svg+xml",
		"X-PlantUML-Diagram-Description": "(0 entities)",
		"X-PlantUML-Diagram-Height":      "310",
		"X-PlantUML-Diagram-Width":       "789",
	})
	c := fake.last(t)
	for _, want := range []string{"@startuml\nlegend\n", "This URL does not look like DEFLATE data.", "\n@@not-valid@@\n"} {
		if !strings.Contains(c.source, want) {
			t.Errorf("rendered source lacks %q:\n%s", want, c.source)
		}
	}

	t.Run("hex form", func(t *testing.T) {
		n := fake.count()
		r := get(t, srv, "/svg/~hzz")
		checkStatus(t, r, http.StatusBadRequest)
		checkHeaders(t, r.header, map[string]string{"Content-Type": "text/plain"})
		if fake.count() != n {
			t.Error("the renderer was called")
		}
	})
}

func TestNoDiagram(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
		return nil, wrapf(gopuml.ErrNoDiagram, "diagram 0 out of range (0 diagrams)")
	})
	r := get(t, srv, "/svg/"+gopuml.Encode("@startuml\nAlice -> Bob"))
	checkStatus(t, r, http.StatusFound)
	checkHeaders(t, r.header, map[string]string{"Location": defaultLocation})

	r = do(t, srv, http.MethodPost, "/render", "application/json", `{"source": "@startuml\nAlice -> Bob"}`)
	checkStatus(t, r, http.StatusBadRequest)
	if r.body != noValidStart {
		t.Errorf("body %q", r.body)
	}
}

func TestRenderFailureReturns503(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	encoded := gopuml.Encode("@startuml\nclass A\n@enduml\n")

	fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
		return nil, wrapf(gopuml.ErrGraphvizFailed, "dot crashed")
	})
	r := get(t, srv, "/png/"+encoded)
	checkStatus(t, r, http.StatusServiceUnavailable)
	checkHeaders(t, r.header, map[string]string{"Content-Type": "text/plain"})
	if !strings.Contains(r.body, "dot crashed") {
		t.Errorf("body %q", r.body)
	}

	t.Run("crash image", func(t *testing.T) {
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
			return &gopuml.Result{Data: []byte("<crash image>"), Width: 400, Height: 300, Description: "(Error)"}, wrapf(gopuml.ErrGraphvizFailed, "dot crashed")
		})
		r := get(t, srv, "/png/"+encoded)
		checkStatus(t, r, http.StatusServiceUnavailable)
		checkHeaders(t, r.header, map[string]string{"Content-Type": "image/png", "X-PlantUML-Diagram-Width": "400"})
		if r.body != "<crash image>" {
			t.Errorf("body %q", r.body)
		}
	})

	t.Run("panic", func(t *testing.T) {
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
			panic("boom")
		})
		r := get(t, srv, "/svg/"+encoded)
		checkStatus(t, r, http.StatusServiceUnavailable)
		if !strings.Contains(r.body, "boom") {
			t.Errorf("body %q", r.body)
		}
	})

	t.Run("no result", func(t *testing.T) {
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) { return nil, nil })
		checkStatus(t, get(t, srv, "/svg/"+encoded), http.StatusServiceUnavailable)
	})
}

func TestTimeoutReturns503(t *testing.T) {
	const limit = 100 * time.Millisecond
	srv, fake := newTestServer(t, Config{Timeout: limit})
	encoded := gopuml.Encode(helloSource)

	t.Run("renderer stops", func(t *testing.T) {
		stopped := make(chan error, 1)
		fake.set(func(ctx context.Context, _ string, _ *gopuml.Options) (*gopuml.Result, error) {
			<-ctx.Done()
			stopped <- ctx.Err()
			return nil, ctx.Err()
		})
		r := get(t, srv, "/svg/"+encoded)
		checkStatus(t, r, http.StatusServiceUnavailable)
		checkHeaders(t, r.header, map[string]string{"Content-Type": "text/plain"})
		select {
		case err := <-stopped:
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("render context error %v, want the deadline", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("the render context was not cancelled")
		}
		if !fake.last(t).hasDeadline {
			t.Error("the render context has no deadline")
		}
	})

	t.Run("renderer ignores the context", func(t *testing.T) {
		release := make(chan struct{})
		defer close(release)
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
			<-release
			return okResult(gopuml.SVG), nil
		})
		start := time.Now()
		r := get(t, srv, "/svg/"+encoded)
		checkStatus(t, r, http.StatusServiceUnavailable)
		if d := time.Since(start); d > 5*time.Second {
			t.Errorf("answered after %v", d)
		}
	})

	t.Run("slow upload", func(t *testing.T) {
		n := fake.count()
		conn, err := net.Dial("tcp", srv.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		// Announce 100 bytes and send 5.
		io.WriteString(conn, "POST /svg HTTP/1.1\r\nHost: x\r\nContent-Type: text/plain\r\nContent-Length: 100\r\n\r\nAlice")
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("status %d, want 503", resp.StatusCode)
		}
		if fake.count() != n {
			t.Error("the renderer was called")
		}
	})

	t.Run("POST /render", func(t *testing.T) {
		fake.set(func(ctx context.Context, _ string, _ *gopuml.Options) (*gopuml.Result, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
		r := do(t, srv, http.MethodPost, "/render", "application/json", `{"source": "Alice -> Bob"}`)
		checkStatus(t, r, http.StatusServiceUnavailable)
	})
}

func TestOversizedSourceReturns413(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	big := "@startuml\n" + strings.Repeat("Alice -> Bob : hello\n", (2<<20)/21) + "@enduml\n"
	if len(big) < 2<<20-64 {
		t.Fatalf("source of %d bytes", len(big))
	}
	for _, path := range []string{"/svg", "/png", "/txt", "/plantuml/svg"} {
		t.Run(path, func(t *testing.T) {
			r := do(t, srv, http.MethodPost, path, "text/plain", big)
			checkStatus(t, r, http.StatusRequestEntityTooLarge)
		})
	}
	t.Run("/render", func(t *testing.T) {
		body := `{"source": "` + strings.Repeat("x", 2<<20) + `"}`
		checkStatus(t, do(t, srv, http.MethodPost, "/render", "application/json", body), http.StatusRequestEntityTooLarge)
	})
	t.Run("/coder", func(t *testing.T) {
		checkStatus(t, do(t, srv, http.MethodPost, "/coder", "text/plain", big), http.StatusRequestEntityTooLarge)
	})
	t.Run("at the limit", func(t *testing.T) {
		src := strings.Repeat("x", int(DefaultMaxBodySize))
		checkStatus(t, do(t, srv, http.MethodPost, "/svg", "text/plain", src), http.StatusOK)
	})
	if n := fake.count(); n != 1 {
		t.Errorf("the renderer was called %d times, want once", n)
	}

	t.Run("decoded source", func(t *testing.T) {
		small, fake := newTestServer(t, Config{MaxBodySize: 64})
		long := "~h" + strings.Repeat("41", 65)
		checkStatus(t, get(t, small, "/svg/"+long), http.StatusRequestEntityTooLarge)
		checkStatus(t, get(t, small, "/coder/"+long), http.StatusRequestEntityTooLarge)
		checkStatus(t, get(t, small, "/check/"+long), http.StatusRequestEntityTooLarge)
		if fake.count() != 0 {
			t.Error("the renderer was called")
		}
	})
}

func TestJSONRenderRequest(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
		return &gopuml.Result{Data: []byte("<svg/>"), Width: 118, Height: 114, Description: "(2 participants)"}, nil
	})
	r := do(t, srv, http.MethodPost, "/render", "application/json", `{"source": "Alice -> Bob", "options": ["--svg"]}`)
	checkStatus(t, r, http.StatusOK)
	checkHeaders(t, r.header, map[string]string{
		"Access-Control-Allow-Origin":    "*",
		"Cache-Control":                  "no-cache",
		"Content-Type":                   "image/svg+xml",
		"Server":                         "PlantUML PicoWebServer 1.2026.8",
		"X-PlantUML-Diagram-Description": "(2 participants)",
		"X-PlantUML-Diagram-Height":      "114",
		"X-PlantUML-Diagram-Width":       "118",
	})
	c := fake.last(t)
	if c.source != "@startuml\nAlice -> Bob\n@enduml" {
		t.Errorf("rendered source %q", c.source)
	}
	if c.opts.Format != gopuml.SVG || c.opts.SecurityProfile != "INTERNET" || !c.hasDeadline {
		t.Errorf("options %+v, deadline %v", c.opts, c.hasDeadline)
	}

	t.Run("defaults", func(t *testing.T) {
		r := do(t, srv, http.MethodPost, "/plantuml/render", "application/json", `{"source": "@startuml\nA -> B\n@enduml", "extra": 1}`)
		checkStatus(t, r, http.StatusOK)
		checkHeaders(t, r.header, map[string]string{"Content-Type": "image/png"})
		if c := fake.last(t); c.source != "@startuml\nA -> B\n@enduml" || c.opts.Format != gopuml.PNG {
			t.Errorf("call %+v", c)
		}
	})

	t.Run("options", func(t *testing.T) {
		body := `{"source": "A -> B", "options": ["-tutxt", "--dark-mode", "-DX=1", "--theme", "cerulean"]}`
		checkStatus(t, do(t, srv, http.MethodPost, "/render", "application/json", body), http.StatusOK)
		c := fake.last(t)
		if c.opts.Format != gopuml.UTXT || !c.opts.DarkMode || len(c.opts.Defines) != 1 || c.opts.Config[0] != "!theme cerulean" {
			t.Errorf("options %+v", c.opts)
		}
	})

	t.Run("syntax error", func(t *testing.T) {
		// handleRenderRequest answers 200 even for an error image.
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
			return errorResult(), &gopuml.SyntaxError{Line: 3, Message: "Syntax Error? (Assumed diagram type: sequence)"}
		})
		r := do(t, srv, http.MethodPost, "/render", "application/json", `{"source": "Alice -> Bob\nfoo bar baz", "options": ["--svg"]}`)
		checkStatus(t, r, http.StatusOK)
		checkHeaders(t, r.header, map[string]string{
			"X-PlantUML-Diagram-Error":      "Syntax Error? (Assumed diagram type: sequence)",
			"X-PlantUML-Diagram-Error-Line": "3",
		})
	})
}

func TestInvalidJSON(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	for _, tt := range []struct{ name, body, msg string }{
		{"not json", `{not json`, "Error parsing request json: "},
		{"empty", ``, "No request body"},
		{"array", `["Alice -> Bob"]`, "Error parsing request json: "},
		{"null", `null`, "Error parsing request json: not an object"},
		{"no source", `{"options": ["--svg"]}`, `Error parsing request json: missing "source"`},
		{"source not a string", `{"source": 1}`, "Error parsing request json: "},
		{"source null", `{"source": null}`, "Error parsing request json: "},
		{"source in another case", `{"Source": "A -> B"}`, `Error parsing request json: missing "source"`},
		{"options not an array", `{"source": "A -> B", "options": "--svg"}`, "Error parsing request json: "},
		{"options null", `{"source": "A -> B", "options": null}`, "Error parsing request json: "},
		{"option not a string", `{"source": "A -> B", "options": [1]}`, "Error parsing request json: "},
		{"trailing data", `{"source": "A -> B"} x`, "Error parsing request json: "},
		{"unsupported format", `{"source": "A -> B", "options": ["--pdf"]}`, "Unsupported output format: PDF"},
		{"missing value", `{"source": "A -> B", "options": ["-D"]}`, "Missing value after -D"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := do(t, srv, http.MethodPost, "/render", "application/json", tt.body)
			checkStatus(t, r, http.StatusBadRequest)
			checkHeaders(t, r.header, map[string]string{"Content-Type": "text/plain"})
			if !strings.HasPrefix(r.body, tt.msg) {
				t.Errorf("body %q, want prefix %q", r.body, tt.msg)
			}
		})
	}
	if fake.count() != 0 {
		t.Error("the renderer was called")
	}
}

func TestRawPOST(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	for _, tt := range []struct {
		path   string
		format gopuml.Format
		mime   string
	}{
		{"/png", gopuml.PNG, "image/png"},
		{"/svg", gopuml.SVG, "image/svg+xml"},
		{"/txt", gopuml.TXT, "text/plain"},
		{"/plantuml/svg", gopuml.SVG, "image/svg+xml"},
		{"/svg/", gopuml.SVG, "image/svg+xml"},
	} {
		t.Run(tt.path, func(t *testing.T) {
			r := do(t, srv, http.MethodPost, tt.path, "text/plain", "Alice -> Bob")
			checkStatus(t, r, http.StatusOK)
			checkHeaders(t, r.header, map[string]string{
				"Content-Type":                   tt.mime,
				"Access-Control-Allow-Origin":    "*",
				"Cache-Control":                  "no-cache",
				"X-PlantUML-Diagram-Description": "(2 participants)",
				"X-PlantUML-Diagram-Width":       "119",
				"X-PlantUML-Diagram-Height":      "129",
			})
			c := fake.last(t)
			if c.source != "@startuml\nAlice -> Bob\n@enduml" || c.opts.Format != tt.format || c.opts.SecurityProfile != "INTERNET" {
				t.Errorf("call %+v", c)
			}
		})
	}
	t.Run("source with @start", func(t *testing.T) {
		checkStatus(t, do(t, srv, http.MethodPost, "/svg", "text/plain", helloSource), http.StatusOK)
		if c := fake.last(t); c.source != helloSource {
			t.Errorf("rendered source %q", c.source)
		}
	})
	t.Run("syntax error", func(t *testing.T) {
		fake.set(func(context.Context, string, *gopuml.Options) (*gopuml.Result, error) {
			return errorResult(), &gopuml.SyntaxError{Line: 2, Message: "Syntax Error?"}
		})
		r := do(t, srv, http.MethodPost, "/svg", "text/plain", "foo bar baz")
		checkStatus(t, r, http.StatusBadRequest)
		checkHeaders(t, r.header, map[string]string{"X-PlantUML-Diagram-Error": "Syntax Error?", "X-PlantUML-Diagram-Error-Line": "2"})
	})
}

func TestRedirect(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, "/"},
		{http.MethodGet, "/plantuml"},
		{http.MethodGet, "/foo"},
		{http.MethodGet, "/svg"},
		{http.MethodGet, "/render"},
		{http.MethodPost, "/foo"},
		{http.MethodPost, "/utxt"},
		{http.MethodPut, "/svg/" + gopuml.Encode(helloSource)},
		{http.MethodDelete, "/render"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := do(t, srv, tt.method, tt.path, "", "")
			checkStatus(t, r, http.StatusFound)
			checkHeaders(t, r.header, map[string]string{"Location": "/plantuml/png/oqbDJyrBuGh8ISmh2VNrKGZ8JCuFJqqAJYqgIotY0aefG5G00000"})
		})
	}
	if fake.count() != 0 {
		t.Error("the renderer was called")
	}
}

func TestSecurityProfile(t *testing.T) {
	encoded := gopuml.Encode("@startuml\n!include /etc/passwd\nAlice -> Bob\n@enduml\n")
	t.Run("default", func(t *testing.T) {
		t.Setenv("PLANTUML_SECURITY_PROFILE", "")
		srv, fake := newTestServer(t, Config{})
		fake.set(func(_ context.Context, _ string, opts *gopuml.Options) (*gopuml.Result, error) {
			if opts.SecurityProfile != "INTERNET" {
				return okResult(opts.Format), nil // the file would be read
			}
			return &gopuml.Result{Data: []byte("<error image>"), Width: 591, Height: 324, Description: "(Error)"},
				&gopuml.SyntaxError{Line: 2, Message: "cannot include /etc/passwd"}
		})
		r := get(t, srv, "/svg/"+encoded)
		checkStatus(t, r, http.StatusBadRequest)
		checkHeaders(t, r.header, map[string]string{
			"X-PlantUML-Diagram-Error":      "cannot include /etc/passwd",
			"X-PlantUML-Diagram-Error-Line": "2",
		})
	})
	for _, tt := range []struct{ env, cfg, want string }{
		{"sandbox", "", "SANDBOX"},
		{"bogus", "", "INTERNET"},
		{"", "legacy", "LEGACY"},
		{"SANDBOX", "internet", "INTERNET"},
	} {
		t.Run(tt.env+"/"+tt.cfg, func(t *testing.T) {
			t.Setenv("PLANTUML_SECURITY_PROFILE", tt.env)
			srv, fake := newTestServer(t, Config{SecurityProfile: tt.cfg})
			checkStatus(t, get(t, srv, "/svg/"+encoded), http.StatusOK)
			if got := fake.last(t).opts.SecurityProfile; got != tt.want {
				t.Errorf("profile %q, want %q", got, tt.want)
			}
		})
	}
	t.Run("unknown in Config", func(t *testing.T) {
		if _, err := New(Config{SecurityProfile: "bogus"}); err == nil {
			t.Error("no error")
		}
	})
}
