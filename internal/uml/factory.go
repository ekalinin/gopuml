// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/core/UmlSource.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/PSystemBuilder.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package uml

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/text"
)

// base64TagStart is upstream's AtomImg.DATA_IMAGE_PNG_BASE64.
const base64TagStart = "data:image/png;base64,"

// base64TagReplacement replaces base64TagStart, followed by the MD5 digest
// of the data.
const base64TagReplacement = "data:image/png;md5,"

// PatchBase64 is upstream's UmlSource.patchBase64, which PSystemBuilder
// runs before the parsers: inline base64 PNG images
// ("data:image/png;base64,XXXX") become "data:image/png;md5,YYYY", YYYY being
// the MD5 digest of XXXX, so that command lines stay short. It returns the
// patched source (src itself when nothing changes) and the data by digest.
func PatchBase64(src *core.UmlSource) (*core.UmlSource, map[string]string) {
	md5map := map[string]string{}
	lines := src.Lines()
	var patched []text.StringLocated
	for i, l := range lines {
		s, changed := patchBase64Line(l.S, md5map)
		if !changed {
			continue
		}
		if patched == nil {
			patched = append([]text.StringLocated(nil), lines...)
		}
		patched[i] = text.StringLocated{S: s, Location: l.Location, PreprocessorError: l.PreprocessorError}
	}
	if patched == nil {
		return src, md5map
	}
	return core.NewUmlSource(patched, src.Raw(), false), md5map
}

func patchBase64Line(line string, md5map map[string]string) (string, bool) {
	from := 0
	var sb strings.Builder
	changed := false
	for {
		idx := strings.Index(line[from:], base64TagStart)
		if idx == -1 {
			break
		}
		start := from + idx
		dataStart := start + len(base64TagStart)
		end := dataStart
		for end < len(line) && isBase64Char(line[end]) {
			end++
		}
		data := line[dataStart:end]
		sum := md5.Sum([]byte(data))
		digest := hex.EncodeToString(sum[:])
		md5map[digest] = data
		sb.WriteString(line[from:start])
		sb.WriteString(base64TagReplacement)
		sb.WriteString(digest)
		from = end
		changed = true
	}
	if !changed {
		return line, false
	}
	sb.WriteString(line[from:])
	return sb.String(), true
}

func isBase64Char(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '+' || c == '/' || c == '='
}

// Factory is a command factory of a UML diagram type, as upstream's
// PSystemCommandFactory subclasses with PSystemBuilder's base64 patching.
// The command list is built once and shared by every parse; the diagram is
// created for each parse with the images of its source.
type Factory[D command.Diagram] struct {
	typ        core.DiagramType
	commands   func() []command.Command[D]
	newDiagram func(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) D

	once sync.Once
	cmds []command.Command[D]
}

// NewFactory returns the factory of diagrams of type typ, with the commands
// built by commands in upstream's order.
func NewFactory[D command.Diagram](typ core.DiagramType, commands func() []command.Command[D],
	newDiagram func(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) D) *Factory[D] {
	return &Factory[D]{typ: typ, commands: commands, newDiagram: newDiagram}
}

// DiagramType implements core.Factory.
func (f *Factory[D]) DiagramType() core.DiagramType { return f.typ }

// Commands returns the command list.
func (f *Factory[D]) Commands() []command.Command[D] {
	f.once.Do(func() { f.cmds = f.commands() })
	return f.cmds
}

// CreateSystem implements core.Factory.
func (f *Factory[D]) CreateSystem(ctx *core.RenderContext, src *core.UmlSource) core.Diagram {
	patched, md5map := PatchBase64(src)
	inner := &command.Factory[D]{
		Type:     f.typ,
		Commands: f.Commands,
		NewDiagram: func(ctx *core.RenderContext, src *core.UmlSource) D {
			return f.newDiagram(ctx, src, md5map)
		},
	}
	return inner.CreateSystem(ctx, patched)
}
