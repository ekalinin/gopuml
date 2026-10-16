// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/sprite/Sprite.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/sprite/SpriteContainer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/sprite/SpriteContainerEmpty.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/sprite/SpriteUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/sprite/SpriteImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/color"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/font"
	"github.com/ekalinin/plantuml-in-go/internal/klimt/geom"
)

// Sprite is a small image that texts can show ("<$name>"). The sprite
// implementations (monochrome, color, SVG, image) are ported with creole.
type Sprite interface {
	AsTextBlock(fontColor, forcedColor color.HColor, scale float64, backColor color.HColor) TextBlock
}

// SpriteContainer holds the sprites of a diagram.
type SpriteContainer interface {
	font.SvgCharSizeHack
	// Sprite returns the sprite named name, nil when there is none.
	Sprite(name string) Sprite
	Guillemet() *Guillemet
	// FromMd5 returns the text of an MD5 hash; ok is false when unknown.
	FromMd5(md5 string) (text string, ok bool)
}

// SpriteUtilsSpriteName is the pattern of sprite names.
const SpriteUtilsSpriteName = `[-\p{L}0-9_/]+`

// SpriteImageFromInternal returns the sprite of upstream's jar resources
// (/sprites/<name>.svg or .png, the archimate sprites), nil when there is
// none. The resources are not bundled yet: a name that upstream would find
// panics, any other gives nil as upstream.
func SpriteImageFromInternal(name string) Sprite {
	if strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".svg") {
		throw(illegalArgument, "")
	}
	if isInternalSprite(name) {
		notPorted("klimt/sprite/SpriteImage.fromInternal", "internal sprite "+name)
	}
	return nil
}

// SpriteContainerEmpty is the sprite container and simple skin of texts
// outside diagrams (error images, legends of tools): no sprites besides the
// internal ones, default values.
type SpriteContainerEmpty struct{}

var _ ISkinSimple = SpriteContainerEmpty{}

// Sprite returns the internal sprite of a name.
func (SpriteContainerEmpty) Sprite(name string) Sprite { return SpriteImageFromInternal(name) }

// Value returns no value.
func (SpriteContainerEmpty) Value(string) (string, bool) { return "", false }

// Padding returns no padding.
func (SpriteContainerEmpty) Padding() ClockwiseTopRightBottomLeft { return ClockwiseNone() }

// Guillemet returns "<<" and ">>".
func (SpriteContainerEmpty) Guillemet() *Guillemet { return GuillemetDoubleComparator() }

// MonospacedFamily returns "monospaced".
func (SpriteContainerEmpty) MonospacedFamily() string { return ParserMonospaced }

// TabSize returns 8.
func (SpriteContainerEmpty) TabSize() int { return 8 }

// IHtmlColorSet returns the color set.
func (SpriteContainerEmpty) IHtmlColorSet() *color.HColorSet { return color.Instance() }

// Dpi returns 96.
func (SpriteContainerEmpty) Dpi() int { return 96 }

// CopyAllFrom panics, as upstream throws UnsupportedOperationException.
func (SpriteContainerEmpty) CopyAllFrom(map[string]string) { throw(unsupportedOperation, "") }

// Values panics, as upstream throws UnsupportedOperationException.
func (SpriteContainerEmpty) Values() map[string]string {
	throw(unsupportedOperation, "")
	return nil
}

// MinClassWidthTOBEREMOVED returns 0.
func (SpriteContainerEmpty) MinClassWidthTOBEREMOVED(*Style) float64 { return 0 }

// TransformStringForSizeHack returns s.
func (SpriteContainerEmpty) TransformStringForSizeHack(s string) string { return s }

// Sheet panics, as upstream throws UnsupportedOperationException.
func (SpriteContainerEmpty) Sheet(*FontConfiguration, geom.HorizontalAlignment, CreoleMode) SheetBuilder {
	throw(unsupportedOperation, "")
	return nil
}

// SheetStereo returns the creole parser of a text.
func (c SpriteContainerEmpty) SheetStereo(fontConfiguration *FontConfiguration,
	horizontalAlignment geom.HorizontalAlignment, creoleMode CreoleMode, stereo *FontConfiguration) SheetBuilder {
	return newCreoleParser(fontConfiguration, horizontalAlignment, c, creoleMode, stereo)
}

// Pragma returns empty pragmas.
func (SpriteContainerEmpty) Pragma() *Pragma { return PragmaCreateEmpty() }

// Options returns empty options.
func (SpriteContainerEmpty) Options() *ConfigurationStore[OptionKey] {
	return NewConfigurationStore[OptionKey]()
}

// PathSystem returns nil: upstream's PathSystem.fetch() (the current
// directory) is the render context's concern.
func (SpriteContainerEmpty) PathSystem() *core.RenderContext { return nil }

// FromMd5 returns no text.
func (SpriteContainerEmpty) FromMd5(string) (string, bool) { return "", false }
