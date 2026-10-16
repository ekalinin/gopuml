// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

// VegaPrefix is the id prefix of cases taken from upstream's Vega suite.
const VegaPrefix = "vega/"

// vegaFormats maps Vega "output" values to the formats this project supports.
var vegaFormats = map[string]string{
	"svg":     "svg",
	"atxt":    "txt",
	"txt":     "txt",
	"utxt":    "utxt",
	"preproc": "preproc",
}

// LoadVegaCases reads the .puml inputs of upstream's Vega suite below root.
// The YAML front matter is removed from Source. Formats are the supported
// formats from the "output" header, or svg when none of them is supported.
// The diagram type is unknown until the reference is rendered.
func LoadVegaCases(fsys fs.FS, root string) ([]*Case, error) {
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
		c, err := parseVegaCase(rel, data)
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

func parseVegaCase(rel string, data []byte) (*Case, error) {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	source := data
	var output string
	if first, rest, ok := bytes.Cut(data, []byte("\n")); ok && bytes.Equal(bytes.TrimSpace(first), headerDelimiter) {
		h, body, ok := cutHeader(rest)
		if !ok {
			return nil, fmt.Errorf("vega %s: unterminated header", rel)
		}
		// Vega headers are not strict YAML (values contain unquoted ": "),
		// and only the top-level output line matters here.
		for line := range strings.SplitSeq(string(h), "\n") {
			if v, ok := strings.CutPrefix(line, "output:"); ok {
				output = v
			}
		}
		source = body
	}

	var formats []string
	for out := range strings.SplitSeq(output, ",") {
		if f, ok := vegaFormats[strings.TrimSpace(out)]; ok && !slices.Contains(formats, f) {
			formats = append(formats, f)
		}
	}
	if len(formats) == 0 {
		formats = []string{"svg"}
	}

	return &Case{
		ID:      VegaPrefix + strings.TrimSuffix(rel, ".puml"),
		Family:  startFamily(source),
		Formats: formats,
		Source:  string(source),
		Path:    rel,
	}, nil
}

// startFamily returns the family of the first @start or \start line, or ""
// when the source has none.
func startFamily(source []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(source))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		rest, ok := strings.CutPrefix(line, "@start")
		if !ok {
			rest, ok = strings.CutPrefix(line, `\start`)
		}
		if !ok {
			continue
		}
		end := strings.IndexFunc(rest, func(r rune) bool {
			return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z')
		})
		if end >= 0 {
			rest = rest[:end]
		}
		return strings.ToLower(rest)
	}
	return ""
}
