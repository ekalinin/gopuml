// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"errors"
	"fmt"
	"html"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
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
		"curl -fLO https://github.com/ekalinin/gopuml/releases/download/" + tag + "/gopuml_" + tag[1:] + "_",
		"curl -fLO https://github.com/ekalinin/gopuml/releases/download/" + tag + "/checksums.txt",
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
		{"no curl", `curl -fLO `, `wget `},
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

var (
	anyTag         = regexp.MustCompile(`\bv\d+\.\d+\.\d+\b`)
	archiveVersion = regexp.MustCompile(`gopuml_(\d+\.\d+\.\d+)_`)
	barTime        = regexp.MustCompile(`(?s)<td[^>]*style="--t: (\d+)"[^>]*>(.*?)</td>`)
	barMax         = regexp.MustCompile(`style="--max: (\d+)"`)
)

// checkLandingRelease checks that the page names one release, the release
// of the README, and that its comparison is the one of the notes of that
// release (landing-page: Consistency with the release, Comparison with
// upstream on the page). notes returns the notes of a release.
func checkLandingRelease(page, readme string, notes func(tag string) (string, error)) error {
	tag, err := pageRelease(page)
	if err != nil {
		return err
	}
	var errs []error
	for _, m := range anyTag.FindAllString(page, -1) {
		if m != tag {
			errs = append(errs, fmt.Errorf("the page names %s besides %s", m, tag))
		}
	}
	for _, m := range archiveVersion.FindAllStringSubmatch(page, -1) {
		if "v"+m[1] != tag {
			errs = append(errs, fmt.Errorf("the page names the archive %s of %s", m[0], tag))
		}
	}
	if releaseTag.FindString(page) == "" {
		errs = append(errs, errors.New("the installation of the page names no release"))
	}
	readmeTags := releaseTag.FindAllStringSubmatch(readme, -1)
	if len(readmeTags) == 0 {
		errs = append(errs, errors.New("the installation of the README names no release"))
	}
	for _, m := range readmeTags {
		if m[1] != tag {
			errs = append(errs, fmt.Errorf("the README names %s, the page %s", m[1], tag))
		}
	}

	n, err := notes(tag)
	if err != nil {
		return errors.Join(append(errs, fmt.Errorf("notes of %s: %w", tag, err))...)
	}
	comparison, err := mdSection(n, "## Comparison with upstream")
	if err != nil {
		return errors.Join(append(errs, fmt.Errorf("notes of %s: %w", tag, err))...)
	}
	section, err := htmlSection(page, "comparison")
	if err != nil {
		return errors.Join(append(errs, err)...)
	}
	if err := checkRows("the comparison of the notes of "+tag, mdRows(comparison), htmlRows(section)); err != nil {
		errs = append(errs, err)
	}
	env := mdItems(comparison)
	if len(env) == 0 {
		errs = append(errs, fmt.Errorf("notes of %s: no environment of the comparison", tag))
	}
	for _, item := range env {
		if !slices.Contains(htmlItems(section), item) {
			errs = append(errs, fmt.Errorf("no %q in the comparison of the page", item))
		}
	}

	// Each activation bar is as long as the time of its cell, relative to
	// the longest time of the table.
	bars := barTime.FindAllStringSubmatch(section, -1)
	if len(bars) == 0 {
		errs = append(errs, errors.New("no activation bar in the comparison"))
	}
	longest := 0
	for _, b := range bars {
		var ms int
		fmt.Sscan(b[1], &ms)
		longest = max(longest, ms)
		if text := htmlText(b[2]); !strings.HasPrefix(text, b[1]+" ms") {
			errs = append(errs, fmt.Errorf("the bar of %q is %s ms long", text, b[1]))
		}
	}
	if m := barMax.FindStringSubmatch(section); m == nil || m[1] != fmt.Sprint(longest) {
		errs = append(errs, fmt.Errorf("the scale of the bars is not the longest time, %d ms", longest))
	}
	return errors.Join(errs...)
}

func TestLandingRelease(t *testing.T) {
	page, readme, notes := readLanding(t)
	notesOf := func(n string) func(string) (string, error) {
		return func(tag string) (string, error) {
			if _, err := os.Stat(filepath.Join("docs", "releases", tag+".md")); err != nil {
				return "", err
			}
			return n, nil
		}
	}
	if err := checkLandingRelease(page, readme, notesOf(notes)); err != nil {
		t.Fatal(err)
	}
	tag, _ := pageRelease(page)
	for _, m := range []struct{ name, file, old, new string }{
		{"another release in the meta tag", "page", `content="` + tag + `"`, `content="v9.9.9"`},
		{"page behind the README", "readme", tag, "v9.9.9"},
		{"another archive", "page", "gopuml_" + tag[1:] + "_darwin", "gopuml_9.9.9_darwin"},
		{"changed cell", "page", "897 ms (17 ms per case)", "897 ms (18 ms per case)"},
		{"comparison out of date", "notes", "| 18 ms |", "| 17 ms |"},
		{"bar of another time", "page", `style="--t: 86"`, `style="--t: 68"`},
		{"other scale", "page", `style="--max: 959"`, `style="--max: 1000"`},
		{"no environment", "page", "<li>Runs: ", "<li>Run: "},
	} {
		p, r, n := page, readme, notes
		target := map[string]*string{"page": &p, "readme": &r, "notes": &n}[m.file]
		mutated := strings.ReplaceAll(*target, m.old, m.new)
		if mutated == *target {
			t.Errorf("%s: %q is not in the %s", m.name, m.old, m.file)
			continue
		}
		*target = mutated
		if checkLandingRelease(p, r, notesOf(n)) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
	noNotes := func(tag string) (string, error) { return "", os.ErrNotExist }
	if checkLandingRelease(page, readme, noNotes) == nil {
		t.Error("release without notes: the check passes")
	}
}

// pagesWorkflow is the part of the Pages workflow that checkPagesWorkflow
// reads.
type pagesWorkflow struct {
	On          map[string]any    `yaml:"on"`
	Permissions map[string]string `yaml:"permissions"`
	Concurrency map[string]any    `yaml:"concurrency"`
	Jobs        map[string]struct {
		If          string            `yaml:"if"`
		Environment map[string]string `yaml:"environment"`
		Steps       []pagesStep       `yaml:"steps"`
	} `yaml:"jobs"`
}

// pagesStep is a step of the Pages workflow, with its id.
type pagesStep struct {
	workflowStep `yaml:",inline"`
	ID           string `yaml:"id"`
}

// checkPagesWorkflow checks the contract of the Pages workflow
// (landing-page: Deployment). release is the name of the release workflow.
func checkPagesWorkflow(data []byte, release string) error {
	var w pagesWorkflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return err
	}
	var errs []error
	triggers := slices.Sorted(maps.Keys(w.On))
	if !slices.Equal(triggers, []string{"push", "workflow_dispatch", "workflow_run"}) {
		errs = append(errs, fmt.Errorf("the workflow runs on %v, not on push, workflow_run and workflow_dispatch", triggers))
	}
	push, _ := w.On["push"].(map[string]any)
	branches, _ := push["branches"].([]any)
	paths, _ := push["paths"].([]any)
	if !slices.Equal(branches, []any{"main"}) {
		errs = append(errs, fmt.Errorf("the workflow runs on pushes to %v, not to main", branches))
	}
	for _, p := range []string{"site/**", "tools/site/**", ".github/workflows/pages.yml", "LICENSE", "NOTICE"} {
		if !slices.Contains(paths, any(p)) {
			errs = append(errs, fmt.Errorf("a push that changes %s does not deploy", p))
		}
	}
	after, _ := w.On["workflow_run"].(map[string]any)
	workflows, _ := after["workflows"].([]any)
	types, _ := after["types"].([]any)
	if !slices.Equal(workflows, []any{release}) || !slices.Equal(types, []any{"completed"}) {
		errs = append(errs, fmt.Errorf("the workflow does not run after the %s workflow", release))
	}
	for k, v := range map[string]string{"contents": "read", "pages": "write", "id-token": "write"} {
		if w.Permissions[k] != v {
			errs = append(errs, fmt.Errorf("permission %s is %q, not %s", k, w.Permissions[k], v))
		}
	}
	if w.Concurrency["group"] != "pages" || w.Concurrency["cancel-in-progress"] != false {
		errs = append(errs, errors.New("deployments are not serialized in the pages group"))
	}
	job, ok := w.Jobs["deploy"]
	if len(w.Jobs) != 1 || !ok {
		return errors.Join(append(errs, errors.New("the workflow does not have one deploy job"))...)
	}
	if job.If != "github.event_name != 'workflow_run' || github.event.workflow_run.conclusion == 'success'" {
		errs = append(errs, fmt.Errorf("the job runs after a failed release workflow: if %q", job.If))
	}
	if job.Environment["name"] != "github-pages" {
		errs = append(errs, errors.New("the job does not deploy in the github-pages environment"))
	}

	find := func(what string, match func(s workflowStep) bool) int {
		i := slices.IndexFunc(job.Steps, func(s pagesStep) bool { return match(s.workflowStep) })
		if i < 0 {
			errs = append(errs, fmt.Errorf("no step that %s", what))
		}
		return i
	}
	checkout := find("checks out the repository", func(s workflowStep) bool { return strings.HasPrefix(s.Uses, "actions/checkout@") })
	setup := find("sets up Go from go.mod", func(s workflowStep) bool {
		return strings.HasPrefix(s.Uses, "actions/setup-go@") && s.With["go-version-file"] == "go.mod"
	})
	site := find("assembles the site with the module of the release", func(s workflowStep) bool {
		return s.Run == "go run ./tools/site -download https://github.com/ekalinin/gopuml/releases/download -out _site" &&
			s.If == "" && s.ContinueOnError == nil
	})
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	id := job.Steps[site].ID
	if id == "" {
		return errors.New("the step that assembles the site has no id")
	}
	// The deployment steps run only when the site was assembled.
	deployed := func(action string, with func(map[string]any) bool) func(s workflowStep) bool {
		return func(s workflowStep) bool {
			return strings.HasPrefix(s.Uses, action+"@") && s.If == "steps."+id+".outputs.deploy == 'true'" && with(s.With)
		}
	}
	anything := func(map[string]any) bool { return true }
	configure := find("configures Pages when the site is assembled", deployed("actions/configure-pages", anything))
	upload := find("uploads _site when the site is assembled", deployed("actions/upload-pages-artifact", func(w map[string]any) bool { return w["path"] == "_site" }))
	deploy := find("deploys when the site is assembled", deployed("actions/deploy-pages", anything))
	if len(errs) == 0 && !(checkout < site && setup < site && site < configure && configure < upload && upload < deploy) {
		errs = append(errs, errors.New("the steps are not in the order checkout, setup, site, configure, upload, deploy"))
	}
	return errors.Join(errs...)
}

func TestPagesWorkflow(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/pages.yml")
	if err != nil {
		t.Fatal(err)
	}
	var release struct {
		Name string `yaml:"name"`
	}
	releaseData, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(releaseData, &release); err != nil || release.Name == "" {
		t.Fatalf("name of the release workflow: %q, %v", release.Name, err)
	}
	if err := checkPagesWorkflow(data, release.Name); err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, m := range []struct{ name, old, new string }{
		{"no workflow_run", "  workflow_run:\n    workflows: [release]\n    types: [completed]\n", ""},
		{"after another workflow", "workflows: [release]", "workflows: [ci]"},
		{"no success condition", "    if: github.event_name != 'workflow_run' || github.event.workflow_run.conclusion == 'success'\n", ""},
		{"runs on pull requests", "on:\n", "on:\n  pull_request:\n"},
		{"runs on every branch", "    branches: [main]\n", ""},
		{"page changes not deployed", "      - \"site/**\"\n", ""},
		{"notice changes not deployed", "      - \"NOTICE\"\n", ""},
		{"no id-token: write", "  id-token: write\n", ""},
		{"cancelled deployments", "cancel-in-progress: false", "cancel-in-progress: true"},
		{"local module", "-download https://github.com/ekalinin/gopuml/releases/download", "-wasm dist/wasm"},
		{"failure tolerated", "        id: site\n", "        id: site\n        continue-on-error: true\n"},
		{"unconditional deployment", "        id: deployment\n        if: steps.site.outputs.deploy == 'true'\n", "        id: deployment\n"},
		{"other directory", "path: _site", "path: site"},
	} {
		mutated := strings.ReplaceAll(s, m.old, m.new)
		if mutated == s {
			t.Errorf("%s: %q is not in the workflow", m.name, m.old)
			continue
		}
		if checkPagesWorkflow([]byte(mutated), release.Name) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
}

// TestPlaygroundStartHint checks the hint of a playground that did not
// start: shown by the page, hidden by page.mjs when it starts, and hidden
// without JavaScript, where the <noscript> message shows instead.
func TestPlaygroundStartHint(t *testing.T) {
	read := func(name string) string {
		data, err := os.ReadFile(filepath.Join("site", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	playground, err := htmlSection(read("index.html"), "playground")
	if err != nil {
		t.Fatal(err)
	}
	hint := regexp.MustCompile(`(?s)<p id="pg-start"[^>]*>(.*?)</p>`).FindStringSubmatch(playground)
	if hint == nil || strings.Contains(hint[0], " hidden") || !strings.Contains(htmlText(hint[1]), "over HTTP") {
		t.Errorf("no shown hint that the playground needs the page over HTTP: %q", hint)
	}
	if !regexp.MustCompile(`(?s)@media \(scripting: none\) \{\s*#pg-start \{\s*display: none;`).MatchString(read("style.css")) {
		t.Error("style.css does not hide #pg-start without JavaScript")
	}
	if !strings.Contains(read("page.mjs"), `document.getElementById('pg-start').hidden = true;`) {
		t.Error("page.mjs does not hide #pg-start when it starts")
	}
}
