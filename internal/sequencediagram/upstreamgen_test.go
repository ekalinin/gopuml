// SPDX-License-Identifier: MIT
// Provenance: original

//go:build upstreamgen

package sequencediagram

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGenerateUpstream runs testdata/upstream/Harness.java on the cases with
// the reference jar and rewrites the golden dumps: testdata/upstream/upstream.txt
// for the committed cases, .cache/sequence-upstream/vega.txt for the Vega
// suite when it was fetched. It runs in the reference image, or with the
// local java when SEQ_LOCAL_JAVA is set:
//
//	go test -tags upstreamgen -run TestGenerateUpstream ./internal/sequencediagram
func TestGenerateUpstream(t *testing.T) {
	jar, err := filepath.Abs("../../.cache/upstream/plantuml-mit-1.2026.9.jar")
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	// Docker only mounts the home directory.
	dir, err := os.MkdirTemp(home, "seq-upstream-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	run := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Stderr = os.Stderr
		if out, err := cmd.Output(); err != nil {
			t.Fatalf("%s: %v\n%s", name, err, out)
		}
	}
	run("javac", "--release", "21", "-cp", jar, "-d", filepath.Join(dir, "classes"), "testdata/upstream/Harness.java")
	harness := func(cases []testCase, out string) {
		in := filepath.Join(dir, "cases.txt")
		if err := writeCases(in, cases); err != nil {
			t.Fatal(err)
		}
		if os.Getenv("SEQ_LOCAL_JAVA") != "" {
			run("java", "-Djava.awt.headless=true", "-cp", jar+":"+filepath.Join(dir, "classes"), "Harness",
				filepath.Join(dir, "out.txt"), in)
		} else {
			run("docker", "run", "--rm", "--platform", "linux/amd64", "-v", dir+":/h", "plantuml-go-reference",
				"java", "-Djava.awt.headless=true", "-cp", "/opt/plantuml.jar:/h/classes", "Harness", "/h/out.txt",
				"/h/cases.txt")
		}
		data, err := os.ReadFile(filepath.Join(dir, "out.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	harness(committedCases(t), goldenFile)
	if _, err := os.Stat(vegaRoot); err == nil {
		harness(vegaCases(t), vegaGoldenFile)
	}
}
