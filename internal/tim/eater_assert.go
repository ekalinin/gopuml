// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterAssert.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// EaterAssert is "!assert condition [: message]".
type EaterAssert struct{ Eater }

func newEaterAssert(sl text.StringLocated) *EaterAssert { return &EaterAssert{newEater(sl)} }

func (e *EaterAssert) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!assert"); err != nil {
		return err
	}
	e.skipSpaces()
	value, err := e.eatExpressionStopAtColon(context, memory)
	if err != nil {
		return err
	}
	e.skipSpaces()
	if !value.ToBoolean() {
		if e.peekChar() == ':' {
			if err := e.checkAndEatChar(':'); err != nil {
				return err
			}
			message, err := e.eatExpression(context, memory)
			if err != nil {
				return err
			}
			return newEaterException("Assertion error : "+message.String(), e.getStringLocated())
		}
		return newEaterException("Assertion error", e.getStringLocated())
	}
	return nil
}
