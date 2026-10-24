// SPDX-License-Identifier: MIT
// Provenance: original

package png

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/png"
	"strings"
	"testing"
)

// withChunks returns a 1x1 PNG with the chunks inserted before IEND.
func withChunks(t *testing.T, chunks ...[]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	iend := len(data) - 12
	out := append([]byte(nil), data[:iend]...)
	for _, c := range chunks {
		out = append(out, c...)
	}
	return append(out, data[iend:]...)
}

func chunk(kind string, data []byte) []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, uint32(len(data)))
	b.WriteString(kind)
	b.Write(data)
	binary.Write(&b, binary.BigEndian, crc32.ChecksumIEEE(append([]byte(kind), data...)))
	return b.Bytes()
}

func deflate(s string) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(s))
	w.Close()
	return b.Bytes()
}

func iTXt(key, text string, compressed bool) []byte {
	data := []byte(key + "\x00")
	if compressed {
		data = append(data, 1, 0)
		data = append(data, "\x00\x00"...)
		return chunk("iTXt", append(data, deflate(text)...))
	}
	data = append(data, 0, 0)
	data = append(data, "\x00\x00"...)
	return chunk("iTXt", append(data, text...))
}

func TestMetadataTag(t *testing.T) {
	src := "@startuml\nAlice -> Bob : héllo\n@enduml\n\nPlantUML version 1.2026.9"
	tests := []struct {
		name   string
		chunks [][]byte
		want   string
		ok     bool
	}{
		{"compressed iTXt", [][]byte{iTXt("plantuml", src, true)}, src, true},
		{"plain iTXt", [][]byte{iTXt("plantuml", src, false)}, src, true},
		{"tEXt", [][]byte{chunk("tEXt", []byte("plantuml\x00caf\xe9"))}, "café", true},
		{"zTXt", [][]byte{chunk("zTXt", append([]byte("plantuml\x00\x00"), deflate("z")...))}, "z", true},
		{"tEXt before iTXt", [][]byte{iTXt("plantuml", "itxt", false), chunk("tEXt", []byte("plantuml\x00text"))}, "text", true},
		{"other keyword", [][]byte{chunk("tEXt", []byte("Software\x00x"))}, "", false},
		{"none", nil, "", false},
	}
	for _, tt := range tests {
		got, ok, err := MetadataTag(bytes.NewReader(withChunks(t, tt.chunks...)), "plantuml")
		if err != nil || ok != tt.ok || got != tt.want {
			t.Errorf("%s: %q, %v, %v", tt.name, got, ok, err)
		}
	}
	if _, _, err := MetadataTag(strings.NewReader("<svg/>"), "plantuml"); err != ErrNotPNG {
		t.Errorf("not a PNG: %v", err)
	}
}
