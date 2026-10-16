// SPDX-License-Identifier: MIT
// Provenance: original

package svg_test

import (
	"bufio"
	"compress/gzip"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/svg"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/svg/svgtest"
)

// TestGoldenScripts runs testdata/cases.txt and cases-insecure.txt and
// compares with the SVG text upstream writes for the same drawing
// operations (golden/<case>.svg, from testdata/javagen/SvgGolden.java). PNG
// data URLs and the plantuml-src metadata are compared decoded (see
// svgtest.Normalize); everything else byte for byte.
func TestGoldenScripts(t *testing.T) {
	cases, err := svgtest.ParseFile("testdata/cases.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) < 50 {
		t.Fatalf("only %d cases", len(cases))
	}
	// Cases upstream runs with the INSECURE profile and javascript links.
	insecure, err := svgtest.ParseFile("testdata/cases-insecure.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases = append(cases, insecure...)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", "golden", c.Name+".svg"))
			if err != nil {
				t.Fatal(err)
			}
			got, err := svgtest.RunSvg(c)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) == string(want) {
				return
			}
			g, w := svgtest.Normalize(string(got)), svgtest.Normalize(string(want))
			if g != w {
				t.Errorf("SVG differs at byte %d\n got: %s\nwant: %s", firstDiff(g, w), around(g, firstDiff(g, w)),
					around(w, firstDiff(g, w)))
			}
		})
	}
}

func firstDiff(a, b string) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func around(s string, i int) string {
	lo := max(0, i-150)
	hi := min(len(s), i+150)
	return s[lo:hi]
}

// TestFormatTable compares the number formatting with upstream's
// SvgGraphics.format, formatPercent, formatOpacity and formatBoolean on the
// values of golden/format.tsv.gz: rounding boundaries of each number of
// decimals and their neighbors, negative zero, NaN, infinities, subnormals,
// large values, random values and several scales.
func TestFormatTable(t *testing.T) {
	f, err := os.Open("testdata/golden/format.tsv.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(zr)
	n, failures := 0, 0
	for sc.Scan() {
		p := strings.Split(sc.Text(), "\t")
		if len(p) != 7 {
			t.Fatalf("bad line %q", sc.Text())
		}
		decimal, _ := strconv.Atoi(p[0])
		scaleBits, _ := strconv.ParseUint(p[1], 16, 64)
		valueBits, _ := strconv.ParseUint(p[2], 16, 64)
		scale := math.Float64frombits(scaleBits)
		v := math.Float64frombits(valueBits)
		g := svg.NewSvgGraphics(42, svg.SvgOptionBasic().WithDecimal(decimal).WithScale(scale))
		got := []string{g.Format(v), g.FormatPercent(v), g.FormatOpacity(v), g.FormatBoolean(v)}
		for i, want := range p[3:] {
			if got[i] != want {
				failures++
				if failures < 30 {
					t.Errorf("decimal %d scale %v value %v (%s): function %d gives %q, upstream %q", decimal, scale,
						v, p[2], i, got[i], want)
				}
			}
		}
		n++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if n < 20000 {
		t.Errorf("only %d values", n)
	}
	if failures > 0 {
		t.Errorf("%d mismatches over %d values", failures, n)
	}
}

// TestXMLWriter builds the document of SvgGolden.xmlWriter and compares its
// serialization, without and with indentation, with golden/xml.txt. Empty
// strings stand for upstream's nulls, which serialize the same.
func TestXMLWriter(t *testing.T) {
	want, err := os.ReadFile("testdata/golden/xml.txt")
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for _, indent := range []int{0, 2, 3} {
		doc := &svg.XMLDocument{}
		root := doc.CreateElement("root")
		doc.SetRoot(root)
		root.SetAttribute("a", "x&y<z>\"q'")
		root.SetAttribute("b", "")
		root.SetAttribute("a", "again")
		root.AppendProcessingInstruction("pi", "data")
		root.AppendProcessingInstruction("empty", "")
		root.AppendProcessingInstruction("nul", "")
		root.AppendChild(doc.CreateElement("empty"))
		text := doc.CreateElement("text")
		text.SetTextContent("a & b < c > d \" e")
		root.AppendChild(text)
		nulText := doc.CreateElement("nul")
		nulText.SetTextContent("")
		root.AppendChild(nulText)
		mixed := doc.CreateElement("mixed")
		mixed.AppendText("before")
		child := doc.CreateElement("child")
		child.AppendComment("c--d-")
		child.AppendComment("")
		child.AppendCData("x]]>y")
		child.AppendCData("")
		mixed.AppendChild(child)
		mixed.AppendText("after")
		mixed.AppendRaw(`<raw attr="1"/>`)
		mixed.AppendRaw("")
		root.AppendChild(mixed)
		deep := doc.CreateElement("l1")
		l2 := doc.CreateElement("l2")
		l3 := doc.CreateElement("l3")
		l3.SetAttribute("k", "v")
		l2.AppendChild(l3)
		l2.AppendText("t")
		deep.AppendChild(l2)
		root.AppendChild(deep)
		sb.WriteString("indent " + strconv.Itoa(indent) + "\n")
		sb.WriteString(doc.ToXML(indent))
		sb.WriteString("\n-----\n")
	}
	if sb.String() != string(want) {
		t.Errorf("got\n%s\nwant\n%s", sb.String(), want)
	}
}

var (
	numericAttribute = regexp.MustCompile(
		` (x|y|x1|y1|x2|y2|cx|cy|rx|ry|width|height|points|d|font-size|textLength|stdDeviation|dx|dy)="([^"]*)"`)
	numberToken = regexp.MustCompile(`-?[0-9]+(\.[0-9]+)?`)
)

// TestReferenceNumbers checks that every coordinate of the reference SVGs
// (testdata/conformance/reference) is in the form the formatter writes:
// formatting the value it denotes gives the same text (no trailing zeros,
// at most 3 decimals, no exponent).
func TestReferenceNumbers(t *testing.T) {
	files, err := filepath.Glob("../../../testdata/conformance/reference/*/*/block*-page*.svg")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 50 {
		t.Skipf("only %d reference SVGs", len(files))
	}
	g := svg.NewSvgGraphics(42, svg.SvgOptionBasic())
	n := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range numericAttribute.FindAllStringSubmatch(string(data), -1) {
			for _, tok := range numberToken.FindAllString(m[2], -1) {
				v, err := strconv.ParseFloat(tok, 64)
				if err != nil {
					t.Fatal(err)
				}
				if got := g.Format(v); got != tok {
					t.Errorf("%s: %s=%q: %q formats as %q", filepath.Base(filepath.Dir(f)), m[1], m[2], tok, got)
				}
				n++
			}
		}
	}
	if n < 5000 {
		t.Errorf("only %d numbers", n)
	}
}
