// SPDX-License-Identifier: MIT
// Provenance: original

package freelist

// List holds up to a fixed number of idle values for reuse. It is safe for
// concurrent use, and neither Get nor Put blocks. A value taken with Get
// belongs to its caller until Put gives it back.
//
// With Get() T and Put(T), a *List[*png.EncoderBuffer] without create is a
// png.EncoderBufferPool.
type List[T any] struct {
	idle   chan T
	create func() T
}

// New returns a list that keeps up to n idle values; create, which may be
// nil, makes a value when the list is empty.
func New[T any](n int, create func() T) *List[T] {
	return &List[T]{idle: make(chan T, n), create: create}
}

// Get returns an idle value, or else a value made by create, or else the
// zero value.
func (l *List[T]) Get() T {
	select {
	case v := <-l.idle:
		return v
	default:
	}
	if l.create != nil {
		return l.create()
	}
	var zero T
	return zero
}

// Put keeps a value for a later Get, or drops it when the list is full.
func (l *List[T]) Put(v T) {
	select {
	case l.idle <- v:
	default:
	}
}
