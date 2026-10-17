// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const classSource = "@startuml\nclass A\nclass B\nA --> B\n@enduml\n"

func TestRenderMissingGraphviz(t *testing.T) {
	// Spec library-api "Missing Graphviz": the error matches
	// ErrGraphvizNotFound, and the error image comes with it.
	res, err := Render(context.Background(), classSource, &Options{DotPath: "/nonexistent/graphviz/dot"})
	if !errors.Is(err, ErrGraphvizNotFound) {
		t.Fatalf("error %v, want ErrGraphvizNotFound", err)
	}
	if res == nil || !strings.Contains(string(res.Data), "Cannot find Graphviz") || res.Width == 0 {
		t.Errorf("no error image: %+v", res)
	}
}

func TestSequenceDiagramRendersWithoutDot(t *testing.T) {
	// Spec graph-layout "Sequence diagram without Graphviz".
	res, err := Render(context.Background(), seqSource, &Options{DotPath: "/nonexistent/graphviz/dot"})
	if err != nil || res == nil || !strings.Contains(string(res.Data), "<svg") {
		t.Fatalf("error %v, result %v", err, res)
	}
}

// fakeDot writes an executable shell script named dot, which answers -V like
// Graphviz 14.0.1 and runs body for a layout.
func fakeDot(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake dot is a shell script")
	}
	path := filepath.Join(t.TempDir(), "dot")
	script := "#!/bin/sh\nif [ \"$1\" = \"-V\" ]; then echo 'dot - graphviz version 14.0.1 (20251006.0113)' >&2; exit 0; fi\n" +
		body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRenderCancelledContext(t *testing.T) {
	// Spec library-api "Cancelled context": dot is stopped and the render
	// returns the context error.
	marker := filepath.Join(t.TempDir(), "started")
	dotPath := fakeDot(t, "cat > /dev/null\ntouch "+marker+"\nexec sleep 30")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for range 1000 {
			if _, err := os.Stat(marker); err == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
	}()
	start := time.Now()
	res, err := Render(ctx, classSource, &Options{DotPath: dotPath})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error %v, want context.Canceled", err)
	}
	if res != nil {
		t.Errorf("a result came with the cancellation: %+v", res)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("dot was not stopped: the render took %s", elapsed)
	}
}

func TestDotFailureOnFirstClassDiagramKeepsSecond(t *testing.T) {
	// Spec graph-layout "dot crashes on one diagram": the first diagram
	// reports dot's error, the second still renders.
	realDot, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("no Graphviz dot on PATH")
	}
	// The first diagram has three classes, so its graph has the node sh0008.
	dotPath := fakeDot(t, "input=$(cat)\ncase \"$input\" in *sh0008*) echo 'Error: fake dot crash' >&2; exit 1;; esac\n"+
		"printf '%s' \"$input\" | exec "+realDot+" \"$@\"")
	src := "@startuml\nclass A\nclass B\nclass C\nA --> B\n@enduml\n\n" + classSource
	res, err := Render(context.Background(), src, &Options{DotPath: dotPath})
	if !errors.Is(err, ErrGraphvizFailed) || !strings.Contains(err.Error(), "fake dot crash") {
		t.Fatalf("first diagram: error %v, want ErrGraphvizFailed with dot's message", err)
	}
	if res == nil || !strings.Contains(string(res.Data), "fake dot crash") {
		t.Errorf("first diagram: the image does not show dot's message")
	}
	res, err = Render(context.Background(), src, &Options{DotPath: dotPath, Diagram: 1})
	if err != nil || res == nil || !strings.Contains(string(res.Data), ">A</text>") {
		t.Fatalf("second diagram: error %v", err)
	}
}

func TestRenderGraphvizTimeout(t *testing.T) {
	// Spec graph-layout "Timeout exceeded", through the library option.
	dotPath := fakeDot(t, "cat > /dev/null\nexec sleep 30")
	start := time.Now()
	res, err := Render(context.Background(), classSource, &Options{DotPath: dotPath, DotTimeout: 500 * time.Millisecond})
	if !errors.Is(err, ErrGraphvizTimeout) {
		t.Fatalf("error %v, want ErrGraphvizTimeout", err)
	}
	if res == nil || len(res.Data) == 0 {
		t.Error("no error image")
	}
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("dot was not stopped: the render took %s", elapsed)
	}
}
