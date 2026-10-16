// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/json/WritingBuffer.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package json

import "io"

// writingBuffer is a lightweight writing buffer that reduces the number of
// write operations on the underlying writer. It is not safe for concurrent
// use. The first write error is kept and makes every later write a no-op;
// flush reports it.
type writingBuffer struct {
	writer io.Writer
	buffer []byte
	fill   int
	err    error
}

func newWritingBuffer(writer io.Writer, bufferSize int) *writingBuffer {
	return &writingBuffer{writer: writer, buffer: make([]byte, bufferSize)}
}

// writeByte writes a single byte.
func (b *writingBuffer) writeByte(c byte) {
	if b.fill > len(b.buffer)-1 {
		b.flush()
	}
	b.buffer[b.fill] = c
	b.fill++
}

// writeString writes s, directly to the underlying writer when it does not
// fit in the buffer.
func (b *writingBuffer) writeString(s string) {
	if b.fill > len(b.buffer)-len(s) {
		b.flush()
		if len(s) > len(b.buffer) {
			b.writeThrough([]byte(s))
			return
		}
	}
	b.fill += copy(b.buffer[b.fill:], s)
}

// flush writes the buffered bytes to the underlying writer, which is not
// flushed itself, and returns the first write error.
func (b *writingBuffer) flush() error {
	b.writeThrough(b.buffer[:b.fill])
	b.fill = 0
	return b.err
}

func (b *writingBuffer) writeThrough(p []byte) {
	if b.err != nil || len(p) == 0 {
		return
	}
	_, b.err = b.writer.Write(p)
}
