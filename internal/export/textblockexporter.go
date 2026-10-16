// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/TextBlockExporter.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package export

import (
	"errors"
	"fmt"
	"io"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/svg"
	"github.com/ekalinin/plantuml-in-go/internal/security"
)

// ErrNotImplemented is wrapped by the errors of output formats, text
// measurers and modes that are not ported yet; the error names the format.
var ErrNotImplemented = errors.New("not implemented yet")

// NotImplementedError is an output format, or a part of one, that is not
// ported yet.
type NotImplementedError struct {
	Format font.FileFormat
	// What names the missing part, for example "UGraphicG2d".
	What string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("export: %s output is not implemented yet (%s)", e.Format, e.What)
}

// Unwrap returns ErrNotImplemented.
func (e *NotImplementedError) Unwrap() error { return ErrNotImplemented }

func notImplemented(format font.FileFormat, what string) error {
	return &NotImplementedError{Format: format, What: what}
}

func nullPointerError() error { return &color.RuntimeError{Class: "java.lang.NullPointerException"} }

func unsupportedOperationError() error {
	return &color.RuntimeError{Class: "java.lang.UnsupportedOperationException"}
}

// TextBlockExporter writes a fully decorated text block (title, header,
// footer... already added) in an output format: it computes the size, the
// scale and dpi, creates the output graphic, draws the margins and the
// border, then writes the stream. Build one with NewBuilder.
//
// Only SVG (and NULL, which writes nothing) is ported; the other formats
// give an ErrNotImplemented error naming the format. The PNG and TXT
// graphics plug into createUGraphic without changing the callers.
type TextBlockExporter struct {
	textBlock        klimt.TextBlock
	fileFormatOption FileFormatOption
	skinParam        klimt.ISkinParam
	stringBounder    font.StringBounder
	stringBounderErr error
	backcolor        color.HColor
	margin           klimt.ClockwiseTopRightBottomLeft
	metadata         string
	seed             int64
	status           int
	scale            Scale
	warningOrError   string
	pragma           *klimt.Pragma
	isHandwritten    bool
	diagramType      core.DiagramType
	hasDiagramType   bool
	securityProfile  security.Profile
	allowJavascript  bool
}

func newTextBlockExporter(b *Builder) *TextBlockExporter {
	e := &TextBlockExporter{textBlock: b.textBlock, fileFormatOption: b.fileFormatOption, skinParam: b.skinParam,
		backcolor: b.backcolor, margin: b.margin, metadata: b.metadata, seed: b.seed, status: b.status,
		scale: b.scale, warningOrError: b.warningOrError, pragma: b.pragma, diagramType: b.diagramType,
		hasDiagramType: b.hasDiagramType, isHandwritten: b.isHandwritten, allowJavascript: b.allowJavascript}
	e.securityProfile = b.securityProfileOrDefault()
	var hack font.SvgCharSizeHack = font.NoHack
	if b.skinParam != nil {
		hack = b.skinParam
	}
	pragma := e.pragma
	if pragma == nil {
		pragma = klimt.PragmaCreateEmpty()
	}
	e.stringBounder, e.stringBounderErr = b.fileFormatOption.DefaultStringBounder(hack, pragma)
	return e
}

// StringBounder returns the text measurer of the output format, nil when it
// is not ported yet.
func (e *TextBlockExporter) StringBounder() font.StringBounder { return e.stringBounder }

// ExportTo writes the text block to w and returns the image data. Formats
// that are not ported give an ErrNotImplemented error before anything is
// written.
func (e *TextBlockExporter) ExportTo(w io.Writer) (ImageData, error) {
	if e.stringBounderErr != nil {
		return nil, e.stringBounderErr
	}
	if e.isHandwritten {
		return nil, notImplemented(e.fileFormatOption.FileFormat(), "UGraphicHandwritten")
	}
	dim := e.calculateFinalDimension()

	if e.fileFormatOption.FileFormat() == font.FormatNULL {
		return NewImageDataSimple(0, 0), nil
	}

	scaleFactor := e.computeScaleFactor(dim)
	if scaleFactor <= 0 {
		panic(&color.RuntimeError{Class: "java.lang.IllegalStateException", Message: "Bad scaleFactor"})
	}

	ug, err := e.createUGraphic(dim, scaleFactor)
	if err != nil {
		return nil, err
	}

	e.maybeDrawBorder(ug, dim)

	ug = ug.Apply(klimt.NewUTranslate(e.margin.Left(), e.margin.Top()))

	e.textBlock.DrawU(ug)
	ug.FlushUg()
	if err := ug.WriteToStream(w, e.metadata, 96); err != nil {
		return nil, err
	}

	// The PNG graphic records the links it draws: they become an image map.
	if c, ok := ug.(urlCollector); ok {
		if urls := c.AllUrlsEncountered(); len(urls) > 0 {
			cmap := klimt.CMapString(urls, scaleFactor)
			return NewImageDataComplex(dim, cmap, e.warningOrError, e.status), nil
		}
	}
	return NewImageDataSimpleDim(dim, e.status), nil
}

// urlCollector is implemented by the graphics that record the links they
// draw (upstream's UGraphicG2d.getAllUrlsEncountered).
type urlCollector interface {
	AllUrlsEncountered() []*klimt.URL
}

func (e *TextBlockExporter) calculateFinalDimension() geom.XDimension2D {
	d := e.textBlock.CalculateDimension(e.stringBounder)
	return geom.NewXDimension2D(d.Width()+e.margin.Left()+e.margin.Right(),
		d.Height()+e.margin.Top()+e.margin.Bottom())
}

func (e *TextBlockExporter) computeScaleFactor(dim geom.XDimension2D) float64 {
	dpi := 96
	if e.skinParam != nil {
		dpi = e.skinParam.Dpi()
	}
	fromScale := 1.0
	if e.scale != nil {
		fromScale = e.scale.Scale(dim.Width(), dim.Height())
	}
	return fromScale * float64(dpi) / 96.0
}

func (e *TextBlockExporter) maybeDrawBorder(ug klimt.UGraphic, dim geom.XDimension2D) {
	if e.skinParam == nil {
		return
	}
	c := klimt.SkinParamUtilsColor(e.skinParam, nil, klimt.ColorParamDiagramBorder)

	stroke := e.skinParam.Thickness(klimt.LineParamDiagramBorder, nil)
	if stroke == nil && c != nil {
		simple := klimt.UStrokeSimple()
		stroke = &simple
	}
	if stroke == nil {
		return
	}

	rectangle := klimt.URectangleBuild(dim.Width()-stroke.Thickness(), dim.Height()-stroke.Thickness()).
		Rounded(e.skinParam.RoundCorner(klimt.CornerParamDiagramBorder, nil))

	if c == nil {
		c = color.Black
	}
	ug.Apply(c).Apply(*stroke).Draw(rectangle)
}

func (e *TextBlockExporter) createUGraphic(dim geom.XDimension2D, scaleFactor float64) (klimt.UGraphic, error) {
	p := e.pragma
	if p == nil {
		p = klimt.PragmaCreateEmpty()
	}
	format := e.fileFormatOption.FileFormat()
	switch format {
	case font.FormatPNG, font.FormatPNGEmpty, font.FormatRAW:
		return nil, notImplemented(format, "UGraphicG2d")
	case font.FormatSVG, font.FormatSVGDeterministic, font.FormatNULL:
		return e.createUGraphicSVG(scaleFactor, dim, p), nil
	case font.FormatPDF:
		return nil, notImplemented(format, "UGraphicPdf")
	case font.FormatEPS, font.FormatEPSText:
		return nil, notImplemented(format, "UGraphicEps")
	case font.FormatHTML5:
		return nil, notImplemented(format, "UGraphicHtml5")
	case font.FormatVDX:
		return nil, notImplemented(format, "UGraphicVdx")
	case font.FormatLATEX, font.FormatLATEXDeterministic, font.FormatLATEXNoPreamble:
		return nil, notImplemented(format, "UGraphicTikz")
	case font.FormatBraillePNG:
		return nil, notImplemented(format, "UGraphicBraille")
	case font.FormatUTXT, font.FormatATXT:
		return nil, notImplemented(format, "UGraphicTxt")
	case font.FormatDEBUG:
		return nil, notImplemented(format, "UGraphicDebug")
	}
	// Upstream throws UnsupportedOperationException.
	return nil, fmt.Errorf("export: unsupported output format %s", format)
}

func (e *TextBlockExporter) createUGraphicSVG(scaleFactor float64, dim geom.XDimension2D,
	p *klimt.Pragma) klimt.UGraphic {
	option := svg.SvgOptionBasic().WithPreserveAspectRatio(e.preserveAspectRatio())
	option = option.WithHoverPathColorRGB(e.hoverPathColorRGB())
	option = option.WithMinDim(dim)
	option = option.WithBackcolor(e.backcolor)
	option = option.WithScale(scaleFactor)
	option = option.WithColorMapper(e.fileFormatOption.ColorMapper())
	option = option.WithLinkTarget(e.svgLinkTarget())
	svgFont, _ := p.Value(klimt.PragmaSvgFont)
	option = option.WithFont(svgFont)
	if e.skinParam != nil {
		option = option.WithConfigurationStore(e.skinParam.Options())
	}
	if e.hasDiagramType {
		option = option.WithRootAttribute("data-diagram-type", e.diagramType.String())
	}

	if e.fileFormatOption.Decimal() >= 0 {
		option = option.WithDecimal(e.fileFormatOption.Decimal())
	}

	if p.IsTrue(klimt.PragmaSvgInteractive) {
		interactiveBaseFilename := "default"
		if e.hasDiagramType && e.diagramType == core.Sequence {
			interactiveBaseFilename = "sequencediagram"
		}
		option = option.WithInteractive(interactiveBaseFilename)
	}

	if e.skinParam != nil {
		option = option.WithLengthAdjust(e.skinParam.LengthAdjust())
		option = option.WithSvgDimensionStyle(e.skinParam.SvgDimensionStyle())
	}
	option = option.WithSecurityProfile(e.securityProfile).WithAllowJavascriptInLink(e.allowJavascript)

	return svg.Build(option, false, e.seed, e.stringBounder, e.fileFormatOption.FileFormat())
}

// hoverPathColorRGB returns the hovered path color, "" for none.
func (e *TextBlockExporter) hoverPathColorRGB() string {
	if v, ok := e.fileFormatOption.HoverColor(); ok {
		return v
	}
	if e.skinParam != nil {
		if c := e.skinParam.HoverPathColor(); c != nil {
			return c.ToRGB(e.fileFormatOption.ColorMapper())
		}
	}
	return ""
}

// svgLinkTarget returns the link target, "" for none (which SVG writes as
// an empty target, as upstream's null).
func (e *TextBlockExporter) svgLinkTarget() string {
	if v, ok := e.fileFormatOption.SvgLinkTarget(); ok {
		return v
	}
	if e.skinParam != nil {
		return e.skinParam.SvgLinkTarget()
	}
	return ""
}

func (e *TextBlockExporter) preserveAspectRatio() string {
	if v, ok := e.fileFormatOption.PreserveAspectRatio(); ok {
		return v
	}
	if e.skinParam != nil {
		return e.skinParam.PreserveAspectRatio()
	}
	return klimt.SkinParamDefaultPreserveAspectRatio
}

// Builder configures a TextBlockExporter.
type Builder struct {
	textBlock        klimt.TextBlock
	fileFormatOption FileFormatOption
	isHandwritten    bool

	skinParam       klimt.ISkinParam
	backcolor       color.HColor
	margin          klimt.ClockwiseTopRightBottomLeft
	metadata        string
	seed            int64
	status          int
	scale           Scale
	warningOrError  string
	pragma          *klimt.Pragma
	diagramType     core.DiagramType
	hasDiagramType  bool
	securityProfile *security.Profile
	allowJavascript bool
}

// NewBuilder returns a builder for a decorated text block and an output
// format: white background (black in dark mode, or the background of the
// text block when it has one), no margin, no metadata, seed 42, status 0.
// isHandwritten selects the handwritten mode, which is not ported yet
// (ExportTo gives an ErrNotImplemented error).
func NewBuilder(textBlock klimt.TextBlock, fileFormatOption FileFormatOption, isHandwritten bool) *Builder {
	b := &Builder{textBlock: textBlock, fileFormatOption: fileFormatOption, isHandwritten: isHandwritten,
		backcolor: color.White.WithDark(color.Black), margin: klimt.ClockwiseNone(), seed: 42}
	if bc := textBlock.Backcolor(); bc != nil {
		b.backcolor = bc
	}
	return b
}

// SkinParam sets the skin parameters (dpi, border, SVG settings); nil means
// none.
func (b *Builder) SkinParam(skinParam klimt.ISkinParam) *Builder {
	b.skinParam = skinParam
	return b
}

// Backcolor sets the background. It must not be nil for SVG, as upstream.
func (b *Builder) Backcolor(backcolor color.HColor) *Builder {
	b.backcolor = backcolor
	return b
}

// Margin sets the margins around the text block.
func (b *Builder) Margin(margin klimt.ClockwiseTopRightBottomLeft) *Builder {
	b.margin = margin
	return b
}

// Metadata sets the diagram source embedded in the output; "" means none.
func (b *Builder) Metadata(metadata string) *Builder {
	b.metadata = metadata
	return b
}

// Seed sets the seed of the SVG ids.
func (b *Builder) Seed(seed int64) *Builder {
	b.seed = seed
	return b
}

// Status sets the status of the image data.
func (b *Builder) Status(status int) *Builder {
	b.status = status
	return b
}

// Scale sets the scale; nil means none.
func (b *Builder) Scale(scale Scale) *Builder {
	b.scale = scale
	return b
}

// WarningOrError sets the warning or error of the image data.
func (b *Builder) WarningOrError(warningOrError string) *Builder {
	b.warningOrError = warningOrError
	return b
}

// Pragma sets the pragmas (SVG font, interactive SVG); nil means none.
func (b *Builder) Pragma(pragma *klimt.Pragma) *Builder {
	b.pragma = pragma
	return b
}

// DiagramType sets the diagram type, written as data-diagram-type.
func (b *Builder) DiagramType(diagramType core.DiagramType) *Builder {
	b.diagramType = diagramType
	b.hasDiagramType = true
	return b
}

// SecurityProfile sets the security profile, which decides whether SVG
// images are inlined (INSECURE). Upstream reads it from process state; by
// default it comes from the render context of the skin parameters, else
// LEGACY.
func (b *Builder) SecurityProfile(profile security.Profile) *Builder {
	b.securityProfile = &profile
	return b
}

// AllowJavascriptInLink keeps "javascript:" links in SVG output, upstream's
// PLANTUML_ALLOW_JAVASCRIPT_IN_LINK environment variable (false by
// default).
func (b *Builder) AllowJavascriptInLink(allow bool) *Builder {
	b.allowJavascript = allow
	return b
}

func (b *Builder) securityProfileOrDefault() security.Profile {
	if b.securityProfile != nil {
		return *b.securityProfile
	}
	if b.skinParam != nil {
		if ctx := b.skinParam.PathSystem(); ctx != nil {
			if ctx.Security != nil {
				return ctx.Security.Profile
			}
			if p, ok := security.ParseProfile(ctx.SecurityProfile); ok {
				return p
			}
		}
	}
	return security.Legacy
}

// StyledDiagram is what Builder.Styled reads from a diagram, upstream's
// TitledDiagram.
type StyledDiagram interface {
	SkinParam() klimt.ISkinParam
	// CalculateBackColor returns the background of the diagram.
	CalculateBackColor() color.HColor
	// DefaultMargins returns the margins used when the document style has
	// none.
	DefaultMargins() klimt.ClockwiseTopRightBottomLeft
	// Metadata returns the source and version embedded in the output.
	Metadata() string
	Seed() int64
	// WarningOrError returns the warning or error, "" for none.
	WarningOrError() string
	// Scale returns the "scale" of the diagram, nil for none.
	Scale() Scale
	Pragma() *klimt.Pragma
	DiagramType() core.DiagramType
}

// Styled configures the builder from a diagram: skin parameters,
// background, margins (of the document style, else the diagram's),
// metadata when the format option has metadata, seed, warning, scale,
// pragmas, type and TikZ font distortion.
func (b *Builder) Styled(diagram StyledDiagram) *Builder {
	b.skinParam = diagram.SkinParam()
	b.backcolor = diagram.CalculateBackColor()
	b.margin = calculateMargin(diagram)
	b.metadata = ""
	if b.fileFormatOption.IsWithMetadata() {
		b.metadata = diagram.Metadata()
	}
	b.seed = diagram.Seed()
	b.warningOrError = diagram.WarningOrError()
	b.status = 0
	b.scale = diagram.Scale()
	b.pragma = diagram.Pragma()
	b.diagramType = diagram.DiagramType()
	b.hasDiagramType = true
	b.fileFormatOption = b.fileFormatOption.WithTikzFontDistortion(diagram.SkinParam().TikzFontDistortion())
	b.fileFormatOption.TikzFontDistortion().UpdateFromPragma(diagram.Pragma())
	return b
}

// Build returns the exporter.
func (b *Builder) Build() *TextBlockExporter { return newTextBlockExporter(b) }

func calculateMargin(diagram StyledDiagram) klimt.ClockwiseTopRightBottomLeft {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument).
		MergedStyle(diagram.SkinParam().CurrentStyleBuilder())
	if style.HasValue(klimt.PNameMargin) {
		return style.Margin()
	}
	return diagram.DefaultMargins()
}
