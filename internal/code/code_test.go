// SPDX-License-Identifier: MIT
// Provenance: original

package code

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"strings"
	"testing"
)

const (
	helloSource = "@startuml\nAlice -> Bob : hello\n@enduml"
	hiSource    = "@startuml\nAlice -> Bob : hi\n@enduml"
)

// Strings encoded by upstream v1.2026.8, from the conformance references.
func TestDecodeUpstreamEncoded(t *testing.T) {
	tests := map[string]string{
		"Syp9J4vLqBLJSCfFKh1Io4ZDoSa70000": helloSource,
		"Syp9J4vLqBLJSCfFKh1IoCW40000":     hiSource,
		"~0Syp9J4vLqBLJSCfFKh1IoCW40000":   hiSource,
		"Syp9J4vLqBLJSCfF0W00":             "@startuml\nAlice -> Bob\n@enduml",
		"~h416c696365202d3e20426f62":       "@startuml\nAlice -> Bob\n@enduml",
		"~h416C696365202D3E20426F62":       "@startuml\nAlice -> Bob\n@enduml",
		"AyaioKbLk33_OiE57HSsNTXwiOdBCISpELL1rqx1AJz9mKeXu_3Ad9nyBX175oW8aRLIoCy60000": "@startuml\ntitle Привет\nAlice -> Bob : héllo\nBob --> Alice : ok\n@enduml",
	}
	for code, want := range tests {
		got, err := Decode(code)
		if err != nil || got != want {
			t.Errorf("Decode(%q) = %q, %v; want %q", code, got, err, want)
		}
	}
}

func TestEncodeRoundTrip(t *testing.T) {
	for _, src := range []string{
		helloSource,
		"Alice -> Bob",
		"@startuml\ntitle Привет\nAlice -> Bob : héllo 😀\n@enduml",
		"@startmindmap\n* root\n@endmindmap",
		strings.Repeat("A -> B : some long text\n", 200),
	} {
		code := Encode(src)
		if strings.Trim(code, "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_") != "" {
			t.Errorf("Encode(%q) uses characters outside the alphabet: %q", src, code)
		}
		got, err := Decode(code)
		if err != nil {
			t.Fatalf("Decode(Encode(%q)): %v", src, err)
		}
		want := decompressSource(compressSource(src))
		if got != want {
			t.Errorf("round trip of %q = %q, want %q", src, got, want)
		}
	}
}

func TestEncodeStripsStartAndEnd(t *testing.T) {
	// The @startuml/@enduml lines are not encoded, and come back on decode.
	if Encode(helloSource) != Encode("Alice -> Bob : hello") {
		t.Error("@startuml/@enduml must not change the encoding")
	}
	if Encode("") != "" {
		t.Error("an empty source encodes to an empty string")
	}
}

func TestDecodeOtherForms(t *testing.T) {
	src := []byte("Alice -> Bob")
	var z bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&z, zlib.HuffmanOnly)
	zw.Write(src)
	zw.Close()
	var g bytes.Buffer
	gw := gzip.NewWriter(&g)
	gw.Write(src)
	gw.Close()
	want := "@startuml\nAlice -> Bob\n@enduml"
	for name, code := range map[string]string{
		"~1 legacy Huffman": "~1" + encodeASCII(z.Bytes()),
		"unprefixed zlib":   encodeASCII(z.Bytes()),
		"~g gzip":           "~g" + encodeASCII(g.Bytes()),
	} {
		got, err := Decode(code)
		if err != nil || got != want {
			t.Errorf("%s: %q, %v", name, got, err)
		}
	}
}

func TestDecodeInvalid(t *testing.T) {
	for _, code := range []string{"@@not-valid@@", "~1@@@@", "~hzz", "~g0000", "é"} {
		if got, err := Decode(code); !errors.Is(err, ErrNoCompression) {
			t.Errorf("Decode(%q) = %q, %v; want ErrNoCompression", code, got, err)
		}
	}
}

func TestDecodeProtected(t *testing.T) {
	got, err := DecodeProtected("@@not-valid@@")
	if err != nil || !strings.Contains(got, "This URL does not look like DEFLATE data.") || !strings.Contains(got, "http://www.plantuml.com/plantuml/png/~1@@not-valid@@") {
		t.Errorf("bad deflate URL: %q, %v", got, err)
	}
	got, _ = DecodeProtected("~1" + strings.Repeat("x", 100))
	if !strings.Contains(got, "This URL does not look like HUFFMAN data.") || !strings.Contains(got, "\n"+strings.Repeat("x", 80)+"\n"+strings.Repeat("x", 20)+"\n") {
		t.Errorf("bad Huffman URL: %q", got)
	}
	if got, err := DecodeProtected("Syp9J4vLqBLJSCfF0W00"); err != nil || got != "@startuml\nAlice -> Bob\n@enduml" {
		t.Errorf("valid code: %q, %v", got, err)
	}
}

func TestUnzipLocalHeader(t *testing.T) {
	// A stored entry "a.txt" with content "Alice -> Bob", plus trailing zeros
	// as left by the 4-character grouping of the alphabet.
	content := "Alice -> Bob"
	header := []byte("PK\x03\x04\x14\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	header = append(header, byte(len(content)), 0, 0, 0, byte(len(content)), 0, 0, 0, 5, 0, 0, 0)
	data := append(append(header, "a.txt"+content...), 0, 0)
	got, err := unzip(data)
	if err != nil || string(got) != content {
		t.Errorf("unzip = %q, %v", got, err)
	}
}
