// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"context"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// newTestSmetana returns the nodes and edges of the JSON text src.
func newTestSmetana(t *testing.T, src string) *smetanaForJSON {
	t.Helper()
	root, err := json.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	sp := klimt.SkinParamCreate(map[string]string{}, nil, core.JSON, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	s := newSmetanaForJSON(sp, klimt.SNameJsonDiagram, font.NewStringBounderSvg(nil))
	s.manageOneNode(root, nil)
	return s
}

func TestPortCellWidths(t *testing.T) {
	heights := make([]float64, 8)
	for i := range heights {
		heights[i] = 20.297
	}
	if got := portCellWidths(heights); !slices.Equal(got, []int{12, 21, 20, 20, 20, 21, 20, 12}) {
		t.Errorf("8 rows of 20.297: %v, want 12 21 20 20 20 21 20 12", got)
	}
	if got := portCellWidths([]float64{20.297}); !slices.Equal(got, []int{4}) {
		t.Errorf("1 row of 20.297: %v, want 4", got)
	}
}

func TestDotSource(t *testing.T) {
	s := newTestSmetana(t, `{"a": {"b": 1}}`)
	src := s.dotSource()
	lines := strings.Split(strings.TrimSpace(src), "\n")
	if len(lines) != 5 || lines[0] != "digraph g {" || !strings.HasPrefix(lines[1], "N0 [") ||
		!strings.HasPrefix(lines[2], "N1 [") || lines[4] != "}" {
		t.Fatalf("DOT %s", src)
	}
	if want := "N0 -> N1 [tailport=P0, arrowsize=.75, arrowtail=none, arrowhead=normal];"; lines[3] != want {
		t.Errorf("edge %q, want %q", lines[3], want)
	}
	n := s.nodes[0]
	if want := "N0 [shape=plaintext, width=" + jfmt.DoubleToString(n.height/72) + ", height=" +
		jfmt.DoubleToString(n.width/72) + ", label=<"; !strings.HasPrefix(lines[1], want) {
		t.Errorf("node %q, want the width and the height of the box swapped: %q", lines[1], want)
	}
	// One row: the port cell is 16 points shorter than the row; the key and
	// value cells are 4 points narrower than their columns.
	m := regexp.MustCompile(`<TR><TD COLSPAN="1" FIXEDSIZE="TRUE" WIDTH="(\d+)" HEIGHT="(\d+)"></TD></TR>` +
		`<TR><TD PORT="P0" FIXEDSIZE="TRUE" WIDTH="(\d+)" HEIGHT="(\d+)"></TD></TR>`).FindStringSubmatch(lines[1])
	if m == nil {
		t.Fatalf("node %q has no table of a key cell and a port cell", lines[1])
	}
	if want := []string{m[3], strconv.Itoa(roundHalfUp(n.heights[0]) - 16), strconv.Itoa(roundHalfUp(n.colA - 4)),
		strconv.Itoa(roundHalfUp(n.colB - 4))}; m[1] != want[0] || m[3] != want[1] || m[2] != want[2] || m[4] != want[3] {
		t.Errorf("cells %q, want the widths %s %s and the heights %s %s", m[1:], want[0], want[1], want[2], want[3])
	}
}

func TestDotSourceEmptyValue(t *testing.T) {
	s := newTestSmetana(t, `{"e": []}`)
	if want := `N1 [shape=plaintext, width=` + jfmt.DoubleToString(15.0/72) + `, height=` +
		jfmt.DoubleToString(16.0/72) + `, label=""];`; !strings.Contains(s.dotSource(), want) {
		t.Errorf("DOT %s, want %s", s.dotSource(), want)
	}
}

func TestDotSourceArray(t *testing.T) {
	s := newTestSmetana(t, `["x", "y"]`)
	n := s.nodes[0]
	want := `<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD PORT="P0" FIXEDSIZE="TRUE" ` +
		`WIDTH="` + strconv.Itoa(roundHalfUp(n.heights[0])-8) + `" HEIGHT="` + strconv.Itoa(roundHalfUp(n.colA-8)) + `"></TD>`
	if !strings.Contains(s.dotSource(), want) {
		t.Errorf("DOT %s, want one row of port cells as high as the box less 8: %s", s.dotSource(), want)
	}
}

func TestEdgeOrder(t *testing.T) {
	s := newTestSmetana(t, `{"a": {"b": 1}, "c": [{}]}`)
	var got []string
	for _, e := range s.edges {
		got = append(got, e.tail.name+"->"+e.head.name+":"+strconv.Itoa(e.port))
	}
	if want := []string{"N0->N1:0", "N2->N3:0", "N0->N2:1"}; !slices.Equal(got, want) {
		t.Errorf("edges %q, want %q", got, want)
	}
}

func TestReadLayout(t *testing.T) {
	s := newTestSmetana(t, `{"a": {"b": 1}}`)
	out := `{"name": "g", "objects": [{"_gvid": 0, "name": "N0", "pos": "20,68.4", "width": "0.55556", "height": "0.5"},
	  {"_gvid": 1, "name": "N1", "pos": "26,7.2", "width": "0.3", "height": "0.2"}],
	  "edges": [{"_gvid": 0, "tail": 0, "head": 1, "tailport": "P0", "pos": "e,26,15.914 26,58.4 26,49.053 26,34.588 26,23.414"}]}`
	if err := s.readLayout([]byte(out)); err != nil {
		t.Fatal(err)
	}
	n0, n1 := s.nodes[0], s.nodes[1]
	if n0.x != 20 || n0.y != 68.4 || math.Abs(n0.lw-40) > 1e-3 || n0.lh != 36 || n1.x != 26 || n1.y != 7.2 {
		t.Errorf("nodes %+v and %+v", n0, n1)
	}
	if s.xMirror.max != 86.4 {
		t.Errorf("mirror %v, want the top of N0, 86.4", s.xMirror.max)
	}
	c := s.edges[0].curve
	if c == nil || len(c.points) != 4 || c.ep == nil || c.ep.X != 26 || c.ep.Y != 15.914 {
		t.Errorf("curve %+v, want 4 points and the end point (26, 15.914)", c)
	}
}

// fakeEngine returns the json0 output of gogv for the DOT it is given.
type fakeEngine struct{ src string }

func (f *fakeEngine) Output(ctx context.Context, src, format string) ([]byte, error) {
	f.src = src
	return (&dot.Gogv{}).Output(ctx, src, format)
}

func TestLayoutWithGogv(t *testing.T) {
	root, err := json.Parse(`{"a": {"b": 1}, "c": [1, 2]}`)
	if err != nil {
		t.Fatal(err)
	}
	sp := klimt.SkinParamCreate(map[string]string{}, nil, core.JSON, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	s := newSmetanaForJSON(sp, klimt.SNameJsonDiagram, font.NewStringBounderSvg(nil))
	engine := &fakeEngine{}
	if err := s.layout(t.Context(), engine, root, nil); err != nil {
		t.Fatal(err)
	}
	if engine.src != s.dotSource() {
		t.Error("the engine was not given the DOT source of the nodes")
	}
	// The root is on the left, the two boxes of its values to its right,
	// one above the other, centered on their rank.
	p0, p1, p2 := s.position(s.nodes[0]), s.position(s.nodes[1]), s.position(s.nodes[2])
	center1, center2 := p1.Dx()+s.nodes[1].width/2, p2.Dx()+s.nodes[2].width/2
	if p0.Dx() != 0 || p1.Dx() <= s.nodes[0].width || p2.Dx() <= s.nodes[0].width || p2.Dy() <= p1.Dy() ||
		math.Abs(center1-center2) > 1 {
		t.Errorf("positions %v, %v and %v", p0, p1, p2)
	}
	for _, e := range s.edges {
		if e.curve == nil || e.curve.ep == nil {
			t.Errorf("edge %s -> %s has no curve with an end point", e.tail.name, e.head.name)
		}
	}
}
