// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ekalinin/gopuml/internal/cli"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
	"github.com/ekalinin/gopuml/internal/security"
)

// cliEngine runs the command line on the diagram engine.
type cliEngine struct{}

func (cliEngine) Read(src *cli.Source) ([]cli.Block, error) {
	rc := core.NewRenderContext()
	cfg := security.ConfigFromEnv(os.Getenv, security.Legacy)
	rc.SecurityProfile, rc.Security = cfg.Profile.String(), &cfg
	rc.Config = src.Config
	rc.DotPath = src.DotPath
	if src.DotTimeout > 0 {
		rc.DotTimeout = src.DotTimeout
	}
	for _, d := range src.Defines {
		rc.Defines[d.Key] = d.Value
	}
	switch {
	case src.File != "":
		rc.Desc = filepath.Base(src.File)
		rc.CurrentDir = filepath.Dir(src.File)
	default:
		if src.Filename != "" {
			rc.Desc = src.Filename
		}
		if src.FileDir != "" {
			rc.CurrentDir = src.FileDir
		}
	}
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{
		Text: src.Text, File: src.File, Filename: src.Filename, FileDir: src.FileDir,
	})
	if err != nil {
		return nil, err
	}
	out := make([]cli.Block, len(blocks))
	for i, b := range blocks {
		out[i] = cliBlock{b}
	}
	return out, nil
}

// cliBlock is an engine block seen by the command line.
type cliBlock struct{ *engine.Block }

// Error maps upstream's PSystemError to a syntax error and PSystemUnsupported
// to an unsupported diagram.
func (b cliBlock) Error() *cli.BlockError {
	switch d := b.Diagram.(type) {
	case *core.ErrorDiagram:
		return &cli.BlockError{Line: d.Position(), Messages: []string{d.Error.Message()}}
	case *core.UnsupportedDiagram:
		return &cli.BlockError{Unsupported: true}
	}
	return nil
}

func (b cliBlock) Type() string {
	if d, ok := b.Diagram.(core.TypedDiagram); ok {
		return d.DiagramType().String()
	}
	return ""
}

func (b cliBlock) Description() string { return b.Diagram.Description() }

func (b cliBlock) EncodedURL() string { return b.Encoded() }

func (b cliBlock) Preprocessed() []string { return b.Source() }

func (b cliBlock) Log() []string { return b.Block.Log }

// formats are the output formats of the command line, by upstream name.
var formats = map[string]font.FileFormat{
	"SVG": font.FormatSVG, "PNG": font.FormatPNG, "ATXT": font.FormatATXT, "UTXT": font.FormatUTXT,
}

func (b cliBlock) Render(w io.Writer, page int, opt cli.RenderOptions) (int, error) {
	f, ok := formats[opt.Format]
	if !ok {
		return 0, fmt.Errorf("format %s is not supported", opt.Format)
	}
	img, err := b.Block.Render(context.Background(), w, page, engine.Options{Format: f, NoMetadata: opt.NoMetadata, DarkMode: opt.DarkMode})
	return img.Status, err
}
