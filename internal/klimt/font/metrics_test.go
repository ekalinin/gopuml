// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"slices"
	"strings"
	"testing"
)

type summaryTable struct {
	Family           string  `json:"family"`
	Style            string  `json:"style"`
	Glyphs           int     `json:"glyphs"`
	Ranges           int     `json:"ranges"`
	Ascent           float64 `json:"ascent"`
	Descent          float64 `json:"descent"`
	Leading          float64 `json:"leading"`
	BoundsHeight     float64 `json:"boundsHeight"`
	MissingCodePoint rune    `json:"missingCodePoint"`
	MissingAdvance   float64 `json:"missingAdvance"`
}

func readSummary(t *testing.T) (refSize float64, dialogEqualsSansSerif bool, tables []summaryTable) {
	t.Helper()
	data, err := os.ReadFile("metrics/summary.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		RefSize               float64        `json:"refSize"`
		DialogEqualsSansSerif bool           `json:"dialogEqualsSansSerif"`
		Tables                []summaryTable `json:"tables"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc.RefSize, doc.DialogEqualsSansSerif, doc.Tables
}

func TestEmbeddedMetricsMatchSummary(t *testing.T) {
	tables, err := loadMetrics()
	if err != nil {
		t.Fatal(err)
	}
	refSize, dialogEqualsSansSerif, summary := readSummary(t)
	if len(tables) != 16 || len(summary) != 16 {
		t.Fatalf("got %d tables and %d summary entries, want 16", len(tables), len(summary))
	}
	for _, want := range summary {
		key := strings.ToLower(want.Family) + "-" + want.Style
		m := tables[key]
		if m == nil {
			t.Errorf("%s: missing table", key)
			continue
		}
		if m.Family != want.Family || m.Style != want.Style {
			t.Errorf("%s: holds %s %s", key, m.Family, m.Style)
		}
		checks := []struct {
			name      string
			got, want float64
		}{
			{"refSize", float64(m.RefSize), refSize},
			{"glyphs", float64(m.Glyphs()), float64(want.Glyphs)},
			{"ranges", float64(m.Ranges()), float64(want.Ranges)},
			{"ascent", float64(m.Ascent), want.Ascent},
			{"descent", float64(m.Descent), want.Descent},
			{"leading", float64(m.Leading), want.Leading},
			{"boundsHeight", float64(m.BoundsHeight), want.BoundsHeight},
			{"missingAdvance", float64(m.MissingAdvance), want.MissingAdvance},
		}
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("%s: %s = %v, want %v", key, c.name, c.got, c.want)
			}
		}
		if adv, ok := m.Advance(want.MissingCodePoint); ok || adv != m.MissingAdvance {
			t.Errorf("%s: missing code point U+%04X: advance %v, displayable %v", key, want.MissingCodePoint, adv, ok)
		}
		if adv, ok := m.Advance('A'); !ok || adv <= 0 {
			t.Errorf("%s: 'A': advance %v, displayable %v", key, adv, ok)
		}
	}
	if !dialogEqualsSansSerif {
		t.Fatal("summary says Dialog differs from SansSerif")
	}
	for _, style := range []string{"plain", "bold", "italic", "bolditalic"} {
		d, s := tables["dialog-"+style], tables["sansserif-"+style]
		equal := slices.EqualFunc(d.ranges, s.ranges, func(a, b advanceRange) bool {
			return a.first == b.first && slices.Equal(a.advances, b.advances)
		})
		if !equal || d.MissingAdvance != s.MissingAdvance || d.BoundsHeight != s.BoundsHeight {
			t.Errorf("dialog-%s differs from sansserif-%s", style, style)
		}
	}
}

func TestMetricsAdvanceLookup(t *testing.T) {
	m := lookupMetrics("sansserif", AwtPlain)
	// Every code point inside a range is displayable, every code point
	// between ranges is not.
	for i, r := range m.ranges {
		for k, want := range r.advances {
			if got, ok := m.Advance(r.first + rune(k)); !ok || got != want {
				t.Fatalf("U+%04X: got %v %v, want %v", r.first+rune(k), got, ok, want)
			}
		}
		if i+1 < len(m.ranges) && r.end() < m.ranges[i+1].first {
			if got, ok := m.Advance(r.end()); ok || got != m.MissingAdvance {
				t.Fatalf("U+%04X: got %v %v, want the missing advance", r.end(), got, ok)
			}
		}
	}
	for _, cp := range []rune{-1, 0, 0x7f, 0xD800, 0x10FFFF, 0x110000} {
		if got, ok := m.Advance(cp); ok || got != m.MissingAdvance {
			t.Errorf("U+%04X: got %v %v, want the missing advance", cp, got, ok)
		}
	}
	// Tabs and line breaks are displayable with no advance, as in Java.
	for _, cp := range []rune{'\t', '\n', '\r'} {
		if got, ok := m.Advance(cp); !ok || got != 0 {
			t.Errorf("U+%04X: got %v %v, want 0 true", cp, got, ok)
		}
	}
}

// buildTable writes a table in the dump format.
func buildTable(family, style string, ranges map[int32][]float32, order []int32) []byte {
	var b bytes.Buffer
	b.WriteString("PUMF")
	w := func(v any) { _ = binary.Write(&b, binary.BigEndian, v) }
	w(uint16(1))
	for _, s := range []string{family, style} {
		w(uint16(len(s)))
		b.WriteString(s)
	}
	for _, v := range []float32{100, 90, 20, 0, 110, 60} {
		w(v)
	}
	w(int32(len(order)))
	for _, first := range order {
		w(first)
		w(int32(len(ranges[first])))
		w(ranges[first])
	}
	return b.Bytes()
}

func TestParseMetrics(t *testing.T) {
	ranges := map[int32][]float32{32: {30, 31}, 65: {70}}
	good := buildTable("SansSerif", "plain", ranges, []int32{32, 65})
	m, err := parseMetrics(good)
	if err != nil {
		t.Fatal(err)
	}
	if m.Family != "SansSerif" || m.Style != "plain" || m.RefSize != 100 || m.Ascent != 90 ||
		m.Descent != 20 || m.Leading != 0 || m.BoundsHeight != 110 || m.MissingAdvance != 60 {
		t.Errorf("header: got %+v", m)
	}
	for cp, want := range map[rune]float32{32: 30, 33: 31, 34: 60, 64: 60, 65: 70, 66: 60} {
		if got, _ := m.Advance(cp); got != want {
			t.Errorf("U+%04X: got %v, want %v", cp, got, want)
		}
	}
	if m.Glyphs() != 3 || m.Ranges() != 2 {
		t.Errorf("got %d glyphs in %d ranges, want 3 in 2", m.Glyphs(), m.Ranges())
	}

	bad := map[string][]byte{
		"empty":       nil,
		"magic":       append([]byte("PUMX"), good[4:]...),
		"version":     append(append([]byte("PUMF"), 0, 2), good[6:]...),
		"truncated":   good[:len(good)-1],
		"trailing":    append(slices.Clone(good), 0),
		"unsorted":    buildTable("SansSerif", "plain", ranges, []int32{65, 32}),
		"empty range": buildTable("SansSerif", "plain", map[int32][]float32{32: {}}, []int32{32}),
		"too far":     buildTable("SansSerif", "plain", map[int32][]float32{0x10FFFF: {1, 2}}, []int32{0x10FFFF}),
	}
	for name, data := range bad {
		if _, err := parseMetrics(data); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestLookupMetricsPanicsOnUnknownTable(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic")
		}
	}()
	lookupMetrics("cursive", AwtPlain)
}

func TestMetricsValuesAreFinite(t *testing.T) {
	tables, err := embeddedMetrics()
	if err != nil {
		t.Fatal(err)
	}
	for key, m := range tables {
		for _, r := range m.ranges {
			for k, adv := range r.advances {
				if math.IsNaN(float64(adv)) || math.IsInf(float64(adv), 0) || adv < 0 {
					t.Fatalf("%s: U+%04X: advance %v", key, r.first+rune(k), adv)
				}
			}
		}
	}
}
