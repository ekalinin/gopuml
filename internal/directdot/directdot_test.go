// SPDX-License-Identifier: MIT
// Provenance: original

package directdot_test

import "testing"

const pending = "the export of dot diagrams comes with task 4.1 of add-dot-diagrams"

// TestDotBuiltinEngine checks the dot-diagram scenario "Laid out by the
// built-in engine".
func TestDotBuiltinEngine(t *testing.T) { t.Skip(pending) }

// TestDotNamedDot checks the dot-diagram scenario "Laid out by the named
// dot".
func TestDotNamedDot(t *testing.T) { t.Skip(pending) }

// TestDotSyntaxErrorOutput checks the dot-diagram scenario "Syntax error in
// the DOT source".
func TestDotSyntaxErrorOutput(t *testing.T) { t.Skip(pending) }

// TestDotGraphvizWarningOutput checks the dot-diagram scenario "Graphviz
// warning in the output".
func TestDotGraphvizWarningOutput(t *testing.T) { t.Skip(pending) }

// TestDotNotSupportedYetWarningDropped checks the dot-diagram scenario
// "Feature not supported by the built-in engine".
func TestDotNotSupportedYetWarningDropped(t *testing.T) { t.Skip(pending) }

// TestDotPNGNamedDot checks the dot-diagram scenario "PNG with the named
// dot".
func TestDotPNGNamedDot(t *testing.T) { t.Skip(pending) }

// TestDotMissingNamedDot checks the dot-diagram scenario "Named dot is
// missing".
func TestDotMissingNamedDot(t *testing.T) { t.Skip(pending) }

// TestDotLayoutTimeout checks the dot-diagram scenario "Layout timeout in a
// dot diagram".
func TestDotLayoutTimeout(t *testing.T) { t.Skip(pending) }
