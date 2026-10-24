// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/builtin/FileExists.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/security/SFile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"path/filepath"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// fileExists is %file_exists(path): whether the file or directory exists
// and the security profile allows reading it (SFile.exists). As upstream,
// which creates a plain SFile, a relative path is resolved against the
// working directory of the process, not the directory of the source; with
// an fs.FS it is resolved against the root of the FS.
type fileExists struct{ simpleReturnFunction }

func (fileExists) Signature() TFunctionSignature { return NewTFunctionSignature("%file_exists", 1) }

func (fileExists) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (fileExists) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	path := values[0].String()
	if path == "" {
		// new File("").exists() is false
		return TValueFromBoolean(false), nil
	}
	if !filepath.IsAbs(path) {
		path = context.fa.launchDir() + "/" + path
	}
	return TValueFromBoolean(context.fa.canRead(javaNormalizePath(path))), nil
}

// javaNormalizePath is the normalization of java.io.File on Unix: repeated
// slashes become one and a trailing slash is removed. Unlike
// filepath.Clean, "." and ".." are kept, so the security check sees them.
func javaNormalizePath(path string) string {
	var b strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] == '/' && i > 0 && path[i-1] == '/' {
			continue
		}
		b.WriteByte(path[i])
	}
	s := b.String()
	if len(s) > 1 {
		s = strings.TrimSuffix(s, "/")
	}
	return s
}
