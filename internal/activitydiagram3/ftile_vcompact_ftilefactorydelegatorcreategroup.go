// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorCreateGroup.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
)

// FtileFactoryDelegatorCreateGroup draws partitions and other blocks
// around their tiles.
type FtileFactoryDelegatorCreateGroup struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorCreateGroup returns the delegator over factory.
func NewFtileFactoryDelegatorCreateGroup(factory FtileFactory) *FtileFactoryDelegatorCreateGroup {
	return &FtileFactoryDelegatorCreateGroup{newFtileFactoryDelegator(factory)}
}

// CreateGroup implements FtileFactory.
func (d *FtileFactoryDelegatorCreateGroup) CreateGroup(list Ftile, name *klimt.Display, backColor color.HColor,
	note *PositionedNote, typ GroupSymbol, style *klimt.Style) Ftile {
	var result Ftile = NewFtileGroup(list, name, backColor, d.SkinParam(), typ, style)
	if note != nil {
		result = NewFtileWithNotes(result, []*PositionedNote{note}, geom.VerticalAlignmentCenter)
	}
	return result
}
