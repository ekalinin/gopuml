// SPDX-License-Identifier: MIT
// Provenance: original

package dot

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Sentinel errors, matched with errors.Is by the typed errors below.
var (
	// ErrNotFound reports that no usable dot executable was found.
	ErrNotFound = errors.New("Graphviz dot executable not found")
	// ErrTimeout reports a dot run stopped after the timeout.
	ErrTimeout = errors.New("Graphviz dot timed out")
	// ErrFailed reports a dot run that failed or gave unusable output.
	ErrFailed = errors.New("Graphviz dot failed")
)

// NotFoundError reports that the dot executable is missing or not usable.
type NotFoundError struct {
	// Path is the executable that was tried, "" when there was none.
	Path string
	// State tells why Path cannot be used.
	State ExeState
}

func (e *NotFoundError) Error() string {
	return "cannot find Graphviz dot: " + e.State.TextMessageFor(e.Path)
}

// Is makes errors.Is(err, ErrNotFound) true.
func (e *NotFoundError) Is(target error) bool { return target == ErrNotFound }

// Lines returns the text of upstream's error diagram for this error.
func (e *NotFoundError) Lines() []string { return NotFoundLines(e.Path, e.State) }

// TimeoutError reports that dot ran longer than the timeout and was stopped.
type TimeoutError struct {
	// Path is the dot executable.
	Path string
	// Timeout is the limit that was exceeded.
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("Graphviz dot timed out after %s and was stopped (%s)", e.Timeout, e.Path)
}

// Is makes errors.Is(err, ErrTimeout) true.
func (e *TimeoutError) Is(target error) bool { return target == ErrTimeout }

// Unwrap makes errors.Is(err, context.DeadlineExceeded) true.
func (e *TimeoutError) Unwrap() error { return context.DeadlineExceeded }

// FailureError reports that dot exited with an error, could not be run, or
// gave output that cannot be used.
type FailureError struct {
	// Path is the dot executable.
	Path string
	// ExitCode is dot's exit status: 0 when dot succeeded but its output is
	// unusable, -1 when dot could not be run or did not exit normally.
	ExitCode int
	// Stderr is dot's error output, trimmed.
	Stderr string
	// Reason describes unusable output ("empty output", "output is not SVG"),
	// "" otherwise.
	Reason string
	// Err is the error that prevented running dot, if any.
	Err error
}

func (e *FailureError) Error() string {
	msg := "Graphviz dot failed"
	switch {
	case e.Err != nil:
		msg += ": " + e.Err.Error()
	case e.Reason != "":
		msg += ": " + e.Reason
	default:
		msg += fmt.Sprintf(": exit status %d", e.ExitCode)
	}
	if e.Stderr != "" {
		msg += ": " + e.Stderr
	}
	return msg
}

// Is makes errors.Is(err, ErrFailed) true.
func (e *FailureError) Is(target error) bool { return target == ErrFailed }

// Unwrap returns the error that prevented running dot, if any.
func (e *FailureError) Unwrap() error { return e.Err }
