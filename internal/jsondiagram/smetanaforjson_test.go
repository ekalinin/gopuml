// SPDX-License-Identifier: MIT
// Provenance: original

package jsondiagram

import (
	"context"
	"math"
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

func TestRecord(t *testing.T) {
	// The root of the Vega JSON case: eight rows of 20.297 points, a key
	// column of 130.094 and a value column of 55.732. Upstream draws the
	// ports of rows 4, 5 and 6 at 10, 30 and 51 points from the center.
	n := &internalNode{width: 185.826, colA: 130.094, colB: 55.732}
	for range 8 {
		n.heights = append(n.heights, 20.297)
	}
	r := n.record()
	if math.Abs(r.width-162.376) > 1e-9 || r.height != 186 || r.portsHeight != 56 {
		t.Errorf("record %v by %v, ports field %d high, want 162.376 by 186 and 56", r.width, r.height, r.portsHeight)
	}
	if !slices.Equal(r.ports, []int{-71, -51, -30, -10, 10, 30, 51, 71}) {
		t.Errorf("ports %v", r.ports)
	}
	// An array: the ports field is as high as the record.
	a := (&internalNode{isArray: true, width: 23.351, colA: 23.351, heights: []float64{20.297, 20.297}}).record()
	if math.Abs(a.width-41) > 1e-9 || a.height != 23 || a.portsHeight != 23 || !slices.Equal(a.ports, []int{-10, 10}) {
		t.Errorf("array record %+v", a)
	}
}

func TestTableLabel(t *testing.T) {
	r := (&internalNode{width: 40, colA: 20, colB: 20, heights: []float64{20.297}}).record()
	want := `<TABLE PORT="h" BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0">` +
		`<TR><TD COLSPAN="3" FIXEDSIZE="TRUE" WIDTH="20" HEIGHT="20"></TD></TR>` +
		`<TR><TD FIXEDSIZE="TRUE" WIDTH="9" HEIGHT="20"></TD><TD PORT="P0" FIXEDSIZE="TRUE" WIDTH="2" HEIGHT="20"></TD>` +
		`<TD FIXEDSIZE="TRUE" WIDTH="9" HEIGHT="20"></TD></TR></TABLE>`
	if got := r.tableLabel(); got != want {
		t.Errorf("table\n%s\nwant\n%s", got, want)
	}
}

func TestDotSource(t *testing.T) {
	s := newTestSmetana(t, `{"a": {"b": 1}}`)
	src := s.dotSource()
	lines := strings.Split(strings.TrimSpace(src), "\n")
	if len(lines) != 6 || lines[0] != "digraph g {" || !strings.HasPrefix(lines[2], "N0 [") ||
		!strings.HasPrefix(lines[3], "N1 [") || lines[5] != "}" {
		t.Fatalf("DOT %s", src)
	}
	if want := "graph [nodesep=" + jfmt.DoubleToString(2.0/72) + ", ranksep=" + jfmt.DoubleToString(28.0/72) + "];"; lines[1] != want {
		t.Errorf("graph %q, want %q", lines[1], want)
	}
	if want := "N0 -> N1 [tailport=P0, headport=h, arrowsize=.75, arrowtail=none, arrowhead=normal];"; lines[4] != want {
		t.Errorf("edge %q, want %q", lines[4], want)
	}
	r := s.nodes[0].record()
	if want := "N0 [shape=plaintext, width=" + jfmt.DoubleToString((r.width+16)/72) + ", height=" +
		jfmt.DoubleToString((r.height+1+8)/72) + ", label=<" + r.tableLabel() + ">];"; lines[2] != want {
		t.Errorf("node %q, want the record of upstream with the padding: %q", lines[2], want)
	}
}

func TestDotSourceEmptyValue(t *testing.T) {
	s := newTestSmetana(t, `{"e": []}`)
	want := `N1 [shape=plaintext, width=` + jfmt.DoubleToString(31.0/72) + `, height=` +
		jfmt.DoubleToString(25.0/72) + `, label=<<TABLE PORT="h" BORDER="0" CELLBORDER="0" CELLSPACING="0" ` +
		`CELLPADDING="0"><TR><TD FIXEDSIZE="TRUE" WIDTH="15" HEIGHT="16"></TD></TR></TABLE>>];`
	if !strings.Contains(s.dotSource(), want) {
		t.Errorf("DOT %s, want %s", s.dotSource(), want)
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
	// Without the padding: 8 points less along the rows, the size less 16
	// along the rows and 8 across.
	if n0.x != 12 || n0.y != 68.4 || math.Abs(n0.lw-24) > 1e-3 || n0.lh != 28 || n1.x != 18 || n1.y != 7.2 {
		t.Errorf("nodes %+v and %+v", n0, n1)
	}
	if s.xMirror.max != 82.4 {
		t.Errorf("mirror %v, want the top of N0, 82.4", s.xMirror.max)
	}
	c := s.edges[0].curve
	if c == nil || len(c.points) != 4 || c.points[0].X != 18 || c.ep == nil || c.ep.X != 18 || c.ep.Y != 15.914 {
		t.Errorf("curve %+v, want 4 points from x 18 and the end point (18, 15.914)", c)
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
