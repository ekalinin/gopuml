// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// releaseConfig is the part of .goreleaser.yaml that checkReleaseConfig reads.
type releaseConfig struct {
	Before struct {
		Hooks []string `yaml:"hooks"`
	} `yaml:"before"`
	Builds   []releaseBuild `yaml:"builds"`
	Archives []struct {
		IDs             []string          `yaml:"ids"`
		NameTemplate    string            `yaml:"name_template"`
		Formats         []string          `yaml:"formats"`
		FormatOverrides []releaseOverride `yaml:"format_overrides"`
		Files           []any             `yaml:"files"`
	} `yaml:"archives"`
	Checksum struct {
		NameTemplate string `yaml:"name_template"`
		Algorithm    string `yaml:"algorithm"`
	} `yaml:"checksum"`
	Changelog struct {
		Disable any `yaml:"disable"`
	} `yaml:"changelog"`
}

type releaseBuild struct {
	ID      string   `yaml:"id"`
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	Env     []string `yaml:"env"`
	Flags   []string `yaml:"flags"`
	Ldflags []string `yaml:"ldflags"`
	Goos    []string `yaml:"goos"`
	Goarch  []string `yaml:"goarch"`
	Ignore  []any    `yaml:"ignore"`
}

type releaseOverride struct {
	Goos    string   `yaml:"goos"`
	Formats []string `yaml:"formats"`
}

// checkReleaseConfig checks the contract of the GoReleaser configuration
// (release: Release artifacts, Version of a release build).
func checkReleaseConfig(data []byte) error {
	var c releaseConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return err
	}
	var errs []error
	// build finds the build of main and checks what both builds share.
	build := func(main string) (releaseBuild, bool) {
		i := slices.IndexFunc(c.Builds, func(b releaseBuild) bool {
			return b.Main == main
		})
		if i < 0 {
			errs = append(errs, fmt.Errorf("no build of %s", main))
			return releaseBuild{}, false
		}
		b := c.Builds[i]
		if b.Binary != "plantuml" {
			errs = append(errs, fmt.Errorf("the binary of %s is %q, not plantuml", main, b.Binary))
		}
		if !slices.Contains(b.Env, "CGO_ENABLED=0") {
			errs = append(errs, fmt.Errorf("the build of %s does not disable cgo", main))
		}
		if !slices.Contains(b.Flags, "-trimpath") {
			errs = append(errs, fmt.Errorf("the build of %s is not built with -trimpath", main))
		}
		return b, true
	}
	sameSet := func(got []string, want ...string) bool {
		got = slices.Clone(got)
		slices.Sort(got)
		slices.Sort(want)
		return slices.Equal(got, want)
	}

	cli, ok := build("./cmd/plantuml")
	if ok {
		if !sameSet(cli.Goos, "darwin", "linux", "windows") || !sameSet(cli.Goarch, "amd64", "arm64") || len(cli.Ignore) > 0 {
			errs = append(errs, fmt.Errorf("the CLI targets are %v × %v (ignored %v), not linux, darwin and windows × amd64 and arm64", cli.Goos, cli.Goarch, cli.Ignore))
		}
		if !slices.ContainsFunc(cli.Ldflags, func(f string) bool {
			return strings.Contains(f, "-X github.com/ekalinin/gopuml/internal/version.Version=v{{ .Version }}")
		}) {
			errs = append(errs, errors.New("the CLI build does not set the version to the tag"))
		}
	}
	wasm, ok := build("./cmd/plantuml-wasm")
	if ok && (!slices.Equal(wasm.Goos, []string{"js"}) || !slices.Equal(wasm.Goarch, []string{"wasm"})) {
		errs = append(errs, fmt.Errorf("the WASM targets are %v × %v, not js × wasm", wasm.Goos, wasm.Goarch))
	}

	// The sources of the files of an archive, given as a path or as src.
	files := func(entries []any) []string {
		var srcs []string
		for _, e := range entries {
			switch e := e.(type) {
			case string:
				srcs = append(srcs, e)
			case map[string]any:
				src, _ := e["src"].(string)
				srcs = append(srcs, src)
			}
		}
		return srcs
	}
	archive := func(id, name string) int {
		i := -1
		for j, a := range c.Archives {
			if id != "" && slices.Equal(a.IDs, []string{id}) {
				i = j
			}
		}
		if i < 0 {
			errs = append(errs, fmt.Errorf("no archive of the %s build alone", name))
		}
		return i
	}
	if i := archive(cli.ID, "CLI"); i >= 0 {
		a := c.Archives[i]
		if a.NameTemplate != "gopuml_{{ .Version }}_{{ .Os }}_{{ .Arch }}" {
			errs = append(errs, fmt.Errorf("the CLI archives are named %q", a.NameTemplate))
		}
		zip := slices.ContainsFunc(a.FormatOverrides, func(o releaseOverride) bool {
			return o.Goos == "windows" && slices.Equal(o.Formats, []string{"zip"})
		})
		if !slices.Equal(a.Formats, []string{"tar.gz"}) || !zip {
			errs = append(errs, errors.New("the CLI archives are not tar.gz, and zip on windows"))
		}
		if got := files(a.Files); !sameSet(got, "LICENSE", "NOTICE", "README.md") {
			errs = append(errs, fmt.Errorf("the CLI archives contain %v besides the binary, not LICENSE, NOTICE and README.md", got))
		}
	}
	if i := archive(wasm.ID, "WASM"); i >= 0 {
		a := c.Archives[i]
		if a.NameTemplate != "gopuml_{{ .Version }}_wasm" || !slices.Equal(a.Formats, []string{"tar.gz"}) {
			errs = append(errs, fmt.Errorf("the WASM archive is %q in %v, not gopuml_{{ .Version }}_wasm in tar.gz", a.NameTemplate, a.Formats))
		}
		got := files(a.Files)
		staged := slices.IndexFunc(got, func(src string) bool { return strings.HasSuffix(src, "/wasm_exec.js") })
		if staged < 0 || !slices.Contains(got, "cmd/plantuml-wasm/plantuml.mjs") || len(got) != 2 {
			errs = append(errs, fmt.Errorf("the WASM archive contains %v besides plantuml.wasm, not wasm_exec.js and plantuml.mjs", got))
		} else if !slices.ContainsFunc(c.Before.Hooks, func(h string) bool {
			return strings.Contains(h, `"$(go env GOROOT)/lib/wasm/wasm_exec.js" `+got[staged])
		}) {
			errs = append(errs, fmt.Errorf("no before hook copies wasm_exec.js of the Go root to %s", got[staged]))
		}
	}

	if c.Checksum.NameTemplate != "checksums.txt" || c.Checksum.Algorithm != "sha256" {
		errs = append(errs, fmt.Errorf("the checksums are %q with %q, not checksums.txt with sha256", c.Checksum.NameTemplate, c.Checksum.Algorithm))
	}
	// A disabled changelog also ignores the notes given with --release-notes.
	if c.Changelog.Disable != nil && c.Changelog.Disable != false {
		errs = append(errs, fmt.Errorf("the changelog is disabled (%v), so the release notes would be ignored", c.Changelog.Disable))
	}
	return errors.Join(errs...)
}

func TestReleaseConfig(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkReleaseConfig(data); err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ name, old, new string }{
		{"no windows target", "      - windows\n", ""},
		{"no arm64 target", "      - arm64\n", ""},
		{"an ignored target", "    goarch:\n      - amd64\n", "    ignore:\n      - goos: windows\n        goarch: arm64\n    goarch:\n      - amd64\n"},
		{"cgo", "CGO_ENABLED=0", "CGO_ENABLED=1"},
		{"no version flag", "-X github.com/ekalinin/gopuml/internal/version.Version=v{{ .Version }}", ""},
		{"version without v", "Version=v{{ .Version }}", "Version={{ .Version }}"},
		{"no NOTICE", "      - NOTICE\n", ""},
		{"no LICENSE", "      - LICENSE\n", ""},
		{"no zip on windows", "          - zip\n", "          - tar.gz\n"},
		{"no plantuml.mjs", "      - src: cmd/plantuml-wasm/plantuml.mjs\n        strip_parent: true\n", ""},
		{"wasm_exec.js not staged", "/lib/wasm/wasm_exec.js\" .cache", "/misc/wasm/wasm_exec.js\" .cache"},
		{"no WASM build", "  - id: wasm\n    main: ./cmd/plantuml-wasm\n", "  - id: wasm\n    main: ./cmd/other\n"},
		{"checksums with SHA-1", "algorithm: sha256", "algorithm: sha1"},
		{"changelog disabled", "  algorithm: sha256\n", "  algorithm: sha256\n\nchangelog:\n  disable: true\n"},
	} {
		mutated := strings.ReplaceAll(string(data), m.old, m.new)
		if mutated == string(data) {
			t.Errorf("%s: %q is not in the configuration", m.name, m.old)
			continue
		}
		if checkReleaseConfig([]byte(mutated)) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
}

// releaseWorkflow is the part of the release workflow that
// checkReleaseWorkflow reads.
type releaseWorkflow struct {
	On          map[string]any    `yaml:"on"`
	Permissions map[string]string `yaml:"permissions"`
	Jobs        map[string]struct {
		Steps []workflowStep `yaml:"steps"`
	} `yaml:"jobs"`
}

type workflowStep struct {
	If              string         `yaml:"if"`
	Uses            string         `yaml:"uses"`
	Run             string         `yaml:"run"`
	With            map[string]any `yaml:"with"`
	ContinueOnError any            `yaml:"continue-on-error"`
}

// checkReleaseWorkflow checks the contract of the release workflow
// (release: Release gate, Release notes).
func checkReleaseWorkflow(data []byte) error {
	var w releaseWorkflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return err
	}
	var errs []error
	push, _ := w.On["push"].(map[string]any)
	tags, _ := push["tags"].([]any)
	if len(w.On) != 1 || len(push) != 1 || !slices.Equal(tags, []any{"v*.*.*"}) {
		errs = append(errs, fmt.Errorf("the workflow runs on %v, not on pushed v*.*.* tags only", w.On))
	}
	if w.Permissions["contents"] != "write" {
		errs = append(errs, errors.New("the workflow cannot write the contents of the repository"))
	}
	job, ok := w.Jobs["release"]
	if len(w.Jobs) != 1 || !ok {
		return errors.Join(append(errs, errors.New("the workflow does not have one release job"))...)
	}
	steps := job.Steps
	find := func(what string, match func(s workflowStep) bool) int {
		i := slices.IndexFunc(steps, match)
		if i < 0 {
			errs = append(errs, fmt.Errorf("no step that %s", what))
		}
		return i
	}
	// A gate step runs unconditionally and fails the job when it fails.
	gate := func(run string) func(s workflowStep) bool {
		return func(s workflowStep) bool {
			return s.Run == run && s.If == "" && s.ContinueOnError == nil
		}
	}
	checkout := find("checks out the full history", func(s workflowStep) bool {
		return strings.HasPrefix(s.Uses, "actions/checkout@") && s.With["fetch-depth"] == 0
	})
	setup := find("sets up Go from go.mod", func(s workflowStep) bool {
		return strings.HasPrefix(s.Uses, "actions/setup-go@") && s.With["go-version-file"] == "go.mod"
	})
	fetch := find("fetches the upstream sources", gate("tools/upstream/fetch.sh"))
	lint := find("runs the lint", gate("make lint"))
	test := find("runs the tests", gate("make test"))
	release := find("publishes with the pinned GoReleaser action and the notes of the tag", func(s workflowStep) bool {
		version, _ := s.With["version"].(string)
		args, _ := s.With["args"].(string)
		return strings.HasPrefix(s.Uses, "goreleaser/goreleaser-action@v") && s.If == "" && strings.HasPrefix(version, "~> v2") &&
			strings.HasPrefix(args, "release --clean ") && strings.Contains(args, "--release-notes docs/releases/${{ github.ref_name }}.md")
	})
	if len(errs) == 0 && !(checkout < fetch && setup < lint && fetch < lint && fetch < test && lint < release && test < release) {
		errs = append(errs, errors.New("the steps are not in the order checkout, setup, upstream sources, lint and tests, release"))
	}
	return errors.Join(errs...)
}

func TestReleaseWorkflow(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkReleaseWorkflow(data); err != nil {
		t.Fatal(err)
	}
	s := string(data)
	testStep := "      - name: Test\n        run: make test\n\n"
	releaseStep := s[strings.Index(s, "      - name: Release\n"):]
	for _, m := range []struct{ name, old, new string }{
		{"no tests", testStep, ""},
		{"tests after the release", testStep + releaseStep, releaseStep + "\n" + strings.TrimSuffix(testStep, "\n")},
		{"tests allowed to fail", "        run: make test\n", "        run: make test\n        continue-on-error: true\n"},
		{"release after a failure", "      - name: Release\n", "      - name: Release\n        if: always()\n"},
		{"no lint", "      - name: Lint\n        run: make lint\n\n", ""},
		{"runs on branches", "  push:\n", "  push:\n    branches: [main]\n"},
		{"runs on pull requests", "on:\n", "on:\n  pull_request:\n"},
		{"read-only contents", "contents: write", "contents: read"},
		{"shallow checkout", "fetch-depth: 0", "fetch-depth: 1"},
		{"no release notes", " --release-notes docs/releases/${{ github.ref_name }}.md", ""},
		{"unpinned GoReleaser", `version: "~> v2"`, "version: latest"},
	} {
		mutated := strings.ReplaceAll(s, m.old, m.new)
		if mutated == s {
			t.Errorf("%s: %q is not in the workflow", m.name, m.old)
			continue
		}
		if checkReleaseWorkflow([]byte(mutated)) == nil {
			t.Errorf("%s: the check passes", m.name)
		}
	}
}

// checkReleaseNotes checks the notes of a release (release: Release notes)
// against the README, whose table of supported diagrams they repeat.
func checkReleaseNotes(notes, readme string) error {
	var errs []error
	for _, section := range []string{
		"## Supported diagrams and formats",
		"## Not supported yet",
		"## Graphviz",
		"## Conformance",
		"## Comparison with upstream",
		"## Installation",
	} {
		if !strings.Contains(notes, "\n"+section+"\n") {
			errs = append(errs, fmt.Errorf("no section %q", section))
		}
	}
	_, table, _ := strings.Cut(readme, "\n## Supported diagrams and formats\n")
	table, _, _ = strings.Cut(table, "\n\n## ")
	var rows int
	for line := range strings.SplitSeq(table, "\n") {
		if strings.HasPrefix(line, "|") {
			rows++
			if !strings.Contains(notes, "\n"+line+"\n") {
				errs = append(errs, fmt.Errorf("no row %q of the table of the README", line))
			}
		}
	}
	if rows == 0 {
		errs = append(errs, errors.New("no table of supported diagrams in the README"))
	}
	for _, want := range []string{
		"| | PlantUML 1.2026.8 (Java) | gopuml |",
		"| Start time (one small sequence diagram) |",
		"- Runs: median of ",
		"xattr -d com.apple.quarantine plantuml",
	} {
		if !strings.Contains(notes, want) {
			errs = append(errs, fmt.Errorf("no %q", want))
		}
	}
	return errors.Join(errs...)
}

func TestReleaseNotes(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	// The releases that the installation of the README names (go install
	// and gh release download), and the releases that have notes.
	var tags []string
	for _, m := range regexp.MustCompile(`(?:plantuml@|release download )(v\d+\.\d+\.\d+)\b`).FindAllStringSubmatch(string(readme), -1) {
		tags = append(tags, m[1])
	}
	notes, _ := filepath.Glob(filepath.Join("docs", "releases", "v*.md"))
	for _, n := range notes {
		tags = append(tags, strings.TrimSuffix(filepath.Base(n), ".md"))
	}
	slices.Sort(tags)
	tags = slices.Compact(tags)
	if !slices.Contains(tags, "v0.1.0") {
		t.Errorf("v0.1.0 is not among the releases %v", tags)
	}
	for _, tag := range tags {
		data, err := os.ReadFile(filepath.Join("docs", "releases", tag+".md"))
		if err != nil {
			t.Errorf("%s: %v", tag, err)
			continue
		}
		if err := checkReleaseNotes(string(data), string(readme)); err != nil {
			t.Errorf("%s: %v", tag, err)
		}
	}
}
