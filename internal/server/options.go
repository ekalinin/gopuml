// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/RenderRequest.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/PicoWebServer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliOptions.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliFlag.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ekalinin/gopuml"
	"github.com/ekalinin/gopuml/internal/klimt/font"
)

// renderRequest is the JSON body of POST /render (upstream's RenderRequest):
// {"source": "...", "options": ["--svg", ...]}.
type renderRequest struct {
	source  string
	options []string
}

// parseRenderRequest is RenderRequest.fromJson with the error messages of
// PicoWebServer.handleRenderRequest. Like upstream, "source" must be a
// string, "options" is optional and must be an array of strings, member
// names are case-sensitive and other members are ignored.
func parseRenderRequest(body []byte) (*renderRequest, error) {
	if len(body) == 0 {
		return nil, errors.New("No request body")
	}
	req, err := decodeRenderRequest(body)
	if err != nil {
		return nil, fmt.Errorf("Error parsing request json: %w", err)
	}
	return req, nil
}

func decodeRenderRequest(body []byte) (*renderRequest, error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(body, &members); err != nil {
		return nil, err
	}
	if members == nil {
		return nil, errors.New("not an object")
	}
	req := &renderRequest{}
	src, ok := members["source"]
	if !ok {
		return nil, errors.New(`missing "source"`)
	}
	if err := jsonString(src, &req.source); err != nil {
		return nil, fmt.Errorf(`"source": %w`, err)
	}
	if opts, ok := members["options"]; ok {
		var list []json.RawMessage
		if string(opts) == "null" || json.Unmarshal(opts, &list) != nil {
			return nil, errors.New(`"options" is not an array`)
		}
		req.options = make([]string, len(list))
		for i, v := range list {
			if err := jsonString(v, &req.options[i]); err != nil {
				return nil, fmt.Errorf(`"options"[%d]: %w`, i, err)
			}
		}
	}
	return req, nil
}

// jsonString decodes a JSON string, refusing null and other types.
func jsonString(data json.RawMessage, s *string) error {
	if len(data) == 0 || data[0] != '"' {
		return errors.New("not a string")
	}
	return json.Unmarshal(data, s)
}

// formatFlag is a CliFlag that selects an output format, matched ignoring
// case.
type formatFlag struct {
	names  []string
	format gopuml.Format
	// unsupported is the upstream format name when this implementation
	// cannot produce the format.
	unsupported string
}

// formatFlags are the format flags of CliFlag, in declaration order.
var formatFlags = []formatFlag{
	{names: []string{"--eps", "-teps", "-eps"}, unsupported: "EPS"},
	{names: []string{"--teps:text", "-teps:text", "-eps:text"}, unsupported: "EPS_TEXT"},
	{names: []string{"--html", "-thtml", "-html"}, unsupported: "HTML"},
	{names: []string{"--latex", "-tlatex", "-latex"}, unsupported: "LATEX"},
	{names: []string{"--latex-nopreamble", "-tlatex:nopreamble", "-latex:nopreamble"}, unsupported: "LATEX_NO_PREAMBLE"},
	{names: []string{"--obfuscate", "-cypher"}, unsupported: "OBFUSCATE"},
	{names: []string{"--pdf", "-tpdf", "-pdf"}, unsupported: "PDF"},
	{names: []string{"--png", "-tpng", "-png"}, format: gopuml.PNG},
	{names: []string{"--preproc", "-preproc"}, unsupported: "PREPROC"},
	{names: []string{"--scxml", "-tscxml"}, unsupported: "SCXML"},
	{names: []string{"--svg", "-tsvg", "-svg"}, format: gopuml.SVG},
	{names: []string{"--null"}, unsupported: "NULL"},
	{names: []string{"--txt", "-ttxt", "-txt"}, format: gopuml.TXT},
	{names: []string{"--utxt", "-tutxt", "-utxt"}, format: gopuml.UTXT},
	{names: []string{"--vdx", "-tvdx", "-vdx"}, unsupported: "VDX"},
	{names: []string{"--xmi", "-txmi", "-xmi"}, unsupported: "XMI_STANDARD"},
	{names: []string{"--xmi:argo", "-xmi:argo"}, unsupported: "XMI_ARGO"},
	{names: []string{"--xmi:custom", "-xmi:custom"}, unsupported: "XMI_CUSTOM"},
	{names: []string{"--xmi:script", "-xmi:script"}, unsupported: "XMI_SCRIPT"},
	{names: []string{"--xmi:star", "-xmi:star"}, unsupported: "XMI_STAR"},
	{names: []string{"--base64"}, unsupported: "BASE64"},
	{names: []string{"--braille"}, unsupported: "BRAILLE_PNG"},
}

// Formats of FileFormat.fromCli that this implementation produces.
var cliFormats = map[font.FileFormat]gopuml.Format{
	font.FormatPNG:  gopuml.PNG,
	font.FormatSVG:  gopuml.SVG,
	font.FormatATXT: gopuml.TXT,
	font.FormatUTXT: gopuml.UTXT,
}

// Flags that take the next argument as their value (BINARY_NEXT_ARGUMENT_VALUE).
var (
	flagFormat   = []string{"--format", "-f"}
	flagTheme    = []string{"--theme", "-theme"}
	flagDefine   = []string{"--define"}
	flagPragma   = []string{"--pragma"}
	flagSkin     = []string{"--skinparam"}
	flagFilename = []string{"-filename"}
)

// Boolean flags.
var (
	flagDarkMode   = []string{"--dark-mode", "-darkmode"}
	flagNoMetadata = []string{"--disable-metadata", "-nometadata"}
)

func matchFlag(arg string, names []string) bool {
	for _, n := range names {
		if strings.EqualFold(arg, n) {
			return true
		}
	}
	return false
}

// keyValue is one -DKEY=VALUE style entry.
type keyValue struct{ key, value string }

// putKeyValue updates a key in place or appends it, like the LinkedHashMap
// of CliParsed.
func putKeyValue(list []keyValue, s string) []keyValue {
	k, v, _ := strings.Cut(s, "=")
	for i := range list {
		if list[i].key == k {
			list[i].value = v
			return list
		}
	}
	return append(list, keyValue{k, v})
}

// parseOptions turns the CLI flags of a POST /render request into render
// options, as CliOptions does for the flags that affect one render:
// the output format, --dark-mode, --disable-metadata, -D/--define,
// --theme, -P/--pragma, -S/--skinparam and -filename. Other arguments are
// ignored, like unknown CLI arguments. Flags that read local files
// (--config, -I, -filedir) are ignored too, because they would bypass the
// server's security profile. A format this implementation cannot produce
// is an error.
func parseOptions(args []string) (gopuml.Options, error) {
	var (
		opts                    gopuml.Options
		formatName              *string
		themes                  []string
		defines, pragmas, skins []keyValue
		formatCounts            = make([]int, len(formatFlags))
	)
	i := 0
	next := func() (string, bool) {
		i++
		if i < len(args) {
			return args[i], true
		}
		return "", false
	}
	// The -DKEY=VALUE style flags, matched as case-sensitive prefixes.
	inline := []struct {
		prefix string
		list   *[]keyValue
	}{{"-D", &defines}, {"-P", &pragmas}, {"-S", &skins}}
	for ; i < len(args); i++ {
		arg := args[i]
		switch {
		case matchFlag(arg, flagDarkMode):
			opts.DarkMode = true
		case matchFlag(arg, flagNoMetadata):
			opts.NoMetadata = true
		case matchFlag(arg, flagFormat):
			if v, ok := next(); ok {
				formatName = &v
			}
		case matchFlag(arg, flagTheme):
			if v, ok := next(); ok {
				themes = append(themes, v)
			}
		case matchFlag(arg, flagDefine):
			if v, ok := next(); ok {
				defines = putKeyValue(defines, v)
			}
		case matchFlag(arg, flagPragma):
			if v, ok := next(); ok {
				pragmas = putKeyValue(pragmas, v)
			}
		case matchFlag(arg, flagSkin):
			if v, ok := next(); ok {
				skins = putKeyValue(skins, v)
			}
		case matchFlag(arg, flagFilename):
			if v, ok := next(); ok {
				opts.Filename = v
			}
		case formatFlagIndex(arg) >= 0:
			formatCounts[formatFlagIndex(arg)]++
		default:
			for _, in := range inline {
				if !strings.HasPrefix(arg, in.prefix) {
					continue
				}
				s := arg[len(in.prefix):]
				if s == "" {
					v, ok := next()
					if !ok {
						return opts, fmt.Errorf("Missing value after %s", in.prefix)
					}
					s = v
				}
				*in.list = putKeyValue(*in.list, s)
				break
			}
		}
	}

	format, err := selectFormat(formatName, formatCounts)
	if err != nil {
		return opts, err
	}
	opts.Format = format
	for _, d := range defines {
		opts.Defines = append(opts.Defines, gopuml.Define{Name: d.key, Value: d.value})
	}
	// The config lines of the CliOptions constructor, in its order.
	for _, t := range themes {
		opts.Config = append(opts.Config, "!theme "+t)
	}
	for _, p := range pragmas {
		opts.Config = append(opts.Config, "!pragma "+p.key+" "+p.value)
	}
	for _, s := range skins {
		opts.Config = append(opts.Config, "skinparamlocked "+s.key+" "+s.value)
	}
	return opts, nil
}

// selectFormat is CliOptions.getFileFormatOption: the --format value when
// FileFormat.fromCli knows it, otherwise the first format flag, in
// declaration order, given exactly once, otherwise PNG.
func selectFormat(name *string, counts []int) (gopuml.Format, error) {
	if name != nil {
		if f, ok := font.FileFormatFromCli(*name); ok {
			if pf, ok := cliFormats[f]; ok {
				return pf, nil
			}
			return 0, unsupportedFormat(f.String())
		}
	}
	for j, n := range counts {
		if n != 1 {
			continue
		}
		if u := formatFlags[j].unsupported; u != "" {
			return 0, unsupportedFormat(u)
		}
		return formatFlags[j].format, nil
	}
	return gopuml.PNG, nil
}

func unsupportedFormat(name string) error {
	return fmt.Errorf("Unsupported output format: %s", name)
}

func formatFlagIndex(arg string) int {
	for j, f := range formatFlags {
		if matchFlag(arg, f.names) {
			return j
		}
	}
	return -1
}
