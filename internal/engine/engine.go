// SPDX-License-Identifier: MIT
// Provenance: original

// Package engine parses and draws diagram sources, as upstream's
// SourceStringReader and BlockUml. The public gopuml package, the command
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

	"github.com/ekalinin/gopuml/internal/activitydiagram"
	"github.com/ekalinin/gopuml/internal/activitydiagram3"
	"github.com/ekalinin/gopuml/internal/cheneer"
	"github.com/ekalinin/gopuml/internal/classdiagram"
	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/descdiagram"
	"github.com/ekalinin/gopuml/internal/errdiagram"
	"github.com/ekalinin/gopuml/internal/export"
	"github.com/ekalinin/gopuml/internal/gantt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/mindmap"
	"github.com/ekalinin/gopuml/internal/packetdiag"
	"github.com/ekalinin/gopuml/internal/salt"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/sequencediagram"
	"github.com/ekalinin/gopuml/internal/statediagram"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/themes"
	"github.com/ekalinin/gopuml/internal/tim"
	"github.com/ekalinin/gopuml/internal/timing"
	"github.com/ekalinin/gopuml/internal/wbs"
)

// Factories are the parsers of the supported diagram types, in upstream's
// order (PSystemBuilder). They are added with each diagram type.
var Factories = []core.Factory{errdiagram.WelcomeFactory{}, sequencediagram.Factory, classdiagram.Factory,
	activitydiagram.Factory, descdiagram.Factory, statediagram.Factory, activitydiagram3.Factory, salt.Factory,
	mindmap.Factory, wbs.Factory, packetdiag.Factory, gantt.Factory, timing.Factory, cheneer.Factory}

// ErrNotDrawable reports a diagram that cannot be drawn yet.
var ErrNotDrawable = errors.New("drawing of this diagram is not implemented yet")

// ErrUnsupportedFormat is wrapped by the error of an output format a
// diagram type does not have in this release.
var ErrUnsupportedFormat = errors.New("output format not supported for this diagram type")

// UnsupportedFormatError is an output format a diagram type does not have
// in this release, such as TXT for a class diagram (upstream has it).
type UnsupportedFormatError struct {
	Format font.FileFormat
	Type   core.DiagramType
}

func (e *UnsupportedFormatError) Error() string {
	name := e.Format.String()
	if e.Format == font.FormatATXT {
		name = "TXT"
	}
	return name + " output is not supported for " + e.Type.HumanReadableName() + " diagrams"
}

// Unwrap returns ErrUnsupportedFormat.
func (e *UnsupportedFormatError) Unwrap() error { return ErrUnsupportedFormat }

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
	rc = withEmbeddedDiagrams(rc)
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
		res, err := tim.PreprocessWith(rc, access, blk.Lines, tim.Options{Definitions: defs, Defines: env.CloneMe(),
			Themes: themes.Bundled})
		if res == nil {
			return nil, err
		}
		b := &Block{Raw: blk.Lines, Data: res.Lines, Log: res.Log}
		defs.Add(res.Lines)
		if err != nil {
			b.Diagram = preprocessorError(res, err)
		} else {
			brc := rc
			if res.Artifact != nil {
				brc = rc.Clone()
				brc.PreprocOptions, brc.PreprocWarnings = res.Artifact.Options(), res.Artifact.Warnings()
			}
			b.Diagram = core.CreateDiagram(brc, res.Lines, blk.Lines, Factories)
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
func (b *Block) Pages() int {
	if p, ok := b.Diagram.(interface{ Pages() int }); ok {
		return p.Pages()
	}
	return 1
}

// Metadata is AbstractDiagram.getMetadata: the source and the version that
// SVG and PNG output embed; ok is false for a diagram without it.
func (b *Block) Metadata() (string, bool) {
	if d, ok := b.Diagram.(interface{ Metadata() string }); ok {
		return d.Metadata(), true
	}
	return "", false
}

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
	// Error, unsupported and welcome images.
	if errdiagram.Handles(b.Diagram) {
		fo := export.NewFileFormatOptionMetadata(opt.Format, !opt.NoMetadata)
		if opt.DarkMode {
			fo = fo.WithColorMapper(color.DarkModeMapper)
		}
		data, err := errdiagram.Export(w, b.Diagram, errdiagram.Source{Data: b.Data, Raw: b.Raw}, fo)
		if err != nil {
			return ImageData{}, err
		}
		return ImageData{Width: float64(data.Width()), Height: float64(data.Height()), Status: data.Status()}, nil
	}
	// Only sequence, gantt and WBS diagrams have the text formats in this
	// release.
	if opt.Format == font.FormatATXT || opt.Format == font.FormatUTXT {
		if d, ok := b.Diagram.(core.TypedDiagram); ok && d.DiagramType() != core.Sequence &&
			d.DiagramType() != core.Gantt && d.DiagramType() != core.WBS {
			return ImageData{}, &UnsupportedFormatError{Format: opt.Format, Type: d.DiagramType()}
		}
	}
	// Diagrams that draw themselves; an output format they do not support
	// yet is reported as not drawable.
	if d, ok := b.Diagram.(DrawableDiagram); ok {
		fo := export.NewFileFormatOptionMetadata(opt.Format, !opt.NoMetadata)
		if opt.DarkMode {
			fo = fo.WithColorMapper(color.DarkModeMapper)
		}
		var data export.ImageData
		var err error
		if dc, ok := d.(ContextDrawableDiagram); ok {
			data, err = dc.ExportDiagramContext(ctx, w, page, fo)
		} else {
			data, err = d.ExportDiagram(w, page, fo)
		}
		if errors.Is(err, export.ErrNotImplemented) {
			return ImageData{}, notDrawable{err}
		}
		if err != nil && data == nil {
			return ImageData{}, err
		}
		// A graph layout error comes with the image that shows it; the image
		// then has an error status (upstream's FileImageData.CRASH), so the
		// command line keeps the file and exits with 200.
		status := data.Status()
		if err != nil && (status == 0 || status == 200) {
			status = 503
		}
		return ImageData{Width: float64(data.Width()), Height: float64(data.Height()), Status: status}, err
	}
	// Diagrams whose drawing is not ported yet report it from Export.
	if d, ok := b.Diagram.(interface{ Export(page int) error }); ok {
		if err := d.Export(page); err != nil {
			return ImageData{}, notDrawable{err}
		}
	}
	return ImageData{}, ErrNotDrawable
}

// DrawableDiagram is a diagram that draws its pages, as upstream's
// Diagram.exportDiagram.
type DrawableDiagram interface {
	ExportDiagram(w io.Writer, page int, opt export.FileFormatOption) (export.ImageData, error)
}

// ContextDrawableDiagram is a drawable diagram whose drawing can be
// cancelled (the graph diagrams, which run Graphviz dot). It may return an
// error together with the data of the image it wrote, which then shows the
// error (dot missing, failing or timed out).
type ContextDrawableDiagram interface {
	ExportDiagramContext(ctx context.Context, w io.Writer, page int, opt export.FileFormatOption) (export.ImageData, error)
}

// notDrawable is the error of a diagram that cannot be drawn yet: it reads
// as the diagram's own message and is ErrNotDrawable.
type notDrawable struct{ err error }

func (e notDrawable) Error() string   { return e.err.Error() }
func (e notDrawable) Unwrap() []error { return []error{ErrNotDrawable, e.err} }
