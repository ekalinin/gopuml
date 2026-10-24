// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/font/UnusedSpace.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dejavu

import (
	_ "embed"
	"encoding/binary"
	"sync"

	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// Upstream's UnusedSpace centers a character by rasterizing it with Java2D
// (hinted, not antialiased) and searching the point that minimizes the
// distance to the farthest pixel. Hinted rasterization is not reproducible
// here, so the results come from a table dumped from the reference JVM
// (the same table as internal/klimt/svg/glyphs/unusedspace.bin, written by
// internal/klimt/svg/testdata/javagen/SvgGolden.java): the printable ASCII
// characters of the 12 embedded faces at sizes 4 to 32.

//go:embed unusedspace.bin
var unusedSpaceTable []byte

type unusedSpaceData struct {
	minSize, maxSize int
	first            rune
	count            int
	centers          []byte
}

var (
	unusedSpaceOnce sync.Once
	unusedSpace     unusedSpaceData
)

func loadUnusedSpace() *unusedSpaceData {
	unusedSpaceOnce.Do(func() {
		c := unusedSpaceTable
		if len(c) < 15 || string(c[:8]) != "PUUNUSE1" || int(c[8]) != 12 {
			panic("dejavu: bad unusedspace.bin")
		}
		unusedSpace = unusedSpaceData{minSize: int(c[9]), maxSize: int(c[10]),
			first: rune(binary.BigEndian.Uint16(c[11:])), count: int(binary.BigEndian.Uint16(c[13:])),
			centers: c[15:]}
	})
	return &unusedSpace
}

// UnusedSpaceCenter returns the center of upstream's UnusedSpace of a
// character drawn with f (its getCenterX and getCenterY); ok is false
// outside of the table.
func UnusedSpaceCenter(f font.UFont, c rune) (centerX, centerY float64, ok bool) {
	awt := f.UnderlyingFont(string(c))
	face := int(FamilyOf(awt))*4 + int(awt.Style()&font.AwtBoldItalic)
	t := loadUnusedSpace()
	size := awt.Size()
	idx := int(c - t.first)
	if size < t.minSize || size > t.maxSize || idx < 0 || idx >= t.count {
		return 0, 0, false
	}
	at := 2 * ((face*(t.maxSize-t.minSize+1)+size-t.minSize)*t.count + idx)
	if at+1 >= len(t.centers) {
		return 0, 0, false
	}
	return float64(int8(t.centers[at])) / 4, float64(int8(t.centers[at+1])) / 4, true
}
