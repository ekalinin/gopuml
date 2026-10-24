// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/image/EntityImageNoteLink.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/rose"
)

// EntityImageNoteLink is the note of a link, drawn in its label.
type EntityImageNoteLink struct {
	klimt.TextBlockDefaults
	memo klimt.TextBlockMemoized
	comp klimt.Component
}

var _ klimt.IEntityImage = (*EntityImageNoteLink)(nil)

func newEntityImageNoteLink(note *klimt.Display, colors *color.Colors, skinParam klimt.ISkinParam,
	styleBuilder *klimt.StyleBuilder) *EntityImageNoteLink {
	skin := rose.NewRose()
	comp := skin.CreateComponentNote(
		[]*klimt.Style{klimt.ComponentTypeNote.StyleSignature().MergedStyle(styleBuilder)}, klimt.ComponentTypeNote,
		skinParam, note, colors)
	return &EntityImageNoteLink{comp: comp}
}

// CalculateDimension returns the preferred size of the note.
func (n *EntityImageNoteLink) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return n.memo.Dimension(stringBounder, func(sb font.StringBounder) geom.XDimension2D {
		height := n.comp.PreferredHeight(sb)
		width := n.comp.PreferredWidth(sb)
		return geom.NewXDimension2D(width, height)
	})
}

// DrawU draws the note.
func (n *EntityImageNoteLink) DrawU(ug klimt.UGraphic) {
	n.comp.DrawU(ug, klimt.NewArea(n.CalculateDimension(ug.StringBounder())), klimt.NewSimpleContext2D(false))
}

// ShapeType returns RECTANGLE.
func (n *EntityImageNoteLink) ShapeType() klimt.ShapeType { return klimt.ShapeTypeRectangle }

// Shield returns no margin.
func (n *EntityImageNoteLink) Shield(font.StringBounder) klimt.Margins { return klimt.MarginsNone() }

// IsHidden returns false.
func (n *EntityImageNoteLink) IsHidden() bool { return false }

// OverscanX returns 0.
func (n *EntityImageNoteLink) OverscanX(font.StringBounder) float64 { return 0 }

// IsCrash returns false.
func (n *EntityImageNoteLink) IsCrash() bool { return false }
