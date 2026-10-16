// SPDX-License-Identifier: MIT
// Provenance: original

package export_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/export"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/svg/svgtest"
)

// scriptBlock is a text block that runs drawing commands.
type scriptBlock struct {
	klimt.TextBlockDefaults
	ops           [][]string
	width, height float64
	back          color.HColor
}

func (b *scriptBlock) DrawU(ug klimt.UGraphic) {
	d := svgtest.Drawer{Root: ug}
	for _, t := range b.ops {
		ug = d.Draw(ug, t)
	}
}

func (b *scriptBlock) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(b.width, b.height)
}

func (b *scriptBlock) Backcolor() color.HColor { return b.back }

func num(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return v
}

// runExport runs an export case: its "export" lines configure the builder,
// its other lines are drawn by the text block.
func runExport(c svgtest.Case) (out []byte, data export.ImageData, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	format := font.FormatSVG
	block := &scriptBlock{width: 100, height: 50}
	var setters []func(b *export.Builder)
	var ffoChanges []func(o export.FileFormatOption) export.FileFormatOption
	var skinParams [][2]string
	withSkinParam := false
	pragma := klimt.PragmaCreateEmpty()
	withPragma := false
	options := klimt.NewConfigurationStore[klimt.OptionKey]()
	diagramType := core.Sequence
	for _, t := range c.Lines {
		if t[0] != "export" {
			block.ops = append(block.ops, t)
			continue
		}
		switch t[1] {
		case "format":
			format = svgtest.FileFormat(t[2])
		case "dim":
			block.width, block.height = num(t[2]), num(t[3])
		case "tbback":
			block.back = svgtest.Color(t[2])
		case "back":
			c := svgtest.Color(t[2])
			setters = append(setters, func(b *export.Builder) { b.Backcolor(c) })
		case "margin":
			m := klimt.ClockwiseTopRightBottomLeftOf(num(t[2]), num(t[3]), num(t[4]), num(t[5]))
			setters = append(setters, func(b *export.Builder) { b.Margin(m) })
		case "seed":
			seed, _ := strconv.ParseInt(t[2], 10, 64)
			setters = append(setters, func(b *export.Builder) { b.Seed(seed) })
		case "scale":
			var s export.Scale
			switch t[2] {
			case "simple":
				s = export.ScaleSimple(num(t[3]))
			case "width":
				s = export.ScaleWidth(num(t[3]))
			case "maxwidth":
				s = export.ScaleMaxWidth(num(t[3]))
			}
			setters = append(setters, func(b *export.Builder) { b.Scale(s) })
		case "diagramtype":
			for dt := core.DiagramType(0); dt <= core.Unknown; dt++ {
				if dt.String() == t[2] {
					diagramType = dt
					setters = append(setters, func(b *export.Builder) { b.DiagramType(dt) })
				}
			}
		case "skinparam":
			withSkinParam = true
			if len(t) > 2 {
				skinParams = append(skinParams, [2]string{t[2], t[3]})
			}
		case "pragma":
			withPragma = true
			pragma.Define(t[2], t[3], true)
		case "svgtitle":
			withSkinParam = true
			options.Define(klimt.OptionSvgTitle, t[2], true)
		case "ffo":
			k, v := t[2], t[3]
			ffoChanges = append(ffoChanges, func(o export.FileFormatOption) export.FileFormatOption {
				switch k {
				case "linktarget":
					return o.WithSvgLinkTarget(v)
				case "hover":
					return o.WithHoverColor(v)
				case "preserve":
					return o.WithPreserveAspectRatio(v)
				case "decimal":
					d, _ := strconv.Atoi(v)
					return o.WithDecimal(d)
				case "mapper":
					return o.WithColorMapper(svgtest.Mapper(v))
				case "watermark":
					return o.WithWatermark(v)
				}
				panic("bad ffo " + k)
			})
		default:
			return nil, nil, fmt.Errorf("bad export setting %q", t[1])
		}
	}
	ffo := export.NewFileFormatOption(format)
	for _, f := range ffoChanges {
		ffo = f(ffo)
	}
	b := export.NewBuilder(block, ffo, false)
	for _, s := range setters {
		s(b)
	}
	b.Metadata(c.Metadata)
	if withPragma {
		b.Pragma(pragma)
	}
	if withSkinParam {
		sp := klimt.SkinParamCreate(map[string]string{}, nil, diagramType, pragma, options)
		for _, kv := range skinParams {
			sp.SetParam(kv[0], kv[1])
		}
		b.SkinParam(sp)
	}
	var buf bytes.Buffer
	data, err = b.Build().ExportTo(&buf)
	return buf.Bytes(), data, err
}

// TestGoldenExport runs testdata/cases.txt and compares the SVG text and
// the image data with upstream's TextBlockExporter (golden/, written by
// internal/klimt/svg/testdata/javagen/SvgGolden.java).
func TestGoldenExport(t *testing.T) {
	cases, err := svgtest.ParseFile("testdata/cases.txt")
	if err != nil {
		t.Fatal(err)
	}
	imageData := map[string]string{}
	raw, err := os.ReadFile("testdata/golden/imagedata.txt")
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
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", "golden", c.Name+".svg"))
			if err != nil {
				t.Fatal(err)
			}
			got, data, err := runExport(c)
			if err != nil {
				t.Fatal(err)
			}
			if d := fmt.Sprintf("%d\t%d\t%d", data.Width(), data.Height(), data.Status()); d != imageData[c.Name] {
				t.Errorf("image data %q, upstream %q", d, imageData[c.Name])
			}
			if string(got) == string(want) {
				return
			}
			if g, w := svgtest.Normalize(string(got)), svgtest.Normalize(string(want)); g != w {
				t.Errorf("SVG differs\n got: %s\nwant: %s", g, w)
			}
		})
	}
}
