// SPDX-License-Identifier: MIT
// Provenance: original

package rose

import (
	"strconv"

	"github.com/ekalinin/gopuml/internal/klimt/color"
)

// Upstream's unchecked exceptions are panics with a *color.RuntimeError
// value, the convention of the klimt packages.
const (
	unsupportedOperation = "java.lang.UnsupportedOperationException"
	illegalArgument      = "java.lang.IllegalArgumentException"
	nullPointer          = "java.lang.NullPointerException"
)

func throw(class, message string) { panic(&color.RuntimeError{Class: class, Message: message}) }

func itoa(i int) string { return strconv.Itoa(i) }

// javaD2I is Java's (int) cast of a double: truncation toward zero, NaN to
// zero and saturation at the int range.
func javaD2I(x float64) int {
	switch {
	case x != x:
		return 0
	case x >= 2147483647:
		return 2147483647
	case x <= -2147483648:
		return -2147483648
	}
	return int(x)
}
