// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"maps"
	"math"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Thresholds from the upstream-compatibility spec. They may only be
// tightened over time.
const (
	SizeTolerance = 0.05
	MinSSIM       = 0.90
)

// Outcome is whether a diagram rendered or failed, and where it failed.
type Outcome struct {
	Failed bool
	// ErrorLine is the 1-based error line reported for a failure.
	ErrorLine int
}

// CompareOutcome checks that both implementations agree on success or
// failure, and on the error line of a failure.
func CompareOutcome(ref, got Outcome) error {
	switch {
	case ref.Failed && !got.Failed:
		return fmt.Errorf("reference fails at line %d but the diagram rendered", ref.ErrorLine)
	case !ref.Failed && got.Failed:
		return fmt.Errorf("reference renders but the diagram failed at line %d", got.ErrorLine)
	case ref.Failed && ref.ErrorLine != got.ErrorLine:
		return fmt.Errorf("error line %d, reference reports line %d", got.ErrorLine, ref.ErrorLine)
	}
	return nil
}

// SVGTexts returns the content of every <text> element in document order,
// with whitespace runs collapsed to single spaces. Empty texts are dropped.
func SVGTexts(svg []byte) ([]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(svg))
	dec.Strict = false
	var texts []string
	depth := 0 // nesting depth inside a <text> element
	var current strings.Builder
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing SVG: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "text" || depth > 0 {
				depth++
			}
		case xml.EndElement:
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 {
				if s := strings.Join(strings.Fields(current.String()), " "); s != "" {
					texts = append(texts, s)
				}
				current.Reset()
			}
		case xml.CharData:
			if depth > 0 {
				current.Write(t)
			}
		}
	}
	return texts, nil
}

// CompareTexts checks that got contains the same texts as ref, in any order.
func CompareTexts(ref, got []string) error {
	count := map[string]int{}
	for _, s := range ref {
		count[s]++
	}
	for _, s := range got {
		count[s]--
	}
	var missing, extra []string
	for _, s := range slices.Sorted(maps.Keys(count)) {
		for n := count[s]; n > 0; n-- {
			missing = append(missing, strconv.Quote(s))
		}
		for n := count[s]; n < 0; n++ {
			extra = append(extra, strconv.Quote(s))
		}
	}
	if len(missing) == 0 && len(extra) == 0 {
		return nil
	}
	var parts []string
	if len(missing) > 0 {
		parts = append(parts, "missing texts "+strings.Join(missing, ", "))
	}
	if len(extra) > 0 {
		parts = append(parts, "unexpected texts "+strings.Join(extra, ", "))
	}
	return errors.New(strings.Join(parts, "; "))
}

var svgRootSize = regexp.MustCompile(`<svg\b[^>]*?\swidth="([0-9.]+)(?:px)?"[^>]*?\sheight="([0-9.]+)(?:px)?"`)

// SVGSize returns the width and height declared on the SVG root element.
// Upstream's recorded image size can differ (for example with scale or
// dpi), so the document is the source of truth.
func SVGSize(svg []byte) (width, height int, ok bool) {
	m := svgRootSize.FindSubmatch(svg)
	if m == nil {
		return 0, 0, false
	}
	w, err1 := strconv.ParseFloat(string(m[1]), 64)
	h, err2 := strconv.ParseFloat(string(m[2]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return int(math.Round(w)), int(math.Round(h)), true
}

// HasSVGSourceMetadata reports whether the SVG embeds its source in a
// plantuml-src processing instruction.
func HasSVGSourceMetadata(svg []byte) bool { return bytes.Contains(svg, []byte("<?plantuml-src ")) }

// HasPNGSourceMetadata reports whether the PNG has a tEXt, zTXt or iTXt
// chunk with the keyword "plantuml".
func HasPNGSourceMetadata(data []byte) bool {
	const signature = "\x89PNG\r\n\x1a\n"
	if !bytes.HasPrefix(data, []byte(signature)) {
		return false
	}
	for p := len(signature); p+8 <= len(data); {
		n := int(data[p])<<24 | int(data[p+1])<<16 | int(data[p+2])<<8 | int(data[p+3])
		kind := string(data[p+4 : p+8])
		body := p + 8
		if n < 0 || body+n > len(data) {
			return false
		}
		if kind == "tEXt" || kind == "zTXt" || kind == "iTXt" {
			if bytes.HasPrefix(data[body:body+n], []byte("plantuml\x00")) {
				return true
			}
		}
		p = body + n + 4 // chunk data and CRC
	}
	return false
}

// CompareSize checks that the output width and height are each within
// tolerance (a fraction) of the reference.
func CompareSize(refW, refH, gotW, gotH int, tolerance float64) error {
	check := func(name string, ref, got int) error {
		if math.Abs(float64(got-ref)) > tolerance*float64(ref) {
			return fmt.Errorf("%s %d is not within %.0f%% of the reference %d", name, got, tolerance*100, ref)
		}
		return nil
	}
	return errors.Join(check("width", refW, gotW), check("height", refH, gotH))
}

// CompareText checks that two text outputs are identical apart from
// trailing whitespace at line ends.
func CompareText(ref, got string) error {
	norm := func(s string) []string {
		lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
		for i, l := range lines {
			lines[i] = strings.TrimRight(l, " \t")
		}
		return lines
	}
	r, g := norm(ref), norm(got)
	for i := range min(len(r), len(g)) {
		if r[i] != g[i] {
			return fmt.Errorf("line %d differs: got %q, reference %q", i+1, g[i], r[i])
		}
	}
	switch {
	case len(g) < len(r):
		return fmt.Errorf("output has %d lines, reference has %d", len(g), len(r))
	case len(g) > len(r):
		return fmt.Errorf("output has %d lines, reference has %d; first extra line %q", len(g), len(r), g[len(r)])
	}
	return nil
}

// CompareImages checks the structural similarity of got against ref after
// resampling got to the reference size.
func CompareImages(ref, got image.Image, minSSIM float64) (float64, error) {
	b := ref.Bounds()
	score := SSIM(gray(ref, b.Dx(), b.Dy()), gray(got, b.Dx(), b.Dy()), b.Dx(), b.Dy())
	if score < minSSIM {
		return score, fmt.Errorf("SSIM %.4f is below %.2f", score, minSSIM)
	}
	return score, nil
}

// gray converts img to grayscale luminance in [0,255], composited onto a
// white background and resampled (bilinear) to w x h.
func gray(img image.Image, w, h int) []float64 {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	src := make([]float64, sw*sh)
	for y := range sh {
		for x := range sw {
			r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			// Premultiplied colors over white: c + (1 - alpha) * white.
			white := float64(0xffff - a)
			lum := 0.299*(float64(r)+white) + 0.587*(float64(g)+white) + 0.114*(float64(bl)+white)
			src[y*sw+x] = lum / 0xffff * 255
		}
	}
	if sw == w && sh == h {
		return src
	}
	dst := make([]float64, w*h)
	for y := range h {
		fy := (float64(y)+0.5)*float64(sh)/float64(h) - 0.5
		y0 := clamp(int(math.Floor(fy)), sh)
		y1 := clamp(y0+1, sh)
		ty := fy - math.Floor(fy)
		for x := range w {
			fx := (float64(x)+0.5)*float64(sw)/float64(w) - 0.5
			x0 := clamp(int(math.Floor(fx)), sw)
			x1 := clamp(x0+1, sw)
			tx := fx - math.Floor(fx)
			top := src[y0*sw+x0]*(1-tx) + src[y0*sw+x1]*tx
			bottom := src[y1*sw+x0]*(1-tx) + src[y1*sw+x1]*tx
			dst[y*w+x] = top*(1-ty) + bottom*ty
		}
	}
	return dst
}

func clamp(v, n int) int { return min(max(v, 0), n-1) }

// SSIM returns the mean structural similarity of two grayscale images of
// size w x h, using an 11x11 Gaussian window with sigma 1.5 (Wang et al.).
func SSIM(a, b []float64, w, h int) float64 {
	const (
		c1 = (0.01 * 255) * (0.01 * 255)
		c2 = (0.03 * 255) * (0.03 * 255)
	)
	blur := gaussianBlur(w, h)
	mul := func(x, y []float64) []float64 {
		out := make([]float64, len(x))
		for i := range x {
			out[i] = x[i] * y[i]
		}
		return out
	}
	muA, muB := blur(a), blur(b)
	sAA, sBB, sAB := blur(mul(a, a)), blur(mul(b, b)), blur(mul(a, b))
	var sum float64
	for i := range muA {
		ma, mb := muA[i], muB[i]
		varA, varB, cov := sAA[i]-ma*ma, sBB[i]-mb*mb, sAB[i]-ma*mb
		sum += ((2*ma*mb + c1) * (2*cov + c2)) / ((ma*ma + mb*mb + c1) * (varA + varB + c2))
	}
	return sum / float64(len(muA))
}

// gaussianBlur returns a separable 11x11 Gaussian filter (sigma 1.5) for
// w x h images. Edges are handled by renormalizing the truncated kernel.
func gaussianBlur(w, h int) func([]float64) []float64 {
	const radius = 5
	var kernel [2*radius + 1]float64
	for i := range kernel {
		d := float64(i - radius)
		kernel[i] = math.Exp(-d * d / (2 * 1.5 * 1.5))
	}
	pass := func(in []float64, horizontal bool) []float64 {
		out := make([]float64, len(in))
		for y := range h {
			for x := range w {
				var sum, weight float64
				for k := -radius; k <= radius; k++ {
					sx, sy := x, y
					if horizontal {
						sx += k
					} else {
						sy += k
					}
					if sx < 0 || sx >= w || sy < 0 || sy >= h {
						continue
					}
					kw := kernel[k+radius]
					sum += kw * in[sy*w+sx]
					weight += kw
				}
				out[y*w+x] = sum / weight
			}
		}
		return out
	}
	return func(in []float64) []float64 { return pass(pass(in, true), false) }
}

// Rasterizer converts an SVG document to an image of the given size.
type Rasterizer interface {
	Rasterize(svg []byte, width, height int) (image.Image, error)
}

// Resvg rasterizes with the resvg command, mapping the generic families to
// the DejaVu fonts of the reference image.
type Resvg struct {
	// Path is the resvg executable; "resvg" when empty.
	Path string
	// FontsDir, when set, is the only font source, so rasterization does not
	// depend on the fonts installed on the host.
	FontsDir string
}

// Rasterize implements Rasterizer.
func (r Resvg) Rasterize(svg []byte, width, height int) (image.Image, error) {
	exe := r.Path
	if exe == "" {
		exe = "resvg"
	}
	args := []string{
		"--sans-serif-family", "DejaVu Sans",
		"--serif-family", "DejaVu Serif",
		"--monospace-family", "DejaVu Sans Mono",
		"-w", strconv.Itoa(width), "-h", strconv.Itoa(height),
		"--resources-dir", ".", // PlantUML embeds images as data URIs
	}
	if r.FontsDir != "" {
		args = append(args, "--skip-system-fonts", "--use-fonts-dir", r.FontsDir)
	}
	cmd := exec.Command(exe, append(args, "-", "-c")...)
	cmd.Stdin = bytes.NewReader(svg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("resvg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return png.Decode(bytes.NewReader(out))
}
