// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	htmlTag     = regexp.MustCompile(`<[^>]*>`)
	htmlRow     = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	htmlCell    = regexp.MustCompile(`(?s)<t[hd][^>]*>(.*?)</t[hd]>`)
	htmlItem    = regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`)
	mdSeparator = regexp.MustCompile(`^\|[-:| ]+\|$`)
	// releaseTag finds the releases that installation commands name.
	releaseTag = regexp.MustCompile(`(?:plantuml@|release download )(v\d+\.\d+\.\d+)\b`)
)

// htmlText is the text of an HTML fragment, with its tags removed, its
// entities decoded and its spaces collapsed.
func htmlText(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(htmlTag.ReplaceAllString(s, ""))), " ")
}

// mdText is the text of a Markdown fragment, without its code spans' marks
// and with its spaces collapsed.
func mdText(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "`", "")), " ")
}

// htmlSection returns the content of the section or footer of the page
// whose id is id.
func htmlSection(page, id string) (string, error) {
	m := regexp.MustCompile(`(?s)<(section|footer)[^>]*\bid="` + regexp.QuoteMeta(id) + `"[^>]*>(.*?)</(?:section|footer)>`).FindStringSubmatch(page)
	if m == nil {
		return "", fmt.Errorf("no section %q", id)
	}
	return m[2], nil
}

// htmlRows returns the text of the cells of the rows of the tables in s.
func htmlRows(s string) [][]string {
	var rows [][]string
	for _, r := range htmlRow.FindAllStringSubmatch(s, -1) {
		var cells []string
		for _, c := range htmlCell.FindAllStringSubmatch(r[1], -1) {
			cells = append(cells, htmlText(c[1]))
		}
		rows = append(rows, cells)
	}
	return rows
}

// htmlItems returns the text of the list items in s.
func htmlItems(s string) []string {
	var items []string
	for _, m := range htmlItem.FindAllStringSubmatch(s, -1) {
		items = append(items, htmlText(m[1]))
	}
	return items
}

// mdSection returns the text of the Markdown section with the given
// heading, up to the next heading of the same level.
func mdSection(doc, heading string) (string, error) {
	_, rest, ok := strings.Cut(doc, "\n"+heading+"\n")
	if !ok {
		return "", fmt.Errorf("no section %q", heading)
	}
	level := heading[:strings.IndexByte(heading, ' ')+1]
	section, _, _ := strings.Cut(rest, "\n"+level)
	return section, nil
}

// mdRows returns the text of the cells of the rows of the Markdown tables in
// s, without their separator rows.
func mdRows(s string) [][]string {
	var rows [][]string
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || mdSeparator.MatchString(line) {
			continue
		}
		var cells []string
		for c := range strings.SplitSeq(strings.Trim(line, "|"), "|") {
			cells = append(cells, mdText(c))
		}
		rows = append(rows, cells)
	}
	return rows
}

// mdItems returns the text of the top-level list items in s, with their
// continuation lines.
func mdItems(s string) []string {
	var items []string
	for line := range strings.SplitSeq(s, "\n") {
		switch {
		case strings.HasPrefix(line, "- "):
			items = append(items, mdText(line[2:]))
		case strings.HasPrefix(line, "  ") && len(items) > 0 && strings.TrimSpace(line) != "":
			items[len(items)-1] = mdText(items[len(items)-1] + " " + line)
		}
	}
	return items
}

// checkRows checks that every row of want is a row of got.
func checkRows(what string, want, got [][]string) error {
	var errs []error
	if len(want) == 0 {
		errs = append(errs, fmt.Errorf("no row in %s", what))
	}
	for _, row := range want {
		if !slices.ContainsFunc(got, func(r []string) bool { return slices.Equal(r, row) }) {
			errs = append(errs, fmt.Errorf("%s: no row %q on the page", what, row))
		}
	}
	return errors.Join(errs...)
}

// pageRelease returns the release named by the gopuml-release meta tag of
// the page.
func pageRelease(page string) (string, error) {
	m := regexp.MustCompile(`<meta name="gopuml-release" content="([^"]*)">`).FindAllStringSubmatch(page, -1)
	if len(m) != 1 {
		return "", fmt.Errorf("%d gopuml-release meta tags, want 1", len(m))
	}
	if !regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(m[0][1]) {
		return "", fmt.Errorf("gopuml-release %q is not a release tag", m[0][1])
	}
	return m[0][1], nil
}

// checkLandingPage checks the content of the landing page (landing-page:
// Published landing page, Project presentation) against the README, whose
// table of supported diagrams it repeats, and the notes of its release,
// whose conformance results it repeats.
func checkLandingPage(page, readme, notes string) error {
	var errs []error
	tag, err := pageRelease(page)
	if err != nil {
		return err
	}

	// The name of the project, what it is, and its links.
	for _, want := range []string{
		`<title>gopuml - `,
		`<a class="wordmark" href="#main">gopuml</a>`,
		`href="https://github.com/ekalinin/gopuml"`,
		`href="https://github.com/ekalinin/gopuml/releases"`,
		`href="https://pkg.go.dev/github.com/ekalinin/gopuml"`,
	} {
		if !strings.Contains(page, want) {
			errs = append(errs, fmt.Errorf("no %q", want))
		}
	}
	sections := map[string]string{}
	for _, id := range []string{"ways", "playground", "diagrams", "install", "comparison", "conformance", "license"} {
		s, err := htmlSection(page, id)
		if err != nil {
			errs = append(errs, err)
		}
		sections[id] = s
	}
	if !strings.Contains(htmlText(page), "gopuml is a Go reimplementation of PlantUML") {
		errs = append(errs, errors.New("the page does not say what gopuml is"))
	}

	// Without JavaScript, everything but the playground is in the static
	// HTML and shown: one script, for the playground, and no hidden element
	// outside the playground, which says that it needs JavaScript.
	if scripts := strings.Count(page, "<script"); scripts != 1 || !strings.Contains(page, `<script type="module" src="page.mjs"></script>`) {
		errs = append(errs, fmt.Errorf("%d scripts, want only page.mjs", scripts))
	}
	if strings.Count(page, " hidden") != strings.Count(sections["playground"], " hidden") {
		errs = append(errs, errors.New("hidden elements outside the playground"))
	}
	if !strings.Contains(sections["playground"], "<noscript>") {
		errs = append(errs, errors.New("no <noscript> message in the playground"))
	}

	// The four forms.
	ways := htmlText(sections["ways"])
	for _, want := range []string{"java -jar plantuml.jar", "Go library", "HTTP server", "WebAssembly module"} {
		if !strings.Contains(ways, want) {
			errs = append(errs, fmt.Errorf("the forms do not name %q", want))
		}
	}

	// The supported diagrams, as the README lists them.
	table, err := mdSection(readme, "## Supported diagrams and formats")
	if err != nil {
		errs = append(errs, fmt.Errorf("README: %w", err))
	}
	table, _, _ = strings.Cut(table, "\n\n")
	if err := checkRows("the supported diagrams of the README", mdRows(table), htmlRows(sections["diagrams"])); err != nil {
		errs = append(errs, err)
	}

	// The installation of the release.
	install := htmlText(sections["install"])
	for _, want := range []string{
		"gh release download " + tag + " ",
		"-p checksums.txt",
		"shasum -a 256 --ignore-missing -c checksums.txt",
		"gopuml_" + tag[1:] + "_",
		"go install github.com/ekalinin/gopuml/cmd/plantuml@" + tag,
	} {
		if !strings.Contains(install, want) {
			errs = append(errs, fmt.Errorf("the installation has no %q", want))
		}
	}

	// The conformance results, as the notes of the release list them.
	results, err := mdSection(notes, "## Conformance")
	if err != nil {
		errs = append(errs, fmt.Errorf("notes of %s: %w", tag, err))
	}
	items := mdItems(results)
	if len(items) == 0 {
		errs = append(errs, fmt.Errorf("notes of %s: no conformance results", tag))
	}
	for _, item := range items {
		if !slices.Contains(htmlItems(sections["conformance"]), item) {
			errs = append(errs, fmt.Errorf("no conformance result %q on the page", item))
		}
	}

	// The license.
	for _, want := range []string{
		`href="https://github.com/ekalinin/gopuml/blob/main/LICENSE"`,
		`href="https://github.com/ekalinin/gopuml/blob/main/NOTICE"`,
	} {
		if !strings.Contains(sections["license"], want) {
			errs = append(errs, fmt.Errorf("the license has no %s", want))
		}
	}
	return errors.Join(errs...)
}

// readLanding reads the landing page, the README and the notes of the
// release the page names.
func readLanding(t *testing.T) (page, readme, notes string) {
	t.Helper()
	for _, f := range []struct {
		dst  *string
		name string
	}{{&page, filepath.Join("site", "index.html")}, {&readme, "README.md"}} {
		data, err := os.ReadFile(f.name)
		if err != nil {
			t.Fatal(err)
		}
		*f.dst = string(data)
	}
	tag, err := pageRelease(page)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("docs", "releases", tag+".md"))
	if err != nil {
		t.Fatal(err)
	}
	return page, readme, string(data)
}

func TestLandingPage(t *testing.T) {
	page, readme, notes := readLanding(t)
	if err := checkLandingPage(page, readme, notes); err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ name, old, new string }{
		{"no repository link", `href="https://github.com/ekalinin/gopuml"`, `href="https://example.com"`},
		{"no section", `id="conformance"`, `id="results"`},
		{"content built by a script", `<script type="module" src="page.mjs"></script>`, `<script type="module" src="page.mjs"></script><script src="content.js"></script>`},
		{"hidden content", `<section id="install"`, `<section hidden id="install"`},
		{"no noscript message", `<noscript>`, `<div>`},
		{"missing diagram type", `<tr><th scope="row">State</th>`, `<tr><th scope="row">Statechart</th>`},
		{"other formats", `<tr><th scope="row">Sequence</th><td>yes</td><td>yes</td><td>yes</td></tr>`, `<tr><th scope="row">Sequence</th><td>yes</td><td>yes</td><td>no</td></tr>`},
		{"no go install", `go install github.com/ekalinin/gopuml/cmd/plantuml@`, `go get github.com/ekalinin/gopuml/cmd/plantuml@`},
		{"no checksums", `-p checksums.txt`, ``},
		{"other conformance result", `<strong>372</strong>`, `<strong>371</strong>`},
		{"no NOTICE link", `blob/main/NOTICE`, `blob/main/README.md`},
	} {
		mutated := strings.ReplaceAll(page, m.old, m.new)
		if mutated == page {
			t.Errorf("%s: %q is not in the page", m.name, m.old)
			continue
		}
		if checkLandingPage(mutated, readme, notes) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
}
