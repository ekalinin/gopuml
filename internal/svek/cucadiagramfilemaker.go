// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/CucaDiagramFileMaker.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/CucaDiagramFileMakerSvek.java
// Provenance: plantuml-mit 1.2026.8 net/atmp/CucaDiagram.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/ekalinin/gopuml/internal/abel"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/uml"
)

// Layout configures the graph layout of a diagram.
type Layout struct {
	// Engine lays the DOT graph out. Nil means the Graphviz dot executable
	// found with Config (the dotexec engine).
	Engine dot.Engine
	// Config finds dot when Engine is nil.
	Config dot.Config
	// Trace, when set, receives the DOT graph and the SVG document of the
	// engine (upstream's -debugsvek trace files).
	Trace func(dotString, svg string)
}

// VersionedEngine is an engine that tells which Graphviz version it
// behaves as; the DOT graph has a few workarounds for old versions. The
// dotexec engine (*dot.Exec) implements it by running `dot -V`.
type VersionedEngine interface {
	dot.Engine
	GraphvizVersion(ctx context.Context) dot.GraphvizVersion
}

// modernGraphvizVersion is the version assumed for an engine that does not
// tell its own: the Graphviz release of the reference renders.
var modernGraphvizVersion = dot.NewGraphvizVersion("dot - graphviz version 14.0.1 (20251006.0113)")

// LayoutFromContext returns the layout of a render: dot found with the
// explicit path and the timeout of the render context.
func LayoutFromContext(rc *core.RenderContext) Layout {
	if rc == nil {
		return Layout{}
	}
	return Layout{Config: dot.Config{DotPath: rc.DotPath, Timeout: rc.DotTimeout}}
}

// versionCache holds the Graphviz version of each dot executable, as
// upstream's GraphvizRuntimeEnvironment caches it per file (running
// `dot -V` for every diagram would double the cost of a layout). The key
// includes the size and modification time of the file, so that replacing
// the executable is noticed.
var versionCache sync.Map

func cachedVersion(ctx context.Context, e *dot.Exec) dot.GraphvizVersion {
	key := e.Path
	if fi, err := os.Stat(e.Path); err == nil {
		key += "\x00" + fi.ModTime().String() + "\x00" + itoa64(fi.Size())
	}
	if v, ok := versionCache.Load(key); ok {
		return v.(dot.GraphvizVersion)
	}
	v := e.GraphvizVersion(ctx)
	if ctx.Err() == nil {
		versionCache.Store(key, v)
	}
	return v
}

// resolve returns the engine of the layout and the Graphviz version the
// DOT graph is written for. When dot cannot be found, it returns the error
// that the diagram shows (with upstream's default version, as upstream
// cannot run `dot -V` either).
func (l Layout) resolve(ctx context.Context) (le *layoutEngine, version dot.GraphvizVersion) {
	le = &layoutEngine{ctx: ctx, trace: l.Trace}
	if l.Engine != nil {
		le.engine = l.Engine
		if v, ok := l.Engine.(VersionedEngine); ok {
			return le, v.GraphvizVersion(ctx)
		}
		return le, modernGraphvizVersion
	}
	if !execSupported {
		le.notFound = &dot.NotFoundError{State: dot.NullUndefined}
		return le, dot.DefaultGraphvizVersion
	}
	exec, err := l.Config.NewExec()
	if err != nil {
		var nf *dot.NotFoundError
		if errors.As(err, &nf) {
			le.notFound = nf
		} else {
			le.notFound = &dot.NotFoundError{State: dot.NullUndefined}
		}
		return le, dot.DefaultGraphvizVersion
	}
	le.engine = exec
	return le, cachedVersion(ctx, exec)
}

// DotStringSkek is CucaDiagram.getDotStringSkek: the settings of the DOT
// graph that svek keeps (nodesep, ranksep, layout), and the aspect and ratio
// pragmas.
func DotStringSkek(d *abel.CucaDiagram, dotStrings []string) []string {
	var result []string
	for _, s := range dotStrings {
		if strings.HasPrefix(s, "nodesep") || strings.HasPrefix(s, "ranksep") || strings.HasPrefix(s, "layout") {
			result = append(result, s)
		}
	}
	if aspect, ok := d.Pragma().Value(klimt.PragmaAspect); ok {
		result = append(result, "aspect="+strings.ReplaceAll(aspect, ",", ".")+";")
	}
	if ratio, ok := d.Pragma().Value(klimt.PragmaRatio); ok {
		result = append(result, "ratio="+ratio+";")
	}
	return result
}

// TextBlock is CucaDiagram.getTextBlock with CucaDiagramFileMakerSvek: the
// drawing of the graph of a diagram laid out by the engine, without the
// annotations (title, legend...). dotStrings are the DOT settings of the
// diagram (AbstractEntityDiagram.getDotStrings).
//
// Upstream uses its Smetana port when "!pragma layout smetana" is set or
// dot is missing, and ELK with "!pragma layout elk"; these engines are not
// ported, so dot is always used (design D16) and a missing dot gives
// upstream's error image. The returned error is a layout error (a
// *dot.NotFoundError, *dot.TimeoutError or *dot.FailureError) that comes
// with a block showing it, or the error of the context, with a nil block.
func TextBlock(ctx context.Context, d *abel.CucaDiagram, dotStrings []string, fileFormatOption export.FileFormatOption,
	layout Layout) (klimt.TextBlock, error) {
	d.EventuallyBuildPhantomGroups()
	stringBounder, err := fileFormatOption.DefaultStringBounder(d.Skin(), d.Pragma())
	if err != nil {
		return nil, err
	}
	// The engine is resolved first, so that the composite states laid out
	// on their own use the same dot, timeout and context. Without dot, they
	// are not laid out: the main run gives the not-found error image.
	le, version := layout.resolve(ctx)
	switch d.DiagramType() {
	case core.Activity:
		notPorted("CucaDiagramSimplifierActivity")
	case core.State:
		if le.notFound == nil {
			if image, err := simplifyState(d, stringBounder, DotModeNormal, le, version); image != nil || err != nil {
				return image, err
			}
		}
	}
	for _, leaf := range d.Leafs() {
		leaf.ClearKals()
	}

	// CucaDiagramFileMaker
	bibliotekon := newBibliotekon(d.Links())
	rootCluster := newRootCluster(d, bibliotekon.colorSequence, d.RootGroup())
	cm := &clusterManager{bibliotekon: bibliotekon, current: rootCluster}

	dotStringFactory := newDotStringFactory(bibliotekon, cm.current, d.DiagramType(), d.SkinParam(), version)
	data := newDotData(d, d.RootGroup(), orderedLinks(d.Links()), d.Leafs(), d, d)
	builder := &graphvizImageBuilder{dotData: data, pragma: d.Pragma(),
		styleName: uml.DiagramTypeStyleName(d.DiagramType()), dotMode: DotModeNormal, dotStringFactory: dotStringFactory,
		clusterManager: cm}
	image, err := builder.buildImage(stringBounder, DotStringSkek(d, dotStrings), le)
	if image == nil {
		return nil, err
	}
	return image, err
}

// orderedLinks puts the links with the same ends next to each other, in
// the order of the first of them.
func orderedLinks(links []*abel.Link) []*abel.Link {
	var result []*abel.Link
	for _, l := range links {
		result = addLinkNew(result, l)
	}
	return result
}

func addLinkNew(result []*abel.Link, link *abel.Link) []*abel.Link {
	for i := 0; i < len(result); i++ {
		if result[i].SameConnections(link) {
			for i < len(result) && result[i].SameConnections(link) {
				i++
			}
			if i == len(result) {
				return append(result, link)
			}
			return insertAt(result, i, link)
		}
	}
	return append(result, link)
}
