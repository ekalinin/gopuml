// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/conformance"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/security"
)

// conformanceKnownDifferences are the cases that cannot pass yet.
var conformanceKnownDifferences = map[string]string{
	"preprocessor/dec2hex": "%dec2hex is not ported yet (task 4.5)",
}

// TestConformancePreproc preprocesses the curated preprocessor cases and
// compares the lines with the PREPROC output of the reference jar
// (testdata/conformance/reference/preprocessor/<name>/block<i>.preproc).
// Fixture files are in the <name>.files directory next to the case.
func TestConformancePreproc(t *testing.T) {
	root := "../../testdata/conformance"
	cases, err := conformance.LoadCases(os.DirFS(root), "cases")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, c := range cases {
		if c.Capability() != "preprocessor" || !slices.Contains(c.Formats, "preproc") {
			continue
		}
		n++
		t.Run(c.ID, func(t *testing.T) {
			if reason, ok := conformanceKnownDifferences[c.ID]; ok {
				t.Skip(reason)
			}
			dir, err := filepath.Abs(filepath.Join(root, "cases", strings.TrimSuffix(c.Path, ".puml")+".files"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(dir); err != nil {
				dir = t.TempDir()
			}
			ctx := core.NewRenderContext()
			ctx.CurrentDir = dir
			access := &security.Access{Config: security.Config{Profile: security.Legacy}}
			defs := &BlockDefinitions{}
			for i, b := range core.BuildBlocks(c.Source, "string", nil) {
				res, _ := Preprocess(ctx, access, b.Lines, defs)
				defs.Add(res.Lines)
				ref := filepath.Join(root, "reference", c.ID, "block"+strconv.Itoa(i)+".preproc")
				want, err := os.ReadFile(ref)
				if err != nil {
					t.Fatal(err)
				}
				got := strings.Join(res.PreprocLines(), "\n") + "\n"
				if got != string(want) {
					t.Errorf("block %d:\ngot:\n%s\nwant:\n%s", i, got, want)
				}
			}
		})
	}
	if n == 0 {
		t.Fatal("no preprocessor case")
	}
}
