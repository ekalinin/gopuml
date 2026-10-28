// SPDX-License-Identifier: MIT
// Provenance: original

package wbs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// TestTxtReferences requires the text of the curated text cases to be
// their references.
func TestTxtReferences(t *testing.T) {
	for _, tc := range []struct {
		id     string
		format font.FileFormat
		file   string
	}{
		{"wbs-diagram/txt-root-children", font.FormatATXT, "block0-page0.atxt"},
		{"output-formats/wbs-txt", font.FormatATXT, "block0-page0.atxt"},
		{"wbs-diagram/utxt-output", font.FormatUTXT, "block0-page0.utxt"},
	} {
		want, err := os.ReadFile(filepath.Join(referencesDir, tc.id, tc.file))
		if err != nil {
			t.Fatal(err)
		}
		d := parseSource(t, caseSource(t, filepath.Join(casesDir, tc.id+".puml")))
		var buf bytes.Buffer
		if _, err := d.ExportDiagram(&buf, 0, export.NewFileFormatOption(tc.format)); err != nil {
			t.Fatalf("%s: %v", tc.id, err)
		}
		if buf.String() != string(want) {
			t.Errorf("%s:\n%s\nwant:\n%s", tc.id, buf.String(), want)
		}
	}
}

func TestTxtWithoutNodes(t *testing.T) {
	d := mustParse(t, "title Plan")
	var buf bytes.Buffer
	if _, err := d.ExportDiagram(&buf, 0, export.NewFileFormatOption(font.FormatATXT)); err == nil || buf.Len() != 0 {
		t.Errorf("error %v, %d bytes written", err, buf.Len())
	}
}
