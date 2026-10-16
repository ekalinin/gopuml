// SPDX-License-Identifier: MIT
// Provenance: original

// Package dot locates and runs the Graphviz dot executable.
//
// It replaces upstream's GraphvizRuntimeEnvironment singleton: the explicit
// dot path, the timeout and the host lookups (environment and files) are
// passed in through Config and Host, so concurrent renders can use different
// settings and tests can fake the host.
//
// The Exec type is the "dotexec" layout engine: it runs `dot -Tsvg`, enforces
// a timeout (30 seconds by default) and reports dot's error message when dot
// fails. Errors can be tested with errors.Is against ErrNotFound, ErrTimeout
// and ErrFailed, or inspected with errors.As as *NotFoundError, *TimeoutError
// and *FailureError.
package dot
