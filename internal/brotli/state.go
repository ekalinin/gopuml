// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/State.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

import "io"

// state is the decoder state. Java's int fields are Go ints; the Huffman
// tables are int32, as their entries fit. Only the 64-bit build of the bit
// reader is ported, so Java's shortBuffer and accumulator32 are dropped.
type state struct {
	ringBuffer     []byte
	contextModes   []byte
	contextMap     []byte
	distContextMap []byte
	output         []byte
	byteBuffer     []byte // bitReader

	intBuffer  []uint32 // bitReader
	rings      []int
	blockTrees []int32
	hGroup0    []int32
	hGroup1    []int32
	hGroup2    []int32

	accumulator64 uint64 // bitReader: pre-fetched bits.

	runningState           int // Default value is 0 == uninitialized
	nextRunningState       int
	bitOffset              int // bitReader: bit-reading position in accumulator.
	halfOffset             int // bitReader: offset of next item in intBuffer.
	tailBytes              int // bitReader: number of bytes in unfinished half.
	endOfStreamReached     int // bitReader: input stream is finished.
	metaBlockLength        int
	inputEnd               int
	isUncompressed         int
	isMetadata             int
	literalBlockLength     int
	numLiteralBlockTypes   int
	commandBlockLength     int
	numCommandBlockTypes   int
	distanceBlockLength    int
	numDistanceBlockTypes  int
	pos                    int
	maxDistance            int
	distRbIdx              int
	trivialLiteralContext  int
	literalTreeIndex       int
	literalTree            int
	j                      int
	insertLength           int
	contextMapSlice        int
	distContextMapSlice    int
	contextLookupOffset1   int
	contextLookupOffset2   int
	treeCommandOffset      int
	distanceCode           int
	numDirectDistanceCodes int
	distancePostfixMask    int
	distancePostfixBits    int
	distance               int
	copyLength             int
	copyDst                int
	maxBackwardDistance    int
	maxRingBufferSize      int
	ringBufferSize         int
	expectedTotalSize      int
	outputOffset           int
	outputLength           int
	outputUsed             int
	bytesWritten           int
	bytesToWrite           int

	input io.Reader // bitReader
	// inputErr is the last error of input, kept by readInput until the bytes
	// read with it are used (Go only: a Java stream throws at once).
	inputErr error
}

func newState() *state {
	s := &state{}
	s.ringBuffer = []byte{}
	s.rings = make([]int, 10)
	s.rings[0] = 16
	s.rings[1] = 15
	s.rings[2] = 11
	s.rings[3] = 4
	return s
}
