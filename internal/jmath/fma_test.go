// SPDX-License-Identifier: MIT
// Provenance: original

package jmath

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fusedOp matches the fused multiply-add instructions of the Go assemblers:
// FMADDD, FMSUBD, FNMADDD and FNMSUBD (arm64, riscv64, loong64), FMADD,
// FMSUB, FNMADD and FNMSUB (ppc64, s390x), and VFMADD231SD and the like
// (amd64).
var fusedOp = regexp.MustCompile(`\bV?FN?M(ADD|SUB)[A-Z0-9]*\b`)

// fmaTarget is an architecture on which the Go compiler fuses a*b+c.
type fmaTarget struct{ goarch, goamd64 string }

var fmaTargets = []fmaTarget{
	{"arm64", ""},
	{"amd64", "v3"},
	{"ppc64le", ""},
	{"s390x", ""},
	{"riscv64", ""},
	{"loong64", ""},
}

// TestNoFusedMultiplyAdd checks that the compiler emits no fused
// multiply-add for this package on any architecture that has one, so that
// results cannot depend on where the code runs. A probe with an unguarded
// a*b+c shows that the check would see fusion on each target.
func TestNoFusedMultiplyAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the package for several architectures")
	}
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go command not found")
	}
	probe := t.TempDir()
	files := map[string]string{
		"go.mod":   "module fmaprobe\n\ngo 1.26\n",
		"probe.go": "package probe\n\nfunc Fused(a, b, c float64) float64 { return a*b + c }\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(probe, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tg := range fmaTargets {
		t.Run(tg.goarch+tg.goamd64, func(t *testing.T) {
			if !fusedOp.Match(assembly(t, goTool, probe, tg)) {
				t.Fatal("the unguarded probe is not fused, so the check cannot detect fusion")
			}
			out := string(assembly(t, goTool, ".", tg))
			if !strings.Contains(out, "jmath.Sin") {
				t.Fatal("no assembly listing for the package")
			}
			for line := range strings.Lines(out) {
				if fusedOp.MatchString(line) {
					t.Errorf("fused multiply-add: %s", strings.TrimSpace(line))
				}
			}
		})
	}
}

// assembly compiles the package in dir for tg and returns the compiler's
// assembly listing.
func assembly(t *testing.T, goTool, dir string, tg fmaTarget) []byte {
	t.Helper()
	cmd := exec.Command(goTool, "build", "-gcflags=-S", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+tg.goarch, "CGO_ENABLED=0", "GOWORK=off")
	if tg.goamd64 != "" {
		cmd.Env = append(cmd.Env, "GOAMD64="+tg.goamd64)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build for %s: %v\n%s", tg.goarch, err, out)
	}
	return out
}
