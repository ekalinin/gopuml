// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/EaterOption.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// EaterOption is "!option name [value]".
type EaterOption struct{ Eater }

func newEaterOption(sl text.StringLocated) *EaterOption { return &EaterOption{newEater(sl)} }

func (e *EaterOption) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!option"); err != nil {
		return err
	}
	e.skipSpaces()
	key, err := e.eatAndGetVarname()
	if err != nil {
		return err
	}
	e.skipSpaces()
	value, err := e.eatExpression(context, memory)
	if err != nil {
		// Sorry, not great...
		value = nil
	}
	e.skipSpaces()
	artifact := context.preprocessingArtifact
	optionKey, ok := lazyFromOptionKey(key)
	switch {
	case !ok:
		// A warning of two lines
		var known []string
		for _, k := range optionKeys {
			known = append(known, strings.ToLower(k.Name))
		}
		artifact.addWarning("Unknown '!option " + key + "': this line is ignored\nKnown options are: " +
			strings.Join(known, ", "))
	case value == nil && optionKey.defaultValue == nil:
		artifact.addWarning("'!option " + key + "' needs a value: write '!option " + key + " <value>'")
	case value == nil:
		artifact.option.define(optionKey, *optionKey.defaultValue)
	default:
		artifact.option.define(optionKey, value.String())
	}
	return nil
}
