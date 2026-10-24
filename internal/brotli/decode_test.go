// SPDX-License-Identifier: MIT
// Provenance: original

package brotli

import (
	"bytes"
	"io"
	"testing"
)

// TestDictionary compares the unpacked dictionary with the one of the
// brotli C library 1.2.0 (RFC 7932, appendix A).
func TestDictionary(t *testing.T) {
	data := getDictionaryData()
	if len(data) != dictionarySize {
		t.Fatalf("dictionary has %d bytes, want %d", len(data), dictionarySize)
	}
	const want = "20e42eb1b511c21806d4d227d07e5dd06877d8ce7b3a817f378f313653f35c70"
	if got := sha256Hex(data); got != want {
		t.Fatalf("dictionary SHA-256 %s, want %s", got, want)
	}
}

// TestDictionaryUsed checks that the prose vectors refer to the dictionary:
// they do not decode with a blank one.
func TestDictionaryUsed(t *testing.T) {
	saved := bytes.Clone(getDictionaryData())
	clear(dictionaryData)
	t.Cleanup(func() { copy(dictionaryData, saved) })
	want := readTestdata(t, "prose.txt")
	for _, name := range []string{"prose.br", "prose-q5.br"} {
		got, err := io.ReadAll(NewReader(bytes.NewReader(readTestdata(t, name))))
		if err == nil && bytes.Equal(got, want) {
			t.Errorf("%s decodes without the dictionary", name)
		}
	}
}

func TestTransformDictionaryWord(t *testing.T) {
	// The first word of length 4 is "time"; the transforms are those of
	// RFC 7932, appendix B.
	tests := []struct {
		transform int
		want      string
	}{
		{0, "time"},
		{1, "time "},
		{2, " time "},
		{3, "ime"},
		{4, "Time "},
		{5, "time the "},
		{9, "Time"},
	}
	for _, tt := range tests {
		dst := make([]byte, maxTransformedWordLength)
		n := transformDictionaryWord(dst, 0, getDictionaryData(), dictionaryOffsetsByLength[4], 4, tt.transform)
		if got := string(dst[:n]); got != tt.want {
			t.Errorf("transform %d = %q, want %q", tt.transform, got, tt.want)
		}
	}
}

// TestUncompressedMetaBlock checks that the random vectors hold
// uncompressed meta-blocks: the stream holds the decoded bytes as they are.
func TestUncompressedMetaBlock(t *testing.T) {
	for _, name := range []string{"random-q0.br", "random-w10.br"} {
		stream := readTestdata(t, name)
		got, err := io.ReadAll(NewReader(bytes.NewReader(stream)))
		if err != nil {
			t.Fatal(err)
		}
		if sha256Hex(got) != randomSHA256 {
			t.Fatalf("%s: decoded %d bytes with SHA-256 %s", name, len(got), sha256Hex(got))
		}
		if !bytes.Contains(stream, got[:256]) {
			t.Errorf("%s: no uncompressed meta-block", name)
		}
	}
}
