// SPDX-License-Identifier: MIT
// Provenance: original

package activitydiagram3

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// The layout tests compare the positions of the nodes (boxes, circles,
// diamonds, bars, arrow heads, lines and texts) of the drawing with the
// coordinates of the reference SVG written by the upstream jar. They
// report the first node that moved, which the byte comparison of the
// export tests does not show. The fixtures are testdata/layout/*.puml with
// the reference *.svg next to them, and the curated conformance cases.

const layoutDir = "testdata/layout"

// layoutNode is a drawn node with its coordinates.
type layoutNode struct {
	kind   string
	coords []float64
	text   string
}

func (n layoutNode) String() string {
	var parts []string
	for _, c := range n.coords {
		parts = append(parts, strconv.FormatFloat(c, 'f', -1, 64))
	}
	s := n.kind + "(" + strings.Join(parts, ",") + ")"
	if n.text != "" {
		s += " " + strconv.Quote(n.text)
	}
	return s
}

var (
	layoutElement = regexp.MustCompile(`<(rect|ellipse|polygon|line|text|path)\b([^>]*)>([^<]*)`)
	layoutAttr    = regexp.MustCompile(`\b(x|y|width|height|cx|cy|rx|ry|x1|y1|x2|y2|points|d)="([^"]*)"`)
	layoutNumber  = regexp.MustCompile(`-?[0-9]+(\.[0-9]+)?`)
)

// layoutAttrs lists, for each element, the attributes that position it.
var layoutAttrs = map[string][]string{
	"rect":    {"x", "y", "width", "height"},
	"ellipse": {"cx", "cy", "rx", "ry"},
	"line":    {"x1", "y1", "x2", "y2"},
	"text":    {"x", "y"},
	"polygon": {"points"},
	"path":    {"d"},
}

// layoutNodes extracts the nodes of an SVG, in drawing order.
func layoutNodes(svg string) []layoutNode {
	var result []layoutNode
	for _, m := range layoutElement.FindAllStringSubmatch(svg, -1) {
		attrs := map[string]string{}
		for _, a := range layoutAttr.FindAllStringSubmatch(m[2], -1) {
			attrs[a[1]] = a[2]
		}
		node := layoutNode{kind: m[1]}
		for _, name := range layoutAttrs[m[1]] {
			for _, num := range layoutNumber.FindAllString(attrs[name], -1) {
				v, _ := strconv.ParseFloat(num, 64)
				node.coords = append(node.coords, v)
			}
		}
		if m[1] == "text" {
			node.text = m[3]
		}
		result = append(result, node)
	}
	return result
}

// compareLayout returns the first node that differs, "" when the nodes
// match within 0.001.
func compareLayout(got, want []layoutNode) string {
	for i := range max(len(got), len(want)) {
		if i >= len(got) {
			return fmt.Sprintf("node %d: missing %v", i, want[i])
		}
		if i >= len(want) {
			return fmt.Sprintf("node %d: extra %v", i, got[i])
		}
		g, w := got[i], want[i]
		same := g.kind == w.kind && g.text == w.text && len(g.coords) == len(w.coords)
		for j := 0; same && j < len(g.coords); j++ {
			same = math.Abs(g.coords[j]-w.coords[j]) < 0.001
		}
		if !same {
			return fmt.Sprintf("node %d: got %v, want %v", i, g, w)
		}
	}
	return ""
}

// layoutFixture is a source with its reference SVG.
type layoutFixture struct {
	name, source, svg string
}

func layoutFixtures(t *testing.T) []layoutFixture {
	t.Helper()
	var result []layoutFixture
	files, err := filepath.Glob(filepath.Join(layoutDir, "*.puml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		svg, err := os.ReadFile(strings.TrimSuffix(f, ".puml") + ".svg")
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, layoutFixture{name: strings.TrimSuffix(filepath.Base(f), ".puml"),
			source: string(src), svg: string(svg)})
	}
	for _, ref := range activityReferences(t, curatedReferenceRoot, conformanceRoot, "") {
		if _, known := exportExceptions[ref.name]; known {
			// The output is known to differ (see exportExceptions).
			continue
		}
		result = append(result, layoutFixture{name: ref.name, source: ref.source, svg: ref.svg})
	}
	slices.SortFunc(result, func(a, b layoutFixture) int { return strings.Compare(a.name, b.name) })
	return result
}

func TestLayoutNodePositions(t *testing.T) {
	fixtures := layoutFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("no layout fixture")
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			got, err := renderSVG(f.source)
			if err != nil {
				t.Fatal(err)
			}
			if diff := compareLayout(layoutNodes(got), layoutNodes(f.svg)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestLayoutNodesParsing(t *testing.T) {
	nodes := layoutNodes(`<rect x="1" y="2.5" width="3" height="4" rx="5"/><text x="7" y="8">a&amp;b</text>` +
		`<polygon points="1,2,3,4" fill="#000"/>`)
	want := []string{"rect(1,2.5,3,4)", `text(7,8) "a&amp;b"`, "polygon(1,2,3,4)"}
	if len(nodes) != len(want) {
		t.Fatalf("got %v", nodes)
	}
	for i, n := range nodes {
		if n.String() != want[i] {
			t.Errorf("node %d: %v, want %s", i, n, want[i])
		}
	}
	if d := compareLayout(nodes, nodes); d != "" {
		t.Error(d)
	}
	moved := slices.Clone(nodes)
	moved[0] = layoutNode{kind: "rect", coords: []float64{1, 2.6, 3, 4}}
	if d := compareLayout(moved, nodes); !strings.Contains(d, "node 0") {
		t.Errorf("moved node not reported: %q", d)
	}
}
