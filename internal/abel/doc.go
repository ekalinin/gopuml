// SPDX-License-Identifier: MIT
// Provenance: original

// Package abel is the entity model shared by the graph diagrams (class,
// object, description, state): entities, links, groups, together blocks,
// notes, the hide, show, remove and restore rules, and the namespace of
// names.
//
// Upstream's abel, cucadiagram and net.atmp (CucaDiagram) packages refer to
// each other, so they form one Go package (design D2); the files carry the
// origin package as prefix: cucadiagram_*.go, atmp_*.go and plasma_*.go
// (the namespace, which only abel uses). The small utils.Position type
// moved here, and abel's LinkStrategy moved to internal/decoration.
//
// The commands that upstream defines in its command, command.note and
// descdiagram.command packages for any CucaDiagram are here too
// (command_*.go, descdiagram_*.go), generic over Diagram: the command
// framework (internal/command) cannot import the entity model. The drawing
// of bodies (cucadiagram_body*.go, MethodsOrFieldsArea, TextBlockMap,
// TextBlockCucaJSon) is here; the graph layout and the images of entities
// are internal/svek, which keeps its qualifier boxes on the entities
// (entity_kal.go).
package abel
