// SPDX-License-Identifier: MIT
// Provenance: original

package gopuml

import (
	"errors"
	"fmt"

	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/security"
)

// Errors returned by the package. Test for them with errors.Is; a syntax
// error is also a *SyntaxError, usable with errors.As.
var (
	// ErrSyntax matches every *SyntaxError.
	ErrSyntax = errors.New("plantuml: syntax error")
	// ErrNoDiagram reports a source without a diagram, or a diagram or
	// page index out of range.
	ErrNoDiagram = errors.New("plantuml: no diagram found")
	// ErrUnsupportedDiagram reports a diagram type or family this
	// implementation does not provide.
	ErrUnsupportedDiagram = errors.New("plantuml: diagram type not supported")
	// ErrUnsupportedFormat reports a format a diagram type cannot produce.
	ErrUnsupportedFormat = errors.New("plantuml: format not supported for this diagram")
	// ErrGraphvizNotFound reports that the Graphviz dot program is missing.
	ErrGraphvizNotFound = errors.New("plantuml: Graphviz dot not found")
	// ErrGraphvizTimeout reports a dot run stopped after the timeout.
	ErrGraphvizTimeout = errors.New("plantuml: Graphviz dot timed out")
	// ErrGraphvizFailed reports a dot run that failed.
	ErrGraphvizFailed = errors.New("plantuml: Graphviz dot failed")
	// ErrAccessDenied reports a file or URL the security profile does not
	// allow.
	ErrAccessDenied = security.ErrDenied
	// ErrDecode reports text that is not encoded diagram source.
	ErrDecode = code.ErrNoCompression
)

// SyntaxError is a diagram with a syntax error.
type SyntaxError struct {
	// Line is the 1-based line of the error in the source.
	Line int
	// Message is upstream's message, for example "Syntax Error? (Assumed
	// diagram type: sequence)".
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("plantuml: line %d: %s", e.Line, e.Message)
}

// Is makes errors.Is(err, ErrSyntax) true for a *SyntaxError.
func (e *SyntaxError) Is(target error) bool { return target == ErrSyntax }

// UnsupportedError is a diagram whose type or family this implementation
// does not provide.
type UnsupportedError struct {
	// Line is the 1-based line of the source that shows the type.
	Line int
	// Message explains what is not supported, for example "mindmap
	// diagrams are not supported yet".
	Message string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("plantuml: line %d: %s", e.Line, e.Message)
}

// Is makes errors.Is(err, ErrUnsupportedDiagram) true for an
// *UnsupportedError.
func (e *UnsupportedError) Is(target error) bool { return target == ErrUnsupportedDiagram }
