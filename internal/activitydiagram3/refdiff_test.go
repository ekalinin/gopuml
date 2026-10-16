// SPDX-License-Identifier: MIT
// Provenance: original

//go:build activityref

package activitydiagram3

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The reference differential test renders activity diagrams with the
// reference jar in the Docker image plantuml-go-reference (built from
// tools/reference/Dockerfile) and compares the SVG byte for byte (except
// <?plantuml-src?>) with this package's output. The references are cached
// in .cache/activity-ref by the hash of the source. Sources: the committed
// parser cases (testdata/upstream/cases, the generated combinations, the
// conformance cases), the random drawing cases of diffgen_test.go and
// testdata/render/*.txt.
//
//	go test -tags activityref -run TestActivityReferenceDiff ./internal/activitydiagram3
//
// ACTIVITY_REF_CACHE overrides the cache directory (an absolute path below
// the home directory, which Docker mounts). ACTIVITY_REF_FILTER selects cases by a regular expression on their names,
// ACTIVITY_REF_EXCLUDE skips the sources matching a regular expression; the
// report is written to ACTIVITY_REF_REPORT (default
// /tmp/activity-refdiff.txt).

const (
	refCacheDir = "../../.cache/activity-ref"
	renderDir   = "testdata/render"
)

// refCache returns the reference cache directory: ACTIVITY_REF_CACHE when
// set, else .cache/activity-ref of the repository.
func refCache(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("ACTIVITY_REF_CACHE"); dir != "" {
		return dir
	}
	dir, err := filepath.Abs(refCacheDir)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func refKey(source string) string {
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:12])
}

// refSource is the source a case renders: unescaped, wrapped in
// @startuml/@enduml when needed, as parseCase does.
func refSource(c testCase) string {
	source := c.source
	if strings.HasSuffix(c.name, " (escaped)") {
		source = unescape(source)
	}
	if !strings.HasPrefix(strings.TrimSpace(source), "@start") {
		source = "@startuml\n" + source + "@enduml\n"
	}
	return source
}

func renderCases(t *testing.T) []testCase {
	t.Helper()
	all := committedCases(t)
	all = append(all, generatedDrawingCasesList()...)
	files, _ := filepath.Glob(filepath.Join(renderDir, "*.txt"))
	sort.Strings(files)
	for _, f := range files {
		cases, err := readCases(f)
		if err != nil {
			t.Fatal(err)
		}
		for i := range cases {
			cases[i].name = "render/" + strings.TrimSuffix(filepath.Base(f), ".txt") + "/" + cases[i].name
		}
		all = append(all, cases...)
	}
	if filter := os.Getenv("ACTIVITY_REF_FILTER"); filter != "" {
		re := regexp.MustCompile(filter)
		var kept []testCase
		for _, c := range all {
			if re.MatchString(c.name) {
				kept = append(kept, c)
			}
		}
		all = kept
	}
	return all
}

// ensureReferences renders the sources without cached reference.
func ensureReferences(t *testing.T, sources map[string]string) {
	t.Helper()
	cache := refCache(t)
	var missing []string
	for key := range sources {
		if _, err := os.Stat(filepath.Join(cache, key, "meta.json")); err != nil {
			missing = append(missing, key)
		}
	}
	if len(missing) == 0 {
		return
	}
	sort.Strings(missing)
	work, err := os.MkdirTemp(filepath.Dir(cache), "activity-ref-work-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(work)
	var manifest strings.Builder
	for _, key := range missing {
		if err := os.WriteFile(filepath.Join(work, key+".puml"), []byte(sources[key]), 0o644); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&manifest, "%s\t/h/%s.puml\tsvg\n", key, key)
	}
	if err := os.WriteFile(filepath.Join(work, "manifest.txt"), []byte(manifest.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("docker", "run", "--rm", "--platform", "linux/amd64", "-v", work+":/h",
		"plantuml-go-reference", "java", "-Djava.awt.headless=true", "-cp", "/opt/plantuml.jar:/opt/tools",
		"ReferenceRender", "/h/manifest.txt", "/h/out")
	cmd.Stderr = os.Stderr
	if out, err := cmd.Output(); err != nil {
		t.Fatalf("reference rendering: %v\n%s", err, out)
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, key := range missing {
		if err := os.Rename(filepath.Join(work, "out", key), filepath.Join(cache, key)); err != nil {
			// Another run may have rendered it meanwhile
			if _, statErr := os.Stat(filepath.Join(cache, key, "meta.json")); statErr != nil {
				t.Fatal(err)
			}
		}
	}
}

type refMeta struct {
	Crash  string `json:"crash"`
	Blocks []struct {
		Type    string `json:"type"`
		Renders []struct {
			File *string `json:"file"`
		} `json:"renders"`
	} `json:"blocks"`
}

func TestActivityReferenceDiff(t *testing.T) {
	cases := renderCases(t)
	sources := map[string]string{}
	for _, c := range cases {
		s := refSource(c)
		sources[refKey(s)] = s
	}
	ensureReferences(t, sources)

	report := os.Getenv("ACTIVITY_REF_REPORT")
	if report == "" {
		report = "/tmp/activity-refdiff.txt"
	}
	var out strings.Builder
	counts := map[string]int{}
	var failures []string
	seen := map[string]bool{}
	var exclude *regexp.Regexp
	if e := os.Getenv("ACTIVITY_REF_EXCLUDE"); e != "" {
		exclude = regexp.MustCompile(e)
	}
	for _, c := range cases {
		source := refSource(c)
		key := refKey(source)
		if seen[key] {
			continue
		}
		seen[key] = true
		dir := filepath.Join(refCache(t), key)
		if _, err := os.Stat(filepath.Join(dir, "source.puml")); err != nil {
			if err := os.WriteFile(filepath.Join(dir, "source.puml"), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if exclude != nil && exclude.MatchString(source) {
			counts["excluded"]++
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, "meta.json"))
		if err != nil {
			t.Fatal(err)
		}
		var meta refMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		if meta.Crash != "" || len(meta.Blocks) == 0 || meta.Blocks[0].Type != "ACTIVITY" ||
			len(meta.Blocks[0].Renders) == 0 || meta.Blocks[0].Renders[0].File == nil {
			counts["skipped (not drawn as an activity diagram upstream)"]++
			continue
		}
		want, err := os.ReadFile(filepath.Join(dir, *meta.Blocks[0].Renders[0].File))
		if err != nil {
			t.Fatal(err)
		}
		got, err := renderSVG(source)
		status := "identical"
		switch {
		case err != nil:
			status = "error"
			failures = append(failures, fmt.Sprintf("%s [%s]: error: %v", c.name, key, err))
		case !sameSVG(got, string(want)):
			status = "differs"
			failures = append(failures, fmt.Sprintf("%s [%s]: differs\n%s", c.name, key,
				firstDifference(got, string(want))))
		}
		counts[status]++
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&out, "%s: %d\n", k, counts[k])
	}
	out.WriteString("\n")
	for _, f := range failures {
		out.WriteString(f)
		out.WriteString("\n\n")
	}
	if err := os.WriteFile(report, []byte(out.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		t.Logf("%s: %d", k, counts[k])
	}
	t.Logf("report: %s", report)
}
