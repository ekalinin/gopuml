// SPDX-License-Identifier: MIT
// Provenance: original

// Command plantuml is a drop-in replacement for java -jar plantuml.jar.
package main

import (
	"os"

	"github.com/ekalinin/gopuml/internal/cli"
)

func main() {
	cmd := &cli.Command{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr, Engine: cliEngine{}}
	os.Exit(cmd.Run(os.Args[1:]))
}
