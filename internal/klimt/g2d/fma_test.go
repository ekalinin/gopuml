// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fusedOp matches the fused multiply-add instructions of the Go assemblers
// (arm64, riscv64, loong64, ppc64, s390x and amd64).
var fusedOp = regexp.MustCompile(`\bV?FN?M(ADD|SUB)[A-Z0-9]*\b`)

// TestNoFusedMultiplyAdd checks that the compiler emits no fused
// multiply-add for the raster packages on any architecture that has one
// (design D17): the images then do not depend on where the code runs.
func TestNoFusedMultiplyAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the packages for several architectures")
	}
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go command not found")
	}
	targets := []struct{ goarch, goamd64 string }{
		{"arm64", ""}, {"amd64", "v3"}, {"ppc64le", ""}, {"s390x", ""}, {"riscv64", ""}, {"loong64", ""},
	}
	probe := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":   "module fmaprobe\n\ngo 1.26\n",
		"probe.go": "package probe\n\nfunc Fused(a, b, c float64) float64 { return a*b + c }\n",
	} {
		if err := os.WriteFile(filepath.Join(probe, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tg := range targets {
		t.Run(tg.goarch+tg.goamd64, func(t *testing.T) {
			if !fusedOp.Match(assembly(t, goTool, probe, tg.goarch, tg.goamd64, ".")) {
				t.Fatal("the unguarded probe is not fused, so the check cannot detect fusion")
			}
			out := string(assembly(t, goTool, ".", tg.goarch, tg.goamd64, ".", "./dejavu", "../../png"))
			for _, fn := range []string{"g2d.rasterExact", "dejavu.decompose", "png.quantify555"} {
				if !strings.Contains(out, fn) {
					t.Fatalf("no assembly listing of %s", fn)
				}
			}
			for line := range strings.Lines(out) {
				if fusedOp.MatchString(line) {
					t.Errorf("fused multiply-add: %s", strings.TrimSpace(line))
				}
			}
		})
	}
}

// assembly compiles packages for a target and returns the compiler's
// assembly listing of them.
func assembly(t *testing.T, goTool, dir, goarch, goamd64 string, pkgs ...string) []byte {
	t.Helper()
	args := []string{"build"}
	for _, p := range pkgs {
		args = append(args, "-gcflags="+p+"=-S")
	}
	args = append(args, pkgs...)
	cmd := exec.Command(goTool, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+goarch, "CGO_ENABLED=0", "GOWORK=off")
	if goamd64 != "" {
		cmd.Env = append(cmd.Env, "GOAMD64="+goamd64)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", goarch, err, out)
	}
	return out
}
