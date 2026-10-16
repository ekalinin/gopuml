// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/classdiagram/ClassDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package classdiagram

import (
	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/command"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// ClassDiagram is a parsed class diagram (object diagrams are parsed as
// class diagrams too).
type ClassDiagram struct {
	AbstractClassOrObjectDiagram

	allowMixing       bool
	useLayoutExplicit int
}

var (
	_ core.TypedDiagram = (*ClassDiagram)(nil)
	_ abel.Diagram      = (*ClassDiagram)(nil)
)

// NewClassDiagram returns an empty class diagram for src; md5map holds the
// base64 images of the source.
func NewClassDiagram(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *ClassDiagram {
	d := &ClassDiagram{}
	d.initAbstractClassOrObjectDiagram(ctx, src, core.Class, md5map)
	return d
}

// SetAllowMixing records "allowmixing": description elements may be used.
func (d *ClassDiagram) SetAllowMixing(allowMixing bool) { d.allowMixing = allowMixing }

// IsAllowMixing reports whether "allowmixing" was used.
func (d *ClassDiagram) IsAllowMixing() bool { return d.allowMixing }

// LayoutNewLine is "layout_new_line": the next elements start a new row.
func (d *ClassDiagram) LayoutNewLine() {
	d.useLayoutExplicit++
	d.IncRawLayout()
}

// UseLayoutExplicit returns the number of "layout_new_line" commands.
func (d *ClassDiagram) UseLayoutExplicit() int { return d.useLayoutExplicit }

// CheckFinalError implements command.Diagram: once the source is read, the
// links between the ends of a link of length 1 get length 1, the
// intermediate packages are packed when "!pragma useIntermediatePackages
// false" is set, and the standalone entities are put in squares.
func (d *ClassDiagram) CheckFinalError() string {
	links := d.Links()
	for _, link := range links {
		if link.Length() == 1 {
			for _, link2 := range links {
				if link2.SameConnections(link) && link2.Length() != 1 {
					link2.SetLength(1)
				}
			}
		}
	}
	if d.Pragma().IsFalse(klimt.PragmaUseIntermediatePackages) {
		d.PackSomePackage()
	}
	d.ApplySingleStrategy()
	return ""
}

// CheckIfPackageHierarchyIsOk returns an error when a parent of the entity
// is a leaf.
func (d *ClassDiagram) CheckIfPackageHierarchyIsOk(entity *abel.Entity) command.Result {
	current := entity.Quark().Parent()
	for !current.IsRoot() {
		if current.Data() != nil && !current.Data().IsGroup() {
			return command.Error("Bad hierarchy for class " + entity.Quark().QualifiedName())
		}
		current = current.Parent()
	}
	return command.Ok()
}

// Export draws a page; the graph layout and the drawing of class diagrams
// are not ported yet.
func (d *ClassDiagram) Export(int) error { return abel.ErrNotDrawable }
