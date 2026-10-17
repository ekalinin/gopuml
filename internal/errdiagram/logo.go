// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/version/PSystemVersion.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package errdiagram

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"image"
	"image/draw"
	"image/png"
	"sync"

	"github.com/ekalinin/gopuml/internal/klimt"
)

// logoPNG is upstream's version/logo.png as Java's ImageIO writes it in SVG
// data URLs (same pixels, different encoding from the resource file).
//
//go:embed logo.png
var logoPNG []byte

// logoPixels decodes the logo once; the result is never changed (each image
// gets a copy).
var logoPixels = sync.OnceValue(func() *klimt.PortableImageARGB {
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		panic("errdiagram: cannot decode the embedded logo: " + err.Error())
	}
	b := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, b.Min, draw.Src)
	result := klimt.NewPortableImageARGB(b.Dx(), b.Dy())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			p := nrgba.Pix[nrgba.PixOffset(x, y):]
			result.SetRGB(x, y, int32(uint32(p[3])<<24|uint32(p[0])<<16|uint32(p[1])<<8|uint32(p[2])))
		}
	}
	return result
})

// logoImage is the logo: its pixels, and upstream's PNG encoding as long as
// they are not changed.
type logoImage struct {
	*klimt.PortableImageARGB
	changed bool
}

// plantumlImage is PSystemVersion.getPlantumlImage: a new copy of the logo.
func plantumlImage() klimt.PortableImage {
	pixels := logoPixels()
	cp := pixels.Subimage(0, 0, pixels.Width(), pixels.Height()).(*klimt.PortableImageARGB)
	return &logoImage{PortableImageARGB: cp}
}

// SetRGB sets the ARGB value of a pixel.
func (l *logoImage) SetRGB(x, y int, rgb int32) {
	l.changed = true
	l.PortableImageARGB.SetRGB(x, y, rgb)
}

// ToPngDataURL returns the image as a PNG data URL, with upstream's bytes
// while the pixels are those of the logo.
func (l *logoImage) ToPngDataURL() string {
	if l.changed {
		return l.PortableImageARGB.ToPngDataURL()
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(logoPNG)
}
