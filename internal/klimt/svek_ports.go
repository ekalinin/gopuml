// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/Ports.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/PortGeometry.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/svek/WithPorts.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import (
	"slices"
	"strings"

	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// PortGeometry is the vertical position of a port (a member of a class box
// that links can attach to). Upstream defines it in svek.
type PortGeometry struct {
	id       string
	position float64
	height   float64
	score    int
}

// NewPortGeometry returns a port geometry.
func NewPortGeometry(id string, position, height float64, score int) PortGeometry {
	return PortGeometry{id: id, position: position, height: height, score: score}
}

// TranslateY returns the geometry moved down by deltaY.
func (p PortGeometry) TranslateY(deltaY float64) PortGeometry {
	return PortGeometry{id: p.id, position: p.position + deltaY, height: p.height, score: p.score}
}

// String is upstream's toString.
func (p PortGeometry) String() string {
	return "pos=" + jstr(p.position) + " height=" + jstr(p.height) + " (" + itoa(p.score) + ")"
}

// Height returns the height.
func (p PortGeometry) Height() float64 { return p.height }

// Position returns the top.
func (p PortGeometry) Position() float64 { return p.position }

// LastY returns the bottom.
func (p PortGeometry) LastY() float64 { return p.position + p.height }

// Score returns the score; a higher score wins among ports of one name.
func (p PortGeometry) Score() int { return p.score }

// ID returns the identifier.
func (p PortGeometry) ID() string { return p.id }

// Ports maps port identifiers to their geometry. Upstream defines it in svek;
// it lives here because text blocks report their ports. Iteration follows
// upstream's HashMap order.
type Ports struct {
	keys []string // insertion order, without duplicates
	ids  map[string]PortGeometry
}

// NewPorts returns an empty set of ports.
func NewPorts() *Ports { return &Ports{ids: map[string]PortGeometry{}} }

// EncodePortNameToID returns the identifier of a port name: "p" and the MD5
// of the name.
func EncodePortNameToID(portName string) string { return "p" + md5Hex(portName) }

func (p *Ports) put(id string, g PortGeometry) {
	if _, ok := p.ids[id]; !ok {
		p.keys = append(p.keys, id)
	}
	p.ids[id] = g
}

// orderedKeys returns the identifiers in upstream's HashMap order.
func (p *Ports) orderedKeys() []string { return javaHashMapOrder(p.keys) }

// String is upstream's toString, in HashMap order.
func (p *Ports) String() string {
	parts := make([]string, 0, len(p.keys))
	for _, k := range p.orderedKeys() {
		parts = append(parts, k+"="+p.ids[k].String())
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// TranslateY returns the ports moved down by deltaY.
func (p *Ports) TranslateY(deltaY float64) *Ports {
	result := NewPorts()
	for _, k := range p.orderedKeys() {
		result.put(k, p.ids[k].TranslateY(deltaY))
	}
	return result
}

// Add records a port, unless one of the same name has a higher or equal
// score.
func (p *Ports) Add(portName string, score int, position, height float64) {
	id := EncodePortNameToID(portName)
	already, ok := p.ids[id]
	if !ok || already.Score() < score {
		p.put(id, NewPortGeometry(id, position, height, score))
	}
}

// AddThis merges other into the ports, keeping the higher score per port.
func (p *Ports) AddThis(other *Ports) {
	for _, key := range other.orderedKeys() {
		g := other.ids[key]
		already, ok := p.ids[key]
		if !ok || already.Score() < g.Score() {
			p.put(key, g)
		}
	}
}

// AllPortGeometry returns the geometries sorted by position; equal
// positions keep the HashMap order, as Java's stable sort does.
func (p *Ports) AllPortGeometry() []PortGeometry {
	keys := p.orderedKeys()
	result := make([]PortGeometry, len(keys))
	for i, k := range keys {
		result[i] = p.ids[k]
	}
	slices.SortStableFunc(result, func(a, b PortGeometry) int { return javaDoubleCompare(a.position, b.position) })
	return result
}

// WithPorts is a text block that has ports. Upstream defines it in svek.
type WithPorts interface {
	Ports(stringBounder font.StringBounder) *Ports
}
