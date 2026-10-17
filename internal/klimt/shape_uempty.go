// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UEmpty.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UComment.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UCenteredCharacter.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UHidden.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/shape/UPixel.java
// Provenance: plantuml-mit 1.2026.8 net/atmp/SpecialText.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/CenteredText.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// UEmpty is an invisible shape that only counts in the bounds.
type UEmpty struct {
	width  float64
	height float64
}

// NewUEmpty returns an invisible shape of a size. A zero width panics, as
// upstream throws IllegalArgumentException.
func NewUEmpty(width, height float64) *UEmpty {
	if width == 0 {
		throw(illegalArgument, "")
	}
	return &UEmpty{width: width, height: height}
}

// UEmptyCreate returns the invisible shape of a dimension.
func UEmptyCreate(d geom.XDimension2D) *UEmpty { return NewUEmpty(d.Width(), d.Height()) }

// Width returns the width.
func (e *UEmpty) Width() float64 { return e.width }

// Height returns the height.
func (e *UEmpty) Height() float64 { return e.height }

// UComment is a comment written in the output.
type UComment struct {
	comment string
}

// NewUComment returns a comment.
func NewUComment(comment string) *UComment { return &UComment{comment: comment} }

// Comment returns the text of the comment.
func (c *UComment) Comment() string { return c.comment }

// UCenteredCharacter is a character drawn centered on the current position.
type UCenteredCharacter struct {
	c    rune
	font font.UFont
}

// NewUCenteredCharacter returns a centered character.
func NewUCenteredCharacter(c rune, f font.UFont) *UCenteredCharacter {
	return &UCenteredCharacter{c: c, font: f}
}

// Char returns the character.
func (c *UCenteredCharacter) Char() rune { return c.c }

// Font returns the font.
func (c *UCenteredCharacter) Font() font.UFont { return c.font }

// UHidden is the change that hides or shows what is drawn next.
type UHidden int

// Visibility changes, in upstream's order.
const (
	UHiddenNormalShown UHidden = iota
	UHiddenHidden
)

// String returns upstream's constant name.
func (h UHidden) String() string {
	return enumName("UHidden", []string{"NORMAL_SHOWN", "HIDDEN"}, int(h))
}

// UPixel is a single point.
type UPixel struct{}

// SpecialText is a title drawn by the graphic itself (upstream's
// net.atmp.SpecialText, used by BigFrame). It is ignored by compression.
type SpecialText struct {
	title TextBlock
}

var _ UShapeIgnorableForCompression = (*SpecialText)(nil)

// NewSpecialText returns the special text of a title.
func NewSpecialText(title TextBlock) *SpecialText { return &SpecialText{title: title} }

// IsIgnoreForCompressionOn returns true.
func (s *SpecialText) IsIgnoreForCompressionOn(CompressionMode) bool { return true }

// DrawWhenCompressed draws a 1x1 empty shape after the title.
func (s *SpecialText) DrawWhenCompressed(ug UGraphic, mode CompressionMode) {
	d := s.title.CalculateDimension(ug.StringBounder())
	ug.Apply(UTranslateDx(d.Width())).Draw(NewUEmpty(1, 1))
}

// Title returns the title.
func (s *SpecialText) Title() TextBlock { return s.title }

// CenteredText is a text centered in a width (upstream's
// activitydiagram3.ftile.CenteredText); the bounds computations ignore it.
type CenteredText struct {
	text       TextBlock
	totalWidth float64
}

// NewCenteredText returns a centered text.
func NewCenteredText(text TextBlock, totalWidth float64) *CenteredText {
	return &CenteredText{text: text, totalWidth: totalWidth}
}

// Text returns the text.
func (c *CenteredText) Text() TextBlock { return c.text }

// TotalWidth returns the width the text is centered in.
func (c *CenteredText) TotalWidth() float64 { return c.totalWidth }
