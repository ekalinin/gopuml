// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func parse(t *testing.T, args ...string) *Parsed {
	t.Helper()
	p, err := Parse(args)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseFormats(t *testing.T) {
	for _, args := range [][]string{{"--svg"}, {"-tsvg"}, {"-svg"}, {"-TSVG"}, {"--SVG"}, {"--Svg"}} {
		if got := parse(t, append(args, "a.puml")...).Format(); got != "SVG" {
			t.Errorf("%v: format %q, want SVG", args, got)
		}
	}
	// -SVG starts with -S: it is the skinparam flag with key "VG", as upstream.
	p := parse(t, "-SVG", "a.puml")
	if p.Format() != "" || len(p.Map(Skinparam, SkinparamLong)) != 1 || p.Map(Skinparam, SkinparamLong)[0].Key != "VG" {
		t.Errorf("-SVG: format %q skinparams %v", p.Format(), p.Map(Skinparam, SkinparamLong))
	}
	// The first format flag in declaration order wins, not the first given.
	if got := parse(t, "--svg", "--png").Format(); got != "PNG" {
		t.Errorf("--svg --png: %q, want PNG", got)
	}
	// A format flag given twice is ignored, as upstream.
	if got := parse(t, "--svg", "--svg").Format(); got != "" {
		t.Errorf("--svg --svg: %q, want none", got)
	}
}

func TestParseDefines(t *testing.T) {
	p := parse(t, "-DUSER=Alice", "-D", "COUNT=3", "--define", "MODE=dark", "-DFLAG", "-dIGNORED=1", "a.puml")
	got := p.Map(Define, DefineLong)
	want := []KeyValue{{"USER", "Alice", true}, {"COUNT", "3", true}, {"FLAG", "", false}, {"MODE", "dark", true}}
	if !slices.Equal(got, want) {
		t.Errorf("defines = %v, want %v", got, want)
	}
	// -d is not a flag: it becomes an input argument, as upstream.
	if !slices.Equal(p.Args(), []string{"-dIGNORED=1", "a.puml"}) {
		t.Errorf("args = %v", p.Args())
	}
	if p := parse(t, "-DA=1", "-DA=2"); !slices.Equal(p.Map(Define, DefineLong), []KeyValue{{"A", "2", true}}) {
		t.Errorf("repeated key = %v", p.Map(Define, DefineLong))
	}
	if _, err := Parse([]string{"-D"}); err == nil {
		t.Error("-D without a value must fail")
	}
}

func TestParseValues(t *testing.T) {
	p := parse(t, "-o", "out", "--http-server:8081", "-pipedelimitor", "===", "-x", "**/draft/**", "x.puml")
	if v, ok := p.String(OutputDir); !ok || v != "out" {
		t.Errorf("output dir = %q, %v", v, ok)
	}
	if l := p.List(Picoweb); len(l) != 1 || l[0].S != "8081" {
		t.Errorf("http server = %v", l)
	}
	if v, _ := p.String(Pipedelimitor); v != "===" {
		t.Errorf("delimiter = %q", v)
	}
	if v, _ := p.String(Charset); v != "UTF-8" {
		t.Errorf("default charset = %q", v)
	}
	if !p.IsTrue(Picoweb) || p.IsTrue(Pipe) {
		t.Error("IsTrue")
	}
	if p := parse(t, "--http-server"); !p.IsTrue(Picoweb) || len(p.List(Picoweb)) != 0 {
		t.Error("--http-server without a port")
	}
	if p := parse(t, "--output-dir"); p.IsTrue(OutputDir) {
		if _, ok := p.String(OutputDir); ok {
			t.Error("a missing value is null")
		}
	}
	if f, ok := parse(t, "--version").ImmediateAction(); !ok || f != Version {
		t.Error("--version is an immediate action")
	}
}

func TestUnsupported(t *testing.T) {
	err := parse(t, "--svg", "--pdf", "a.puml").CheckSupported()
	if u, ok := errors.AsType[*UnsupportedError](err); !ok || u.Flag != TPdf || !strings.Contains(err.Error(), "--pdf") {
		t.Fatalf("err = %v", err)
	}
	for _, f := range []string{"--gui", "--ftp-server", "--encode-sprite", "--author", "-printfonts"} {
		if parse(t, f).CheckSupported() == nil {
			t.Errorf("%s must be unsupported", f)
		}
	}
	// Unknown flags are input arguments, as upstream.
	if p := parse(t, "--foo", "--svg"); p.CheckSupported() != nil || !slices.Equal(p.Args(), []string{"--foo"}) {
		t.Errorf("--foo: %v %v", p.CheckSupported(), p.Args())
	}
}

func TestHelpMatchesUpstream(t *testing.T) {
	for file, more := range map[string]bool{"testdata/help.txt": false, "testdata/help-more.txt": true} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		want := strings.ReplaceAll(string(data), "java -jar plantuml.jar", "plantuml")
		var b bytes.Buffer
		PrintHelp(&b, "plantuml", more)
		if b.String() != want {
			got, wantLines := strings.Split(b.String(), "\n"), strings.Split(want, "\n")
			for i := range min(len(got), len(wantLines)) {
				if got[i] != wantLines[i] {
					t.Fatalf("%s line %d:\n got %q\nwant %q", file, i+1, got[i], wantLines[i])
				}
			}
			t.Fatalf("%s: %d lines, want %d", file, len(got), len(wantLines))
		}
	}
}

func TestMainEntry(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Main(nil, &out, &errOut); code != ExitOK || !strings.HasPrefix(out.String(), "plantuml - generate diagrams") {
		t.Errorf("no arguments: exit %d, output %q", code, out.String())
	}
	out.Reset()
	if code := Main([]string{"--pdf", "diagram.puml"}, &out, &errOut); code != ExitUnsupported || !strings.Contains(errOut.String(), "--pdf") || out.Len() != 0 {
		t.Errorf("--pdf: exit %d, stderr %q", code, errOut.String())
	}
	out.Reset()
	if code := Main([]string{"-h"}, &out, &errOut); code != ExitOK || !strings.Contains(out.String(), "--help-more") {
		t.Errorf("-h: exit %d", code)
	}
}
