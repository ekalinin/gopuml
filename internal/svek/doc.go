// SPDX-License-Identifier: MIT
// Provenance: original

// Package svek lays out graph-based diagrams (class and object diagrams in
// this release) with Graphviz dot and draws them, as upstream's svek
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
// dotexec engine running the dot executable found with the explicit path,
// GRAPHVIZ_DOT, PATH and the standard locations, with a timeout (the render
// context's DotPath and DotTimeout). The engine is the only part that knows
// about the dot program, so a pure Go engine can replace it. A missing dot
// gives upstream's error image ("Cannot find Graphviz") together with a
// *dot.NotFoundError; a dot run that fails or times out gives an image that
// shows dot's message with its error (design D16); cancelling the context
// stops dot. The smetana, elk and vizjs layout pragmas are laid out with
// dot (design D16). In the WebAssembly build, programs cannot run and the
// layout always reports dot as not found.
//
// Not ported yet, because only diagram types of later changes use them:
// the images of activities and Chen diagrams, the Kermor layout pragma and
// the diagram simplifier of activity diagrams; they panic with "not ported
// yet".
package svek
