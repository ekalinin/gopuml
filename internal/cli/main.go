// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"io"
	"os"
)

// Prog is the program name shown in the help.
const Prog = "plantuml"

// Main runs the command line with standard input and returns the exit
// status. Rendering is added by later tasks; until then a render request
// reports that it is not implemented.
func Main(args []string, stdout, stderr io.Writer) int {
	return (&Command{Stdin: os.Stdin, Stdout: stdout, Stderr: stderr}).Run(args)
}
