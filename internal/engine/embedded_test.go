// SPDX-License-Identifier: MIT
// Provenance: original

package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
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

// TestEmbeddedUnsupportedType checks the styling scenario "Embedded
// diagram of an unsupported type": the note shows the "not supported yet"
// error image of the nested state diagram, and the rest of the diagram
// renders normally.
func TestEmbeddedUnsupportedType(t *testing.T) {
	src := "@startuml\nAlice -> Bob : hello\nnote right\n{{\n[*] --> Idle\n}}\nend note\n@enduml\n"
	blocks, err := Parse(context.Background(), core.NewRenderContext(), Source{Text: src})
	if err != nil || len(blocks) != 1 {
		t.Fatalf("%d blocks, error %v", len(blocks), err)
	}
	var buf bytes.Buffer
	if _, err := blocks[0].Render(context.Background(), &buf, 0, Options{Format: font.FormatSVG}); err != nil {
		t.Fatal(err)
	}
	svg := buf.String()
	if !strings.Contains(svg, ">hello</text>") {
		t.Error("the sequence diagram is not drawn")
	}
	m := regexp.MustCompile(`data:image/svg\+xml;base64,([A-Za-z0-9+/=]+)`).FindStringSubmatch(svg)
	if m == nil {
		t.Fatal("no embedded diagram")
	}
	nested, err := base64.StdEncoding.DecodeString(m[1])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(nested), "state diagrams are not supported yet") {
		t.Errorf("the embedded diagram is not the error image: %.300s", nested)
	}
}
