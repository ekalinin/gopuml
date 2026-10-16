// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/MethodsOrFieldsArea.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/Elected.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// MethodsOrFieldsArea is the area of the fields or of the methods of a
// class: one line per member, with the visibility icons in a column on the
// left when the diagram shows them, and the embedded diagrams ("{{" ...
// "}}") of the members after them.
type MethodsOrFieldsArea struct {
	klimt.TextBlockDefaults
	memo klimt.TextBlockMemoized

	skinParam klimt.ISkinParam
	members   *klimt.Display
	align     geom.HorizontalAlignment
	embeddeds []klimt.TextBlock
	leaf      *Entity
	style     *klimt.Style
}

var _ klimt.WithPorts = (*MethodsOrFieldsArea)(nil)

// NewMethodsOrFieldsArea returns the area of members, aligned on the left.
func NewMethodsOrFieldsArea(members *klimt.Display, skinParam klimt.ISkinParam, leaf *Entity,
	style *klimt.Style) *MethodsOrFieldsArea {
	return NewMethodsOrFieldsAreaAligned(members, skinParam, geom.HorizontalAlignmentLeft, leaf, style)
}

// NewMethodsOrFieldsAreaAligned returns the area of members with an
// alignment.
func NewMethodsOrFieldsAreaAligned(members *klimt.Display, skinParam klimt.ISkinParam,
	align geom.HorizontalAlignment, leaf *Entity, style *klimt.Style) *MethodsOrFieldsArea {
	a := &MethodsOrFieldsArea{style: style, leaf: leaf, align: align, skinParam: skinParam}
	var result []klimt.CharSequence
	lines := members.AsList()
	for i := 0; i < len(lines); i++ {
		cs := lines[i]
		if typ, ok := klimt.EmbeddedDiagramType(cs.String()); ok {
			var embedded klimt.TextBlock
			embedded, i = klimt.EmbeddedDiagramCreateAndSkip(typ, lines, i+1, skinParam)
			a.embeddeds = append(a.embeddeds, embedded)
		} else {
			result = append(result, cs)
		}
	}
	a.members = klimt.DisplayCreateSeq(result)
	return a
}

// AsBlockMemberImpl returns the area with its margins after a separator
// line, as a compartment of a class box.
func (a *MethodsOrFieldsArea) AsBlockMemberImpl() klimt.TextBlock {
	return klimt.NewTextBlockLineBefore(a.style.Value(klimt.PNameLineThickness).AsDouble(),
		klimt.TextBlockUtilsWithMargin(a, 6, 4), 0, nil)
}

func (a *MethodsOrFieldsArea) hasSmallIcon() bool {
	if a.skinParam.ClassAttributeIconSize() == 0 {
		return false
	}
	for _, cs := range a.members.AsList() {
		m, ok := cs.(*Member)
		if !ok {
			continue
		}
		if _, ok := m.VisibilityModifier(); ok {
			return true
		}
	}
	return false
}

// CalculateDimension returns the size of the members and the embedded
// diagrams, memoized.
func (a *MethodsOrFieldsArea) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return a.memo.Dimension(stringBounder, a.calculateDimensionSlow)
}

func (a *MethodsOrFieldsArea) calculateDimensionSlow(stringBounder font.StringBounder) geom.XDimension2D {
	dim1 := a.calculateDimensionOnlyMembers(stringBounder)
	x := dim1.Width()
	y := dim1.Height()
	for _, embedded := range a.embeddeds {
		dim := embedded.CalculateDimension(stringBounder)
		x = math.Max(dim.Width(), x)
		y += dim.Height()
	}
	return geom.NewXDimension2D(x, y)
}

func (a *MethodsOrFieldsArea) calculateDimensionOnlyMembers(stringBounder font.StringBounder) geom.XDimension2D {
	smallIcon := 0.0
	if a.hasSmallIcon() {
		smallIcon = float64(a.skinParam.CircledCharacterRadius() + 3)
	}
	x, y := 0.0, 0.0
	for _, cs := range a.members.AsList() {
		block := a.createTextBlock(cs)
		dim := block.CalculateDimension(stringBounder)
		x = math.Max(dim.Width(), x)
		y += dim.Height()
		if stringBounder.MatchesProperty("TIKZ") && dim.Height() == 10 {
			// The modifier would add a hard-coded 1 to the height, as
			// upstream's AtomText and VisibilityModifier.getUBlock do.
			y++
		}
	}
	x += smallIcon
	return geom.NewXDimension2D(x, y)
}

// javaStringLength is String.length: the number of UTF-16 units.
func javaStringLength(s string) int { return len(utf16.Encode([]rune(s))) }

// javaCompareTo is String.compareTo: the lexicographic order of the UTF-16
// units.
func javaCompareTo(s1, s2 string) int {
	u1, u2 := utf16.Encode([]rune(s1)), utf16.Encode([]rune(s2))
	for i := 0; i < len(u1) && i < len(u2); i++ {
		if u1[i] != u2[i] {
			return int(u1[i]) - int(u2[i])
		}
	}
	return len(u1) - len(u2)
}

// sortBySize returns the names, the longest first, then in alphabetical
// order.
func sortBySize(all []string) []string {
	result := slices.Clone(all)
	slices.SortStableFunc(result, func(s1, s2 string) int {
		if diff := javaStringLength(s2) - javaStringLength(s1); diff != 0 {
			return diff
		}
		return javaCompareTo(s1, s2)
	})
	return result
}

// Ports implements klimt.WithPorts: the members that name a port of the
// entity, with their vertical position.
func (a *MethodsOrFieldsArea) Ports(stringBounder font.StringBounder) *klimt.Ports {
	ports := klimt.NewPorts()
	y := 0.0
	shortNames := sortBySize(a.leaf.PortShortNames())
	for _, cs := range a.members.AsList() {
		block := a.createTextBlock(cs)
		dim := block.CalculateDimension(stringBounder)
		if elected := a.Elected(methodsOrFieldsConvert(cs), shortNames); elected != nil {
			ports.Add(elected.ShortName(), elected.Score(), y, dim.Height())
		}
		y += dim.Height()
	}
	return ports
}

func methodsOrFieldsConvert(cs klimt.CharSequence) string {
	if m, ok := cs.(*Member); ok {
		return m.Display(false)
	}
	return cs.String()
}

// Elected is upstream's Elected: the port name a member matches, with the
// score of the match.
type Elected struct {
	shortName string
	score     int
}

// String is upstream's toString.
func (e *Elected) String() string { return e.shortName + "/" + itoa(e.score) }

// ShortName returns the port name.
func (e *Elected) ShortName() string { return e.shortName }

// Score returns the score: 100 for a whole word, 50 for a part.
func (e *Elected) Score() int { return e.score }

// Elected returns the first port name that cs matches, nil when none does.
func (a *MethodsOrFieldsArea) Elected(cs string, shortNames []string) *Elected {
	for _, shortName := range shortNames {
		if score := methodsOrFieldsScore(cs, shortName); score > 0 {
			return &Elected{shortName: shortName, score: score}
		}
	}
	return nil
}

// methodsOrFieldsScore is upstream's getScore. As upstream, the port name
// is used as a regular expression.
func methodsOrFieldsScore(cs, shortName string) int {
	re, err := regexp.Compile(`^(?:` + uml.JavaDots(`.*\b`+shortName+`\b.*`) + `)$`)
	if err != nil {
		panic(&color.RuntimeError{Class: "java.util.regex.PatternSyntaxException", Message: err.Error()})
	}
	if re.MatchString(cs) {
		return 100
	}
	if strings.Contains(cs, shortName) {
		return 50
	}
	return 0
}

func (a *MethodsOrFieldsArea) createTextBlock(cs klimt.CharSequence) klimt.TextBlock {
	config := klimt.FontConfigurationFromStyle(a.skinParam, a.style, a.leaf.Colors())
	if m, ok := cs.(*Member); ok {
		withVisibilityChar := a.skinParam.ClassAttributeIconSize() == 0
		s := m.Display(withVisibilityChar)
		if withVisibilityChar && strings.HasPrefix(s, "#") {
			s = klimt.CharHidderAddTileAtBegin(s)
		}
		if m.IsAbstract() {
			config = config.Italic()
		}
		if m.IsStatic() {
			config = config.Underline()
		}
		block := klimt.DisplayWithNewlines(a.skinParam.Pragma(), s).Create8(config, a.align, a.skinParam,
			klimt.CreoleModeSimpleLine, a.style.WrapWidth())
		block = &fullInnerPositionBlock{block: block, display: m}
		return &textBlockTracer{block: block, url: m.URL()}
	}
	return klimt.DisplayWithNewlines(a.skinParam.Pragma(), cs.String()).Create8(config, a.align, a.skinParam,
		klimt.CreoleModeSimpleLine, a.style.WrapWidth())
}

// fullInnerPositionBlock is upstream's anonymous block of
// fullInnerPosition: the whole block is the inner position of its member.
type fullInnerPositionBlock struct {
	block   klimt.TextBlock
	display *Member
}

func (b *fullInnerPositionBlock) DrawU(ug klimt.UGraphic) { b.block.DrawU(ug) }

func (b *fullInnerPositionBlock) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return b.block.CalculateDimension(stringBounder)
}

func (b *fullInnerPositionBlock) MinMax(stringBounder font.StringBounder) geom.MinMax {
	return b.block.MinMax(stringBounder)
}

func (b *fullInnerPositionBlock) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	if b.display.String() == member {
		dim := b.CalculateDimension(stringBounder)
		r := geom.NewXRectangle2D(0, 0, dim.Width(), dim.Height())
		return &r
	}
	return nil
}

func (b *fullInnerPositionBlock) MagneticBorder() klimt.MagneticBorder {
	return b.block.MagneticBorder()
}

func (b *fullInnerPositionBlock) Backcolor() color.HColor { return b.block.Backcolor() }

// textBlockTracer is upstream's TextBlockTracer: the block of a member
// inside its link.
type textBlockTracer struct {
	klimt.TextBlockDefaults
	block klimt.TextBlock
	url   *klimt.URL
}

func (t *textBlockTracer) DrawU(ug klimt.UGraphic) {
	if t.url != nil {
		ug.StartURL(t.url)
	}
	t.block.DrawU(ug)
	if t.url != nil {
		ug.CloseURL()
	}
}

func (t *textBlockTracer) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return t.block.CalculateDimension(stringBounder)
}

func (t *textBlockTracer) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	return t.block.InnerPosition(member, stringBounder)
}

// noVisibilityBlock is upstream's anonymous block of getUBlock for a member
// without visibility: 1x1 and empty.
type noVisibilityBlock struct{ klimt.TextBlockDefaults }

func (noVisibilityBlock) DrawU(klimt.UGraphic) {}

func (noVisibilityBlock) InnerPosition(string, font.StringBounder) *geom.XRectangle2D { return nil }

func (noVisibilityBlock) CalculateDimension(font.StringBounder) geom.XDimension2D {
	return geom.NewXDimension2D(1, 1)
}

func (a *MethodsOrFieldsArea) uBlock(modifier klimt.VisibilityModifier, hasModifier bool,
	url *klimt.URL) klimt.TextBlock {
	if !hasModifier {
		return &noVisibilityBlock{}
	}
	style := modifier.StyleSignature().MergedStyle(a.skinParam.CurrentStyleBuilder())
	borderColor := style.Value(klimt.PNameLineColor).AsColor(a.skinParam.IHtmlColorSet())
	var backColor color.HColor
	if !modifier.IsField() {
		backColor = style.Value(klimt.PNameBackGroundColor).AsColor(a.skinParam.IHtmlColorSet())
	}
	uBlock := modifier.UBlock(a.skinParam.ClassAttributeIconSize(), borderColor, backColor, url != nil)
	return klimt.TextBlockWithURLOf(uBlock, url)
}

// Contains reports whether a member starts with member. A line that is not
// a member panics, as upstream throws ClassCastException.
func (a *MethodsOrFieldsArea) Contains(member string) bool {
	for _, cs := range a.members.AsList() {
		att := asMember(cs)
		if strings.HasPrefix(att.Display(false), member) {
			return true
		}
	}
	return false
}

func asMember(cs klimt.CharSequence) *Member {
	m, ok := cs.(*Member)
	if !ok {
		panic(&color.RuntimeError{Class: "java.lang.ClassCastException", Message: "class " +
			"java.lang.String cannot be cast to class net.sourceforge.plantuml.cucadiagram.Member"})
	}
	return m
}

// InnerPosition returns the position of a member, its visibility icon
// included.
func (a *MethodsOrFieldsArea) InnerPosition(member string, stringBounder font.StringBounder) *geom.XRectangle2D {
	group := a.layout(stringBounder)
	dim := a.CalculateDimension(stringBounder)
	result := group.InnerPosition(member, dim.Width(), dim.Height(), stringBounder)
	if result != nil && a.hasSmallIcon() {
		// Extend the inner position to include the visibility modifier on
		// its left.
		smallIcon := float64(a.skinParam.CircledCharacterRadius() + 3)
		r := geom.NewXRectangle2D(result.MinX()-smallIcon, result.MinY(), result.Width+smallIcon, result.Height)
		return &r
	}
	return result
}

func (a *MethodsOrFieldsArea) layout(stringBounder font.StringBounder) *klimt.ULayoutGroup {
	var group *klimt.ULayoutGroup
	if a.hasSmallIcon() {
		group = klimt.NewULayoutGroup(klimt.NewPlacementStrategyVisibility(stringBounder,
			a.skinParam.CircledCharacterRadius()+3))
		for _, cs := range a.members.AsList() {
			block := a.createTextBlock(cs)
			att := asMember(cs)
			modifier, ok := att.VisibilityModifier()
			group.Add(a.uBlock(modifier, ok, att.URL()))
			group.Add(block)
		}
		return group
	}
	var placementStrategy klimt.PlacementStrategy
	switch a.align {
	case geom.HorizontalAlignmentLeft:
		placementStrategy = klimt.NewPlacementStrategyY1Y2Left(stringBounder)
	case geom.HorizontalAlignmentCenter:
		placementStrategy = klimt.NewPlacementStrategyY1Y2Center(stringBounder)
	default:
		placementStrategy = klimt.NewPlacementStrategyY1Y2Right(stringBounder)
	}
	group = klimt.NewULayoutGroup(placementStrategy)
	for _, cs := range a.members.AsList() {
		group.Add(a.createTextBlock(cs))
	}
	return group
}

// DrawU draws the members, then the embedded diagrams below them.
func (a *MethodsOrFieldsArea) DrawU(ug klimt.UGraphic) {
	stringBounder := ug.StringBounder()
	group := a.layout(stringBounder)
	dim := a.calculateDimensionOnlyMembers(stringBounder)
	group.DrawU(ug, dim.Width(), dim.Height())
	ug = ug.Apply(klimt.UTranslateDy(dim.Height()))
	for _, embedded := range a.embeddeds {
		embedded.DrawU(ug)
		ug = ug.Apply(klimt.UTranslateDy(embedded.CalculateDimension(stringBounder).Height()))
	}
}
