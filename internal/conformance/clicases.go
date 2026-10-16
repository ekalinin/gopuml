// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	cliCaseSuffix  = ".cli.yaml"
	httpCaseSuffix = ".http.yaml"
)

// What a CLI case can compare.
var knownCLICompare = []string{"exit", "files", "stdout", "stderr"}

// What an HTTP case can compare.
var knownHTTPCompare = []string{"status", "headers", "body"}

// CLICase runs the command line with arguments in a fresh working directory
// and records the exit code, outputs and created files. It is stored as
// cases/<capability>/<name>.cli.yaml.
type CLICase struct {
	ID        string   `yaml:"id"`
	Scenarios []string `yaml:"scenarios"`
	// Args are passed to the command, without the program name.
	Args []string `yaml:"args"`
	// Stdin is written to the standard input of the command.
	Stdin string `yaml:"stdin"`
	// Files are created in the working directory before the run, keyed by
	// slash-separated relative path.
	Files map[string]string `yaml:"files"`
	// Env sets environment variables for the run.
	Env map[string]string `yaml:"env"`
	// Compare lists what must match the reference. Defaults to exit and files.
	Compare []string `yaml:"compare"`

	Path string `yaml:"-"`
}

// HTTPCase sends one request to a running server and records the response.
// It is stored as cases/<capability>/<name>.http.yaml.
type HTTPCase struct {
	ID        string   `yaml:"id"`
	Scenarios []string `yaml:"scenarios"`
	// Server is the reference server: "builtin" (the --http-server of the
	// CLI) or "plantuml-server".
	Server string `yaml:"server"`
	Method string `yaml:"method"`
	// RequestPath may contain {encoded}, replaced by the encoded Source.
	RequestPath string            `yaml:"path"`
	Source      string            `yaml:"source"`
	Body        string            `yaml:"body"`
	Headers     map[string]string `yaml:"headers"`
	// Compare lists what must match the reference: status, headers (the
	// X-PlantUML-* headers and Content-Type) and body.
	Compare []string `yaml:"compare"`

	Path string `yaml:"-"`
}

func decodeStrict(relPath string, data []byte, v any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s: invalid YAML: %w", relPath, err)
	}
	return nil
}

func checkID(id, relPath, suffix string) error {
	want := strings.TrimSuffix(relPath, suffix)
	switch {
	case id == "":
		return errors.New("missing id")
	case id != want:
		return fmt.Errorf("id %q does not match the file location %q", id, want)
	}
	return nil
}

func checkScenarios(scenarios []string) error {
	for _, s := range scenarios {
		capability, name, ok := strings.Cut(s, ": ")
		if !ok || capability == "" || strings.TrimSpace(name) == "" {
			return fmt.Errorf("scenario %q is not in the form \"<capability>: <scenario name>\"", s)
		}
	}
	return nil
}

func checkCompare(values, known []string) error {
	for _, v := range values {
		if !slices.Contains(known, v) {
			return fmt.Errorf("unknown compare value %q (known: %s)", v, strings.Join(known, ", "))
		}
	}
	return nil
}

// ParseCLICase parses one .cli.yaml file.
func ParseCLICase(relPath string, data []byte) (*CLICase, error) {
	var c CLICase
	if err := decodeStrict(relPath, data, &c); err != nil {
		return nil, err
	}
	c.Path = relPath
	if len(c.Compare) == 0 {
		c.Compare = []string{"exit", "files"}
	}
	err := errors.Join(
		checkID(c.ID, relPath, cliCaseSuffix),
		checkScenarios(c.Scenarios),
		checkCompare(c.Compare, knownCLICompare),
	)
	for name := range c.Files {
		if path.IsAbs(name) || strings.HasPrefix(path.Clean(name), "..") {
			err = errors.Join(err, fmt.Errorf("file %q must be a relative path inside the working directory", name))
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", relPath, err)
	}
	return &c, nil
}

// ParseHTTPCase parses one .http.yaml file.
func ParseHTTPCase(relPath string, data []byte) (*HTTPCase, error) {
	var c HTTPCase
	if err := decodeStrict(relPath, data, &c); err != nil {
		return nil, err
	}
	c.Path = relPath
	if len(c.Compare) == 0 {
		c.Compare = []string{"status", "headers", "body"}
	}
	err := errors.Join(
		checkID(c.ID, relPath, httpCaseSuffix),
		checkScenarios(c.Scenarios),
		checkCompare(c.Compare, knownHTTPCompare),
	)
	if c.Server != "builtin" && c.Server != "plantuml-server" {
		err = errors.Join(err, fmt.Errorf("server must be builtin or plantuml-server, got %q", c.Server))
	}
	if c.Method == "" || !strings.HasPrefix(c.RequestPath, "/") {
		err = errors.Join(err, errors.New("method and an absolute path are required"))
	}
	if strings.Contains(c.RequestPath, "{encoded}") && c.Source == "" {
		err = errors.Join(err, errors.New("path uses {encoded} but source is empty"))
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", relPath, err)
	}
	return &c, nil
}

// loadBySuffix parses every file below root whose name ends with suffix.
func loadBySuffix[T any](fsys fs.FS, root, suffix string, parse func(string, []byte) (T, error)) ([]T, error) {
	var out []T
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, suffix) {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		c, err := parse(strings.TrimPrefix(strings.TrimPrefix(p, root), "/"), data)
		if err != nil {
			return err
		}
		out = append(out, c)
		return nil
	})
	return out, err
}

// LoadCLICases reads every .cli.yaml case below root, sorted by id.
func LoadCLICases(fsys fs.FS, root string) ([]*CLICase, error) {
	cases, err := loadBySuffix(fsys, root, cliCaseSuffix, ParseCLICase)
	slices.SortFunc(cases, func(a, b *CLICase) int { return strings.Compare(a.ID, b.ID) })
	return cases, err
}

// LoadHTTPCases reads every .http.yaml case below root, sorted by id.
func LoadHTTPCases(fsys fs.FS, root string) ([]*HTTPCase, error) {
	cases, err := loadBySuffix(fsys, root, httpCaseSuffix, ParseHTTPCase)
	slices.SortFunc(cases, func(a, b *HTTPCase) int { return strings.Compare(a.ID, b.ID) })
	return cases, err
}
