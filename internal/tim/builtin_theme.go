// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/GetAllTheme.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/GetCurrentTheme.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/theme/ThemeUtils.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"slices"

	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// ThemeLister is implemented by a ThemeSource that can list its themes. It
// gives %get_all_theme its names; without it, the list is empty.
type ThemeLister interface {
	// ThemeNames returns the names of the bundled themes, without the
	// puml-theme- prefix and the .puml suffix.
	ThemeNames() []string
}

// getAllTheme is %get_all_theme(): the names of the bundled themes, sorted
// (ThemeUtils.getAllThemeNames).
type getAllTheme struct{ simpleReturnFunction }

func (getAllTheme) Signature() TFunctionSignature { return NewTFunctionSignature("%get_all_theme", 0) }

func (getAllTheme) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (getAllTheme) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, _ []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	result := json.NewArray()
	if lister, ok := context.themes.(ThemeLister); ok {
		names := slices.Clone(lister.ThemeNames())
		slices.SortFunc(names, javaCompare)
		for _, theme := range names {
			result.Add(json.ValueOfString(theme))
		}
	}
	return TValueFromJSON(result), nil
}

// getCurrentTheme is %get_current_theme(): the YAML header of the last
// theme loaded, as a JSON object (empty before any !theme).
type getCurrentTheme struct{ simpleReturnFunction }

func (getCurrentTheme) Signature() TFunctionSignature {
	return NewTFunctionSignature("%get_current_theme", 0)
}

func (getCurrentTheme) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (getCurrentTheme) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, _ []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	return TValueFromJSON(context.themeMetadata), nil
}
