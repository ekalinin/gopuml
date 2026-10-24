// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckFile(t *testing.T) {
	const notice = derivedNotice + "\n"
	tests := []struct {
		name    string
		content string
		wantErr string // substring of the first problem, empty for no problems
	}{
		{
			name:    "original",
			content: "// SPDX-License-Identifier: MIT\n// Provenance: original\n\npackage x\n",
		},
		{
			name: "translated with notice",
			content: "// SPDX-License-Identifier: MIT\n" +
				"// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/AsciiEncoder.java\n" +
				"// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/Base64Coder.java\n" +
				notice + "\npackage x\n",
		},
		{
			name:    "fdlibm with notice",
			content: "// SPDX-License-Identifier: MIT\n// Provenance: fdlibm 5.3 e_atan2.c\n" + fdlibmNotice + "\n\npackage x\n",
		},
		{
			name:    "fdlibm without notice",
			content: "// Provenance: fdlibm 5.3 e_atan2.c\npackage x\n",
			wantErr: "lacks Sun's notice",
		},
		{
			name:    "missing header",
			content: "package x\n",
			wantErr: "missing provenance header",
		},
		{
			name:    "header after the first lines",
			content: strings.Repeat("//\n", headerLines) + "// Provenance: original\npackage x\n",
			wantErr: "missing provenance header",
		},
		{
			name:    "translated without notice",
			content: "// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/AsciiEncoder.java\npackage x\n",
			wantErr: "lacks the upstream copyright notice",
		},
		{
			name:    "excluded ditaa",
			content: "// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/ditaa/PSystemDitaa.java\n" + notice + "package x\n",
			wantErr: "excluded component (ditaa)",
		},
		{
			name:    "excluded ditaa library",
			content: "// Provenance: plantuml-mit 1.2026.9 org/stathissideris/ascii2image/core/Shape.java\n" + notice + "package x\n",
			wantErr: "excluded component",
		},
		{
			name:    "excluded jcckit",
			content: "// Provenance: plantuml-mit 1.2026.9 jcckit/plot/Plot.java\n" + notice + "package x\n",
			wantErr: "excluded component (jcckit)",
		},
		{
			name:    "both original and upstream",
			content: "// Provenance: original\n// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/AsciiEncoder.java\n" + notice + "package x\n",
			wantErr: "both original and upstream",
		},
		{
			name:    "unknown provenance",
			content: "// Provenance: plantuml-gpl net/sourceforge/plantuml/Run.java\npackage x\n",
			wantErr: "unrecognized provenance",
		},
		{
			name:    "not a java path",
			content: "// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code\n" + notice + "package x\n",
			wantErr: "not a Java source file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := checkFile("x.go", tt.content, "")
			if tt.wantErr == "" {
				if len(problems) != 0 {
					t.Fatalf("unexpected problems: %v", problems)
				}
				return
			}
			if len(problems) == 0 {
				t.Fatalf("expected a problem containing %q, got none", tt.wantErr)
			}
			if !strings.Contains(problems[0].Message, tt.wantErr) {
				t.Fatalf("problem %q does not contain %q", problems[0].Message, tt.wantErr)
			}
		})
	}
}

func TestCheckFileUpstreamPathMustExist(t *testing.T) {
	src := t.TempDir()
	existing := filepath.Join(src, "net", "sourceforge", "plantuml", "code")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(existing, "AsciiEncoder.java"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ok := "// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/AsciiEncoder.java\n" + derivedNotice + "\n"
	if problems := checkFile("x.go", ok, src); len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	missing := "// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/code/Missing.java\n" + derivedNotice + "\n"
	problems := checkFile("x.go", missing, src)
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "does not exist") {
		t.Fatalf("expected a does-not-exist problem, got %v", problems)
	}
}

func TestCheckTree(t *testing.T) {
	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a/good.go", "// Provenance: original\npackage a\n")
	write("a/bad.go", "package a\n")
	write("a/notes.txt", "not go")

	problems, count, err := checkTree([]string{root}, "")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("checked %d files, want 2", count)
	}
	if len(problems) != 1 || !strings.HasSuffix(problems[0].File, "a/bad.go") {
		t.Fatalf("want one problem for bad.go, got %v", problems)
	}
}
