// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/JsonDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"context"
	"io"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// JsonDiagram is a JSON, YAML or HCL diagram: the boxes of its data, or the
// text that says the data cannot be read.
type JsonDiagram struct {
	uml.Base
	klimt.TextBlockDefaults

	// root is nil when the data cannot be read.
	root        json.Value
	highlighted []*Highlighted
	handwritten bool
	// smetana holds the layout of the boxes, made by ExportDiagramContext.
	smetana *smetanaForJSON
	// engine lays the boxes out; nil means gogv with the timeout of the
	// render.
	engine layoutEngine
}

var (
	_ core.TypedDiagram = (*JsonDiagram)(nil)
	_ klimt.TextBlock   = (*JsonDiagram)(nil)
)

// NewJsonDiagram returns the diagram of the data root, of type JSON, YAML or
// HCL. A root that is a string, a number, a boolean or null becomes an array
// of one item, and an empty object or array an array of one empty string,
// as upstream. The scale line of styleExtractor is applied.
func NewJsonDiagram(ctx *core.RenderContext, src *core.UmlSource, diagramType core.DiagramType, root json.Value,
	highlighted []*Highlighted, styleExtractor *StyleExtractor) *JsonDiagram {
	d := &JsonDiagram{Base: uml.NewBase(ctx, src, diagramType, nil), highlighted: highlighted,
		handwritten: styleExtractor.IsHandwritten()}
	switch {
	case root != nil && (root.IsString() || root.IsBoolean() || root.IsNumber() || root.IsNull()):
		d.root = json.NewArray().Add(root)
	case root != nil && ((root.IsArray() && root.AsArray().IsEmpty()) || (root.IsObject() && root.AsObject().IsEmpty())):
		d.root = json.NewArray().Add(json.ValueOfString(""))
	default:
		d.root = root
	}
	if scale := styleExtractor.Scale(); scale != "" {
		lines := command.SingleString(scale)
		for _, cmd := range uml.AddCommonScaleCommands([]command.Command[*JsonDiagram](nil)) {
			if cmd.IsValid(lines) == command.OK {
				cmd.Execute(d, lines, command.One)
			}
		}
	}
	return d
}

// Description implements core.Diagram: "(Json)", "(Yaml)" or "(HCL)".
func (d *JsonDiagram) Description() string {
	switch d.DiagramType() {
	case core.YAML:
		return "(Yaml)"
	case core.HCL:
		return "(HCL)"
	}
	return "(Json)"
}

// DrawU draws the boxes, or the text that says the data cannot be read. The
// handwritten drawing is not ported, so it draws the boxes as they are.
func (d *JsonDiagram) DrawU(ug klimt.UGraphic) {
	if d.root == nil {
		display := klimt.DisplayWithNewlines(d.Pragma(),
			"Your data does not sound like "+d.DiagramType().String()+" data")
		fontConfiguration := klimt.FontConfigurationBlackBlueTrue(font.Monospace(14))
		result := display.Create(fontConfiguration, geom.HorizontalAlignmentLeft, d.SkinParam())
		result = klimt.TextBlockUtilsWithMargin(result, 5, 2)
		result.DrawU(ug)
		return
	}
	d.smetana.drawMe(ug)
}

// CalculateDimension is the extent of the drawing.
func (d *JsonDiagram) CalculateDimension(stringBounder font.StringBounder) geom.XDimension2D {
	return klimt.TextBlockUtilsMinMax(d, stringBounder, true).Dimension()
}

// ExportDiagram draws the diagram, as ExportDiagramContext without
// cancellation.
func (d *JsonDiagram) ExportDiagram(w io.Writer, index int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	return d.ExportDiagramContext(context.Background(), w, index, fileFormatOption)
}

// ExportDiagramContext lays the boxes out with gogv, then draws them with
// the annotations around them, as upstream's UgDiagram.exportDiagram. A
// layout that fails or times out returns its error, and a cancelled
// context its own error, without output.
func (d *JsonDiagram) ExportDiagramContext(ctx context.Context, w io.Writer, _ int,
	fileFormatOption export.FileFormatOption) (export.ImageData, error) {
	if d.root != nil {
		stringBounder, err := fileFormatOption.DefaultStringBounder(d.Skin(), d.Pragma())
		if err != nil {
			return nil, err
		}
		engine := d.engine
		if engine == nil {
			engine = &dot.Gogv{Timeout: d.RenderContext().DotTimeout}
		}
		smetana := newSmetanaForJSON(d.SkinParam(), diagramSName(d.DiagramType()), stringBounder)
		if err := smetana.layout(ctx, engine, d.root, d.highlighted); err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return nil, err
		}
		d.smetana = smetana
	}
	return uml.ExportTitled(w, d, d.Title(), func() klimt.TextBlock { return d }, fileFormatOption)
}
