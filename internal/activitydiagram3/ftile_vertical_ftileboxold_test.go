// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"fmt"
	"testing"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// boxRec records the rectangles and the texts drawn, with the translation
// and the background color they are drawn with.
type boxRec struct {
	klimt.UGraphic
	log    *[]string
	dx, dy float64
	back   color.HColor
}

func (r *boxRec) StringBounder() font.StringBounder { return font.NewStringBounderSvg(nil) }

func (r *boxRec) Apply(change klimt.UChange) klimt.UGraphic {
	c := *r
	switch ch := change.(type) {
	case klimt.UTranslate:
		c.dx += ch.Dx()
		c.dy += ch.Dy()
	case klimt.UBackground:
		c.back = ch.BackColor()
	}
	return &c
}

func (r *boxRec) Draw(shape klimt.UShape) {
	switch s := shape.(type) {
	case *klimt.URectangle:
		*r.log = append(*r.log, fmt.Sprintf("rect %.3f,%.3f %.3fx%.3f rx=%v %s", r.dx, r.dy, s.Width(), s.Height(),
			s.Rx(), r.back.AsString()))
	case *klimt.UText:
		*r.log = append(*r.log, fmt.Sprintf("text %.3f %s", r.dx, s.Text()))
	}
}

// mindMapBox returns the box of label with the node style of a mindmap
// diagram, parsed from an activity diagram with the lines before; back, when
// set, is the background color of the skin parameters.
func mindMapBox(t *testing.T, before, label, back string) klimt.TextBlock {
	t.Helper()
	d, _, err := parseCase("@startuml\n" + before + "start\n@enduml\n")
	if err != nil {
		t.Fatal(err)
	}
	var skinParam klimt.ISkinParam = d.(*ActivityDiagram3).SkinParam()
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameMindmapDiagram,
		klimt.SNameNode).MergedStyle(skinParam.CurrentStyleBuilder())
	if back != "" {
		c, err := skinParam.IHtmlColorSet().GetColor(back)
		if err != nil {
			t.Fatal(err)
		}
		skinParam = klimt.NewSkinParamColors(skinParam, color.EmptyColors().Add(color.TypeBack, c))
	}
	return FtileBoxOldCreateMindMap(style, skinParam, klimt.DisplayCreate(label))
}

func TestFtileBoxOld(t *testing.T) {
	// The sizes are those of the mindmap references. "Node 1" is 49.896
	// wide, so it is drawn at 100 - 10 - 49.896 on the right of a box of
	// MinimumWidth 100 with a padding of 10, and at 10 + (80 - 49.896) / 2
	// in the middle. The corner size is twice the radius of the SVG.
	nodeStyle := func(lines string) string {
		return "<style>\nmindmapDiagram {\n node {\n" + lines + " }\n}\n</style>\n"
	}
	for _, tc := range []struct {
		name, before, label, back string
		dim                       string
		log                       []string
	}{
		{"default style", "", "a", "", "28.579x36.297",
			[]string{"rect 0.000,0.000 28.579x36.297 rx=25 #F1F1F1", "text 10.000 a"}},
		// The padding of the skin parameters does not count.
		{"skinparam padding", "skinparam padding 5\n", "a", "", "28.579x36.297",
			[]string{"rect 0.000,0.000 28.579x36.297 rx=25 #F1F1F1", "text 10.000 a"}},
		{"style padding", nodeStyle("  Padding 4\n"), "a", "", "16.579x24.297",
			[]string{"rect 0.000,0.000 16.579x24.297 rx=25 #F1F1F1", "text 4.000 a"}},
		{"minimum width", nodeStyle("  MinimumWidth 100\n"), "Node 1", "", "100.000x36.297",
			[]string{"rect 0.000,0.000 100.000x36.297 rx=25 #F1F1F1", "text 10.000 Node 1"}},
		{"right", nodeStyle("  MinimumWidth 100\n  HorizontalAlignment Right\n"), "Node 1", "", "100.000x36.297",
			[]string{"rect 0.000,0.000 100.000x36.297 rx=25 #F1F1F1", "text 40.104 Node 1"}},
		{"center", nodeStyle("  MinimumWidth 100\n  HorizontalAlignment Center\n"), "Node 2", "", "100.000x36.297",
			[]string{"rect 0.000,0.000 100.000x36.297 rx=25 #F1F1F1", "text 25.052 Node 2"}},
		// The color of SkinParamColors goes over the style's.
		{"background color", nodeStyle("  BackGroundColor lightgray\n"), "a", "#Orange", "28.579x36.297",
			[]string{"rect 0.000,0.000 28.579x36.297 rx=25 #FFA500", "text 10.000 a"}},
		{"round corner", nodeStyle("  RoundCorner 8\n"), "a", "", "28.579x36.297",
			[]string{"rect 0.000,0.000 28.579x36.297 rx=8 #F1F1F1", "text 10.000 a"}},
	} {
		box := mindMapBox(t, tc.before, tc.label, tc.back)
		dim := box.CalculateDimension(font.NewStringBounderSvg(nil))
		if got := fmt.Sprintf("%.3fx%.3f", dim.Width(), dim.Height()); got != tc.dim {
			t.Errorf("%s: dimension %s, want %s", tc.name, got, tc.dim)
		}
		r := &boxRec{UGraphic: klimt.NewUGraphicNull(), log: new([]string)}
		box.DrawU(r)
		if fmt.Sprint(*r.log) != fmt.Sprint(tc.log) {
			t.Errorf("%s: drawn %q, want %q", tc.name, *r.log, tc.log)
		}
	}
}
