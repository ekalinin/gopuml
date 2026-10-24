// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/Context.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

// contextLookup is upstream's Context.LOOKUP: the common context lookup
// table for all context modes.
var contextLookup [2048]int

// The packed tables are Java strings of characters below 0x100, here bytes:
// the \x81 of utfRle is Java's \u0081.
const (
	utfMap = "         !!  !                  \"#$##%#$&'##(#)#+++++++++" +
		"+((&*'##,---,---,-----,-----,-----&#'###.///.///./////./////./////&#'# "
	utfRle = "A/*  ':  & : $  \x81 @"
)

func unpackLookupTable(lookup []int, m, rle string) {
	// LSB6, MSB6, SIGNED
	for i := range 256 {
		lookup[i] = i & 0x3F
		lookup[512+i] = i >> 2
		lookup[1792+i] = 2 + (i >> 6)
	}
	// UTF8
	for i := range 128 {
		lookup[1024+i] = 4 * (int(m[i]) - 32)
	}
	for i := range 64 {
		lookup[1152+i] = i & 1
		lookup[1216+i] = 2 + (i & 1)
	}
	offset := 1280
	for k := range 19 {
		value := k & 3
		rep := int(rle[k]) - 32
		for range rep {
			lookup[offset] = value
			offset++
		}
	}
	// SIGNED
	for i := range 16 {
		lookup[1792+i] = 1
		lookup[2032+i] = 6
	}
	lookup[1792] = 0
	lookup[2047] = 7
	for i := range 256 {
		lookup[1536+i] = lookup[1792+i] << 3
	}
}

func init() {
	unpackLookupTable(contextLookup[:], utfMap, utfRle)
}
