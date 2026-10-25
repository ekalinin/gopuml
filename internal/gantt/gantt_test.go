// SPDX-License-Identifier: MIT
// Provenance: original

package gantt_test

import (
	"bytes"
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/gantt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

func render(t *testing.T, rc *core.RenderContext, src string) (string, error) {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	if _, ok := blocks[0].Diagram.(*gantt.GanttDiagram); !ok {
		t.Fatalf("diagram %T", blocks[0].Diagram)
	}
	var buf bytes.Buffer
	_, err = blocks[0].Render(context.Background(), &buf, 0, engine.Options{Format: font.FormatSVG})
	return buf.String(), err
}

// TestGanttWithoutDot checks the gantt-diagram scenario "Gantt diagram
// without Graphviz".
func TestGanttWithoutDot(t *testing.T) {
	rc := core.NewRenderContext()
	rc.DotPath = "/nonexistent/graphviz/dot"
	svg, err := render(t, rc, "@startgantt\n[Prototype design] lasts 13 days\n@endgantt\n")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains([]byte(svg), []byte(">Prototype design</text>")) {
		t.Errorf("the drawing has no task: %.400s", svg)
	}
}

// TestGanttCurrentDay checks the gantt-diagram scenario "Current day": on
// January 8, 2020, "today is colored in #AAF" draws the drawing of "today is
// 2020-01-08 and is colored in #AAF".
func TestGanttCurrentDay(t *testing.T) {
	gantt.SetNow(t, time.Date(2020, 1, 8, 15, 30, 0, 0, time.Local))
	got, err := render(t, core.NewRenderContext(),
		"@startgantt\nProject starts 2020-01-06\ntoday is colored in #AAF\n[A] lasts 5 days\n@endgantt\n")
	if err != nil {
		t.Fatal(err)
	}
	want, err := render(t, core.NewRenderContext(),
		"@startgantt\nProject starts 2020-01-06\ntoday is 2020-01-08 and is colored in #AAF\n[A] lasts 5 days\n@endgantt\n")
	if err != nil {
		t.Fatal(err)
	}
	source := regexp.MustCompile(`<\?plantuml-src [^?]*\?>`)
	if source.ReplaceAllString(got, "") != source.ReplaceAllString(want, "") {
		t.Errorf("got %s\nwant %s", got, want)
	}
	if !bytes.Contains([]byte(got), []byte(`<rect x="161.707" y="39" width="16" height="16.805" fill="#AAF"`)) {
		t.Errorf("January 8 is not colored: %s", got)
	}
}

// TestGanttDrawingFailure checks the gantt-diagram scenario "Drawing
// failure": upstream fails to draw a resource of a task in a group and
// writes a crash report, so the render returns an error and no drawing.
func TestGanttDrawingFailure(t *testing.T) {
	svg, err := render(t, core.NewRenderContext(), "@startgantt\nProject starts 2020-01-06\ngroup [G]\n"+
		"[A] on {Alice} lasts 2 days\nend group\n@endgantt\n")
	if err == nil {
		t.Fatal("the diagram rendered without an error")
	}
	if svg != "" {
		t.Errorf("%d bytes written", len(svg))
	}
}
