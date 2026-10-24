// SPDX-License-Identifier: MIT
// Provenance: original

package brotli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

// vector is a stream of testdata/, made with the brotli CLI 1.2.0, and its
// decoded content: the file plain, or the size and SHA-256 of a content that
// is not committed.
type vector struct {
	name   string
	plain  string
	size   int
	sha256 string
}

// The streams were made with "brotli -c -q 11 -w 24" unless the name says
// otherwise: prose-q5.br and prose-q1.br with "-q 5" and "-q 1";
// large-w16.br (a 4 MB repetitive text) with "-w 16", so that the ring
// buffer wraps; the random bytes of random-q0.br and random-w10.br with
// "-q 0" and "-q 5 -w 10", which store them in uncompressed meta-blocks;
// shuffled-w10.br and shuffled-q4-w10.br (40 KB of the sentences of
// prose.txt in random orders) with "-w 10" and "-q 4 -w 10", so that
// literals and dictionary words cross the end of the ring buffer.
var vectors = []vector{
	{name: "empty.br", size: 0, sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	{name: "hello.br", plain: "hello.txt"},
	{name: "prose.br", plain: "prose.txt"},
	{name: "prose-q5.br", plain: "prose.txt"},
	{name: "prose-q1.br", plain: "prose.txt"},
	largeVector,
	{name: "random-q0.br", size: 8000, sha256: randomSHA256},
	{name: "random-w10.br", size: 8000, sha256: randomSHA256},
	{name: "shuffled-w10.br", size: 40632, sha256: shuffledSHA256},
	{name: "shuffled-q4-w10.br", size: 40632, sha256: shuffledSHA256},
}

var largeVector = vector{name: "large-w16.br", size: 4134690, sha256: "617282654ef1583c942ed8b887c1d7a6e0a8744b8b13b8cd21acb64c8ae86bd2"}

const (
	randomSHA256   = "b26df339d8edee0cd910bc6c00bb400179a657f0a872be02393ad2cd68843748"
	shuffledSHA256 = "50d2111c26efa0720e0c8a9b74ba95667c8ed109b7f2fb298e7471cd4674ebf8"
)

func readTestdata(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// checkDecoded compares got with the content of v.
func checkDecoded(t *testing.T, v vector, got []byte) {
	t.Helper()
	if v.plain != "" {
		if want := readTestdata(t, v.plain); !bytes.Equal(got, want) {
			t.Errorf("decoded %d bytes, want the %d bytes of %s", len(got), len(want), v.plain)
		}
		return
	}
	if len(got) != v.size || sha256Hex(got) != v.sha256 {
		t.Errorf("decoded %d bytes with SHA-256 %s, want %d bytes with %s", len(got), sha256Hex(got), v.size, v.sha256)
	}
}

func TestDecodeVectors(t *testing.T) {
	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			data := readTestdata(t, v.name)
			got, err := io.ReadAll(NewReader(bytes.NewReader(data)))
			if err != nil {
				t.Fatal(err)
			}
			checkDecoded(t, v, got)
		})
	}
}

// TestDecodeSmallReads decodes with reads of a few bytes, on both sides, so
// that the decoder stops and resumes in every state that writes, and reads
// its source in small pieces, with io.EOF coming with the last bytes.
func TestDecodeSmallReads(t *testing.T) {
	for _, v := range vectors {
		if v.size > 100000 {
			continue
		}
		t.Run(v.name, func(t *testing.T) {
			data := readTestdata(t, v.name)
			sources := map[string]io.Reader{
				"one byte": iotest.OneByteReader(bytes.NewReader(data)),
				"half":     iotest.HalfReader(bytes.NewReader(data)),
				"data+EOF": iotest.DataErrReader(bytes.NewReader(data)),
				"whole":    bytes.NewReader(data),
			}
			for name, src := range sources {
				var got []byte
				r := NewReader(src)
				buf := make([]byte, 7)
				for {
					n, err := r.Read(buf)
					got = append(got, buf[:n]...)
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatalf("%s: %v", name, err)
					}
				}
				checkDecoded(t, v, got)
			}
		})
	}
}

// TestDecodeLargeVectorHalfReads decodes the large vector with reads of
// half the requested size, on both sides.
func TestDecodeLargeVectorHalfReads(t *testing.T) {
	data := readTestdata(t, "large-w16.br")
	got, err := io.ReadAll(iotest.HalfReader(NewReader(iotest.HalfReader(bytes.NewReader(data)))))
	if err != nil {
		t.Fatal(err)
	}
	checkDecoded(t, largeVector, got)
}

func TestReadEmptyBuffer(t *testing.T) {
	r := NewReader(bytes.NewReader(readTestdata(t, "hello.br")))
	if n, err := r.Read(nil); n != 0 || err != nil {
		t.Fatalf("Read(nil) = %d, %v, want 0, nil", n, err)
	}
	got, err := io.ReadAll(r)
	if err != nil || string(got) != "Hello, Brotli!\n" {
		t.Fatalf("ReadAll = %q, %v", got, err)
	}
	if n, err := r.Read(make([]byte, 10)); n != 0 || err != io.EOF {
		t.Fatalf("Read after the end = %d, %v, want 0, io.EOF", n, err)
	}
}

// TestMetadataBlock decodes a hand-made stream (accepted by the brotli CLI):
// a metadata meta-block, an uncompressed meta-block and an empty last one.
func TestMetadataBlock(t *testing.T) {
	stream := []byte{
		// WBITS 0 (16), ISLAST 0, MNIBBLES 3 (metadata), reserved 0,
		// MSKIPBYTES 1, MSKIPLEN-1 2, padding.
		0x2C, 0x01,
		'a', 'b', 'c',
		// ISLAST 0, MNIBBLES 0 (4), MLEN-1 1, ISUNCOMPRESSED 1, padding.
		0x08, 0x00, 0x08,
		'h', 'i',
		// ISLAST 1, ISLASTEMPTY 1.
		0x03,
	}
	got, err := io.ReadAll(NewReader(bytes.NewReader(stream)))
	if err != nil || string(got) != "hi" {
		t.Fatalf("ReadAll = %q, %v, want \"hi\"", got, err)
	}
}

// decodeLimited decodes data up to limit bytes and returns the error, nil
// when the stream ends or the limit is reached.
func decodeLimited(data []byte, limit int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(NewReader(bytes.NewReader(data)), limit))
}

func TestTruncatedStream(t *testing.T) {
	for _, name := range []string{"hello.br", "prose.br", "prose-q1.br", "random-q0.br", "random-w10.br"} {
		data := readTestdata(t, name)
		for n := range len(data) {
			if _, err := decodeLimited(data[:n], 1<<26); err == nil {
				t.Errorf("%s cut to %d of %d bytes: no error", name, n, len(data))
			}
		}
	}
	data := readTestdata(t, "large-w16.br")
	_, err := decodeLimited(data[:len(data)/2], 1<<26)
	if err == nil {
		t.Fatal("large-w16.br cut in half: no error")
	}
	if !strings.HasPrefix(err.Error(), "Brotli stream decoding failed: ") {
		t.Errorf("error %q, want upstream's message", err)
	}
}

func TestCorruptedStream(t *testing.T) {
	data := readTestdata(t, "prose.br")
	// Flipped bytes that upstream's decoder rejects.
	for _, pos := range []int{0, 1, 2, len(data) / 2, len(data) - 1} {
		corrupted := bytes.Clone(data)
		corrupted[pos] ^= 0xFF
		if _, err := decodeLimited(corrupted, 1<<26); err == nil {
			t.Errorf("byte %d flipped: no error", pos)
		}
	}
	// Every flipped byte either fails or decodes to something, without
	// panicking or looping forever. Brotli has no checksum, so that some
	// corruptions decode.
	for _, name := range []string{"hello.br", "prose.br", "prose-q1.br"} {
		data := readTestdata(t, name)
		for pos := range len(data) {
			for _, mask := range []byte{0x01, 0x10, 0xFF} {
				corrupted := bytes.Clone(data)
				corrupted[pos] ^= mask
				decodeLimited(corrupted, 1<<20)
			}
		}
	}
}

func TestErrorIsSticky(t *testing.T) {
	data := readTestdata(t, "prose.br")
	r := NewReader(bytes.NewReader(data[:len(data)/2]))
	_, err := io.ReadAll(r)
	if err == nil {
		t.Fatal("no error")
	}
	if n, again := r.Read(make([]byte, 10)); n != 0 || again != err {
		t.Fatalf("Read after the error = %d, %v, want 0, %v", n, again, err)
	}
}

func TestSourceError(t *testing.T) {
	errSource := errors.New("source failed")
	data := readTestdata(t, "large-w16.br")
	src := io.MultiReader(bytes.NewReader(data[:len(data)/2]), iotest.ErrReader(errSource))
	_, err := io.ReadAll(NewReader(src))
	if !errors.Is(err, errSource) {
		t.Fatalf("error %v, want one that wraps %v", err, errSource)
	}

	_, err = io.ReadAll(NewReader(iotest.ErrReader(errSource)))
	if !errors.Is(err, errSource) || !strings.HasPrefix(err.Error(), "Brotli decoder initialization failed: ") {
		t.Fatalf("error %v, want the initialization failure of %v", err, errSource)
	}
}

func TestReservedWindowBits(t *testing.T) {
	// WBITS 1, 000, 001: the reserved window of 9 bits.
	_, err := io.ReadAll(NewReader(bytes.NewReader([]byte{0x11})))
	if err == nil || err.Error() != "Brotli decoder initialization failed: Invalid 'windowBits' code" {
		t.Fatalf("error %v, want the invalid window bits", err)
	}
}

// emptyReader returns no byte and no error.
type emptyReader struct{}

func (emptyReader) Read([]byte) (int, error) { return 0, nil }

func TestSourceWithoutProgress(t *testing.T) {
	_, err := io.ReadAll(NewReader(emptyReader{}))
	if !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("error %v, want io.ErrNoProgress", err)
	}
}

func FuzzReader(f *testing.F) {
	for _, v := range vectors {
		if v.size <= 100000 {
			f.Add(readTestdata(f, v.name))
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		decodeLimited(data, 1<<20)
	})
}
