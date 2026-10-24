// SPDX-License-Identifier: MIT
// Provenance: original

// Package brotli decodes Brotli streams (RFC 7932), such as the .spm files of
// PlantUML's standard library. It ports upstream's
// net.sourceforge.plantuml.brotli package, which is Google's Java decoder
// (Copyright 2015 Google Inc., MIT License,
// https://github.com/google/brotli/blob/master/LICENSE).
//
// The Java decoder throws an unchecked BrotliRuntimeException on malformed
// input. The port panics with it inside the package, and Reader.Read
// recovers it and returns an error, as BrotliInputStream wraps it into an
// IOException.
package brotli
