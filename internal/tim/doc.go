// SPDX-License-Identifier: MIT
// Provenance: original

// Package tim implements the PlantUML preprocessor (variables, macros, control flow, functions, builtins and includes).
//
// It translates upstream's packages tim, tim.expression, tim.iterator and
// tim.builtin, which reference each other, into one Go package, together
// with the parts of preproc, preproc2 and nio that the preprocessor uses.
// Files are named after the upstream classes. Preprocess runs the
// preprocessor on one block.
//
// Builtins are registered in the standardFunctions table (builtins.go).
package tim
