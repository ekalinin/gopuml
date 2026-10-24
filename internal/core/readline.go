// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/ReadLineReader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/UncommentReadLine.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc/ReadLineList.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc2/ReadFilterAddConfig.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc2/ReadFilterMergeLines.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/preproc2/ReadFilterQuoteComment.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package core

import (
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// ReadLine is a source of located lines. ReadLine returns ok false at the end.
type ReadLine interface {
	ReadLine() (line text.StringLocated, ok bool)
}

// lineReader splits text into lines like Java's BufferedReader.readLine:
// lines end at \n, \r or \r\n, and a final line terminator does not start a
// new line. A byte order mark at the start of a line is removed and en
// dashes are replaced with hyphens, as upstream does.
type lineReader struct {
	rest     string
	done     bool
	location *text.LineLocation
}

func newLineReader(src, desc string, parent *text.LineLocation) *lineReader {
	return &lineReader{rest: strings.ToValidUTF8(src, "\uFFFD"), location: text.NewLineLocation(desc, parent)}
}

func (r *lineReader) ReadLine() (text.StringLocated, bool) {
	r.location = r.location.OneLineRead()
	if r.done || r.rest == "" {
		r.done = true
		return text.StringLocated{}, false
	}
	line := r.rest
	if i := strings.IndexAny(r.rest, "\r\n"); i >= 0 {
		line = r.rest[:i]
		if strings.HasPrefix(r.rest[i:], "\r\n") {
			r.rest = r.rest[i+2:]
		} else {
			r.rest = r.rest[i+1:]
		}
	} else {
		r.rest = ""
	}
	line = strings.TrimPrefix(line, "\uFEFF")
	line = strings.ReplaceAll(line, "\u2013", "-")
	return text.NewStringLocated(line, r.location), true
}

// listReader returns fixed lines, all at the same location.
type listReader struct {
	lines    []string
	location *text.LineLocation
}

func (r *listReader) ReadLine() (text.StringLocated, bool) {
	if len(r.lines) == 0 {
		return text.StringLocated{}, false
	}
	line := r.lines[0]
	r.lines = r.lines[1:]
	return text.NewStringLocated(line, r.location), true
}

// uncommentReader removes the text that precedes @startuml on the start line
// (for example "* " in a Java comment) from the following lines.
type uncommentReader struct {
	raw            ReadLine
	headerToRemove *string
	paused         bool
}

func (r *uncommentReader) ReadLine() (text.StringLocated, bool) {
	result, ok := r.raw.ReadLine()
	if !ok {
		return result, false
	}
	if prefix, found := BeforeStartUml(result.S); found {
		r.headerToRemove = &prefix
	}
	if r.paused {
		if m := pausePattern.FindStringSubmatch(result.S); m != nil {
			h := m[1]
			r.headerToRemove = &h
		}
	}
	if h := r.headerToRemove; h != nil {
		if strings.HasPrefix(*h, result.S) {
			return text.NewStringLocated("", result.Location), true
		}
		if strings.HasPrefix(result.S, *h) {
			return result.Substring(len(*h), len(result.S)), true
		}
	}
	return result, true
}

// addConfigReader inserts configuration lines after every start directive.
type addConfigReader struct {
	raw      ReadLine
	config   []string
	inserted ReadLine
}

func (r *addConfigReader) ReadLine() (text.StringLocated, bool) {
	if r.inserted != nil {
		if line, ok := r.inserted.ReadLine(); ok {
			return line, true
		}
		r.inserted = nil
	}
	result, ok := r.raw.ReadLine()
	if ok && IsStartDirective(result.S) && len(r.config) > 0 {
		r.inserted = &quoteCommentReader{source: &listReader{lines: r.config, location: result.Location}}
	}
	return result, ok
}

// mergeLinesReader joins lines that end with a backslash with the following
// line, skipping comment lines in between. Ditaa blocks are left alone.
type mergeLinesReader struct {
	source              ReadLine
	keepEndingBackslash bool
}

func (r *mergeLinesReader) ReadLine() (text.StringLocated, bool) {
	result, ok := r.source.ReadLine()
	if ok && IsStartDirective(result.S) && isDitaaStart(result.S) {
		r.keepEndingBackslash = true
	}
	if ok && IsEndDirective(result.S) {
		r.keepEndingBackslash = false
	}
	var withoutComment ReadLine
	for ok && !r.keepEndingBackslash && text.EndsWithBackslash(result.S) {
		if withoutComment == nil {
			withoutComment = &quoteCommentReader{source: r.source}
		}
		next, more := withoutComment.ReadLine()
		if !more {
			break
		}
		result = result.MergeEndBackslash(next)
	}
	return result, ok
}

// isDitaaStart reports whether the start line opens a ditaa diagram. The MIT
// variant does not recognize ditaa, so this is always false, as upstream.
func isDitaaStart(s string) bool {
	return slices.Contains(FindStartTypes(text.Trin(s)), Ditaa)
}

// quoteCommentReader skips comment lines: lines starting with ' and
// /' ... '/ blocks, possibly spanning several lines.
type quoteCommentReader struct {
	source ReadLine
}

func (r *quoteCommentReader) ReadLine() (text.StringLocated, bool) {
	longComment := false
	for {
		result, ok := r.source.ReadLine()
		if !ok {
			return result, false
		}
		trim := text.Trin(strings.ReplaceAll(result.S, "\t", " "))
		switch {
		case longComment && strings.HasSuffix(trim, "'/"):
			longComment = false
			continue
		case longComment:
			continue
		case strings.HasPrefix(trim, "'"):
			continue
		case strings.HasPrefix(trim, "/'") && strings.HasSuffix(trim, "'/"):
			continue
		case strings.HasPrefix(trim, "/'") && !strings.Contains(trim, "'/"):
			longComment = true
			continue
		}
		return result.RemoveInnerComment(), true
	}
}

// ReadUncommented splits src into lines like the block reader does, removing
// the text before @startuml (for example "* " in a code comment) from the
// following lines.
func ReadUncommented(src, desc string) []text.StringLocated {
	r := &uncommentReader{raw: newLineReader(src, desc, nil)}
	var out []text.StringLocated
	for {
		l, ok := r.ReadLine()
		if !ok {
			return out
		}
		out = append(out, l)
	}
}

// NewLineReader returns the lines of src like upstream's
// ReadLineReader.create: the locations are described by desc and have the
// given parent (the location of an !include directive, or nil).
func NewLineReader(src, desc string, parent *text.LineLocation) ReadLine {
	return newLineReader(src, desc, parent)
}

// NewUncommentReader wraps r like upstream's UncommentReadLine.
func NewUncommentReader(r ReadLine) ReadLine { return &uncommentReader{raw: r} }

// NewMergeLinesReader wraps r like upstream's ReadFilterMergeLines: lines
// ending with a backslash are joined with the next line.
func NewMergeLinesReader(r ReadLine) ReadLine { return &mergeLinesReader{source: r} }
