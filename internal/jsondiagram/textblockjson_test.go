// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"math"
	"slices"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// newTestBlock returns the box of the JSON text src in a JSON diagram.
func newTestBlock(t *testing.T, src string, highlighted ...*Highlighted) *TextBlockJson {
	t.Helper()
	root, err := json.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	sp := klimt.SkinParamCreate(map[string]string{}, nil, core.JSON, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	return NewTextBlockJson(sp, klimt.SNameJsonDiagram, root, highlighted)
}

func TestTextBlockJsonObject(t *testing.T) {
	b := newTestBlock(t, `{"a": 1, "bb": {}}`)
	if !slices.Equal(b.keys(), []string{"a", "bb"}) {
		t.Errorf("keys %q, want a and bb", b.keys())
	}
	children := b.children()
	if len(children) != 2 || children[0] != nil || children[1] == nil || !children[1].IsObject() {
		t.Errorf("children %v, want nil and the object {}", children)
	}
	if got := []string{shortString(json.ValueOfInt(1)), shortString(json.NewObject())}; !slices.Equal(got, []string{"1", "   "}) {
		t.Errorf("values %q, want 1 and three spaces", got)
	}
	sb := font.NewStringBounderSvg(nil)
	dim := b.CalculateDimension(sb)
	if len(b.lines) != 2 || math.Abs(dim.Width()-(b.widthColA(sb)+b.widthColB(sb))) > 1e-9 || b.widthColB(sb) == 0 {
		t.Errorf("%d rows, width %v, columns %v and %v", len(b.lines), dim.Width(), b.widthColA(sb), b.widthColB(sb))
	}
	if heights := b.allHeights(sb); len(heights) != 2 || math.Abs(dim.Height()-(heights[0]+heights[1])) > 1e-9 {
		t.Errorf("height %v, row heights %v", dim.Height(), heights)
	}
}

func TestTextBlockJsonArray(t *testing.T) {
	b := newTestBlock(t, `[true, null]`)
	if len(b.lines) != 2 || b.lines[0].b2 != nil || b.lines[1].b2 != nil {
		t.Fatalf("%d rows, want two rows of one column", len(b.lines))
	}
	if got := []string{shortString(json.True), shortString(json.Null), shortString(json.False)}; !slices.Equal(got,
		[]string{"<U+2611> true", "<U+2400>", "<U+2610> false"}) {
		t.Errorf("values %q", got)
	}
	if !slices.Equal(b.keys(), []string{"0", "1"}) {
		t.Errorf("keys %q, want the indexes 0 and 1", b.keys())
	}
	if b.widthColB(font.NewStringBounderSvg(nil)) != 0 {
		t.Error("an array has a value column")
	}
}

func TestTextBlockJsonEmpty(t *testing.T) {
	b := newTestBlock(t, `[]`)
	dim := b.CalculateDimension(font.NewStringBounderSvg(nil))
	if len(b.lines) != 0 || dim.Width() != 0 || dim.Height() != textBlockJSONMinHeight {
		t.Errorf("%d rows, size %v, want no row, the width 0 and the height 15", len(b.lines), dim)
	}
}

func TestTextBlockJsonHighlighted(t *testing.T) {
	b := newTestBlock(t, `{"firstName": "John", "lastName": "Smith"}`, BuildHighlighted(`#highlight "lastName"`))
	if b.lines[0].highlighted != nil || b.lines[1].highlighted == nil {
		t.Errorf("highlighted rows %v and %v, want the second only", b.lines[0].highlighted, b.lines[1].highlighted)
	}
}
