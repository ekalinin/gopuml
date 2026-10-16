// SPDX-License-Identifier: MIT
// Provenance: original

package png

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"image"
	"image/png"
	"io"
	"strings"
	"testing"
)

// chunkKinds returns the kinds of the chunks of a PNG stream.
func chunkKinds(t *testing.T, data []byte) []string {
	t.Helper()
	if !bytes.HasPrefix(data, signature) {
		t.Fatal("no PNG signature")
	}
	var kinds []string
	for p := len(signature); p+8 <= len(data); {
		n := int(binary.BigEndian.Uint32(data[p:]))
		kinds = append(kinds, string(data[p+4:p+8]))
		p += 12 + n
	}
	return kinds
}

// chunkData returns the data of the first chunk of a kind.
func chunkData(data []byte, kind string) []byte {
	for p := len(signature); p+8 <= len(data); {
		n := int(binary.BigEndian.Uint32(data[p:]))
		if string(data[p+4:p+8]) == kind {
			return data[p+8 : p+8+n]
		}
		p += 12 + n
	}
	return nil
}

func decodeARGB(t *testing.T, data []byte) (image.Image, []uint32) {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	out := make([]uint32, 0, b.Dx()*b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a != 0 {
				r, g, bl = r*0xFFFF/a, g*0xFFFF/a, bl*0xFFFF/a
			}
			out = append(out, (a>>8)<<24|(r>>8)<<16|(g>>8)<<8|bl>>8)
		}
	}
	return img, out
}

func TestWriteMetadataChunks(t *testing.T) {
	img := RasterImage{Width: 2, Height: 1, Pix: []uint32{0xFFFFFFFF, 0xFF000000}}
	var buf bytes.Buffer
	meta := "@startuml\nAlice -> Bob : привет 😀\n@enduml"
	if err := Write(&buf, img, meta, 96); err != nil {
		t.Fatal(err)
	}
	kinds := strings.Join(chunkKinds(t, buf.Bytes()), " ")
	if !strings.HasPrefix(kinds, "IHDR tEXt iTXt IDAT") || !strings.HasSuffix(kinds, "IEND") {
		t.Errorf("chunks %s", kinds)
	}
	if got, ok, err := MetadataTag(bytes.NewReader(buf.Bytes()), "plantuml"); got != meta || !ok || err != nil {
		t.Errorf("metadata %q %v %v", got, ok, err)
	}
	if got, ok, _ := MetadataTag(bytes.NewReader(buf.Bytes()), "copyleft"); got != copyleft || !ok {
		t.Errorf("copyleft %q", got)
	}
	// iTXt: keyword, compressed flag 1, method 0, empty language tag and
	// translated keyword, zlib data.
	it := chunkData(buf.Bytes(), "iTXt")
	if !bytes.HasPrefix(it, []byte("plantuml\x00\x01\x00\x00\x00")) {
		t.Errorf("iTXt header %q", it[:min(len(it), 13)])
	}
	zr, err := zlib.NewReader(bytes.NewReader(it[13:]))
	if err != nil {
		t.Fatal(err)
	}
	text, _ := io.ReadAll(zr)
	if string(text) != meta {
		t.Errorf("iTXt text %q", text)
	}

	buf.Reset()
	if err := Write(&buf, img, "", 96); err != nil {
		t.Fatal(err)
	}
	if kinds := strings.Join(chunkKinds(t, buf.Bytes()), " "); strings.Contains(kinds, "iTXt") ||
		strings.Contains(kinds, "pHYs") {
		t.Errorf("chunks without metadata: %s", kinds)
	}
}

func TestWriteDpi(t *testing.T) {
	img := RasterImage{Width: 1, Height: 1, Pix: []uint32{0xFFFFFFFF}}
	var buf bytes.Buffer
	if err := Write(&buf, img, "x", 300); err != nil {
		t.Fatal(err)
	}
	kinds := strings.Join(chunkKinds(t, buf.Bytes()), " ")
	if !strings.HasPrefix(kinds, "IHDR pHYs tEXt iTXt IDAT") {
		t.Errorf("chunks %s", kinds)
	}
	// Upstream gives HorizontalPixelSize = dpi / 0.0254 / 1000, so the chunk
	// holds 1000 / 11.811024 pixels per meter.
	phys := chunkData(buf.Bytes(), "pHYs")
	if binary.BigEndian.Uint32(phys) != 85 || binary.BigEndian.Uint32(phys[4:]) != 85 || phys[8] != 1 {
		t.Errorf("pHYs %v", phys)
	}
}

func TestWriteQuantization(t *testing.T) {
	// 40 colors in distinct cubes, the first with two sub-colors: an opaque
	// image becomes a palette of the most frequent color of each cube.
	pix := make([]uint32, 0, 44)
	for i := range 40 {
		pix = append(pix, 0xFF000000|uint32(i%20*8)<<16|uint32(i/20*8)<<8)
	}
	pix = append(pix, 0xFF010203, 0xFF010203, 0xFF010203, 0x12000000) // alpha ignored for RGB
	img := RasterImage{Width: 44, Height: 1, Pix: pix}
	var buf bytes.Buffer
	if err := Write(&buf, img, "", 96); err != nil {
		t.Fatal(err)
	}
	decoded, got := decodeARGB(t, buf.Bytes())
	if _, ok := decoded.(*image.Paletted); !ok {
		t.Fatalf("not a palette image: %T", decoded)
	}
	if got[0] != 0xFF010203 || got[43] != 0xFF010203 || got[1] != 0xFF080000 {
		t.Errorf("pixels %x %x %x", got[0], got[43], got[1])
	}

	// Fewer than 32 cubes: the low bit of each channel takes its high bit.
	img = RasterImage{Width: 2, Height: 1, Pix: []uint32{0xFFFEFECE, 0xFF7F80AA}}
	buf.Reset()
	if err := Write(&buf, img, "", 96); err != nil {
		t.Fatal(err)
	}
	if _, got := decodeARGB(t, buf.Bytes()); got[0] != 0xFFFFFFCF || got[1] != 0xFF7E81AB {
		t.Errorf("packed pixels %x %x", got[0], got[1])
	}

	// More than 256 cubes: packed too.
	pix = pix[:0]
	for i := range 300 {
		pix = append(pix, 0xFF000000|uint32(i%32*8)<<16|uint32(i/32*8)<<8)
	}
	img = RasterImage{Width: 300, Height: 1, Pix: pix}
	buf.Reset()
	if err := Write(&buf, img, "", 96); err != nil {
		t.Fatal(err)
	}
	if decoded, _ := decodeARGB(t, buf.Bytes()); func() bool { _, ok := decoded.(*image.Paletted); return ok }() {
		t.Error("palette image for more than 256 cubes")
	}

	// An image with alpha keeps it (packed as well).
	img = RasterImage{Width: 3, Height: 1, Pix: []uint32{0x00000000, 0x80FF0000, 0x7F112233}, HasAlpha: true}
	buf.Reset()
	if err := Write(&buf, img, "", 96); err != nil {
		t.Fatal(err)
	}
	if _, got := decodeARGB(t, buf.Bytes()); got[0]>>24 != 0 || got[1] != 0x81FF0000 || got[2]>>24 != 0x7E {
		t.Errorf("alpha pixels %x %x %x", got[0], got[1], got[2])
	}
}

func TestQuantify555(t *testing.T) {
	// The most frequent sub-color wins, the first one on ties; transparent
	// pixels share one cube, last in the palette.
	pix := []uint32{0xFF000001, 0xFF000002, 0xFF000002, 0x00FFFFFF, 0x10FFFFFF}
	for i := 1; i < 40; i++ {
		pix = append(pix, 0xFF000000|uint32(i*8)<<8)
	}
	palette, indices, ok := quantify555(pix)
	if !ok {
		t.Fatal("not quantified")
	}
	if palette[indices[0]] != 0xFF000002 || palette[indices[3]] != 0 || indices[3] != indices[4] ||
		int(indices[3]) != len(palette)-1 {
		t.Errorf("palette %x indices %v", palette, indices[:5])
	}
	if _, _, ok := quantify555([]uint32{0xFF000000, 0xFFFFFFFF}); ok {
		t.Error("two colors quantified")
	}
	if representativeARGB(transparentCube, 0) != 0 || compressPackedARGB(0x7F8055AA) != 0x7E8154AB {
		t.Error("helpers")
	}
}

func TestWriteErrors(t *testing.T) {
	if err := Write(io.Discard, RasterImage{Width: 0, Height: 1}, "", 96); err == nil {
		t.Error("empty image")
	}
	if err := Write(io.Discard, RasterImage{Width: 2, Height: 2, Pix: []uint32{0}}, "", 96); err == nil {
		t.Error("short pixels")
	}
	if err := insertChunks(io.Discard, []byte("x"), nil); err == nil {
		t.Error("short stream")
	}
	if got := string(utf8Bytes("a\xffb")); got != "a?b" {
		t.Errorf("utf8 %q", got)
	}
	if got := string(latin1Bytes("aé€")); got != "a\xe9?" {
		t.Errorf("latin1 %q", got)
	}
}
