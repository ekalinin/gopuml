// SPDX-License-Identifier: MIT
// Provenance: original

// Package svek lays out graph-based diagrams (class and object diagrams in
// this release) with a dot layout engine and draws them, as upstream's svek
// package: it measures the image of each entity, writes the DOT graph
// (DotStringFactory, with a subgraph per group and the HTML tables that
// reserve the room of labels and quantifiers), has a layout engine compute
// it, reads node positions, cluster boxes and edge paths back from the SVG
// document the engine returns (SvgResult), and draws the result
// (SvekResult): clusters, nodes, then edges with their arrow heads
// (internal/svek/extremity), labels, quantifiers and qualifier boxes.
//
// Upstream's svek, svek.image and the drawing classes of the dot and
// descdiagram packages refer to each other, so they form this one Go
// package (design D2): files from svek.image are prefixed image_, those of
// dot and descdiagram dot_ and descdiagram_. The extremity package has no
// cycle and stays separate.
//
// The layout runs through dot.Engine (design D7): Layout.Engine, or the
// engine that dot.Config.NewEngine chooses with the render context's
// DotPath and DotTimeout: the dotexec engine running the dot that the
// explicit path or GRAPHVIZ_DOT names, otherwise the built-in engine gogv,
// a Go implementation of dot. A missing named dot gives upstream's error
// image ("Cannot find Graphviz") together with a *dot.NotFoundError; a
// layout that fails or times out gives an image that shows the engine's
// message with its error (design D16); cancelling the context stops the
// layout. The smetana, elk and vizjs layout pragmas are laid out with the
// engine of the render (design D16). In the WebAssembly build, programs
// cannot run and the layout always uses gogv.
//
// Not ported yet, because only diagram types of later changes use them:
// the Kermor layout pragma; it panics with "not ported yet".
package svek
