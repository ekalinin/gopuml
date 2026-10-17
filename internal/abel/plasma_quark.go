// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/plasma/Plasma.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/plasma/Quark.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/plasma/PEntry.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package abel

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/uml"
)

// MagicSeparator is upstream's Plasma.MAGIC_SEPARATOR, the separator of a
// namespace without separator.
const MagicSeparator = "\u0001"

// Plasma is the namespace of the quarks of a diagram: a tree of named
// nodes, each holding an entity or nothing. Upstream's Plasma<DATA> is only
// used with entities, so the Go types are not generic.
//
// Upstream keeps the separator in the Plasma and CucaDiagram sets it
// whenever the namespace separator of the diagram changes: here it is read
// from the diagram.
type Plasma struct {
	base   *uml.Base
	root   *Quark
	quarks []*Quark
	stats  map[string]*pentry
}

// pentry is upstream's PEntry: the first quark of a name and how many
// quarks have it.
type pentry struct {
	first   *Quark
	counter int
}

// newPlasma returns the namespace of a diagram, with its root quark.
func newPlasma(base *uml.Base) *Plasma {
	p := &Plasma{base: base, stats: map[string]*pentry{}}
	p.root = newQuark(p, nil, "")
	return p
}

func (p *Plasma) register(q *Quark) {
	p.quarks = append(p.quarks, q)
	if ent, ok := p.stats[q.name]; ok {
		ent.counter++
	} else {
		p.stats[q.name] = &pentry{first: q, counter: 1}
	}
}

// Root returns the root quark.
func (p *Plasma) Root() *Quark { return p.root }

// Separator returns the namespace separator, MagicSeparator when the
// diagram has none.
func (p *Plasma) Separator() string {
	if s, ok := p.base.NamespaceSeparator(); ok {
		return s
	}
	return MagicSeparator
}

// HasSeparator reports whether the diagram has a namespace separator.
func (p *Plasma) HasSeparator() bool { return p.Separator() != MagicSeparator }

// Quarks returns every quark, the root first, in creation order.
func (p *Plasma) Quarks() []*Quark { return p.quarks }

// FirstWithName returns the first quark created with a name, nil for none.
func (p *Plasma) FirstWithName(name string) *Quark {
	if ent, ok := p.stats[name]; ok {
		return ent.first
	}
	return nil
}

// CountByName returns the number of quarks of a name.
func (p *Plasma) CountByName(name string) int {
	if ent, ok := p.stats[name]; ok {
		return ent.counter
	}
	return 0
}

// Quark is a named node of the namespace, holding at most one entity.
type Quark struct {
	plasma        *Plasma
	parent        *Quark
	name          string
	data          *Entity
	childNames    []string
	children      map[string]*Quark
	qualifiedName string
}

func newQuark(plasma *Plasma, parent *Quark, name string) *Quark {
	q := &Quark{plasma: plasma, parent: parent, name: name, children: map[string]*Quark{}}
	if parent == nil || parent.parent == nil {
		q.qualifiedName = name
	} else {
		q.qualifiedName = parent.qualifiedName + plasma.Separator() + name
	}
	plasma.register(q)
	return q
}

// Parent returns the parent quark, nil for the root.
func (q *Quark) Parent() *Quark { return q.parent }

// String is upstream's toString: the qualified name.
func (q *Quark) String() string { return q.qualifiedName }

// StringPoint returns the qualified name with "." as separator.
func (q *Quark) StringPoint() string {
	if q.parent == nil || q.parent.parent == nil {
		return q.name
	}
	return q.parent.StringPoint() + "." + q.name
}

// Name returns the name.
func (q *Quark) Name() string { return q.name }

// QualifiedName returns the names from the root, joined by the separator
// of the namespace when the quark was created.
func (q *Quark) QualifiedName() string { return q.qualifiedName }

// IsRoot reports whether the quark is the root.
func (q *Quark) IsRoot() bool { return q.parent == nil }

// Plasma returns the namespace.
func (q *Quark) Plasma() *Plasma { return q.plasma }

// Data returns the entity, nil for none.
func (q *Quark) Data() *Entity { return q.data }

// SetData sets the entity; it panics when the quark has one, as upstream
// throws IllegalStateException.
func (q *Quark) SetData(data *Entity) {
	if q.data != nil {
		uml.Throw(uml.IllegalStateException, "")
	}
	q.data = data
}

// ChildIfExists returns the direct child of a name, nil for none. A name
// containing the separator panics, as upstream throws
// IllegalArgumentException.
func (q *Quark) ChildIfExists(name string) *Quark {
	if q.plasma.HasSeparator() && strings.Contains(name, q.plasma.Separator()) {
		uml.Throw(uml.IllegalArgumentException, "")
	}
	return q.children[name]
}

// Child returns the descendant of a qualified name, creating the missing
// quarks.
func (q *Quark) Child(full string) *Quark {
	if !q.plasma.HasSeparator() {
		return q.directChild(full)
	}
	separator := q.plasma.Separator()
	full = q.clean(full)
	current := q
	for {
		idx := strings.Index(full, separator)
		if idx == -1 {
			return current.directChild(full)
		}
		current = current.directChild(full[:idx])
		full = q.clean(full[idx+len(separator):])
	}
}

func (q *Quark) directChild(name string) *Quark {
	if result, ok := q.children[name]; ok {
		return result
	}
	result := newQuark(q.plasma, q, name)
	q.children[name] = result
	q.childNames = append(q.childNames, name)
	return result
}

func (q *Quark) clean(full string) string {
	separator := q.plasma.Separator()
	for strings.HasPrefix(full, separator) {
		full = full[len(separator):]
	}
	for strings.HasSuffix(full, separator) {
		full = full[:len(full)-len(separator)]
	}
	return full
}

// Children returns the direct children in creation order.
func (q *Quark) Children() []*Quark {
	result := make([]*Quark, len(q.childNames))
	for i, name := range q.childNames {
		result[i] = q.children[name]
	}
	return result
}

// CountChildren returns the number of direct children.
func (q *Quark) CountChildren() int { return len(q.childNames) }
