// SPDX-License-Identifier: MIT
// Provenance: original

package classdiagram

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/svek"
)

// recordingEngine records the DOT graphs given to the engine it wraps.
type recordingEngine struct {
	inner dot.Engine
	dots  []string
}

func (r *recordingEngine) Layout(ctx context.Context, dotString string) ([]byte, error) {
	r.dots = append(r.dots, dotString)
	return r.inner.Layout(ctx, dotString)
}

var (
	entityRect = regexp.MustCompile(`data-qualified-name="(\w+)"[^>]*>\s*<rect x="[-\d.]+" y="([-\d.]+)" width="[-\d.]+" height="([-\d.]+)"`)
	linkPath   = regexp.MustCompile(`<path d="M([-\d.]+),([-\d.]+) C[^"]* ([-\d.]+),([-\d.]+)"`)
)

func TestClassLayoutUsesDot(t *testing.T) {
	// Spec class-diagram "Rendering requires dot" and graph-layout "Class
	// diagram layout": B below A, the arrow from A to B, laid out by dot.
	exec, err := dot.Config{}.NewExec()
	if err != nil {
		t.Skip("no Graphviz dot executable: ", err)
	}
	d := parseClass(t, "@startuml\nclass A\nclass B\nA --> B\n@enduml")
	eng := &recordingEngine{inner: exec}
	var buf bytes.Buffer
	if _, err := d.ExportDiagramLayout(context.Background(), &buf, 0, export.NewFileFormatOption(font.FormatSVG),
		svek.Layout{Engine: eng}); err != nil {
		t.Fatal(err)
	}
	if len(eng.dots) != 1 || !strings.Contains(eng.dots[0], "sh0006->sh0007") {
		t.Fatalf("dot was not given the graph: %q", eng.dots)
	}
	svg := buf.String()
	pos := map[string][2]float64{}
	for _, m := range entityRect.FindAllStringSubmatch(svg, -1) {
		y, _ := strconv.ParseFloat(m[2], 64)
		h, _ := strconv.ParseFloat(m[3], 64)
		pos[m[1]] = [2]float64{y, h}
	}
	a, okA := pos["A"]
	b, okB := pos["B"]
	if !okA || !okB || b[0] < a[0]+a[1] {
		t.Fatalf("B is not below A: %v", pos)
	}
	m := linkPath.FindStringSubmatch(svg)
	if m == nil {
		t.Fatal("no link path")
	}
	y1, _ := strconv.ParseFloat(m[2], 64)
	y2, _ := strconv.ParseFloat(m[4], 64)
	if y1 > a[0]+a[1]+1 || y2 < b[0]-10 {
		t.Errorf("the arrow does not run from A to B: %v", m[1:])
	}
}
