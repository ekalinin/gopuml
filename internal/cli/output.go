// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/SourceFileReader.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/SourceFileReaderAbstract.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/file/SuggestedFile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/FileFormat.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliOptions.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
)

// Output file suffixes of the formats this implementation writes.
var formatSuffix = map[string]string{"SVG": ".svg", "PNG": ".png", "ATXT": ".atxt", "UTXT": ".utxt", "PREPROC": ".preproc"}

// FormatSuffix returns the file suffix of an output format.
func FormatSuffix(format string) string { return formatSuffix[format] }

// OutputFormat is CliOptions.getFileFormatOption: the -f value when upstream
// knows it, otherwise the format flag, otherwise PNG.
func (p *Parsed) OutputFormat() string {
	if v, ok := p.String(Format); ok {
		if f, ok := font.FileFormatFromCli(v); ok {
			return f.String()
		}
	}
	if f := p.Format(); f != "" {
		return f
	}
	return "PNG"
}

var lastExtension = regexp.MustCompile(`\.\w+$`)

// changeName is FileFormat.changeName: the first output replaces the
// extension; later ones add a separator and a 3-digit counter.
func changeName(fileName string, cpt int, separator, suffix string) string {
	replacement := suffix
	if cpt != 0 {
		replacement = fmt.Sprintf("%s%03d%s", separator, cpt, suffix)
	}
	result := lastExtension.ReplaceAllLiteralString(fileName, replacement)
	if result == fileName {
		result = fileName + replacement
	}
	return result
}

// OutputNamer names the output files of one source file, as upstream's
// SourceFileReader and SuggestedFile do.
type OutputNamer struct {
	// Source is the input file path; OutputDir is the resolved output
	// directory.
	Source, OutputDir string
	// Suffix is the output file suffix, for example ".svg".
	Suffix string
	// Separator precedes the counter: "_", or "-" with -useseparatorminus.
	Separator string
	// IgnoreNames uses the source name even for named diagrams.
	IgnoreNames bool

	cpt int
}

// ResolveOutputDir is the output directory of a source file: the -o
// directory when given (relative to the source file's directory), otherwise
// the source file's directory.
func ResolveOutputDir(source, outputDir string) string {
	sourceDir := filepath.Dir(source)
	switch {
	case outputDir == "":
		return sourceDir
	case filepath.IsAbs(outputDir):
		return outputDir
	}
	return filepath.Join(sourceDir, outputDir)
}

// dirIfDirectory is SourceFileReader.getDirIfDirectory: a name ending with
// a slash is a directory to create, and an existing directory is used as is.
func (n *OutputNamer) dirIfDirectory(name string) (string, bool) {
	if strings.HasSuffix(name, "/") || strings.HasSuffix(name, `\`) {
		name = name[:len(name)-1]
		dir := name
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(n.OutputDir, name)
		}
		if _, err := os.Stat(dir); err != nil {
			os.MkdirAll(dir, 0o755)
			if st, err := os.Stat(dir); err == nil && st.IsDir() {
				return dir, true
			}
		} else if st, _ := os.Stat(dir); !st.IsDir() {
			return "", false
		}
		return dir, true
	}
	dir := name
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(n.OutputDir, name)
	}
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return dir, true
	}
	return "", false
}

// Files returns the output files of the next block of the source, one per
// page. hasName and name come from the start line; pages is the number of
// images the diagram produces.
func (n *OutputNamer) Files(name string, hasName bool, pages int) []string {
	s := n.Suggest(name, hasName)
	pages = max(pages, 1)
	n.AddPages(pages)
	files := make([]string, pages)
	for i := range pages {
		files[i] = s.File(i)
	}
	return files
}

// Suggested is upstream's SuggestedFile: the output file of a block, whose
// pages are numbered from an initial counter.
type Suggested struct {
	outFile, separator, suffix string
	initial                    int
}

// File returns the file of page i.
func (s Suggested) File(i int) string {
	dir, file := filepath.Split(s.outFile)
	return filepath.Join(dir, changeName(file, s.initial+i, s.separator, s.suffix))
}

// Suggest is SourceFileReader.getSuggestedFile for the next block.
func (n *OutputNamer) Suggest(name string, hasName bool) Suggested {
	base := filepath.Base(n.Source)
	s := Suggested{separator: n.Separator, suffix: n.Suffix}
	switch {
	case !hasName || n.IgnoreNames:
		s.outFile = filepath.Join(n.OutputDir, base)
		s.initial = n.cpt
		n.cpt++
	default:
		if dir, ok := n.dirIfDirectory(name); ok {
			s.outFile = filepath.Join(dir, base)
		} else {
			s.outFile = filepath.Join(n.OutputDir, name)
		}
	}
	return s
}

// AddPages counts the extra pages of the block just suggested, as
// SourceFileReaderAbstract.getGeneratedImages does.
func (n *OutputNamer) AddPages(pages int) {
	if pages > 1 {
		n.cpt += pages - 1
	}
}
