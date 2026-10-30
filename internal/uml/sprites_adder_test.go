// SPDX-License-Identifier: MIT
// Provenance: original

package uml

import (
	"testing"

	"github.com/ekalinin/gopuml/internal/command"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/klimt"
)

// spriteDiagram keeps its sprites itself, as upstream's PSystemSalt.
type spriteDiagram struct {
	Base
	sprites map[string]klimt.Sprite
}

func (*spriteDiagram) Description() string { return "(sprites)" }
func (d *spriteDiagram) AddSprite(name string, sprite klimt.Sprite) {
	d.sprites[name] = sprite
}

// TestSpriteAdder checks that the sprite commands give the sprites to the
// diagram's own AddSprite when it has one, and to the skin parameters of its
// Base otherwise.
func TestSpriteAdder(t *testing.T) {
	sprite := []string{"sprite $dot [4x4/16] {", "0FF0", "FFFF", "FFFF", "0FF0", "}",
		"sprite $svg <svg viewBox=\"0 0 1 1\"></svg>"}

	own := NewFactory(core.Salt, func() []command.Command[*spriteDiagram] {
		return AddCommonCommands1[*spriteDiagram](nil)
	}, func(ctx *core.RenderContext, src *core.UmlSource, md5map map[string]string) *spriteDiagram {
		return &spriteDiagram{Base: NewBase(ctx, src, core.Salt, md5map), sprites: map[string]klimt.Sprite{}}
	})
	src := source(append(append([]string{"@startuml"}, sprite...), "@enduml")...)
	r := own.CreateSystem(core.NewRenderContext(), src)
	d, ok := r.(*spriteDiagram)
	if !ok {
		t.Fatalf("the diagram with its own sprites was not created: %#v", r)
	}
	for _, name := range []string{"dot", "svg"} {
		if d.sprites[name] == nil {
			t.Errorf("sprite %q did not reach the diagram's store", name)
		}
		if d.SkinParam().Sprite(name) != nil {
			t.Errorf("sprite %q reached the skin parameters", name)
		}
	}

	plain, ok := parse(t, testFactory(), sprite...).(*testDiagram)
	if !ok {
		t.Fatal("the test diagram was not created")
	}
	for _, name := range []string{"dot", "svg"} {
		if plain.SkinParam().Sprite(name) == nil {
			t.Errorf("sprite %q did not reach the skin parameters", name)
		}
	}
}
