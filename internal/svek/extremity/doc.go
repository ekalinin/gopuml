// SPDX-License-Identifier: MIT
// Provenance: original

// Package extremity is upstream's svek.extremity: the decorations drawn at
// the ends of the links of graph diagrams (arrow heads, diamonds, crow's
// feet, circles...) and in their middle, with upstream's
// AbstractExtremityFactory.
//
// Upstream's LinkDecor and LinkMiddleDecor enums (internal/decoration)
// choose the factories; those methods are the functions FactoryLegacy,
// FactoryComplete and MiddleFactoryFor here, so that internal/decoration
// does not depend on the drawing code.
package extremity
