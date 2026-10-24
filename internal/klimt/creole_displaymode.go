// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/klimt/creole/Display.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package klimt

// DefaultCreoleMode returns the creole mode the display is drawn with
// (upstream's private defaultCreoleMode field, which the model tests of
// graph diagrams read).
func (d *Display) DefaultCreoleMode() CreoleMode { return d.defaultCreoleMode }
