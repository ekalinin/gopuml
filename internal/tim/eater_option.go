// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterOption.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

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
		artifact.addWarning("No such !option " + key)
	case value == nil && optionKey.defaultValue == nil:
		artifact.addWarning("No default value for " + key)
	case value == nil:
		artifact.option.define(optionKey, *optionKey.defaultValue)
	default:
		artifact.option.define(optionKey, value.String())
	}
	return nil
}
