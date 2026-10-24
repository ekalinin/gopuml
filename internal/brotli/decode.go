// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/Decode.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

import (
	"io"
	"strconv"
)

// Brotli decompression.

// RunningState
const (
	uninitialized        = 0
	blockStart           = 1
	compressedBlockStart = 2
	mainLoop             = 3
	readMetadata         = 4
	copyUncompressed     = 5
	insertLoop           = 6
	copyLoop             = 7
	copyWrapBuffer       = 8
	transform            = 9
	finished             = 10
	closed               = 11
	write                = 12
)

const (
	defaultCodeLength     = 8
	codeLengthRepeatCode  = 16
	numLiteralCodes       = 256
	numInsertAndCopyCodes = 704
	numBlockLengthCodes   = 26
	literalContextBits    = 6
	distanceContextBits   = 2

	huffmanTableBits = 8
	huffmanTableMask = 0xFF

	// huffmanTableSize is the maximum possible Huffman table size for an
	// alphabet size of 704, max code length 15 and root table bits 8.
	huffmanTableSize = 1080

	codeLengthCodes = 18

	numDistanceShortCodes = 16

	minWordLength = 4

	maxWordLength = 24

	maxTransformedWordLength = 5 + maxWordLength + 8
)

var codeLengthCodeOrder = [...]int{1, 2, 3, 4, 0, 5, 17, 6, 16, 7, 8, 9, 10, 11, 12, 13, 14, 15}

var distanceShortCodeIndexOffset = [...]int{3, 2, 1, 0, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2}

var distanceShortCodeValueOffset = [...]int{0, 0, 0, 0, -1, 1, -2, 2, -3, 3, -1, 1, -2, 2, -3, 3}

// fixedTable is the static Huffman code for the code length code lengths.
var fixedTable = [...]int{0x020000, 0x020004, 0x020003, 0x030002, 0x020000, 0x020004, 0x020003,
	0x040001, 0x020000, 0x020004, 0x020003, 0x030002, 0x020000, 0x020004, 0x020003, 0x040005}

var dictionaryOffsetsByLength = [...]int{0, 0, 0, 0, 0, 4096, 9216, 21504, 35840, 44032, 53248, 63488,
	74752, 87040, 93696, 100864, 104704, 106752, 108928, 113536, 115968, 118528, 119872, 121280, 122016}

var dictionarySizeBitsByLength = [...]int{0, 0, 0, 0, 10, 10, 11, 11, 10, 10, 10, 10, 10, 9, 9, 8, 7, 7,
	8, 7, 7, 6, 6, 5, 5}

// Prefix code LUT.
var blockLengthOffset = [...]int{1, 5, 9, 13, 17, 25, 33, 41, 49, 65, 81, 97, 113, 145, 177, 209, 241,
	305, 369, 497, 753, 1265, 2289, 4337, 8433, 16625}

var blockLengthNBits = [...]int{2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 6, 6, 7, 8, 9, 10, 11,
	12, 13, 24}

var insertLengthOffset = [...]int{0, 1, 2, 3, 4, 5, 6, 8, 10, 14, 18, 26, 34, 50, 66, 98, 130, 194, 322,
	578, 1090, 2114, 6210, 22594}

var insertLengthNBits = [...]int{0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 7, 8, 9, 10, 12, 14,
	24}

var copyLengthOffset = [...]int{2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 14, 18, 22, 30, 38, 54, 70, 102, 134, 198,
	326, 582, 1094, 2118}

var copyLengthNBits = [...]int{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 7, 8, 9, 10,
	24}

var insertRangeLut = [...]int{0, 0, 8, 8, 0, 16, 8, 16, 16}

var copyRangeLut = [...]int{0, 8, 0, 8, 16, 0, 16, 8, 16}

func decodeWindowBits(s *state) int {
	fillBitWindow(s)
	if readFewBits(s, 1) == 0 {
		return 16
	}
	n := readFewBits(s, 3)
	if n != 0 {
		return 17 + n
	}
	n = readFewBits(s, 3)
	if n != 0 {
		return 8 + n
	}
	return 17
}

// initState associates input with s, an uninitialized state without
// associated input. Upstream's IllegalStateException is a
// brotliRuntimeException here.
func initState(s *state, input io.Reader) {
	if s.runningState != uninitialized {
		throw("State MUST be uninitialized")
	}
	s.blockTrees = make([]int32, 6*huffmanTableSize)
	s.input = input
	initBitReader(s)
	windowBits := decodeWindowBits(s)
	if windowBits == 9 { // Reserved case for future expansion.
		throw("Invalid 'windowBits' code")
	}
	s.maxRingBufferSize = 1 << windowBits
	s.maxBackwardDistance = s.maxRingBufferSize - 16
	s.runningState = blockStart
}

// decodeVarLenUnsignedByte decodes a number in the range [0..255], by
// reading 1 - 11 bits.
func decodeVarLenUnsignedByte(s *state) int {
	fillBitWindow(s)
	if readFewBits(s, 1) != 0 {
		n := readFewBits(s, 3)
		if n == 0 {
			return 1
		}
		return readFewBits(s, n) + (1 << n)
	}
	return 0
}

func decodeMetaBlockLength(s *state) {
	fillBitWindow(s)
	s.inputEnd = readFewBits(s, 1)
	s.metaBlockLength = 0
	s.isUncompressed = 0
	s.isMetadata = 0
	if s.inputEnd != 0 && readFewBits(s, 1) != 0 {
		return
	}
	sizeNibbles := readFewBits(s, 2) + 4
	if sizeNibbles == 7 {
		s.isMetadata = 1
		if readFewBits(s, 1) != 0 {
			throw("Corrupted reserved bit")
		}
		sizeBytes := readFewBits(s, 2)
		if sizeBytes == 0 {
			return
		}
		for i := range sizeBytes {
			fillBitWindow(s)
			bits := readFewBits(s, 8)
			if bits == 0 && i+1 == sizeBytes && sizeBytes > 1 {
				throw("Exuberant nibble")
			}
			s.metaBlockLength |= bits << (i * 8)
		}
	} else {
		for i := range sizeNibbles {
			fillBitWindow(s)
			bits := readFewBits(s, 4)
			if bits == 0 && i+1 == sizeNibbles && sizeNibbles > 4 {
				throw("Exuberant nibble")
			}
			s.metaBlockLength |= bits << (i * 4)
		}
	}
	s.metaBlockLength++
	if s.inputEnd == 0 {
		s.isUncompressed = readFewBits(s, 1)
	}
}

// readSymbol decodes the next Huffman code from bit-stream.
func readSymbol(table []int32, offset int, s *state) int {
	val := peekBits(s)
	offset += val & huffmanTableMask
	bits := int(table[offset] >> 16)
	sym := int(table[offset] & 0xFFFF)
	if bits <= huffmanTableBits {
		s.bitOffset += bits
		return sym
	}
	offset += sym
	mask := (1 << bits) - 1
	offset += (val & mask) >> huffmanTableBits
	s.bitOffset += int(table[offset]>>16) + huffmanTableBits
	return int(table[offset] & 0xFFFF)
}

func readBlockLength(table []int32, offset int, s *state) int {
	fillBitWindow(s)
	code := readSymbol(table, offset, s)
	n := blockLengthNBits[code]
	fillBitWindow(s)
	return blockLengthOffset[code] + readBits(s, n)
}

func translateShortCodes(code int, ringBuffer []int, index int) int {
	if code < numDistanceShortCodes {
		index += distanceShortCodeIndexOffset[code]
		index &= 3
		return ringBuffer[index] + distanceShortCodeValueOffset[code]
	}
	return code - numDistanceShortCodes + 1
}

func moveToFront(v []int, index int) {
	value := v[index]
	for ; index > 0; index-- {
		v[index] = v[index-1]
	}
	v[0] = value
}

func inverseMoveToFrontTransform(v []byte, vLen int) {
	var mtf [256]int
	for i := range 256 {
		mtf[i] = i
	}
	for i := range vLen {
		index := int(v[i])
		v[i] = byte(mtf[index])
		if index != 0 {
			moveToFront(mtf[:], index)
		}
	}
}

func readHuffmanCodeLengths(codeLengthCodeLengths []int, numSymbols int, codeLengths []int, s *state) {
	symbol := 0
	prevCodeLen := defaultCodeLength
	repeat := 0
	repeatCodeLen := 0
	space := 32768
	table := make([]int32, 32)

	buildHuffmanTable(table, 0, 5, codeLengthCodeLengths, codeLengthCodes)

	for symbol < numSymbols && space > 0 {
		readMoreInput(s)
		fillBitWindow(s)
		p := peekBits(s) & 31
		s.bitOffset += int(table[p] >> 16)
		codeLen := int(table[p] & 0xFFFF)
		if codeLen < codeLengthRepeatCode {
			repeat = 0
			codeLengths[symbol] = codeLen
			symbol++
			if codeLen != 0 {
				prevCodeLen = codeLen
				space -= 32768 >> codeLen
			}
		} else {
			extraBits := codeLen - 14
			newLen := 0
			if codeLen == codeLengthRepeatCode {
				newLen = prevCodeLen
			}
			if repeatCodeLen != newLen {
				repeat = 0
				repeatCodeLen = newLen
			}
			oldRepeat := repeat
			if repeat > 0 {
				repeat -= 2
				repeat <<= extraBits
			}
			fillBitWindow(s)
			repeat += readFewBits(s, extraBits) + 3
			repeatDelta := repeat - oldRepeat
			if symbol+repeatDelta > numSymbols {
				throw("symbol + repeatDelta > numSymbols") // COV_NF_LINE
			}
			for range repeatDelta {
				codeLengths[symbol] = repeatCodeLen
				symbol++
			}
			if repeatCodeLen != 0 {
				space -= repeatDelta << (15 - repeatCodeLen)
			}
		}
	}
	if space != 0 {
		throw("Unused space") // COV_NF_LINE
	}
	// TODO: Pass max_symbol to Huffman table builder instead?
	fillIntsWithZeroes(codeLengths, symbol, numSymbols)
}

func checkDupes(symbols []int, length int) int {
	for i := 0; i < length-1; i++ {
		for j := i + 1; j < length; j++ {
			if symbols[i] == symbols[j] {
				return 0
			}
		}
	}
	return 1
}

// readHuffmanCode reads a Huffman code of alphabetSize symbols and builds
// its table at table[offset:].
// TODO: Use specialized versions for smaller tables.
func readHuffmanCode(alphabetSize int, table []int32, offset int, s *state) {
	ok := 1
	var simpleCodeOrSkip int
	readMoreInput(s)
	// TODO: Avoid allocation.
	codeLengths := make([]int, alphabetSize)
	fillBitWindow(s)
	simpleCodeOrSkip = readFewBits(s, 2)
	if simpleCodeOrSkip == 1 { // Read symbols, codes & code lengths directly.
		maxBitsCounter := alphabetSize - 1
		maxBits := 0
		var symbols [4]int
		numSymbols := readFewBits(s, 2) + 1
		for maxBitsCounter != 0 {
			maxBitsCounter >>= 1
			maxBits++
		}
		// TODO: uncomment when codeLengths is reused.
		// fillWithZeroes(codeLengths, 0, alphabetSize)
		for i := range numSymbols {
			fillBitWindow(s)
			symbols[i] = readFewBits(s, maxBits) % alphabetSize
			codeLengths[symbols[i]] = 2
		}
		codeLengths[symbols[0]] = 1
		switch numSymbols {
		case 2:
			codeLengths[symbols[1]] = 1
		case 4:
			if readFewBits(s, 1) == 1 {
				codeLengths[symbols[2]] = 3
				codeLengths[symbols[3]] = 3
			} else {
				codeLengths[symbols[0]] = 2
			}
		}
		ok = checkDupes(symbols[:], numSymbols)
	} else { // Decode Huffman-coded code lengths.
		var codeLengthCodeLengths [codeLengthCodes]int
		space := 32
		numCodes := 0
		for i := simpleCodeOrSkip; i < codeLengthCodes && space > 0; i++ {
			codeLenIdx := codeLengthCodeOrder[i]
			fillBitWindow(s)
			p := peekBits(s) & 15
			// TODO: Demultiplex FIXED_TABLE.
			s.bitOffset += fixedTable[p] >> 16
			v := fixedTable[p] & 0xFFFF
			codeLengthCodeLengths[codeLenIdx] = v
			if v != 0 {
				space -= 32 >> v
				numCodes++
			}
		}
		if space != 0 && numCodes != 1 {
			ok = 0
		}
		readHuffmanCodeLengths(codeLengthCodeLengths[:], alphabetSize, codeLengths, s)
	}
	if ok == 0 {
		throw("Can't readHuffmanCode") // COV_NF_LINE
	}
	buildHuffmanTable(table, offset, huffmanTableBits, codeLengths, alphabetSize)
}

func decodeContextMap(contextMapSize int, contextMap []byte, s *state) int {
	readMoreInput(s)
	numTrees := decodeVarLenUnsignedByte(s) + 1

	if numTrees == 1 {
		fillBytesWithZeroes(contextMap, 0, contextMapSize)
		return numTrees
	}

	fillBitWindow(s)
	useRleForZeros := readFewBits(s, 1)
	maxRunLengthPrefix := 0
	if useRleForZeros != 0 {
		maxRunLengthPrefix = readFewBits(s, 4) + 1
	}
	table := make([]int32, huffmanTableSize)
	readHuffmanCode(numTrees+maxRunLengthPrefix, table, 0, s)
	for i := 0; i < contextMapSize; {
		readMoreInput(s)
		fillBitWindow(s)
		code := readSymbol(table, 0, s)
		if code == 0 {
			contextMap[i] = 0
			i++
		} else if code <= maxRunLengthPrefix {
			fillBitWindow(s)
			reps := (1 << code) + readFewBits(s, code)
			for reps != 0 {
				if i >= contextMapSize {
					throw("Corrupted context map") // COV_NF_LINE
				}
				contextMap[i] = 0
				i++
				reps--
			}
		} else {
			contextMap[i] = byte(code - maxRunLengthPrefix)
			i++
		}
	}
	fillBitWindow(s)
	if readFewBits(s, 1) == 1 {
		inverseMoveToFrontTransform(contextMap, contextMapSize)
	}
	return numTrees
}

func decodeBlockTypeAndLength(s *state, treeType, numBlockTypes int) int {
	ringBuffers := s.rings
	offset := 4 + treeType*2
	fillBitWindow(s)
	blockType := readSymbol(s.blockTrees, treeType*huffmanTableSize, s)
	result := readBlockLength(s.blockTrees, (treeType+3)*huffmanTableSize, s)

	if blockType == 1 {
		blockType = ringBuffers[offset+1] + 1
	} else if blockType == 0 {
		blockType = ringBuffers[offset]
	} else {
		blockType -= 2
	}
	if blockType >= numBlockTypes {
		blockType -= numBlockTypes
	}
	ringBuffers[offset] = ringBuffers[offset+1]
	ringBuffers[offset+1] = blockType
	return result
}

func decodeLiteralBlockSwitch(s *state) {
	s.literalBlockLength = decodeBlockTypeAndLength(s, 0, s.numLiteralBlockTypes)
	literalBlockType := s.rings[5]
	s.contextMapSlice = literalBlockType << literalContextBits
	s.literalTreeIndex = int(s.contextMap[s.contextMapSlice])
	s.literalTree = int(s.hGroup0[s.literalTreeIndex])
	contextMode := int(s.contextModes[literalBlockType])
	s.contextLookupOffset1 = contextMode << 9
	s.contextLookupOffset2 = s.contextLookupOffset1 + 256
}

func decodeCommandBlockSwitch(s *state) {
	s.commandBlockLength = decodeBlockTypeAndLength(s, 1, s.numCommandBlockTypes)
	s.treeCommandOffset = int(s.hGroup1[s.rings[7]])
}

func decodeDistanceBlockSwitch(s *state) {
	s.distanceBlockLength = decodeBlockTypeAndLength(s, 2, s.numDistanceBlockTypes)
	s.distContextMapSlice = s.rings[9] << distanceContextBits
}

func maybeReallocateRingBuffer(s *state) {
	newSize := s.maxRingBufferSize
	if newSize > s.expectedTotalSize {
		// TODO: Handle 2GB+ cases more gracefully.
		minimalNewSize := s.expectedTotalSize
		for (newSize >> 1) > minimalNewSize {
			newSize >>= 1
		}
		if s.inputEnd == 0 && newSize < 16384 && s.maxRingBufferSize >= 16384 {
			newSize = 16384
		}
	}
	if newSize <= s.ringBufferSize {
		return
	}
	ringBufferSizeWithSlack := newSize + maxTransformedWordLength
	newBuffer := make([]byte, ringBufferSizeWithSlack)
	if len(s.ringBuffer) != 0 {
		copy(newBuffer, s.ringBuffer[:s.ringBufferSize])
	}
	s.ringBuffer = newBuffer
	s.ringBufferSize = newSize
}

func readNextMetablockHeader(s *state) {
	if s.inputEnd != 0 {
		s.nextRunningState = finished
		s.bytesToWrite = s.pos
		s.bytesWritten = 0
		s.runningState = write
		return
	}
	// TODO: Reset? Do we need this?
	s.hGroup0 = []int32{}
	s.hGroup1 = []int32{}
	s.hGroup2 = []int32{}

	readMoreInput(s)
	decodeMetaBlockLength(s)
	if s.metaBlockLength == 0 && s.isMetadata == 0 {
		return
	}
	if s.isUncompressed != 0 || s.isMetadata != 0 {
		jumpToByteBoundary(s)
		if s.isMetadata != 0 {
			s.runningState = readMetadata
		} else {
			s.runningState = copyUncompressed
		}
	} else {
		s.runningState = compressedBlockStart
	}

	if s.isMetadata != 0 {
		return
	}
	s.expectedTotalSize += s.metaBlockLength
	if s.expectedTotalSize > 1<<30 {
		s.expectedTotalSize = 1 << 30
	}
	if s.ringBufferSize < s.maxRingBufferSize {
		maybeReallocateRingBuffer(s)
	}
}

func readMetablockPartition(s *state, treeType, numBlockTypes int) int {
	if numBlockTypes <= 1 {
		return 1 << 28
	}
	readHuffmanCode(numBlockTypes+2, s.blockTrees, treeType*huffmanTableSize, s)
	readHuffmanCode(numBlockLengthCodes, s.blockTrees, (treeType+3)*huffmanTableSize, s)
	return readBlockLength(s.blockTrees, (treeType+3)*huffmanTableSize, s)
}

func readMetablockHuffmanCodesAndContextMaps(s *state) {
	s.numLiteralBlockTypes = decodeVarLenUnsignedByte(s) + 1
	s.literalBlockLength = readMetablockPartition(s, 0, s.numLiteralBlockTypes)
	s.numCommandBlockTypes = decodeVarLenUnsignedByte(s) + 1
	s.commandBlockLength = readMetablockPartition(s, 1, s.numCommandBlockTypes)
	s.numDistanceBlockTypes = decodeVarLenUnsignedByte(s) + 1
	s.distanceBlockLength = readMetablockPartition(s, 2, s.numDistanceBlockTypes)

	readMoreInput(s)
	fillBitWindow(s)
	s.distancePostfixBits = readFewBits(s, 2)
	s.numDirectDistanceCodes = numDistanceShortCodes + (readFewBits(s, 4) << s.distancePostfixBits)
	s.distancePostfixMask = (1 << s.distancePostfixBits) - 1
	numDistanceCodes := s.numDirectDistanceCodes + (48 << s.distancePostfixBits)
	// TODO: Reuse?
	s.contextModes = make([]byte, s.numLiteralBlockTypes)
	for i := 0; i < s.numLiteralBlockTypes; {
		// Ensure that less than 256 bits read between readMoreInput.
		limit := min(i+96, s.numLiteralBlockTypes)
		for ; i < limit; i++ {
			fillBitWindow(s)
			s.contextModes[i] = byte(readFewBits(s, 2))
		}
		readMoreInput(s)
	}

	// TODO: Reuse?
	s.contextMap = make([]byte, s.numLiteralBlockTypes<<literalContextBits)
	numLiteralTrees := decodeContextMap(s.numLiteralBlockTypes<<literalContextBits, s.contextMap, s)
	s.trivialLiteralContext = 1
	for j := range s.numLiteralBlockTypes << literalContextBits {
		// Java compares the signed byte, so that block types from 128 take
		// the general path of INSERT_LOOP: both paths decode the same.
		if int(s.contextMap[j]) != j>>literalContextBits {
			s.trivialLiteralContext = 0
			break
		}
	}

	// TODO: Reuse?
	s.distContextMap = make([]byte, s.numDistanceBlockTypes<<distanceContextBits)
	numDistTrees := decodeContextMap(s.numDistanceBlockTypes<<distanceContextBits, s.distContextMap, s)

	s.hGroup0 = decodeHuffmanTreeGroup(numLiteralCodes, numLiteralTrees, s)
	s.hGroup1 = decodeHuffmanTreeGroup(numInsertAndCopyCodes, s.numCommandBlockTypes, s)
	s.hGroup2 = decodeHuffmanTreeGroup(numDistanceCodes, numDistTrees, s)

	s.contextMapSlice = 0
	s.distContextMapSlice = 0
	s.contextLookupOffset1 = int(s.contextModes[0]) << 9
	s.contextLookupOffset2 = s.contextLookupOffset1 + 256
	s.literalTreeIndex = 0
	s.literalTree = int(s.hGroup0[0])
	s.treeCommandOffset = int(s.hGroup1[0])

	s.rings[4] = 1
	s.rings[5] = 0
	s.rings[6] = 1
	s.rings[7] = 0
	s.rings[8] = 1
	s.rings[9] = 0
}

func copyUncompressedData(s *state) {
	ringBuffer := s.ringBuffer

	// Could happen if block ends at ring buffer end.
	if s.metaBlockLength <= 0 {
		reload(s)
		s.runningState = blockStart
		return
	}

	chunkLength := min(s.ringBufferSize-s.pos, s.metaBlockLength)
	copyBytes(s, ringBuffer, s.pos, chunkLength)
	s.metaBlockLength -= chunkLength
	s.pos += chunkLength
	if s.pos == s.ringBufferSize {
		s.nextRunningState = copyUncompressed
		s.bytesToWrite = s.ringBufferSize
		s.bytesWritten = 0
		s.runningState = write
		return
	}

	reload(s)
	s.runningState = blockStart
}

func writeRingBuffer(s *state) int {
	toWrite := min(s.outputLength-s.outputUsed, s.bytesToWrite-s.bytesWritten)
	if toWrite != 0 {
		copy(s.output[s.outputOffset+s.outputUsed:], s.ringBuffer[s.bytesWritten:s.bytesWritten+toWrite])
		s.outputUsed += toWrite
		s.bytesWritten += toWrite
	}

	if s.outputUsed < s.outputLength {
		return 1
	}
	return 0
}

func decodeHuffmanTreeGroup(alphabetSize, n int, s *state) []int32 {
	group := make([]int32, n+(n*huffmanTableSize))
	next := n
	for i := range n {
		group[i] = int32(next)
		readHuffmanCode(alphabetSize, group, next, s)
		next += huffmanTableSize
	}
	return group
}

// decompress is the actual decompress implementation: it decodes into
// s.output until the output is full or the stream ends. Upstream's
// IllegalStateException is a brotliRuntimeException here.
func decompress(s *state) {
	if s.runningState == uninitialized {
		throw("Can't decompress until initialized")
	}
	if s.runningState == closed {
		throw("Can't decompress after close")
	}
	ringBufferMask := s.ringBufferSize - 1
	ringBuffer := s.ringBuffer

	for s.runningState != finished {
		// TODO: extract cases to methods for the better readability.
		switch s.runningState {
		case blockStart:
			if s.metaBlockLength < 0 {
				throw("Invalid metablock length")
			}
			readNextMetablockHeader(s)
			// Ring-buffer would be reallocated here.
			ringBufferMask = s.ringBufferSize - 1
			ringBuffer = s.ringBuffer
			continue

		case compressedBlockStart:
			readMetablockHuffmanCodesAndContextMaps(s)
			s.runningState = mainLoop
			fallthrough

		case mainLoop:
			if s.metaBlockLength <= 0 {
				s.runningState = blockStart
				continue
			}
			readMoreInput(s)
			if s.commandBlockLength == 0 {
				decodeCommandBlockSwitch(s)
			}
			s.commandBlockLength--
			fillBitWindow(s)
			cmdCode := readSymbol(s.hGroup1, s.treeCommandOffset, s)
			rangeIdx := cmdCode >> 6
			s.distanceCode = 0
			if rangeIdx >= 2 {
				rangeIdx -= 2
				s.distanceCode = -1
			}
			insertCode := insertRangeLut[rangeIdx] + ((cmdCode >> 3) & 7)
			fillBitWindow(s)
			insertBits := insertLengthNBits[insertCode]
			insertExtra := readBits(s, insertBits)
			s.insertLength = insertLengthOffset[insertCode] + insertExtra
			copyCode := copyRangeLut[rangeIdx] + (cmdCode & 7)
			fillBitWindow(s)
			copyBits := copyLengthNBits[copyCode]
			copyExtra := readBits(s, copyBits)
			s.copyLength = copyLengthOffset[copyCode] + copyExtra

			s.j = 0
			s.runningState = insertLoop
			fallthrough

		case insertLoop:
			if s.trivialLiteralContext != 0 {
				for s.j < s.insertLength {
					readMoreInput(s)
					if s.literalBlockLength == 0 {
						decodeLiteralBlockSwitch(s)
					}
					s.literalBlockLength--
					fillBitWindow(s)
					ringBuffer[s.pos] = byte(readSymbol(s.hGroup0, s.literalTree, s))
					s.j++
					s.pos++
					if s.pos-1 == ringBufferMask {
						s.nextRunningState = insertLoop
						s.bytesToWrite = s.ringBufferSize
						s.bytesWritten = 0
						s.runningState = write
						break
					}
				}
			} else {
				prevByte1 := int(ringBuffer[(s.pos-1)&ringBufferMask])
				prevByte2 := int(ringBuffer[(s.pos-2)&ringBufferMask])
				for s.j < s.insertLength {
					readMoreInput(s)
					if s.literalBlockLength == 0 {
						decodeLiteralBlockSwitch(s)
					}
					literalTreeIndex := int(s.contextMap[s.contextMapSlice+
						(contextLookup[s.contextLookupOffset1+prevByte1]|
							contextLookup[s.contextLookupOffset2+prevByte2])])
					s.literalBlockLength--
					prevByte2 = prevByte1
					fillBitWindow(s)
					prevByte1 = readSymbol(s.hGroup0, int(s.hGroup0[literalTreeIndex]), s)
					ringBuffer[s.pos] = byte(prevByte1)
					s.j++
					s.pos++
					if s.pos-1 == ringBufferMask {
						s.nextRunningState = insertLoop
						s.bytesToWrite = s.ringBufferSize
						s.bytesWritten = 0
						s.runningState = write
						break
					}
				}
			}
			if s.runningState != insertLoop {
				continue
			}
			s.metaBlockLength -= s.insertLength
			if s.metaBlockLength <= 0 {
				s.runningState = mainLoop
				continue
			}
			if s.distanceCode < 0 {
				readMoreInput(s)
				if s.distanceBlockLength == 0 {
					decodeDistanceBlockSwitch(s)
				}
				s.distanceBlockLength--
				fillBitWindow(s)
				distTreeIndex := 3
				if s.copyLength <= 4 {
					distTreeIndex = s.copyLength - 2
				}
				s.distanceCode = readSymbol(s.hGroup2,
					int(s.hGroup2[s.distContextMap[s.distContextMapSlice+distTreeIndex]]), s)
				if s.distanceCode >= s.numDirectDistanceCodes {
					s.distanceCode -= s.numDirectDistanceCodes
					postfix := s.distanceCode & s.distancePostfixMask
					s.distanceCode >>= s.distancePostfixBits
					n := (s.distanceCode >> 1) + 1
					offset := ((2 + (s.distanceCode & 1)) << n) - 4
					fillBitWindow(s)
					distanceExtra := readBits(s, n)
					s.distanceCode = s.numDirectDistanceCodes + postfix +
						((offset + distanceExtra) << s.distancePostfixBits)
				}
			}

			// Convert the distance code to the actual distance by possibly
			// looking up past distances from the ringBuffer.
			s.distance = translateShortCodes(s.distanceCode, s.rings, s.distRbIdx)
			if s.distance < 0 {
				throw("Negative distance") // COV_NF_LINE
			}

			if s.maxDistance != s.maxBackwardDistance && s.pos < s.maxBackwardDistance {
				s.maxDistance = s.pos
			} else {
				s.maxDistance = s.maxBackwardDistance
			}

			s.copyDst = s.pos
			if s.distance > s.maxDistance {
				s.runningState = transform
				continue
			}

			if s.distanceCode > 0 {
				s.rings[s.distRbIdx&3] = s.distance
				s.distRbIdx++
			}

			if s.copyLength > s.metaBlockLength {
				throw("Invalid backward reference") // COV_NF_LINE
			}
			s.j = 0
			s.runningState = copyLoop
			fallthrough

		case copyLoop:
			src := (s.pos - s.distance) & ringBufferMask
			dst := s.pos
			copyLength := s.copyLength - s.j
			srcEnd := src + copyLength
			dstEnd := dst + copyLength
			if srcEnd < ringBufferMask && dstEnd < ringBufferMask {
				if copyLength < 12 || (srcEnd > dst && dstEnd > src) {
					for range copyLength {
						ringBuffer[dst] = ringBuffer[src]
						dst++
						src++
					}
				} else {
					copyBytesWithin(ringBuffer, dst, src, srcEnd)
				}
				s.j += copyLength
				s.metaBlockLength -= copyLength
				s.pos += copyLength
			} else {
				for s.j < s.copyLength {
					ringBuffer[s.pos] = ringBuffer[(s.pos-s.distance)&ringBufferMask]
					s.metaBlockLength--
					s.j++
					s.pos++
					if s.pos-1 == ringBufferMask {
						s.nextRunningState = copyLoop
						s.bytesToWrite = s.ringBufferSize
						s.bytesWritten = 0
						s.runningState = write
						break
					}
				}
			}
			if s.runningState == copyLoop {
				s.runningState = mainLoop
			}
			continue

		case transform:
			if s.copyLength >= minWordLength && s.copyLength <= maxWordLength {
				offset := dictionaryOffsetsByLength[s.copyLength]
				wordID := s.distance - s.maxDistance - 1
				shift := dictionarySizeBitsByLength[s.copyLength]
				mask := (1 << shift) - 1
				wordIdx := wordID & mask
				transformIdx := wordID >> shift
				offset += wordIdx * s.copyLength
				if transformIdx < numTransforms {
					length := transformDictionaryWord(ringBuffer, s.copyDst, getDictionaryData(), offset,
						s.copyLength, transformIdx)
					s.copyDst += length
					s.pos += length
					s.metaBlockLength -= length
					if s.copyDst >= s.ringBufferSize {
						s.nextRunningState = copyWrapBuffer
						s.bytesToWrite = s.ringBufferSize
						s.bytesWritten = 0
						s.runningState = write
						continue
					}
				} else {
					throw("Invalid backward reference") // COV_NF_LINE
				}
			} else {
				throw("Invalid backward reference") // COV_NF_LINE
			}
			s.runningState = mainLoop
			continue

		case copyWrapBuffer:
			copyBytesWithin(ringBuffer, 0, s.ringBufferSize, s.copyDst)
			s.runningState = mainLoop
			continue

		case readMetadata:
			for s.metaBlockLength > 0 {
				readMoreInput(s)
				// Optimize
				fillBitWindow(s)
				readFewBits(s, 8)
				s.metaBlockLength--
			}
			s.runningState = blockStart
			continue

		case copyUncompressed:
			copyUncompressedData(s)
			continue

		case write:
			if writeRingBuffer(s) == 0 {
				// Output buffer is full.
				return
			}
			if s.pos >= s.maxBackwardDistance {
				s.maxDistance = s.maxBackwardDistance
			}
			s.pos &= ringBufferMask
			s.runningState = s.nextRunningState
			continue

		default:
			throw("Unexpected state " + strconv.Itoa(s.runningState))
		}
	}
	if s.runningState == finished {
		if s.metaBlockLength < 0 {
			throw("Invalid metablock length")
		}
		jumpToByteBoundary(s)
		checkHealth(s, 1)
	}
}
