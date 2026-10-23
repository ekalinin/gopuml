// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
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

// archive returns a .tar.gz archive with the files at its root, as
// GoReleaser writes them, and a file that is not part of the module.
func archive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	add := func(name, data string) {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range files {
		add(name, data)
	}
	add("README.md", "not part of the module")
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// releaseServer serves the files of the release v0.1.0 as GitHub serves
// release downloads, and 404 for everything else.
func releaseServer(t *testing.T, files map[string][]byte) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, ok := files[strings.TrimPrefix(r.URL.Path, "/v0.1.0/")]
		if !ok || !strings.HasPrefix(r.URL.Path, "/v0.1.0/") {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// published returns the files of a published release v0.1.0 whose WASM
// archive has the files, with sum as its checksum (the right one when sum
// is empty).
func published(t *testing.T, files map[string]string, sum string) map[string][]byte {
	t.Helper()
	data := archive(t, files)
	if sum == "" {
		s := sha256.Sum256(data)
		sum = hex.EncodeToString(s[:])
	}
	return map[string][]byte{
		"checksums.txt":                   []byte(strings.Repeat("0", 64) + "  gopuml_0.1.0_linux_amd64.tar.gz\n" + sum + "  gopuml_0.1.0_wasm.tar.gz\n"),
		"gopuml_0.1.0_wasm.tar.gz":        data,
		"gopuml_0.1.0_linux_amd64.tar.gz": []byte("other archive"),
	}
}

func TestDownloadRelease(t *testing.T) {
	src := sources(t)
	output := githubOutput(t)
	out := filepath.Join(t.TempDir(), "site")
	var stdout bytes.Buffer
	if err := run([]string{"-src", src, "-out", out, "-download", releaseServer(t, published(t, moduleFiles, "")) + "/"}, &stdout); err != nil {
		t.Fatal(err)
	}
	checkSite(t, src, out)
	if got := output(); got != "deploy=true\n" {
		t.Errorf("GITHUB_OUTPUT = %q, want deploy=true", got)
	}

	// An archive without one of the module files writes nothing.
	partial := map[string]string{"plantuml.wasm": moduleFiles["plantuml.wasm"], "plantuml.mjs": moduleFiles["plantuml.mjs"]}
	out2 := filepath.Join(t.TempDir(), "site")
	err := run([]string{"-src", src, "-out", out2, "-download", releaseServer(t, published(t, partial, ""))}, &stdout)
	if err == nil || !strings.Contains(err.Error(), "has no wasm_exec.js") {
		t.Errorf("archive without wasm_exec.js: %v", err)
	}
	noOutput(t, out2)
}

func TestDownloadMismatch(t *testing.T) {
	src := sources(t)
	output := githubOutput(t)
	out := filepath.Join(t.TempDir(), "site")
	files := published(t, moduleFiles, strings.Repeat("ab", 32))
	err := run([]string{"-src", src, "-out", out, "-download", releaseServer(t, files)}, new(bytes.Buffer))
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Errorf("mismatch: %v", err)
	}
	noOutput(t, out)
	if got := output(); got != "" {
		t.Errorf("GITHUB_OUTPUT = %q, want nothing", got)
	}

	// A checksums file without the WASM archive writes nothing either.
	delete(files, "gopuml_0.1.0_wasm.tar.gz")
	files["checksums.txt"] = []byte(strings.Repeat("0", 64) + "  gopuml_0.1.0_linux_amd64.tar.gz\n")
	if err := run([]string{"-src", src, "-out", out, "-download", releaseServer(t, files)}, new(bytes.Buffer)); err == nil {
		t.Error("no error without the WASM archive in checksums.txt")
	}
	noOutput(t, out)
}

func TestDownloadUnpublished(t *testing.T) {
	src := sources(t)
	output := githubOutput(t)
	out := filepath.Join(t.TempDir(), "site")
	var stdout bytes.Buffer
	if err := run([]string{"-src", src, "-out", out, "-download", releaseServer(t, nil)}, &stdout); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout.String(), "::notice::release v0.1.0 is not published: ") {
		t.Errorf("stdout = %q, want a notice", stdout.String())
	}
	noOutput(t, out)
	if got := output(); got != "deploy=false\n" {
		t.Errorf("GITHUB_OUTPUT = %q, want deploy=false", got)
	}

	// Another error status is a failure, not an unpublished release.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	if err := run([]string{"-src", src, "-out", out, "-download", srv.URL}, new(bytes.Buffer)); err == nil {
		t.Error("no error for status 503")
	}
	noOutput(t, out)
}
