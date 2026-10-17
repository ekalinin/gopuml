// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/RandomFunction.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"math/rand/v2"

	"github.com/ekalinin/gopuml/internal/text"
)

// randomFunction is %random() (0 or 1), %random(max) (from 0 to max-1) and
// %random(min, max) (from min to max-1). Like upstream, each block has its
// own java.util.Random, seeded with Options.RandomSeed or at random. A bound
// that is not positive throws IllegalArgumentException, which becomes
// "Fatal parsing error".
type randomFunction struct {
	simpleReturnFunction
	random *javaRandom
}

func (*randomFunction) Signature() TFunctionSignature { return NewTFunctionSignature("%random", 2) }

func (*randomFunction) CanCover(nbArg int, _ []string) bool {
	return nbArg == 0 || nbArg == 1 || nbArg == 2
}

func (f *randomFunction) ExecuteReturnFunction(context *TContext, _ TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	if f.random == nil {
		seed := rand.Int64()
		if context.randomSeed != nil {
			seed = *context.randomSeed
		}
		f.random = newJavaRandom(seed)
	}
	var bound, offset int32
	switch len(values) {
	case 0:
		bound = 2
	case 1:
		bound = values[0].ToInt()
	case 2:
		lo, hi := values[0].ToInt(), values[1].ToInt()
		bound, offset = hi-lo, lo
	default:
		return nil, newEaterException("Error on Random: Too many argument", location)
	}
	v, ok := f.random.nextInt(bound)
	if !ok {
		throwRuntime("IllegalArgumentException", "bound must be positive")
	}
	return TValueFromInt(v + offset), nil
}
