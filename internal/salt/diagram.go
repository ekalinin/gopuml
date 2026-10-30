// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/PSystemSalt.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/jfmt"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/text"
	"github.com/ekalinin/gopuml/internal/uml"
)

// PSystemSalt is upstream's PSystemSalt: the lines of a salt diagram, kept
// for the drawing, which turns them into a tree of widgets.
type PSystemSalt struct {
	uml.Base

	data       []string
	dictionary *saltDictionary
}

// NewPSystemSalt returns an empty salt diagram for src; md5map holds the
// base64 images of the source.
func NewPSystemSalt(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *PSystemSalt {
	d := &PSystemSalt{Base: uml.NewBase(ctx, src, core.Salt, md5map)}
	d.dictionary = newSaltDictionary(ctx, d.Skin().Options())
	return d
}

var (
	_ core.TypedDiagram = (*PSystemSalt)(nil)
	_ uml.SpriteAdder   = (*PSystemSalt)(nil)
)

// Description implements core.Diagram: "(Salt)".
func (d *PSystemSalt) Description() string { return "(Salt)" }

// DefaultMargins are the margins of salt diagrams: 5 on each side.
func (d *PSystemSalt) DefaultMargins() klimt.ClockwiseTopRightBottomLeft {
	return klimt.ClockwiseSame(5)
}

// AddSprite keeps a sprite for the widgets only, as upstream's override of
// addSprite: the annotations do not see it.
func (d *PSystemSalt) AddSprite(name string, sprite klimt.Sprite) {
	d.dictionary.addSprite(name, sprite)
}

// add keeps a line, cut at the newlines of the preprocessor with Java's
// split.
func (d *PSystemSalt) add(s string) {
	d.data = append(d.data, klimt.JavaSplit(s, string(klimt.JawsBlockE1Newline))...)
}

func (d *PSystemSalt) size() int { return len(d.data) }

// manageSprite returns the lines of the widgets: it drops "hide stereotype"
// and the skin parameters, sets the scale, and reads the sprites. The common
// commands read most of these lines before they get here.
func (d *PSystemSalt) manageSprite() []string {
	var result []string
	for i := 0; i < len(d.data); i++ {
		s := d.data[i]
		switch {
		case s == "hide stereotype":
		case strings.HasPrefix(s, "skinparam "):
		case strings.HasPrefix(s, "scale "):
			scale, err := jfmt.ParseDouble(s[len("scale "):])
			if err != nil {
				uml.Throw(uml.NumberFormatException, err.Error())
			}
			d.SetScale(&uml.Scale{Kind: uml.ScaleSimple, A: scale})
		case strings.HasPrefix(s, "sprite $"):
			block := command.SingleString(s)
			for {
				i++
				if i >= len(d.data) {
					uml.Throw("java.util.NoSuchElementException", "")
				}
				s = d.data[i]
				block = block.Add(text.NewStringLocated(s, nil))
				if s == "}" {
					break
				}
			}
			uml.CommandFactorySpriteMultiLine[*PSystemSalt]().Execute(d, block, command.One)
		default:
			result = append(result, s)
		}
	}
	return result
}
