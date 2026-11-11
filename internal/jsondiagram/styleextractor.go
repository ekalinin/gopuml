// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/jsondiagram/StyleExtractor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package jsondiagram

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// StyleExtractor takes the style blocks, and the commands that come before
// the data, out of the lines of a JSON, YAML or HCL diagram.
type StyleExtractor struct {
	list  []string
	style []text.StringLocated
	// title, scale and newSkin are empty when the source has none.
	title       string
	handwritten bool
	scale       string
	newSkin     string
}

// NewStyleExtractor reads the lines of a block, the start and end lines
// included. Empty lines are dropped. The <style> blocks are kept apart
// wherever they are. Before the first data line, "!assume", "!pragma" and
// "hide" lines are dropped, and "scale", "title", "skin" and "skinparam"
// lines are kept apart, a "skinparam" block with its lines.
func NewStyleExtractor(data []text.StringLocated) *StyleExtractor {
	e := &StyleExtractor{}
	for i := 0; i < len(data); i++ {
		line := data[i]
		s := text.JavaTrim(line.S)
		if s == "" {
			continue
		}
		beforeData := len(e.list) <= 1
		switch {
		case s == "<style>":
			// Neither the opening "<style>" nor the closing "</style>" joins the
			// style lines, so that the lines of several blocks follow each
			// other, as upstream collects them.
			for i+1 < len(data) {
				i++
				line = data[i]
				if text.JavaTrim(line.S) == "</style>" {
					break
				}
				e.style = append(e.style, line)
			}
		case beforeData && (strings.HasPrefix(s, "!assume ") || strings.HasPrefix(s, "!pragma ") ||
			strings.HasPrefix(s, "hide ")):
		case beforeData && strings.HasPrefix(s, "scale "):
			e.scale = s
		case beforeData && strings.HasPrefix(s, "title "):
			e.title = text.JavaTrim(s[len("title "):])
		case beforeData && strings.HasPrefix(s, "skin "):
			e.newSkin = text.JavaTrim(s[len("skin "):])
		case beforeData && strings.HasPrefix(s, "skinparam "):
			if strings.Contains(s, "handwritten") && strings.Contains(s, "true") {
				e.handwritten = true
			}
			if strings.Contains(s, "{") {
				for i+1 < len(data) && text.JavaTrim(line.S) != "}" {
					i++
					line = data[i]
				}
			}
		default:
			e.list = append(e.list, line.S)
		}
	}
	return e
}

// applyStyles loads the skin, then merges the style blocks into the style
// sheet of d. When the skin cannot be loaded or the style blocks cannot be
// parsed, it stops there, as upstream throws a StyleParsingException that
// its factories log.
func (e *StyleExtractor) applyStyles(d *uml.Base) {
	if e.newSkin != "" && !d.LoadSkin(e.newSkin).IsOK() {
		return
	}
	if len(e.style) > 0 {
		styles, err := klimt.StyleLoaderParseStyleText(command.NewBlocLines(e.style), d.CurrentStyleBuilder())
		if err != nil {
			return
		}
		d.Skin().MuteStyle(styles)
	}
}

// Lines returns the lines that are not taken out: the start line, the data
// lines and the end line.
func (e *StyleExtractor) Lines() []string { return e.list }

// Title returns the text of the title line, empty for none.
func (e *StyleExtractor) Title() string { return e.title }

// IsHandwritten reports whether a skinparam line before the data asks for a
// handwritten drawing, which this release does not draw.
func (e *StyleExtractor) IsHandwritten() bool { return e.handwritten }

// Scale returns the scale line, empty for none.
func (e *StyleExtractor) Scale() string { return e.scale }
