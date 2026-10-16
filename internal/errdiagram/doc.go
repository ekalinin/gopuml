// SPDX-License-Identifier: MIT
// Provenance: original

// Package errdiagram draws the built-in images of upstream's
// net.sourceforge.plantuml.error package and its friends:
//
//   - the error image of a block that is not a valid diagram
//     (PSystemError, with PSystemErrorV2 and PSystemErrorPreprocessor, which
//     only differ in how they are built): the "[From ...]" location of the
//     error, the source lines read before it, the last one underlined with a
//     red wave, the message, the version line, and the welcome help on top
//     when the block has fewer than five lines. core.ErrorDiagram values
//     with NotSupportedYet set are drawn the same way;
//   - the image of an unknown start directive (PSystemUnsupported);
//   - the welcome screen of an empty @startuml block (PSystemWelcome, built
//     by WelcomeFactory, which goes first in the factory list as upstream's
//     PSystemWelcomeFactory);
//   - the image of a source without any @start/@end block
//     (SourceStringReader.noValidStartFound).
//
// Upstream adds sponsor messages to error images during some minutes of each
// hour, and the "This version of PlantUML is N days old" warning when its
// build is more than 140 days old. Both depend on the clock: they are never
// added here (design D16), as with PSystemError.disableTimeBasedErrorDecorations,
// which the references are generated with. The image that upstream adds
// next to an error whose source contains "arecibo" is not added either: with
// those decorations disabled, upstream does not add it.
//
// The PlantUML logo of the welcome and unsupported-diagram images is
// embedded as upstream writes it in SVG data URLs (logo.png holds the bytes
// that Java's ImageIO writes for upstream's version/logo.png, the same
// pixels), so the image data is byte-identical to upstream's.
package errdiagram
