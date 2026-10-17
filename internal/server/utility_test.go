// SPDX-License-Identifier: MIT
// Provenance: original

package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/gopuml"
)

func TestCoderRoundTrip(t *testing.T) {
	srv, fake := newTestServer(t, Config{})
	for _, source := range []string{"Alice -> Bob", helloSource, "@startuml\nnote over A: héllo\n@enduml"} {
		r := do(t, srv, http.MethodPost, "/coder", "text/plain", source)
		checkStatus(t, r, http.StatusOK)
		// The header set of the encode-through-coder reference.
		checkHeaders(t, r.header, map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "text/plain;charset=utf-8",
		})
		for _, c := range r.body {
			if !strings.ContainsRune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_", c) {
				t.Fatalf("encoded text %q is not URL-safe", r.body)
			}
		}
		decoded, err := gopuml.Decode(r.body)
		if err != nil {
			t.Fatal(err)
		}
		want := strings.TrimSuffix(wrapSource(source), "\n")
		if strings.TrimSuffix(decoded, "\n") != want {
			t.Errorf("%q decodes to %q, want %q", r.body, decoded, want)
		}

		g := get(t, srv, "/plantuml/coder/"+r.body)
		checkStatus(t, g, http.StatusOK)
		checkHeaders(t, g.header, map[string]string{"Content-Type": "text/plain;charset=utf-8", "Access-Control-Allow-Origin": "*"})
		if g.body != decoded {
			t.Errorf("GET /coder: %q, want %q", g.body, decoded)
		}
	}
	t.Run("undecodable", func(t *testing.T) {
		checkStatus(t, get(t, srv, "/coder/@@not-valid@@"), http.StatusBadRequest)
	})
	if fake.count() != 0 {
		t.Error("the renderer was called")
	}
}

func TestCheck(t *testing.T) {
	var blocks []*gopuml.Block
	var inspectErr error
	var gotOpts gopuml.Options
	inspect := func(_ context.Context, _ string, opts *gopuml.Options) ([]*gopuml.Block, error) {
		gotOpts = *opts
		return blocks, inspectErr
	}
	srv, fake := newTestServer(t, Config{Inspect: inspect})
	encoded := gopuml.Encode(helloSource)
	for _, tt := range []struct {
		name   string
		blocks []*gopuml.Block
		err    error
		want   string
	}{
		{"valid", []*gopuml.Block{{Type: "SEQUENCE", Description: "(2 participants)"}}, nil, "SEQUENCE\n(2 participants)\n"},
		{"other", []*gopuml.Block{{Description: "(Empty)"}}, nil, "OTHER\n(Empty)\n"},
		{"syntax error", []*gopuml.Block{{Description: "(Error)", Err: &gopuml.SyntaxError{Line: 3, Message: "Syntax Error? (Assumed diagram type: sequence)"}}}, nil,
			"ERROR\n2\nSyntax Error? (Assumed diagram type: sequence)\n"},
		{"unsupported", []*gopuml.Block{{Err: wrapf(gopuml.ErrUnsupportedDiagram, "mindmap")}}, nil, ""},
		{"no block", nil, nil, ""},
		{"no diagram", nil, gopuml.ErrNoDiagram, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			blocks, inspectErr = tt.blocks, tt.err
			r := get(t, srv, "/check/"+encoded)
			checkStatus(t, r, http.StatusOK)
			checkHeaders(t, r.header, map[string]string{"Content-Type": "text/plain;charset=utf-8", "Access-Control-Allow-Origin": "*"})
			if r.body != tt.want {
				t.Errorf("body %q, want %q", r.body, tt.want)
			}
			if gotOpts.SecurityProfile != "INTERNET" {
				t.Errorf("profile %q", gotOpts.SecurityProfile)
			}
		})
	}
	t.Run("failure", func(t *testing.T) {
		blocks, inspectErr = nil, errors.New("broken")
		checkStatus(t, get(t, srv, "/plantuml/check/"+encoded), http.StatusServiceUnavailable)
	})
	t.Run("undecodable", func(t *testing.T) {
		checkStatus(t, get(t, srv, "/check/@@not-valid@@"), http.StatusBadRequest)
	})
	t.Run("timeout", func(t *testing.T) {
		slow := func(ctx context.Context, _ string, _ *gopuml.Options) ([]*gopuml.Block, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		srv, _ := newTestServer(t, Config{Inspect: slow, Timeout: 50 * time.Millisecond})
		checkStatus(t, get(t, srv, "/check/"+encoded), http.StatusServiceUnavailable)
	})
	if fake.count() != 0 {
		t.Error("the renderer was called")
	}
}

func TestServerInfo(t *testing.T) {
	srv, _ := newTestServer(t, Config{})
	for _, path := range []string{"/serverinfo", "/plantuml/serverinfo"} {
		r := get(t, srv, path)
		checkStatus(t, r, http.StatusOK)
		checkHeaders(t, r.header, map[string]string{
			"Content-Type":  "application/json",
			"Cache-Control": "no-cache",
			"Server":        "PlantUML PicoWebServer 1.2026.8",
		})
		want := `{"version":"1.2026.8","PicoWebServer":true,"formats":["png","svg","txt"]}` + "\r\n"
		if r.body != want {
			t.Errorf("body %q, want %q", r.body, want)
		}
	}
}

func TestLanguage(t *testing.T) {
	srv, _ := newTestServer(t, Config{})
	r := get(t, srv, "/language")
	checkStatus(t, r, http.StatusOK)
	checkHeaders(t, r.header, map[string]string{
		"Content-Type":  "text/text",
		"Cache-Control": "no-cache",
		"Server":        "PlantUML PicoWebServer 1.2026.8",
	})
	for _, want := range []string{
		";type\n;53\nabstract\naction\n",
		"\n;keyword\n;196\n@endboard\n",
		"\n@startuml\n",
		"\nleft to right direction\n",
		"\nstackMode\nstart\n",
		"\n;preprocessor\n;35\n!assert\n",
		"\n;skinparameter\n",
		"\n;color\n",
	} {
		if !strings.Contains(r.body, want) {
			t.Errorf("language lacks %q", want)
		}
	}
	if !strings.HasSuffix(r.body, "\n\n;EOF\n") {
		t.Errorf("language ends with %q", r.body[max(0, len(r.body)-20):])
	}
	if strings.Count(r.body, "\nelse\n") != 1 {
		t.Error("duplicate keyword")
	}

	t.Run("custom", func(t *testing.T) {
		srv, _ := newTestServer(t, Config{Language: func(w io.Writer) error {
			_, err := io.WriteString(w, ";type\n;0\n\n;EOF\n")
			return err
		}})
		r := get(t, srv, "/plantuml/language")
		checkStatus(t, r, http.StatusOK)
		if r.body != ";type\n;0\n\n;EOF\n" {
			t.Errorf("body %q", r.body)
		}
	})
	t.Run("failing", func(t *testing.T) {
		srv, _ := newTestServer(t, Config{Language: func(io.Writer) error { return errors.New("no keywords") }})
		checkStatus(t, get(t, srv, "/language"), http.StatusInternalServerError)
	})
}

func TestWriteLanguageMatchesUpstream(t *testing.T) {
	// testdata/language-upstream.txt is the output of
	// "java -jar plantuml.jar -language" in the reference image.
	want, err := os.ReadFile("testdata/language-upstream.txt")
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := WriteLanguage(&got); err != nil {
		t.Fatal(err)
	}
	if got.String() != string(want) {
		t.Error("the language description differs from upstream's")
	}
}
