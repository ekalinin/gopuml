// SPDX-License-Identifier: MIT
// Provenance: original

// Package decoration is upstream's net.sourceforge.plantuml.decoration: the
// types of the links of graph diagrams (LinkType with its two LinkDecor
// heads, its LinkStyle and its LinkMiddleDecor) and WithLinkType, the base
// of the links.
//
// Upstream's LinkType imports abel.LinkStrategy, and abel.Link extends
// WithLinkType: per design D2, the small LinkStrategy type moves here so
// that the packages do not form a cycle. The drawing parts of LinkDecor and
// LinkMiddleDecor (the extremity and middle factories) are functions of
// internal/svek/extremity. Rainbow is here with the part the graph links
// use (RainbowBuild from the arrow style); the activity diagrams have the
// full Rainbow and HtmlColorAndStyle in internal/activitydiagram3.
package decoration
