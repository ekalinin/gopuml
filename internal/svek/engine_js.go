// SPDX-License-Identifier: MIT
// Provenance: original

//go:build js

package svek

// execSupported is false in the WebAssembly build: it cannot run
// programs, so graphs are always laid out by the built-in engine gogv
// (wasm-build spec).
const execSupported = false
