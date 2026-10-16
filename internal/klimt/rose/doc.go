// SPDX-License-Identifier: MIT
// Provenance: original

// Package rose is upstream's skin.rose: the components that draw the
// elements of sequence diagrams (message arrows, participant heads and
// lines, notes, groupings, dividers, delays, references, boxes) and the Rose
// skin that builds them.
//
// Upstream's skin.rose imports two sequencediagram types, which imports
// skin.rose. To keep Go's packages acyclic (design D2), NotePosition is
// defined here, and the participant is the small Participant interface that
// sequencediagram's participant implements.
//
// Upstream's unchecked exceptions are panics with a *color.RuntimeError
// value, the convention of the klimt packages.
package rose
