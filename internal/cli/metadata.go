// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/Run.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/png"
	"github.com/ekalinin/gopuml/internal/text"
)

const (
	metaHeaderNew = "<?plantuml-src "
	metaHeaderOld = "<!--SRC=["
	frame         = "------------------------"
)

// extractMetadata is Run.extractMetadata for each file: the source embedded
// in an SVG or PNG file, framed by dashed lines.
func (r *run) extractMetadata(files []string) {
	for _, f := range files {
		fmt.Fprintln(r.Stdout, frame)
		fmt.Fprintln(r.Stdout, f)
		fmt.Fprintln(r.Stdout)
		if strings.HasSuffix(f, ".svg") {
			if decoded, ok := r.metadataFromSVG(f); ok {
				fmt.Fprintln(r.Stdout, decoded)
			}
		} else {
			// Upstream prints "null" when the file has no source.
			data := "null"
			if file, err := os.Open(f); err != nil {
				fmt.Fprintln(r.Stderr, err)
			} else {
				if v, ok, err := png.MetadataTag(file, "plantuml"); err != nil {
					fmt.Fprintf(r.Stderr, "plantuml: %s: %v\n", f, err)
				} else if ok {
					data = v
				}
				file.Close()
			}
			fmt.Fprintln(r.Stdout, data)
		}
		fmt.Fprintln(r.Stdout, frame)
	}
}

// metadataFromSVG is Run.extractMetadataFromSvg: the decoded text of the
// last <?plantuml-src ...?> instruction, or of the older <!--SRC=[...]-->
// comment.
func (r *run) metadataFromSVG(f string) (string, bool) {
	data, err := os.ReadFile(f)
	if err != nil {
		fmt.Fprintln(r.Stderr, err)
		return "", false
	}
	svg := text.DecodeUTF8(data)
	var part string
	if i := strings.LastIndex(svg, metaHeaderNew); i > 0 {
		part = svg[i+len(metaHeaderNew):]
		if end := strings.Index(part, "?>"); end > 0 {
			part = part[:end]
		} else {
			part = ""
		}
	}
	if part == "" {
		if i := strings.LastIndex(svg, metaHeaderOld); i > 0 {
			part = svg[i+len(metaHeaderOld):]
			if end := strings.Index(part, "]"); end > 0 {
				part = strings.ReplaceAll(part[:end], "- -", "--")
			} else {
				part = ""
			}
		}
	}
	if part == "" {
		return "", false
	}
	decoded, err := code.DecodeProtected(part)
	if err != nil {
		fmt.Fprintf(r.Stderr, "plantuml: %s: %v\n", f, err)
		return "", false
	}
	return decoded, true
}

// sameMetadata is FileFormat.equalsMetadata for --skip-fresh: whether the
// existing output file already embeds metadata. For SVG, upstream compares
// the encoded text; encodings of different deflate implementations may
// differ byte-wise (design D10), so the decoded texts are compared.
func (r *run) sameMetadata(file, metadata string) bool {
	switch r.format {
	case "SVG":
		if _, err := os.Stat(file); err != nil {
			return false
		}
		existing, ok := r.metadataFromSVG(file)
		if !ok {
			return false
		}
		current, err := code.Decode(code.Encode(metadata))
		return err == nil && existing == current
	case "PNG":
		f, err := os.Open(file)
		if err != nil {
			return false
		}
		defer f.Close()
		existing, ok, err := png.MetadataTag(f, "plantuml")
		return err == nil && ok && existing == metadata
	}
	return false
}
