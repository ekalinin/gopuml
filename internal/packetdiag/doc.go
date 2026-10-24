// SPDX-License-Identifier: MIT
// Provenance: original

// Package packetdiag parses and draws packet diagrams (@startpacketdiag):
// the PacketDiagram model with its layout in rows of bits, the drawing of
// its fields (PacketBlock) and of its bit scale (PacketIndicator), and the
// factory and its commands in upstream's order (upstream's packetdiag and
// packetdiag.command packages). It needs no graph layout.
package packetdiag
