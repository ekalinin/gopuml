// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Meta is the recorded result of rendering one diagram case (meta.json).
type Meta struct {
	ID     string      `json:"id"`
	Blocks []BlockMeta `json:"blocks"`
	Crash  string      `json:"crash"`
}

// BlockMeta describes one @start/@end block of a diagram case.
type BlockMeta struct {
	Index       int          `json:"index"`
	Name        *string      `json:"name"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	Pages       int          `json:"pages"`
	Error       *ErrorMeta   `json:"error"`
	Renders     []RenderMeta `json:"renders"`
}

// ErrorMeta is the error of a failed block.
type ErrorMeta struct {
	Line     int      `json:"line"`
	Messages []string `json:"messages"`
}

// RenderMeta is one rendered page in one format.
type RenderMeta struct {
	File        *string `json:"file"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Status      int     `json:"status"`
	CrashReport bool    `json:"crashReport"`
}

// CLIRecord is the recorded result of one CLI case (cli.json).
type CLIRecord struct {
	Args   []string     `json:"args"`
	Exit   int          `json:"exit"`
	Stdout string       `json:"stdout"`
	Stderr string       `json:"stderr"`
	Files  []FileRecord `json:"files"`
}

// FileRecord describes a file created by a CLI case or an HTTP body.
type FileRecord struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	Stored bool   `json:"stored"`
}

// HTTPRecord is the recorded result of one HTTP case (http.json).
type HTTPRecord struct {
	Server  string            `json:"server"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Encoded string            `json:"encoded,omitempty"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    FileRecord        `json:"body"`
}

// KnownDifference is an accepted, justified failure of one case.
type KnownDifference struct {
	ID     string `yaml:"id"`
	Reason string `yaml:"reason"`
	// ResolvedBy names the follow-up change expected to fix the case.
	ResolvedBy string `yaml:"resolvedBy"`
}

// LoadKnownDifferences reads a known-differences YAML list. A missing file
// means no known differences.
func LoadKnownDifferences(file string) (map[string]KnownDifference, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]KnownDifference{}, nil
	}
	if err != nil {
		return nil, err
	}
	var list []KnownDifference
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	out := make(map[string]KnownDifference, len(list))
	for _, k := range list {
		switch {
		case k.ID == "" || strings.TrimSpace(k.Reason) == "":
			return nil, fmt.Errorf("%s: every entry needs an id and a reason", file)
		case out[k.ID].ID != "":
			return nil, fmt.Errorf("%s: duplicate entry %q", file, k.ID)
		}
		out[k.ID] = k
	}
	return out, nil
}

// Result statuses.
const (
	StatusPass      = "pass"
	StatusFail      = "fail"
	StatusKnown     = "known"      // fails, listed as a known difference
	StatusKnownPass = "known-pass" // listed as a known difference but passes
	// StatusUnsupported is a case of a diagram type this implementation
	// does not provide yet, reported with an explicit "not supported yet"
	// error instead of a wrong output.
	StatusUnsupported = "unsupported"
)

// CaseResult is the outcome of comparing one case with its reference.
type CaseResult struct {
	ID string
	// Group is the diagram type for diagram cases, or "cli" / "http".
	Group string
	// Failures not tied to one format (outcome, block count, missing files).
	Failures []string
	// FormatFailures are failures per output format.
	FormatFailures map[string][]string
	// Formats lists the formats the case covers.
	Formats []string
	Known   *KnownDifference
	// Unsupported lists the types of the blocks that the implementation
	// reported as not supported yet.
	Unsupported []string
}

// Passed reports whether the case matches its reference.
func (r *CaseResult) Passed() bool {
	if len(r.Failures) > 0 {
		return false
	}
	for _, f := range r.FormatFailures {
		if len(f) > 0 {
			return false
		}
	}
	return true
}

// Status returns pass, fail, known, known-pass or unsupported.
func (r *CaseResult) Status() string {
	switch {
	case r.Passed() && len(r.Unsupported) > 0:
		return StatusUnsupported
	case r.Passed() && r.Known != nil:
		return StatusKnownPass
	case r.Passed():
		return StatusPass
	case r.Known != nil:
		return StatusKnown
	}
	return StatusFail
}

// FormatPassed reports whether one format of the case matches.
func (r *CaseResult) FormatPassed(format string) bool {
	return len(r.Failures) == 0 && len(r.FormatFailures[format]) == 0
}

func (r *CaseResult) fail(format, msg string) {
	if format == "" {
		r.Failures = append(r.Failures, msg)
		return
	}
	if r.FormatFailures == nil {
		r.FormatFailures = map[string][]string{}
	}
	r.FormatFailures[format] = append(r.FormatFailures[format], msg)
}

// Runner compares the output of an implementation with the references.
type Runner struct {
	// Reference returns the reference directory of a case id.
	Reference func(id string) string
	// Actual is the root of the implementation's output tree (<Actual>/<id>).
	Actual string
	// PNGReference and PNGActual hold PNG outputs, when PNG is compared.
	PNGReference, PNGActual string
	// Rasterizer is used for SSIM on SVG outputs; nil skips that check.
	Rasterizer Rasterizer
	Known      map[string]KnownDifference
}

func readJSON(file string, v any) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// CompareDiagram compares one diagram case.
func (r *Runner) CompareDiagram(c *Case) (res CaseResult) {
	res = CaseResult{ID: c.ID, Formats: c.Formats, Group: "none"}
	defer r.markKnown(&res)
	refDir, actDir := r.Reference(c.ID), filepath.Join(r.Actual, filepath.FromSlash(c.ID))

	var ref, act Meta
	if err := readJSON(filepath.Join(refDir, "meta.json"), &ref); err != nil {
		res.fail("", "reading reference: "+err.Error())
		return res
	}
	res.Group = groupOf(c, &ref)
	if err := readJSON(filepath.Join(actDir, "meta.json"), &act); err != nil {
		res.fail("", "reading output: "+err.Error())
		return res
	}
	if act.Crash != "" && ref.Crash == "" {
		res.fail("", "implementation crashed: "+act.Crash)
		return res
	}
	if len(ref.Blocks) != len(act.Blocks) {
		res.fail("", fmt.Sprintf("%d block(s), reference has %d", len(act.Blocks), len(ref.Blocks)))
		return res
	}
	for i := range ref.Blocks {
		r.compareBlock(&res, c, refDir, actDir, &ref.Blocks[i], &act.Blocks[i])
	}
	return res
}

func outcome(b *BlockMeta) Outcome {
	if b.Error != nil {
		return Outcome{Failed: true, ErrorLine: b.Error.Line}
	}
	return Outcome{}
}

// reportsUnsupported tells whether the implementation reported the block as
// a diagram type it does not provide yet. Sequence and class (object)
// diagrams are always in scope; an activity diagram may only be the legacy
// syntax. A reference error counts when the case is of another type
// (caseType, from the case header), whose parser upstream ran.
func reportsUnsupported(ref, act *BlockMeta, caseType string) bool {
	message := ""
	if act.Error != nil && len(act.Error.Messages) > 0 {
		message = act.Error.Messages[0]
	}
	if ref.Error != nil {
		switch caseType {
		case "", "uml", "sequence", "class", "object", "activity":
			return false
		}
		return strings.HasSuffix(message, " not supported yet")
	}
	switch ref.Type {
	case "SEQUENCE", "CLASS", "OBJECT":
		return false
	case "ACTIVITY":
		return message == "legacy activity diagrams are not supported yet"
	}
	// OTHER is the implementation's "Diagram not supported by this release".
	return strings.HasSuffix(message, " not supported yet") || (act.Error == nil && act.Type == "OTHER" && ref.Type != "OTHER")
}

func (r *Runner) compareBlock(res *CaseResult, c *Case, refDir, actDir string, ref, act *BlockMeta) {
	prefix := fmt.Sprintf("block %d: ", ref.Index)
	caseType := c.Type
	if caseType == "" && c.Family != "uml" {
		// A Vega case only records its family, such as "packetdiag".
		caseType = c.Family
	}
	if reportsUnsupported(ref, act, caseType) {
		res.Unsupported = append(res.Unsupported, res.Group)
		return
	}
	if err := CompareOutcome(outcome(ref), outcome(act)); err != nil {
		res.fail("", prefix+err.Error())
		return
	}
	if ref.Pages != act.Pages {
		res.fail("", fmt.Sprintf("%s%d page(s), reference has %d", prefix, act.Pages, ref.Pages))
		return
	}
	actual := map[string]RenderMeta{}
	for _, a := range act.Renders {
		if a.File != nil {
			actual[*a.File] = a
		}
	}
	for _, rr := range ref.Renders {
		if rr.CrashReport {
			// Upstream crashed; only a failure of the implementation matches.
			if act.Error == nil && !slices.ContainsFunc(act.Renders, func(a RenderMeta) bool { return a.CrashReport || a.Status >= 400 }) {
				res.fail("", prefix+"reference is a crash report but the diagram rendered")
			}
			continue
		}
		name := *rr.File
		format := formatOf(name)
		ar, ok := actual[name]
		if !ok {
			res.fail(format, prefix+"missing output "+name)
			continue
		}
		r.compareRender(res, format, prefix+name+": ", filepath.Join(refDir, name), filepath.Join(actDir, name), rr, ar)
	}
	if slices.Contains(c.Formats, "preproc") {
		name := fmt.Sprintf("block%d.preproc", ref.Index)
		refText, err1 := os.ReadFile(filepath.Join(refDir, name))
		actText, err2 := os.ReadFile(filepath.Join(actDir, name))
		switch {
		case err1 != nil:
			res.fail("preproc", prefix+"reading reference: "+err1.Error())
		case err2 != nil:
			res.fail("preproc", prefix+"missing output "+name)
		default:
			if err := CompareText(string(refText), string(actText)); err != nil {
				res.fail("preproc", prefix+err.Error())
			}
		}
	}
	if slices.Contains(c.Formats, "png") && r.PNGReference != "" {
		r.comparePNG(res, c.ID, prefix, ref)
	}
}

func (r *Runner) compareRender(res *CaseResult, format, prefix, refFile, actFile string, ref, act RenderMeta) {
	refData, err := os.ReadFile(refFile)
	if err != nil {
		res.fail(format, prefix+"reading reference: "+err.Error())
		return
	}
	actData, err := os.ReadFile(actFile)
	if err != nil {
		res.fail(format, prefix+"reading output: "+err.Error())
		return
	}
	switch format {
	case "txt", "utxt":
		if err := CompareText(string(refData), string(actData)); err != nil {
			res.fail(format, prefix+err.Error())
		}
	case "svg":
		refTexts, err1 := SVGTexts(refData)
		actTexts, err2 := SVGTexts(actData)
		if err := errors.Join(err1, err2); err != nil {
			res.fail(format, prefix+err.Error())
			return
		}
		if err := CompareTexts(refTexts, actTexts); err != nil {
			res.fail(format, prefix+err.Error())
		}
		if HasSVGSourceMetadata(refData) != HasSVGSourceMetadata(actData) {
			res.fail(format, prefix+"plantuml-src metadata presence differs from the reference")
		}
		refW, refH := svgSizeOr(refData, ref)
		actW, actH := svgSizeOr(actData, act)
		if err := CompareSize(refW, refH, actW, actH, SizeTolerance); err != nil {
			res.fail(format, prefix+err.Error())
			return
		}
		if r.Rasterizer == nil || refW == 0 || refH == 0 {
			return
		}
		refImg, err1 := r.Rasterizer.Rasterize(refData, refW, refH)
		actImg, err2 := r.Rasterizer.Rasterize(actData, refW, refH)
		if err := errors.Join(err1, err2); err != nil {
			res.fail(format, prefix+err.Error())
			return
		}
		if _, err := CompareImages(refImg, actImg, MinSSIM); err != nil {
			res.fail(format, prefix+err.Error())
		}
	}
}

func (r *Runner) comparePNG(res *CaseResult, id, prefix string, ref *BlockMeta) {
	for page := range ref.Pages {
		name := fmt.Sprintf("block%d-page%d.png", ref.Index, page)
		refData, err1 := os.ReadFile(filepath.Join(r.PNGReference, filepath.FromSlash(id), name))
		actData, err2 := os.ReadFile(filepath.Join(r.PNGActual, filepath.FromSlash(id), name))
		if err := errors.Join(err1, err2); err != nil {
			res.fail("png", prefix+err.Error())
			continue
		}
		if HasPNGSourceMetadata(refData) != HasPNGSourceMetadata(actData) {
			res.fail("png", prefix+name+": plantuml metadata presence differs from the reference")
		}
		refImg, err1 := png.Decode(bytes.NewReader(refData))
		actImg, err2 := png.Decode(bytes.NewReader(actData))
		if err := errors.Join(err1, err2); err != nil {
			res.fail("png", prefix+name+": "+err.Error())
			continue
		}
		rb, ab := refImg.Bounds(), actImg.Bounds()
		if err := CompareSize(rb.Dx(), rb.Dy(), ab.Dx(), ab.Dy(), SizeTolerance); err != nil {
			res.fail("png", prefix+name+": "+err.Error())
			continue
		}
		if _, err := CompareImages(refImg, actImg, MinSSIM); err != nil {
			res.fail("png", prefix+name+": "+err.Error())
		}
	}
}

// svgSizeOr returns the size declared by the SVG document, or the recorded
// size when the document declares none.
func svgSizeOr(svg []byte, meta RenderMeta) (int, int) {
	if w, h, ok := SVGSize(svg); ok {
		return w, h
	}
	return meta.Width, meta.Height
}

// CompareCLI compares one CLI case.
func (r *Runner) CompareCLI(c *CLICase) (res CaseResult) {
	res = CaseResult{ID: c.ID, Group: "cli", Formats: []string{"cli"}}
	defer r.markKnown(&res)
	refDir, actDir := r.Reference(c.ID), filepath.Join(r.Actual, filepath.FromSlash(c.ID))
	var ref, act CLIRecord
	if err := readJSON(filepath.Join(refDir, "cli.json"), &ref); err != nil {
		res.fail("", "reading reference: "+err.Error())
		return res
	}
	if err := readJSON(filepath.Join(actDir, "cli.json"), &act); err != nil {
		res.fail("", "reading output: "+err.Error())
		return res
	}
	for _, what := range c.Compare {
		switch what {
		case "exit":
			if ref.Exit != act.Exit {
				res.fail("cli", fmt.Sprintf("exit status %d, reference %d", act.Exit, ref.Exit))
			}
		case "stdout":
			if err := CompareText(ref.Stdout, act.Stdout); err != nil {
				res.fail("cli", "stdout: "+err.Error())
			}
		case "stderr":
			if err := CompareText(ref.Stderr, act.Stderr); err != nil {
				res.fail("cli", "stderr: "+err.Error())
			}
		case "files":
			r.compareCLIFiles(&res, refDir, actDir, ref.Files, act.Files)
		}
	}
	return res
}

func (r *Runner) compareCLIFiles(res *CaseResult, refDir, actDir string, ref, act []FileRecord) {
	paths := func(files []FileRecord) []string {
		var out []string
		for _, f := range files {
			out = append(out, f.Path)
		}
		slices.Sort(out)
		return out
	}
	if rp, ap := paths(ref), paths(act); !slices.Equal(rp, ap) {
		res.fail("cli", fmt.Sprintf("created files %q, reference %q", ap, rp))
		return
	}
	for _, f := range ref {
		if !f.Stored {
			continue
		}
		refData, err1 := os.ReadFile(filepath.Join(refDir, "files", filepath.FromSlash(f.Path)))
		actData, err2 := os.ReadFile(filepath.Join(actDir, "files", filepath.FromSlash(f.Path)))
		if err := errors.Join(err1, err2); err != nil {
			res.fail("cli", f.Path+": "+err.Error())
			continue
		}
		if err := compareBody(path.Ext(f.Path), refData, actData); err != nil {
			res.fail("cli", f.Path+": "+err.Error())
		}
	}
}

// compareBody compares stored outputs by kind: SVG by texts, others as text.
func compareBody(ext string, ref, act []byte) error {
	if ext == ".svg" {
		rt, err1 := SVGTexts(ref)
		at, err2 := SVGTexts(act)
		if err := errors.Join(err1, err2); err != nil {
			return err
		}
		return CompareTexts(rt, at)
	}
	return CompareText(string(ref), string(act))
}

// CompareHTTP compares one HTTP case.
func (r *Runner) CompareHTTP(c *HTTPCase) (res CaseResult) {
	res = CaseResult{ID: c.ID, Group: "http", Formats: []string{"http"}}
	defer r.markKnown(&res)
	refDir, actDir := r.Reference(c.ID), filepath.Join(r.Actual, filepath.FromSlash(c.ID))
	var ref, act HTTPRecord
	if err := readJSON(filepath.Join(refDir, "http.json"), &ref); err != nil {
		res.fail("", "reading reference: "+err.Error())
		return res
	}
	if err := readJSON(filepath.Join(actDir, "http.json"), &act); err != nil {
		res.fail("", "reading output: "+err.Error())
		return res
	}
	for _, what := range c.Compare {
		switch what {
		case "status":
			if ref.Status != act.Status {
				res.fail("http", fmt.Sprintf("status %d, reference %d", act.Status, ref.Status))
			}
		case "headers":
			compareHeaders(&res, ref.Headers, act.Headers)
		case "body":
			if ref.Body.Stored != act.Body.Stored {
				res.fail("http", "body kind differs from the reference")
				continue
			}
			if !ref.Body.Stored {
				continue // binary bodies are compared as images by format cases
			}
			refData, err1 := os.ReadFile(filepath.Join(refDir, ref.Body.Path))
			actData, err2 := os.ReadFile(filepath.Join(actDir, act.Body.Path))
			if err := errors.Join(err1, err2); err != nil {
				res.fail("http", "body: "+err.Error())
				continue
			}
			if err := compareBody(path.Ext(ref.Body.Path), refData, actData); err != nil {
				res.fail("http", "body: "+err.Error())
			}
		}
	}
	return res
}

// sizeHeaders are compared within the size tolerance instead of exactly.
var sizeHeaders = []string{"X-Plantuml-Diagram-Width", "X-Plantuml-Diagram-Height"}

func compareHeaders(res *CaseResult, ref, act map[string]string) {
	for _, k := range slices.Sorted(maps.Keys(ref)) {
		rv, av := ref[k], act[k]
		if slices.Contains(sizeHeaders, k) {
			rn, err1 := strconv.Atoi(rv)
			an, err2 := strconv.Atoi(av)
			if err1 == nil && err2 == nil && float64(abs(rn-an)) <= SizeTolerance*float64(rn) {
				continue
			}
		}
		if rv != av {
			res.fail("http", fmt.Sprintf("header %s %q, reference %q", k, av, rv))
		}
	}
}

func abs(v int) int { return max(v, -v) }

// markKnown attaches the known difference of the case, if any. It is
// deferred with the named result, so it applies to every return path.
func (r *Runner) markKnown(res *CaseResult) {
	if k, ok := r.Known[res.ID]; ok {
		res.Known = &k
	}
}

// groupOf is the diagram type used to group a case in the report.
func groupOf(c *Case, ref *Meta) string {
	switch {
	case ref.Crash != "":
		return "crash"
	case len(ref.Blocks) == 0:
		return "none"
	}
	t := strings.ToLower(ref.Blocks[0].Type)
	if t == "error" && c.Type != "" {
		return c.Type
	}
	return t
}

func formatOf(file string) string {
	switch path.Ext(file) {
	case ".atxt":
		return "txt"
	default:
		return strings.TrimPrefix(path.Ext(file), ".")
	}
}

// Summary counts results of one group or format.
type Summary struct {
	Name                                      string
	Cases, Passed, Known, Failed, Unsupported int
}

// Rate is the pass rate of the supported cases, counting known
// differences as passing.
func (s Summary) Rate() float64 {
	if s.Cases == s.Unsupported {
		return 1
	}
	return float64(s.Passed+s.Known) / float64(s.Cases-s.Unsupported)
}

// Summarize counts results per group and per format, sorted by name.
func Summarize(results []CaseResult) (groups, formats []Summary) {
	g, f := map[string]*Summary{}, map[string]*Summary{}
	add := func(m map[string]*Summary, name string, passed, known, unsupported bool) {
		s := m[name]
		if s == nil {
			s = &Summary{Name: name}
			m[name] = s
		}
		s.Cases++
		switch {
		case passed && unsupported:
			s.Unsupported++
		case passed:
			s.Passed++
		case known:
			s.Known++
		default:
			s.Failed++
		}
	}
	for i := range results {
		r := &results[i]
		unsupported := len(r.Unsupported) > 0
		add(g, r.Group, r.Passed(), r.Known != nil, unsupported)
		for _, format := range r.Formats {
			add(f, format, r.FormatPassed(format), r.Known != nil, unsupported)
		}
	}
	sorted := func(m map[string]*Summary) []Summary {
		var out []Summary
		for _, s := range m {
			out = append(out, *s)
		}
		slices.SortFunc(out, func(a, b Summary) int { return strings.Compare(a.Name, b.Name) })
		return out
	}
	return sorted(g), sorted(f)
}

// WriteReport writes the Markdown conformance report.
func WriteReport(w io.Writer, implementation string, results []CaseResult) error {
	groups, formats := Summarize(results)
	var b strings.Builder
	fmt.Fprintf(&b, "# Conformance report\n\n")
	fmt.Fprintf(&b, "- Implementation under test: %s\n- Reference: PlantUML 1.2026.8 (MIT variant)\n- Cases: %d\n\n", implementation, len(results))
	table := func(title string, rows []Summary) {
		fmt.Fprintf(&b, "## %s\n\n| Name | Cases | Passed | Known differences | Failed | Not supported yet | Pass rate |\n|---|---|---|---|---|---|---|\n", title)
		for _, s := range rows {
			fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %.1f%% |\n", s.Name, s.Cases, s.Passed, s.Known, s.Failed, s.Unsupported, s.Rate()*100)
		}
		b.WriteString("\n")
	}
	table("By diagram type", groups)
	table("By format", formats)

	var failed, known, knownPass, unsupported []*CaseResult
	for i := range results {
		switch results[i].Status() {
		case StatusUnsupported:
			unsupported = append(unsupported, &results[i])
		case StatusFail:
			failed = append(failed, &results[i])
		case StatusKnown:
			known = append(known, &results[i])
		case StatusKnownPass:
			knownPass = append(knownPass, &results[i])
		}
	}
	fmt.Fprintf(&b, "## Failures (%d)\n\n", len(failed))
	for _, r := range failed {
		fmt.Fprintf(&b, "- `%s`: %s\n", r.ID, strings.Join(allFailures(r), "; "))
	}
	fmt.Fprintf(&b, "\n## Not supported yet (%d)\n\nCases of diagram types this implementation reports as not supported yet, instead of drawing them:\n\n", len(unsupported))
	for _, r := range unsupported {
		fmt.Fprintf(&b, "- `%s`: %s\n", r.ID, strings.Join(slices.Compact(slices.Sorted(slices.Values(r.Unsupported))), ", "))
	}
	fmt.Fprintf(&b, "\n## Known differences (%d)\n\n", len(known))
	for _, r := range known {
		fmt.Fprintf(&b, "- `%s`: %s", r.ID, r.Known.Reason)
		if r.Known.ResolvedBy != "" {
			fmt.Fprintf(&b, " (resolved by %s)", r.Known.ResolvedBy)
		}
		b.WriteString("\n")
	}
	if len(knownPass) > 0 {
		fmt.Fprintf(&b, "\n## Known differences that now pass (%d)\n\nRemove these entries from known-differences.yaml:\n\n", len(knownPass))
		for _, r := range knownPass {
			fmt.Fprintf(&b, "- `%s`\n", r.ID)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func allFailures(r *CaseResult) []string {
	out := slices.Clone(r.Failures)
	for _, f := range slices.Sorted(maps.Keys(r.FormatFailures)) {
		for _, msg := range r.FormatFailures[f] {
			out = append(out, f+": "+msg)
		}
	}
	return out
}
