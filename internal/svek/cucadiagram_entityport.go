// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cucadiagram/EntityPort.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/abel/Link.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package svek

import (
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/abel"
	"github.com/ekalinin/plantuml-in-go/internal/klimt"
)

// entityPort is upstream's cucadiagram.EntityPort: the dot node of a link
// end, with its port. It lives here because upstream's Link.getEntityPort1
// asks the Bibliotekon for the node.
type entityPort struct {
	entityUID string
	portID    string
	hasPort   bool
}

func entityPortCreate(entityUID string, portName abel.OptString) entityPort {
	if !portName.OK {
		return entityPort{entityUID: entityUID}
	}
	return entityPort{entityUID: entityUID, portID: klimt.EncodePortNameToID(portName.S), hasPort: true}
}

func entityPortForPort(entityUID string) entityPort {
	return entityPort{entityUID: entityUID, portID: "P", hasPort: true}
}

// fullString is "uid:port", or the uid.
func (p entityPort) fullString() string {
	if p.hasPort {
		return p.entityUID + ":" + p.portID
	}
	return p.entityUID
}

func (p entityPort) isShielded() bool { return strings.HasSuffix(p.entityUID, ":h") }

// prefix is the uid without the ":h" of shielded nodes.
func (p entityPort) prefix() string {
	if p.isShielded() {
		return p.entityUID[:len(p.entityUID)-2]
	}
	return p.entityUID
}

func (p entityPort) startsWith(centerID string) bool { return strings.HasPrefix(p.entityUID, centerID) }

func (p entityPort) equalsID(other entityPort) bool { return p.entityUID == other.entityUID }

// linkEntityPort1 is Link.getEntityPort1.
func linkEntityPort1(l *abel.Link, bibliotekon *Bibliotekon) entityPort {
	return linkEntityPort(l.Entity1(), l.PortName1(), bibliotekon)
}

// linkEntityPort2 is Link.getEntityPort2.
func linkEntityPort2(l *abel.Link, bibliotekon *Bibliotekon) entityPort {
	return linkEntityPort(l.Entity2(), l.PortName2(), bibliotekon)
}

func linkEntityPort(leaf *abel.Entity, port abel.OptString, bibliotekon *Bibliotekon) entityPort {
	if leaf.EntityPosition().UsePortP() {
		return entityPortForPort(bibliotekon.nodeUID(leaf))
	}
	return entityPortCreate(bibliotekon.nodeUID(leaf), port)
}
