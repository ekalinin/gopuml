// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/Swimlanes.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/LaneDivider.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"math"
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Swimlanes holds the lanes and the instruction tree of a diagram while it
// is parsed: the root list, the instruction that receives the next ones,
// the current lane and the arrow that leads to the next instruction. It is
// also the text block that lays out and draws the diagram.
type Swimlanes struct {
	klimt.TextBlockDefaults

	skinParam *klimt.SkinParam
	pragma    *klimt.Pragma

	swimlanesRaw     []*Swimlane
	swimlanesSpecial []*Swimlane
	dividers         []*LaneDivider
	currentSwimlane  *Swimlane

	root               Instruction
	currentInstruction Instruction

	nextLinkRenderer *LinkRendering
	style            *klimt.Style

	sizeComputed bool
	cachedMinMax *geom.MinMax
}

var _ klimt.TextBlock = (*Swimlanes)(nil)

// NewSwimlanes returns the lanes of a diagram, with an empty root list.
func NewSwimlanes(skinParam *klimt.SkinParam, pragma *klimt.Pragma) *Swimlanes {
	root := InstructionListEmpty()
	return &Swimlanes{skinParam: skinParam, pragma: pragma, root: root, currentInstruction: root,
		nextLinkRenderer: LinkRenderingNone()}
}

func (s *Swimlanes) swimlanes() []*Swimlane { return s.swimlanesRaw }

func (s *Swimlanes) swimlanesSpecialList() []*Swimlane {
	if len(s.swimlanesSpecial) == 0 {
		s.swimlanesSpecial = append(s.swimlanesSpecial, s.swimlanesRaw...)
		last := NewSwimlane("", math.MaxInt32, s.pragma)
		last.SetMinMax(geom.MinMaxEmpty(true))
		s.swimlanesSpecial = append(s.swimlanesSpecial, last)
	}
	return s.swimlanesSpecial
}

// StyleSignature returns the style signature of the lanes.
func (s *Swimlanes) StyleSignature() klimt.StyleSignatureBasic {
	return klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameSwimlane)
}

func (s *Swimlanes) getStyle() *klimt.Style {
	if s.style == nil {
		s.style = s.StyleSignature().MergedStyle(s.skinParam.CurrentStyleBuilder())
	}
	return s.style
}

// ftileFactory is upstream's getFtileFactory: the factory chain.
func (s *Swimlanes) ftileFactory(stringBounder font.StringBounder) FtileFactory {
	var factory FtileFactory = NewVCompactFactory(s.skinParam, stringBounder)
	factory = NewFtileFactoryDelegatorAddUrl(factory)
	factory = NewFtileFactoryDelegatorAssembly(factory)
	factory = NewFtileFactoryDelegatorIf(factory, s.pragma)
	factory = NewFtileFactoryDelegatorSwitch(factory)
	factory = NewFtileFactoryDelegatorWhile(factory)
	factory = NewFtileFactoryDelegatorRepeat(factory)
	factory = NewFtileFactoryDelegatorCreateParallel(factory)
	factory = NewFtileFactoryDelegatorAddNote(factory)
	factory = NewFtileFactoryDelegatorCreateGroup(factory)
	return factory
}

// List returns the lanes in order of creation.
func (s *Swimlanes) List() []*Swimlane { return s.swimlanesRaw }

// Swimlane makes the lane name current, creating it when it does not
// exist; color, when not nil, becomes its background and label, when not
// null, its title.
func (s *Swimlanes) Swimlane(name string, c color.HColor, label *klimt.Display) {
	s.currentSwimlane = s.getOrCreate(name)
	if c != nil {
		s.currentSwimlane.SetSpecificColorTOBEREMOVED(color.TypeBack, c)
	}
	if !klimt.DisplayIsNull(label) {
		s.currentSwimlane.SetDisplay(label)
	}
}

func (s *Swimlanes) getOrCreate(name string) *Swimlane {
	for _, sw := range s.swimlanesRaw {
		if sw.Name() == name {
			return sw
		}
	}
	result := NewSwimlane(name, len(s.swimlanesRaw), s.pragma)
	s.swimlanesRaw = append(s.swimlanesRaw, result)
	return result
}

// swimlanesCross draws the connections between tiles of different lanes,
// upstream's inner class Cross.
type swimlanesCross struct {
	klimt.UGraphicDelegator
}

func newSwimlanesCross(ug klimt.UGraphic) *swimlanesCross {
	return &swimlanesCross{UGraphicDelegator: klimt.NewUGraphicDelegator(ug)}
}

func (c *swimlanesCross) Draw(shape klimt.UShape) {
	switch s := shape.(type) {
	case Ftile:
		s.DrawU(c)
	case Connection:
		tile1 := s.Ftile1()
		tile2 := s.Ftile2()
		if tile1 == nil || tile2 == nil {
			return
		}
		if tile1.SwimlaneOut() != tile2.SwimlaneIn() {
			NewConnectionCross(s).DrawU(c.Ug())
		}
	}
}

func (c *swimlanesCross) Apply(change klimt.UChange) klimt.UGraphic {
	return newSwimlanesCross(c.Ug().Apply(change))
}

func (s *Swimlanes) ensureSizeComputed(stringBounder font.StringBounder) {
	if !s.sizeComputed {
		s.sizeComputed = true
		if len(s.swimlanes()) > 1 {
			full := s.root.CreateFtile(s.ftileFactory(stringBounder))
			s.computeSizeInternal(stringBounder, full)
		}
	}
}

// DrawU implements klimt.TextBlock.
func (s *Swimlanes) DrawU(ug klimt.UGraphic) {
	s.ensureSizeComputed(ug.StringBounder())
	var full klimt.TextBlock = s.root.CreateFtile(s.ftileFactory(ug.StringBounder()))
	style := klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
		klimt.SNameGoto).MergedStyle(s.skinParam.CurrentStyleBuilder())
	gotoColor := style.Value(klimt.PNameLineColor).AsColor(s.skinParam.IHtmlColorSet())
	isDebug := isDebugOption(s.skinParam)

	ug = NewUGraphicForSnake(ug)
	if len(s.swimlanes()) > 1 {
		s.drawWhenSwimlanes(ug, full)
	} else {
		// BUG42
		full = newTextBlockInterceptorUDrawable(full, gotoColor, isDebug)
		full.DrawU(ug)
		ug.FlushUg()
	}
}

func (s *Swimlanes) title(swimlane *Swimlane) klimt.TextBlock {
	horizontalAlignment := geom.HorizontalAlignmentLeft
	fontConfiguration := s.getStyle().FontConfiguration(s.skinParam.IHtmlColorSet(), nil)
	wrap := s.wrap()
	if wrap.IsAuto() {
		wrap = klimt.NewLineBreakStrategy(strconv.Itoa(int(swimlane.ActualWidth())))
	}
	return swimlane.Display().Create9(fontConfiguration, horizontalAlignment, s.skinParam, wrap)
}

// wrap is upstream's getWrap. Upstream falls back to the wrap width of the
// style when the skin parameter is LineBreakStrategy.NONE, compared by
// identity: the skin parameter is always a new object, so the fallback
// never happens.
func (s *Swimlanes) wrap() klimt.LineBreakStrategy { return s.skinParam.SwimlaneWrapTitleWidth() }

func (s *Swimlanes) titleHeightTranslate(stringBounder font.StringBounder) klimt.UTranslate {
	titlesHeight := s.titlesHeight(stringBounder)
	if titlesHeight > 0 {
		return klimt.UTranslateDy(titlesHeight + 5)
	}
	return klimt.UTranslateDy(0)
}

func (s *Swimlanes) titlesHeight(stringBounder font.StringBounder) float64 {
	titlesHeight := 0.0
	for _, swimlane := range s.swimlanes() {
		swTitle := s.title(swimlane)
		titlesHeight = math.Max(titlesHeight, swTitle.CalculateDimension(stringBounder).Height())
	}
	return titlesHeight
}

func (s *Swimlanes) drawWhenSwimlanes(ug klimt.UGraphic, full klimt.TextBlock) {
	stringBounder := ug.StringBounder()
	titleHeightTranslate := s.titleHeightTranslate(stringBounder)

	s.drawTitlesBackground(ug)

	dimensionFull := full.CalculateDimension(stringBounder)
	for i, swimlane := range s.swimlanesSpecialList() {
		divider1 := s.dividers[i]
		xpos := swimlane.Translate().Dx() + swimlane.MinMax().MinX()
		back := swimlane.Colors().Color(color.TypeBack)
		if back != nil && !back.IsTransparent() {
			divider2 := s.dividers[i+1]
			background := ug.Apply(klimt.Bg(back)).Apply(back).Apply(klimt.UTranslateDx(xpos - divider1.X2()))
			width := swimlane.ActualWidth() + divider1.X2() + divider2.X1()
			height := dimensionFull.Height() + titleHeightTranslate.Dy()
			background.Draw(klimt.URectangleBuild(width, height).IgnoreForCompressionOnX().IgnoreForCompressionOnY())
		}
		full.DrawU(newUGraphicInterceptorOneSwimlane(ug, swimlane, s.swimlanes()).Apply(swimlane.Translate()).
			Apply(s.titleHeightTranslate(stringBounder)))
		dividerWith := divider1.CalculateDimension(stringBounder).Width()
		divider1.DrawU(ug.Apply(klimt.UTranslateDx(xpos - dividerWith)))
	}

	cross := newSwimlanesCross(ug.Apply(s.titleHeightTranslate(stringBounder)))
	full.DrawU(cross)
	cross.FlushUg()

	s.drawTitles(ug)
}

func (s *Swimlanes) drawTitlesBackground(ug klimt.UGraphic) {
	c := s.getStyle().Value(klimt.PNameBackGroundColor).AsColor(s.skinParam.IHtmlColorSet())
	if c != nil {
		titleHeight := s.titlesHeight(ug.StringBounder())
		special := s.swimlanesSpecialList()
		fullWidth := special[len(special)-1].Translate().Dx() - 2*5 - 1
		back := klimt.URectangleBuild(fullWidth, titleHeight).IgnoreForCompressionOnX().IgnoreForCompressionOnY()
		ug.Apply(klimt.UTranslateDx(5)).Apply(klimt.Bg(c)).Apply(c).Draw(back)
	}
}

func (s *Swimlanes) drawTitles(ug klimt.UGraphic) {
	for _, swimlane := range s.swimlanes() {
		swTitle := s.title(swimlane)
		x2 := swimlane.Translate().Dx() + swimlane.MinMax().MinX()
		centeredText := klimt.NewCenteredText(swTitle, s.widthWithoutTitle(swimlane))
		ug.Apply(klimt.UTranslateDx(x2)).Draw(centeredText)
	}
}

func (s *Swimlanes) computeDrawingWidths(stringBounder font.StringBounder, full klimt.TextBlock) {
	finders := make(map[*Swimlane]*klimt.LimitFinder, len(s.swimlanes()))
	swimlaneToUg := newSwimlaneUgMap()
	for _, swimlane := range s.swimlanes() {
		limitFinder := klimt.LimitFinderCreate(stringBounder, false)
		finders[swimlane] = limitFinder
		swimlaneToUg.put(swimlane, NewUGraphicForSnake(limitFinder))
	}
	interceptor := newUGraphicInterceptorAllSwimlanes(swimlaneToUg, s.swimlanes(), stringBounder)
	full.DrawU(interceptor)
	interceptor.FlushUg()
	for _, swimlane := range s.swimlanes() {
		swimlane.SetMinMax(finders[swimlane].MinMax())
	}
}

func (s *Swimlanes) computeSizeInternal(stringBounder font.StringBounder, full klimt.TextBlock) {
	s.computeDrawingWidths(stringBounder, full)

	minWidth := float64(s.skinParam.SwimlaneWidth())
	if minWidth == klimt.ISkinParamSwimlaneWidthSame {
		for _, swimlane := range s.swimlanes() {
			minWidth = math.Max(minWidth, s.widthWithoutTitle(swimlane))
		}
	}
	for _, swimlane := range s.swimlanesSpecialList() {
		swimlaneActualWidth := math.Max(minWidth, s.widthWithoutTitle(swimlane))
		swimlane.SetWidth(swimlaneActualWidth)
	}

	titleHeightTranslate := s.titleHeightTranslate(stringBounder)
	dimensionFull := full.CalculateDimension(stringBounder)

	s.dividers = s.dividers[:0]
	xpos := 0.0
	special := s.swimlanesSpecialList()
	for i, swimlane := range special {
		x1 := s.HalfMissingSpace(stringBounder, i, minWidth)
		x2 := s.HalfMissingSpace(stringBounder, i+1, minWidth)
		laneDivider := NewLaneDivider(s.skinParam, x1, x2, dimensionFull.Height()+titleHeightTranslate.Dy())
		s.dividers = append(s.dividers, laneDivider)
		xx := xpos + laneDivider.Width() - swimlane.MinMax().MinX() +
			(swimlane.ActualWidth()-s.widthWithoutTitle(swimlane))/2.0
		swimlane.SetTranslate(klimt.UTranslateDx(xx))
		xpos += swimlane.ActualWidth() + laneDivider.Width()
	}
}

// HalfMissingSpace returns the space a title wider than its lane adds on
// each side of divider i.
func (s *Swimlanes) HalfMissingSpace(stringBounder font.StringBounder, i int, minWidth float64) float64 {
	special := s.swimlanesSpecialList()
	if i == 0 || i > len(special) {
		return 5
	}
	swimlane := special[i-1]
	swimlaneActualWidth := math.Max(minWidth, s.widthWithoutTitle(swimlane))
	titleWidth := s.title(swimlane).CalculateDimension(stringBounder).Width()
	if titleWidth <= swimlaneActualWidth {
		return 5
	}
	return math.Max(5, 5+(titleWidth-swimlaneActualWidth)/2)
}

func (s *Swimlanes) widthWithoutTitle(swimlane *Swimlane) float64 { return swimlane.MinMax().Width() }

// CalculateDimension implements klimt.TextBlock.
func (s *Swimlanes) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	s.ensureSizeComputed(stringBounder)
	return s.MinMax(stringBounder).Dimension()
}

// Root returns the root instruction list.
func (s *Swimlanes) Root() Instruction { return s.root }

// Current returns the instruction that receives the next instructions.
func (s *Swimlanes) Current() Instruction { return s.currentInstruction }

// SetCurrent sets the instruction that receives the next instructions.
func (s *Swimlanes) SetCurrent(current Instruction) { s.currentInstruction = current }

// NextLinkRenderer returns the arrow that leads to the next instruction.
func (s *Swimlanes) NextLinkRenderer() *LinkRendering { return s.nextLinkRenderer }

// SetNextLinkRenderer sets the arrow that leads to the next instruction; a
// nil link panics, as upstream's requireNonNull.
func (s *Swimlanes) SetNextLinkRenderer(link *LinkRendering) {
	if link == nil {
		uml.ThrowNPE("")
	}
	s.nextLinkRenderer = link
}

// CurrentSwimlane returns the current lane, nil before the first one.
func (s *Swimlanes) CurrentSwimlane() *Swimlane { return s.currentSwimlane }

// MinMax implements klimt.TextBlock: the bounds of the drawing, cached.
func (s *Swimlanes) MinMax(stringBounder font.StringBounder) geom.MinMax {
	if s.cachedMinMax == nil {
		m := klimt.TextBlockUtilsMinMax(s, stringBounder, false)
		s.cachedMinMax = &m
	}
	return *s.cachedMinMax
}

// CurrentStyleBuilder returns the style builder of the diagram.
func (s *Swimlanes) CurrentStyleBuilder() *klimt.StyleBuilder {
	return s.skinParam.CurrentStyleBuilder()
}

// LaneDivider is the vertical line between two lanes, with the space on
// each side.
type LaneDivider struct {
	klimt.TextBlockDefaults
	skinParam klimt.ISkinParam
	x1, x2    float64
	height    float64
	style     *klimt.Style
}

// NewLaneDivider returns a divider with x1 on the left of the line and x2
// on its right.
func NewLaneDivider(skinParam klimt.ISkinParam, x1, x2, height float64) *LaneDivider {
	return &LaneDivider{skinParam: skinParam, x1: x1, x2: x2, height: height}
}

func (d *LaneDivider) getStyle() *klimt.Style {
	if d.style == nil {
		d.style = klimt.StyleSignatureBasicOf(klimt.SNameRoot, klimt.SNameElement, klimt.SNameActivityDiagram,
			klimt.SNameSwimlane).MergedStyle(d.skinParam.CurrentStyleBuilder())
	}
	return d.style
}

// CalculateDimension implements klimt.TextBlock.
func (d *LaneDivider) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(d.x1+d.x2, d.height)
}

// DrawU implements klimt.TextBlock.
func (d *LaneDivider) DrawU(ug klimt.UGraphic) {
	ug.Draw(klimt.NewUEmpty(d.x1+d.x2, 1))
	c := d.getStyle().Value(klimt.PNameLineColor).AsColor(d.skinParam.IHtmlColorSet())
	thickness := d.getStyle().Stroke()
	ug.Apply(klimt.UTranslateDx(d.x1)).Apply(thickness).Apply(c).Draw(klimt.ULineV(d.height))
}

// Width returns the width of the divider.
func (d *LaneDivider) Width() float64 { return d.x1 + d.x2 }

// X1 returns the space on the left of the line.
func (d *LaneDivider) X1() float64 { return d.x1 }

// X2 returns the space on the right of the line.
func (d *LaneDivider) X2() float64 { return d.x2 }
