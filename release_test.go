// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"errors"
	"fmt"
	"os"
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
