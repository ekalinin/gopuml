// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Modulo.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// modulo is %mod(dividend, divisor): the remainder of the int division,
// with the sign of the dividend.
type modulo struct{ simpleReturnFunction }

func (modulo) Signature() TFunctionSignature { return NewTFunctionSignature("%mod", 2) }

func (modulo) CanCover(nbArg int, _ []string) bool { return nbArg == 2 }

func (modulo) ExecuteReturnFunction(_ *TContext, _ TMemory, location text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	dividend := values[0].ToInt()
	divisor := values[1].ToInt()
	if divisor == 0 {
		return nil, newEaterException("Divide by zero", location)
	}
	// As in Java, MinInt32 % -1 is 0.
	return TValueFromInt(dividend % divisor), nil
}
