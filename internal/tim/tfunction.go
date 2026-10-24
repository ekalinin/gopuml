// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TFunction.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TFunctionType.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TFunctionArgument.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"github.com/ekalinin/gopuml/internal/text"
)

// TFunction is a preprocessor function or procedure: a builtin, a user
// definition or a legacy !define.
type TFunction interface {
	Signature() TFunctionSignature
	// CanCover reports whether a call with nbArg positional arguments and
	// the given named arguments can use this function.
	CanCover(nbArg int, namedArguments []string) bool
	FunctionType() TFunctionType
	ExecuteReturnFunction(context *TContext, memory TMemory, location text.StringLocated, args []*TValue,
		named map[string]*TValue) (*TValue, error)
	ExecuteProcedureInternal(context *TContext, memory TMemory, location text.StringLocated, args []*TValue,
		named map[string]*TValue) error
	IsUnquoted() bool
}

// TFunctionType is the kind of a TFunction. The zero value is Java's null,
// the type of top-level code.
type TFunctionType int

// Function types.
const (
	functionNone TFunctionType = iota
	functionProcedure
	functionReturn
	functionLegacyDefine
	functionLegacyDefineLong
)

var tFunctionTypeNames = [...]string{"null", "PROCEDURE", "RETURN_FUNCTION", "LEGACY_DEFINE", "LEGACY_DEFINELONG"}

func (t TFunctionType) String() string { return tFunctionTypeNames[t] }

func (t TFunctionType) isLegacy() bool {
	return t == functionLegacyDefine || t == functionLegacyDefineLong
}

// TFunctionArgument is a declared argument, with its optional default value.
type TFunctionArgument struct {
	name string
	def  *TValue
}

func (a TFunctionArgument) String() string { return "ARG:" + a.name }
