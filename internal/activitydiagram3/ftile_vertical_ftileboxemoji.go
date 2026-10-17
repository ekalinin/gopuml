// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vertical/FtileBoxEmoji.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"strings"
	"unicode/utf16"

	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// ftileBoxEmojiMargin is the space between the emoji and the name.
const ftileBoxEmojiMargin = 5

// FtileBoxEmoji is the action of the <<icon>> stereotype: an emoji (the
// first "<:name:>" of the label, drawn twice as big) with the rest of the
// label beside it, without box.
type FtileBoxEmoji struct {
	AbstractFtile
	emoji       klimt.TextBlock
	name        klimt.TextBlock
	inRendering *LinkRendering
	swimlane    *Swimlane
}

// FtileBoxEmojiCreate returns the icon action of label; styleBuilder may
// be nil.
func FtileBoxEmojiCreate(skinParam klimt.ISkinParam, label *klimt.Display, swimlane *Swimlane,
	stereotype *klimt.Stereotype, styleBuilder *klimt.StyleBuilder) *FtileBoxEmoji {
	if styleBuilder == nil {
		styleBuilder = skinParam.CurrentStyleBuilder()
	}
	style := FtileBoxStyleSignature().WithTOBECHANGED(stereotype).MergedStyle(styleBuilder)
	styleArrow := FtileBoxStyleSignatureArrow().MergedStyle(styleBuilder)
	f := &FtileBoxEmoji{swimlane: swimlane}
	f.initFtile(f, skinParam)
	f.inRendering = LinkRenderingCreate(RainbowFromStyle(styleArrow, f.IHtmlColorSet()))
	fc := style.FontConfiguration(f.IHtmlColorSet(), nil)
	s := label.Get(0).String()
	if klimt.CreoleEmojiMatchingSize(s) == 0 {
		f.emoji = label.Create(fc, geom.HorizontalAlignmentLeft, skinParam)
		f.name = klimt.TextBlockUtilsEmptyTextBlock()
	} else {
		position := klimt.CreoleEmojiMatchingSize(s)
		units := utf16.Encode([]rune(s))
		part1 := string(utf16.Decode(units[:position]))
		part2 := string(utf16.Decode(units[position:]))
		f.emoji = klimt.DisplayCreate(ftileBoxEmojiHackSize(part1, "2")).Create(fc, geom.HorizontalAlignmentLeft,
			skinParam)
		f.name = klimt.DisplayCreate(part2).Create(fc, geom.HorizontalAlignmentLeft, skinParam)
	}
	return f
}

func ftileBoxEmojiHackSize(part1, size string) string {
	if strings.Contains(part1, "*") {
		return part1
	}
	return strings.ReplaceAll(part1, ":>", ":*"+size+">")
}

// InLinkRendering implements Ftile.
func (f *FtileBoxEmoji) InLinkRendering() *LinkRendering { return f.inRendering }

// Swimlanes implements Swimable.
func (f *FtileBoxEmoji) Swimlanes() SwimlaneSet { return singleSwimlane(f.swimlane) }

// SwimlaneIn implements Swimable.
func (f *FtileBoxEmoji) SwimlaneIn() *Swimlane { return f.swimlane }

// SwimlaneOut implements Swimable.
func (f *FtileBoxEmoji) SwimlaneOut() *Swimlane { return f.swimlane }

// DrawU implements klimt.TextBlock.
func (f *FtileBoxEmoji) DrawU(ug klimt.UGraphic) {
	if f.name == klimt.TextBlockUtilsEmptyTextBlock() {
		f.emoji.DrawU(ug)
		return
	}
	dimEmoji := f.emoji.CalculateDimension(ug.StringBounder())
	dimName := f.name.CalculateDimension(ug.StringBounder())
	deltaX := dimEmoji.Width() + ftileBoxEmojiMargin
	deltaY := (dimEmoji.Height() - dimName.Height()) / 2
	if deltaY > 0 {
		f.emoji.DrawU(ug)
		f.name.DrawU(ug.Apply(klimt.NewUTranslate(deltaX, deltaY)))
	} else {
		f.emoji.DrawU(ug.Apply(klimt.UTranslateDy(-deltaY)))
		f.name.DrawU(ug.Apply(klimt.UTranslateDx(deltaX)))
	}
}

func (f *FtileBoxEmoji) calculateDimensionFtile(stringBounder font.StringBounder) FtileGeometry {
	dimEmoji := f.emoji.CalculateDimension(stringBounder)
	dimName := f.name.CalculateDimension(stringBounder)
	width := dimEmoji.Width() + ftileBoxEmojiMargin + dimName.Width()
	height := max(dimEmoji.Height(), dimName.Height())
	deltaY := (dimEmoji.Height() - dimName.Height()) / 2
	if deltaY > 0 {
		return NewFtileGeometryOut(width, height, dimEmoji.Width()/2, 0, dimEmoji.Height())
	}
	return NewFtileGeometryOut(width, height, dimEmoji.Width()/2, -deltaY, dimEmoji.Height()-deltaY)
}

// MyChildren implements Ftile: none.
func (f *FtileBoxEmoji) MyChildren() []Ftile { return nil }
