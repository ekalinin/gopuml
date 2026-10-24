// SPDX-License-Identifier: MIT
// Provenance: original

package svek

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/klimt/svg"
)

// drawRole draws the block of a role and returns the SVG.
func drawRole(t *testing.T, role string, iconSize string) string {
	t.Helper()
	skinParam := klimt.SkinParamCreate(map[string]string{}, nil, core.Class, klimt.PragmaCreateEmpty(),
		klimt.NewConfigurationStore[klimt.OptionKey]())
	if iconSize != "" {
		skinParam.SetParam("classAttributeIconSize", iconSize)
	}
	block := roleLabelsCreate(role, skinParam.Pragma(), klimt.FontConfigurationBlackBlueTrue(font.SansSerif(13)),
		skinParam)
	ug := svg.Build(svg.SvgOptionBasic().WithBackcolor(color.White), false, 42, font.NewStringBounderSvg(nil),
		font.FormatSVG)
	block.DrawU(ug)
	var buf bytes.Buffer
	if err := ug.WriteToStream(&buf, "", 96); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// A role starting with a visibility character gets the icon of a field and
// loses the character, unless the icons are disabled.
func TestRoleLabelsVisibility(t *testing.T) {
	cases := []struct {
		role, iconSize, text, icon string
	}{
		{"-owner", "", ">owner<", `data-visibility-modifier="PRIVATE_FIELD"`},
		{"#owner", "", ">owner<", `data-visibility-modifier="PROTECTED_FIELD"`},
		{"owner", "", ">owner<", ""},
		{"-owner", "0", ">-owner<", ""},
	}
	for _, c := range cases {
		got := drawRole(t, c.role, c.iconSize)
		if !strings.Contains(got, c.text) {
			t.Errorf("%q (icon size %q): no text %s in %s", c.role, c.iconSize, c.text, got)
		}
		hasIcon := strings.Contains(got, "data-visibility-modifier")
		if c.icon == "" && hasIcon || c.icon != "" && !strings.Contains(got, c.icon) {
			t.Errorf("%q (icon size %q): icon %q expected, got %s", c.role, c.iconSize, c.icon, got)
		}
	}
}

// covers reports whether a rectangle at pos holds one of the points.
func covers(pos geom.XPoint2D, dim geom.XDimension2D, points []geom.XPoint2D) bool {
	for _, pt := range points {
		if pt.X > pos.X && pt.X < pos.X+dim.Width() && pt.Y > pos.Y && pt.Y < pos.Y+dim.Height() {
			return true
		}
	}
	return false
}

func intersects(pos geom.XPoint2D, dim geom.XDimension2D, area geom.RectangleArea) bool {
	return pos.X < area.MaxX() && pos.X+dim.Width() > area.MinX() && pos.Y < area.MaxY() &&
		pos.Y+dim.Height() > area.MinY()
}

// The role is drawn on the other side of the line than the quantifier, moved
// off a line that is not vertical at the height of the role.
func TestRoleLabelsPositionOffTheLine(t *testing.T) {
	qDim := geom.NewXDimension2D(10, 14)
	rDim := geom.NewXDimension2D(40, 14)
	quantifier := geom.NewXPoint2D(80, 105)
	end, other := geom.NewXPoint2D(100, 100), geom.NewXPoint2D(130, 160)
	line := roleLabelsSampleSegment(end, other)
	// The mirror image at the end of the line, as 1.2026.8 placed it
	if old := geom.NewXPoint2D(102, 105); !covers(old, rDim, line) {
		t.Fatalf("the old position does not cover the line")
	}
	got := roleLabelsPosition(qDim, rDim, quantifier, end, other, nil, nil)
	if covers(got, rDim, line) {
		t.Errorf("role at %v covers the line", got)
	}
	if got.X < end.X {
		t.Errorf("role at %v is on the side of the quantifier", got)
	}
}

// A role whose mirror image falls in the node the line leaves is moved out
// of it, and out of the quantifier.
func TestRoleLabelsPositionOutOfNode(t *testing.T) {
	qDim := geom.NewXDimension2D(10, 14)
	rDim := geom.NewXDimension2D(40, 14)
	quantifier := geom.NewXPoint2D(95, 52)
	end, other := geom.NewXPoint2D(90, 50), geom.NewXPoint2D(200, 150)
	node := geom.NewRectangleArea(0, 0, 100, 50)
	preferred := roleLabelsMirroredPosition(qDim, rDim, quantifier, end, 110, 100,
		roleLabelsSampleSegment(end, other))
	if !intersects(preferred, rDim, node) {
		t.Fatalf("the mirror image %v is not in the node", preferred)
	}
	got := roleLabelsPosition(qDim, rDim, quantifier, end, other, nil, []geom.RectangleArea{node})
	if intersects(got, rDim, node) {
		t.Errorf("role at %v is in the node", got)
	}
	q := geom.NewRectangleArea(quantifier.X, quantifier.Y, quantifier.X+qDim.Width(), quantifier.Y+qDim.Height())
	if intersects(got, rDim, q) {
		t.Errorf("role at %v is on the quantifier", got)
	}
	if covers(got, rDim, roleLabelsSampleSegment(end, other)) {
		t.Errorf("role at %v covers the line", got)
	}
}
