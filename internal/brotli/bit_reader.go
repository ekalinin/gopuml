// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/BitReader.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

// Bit reading helpers. Upstream selects a 32-bit or a 64-bit build with
// LOG_BITNESS and uses the 64-bit one: only that one is ported (readBits
// never needs readManyBits there).

const (
	// Added by Arnaud... not very beautiful
	checkUnusedBytesAfterEnd = false

	logBitness = 6
	bitness    = 1 << logBitness

	byteness = bitness / 8
	capacity = 4096
	// After encountering the end of the input stream, this amount of zero
	// bytes will be appended.
	slack      = 64
	bufferSize = capacity + slack
	// Don't bother to replenish the buffer while this number of bytes is
	// available.
	safeguard = 36
	waterline = capacity - safeguard

	// "Half" refers to "half of native integer type", i.e. on 64-bit
	// machines it is 32-bit type.
	halfBitness    = bitness / 2
	halfSize       = byteness / 2
	halvesCapacity = capacity / halfSize
	halfBufferSize = bufferSize / halfSize
	halfWaterline  = waterline / halfSize

	logHalfSize = logBitness - 4
)

// readMoreInput fills up the input buffer. It is a no-op if there are at
// least 36 bytes present after the current position.
func readMoreInput(s *state) {
	if s.halfOffset > halfWaterline {
		doReadMoreInput(s)
	}
}

func doReadMoreInput(s *state) {
	if s.endOfStreamReached != 0 {
		if halfAvailable(s) >= -2 {
			return
		}
		throw("No more input")
	}
	readOffset := s.halfOffset << logHalfSize
	bytesInBuffer := capacity - readOffset
	// Move unused bytes to the head of the buffer.
	copyBytesWithin(s.byteBuffer, 0, readOffset, capacity)
	s.halfOffset = 0
	for bytesInBuffer < capacity {
		spaceLeft := capacity - bytesInBuffer
		n := readInput(s, s.byteBuffer, bytesInBuffer, spaceLeft)
		// EOF is -1 in Java, but 0 in C#.
		if n <= 0 {
			s.endOfStreamReached = 1
			s.tailBytes = bytesInBuffer
			bytesInBuffer += halfSize - 1
			break
		}
		bytesInBuffer += n
	}
	bytesToNibbles(s, bytesInBuffer)
}

func checkHealth(s *state, endOfStream int) {
	if s.endOfStreamReached == 0 {
		return
	}
	byteOffset := (s.halfOffset << logHalfSize) + ((s.bitOffset + 7) >> 3) - byteness
	if byteOffset > s.tailBytes {
		throw("Read after end")
	}
	if checkUnusedBytesAfterEnd && endOfStream != 0 && byteOffset != s.tailBytes {
		throw("Unused bytes after end")
	}
}

// fillBitWindow is doFillBitWindow when half of the accumulator is used.
// Java repeats the body because the JVM fails to inline it; Go inlines it.
func fillBitWindow(s *state) {
	if s.bitOffset >= halfBitness {
		doFillBitWindow(s)
	}
}

func doFillBitWindow(s *state) {
	s.accumulator64 = uint64(s.intBuffer[s.halfOffset])<<halfBitness | s.accumulator64>>halfBitness
	s.halfOffset++
	s.bitOffset -= halfBitness
}

// peekBits returns the low 32 bits of the accumulator from bitOffset. The
// shift count is masked as Java masks the shift count of a long.
func peekBits(s *state) int {
	return int(uint32(s.accumulator64 >> (uint(s.bitOffset) & 63)))
}

func readFewBits(s *state, n int) int {
	val := peekBits(s) & (1<<n - 1)
	s.bitOffset += n
	return val
}

func readBits(s *state, n int) int {
	return readFewBits(s, n)
}

func initBitReader(s *state) {
	s.byteBuffer = make([]byte, bufferSize)
	s.accumulator64 = 0
	s.intBuffer = make([]uint32, halfBufferSize)
	s.bitOffset = bitness
	s.halfOffset = halvesCapacity
	s.endOfStreamReached = 0
	prepare(s)
}

func prepare(s *state) {
	readMoreInput(s)
	checkHealth(s, 0)
	doFillBitWindow(s)
	doFillBitWindow(s)
}

func reload(s *state) {
	if s.bitOffset == bitness {
		prepare(s)
	}
}

func jumpToByteBoundary(s *state) {
	padding := (bitness - s.bitOffset) & 7
	if padding != 0 {
		paddingBits := readFewBits(s, padding)
		if paddingBits != 0 {
			throw("Corrupted padding bits")
		}
	}
}

func halfAvailable(s *state) int {
	limit := halvesCapacity
	if s.endOfStreamReached != 0 {
		limit = (s.tailBytes + (halfSize - 1)) >> logHalfSize
	}
	return limit - s.halfOffset
}

func copyBytes(s *state, data []byte, offset, length int) {
	if s.bitOffset&7 != 0 {
		throw("Unaligned copyBytes")
	}

	// Drain accumulator.
	for s.bitOffset != bitness && length != 0 {
		data[offset] = byte(peekBits(s))
		offset++
		s.bitOffset += 8
		length--
	}
	if length == 0 {
		return
	}

	// Get data from shadow buffer with "sizeof(int)" granularity.
	copyNibbles := min(halfAvailable(s), length>>logHalfSize)
	if copyNibbles > 0 {
		readOffset := s.halfOffset << logHalfSize
		delta := copyNibbles << logHalfSize
		copy(data[offset:offset+delta], s.byteBuffer[readOffset:readOffset+delta])
		offset += delta
		length -= delta
		s.halfOffset += copyNibbles
	}
	if length == 0 {
		return
	}

	// Read tail bytes.
	if halfAvailable(s) > 0 {
		// length = 1..3
		fillBitWindow(s)
		for length != 0 {
			data[offset] = byte(peekBits(s))
			offset++
			s.bitOffset += 8
			length--
		}
		checkHealth(s, 0)
		return
	}

	// Now it is possible to copy bytes directly.
	for length > 0 {
		n := readInput(s, data, offset, length)
		if n == -1 {
			throw("Unexpected end of input")
		}
		offset += n
		length -= n
	}
}

// bytesToNibbles translates bytes to halves (little-endian uint32).
func bytesToNibbles(s *state, byteLen int) {
	byteBuffer := s.byteBuffer
	halfLen := byteLen >> logHalfSize
	intBuffer := s.intBuffer
	for i := range halfLen {
		intBuffer[i] = uint32(byteBuffer[i*4]) | uint32(byteBuffer[(i*4)+1])<<8 |
			uint32(byteBuffer[(i*4)+2])<<16 | uint32(byteBuffer[(i*4)+3])<<24
	}
}
