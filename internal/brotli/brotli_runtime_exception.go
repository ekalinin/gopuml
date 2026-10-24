// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/BrotliRuntimeException.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

// brotliRuntimeException is upstream's unchecked exception of the decoder,
// with its message and cause. The decoder panics with it, and Reader.Read
// recovers it.
type brotliRuntimeException struct {
	message string
	cause   error
}

func (e *brotliRuntimeException) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e *brotliRuntimeException) Unwrap() error { return e.cause }

// throw panics with a brotliRuntimeException, where upstream throws one.
func throw(message string) {
	panic(&brotliRuntimeException{message: message})
}

// throwCause panics with a brotliRuntimeException that has a cause.
func throwCause(message string, cause error) {
	panic(&brotliRuntimeException{message: message, cause: cause})
}
