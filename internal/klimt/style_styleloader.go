// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/style/StyleLoader.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/BlocLines.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"embed"
	"errors"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// skinFiles are upstream's /skin resources: plantuml.skin (the default
// style sheet), rose.skin, debug.skin, sonyxperiadev.skin and
// strictuml.skin.
//
//go:embed skin/*.skin
var skinFiles embed.FS

// skinEntry is a parsed embedded skin, parsed once and read-only after.
type skinEntry struct {
	once    sync.Once
	builder *StyleBuilder
	err     error
}

// embeddedSkins holds one entry per embedded skin; the map is built at
// initialization and never modified, upstream's StyleLoader cache.
var embeddedSkins = func() map[string]*skinEntry {
	m := map[string]*skinEntry{}
	entries, err := fs.ReadDir(skinFiles, "skin")
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		m[e.Name()] = &skinEntry{}
	}
	return m
}()

// loadBlocLines splits a file into lines as Java's BufferedReader.readLine
// does ("\n", "\r" or "\r\n"), each with the same location.
func loadBlocLines(data string, location *text.LineLocation) command.BlocLines {
	var lines []text.StringLocated
	for len(data) > 0 {
		i := strings.IndexAny(data, "\r\n")
		if i < 0 {
			lines = append(lines, text.NewStringLocated(data, location))
			break
		}
		lines = append(lines, text.NewStringLocated(data[:i], location))
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			i++
		}
		data = data[i+1:]
	}
	return command.NewBlocLines(lines)
}

// embeddedSkinLines returns the lines of an embedded skin.
func embeddedSkinLines(filename string) (command.BlocLines, error) {
	data, err := skinFiles.ReadFile("skin/" + filename)
	if err != nil {
		return command.BlocLines{}, err
	}
	return loadBlocLines(string(data), nil), nil
}

// StyleLoaderLoadSkin returns a new style builder loaded from a skin file:
// from the render context's file system first (the file relative to the
// current directory, as upstream looks for a local file), then from the
// embedded skins. An embedded skin is parsed once and cloned for each call.
// A missing skin is a *NoStyleAvailableError, a skin that does not parse or
// defines no style a *StyleParsingError.
func StyleLoaderLoadSkin(ctx *core.RenderContext, filename string) (*StyleBuilder, error) {
	if ctx != nil && ctx.FS != nil {
		name := path.Join(ctx.CurrentDir, filename)
		if !fs.ValidPath(name) {
			name = ""
		}
		if name != "" {
			if data, err := fs.ReadFile(ctx.FS, name); err == nil {
				return loadSkinSlow(filename, string(data))
			} else if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		}
	}
	entry, ok := embeddedSkins[filename]
	if !ok {
		return nil, &NoStyleAvailableError{}
	}
	entry.once.Do(func() {
		data, err := skinFiles.ReadFile("skin/" + filename)
		if err != nil {
			entry.err = err
			return
		}
		entry.builder, entry.err = loadSkinSlow(filename, string(data))
	})
	if entry.err != nil {
		return nil, entry.err
	}
	return entry.builder.CloneMe(), nil
}

func loadSkinSlow(filename, data string) (*StyleBuilder, error) {
	styleBuilder := NewStyleBuilder()
	lines2 := loadBlocLines(data, text.NewLineLocation(filename, nil))
	styles, err := StyleLoaderParseStyleText(lines2, styleBuilder)
	if err != nil {
		return nil, err
	}
	// A file that parses without any error but that defines no style at all
	// is not a style sheet: this happens with legacy skinparam files, where
	// every line is silently ignored because the key is unknown.
	if len(styles) == 0 {
		return nil, &StyleParsingError{Message: "No style found in " + filename}
	}
	for _, newStyle := range styles {
		styleBuilder.LoadInternal(newStyle.Signature(), newStyle)
	}
	return styleBuilder, nil
}

// StyleLoaderParseStyleText parses the text of a style sheet into styles,
// numbering their values with counter (usually the style builder they are
// for, so that they win over its styles): the blocks are parsed, merged by
// selector (a dark value with its regular one), then flattened into one
// style per selector.
func StyleLoaderParseStyleText(lines command.BlocLines, counter AutomaticCounter) ([]*Style, error) {
	if lines.Size() == 0 {
		return nil, nil
	}
	rules, err := parseRawStyleSheet(lines)
	if err != nil {
		return nil, err
	}
	root := newMergedStyleNode(false)
	root.mergeAll(rules, counter)
	return flattenStyles(root), nil
}

// addStereotypeCount is upstream's StyleLoader.addStereotypeCount: the
// values of a selector with count stereotypes.
func addStereotypeCount(tmp styleValues, count int) styleValues {
	var result styleValues
	for k, v := range tmp {
		if v != nil {
			result[k] = asValueImpl(v).WithStereotypeCount(count)
		}
	}
	return result
}

// mandatoryRootProperties are the properties that any complete style sheet
// defines on its root style (a fragment such as strictuml.skin does not).
var mandatoryRootProperties = []PName{PNameFontName, PNameFontSize, PNameFontStyle, PNameFontColor,
	PNameLineColor, PNameLineThickness, PNameBackGroundColor, PNameHorizontalAlignment}

// StyleLoaderMissingRootProperties returns the mandatory root properties
// that a style sheet does not define; none means it is complete.
func StyleLoaderMissingRootProperties(styleBuilder *StyleBuilder) []PName {
	var root *Style
	if styleBuilder != nil {
		root = styleBuilder.MergedStyle(StyleSignatureBasicOf(SNameRoot))
	}
	var result []PName
	for _, property := range mandatoryRootProperties {
		if root == nil || !root.HasValue(property) {
			result = append(result, property)
		}
	}
	return result
}
