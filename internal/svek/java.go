// SPDX-License-Identifier: MIT
// Provenance: original

package svek

import "github.com/ekalinin/gopuml/internal/uml"

// The Java exceptions thrown by the ported code, as panics with
// *color.RuntimeError (see internal/uml).

func throwIllegalState(message string) { uml.Throw(uml.IllegalStateException, message) }

func throwIllegalArgument(message string) { uml.Throw(uml.IllegalArgumentException, message) }

func throwUnsupported(message string) { uml.Throw(uml.UnsupportedOperationException, message) }

func throwIndexOutOfBounds(index, length int) { uml.ThrowIndexOutOfBounds(index, length) }

func throwNumberFormat(s string) {
	uml.Throw(uml.NumberFormatException, `For input string: "`+s+`"`)
}

// notPorted panics for an upstream class that the class diagrams do not
// need and that is not ported yet.
func notPorted(class string) {
	throwUnsupported("not ported yet: " + class)
}

func throwNPE() { uml.ThrowNPE("") }
