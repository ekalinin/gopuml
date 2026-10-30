// SPDX-License-Identifier: MIT
// Provenance: original

package svek

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/svg"
	"github.com/ekalinin/gopuml/internal/text"
)

// chenLeaf returns a leaf of a Chen diagram, created as the Chen commands
// create it. The Chen parser imports this package, so the test cannot use
// it.
func chenLeaf(t *testing.T, typ abel.LeafType, name, stereotype string) *abel.Entity {
	t.Helper()
	d := &abel.AbstractEntityDiagram{}
	lines := []text.StringLocated{{S: "@startchen"}, {S: "@endchen"}}
	d.InitCucaDiagram(core.NewRenderContext(), core.NewUmlSource(lines, lines, true), core.ChenEER, nil, nil)
	e := d.ReallyCreateLeaf(nil, d.QuarkInContext(true, name), klimt.DisplayWithNewlines(d.Pragma(), name), typ,
		abel.USymbolNull)
	if stereotype != "" {
		e.SetStereotype(klimt.StereotypeBuild(stereotype))
	}
	e.SetColors(color.EmptyColors())
	return e
}

// TestChenImageDimensions checks the sizes of the Chen images against the
// shapes of the reference chen-eer-diagram/entities-relationship-and-attributes.
func TestChenImageDimensions(t *testing.T) {
	stringBounder := font.NewStringBounderSvg(nil)
	for _, c := range []struct {
		image         klimt.IEntityImage
		shape         klimt.ShapeType
		width, height float64
	}{
		{newEntityImageChenEntity(chenLeaf(t, abel.LeafTypeChenEntity, "CUSTOMER", "")),
			klimt.ShapeTypeRectangle, 99.133, 36.297},
		{newEntityImageChenRelationship(chenLeaf(t, abel.LeafTypeChenRelationship, "RENTED_TO", "")),
			klimt.ShapeTypeDiamond, 138.749, 69.375},
		{newEntityImageChenAttribute(chenLeaf(t, abel.LeafTypeChenAttribute, "Number", "")),
			klimt.ShapeTypeOval, 85.534, 29.048},
		{newEntityImageChenCircle(chenLeaf(t, abel.LeafTypeChenCircle, "d", "")),
			klimt.ShapeTypeOval, 25, 25},
	} {
		dim := c.image.CalculateDimension(stringBounder)
		if math.Abs(dim.Width()-c.width) > 0.001 || math.Abs(dim.Height()-c.height) > 0.001 {
			t.Errorf("%T: %v x %v, want %v x %v", c.image, dim.Width(), dim.Height(), c.width, c.height)
		}
		if got := c.image.ShapeType(); got != c.shape {
			t.Errorf("%T: shape %v, want %v", c.image, got, c.shape)
		}
	}
	// The attribute takes the size of the smallest ellipse around its name.
	attribute := newEntityImageChenAttribute(chenLeaf(t, abel.LeafTypeChenAttribute, "Number", ""))
	want := klimt.NewTextBlockInContainingEllipse(attribute.title, stringBounder).CalculateDimension(stringBounder)
	if got := attribute.CalculateDimension(stringBounder); got != want {
		t.Errorf("attribute %v, ellipse %v", got, want)
	}
}

// drawChen draws an image alone and returns its SVG.
func drawChen(t *testing.T, image klimt.IEntityImage) string {
	t.Helper()
	ug := svg.Build(svg.SvgOptionBasic().WithBackcolor(color.White), false, 42, font.NewStringBounderSvg(nil),
		font.FormatSVG)
	image.DrawU(ug)
	var buf bytes.Buffer
	if err := ug.WriteToStream(&buf, "", 96); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// TestChenImageDrawing checks the group of each image, which has only an id
// and so is drawn as a bare <g>, and the shapes of the stereotypes.
func TestChenImageDrawing(t *testing.T) {
	for _, c := range []struct {
		image       klimt.IEntityImage
		shape       string
		count       int
		mustContain string
	}{
		{newEntityImageChenEntity(chenLeaf(t, abel.LeafTypeChenEntity, "A", "")), "<rect", 1, ""},
		{newEntityImageChenEntity(chenLeaf(t, abel.LeafTypeChenEntity, "A", "<<weak>>")), "<rect", 2, ""},
		{newEntityImageChenRelationship(chenLeaf(t, abel.LeafTypeChenRelationship, "R", "")), "<polygon", 1, ""},
		{newEntityImageChenRelationship(chenLeaf(t, abel.LeafTypeChenRelationship, "R", "<<identifying>>")),
			"<polygon", 2, ""},
		{newEntityImageChenAttribute(chenLeaf(t, abel.LeafTypeChenAttribute, "x", "<<key>>")), "<ellipse", 1,
			`text-decoration="underline"`},
		{newEntityImageChenAttribute(chenLeaf(t, abel.LeafTypeChenAttribute, "x", "<<multi>>")), "<ellipse", 2, ""},
		{newEntityImageChenAttribute(chenLeaf(t, abel.LeafTypeChenAttribute, "x", "<<derived>>")), "<ellipse", 1,
			"stroke-dasharray:10,10"},
		{newEntityImageChenCircle(chenLeaf(t, abel.LeafTypeChenCircle, "U", "")), "<ellipse", 1, ">U<"},
	} {
		got := drawChen(t, c.image)
		if !strings.Contains(got, "<g><"+strings.TrimPrefix(c.shape, "<")) {
			t.Errorf("%T: no bare group around the shape in %s", c.image, got)
		}
		if n := strings.Count(got, c.shape+" "); n != c.count {
			t.Errorf("%T: %d %s, want %d", c.image, n, c.shape, c.count)
		}
		if c.mustContain != "" && !strings.Contains(got, c.mustContain) {
			t.Errorf("%T: no %s in %s", c.image, c.mustContain, got)
		}
	}
}
