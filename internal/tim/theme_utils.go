// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/theme/ThemeUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/theme/Theme.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// ThemeSource provides the themes bundled with the program, which !theme
// finds before local theme files. A ThemeSource that also implements
// ThemeLister gives %get_all_theme its list.
type ThemeSource interface {
	// Theme returns the content of the bundled file puml-theme-<name>.puml.
	Theme(name string) (content []byte, ok bool)
}

// theme is the source of a theme, without its YAML header.
type theme struct {
	source *readLineWithYamlHeader
}

func newTheme(source core.ReadLine) *theme { return &theme{source: newReadLineWithYamlHeader(source)} }

func (t *theme) readLine() (text.StringLocated, bool) { return t.source.ReadLine() }

func (t *theme) getMetadata() *json.Object {
	result := json.NewObject()
	for _, ent := range t.source.metadata {
		result.Add(ent.key, json.ValueOfString(ent.value))
	}
	return result
}

const (
	themeFilePrefix = "puml-theme-"
	themeFileSuffix = ".puml"
	themePath       = "themes"
)

func getThemeFilename(filename string) string { return themeFilePrefix + filename + themeFileSuffix }

// loadTheme is ThemeUtils.loadTheme: nil when the theme does not exist.
func (c *TContext) loadTheme(pathSystem *PathSystem, name string, from *string, location text.StringLocated) (*theme, error) {
	if from == nil {
		return c.loadBundledOrLocalTheme(pathSystem, name)
	}
	f := *from
	switch {
	case strings.HasPrefix(f, "<") && strings.HasSuffix(f, ">"):
		return c.loadStdlibTheme(name, f), nil
	case strings.HasPrefix(f, "http://") || strings.HasPrefix(f, "https://"):
		return c.loadHTTPTheme(name, f, location)
	}
	return c.loadFileTheme(pathSystem, name, f)
}

func (c *TContext) loadBundledOrLocalTheme(pathSystem *PathSystem, name string) (*theme, error) {
	// First, try bundled themes
	res := "/" + themePath + "/" + getThemeFilename(name)
	if c.themes != nil {
		if content, ok := c.themes.Theme(name); ok {
			return newTheme(core.NewLineReader(string(content), "<"+res+">", nil)), nil
		}
	}
	// Then try local file
	localFile, err := pathSystem.getInputFile(getThemeFilename(name))
	if err != nil {
		return nil, err
	}
	if localFile != nil {
		data, err := localFile.newInputStream()
		if err != nil {
			return nil, err
		}
		if data == nil {
			throwRuntime("NullPointerException", "")
		}
		c.addIncluded(localFile.String())
		return newTheme(core.NewLineReader(string(data), "theme "+name, nil)), nil
	}
	return nil, nil
}

func (c *TContext) loadStdlibTheme(name, from string) *theme {
	realFrom := from[1 : len(from)-1]
	content := c.stdlib.getPumlResource(realFrom + "/" + getThemeFilename(name))
	if content == nil {
		return nil
	}
	return newTheme(core.NewLineReader(string(content), name+" from "+from, nil))
}

func (c *TContext) loadHTTPTheme(name, from string, location text.StringLocated) (*theme, error) {
	u, ok := createURL(getThemeFullPath(from, name))
	if !ok {
		return nil, newEaterException("Cannot open URL", location)
	}
	reader, err := c.getReaderInclude(u, location)
	if err != nil {
		return nil, err
	}
	return newTheme(reader), nil
}

func getThemeFullPath(from, filename string) string {
	var sb strings.Builder
	sb.WriteString(from)
	if !strings.HasSuffix(from, "/") {
		sb.WriteString("/")
	}
	return sb.String() + getThemeFilename(filename)
}

func (c *TContext) loadFileTheme(pathSystem *PathSystem, name, from string) (*theme, error) {
	file, err := pathSystem.getInputFile(getThemeFullPath(from, name))
	if err != nil {
		return nil, err
	}
	if file == nil {
		throwRuntime("NullPointerException", "")
	}
	data, err := file.newInputStream()
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	c.addIncluded(file.String())
	return newTheme(core.NewLineReader(string(data), name+" from "+from, nil)), nil
}
