// SPDX-License-Identifier: MIT
// Provenance: original

package core

import (
	"io/fs"
	"maps"
	"slices"
	"time"

	"github.com/ekalinin/gopuml/internal/security"
)

// DefaultDotTimeout is upstream's Graphviz timeout.
const DefaultDotTimeout = 30 * time.Second

// RenderContext holds everything that upstream keeps in process-wide state
// (GlobalConfig, the current directory, the security profile, pragmas, the
// Graphviz settings). One context belongs to one render, so concurrent
// renders with different settings do not affect each other.
type RenderContext struct {
	// Desc describes the source in line locations, for example a file name.
	Desc string
	// Config lines are inserted after every start line (--config).
	Config []string
	// Defines are preprocessor definitions (-D).
	Defines map[string]string
	// Pragmas are pragmas given outside the source (-P).
	Pragmas map[string]string
	// FS resolves includes and themes; CurrentDir is the directory of the
	// source inside FS.
	FS         fs.FS
	CurrentDir string
	// SecurityProfile is the profile name, for example "LEGACY".
	SecurityProfile string
	// Security is the full security setting; nil means SecurityProfile
	// alone.
	Security *security.Config
	// DotPath is an explicit Graphviz dot executable, or "".
	DotPath string
	// DotTimeout stops a dot process that runs longer.
	DotTimeout time.Duration
	// PreprocOptions and PreprocWarnings are the "!option" values (by
	// upstream key name, for example "SVG_TITLE") and the warnings of the
	// block being built, upstream's PreprocessingArtifact.
	PreprocOptions  map[string]string
	PreprocWarnings []string
	// Previous are the skinparam values a new diagram starts with,
	// upstream's Previous: those of the diagram that embeds it ("{{" ...
	// "}}"), nil otherwise.
	Previous map[string]string
	// EmbeddedDiagram renders an embedded diagram; the engine sets it. nil
	// means that embedded diagrams cannot be drawn.
	EmbeddedDiagram EmbeddedDiagramRenderer
}

// EmbeddedDiagramRenderer renders the lines of an embedded diagram, from
// "@start<type>" to "@end<type>", as upstream's EmbeddedDiagram builds and
// exports a BlockUml: previous are the skinparam values of the embedding
// diagram, and the image is SVG when svg is true, PNG otherwise.
type EmbeddedDiagramRenderer func(lines []string, previous map[string]string, svg bool) ([]byte, error)

// NewRenderContext returns a context with upstream's defaults.
func NewRenderContext() *RenderContext {
	return &RenderContext{
		Desc:            "string",
		Defines:         map[string]string{},
		Pragmas:         map[string]string{},
		CurrentDir:      ".",
		SecurityProfile: "LEGACY",
		DotTimeout:      DefaultDotTimeout,
	}
}

// Clone returns an independent copy of the context.
func (c *RenderContext) Clone() *RenderContext {
	out := *c
	out.Config = slices.Clone(c.Config)
	out.Defines = maps.Clone(c.Defines)
	out.Pragmas = maps.Clone(c.Pragmas)
	out.Previous = maps.Clone(c.Previous)
	return &out
}

// Parse extracts the blocks of src and turns each into a diagram, in order.
func Parse(ctx *RenderContext, src string, factories []Factory) []Diagram {
	var out []Diagram
	for _, b := range BuildBlocks(src, ctx.Desc, ctx.Config) {
		out = append(out, CreateDiagram(ctx, b.Lines, b.Lines, factories))
	}
	return out
}
