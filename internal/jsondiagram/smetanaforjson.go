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
	// nodeSep and rankSep are the separations of Graphviz's defaults, which
	// Smetana keeps, less the padding of the nodes on both sides, so that
	// the gaps between the boxes are Smetana's (design D3).
	nodeSep = 18 - plaintextPadX
	rankSep = 36 - plaintextPadY
	// emptyNodeWidth and emptyNodeHeight are the size of upstream's node of
	// a box without rows, along the rows and across them, which Smetana
	// sizes from its default label.
	emptyNodeWidth  = textBlockJSONMinHeight
	emptyNodeHeight = 16
	// headPort is the port of the table of every node.
	headPort = "h"
	// portCellWidth is the width of the cell of a port along the rows.
	portCellWidth = 2
)

// internalNode is the box of an object or an array and its node in the
// layout: center and size, in points, in the top-to-bottom layout, without
// the padding of plaintext.
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
	sb.WriteString("graph [nodesep=" + jfmt.DoubleToString(nodeSep/72.0) + ", ranksep=" +
		jfmt.DoubleToString(rankSep/72.0) + "];\n")
	for _, n := range s.nodes {
		r := n.record()
		sb.WriteString(n.name + " [shape=plaintext, width=" + jfmt.DoubleToString((r.width+plaintextPadX)/72) +
			", height=" + jfmt.DoubleToString((r.height+1+plaintextPadY)/72) + ", label=<" + r.tableLabel() + ">];\n")
	}
	for _, e := range s.edges {
		fmt.Fprintf(&sb, "%s -> %s [tailport=P%d, headport=%s, arrowsize=.75, arrowtail=none, arrowhead=normal];\n",
			e.tail.name, e.head.name, e.port, headPort)
	}
	sb.WriteString("}\n")
	return sb.String()
}

// record is the geometry of upstream's record node, in the top-to-bottom
// layout: its size, the center of each port relative to the center of the
// node, rounded as Smetana rounds port points, and the size across the rows
// of the field of the ports.
type record struct {
	width, height float64
	ports         []int
	portsHeight   int
	isArray       bool
}

// record computes the record of upstream's SmetanaForJson for the node, as
// Smetana's record_init sizes it: fields with the sizes of the _dim_ hack
// (the rows along, the columns less 8 across), and the size of the node at
// least the width and the height attributes rounded, without padding.
// Along the rows, the fields keep their sizes from the left; across them,
// the room left is shared between the key and the value fields as
// resize_reclbl shares it.
func (n *internalNode) record() record {
	if len(n.heights) == 0 {
		return record{width: emptyNodeWidth, height: emptyNodeHeight}
	}
	total := 0.0
	for _, h := range n.heights {
		total += h
	}
	r := record{width: max(total, float64(roundHalfUp(total))), isArray: n.isArray}
	var natural, portsHeight float64
	if n.isArray {
		natural = n.colA - 8
	} else {
		natural = n.colA - 8 + n.colB - 8
	}
	r.height = max(natural, float64(roundHalfUp(n.width)))
	if n.isArray {
		portsHeight = r.height
	} else {
		inc := (r.height - natural) / 2
		portsHeight = n.colB - 8 + float64(int(2*inc)-int(inc))
	}
	r.portsHeight = roundHalfUp(portsHeight)
	sum := 0.0
	for _, h := range n.heights {
		r.ports = append(r.ports, roundHalfUp(-r.width/2+sum+h/2))
		sum += h
	}
	return r
}

// tableLabel is the table that stands for the record: as high as the record,
// so that the edges start and end at its border and not at the padding of
// plaintext, and a little narrower, of an even width, so that each port cell,
// 2 points wide between empty cells, is centered on the rounded center of its
// field. For an object, a first row holds the key field. The table is the
// port of the edges that end at the node (design D3).
func (r record) tableLabel() string {
	const table = `<TABLE PORT="` + headPort + `" BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0">`
	height := int(r.height)
	if len(r.ports) == 0 {
		return fmt.Sprintf(table+`<TR><TD FIXEDSIZE="TRUE" WIDTH="%d" HEIGHT="%d"></TD></TR></TABLE>`,
			int(r.width), height)
	}
	width := 2 * int(r.width/2)
	var cells strings.Builder
	count, end := 0, 0
	cell := func(port string, w int) {
		if w <= 0 {
			return
		}
		fmt.Fprintf(&cells, `<TD%s FIXEDSIZE="TRUE" WIDTH="%d" HEIGHT="%d"></TD>`, port, w, r.portsHeight)
		count++
		end += w
	}
	for i, c := range r.ports {
		start := max(c+width/2-portCellWidth/2, end)
		cell("", start-end)
		cell(` PORT="P`+strconv.Itoa(i)+`"`, portCellWidth)
	}
	cell("", width-end)
	var sb strings.Builder
	sb.WriteString(table)
	if !r.isArray {
		fmt.Fprintf(&sb, `<TR><TD COLSPAN="%d" FIXEDSIZE="TRUE" WIDTH="%d" HEIGHT="%d"></TD></TR>`, count, end,
			height-r.portsHeight)
	}
	sb.WriteString("<TR>" + cells.String() + "</TR></TABLE>")
	return sb.String()
}

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
		// The node without its padding: the coordinates along the rows move
		// by the padding of the first node, as do those of the edges.
		n.x, n.y, n.lw, n.lh = pos.X-plaintextPadX/2, pos.Y, w*72-plaintextPadX, h*72-plaintextPadY
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
		for i := range curve.points {
			curve.points[i].X -= plaintextPadX / 2
		}
		if curve.ep != nil {
			curve.ep.X -= plaintextPadX / 2
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
