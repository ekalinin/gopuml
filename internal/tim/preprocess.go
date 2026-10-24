// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"maps"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/text"
)

// Result is the outcome of preprocessing one block.
type Result struct {
	// Lines are the preprocessed lines with their source locations
	// (upstream's BlockUml data, which the PREPROC output prints). After an
	// error, the last line is the offending line, with PreprocessorError
	// set to the message.
	Lines []text.StringLocated
	// Debug lists every line the preprocessor read, in order; after an
	// error, the last one carries the message (the trace of the error
	// diagram).
	Debug []text.StringLocated
	// Artifact holds the !option values and the warnings.
	Artifact *PreprocessingArtifact
	// Included lists the files and URLs read by includes and themes.
	// Upstream returns an empty set here (its "used files" set is never
	// filled); the list is informational.
	Included []string
	// Log holds the !log lines ("[Log] ..."), which upstream writes to the
	// standard error stream.
	Log []string
	// Info holds the diagnostic lines of !dump_memory, which upstream
	// prints with -verbose only.
	Info []string
	// ThemeMetadata is the YAML header of the last theme loaded.
	ThemeMetadata *json.Object
}

// PreprocLines returns the text of the lines, as the PREPROC output
// prints them (BlockUml.getDefinition(true)).
func (r *Result) PreprocLines() []string {
	out := make([]string, len(r.Lines))
	for i, l := range r.Lines {
		out[i] = l.S
	}
	return out
}

// Options are the optional inputs of PreprocessWith.
type Options struct {
	// Definitions gives !includedef the @startdef blocks of the source.
	Definitions DefinitionsContainer
	// Defines are the -D definitions and the file environment; nil builds
	// them from the context's Defines.
	Defines *Defines
	// Themes are the bundled themes; nil means none.
	Themes ThemeSource
	// Stdlib is the standard library index; nil means the embedded one.
	Stdlib *Stdlib
	// Now returns the current time of %now and %date; nil uses time.Now.
	Now func() time.Time
	// Location is the default time zone of %date (upstream's
	// TimeZone.getDefault()); nil uses time.Local.
	Location *time.Location
	// Getenv looks up the names given to %getenv, which upstream reads as
	// Java system properties, then as environment variables; nil reads the
	// environment of the process and answers Java's line.separator,
	// path.separator and file.separator properties.
	Getenv func(name string) (string, bool)
	// RandomSeed seeds the java.util.Random generator of %random; nil
	// seeds it at random, as upstream does.
	RandomSeed *int64
}

// Preprocess runs the preprocessor on the lines of one block, from its
// start line to its end line. Files and URLs are read through access; a
// nil access uses ctx.FS and ctx.SecurityProfile. On a preprocessing error
// it returns the result together with an *EaterException holding the
// message and the offending line (upstream's PSystemErrorPreprocessor).
func Preprocess(ctx *core.RenderContext, access *security.Access, lines []text.StringLocated, defs DefinitionsContainer) (*Result, error) {
	return PreprocessWith(ctx, access, lines, Options{Definitions: defs})
}

// PreprocessWith is Preprocess with all options.
func PreprocessWith(ctx *core.RenderContext, access *security.Access, lines []text.StringLocated, opts Options) (*Result, error) {
	if access == nil {
		profile, _ := security.ParseProfile(ctx.SecurityProfile)
		access = &security.Access{Config: security.Config{Profile: profile}, FS: ctx.FS}
	}
	defines := opts.Defines
	if defines == nil {
		defines = NewDefines()
		for _, name := range sortedKeys(ctx.Defines) {
			defines.Define(name, []string{ctx.Defines[name]})
		}
	}
	stdlib := opts.Stdlib
	if stdlib == nil {
		stdlib = embeddedStdlib
	}
	fa := &fileAccess{access: access}
	currentDir := currentDirectory(ctx.CurrentDir, access)
	pathSystem := &PathSystem{fa: fa, stdlib: stdlib, currentFolder: &nFolderRegular{fa: fa, dir: currentDir},
		importedFolders: &[]*nFolderZip{}}

	var location text.StringLocated
	if len(lines) > 0 {
		location = lines[0]
	}
	loader := newTimLoader(pathSystem, defines, opts.Definitions, location)
	loader.context.themes = opts.Themes
	loader.context.now = opts.Now
	loader.context.location = opts.Location
	loader.context.getenv = opts.Getenv
	loader.context.randomSeed = opts.RandomSeed
	loader.context.currentDir = currentDir
	loader.load(lines)

	c := loader.context
	result := &Result{
		Lines:         mutateExpands1(loader.resultList),
		Debug:         c.debug,
		Artifact:      c.preprocessingArtifact,
		Included:      c.included,
		Log:           c.userLog,
		Info:          c.info,
		ThemeMetadata: c.themeMetadata,
	}
	if loader.preprocessorError != nil {
		return result, loader.preprocessorError
	}
	return result, nil
}

// currentDirectory returns the absolute directory of the source: a slash
// path inside access.FS, or an operating system path.
func currentDirectory(dir string, access *security.Access) string {
	if dir == "" {
		dir = "."
	}
	if access.FS != nil {
		return path.Clean("/" + filepath.ToSlash(dir))
	}
	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}
	return dir
}

func sortedKeys(m map[string]string) []string { return slices.Sorted(maps.Keys(m)) }
