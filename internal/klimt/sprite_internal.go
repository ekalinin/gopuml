// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/sprite/SpriteImage.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

import "embed"

// internalSprites are upstream's /sprites resources of the jar, the
// ArchiMate icons that upstream takes from Archi: sprites/archimate/<name>.svg
// or .png, unchanged (NOTICE).
//
//go:embed sprites/archimate
var internalSprites embed.FS
