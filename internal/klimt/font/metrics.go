// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"path"
	"slices"
	"strings"
	"sync"
)

// The tables are dumped from the reference JVM by
// tools/reference/FontMetricsDump.java, one file per Java logical family and
// style, named "<family>-<style>.bin".
//
//go:embed metrics/*.bin
var metricsFS embed.FS

// metricsMagic and metricsVersion identify the binary table format.
const (
	metricsMagic   = "PUMF"
	metricsVersion = 1
)

// Metrics is the text metrics table of one Java logical font family and
// style, as Java2D measures it with fractional metrics. Every value is
// measured at RefSize and scales linearly with the font size.
type Metrics struct {
	// Family is the Java logical family the table was measured on, for
	// example "SansSerif".
	Family string
	// Style is "plain", "bold", "italic" or "bolditalic".
	Style string
	// RefSize is the font size the table was measured at.
	RefSize float32
	// Ascent, Descent and Leading are Java's LineMetrics values.
	Ascent, Descent, Leading float32
	// BoundsHeight is the height of FontMetrics.getStringBounds, which does
	// not depend on the text.
	BoundsHeight float32
	// MissingAdvance is the advance of a code point that no font can display.
	MissingAdvance float32

	ranges []advanceRange
}

// advanceRange holds the advances of consecutive displayable code points.
type advanceRange struct {
	first    rune
	advances []float32
}

func (r advanceRange) end() rune { return r.first + rune(len(r.advances)) }

// Advance returns the advance of cp at RefSize and whether the font can
// display cp, as Java's Font.canDisplay. Code points the font cannot display
// get MissingAdvance.
func (m *Metrics) Advance(cp rune) (float32, bool) {
	i, found := slices.BinarySearchFunc(m.ranges, cp, func(r advanceRange, cp rune) int {
		switch {
		case cp < r.first:
			return 1
		case cp >= r.end():
			return -1
		}
		return 0
	})
	if !found {
		return m.MissingAdvance, false
	}
	r := m.ranges[i]
	return r.advances[cp-r.first], true
}

// CanDisplay reports whether the font can display cp, as Java's
// Font.canDisplay.
func (m *Metrics) CanDisplay(cp rune) bool {
	_, ok := m.Advance(cp)
	return ok
}

// Glyphs returns the number of code points the font can display.
func (m *Metrics) Glyphs() int {
	n := 0
	for _, r := range m.ranges {
		n += len(r.advances)
	}
	return n
}

// Ranges returns the number of ranges of consecutive displayable code points.
func (m *Metrics) Ranges() int { return len(m.ranges) }

// metricsReader decodes the big-endian values written by Java's
// DataOutputStream. The first error sticks.
type metricsReader struct {
	data []byte
	err  error
}

func (r *metricsReader) next(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || len(r.data) < n {
		r.err = errors.New("unexpected end of data")
		return nil
	}
	b := r.data[:n]
	r.data = r.data[n:]
	return b
}

func (r *metricsReader) u16() uint16 {
	if b := r.next(2); b != nil {
		return binary.BigEndian.Uint16(b)
	}
	return 0
}

func (r *metricsReader) i32() int32 {
	if b := r.next(4); b != nil {
		return int32(binary.BigEndian.Uint32(b))
	}
	return 0
}

func (r *metricsReader) f32() float32 {
	if b := r.next(4); b != nil {
		return math.Float32frombits(binary.BigEndian.Uint32(b))
	}
	return 0
}

// utf reads a DataOutputStream.writeUTF string. The family and style names
// are ASCII, where Java's modified UTF-8 is plain UTF-8.
func (r *metricsReader) utf() string {
	n := r.u16()
	return string(r.next(int(n)))
}

// parseMetrics decodes one table:
//
//	magic "PUMF", u16 version (1), UTF family, UTF style, f32 refSize,
//	f32 ascent, f32 descent, f32 leading, f32 boundsHeight, f32 missingAdvance,
//	i32 rangeCount, then per range: i32 firstCodePoint, i32 count, count x f32.
func parseMetrics(data []byte) (*Metrics, error) {
	r := &metricsReader{data: data}
	if magic := string(r.next(len(metricsMagic))); r.err == nil && magic != metricsMagic {
		return nil, fmt.Errorf("bad magic %q", magic)
	}
	if v := r.u16(); r.err == nil && v != metricsVersion {
		return nil, fmt.Errorf("unsupported version %d", v)
	}
	m := &Metrics{Family: r.utf(), Style: r.utf()}
	m.RefSize = r.f32()
	m.Ascent = r.f32()
	m.Descent = r.f32()
	m.Leading = r.f32()
	m.BoundsHeight = r.f32()
	m.MissingAdvance = r.f32()
	count := r.i32()
	if r.err != nil {
		return nil, r.err
	}
	if count < 0 || int(count) > len(r.data)/8 {
		return nil, fmt.Errorf("bad range count %d", count)
	}
	if !(m.RefSize > 0) {
		return nil, fmt.Errorf("bad reference size %v", m.RefSize)
	}
	m.ranges = make([]advanceRange, 0, count)
	var prevEnd rune
	for i := range int(count) {
		first, n := r.i32(), r.i32()
		if r.err != nil {
			return nil, r.err
		}
		if first < prevEnd || n <= 0 || int64(first)+int64(n) > 0x110000 || int(n) > len(r.data)/4 {
			return nil, fmt.Errorf("range %d: bad code points %d+%d", i, first, n)
		}
		advances := make([]float32, n)
		for k := range advances {
			advances[k] = r.f32()
		}
		m.ranges = append(m.ranges, advanceRange{first: first, advances: advances})
		prevEnd = first + n
	}
	if r.err != nil {
		return nil, r.err
	}
	if len(r.data) != 0 {
		return nil, fmt.Errorf("%d trailing bytes", len(r.data))
	}
	return m, nil
}

// loadMetrics parses every embedded table, keyed by file name without the
// extension, for example "sansserif-plain".
func loadMetrics() (map[string]*Metrics, error) {
	names, err := fs.Glob(metricsFS, "metrics/*.bin")
	if err != nil {
		return nil, err
	}
	tables := make(map[string]*Metrics, len(names))
	for _, name := range names {
		data, err := metricsFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		m, err := parseMetrics(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		key := strings.TrimSuffix(path.Base(name), ".bin")
		if want := strings.ToLower(m.Family) + "-" + m.Style; key != want {
			return nil, fmt.Errorf("%s: holds %s %s", name, m.Family, m.Style)
		}
		tables[key] = m
	}
	return tables, nil
}

var embeddedMetrics = sync.OnceValues(loadMetrics)

// lookupMetrics returns the embedded table of a logical family ("sansserif",
// "serif", "monospaced" or "dialog") in a style. The tables are fixed data
// checked by the tests, so a missing or corrupt table is a build defect and
// panics.
func lookupMetrics(table string, style AwtStyle) *Metrics {
	tables, err := embeddedMetrics()
	if err != nil {
		panic("font: embedded metrics: " + err.Error())
	}
	m, ok := tables[table+"-"+style.String()]
	if !ok {
		panic("font: no embedded metrics for " + table + " " + style.String())
	}
	return m
}
