// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/LoadJson.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/preproc/Stdlib.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"errors"
	"strings"

	"github.com/ekalinin/gopuml/internal/json"
	"github.com/ekalinin/gopuml/internal/text"
)

// loadJSON is %load_json(source[, default[, charset]]): the JSON data of a
// file, a URL or a standard library resource ("<library/path>"), or the
// default JSON (by default "{}") when the source has no data. Files and URLs
// are read through the security layer (design D9): a denied source has no
// data.
type loadJSON struct{ simpleReturnFunction }

func (loadJSON) Signature() TFunctionSignature { return NewTFunctionSignature("%load_json", 3) }

func (loadJSON) CanCover(nbArg int, _ []string) bool { return nbArg == 1 || nbArg == 2 || nbArg == 3 }

func (loadJSON) ExecuteReturnFunction(context *TContext, _ TMemory, location text.StringLocated,
	values []*TValue, _ map[string]*TValue) (*TValue, error) {
	path := values[0].String()
	charset := "UTF-8"
	if len(values) == 3 {
		charset = values[2].String()
	}
	data, err := context.loadJSONData(path, charset)
	if err == nil && data == nil {
		defaultJSON := "{}"
		if len(values) > 1 {
			defaultJSON = values[1].String()
		}
		data, err = json.Parse(defaultJSON)
	}
	if pe, ok := errors.AsType[*json.ParseError](err); ok {
		return nil, newEaterException("JSON parse issue in source "+path+" on location "+pe.Location.String(), location)
	}
	if err != nil {
		return nil, newEaterException("JSON encoding issue in source "+path+": "+err.Error(), location)
	}
	return TValueFromJSON(data), nil
}

// errUnsupportedEncoding is UnsupportedEncodingException; its message is the
// charset name.
type errUnsupportedEncoding struct{ charset string }

func (e *errUnsupportedEncoding) Error() string { return e.charset }

// loadJSONData is LoadJson.loadJsonData: nil when the source has no data.
func (c *TContext) loadJSONData(path, charset string) (json.Value, error) {
	if len(path) > 0 && (path[0] == '<' || path[0] == '>') {
		u := toUTF16(path)
		if len(u) < 2 {
			throwRuntime("StringIndexOutOfBoundsException", "")
		}
		return c.stdlib.getJSONResource(fromUTF16(u[1 : len(u)-1]))
	}
	var byteData []byte
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if u, ok := createURL(path); ok {
			byteData = c.fa.openURL(u)
		}
	} else if file := c.fa.fileSystemGetFile(c.currentDir, path); c.fa.isFile(file) {
		byteData = c.fa.openFile(file)
	}
	if len(byteData) == 0 {
		// no length, no data (we want the default)
		return nil, nil
	}
	cs, ok := text.LookupCharset(charset)
	if !ok {
		return nil, &errUnsupportedEncoding{charset}
	}
	return json.Parse(cs.Decode(byteData))
}
