// SPDX-License-Identifier: MIT
// Provenance: original

package plantuml

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/code"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/engine"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/security"
)

// Format is an output format.
type Format int

// Formats. The zero value is SVG.
const (
	SVG  Format = iota
	PNG         // PNG image
	TXT         // ASCII art
	UTXT        // Unicode art
)

var formatNames = [...]string{SVG: "svg", PNG: "png", TXT: "txt", UTXT: "utxt"}

var fileFormats = [...]font.FileFormat{SVG: font.FormatSVG, PNG: font.FormatPNG, TXT: font.FormatATXT, UTXT: font.FormatUTXT}

func (f Format) String() string {
	if f < 0 || int(f) >= len(formatNames) {
		return fmt.Sprintf("Format(%d)", int(f))
	}
	return formatNames[f]
}

// Define is a preprocessor definition, like -DNAME=VALUE on the command
// line.
type Define struct {
	Name, Value string
}

// Options configure one call; they do not affect other calls. A nil
// *Options means the zero value.
type Options struct {
	// Format is the output format.
	Format Format
	// Diagram and Page select what Render draws: the index of the
	// @start/@end block and the page within it, both from 0.
	Diagram, Page int
	// NoMetadata omits the embedded source from SVG and PNG output.
	NoMetadata bool
	// NoErrorImage returns syntax errors without an error image.
	NoErrorImage bool
	// Defines are preprocessor definitions, in order.
	Defines []Define
	// Config lines are inserted after the start line of every diagram.
	Config []string
	// IncludePaths are directories searched by !include.
	IncludePaths []string
	// Theme is applied to every diagram, like !theme.
	Theme string
	// DarkMode renders with upstream's dark color mapping.
	DarkMode bool
	// SecurityProfile is upstream's profile name, such as SANDBOX or
	// INTERNET; empty means LEGACY.
	SecurityProfile string
	// DotPath is the Graphviz dot program; empty means it is searched for.
	DotPath string
	// DotTimeout stops dot after this duration; zero means 30 seconds.
	DotTimeout time.Duration
	// FS resolves includes and themes. When nil, the operating system's
	// file system is used, subject to the security profile.
	FS fs.FS
	// Dir is the directory of the source, against which relative includes
	// are resolved.
	Dir string
	// Filename is the name of the source, as %filename() returns it.
	Filename string
}

// Result is a rendered page.
type Result struct {
	// Data is the output: an image, or text for TXT and UTXT.
	Data []byte
	// Width and Height are the size of the image in pixels.
	Width, Height int
	// Description is upstream's description, such as "(2 participants)".
	Description string
	// Pages is the number of pages of the rendered diagram.
	Pages int
	// Diagrams is the number of @start/@end blocks of the source.
	Diagrams int
}

// Block is one @start/@end block of a source, with its diagram.
type Block struct {
	// Name is the name given on the start line, such as "login" for
	// "@startuml login", or "".
	Name string
	// Family is the word after @start, in lower case, such as "uml".
	Family string
	// Type is the diagram type upstream detects, such as "SEQUENCE" or
	// "CLASS", or "" for a block with an error.
	Type string
	// Description is upstream's description, such as "(2 participants)" or
	// "(Error)".
	Description string
	// Err is the syntax status: nil, a *SyntaxError, or an error matching
	// ErrUnsupportedDiagram.
	Err error

	b *engine.Block
}

// renderContext turns options into the context of one render.
func renderContext(opts *Options) (*core.RenderContext, error) {
	rc := core.NewRenderContext()
	if opts == nil {
		return rc, nil
	}
	// The CliOptions constructor adds the --theme line before the --config
	// lines.
	if opts.Theme != "" {
		rc.Config = append(rc.Config, "!theme "+opts.Theme)
	}
	rc.Config = append(rc.Config, opts.Config...)
	for _, d := range opts.Defines {
		rc.Defines[d.Name] = d.Value
	}
	if opts.Filename != "" {
		rc.Desc = opts.Filename
	}
	cfg := security.Config{Profile: security.Legacy, IncludePaths: opts.IncludePaths, HostFiles: opts.FS != nil}
	if opts.SecurityProfile != "" {
		p, ok := security.ParseProfile(opts.SecurityProfile)
		if !ok {
			return nil, fmt.Errorf("plantuml: unknown security profile %q", opts.SecurityProfile)
		}
		cfg.Profile = p
	}
	rc.SecurityProfile, rc.Security = cfg.Profile.String(), &cfg
	rc.FS = opts.FS
	if opts.Dir != "" {
		rc.CurrentDir = opts.Dir
	}
	rc.DotPath = opts.DotPath
	if opts.DotTimeout > 0 {
		rc.DotTimeout = opts.DotTimeout
	}
	return rc, nil
}

// Inspect lists the blocks of a source with their diagram type and syntax
// status, without drawing them.
func Inspect(ctx context.Context, source string, opts *Options) ([]*Block, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rc, err := renderContext(opts)
	if err != nil {
		return nil, err
	}
	src := engine.Source{Text: source}
	if opts != nil {
		src.Filename = opts.Filename
	}
	blocks, err := engine.Parse(ctx, rc, src)
	if err != nil {
		return nil, err
	}
	out := make([]*Block, len(blocks))
	for i, eb := range blocks {
		b := &Block{Family: eb.Family(), Description: eb.Diagram.Description(), b: eb}
		b.Name, _ = eb.Name()
		switch d := eb.Diagram.(type) {
		case *core.ErrorDiagram:
			if d.NotSupportedYet {
				b.Err = &UnsupportedError{Line: d.Position() + 1, Message: d.Error.Message()}
			} else {
				b.Err = &SyntaxError{Line: d.Position() + 1, Message: d.Error.Message()}
			}
		case *core.UnsupportedDiagram:
			line := 1
			if loc := eb.Raw[0].Location; loc != nil {
				line = loc.Position + 1
			}
			b.Err = &UnsupportedError{Line: line, Message: "Diagram not supported by this release of PlantUML: " + d.Directive}
		case core.TypedDiagram:
			b.Type = d.DiagramType().String()
		}
		out[i] = b
	}
	return out, nil
}

// Source returns the lines of the block, from the start line to the end
// line.
func (b *Block) Source() []string { return b.b.Source() }

// Encoded returns the block in the text format of PlantUML server URLs.
func (b *Block) Encoded() string { return b.b.Encoded() }

// Pages returns the number of pages of the diagram.
func (b *Block) Pages() int { return b.b.Pages() }

// Render draws one page of the block to w. A block with an error returns
// its error (a *SyntaxError or an *UnsupportedError) together with the
// error image, unless NoErrorImage is set.
func (b *Block) Render(ctx context.Context, w io.Writer, page int, opts *Options) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if page < 0 || page >= b.Pages() {
		return nil, fmt.Errorf("%w: page %d out of range (%d pages)", ErrNoDiagram, page, b.Pages())
	}
	var o Options
	if opts != nil {
		o = *opts
	}
	if b.Err != nil && o.NoErrorImage {
		return nil, b.Err
	}
	if o.Format < 0 || int(o.Format) >= len(fileFormats) {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedFormat, o.Format)
	}
	img, err := b.b.Render(ctx, w, page, engine.Options{Format: fileFormats[o.Format], NoMetadata: o.NoMetadata, DarkMode: o.DarkMode})
	if errors.Is(err, engine.ErrNotDrawable) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDiagram, err)
	}
	if err != nil {
		return nil, err
	}
	// A block with an error returns its error image with the error.
	return &Result{
		Width:       int(math.Round(img.Width)),
		Height:      int(math.Round(img.Height)),
		Description: b.Description,
		Pages:       b.Pages(),
	}, b.Err
}

// Render draws the diagram and page selected by opts.Diagram and opts.Page.
// For a syntax error it returns the error image, unless opts.NoErrorImage
// is set, together with the *SyntaxError.
func Render(ctx context.Context, source string, opts *Options) (*Result, error) {
	blocks, err := Inspect(ctx, source, opts)
	if err != nil {
		return nil, err
	}
	var o Options
	if opts != nil {
		o = *opts
	}
	if o.Diagram < 0 || o.Diagram >= len(blocks) {
		return nil, fmt.Errorf("%w: diagram %d out of range (%d diagrams)", ErrNoDiagram, o.Diagram, len(blocks))
	}
	var buf bytes.Buffer
	res, err := blocks[o.Diagram].Render(ctx, &buf, o.Page, opts)
	if res != nil {
		res.Data = buf.Bytes()
		res.Diagrams = len(blocks)
	}
	return res, err
}

// Encode returns source in the text format of PlantUML server URLs.
func Encode(source string) string { return code.Encode(source) }

// Decode returns the source of text in any format PlantUML server URLs use.
// It fails with an error matching ErrDecode.
func Decode(text string) (string, error) { return code.Decode(text) }
