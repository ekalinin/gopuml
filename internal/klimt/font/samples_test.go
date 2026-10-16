// SPDX-License-Identifier: MIT
// Provenance: original

package font

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// sample is a text measured directly by the reference JVM, at its size, with
// FontMetrics.getStringBounds and LineMetrics.getDescent.
type sample struct {
	Family  string  `json:"family"`
	Style   string  `json:"style"`
	Size    int     `json:"size"`
	Text    string  `json:"text"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Descent float64 `json:"descent"`
}

func readSamples(t *testing.T) []sample {
	t.Helper()
	data, err := os.ReadFile("metrics/testdata/samples.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Samples []sample `json:"samples"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Samples) != 1000 {
		t.Fatalf("got %d samples, want 1000", len(doc.Samples))
	}
	return doc.Samples
}

func faceOfStyle(t *testing.T, style string) UFontFace {
	t.Helper()
	switch style {
	case "plain":
		return NormalFace()
	case "bold":
		return BoldFace()
	case "italic":
		return ItalicFace()
	case "bolditalic":
		return BoldItalicFace()
	}
	t.Fatalf("unknown style %q", style)
	return UFontFace{}
}

// TestSamples checks the 1000 texts measured by the reference JVM. The
// acceptance criterion is 0.01 px; the tables reproduce Java's values
// exactly.
func TestSamples(t *testing.T) {
	const tolerance = 0.01
	bounders := []StringBounder{NewStringBounderSvg(nil), StringBounderAwt{}}
	var maxWidth, maxHeight, maxDescent float64
	for i, s := range readSamples(t) {
		font := Build(s.Family, faceOfStyle(t, s.Style), s.Size)
		for _, b := range bounders {
			dim := b.CalculateDimension(font, s.Text)
			descent := b.Descent(font, s.Text)
			dw := math.Abs(dim.Width - s.Width)
			dh := math.Abs(dim.Height - s.Height)
			dd := math.Abs(descent - s.Descent)
			maxWidth, maxHeight, maxDescent = max(maxWidth, dw), max(maxHeight, dh), max(maxDescent, dd)
			if dw > tolerance || dh > tolerance || dd > tolerance {
				t.Errorf("sample %d (%s %s %d %q) with %v: got %v x %v descent %v, want %v x %v descent %v",
					i, s.Family, s.Style, s.Size, s.Text, b.FileFormat(), dim.Width, dim.Height, descent,
					s.Width, s.Height, s.Descent)
			}
		}
	}
	t.Logf("maximum errors: width %g, height %g, descent %g", maxWidth, maxHeight, maxDescent)
	if maxWidth != 0 || maxHeight != 0 || maxDescent != 0 {
		t.Errorf("measurements are not exact: maximum errors width %g, height %g, descent %g",
			maxWidth, maxHeight, maxDescent)
	}
}
