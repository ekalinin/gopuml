// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/salt/SaltDictionary.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package salt

import (
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
	"github.com/ekalinin/gopuml/internal/klimt/color"
	"github.com/ekalinin/gopuml/internal/klimt/geom"
	"github.com/ekalinin/gopuml/internal/uml"
)

// saltDictionary is the skin of every widget of a salt diagram: no skin
// values, and the pixel images and the sprites of the diagram by name.
type saltDictionary struct {
	data       map[string]element
	sprites    map[string]klimt.Sprite
	option     *klimt.ConfigurationStore[klimt.OptionKey]
	pathSystem *core.RenderContext
}

var _ klimt.ISkinSimple = (*saltDictionary)(nil)

func newSaltDictionary(pathSystem *core.RenderContext,
	option *klimt.ConfigurationStore[klimt.OptionKey]) *saltDictionary {
	return &saltDictionary{data: map[string]element{}, sprites: map[string]klimt.Sprite{}, option: option,
		pathSystem: pathSystem}
}

func (d *saltDictionary) put(name string, e element) { d.data[name] = e }

// get returns the image of a name, wrapped so that it is a new element. An
// unknown name throws, as Objects.requireNonNull.
func (d *saltDictionary) get(name string) element {
	result, ok := d.data[name]
	if !ok {
		uml.ThrowNPE("")
	}
	return &wrappedElement{wrapped: result}
}

func (d *saltDictionary) addSprite(name string, sprite klimt.Sprite) { d.sprites[name] = sprite }

// Sprite returns a sprite of the diagram, nil when there is none.
func (d *saltDictionary) Sprite(name string) klimt.Sprite { return d.sprites[name] }

func (*saltDictionary) Value(string) (string, bool)                { return "", false }
func (*saltDictionary) Padding() klimt.ClockwiseTopRightBottomLeft { return klimt.ClockwiseNone() }
func (*saltDictionary) Guillemet() *klimt.Guillemet                { return klimt.GuillemetGuillemet() }
func (*saltDictionary) MonospacedFamily() string                   { return klimt.ParserMonospaced }
func (*saltDictionary) TabSize() int                               { return 8 }
func (*saltDictionary) IHtmlColorSet() *color.HColorSet            { return color.Instance() }
func (*saltDictionary) Dpi() int                                   { return 96 }

// CopyAllFrom panics, as upstream throws UnsupportedOperationException.
func (*saltDictionary) CopyAllFrom(map[string]string) {
	uml.Throw(uml.UnsupportedOperationException, "")
}

// Values panics, as upstream throws UnsupportedOperationException.
func (*saltDictionary) Values() map[string]string {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

func (*saltDictionary) TransformStringForSizeHack(s string) string { return s }

// Sheet panics, as upstream throws UnsupportedOperationException.
func (*saltDictionary) Sheet(*klimt.FontConfiguration, geom.HorizontalAlignment, klimt.CreoleMode) klimt.SheetBuilder {
	uml.Throw(uml.UnsupportedOperationException, "")
	return nil
}

// SheetStereo returns the creole parser of a text.
func (d *saltDictionary) SheetStereo(fontConfiguration *klimt.FontConfiguration,
	horizontalAlignment geom.HorizontalAlignment, creoleMode klimt.CreoleMode,
	stereo *klimt.FontConfiguration) klimt.SheetBuilder {
	return klimt.NewCreoleParser(fontConfiguration, horizontalAlignment, d, creoleMode, stereo)
}

func (*saltDictionary) Pragma() *klimt.Pragma { return klimt.PragmaCreateEmpty() }

func (d *saltDictionary) Options() *klimt.ConfigurationStore[klimt.OptionKey] { return d.option }

func (d *saltDictionary) PathSystem() *core.RenderContext { return d.pathSystem }

func (*saltDictionary) FromMd5(string) (string, bool) { return "", false }
