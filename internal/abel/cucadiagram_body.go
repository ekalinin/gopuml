// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/BodierLikeClassOrObject.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/BodierMap.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/BodierJSon.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/BodierSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/LinkConstraint.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
	"github.com/ekalinin/plantuml-in-go/internal/uml"
)

// The drawing parts of the bodiers (upstream's getBody) and of the link
// constraints.

// isBodyEnhanced reports whether a line needs the enhanced body: a
// separator, a creole table or tree, or an embedded diagram.
func (b *BodierLikeClassOrObject) isBodyEnhanced() bool {
	for _, cs := range b.rawBody {
		s := cs.String()
		if _, embedded := klimt.EmbeddedDiagramType(s); klimt.BodyEnhancedIsBlockSeparator(s) ||
			klimt.CreoleParserIsTableLine(s) || klimt.ParserIsTreeStart(s) || embedded {
			return true
		}
	}
	return false
}

// Body implements Bodier: the enhanced body when a line needs it, the
// fields of an object, else the compartments of the fields and of the
// methods that are shown.
func (b *BodierLikeClassOrObject) Body(skinParam klimt.ISkinParam, showMethods, showFields bool,
	stereotype *klimt.Stereotype, style *klimt.Style, _ *klimt.FontConfiguration) klimt.TextBlock {
	if b.typ.IsLikeClass() && b.isBodyEnhanced() {
		if showMethods || showFields {
			return BodyFactoryCreate1(skinParam.DefaultTextAlignment(geom.HorizontalAlignmentLeft),
				b.RawBodyWithoutHidden(), skinParam, stereotype, b.leaf, style)
		}
		return nil
	}
	if b.leaf == nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	if b.typ == LeafTypeObject {
		if !showFields {
			return klimt.TextBlockUtilsEmpty(0, 0)
		}
		return BodyFactoryCreate1(skinParam.DefaultTextAlignment(geom.HorizontalAlignmentLeft),
			b.RawBodyWithoutHidden(), skinParam, stereotype, b.leaf, style)
	}
	fields := NewMethodsOrFieldsArea(b.FieldsToDisplay(), skinParam, b.leaf, style)
	methods := NewMethodsOrFieldsArea(b.MethodsToDisplay(), skinParam, b.leaf, style)
	switch {
	case showFields && !showMethods:
		return fields.AsBlockMemberImpl()
	case showMethods && !showFields:
		return methods.AsBlockMemberImpl()
	case !showFields && !showMethods:
		return klimt.TextBlockUtilsEmpty(0, 0)
	}
	bb1 := fields.AsBlockMemberImpl()
	bb2 := methods.AsBlockMemberImpl()
	return klimt.TextBlockUtilsMergeTB(bb1, bb2, geom.HorizontalAlignmentLeft)
}

// Body implements Bodier: the table of the entries, drawn with
// fontConfiguration.
func (b *BodierMap) Body(skinParam klimt.ISkinParam, _, _ bool, _ *klimt.Stereotype, style *klimt.Style,
	fontConfiguration *klimt.FontConfiguration) klimt.TextBlock {
	wordWrap := style.WrapWidth()
	horizontalAlignment := style.HorizontalAlignment()
	return NewTextBlockMap(fontConfiguration, skinParam, b.keys, b.values, wordWrap, horizontalAlignment)
}

// Body implements Bodier: the data, drawn with fontConfiguration.
func (b *BodierJSon) Body(skinParam klimt.ISkinParam, _, _ bool, _ *klimt.Stereotype, style *klimt.Style,
	fontConfiguration *klimt.FontConfiguration) klimt.TextBlock {
	return NewTextBlockCucaJSon(fontConfiguration, skinParam, b.json, style.WrapWidth())
}

// Body implements Bodier: the enhanced body of the lines.
func (b *BodierSimple) Body(skinParam klimt.ISkinParam, _, _ bool, stereotype *klimt.Stereotype,
	style *klimt.Style, _ *klimt.FontConfiguration) klimt.TextBlock {
	return BodyFactoryCreate1(skinParam.DefaultTextAlignment(geom.HorizontalAlignmentLeft), b.rawBody, skinParam,
		stereotype, b.leaf, style)
}

// DrawMe draws the constraint: a dashed line between the two links and its
// label in the middle, once both positions are known.
func (c *LinkConstraint) DrawMe(ug klimt.UGraphic, skinParam klimt.ISkinParam) {
	if c.x1 == 0 && c.y1 == 0 {
		return
	}
	if c.x2 == 0 && c.y2 == 0 {
		return
	}
	ug = ug.Apply(color.Black)
	line := klimt.NewULine(c.x2-c.x1, c.y2-c.y1)
	ug = ug.Apply(klimt.NewUStroke(3, 3, 1))
	ug.Apply(klimt.NewUTranslate(c.x1, c.y1)).Draw(line)

	label := c.display.Create(klimt.FontConfigurationFromSkinParam(skinParam, klimt.FontParamArrow, nil),
		geom.HorizontalAlignmentCenter, skinParam)
	dimLabel := label.CalculateDimension(ug.StringBounder())
	x := (c.x1+c.x2)/2 - dimLabel.Width()/2
	y := (c.y1+c.y2)/2 - dimLabel.Height()/2
	label.DrawU(ug.Apply(klimt.NewUTranslate(x, y)))
}
