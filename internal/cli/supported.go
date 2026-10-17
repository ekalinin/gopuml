// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"fmt"

	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// supported are the upstream flags this implementation provides; the cli
// spec lists them. Any other upstream flag is rejected with exit status 1.
var supported = map[Flag]bool{
	// Formats.
	TSvg: true, TPng: true, TTxt: true, TUtxt: true, Preprocess: true, Format: true,
	// Pipe mode.
	Pipe: true, Pipedelimitor: true, Pipenostderr: true, PipeImageIndex: true,
	// Input and output.
	OutputDir: true, Charset: true, Exclude: true, Overwrite: true, IgnoreStartumlFilename: true,
	UseSeparatorMinus: true, FileDir: true, Filename: true,
	// Preprocessing and styling.
	Define: true, DefineLong: true, Include: true, IncludeLong: true, Pragma: true, PragmaLong: true,
	Skinparam: true, SkinparamLong: true, Theme: true, Config: true, DarkMode: true,
	// Error handling.
	CheckOnly: true, Syntax: true, FailFast: true, FailFast2: true, NoErrorImage: true,
	// Graphviz.
	Timeout: true, GraphvizDot: true, TestDot: true,
	// Utilities.
	NbThread: true, ComputeUrl: true, DecodeUrl: true, RetrieveMetadata: true, NoMetadata: true,
	CheckMetadata: true, Language: true, StdLib: true, Picoweb: true,
	// Information and logging.
	Version: true, Help: true, HelpMore: true, License: true, Verbose: true, Duration: true, Headless: true,
}

// SupportedFormats are the -f/--format values this implementation provides,
// with the output format each selects.
var SupportedFormats = map[string]string{"svg": "SVG", "png": "PNG", "txt": "ATXT", "utxt": "UTXT", "preproc": "PREPROC"}

// Supported reports whether the implementation provides a flag.
func (f Flag) Supported() bool { return supported[f] }

// CheckSupported returns an *UnsupportedError for the first flag of the
// command line that this implementation does not provide, or an
// *UnsupportedFormatError for a -f value upstream knows but this
// implementation does not provide. Other -f values are ignored, as upstream.
func (p *Parsed) CheckSupported() error {
	for _, f := range p.Given() {
		if !f.Supported() {
			return &UnsupportedError{Flag: f}
		}
	}
	if v, ok := p.String(Format); ok {
		if _, known := font.FileFormatFromCli(v); known && SupportedFormats[v] == "" {
			return &UnsupportedFormatError{Name: v}
		}
	}
	return nil
}

// UnsupportedFormatError reports a -f value this implementation does not
// provide; the command line fails with exit status 1.
type UnsupportedFormatError struct{ Name string }

func (e *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("the format %q is not supported by this implementation", e.Name)
}
