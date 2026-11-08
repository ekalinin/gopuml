// SPDX-License-Identifier: MIT
// Provenance: original

package freelist

import (
	"image/png"
	"runtime"
	"sync"
	"testing"
)

// A list of PNG encoder buffers is the BufferPool of a png.Encoder.
var _ png.EncoderBufferPool = (*List[*png.EncoderBuffer])(nil)

func TestGetEmpty(t *testing.T) {
	if v := New(4, func() *int { return new(1) }).Get(); v == nil || *v != 1 {
		t.Errorf("Get of an empty list = %v, want a created 1", v)
	}
	if v := New[*int](4, nil).Get(); v != nil {
		t.Errorf("Get of an empty list without create = %v, want nil", v)
	}
}

func TestPutThenGet(t *testing.T) {
	l := New(4, func() *int { return new(0) })
	v := new(7)
	l.Put(v)
	if got := l.Get(); got != v {
		t.Errorf("Get = %p, want the value put back %p", got, v)
	}
}

func TestFull(t *testing.T) {
	l := New[*int](4, nil)
	var kept []*int
	for i := range 4 {
		v := new(i)
		kept = append(kept, v)
		l.Put(v)
	}
	l.Put(new(4)) // dropped
	for i := range 4 {
		if got := l.Get(); got != kept[i] {
			t.Errorf("Get %d = %p, want %p", i, got, kept[i])
		}
	}
	if got := l.Get(); got != nil {
		t.Errorf("Get after four values = %v, want nil: the fifth value is dropped", *got)
	}
}

// A value put back survives garbage collections, which empty a sync.Pool.
func TestSurvivesGC(t *testing.T) {
	l := New(4, func() *int { return new(0) })
	v := new(7)
	l.Put(v)
	runtime.GC()
	runtime.GC()
	if got := l.Get(); got != v {
		t.Errorf("Get after two collections = %p, want the value put back %p", got, v)
	}
}

// Run with -race.
func TestConcurrent(t *testing.T) {
	l := New(4, func() *int { return new(0) })
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 1000 {
				v := l.Get()
				*v++
				l.Put(v)
			}
		})
	}
	wg.Wait()
	if n := len(l.idle); n > 4 {
		t.Errorf("%d idle values, want at most 4", n)
	}
}
