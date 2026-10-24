// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/activitydiagram3/ftile/vcompact/FtileFactoryDelegatorCreateParallel.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package activitydiagram3

import (
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/uml"
)

// FtileFactoryDelegatorCreateParallel draws the parallel blocks: fork,
// merge and split.
type FtileFactoryDelegatorCreateParallel struct {
	FtileFactoryDelegator
}

// NewFtileFactoryDelegatorCreateParallel returns the delegator over
// factory.
func NewFtileFactoryDelegatorCreateParallel(factory FtileFactory) *FtileFactoryDelegatorCreateParallel {
	return &FtileFactoryDelegatorCreateParallel{newFtileFactoryDelegator(factory)}
}

// CreateParallel implements FtileFactory.
func (d *FtileFactoryDelegatorCreateParallel) CreateParallel(all []Ftile, style ForkStyle, label string,
	hasLabel bool, in, out *Swimlane, colors *color.Colors) Ftile {
	var builder parallelFtilesBuilder
	switch style {
	case ForkStyleSplit:
		builder = NewParallelBuilderSplit(d.SkinParam(), d.StringBounder(), all, colors)
	case ForkStyleMerge:
		builder = NewParallelBuilderMerge(d.SkinParam(), d.StringBounder(), all, colors)
	case ForkStyleFork:
		builder = NewParallelBuilderFork(d.SkinParam(), d.StringBounder(), label, hasLabel, in, out, all, colors)
	default:
		uml.Throw(uml.IllegalStateException, "")
	}
	inner := d.FtileFactoryDelegator.CreateParallel(builder.List99(), style, label, hasLabel, in, out, colors)
	return builder.Build(inner)
}
