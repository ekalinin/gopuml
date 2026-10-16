// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/CreoleParser.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/CreoleStripeSimpleParser.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/StripeSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/StripeTable.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/StripeTree.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/StripeCode.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/StripeLatex.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/legacy/StripeRaw.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/EmbeddedDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"math"
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// creoleParser is upstream's legacy CreoleParser, the SheetBuilder of the
// skin parameters: it turns each line of a display into stripes.
type creoleParser struct {
	fontConfiguration   *FontConfiguration
	skinParam           ISkinSimple
	horizontalAlignment geom.HorizontalAlignment
	creoleMode          CreoleMode
	stereotype          *FontConfiguration
	cache               []creoleCacheEntry
}

// creoleCacheEntry is an entry of upstream's cache (a HashMap keyed by
// Display equality).
type creoleCacheEntry struct {
	display *Display
	sheet   *Sheet
}

// newCreoleParser returns the creole parser of a text; skinParam must not
// be nil, as upstream's Objects.requireNonNull.
func newCreoleParser(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	skinParam ISkinSimple, creoleMode CreoleMode, stereotype *FontConfiguration) SheetBuilder {
	if skinParam == nil {
		throw(nullPointer, "")
	}
	return &creoleParser{stereotype: stereotype, creoleMode: creoleMode, fontConfiguration: fontConfiguration,
		skinParam: skinParam, horizontalAlignment: horizontalAlignment}
}

// NewCreoleParser is upstream's legacy CreoleParser constructor, for the
// ISkinSimple implementations outside this package.
func NewCreoleParser(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	skinParam ISkinSimple, creoleMode CreoleMode, stereotype *FontConfiguration) SheetBuilder {
	return newCreoleParser(fontConfiguration, horizontalAlignment, skinParam, creoleMode, stereotype)
}

func (p *creoleParser) createStripes(line string, context *CreoleContext, lastStripe Stripe,
	fontConfiguration *FontConfiguration) []Stripe {
	if code, ok := lastStripe.(StripeRaw); ok {
		if code.IsTerminated() {
			lastStripe = nil
		} else {
			code.AddAndCheckTermination(line)
			return nil
		}
	}
	table, isTable := lastStripe.(*StripeTable)
	tree, isTree := lastStripe.(*StripeTree)
	switch {
	case isTable && CreoleParserIsTableLine(line):
		table.AnalyzeAndAddLine(line)
		return nil
	case isTree && ParserIsTreeStart(trin(line)):
		tree.AnalyzeAndAdd(line)
		return nil
	case CreoleParserIsTableLine(line):
		return []Stripe{NewStripeTable(fontConfiguration, p.skinParam, line)}
	case ParserIsTreeStart(line):
		return []Stripe{NewStripeTree(fontConfiguration, p.skinParam, line)}
	case ParserIsCodeStart(line):
		return []Stripe{NewStripeCode(fontConfiguration.ChangeFamily(ParserMonospaced))}
	case ParserIsLatexStart(line):
		return []Stripe{NewStripeLatex(fontConfiguration)}
	}
	align := p.horizontalAlignment
	if simple, ok := lastStripe.(*StripeSimple); ok {
		align = simple.CellAlignment()
	}
	return newCreoleStripeSimpleParser(line, context, fontConfiguration, p.skinParam, p.creoleMode).
		createStripes(context, align)
}

var (
	creoleTableLinePattern       = regexp.MustCompile(javaDots(`^(\<#\w+(,#?\w+)?\>)?\|(\=)?.*\|$`))
	creoleStartsByColorPattern   = regexp.MustCompile(javaDots(`^\=?[\t\n\v\f\r ]*(\<#\w+(,#?\w+)?\>).*$`))
	creoleTreeLeadingPipePattern = regexp.MustCompile(`^[\t\n\v\f\r ]*\|_`)
)

// CreoleParserIsTableLine reports whether line is a table row: "|...|",
// possibly after a "<#color>".
func CreoleParserIsTableLine(line string) bool { return creoleTableLinePattern.MatchString(line) }

// CreoleParserDoesStartByColor reports whether line starts with a color
// tag ("<#color>" or "<#back,#line>"), possibly after "=" and spaces.
func CreoleParserDoesStartByColor(line string) bool {
	return creoleStartsByColorPattern.MatchString(line)
}

// CreateSheet returns the sheet of a display, cached by display equality.
func (p *creoleParser) CreateSheet(display *Display) *Sheet {
	for _, e := range p.cache {
		if e.display == display ||
			(!DisplayIsNull(display) && !DisplayIsNull(e.display) && display.HashCode() == e.display.HashCode() &&
				display.Equals(e.display)) {
			return e.sheet
		}
	}
	result := p.createSheetSlow(display)
	p.cache = append(p.cache, creoleCacheEntry{display, result})
	return result
}

func (p *creoleParser) createSheetSlow(display *Display) *Sheet {
	sheet := NewSheet(p.horizontalAlignment)
	if DisplayIsNull(display) {
		return sheet
	}
	context := NewCreoleContext()
	lines := display.AsList()
	for i := 0; i < len(lines); i++ {
		cs := lines[i]
		var stripes []Stripe
		if typ, ok := EmbeddedDiagramType(cs.String()); ok {
			var embeddedDiagram Atom
			embeddedDiagram, i = embeddedDiagramCreateAndSkip(typ, lines, i+1, p.skinParam)
			stripes = []Stripe{&embeddedStripe{atom: embeddedDiagram}}
		} else if st, ok := cs.(*Stereotype); ok {
			if display.ShowStereotype() {
				// Upstream throws NullPointerException for a stereotype
				// without labels ("<<O-O>>"); it gives no stripe here.
				for _, label := range st.Labels(p.skinParam.Guillemet()) {
					sheet.Add(p.createStripes(label, context, sheet.LastStripe(), p.stereotype)...)
				}
			}
			continue
		} else {
			stripes = p.createStripes(p.skinParam.Guillemet().ManageGuillemet(cs.String()), context,
				sheet.LastStripe(), p.fontConfiguration)
		}
		sheet.Add(stripes...)
	}
	return sheet
}

// embeddedStripe is the anonymous stripe of an embedded diagram.
type embeddedStripe struct{ atom Atom }

func (s *embeddedStripe) LHeader() Atom  { return nil }
func (s *embeddedStripe) Atoms() []Atom  { return []Atom{s.atom} }
func (s *embeddedStripe) String() string { return "EmbeddedDiagram" }

// EmbeddedDiagramMaker is implemented by the skin parameters that can draw
// the embedded diagrams of a text ("{{" ... "}}"). Upstream's
// EmbeddedDiagram builds and exports the diagram, which needs the diagram
// engine: the engine provides it through the ISkinSimple of the text.
type EmbeddedDiagramMaker interface {
	// MakeEmbeddedDiagram returns the atom of the diagram of lines, from
	// "@start<type>" to "@end<type>".
	MakeEmbeddedDiagram(lines []string) Atom
}

// EmbeddedDiagramStart and EmbeddedDiagramEnd are upstream's
// EmbeddedDiagram.EMBEDDED_START and EMBEDDED_END.
const (
	EmbeddedDiagramStart = "{{"
	EmbeddedDiagramEnd   = embeddedEnd
)

// embeddedDiagramCreateAndSkip is upstream's EmbeddedDiagram.createAndSkip:
// the lines from start up to the matching "}}" make the diagram. It returns
// the atom and the index of the closing line.
func embeddedDiagramCreateAndSkip(typ string, lines []CharSequence, start int, skinParam ISkinSimple) (Atom, int) {
	result := []string{"@start" + typ}
	nested := 1
	i := start
	for ; i < len(lines); i++ {
		s2 := lines[i].String()
		if _, ok := EmbeddedDiagramType(s2); ok {
			nested++
		} else if trin(s2) == EmbeddedDiagramEnd {
			nested--
			if nested == 0 {
				break
			}
		}
		result = append(result, s2)
	}
	result = append(result, "@end"+typ)
	maker, ok := skinParam.(EmbeddedDiagramMaker)
	if !ok {
		notPorted("EmbeddedDiagram", "embedded diagram {{"+typ)
	}
	return maker.MakeEmbeddedDiagram(result), i
}

// embeddedDiagramTypes are the keywords after "{{" of upstream's
// getEmbeddedType.
var embeddedDiagramTypes = map[string]bool{"board": true, "creole": true, "chronology": true, "chen": true,
	"chart": true, "ditaa": true, "ebnf": true, "files": true, "gantt": true, "json": true, "mindmap": true,
	"nwdiag": true, "packetdiag": true, "regex": true, "salt": true, "uml": true, "wbs": true, "wire": true,
	"yaml": true}

// EmbeddedDiagramType returns the diagram type of a line that starts an
// embedded diagram ("{{" alone is "uml", "{{json" is "json"); ok is false
// for other lines.
func EmbeddedDiagramType(cs string) (string, bool) {
	runes := []rune(cs)
	p := 0
	for p < len(runes) && javaIsWhitespace(runes[p]) {
		p++
	}
	if p+2 > len(runes) || runes[p] != '{' || runes[p+1] != '{' {
		return "", false
	}
	p += 2
	end := len(runes)
	for end > p && javaIsWhitespace(runes[end-1]) {
		end--
	}
	suffix := string(runes[p:end])
	if suffix == "" {
		return "uml", true
	}
	if embeddedDiagramTypes[suffix] {
		return suffix, true
	}
	return "", false
}

// creoleStripeSimpleParser is upstream's CreoleStripeSimpleParser: it
// recognizes the line style (heading, list item, horizontal rule) and
// makes the stripes of the line.
type creoleStripeSimpleParser struct {
	line              string
	style             StripeStyle
	modeSimpleLine    CreoleMode
	fontConfiguration *FontConfiguration
	skinParam         ISkinSimple
}

var (
	creoleSectionTitlePattern        = creolePattern("^==([^=]*)==$")
	creoleSectionSeparatorPattern    = creolePattern("^===*==$")
	creoleDoubleDotDelimitedLine     = creolePattern(`^\.\.([^\.]*)\.\.$`)
	creoleAsteriskPrefixedLine       = creolePattern(`^(\*+)([^*]+(?:[^*]|\*\*[^*]+\*\*)*)$`)
	creoleAsteriskHeaderLinePattern  = creolePattern(`^(\*+)([%s].+)$`)
	creoleHashHeadingPattern         = creolePattern("^(#+)(.+)$")
	creoleEqualsHeadingPattern       = creolePattern("^(=+)(.+)$")
	creoleSectionHeaderPattern       = creolePattern("^--([^-]*)--$")
	creoleNormalStyle                = NewStripeStyle(StripeStyleTypeNormal, 0, 0)
	creoleTableStyle                 = NewStripeStyle(StripeStyleTypeNormal, 0, 0)
	creoleTreeStyle                  = NewStripeStyle(StripeStyleTypeTree, 0, 0)
	creoleHorizontalLineStyleDash    = NewStripeStyle(StripeStyleTypeHorizontalLine, 0, '-')
	creoleHorizontalLineStyleEquals  = NewStripeStyle(StripeStyleTypeHorizontalLine, 0, '=')
	creoleHorizontalLineStyleDot     = NewStripeStyle(StripeStyleTypeHorizontalLine, 0, '.')
	creoleSplitBlockE1NewlineDivider = string(JawsBlockE1Newline)
)

func newCreoleStripeSimpleParser(line string, creoleContext *CreoleContext, fontConfiguration *FontConfiguration,
	skinParam ISkinSimple, mode CreoleMode) *creoleStripeSimpleParser {
	if skinParam == nil {
		throw(nullPointer, "")
	}
	p := &creoleStripeSimpleParser{fontConfiguration: fontConfiguration, modeSimpleLine: mode,
		skinParam: skinParam, line: line, style: creoleNormalStyle}
	if mode == CreoleModeNoCreole {
		return p
	}
	find := func(re *regexp.Regexp, s string) []string { return re.FindStringSubmatch(s) }
	if m := find(creoleSectionHeaderPattern, line); m != nil {
		p.line, p.style = m[1], creoleHorizontalLineStyleDash
		return p
	}
	if m := find(creoleSectionTitlePattern, line); m != nil {
		p.line, p.style = m[1], creoleHorizontalLineStyleEquals
		return p
	}
	if find(creoleSectionSeparatorPattern, line) != nil {
		p.line, p.style = "", creoleHorizontalLineStyleEquals
		return p
	}
	if m := find(creoleDoubleDotDelimitedLine, line); m != nil {
		p.line, p.style = m[1], creoleHorizontalLineStyleDot
		return p
	}
	if mode == CreoleModeFull {
		if m := find(creoleAsteriskPrefixedLine, line); m != nil {
			p.line = trin(m[2])
			p.style = NewStripeStyle(StripeStyleTypeListWithoutNumber, javaLen(m[1])-1, 0)
			return p
		}
		if m := find(creoleAsteriskHeaderLinePattern, line); m != nil {
			p.line = trin(m[2])
			p.style = NewStripeStyle(StripeStyleTypeListWithoutNumber, javaLen(m[1])-1, 0)
			return p
		}
		if m := find(creoleHashHeadingPattern, CharHidderHide(line)); m != nil {
			p.line = trin(CharHidderUnhide(m[2]))
			p.style = NewStripeStyle(StripeStyleTypeListWithNumber, javaLen(CharHidderUnhide(m[1]))-1, 0)
			return p
		}
	}
	if m := find(creoleEqualsHeadingPattern, line); m != nil {
		p.line = trin(m[2])
		p.style = NewStripeStyle(StripeStyleTypeHeading, javaLen(m[1])-1, 0)
		return p
	}
	return p
}

func (p *creoleStripeSimpleParser) createStripes(context *CreoleContext, align geom.HorizontalAlignment) []Stripe {
	var result []Stripe
	for _, singleLine := range javaSplit(p.line, creoleSplitBlockE1NewlineDivider) {
		stripe := NewStripeSimple(p.fontConfiguration, p.style, context, p.skinParam, p.modeSimpleLine)
		stripe.SetCellAlignment(align)
		stripe.AnalyzeAndAdd(singleLine)
		result = append(result, stripe)
	}
	return result
}

// StripeSimple is a line of text: atoms in the font configurations of the
// markup, after an optional header (bullet or number).
type StripeSimple struct {
	header            Atom
	atoms             []Atom
	commands          creoleCommands
	align             geom.HorizontalAlignment
	fontConfiguration *FontConfiguration
	style             StripeStyle
	skinParam         ISkinSimple
}

var _ Stripe = (*StripeSimple)(nil)

// NewStripeSimple returns an empty line of a style; the header of a list
// item takes its number from context.
func NewStripeSimple(fontConfiguration *FontConfiguration, style StripeStyle, context *CreoleContext,
	skinParam ISkinSimple, modeSimpleLine CreoleMode) *StripeSimple {
	s := &StripeSimple{fontConfiguration: fontConfiguration, style: style, skinParam: skinParam,
		align: geom.HorizontalAlignmentLeft, commands: creoleCommandsOther}
	if modeSimpleLine == CreoleModeFull {
		s.commands = creoleCommandsFull
	}
	s.header = style.Header(fontConfiguration, context)
	if s.header != nil {
		s.atoms = append(s.atoms, s.header)
	}
	return s
}

// SetCellAlignment sets the alignment of the line.
func (s *StripeSimple) SetCellAlignment(align geom.HorizontalAlignment) { s.align = align }

// CellAlignment returns the alignment of the line.
func (s *StripeSimple) CellAlignment() geom.HorizontalAlignment { return s.align }

// String is upstream's toString (without the identity hash).
func (s *StripeSimple) String() string {
	parts := make([]string, len(s.atoms))
	for i, a := range s.atoms {
		parts[i] = textBlockString(a)
	}
	return "StripeSimple [" + strings.Join(parts, ", ") + "]"
}

// LHeader returns the header, nil for none.
func (s *StripeSimple) LHeader() Atom { return s.header }

// Atoms returns the atoms; an empty line first gets a space, as upstream.
func (s *StripeSimple) Atoms() []Atom {
	if len(s.atoms) == 0 {
		s.atoms = append(s.atoms, AtomTextUtilsCreateLegacy(" ", s.fontConfiguration))
	}
	return s.atoms
}

// ActualFontConfiguration returns the current font configuration.
func (s *StripeSimple) ActualFontConfiguration() *FontConfiguration { return s.fontConfiguration }

// SetActualFontConfiguration sets the current font configuration.
func (s *StripeSimple) SetActualFontConfiguration(fontConfiguration *FontConfiguration) {
	s.fontConfiguration = fontConfiguration
}

// AnalyzeAndAdd parses a line of markup and adds its atoms. A line with a
// newline character panics, as upstream throws IllegalArgumentException.
func (s *StripeSimple) AnalyzeAndAdd(line string) {
	if strings.ContainsRune(line, JawsBlockE1Newline) {
		throw(illegalArgument, line)
	}
	line = s.manageCellAlignment(line)
	line = CharHidderHide(line)
	switch s.style.Type() {
	case StripeStyleTypeHeading:
		s.fontConfiguration = fontConfigurationForHeading(s.fontConfiguration, s.style.Order())
		s.modifyStripe(line)
	case StripeStyleTypeHorizontalLine:
		s.atoms = append(s.atoms, CreoleHorizontalLineCreate(s.fontConfiguration, line, s.style.Style(),
			s.skinParam))
	default:
		s.modifyStripe(line)
	}
}

// The alignment prefixes of a line.
var stripeAlignments = []struct {
	prefix string
	align  geom.HorizontalAlignment
}{
	{"<l>", geom.HorizontalAlignmentLeft},
	{"<left>", geom.HorizontalAlignmentLeft},
	{"<center>", geom.HorizontalAlignmentCenter},
	{"<c>", geom.HorizontalAlignmentCenter},
	{"<right>", geom.HorizontalAlignmentRight},
	{"<r>", geom.HorizontalAlignmentRight},
}

func (s *StripeSimple) manageCellAlignment(line string) string {
	if strings.HasPrefix(line, "<") {
		for _, a := range stripeAlignments {
			if rest, ok := strings.CutPrefix(line, a.prefix); ok {
				s.SetCellAlignment(a.align)
				return rest
			}
		}
	}
	return line
}

func fontConfigurationForHeading(fontConfiguration *FontConfiguration, order int) *FontConfiguration {
	switch order {
	case 0:
		return fontConfiguration.Bigger(4).Bold()
	case 1:
		return fontConfiguration.Bigger(2).Bold()
	case 2:
		return fontConfiguration.Bigger(1).Bold()
	}
	return fontConfiguration.Italic()
}

// AddImage adds an image.
func (s *StripeSimple) AddImage(src string, scale float64) {
	s.atoms = append(s.atoms, AtomImgCreate(s.skinParam, src, geom.ImgValignTop, 0, scale, nil))
}

// AddQrcode adds a QR code; QR codes are not ported yet and panic.
func (s *StripeSimple) AddQrcode(src string, scale float64) {
	notPorted("klimt/creole/atom/AtomImg.createQrcode", "QR code <qrcode:"+src+">")
}

// AddSpace adds a space of size pixels.
func (s *StripeSimple) AddSpace(size int) { s.atoms = append(s.atoms, AtomSpaceCreate(float64(size))) }

// AddURL adds a link.
func (s *StripeSimple) AddURL(url *URL) {
	if url == nil {
		throw(nullPointer, "")
	}
	s.atoms = append(s.atoms, AtomTextUtilsCreateURL(url, s.fontConfiguration, s.skinParam))
}

// AddSprite adds a sprite, or its literal markup when the sprite is
// unknown. Upstream adds nothing for an unknown sprite; until sprites are
// ported (none can be defined yet), the specification asks for the literal
// text.
func (s *StripeSimple) AddSprite(src string, scale float64, forcedColor color.HColor, literal string) {
	sprite := s.skinParam.Sprite(src)
	if sprite == nil {
		s.atoms = append(s.atoms, AtomTextUtilsCreateLegacy(literal, s.fontConfiguration))
		return
	}
	var backColor color.HColor
	if s.fontConfiguration.ContainsStyle(font.FontStyleBackcolor) {
		backColor = s.fontConfiguration.ExtendedColor()
	}
	s.atoms = append(s.atoms, NewAtomSprite(s.fontConfiguration.Color(), forcedColor, scale, sprite, nil,
		backColor))
}

// AddOpenIcon adds an OpenIconic icon. The icons are not ported yet: the
// specification asks for the literal markup until they are.
func (s *StripeSimple) AddOpenIcon(src string, scale float64, c color.HColor, literal string) {
	s.atoms = append(s.atoms, AtomTextUtilsCreateLegacy(literal, s.fontConfiguration))
}

// AddEmoji adds an emoji. The emoji are not ported yet: the specification
// asks for the literal markup until they are (upstream draws "¿name?" in
// red for an unknown emoji).
func (s *StripeSimple) AddEmoji(emojiName string, scale float64, forcedColor string, hasColor bool,
	literal string) {
	s.atoms = append(s.atoms, AtomTextUtilsCreateLegacy(literal, s.fontConfiguration))
}

// AddMath adds a formula; formulas are not ported yet and panic.
func (s *StripeSimple) AddMath(kind, formula string) {
	notPorted("klimt/creole/atom/AtomMath", kind+" formula "+formula)
}

func (s *StripeSimple) modifyStripe(line string) {
	u := utf16Of(line)
	var pending []uint16
	pos := 0
	for pos < len(u) {
		cmd := s.searchCommand(u, pos)
		if cmd == nil {
			pending = append(pending, u[pos])
			pos++
		} else {
			pending = s.addPending(pending)
			pos += cmd.executeAndAdvance(s.skinParam, u, pos, s)
		}
	}
	s.addPending(pending)
}

func (s *StripeSimple) addPending(pending []uint16) []uint16 {
	if len(pending) == 0 {
		return pending
	}
	s.atoms = append(s.atoms, AtomTextUtilsCreateLegacy(stringOf(pending), s.fontConfiguration))
	return pending[:0]
}

func (s *StripeSimple) searchCommand(line []uint16, pos int) creoleCommand {
	if len(line) > pos+2 {
		for _, cmd := range s.commands[[2]uint16{line[pos], line[pos+1]}] {
			if cmd.matchingSize(line, pos) != 0 {
				return cmd
			}
		}
	}
	return nil
}

// SkinParam returns the skin parameters.
func (s *StripeSimple) SkinParam() ISkinSimple { return s.skinParam }

// StripeTable is a creole table: consecutive "|...|" lines, "|=" cells are
// headers in bold, and "<#color>" sets row, cell and grid colors.
type StripeTable struct {
	fontConfiguration *FontConfiguration
	skinParam         ISkinSimple
	table             *AtomTable
	marged            Atom
}

var _ Stripe = (*StripeTable)(nil)

// NewStripeTable returns the table of its first line.
func NewStripeTable(fontConfiguration *FontConfiguration, skinParam ISkinSimple, line string) *StripeTable {
	t := &StripeTable{skinParam: skinParam, fontConfiguration: fontConfiguration}
	lineColor := t.backOrFrontColor(line, 1)
	if lineColor == nil {
		lineColor = fontConfiguration.Color()
	}
	t.table = NewAtomTable(lineColor)
	t.marged = NewAtomWithMargin(t.table, 2, 2)
	t.analyzeAndAddInternal(line)
	return t
}

// Atoms returns the table with its margins.
func (t *StripeTable) Atoms() []Atom { return []Atom{t.marged} }

// LHeader returns nil.
func (t *StripeTable) LHeader() Atom { return nil }

// String describes the stripe.
func (t *StripeTable) String() string { return "StripeTable" }

// stripeTableAsAtom is upstream's StripeTable.asAtom: the cell lines as a
// left-aligned sheet.
func stripeTableAsAtom(cells []*StripeSimple, padding ClockwiseTopRightBottomLeft) Atom {
	sheet := NewSheet(geom.HorizontalAlignmentLeft)
	for _, cell := range cells {
		sheet.Add(cell)
	}
	return NewSheetBlock1Padding(sheet, LineBreakStrategyNone, padding)
}

func (t *StripeTable) backOrFrontColor(line string, idx int) color.HColor {
	if CreoleParserDoesStartByColor(line) {
		idx1 := strings.IndexByte(line, '#')
		idx2 := strings.IndexByte(line, '>')
		if idx2 == -1 {
			throw(illegalState, "")
		}
		if idx1 > idx2 {
			throw("java.lang.StringIndexOutOfBoundsException", "")
		}
		colors := javaSplit(line[idx1:idx2], ",")
		if idx < len(colors) {
			return t.skinParam.IHtmlColorSet().GetColorOrWhite(colors[idx])
		}
	}
	return nil
}

func stripeTableWithoutBackColor(line string) string {
	idx2 := strings.IndexByte(line, '>')
	if idx2 == -1 {
		throw(illegalState, "")
	}
	return line[idx2+1:]
}

const stripeTableHiddenBar = "\uE000"

func (t *StripeTable) analyzeAndAddInternal(line string) {
	line = strings.ReplaceAll(line, `\|`, stripeTableHiddenBar)
	lineBackColor := t.backOrFrontColor(line, 0)
	if lineBackColor != nil {
		line = stripeTableWithoutBackColor(line)
	}
	t.table.NewLine(lineBackColor)
	for _, token := range javaTokenize(line, "|") {
		header := false
		v := strings.ReplaceAll(token, stripeTableHiddenBar, "|")
		if rest, ok := strings.CutPrefix(v, "="); ok {
			v = rest
			header = true
		}
		cellBackColor := t.backOrFrontColor(v, 0)
		if cellBackColor != nil {
			v = stripeTableWithoutBackColor(v)
		}
		var cells []*StripeSimple
		for _, s := range stripeTableWithNewlinesInternal(v) {
			cell := NewStripeSimple(t.cellFontConfiguration(header), creoleTableStyle, NewCreoleContext(),
				t.skinParam, CreoleModeFull)
			if rest, ok := strings.CutPrefix(s, "<r>"); ok {
				cell.SetCellAlignment(geom.HorizontalAlignmentRight)
				s = rest
			}
			cell.AnalyzeAndAdd(s)
			cells = append(cells, cell)
		}
		t.table.AddCell(stripeTableAsAtom(cells, t.skinParam.Padding()), cellBackColor)
	}
}

// stripeTableWithNewlinesInternal splits a cell at "\n" (and the hidden
// newline character); "\\" is a backslash.
func stripeTableWithNewlinesInternal(s string) []string {
	var result []string
	u := utf16Of(s)
	var current []uint16
	if PragmaLegacyReplaceBackslashNByNewline() {
		for i := 0; i < len(u); i++ {
			c := u[i]
			switch {
			case c == '\\' && i < len(u)-1:
				c2 := u[i+1]
				i++
				switch c2 {
				case 'n':
					result = append(result, stringOf(current))
					current = current[:0]
				case '\\':
					current = append(current, c2)
				default:
					current = append(current, c, c2)
				}
			case rune(c) == hiddenNewLine():
				result = append(result, stringOf(current))
				current = current[:0]
			default:
				current = append(current, c)
			}
		}
		return append(result, stringOf(current))
	}
	for _, c := range u {
		if rune(c) == JawsBlockE1Newline {
			result = append(result, stringOf(current))
			current = current[:0]
		} else {
			current = append(current, c)
		}
	}
	return append(result, stringOf(current))
}

func (t *StripeTable) cellFontConfiguration(header bool) *FontConfiguration {
	if !header {
		return t.fontConfiguration
	}
	return t.fontConfiguration.Bold()
}

// AnalyzeAndAddLine adds a row.
func (t *StripeTable) AnalyzeAndAddLine(line string) { t.analyzeAndAddInternal(line) }

// StripeTree is a creole tree: consecutive "|_" lines, indented by two
// spaces or a tabulation per level.
type StripeTree struct {
	fontConfiguration *FontConfiguration
	skinParam         ISkinSimple
	tree              *AtomTree
	marged            Atom
}

var _ Stripe = (*StripeTree)(nil)

// NewStripeTree returns the tree of its first line.
func NewStripeTree(fontConfiguration *FontConfiguration, skinParam ISkinSimple, line string) *StripeTree {
	t := &StripeTree{fontConfiguration: fontConfiguration, skinParam: skinParam}
	t.tree = NewAtomTree(fontConfiguration.Color())
	t.marged = NewAtomWithMargin(t.tree, 2, 2)
	t.AnalyzeAndAdd(line)
	return t
}

// Atoms returns the tree with its margins.
func (t *StripeTree) Atoms() []Atom { return []Atom{t.marged} }

// LHeader returns nil.
func (t *StripeTree) LHeader() Atom { return nil }

// String describes the stripe.
func (t *StripeTree) String() string { return "StripeTree" }

// AnalyzeAndAdd adds the cells of a line.
func (t *StripeTree) AnalyzeAndAdd(line string) {
	for _, s := range stripeTableWithNewlinesInternal(line) {
		cell := NewStripeSimple(t.fontConfiguration, creoleTreeStyle, NewCreoleContext(), t.skinParam,
			CreoleModeFull)
		text := s
		if loc := creoleTreeLeadingPipePattern.FindStringIndex(s); loc != nil {
			text = s[loc[1]:]
		}
		level := stripeTreeComputeLevel(s)
		cell.AnalyzeAndAdd(text)
		t.tree.AddCell(stripeTableAsAtom([]*StripeSimple{cell}, ClockwiseNone()), level)
	}
}

func stripeTreeComputeLevel(s string) int {
	result := 1
	for s != "" {
		if rest, ok := strings.CutPrefix(s, "  "); ok {
			result++
			s = rest
			continue
		} else if rest, ok := strings.CutPrefix(s, "\t"); ok {
			result++
			s = rest
			continue
		}
		return result
	}
	return result
}

// StripeRaw is a stripe that takes the following lines verbatim until its
// end marker: code and LaTeX blocks.
type StripeRaw interface {
	Stripe
	Atom
	// AddAndCheckTermination adds a line, or ends the block at its end
	// marker (reporting true).
	AddAndCheckTermination(line string) bool
	IsTerminated() bool
}

// StripeCode is a "<code>" block: lines drawn verbatim in the monospaced
// font.
type StripeCode struct {
	AtomBase
	fontConfiguration *FontConfiguration
	raw               []string
	terminated        bool
}

var _ StripeRaw = (*StripeCode)(nil)

// NewStripeCode returns an empty code block.
func NewStripeCode(fontConfiguration *FontConfiguration) *StripeCode {
	return &StripeCode{fontConfiguration: fontConfiguration}
}

// Atoms returns the block itself.
func (c *StripeCode) Atoms() []Atom { return []Atom{c} }

// LHeader returns nil.
func (c *StripeCode) LHeader() Atom { return nil }

// String describes the stripe.
func (c *StripeCode) String() string { return "StripeCode" }

// AddAndCheckTermination adds a line, or ends the block at "</code>".
func (c *StripeCode) AddAndCheckTermination(line string) bool {
	if ParserIsCodeEnd(line) {
		c.terminated = true
		return true
	}
	c.raw = append(c.raw, line)
	return false
}

// IsTerminated reports whether "</code>" was seen.
func (c *StripeCode) IsTerminated() bool { return c.terminated }

// CalculateDimension returns the largest width and the sum of the heights
// of the lines.
func (c *StripeCode) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return c.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		width, height := 0.0, 0.0
		for _, s := range c.raw {
			dim := sb.CalculateDimension(c.fontConfiguration.Font(), s)
			width = math.Max(width, dim.Width)
			height += dim.Height
		}
		return geom.NewXDimension2D(width, height)
	})
}

// StartingAltitude returns 0.
func (c *StripeCode) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU draws the lines.
func (c *StripeCode) DrawU(ug UGraphic) {
	y := 0.0
	for _, s := range c.raw {
		useFontConfiguration := c.fontConfiguration.AdjustColorForBackground(ug)
		shape := UTextBuild(s, useFontConfiguration)
		stringBounder := ug.StringBounder()
		dim := stringBounder.CalculateDimension(c.fontConfiguration.Font(), s)
		y += dim.Height
		ug.Apply(UTranslateDy(y - shape.Descent(stringBounder))).Draw(shape)
	}
}

// Neutrons returns the block as one neutron.
func (c *StripeCode) Neutrons() []*Neutron { return atomNeutrons(c) }

// StripeLatex is a "<latex>" block. Formulas are not ported yet: measuring
// or drawing the block panics.
type StripeLatex struct {
	TextBlockDefaults
	fontConfiguration *FontConfiguration
	formula           strings.Builder
	terminated        bool
}

var _ StripeRaw = (*StripeLatex)(nil)

// NewStripeLatex returns an empty LaTeX block.
func NewStripeLatex(fontConfiguration *FontConfiguration) *StripeLatex {
	return &StripeLatex{fontConfiguration: fontConfiguration}
}

// Atoms returns the block itself.
func (l *StripeLatex) Atoms() []Atom { return []Atom{l} }

// LHeader returns nil.
func (l *StripeLatex) LHeader() Atom { return nil }

// String describes the stripe.
func (l *StripeLatex) String() string { return "StripeLatex" }

// AddAndCheckTermination adds a line, or ends the block at "</latex>".
func (l *StripeLatex) AddAndCheckTermination(line string) bool {
	if ParserIsLatexEnd(line) {
		l.terminated = true
		return true
	}
	l.formula.WriteString(line)
	return false
}

// IsTerminated reports whether "</latex>" was seen.
func (l *StripeLatex) IsTerminated() bool { return l.terminated }

func (l *StripeLatex) atom() Atom {
	notPorted("klimt/creole/atom/AtomMath", "LaTeX formula "+l.formula.String())
	return nil
}

// CalculateDimension panics: formulas are not ported yet.
func (l *StripeLatex) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return l.atom().CalculateDimension(stringBounder)
}

// StartingAltitude returns 0.
func (l *StripeLatex) StartingAltitude(font.StringBounder) float64 { return 0 }

// DrawU panics: formulas are not ported yet.
func (l *StripeLatex) DrawU(ug UGraphic) { l.atom().DrawU(ug) }

// Neutrons panics, as upstream throws UnsupportedOperationException.
func (l *StripeLatex) Neutrons() []*Neutron {
	throw(unsupportedOperation, "")
	return nil
}
