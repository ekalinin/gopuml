// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalAnd.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalNand.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalNor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalNxor.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalOr.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LogicalXor.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/plantuml-in-go/internal/text"

// logicalFunction is one of the logical builtins with two or more
// arguments: %and, %nand, %nor, %nxor, %or and %xor. eval gets the number
// of true arguments and the number of arguments.
type logicalFunction struct {
	simpleReturnFunction
	name string
	eval func(nbTrue, nb int) bool
}

func (f logicalFunction) Signature() TFunctionSignature { return NewTFunctionSignature(f.name, 2) }

func (logicalFunction) CanCover(nbArg int, _ []string) bool { return nbArg >= 2 }

func (f logicalFunction) ExecuteReturnFunction(_ *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	nbTrue := 0
	for _, v := range values {
		if v.ToBoolean() {
			nbTrue++
		}
	}
	return TValueFromBoolean(f.eval(nbTrue, len(values))), nil
}

func newLogicalFunction(name string, eval func(nbTrue, nb int) bool) func(*Defines) TFunction {
	return func(*Defines) TFunction { return logicalFunction{name: name, eval: eval} }
}

// The logical operations of LogicalAnd, LogicalNand, LogicalNor, LogicalNxor,
// LogicalOr and LogicalXor. Upstream stops at the first decisive argument;
// the arguments are already evaluated, so counting gives the same result.
var (
	logicalAnd  = func(nbTrue, nb int) bool { return nbTrue == nb }
	logicalNand = func(nbTrue, nb int) bool { return nbTrue != nb }
	logicalNor  = func(nbTrue, _ int) bool { return nbTrue == 0 }
	logicalNxor = func(nbTrue, _ int) bool { return nbTrue != 1 }
	logicalOr   = func(nbTrue, _ int) bool { return nbTrue > 0 }
	logicalXor  = func(nbTrue, _ int) bool { return nbTrue == 1 }
)
