// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"io"
	"time"
)

// Engine reads the diagrams of a source for the command line. The library
// provides the implementation; tests use a fake.
type Engine interface {
	// Read splits a source into blocks and builds the diagram of each.
	Read(src *Source) ([]Block, error)
}

// Source is one input: a file, or a diagram read from standard input.
type Source struct {
	// Text is the source text.
	Text string
	// File is the input path, empty for standard input.
	File string
	// FileDir and Filename are -filedir and -filename, for standard input.
	FileDir, Filename string
	// Defines are the -D definitions, in order.
	Defines []KeyValue
	// Config lines are inserted at the start of each diagram.
	Config []string
	// DotPath and DotTimeout are --dot-path and --graphviz-timeout; empty
	// and zero mean the defaults.
	DotPath    string
	DotTimeout time.Duration
}

// Block is one @start/@end block of a source and its diagram.
type Block interface {
	// Name returns the name given on the start line.
	Name() (string, bool)
	// Error returns the error of a diagram that failed, or nil.
	Error() *BlockError
	// Type is upstream's diagram type name, such as SEQUENCE, or "" when
	// the diagram has none.
	Type() string
	// Description is upstream's description, such as "(2 participants)".
	Description() string
	// EncodedURL is the source in the text format of server URLs.
	EncodedURL() string
	// Preprocessed returns the lines of the block after preprocessing,
	// including the start and end lines.
	Preprocessed() []string
	// Pages is the number of images of the diagram.
	Pages() int
	// Metadata is the text upstream embeds in SVG and PNG output (the
	// diagram source and the version); ok is false when the diagram has
	// none.
	Metadata() (string, bool)
	// Log returns the !log lines of the block, which upstream prints to
	// standard error while preprocessing.
	Log() []string
	// Render writes one page and returns upstream's image status: 0 or 200
	// on success, another code (such as 400 or 503) when the image shows a
	// failure.
	Render(w io.Writer, page int, opt RenderOptions) (status int, err error)
}

// BlockError describes a diagram that failed.
type BlockError struct {
	// Unsupported marks a diagram type this implementation does not
	// provide; otherwise the block has a syntax error.
	Unsupported bool
	// Line is the 0-based line of the error in the source.
	Line int
	// Messages are the error texts.
	Messages []string
}

// RenderOptions select the output of Block.Render.
type RenderOptions struct {
	// Format is upstream's format name, such as SVG.
	Format string
	// NoMetadata omits the embedded source.
	NoMetadata bool
	// DarkMode renders with upstream's dark color mapping.
	DarkMode bool
}
