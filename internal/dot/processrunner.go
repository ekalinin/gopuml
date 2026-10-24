// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/ProcessRunner.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/ProcessState.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/GlobalConfigKey.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// DefaultTimeout is the time limit of one dot run when none is configured
// (upstream's TIMEOUT_MS default).
const DefaultTimeout = 30 * time.Second

// killDelay is how long a stopped process may take to exit before it is
// killed, as upstream waits 500 ms between destroy and destroyForcibly. It
// also bounds the wait for output pipes held open by child processes.
const killDelay = 500 * time.Millisecond

// processState is the outcome of a process run.
type processState struct {
	name  string
	cause error
}

var (
	terminatedOK = processState{name: "TERMINATED_OK"}
	timeoutState = processState{name: "TIMEOUT"}
)

func exceptionState(cause error) processState {
	return processState{name: "EXCEPTION", cause: cause}
}

func (s processState) String() string {
	if s.cause == nil {
		return s.name
	}
	return s.name + " " + s.cause.Error()
}

func (s processState) differs(other processState) bool {
	return s.name != other.name
}

// errProcessTimeout is the context cause set when the run timeout expires.
var errProcessTimeout = errors.New("dot: process timeout")

// processRunner runs one command, feeding it stdin and capturing its output.
//
// Upstream applies its timeout only after the whole output has been read, so a
// dot process that never finishes is never stopped. Here the timeout covers
// the whole run: when it expires the process is asked to terminate, then
// killed after killDelay.
type processRunner struct {
	cmd     []string
	timeout time.Duration

	stdout   []byte
	stderr   []byte
	exitCode int
}

// run starts the command, writes in to its stdin (when not nil) and waits for
// it. With redirectErrorStream, stderr is captured with stdout, as upstream
// always does; otherwise it is kept apart. The exit status does not change the
// state, as upstream; it is kept in exitCode (-1 when the process did not
// exit normally).
func (p *processRunner) run(ctx context.Context, in []byte, redirectErrorStream bool) processState {
	ctx, cancel := context.WithTimeoutCause(ctx, p.timeout, errProcessTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.cmd[0], p.cmd[1:]...)
	if in != nil {
		cmd.Stdin = bytes.NewReader(in)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	if redirectErrorStream {
		cmd.Stderr = &stdout
	} else {
		cmd.Stderr = &stderr
	}
	cmd.Cancel = func() error { return terminate(cmd.Process) }
	cmd.WaitDelay = killDelay

	err := cmd.Run()
	p.stdout, p.stderr = stdout.Bytes(), stderr.Bytes()
	p.exitCode = -1
	if cmd.ProcessState != nil {
		p.exitCode = cmd.ProcessState.ExitCode()
	}

	if err == nil {
		return terminatedOK
	}
	if errors.Is(context.Cause(ctx), errProcessTimeout) {
		return timeoutState
	}
	if ctx.Err() != nil {
		// The caller's context was canceled.
		return exceptionState(context.Cause(ctx))
	}
	if _, exited := errors.AsType[*exec.ExitError](err); exited || errors.Is(err, exec.ErrWaitDelay) {
		return terminatedOK
	}
	return exceptionState(err)
}

// terminate asks the process to stop, as Java's Process.destroy does.
func terminate(p *os.Process) error {
	if runtime.GOOS == "windows" {
		return p.Kill()
	}
	return p.Signal(syscall.SIGTERM)
}
