// SPDX-License-Identifier: MIT
// Provenance: original

package export_test

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/security"
)

func rectBlock() *scriptBlock {
	return &scriptBlock{width: 40, height: 20, ops: [][]string{{"color", "#181818"}, {"rect", "10", "10"}}}
}

func TestNotImplementedFormats(t *testing.T) {
	for _, f := range []font.FileFormat{font.FormatATXT,
		font.FormatUTXT, font.FormatSVGDeterministic, font.FormatEPS, font.FormatPDF, font.FormatLATEX,
		font.FormatHTML5, font.FormatVDX, font.FormatBraillePNG, font.FormatDEBUG} {
		var buf bytes.Buffer
		_, err := export.NewBuilder(rectBlock(), export.NewFileFormatOption(f), false).Build().ExportTo(&buf)
		var nie *export.NotImplementedError
		if !errors.Is(err, export.ErrNotImplemented) || !errors.As(err, &nie) || nie.Format != f ||
			!strings.Contains(err.Error(), f.String()) {
			t.Errorf("%s: error %v", f, err)
		}
		if buf.Len() != 0 {
			t.Errorf("%s: wrote %d bytes", f, buf.Len())
		}
	}
	_, err := export.NewBuilder(rectBlock(), export.NewFileFormatOption(font.FormatSVG), true).Build().
		ExportTo(&bytes.Buffer{})
	if !errors.Is(err, export.ErrNotImplemented) || !strings.Contains(err.Error(), "UGraphicHandwritten") {
		t.Errorf("handwritten: %v", err)
	}
	_, err = export.NewBuilder(rectBlock(), export.NewFileFormatOption(font.FormatPREPROC), false).Build().
		ExportTo(&bytes.Buffer{})
	if err == nil || errors.Is(err, export.ErrNotImplemented) {
		t.Errorf("unsupported format: %v", err)
	}
}

func TestStringBounders(t *testing.T) {
	sb, err := export.NewFileFormatOption(font.FormatSVG).DefaultStringBounder(nil, klimt.PragmaCreateEmpty())
	if err != nil || sb.FileFormat() != font.FormatSVG {
		t.Errorf("SVG: %v %v", sb, err)
	}
	sb, err = export.NewFileFormatOption(font.FormatPNG).DefaultStringBounder(nil, klimt.PragmaCreateEmpty())
	if err != nil || sb.FileFormat() != font.FormatPNG {
		t.Errorf("PNG: %v %v", sb, err)
	}
	e := export.NewBuilder(rectBlock(), export.NewFileFormatOption(font.FormatSVG), false).Build()
	if e.StringBounder() == nil {
		t.Error("exporter bounder")
	}
}

func TestScales(t *testing.T) {
	tests := []struct {
		s    export.Scale
		want float64
	}{
		{export.ScaleSimple(1.5), 1.5},
		{export.ScaleSimple(0), 1},
		{export.ScaleSimple(-2), 1},
		{export.ScaleSimple(9), 4},
		{export.ScaleWidth(300), 3},
		{export.ScaleHeight(25), 0.5},
		{export.ScaleMaxWidth(300), 1},
		{export.ScaleMaxWidth(50), 0.5},
		{export.ScaleMaxHeight(10), 0.2},
		{export.ScaleMaxHeight(100), 1},
		{export.ScaleWidthAndHeight(200, 200), 2},
		{export.ScaleMaxWidthAndHeight(200, 200), 1},
		{export.ScaleMaxWidthAndHeight(50, 200), 0.5},
	}
	for i, tt := range tests {
		if got := tt.s.Scale(100, 50); got != tt.want {
			t.Errorf("scale %d: %v, want %v", i, got, tt.want)
		}
	}
	// As upstream, NaN passes through the protection.
	if got := export.ScaleSimple(math.NaN()).Scale(1, 1); !math.IsNaN(got) {
		t.Errorf("NaN: %v", got)
	}
	if got := export.ScaleMaxWidth(10).Scale(0, 1); got != 1 {
		t.Errorf("division by zero: %v", got)
	}
}

func TestImageData(t *testing.T) {
	ok := export.ImageDataOK()
	if ok.Width() != 0 || ok.Status() != 0 || ok.ContainsCMapData() || ok.RootCause() != nil || ok.WarningOrError() != "" {
		t.Error("ok")
	}
	cause := errors.New("boom")
	ko := export.ImageDataError(cause)
	if ko.Status() != 503 || ko.RootCause() != cause {
		t.Error("error")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("CMapData of a simple image does not panic")
			}
		}()
		ok.CMapData("x")
	}()
	cmap := klimt.CMapString([]*klimt.URL{}, 1)
	c := export.NewImageDataComplex(geom.NewXDimension2D(12.9, 3.5), cmap, "warn", 7)
	if c.Width() != 12 || c.Height() != 3 || c.Status() != 7 || c.ContainsCMapData() || c.WarningOrError() != "warn" {
		t.Errorf("complex %d %d", c.Width(), c.Height())
	}
	if export.NewImageDataComplex(geom.NewXDimension2D(1, 1), nil, "", 0).ContainsCMapData() {
		t.Error("nil cmap")
	}
}

func TestFileFormatOption(t *testing.T) {
	o := export.NewFileFormatOption(font.FormatSVG)
	if !o.IsWithMetadata() || o.Decimal() != -1 || o.ScaleCoef() != 1 || o.ColorMapper() != color.IdentityMapper ||
		o.String() != "SVG" || o.IsUseRedForError() || o.IsDebugSvek() {
		t.Error("defaults")
	}
	if _, ok := o.SvgLinkTarget(); ok {
		t.Error("link target set")
	}
	o2 := o.WithSvgLinkTarget("").WithHoverColor("#FF0000").WithPreserveAspectRatio("xMinYMin").WithScale(2).
		WithWatermark("w").WithDecimal(1).WithColorMapper(color.DarkModeMapper).WithUseRedForError()
	if v, ok := o2.SvgLinkTarget(); !ok || v != "" {
		t.Error("empty link target is set")
	}
	if v, _ := o2.HoverColor(); v != "#FF0000" {
		t.Error("hover")
	}
	if v, _ := o2.PreserveAspectRatio(); v != "xMinYMin" {
		t.Error("preserve")
	}
	if v, _ := o2.Watermark(); v != "w" || o2.ScaleCoef() != 2 || o2.Decimal() != 1 || !o2.IsUseRedForError() {
		t.Error("with")
	}
	if _, ok := o.HoverColor(); ok {
		t.Error("copies share state")
	}
	o2.HideMetadata()
	o2.SetDebugSvek(true)
	if o2.IsWithMetadata() || !o2.IsDebugSvek() || !o.IsWithMetadata() {
		t.Error("in place changes")
	}
	if !export.NewFileFormatOptionMetadata(font.FormatPNG, true).IsWithMetadata() ||
		export.NewFileFormatOptionMetadata(font.FormatPNG, false).IsWithMetadata() {
		t.Error("metadata constructor")
	}
}

func TestNullFormat(t *testing.T) {
	var buf bytes.Buffer
	data, err := export.NewBuilder(rectBlock(), export.NewFileFormatOption(font.FormatNULL), false).
		Metadata("@startuml\na\n@enduml").Build().ExportTo(&buf)
	if err != nil || buf.Len() != 0 || data.Width() != 0 || data.Height() != 0 {
		t.Errorf("NULL: %v %d", err, buf.Len())
	}
}

func TestSecurityProfile(t *testing.T) {
	svgImage := `<svg xmlns="http://www.w3.org/2000/svg" width="4" height="4"><g/></svg>`
	block := &scriptBlock{width: 10, height: 10, ops: [][]string{{"imagesvg", "1", svgImage}}}
	run := func(b *export.Builder) string {
		var buf bytes.Buffer
		if _, err := b.Build().ExportTo(&buf); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	ffo := export.NewFileFormatOption(font.FormatSVG)
	if out := run(export.NewBuilder(block, ffo, false)); !strings.Contains(out, "data:image/svg+xml;base64,") {
		t.Errorf("default profile inlines: %s", out)
	}
	if out := run(export.NewBuilder(block, ffo, false).SecurityProfile(security.Insecure)); !strings.Contains(out,
		`<svg x="0" y="0"><g/></svg>`) {
		t.Errorf("insecure does not inline: %s", out)
	}
	ctx := core.NewRenderContext()
	ctx.Security = &security.Config{Profile: security.Insecure}
	sp := klimt.SkinParamCreate(map[string]string{}, ctx, core.Sequence, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	if out := run(export.NewBuilder(block, ffo, false).SkinParam(sp)); !strings.Contains(out, `<svg x="0" y="0">`) {
		t.Errorf("render context profile ignored: %s", out)
	}
	linkBlock := &scriptBlock{width: 10, height: 10, ops: [][]string{{"url", "javascript:x()"}, {"rect", "1", "1"},
		{"endurl"}}}
	if out := run(export.NewBuilder(linkBlock, ffo, false)); !strings.Contains(out, `href=""`) {
		t.Errorf("javascript link kept: %s", out)
	}
	if out := run(export.NewBuilder(linkBlock, ffo, false).AllowJavascriptInLink(true)); !strings.Contains(out,
		`href="javascript:x()"`) {
		t.Errorf("javascript link removed: %s", out)
	}
}

// fakeDiagram is a StyledDiagram.
type fakeDiagram struct {
	sp *klimt.SkinParam
}

func (d fakeDiagram) SkinParam() klimt.ISkinParam      { return d.sp }
func (d fakeDiagram) CalculateBackColor() color.HColor { return color.Red }
func (d fakeDiagram) DefaultMargins() klimt.ClockwiseTopRightBottomLeft {
	return klimt.ClockwiseSame(7)
}
func (d fakeDiagram) Metadata() string       { return "@startuml\nx\n@enduml" }
func (d fakeDiagram) Seed() int64            { return 99 }
func (d fakeDiagram) WarningOrError() string { return "" }
func (d fakeDiagram) Scale() export.Scale    { return export.ScaleSimple(2) }
func (d fakeDiagram) Pragma() *klimt.Pragma  { return klimt.PragmaCreateEmpty() }
func (d fakeDiagram) DiagramType() core.DiagramType {
	return core.Class
}

func TestStyled(t *testing.T) {
	sp := klimt.SkinParamCreate(map[string]string{}, nil, core.Class, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	run := func(ffo export.FileFormatOption) (string, export.ImageData) {
		var buf bytes.Buffer
		data, err := export.NewBuilder(rectBlock(), ffo, false).Styled(fakeDiagram{sp: sp}).Build().ExportTo(&buf)
		if err != nil {
			t.Fatal(err)
		}
		return buf.String(), data
	}
	out, data := run(export.NewFileFormatOption(font.FormatSVG))
	for _, want := range []string{`data-diagram-type="CLASS"`, "background:#FF0000;", "<?plantuml-src ",
		`style="width:`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The document style of the plantuml skin has no margin: the diagram's
	// default margins apply.
	if data.Width() != 40+7+7 || data.Height() != 20+7+7 {
		t.Errorf("size %dx%d", data.Width(), data.Height())
	}
	out, _ = run(export.NewFileFormatOptionMetadata(font.FormatSVG, false))
	if strings.Contains(out, "plantuml-src") {
		t.Error("metadata without metadata")
	}
}

// A document style margin wins over the diagram's default margins.
func TestStyledDocumentMargin(t *testing.T) {
	sp := klimt.SkinParamCreate(map[string]string{}, nil, core.Class, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	styles, err := klimt.StyleLoaderParseStyleText(command.SingleString("document { Margin 3 }"), &klimt.AutomaticCounterBasic{})
	if err != nil {
		t.Fatal(err)
	}
	sp.MuteStyle(styles)
	var buf bytes.Buffer
	data, err := export.NewBuilder(rectBlock(), export.NewFileFormatOption(font.FormatSVG), false).
		Styled(fakeDiagram{sp: sp}).Build().ExportTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if data.Width() != 40+3+3 {
		t.Errorf("width %d", data.Width())
	}
}

func TestBuilderDefaults(t *testing.T) {
	var buf bytes.Buffer
	block := rectBlock()
	block.back = color.Blue
	if _, err := export.NewBuilder(block, export.NewFileFormatOption(font.FormatSVG), false).Status(3).
		WarningOrError("w").Build().ExportTo(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "background:#0000FF;") {
		t.Errorf("text block background ignored: %s", buf.String())
	}
}
