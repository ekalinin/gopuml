// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/RoundedContainer.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/RoundedNorth.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/svek/RoundedSouth.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// roundedContainer is upstream's RoundedContainer: the rounded box of a
// state with a name part (north), a description part (center) and a body
// part (south), each with its own background.
type roundedContainer struct {
	dim                                             geom.XDimension2D
	nameHeight, descriptionHeight                   float64
	borderColor                                     color.HColor
	northBackcolor, centerBackcolor, southBackcolor color.HColor
	stroke                                          klimt.UStroke
	rounded, shadowing                              float64
}

func newRoundedContainer(borderColor color.HColor, dim geom.XDimension2D, nameHeight, descriptionHeight float64,
	northBackcolor, centerBackcolor, southBackcolor color.HColor, stroke klimt.UStroke,
	rounded, shadowing float64) *roundedContainer {
	if dim.Width() == 0 {
		throwIllegalArgument("")
	}
	return &roundedContainer{dim: dim, nameHeight: nameHeight, descriptionHeight: descriptionHeight,
		borderColor: borderColor, northBackcolor: northBackcolor, centerBackcolor: centerBackcolor,
		southBackcolor: southBackcolor, stroke: stroke, rounded: rounded, shadowing: shadowing}
}

func (r *roundedContainer) DrawU(ug klimt.UGraphic) {
	ug = ug.Apply(r.borderColor).Apply(r.stroke)
	rect := klimt.URectangleBuild(r.dim.Width(), r.dim.Height()).Rounded(r.rounded)
	if r.shadowing > 0 {
		rect.SetDeltaShadow(r.shadowing)
		ug.Apply(klimt.Bg(color.Transparent())).Draw(rect)
		rect.SetDeltaShadow(0)
	}
	newRoundedNorth(r.dim.Width(), r.nameHeight, r.northBackcolor, r.rounded).DrawU(ug)
	r.drawCenter(ug)
	newRoundedSouth(r.dim.Width(), r.dim.Height()-r.nameHeight-r.descriptionHeight, r.southBackcolor, r.rounded).
		DrawU(ug.Apply(klimt.UTranslateDy(r.nameHeight + r.descriptionHeight)))
	ug.Apply(klimt.Bg(color.Transparent())).Draw(rect)
	if r.nameHeight > 0 {
		ug.Apply(klimt.UTranslateDy(r.nameHeight)).Draw(klimt.ULineH(r.dim.Width()))
	}
	if r.descriptionHeight > 0 && r.nameHeight+r.descriptionHeight > 0 {
		ug.Apply(klimt.UTranslateDy(r.nameHeight + r.descriptionHeight)).Draw(klimt.ULineH(r.dim.Width()))
	}
}

func (r *roundedContainer) drawCenter(ug klimt.UGraphic) {
	if r.descriptionHeight == 0 {
		return
	}
	rect := klimt.URectangleBuild(r.dim.Width(), r.descriptionHeight)
	ug = ug.Apply(klimt.UStrokeSimple()).Apply(r.centerBackcolor).Apply(klimt.Bg(r.centerBackcolor))
	ug.Apply(klimt.UTranslateDy(r.nameHeight)).Draw(rect)
}

// roundedNorth is upstream's RoundedNorth: the background of the top of a
// rounded box, with rounded top corners.
type roundedNorth struct {
	width, height, rounded float64
	backColor              color.HColor
}

func newRoundedNorth(width, height float64, backColor color.HColor, rounded float64) *roundedNorth {
	if width == 0 || height == 0 {
		throwIllegalArgument("")
	}
	return &roundedNorth{width: width, height: height, rounded: rounded, backColor: backColor}
}

func (r *roundedNorth) DrawU(ug klimt.UGraphic) {
	if r.backColor.IsTransparent() {
		return
	}
	var header klimt.UShape
	if r.rounded == 0 {
		header = klimt.URectangleBuild(r.width, r.height)
	} else {
		path := klimt.UPathNone()
		path.MoveTo(r.rounded/2, 0)
		path.LineTo(r.width-r.rounded/2, 0)
		path.ArcTo(r.rounded/2, r.rounded/2, 0, 0, 1, r.width, r.rounded/2)
		path.LineTo(r.width, r.height)
		path.LineTo(0, r.height)
		path.LineTo(0, r.rounded/2)
		path.ArcTo(r.rounded/2, r.rounded/2, 0, 0, 1, r.rounded/2, 0)
		path.ClosePath()
		header = path
	}
	ug.Apply(klimt.UStrokeSimple()).Apply(r.backColor).Apply(klimt.Bg(r.backColor)).Draw(header)
}

// roundedSouth is upstream's RoundedSouth: the background of the bottom of
// a rounded box, with rounded bottom corners.
type roundedSouth struct {
	width, height, rounded float64
	backColor              color.HColor
}

func newRoundedSouth(width, height float64, backColor color.HColor, rounded float64) *roundedSouth {
	if width == 0 || height == 0 {
		throwIllegalArgument("")
	}
	return &roundedSouth{width: width, height: height, rounded: rounded, backColor: backColor}
}

func (r *roundedSouth) DrawU(ug klimt.UGraphic) {
	if r.backColor.IsTransparent() {
		return
	}
	var header klimt.UShape
	if r.rounded == 0 {
		header = klimt.URectangleBuild(r.width, r.height)
	} else {
		path := klimt.UPathNone()
		path.MoveTo(0, 0)
		path.LineTo(r.width, 0)
		path.LineTo(r.width, r.height-r.rounded/2)
		path.ArcTo(r.rounded/2, r.rounded/2, 0, 0, 1, r.width-r.rounded/2, r.height)
		path.LineTo(r.rounded/2, r.height)
		path.ArcTo(r.rounded/2, r.rounded/2, 0, 0, 1, 0, r.height-r.rounded/2)
		path.LineTo(0, 0)
		path.ClosePath()
		header = path
	}
	ug.Apply(klimt.UStrokeSimple()).Apply(r.backColor).Apply(klimt.Bg(r.backColor)).Draw(header)
}
