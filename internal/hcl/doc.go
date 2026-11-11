// SPDX-License-Identifier: MIT
// Provenance: original

// Package hcl draws HCL diagrams, as upstream's hcl package: a subset of
// HCL (blocks of attributes, with lists, objects and function calls) read
// into the JSON value that internal/jsondiagram draws, with the yamlDiagram
// styles and without the style blocks and the title of the source, as
// upstream does.
//
// Upstream throws exceptions where the HCL is not of its subset. The parser
// of this package returns them as an *Error, which the factory draws as
// data that cannot be read.
package hcl
