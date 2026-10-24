// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/BrotliInputStream.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

import (
	"fmt"
	"io"
)

// Reader is upstream's BrotliInputStream: an io.Reader that decompresses
// the Brotli data of its source. It is not safe for concurrent use.
//
// Unlike Java's constructor, NewReader does not read the source: the decoder
// is initialized by the first Read, which reports an initialization
// failure. Java's internal buffer for byte-by-byte reading and close are
// not ported.
type Reader struct {
	source      io.Reader
	state       *state
	initialized bool
	// err is returned by every Read after the first error: io.EOF at the
	// end of the stream, or the failure of the decoder, whose state is
	// unusable after it.
	err error
}

// NewReader returns a Reader that decompresses the Brotli stream read from
// r.
func NewReader(r io.Reader) *Reader {
	return &Reader{source: r, state: newState()}
}

// Read decompresses data into p. It returns io.EOF at the end of the stream,
// and an error, starting with the message of upstream's IOException, when
// the stream is malformed or truncated or the source fails.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	message := "Brotli stream decoding failed"
	defer func() {
		if v := recover(); v != nil {
			r.state.output = nil
			r.err = recoveredError(message, v)
			n, err = 0, r.err
		}
	}()
	s := r.state
	if !r.initialized {
		message = "Brotli decoder initialization failed"
		r.initialized = true
		initState(s, r.source)
		message = "Brotli stream decoding failed"
	}
	s.output = p
	s.outputOffset = 0
	s.outputLength = len(p)
	s.outputUsed = 0
	decompress(s)
	// Do not keep the caller's buffer.
	s.output = nil
	if s.outputUsed == 0 {
		r.err = io.EOF
		return 0, io.EOF
	}
	return s.outputUsed, nil
}

// recoveredError turns a panic of the decoder into the error of Read, as
// upstream wraps a BrotliRuntimeException into an IOException with message.
// A Go runtime error, such as an index out of range where Java would throw
// an unchecked exception out of the stream, is returned the same way.
func recoveredError(message string, v any) error {
	cause, ok := v.(error)
	if !ok {
		cause = fmt.Errorf("%v", v)
	}
	return fmt.Errorf("%s: %w", message, cause)
}
