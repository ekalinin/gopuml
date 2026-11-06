// SPDX-License-Identifier: MIT
// Provenance: original

// Package directdot reads and draws dot diagrams (@startdot): a Graphviz
// graph that the layout engine of the render draws itself, as upstream's
// directdot package. The SVG is the output of the engine for -Tsvg in the
// INSECURE and INTERNET_WITH_DOTSVG security profiles and upstream's
// suppressed-feature document in the others; the PNG is the output of a
// named dot for -Tpng, or upstream's "GraphViz has crashed" image with the
// built-in engine, which does not write PNG.
package directdot
