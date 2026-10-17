// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ekalinin/gopuml/internal/conformance"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/engine"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// goFormats are the formats the Go implementation renders in process, by
// case format name.
var goFormats = map[string]font.FileFormat{"svg": font.FormatSVG, "png": font.FormatPNG, "txt": font.FormatATXT, "utxt": font.FormatUTXT}

// renderDiagramsGo is renderDiagrams for the Go implementation: each case is
// parsed and drawn in process, as tools/reference/ReferenceRender.java does
// with the jar, and preproc outputs are the preprocessed blocks, as the CLI
// writes them with --preproc.
func (g *generator) renderDiagramsGo(jobs []diagramJob) error {
	for _, j := range jobs {
		out := filepath.Join(g.root, j.outRoot, filepath.FromSlash(j.c.ID))
		if err := os.MkdirAll(out, 0o755); err != nil {
			return err
		}
		meta := conformance.Meta{ID: j.c.ID}
		if err := g.renderCaseGo(&meta, j, out); err != nil {
			meta.Crash = err.Error()
		}
		if err := writeJSON(filepath.Join(out, "meta.json"), meta); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) renderCaseGo(meta *conformance.Meta, j diagramJob, out string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()
	file := filepath.Join(g.root, j.file)
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	// The jar's ReferenceRender reads the case through a SourceStringReader,
	// whose line locations are named "string", the context's default.
	rc := core.NewRenderContext()
	rc.CurrentDir = filepath.Dir(file)
	blocks, err := engine.Parse(context.Background(), rc, engine.Source{Text: string(data), File: file})
	if err != nil {
		return err
	}
	meta.Blocks = []conformance.BlockMeta{}
	for i, b := range blocks {
		bm := conformance.BlockMeta{Index: i, Type: "OTHER", Description: b.Diagram.Description(), Pages: b.Pages(), Renders: []conformance.RenderMeta{}}
		if name, ok := b.Name(); ok {
			bm.Name = &name
		}
		switch d := b.Diagram.(type) {
		case core.TypedDiagram:
			bm.Type = d.DiagramType().String()
		case *core.ErrorDiagram:
			bm.Type = "ERROR"
			bm.Error = &conformance.ErrorMeta{Line: d.Position() + 1, Messages: []string{d.Error.Message()}}
		}
		for _, f := range j.c.Formats {
			if f == "preproc" {
				text := strings.Join(b.Source(), "\n") + "\n"
				if err := writeFile(filepath.Join(out, fmt.Sprintf("block%d.preproc", i)), []byte(text)); err != nil {
					return err
				}
				continue
			}
			ff, ok := goFormats[f]
			if !ok || (f == "png" && !g.png) {
				continue
			}
			dir := out
			if f == "png" {
				dir = filepath.Join(g.root, g.outPNG, filepath.FromSlash(j.c.ID))
			}
			for page := range b.Pages() {
				var buf bytes.Buffer
				img, err := b.Render(context.Background(), &buf, page, engine.Options{Format: ff})
				if errors.Is(err, engine.ErrNotDrawable) {
					// No output: the runner reports the missing file.
					continue
				}
				name := fmt.Sprintf("block%d-page%d%s", i, page, ff.FileSuffix())
				rm := conformance.RenderMeta{Width: int(img.Width), Height: int(img.Height), Status: img.Status}
				if err != nil {
					rm.Status = 503
				} else {
					if err := writeFile(filepath.Join(dir, name), buf.Bytes()); err != nil {
						return err
					}
					rm.File = &name
				}
				bm.Renders = append(bm.Renders, rm)
			}
		}
		meta.Blocks = append(meta.Blocks, bm)
	}
	return nil
}

// goBinary builds the command line once.
func (g *generator) goBinary() (string, error) {
	if g.binary != "" {
		return g.binary, nil
	}
	bin := filepath.Join(g.root, ".cache", "plantuml-go")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/plantuml")
	cmd.Dir, cmd.Stdout, cmd.Stderr = g.root, os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("building the command line: %w", err)
	}
	g.binary = bin
	return bin, nil
}

// runCLIGo runs a CLI case with the Go command line in dir.
func (g *generator) runCLIGo(c *conformance.CLICase, dir string, stdout, stderr *bytes.Buffer) (int, error) {
	bin, err := g.goBinary()
	if err != nil {
		return 0, err
	}
	cmd := exec.Command(bin, c.Args...)
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = dir, strings.NewReader(c.Stdin), stdout, stderr
	cmd.Env = append(os.Environ(), "HOME="+os.TempDir())
	for k, v := range c.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	err = cmd.Run()
	if ee, ok := errors.AsType[*exec.ExitError](err); ok {
		return ee.ExitCode(), nil
	}
	return 0, err
}

// startGoServer starts the Go command line's HTTP server on a free port.
func (g *generator) startGoServer() (*server, error) {
	bin, err := g.goBinary()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(bin, "--http-server:0:127.0.0.1")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	s := &server{g: g, cmd: cmd}
	sc := bufio.NewScanner(stderr)
	for sc.Scan() {
		if port, ok := strings.CutPrefix(sc.Text(), "webPort="); ok {
			s.url = "http://127.0.0.1:" + port
			go io.Copy(io.Discard, stderr)
			return s, nil
		}
	}
	s.stop()
	return nil, errors.New("the Go server did not start")
}

func readJSONFile(p string, v any) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
