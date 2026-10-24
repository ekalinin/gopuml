// SPDX-License-Identifier: MIT
// Provenance: original

// Command stdlib imports the standard library of the upstream jar into
// internal/stdlib: the .spm channel files of every library that
// stdlib.Excluded does not list, unchanged, and home.spm without the
// excluded names. The .sha1 files and README.md, which no code reads, are
// left out. It reads the stdlib/ folder of the sources jar that
// tools/upstream/fetch.sh unpacks (its files are those of the binary jar),
// then checks the result as TestUpstreamStdlib does:
//
//	go run ./tools/stdlib -src .cache/upstream/src
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ekalinin/gopuml/internal/stdlib"
)

func main() {
	src := flag.String("src", ".cache/upstream/src", "unpacked sources jar")
	out := flag.String("out", "internal/stdlib", "directory of the bundled files")
	flag.Parse()
	if err := run(filepath.Join(*src, "stdlib"), *out); err != nil {
		fmt.Fprintln(os.Stderr, "stdlib:", err)
		os.Exit(1)
	}
}

func run(src, out string) error {
	upstream, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, e := range upstream {
		if e.IsDir() && !stdlib.IsExcluded(e.Name()) {
			keep[e.Name()] = true
		}
	}
	// Folders that are no longer bundled, and every bundled .spm file, are
	// removed before the copy, so that files dropped upstream go too.
	ours, err := os.ReadDir(out)
	if err != nil {
		return err
	}
	for _, e := range ours {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(out, e.Name())
		if !keep[e.Name()] {
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
			continue
		}
		old, err := filepath.Glob(filepath.Join(dir, "*.spm"))
		if err != nil {
			return err
		}
		for _, f := range old {
			if err := os.Remove(f); err != nil {
				return err
			}
		}
	}
	libraries := 0
	for name := range keep {
		files, err := filepath.Glob(filepath.Join(src, name, "*.spm"))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(out, name), 0o755); err != nil {
			return err
		}
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(out, name, filepath.Base(f)), data, 0o644); err != nil {
				return err
			}
		}
		libraries++
	}
	home, err := os.ReadFile(filepath.Join(src, "home.spm"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "home.spm"), stdlib.FilterHome(home), 0o644); err != nil {
		return err
	}
	if problems := stdlib.CheckUpstream(os.DirFS(out), os.DirFS(src)); problems != nil {
		return fmt.Errorf("the imported files differ from upstream:\n%s", strings.Join(problems, "\n"))
	}
	fmt.Printf("imported %d libraries into %s (%d excluded)\n", libraries, out, len(stdlib.Excluded))
	return nil
}
