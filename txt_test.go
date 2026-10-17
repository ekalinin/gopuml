// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRenderSequenceAsTXT(t *testing.T) {
	// Spec scenario "ASCII sequence diagram"; the texts are the reference
	// jar's.
	for format, want := range map[Format]string{
		TXT: ",-----.   ,---.\n|Alice|   |Bob|\n`--+--'   `-+-'\n   | hello  |\n   |------->|\n   |        |\n" +
			",--+--.   ,-+-.\n|Alice|   |Bob|\n`-----'   `---'\n",
		UTXT: "┌─────┐   ┌───┐\n│Alice│   │Bob│\n└──┬──┘   └─┬─┘\n   │ hello  │\n   │───────>│\n   │        │\n" +
			"┌──┴──┐   ┌─┴─┐\n│Alice│   │Bob│\n└─────┘   └───┘\n",
	} {
		res, err := Render(context.Background(), seqSource, &Options{Format: format})
		if err != nil {
			t.Fatalf("%v: %v", format, err)
		}
		if string(res.Data) != want {
			t.Errorf("%v: got\n%s\nwant\n%s", format, res.Data, want)
		}
		if res.Description != "(2 participants)" || res.Width != 0 || res.Height != 0 {
			t.Errorf("%v: metadata %+v", format, res)
		}
	}
}

func TestRenderClassDiagramAsTXTIsUnsupported(t *testing.T) {
	// Spec scenario "TXT for a class diagram": upstream draws it, this
	// release reports that the format is not supported for the type.
	src := "@startuml\nclass Order\nOrder --> Customer\n@enduml\n"
	for _, format := range []Format{TXT, UTXT} {
		res, err := Render(context.Background(), src, &Options{Format: format})
		if !errors.Is(err, ErrUnsupportedFormat) {
			t.Fatalf("%v: err = %v, want ErrUnsupportedFormat", format, err)
		}
		if msg := err.Error(); !strings.Contains(msg, "TXT output is not supported for class diagrams") {
			t.Errorf("%v: message %q does not name the format and the diagram type", format, msg)
		}
		if res != nil {
			t.Errorf("%v: got a result %+v, want none", format, res)
		}
	}
}
