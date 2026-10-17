// SPDX-License-Identifier: MIT
// Provenance: original

package g2d

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/conformance"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/svg/svgtest"
	ipng "github.com/ekalinin/gopuml/internal/png"
)

var outDir = flag.String("g2d-out", "", "write the PNG files of the driver tests to this directory")

// minSSIM is the similarity each driver test case must reach against
// upstream's image (task 13.1).
const minSSIM = 0.98

// g2dCase holds the settings of a driver test case.
type g2dCase struct {
	width, height int
	back          color.HColor
	dpi           float64
	format        font.FileFormat
}

// runCase draws a case of testdata/cases.txt and returns the PNG.
func runCase(c svgtest.Case) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return runCaseRaw(c)
}

func runCaseRaw(c svgtest.Case) ([]byte, error) {
	s := g2dCase{width: 200, height: 100, back: color.White, dpi: 1, format: font.FormatPNG}
	var root *UGraphicG2d
	var d svgtest.Drawer
	var ug klimt.UGraphic
	for _, t := range c.Lines {
		if t[0] == "opt" {
			switch t[1] {
			case "size":
				s.width, _ = strconv.Atoi(t[2])
				s.height, _ = strconv.Atoi(t[3])
			case "back":
				s.back = svgtest.Color(t[2])
			case "dpi":
				s.dpi, _ = strconv.ParseFloat(t[2], 64)
			case "format":
				s.format = svgtest.FileFormat(t[2])
			default:
				return nil, fmt.Errorf("bad option %q", t[1])
			}
			continue
		}
		if root == nil {
			root = newCaseGraphic(s)
			d.Root = root
			ug = root
		}
		switch t[0] {
		case "aa":
			aa := klimt.AntiAliasingOn
			if t[1] == "off" {
				aa = klimt.AntiAliasingOff
			}
			ug = ug.Apply(aa)
		case "pattern":
			p := map[string]klimt.UPattern{"FULL": klimt.UPatternFull,
				"HORIZONTAL_STRIPE": klimt.UPatternHorizontalStripe, "VERTICAL_STRIPE": klimt.UPatternVerticalStripe,
				"SMALL_CIRCLE": klimt.UPatternSmallCircle}[t[1]]
			ug = ug.Apply(p)
		default:
			ug = d.Draw(ug, t)
		}
	}
	if root == nil {
		root = newCaseGraphic(s)
	}
	var buf bytes.Buffer
	if err := root.WriteToStream(&buf, c.Metadata, 96); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func newCaseGraphic(s g2dCase) *UGraphicG2d {
	b := NewEmptyImageBuilder("", s.width, s.height, s.back.ToColor(color.IdentityMapper), DefaultImageLimit)
	return New(color.White, color.IdentityMapper, font.StringBounderAwt{}, b, s.dpi, s.format)
}

// TestDriverMatchesUpstream draws testdata/cases.txt and compares each
// image with upstream's (golden/, written by testdata/javagen/G2dGolden.java
// in the reference image): same size, SSIM at least minSSIM.
func TestDriverMatchesUpstream(t *testing.T) {
	cases, err := svgtest.ParseFile("testdata/cases.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) < 30 {
		t.Fatalf("only %d cases", len(cases))
	}
	scores := map[string]float64{}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := runCase(c)
			if err != nil {
				t.Fatal(err)
			}
			if *outDir != "" {
				if err := os.WriteFile(filepath.Join(*outDir, c.Name+".png"), got, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			wantData, err := os.ReadFile(filepath.Join("testdata", "golden", c.Name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := png.Decode(bytes.NewReader(wantData))
			if err != nil {
				t.Fatal(err)
			}
			gotImg, err := png.Decode(bytes.NewReader(got))
			if err != nil {
				t.Fatal(err)
			}
			if want.Bounds() != gotImg.Bounds() {
				t.Fatalf("size %v, upstream %v", gotImg.Bounds().Size(), want.Bounds().Size())
			}
			score, err := conformance.CompareImages(want, gotImg, minSSIM)
			scores[c.Name] = score
			if err != nil {
				t.Error(err)
			}
			if _, isGray := want.(*image.Gray); isGray != false {
				t.Errorf("unexpected gray image")
			}
		})
	}
	names := make([]string, 0, len(scores))
	for n := range scores {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintf(&b, "%-24s %.4f\n", n, scores[n])
	}
	t.Logf("SSIM per case:\n%s", b.String())
}

// TestMetadataRoundTrip checks that the plantuml text chunk written by the
// driver reads back.
func TestMetadataRoundTrip(t *testing.T) {
	c := svgtest.Case{Name: "meta", Lines: [][]string{{"color", "#181818"}, {"rect", "10", "10"}},
		Metadata: "@startuml\nAlice -> Bob : привет\n@enduml"}
	out, err := runCase(c)
	if err != nil {
		t.Fatal(err)
	}
	text, ok, err := ipng.MetadataTag(bytes.NewReader(out), "plantuml")
	if err != nil || !ok || text != c.Metadata {
		t.Errorf("metadata %q %v %v", text, ok, err)
	}
	copyleft, ok, err := ipng.MetadataTag(bytes.NewReader(out), "copyleft")
	if err != nil || !ok || copyleft != "Generated by https://plantuml.com" {
		t.Errorf("copyleft %q %v %v", copyleft, ok, err)
	}
	c.Metadata = ""
	out, err = runCase(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := ipng.MetadataTag(bytes.NewReader(out), "plantuml"); ok {
		t.Error("metadata written without metadata")
	}
}
