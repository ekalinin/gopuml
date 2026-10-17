// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/Pipe.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/StdrptPipe0.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/SourceStringReader.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// javaLineReader reads lines like Java's BufferedReader.readLine: a line
// ends with "\n", "\r" or "\r\n".
type javaLineReader struct{ r *bufio.Reader }

func (l *javaLineReader) readLine() (string, bool) {
	var b []byte
	for {
		c, err := l.r.ReadByte()
		if err != nil {
			return string(b), len(b) > 0
		}
		switch c {
		case '\n':
			return string(b), true
		case '\r':
			if next, err := l.r.Peek(1); err == nil && next[0] == '\n' {
				l.r.ReadByte()
			}
			return string(b), true
		}
		b = append(b, c)
	}
}

// pipeState is Pipe.State.
type pipeState int

const (
	noContent pipeState = iota
	startMarkNotFound
	startMarkFound
	complete
)

var startMark = regexp.MustCompile(`^@start([A-Za-z]*)$`)

// javaTrimEmpty reports whether String.trim() of s is empty.
func javaTrimEmpty(s string) bool {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' }) == ""
}

// readSingleDiagram is Pipe.readSingleDiagram: it returns the next diagram
// of the input. Only the first diagram may lack a start line; it is then
// wrapped in @startuml and @enduml.
func (r *run) readSingleDiagram(in *javaLineReader, unmarkedAllowed bool) (string, bool) {
	state := noContent
	var expectedEnd string
	var sb strings.Builder
	for state != complete {
		line, ok := in.readLine()
		if !ok {
			break
		}
		if strings.HasPrefix(line, "@@@format ") {
			r.manageFormat(line)
			continue
		}
		if state == noContent && !javaTrimEmpty(line) {
			state = startMarkNotFound
		}
		if state == startMarkNotFound && strings.HasPrefix(line, "@start") {
			sb.Reset()
			state = startMarkFound
			if m := startMark.FindStringSubmatch(line); m != nil {
				expectedEnd = "@end" + m[1]
			} else {
				expectedEnd = "@end"
			}
		} else if state == startMarkFound && strings.HasPrefix(line, expectedEnd) {
			state = complete
		}
		if state != noContent {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
	}
	switch state {
	case startMarkNotFound:
		if unmarkedAllowed {
			return "@startuml\n" + sb.String() + "@enduml\n", true
		}
	case startMarkFound:
		return sb.String() + expectedEnd, true
	case complete:
		return sb.String(), true
	}
	return "", false
}

// manageFormat is Pipe.manageFormat: an "@@@format" line changes the
// format of the following diagrams.
func (r *run) manageFormat(s string) {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "png"):
		r.format = "PNG"
	case strings.Contains(s, "svg"):
		r.format = "SVG"
	case strings.Contains(s, "atxt"):
		r.format = "ATXT"
	case strings.Contains(s, "utxt"):
		r.format = "UTXT"
	}
}

// imageIndex is CliOptions.getImageIndex. A negative index, which makes
// upstream fail, selects the first image.
func (r *run) imageIndex() int {
	v, _ := r.p.String(PipeImageIndex)
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil || i < 0 {
		return 0
	}
	return int(i)
}

// pipe is Pipe.managePipe: each diagram of standard input is rendered to
// standard output.
func (r *run) pipe() {
	noStdErr := r.p.IsTrue(Pipenostderr)
	noErrorImage := noStdErr || r.p.IsTrue(NoErrorImage)
	fileDir, _ := r.p.String(FileDir)
	filename, _ := r.p.String(Filename)
	in := &javaLineReader{r: bufio.NewReader(r.charset.NewReader(r.Stdin))}
	for source, ok := r.readSingleDiagram(in, true); ok; source, ok = r.readSingleDiagram(in, false) {
		src := &Source{Text: text.DecodeUTF8([]byte(source)), FileDir: fileDir, Filename: filename, Defines: r.defines(), Config: r.config,
			DotPath: r.dot.DotPath, DotTimeout: r.dot.Timeout}
		blocks, err := r.Engine.Read(src)
		printLog(r.Stderr, blocks)
		if err != nil {
			fmt.Fprintln(r.Stderr, err)
			continue
		}
		r.exit.update(blocks)
		switch {
		case r.p.IsTrue(ComputeUrl):
			for _, b := range blocks {
				fmt.Fprintln(r.Stdout, b.EncodedURL())
			}
		case r.p.IsTrue(Syntax):
			r.syntaxCheckDiagram(blocks)
		default:
			r.generateDiagram(blocks, noStdErr, noErrorImage)
		}
	}
}

// printInfo is StdrptPipe0.printInfo: only a syntax error is reported.
func printInfo(w io.Writer, blocks []Block) {
	if len(blocks) == 0 {
		return
	}
	e := blocks[0].Error()
	if e == nil || e.Unsupported {
		return
	}
	fmt.Fprintln(w, "ERROR")
	fmt.Fprintln(w, e.Line)
	for _, m := range e.Messages {
		fmt.Fprintln(w, m)
	}
}

// syntaxCheckDiagram is Pipe.syntaxCheckDiagram, for -syntax.
func (r *run) syntaxCheckDiagram(blocks []Block) {
	if len(blocks) == 0 {
		return
	}
	b := blocks[0]
	switch {
	case b.Error() != nil:
		r.exit.hasErrors.Store(true)
		printInfo(r.Stdout, blocks)
	case b.Type() != "":
		fmt.Fprintln(r.Stdout, b.Type())
		fmt.Fprintln(r.Stdout, b.Description())
	default:
		fmt.Fprintln(r.Stdout, "OTHER")
		fmt.Fprintln(r.Stdout, b.Description())
	}
}

// outputImage is SourceStringReader.outputImage: image n counts the pages
// of all blocks. It returns the rendered block, or nil.
func (r *run) outputImage(w io.Writer, blocks []Block, n int) Block {
	if len(blocks) == 0 {
		return nil
	}
	if r.format == "PREPROC" {
		for _, line := range blocks[0].Preprocessed() {
			io.WriteString(w, line+"\n")
		}
		return nil
	}
	for _, b := range blocks {
		if pages := b.Pages(); n >= pages {
			n -= pages
			continue
		}
		status, err := b.Render(w, n, r.renderOptions())
		if err != nil {
			fmt.Fprintln(r.Stderr, err)
			status = statusCrash
		}
		if isErrorStatus(status) {
			r.exit.hasErrors.Store(true)
		}
		return b
	}
	fmt.Fprintf(r.Stderr, "numImage is too big = %d\n", n)
	return nil
}

// generateDiagram is Pipe.generateDiagram. With -pipenostderr or
// --no-error-image, the image is buffered and dropped for a failed diagram.
func (r *run) generateDiagram(blocks []Block, noStdErr, noErrorImage bool) {
	buffered := noStdErr || noErrorImage
	var buf bytes.Buffer
	out := r.Stdout
	if buffered {
		out = &buf
	}
	b := r.outputImage(out, blocks, r.imageIndex())
	info := r.Stderr
	if noStdErr {
		info = r.Stdout
	}
	printInfo(info, blocks)
	if buffered && (b == nil || !strings.EqualFold(b.Description(), "(error)")) {
		r.Stdout.Write(buf.Bytes())
	}
	if delimiter, ok := r.p.String(Pipedelimitor); ok {
		fmt.Fprintln(r.Stdout, delimiter)
	}
}
