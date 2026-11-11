// SPDX-License-Identifier: MIT
// Provenance: original

// Package jsondiagram draws JSON, YAML and HCL diagrams, as upstream's
// jsondiagram package: every object and array of the data is a box of rows,
// and the boxes are linked from left to right by arrows from the rows whose
// values they are.
//
// Upstream's jsondiagram and yaml packages refer to each other, so they form
// this one Go package: files from yaml are prefixed yaml_. The YAML parser
// (internal/yamlparser) and the HCL diagrams (internal/hcl) have no cycle
// and stay separate.
//
// Upstream lays the boxes out with its Smetana port, on records whose field
// sizes it sets with a hack of Smetana. This package gives the same graph to
// gogv, with tables of empty cells in place of the records, and reads the
// layout back from gogv's json0 output (smetanaforjson.go).
package jsondiagram
