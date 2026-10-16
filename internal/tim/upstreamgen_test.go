// SPDX-License-Identifier: MIT
// Provenance: original

//go:build upstreamgen

package tim

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGenerateUpstream runs testdata/upstream/Harness.java on the case files
// with the reference jar in the reference image and rewrites
// testdata/upstream/upstream.json. It needs javac and the reference image:
//
//	go test -tags upstreamgen -run TestGenerateUpstream ./internal/tim
func TestGenerateUpstream(t *testing.T) {
	jar, err := filepath.Abs("../../.cache/upstream/plantuml-mit-1.2026.8.jar")
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	// Docker only mounts the home directory.
	dir, err := os.MkdirTemp(home, "tim-upstream-")
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
	cases, err := filepath.Glob("testdata/upstream/cases/*.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines, err := filepath.Glob("testdata/upstream/cases/*.lines")
	if err != nil {
		t.Fatal(err)
	}
	cases = append(lines, cases...)
	if err := os.MkdirAll(filepath.Join(dir, "cases"), 0o755); err != nil {
		t.Fatal(err)
	}
	args := []string{"run", "--rm", "--platform", "linux/amd64", "-v", dir + ":/h", "plantuml-go-reference",
		"java", "-cp", "/opt/plantuml.jar:/h/classes", "Harness", "/h/out.json", "/h/work"}
	for _, c := range cases {
		data, err := os.ReadFile(c)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cases", filepath.Base(c)), data, 0o644); err != nil {
			t.Fatal(err)
		}
		args = append(args, "/h/cases/"+filepath.Base(c))
	}
	run("docker", args...)
	out, err := os.ReadFile(filepath.Join(dir, "out.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("testdata/upstream/upstream.json", out, 0o644); err != nil {
		t.Fatal(err)
	}
}
