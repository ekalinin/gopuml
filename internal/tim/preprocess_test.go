// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/security"
)

// preprocessBlock preprocesses the first block of src, read from dir.
func preprocessBlock(t *testing.T, src string, access *security.Access, dir string) (*Result, error) {
	t.Helper()
	ctx := core.NewRenderContext()
	if dir != "" {
		ctx.CurrentDir = dir
	}
	blocks := core.BuildBlocks(src, "string", nil)
	if len(blocks) == 0 {
		t.Fatal("no block")
	}
	return Preprocess(ctx, access, blocks[0].Lines, nil)
}

func memAccess(profile security.Profile, files map[string]string) *security.Access {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return &security.Access{Config: security.Config{Profile: profile}, FS: fsys}
}

func wantLines(t *testing.T, res *Result, err error, want ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	got := res.PreprocLines()
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func wantError(t *testing.T, err error, message string, position int) {
	t.Helper()
	e, ok := errors.AsType[*EaterException](err)
	if !ok {
		t.Fatalf("got %v, want the error %q", err, message)
	}
	if e.Message != message {
		t.Errorf("message %q, want %q", e.Message, message)
	}
	if e.Location.Location == nil || e.Location.Location.Position != position {
		t.Errorf("error at %+v, want position %d", e.Location.Location, position)
	}
}

func TestSpecScenarios(t *testing.T) {
	access := memAccess(security.Legacy, nil)
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"variable in a message", "@startuml\n!$user = \"Alice\"\n$user -> Bob : hi\n@enduml\n",
			[]string{"@startuml", "Alice -> Bob : hi", "@enduml"}},
		{"conditional assignment keeps the existing value", "@startuml\n!$color = \"red\"\n!$color ?= \"blue\"\nA : $color\n@enduml\n",
			[]string{"@startuml", "A : red", "@enduml"}},
		{"legacy macro with arguments", "@startuml\n!define SEND(a,b) a -> b\nSEND(Alice,Bob)\n@enduml\n",
			[]string{"@startuml", "Alice -> Bob", "@enduml"}},
		{"if-else branch", "@startuml\n!$n = 2\n!if $n > 1\nA -> B : many\n!else\nA -> B : one\n!endif\n@enduml\n",
			[]string{"@startuml", "A -> B : many", "@enduml"}},
		{"while loop", "@startuml\n!$i = 0\n!while $i < 3\nA -> B : $i\n!$i = $i + 1\n!endwhile\n@enduml\n",
			[]string{"@startuml", "A -> B : 0", "A -> B : 1", "A -> B : 2", "@enduml"}},
		{"procedure call with default argument", "@startuml\n!procedure $msg($to, $text=\"ping\")\nA -> $to : $text\n!endprocedure\n$msg(\"B\")\n@enduml\n",
			[]string{"@startuml", "A -> B : ping", "@enduml"}},
		{"function return value", "@startuml\n!function $double($x)\n!return $x * 2\n!endfunction\nA -> B : $double(21)\n@enduml\n",
			[]string{"@startuml", "A -> B : 42", "@enduml"}},
		{"string builtins", "@startuml\nA : %strlen(\"abc\") %upper(\"abc\")\n@enduml\n",
			[]string{"@startuml", "A : 3 ABC", "@enduml"}},
		{"JSON member access", "@startuml\n!$data = { \"name\": \"api\", \"ports\": [80, 443] }\nA -> B : $data.ports[1]\n@enduml\n",
			[]string{"@startuml", "A -> B : 443", "@enduml"}},
		{"layout pragma is passed on", "@startuml\n!pragma layout elk\nclass A\n@enduml\n",
			[]string{"@startuml", "!pragma layout elk", "class A", "@enduml"}},
		{"undefined function in diagram text", "@startuml\nA -> B : $missing(1)\n@enduml\n",
			[]string{"@startuml", "A -> B : $missing(1)", "@enduml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := preprocessBlock(t, tt.src, access, "")
			wantLines(t, res, err, tt.want...)
		})
	}
}

func TestSpecErrors(t *testing.T) {
	access := memAccess(security.Legacy, nil)
	tests := []struct {
		name     string
		src      string
		message  string
		position int
	}{
		{"undefined function", "@startuml\nAlice -> Bob : hi\n!$x = $missing(1)\n@enduml\n", "Unknown built-in function $missing", 2},
		{"failing assertion", "@startuml\nAlice -> Bob : hi\n!assert 1 == 2 : \"numbers differ\"\n@enduml\n", "Assertion error : numbers differ", 2},
		{"missing include", "@startuml\nAlice -> Bob : hi\n!include missing.iuml\n@enduml\n", "cannot include missing.iuml", 2},
		{"unbalanced endif", "@startuml\n!endif\n@enduml\n", "No if related to this endif", 1},
		{"wrong argument count", "@startuml\n!function $f($a)\n!return $a\n!endfunction\n!$x = $f(1, 2)\n@enduml\n", "Unknown built-in function $f", 4},
		{"invalid expression", "@startuml\n!$x = \"abc\n@enduml\n", "a001", 1},
		{"runtime exception at top level", "@startuml\n!$x = 1 / 0\n@enduml\n", "Fatal parsing error", 1},
		{"infinite recursion", "@startuml\n!function $r($n)\n!return $r($n + 1)\n!endfunction\nA : $r(1)\n@enduml\n", "Too many nested calls (infinite recursion?)", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := preprocessBlock(t, tt.src, access, "")
			wantError(t, err, tt.message, tt.position)
			last := res.Lines[len(res.Lines)-1]
			if last.PreprocessorError != tt.message {
				t.Errorf("last line %+v, want the error %q", last, tt.message)
			}
		})
	}
}

// TestStdlibIncludeNotBundledError checks the spec scenario "Library not
// bundled": upstream bundles C4, this build bundles no library.
func TestStdlibIncludeNotBundledError(t *testing.T) {
	src := "@startuml\nAlice -> Bob : hi\n!include <C4/C4_Container>\n@enduml\n"
	res, err := preprocessBlock(t, src, memAccess(security.Legacy, nil), "")
	wantError(t, err, "cannot include <C4/C4_Container> (library C4 is not bundled)", 2)
	if got := res.PreprocLines(); !slices.Equal(got, []string{"@startuml", "Alice -> Bob : hi", "!include <C4/C4_Container>"}) {
		t.Errorf("lines %q", got)
	}

	// A bundled library is included, relative to its own folder.
	stdlib := NewStdlib(map[string]map[string][]byte{
		"mylib": {"sub/part.puml": []byte("A -> B : from mylib\n")},
	})
	ctx := core.NewRenderContext()
	blocks := core.BuildBlocks("@startuml\n!include <MyLib/sub/part>\n@enduml\n", "string", nil)
	res, err = PreprocessWith(ctx, memAccess(security.Legacy, nil), blocks[0].Lines, Options{Stdlib: stdlib})
	wantLines(t, res, err, "@startuml", "A -> B : from mylib", "@enduml")
	if loc := res.Lines[1].Location; loc.Desc != "<MyLib/sub/part>" {
		t.Errorf("location %+v", loc)
	}
}

func TestIncludeFromFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("diagrams/common/style.iuml", "Bob -> Carol : style\n")
	write("common/style.iuml", "wrong file\n")
	write("a.iuml", "Alice -> Bob : from a.iuml\n")
	src := "@startuml\n!include common/style.iuml\n@enduml\n"

	t.Run("relative to the including file", func(t *testing.T) {
		access := &security.Access{Config: security.Config{Profile: security.Legacy}}
		res, err := preprocessBlock(t, src, access, filepath.Join(dir, "diagrams"))
		wantLines(t, res, err, "@startuml", "Bob -> Carol : style", "@enduml")
		if len(res.Included) != 1 || !strings.HasSuffix(res.Included[0], filepath.Join("diagrams", "common", "style.iuml")) {
			t.Errorf("included %q", res.Included)
		}
	})
	t.Run("include once twice", func(t *testing.T) {
		access := &security.Access{Config: security.Config{Profile: security.Legacy}}
		res, err := preprocessBlock(t, "@startuml\n!include_once a.iuml\n!include_once a.iuml\n@enduml\n", access, dir)
		wantLines(t, res, err, "@startuml", "Alice -> Bob : from a.iuml", "Alice -> Bob : from a.iuml", "@enduml")
	})
	t.Run("sandbox blocks local files", func(t *testing.T) {
		access := &security.Access{Config: security.Config{Profile: security.Sandbox}}
		_, err := preprocessBlock(t, "@startuml\nAlice -> Bob : hi\n!include a.iuml\n@enduml\n", access, dir)
		wantError(t, err, "cannot include a.iuml", 2)
	})
	t.Run("system paths are denied by default", func(t *testing.T) {
		access := &security.Access{Config: security.Config{Profile: security.Legacy}}
		_, err := preprocessBlock(t, "@startuml\nAlice -> Bob : hi\n!include /etc/passwd\n@enduml\n", access, dir)
		wantError(t, err, "cannot include /etc/passwd", 2)
	})
	t.Run("default access from the render context", func(t *testing.T) {
		ctx := core.NewRenderContext()
		ctx.CurrentDir = dir
		blocks := core.BuildBlocks("@startuml\n!include a.iuml\n@enduml\n", "string", nil)
		res, err := Preprocess(ctx, nil, blocks[0].Lines, nil)
		wantLines(t, res, err, "@startuml", "Alice -> Bob : from a.iuml", "@enduml")
		ctx.SecurityProfile = "SANDBOX"
		_, err = Preprocess(ctx, nil, blocks[0].Lines, nil)
		wantError(t, err, "cannot include a.iuml", 1)
	})
}

// fakeResolver makes every host resolve to a public address, so that the
// URL checks of the security profile accept the test server.
func fakeResolver(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }

// clientTo returns an HTTP client that sends every request to server.
func clientTo(server *httptest.Server) *http.Client {
	addr := server.Listener.Addr().String()
	return &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}}
}

func TestURLInclude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/style.iuml":
			w.Write([]byte("Bob -> Carol : from url\n"))
		case "/multi.puml":
			w.Write([]byte("@startuml(id=one)\nfirst\n@enduml\n@startuml(id=two)\nsecond\n@enduml\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	access := func(p security.Profile) *security.Access {
		return &security.Access{Config: security.Config{Profile: p, LookupIP: fakeResolver}, HTTP: clientTo(server)}
	}

	res, err := preprocessBlock(t, "@startuml\n!include http://example.com/style.iuml\n@enduml\n", access(security.Internet), "")
	wantLines(t, res, err, "@startuml", "Bob -> Carol : from url", "@enduml")
	if loc := res.Lines[1].Location; loc.Desc != "http://example.com/style.iuml" || loc.Parent == nil {
		t.Errorf("location %+v", loc)
	}

	res, err = preprocessBlock(t, "@startuml\n!include http://example.com/multi.puml!two\n!include http://example.com/multi.puml!0\n@enduml\n", access(security.Internet), "")
	wantLines(t, res, err, "@startuml", "second", "first", "@enduml")

	_, err = preprocessBlock(t, "@startuml\n!include http://example.com/missing.iuml\n@enduml\n", access(security.Internet), "")
	wantError(t, err, "Cannot open URL", 1)

	_, err = preprocessBlock(t, "@startuml\n!include http://example.com/style.iuml\n@enduml\n", access(security.Sandbox), "")
	wantError(t, err, "Cannot open URL", 1)
}

type mapThemes map[string]string

func (m mapThemes) Theme(name string) ([]byte, bool) {
	s, ok := m[name]
	return []byte(s), ok
}

func TestThemeHook(t *testing.T) {
	src := "@startuml\n!theme demo\nA : $themevar\n@enduml\n"
	blocks := core.BuildBlocks(src, "string", nil)
	ctx := core.NewRenderContext()

	_, err := PreprocessWith(ctx, memAccess(security.Legacy, nil), blocks[0].Lines, Options{})
	wantError(t, err, "Cannot load theme demo", 1)

	themes := mapThemes{"demo": "---\nauthor: me\n---\nskinparam backgroundColor red\n!$themevar = \"themed\"\n"}
	res, err := PreprocessWith(ctx, memAccess(security.Legacy, nil), blocks[0].Lines, Options{Themes: themes})
	wantLines(t, res, err, "@startuml", "skinparam backgroundColor red", "A : themed", "@enduml")
	if got := res.ThemeMetadata.String(); got != `{"author":"me"}` {
		t.Errorf("metadata %s", got)
	}
}

func TestLogDumpMemoryOptionAndIncludeDef(t *testing.T) {
	src := "@startdef(id=d)\nA -> B : def\n@enddef\n@startuml\n!$b = 2\n!log value $b\n!dump_memory\n!option handwritten\n!includedef d\n@enduml\n"
	ctx := core.NewRenderContext()
	defs := &BlockDefinitions{}
	var last *Result
	for _, b := range core.BuildBlocks(src, "string", nil) {
		res, err := Preprocess(ctx, memAccess(security.Legacy, nil), b.Lines, defs)
		if err != nil {
			t.Fatal(err)
		}
		defs.Add(res.Lines)
		last = res
	}
	wantLines(t, last, nil, "@startuml", "A -> B : def", "@enduml")
	if !slices.Equal(last.Log, []string{"[Log] value 2"}) {
		t.Errorf("log %q", last.Log)
	}
	if !slices.Equal(last.Info, []string{"[MemGlobal] Start of memory_dump ", "[MemGlobal] Number of variable(s) : 1", "[MemGlobal] $b = 2", "[MemGlobal] End of memory_dump"}) {
		t.Errorf("info %q", last.Info)
	}
	if v, ok := last.Artifact.OptionValue("HANDWRITTEN"); !ok || v != "true" || !last.Artifact.OptionIsTrue("HANDWRITTEN") {
		t.Errorf("options %v", last.Artifact.Options())
	}
}

func TestDefinesFromContext(t *testing.T) {
	ctx := core.NewRenderContext()
	ctx.Defines = map[string]string{"FOO": "bar", "$n": "1"}
	blocks := core.BuildBlocks("@startuml\nA : FOO $n\n@enduml\n", "string", nil)
	res, err := Preprocess(ctx, memAccess(security.Legacy, nil), blocks[0].Lines, nil)
	wantLines(t, res, err, "@startuml", "A : bar 1", "@enduml")
}

// TestConcurrentPreprocess runs renders in parallel: there is no shared
// mutable state (run with -race).
func TestConcurrentPreprocess(t *testing.T) {
	access := memAccess(security.Legacy, map[string]string{"a.iuml": "!$v = \"A\"\n"})
	src := "@startuml\n!include a.iuml\n!function $f($x)\n!return $x + $v\n!endfunction\n!$i = 0\n!while $i < 20\nA : $f($i)\n!$i = $i + 1\n!endwhile\n@enduml\n"
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			res, err := preprocessBlock(t, src, access, "/")
			if err != nil || len(res.Lines) != 22 || res.Lines[20].S != "A : 19A" {
				t.Errorf("got %v %v", res, err)
			}
		})
	}
	wg.Wait()
}
