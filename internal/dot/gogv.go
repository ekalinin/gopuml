// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ekalinin/gogv"
)

// gogvModule is the module of gogv, whose version the build information
// gives.
const gogvModule = "github.com/ekalinin/gogv"

// gogvPath names the built-in engine in its errors, where the dotexec
// engine names the dot executable.
const gogvPath = "gogv"

// Gogv is the built-in layout engine: it lays graphs out in the process with
// gogv, a Go implementation of Graphviz dot, so that no Graphviz has to be
// installed. It holds no mutable state and is safe for concurrent use.
type Gogv struct {
	// Timeout limits each layout. Zero or negative means DefaultTimeout.
	Timeout time.Duration
	// Render lays out the graphs of src in format. Nil means gogv.Render;
	// tests replace it with fakes.
	Render func(ctx context.Context, src, format string) ([]byte, []gogv.Warning, error)
}

var (
	_ Engine   = (*Gogv)(nil)
	_ Renderer = (*Gogv)(nil)
)

// Layout lays the DOT source out with gogv and returns the SVG document that
// `dot -Tsvg` returns for it. It returns a *TimeoutError when the layout runs
// longer than the timeout, the context's error when ctx is done, and a
// *FailureError when gogv fails, panics or returns output that is not SVG.
// gogv's warnings are dropped, as the dotexec engine drops what dot writes
// on stderr when it succeeds.
func (g *Gogv) Layout(ctx context.Context, dot string) (svg []byte, err error) {
	timeout := g.timeout()
	lctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer func() {
		if v := recover(); v != nil {
			svg, err = nil, &FailureError{Path: gogvPath, ExitCode: -1, Err: fmt.Errorf("panic: %v", v)}
		}
	}()
	render := g.Render
	if render == nil {
		render = gogv.Render
	}
	out, _, err := render(lctx, dot, "svg")
	switch {
	case err == nil && bytes.Contains(out, []byte("<svg")):
		return out, nil
	case err == nil:
		return nil, &FailureError{Path: gogvPath, Reason: "output is not SVG"}
	case ctx.Err() != nil:
		return nil, fmt.Errorf("dot: %w", ctx.Err())
	case errors.Is(lctx.Err(), context.DeadlineExceeded):
		return nil, &TimeoutError{Path: gogvPath, Timeout: timeout}
	}
	return nil, &FailureError{Path: gogvPath, ExitCode: -1, Err: err}
}

// RenderFormat lays the DOT source out with gogv and returns what gogv's dot
// writes for -T<format> on stdout and stderr, merged as upstream merges them
// for dot: a "Warning: " line for each warning, the output, then
// "Error: <stdin>: " and the message for a syntax error, which is not an
// error of RenderFormat. The warnings about Graphviz features that gogv does
// not support yet are dropped, as Graphviz does not give them. It returns a
// *TimeoutError when the layout runs longer than the timeout, the context's
// error when ctx is done, and a *FailureError when gogv fails, for example
// for a format it does not write, or panics.
func (g *Gogv) RenderFormat(ctx context.Context, src, format string) (out []byte, err error) {
	timeout := g.timeout()
	lctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer func() {
		if v := recover(); v != nil {
			out, err = nil, &FailureError{Path: gogvPath, ExitCode: -1, Err: fmt.Errorf("panic: %v", v)}
		}
	}()
	render := g.Render
	if render == nil {
		render = gogv.Render
	}
	data, warnings, err := render(lctx, src, format)
	syntax, isSyntax := errors.AsType[*gogv.SyntaxError](err)
	switch {
	case err == nil || isSyntax:
		var b bytes.Buffer
		for _, w := range warnings {
			if !w.NotSupportedYet() {
				b.WriteString("Warning: " + w.String() + "\n")
			}
		}
		b.Write(data)
		if isSyntax {
			b.WriteString("Error: <stdin>: " + syntax.Error() + "\n")
		}
		return b.Bytes(), nil
	case ctx.Err() != nil:
		return nil, fmt.Errorf("dot: %w", ctx.Err())
	case errors.Is(lctx.Err(), context.DeadlineExceeded):
		return nil, &TimeoutError{Path: gogvPath, Timeout: timeout}
	}
	return nil, &FailureError{Path: gogvPath, ExitCode: -1, Err: err}
}

// Output lays the DOT source out with gogv and returns its output for
// -T<format>, without the warnings, for a caller that reads the layout
// itself, such as the JSON diagrams with the json0 format. It returns a
// *TimeoutError when the layout runs longer than the timeout, the context's
// error when ctx is done, and a *FailureError when gogv fails or panics.
func (g *Gogv) Output(ctx context.Context, src, format string) (out []byte, err error) {
	timeout := g.timeout()
	lctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer func() {
		if v := recover(); v != nil {
			out, err = nil, &FailureError{Path: gogvPath, ExitCode: -1, Err: fmt.Errorf("panic: %v", v)}
		}
	}()
	render := g.Render
	if render == nil {
		render = gogv.Render
	}
	data, _, err := render(lctx, src, format)
	switch {
	case err == nil:
		return data, nil
	case ctx.Err() != nil:
		return nil, fmt.Errorf("dot: %w", ctx.Err())
	case errors.Is(lctx.Err(), context.DeadlineExceeded):
		return nil, &TimeoutError{Path: gogvPath, Timeout: timeout}
	}
	return nil, &FailureError{Path: gogvPath, ExitCode: -1, Err: err}
}

// GraphvizVersion returns the layout tweaks for the Graphviz version that
// gogv follows.
func (g *Gogv) GraphvizVersion(context.Context) GraphvizVersion {
	return NewGraphvizVersion(gogvVersionLine())
}

func (g *Gogv) timeout() time.Duration {
	if g.Timeout <= 0 {
		return DefaultTimeout
	}
	return g.Timeout
}

// gogvVersionLine is the version line of the built-in engine, as gogv's
// `dot -V` writes it: "dot - graphviz version 16.1.0 (gogv 0.2.0)".
func gogvVersionLine() string {
	return "dot - graphviz version " + gogv.GraphvizVersion + " (gogv " + gogvVersion() + ")"
}

// gogvVersion is the version of the gogv module in the build information,
// without its leading "v", or "devel" when it is not there, as gogv's
// `dot -V` writes it. It is read once, as every layout asks for it.
var gogvVersion = sync.OnceValue(func() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	for _, dep := range info.Deps {
		if dep.Path != gogvModule {
			continue
		}
		if dep.Replace != nil {
			dep = dep.Replace
		}
		if dep.Version != "" && dep.Version != "(devel)" {
			return strings.TrimPrefix(dep.Version, "v")
		}
	}
	return "devel"
})
