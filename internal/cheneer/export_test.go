// SPDX-License-Identifier: MIT
// Provenance: original

package cheneer

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// TestExportSVG draws the curated case "Entities, relationship and
// attributes" with the local dot: the shapes are upstream's, each element in
// a bare <g>, as the references show.
func TestExportSVG(t *testing.T) {
	d := parse(t, `entity CUSTOMER {`, `Number`, `}`, `entity MOVIE {`, `Code`, `}`, `relationship RENTED_TO {`,
		`Date`, `}`, `RENTED_TO -1- CUSTOMER`, `RENTED_TO -N- MOVIE`)
	var buf bytes.Buffer
	_, err := d.ExportDiagram(&buf, 0, export.NewFileFormatOptionMetadata(font.FormatSVG, true))
	if errors.Is(err, dot.ErrNotFound) {
		t.Skip("no dot")
	}
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, c := range []struct {
		pattern string
		count   int
	}{
		{`data-diagram-type="CHEN_EER"`, 1},
		{`<g><rect [^>]*/><text [^>]*>CUSTOMER</text></g>`, 1},
		{`<g><rect [^>]*/><text [^>]*>MOVIE</text></g>`, 1},
		{`<g><polygon [^>]*/><text [^>]*>RENTED_TO</text></g>`, 1},
		{`<g><ellipse [^>]*/><text [^>]*>(Number|Code|Date)</text></g>`, 3},
		{`<g class="link" [^>]*><path [^>]*id="RENTED_TO-CUSTOMER"/><text [^>]*>1</text></g>`, 1},
		{`<g class="link" [^>]*><path [^>]*id="RENTED_TO-MOVIE"/><text [^>]*>N</text></g>`, 1},
		{`id="CUSTOMER/Number-CUSTOMER"`, 1},
	} {
		if n := len(regexp.MustCompile(c.pattern).FindAllString(got, -1)); n != c.count {
			t.Errorf("%d matches of %s, want %d", n, c.pattern, c.count)
		}
	}
	if strings.Contains(got, `class="entity"`) {
		t.Error("the elements are in entity groups")
	}
}
