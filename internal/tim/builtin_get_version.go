// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/GetVersion.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/plantuml-in-go/internal/text"
	"github.com/ekalinin/plantuml-in-go/internal/version"
)

// getVersion is %version(): the upstream version this implementation
// targets (design D15).
type getVersion struct{ simpleReturnFunction }

func (getVersion) Signature() TFunctionSignature { return NewTFunctionSignature("%version", 0) }

func (getVersion) CanCover(nbArg int, _ []string) bool { return nbArg == 0 }

func (getVersion) ExecuteReturnFunction(*TContext, TMemory, text.StringLocated, []*TValue,
	map[string]*TValue) (*TValue, error) {
	return TValueFromString(version.Upstream), nil
}
