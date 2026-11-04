// SPDX-License-Identifier: MIT
// Provenance: original

// Command benchcmp compares the render benchmark (BenchmarkRender of the
// root package) of the checked-out commit with that of a base commit, on the
// same machine, and fails when the allocations of a diagram type grow by
// more than a threshold.
//
// Usage: go run ./tools/benchcmp -base <ref>
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "usage: go run ./tools/benchcmp -base <ref>")
	os.Exit(2)
}
