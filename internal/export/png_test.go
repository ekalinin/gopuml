// SPDX-License-Identifier: MIT
// Provenance: original

package export_test

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/conformance"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/svg/svgtest"
	ipng "github.com/ekalinin/plantuml-in-go/internal/png"
)

var pngOut = flag.String("png-out", "", "write the PNG files of the export tests to this directory")

// pngKnownDifferences lowers the similarity required for cases this port
// cannot reproduce exactly, with the reason.
var pngKnownDifferences = map[string]struct {
	minSSIM float64
	reason  string
}{
	// Upstream draws the watermark before any text, with fractional metrics
	// off: FreeType hints the glyphs, which this port does not (the
	// positions match, with the hinted advances of the default font).
	"watermark": {0.85, "hinted glyphs of the watermark"},
}

// TestGoldenExportPNG runs testdata/png-cases.txt and compares, with
// upstream's TextBlockExporter (golden-png/, written by
// internal/klimt/g2d/testdata/javagen/G2dGolden.java): the image data, the
// PNG size, the plantuml and copyleft text chunks, and the image (SSIM).
func TestGoldenExportPNG(t *testing.T) {
	cases, err := svgtest.ParseFile("testdata/png-cases.txt")
	if err != nil {
		t.Fatal(err)
	}
	imageData := map[string]string{}
	raw, err := os.ReadFile("testdata/golden-png/imagedata.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		name, rest, _ := strings.Cut(line, "\t")
		imageData[name] = rest
	}
	if len(cases) < 20 || len(cases) != len(imageData) {
		t.Fatalf("%d cases, %d image data", len(cases), len(imageData))
	}
	var report strings.Builder
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", "golden-png", c.Name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			got, data, err := runExport(c)
			if err != nil {
				t.Fatal(err)
			}
			if *pngOut != "" {
				if err := os.WriteFile(filepath.Join(*pngOut, c.Name+".png"), got, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			size := "0\t0"
			if len(got) > 0 {
				im, err := png.Decode(bytes.NewReader(got))
				if err != nil {
					t.Fatal(err)
				}
				size = fmt.Sprintf("%d\t%d", im.Bounds().Dx(), im.Bounds().Dy())
			}
			if d := fmt.Sprintf("%d\t%d\t%d\t%s", data.Width(), data.Height(), data.Status(), size); d != imageData[c.Name] {
				t.Errorf("image data and size %q, upstream %q", d, imageData[c.Name])
			}
			if len(want) == 0 || len(got) == 0 {
				if len(want) != len(got) {
					t.Errorf("%d bytes, upstream %d", len(got), len(want))
				}
				return
			}
			for _, key := range []string{"plantuml", "copyleft"} {
				gv, gok, gerr := ipng.MetadataTag(bytes.NewReader(got), key)
				wv, wok, werr := ipng.MetadataTag(bytes.NewReader(want), key)
				if gv != wv || gok != wok || gerr != nil || werr != nil {
					t.Errorf("%s chunk %q %v %v, upstream %q %v %v", key, gv, gok, gerr, wv, wok, werr)
				}
			}
			wantImg, err := png.Decode(bytes.NewReader(want))
			if err != nil {
				t.Fatal(err)
			}
			gotImg, _ := png.Decode(bytes.NewReader(got))
			minSSIM := 0.98
			if k, ok := pngKnownDifferences[c.Name]; ok {
				minSSIM = k.minSSIM
			}
			score, err := conformance.CompareImages(wantImg, gotImg, minSSIM)
			fmt.Fprintf(&report, "%-24s %.4f\n", c.Name, score)
			if err != nil {
				t.Error(err)
			}
		})
	}
	t.Logf("SSIM per case:\n%s", report.String())
}

// TestPNGLinks checks that the PNG graphic records the links it draws in
// the image map, without javascript links.
func TestPNGLinks(t *testing.T) {
	block := &scriptBlock{width: 120, height: 60, ops: [][]string{
		{"url", "https://example.com", "tooltip"}, {"color", "#181818"}, {"rect", "50", "30"}, {"endurl"},
		{"url", "javascript:alert(1)"}, {"translate", "60", "0"}, {"rect", "50", "30"}, {"endurl"}}}
	var buf bytes.Buffer
	data, err := export.NewBuilder(block, export.NewFileFormatOption(font.FormatPNG), false).Build().ExportTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !data.ContainsCMapData() {
		t.Fatal("no image map")
	}
	cmap := data.CMapData("plantuml")
	if !strings.Contains(cmap, "https://example.com") || strings.Contains(cmap, "javascript") {
		t.Errorf("image map %q", cmap)
	}
	buf.Reset()
	data, err = export.NewBuilder(block, export.NewFileFormatOption(font.FormatPNG), false).
		AllowJavascriptInLink(true).Build().ExportTo(&buf)
	if err != nil || !strings.Contains(data.CMapData("plantuml"), "javascript") {
		t.Errorf("javascript allowed: %v", err)
	}
}

// TestPNGImageLimit checks the limit of the image size.
func TestPNGImageLimit(t *testing.T) {
	block := &scriptBlock{width: 300, height: 50, ops: [][]string{{"color", "#181818"}, {"rect", "290", "40"}}}
	var buf bytes.Buffer
	if _, err := export.NewBuilder(block, export.NewFileFormatOption(font.FormatPNG), false).ImageLimit(200).
		Build().ExportTo(&buf); err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(&buf)
	if err != nil || im.Bounds().Dx() != 200 || im.Bounds().Dy() != 50 {
		t.Errorf("size %v %v", im.Bounds(), err)
	}
}
