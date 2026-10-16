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

// Formats that a curated case can request.
var knownFormats = []string{"svg", "png", "txt", "utxt", "preproc"}

// Case is one curated conformance case: a PlantUML source preceded by a YAML
// front matter header, stored as cases/<capability>/<name>.puml.
//
//	---
//	id: sequence-diagram/self-message
//	family: uml
//	type: sequence
//	formats: [svg, txt]
//	scenarios:
//	  - "sequence-diagram: Self message"
//	---
//	@startuml
//	Alice -> Alice : think
//	@enduml
type Case struct {
	// ID is "<capability>/<name>" and must match the file location.
	ID string `yaml:"id"`
	// Family is the start tag family, for example "uml" or "mindmap".
	Family string `yaml:"family"`
	// Type is the diagram type within the family, for example "sequence".
	Type string `yaml:"type"`
	// Formats lists the output formats to compare.
	Formats []string `yaml:"formats"`
	// Scenarios references the spec scenarios the case covers, written as
	// "<capability>: <scenario name>".
	Scenarios []string `yaml:"scenarios"`

	// Source is the PlantUML text after the header.
	Source string `yaml:"-"`
	// Path is the slash-separated path of the case file relative to the
	// cases root.
	Path string `yaml:"-"`
}

// Capability returns the capability directory of the case.
func (c *Case) Capability() string {
	capability, _, _ := strings.Cut(c.ID, "/")
	return capability
}

var headerDelimiter = []byte("---")

// ParseCase parses one case file. relPath is the file path relative to the
// cases root, using forward slashes.
func ParseCase(relPath string, data []byte) (*Case, error) {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	first, rest, ok := bytes.Cut(data, []byte("\n"))
	if !ok || !bytes.Equal(bytes.TrimSpace(first), headerDelimiter) {
		return nil, fmt.Errorf("%s: missing YAML header: the first line must be ---", relPath)
	}
	header, body, ok := cutHeader(rest)
	if !ok {
		return nil, fmt.Errorf("%s: unterminated YAML header: no closing --- line", relPath)
	}

	var c Case
	dec := yaml.NewDecoder(bytes.NewReader(header))
	dec.KnownFields(true)
	// An empty header decodes to io.EOF; validate reports the missing fields.
	if err := dec.Decode(&c); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%s: invalid YAML header: %w", relPath, err)
	}
	c.Source = string(body)
	c.Path = relPath
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", relPath, err)
	}
	return &c, nil
}

// cutHeader splits the text after the opening delimiter into the header and
// the body that follows the closing delimiter line.
func cutHeader(rest []byte) (header, body []byte, ok bool) {
	offset := 0
	for offset <= len(rest) {
		line, _, _ := bytes.Cut(rest[offset:], []byte("\n"))
		if bytes.Equal(bytes.TrimSpace(line), headerDelimiter) {
			end := min(offset+len(line)+1, len(rest))
			return rest[:offset], rest[end:], true
		}
		next := bytes.IndexByte(rest[offset:], '\n')
		if next < 0 {
			break
		}
		offset += next + 1
	}
	return nil, nil, false
}

func (c *Case) validate() error {
	wantID := strings.TrimSuffix(c.Path, path.Ext(c.Path))
	switch {
	case c.ID == "":
		return errors.New("missing id")
	case c.ID != wantID:
		return fmt.Errorf("id %q does not match the file location %q", c.ID, wantID)
	case c.Family == "":
		return errors.New("missing family")
	case c.Type == "":
		return errors.New("missing type")
	case len(c.Formats) == 0:
		return errors.New("missing formats")
	case strings.TrimSpace(c.Source) == "":
		return errors.New("empty diagram source")
	}
	for _, f := range c.Formats {
		if !slices.Contains(knownFormats, f) {
			return fmt.Errorf("unknown format %q (known: %s)", f, strings.Join(knownFormats, ", "))
		}
	}
	for _, s := range c.Scenarios {
		capability, name, ok := strings.Cut(s, ": ")
		if !ok || capability == "" || strings.TrimSpace(name) == "" {
			return fmt.Errorf("scenario %q is not in the form \"<capability>: <scenario name>\"", s)
		}
	}
	return nil
}

// LoadCases reads every .puml case below root in fsys, sorted by id. Ids are
// unique because each id must match its file location.
func LoadCases(fsys fs.FS, root string) ([]*Case, error) {
	var cases []*Case
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Ext(p) != ".puml" {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, root), "/")
		c, err := ParseCase(rel, data)
		if err != nil {
			return err
		}
		cases = append(cases, c)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(cases, func(a, b *Case) int { return strings.Compare(a.ID, b.ID) })
	return cases, nil
}
