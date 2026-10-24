// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/cli/CliOptions.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/text"
)

// ConfigError is upstream's CliParsingException for a file of -I or
// --config that cannot be read; the command line fails with exit status 42.
type ConfigError struct{ Message string }

func (e *ConfigError) Error() string { return e.Message }

// readConfigFile is CliOptions.addFileInConfig: the lines of a file, read
// as UTF-8.
func readConfigFile(name string) ([]string, error) {
	abs, _ := filepath.Abs(name)
	st, err := os.Stat(name)
	if err != nil {
		return nil, &ConfigError{"Cannot read " + abs}
	}
	if st.IsDir() {
		return nil, &ConfigError{"Cannot have a directory here " + abs}
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, &ConfigError{"Cannot read " + abs}
	}
	var lines []string
	in := &javaLineReader{r: bufio.NewReader(bytes.NewReader(data))}
	for line, ok := in.readLine(); ok; line, ok = in.readLine() {
		lines = append(lines, text.DecodeUTF8([]byte(line)))
	}
	return lines, nil
}

// javaString is Java's string conversion of a possibly null map value.
func javaString(kv KeyValue) string {
	if !kv.HasValue {
		return "null"
	}
	return kv.Value
}

// buildConfig is the CliOptions constructor: the lines inserted at the
// start of each diagram, from PLANTUML_DEFAULT_CONFIG_FILENAME, -I,
// --theme, --config, -P and -S, in this order.
func buildConfig(p *Parsed, getenv func(string) string) ([]string, error) {
	var config []string
	addFile := func(name string) error {
		lines, err := readConfigFile(name)
		config = append(config, lines...)
		return err
	}
	// initInclude: a name with * is a pattern of files.
	initInclude := func(name string) error {
		if name == "" {
			return nil
		}
		if !strings.Contains(name, "*") {
			return addFile(name)
		}
		for _, f := range fileGroup(name, nil) {
			if err := addFile(f); err != nil {
				return err
			}
		}
		return nil
	}
	if err := initInclude(security.Getenv(getenv, "PLANTUML_DEFAULT_CONFIG_FILENAME")); err != nil {
		return nil, err
	}
	for _, kv := range p.Map(Include, IncludeLong) {
		if err := initInclude(kv.Key); err != nil {
			return nil, err
		}
	}
	for _, v := range p.List(Theme) {
		config = append(config, "!theme "+v.S)
	}
	for _, v := range p.List(Config) {
		if err := addFile(v.S); err != nil {
			return nil, err
		}
	}
	for _, kv := range p.Map(Pragma, PragmaLong) {
		config = append(config, "!pragma "+kv.Key+" "+javaString(kv))
	}
	for _, kv := range p.Map(Skinparam, SkinparamLong) {
		config = append(config, "skinparamlocked "+kv.Key+" "+javaString(kv))
	}
	return config, nil
}
