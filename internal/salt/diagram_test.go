// SPDX-License-Identifier: MIT
// Provenance: original

package salt

import (
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
)

func TestNeedlessSaltKeyword(t *testing.T) {
	d, ok := parseSource(t, "@startsalt\nsalt\n{\n[OK]\n}\n@endsalt\n").(*core.ErrorDiagram)
	if !ok {
		t.Fatal("a first line salt does not fail")
	}
	if d.Error.Error != "This is not needed anymore." {
		t.Errorf("error %q, want This is not needed anymore.", d.Error.Error)
	}
	// After the first line, salt is a text.
	s := parseSalt(t, "@startsalt\n{\nName\nsalt\n}\n@endsalt\n")
	if got := strings.Join(s.manageSprite(), "|"); got != "{|Name|salt|}" {
		t.Errorf("lines %q, want {|Name|salt|}", got)
	}
}

func TestRowReadAsCommand(t *testing.T) {
	d := parseSalt(t, "@startsalt\n{\nTitle | \"text\"\nName | \"me\"\n}\n@endsalt\n")
	if got := d.TitleDisplay().String(); !strings.Contains(got, `| "text"`) {
		t.Errorf("title %q, want | \"text\"", got)
	}
	if got := strings.Join(d.manageSprite(), "|"); got != `{|Name | "me"|}` {
		t.Errorf("lines %q, want the row of Name only", got)
	}
}

func TestSaltSprites(t *testing.T) {
	d := parseSalt(t, "@startsalt\nsprite $dot [4x4/16] {\n0FF0\nFFFF\nFFFF\n0FF0\n}\n{\n<$dot> Ready\n}\n@endsalt\n")
	if d.dictionary.Sprite("dot") == nil {
		t.Error("the sprite did not reach the dictionary")
	}
	if d.SkinParam().Sprite("dot") != nil {
		t.Error("the sprite reached the skin parameters")
	}
}

func TestSaltDescriptionAndMargins(t *testing.T) {
	d := parseSalt(t, "@startsalt\n{\n[OK]\n}\n@endsalt\n")
	if d.Description() != "(Salt)" || d.DiagramType() != core.Salt {
		t.Errorf("description %q and type %v, want (Salt) and salt", d.Description(), d.DiagramType())
	}
	if m := d.DefaultMargins(); m.Top() != 5 || m.Left() != 5 {
		t.Errorf("margins %v, want 5", m)
	}
}
