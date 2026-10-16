// SPDX-License-Identifier: MIT
// Provenance: original

//go:build js

package svek

// execSupported is false in the WebAssembly build: it cannot run
// programs, so graph layouts are not available (wasm-build spec) until a
// pure Go engine exists.
const execSupported = false
