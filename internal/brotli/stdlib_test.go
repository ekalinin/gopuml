// SPDX-License-Identifier: MIT
// Provenance: original

package brotli

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// stdlibDir is the data directory of internal/stdlib, which holds the
// channel files of the bundled libraries as <lib>/<channel>.spm.
const stdlibDir = "../stdlib"

// stdlibEntry is a line of testdata/stdlib.txt: the size and SHA-256 of a
// channel file of the plantuml-mit 1.2026.9 jar, decoded with the brotli CLI.
type stdlibEntry struct {
	sha256 string
	size   int64
}

func readStdlibFixture(t testing.TB) map[string]stdlibEntry {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "stdlib.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	entries := map[string]stdlibEntry{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 3 {
			t.Fatalf("bad fixture line %q", sc.Text())
		}
		size, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		entries[fields[2]] = stdlibEntry{sha256: fields[0], size: size}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return entries
}

// decodeFile decodes the file at path and returns the size and SHA-256 of
// the decoded bytes.
func decodeFile(path string) (int64, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, NewReader(f))
	return n, hex.EncodeToString(h.Sum(nil)), err
}

// TestStdlibChannels decodes every channel file found in the stdlib data
// directory and compares it with the fixture.
func TestStdlibChannels(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(stdlibDir, "*", "*.spm"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skipf("no channel file in %s", stdlibDir)
	}
	fixture := readStdlibFixture(t)
	for _, path := range paths {
		name := filepath.ToSlash(filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path)))
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			want, ok := fixture[name]
			if !ok {
				t.Fatalf("%s is not in testdata/stdlib.txt", name)
			}
			start := time.Now()
			size, sum, err := decodeFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if size != want.size || sum != want.sha256 {
				t.Errorf("decoded %d bytes with SHA-256 %s, want %d bytes with %s", size, sum, want.size, want.sha256)
			}
			t.Logf("decoded %d bytes in %v", size, time.Since(start))
		})
	}
}

// BenchmarkStdlib decodes the two largest channel files of the stdlib data
// directory.
func BenchmarkStdlib(b *testing.B) {
	for _, name := range []string{"ibm/image.spm", "tupadr3/sprite.spm"} {
		b.Run(name, func(b *testing.B) {
			path := filepath.Join(stdlibDir, filepath.FromSlash(name))
			data, err := os.ReadFile(path)
			if err != nil {
				b.Skip(err)
			}
			want := readStdlibFixture(b)[name]
			b.SetBytes(want.size)
			for b.Loop() {
				n, err := io.Copy(io.Discard, NewReader(bytes.NewReader(data)))
				if err != nil || n != want.size {
					b.Fatalf("decoded %d bytes, %v", n, err)
				}
			}
		})
	}
}
