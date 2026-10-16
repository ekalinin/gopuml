// SPDX-License-Identifier: MIT
// Provenance: original

package classdiagram

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/decoration"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

func parseClass(t *testing.T, source string) *ClassDiagram {
	t.Helper()
	d, _, err := parseCase(source + "\n")
	if err != nil {
		t.Fatal(err)
	}
	cd, ok := d.(*ClassDiagram)
	if !ok {
		if e, ok := d.(*core.ErrorDiagram); ok {
			t.Fatalf("error at line %d: %s", e.Position(), e.Error.Message())
		}
		t.Fatalf("result %T", d)
	}
	return cd
}

// TestLinkMiddleCircle checks the RE2 rewrite of upstream's middle circle
// lookahead: a circle followed by the second body is a middle decoration,
// a circle at the end is a head.
func TestLinkMiddleCircle(t *testing.T) {
	for _, c := range []struct {
		line   string
		middle decoration.LinkMiddleDecor
		decor1 decoration.LinkDecor
	}{
		{"A -0- B", decoration.LinkMiddleDecorCircle, decoration.LinkDecorNone},
		{"A -(0)-> B", decoration.LinkMiddleDecorCircleCircled, decoration.LinkDecorArrow},
		{"A --0 B", decoration.LinkMiddleDecorNone, decoration.LinkDecorCircle},
		{"A --(0 B", decoration.LinkMiddleDecorNone, decoration.LinkDecorCircleConnect},
	} {
		cd := parseClass(t, c.line)
		lt := cd.Links()[0].Type()
		if lt.MiddleDecor() != c.middle || lt.Decor1() != c.decor1 {
			t.Errorf("%q: middle %v decor1 %v", c.line, lt.MiddleDecor(), lt.Decor1())
		}
	}
}

func TestDirectionInvertsLink(t *testing.T) {
	cd := parseClass(t, "A -left-> B\nC -up-|> D\nE -right-> F")
	links := cd.Links()
	if links[0].Entity1().Name() != "B" || !links[0].IsInverted() || links[0].Length() != 1 {
		t.Errorf("left: %s inverted %v length %d", links[0].Entity1().Name(), links[0].IsInverted(), links[0].Length())
	}
	if links[1].Entity1().Name() != "D" || links[1].Type().Decor2() != decoration.LinkDecorExtends {
		t.Errorf("up: %s %v", links[1].Entity1().Name(), links[1].Type())
	}
	if links[2].IsInverted() || links[2].Length() != 1 {
		t.Error("right")
	}
}

func TestClassDeclaration(t *testing.T) {
	cd := parseClass(t, "class Repo<T> <<service>> #red\nabstract class A extends B implements I, J")
	repo := cd.QuarkInContext(true, "Repo").Data()
	if repo.LeafType() != abel.LeafTypeClass || repo.Generic().S != "T" || repo.Stereotype().String() != "<<service>>" {
		t.Errorf("Repo: %v %v %v", repo.LeafType(), repo.Generic(), repo.Stereotype())
	}
	i := cd.QuarkInContext(true, "I").Data()
	if i.LeafType() != abel.LeafTypeInterface {
		t.Errorf("implemented type %v", i.LeafType())
	}
	var kinds []string
	for _, l := range cd.Links() {
		kinds = append(kinds, l.Entity1().Name()+">"+l.Entity2().Name()+":"+l.Type().String())
	}
	if got := strings.Join(kinds, " "); got != "B>A:NONE-NORMAL(null)-EXTENDS I>A:NONE-DASHED(null)-EXTENDS "+
		"J>A:NONE-DASHED(null)-EXTENDS" {
		t.Errorf("links %s", got)
	}
}

func TestNewpage(t *testing.T) {
	d, _, err := parseCase("class A\nnewpage\nclass B\nnewpage\nclass C\n")
	if err != nil {
		t.Fatal(err)
	}
	n, ok := d.(*abel.NewpagedDiagram)
	if !ok || len(n.Diagrams()) != 3 || n.Description() != "(1 entities) (1 entities) (1 entities)" {
		t.Fatalf("newpage: %T %s", d, d.Description())
	}
	// Upstream draws the first page only (a single image).
	var buf bytes.Buffer
	if _, err := n.ExportDiagram(&buf, 0, export.NewFileFormatOption(font.FormatSVG)); err != nil {
		t.Fatal(err)
	}
	if svg := buf.String(); !strings.Contains(svg, ">A</text>") || strings.Contains(svg, ">B</text>") {
		t.Errorf("export: %s", svg)
	}
}

func TestMixingNeedsAllowmixing(t *testing.T) {
	d, _, _ := parseCase("component C\n")
	e, ok := d.(*core.ErrorDiagram)
	if !ok || e.Error.Error != "Use 'allowmixing' if you want to mix classes and other UML elements." {
		t.Fatalf("without allowmixing: %T", d)
	}
	cd := parseClass(t, "allowmixing\ncomponent C\nnode N")
	if got := cd.QuarkInContext(true, "N").Data().USymbol(); got != abel.USymbolNode {
		t.Errorf("node symbol %v", got)
	}
}

func TestDiagramType(t *testing.T) {
	cd := parseClass(t, "object O")
	if cd.DiagramType() != core.Class || cd.Description() != "(1 entities)" {
		t.Errorf("type %v description %q", cd.DiagramType(), cd.Description())
	}
	cd = parseClass(t, "@startuml id=foo\nclass A\n@enduml")
	if cd.Description() != "(1 entities, foo)" {
		t.Errorf("description with id %q", cd.Description())
	}
}

// TestConcurrentParse parses in parallel: the factory and the commands keep
// no mutable state (design D3).
func TestConcurrentParse(t *testing.T) {
	src := "class A<T> <<s>> {\n -f\n +m()\n}\nA \"1\" *-- \"*\" B : c >\npackage p {\n  class C\n}\n" +
		"note right of A : n\nhide empty members\n(A, B) .. D\n"
	want := dumpCase(testCase{name: "x", source: src})
	done := make(chan []string)
	for range 8 {
		go func() { done <- dumpCase(testCase{name: "x", source: src}) }()
	}
	for range 8 {
		if got := <-done; strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Error("parallel parse differs")
		}
	}
}
