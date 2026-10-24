// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/TFunctionSignature.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"strconv"
	"strings"
)

// TFunctionSignature names a function and its number of arguments. Two
// signatures are equal when the name and the number match; the named
// arguments are not compared.
type TFunctionSignature struct {
	FunctionName   string
	NbArg          int
	NamedArguments []string
}

// NewTFunctionSignature returns a signature without named arguments.
func NewTFunctionSignature(functionName string, nbArg int) TFunctionSignature {
	return TFunctionSignature{FunctionName: functionName, NbArg: nbArg}
}

func (s TFunctionSignature) sameFunctionNameAs(other TFunctionSignature) bool {
	return s.FunctionName == other.FunctionName
}

func (s TFunctionSignature) String() string {
	return s.FunctionName + "/" + strconv.Itoa(s.NbArg) + " [" + strings.Join(s.NamedArguments, ", ") + "]"
}

// signatureKey is the identity of a signature in maps.
type signatureKey struct {
	name  string
	nbArg int
}

func (s TFunctionSignature) key() signatureKey { return signatureKey{s.FunctionName, s.NbArg} }

// hashCode is TFunctionSignature.hashCode.
func (s TFunctionSignature) hashCode() int32 {
	const prime = 31
	result := int32(1)
	result = prime*result + javaHashCode(s.FunctionName)
	result = prime*result + int32(s.NbArg)
	return result
}
