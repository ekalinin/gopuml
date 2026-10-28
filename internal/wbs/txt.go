// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/Fork.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/ITFComposed.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/ITFLeaf.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSDiagram.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/wbs/WBSTextBlock.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package wbs

import (
	"fmt"
	"io"

	"github.com/ekalinin/gopuml/internal/asciiverse"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// exportTxt is upstream's WBSDiagram.exportTxt: the root and its children
// as boxes, drawn in ASCII for TXT and UTXT alike, without the title and
// the other annotations. A panic is returned as an error and nothing is
// written.
func (d *WBSDiagram) exportTxt(w io.Writer) (data export.ImageData, err error) {
	var plan *asciiverse.InfinitePlan
	func() {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = fmt.Errorf("drawing crashed: %w", e)
				} else {
					err = fmt.Errorf("drawing crashed: %v", r)
				}
			}
		}()
		plan = asciiverse.NewInfinitePlan(font.FormatATXT)
		d.drawingElement().asciiDraw(plan)
	}()
	if err != nil {
		return nil, err
	}
	if err := plan.ExportTxt(w); err != nil {
		return nil, err
	}
	return export.ImageDataOK(), nil
}

// asciiDrawLabel draws label in a box with padding spaces on each side.
func asciiDrawLabel(plan *asciiverse.InfinitePlan, label *klimt.Display, padding int) {
	table := asciiverse.NewATable(1, 1).WithLeftPadding(padding).WithRightPadding(padding)
	table.SetCellContent(0, 0, label)
	table.AsciiDraw(plan)
}

// asciiDraw is the box of the label of the node.
func (t *wbsTextBlock) asciiDraw(plan *asciiverse.InfinitePlan) {
	asciiDrawLabel(plan, t.idea.label, 1)
}

// asciiDraw is the box of the root, then the box of each child of the root,
// each one 4 columns right of and 4 lines below the previous one.
func (f *fork) asciiDraw(plan *asciiverse.InfinitePlan) {
	f.wbsTextBlock.asciiDraw(plan)
	for _, child := range f.right {
		plan = plan.Move(4, 4)
		child.asciiDraw(plan)
	}
}

// asciiDraw is the box of the label of the node, without its children.
func (c *itfComposed) asciiDraw(plan *asciiverse.InfinitePlan) { asciiDrawLabel(plan, c.idea.label, 7) }

// asciiDraw is the box of the label of the node.
func (l *itfLeaf) asciiDraw(plan *asciiverse.InfinitePlan) { asciiDrawLabel(plan, l.idea.label, 3) }
