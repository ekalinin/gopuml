// SPDX-License-Identifier: MIT
// Provenance: original

package svek_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ekalinin/gopuml/internal/classdiagram"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/svek"
)

var sourceLines = regexp.MustCompile(` (data-source-line|codeLine)="\d+"`)

const twoClasses = "@startuml\nclass A\nclass B\nA --> B\n@enduml\n"

func parseSource(t *testing.T, src string) *classdiagram.ClassDiagram {
	t.Helper()
	blocks, err := engine.Parse(context.Background(), core.NewRenderContext(), engine.Source{Text: src})
	if err != nil {
		t.Fatal(err)
	}
	d, ok := blocks[0].Diagram.(*classdiagram.ClassDiagram)
	if !ok {
		t.Fatalf("not a class diagram: %T", blocks[0].Diagram)
	}
	return d
}

func export1(ctx context.Context, d *classdiagram.ClassDiagram, layout svek.Layout) (string, export.ImageData, error) {
	var buf bytes.Buffer
	data, err := d.ExportDiagramLayout(ctx, &buf, 0, export.NewFileFormatOptionMetadata(font.FormatSVG, false), layout)
	return buf.String(), data, err
}

// realDot returns the layout with the dot executable of the host, and skips
// the test when there is none.
func realDot(t *testing.T) svek.Layout {
	t.Helper()
	exec, err := dot.Config{}.NewExec()
	if err != nil {
		t.Skip("no Graphviz dot executable: ", err)
	}
	return svek.Layout{Engine: exec}
}

// countingEngine counts the layouts of the engine it wraps.
type countingEngine struct {
	inner dot.Engine
	calls int
}

func (c *countingEngine) Layout(ctx context.Context, dotString string) ([]byte, error) {
	c.calls++
	return c.inner.Layout(ctx, dotString)
}

func TestClassDiagramWithoutDotReportsError(t *testing.T) {
	// Spec graph-layout "Class diagram without Graphviz": an error stating
	// that dot was not found, with an error image.
	d := parseSource(t, twoClasses)
	layout := svek.Layout{Config: dot.Config{DotPath: "/nonexistent/graphviz/dot"}}
	svg, data, err := export1(context.Background(), d, layout)
	if !errors.Is(err, dot.ErrNotFound) {
		t.Fatalf("error %v, want dot.ErrNotFound", err)
	}
	if data == nil || svg == "" {
		t.Fatal("no error image")
	}
	for _, want := range []string{"Dot Executable: /nonexistent/graphviz/dot", "Dot executable does not exist",
		"Cannot find Graphviz. You should try"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the image does not say %q", want)
		}
	}
}

func TestSmetanaPragmaFallsBackToDot(t *testing.T) {
	// Spec graph-layout "Smetana pragma before the port" and design D16:
	// smetana, elk and vizjs are laid out with dot.
	layout := realDot(t)
	_, _, err := export1(context.Background(), parseSource(t, twoClasses), layout)
	if err != nil {
		t.Fatal(err)
	}
	want, _, _ := export1(context.Background(), parseSource(t, twoClasses), layout)
	for _, engineName := range []string{"smetana", "elk", "vizjs"} {
		src := strings.Replace(twoClasses, "@startuml\n", "@startuml\n!pragma layout "+engineName+"\n", 1)
		counting := &countingEngine{inner: layout.Engine}
		got, _, err := export1(context.Background(), parseSource(t, src), svek.Layout{Engine: counting})
		if err != nil {
			t.Fatalf("%s: %v", engineName, err)
		}
		if counting.calls != 1 {
			t.Errorf("%s: dot ran %d times", engineName, counting.calls)
		}
		// The pragma line moves the source lines of the elements.
		if sourceLines.ReplaceAllString(got, "") != sourceLines.ReplaceAllString(want, "") {
			t.Errorf("%s: the drawing differs from the dot layout", engineName)
		}
	}
}

// fakeEngine returns a fixed result.
type fakeEngine struct {
	svg []byte
	err error
}

func (f fakeEngine) Layout(context.Context, string) ([]byte, error) { return f.svg, f.err }

func TestLayoutFailureImage(t *testing.T) {
	// Spec graph-layout "dot failure" and "dot timeout": the error comes
	// with an image that shows dot's message (design D16).
	for _, tc := range []struct {
		err     error
		is      error
		message string
	}{
		{&dot.FailureError{Path: "/usr/bin/dot", ExitCode: 1, Stderr: "Error: syntax error in line 1"}, dot.ErrFailed,
			"syntax error in line 1"},
		{&dot.TimeoutError{Path: "/usr/bin/dot", Timeout: time.Second}, dot.ErrTimeout, "timed out after 1s"},
	} {
		svg, data, err := export1(context.Background(), parseSource(t, twoClasses),
			svek.Layout{Engine: fakeEngine{err: tc.err}})
		if !errors.Is(err, tc.is) {
			t.Errorf("error %v, want %v", err, tc.is)
		}
		if data == nil || !strings.Contains(svg, tc.message) {
			t.Errorf("the image does not show %q", tc.message)
		}
	}
	// Output that is not dot's SVG is reported as a failure too.
	_, data, err := export1(context.Background(), parseSource(t, twoClasses), svek.Layout{Engine: fakeEngine{}})
	if !errors.Is(err, dot.ErrFailed) || data == nil {
		t.Errorf("empty output: error %v, image %v", err, data != nil)
	}
}

// blockingEngine waits for the context to be done.
type blockingEngine struct{ started chan struct{} }

func (b blockingEngine) Layout(ctx context.Context, _ string) ([]byte, error) {
	close(b.started)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestCancelledLayoutWritesNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	eng := blockingEngine{started: make(chan struct{})}
	go func() {
		<-eng.started
		cancel()
	}()
	var buf bytes.Buffer
	d := parseSource(t, twoClasses)
	data, err := d.ExportDiagramLayout(ctx, &buf, 0, export.NewFileFormatOption(font.FormatSVG),
		svek.Layout{Engine: eng})
	if !errors.Is(err, context.Canceled) || data != nil || buf.Len() != 0 {
		t.Errorf("error %v, data %v, %d bytes written", err, data, buf.Len())
	}
}

func TestDegeneratedDiagramDoesNotNeedDot(t *testing.T) {
	// Upstream draws a diagram of a single entity without Graphviz.
	d := parseSource(t, "@startuml\nclass A\n@enduml\n")
	counting := &countingEngine{inner: fakeEngine{err: errors.New("must not run")}}
	svg, _, err := export1(context.Background(), d, svek.Layout{Engine: counting})
	if err != nil || counting.calls != 0 || !strings.Contains(svg, ">A</text>") {
		t.Errorf("error %v, %d layouts", err, counting.calls)
	}
}

func TestDotStringSkek(t *testing.T) {
	d := parseSource(t, "@startuml\n!pragma aspect 0,5\n!pragma ratio fill\nclass A\n@enduml\n")
	got := svek.DotStringSkek(d.Cuca(), d.DotStrings())
	want := []string{"nodesep=.35;", "ranksep=0.8;", "aspect=0.5;", "ratio=fill;"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConcurrentLayouts(t *testing.T) {
	// Each render has its own layout state (design D3): concurrent renders
	// of the same source give the same drawing.
	dotSvg, err := os.ReadFile("testdata/golden/bigdiagram.dot14.svg")
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile("testdata/cases/bigdiagram.puml")
	if err != nil {
		t.Fatal(err)
	}
	want, _, err := export1(context.Background(), parseSource(t, string(src)), svek.Layout{Engine: fakeEngine{svg: dotSvg}})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d := parseSource(t, string(src))
			got, _, err := export1(context.Background(), d, svek.Layout{Engine: fakeEngine{svg: dotSvg}})
			if err != nil || got != want {
				errs <- fmt.Sprint("concurrent render differs: ", err)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
