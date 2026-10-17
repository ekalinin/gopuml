// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/core/DiagramChromeFactory.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/DecorateEntityImage.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/EntityImageLegend.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/DisplayPositioned.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"github.com/ekalinin/gopuml/internal/asciiverse"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/jmath"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// Annotated is what DiagramChromeCreate reads from a diagram: its
// annotations (upstream's Annotated).
type Annotated interface {
	Title() DisplayPositioned
	Caption() DisplayPositioned
	Legend() DisplayPositioned
	Header() DisplayPositioned
	Footer() DisplayPositioned
	MainFrame() DisplayPositioned
}

// DiagramChromeCreate wraps the drawing of a diagram with its warnings,
// main frame, legend, title, caption, header and footer, as upstream's
// DiagramChromeFactory.create. title is the title of the page.
func DiagramChromeCreate(raw klimt.TextBlock, annotated Annotated, skinParam klimt.ISkinParam,
	warnings []*klimt.Warning, title DisplayPositioned) klimt.TextBlock {
	result := raw
	result = chromeAddWarnings(result, warnings)
	result = chromeDecorateWithFrame(result, annotated, skinParam)
	result = chromeAddLegend(result, annotated, skinParam)
	result = chromeAddTitle(result, title, skinParam)
	result = chromeAddCaption(result, annotated, skinParam)
	result = chromeAddHeaderAndFooter(result, annotated, skinParam)
	return result
}

// DiagramChromeCreateASCII wraps the text drawing (ATXT, UTXT) of a
// diagram with its title, centered above it, as upstream's
// DiagramChromeFactory.createAscii. As upstream, the caption, header,
// footer, legend, main frame and warnings are not drawn in text.
func DiagramChromeCreateASCII(raw asciiverse.AsciiBlock, annotated Annotated) asciiverse.AsciiBlock {
	return chromeAddASCIITitle(raw, annotated)
}

func chromeAddASCIITitle(original asciiverse.AsciiBlock, annotated Annotated) asciiverse.AsciiBlock {
	title := annotated.Title()
	if title.IsNull() {
		return original
	}
	return asciiverse.NewAsciiBlockStackTB(asciiverse.DisplayBlock(title.Display()), original)
}

func chromeAddWarnings(original klimt.TextBlock, warnings []*klimt.Warning) klimt.TextBlock {
	if len(warnings) == 0 {
		return original
	}
	return &warningsBlock{original: original, banner: &warningBannerBlock{warnings: warnings}}
}

// warningsBlock draws the warning banner above the diagram.
type warningsBlock struct {
	klimt.TextBlockDefaults
	original klimt.TextBlock
	banner   *warningBannerBlock
}

func (b *warningsBlock) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	totalWidth := b.CalculateDimension(stringBounder).Width()
	b.banner.drawUForce(ug, totalWidth)
	bannerHeight := b.banner.CalculateDimension(stringBounder).Height()
	b.original.DrawU(ug.Apply(klimt.UTranslateDy(bannerHeight)))
}

func (b *warningsBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimBanner := b.banner.CalculateDimension(stringBounder)
	dimOriginal := b.original.CalculateDimension(stringBounder)
	return geom.NewXDimension2D(jmath.Max(dimBanner.Width(), dimOriginal.Width()),
		dimBanner.Height()+dimOriginal.Height())
}

func (b *warningsBlock) Backcolor() color.HColor { return b.original.Backcolor() }

// warningBannerBlock is the yellow warning banner.
type warningBannerBlock struct {
	klimt.TextBlockDefaults
	warnings []*klimt.Warning
}

const (
	warningLineSpacing  = 10
	warningCornerRadius = 5
)

func warningFontConfiguration() *klimt.FontConfiguration {
	return klimt.FontConfigurationBlackBlueTrue(font.Monospace(10))
}

func (b *warningBannerBlock) DrawU(ug klimt.UGraphic) { b.drawUForce(ug, 0) }

func (b *warningBannerBlock) drawUForce(ug klimt.UGraphic, forceWidth float64) {
	stringBounder := ug.StringBounder()
	dim := b.CalculateDimension(stringBounder)
	effectiveWidth := jmath.Max(dim.Width(), forceWidth)

	set := color.Instance()
	back := set.GetColorOrWhite("ffffcc").WithDark(set.GetColorOrWhite("774400"))
	border := set.GetColorOrWhite("ffdd88").WithDark(set.GetColorOrWhite("aa5500"))

	rect := klimt.URectangleBuild(effectiveWidth-10, dim.Height()-5).Rounded(warningCornerRadius)
	ug.Apply(klimt.Bg(back)).Apply(border).Apply(klimt.UStrokeWithThickness(3)).Apply(klimt.NewUTranslate(3, 3)).
		Draw(rect)

	ugText := ug.Apply(color.Black).Apply(klimt.NewUTranslate(10, 2))
	fc := warningFontConfiguration()
	for _, w := range b.warnings {
		for _, s := range w.Message() {
			text := klimt.UTextBuild(s, fc)
			height := text.CalculateDimension(stringBounder).Height()
			ugText = ugText.Apply(klimt.UTranslateDy(height))
			ugText.Draw(text)
		}
		ugText = ugText.Apply(klimt.UTranslateDy(warningLineSpacing))
	}
}

func (b *warningBannerBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	width := 0.0
	height := 0.0
	fc := warningFontConfiguration()
	for _, w := range b.warnings {
		for _, s := range w.Message() {
			lineDim := klimt.UTextBuild(s, fc).CalculateDimension(stringBounder)
			width = jmath.Max(width, lineDim.Width())
			height += lineDim.Height()
		}
		height += warningLineSpacing
	}
	// Remove trailing spacing from last warning
	if len(b.warnings) > 0 {
		height -= warningLineSpacing
	}
	return geom.NewXDimension2D(width+20, height+10)
}

func (b *warningBannerBlock) Backcolor() color.HColor { return nil }

func chromeDecorateWithFrame(original klimt.TextBlock, annotated Annotated, skinParam klimt.ISkinParam) klimt.TextBlock {
	mainFrame := annotated.MainFrame().Display()
	if klimt.DisplayIsNull(mainFrame) {
		return original
	}
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument, klimt.SNameMainframe).
		MergedStyle(skinParam.CurrentStyleBuilder())
	fontConfiguration := klimt.FontConfigurationFromStyle(skinParam, style, nil)
	title := mainFrame.Create(fontConfiguration, geom.HorizontalAlignmentCenter, skinParam)

	symbolContext := style.SymbolContext(skinParam.IHtmlColorSet(), nil)
	margin := style.Margin()

	frame := klimt.NewBigFrame(title, original, style.Padding(), symbolContext)
	return &frameBlock{original: original, title: title, frame: frame, style: style, margin: margin,
		symbolContext: symbolContext}
}

// frameBlock is the anonymous text block of upstream's decorateWithFrame.
type frameBlock struct {
	klimt.TextBlockDefaults
	original      klimt.TextBlock
	title         klimt.TextBlock
	frame         klimt.TextBlock
	style         *klimt.Style
	margin        klimt.ClockwiseTopRightBottomLeft
	symbolContext *klimt.Fashion
}

func (b *frameBlock) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	padding := b.computePadding(stringBounder)
	delta := b.computeDelta(stringBounder)
	b.frame.DrawU(ug.Apply(b.margin.Translate()))
	b.original.DrawU(ug.Apply(b.margin.Translate().Compose(padding.Translate().Compose(delta))))
}

func (b *frameBlock) InnerPosition(member string, sb font.StringBounder) *geom.XRectangle2D {
	padding := b.computePadding(sb)
	dimTitle := b.title.CalculateDimension(sb)
	originalMinMax := klimt.TextBlockUtilsMinMax(b.original, sb, false)
	dx, dy := 0.0, 0.0
	if originalMinMax.MinX() < 0 {
		dx = -originalMinMax.MinX()
	}
	if originalMinMax.MinY() < 0 {
		dy = -originalMinMax.MinY()
	}
	rect := b.original.InnerPosition(member, sb)
	result := geom.NewXRectangle2D(dx+rect.X+b.margin.Left()+padding.Left(),
		dy+rect.Y+b.margin.Top()+padding.Top()+dimTitle.Height(), rect.Width, rect.Height)
	return &result
}

func (b *frameBlock) CalculateDimension(sb font.StringBounder) geom.XDimension2D {
	frameDim := b.frame.CalculateDimension(sb)
	return geom.NewXDimension2D(b.margin.Left()+frameDim.Width()+b.margin.Right(),
		b.margin.Top()+frameDim.Height()+b.margin.Bottom())
}

func (b *frameBlock) Backcolor() color.HColor { return b.symbolContext.BackColor() }

func (b *frameBlock) computePadding(sb font.StringBounder) klimt.ClockwiseTopRightBottomLeft {
	dimTitle := b.title.CalculateDimension(sb)
	return b.style.Padding().IncTop(dimTitle.Height() + 10)
}

func (b *frameBlock) computeDelta(sb font.StringBounder) klimt.UTranslate {
	originalMinMax := klimt.TextBlockUtilsMinMax(b.original, sb, false)
	dx, dy := 0.0, 0.0
	if originalMinMax.MinX() < 0 {
		dx = -originalMinMax.MinX()
	}
	if originalMinMax.MinY() < 0 {
		dy = -originalMinMax.MinY()
	}
	return klimt.NewUTranslate(dx, dy)
}

func chromeAddLegend(original klimt.TextBlock, annotated Annotated, skinParam klimt.ISkinParam) klimt.TextBlock {
	legend := annotated.Legend()
	if legend.IsNull() {
		return original
	}
	group := klimt.NewUGroupAt(legend.LineLocation())
	group.Put(klimt.UGroupClass, "legend")

	legendBlock := EntityImageLegendCreate(legend.Display(), skinParam)
	return DecorateEntityImageAdd(group, original, legendBlock, legend.HorizontalAlignment(),
		legend.VerticalAlignment())
}

// EntityImageLegendCreate returns the box of a legend, upstream's
// EntityImageLegend.create.
func EntityImageLegendCreate(note *klimt.Display, skinParam klimt.ISkinParam) klimt.TextBlock {
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameRoot, klimt.SNameDocument,
		DiagramTypeStyleName(skinParam.DiagramType()), klimt.SNameLegend).
		MergedStyle(skinParam.CurrentStyleBuilder())
	return style.CreateTextBlockBordered(note, skinParam.IHtmlColorSet(), skinParam, klimt.StyleIDLegend,
		style.WrapWidth())
}

func chromeAddTitle(original klimt.TextBlock, title DisplayPositioned, skinParam klimt.ISkinParam) klimt.TextBlock {
	if title.IsNull() {
		return original
	}
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument, klimt.SNameTitle).
		MergedStyle(skinParam.CurrentStyleBuilder())
	titleBlock := style.CreateTextBlockBordered(title.Display(), skinParam.IHtmlColorSet(), skinParam,
		klimt.StyleIDTitle, klimt.LineBreakStrategyNone)

	group := klimt.NewUGroupAt(title.LineLocation())
	group.Put(klimt.UGroupClass, "title")
	return DecorateEntityImageAddTop(group, original, titleBlock, geom.HorizontalAlignmentCenter)
}

func chromeAddCaption(original klimt.TextBlock, annotated Annotated, skinParam klimt.ISkinParam) klimt.TextBlock {
	caption := annotated.Caption()
	if caption.IsNull() {
		return original
	}
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument, klimt.SNameCaption).
		MergedStyle(skinParam.CurrentStyleBuilder())
	captionBlock := style.CreateTextBlockBordered(caption.Display(), skinParam.IHtmlColorSet(), skinParam,
		klimt.StyleIDCaption, klimt.LineBreakStrategyNone)

	group := klimt.NewUGroupAt(caption.LineLocation())
	group.Put(klimt.UGroupClass, "caption")
	return DecorateEntityImageAddBottom(group, original, captionBlock, geom.HorizontalAlignmentCenter)
}

func chromeAddHeaderAndFooter(original klimt.TextBlock, annotated Annotated,
	skinParam klimt.ISkinParam) klimt.TextBlock {
	header := annotated.Header()
	footer := annotated.Footer()
	if footer.IsNull() && header.IsNull() {
		return original
	}

	group1 := klimt.NewUGroupAt(header.LineLocation())
	group1.Put(klimt.UGroupClass, "header")

	var textHeader klimt.TextBlock
	if !header.IsNull() {
		style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument, klimt.SNameHeader).
			MergedStyle(skinParam.CurrentStyleBuilder())
		textHeader = header.CreateRibbon(skinParam, style)
	}

	group2 := klimt.NewUGroupAt(footer.LineLocation())
	group2.Put(klimt.UGroupClass, "footer")

	var textFooter klimt.TextBlock
	if !footer.IsNull() {
		style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameDocument, klimt.SNameFooter).
			MergedStyle(skinParam.CurrentStyleBuilder())
		textFooter = footer.CreateRibbon(skinParam, style)
	}

	return DecorateEntityImageAddTopAndBottom(original, group1, textHeader, header.HorizontalAlignment(), group2,
		textFooter, footer.HorizontalAlignment())
}

// CreateRibbon returns the text block of a header or footer drawn with
// its style, nil for no text (upstream's DisplayPositioned.createRibbon
// with a style, which is always given).
func (d DisplayPositioned) CreateRibbon(spriteContainer klimt.ISkinSimple, style *klimt.Style) klimt.TextBlock {
	display := d.Display()
	if klimt.DisplayIsNull(display) || display.Size() == 0 {
		return nil
	}
	return style.CreateTextBlockBordered(display, spriteContainer.IHtmlColorSet(), spriteContainer, "",
		klimt.LineBreakStrategyNone)
}

// DiagramTypeStyleName returns the style name of a diagram type, upstream's
// DiagramType.getStyleName.
func DiagramTypeStyleName(t core.DiagramType) klimt.SName {
	switch t {
	case core.Sequence:
		return klimt.SNameSequenceDiagram
	case core.State:
		return klimt.SNameStateDiagram
	case core.Class:
		return klimt.SNameClassDiagram
	case core.Object:
		return klimt.SNameObjectDiagram
	case core.Activity:
		return klimt.SNameActivityDiagram
	case core.Description:
		return klimt.SNameComponentDiagram
	case core.MindMap:
		return klimt.SNameMindmapDiagram
	case core.WBS:
		return klimt.SNameWbsDiagram
	case core.Gantt, core.Chronology:
		return klimt.SNameGanttDiagram
	case core.Salt:
		return klimt.SNameSaltDiagram
	case core.YAML, core.HCL:
		return klimt.SNameYamlDiagram
	case core.JSON:
		return klimt.SNameJsonDiagram
	case core.Timing:
		return klimt.SNameTimingDiagram
	case core.NwDiag:
		return klimt.SNameNwdiagDiagram
	case core.EBNF:
		return klimt.SNameEbnf
	case core.Regex:
		return klimt.SNameRegex
	case core.Chart:
		return klimt.SNameChartDiagram
	case core.Packet:
		return klimt.SNamePacketdiagDiagram
	}
	return klimt.SNameActivityDiagram
}

// DecorateEntityImage draws a text block with a text above it and a text
// below it, upstream's svek.DecorateEntityImage.
type DecorateEntityImage struct {
	klimt.TextBlockDefaults
	original klimt.TextBlock

	group1      *klimt.UGroup
	horizontal1 geom.HorizontalAlignment
	text1       klimt.TextBlock

	group2      *klimt.UGroup
	horizontal2 geom.HorizontalAlignment
	text2       klimt.TextBlock

	deltaX float64
	deltaY float64
}

// DecorateEntityImageAddTop adds a text above original.
func DecorateEntityImageAddTop(group *klimt.UGroup, original, text klimt.TextBlock,
	horizontal geom.HorizontalAlignment) klimt.TextBlock {
	return newDecorateEntityImage(original, group, text, horizontal, nil, nil, klimt.HorizontalAlignmentNull)
}

// DecorateEntityImageAddBottom adds a text below original.
func DecorateEntityImageAddBottom(group *klimt.UGroup, original, text klimt.TextBlock,
	horizontal geom.HorizontalAlignment) klimt.TextBlock {
	return newDecorateEntityImage(original, nil, nil, klimt.HorizontalAlignmentNull, group, text, horizontal)
}

// DecorateEntityImageAdd adds a text above original for the TOP vertical
// alignment, below it otherwise.
func DecorateEntityImageAdd(group *klimt.UGroup, original, text klimt.TextBlock, horizontal geom.HorizontalAlignment,
	verticalAlignment geom.VerticalAlignment) klimt.TextBlock {
	if verticalAlignment == geom.VerticalAlignmentTop {
		return DecorateEntityImageAddTop(group, original, text, horizontal)
	}
	return DecorateEntityImageAddBottom(group, original, text, horizontal)
}

// DecorateEntityImageAddTopAndBottom adds texts above and below original;
// a nil text is absent.
func DecorateEntityImageAddTopAndBottom(original klimt.TextBlock, group1 *klimt.UGroup, text1 klimt.TextBlock,
	horizontal1 geom.HorizontalAlignment, group2 *klimt.UGroup, text2 klimt.TextBlock,
	horizontal2 geom.HorizontalAlignment) klimt.TextBlock {
	return newDecorateEntityImage(original, group1, text1, horizontal1, group2, text2, horizontal2)
}

func newDecorateEntityImage(original klimt.TextBlock, group1 *klimt.UGroup, text1 klimt.TextBlock,
	horizontal1 geom.HorizontalAlignment, group2 *klimt.UGroup, text2 klimt.TextBlock,
	horizontal2 geom.HorizontalAlignment) *DecorateEntityImage {
	if original == nil {
		ThrowNPE("")
	}
	return &DecorateEntityImage{original: original, group1: group1, horizontal1: horizontal1, text1: text1,
		group2: group2, horizontal2: horizontal2, text2: text2}
}

// DrawU draws the texts and the original.
func (d *DecorateEntityImage) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	dimOriginal := d.original.CalculateDimension(stringBounder)
	dimText1 := textDim(d.text1, stringBounder)
	dimText2 := textDim(d.text2, stringBounder)
	dimTotal := d.CalculateDimension(stringBounder)

	yImage := dimText1.Height()
	yText2 := yImage + dimOriginal.Height()

	xImage := (dimTotal.Width() - dimOriginal.Width()) / 2

	if d.text1 != nil {
		xText1 := textX(dimText1, dimTotal, d.horizontal1)
		if d.group1 != nil {
			ug.StartGroup(d.group1)
		}
		d.text1.DrawU(ug.Apply(klimt.UTranslateDx(xText1)))
		if d.group1 != nil {
			ug.CloseGroup()
		}
	}
	d.original.DrawU(ug.Apply(klimt.NewUTranslate(xImage, yImage)))
	d.deltaX = xImage
	d.deltaY = yImage
	if d.text2 != nil {
		xText2 := textX(dimText2, dimTotal, d.horizontal2)
		if d.group2 != nil {
			ug.StartGroup(d.group2)
		}
		d.text2.DrawU(ug.Apply(klimt.NewUTranslate(xText2, yText2)))
		if d.group2 != nil {
			ug.CloseGroup()
		}
	}
}

func textDim(text klimt.TextBlock, stringBounder font.StringBounder) geom.XDimension2D {
	if text == nil {
		return geom.NewXDimension2D(0, 0)
	}
	return text.CalculateDimension(stringBounder)
}

func textX(dimText, dimTotal geom.XDimension2D, h geom.HorizontalAlignment) float64 {
	switch h {
	case geom.HorizontalAlignmentCenter:
		return (dimTotal.Width() - dimText.Width()) / 2
	case geom.HorizontalAlignmentLeft:
		return 0
	case geom.HorizontalAlignmentRight:
		return dimTotal.Width() - dimText.Width()
	}
	Throw(IllegalStateException, "")
	return 0
}

// Backcolor returns the background of the original.
func (d *DecorateEntityImage) Backcolor() color.HColor { return d.original.Backcolor() }

// CalculateDimension returns the size of the texts and the original
// stacked.
func (d *DecorateEntityImage) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	dimOriginal := d.original.CalculateDimension(stringBounder)
	dim1 := textDim(d.text1, stringBounder)
	dim2 := textDim(d.text2, stringBounder)
	dimText := dim1.MergeTB(dim2)
	return dimOriginal.MergeTB(dimText)
}

// MinMax returns the bounds of the dimension.
func (d *DecorateEntityImage) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return geom.MinMaxFromDim(d.CalculateDimension(stringBounder))
}

// DeltaX returns the horizontal offset of the original after drawing.
func (d *DecorateEntityImage) DeltaX() float64 {
	if o, ok := d.original.(*DecorateEntityImage); ok {
		return d.deltaX + o.deltaX
	}
	return d.deltaX
}

// DeltaY returns the vertical offset of the original after drawing.
func (d *DecorateEntityImage) DeltaY() float64 {
	if o, ok := d.original.(*DecorateEntityImage); ok {
		return d.deltaY + o.deltaY
	}
	return d.deltaY
}
