// SPDX-License-Identifier: MIT
// Provenance: original

// Package dot provides the layout engines of graph-based diagrams: the
// built-in engine gogv, and the Graphviz dot executable that the caller
// names.
//
// It replaces upstream's GraphvizRuntimeEnvironment singleton: the explicit
// dot path, the timeout and the host lookups (environment and files) are
// passed in through Config and Host, so concurrent renders can use different
// settings and tests can fake the host. Config.NewEngine chooses the engine:
// the dot named by the explicit path or GRAPHVIZ_DOT, otherwise gogv.
//
// The Gogv type is the built-in engine: it lays graphs out in the process
// with github.com/ekalinin/gogv, a Go implementation of dot. The Exec type is
// the "dotexec" layout engine: it runs `dot -Tsvg`. Both enforce a timeout
// (30 seconds by default) and report the engine's error message when the
// layout fails. Errors can be tested with errors.Is against ErrNotFound,
// ErrTimeout and ErrFailed, or inspected with errors.As as *NotFoundError,
// *TimeoutError and *FailureError.
package dot
