// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/skin/SkinParam.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/text"
)

// SkinParamDefaultPreserveAspectRatio is the default SVG
// preserveAspectRatio.
const SkinParamDefaultPreserveAspectRatio = "none"

// SkinParamDefaultSkin is the default style sheet.
const SkinParamDefaultSkin = "plantuml.skin"

// SkinParamIsDark reports whether "skinparam mode" is "dark".
func SkinParamIsDark(skinParam ISkinParam) bool {
	v, ok := skinParam.Value("mode")
	return ok && javaEqualsIgnoreCase("dark", v)
}

// optString is a string that may be missing (null upstream).
type optString struct {
	s  string
	ok bool
}

// sheetCacheEntry is an entry of upstream's creole parser cache, keyed by
// (alignment, mode, font configuration, stereotype font configuration).
type sheetCacheEntry struct {
	alignment geom.HorizontalAlignment
	mode      CreoleMode
	fc        *FontConfiguration
	stereo    *FontConfiguration
	parser    SheetBuilder
}

// SkinParam is the skin parameters of a diagram: the "skinparam" values and
// the style sheet they are converted into.
//
// As upstream, Value caches its results, including missing values, and the
// cache is not cleared by SetParam: a value read before it is set stays
// missing.
type SkinParam struct {
	skin         string
	styleBuilder *StyleBuilder
	pragma       *Pragma
	pathSystem   *core.RenderContext
	option       *ConfigurationStore[OptionKey]
	md5map       map[string]string

	params                         map[string]string
	paramsPendingForStyleMigration *orderedMap[string, string]
	svgCharSizes                   *orderedMap[string, string]
	rankdir                        geom.Rankdir
	diagramType                    core.DiagramType
	useVizJs                       bool

	cacheCleanForKey map[string][]string
	cacheKeyValue    map[string]optString
	cacheThickness   map[string]*UStroke
	sprites          map[string]Sprite
	paramSameClass   float64
	sheetCache       []sheetCacheEntry
}

var _ ISkinParam = (*SkinParam)(nil)

var (
	skinDigits       = regexp.MustCompile(`^[0-9]+$`)
	skinDigitsDot    = regexp.MustCompile(`^[0-9.]+$`)
	skinIntOrDecimal = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
)

func isDigits(s string, ok bool) bool       { return ok && skinDigits.MatchString(s) }
func isDigitsOrDot(s string, ok bool) bool  { return ok && skinDigitsDot.MatchString(s) }
func isIntOrDecimal(s string, ok bool) bool { return ok && skinIntOrDecimal.MatchString(s) }

// SkinParamCreate returns the skin parameters of a diagram. ctx stands for
// upstream's PathSystem; it may be nil.
func SkinParamCreate(md5map map[string]string, ctx *core.RenderContext, diagramType core.DiagramType,
	pragma *Pragma, option *ConfigurationStore[OptionKey]) *SkinParam {
	return &SkinParam{skin: SkinParamDefaultSkin, diagramType: diagramType, pragma: pragma, option: option,
		pathSystem: ctx, md5map: md5map, params: map[string]string{},
		paramsPendingForStyleMigration: newOrderedMap[string, string](),
		svgCharSizes:                   newOrderedMap[string, string](), rankdir: geom.RankdirTopToBottom,
		cacheCleanForKey: map[string][]string{}, cacheKeyValue: map[string]optString{},
		cacheThickness: map[string]*UStroke{}, sprites: map[string]Sprite{}}
}

// CurrentStyleBuilder returns the style sheet, loading the default skin on
// first use.
func (s *SkinParam) CurrentStyleBuilder() *StyleBuilder {
	if s.styleBuilder == nil {
		s.styleBuilder = s.currentStyleBuilderInternal()
	}
	return s.styleBuilder
}

// MuteStyle merges styles into the style sheet.
func (s *SkinParam) MuteStyle(modifiedStyles []*Style) {
	s.styleBuilder = s.CurrentStyleBuilder().MuteStyle(modifiedStyles)
}

// DefaultSkin returns the style sheet file name.
func (s *SkinParam) DefaultSkin() string { return s.skin }

// SetDefaultSkin changes the style sheet file name.
func (s *SkinParam) SetDefaultSkin(newSkin string) { s.skin = newSkin }

// currentStyleBuilderInternal loads the skin, falling back on the default
// skin when it cannot be loaded. It never returns nil.
func (s *SkinParam) currentStyleBuilderInternal() *StyleBuilder {
	if b, err := StyleLoaderLoadSkin(s.pathSystem, s.DefaultSkin()); err == nil {
		return b
	}
	b, err := StyleLoaderLoadSkin(s.pathSystem, SkinParamDefaultSkin)
	if err != nil {
		throw(illegalState, "Cannot load "+SkinParamDefaultSkin)
	}
	return b
}

// SkinParamZeroMargin returns defaultValue.
func SkinParamZeroMargin(defaultValue int) int { return defaultValue }

const skinStereoPatternString = `\<\<(.*?)\>\>`

var skinStereoPattern = regexp.MustCompile(javaDots(`(?i)<<(.*?)>>`))

// CopyAllFrom copies skinparam values.
func (s *SkinParam) CopyAllFrom(other map[string]string) { maps.Copy(s.params, other) }

// Values returns the skinparam values; the map must not be modified.
func (s *SkinParam) Values() map[string]string { return s.params }

// SetParam sets a skinparam and converts it to styles. "style strictuml"
// also loads strictuml.skin.
func (s *SkinParam) SetParam(key, value string) {
	for _, key2 := range s.cleanForKey(key) {
		s.params[key2] = trin(value)
		s.ApplyPendingStyleMigration()
		converter := NewFromSkinparamToStyle(key2)
		converter.ConvertNow(value, s.CurrentStyleBuilder())
		s.MuteStyle(converter.Styles())
	}
	if javaEqualsIgnoreCase("style", key) && javaEqualsIgnoreCase("strictuml", value) {
		styleBuilder := s.CurrentStyleBuilder()
		lines, err := embeddedSkinLines("strictuml.skin")
		if err == nil {
			if styles, err := StyleLoaderParseStyleText(lines, styleBuilder); err == nil {
				s.MuteStyle(styles)
			}
		}
	}
}

// ApplyPendingStyleMigration converts the pending skinparams to styles.
// Upstream never adds pending skinparams, so this does nothing in practice.
func (s *SkinParam) ApplyPendingStyleMigration() {
	for _, k := range s.paramsPendingForStyleMigration.keys {
		converter := NewFromSkinparamToStyle(k)
		converter.ConvertNow(s.paramsPendingForStyleMigration.values[k], s.CurrentStyleBuilder())
		s.MuteStyle(converter.Styles())
	}
	s.paramsPendingForStyleMigration = newOrderedMap[string, string]()
}

func (s *SkinParam) cleanForKey(key string) []string {
	if result, ok := s.cacheCleanForKey[key]; ok {
		return result
	}
	result := cleanForKeySlow(key)
	s.cacheCleanForKey[key] = result
	return result
}

var (
	patternCleanUnderscoreDot = regexp.MustCompile(`_|\.`)
	patternCleanSequence      = regexp.MustCompile(`sequence(participant|actor)`)
	patternCleanArrow         = regexp.MustCompile(`(activity|class|component|object|sequence|state|usecase)arrow`)
	patternCleanAlign         = regexp.MustCompile(`align$`)
	patternStereoRemove       = regexp.MustCompile(javaDots(skinStereoPatternString))
)

// cleanForKeySlow normalizes a skinparam key: lower case, without "_" and
// ".", with the aliases of participants, arrows and alignments; one key per
// stereotype of the key.
func cleanForKeySlow(key string) []string {
	key = trin(javaLower(key))
	key = patternCleanUnderscoreDot.ReplaceAllString(key, "")
	key = patternCleanSequence.ReplaceAllString(key, "$1")
	key = patternCleanArrow.ReplaceAllString(key, "arrow")
	key = patternCleanAlign.ReplaceAllString(key, "alignment")
	var result []string
	for _, m := range skinStereoPattern.FindAllStringSubmatch(key, -1) {
		result = append(result, patternStereoRemove.ReplaceAllString(key, "")+"<<"+m[1]+">>")
	}
	if len(result) == 0 {
		result = append(result, key)
	}
	return result
}

// HyperlinkColor returns the color of links, blue by default.
func (s *SkinParam) HyperlinkColor() color.HColor {
	if result := s.HtmlColor(ColorParamHyperlink, nil, false); result != nil {
		return result
	}
	return color.Blue
}

// BackgroundColor returns the background color: the skinparam, else the
// document style.
func (s *SkinParam) BackgroundColor() color.HColor {
	if result := s.HtmlColor(ColorParamBackground, nil, false); result != nil {
		return result
	}
	style := s.CurrentStyleBuilder().MergedStyle(StyleSignatureBasicOf(SNameRoot, SNameDocument))
	return style.Value(PNameBackGroundColor).AsColor(s.IHtmlColorSet())
}

// Value returns a skinparam value, trying each cleaned key; the result is
// cached.
func (s *SkinParam) Value(key string) (string, bool) {
	if v, ok := s.cacheKeyValue[key]; ok {
		return v.s, v.ok
	}
	s.ApplyPendingStyleMigration()
	for _, key2 := range s.cleanForKey(key) {
		if result, ok := s.params[key2]; ok {
			s.cacheKeyValue[key] = optString{result, true}
			return result, true
		}
	}
	s.cacheKeyValue[key] = optString{}
	return "", false
}

func (s *SkinParam) valueOr(key, defaultValue string) string {
	if v, ok := s.Value(key); ok {
		return v
	}
	return defaultValue
}

func (s *SkinParam) valueIs(key, expected string) bool {
	v, ok := s.Value(key)
	return ok && javaEqualsIgnoreCase(expected, v)
}

func (s *SkinParam) isTrue(key string) bool { return s.valueIs(key, "true") }

// SkinParamHumanName is upstream's humanName: "CLASS_ATTRIBUTE" gives
// "ClassAttribute".
func SkinParamHumanName(key string) string {
	var sb strings.Builder
	upper := true
	for _, c := range key {
		if c == '_' {
			upper = true
		} else {
			if upper {
				sb.WriteString(javaUpper(string(c)))
			} else {
				sb.WriteString(javaLower(string(c)))
			}
			upper = false
		}
	}
	return sb.String()
}

// HtmlColor returns the color of a parameter: for a stereotype, the
// "<param>color<<label>>" value of its first label that has one; else the
// "<param>color" value (with "Clickable" for clickable elements); nil when
// none is set. A transparent background or arrow head is "transparent" or
// "none"; an unknown color is white.
func (s *SkinParam) HtmlColor(param ColorParam, stereotype *Stereotype, clickable bool) color.HColor {
	if stereotype != nil {
		for _, label := range stereotype.MultipleLabels() {
			if value2, ok := s.Value(param.String() + "color" + "<<" + label + ">>"); ok {
				return s.IHtmlColorSet().GetColorOrWhite(value2)
			}
		}
	}
	value, ok := s.Value(paramName(param, clickable))
	if !ok {
		return nil
	}
	if (param == ColorParamBackground || param == ColorParamArrowHead) &&
		(javaEqualsIgnoreCase(value, "transparent") || javaEqualsIgnoreCase(value, "none")) {
		return color.Transparent()
	}
	return s.IHtmlColorSet().GetColorOrWhite(value)
}

// CircledCharacter returns the first character of the "spotchar<<label>>"
// value of a stereotype, 0 for none.
func (s *SkinParam) CircledCharacter(stereotype *Stereotype) rune {
	if stereotype == nil {
		throw(nullPointer, "")
	}
	value2, ok := s.Value("spotchar" + stereotype.labelOrNull(GuillemetDoubleComparator()))
	if ok && value2 != "" {
		return javaCharAt(value2, 0)
	}
	return 0
}

// Colors returns the colors of a parameter, for a stereotype first.
func (s *SkinParam) Colors(param ColorParam, stereotype *Stereotype) (*color.Colors, error) {
	if stereotype != nil {
		if value2, ok := s.Value(param.String() + "color" +
			stereotype.labelOrNull(GuillemetDoubleComparator())); ok {
			return newColorsOfParam(value2, s.IHtmlColorSet(), param)
		}
	}
	value, ok := s.Value(paramName(param, false))
	if !ok {
		return color.EmptyColors(), nil
	}
	return newColorsOfParam(value, s.IHtmlColorSet(), param)
}

// newColorsOfParam is upstream's new Colors(value, set, param.getColorType()).
// For the parameters without color type upstream passes null, and the first
// plain color part makes its EnumMap throw NullPointerException; the parts
// are checked in order so that the same exception (or error) comes first.
func newColorsOfParam(value string, set *color.HColorSet, param ColorParam) (*color.Colors, error) {
	t, ok := param.ColorType()
	if ok {
		return color.NewColors(value, set, t)
	}
	data := strings.ReplaceAll(javaLower(value), "#", "")
	for _, part := range javaTokenize(data, ";") {
		name, _, found := strings.Cut(part, ":")
		if found {
			if !javaEqualsIgnoreCase(name, "shadowing") {
				color.ColorTypeOf(name)
			}
			continue
		}
		if !strings.Contains(part, ".") {
			if _, err := set.GetColor(part); err != nil {
				return nil, err
			}
			throw(nullPointer, "")
		}
	}
	return color.NewColors(value, set, color.TypeBack)
}

func paramName(param ColorParam, clickable bool) string {
	n := param.String()
	if clickable && strings.HasSuffix(n, "Background") {
		n = strings.ReplaceAll(n, "Background", "ClickableBackground")
	} else if clickable && strings.HasSuffix(n, "Border") {
		n = strings.ReplaceAll(n, "Border", "ClickableBorder")
	}
	return n + "color"
}

func (s *SkinParam) fontSize(stereotype *Stereotype, param ...FontParam) int {
	if stereotype != nil {
		value2, ok := s.firstValueNonNullWithSuffix("fontsize"+stereotype.labelOrNull(GuillemetDoubleComparator()),
			param...)
		if isDigits(value2, ok) {
			return javaParseInt(value2)
		}
	}
	value, ok := s.firstValueNonNullWithSuffix("fontsize", param...)
	if !isDigits(value, ok) {
		value, ok = s.Value("defaultfontsize")
	}
	if !isDigits(value, ok) {
		return param[0].DefaultSize(s)
	}
	return javaParseInt(value)
}

func (s *SkinParam) fontFamily(stereotype *Stereotype, param ...FontParam) string {
	if stereotype != nil {
		if value2, ok := s.firstValueNonNullWithSuffix(
			"fontname"+stereotype.labelOrNull(GuillemetDoubleComparator()), param...); ok {
			return eventuallyRemoveStartingAndEndingDoubleQuote(value2)
		}
	}
	// Times, Helvetica, Courier or Symbol
	if value, ok := s.firstValueNonNullWithSuffix("fontname", param...); ok {
		return eventuallyRemoveStartingAndEndingDoubleQuote(value)
	}
	if param[0] != FontParamCircledCharacter {
		if value, ok := s.Value("defaultfontname"); ok {
			return eventuallyRemoveStartingAndEndingDoubleQuote(value)
		}
	}
	return param[0].DefaultFamily()
}

// FontHtmlColor returns the font color of the parameters: for the
// stereotype, else the parameters, else the default font color, else the
// default of the first parameter.
func (s *SkinParam) FontHtmlColor(stereotype *Stereotype, param ...FontParam) color.HColor {
	value, ok := "", false
	if stereotype != nil {
		value, ok = s.firstValueNonNullWithSuffix("fontcolor"+stereotype.labelOrNull(GuillemetDoubleComparator()),
			param...)
	}
	if !ok {
		value, ok = s.firstValueNonNullWithSuffix("fontcolor", param...)
	}
	if !ok {
		value, ok = s.Value("defaultfontcolor")
	}
	if !ok {
		value, ok = param[0].DefaultColor(), true
	}
	return s.IHtmlColorSet().GetColorOrWhite(value)
}

func (s *SkinParam) firstValueNonNullWithSuffix(suffix string, param ...FontParam) (string, bool) {
	for _, p := range param {
		if v, ok := s.Value(p.String() + suffix); ok {
			return v, true
		}
	}
	return "", false
}

func (s *SkinParam) fontFace(stereotype *Stereotype, inPackageTitle bool, param ...FontParam) font.UFontFace {
	value, ok := "", false
	if stereotype != nil {
		value, ok = s.firstValueNonNullWithSuffix("fontstyle"+stereotype.labelOrNull(GuillemetDoubleComparator()),
			param...)
	}
	if !ok {
		value, ok = s.firstValueNonNullWithSuffix("fontstyle", param...)
	}
	if !ok {
		value, ok = s.Value("defaultfontstyle")
	}
	if !ok {
		return param[0].DefaultFontFace(s, inPackageTitle)
	}
	lower := javaLower(value)
	bold := strings.Contains(lower, "bold")
	italic := strings.Contains(lower, "italic")
	switch {
	case bold && italic:
		return font.BoldItalicFace()
	case bold:
		return font.BoldFace()
	case italic:
		return font.ItalicFace()
	}
	return font.NormalFace()
}

// Font returns the font of the parameters for a stereotype.
func (s *SkinParam) Font(stereotype *Stereotype, inPackageTitle bool, fontParam ...FontParam) font.UFont {
	fontFamily := s.fontFamily(stereotype, fontParam...)
	face := s.fontFace(stereotype, inPackageTitle, fontParam...)
	fontSize := s.fontSize(stereotype, fontParam...)
	return font.Build(fontFamily, face, fontSize)
}

// CircledCharacterRadius returns the radius of class spots.
func (s *SkinParam) CircledCharacterRadius() int {
	value := s.asInt("circledCharacterRadius", -1)
	if value == -1 {
		return s.fontSize(nil, FontParamCircledCharacter)/3 + 6
	}
	return value
}

// ClassAttributeIconSize returns the size of visibility icons.
func (s *SkinParam) ClassAttributeIconSize() int { return s.asInt("classAttributeIconSize", 10) }

// SkinParamPossibleValues returns the known skinparam names, sorted.
func SkinParamPossibleValues() []string {
	set := map[string]bool{}
	for _, v := range []string{"Monochrome", "CircledCharacterRadius", "ClassAttributeIconSize",
		"DefaultFontName", "DefaultFontStyle", "DefaultFontSize", "DefaultFontColor", "MinClassWidth", "Dpi",
		"DefaultTextAlignment", "Shadowing", "NoteShadowing", "Handwritten", "Linetype", "PackageStyle",
		"ComponentStyle", "StereotypePosition", "Nodesep", "Ranksep", "RoundCorner", "TitleBorderRoundCorner",
		"MaxMessageSize", "Style", "SequenceParticipant", "ConditionStyle", "ConditionEndStyle", "SameClassWidth",
		"HyperlinkUnderline", "Padding", "BoxPadding", "ParticipantPadding", "Guillemet", "SvglinkTarget",
		"DefaultMonospacedFontName", "TabSize", "MaxAsciiMessageLength", "ColorArrowSeparationSpace",
		"ResponseMessageBelowArrow", "GenericDisplay", "PathHoverColor", "SwimlaneWidth", "PageBorderColor",
		"PageExternalColor", "PageMargin", "WrapWidth", "SwimlaneWrapTitleWidth", "FixCircleLabelOverlapping"} {
		set[v] = true
	}
	for p := FontParam(0); p < fontParamCount; p++ {
		h := SkinParamHumanName(p.String())
		set[h+"FontStyle"] = true
		set[h+"FontName"] = true
		set[h+"FontSize"] = true
		set[h+"FontColor"] = true
	}
	for p := ColorParam(0); p < colorParamCount; p++ {
		set[capitalize(p.String())+"Color"] = true
	}
	for p := LineParam(0); p < lineParamCount; p++ {
		set[capitalize(p.String())+"Thickness"] = true
	}
	for p := AlignmentParam(0); p < alignmentParamCount; p++ {
		set[capitalize(p.String())] = true
	}
	return slices.Sorted(maps.Keys(set))
}

func capitalize(name string) string {
	r := []rune(name)
	return javaUpper(string(r[0])) + string(r[1:])
}

// Dpi returns the resolution, 96 by default.
func (s *SkinParam) Dpi() int {
	const defaultValue = 96
	dpi := s.asInt("dpi", defaultValue)
	if dpi <= 0 {
		return defaultValue
	}
	return dpi
}

// DotSplines returns the Graphviz splines of "skinparam linetype".
func (s *SkinParam) DotSplines() DotSplines {
	value, ok := s.Value("linetype")
	if ok && javaEqualsIgnoreCase("polyline", value) {
		return DotSplinesPolyline
	}
	if ok && javaEqualsIgnoreCase("ortho", value) {
		return DotSplinesOrtho
	}
	return DotSplinesSplines
}

// HorizontalAlignment returns the alignment of a parameter: "first",
// "direction" and "reversedirection" depend on the arrow direction.
func (s *SkinParam) HorizontalAlignment(param AlignmentParam, arrowDirection ArrowDirection,
	isReverseDefine bool, overrideDefault geom.HorizontalAlignment) geom.HorizontalAlignment {
	var value string
	var ok bool
	switch param {
	case AlignmentParamSequenceMessageAlignment:
		value, ok = skinArg(s.Value(AlignmentParamSequenceMessageAlignment.String()))(0)
	case AlignmentParamSequenceMessageTextAlignment:
		value, ok = skinArg(s.Value(AlignmentParamSequenceMessageAlignment.String()))(1)
	default:
		value, ok = s.Value(param.String())
	}
	if ok && javaEqualsIgnoreCase("first", value) {
		if arrowDirection == ArrowDirectionRightToLeftReverse {
			if isReverseDefine {
				return geom.HorizontalAlignmentLeft
			}
			return geom.HorizontalAlignmentRight
		}
		if isReverseDefine {
			return geom.HorizontalAlignmentRight
		}
		return geom.HorizontalAlignmentLeft
	}
	if ok && javaEqualsIgnoreCase("direction", value) {
		switch arrowDirection {
		case ArrowDirectionLeftToRightNormal:
			return geom.HorizontalAlignmentLeft
		case ArrowDirectionRightToLeftReverse:
			return geom.HorizontalAlignmentRight
		case ArrowDirectionBothDirection:
			return geom.HorizontalAlignmentCenter
		}
	}
	if ok && javaEqualsIgnoreCase("reversedirection", value) {
		switch arrowDirection {
		case ArrowDirectionLeftToRightNormal:
			return geom.HorizontalAlignmentRight
		case ArrowDirectionRightToLeftReverse:
			return geom.HorizontalAlignmentLeft
		case ArrowDirectionBothDirection:
			return geom.HorizontalAlignmentCenter
		}
	}
	result := HorizontalAlignmentNull
	if ok {
		result = horizontalAlignmentFromString(value)
	}
	if result == HorizontalAlignmentNull && param == AlignmentParamNoteTextAlignment {
		if overrideDefault == HorizontalAlignmentNull {
			return s.DefaultTextAlignment(geom.HorizontalAlignmentLeft)
		}
		return s.DefaultTextAlignment(overrideDefault)
	} else if result == HorizontalAlignmentNull && param == AlignmentParamStateMessageAlignment {
		return s.DefaultTextAlignment(geom.HorizontalAlignmentCenter)
	} else if result == HorizontalAlignmentNull {
		return param.DefaultValue()
	}
	return result
}

// DefaultTextAlignment returns "skinparam defaultTextAlignment", or
// defaultValue.
func (s *SkinParam) DefaultTextAlignment(defaultValue geom.HorizontalAlignment) geom.HorizontalAlignment {
	value, ok := s.Value("defaulttextalignment")
	if !ok {
		return defaultValue
	}
	if result := horizontalAlignmentFromString(value); result != HorizontalAlignmentNull {
		return result
	}
	return defaultValue
}

// StereotypeAlignment returns the alignment of stereotypes, centered by
// default.
func (s *SkinParam) StereotypeAlignment() geom.HorizontalAlignment {
	value, ok := s.Value("stereotypealignment")
	if !ok {
		return geom.HorizontalAlignmentCenter
	}
	if result := horizontalAlignmentFromString(value); result != HorizontalAlignmentNull {
		return result
	}
	return geom.HorizontalAlignmentCenter
}

// skinArg returns upstream's getArg(value, i): the part i of a
// ":"-separated value, or its first part.
func skinArg(value string, ok bool) func(i int) (string, bool) {
	return func(i int) (string, bool) {
		if !ok {
			return "", false
		}
		split := javaSplit(value, ":")
		if len(split) == 0 {
			// Java's split of a value made of separators only gives an empty
			// array, and split[0] throws.
			throw("java.lang.ArrayIndexOutOfBoundsException", "Index 0 out of bounds for length 0")
		}
		if i >= len(split) {
			return split[0], true
		}
		return split[i], true
	}
}

// Shadowing reports whether elements of a stereotype have shadows.
func (s *SkinParam) Shadowing(stereotype *Stereotype) bool {
	if stereotype != nil {
		if value2, ok := s.Value("shadowing" + stereotype.labelOrNull(GuillemetDoubleComparator())); ok {
			return javaEqualsIgnoreCase(value2, "true")
		}
	}
	value, ok := s.Value("shadowing")
	if ok && javaEqualsIgnoreCase("false", value) {
		return false
	}
	if ok && javaEqualsIgnoreCase("true", value) {
		return true
	}
	if s.StrictUmlStyle() {
		return false
	}
	return true
}

// ShadowingForNote reports whether notes of a stereotype have shadows.
func (s *SkinParam) ShadowingForNote(stereotype *Stereotype) bool {
	if stereotype != nil {
		if value2, ok := s.Value("note" + "shadowing" + stereotype.labelOrNull(GuillemetDoubleComparator())); ok {
			return javaEqualsIgnoreCase(value2, "true")
		}
	}
	if value2, ok := s.Value("note" + "shadowing"); ok {
		return javaEqualsIgnoreCase(value2, "true")
	}
	return s.Shadowing(stereotype)
}

// AllSpriteNames returns the names of the declared sprites, sorted.
func (s *SkinParam) AllSpriteNames() []string {
	names := slices.Collect(maps.Keys(s.sprites))
	slices.SortFunc(names, javaStringCompare)
	return names
}

// AddSprite declares a sprite.
func (s *SkinParam) AddSprite(name string, sprite Sprite) { s.sprites[name] = sprite }

// Sprite returns a declared or internal sprite, nil when there is none.
func (s *SkinParam) Sprite(name string) Sprite {
	if result := s.sprites[name]; result != nil {
		return result
	}
	return SpriteImageFromInternal(name)
}

// PackageStyle returns "skinparam packageStyle", folder by default.
func (s *SkinParam) PackageStyle() PackageStyle {
	value, ok := s.Value("packageStyle")
	if !ok {
		return PackageStyleFolder
	}
	if p, ok := PackageStyleFromString(value); ok {
		return p
	}
	return PackageStyleFolder
}

// ComponentStyle returns "skinparam componentStyle", UML2 by default and in
// strict UML style.
func (s *SkinParam) ComponentStyle() ComponentStyle {
	if s.StrictUmlStyle() {
		return ComponentStyleUML2
	}
	value, ok := s.Value("componentstyle")
	switch {
	case ok && javaEqualsIgnoreCase("uml1", value):
		return ComponentStyleUML1
	case ok && javaEqualsIgnoreCase("uml2", value):
		return ComponentStyleUML2
	case ok && javaEqualsIgnoreCase("rectangle", value):
		return ComponentStyleRectangle
	}
	return ComponentStyleUML2
}

// StereotypePositionTop reports whether stereotypes are above names.
func (s *SkinParam) StereotypePositionTop() bool { return !s.valueIs("stereotypePosition", "bottom") }

// UseSwimlanes reports whether an activity diagram uses swimlanes.
func (s *SkinParam) UseSwimlanes(diagramType core.DiagramType) bool {
	if diagramType != core.Activity {
		return false
	}
	return s.Swimlanes()
}

// Swimlanes reports whether "skinparam swimlane(s)" is true.
func (s *SkinParam) Swimlanes() bool { return s.isTrue("swimlane") || s.isTrue("swimlanes") }

// Nodesep returns the Graphviz nodesep, an integer.
func (s *SkinParam) Nodesep() float64 { return float64(s.asInt("nodesep", 0)) }

// Ranksep returns the Graphviz ranksep, an integer.
func (s *SkinParam) Ranksep() float64 { return float64(s.asInt("ranksep", 0)) }

// DiagonalCorner returns the diagonal corner of a parameter for a
// stereotype, falling back on the default corner.
func (s *SkinParam) DiagonalCorner(param CornerParam, stereotype *Stereotype) float64 {
	key := param.DiagonalKey()
	if result, ok := s.cornerInternal(key, stereotype); ok {
		return result
	}
	if result, ok := s.cornerInternal(key, nil); ok {
		return result
	}
	if param == CornerParamDefault {
		return 0
	}
	return s.DiagonalCorner(CornerParamDefault, stereotype)
}

// RoundCorner returns the round corner of a parameter for a stereotype,
// falling back on the default corner.
func (s *SkinParam) RoundCorner(param CornerParam, stereotype *Stereotype) float64 {
	key := param.RoundKey()
	if result, ok := s.cornerInternal(key, stereotype); ok {
		return result
	}
	if result, ok := s.cornerInternal(key, nil); ok {
		return result
	}
	if param == CornerParamDefault {
		return 0
	}
	return s.RoundCorner(CornerParamDefault, stereotype)
}

func (s *SkinParam) cornerInternal(key string, stereotype *Stereotype) (float64, bool) {
	if stereotype != nil {
		key += stereotype.labelOrNull(GuillemetDoubleComparator())
	}
	value, ok := s.Value(key)
	if isDigits(value, ok) {
		return javaParseDouble(value), true
	}
	return 0, false
}

// Thickness returns the stroke of a line parameter for a stereotype, nil
// when neither its thickness nor a non-normal style is set.
func (s *SkinParam) Thickness(param LineParam, stereotype *Stereotype) *UStroke {
	key := param.String()
	if stereotype != nil {
		key = param.String() + "\x00" + stereotype.labelOrNull(GuillemetDoubleComparator())
	}
	if result, ok := s.cacheThickness[key]; ok {
		return result
	}
	result := s.thicknessSlow(param, stereotype)
	s.cacheThickness[key] = result
	return result
}

// linkStyle is the part of upstream's decoration.LinkStyle that thickness
// parameters use: a type and an optional thickness.
type linkStyle struct {
	style        color.LineStyle
	thickness    float64
	hasThickness bool
}

func linkStyleFromString2(s string) (linkStyle, bool) {
	switch {
	case javaEqualsIgnoreCase("dashed", s):
		return linkStyle{style: color.LineStyleDashed}, true
	case javaEqualsIgnoreCase("dotted", s):
		return linkStyle{style: color.LineStyleDotted}, true
	case javaEqualsIgnoreCase("bold", s):
		return linkStyle{style: color.LineStyleBold}, true
	case javaEqualsIgnoreCase("hidden", s):
		return linkStyle{style: color.LineStyleInvisible}, true
	}
	return linkStyle{}, false
}

func (l linkStyle) stroke3() *UStroke {
	t := 1.0
	if l.hasThickness {
		t = l.thickness
	}
	var st UStroke
	switch l.style {
	case color.LineStyleDashed:
		st = NewUStroke(7, 7, t)
	case color.LineStyleDotted:
		st = NewUStroke(1, 3, t)
	case color.LineStyleBold:
		st = UStrokeWithThickness(2)
	default:
		st = UStrokeWithThickness(t)
	}
	return &st
}

func (s *SkinParam) thicknessSlow(param LineParam, stereotype *Stereotype) *UStroke {
	var style *linkStyle
	if stereotype != nil {
		label := stereotype.labelOrNull(GuillemetDoubleComparator())
		if styleValue, ok := s.Value(param.String() + "style" + label); ok {
			if ls, found := linkStyleFromString2(styleValue); found {
				style = &ls
			}
		}
		value2, ok := s.Value(param.String() + "thickness" + label)
		if isDigitsOrDot(value2, ok) {
			if style == nil {
				style = &linkStyle{style: color.LineStyleNormal}
			}
			return linkStyle{style: style.style, thickness: javaParseDouble(value2), hasThickness: true}.stroke3()
		}
	}
	value, ok := s.Value(param.String() + "thickness")
	if isDigitsOrDot(value, ok) {
		if style == nil {
			style = &linkStyle{style: color.LineStyleNormal}
		}
		return linkStyle{style: style.style, thickness: javaParseDouble(value), hasThickness: true}.stroke3()
	}
	if style == nil {
		if styleValue, ok := s.Value(param.String() + "style"); ok {
			if ls, found := linkStyleFromString2(styleValue); found {
				style = &ls
			}
		}
	}
	if style != nil && style.style != color.LineStyleNormal {
		return style.stroke3()
	}
	return nil
}

// MaxMessageSize returns "skinparam wrapMessageWidth" or "maxMessageSize".
func (s *SkinParam) MaxMessageSize() LineBreakStrategy {
	value, ok := s.Value("wrapmessagewidth")
	if !ok {
		value, ok = s.Value("maxmessagesize")
	}
	return NewLineBreakStrategyOpt(value, ok)
}

// SwimlaneWrapTitleWidth returns "skinparam swimlaneWrapTitleWidth".
func (s *SkinParam) SwimlaneWrapTitleWidth() LineBreakStrategy {
	return NewLineBreakStrategyOpt(s.Value("swimlanewraptitlewidth"))
}

// StrictUmlStyle reports whether "skinparam style strictuml" is set.
func (s *SkinParam) StrictUmlStyle() bool { return s.valueIs("style", "strictuml") }

// ForceSequenceParticipantUnderlined reports whether participants are
// underlined.
func (s *SkinParam) ForceSequenceParticipantUnderlined() bool {
	return s.valueIs("sequenceParticipant", "underline")
}

// ConditionStyle returns "skinparam conditionStyle", inside hexagon by
// default.
func (s *SkinParam) ConditionStyle() ConditionStyle {
	if value, ok := s.Value("conditionStyle"); ok {
		if p, ok := ConditionStyleFromString(value); ok {
			return p
		}
	}
	return ConditionStyleInsideHexagon
}

// ConditionEndStyle returns "skinparam conditionEndStyle", diamond by
// default.
func (s *SkinParam) ConditionEndStyle() ConditionEndStyle {
	if value, ok := s.Value("conditionEndStyle"); ok {
		if p, ok := ConditionEndStyleFromString(value); ok {
			return p
		}
	}
	return ConditionEndStyleDiamond
}

// SameClassWidth reports whether classes have the same width.
func (s *SkinParam) SameClassWidth() bool { return s.isTrue("sameclasswidth") }

// Rankdir returns the layout direction.
func (s *SkinParam) Rankdir() geom.Rankdir { return s.rankdir }

// SetRankdir changes the layout direction.
func (s *SkinParam) SetRankdir(rankdir geom.Rankdir) { s.rankdir = rankdir }

// UseOctagonForActivity reports whether activities are octagons.
func (s *SkinParam) UseOctagonForActivity(stereotype *Stereotype) bool {
	value, ok := s.Value("activityshape")
	if stereotype != nil {
		if value2, ok2 := s.Value("activityshape" + stereotype.labelOrNull(GuillemetDoubleComparator())); ok2 {
			value, ok = value2, true
		}
	}
	if ok && javaEqualsIgnoreCase("roundedbox", value) {
		return false
	}
	if ok && javaEqualsIgnoreCase("octagon", value) {
		return true
	}
	return false
}

// IHtmlColorSet returns the color set.
func (s *SkinParam) IHtmlColorSet() *color.HColorSet { return color.Instance() }

// UseUnderlineForHyperlink returns the underline of links: the simple
// stroke unless "skinparam hyperlinkUnderline false".
func (s *SkinParam) UseUnderlineForHyperlink() *UStroke {
	if !s.valueIs("hyperlinkunderline", "false") {
		return strokePtr(UStrokeSimple())
	}
	return nil
}

// GroupInheritance returns "skinparam groupInheritance"; 1 or less means no
// grouping.
func (s *SkinParam) GroupInheritance() int {
	value := s.asInt("groupinheritance", maxInt32)
	if value <= 1 {
		return maxInt32
	}
	return value
}

// Guillemet returns the guillemets of stereotypes.
func (s *SkinParam) Guillemet() *Guillemet {
	value, ok := s.Value("guillemet")
	return GuillemetGuillemet().FromDescription(value, ok)
}

// Handwritten reports whether "skinparam handwritten" is true.
func (s *SkinParam) Handwritten() bool { return s.isTrue("handwritten") }

// SvgLinkTarget returns the target of SVG links, "_top" by default.
func (s *SkinParam) SvgLinkTarget() string { return s.valueOr("svglinktarget", "_top") }

// PreserveAspectRatio returns the SVG preserveAspectRatio.
func (s *SkinParam) PreserveAspectRatio() string {
	return s.valueOr("preserveaspectratio", SkinParamDefaultPreserveAspectRatio)
}

// MonospacedFamily returns the monospaced font family.
func (s *SkinParam) MonospacedFamily() string {
	return s.valueOr("defaultMonospacedFontName", ParserMonospaced)
}

// TabSize returns the tab size, 8 by default.
func (s *SkinParam) TabSize() int { return s.asInt("tabsize", 8) }

// MaxAsciiMessageLength returns the maximum message length of text output,
// -1 for none.
func (s *SkinParam) MaxAsciiMessageLength() int { return s.asInt("maxasciimessagelength", -1) }

// ColorArrowSeparationSpace returns the space between colored arrow parts.
func (s *SkinParam) ColorArrowSeparationSpace() int { return s.asInt("colorarrowseparationspace", 0) }

// SplitParam returns the page parameters of split images.
func (s *SkinParam) SplitParam() *SplitParam {
	var borderColor, externalColor color.HColor
	if border, ok := s.Value("pageBorderColor"); ok {
		borderColor = s.IHtmlColorSet().GetColorOrWhite(border)
	}
	if external, ok := s.Value("pageExternalColor"); ok {
		externalColor = s.IHtmlColorSet().GetColorOrWhite(external)
	}
	margin := s.asInt("pageMargin", 0)
	return NewSplitParam(borderColor, externalColor, margin)
}

// SwimlaneWidth returns "skinparam swimlaneWidth", ISkinParamSwimlaneWidthSame
// for "same", 0 by default.
func (s *SkinParam) SwimlaneWidth() int {
	value, ok := s.Value("swimlanewidth")
	if ok && javaEqualsIgnoreCase("same", value) {
		return ISkinParamSwimlaneWidthSame
	}
	if isDigits(value, ok) {
		return javaParseInt(value)
	}
	return 0
}

// DiagramType returns the diagram type.
func (s *SkinParam) DiagramType() core.DiagramType { return s.diagramType }

// HoverPathColor returns "skinparam pathHoverColor", nil for none.
func (s *SkinParam) HoverPathColor() color.HColor {
	value, ok := s.Value("pathhovercolor")
	if !ok {
		return nil
	}
	return s.IHtmlColorSet().GetColorOrWhite(value)
}

// Padding returns "skinparam padding" on every side.
func (s *SkinParam) Padding() ClockwiseTopRightBottomLeft { return s.asDouble("padding") }

// PaddingTOBEREMOVED returns a padding parameter on every side.
func (s *SkinParam) PaddingTOBEREMOVED(param PaddingParam) ClockwiseTopRightBottomLeft {
	return s.asDouble(param.SkinName())
}

func (s *SkinParam) asDouble(name string) ClockwiseTopRightBottomLeft {
	value, ok := s.Value(name)
	if isIntOrDecimal(value, ok) {
		return ClockwiseSame(javaParseDouble(value))
	}
	return ClockwiseSame(0)
}

func (s *SkinParam) asInt(key string, defaultValue int) int {
	value, ok := s.Value(key)
	if isDigits(value, ok) {
		return javaParseInt(value)
	}
	return defaultValue
}

// UseRankSame returns false.
func (s *SkinParam) UseRankSame() bool { return false }

// DisplayGenericWithOldFashion reports whether "skinparam genericDisplay
// old" is set.
func (s *SkinParam) DisplayGenericWithOldFashion() bool { return s.valueIs("genericDisplay", "old") }

// ResponseMessageBelowArrow reports whether "skinparam
// responseMessageBelowArrow" is true.
func (s *SkinParam) ResponseMessageBelowArrow() bool { return s.isTrue("responsemessagebelowarrow") }

// TikzFontDistortion returns "skinparam tikzFont".
func (s *SkinParam) TikzFontDistortion() *TikzFontDistortion {
	return TikzFontDistortionFromValue(s.Value("tikzFont"))
}

// SvgDimensionStyle reports whether SVG sizes are in the style attribute.
func (s *SkinParam) SvgDimensionStyle() bool { return !s.valueIs("svgdimensionstyle", "false") }

// FixCircleLabelOverlapping reports whether "skinparam
// fixCircleLabelOverlapping" is true.
func (s *SkinParam) FixCircleLabelOverlapping() bool { return s.isTrue("fixcirclelabeloverlapping") }

// SetUseVizJs records the layout engine choice.
func (s *SkinParam) SetUseVizJs(useVizJs bool) { s.useVizJs = useVizJs }

// IsUseVizJs reports the layout engine choice.
func (s *SkinParam) IsUseVizJs() bool { return s.useVizJs }

// SequenceDiagramPadder returns the padder of sequence messages.
func (s *SkinParam) SequenceDiagramPadder() *Padder {
	padding := s.asDouble("SequenceMessagePadding")
	margin := s.asDouble("SequenceMessageMargin")
	borderColor, okBorder := s.Value("SequenceMessageBorderColor")
	backgroundColor, okBack := s.Value("SequenceMessageBackGroundColor")
	if padding.IsZero() && margin.IsZero() && !okBorder && !okBack {
		return PadderNone()
	}
	var border, background color.HColor
	if okBorder {
		border = s.IHtmlColorSet().GetColorOrWhite(borderColor)
	}
	if okBack {
		background = s.IHtmlColorSet().GetColorOrWhite(backgroundColor)
	}
	roundCorner := s.RoundCorner(CornerParamDefault, nil)
	return PadderNone().WithMargin(margin).WithPadding(padding).WithBackgroundColor(background).
		WithBorderColor(border).WithRoundCorner(roundCorner)
}

// ActorStyle returns "skinparam actorStyle", stickman by default.
func (s *SkinParam) ActorStyle() ActorStyle {
	value, ok := s.Value("actorstyle")
	if ok && javaEqualsIgnoreCase("awesome", value) {
		return ActorStyleAwesome
	}
	if ok && javaEqualsIgnoreCase("hollow", value) {
		return ActorStyleHollow
	}
	return ActorStyleStickman
}

// SetSvgSize records a text replacement used to measure SVG texts.
func (s *SkinParam) SetSvgSize(origin, sizeToUse string) {
	s.svgCharSizes.put(manageUnicodeNotationUplus(origin), manageUnicodeNotationUplus(sizeToUse))
}

// TransformStringForSizeHack applies the replacements of SetSvgSize, in
// upstream's HashMap order.
func (s *SkinParam) TransformStringForSizeHack(str string) string {
	for _, k := range javaHashMapOrder(s.svgCharSizes.keys) {
		str = strings.ReplaceAll(str, k, s.svgCharSizes.values[k])
	}
	return str
}

// LengthAdjust returns "skinparam lengthAdjust", spacing by default.
func (s *SkinParam) LengthAdjust() LengthAdjust {
	value, ok := s.Value("lengthAdjust")
	switch {
	case ok && javaEqualsIgnoreCase("spacingAndGlyphs", value):
		return LengthAdjustSpacingAndGlyphs
	case ok && javaEqualsIgnoreCase("spacing", value):
		return LengthAdjustSpacing
	case ok && javaEqualsIgnoreCase("none", value):
		return LengthAdjustNone
	}
	return LengthAdjustDefault()
}

// SetParamSameClassWidth records the width shared by classes.
func (s *SkinParam) SetParamSameClassWidth(width float64) { s.paramSameClass = width }

// ParamSameClassWidth returns the width shared by classes.
func (s *SkinParam) ParamSameClassWidth() float64 { return s.paramSameClass }

// Sheet returns the creole parser of a text, with an italic stereotype
// font.
func (s *SkinParam) Sheet(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	creoleMode CreoleMode) SheetBuilder {
	stereotype := fontConfiguration.ForceFont(nil, nil)
	return s.SheetStereo(fontConfiguration, horizontalAlignment, creoleMode, stereotype)
}

// SheetStereo returns the creole parser of a text, cached by its
// parameters.
func (s *SkinParam) SheetStereo(fontConfiguration *FontConfiguration, horizontalAlignment geom.HorizontalAlignment,
	creoleMode CreoleMode, stereo *FontConfiguration) SheetBuilder {
	for _, e := range s.sheetCache {
		if e.alignment == horizontalAlignment && e.mode == creoleMode && e.fc.Equals(fontConfiguration) &&
			sameFontConfiguration(e.stereo, stereo) {
			return e.parser
		}
	}
	result := newCreoleParser(fontConfiguration, horizontalAlignment, s, creoleMode, stereo)
	s.sheetCache = append(s.sheetCache, sheetCacheEntry{alignment: horizontalAlignment, mode: creoleMode,
		fc: fontConfiguration, stereo: stereo, parser: result})
	return result
}

func sameFontConfiguration(a, b *FontConfiguration) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equals(b)
}

// Arrows returns the arrow heads: triangles in strict UML style.
func (s *SkinParam) Arrows() Arrows {
	if s.StrictUmlStyle() {
		return ArrowsTriangle{}
	}
	return ArrowsRegular{}
}

// Pragma returns the pragmas.
func (s *SkinParam) Pragma() *Pragma { return s.pragma }

// Options returns the options.
func (s *SkinParam) Options() *ConfigurationStore[OptionKey] { return s.option }

// PathSystem returns the render context.
func (s *SkinParam) PathSystem() *core.RenderContext { return s.pathSystem }

// FromMd5 returns the text of an MD5 hash.
func (s *SkinParam) FromMd5(md5 string) (string, bool) {
	v, ok := s.md5map[md5]
	return v, ok
}

// trin is upstream's StringUtils.trin.
func trin(s string) string { return text.Trin(s) }
