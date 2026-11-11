// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/Peeker.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/utils/PeekerUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package yamlparser

// peeker reads a list, looking ahead of the current element.
type peeker[O any] struct {
	list []O
	pos  int
}

func newPeeker[O any](list []O) *peeker[O] { return &peeker[O]{list: list} }

// peek returns the element ahead of the current one; ok is false past the
// end, where upstream returns null.
func (p *peeker[O]) peek(ahead int) (o O, ok bool) {
	if tmp := p.pos + ahead; tmp < len(p.list) {
		return p.list[tmp], true
	}
	return o, false
}

// jump moves to the next element.
func (p *peeker[O]) jump() { p.pos++ }
