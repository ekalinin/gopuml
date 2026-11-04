// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/gogv"
)

func TestGogvLayout(t *testing.T) {
	svg, err := (&Gogv{}).Layout(context.Background(), testDotSource)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<svg", `class="node"`} {
		if !strings.Contains(string(svg), want) {
			t.Errorf("the SVG of %q has no %s", testDotSource, want)
		}
	}
}

func TestGogvFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		render func(context.Context, string, string) ([]byte, []gogv.Warning, error)
		want   string
	}{
		{"panic", func(context.Context, string, string) ([]byte, []gogv.Warning, error) { panic("boom") }, "panic: boom"},
		{"error", func(context.Context, string, string) ([]byte, []gogv.Warning, error) {
			return nil, nil, errors.New("syntax error in line 1")
		}, "syntax error in line 1"},
		{"not SVG", func(context.Context, string, string) ([]byte, []gogv.Warning, error) {
			return []byte("graph 1 1 1\n"), nil, nil
		}, "output is not SVG"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&Gogv{Render: tc.render}).Layout(context.Background(), testDotSource)
			var fe *FailureError
			if !errors.As(err, &fe) || !errors.Is(err, ErrFailed) || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %v, want a *FailureError with %q", err, tc.want)
			}
		})
	}
}

// waitForContext is a gogv that does not finish before its context is done.
func waitForContext(ctx context.Context, _, _ string) ([]byte, []gogv.Warning, error) {
	<-ctx.Done()
	return nil, nil, ctx.Err()
}

func TestGogvTimeout(t *testing.T) {
	// Spec graph-layout "Built-in layout timeout": with a timeout of 1
	// second, the layout is stopped and reports a timeout.
	start := time.Now()
	_, err := (&Gogv{Timeout: time.Second, Render: waitForContext}).Layout(context.Background(), testDotSource)
	elapsed := time.Since(start)
	var te *TimeoutError
	if !errors.As(err, &te) || !errors.Is(err, ErrTimeout) || te.Timeout != time.Second {
		t.Fatalf("error %v, want a *TimeoutError of 1s", err)
	}
	if elapsed < time.Second || elapsed > 5*time.Second {
		t.Errorf("the layout stopped after %s, want about 1s", elapsed)
	}
}

func TestGogvCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(10*time.Millisecond, cancel)
	_, err := (&Gogv{Render: waitForContext}).Layout(ctx, testDotSource)
	if !errors.Is(err, context.Canceled) || errors.Is(err, ErrTimeout) {
		t.Errorf("error %v, want context.Canceled and no timeout", err)
	}
}

func TestGogvVersion(t *testing.T) {
	line := gogvVersionLine()
	// The test binary names the gogv of go.mod, without its leading "v".
	if !regexp.MustCompile(`^dot - graphviz version 16\.1\.0 \(gogv [^v ][^ ]*\)$`).MatchString(line) {
		t.Errorf("version line %q", line)
	}
	if v := (&Gogv{}).GraphvizVersion(context.Background()); v != numberedGraphvizVersion(1601) {
		t.Errorf("Graphviz version %v, want 1601", v)
	}
}
