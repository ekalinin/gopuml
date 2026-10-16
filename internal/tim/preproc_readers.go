// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/DiagramDetector.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/DiagramExtractor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/ReadLineWithYamlHeader.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/ReadLineSimple.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/ReadLineList.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc2/PreprocessorUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc2/PreprocessorIncludeStrategy.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"regexp"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// PreprocessorIncludeStrategy is the variant of !include.
type PreprocessorIncludeStrategy int

// Include strategies.
const (
	includeOnce PreprocessorIncludeStrategy = iota
	includeMany
	includeDefault
)

// readAll returns the remaining lines of a reader.
func readAll(r core.ReadLine) []text.StringLocated {
	var out []text.StringLocated
	for {
		l, ok := r.ReadLine()
		if !ok {
			return out
		}
		out = append(out, l)
	}
}

// uncommentAndMerge is DiagramDetector.uncommentAndMerge.
func uncommentAndMerge(r core.ReadLine) core.ReadLine {
	return core.NewUncommentReader(core.NewMergeLinesReader(r))
}

func containsStartDiagram(r core.ReadLine) bool {
	for {
		s, ok := r.ReadLine()
		if !ok {
			return false
		}
		if core.IsStartDirective(s.S) {
			return true
		}
	}
}

// extractFromData is DiagramDetector.extractFromFile and extractFromBytes:
// when data contains a start directive, the lines of its first diagram (or
// of the diagram selected by suf), else nil.
func extractFromData(data []byte, description string, suf *string) core.ReadLine {
	if !containsStartDiagram(uncommentAndMerge(core.NewLineReader(string(data), description, nil))) {
		return nil
	}
	return newDiagramExtractor(uncommentAndMerge(core.NewLineReader(string(data), description, nil)), suf)
}

// diagramExtractor reads the lines of one diagram of a source, without its
// start and end lines.
type diagramExtractor struct {
	raw      core.ReadLine
	finished bool
}

var digits = regexp.MustCompile(`^[0-9]+$`)

// newDiagramExtractor skips to the diagram selected by suf: a block
// number, an id (matched against "id=<suf>" on the start line) or, when
// nil, the first diagram.
func newDiagramExtractor(raw core.ReadLine, suf *string) *diagramExtractor {
	block := 0
	var uid *string
	if suf != nil && digits.MatchString(*suf) {
		v, ok := javaParseInt(*suf)
		if !ok {
			throwRuntime("NumberFormatException", `For input string: "`+*suf+`"`)
		}
		block = int(v)
	} else {
		uid = suf
	}
	d := &diagramExtractor{raw: raw}
	for {
		s, ok := raw.ReadLine()
		if !ok {
			break
		}
		if core.IsStartDirective(s.S) && checkUID(uid, s) {
			if block == 0 {
				return d
			}
			block--
		}
	}
	d.finished = true
	return d
}

// checkUID tests the start line against Java's ".*id=<uid>\W.*" on the
// toString of the line ("(SL) " followed by the text). Java's "." does not
// match line terminators.
func checkUID(uid *string, s text.StringLocated) bool {
	if uid == nil {
		return true
	}
	const notEOL = `[^\n\r\x{85}\x{2028}\x{2029}]*`
	re, err := regexp.Compile(`^` + notEOL + `id=` + *uid + `\W` + notEOL + `$`)
	if err != nil {
		throwRuntime("PatternSyntaxException", err.Error())
	}
	return re.MatchString("(SL) " + s.S)
}

func (d *diagramExtractor) ReadLine() (text.StringLocated, bool) {
	if d.finished {
		return text.StringLocated{}, false
	}
	result, ok := d.raw.ReadLine()
	if ok && core.IsEndDirective(result.S) {
		d.finished = true
		return text.StringLocated{}, false
	}
	return result, ok
}

// readLineWithYamlHeader skips a "---" delimited header at the start of a
// source and keeps its "key: value" lines as metadata.
type readLineWithYamlHeader struct {
	source     core.ReadLine
	yamlHeader []string
	started    bool
	metadata   []metadataEntry
}

type metadataEntry struct{ key, value string }

func newReadLineWithYamlHeader(source core.ReadLine) *readLineWithYamlHeader {
	return &readLineWithYamlHeader{source: source}
}

func (r *readLineWithYamlHeader) ReadLine() (text.StringLocated, bool) {
	line, ok := r.source.ReadLine()
	if !r.started {
		// At this point, "line" is the very first line of the file
		r.started = true
		if ok && line.S == "---" {
			for {
				line, ok = r.source.ReadLine()
				if !ok || line.S == "---" {
					break
				}
				tmp := line.S
				r.yamlHeader = append(r.yamlHeader, tmp)
				if idx := strings.Index(tmp, ":"); idx > 0 {
					r.putMetadata(text.Trin(tmp[:idx]), text.Trin(tmp[idx+1:]))
				}
			}
			// Skip the second separator
			line, ok = r.source.ReadLine()
		}
	}
	return line, ok
}

// putMetadata is LinkedHashMap.put: a key keeps its first position.
func (r *readLineWithYamlHeader) putMetadata(key, value string) {
	for i := range r.metadata {
		if r.metadata[i].key == key {
			r.metadata[i].value = value
			return
		}
	}
	r.metadata = append(r.metadata, metadataEntry{key, value})
}

// readLineSimple returns one line, marked with an error.
type readLineSimple struct {
	data    text.StringLocated
	err     string
	current int
}

func (r *readLineSimple) ReadLine() (text.StringLocated, bool) {
	if r.current > 0 {
		return text.StringLocated{}, false
	}
	r.current++
	d := r.data
	d.PreprocessorError = r.err
	return d, true
}

// readLineList returns fixed lines, all at the same location.
type readLineList struct {
	lines    []string
	location *text.LineLocation
}

func (r *readLineList) ReadLine() (text.StringLocated, bool) {
	if len(r.lines) == 0 {
		return text.StringLocated{}, false
	}
	line := r.lines[0]
	r.lines = r.lines[1:]
	return text.NewStringLocated(line, r.location), true
}

// getReaderStdlibInclude is PreprocessorUtils.getReaderStdlibInclude.
func getReaderStdlibInclude(stdlib *Stdlib, filename string) core.ReadLine {
	puml := stdlib.getPumlResource(filename)
	if puml == nil {
		return nil
	}
	description := "<" + filename + ">"
	if tmp := extractFromData(puml, description, nil); tmp != nil {
		return tmp
	}
	return core.NewLineReader(string(puml), description, nil)
}

// getReaderIncludeURL is PreprocessorUtils.getReaderIncludeUrl.
func (c *TContext) getReaderIncludeURL(url string, s text.StringLocated, suf *string) (core.ReadLine, error) {
	if tmp := c.extractFromURL(url, s, suf); tmp != nil {
		return tmp, nil
	}
	return c.getReaderInclude(url, s)
}

// extractFromURL is DiagramDetector.extractFromUrl. Like upstream, it
// fetches the URL a second time to read the diagram.
func (c *TContext) extractFromURL(url string, s text.StringLocated, uid *string) core.ReadLine {
	if !containsStartDiagram(c.newReadLineFromURL(url, s)) {
		return nil
	}
	return newDiagramExtractor(c.newReadLineFromURL(url, s), uid)
}

func (c *TContext) newReadLineFromURL(url string, s text.StringLocated) core.ReadLine {
	data := c.fa.openURL(url)
	if data == nil {
		return &readLineSimple{data: s, err: "Cannot connect"}
	}
	return uncommentAndMerge(core.NewLineReader(string(data), url, nil))
}

// getReaderInclude is PreprocessorUtils.getReaderInclude.
func (c *TContext) getReaderInclude(url string, s text.StringLocated) (core.ReadLine, error) {
	data := c.fa.openURL(url)
	if data == nil {
		return nil, newEaterException("Cannot open URL", s)
	}
	return core.NewLineReader(string(data), url, s.Location), nil
}
