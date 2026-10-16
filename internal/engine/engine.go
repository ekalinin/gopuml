// SPDX-License-Identifier: MIT
// Provenance: original

// Package engine parses and draws diagram sources, as upstream's
// SourceStringReader and BlockUml. The public plantuml package, the command
// line and the HTTP server are built on it.
package engine

import (
	"context"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/code"
	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/security"
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/tim"
)

// Factories are the parsers of the supported diagram types, in upstream's
// order. They are added with each diagram type.
var Factories []core.Factory

// ErrNotDrawable reports a diagram that cannot be drawn yet.
var ErrNotDrawable = errors.New("drawing of this diagram is not implemented yet")

// Block is one @start/@end block of a source and its diagram, as upstream's
// BlockUml.
type Block struct {
	// Raw holds the lines of the block; Data holds them after
	// preprocessing.
	Raw, Data []text.StringLocated
	Diagram   core.Diagram
	// Log holds the !log lines, which upstream writes to standard error.
	Log []string
}

// Source is a diagram source and where it comes from.
type Source struct {
	Text string
	// File is the path of the source file, empty for a source that is not
	// a file.
	File string
	// Filename and FileDir set %filename and %dirpath for a source that is
	// not a file (-filename and -filedir).
	Filename, FileDir string
}

// defines is CliOptions.getDefaultDefines: the file environment and the
// definitions of the context.
func defines(rc *core.RenderContext, src Source) *tim.Defines {
	var d *tim.Defines
	if src.File != "" {
		// java.io.File.lastModified is 0 for a file that cannot be read.
		modTime := time.UnixMilli(0)
		if st, err := os.Stat(src.File); err == nil {
			modTime = st.ModTime()
		}
		dirpath := ""
		if rc.SecurityProfile == security.Insecure.String() {
			if abs, err := filepath.Abs(filepath.Dir(src.File)); err == nil {
				dirpath = abs
			}
		}
		d = tim.DefinesWithFileName(filepath.Base(src.File), modTime, dirpath)
	} else {
		d = tim.NewDefines()
		d.OverrideFilename(src.Filename)
		d.OverrideDirPath(src.FileDir)
	}
	for _, name := range slices.Sorted(maps.Keys(rc.Defines)) {
		d.Define(name, []string{rc.Defines[name]})
	}
	return d
}

// Parse splits a source into blocks, preprocesses each and builds its
// diagram, as upstream's BlockUmlBuilder.
func Parse(ctx context.Context, rc *core.RenderContext, src Source) ([]*Block, error) {
	defs := &tim.BlockDefinitions{}
	env := defines(rc, src)
	access := &security.Access{FS: rc.FS}
	if rc.Security != nil {
		access.Config = *rc.Security
	} else {
		access.Config.Profile, _ = security.ParseProfile(rc.SecurityProfile)
	}
	var out []*Block
	for _, blk := range core.BuildBlocks(src.Text, rc.Desc, rc.Config) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := tim.PreprocessWith(rc, access, blk.Lines, tim.Options{Definitions: defs, Defines: env.CloneMe()})
		if res == nil {
			return nil, err
		}
		b := &Block{Raw: blk.Lines, Data: res.Lines, Log: res.Log}
		defs.Add(res.Lines)
		if err != nil {
			b.Diagram = preprocessorError(res, err)
		} else {
			b.Diagram = core.CreateDiagram(rc, res.Lines, blk.Lines, Factories)
		}
		out = append(out, b)
	}
	return out, nil
}

// preprocessorError is upstream's PSystemErrorPreprocessor: a syntax error
// on the last line the preprocessor read.
func preprocessorError(res *tim.Result, err error) *core.ErrorDiagram {
	var last text.StringLocated
	if n := len(res.Debug); n > 0 {
		last = res.Debug[n-1]
	}
	message := last.PreprocessorError
	if message == "" {
		message = err.Error()
	}
	return &core.ErrorDiagram{
		Source: res.Lines,
		Trace:  res.Debug,
		Error:  core.ErrorUml{Type: core.SyntaxError, Error: message, Line: &last},
	}
}

// Name is BlockUml.getFileOrDirname: the name given on the preprocessed
// start line.
func (b *Block) Name() (string, bool) { return (&core.Block{Lines: b.Data}).Name() }

var familyPattern = regexp.MustCompile(`^@start([A-Za-z]*)`)

// Family returns the word after @start, in lower case, such as "uml".
func (b *Block) Family() string {
	if m := familyPattern.FindStringSubmatch(strings.TrimSpace(b.Raw[0].S)); m != nil {
		return strings.ToLower(m[1])
	}
	return ""
}

// Source is BlockUml.getDefinition(true): the preprocessed lines of the
// block, from the start line to the end line.
func (b *Block) Source() []string {
	out := make([]string, len(b.Data))
	for i, l := range b.Data {
		out[i] = l.S
	}
	return out
}

// Encoded is BlockUml.getEncodedUrl: the block in the text format of server
// URLs.
func (b *Block) Encoded() string { return code.Encode(strings.Join(b.Source(), "\n")) }

// Pages returns the number of images of the diagram.
func (b *Block) Pages() int { return 1 }

// ImageData is upstream's ImageData: the size of the drawn image and its
// status, 0 or 200 on success.
type ImageData struct {
	Width, Height float64
	Status        int
}

// Options select the output of Render.
type Options struct {
	Format     font.FileFormat
	NoMetadata bool
	DarkMode   bool
}

// Render draws one page of the diagram to w.
func (b *Block) Render(ctx context.Context, w io.Writer, page int, opt Options) (ImageData, error) {
	if err := ctx.Err(); err != nil {
		return ImageData{}, err
	}
	// Drawing is added with the first diagram type.
	return ImageData{}, ErrNotDrawable
}
