// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/Run.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/ExitStatus.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/StdrptNull.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/SourceFileReaderAbstract.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/PSystemUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/dot"
	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/server"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/tim"
	"github.com/ekalinin/gopuml/internal/version"
)

// Image statuses of upstream's FileImageData.
const (
	statusError = 400
	statusCrash = 503
)

// isErrorStatus is ExitStatus.isErrorStatus: 0 and 200 mean success.
func isErrorStatus(status int) bool { return status != 0 && status != 200 }

// exitStatus is upstream's ExitStatus. Files are processed in parallel, so
// the flags are atomic.
type exitStatus struct {
	hasFiles, hasErrors, hasBlocks atomic.Bool
}

// code is ExitStatus.getExitCode: errors take precedence over a missing
// file, which takes precedence over a missing diagram.
func (e *exitStatus) code() int {
	switch {
	case e.hasErrors.Load():
		return ExitSomeErrors
	case !e.hasFiles.Load():
		return ExitNoFileFound
	case !e.hasBlocks.Load():
		return ExitNoDiagram
	}
	return ExitOK
}

// update is SourceFileReaderAbstract.updateStatus.
func (e *exitStatus) update(blocks []Block) {
	e.hasFiles.Store(true)
	for _, b := range blocks {
		e.hasBlocks.Store(true)
		if b.Error() != nil {
			e.hasErrors.Store(true)
		}
	}
}

// finalMessage is StdrptNull.finalMessage.
func (e *exitStatus) finalMessage(w io.Writer) {
	if e.hasErrors.Load() {
		fmt.Fprintln(w, "Some diagram description contains errors")
	}
	if !e.hasBlocks.Load() {
		fmt.Fprintln(w, "No diagram found")
	}
}

// Command is the command line with its standard streams.
type Command struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	// Engine builds and renders diagrams. Without it, the commands that
	// need diagrams report that rendering is not implemented.
	Engine Engine
	// Getenv reads environment variables; nil means os.Getenv.
	Getenv func(string) string
}

// errNotImplemented is reported until the command has an engine.
var errNotImplemented = errors.New("rendering is not implemented yet")

// Run runs the command line and returns the exit status, following
// upstream's Run.main.
func (c *Command) Run(args []string) int {
	start := time.Now()
	if len(args) == 0 {
		// Upstream opens the GUI here; this implementation prints the usage.
		PrintHelp(c.Stdout, Prog, false)
		return ExitOK
	}
	p, err := Parse(args)
	if err != nil {
		fmt.Fprintln(c.Stderr, err)
		return ExitParseError
	}
	getenv := c.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	config, err := buildConfig(p, getenv)
	if err != nil {
		fmt.Fprintln(c.Stderr, err)
		return ExitParseError
	}
	if err := p.CheckSupported(); err != nil {
		fmt.Fprintln(c.Stderr, "plantuml:", err)
		return ExitUnsupported
	}
	if f, ok := p.ImmediateAction(); ok {
		switch f {
		case Help:
			PrintHelp(c.Stdout, Prog, false)
			return ExitOK
		case HelpMore:
			PrintHelp(c.Stdout, Prog, true)
			return ExitOK
		case Version:
			fmt.Fprintln(c.Stdout, version.Full())
			return ExitOK
		case License:
			for _, line := range version.LicenseText() {
				fmt.Fprintln(c.Stdout, line)
			}
			return ExitOK
		case TestDot:
			return c.checkGraphviz(p, getenv)
		case StdLib:
			for _, line := range tim.PrintStdLib() {
				fmt.Fprintln(c.Stdout, line)
			}
			return ExitOK
		case Language:
			if err := server.WriteLanguage(c.Stdout); err != nil {
				fmt.Fprintln(c.Stderr, "plantuml:", err)
				return ExitUnsupported
			}
			return ExitOK
		}
		return c.notImplemented()
	}
	// Upstream fails on an unknown charset only when it reads a file, and
	// then silently; this implementation reports it first.
	name, _ := p.String(Charset)
	charset, ok := text.LookupCharset(name)
	if !ok {
		fmt.Fprintf(c.Stderr, "plantuml: unsupported charset %q\n", name)
		return ExitUnsupported
	}
	r := &run{Command: c, p: p, exit: &exitStatus{}, format: p.OutputFormat(), config: config, charset: charset,
		dot: dotConfig(p, getenv)}
	// duration is upstream's finally block; it does not run when upstream
	// calls System.exit.
	duration := func(code int) int {
		if p.IsTrue(Duration) {
			fmt.Fprintf(c.Stderr, "Duration = %s seconds\n", jfmt.DoubleToString(float64(time.Since(start).Milliseconds())/1000))
		}
		return code
	}
	switch {
	case p.IsTrue(Picoweb):
		return r.serve()
	case p.IsTrue(Pipe) || p.IsTrue(Syntax):
		if c.Engine == nil {
			return c.notImplemented()
		}
		r.pipe()
		if code := r.exit.code(); code != ExitOK {
			return code
		}
		return duration(ExitOK)
	case p.IsTrue(DecodeUrl):
		return duration(r.decodeURLs())
	case p.IsTrue(RetrieveMetadata):
		r.extractMetadata(p.InputFiles())
		return duration(ExitOK)
	}
	files := p.InputFiles()
	if len(files) == 0 {
		fmt.Fprintln(c.Stderr, "No file found")
		return ExitNoFileFound
	}
	if c.Engine == nil {
		return c.notImplemented()
	}
	switch {
	case p.IsTrue(ComputeUrl):
		r.computeURL(files)
		return duration(ExitOK)
	case p.IsTrue(FailFast2) && r.checkError(files):
	default:
		r.processInputs(files)
	}
	duration(0)
	r.exit.finalMessage(c.Stderr)
	return r.exit.code()
}

// dotConfig is the Graphviz setting of --dot-path and --graphviz-timeout.
// Upstream ignores a timeout that is not a number of seconds.
func dotConfig(p *Parsed, getenv func(string) string) dot.Config {
	c := dot.Config{Host: dot.Host{Getenv: getenv}}
	if v, ok := p.String(GraphvizDot); ok {
		c.DotPath = text.RemoveQuotes(v)
	}
	if v, ok := p.String(Timeout); ok && isDigits(v) {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			c.Timeout = time.Duration(n) * time.Second
		}
	}
	return c
}

func isDigits(s string) bool {
	return s != "" && strings.Trim(s, "0123456789") == ""
}

// checkGraphviz is OptionPrint.printCheckGraphviz: the report goes to
// standard output on success and to standard error otherwise. Upstream
// runs it before reading --dot-path; this implementation uses the flag.
func (c *Command) checkGraphviz(p *Parsed, getenv func(string) string) int {
	insecure := security.ConfigFromEnv(getenv, security.Legacy).Profile == security.Insecure
	lines, code := dot.CheckGraphviz(context.Background(), dotConfig(p, getenv), insecure)
	w := c.Stdout
	if code != ExitOK {
		w = c.Stderr
	}
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return code
}

func (c *Command) notImplemented() int {
	fmt.Fprintln(c.Stderr, "plantuml:", errNotImplemented)
	return ExitUnsupported
}

// run holds the state of one command line run.
type run struct {
	*Command
	p    *Parsed
	exit *exitStatus
	// format is upstream's output format name; pipe mode can change it.
	format string
	// config lines go before each diagram; charset decodes the inputs.
	config  []string
	charset text.Charset
	// dot is the Graphviz setting.
	dot dot.Config
}

// threads is CliOptions.getNbThreads. Upstream fails on a value below 1;
// this implementation uses one thread.
func (r *run) threads() int {
	n := runtime.NumCPU()
	if v, ok := r.p.String(NbThread); ok && !strings.EqualFold(v, "auto") {
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			n = int(i)
		}
	}
	return max(n, 1)
}

// failFast reports --stop-on-error or --check-before-run.
func (r *run) failFast() bool { return r.p.IsTrue(FailFast) || r.p.IsTrue(FailFast2) }

func (r *run) renderOptions() RenderOptions {
	return RenderOptions{Format: r.format, NoMetadata: r.p.IsTrue(NoMetadata), DarkMode: r.p.IsTrue(DarkMode)}
}

// outputDir is CliOptions.getOutputDir: the -o value without enclosing
// quotes.
func (r *run) outputDir() string {
	out, _ := r.p.String(OutputDir)
	return text.RemoveQuotes(out)
}

func (r *run) defines() []KeyValue { return r.p.Map(Define, DefineLong) }

// decodeURLs prints each argument decoded, wrapped in @startuml and
// @enduml. Upstream ends with a stack trace on text it cannot decode;
// this implementation prints the error with exit status 1, the status of
// that stack trace.
func (r *run) decodeURLs() int {
	for _, s := range r.p.Args() {
		decoded, err := code.Decode(s)
		if err != nil {
			fmt.Fprintln(r.Stderr, "plantuml:", err)
			return ExitUnsupported
		}
		fmt.Fprintf(r.Stdout, "@startuml\n%s\n@enduml\n", decoded)
	}
	return ExitOK
}

// read reads a source file and builds its diagrams, as upstream's
// SourceFileReader, which also creates the output directory. The !log
// lines of the preprocessor go to stderr.
func (r *run) read(file string, stderr io.Writer) ([]Block, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	os.MkdirAll(ResolveOutputDir(file, r.outputDir()), 0o755)
	blocks, err := r.Engine.Read(&Source{Text: r.charset.Decode(data), File: file, Defines: r.defines(), Config: r.config,
		DotPath: r.dot.DotPath, DotTimeout: r.dot.Timeout})
	printLog(stderr, blocks)
	return blocks, err
}

// printLog prints the !log lines of the blocks, as Log.userLog.
func printLog(w io.Writer, blocks []Block) {
	for _, b := range blocks {
		for _, line := range b.Log() {
			fmt.Fprintln(w, line)
		}
	}
}

func hasError(blocks []Block) bool {
	for _, b := range blocks {
		if b.Error() != nil {
			return true
		}
	}
	return false
}

// forEachFile runs task for each file on the given number of goroutines.
// What a task writes to its stderr is printed in file order, so the output
// does not depend on scheduling. A panic is reported as an error of that
// file instead of ending the program.
func (r *run) forEachFile(files []string, threads int, task func(file string, stderr io.Writer)) {
	outs := make([]bytes.Buffer, len(files))
	done := make([]chan struct{}, len(files))
	for i := range done {
		done[i] = make(chan struct{})
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(threads, len(files)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				func() {
					defer close(done[i])
					defer func() {
						if v := recover(); v != nil {
							r.exit.hasErrors.Store(true)
							fmt.Fprintf(&outs[i], "plantuml: %s: %v\n", files[i], v)
						}
					}()
					task(files[i], &outs[i])
				}()
			}
		}()
	}
	go func() {
		for i := range files {
			jobs <- i
		}
		close(jobs)
	}()
	for i := range files {
		<-done[i]
		r.Stderr.Write(outs[i].Bytes())
	}
	wg.Wait()
}

// checkError is Run.checkError: it reads every input and reports whether a
// diagram has an error.
func (r *run) checkError(files []string) bool {
	var found atomic.Bool
	r.forEachFile(files, r.threads(), func(file string, stderr io.Writer) {
		if found.Load() {
			return
		}
		blocks, err := r.read(file, stderr)
		r.exit.hasFiles.Store(true)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return
		}
		if hasError(blocks) {
			found.Store(true)
			r.exit.hasErrors.Store(true)
		}
	})
	return found.Load()
}

// computeURL is Run.computeUrl.
func (r *run) computeURL(files []string) {
	for _, file := range files {
		blocks, err := r.read(file, r.Stderr)
		if err != nil {
			fmt.Fprintln(r.Stderr, err)
			continue
		}
		for _, b := range blocks {
			fmt.Fprintln(r.Stdout, b.EncodedURL())
		}
	}
}

// processInputs is Run.processInputsInParallel. Upstream checks for an
// earlier error before each file, which only stops at the first error when
// files run one at a time; this implementation then uses one thread.
func (r *run) processInputs(files []string) {
	threads := r.threads()
	if r.failFast() {
		threads = 1
	}
	r.forEachFile(files, threads, func(file string, stderr io.Writer) {
		r.exit.hasFiles.Store(true)
		if r.exit.hasErrors.Load() && r.failFast() {
			return
		}
		r.manageFile(file, stderr)
	})
}

// generatedImage is upstream's GeneratedImage.
type generatedImage struct {
	file   string // empty when no file is written
	status int
	block  Block
}

// lineErrorRaw is GeneratedImageImpl.lineErrorRaw: the 0-based error line
// of a syntax error, or -1.
func (g generatedImage) lineErrorRaw() int {
	if e := g.block.Error(); e != nil && !e.Unsupported {
		return e.Line
	}
	return -1
}

// manageFile is Run.manageFileInternal.
func (r *run) manageFile(file string, stderr io.Writer) {
	blocks, err := r.read(file, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return
	}
	r.exit.update(blocks)
	if r.p.IsTrue(CheckOnly) {
		return
	}
	if r.format == "PREPROC" {
		r.extractPreprocessingSource(file, blocks, stderr)
		return
	}
	images := r.generatedImages(file, blocks, stderr)
	if len(images) == 0 {
		fmt.Fprintln(stderr, "Warning: no image in "+file)
		return
	}
	for _, im := range images {
		if isErrorStatus(im.status) {
			r.exit.hasErrors.Store(true)
		}
	}
	for _, im := range images {
		if line := im.lineErrorRaw(); line != -1 {
			fmt.Fprintf(stderr, "Error line %d in file: %s\n", line+1, file)
			return
		}
	}
}

// namer names the outputs of a source file, as SourceFileReader.
func (r *run) namer(file, suffix string) *OutputNamer {
	sep := "_"
	if r.p.IsTrue(UseSeparatorMinus) {
		sep = "-"
	}
	return &OutputNamer{
		Source:      file,
		OutputDir:   ResolveOutputDir(file, r.outputDir()),
		Suffix:      suffix,
		Separator:   sep,
		IgnoreNames: r.p.IsTrue(IgnoreStartumlFilename),
	}
}

// generatedImages is SourceFileReaderAbstract.getGeneratedImages with
// PSystemUtils.exportDiagrams: each page of each block goes to its file.
func (r *run) generatedImages(file string, blocks []Block, stderr io.Writer) []generatedImage {
	namer := r.namer(file, FormatSuffix(r.format))
	var result []generatedImage
	for _, b := range blocks {
		name, hasName := b.Name()
		if b.Error() != nil && r.p.IsTrue(NoErrorImage) {
			namer.Files(name, hasName, 1)
			result = append(result, generatedImage{status: statusError, block: b})
			continue
		}
		s := namer.Suggest(name, hasName)
		// --skip-fresh (PSystemUtils.exportDiagrams): an existing output that
		// embeds the same source is kept, as one image.
		if r.p.IsTrue(CheckMetadata) {
			if meta, ok := b.Metadata(); ok && r.sameMetadata(s.File(0), meta) {
				result = append(result, generatedImage{file: s.File(0), block: b})
				continue
			}
		}
		pages := max(b.Pages(), 1)
		namer.AddPages(pages)
		for page := range pages {
			f := s.File(page)
			if !r.canFileBeWritten(f, stderr) {
				break
			}
			status, err := r.writeImage(f, b, page)
			if err != nil {
				fmt.Fprintf(stderr, "plantuml: %s: %v\n", file, err)
				if !isErrorStatus(status) {
					status = statusCrash
				}
			}
			result = append(result, generatedImage{file: f, status: status, block: b})
		}
	}
	return result
}

// canFileBeWritten is PSystemUtils.canFileBeWritten: a read-only file is
// replaced only with --overwrite.
func (r *run) canFileBeWritten(file string, stderr io.Writer) bool {
	st, err := os.Stat(file)
	if err != nil {
		return true
	}
	if f, err := os.OpenFile(file, os.O_WRONLY, 0); err == nil {
		f.Close()
		return true
	}
	if r.p.IsTrue(Overwrite) {
		os.Chmod(file, st.Mode().Perm()|0o200)
		os.Remove(file)
		return true
	}
	abs, _ := filepath.Abs(file)
	fmt.Fprintln(stderr, "Cannot write to file "+abs)
	return false
}

func (r *run) writeImage(file string, b Block, page int) (int, error) {
	f, err := os.Create(file)
	if err != nil {
		return 0, err
	}
	w := bufio.NewWriter(f)
	status, err := b.Render(w, page, r.renderOptions())
	if ferr := w.Flush(); err == nil {
		err = ferr
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	// An error with an error status comes with the image that shows it (for
	// example Graphviz missing or failing): keep it. Otherwise do not leave
	// an empty or partial image behind.
	if err != nil && !isErrorStatus(status) {
		os.Remove(file)
	}
	return status, err
}

// extractPreprocessingSource is Run.extractPreprocessingSource for the
// PREPROC format: each block's preprocessed lines go to its .preproc file.
func (r *run) extractPreprocessingSource(file string, blocks []Block, stderr io.Writer) {
	namer := r.namer(file, FormatSuffix("PREPROC"))
	for _, b := range blocks {
		name, hasName := b.Name()
		var sb strings.Builder
		for _, line := range b.Preprocessed() {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
		if err := os.WriteFile(namer.Files(name, hasName, 1)[0], []byte(sb.String()), 0o644); err != nil {
			fmt.Fprintln(stderr, err)
		}
	}
}
