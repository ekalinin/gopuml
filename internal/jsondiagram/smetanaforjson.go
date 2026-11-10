// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/SmetanaForJson.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"context"
	stdjson "encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// layoutEngine lays a DOT source out and returns the output of a format,
// as dot.Gogv.Output.
type layoutEngine interface {
	Output(ctx context.Context, src, format string) ([]byte, error)
}

const (
	// veryFirstLine is the length of the straight start of the arrows,
	// between the dot and the spline.
	veryFirstLine = 13
	// plaintextPadX and plaintextPadY are the padding of a plaintext node
	// around its label, along the rows and across them, which gogv adds
	// as it supports neither margin nor fixedsize (design D3).
	plaintextPadX = 16
	plaintextPadY = 8
	// emptyNodeHeight is the size across the rows of upstream's node of a
	// box without rows, which Smetana sizes from its default label.
	emptyNodeHeight = 16
)

// internalNode is the box of an object or an array and its node in the
// layout: center and size, in points, in the top-to-bottom layout.
type internalNode struct {
	block   *TextBlockJson
	name    string
	isArray bool
	// dimension of the box and the columns and rows of its table
	width, height float64
	colA, colB    float64
	heights       []float64
	x, y, lw, lh  float64
}

// maxX is the top of the node in the layout, where the drawing starts.
func (n *internalNode) maxX() float64 { return n.y + n.lh/2 }

// jsonEdge is the link from the row port of a box to the box of its value.
type jsonEdge struct {
	tail, head *internalNode
	port       int
	curve      *jsonCurve
}

// smetanaForJSON lays the boxes of a JSON, YAML or HCL diagram out and draws
// them. Upstream gives its Smetana port a top-to-bottom graph of record
// nodes with the width and height of each box swapped, then turns the
// result a quarter; this release gives gogv the same graph, with tables of
// empty cells in place of the records (design D3).
type smetanaForJSON struct {
	skinParam     klimt.ISkinParam
	diagramType   klimt.SName
	stringBounder font.StringBounder
	nodes         []*internalNode
	edges         []*jsonEdge
	xMirror       Mirror
}

func newSmetanaForJSON(skinParam klimt.ISkinParam, diagramType klimt.SName,
	stringBounder font.StringBounder) *smetanaForJSON {
	return &smetanaForJSON{skinParam: skinParam, diagramType: diagramType, stringBounder: stringBounder}
}

func (s *smetanaForJSON) styleArrow() *klimt.Style {
	return styleQuery(s.diagramType, klimt.SNameArrow).MergedStyle(s.skinParam.CurrentStyleBuilder())
}

func (s *smetanaForJSON) styleNode() *klimt.Style {
	return styleQuery(s.diagramType, klimt.SNameNode).MergedStyle(s.skinParam.CurrentStyleBuilder())
}

// manageOneNode creates the node of current, then for each row whose value
// is an object or an array the nodes of that value and the edge to it, in
// upstream's order.
func (s *smetanaForJSON) manageOneNode(current json.Value, highlighted []*Highlighted) *internalNode {
	block := NewTextBlockJson(s.skinParam, s.diagramType, current, highlighted)
	dim := block.CalculateDimension(s.stringBounder)
	node := &internalNode{block: block, name: "N" + strconv.Itoa(len(s.nodes)), isArray: current.IsArray(),
		width: dim.Width(), height: dim.Height(), colA: block.widthColA(s.stringBounder),
		colB: block.widthColB(s.stringBounder), heights: block.allHeights(s.stringBounder)}
	s.nodes = append(s.nodes, node)
	keys := block.keys()
	for i, child := range block.children() {
		if child != nil {
			childNode := s.manageOneNode(child, upOneLevel(keys[i], highlighted))
			s.edges = append(s.edges, &jsonEdge{tail: node, head: childNode, port: i})
		}
	}
	return node
}

func upOneLevel(key string, list []*Highlighted) []*Highlighted {
	var result []*Highlighted
	for _, tmp := range list {
		if parent := tmp.upOneLevel(key); parent != nil {
			result = append(result, parent)
		}
	}
	return result
}

// dotSource returns the graph of the nodes and edges, in their order of
// creation.
func (s *smetanaForJSON) dotSource() string {
	var sb strings.Builder
	sb.WriteString("digraph g {\n")
	for _, n := range s.nodes {
		sb.WriteString(n.name + " [shape=plaintext, width=" + jfmt.DoubleToString(n.height/72) + ", height=")
		if len(n.heights) == 0 {
			sb.WriteString(jfmt.DoubleToString(emptyNodeHeight/72.0) + `, label=""];` + "\n")
			continue
		}
		sb.WriteString(jfmt.DoubleToString(n.width/72) + ", label=<" + n.tableLabel() + ">];\n")
	}
	for _, e := range s.edges {
		fmt.Fprintf(&sb, "%s -> %s [tailport=P%d, arrowsize=.75, arrowtail=none, arrowhead=normal];\n",
			e.tail.name, e.head.name, e.port)
	}
	sb.WriteString("}\n")
	return sb.String()
}

// tableLabel is the table that stands for upstream's record: along the
// rows, a cell per port P0 ... P<n-1>; for an object, a first row with one
// cell for the key column. The cells are smaller than the box by the
// padding of plaintext, so that the padded node has the size of the record
// (design D3).
func (n *internalNode) tableLabel() string {
	widths := portCellWidths(n.heights)
	var cells strings.Builder
	total := 0
	var across int
	if n.isArray {
		across = cellSize(n.colA - plaintextPadY)
	} else {
		across = cellSize(n.colB - plaintextPadY/2)
	}
	for i, w := range widths {
		total += w
		fmt.Fprintf(&cells, `<TD PORT="P%d" FIXEDSIZE="TRUE" WIDTH="%d" HEIGHT="%d"></TD>`, i, w, across)
	}
	var sb strings.Builder
	sb.WriteString(`<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0">`)
	if !n.isArray {
		fmt.Fprintf(&sb, `<TR><TD COLSPAN="%d" FIXEDSIZE="TRUE" WIDTH="%d" HEIGHT="%d"></TD></TR>`, len(widths), total,
			cellSize(n.colA-plaintextPadY/2))
	}
	sb.WriteString("<TR>" + cells.String() + "</TR></TABLE>")
	return sb.String()
}

// portCellWidths returns the whole sizes of the port cells along the rows:
// each row ends where the rounded sum of the heights up to it ends, and the
// first and the last cells leave room for the padding of plaintext.
func portCellWidths(heights []float64) []int {
	result := make([]int, len(heights))
	sum := 0.0
	for i, h := range heights {
		start := roundHalfUp(sum)
		sum += h
		result[i] = roundHalfUp(sum) - start
	}
	result[0] -= plaintextPadX / 2
	result[len(result)-1] -= plaintextPadX / 2
	for i := range result {
		result[i] = max(result[i], 1)
	}
	return result
}

// cellSize is a size rounded to a whole number of points, at least 1.
func cellSize(v float64) int { return max(roundHalfUp(v), 1) }

func roundHalfUp(v float64) int { return int(math.Floor(v + 0.5)) }

// json0Layout is the part of gogv's json0 output that the drawing reads.
type json0Layout struct {
	Objects []struct {
		GVID   int    `json:"_gvid"`
		Name   string `json:"name"`
		Pos    string `json:"pos"`
		Width  string `json:"width"`
		Height string `json:"height"`
	} `json:"objects"`
	Edges []struct {
		Tail     int    `json:"tail"`
		Head     int    `json:"head"`
		TailPort string `json:"tailport"`
		Pos      string `json:"pos"`
	} `json:"edges"`
}

// layout builds the graph of root and lays it out with engine.
func (s *smetanaForJSON) layout(ctx context.Context, engine layoutEngine, root json.Value,
	highlighted []*Highlighted) error {
	s.manageOneNode(root, highlighted)
	out, err := engine.Output(ctx, s.dotSource(), "json0")
	if err != nil {
		return err
	}
	return s.readLayout(out)
}

// readLayout reads the positions of the nodes and the splines of the edges
// from gogv's json0 output.
func (s *smetanaForJSON) readLayout(out []byte) error {
	var layout json0Layout
	if err := stdjson.Unmarshal(out, &layout); err != nil {
		return fmt.Errorf("json0 layout: %w", err)
	}
	byName := make(map[string]*internalNode, len(s.nodes))
	for _, n := range s.nodes {
		byName[n.name] = n
	}
	byGVID := make(map[int]*internalNode, len(s.nodes))
	for _, o := range layout.Objects {
		n := byName[o.Name]
		if n == nil {
			continue
		}
		byGVID[o.GVID] = n
		pos, err := parsePoint(o.Pos)
		if err != nil {
			return fmt.Errorf("json0 layout of %s: %w", o.Name, err)
		}
		w, errW := strconv.ParseFloat(o.Width, 64)
		h, errH := strconv.ParseFloat(o.Height, 64)
		if errW != nil || errH != nil {
			return fmt.Errorf("json0 layout of %s: size %q by %q", o.Name, o.Width, o.Height)
		}
		n.x, n.y, n.lw, n.lh = pos.X, pos.Y, w*72, h*72
	}
	maxX := 0.0
	for _, n := range s.nodes {
		maxX = max(maxX, n.maxX())
	}
	s.xMirror = Mirror{max: maxX}
	for _, e := range s.edges {
		pos := ""
		for _, le := range layout.Edges {
			if byGVID[le.Tail] == e.tail && byGVID[le.Head] == e.head && le.TailPort == "P"+strconv.Itoa(e.port) {
				pos = le.Pos
				break
			}
		}
		curve, err := newJSONCurve(pos, s.xMirror, veryFirstLine)
		if err != nil {
			return fmt.Errorf("json0 layout of %s -> %s: %w", e.tail.name, e.head.name, err)
		}
		e.curve = curve
	}
	return nil
}

// drawMe draws the boxes, then the arrows with their dots, as upstream.
func (s *smetanaForJSON) drawMe(ug klimt.UGraphic) {
	set := s.skinParam.IHtmlColorSet()
	for _, n := range s.nodes {
		n.block.DrawU(s.styleNode().ApplyStrokeAndLineColor(ug, set).Apply(s.position(n)))
	}
	styleArrow := s.styleArrow()
	hcolor := styleArrow.Value(klimt.PNameLineColor).AsColor(set)
	for _, e := range s.edges {
		e.curve.drawCurve(hcolor, styleArrow.ApplyStrokeAndLineColor(ug, set))
		e.curve.drawSpot(styleArrow.ApplyStrokeAndLineColor(ug, set).Apply(klimt.Bg(hcolor)))
	}
}

// position is the top left corner of the box of n in the drawing.
func (s *smetanaForJSON) position(n *internalNode) klimt.UTranslate {
	return klimt.NewUTranslate(n.x-n.lw/2, s.xMirror.inv(n.y+n.lh/2)).Sym()
}
