// SPDX-License-Identifier: MIT
// Provenance: original

// Package yamlparser reads the YAML of YAML diagrams, as upstream's
// yaml.parser package: a subset of YAML (mappings, sequences, flow
// sequences, block values and comments), read line by line into a
// Monomorph, a value that becomes a scalar, a list or a map as the lines
// fill it, then converted into the JSON value that JSON diagrams draw.
//
// Upstream throws exceptions where the YAML is not of its subset. The
// functions of this package return them as an *Error, which the YAML
// factory of internal/jsondiagram draws as data that cannot be read.
package yamlparser
