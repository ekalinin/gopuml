// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/klimt/creole/command/CommandCreoleEmoji.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "unicode/utf16"

// CreoleEmojiMatchingSize is upstream's
// CommandCreoleEmoji.create().matchingSize(line, 0), which is also what its
// executeAndAdvance returns without stripe: the length, in UTF-16 units, of
// the "<:name:>" emoji at the start of line, 0 for none. Activity diagrams
// use it for the boxes of the <<icon>> stereotype.
func CreoleEmojiMatchingSize(line string) int {
	return commandCreoleEmoji().matchingSize(utf16.Encode([]rune(line)), 0)
}
