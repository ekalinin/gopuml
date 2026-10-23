// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// page is the index.html of the test sites.
const page = `<!doctype html><meta name="gopuml-release" content="v0.1.0"><title>t</title>`

// moduleFiles are the files of a fake module.
var moduleFiles = map[string]string{
	"plantuml.wasm": "\x00asm fake module",
	"wasm_exec.js":  "// wasm_exec",
	"plantuml.mjs":  "export function load() {}",
}

// sources writes a page with a style sheet and an image in a subdirectory
// into a temporary directory and returns it.
func sources(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	for name, data := range map[string]string{
		"index.html":     page,
		"style.css":      "body {}",
		"img/hero.svg":   "<svg/>",
		"playground.mjs": "export {}",
	} {
		writeFile(t, filepath.Join(src, name), data)
	}
	return src
}

func writeFile(t *testing.T, name, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// githubOutput points GITHUB_OUTPUT to a temporary file and returns a
// function that reads it.
func githubOutput(t *testing.T) func() string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "output")
	t.Setenv("GITHUB_OUTPUT", name)
	return func() string {
		data, _ := os.ReadFile(name)
		return string(data)
	}
}

// checkSite checks that out has the files of src and the module files.
func checkSite(t *testing.T, src, out string) {
	t.Helper()
	for _, name := range []string{"index.html", "style.css", "img/hero.svg", "playground.mjs"} {
		want, _ := os.ReadFile(filepath.Join(src, name))
		got, err := os.ReadFile(filepath.Join(out, name))
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("%s: %q, %v; want %q", name, got, err, want)
		}
	}
	for name, want := range moduleFiles {
		got, err := os.ReadFile(filepath.Join(out, "wasm", name))
		if err != nil || string(got) != want {
			t.Errorf("wasm/%s: %q, %v; want %q", name, got, err, want)
		}
	}
}

func noOutput(t *testing.T, out string) {
	t.Helper()
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("%s exists: %v", out, err)
	}
}

func TestLocalBuild(t *testing.T) {
	src := sources(t)
	build := t.TempDir()
	for name, data := range moduleFiles {
		writeFile(t, filepath.Join(build, name), data)
	}
	output := githubOutput(t)
	out := filepath.Join(t.TempDir(), "site")
	// A stale file of an earlier run is removed.
	writeFile(t, filepath.Join(out, "stale.txt"), "old")

	var stdout bytes.Buffer
	if err := run([]string{"-src", src, "-out", out, "-wasm", build}, &stdout); err != nil {
		t.Fatal(err)
	}
	checkSite(t, src, out)
	if _, err := os.Stat(filepath.Join(out, "stale.txt")); !os.IsNotExist(err) {
		t.Errorf("stale.txt is still in the site: %v", err)
	}
	if got := output(); got != "deploy=true\n" {
		t.Errorf("GITHUB_OUTPUT = %q, want deploy=true", got)
	}

	// A build without one of the module files writes nothing.
	os.Remove(filepath.Join(build, "wasm_exec.js"))
	out2 := filepath.Join(t.TempDir(), "site")
	if err := run([]string{"-src", src, "-out", out2, "-wasm", build}, &stdout); err == nil {
		t.Error("no error without wasm_exec.js")
	}
	noOutput(t, out2)
}

func TestUsage(t *testing.T) {
	src := sources(t)
	for _, args := range [][]string{
		{"-src", src},
		{"-src", src, "-out", t.TempDir()},
		{"-src", src, "-out", t.TempDir(), "-wasm", "a", "-download", "http://localhost"},
	} {
		if err := run(args, new(bytes.Buffer)); err == nil {
			t.Errorf("%q: no error", args)
		}
	}
	// A page without its release.
	writeFile(t, filepath.Join(src, "index.html"), strings.Replace(page, "gopuml-release", "release", 1))
	if err := run([]string{"-src", src, "-out", filepath.Join(t.TempDir(), "site"), "-wasm", t.TempDir()}, new(bytes.Buffer)); err == nil || !strings.Contains(err.Error(), "0 gopuml-release meta tags") {
		t.Errorf("page without release: %v", err)
	}
}
