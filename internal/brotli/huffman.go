// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/Huffman.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

// Utilities for building Huffman decoding tables.

const maxLength = 15

// getNextKey returns reverse(reverse(key, length) + 1, length), where
// reverse(key, length) is the bit-wise reversal of the length least significant
// bits of key.
func getNextKey(key, length int) int {
	step := 1 << (length - 1)
	for key&step != 0 {
		step >>= 1
	}
	return (key & (step - 1)) + step
}

// replicateValue stores item in table[0], table[step], table[2*step] ..,
// table[end] (from offset). It assumes that end is an integer multiple of
// step.
func replicateValue(table []int32, offset, step, end int, item int32) {
	for {
		end -= step
		table[offset+end] = item
		if end <= 0 {
			break
		}
	}
}

// nextTableBitSize returns the table width of the next 2nd level table, from
// count, the histogram of bit lengths for the remaining symbols, and length,
// the code length of the next processed symbol.
func nextTableBitSize(count []int, length, rootBits int) int {
	left := 1 << (length - rootBits)
	for length < maxLength {
		left -= count[length]
		if left <= 0 {
			break
		}
		length++
		left <<= 1
	}
	return length - rootBits
}

// buildHuffmanTable builds the Huffman lookup table at rootTable[tableOffset:]
// assuming code lengths are in symbol order.
func buildHuffmanTable(rootTable []int32, tableOffset, rootBits int, codeLengths []int, codeLengthsSize int) {
	var key int                            // Reversed prefix code.
	sorted := make([]int, codeLengthsSize) // Symbols sorted by code length.
	// TODO: fill with zeroes?
	var count [maxLength + 1]int  // Number of codes of each length.
	var offset [maxLength + 1]int // Offsets in sorted table for each length.
	var symbol int

	// Build histogram of code lengths.
	for symbol = 0; symbol < codeLengthsSize; symbol++ {
		count[codeLengths[symbol]]++
	}

	// Generate offsets into sorted symbol table by code length.
	offset[1] = 0
	for length := 1; length < maxLength; length++ {
		offset[length+1] = offset[length] + count[length]
	}

	// Sort symbols by length, by symbol order within each length.
	for symbol = 0; symbol < codeLengthsSize; symbol++ {
		if codeLengths[symbol] != 0 {
			sorted[offset[codeLengths[symbol]]] = symbol
			offset[codeLengths[symbol]]++
		}
	}

	tableBits := rootBits
	tableSize := 1 << tableBits
	totalSize := tableSize

	// Special case code with only one value.
	if offset[maxLength] == 1 {
		for key = 0; key < totalSize; key++ {
			rootTable[tableOffset+key] = int32(sorted[0])
		}
		return
	}

	// Fill in root table.
	key = 0
	symbol = 0
	for length, step := 1, 2; length <= rootBits; length, step = length+1, step<<1 {
		for ; count[length] > 0; count[length]-- {
			replicateValue(rootTable, tableOffset+key, step, tableSize, int32(length<<16|sorted[symbol]))
			symbol++
			key = getNextKey(key, length)
		}
	}

	// Fill in 2nd level tables and add pointers to root table.
	mask := totalSize - 1
	low := -1
	currentOffset := tableOffset
	for length, step := rootBits+1, 2; length <= maxLength; length, step = length+1, step<<1 {
		for ; count[length] > 0; count[length]-- {
			if key&mask != low {
				currentOffset += tableSize
				tableBits = nextTableBitSize(count[:], length, rootBits)
				tableSize = 1 << tableBits
				totalSize += tableSize
				low = key & mask
				rootTable[tableOffset+low] = int32((tableBits+rootBits)<<16 | (currentOffset - tableOffset - low))
			}
			replicateValue(rootTable, currentOffset+(key>>rootBits), step, tableSize,
				int32((length-rootBits)<<16|sorted[symbol]))
			symbol++
			key = getNextKey(key, length)
		}
	}
}
