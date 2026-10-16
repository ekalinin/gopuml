// SPDX-License-Identifier: MIT
// Provenance: original

// Command provenance checks that every Go source file declares where its code
// comes from, so that only MIT-licensed upstream code is translated.
//
// Each checked file must contain, within its first lines, either
//
//	// Provenance: original
//
// or one or more lines naming the upstream MIT source it was translated from,
// followed by the upstream copyright notice:
//
//	// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/code/AsciiEncoder.java
//	// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.
//
// Code translated from fdlibm (the C library behind Java's StrictMath)
// names its source file and keeps Sun's notice:
//
//	// Provenance: fdlibm 5.3 e_atan2.c
//	// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.
//
// Upstream paths inside components that are not part of the MIT variant
// (ditaa, jcckit, sudoku, zxing, webp) are rejected. When -src points to the
// unpacked sources jar, every referenced upstream path must also exist there.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	provenancePrefix = "// Provenance: "
	upstreamPrefix   = "plantuml-mit 1.2026.8 "
	derivedNotice    = "// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License."
	fdlibmPrefix     = "fdlibm 5.3 "
	fdlibmNotice     = "// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved."
	headerLines      = 15
)

// excludedSegments are path fragments of upstream components that are not
// part of the MIT variant and must never be translated.
var excludedSegments = []string{
	"/ditaa/",
	"org/stathissideris/",
	"jcckit/",
	"/sudoku/",
	"/zext/",
	"/webp/",
}

// Problem is a single provenance violation.
type Problem struct {
	File    string
	Message string
}

func (p Problem) String() string { return p.File + ": " + p.Message }

// checkFile validates the provenance header of one Go file. srcRoot may be
// empty, in which case upstream paths are not checked for existence.
func checkFile(path, content, srcRoot string) []Problem {
	var problems []Problem
	add := func(format string, args ...any) {
		problems = append(problems, Problem{File: path, Message: fmt.Sprintf(format, args...)})
	}

	var original bool
	var upstream, fdlibm []string
	var hasNotice, hasFdlibmNotice bool
	sc := bufio.NewScanner(strings.NewReader(content))
	for i := 0; i < headerLines && sc.Scan(); i++ {
		line := strings.TrimRight(sc.Text(), " \t\r")
		if line == derivedNotice {
			hasNotice = true
			continue
		}
		if line == fdlibmNotice {
			hasFdlibmNotice = true
			continue
		}
		value, ok := strings.CutPrefix(line, provenancePrefix)
		if !ok {
			continue
		}
		switch {
		case value == "original":
			original = true
		case strings.HasPrefix(value, upstreamPrefix):
			upstream = append(upstream, strings.TrimSpace(strings.TrimPrefix(value, upstreamPrefix)))
		case strings.HasPrefix(value, fdlibmPrefix) && strings.HasSuffix(value, ".c"):
			fdlibm = append(fdlibm, value)
		default:
			add("unrecognized provenance %q", value)
		}
	}

	switch {
	case original && len(upstream)+len(fdlibm) > 0:
		add("declares both original and upstream provenance")
	case !original && len(upstream)+len(fdlibm) == 0 && len(problems) == 0:
		add("missing provenance header")
	case len(upstream) > 0 && !hasNotice:
		add("translated file lacks the upstream copyright notice %q", derivedNotice)
	case len(fdlibm) > 0 && !hasFdlibmNotice:
		add("fdlibm file lacks Sun's notice %q", fdlibmNotice)
	}

	for _, up := range upstream {
		if !strings.HasSuffix(up, ".java") {
			add("upstream path %q is not a Java source file", up)
			continue
		}
		for _, seg := range excludedSegments {
			if strings.Contains("/"+up, seg) {
				add("upstream path %q belongs to an excluded component (%s)", up, strings.Trim(seg, "/"))
			}
		}
		if srcRoot != "" {
			if _, err := os.Stat(filepath.Join(srcRoot, filepath.FromSlash(up))); err != nil {
				add("upstream path %q does not exist in %s", up, srcRoot)
			}
		}
	}
	return problems
}

// checkTree walks the given roots and checks every .go file.
func checkTree(roots []string, srcRoot string) ([]Problem, int, error) {
	var problems []Problem
	var count int
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			count++
			problems = append(problems, checkFile(filepath.ToSlash(path), string(data), srcRoot)...)
			return nil
		})
		if err != nil {
			return nil, count, err
		}
	}
	return problems, count, nil
}

func main() {
	src := flag.String("src", "", "directory of the unpacked plantuml-mit sources jar (optional)")
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"internal", "cmd"}
	}

	problems, count, err := checkTree(roots, *src)
	if err != nil {
		fmt.Fprintln(os.Stderr, "provenance:", err)
		os.Exit(2)
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "provenance: %d problem(s) in %d file(s) checked\n", len(problems), count)
		os.Exit(1)
	}
	fmt.Printf("provenance: %d file(s) OK\n", count)
}
