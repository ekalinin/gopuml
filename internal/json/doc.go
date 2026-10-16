// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/package-info.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

// Package json manages JSON data with the semantics of PlantUML's fork of
// the minimal-json library (net.sourceforge.plantuml.json).
//
// The PlantUML preprocessor prints JSON text, so this package reproduces the
// upstream behavior exactly: the minimal and pretty-printed serialization,
// the string escape table, the member order and duplicate members of
// objects, number text kept verbatim as parsed, the parser leniency (line
// comments inside objects) and its error messages and locations.
//
// # Mapping from upstream
//
//   - JsonValue is the [Value] interface, JsonObject is [*Object] and
//     JsonArray is [*Array]. JsonString, JsonNumber and JsonLiteral are
//     unexported implementations of [Value].
//   - Json.NULL, Json.TRUE and Json.FALSE are [Null], [True] and [False].
//   - The Json.value overloads are the ValueOf* functions, for example
//     [ValueOfInt] and [ValueOfDouble]. The typed add and set overloads of
//     JsonArray and JsonObject, and the typed Json.array overloads, are
//     expressed with [Array.Add], [Object.Add], [Object.Set] and a ValueOf*
//     argument.
//   - ParseException is [*ParseError]; its Error text is upstream's
//     getMessage, for example "Unexpected end of input at 1:5".
//   - WriterConfig.MINIMAL and WriterConfig.PRETTY_PRINT are [Minimal] and
//     [Pretty]; PrettyPrint is [*PrettyPrint].
//   - toString() is String() and toString(WriterConfig) is StringConfig. A
//     nil WriterConfig means the minimal form, where upstream throws
//     NullPointerException.
//   - JsonValue.hashCode is not ported; Equal compares values like equals.
//   - Deprecated upstream members (readFrom, valueOf, the deprecated
//     constants and ParseException's offset getters) are not ported.
//
// # Errors and panics
//
// Where upstream throws a checked or runtime exception for a condition that
// depends on the data (a parse error, a number that does not fit an int, an
// index out of range, NaN given to a factory), the Go function returns an
// error. Upstream's UnsupportedOperationException for wrong-type access (for
// example AsObject on a number) and for modifying an unmodifiable view
// panics with an [*UnsupportedOperationError]; callers check the type with
// the Is* methods first, as upstream callers do. Passing a nil [Value],
// [*Array], [*Object] or [Handler] where upstream throws NullPointerException
// is a programming error and panics as well.
//
// # Strings
//
// Java strings are sequences of UTF-16 code units. The parser works on the
// UTF-16 form of its input, so the offsets and columns of a [Location] count
// UTF-16 units, as upstream does. Parsed strings are returned as UTF-8; a
// lone surrogate produced by a \u escape becomes U+FFFD because a Go string
// cannot hold it.
package json
