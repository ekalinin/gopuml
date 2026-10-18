// SPDX-License-Identifier: MIT
// Provenance: original

package engine

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
)

var nestedSequence = []string{"@startuml", "Bob -> Carol : nested", "@enduml"}

func TestEmbeddedDiagramRenderer(t *testing.T) {
	rc := withEmbeddedDiagrams(core.NewRenderContext())
	if rc.EmbeddedDiagram == nil {
		t.Fatal("no renderer")
	}
	svg, err := rc.EmbeddedDiagram(nestedSequence, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(svg), ">nested</text>") {
		t.Errorf("the SVG has no message: %.200s", svg)
	}
	img := klimt.NewUImageSvg(string(svg), 1)
	if img.Width() <= 0 || img.Height() <= 0 {
		t.Errorf("SVG size %vx%v", img.Width(), img.Height())
	}
	data, err := rc.EmbeddedDiagram(nestedSequence, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if w, h := pngSize(t, data); w <= 0 || h <= 0 {
		t.Errorf("PNG size %dx%d", w, h)
	}
}

func pngSize(t *testing.T, data []byte) (int, int) {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return decoded.Bounds().Dx(), decoded.Bounds().Dy()
}

// TestEmbeddedDiagramPrevious checks upstream's Previous: the nested
// diagram starts with the skinparam values of the embedding diagram, such
// as "dpi", which the PNG export reads from them.
func TestEmbeddedDiagramPrevious(t *testing.T) {
	rc := withEmbeddedDiagrams(core.NewRenderContext())
	plain, err := rc.EmbeddedDiagram(nestedSequence, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	scaled, err := rc.EmbeddedDiagram(nestedSequence, map[string]string{"dpi": "192"}, false)
	if err != nil {
		t.Fatal(err)
	}
	w1, _ := pngSize(t, plain)
	w2, _ := pngSize(t, scaled)
	if w2 < w1*19/10 {
		t.Errorf("width %d with dpi 192, %d without", w2, w1)
	}
}
