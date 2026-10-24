// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/SheetBlock1.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/SheetBlock2.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/Fission.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/Sea.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"reflect"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// SheetBlock1 lays out a sheet: it breaks the stripes at the maximum width,
// aligns the atoms of each line on their baseline and stacks the lines. It
// is a TextBlock, an Atom and a Stencil (its own outline).
type SheetBlock1 struct {
	AtomBase
	sheet        *Sheet
	stripes      []Stripe
	widths       []stripeWidth
	positions    map[Atom]Position
	minMax       geom.MinMax
	maxWidth     LineBreakStrategy
	padding      ClockwiseTopRightBottomLeft
	marginX1     float64
	marginX2     float64
	lastCaller   reflect.Type
	minimumWidth float64
}

// stripeWidth is an entry of upstream's widths map (a LinkedHashMap keyed
// by stripe identity).
type stripeWidth struct {
	stripe Stripe
	width  float64
}

var (
	_ Atom    = (*SheetBlock1)(nil)
	_ Stencil = (*SheetBlock1)(nil)
)

// NewSheetBlock1 returns the layout of a sheet with a maximum width, a
// padding and horizontal margins (which only extend its stencil).
func NewSheetBlock1(sheet *Sheet, maxWidth LineBreakStrategy, padding ClockwiseTopRightBottomLeft,
	marginX1, marginX2 float64) *SheetBlock1 {
	return &SheetBlock1{sheet: sheet, maxWidth: maxWidth, padding: padding, marginX1: marginX1, marginX2: marginX2}
}

// NewSheetBlock1Padding is NewSheetBlock1 without margins.
func NewSheetBlock1Padding(sheet *Sheet, maxWidth LineBreakStrategy, padding ClockwiseTopRightBottomLeft) *SheetBlock1 {
	return NewSheetBlock1(sheet, maxWidth, padding, 0, 0)
}

// NewSheetBlock1Same is NewSheetBlock1 with the same padding on each side.
func NewSheetBlock1Same(sheet *Sheet, maxWidth LineBreakStrategy, padding, marginX1, marginX2 float64) *SheetBlock1 {
	return NewSheetBlock1(sheet, maxWidth, ClockwiseSame(padding), marginX1, marginX2)
}

// NewSheetBlock1Style is NewSheetBlock1 with the padding and minimum width
// of a style.
func NewSheetBlock1Style(sheet *Sheet, maxWidth LineBreakStrategy, style *Style) *SheetBlock1 {
	result := NewSheetBlock1(sheet, maxWidth, style.Padding(), 0, 0)
	result.minimumWidth = style.Value(PNameMinimumWidth).AsDouble()
	return result
}

// String is upstream's toString: the sheet.
func (b *SheetBlock1) String() string { return b.sheet.String() }

// CellAlignment returns the alignment of a single-line sheet (a table
// cell), LEFT otherwise. It is only valid after the layout (a dimension or a
// drawing), as upstream.
func (b *SheetBlock1) CellAlignment() geom.HorizontalAlignment {
	if len(b.stripes) != 1 {
		return geom.HorizontalAlignmentLeft
	}
	if simple, ok := b.stripes[0].(*StripeSimple); ok {
		return simple.CellAlignment()
	}
	return geom.HorizontalAlignmentLeft
}

// HorizontalAlignment returns the alignment of the sheet.
func (b *SheetBlock1) HorizontalAlignment() geom.HorizontalAlignment {
	return b.sheet.HorizontalAlignment()
}

func (b *SheetBlock1) initMap(stringBounder font.StringBounder) {
	currentCaller := reflect.TypeOf(stringBounder)
	if b.lastCaller == currentCaller {
		return
	}
	b.lastCaller = currentCaller
	b.stripes = nil
	for _, stripe := range b.sheet.Stripes() {
		b.stripes = append(b.stripes, newFission(stripe, b.maxWidth).splitted(stringBounder)...)
	}
	b.positions = map[Atom]Position{}
	b.widths = nil
	b.minMax = geom.MinMaxEmpty(true)
	y := 0.0
	for _, stripe := range b.stripes {
		if len(stripe.Atoms()) == 0 {
			continue
		}
		sea := newSea(stringBounder)
		for _, atom := range stripe.Atoms() {
			sea.add(atom)
		}
		sea.doAlign()
		if stringBounder.MatchesProperty("TIKZ") {
			sea.doAlignTikz()
		}
		sea.translateMinYto(y)
		b.putWidth(stripe, sea.width())
		b.minMax = sea.update(b.minMax)
		height := sea.height()
		y += height
		if stringBounder.MatchesProperty("TIKZ") {
			sea.doAlignTikzBaseline()
		}
		sea.exportAllPositions(b.positions)
	}
	maxWidth := 0.0
	for _, w := range b.widths {
		if w.width > maxWidth {
			maxWidth = w.width
		}
	}
	for _, w := range b.widths {
		diff := maxWidth - w.width
		if diff > 0 {
			coef := b.coefOfStripe(w.stripe)
			if coef > 0 {
				for _, atom := range w.stripe.Atoms() {
					b.positions[atom] = b.positions[atom].TranslateX(diff / float64(coef))
				}
			}
		}
	}
}

func (b *SheetBlock1) putWidth(stripe Stripe, width float64) {
	for i, w := range b.widths {
		if w.stripe == stripe {
			b.widths[i].width = width
			return
		}
	}
	b.widths = append(b.widths, stripeWidth{stripe, width})
}

func (b *SheetBlock1) coefOfStripe(key Stripe) int {
	if simple, ok := key.(*StripeSimple); ok {
		return creoleCoef(simple.CellAlignment())
	}
	return creoleCoef(b.sheet.HorizontalAlignment())
}

func creoleCoef(alignment geom.HorizontalAlignment) int {
	switch alignment {
	case geom.HorizontalAlignmentCenter:
		return 2
	case geom.HorizontalAlignmentRight:
		return 1
	}
	return 0
}

// CalculateDimension returns the size of the laid out sheet. As upstream,
// the vertical padding is added to the width too.
func (b *SheetBlock1) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		b.initMap(sb)
		return b.minMax.Dimension().Delta(b.padding.Bottom() + b.padding.Top())
	})
}

// InnerPosition returns nil.
func (b *SheetBlock1) InnerPosition(string, font.StringBounder) *geom.XRectangle2D { return nil }

// DrawU draws the atoms at their positions, after the left and top padding.
func (b *SheetBlock1) DrawU(ug UGraphic) {
	b.initMap(ug.StringBounder())
	if b.padding.Left() > 0 || b.padding.Top() > 0 {
		ug = ug.Apply(NewUTranslate(b.padding.Left(), b.padding.Top()))
	}
	for _, stripe := range b.stripes {
		for _, atom := range stripe.Atoms() {
			position, ok := b.positions[atom]
			if !ok {
				throw(nullPointer, "")
			}
			atom.DrawU(position.Translate(ug))
		}
	}
}

// StartingAltitude returns 0.
func (b *SheetBlock1) StartingAltitude(font.StringBounder) float64 { return 0 }

// StartingX returns the left margin, negated.
func (b *SheetBlock1) StartingX(font.StringBounder, float64) float64 { return -b.marginX1 }

// EndingX returns the width plus the right margin.
func (b *SheetBlock1) EndingX(stringBounder font.StringBounder, _ float64) float64 {
	return b.CalculateDimension(stringBounder).Width() + b.marginX2
}

// MinimumWidth returns the minimum width of the style, 0 without style.
func (b *SheetBlock1) MinimumWidth() float64 { return b.minimumWidth }

// Neutrons panics, as upstream throws UnsupportedOperationException.
func (b *SheetBlock1) Neutrons() []*Neutron {
	throw(unsupportedOperation, "")
	return nil
}

// SheetBlock2 draws a SheetBlock1 inside a stencil, so that its horizontal
// lines span the stencil, and centers or right-aligns it within the minimum
// width of its style.
type SheetBlock2 struct {
	TextBlockDefaults
	block         *SheetBlock1
	defaultStroke UStroke
	stencil       Stencil
}

var _ Atom = (*SheetBlock2)(nil)

// NewSheetBlock2 returns block drawn within stencil; the horizontal lines
// use defaultStroke. A nil stencil panics, as upstream's
// Objects.requireNonNull.
func NewSheetBlock2(block *SheetBlock1, stencil Stencil, defaultStroke UStroke) *SheetBlock2 {
	if stencil == nil {
		throw(nullPointer, "")
	}
	return &SheetBlock2{block: block, stencil: stencil, defaultStroke: defaultStroke}
}

// EnlargeMe returns the block with its stencil extended by delta1 on the
// left and delta2 on the right.
func (b *SheetBlock2) EnlargeMe(delta1, delta2 float64) *SheetBlock2 {
	stencil := b.stencil
	newStencil := StencilFunc{
		Starting: func(stringBounder font.StringBounder, y float64) float64 {
			return stencil.StartingX(stringBounder, y) - delta1
		},
		Ending: func(stringBounder font.StringBounder, y float64) float64 {
			return stencil.EndingX(stringBounder, y) + delta2
		},
	}
	return NewSheetBlock2(b.block, newStencil, b.defaultStroke)
}

// String is upstream's toString: the sheet.
func (b *SheetBlock2) String() string { return b.block.String() }

// CalculateDimension returns the size of the block.
func (b *SheetBlock2) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.block.CalculateDimension(stringBounder)
}

// DrawU draws the block.
func (b *SheetBlock2) DrawU(ug UGraphic) {
	ug = UGraphicStencilCreate(ug, b.stencil, b.defaultStroke)
	alignment := b.block.HorizontalAlignment()
	if alignment == geom.HorizontalAlignmentCenter && b.block.MinimumWidth() > 0 {
		width := b.CalculateDimension(ug.StringBounder()).Width()
		dx := (b.block.MinimumWidth() - width) / 2
		ug = ug.Apply(UTranslateDx(dx))
	} else if alignment == geom.HorizontalAlignmentRight && b.block.MinimumWidth() > 0 {
		width := b.CalculateDimension(ug.StringBounder()).Width()
		dx := b.block.MinimumWidth() - width
		ug = ug.Apply(UTranslateDx(dx))
	}
	b.block.DrawU(ug)
}

// StartingAltitude returns 0.
func (b *SheetBlock2) StartingAltitude(font.StringBounder) float64 { return 0 }

// InnerPosition returns nil, as the block.
func (b *SheetBlock2) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return b.block.InnerPosition(member, stringBounder)
}

// Ports returns no ports.
func (b *SheetBlock2) Ports(font.StringBounder) *Ports { return NewPorts() }

// Neutrons panics, as upstream throws UnsupportedOperationException.
func (b *SheetBlock2) Neutrons() []*Neutron {
	throw(unsupportedOperation, "")
	return nil
}

// fission breaks a stripe into lines no wider than a maximum width, at the
// break opportunities (ZWSP neutrons) of its atoms.
type fission struct {
	stripe   Stripe
	maxWidth LineBreakStrategy
}

func newFission(stripe Stripe, maxWidth LineBreakStrategy) *fission {
	return &fission{stripe: stripe, maxWidth: maxWidth}
}

func (f *fission) splitted(stringBounder font.StringBounder) []Stripe {
	valueMaxWidth := math.Abs(f.maxWidth.MaxWidth())
	if valueMaxWidth == 0 {
		return []Stripe{f.stripe}
	}
	line := newStripeSimpleInternal(false, stringBounder, f.stripe.LHeader())
	result := []*stripeSimpleInternal{line}
	var all []*Neutron
	for _, atom := range f.noHeader() {
		all = append(all, atom.Neutrons()...)
	}
	// Upstream fails with a NullPointerException when the stripe has no
	// neutron (a list item without text, such as "* "); the line then only
	// holds the header.
	if len(all) > 0 && all[len(all)-1].Type() != NeutronTypeZwspSeparator {
		all = append(all, NeutronZwspSeparator())
	}
	for len(all) > 0 {
		current := all[0]
		all = all[1:]
		if current.Type() == NeutronTypeZwspSeparator && line.Width() > valueMaxWidth {
			removed := line.slightyShorten()
			all = append(append(removed, current), all...)
			line = newStripeSimpleInternal(true, stringBounder, blankAtomOf(f.stripe.LHeader()))
			result = append(result, line)
		} else {
			line.addNeutron(current)
		}
	}
	for _, l := range result {
		l.removeFinalSpaces()
	}
	for len(result) > 1 && result[len(result)-1].isWhite() {
		result = result[:len(result)-1]
	}
	stripes := make([]Stripe, len(result))
	for i, l := range result {
		stripes[i] = l
	}
	return stripes
}

func (f *fission) noHeader() []Atom {
	atoms := f.stripe.Atoms()
	if f.stripe.LHeader() == nil {
		return atoms
	}
	return atoms[1:]
}

// stripeSimpleInternal is upstream's Fission.StripeSimpleInternal: a line
// made by the fission.
type stripeSimpleInternal struct {
	removeInitialSpaces bool
	header              Atom
	neutrons            []*Neutron
	stringBounder       font.StringBounder
	width               float64
}

func newStripeSimpleInternal(removeInitialSpaces bool, stringBounder font.StringBounder,
	header Atom) *stripeSimpleInternal {
	s := &stripeSimpleInternal{removeInitialSpaces: removeInitialSpaces, stringBounder: stringBounder,
		header: header}
	if header != nil {
		s.width += header.CalculateDimension(stringBounder).Width()
	}
	return s
}

// Width returns the width of the line.
func (s *stripeSimpleInternal) Width() float64 {
	if s.width == -1 {
		throw(illegalState, "")
	}
	return s.width
}

// String is upstream's toString.
func (s *stripeSimpleInternal) String() string {
	parts := make([]string, len(s.neutrons))
	for i, n := range s.neutrons {
		parts[i] = n.String()
	}
	list := "[" + strings.Join(parts, ", ") + "]"
	if s.header != nil {
		return textBlockString(s.header) + " " + list
	}
	return list
}

func (s *stripeSimpleInternal) slightyShorten() []*Neutron {
	if len(s.neutrons) == 0 {
		throw(illegalState, "")
	}
	lastZwsp := s.lastZwsp()
	if lastZwsp == -1 {
		return []*Neutron{}
	}
	s.width = -1
	removed := append([]*Neutron(nil), s.neutrons[lastZwsp:]...)
	s.neutrons = s.neutrons[:lastZwsp]
	return removed
}

func (s *stripeSimpleInternal) isWhite() bool {
	for _, n := range s.neutrons {
		if n.Type() != NeutronTypeZwspSeparator && n.Type() != NeutronTypeWhitespace {
			return false
		}
	}
	return true
}

func (s *stripeSimpleInternal) removeFinalSpaces() {
	for len(s.neutrons) > 0 && s.neutrons[0].Type() == NeutronTypeZwspSeparator {
		s.neutrons = s.neutrons[1:]
	}
	for len(s.neutrons) > 1 && (s.last().Type() == NeutronTypeWhitespace ||
		s.last().Type() == NeutronTypeZwspSeparator) {
		s.neutrons = s.neutrons[:len(s.neutrons)-1]
	}
}

func (s *stripeSimpleInternal) last() *Neutron { return s.neutrons[len(s.neutrons)-1] }

func (s *stripeSimpleInternal) lastZwsp() int {
	for i := len(s.neutrons) - 1; i >= 0; i-- {
		if s.neutrons[i].Type() == NeutronTypeZwspSeparator {
			return i
		}
	}
	return -1
}

// Atoms returns the header and the atoms of the neutrons, without the
// break opportunities (and without the leading spaces of a continuation
// line).
func (s *stripeSimpleInternal) Atoms() []Atom {
	var result []Atom
	if s.header != nil {
		result = append(result, s.header)
	}
	for _, n := range s.neutrons {
		if n.Type() != NeutronTypeZwspSeparator {
			if s.removeInitialSpaces && len(result) == 0 && n.Type() == NeutronTypeWhitespace {
				continue
			}
			result = append(result, n.AsAtom())
		}
	}
	return result
}

func (s *stripeSimpleInternal) addNeutron(neutron *Neutron) {
	if s.width == -1 {
		throw(illegalState, "")
	}
	if neutron.Type() == NeutronTypeZwspSeparator && len(s.neutrons) == 0 {
		return
	}
	if neutron.Type() == NeutronTypeZwspSeparator && len(s.neutrons) > 0 &&
		s.last().Type() == NeutronTypeZwspSeparator {
		return
	}
	if s.removeInitialSpaces && len(s.neutrons) == 0 && neutron.Type() == NeutronTypeWhitespace {
		return
	}
	s.neutrons = append(s.neutrons, neutron)
	s.width += neutron.Width(s.stringBounder)
}

// LHeader returns nil.
func (s *stripeSimpleInternal) LHeader() Atom { return nil }

// blankAtom is the anonymous atom of upstream's Fission.blank: the space of
// a header, drawn empty.
type blankAtom struct {
	AtomBase
	header Atom
}

func blankAtomOf(header Atom) Atom {
	if header == nil {
		return nil
	}
	return &blankAtom{header: header}
}

func (a *blankAtom) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.Dimension(stringBounder, a.header.CalculateDimension)
}

func (a *blankAtom) StartingAltitude(stringBounder font.StringBounder) float64 {
	return a.header.StartingAltitude(stringBounder)
}

func (a *blankAtom) DrawU(UGraphic) {}

func (a *blankAtom) Neutrons() []*Neutron { return atomNeutrons(a) }

// sea lays out the atoms of one line: side by side, aligned on their
// baseline.
type sea struct {
	currentX      float64
	atoms         []Atom // the keys of positions, in insertion order
	positions     map[Atom]Position
	stringBounder font.StringBounder
}

func newSea(stringBounder font.StringBounder) *sea {
	if stringBounder == nil {
		throw(nullPointer, "")
	}
	return &sea{positions: map[Atom]Position{}, stringBounder: stringBounder}
}

func (s *sea) put(atom Atom, position Position) {
	if _, ok := s.positions[atom]; !ok {
		s.atoms = append(s.atoms, atom)
	}
	s.positions[atom] = position
}

func (s *sea) add(atom Atom) {
	dim := atom.CalculateDimension(s.stringBounder)
	s.put(atom, NewPosition(s.currentX, 0, dim))
	s.currentX += dim.Width()
}

func (s *sea) doAlign() {
	for _, atom := range s.atoms {
		pos := s.positions[atom]
		height := atom.CalculateDimension(s.stringBounder).Height()
		s.positions[atom] = pos.TranslateY(-height + atom.StartingAltitude(s.stringBounder))
	}
}

func (s *sea) translateMinYto(newValue float64) {
	delta := newValue - s.minY()
	for _, atom := range s.atoms {
		s.positions[atom] = s.positions[atom].TranslateY(delta)
	}
}

func (s *sea) findFirstAtomText() *AtomText {
	for _, atom := range s.atoms {
		if atomText, ok := atom.(*AtomText); ok {
			if text.JavaTrim(atomText.Text()) == "" {
				continue
			}
			return atomText
		}
	}
	return nil
}

// doAlignTikz centers the non-text atoms vertically on the first text
// (#1628).
func (s *sea) doAlignTikz() {
	firstTextAtom := s.findFirstAtomText()
	if firstTextAtom == nil {
		return
	}
	firstTextPosition := s.positions[firstTextAtom]
	firstTextHeight := firstTextAtom.FontHeight(s.stringBounder)
	for _, atom := range s.atoms {
		if _, ok := atom.(*AtomText); ok {
			continue
		}
		position := s.positions[atom]
		targetY := firstTextPosition.MinY() - (position.Height()-firstTextHeight)/2
		if delta := targetY - position.MinY(); delta != 0 {
			s.positions[atom] = position.TranslateY(delta)
		}
	}
}

// doAlignTikzBaseline puts the texts on the same baseline (#1606).
func (s *sea) doAlignTikzBaseline() {
	firstTextAtom := s.findFirstAtomText()
	if firstTextAtom == nil {
		return
	}
	firstTextHeight := firstTextAtom.FontHeight(s.stringBounder)
	for _, atom := range s.atoms {
		atomText, ok := atom.(*AtomText)
		if !ok {
			continue
		}
		if delta := firstTextHeight - atomText.FontHeight(s.stringBounder); delta != 0 {
			s.positions[atom] = s.positions[atom].TranslateY(delta)
		}
	}
	if len(s.positions) == 1 {
		// the only atom text, and we have hard-coded the minimum height to
		// 10 (AtomText's dimension)
		if delta := (10 - firstTextHeight) / 2; delta > 0 {
			s.positions[firstTextAtom] = s.positions[firstTextAtom].TranslateY(delta)
		}
	}
}

func (s *sea) exportAllPositions(destination map[Atom]Position) {
	for _, atom := range s.atoms {
		destination[atom] = s.positions[atom]
	}
}

func (s *sea) minY() float64 {
	if len(s.positions) == 0 {
		throw(illegalState, "")
	}
	result := math.MaxFloat64
	for _, atom := range s.atoms {
		if pos := s.positions[atom]; result > pos.MinY() {
			result = pos.MinY()
		}
	}
	return result
}

func (s *sea) maxY() float64 {
	if len(s.positions) == 0 {
		throw(illegalState, "")
	}
	var atom Atom
	result := -math.MaxFloat64
	for _, a := range s.atoms {
		if pos := s.positions[a]; result < pos.MaxY() {
			atom = a
			result = pos.MaxY()
		}
	}
	if !s.stringBounder.MatchesProperty("TIKZ") {
		return result
	}
	// For TKIZ, make sure the strip has at least 1pt
	if atomText, ok := atom.(*AtomText); ok {
		// the delta in AtomText should be larger than 1 already
		height := s.stringBounder.CalculateDimension(atomText.FontConfiguration().Font(), atomText.Text()).Height
		delta := result - height
		return result + math.Max(1-delta, 0)
	}
	return result + 1
}

func (s *sea) height() float64 { return s.maxY() - s.minY() }

func (s *sea) update(minMax geom.MinMax) geom.MinMax {
	for _, atom := range s.atoms {
		minMax = s.positions[atom].Update(minMax)
	}
	return minMax
}

func (s *sea) width() float64 { return s.currentX }

// textBlockString returns the toString of a text block, for debug output.
func textBlockString(tb TextBlock) string {
	if st, ok := tb.(interface{ String() string }); ok {
		return st.String()
	}
	return reflect.TypeOf(tb).String()
}
