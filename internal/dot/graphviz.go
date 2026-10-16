// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/Graphviz.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/AbstractGraphviz.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/dot/GraphvizVersionFinder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/DotStringFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/GraphvizImageBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

// Engine lays out a graph described in the DOT language. Implementations must
// be safe for concurrent use.
type Engine interface {
	// Layout lays out the graph and returns the SVG document that
	// `dot -Tsvg` produces for it.
	Layout(ctx context.Context, dot string) ([]byte, error)
}

// Exec is the dotexec engine: it lays graphs out by running the Graphviz dot
// executable, as upstream's GraphvizLinux and GraphvizWindows* do. It holds no
// mutable state and is safe for concurrent use.
type Exec struct {
	// Path is the dot executable, usually found with Config.Find.
	Path string
	// Timeout limits each run of dot. Zero or negative means DefaultTimeout.
	Timeout time.Duration
}

var _ Engine = (*Exec)(nil)

// Layout runs `dot -Tsvg` with the DOT source on stdin and returns its
// output. It returns a *NotFoundError when Path is empty or missing, a
// *TimeoutError when dot runs longer than the timeout (dot is then stopped),
// and a *FailureError, carrying dot's stderr, when dot exits with an error or
// its output is empty or not SVG. When ctx is canceled the context's error is
// returned.
func (e *Exec) Layout(ctx context.Context, dot string) ([]byte, error) {
	if e.Path == "" {
		return nil, &NotFoundError{State: NullUndefined}
	}
	p, state := e.createFile3(ctx, dot, false, "svg")
	switch state.name {
	case timeoutState.name:
		return nil, &TimeoutError{Path: e.Path, Timeout: e.timeout()}
	case "EXCEPTION":
		if ctx.Err() != nil {
			return nil, fmt.Errorf("dot: %w", state.cause)
		}
		if errors.Is(state.cause, fs.ErrNotExist) {
			return nil, &NotFoundError{Path: e.Path, State: DoesNotExist}
		}
		return nil, &FailureError{Path: e.Path, ExitCode: -1, Err: state.cause}
	}

	stderr := trin(string(p.stderr))
	switch {
	case p.exitCode != 0:
		return nil, &FailureError{Path: e.Path, ExitCode: p.exitCode, Stderr: stderr}
	case len(p.stdout) == 0:
		// GraphvizImageBuilder turns an empty SVG into a crash report.
		return nil, &FailureError{Path: e.Path, Stderr: stderr, Reason: "empty output"}
	case !bytes.Contains(p.stdout, []byte("<svg")):
		return nil, &FailureError{Path: e.Path, Stderr: stderr, Reason: "output is not SVG"}
	}
	return p.stdout, nil
}

// DotVersion returns the output of `dot -V` (which dot prints on stderr) on
// one line, or "?" when dot cannot be run or times out.
func (e *Exec) DotVersion(ctx context.Context) string {
	p := &processRunner{cmd: e.commandLineVersion(), timeout: e.timeout()}
	if p.run(ctx, nil, true).differs(terminatedOK) {
		return "?"
	}
	// Upstream also appends the runner's error text, which is always null
	// after a run that terminated.
	return trin(strings.ReplaceAll(string(p.stdout), "\n", " "))
}

// GraphvizVersion runs `dot -V` and returns the layout tweaks for that
// version (upstream's GraphvizVersionFinder.getVersion). Callers should cache
// the result per executable.
func (e *Exec) GraphvizVersion(ctx context.Context) GraphvizVersion {
	return NewGraphvizVersion(e.DotVersion(ctx))
}

// createFile3 runs dot with the DOT source on stdin, one -T option per type.
func (e *Exec) createFile3(ctx context.Context, dot string, redirectErrorStream bool, types ...string) (*processRunner, processState) {
	p := &processRunner{cmd: e.commandLine(types...), timeout: e.timeout()}
	return p, p.run(ctx, []byte(dot), redirectErrorStream)
}

func (e *Exec) timeout() time.Duration {
	if e.Timeout <= 0 {
		return DefaultTimeout
	}
	return e.Timeout
}

// commandLine is the dot command with its absolute path, so that a relative
// path is never searched on PATH.
func (e *Exec) commandLine(types ...string) []string {
	result := []string{absolutePath(e.Path)}
	for _, t := range types {
		result = append(result, "-T"+t)
	}
	return result
}

func (e *Exec) commandLineVersion() []string {
	return []string{absolutePath(e.Path), "-V"}
}
