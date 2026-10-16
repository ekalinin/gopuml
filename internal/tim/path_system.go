// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/PathSystem.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/NFolder.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/NFolderRegular.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/InputFile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/InputFileUrl.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/NFolderStdlib.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/nio/InputFileStdlib.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/FileSystem.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"context"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ekalinin/plantuml-in-go/internal/security"
)

// fileAccess reads files and URLs through the security layer (design D9).
// Files are absolute paths: in the operating system when the Access has no
// FS, else slash paths inside the FS.
type fileAccess struct {
	access *security.Access
}

func (fa *fileAccess) stat(abs string) (fs.FileInfo, error) {
	if fa.access.FS == nil {
		return os.Stat(abs)
	}
	name := strings.TrimPrefix(path.Clean("/"+filepath.ToSlash(abs)), "/")
	if name == "" {
		// the root of the FS
		name = "."
	}
	return fs.Stat(fa.access.FS, name)
}

// canRead is SFile.canRead: the profile allows the file and it exists
// (directories included, as in Java).
func (fa *fileAccess) canRead(abs string) bool {
	if !fa.access.Config.FileOK(abs) {
		return false
	}
	_, err := fa.stat(abs)
	return err == nil
}

// isFile reports whether abs exists and is not a directory.
func (fa *fileAccess) isFile(abs string) bool {
	if !fa.access.Config.FileOK(abs) {
		return false
	}
	st, err := fa.stat(abs)
	return err == nil && !st.IsDir()
}

// openFile is SFile.openFile: the content, or nil when the file is not
// allowed or cannot be read.
func (fa *fileAccess) openFile(abs string) []byte {
	data, err := fa.access.ReadFile(abs)
	if err != nil {
		return nil
	}
	return data
}

// canonical is SFile.getCanonicalFile: the cleaned path without symbolic
// links.
func (fa *fileAccess) canonical(abs string) string {
	abs = filepath.Clean(abs)
	if fa.access.FS == nil {
		if real, err := filepath.EvalSymlinks(abs); err == nil {
			return real
		}
	}
	return abs
}

// launchDir is the directory that "::" paths are resolved from.
func (fa *fileAccess) launchDir() string {
	if fa.access.FS == nil {
		if wd, err := os.Getwd(); err == nil {
			return wd
		}
	}
	return "/"
}

// openURL is SURL.openStream: the content, or nil when the profile denies
// the URL or the fetch fails.
func (fa *fileAccess) openURL(u string) []byte {
	data, err := fa.access.ReadURL(context.Background(), u)
	if err != nil {
		return nil
	}
	return data
}

// createURL is SURL.create: nil unless s is an http or https URL.
func createURL(s string) (string, bool) {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "", false
	}
	if _, err := url.Parse(s); err != nil || strings.ContainsAny(s, " \"<>\\^`{|}") {
		return "", false
	}
	return s, true
}

// errNotAllowed is the IOException of a file that cannot be opened.
var errNotAllowed = errors.New("cannot open file")

// inputFile is a file that can be included.
type inputFile interface {
	// newInputStream returns the content; nil without error is Java's null
	// stream (an SFile that cannot be opened).
	newInputStream() ([]byte, error)
	getParentFolder() (nFolder, error)
	String() string
}

// nFolder is a directory that includes are resolved against.
type nFolder interface {
	getInputFile(nameOrPath string) (inputFile, error)
	getSubfolder(nameOrPath string) (nFolder, error)
}

// sFile is a file of the file system (upstream's SFile as an InputFile).
type sFile struct {
	fa   *fileAccess
	path string
}

func (f *sFile) newInputStream() ([]byte, error) { return f.fa.openFile(f.path), nil }

func (f *sFile) getParentFolder() (nFolder, error) {
	return &nFolderRegular{fa: f.fa, dir: filepath.Dir(f.fa.canonical(f.path))}, nil
}

func (f *sFile) String() string { return f.path }

// nFolderRegular is a directory of the file system.
type nFolderRegular struct {
	fa  *fileAccess
	dir string
}

func (d *nFolderRegular) getInputFile(nameOrPath string) (inputFile, error) {
	if filepath.IsAbs(nameOrPath) {
		if d.fa.canRead(nameOrPath) {
			return &sFile{fa: d.fa, path: filepath.Clean(nameOrPath)}, nil
		}
		return nil, nil
	}
	if p := filepath.Join(d.dir, nameOrPath); d.fa.canRead(p) {
		return &sFile{fa: d.fa, path: p}, nil
	}
	for _, dir := range d.fa.access.Config.IncludePaths {
		if p := filepath.Join(dir, nameOrPath); d.fa.canRead(p) {
			return &sFile{fa: d.fa, path: d.fa.canonical(p)}, nil
		}
	}
	return nil, nil
}

func (d *nFolderRegular) getSubfolder(nameOrPath string) (nFolder, error) {
	return &nFolderRegular{fa: d.fa, dir: filepath.Join(d.dir, nameOrPath)}, nil
}

func (d *nFolderRegular) String() string { return "NFolderRegular[" + d.dir + "]" }

// inputFileURL is a URL that can be included.
type inputFileURL struct {
	fa  *fileAccess
	url string
}

func (f *inputFileURL) newInputStream() ([]byte, error) {
	data := f.fa.openURL(f.url)
	if data == nil {
		return nil, errors.New("Cannot open URL " + f.url)
	}
	return data, nil
}

func (f *inputFileURL) getParentFolder() (nFolder, error) {
	throwRuntime("UnsupportedOperationException", "")
	return nil, nil
}

func (f *inputFileURL) String() string { return f.url }

// nFolderStdlib is a directory inside a standard library.
type nFolderStdlib struct {
	stdlib           *stdlibLibrary
	pathInsideStdlib string
}

func (d *nFolderStdlib) getInputFile(nameOrPath string) (inputFile, error) {
	return &inputFileStdlib{stdlib: d.stdlib, stdlibPath: path.Join(d.pathInsideStdlib, nameOrPath)}, nil
}

func (d *nFolderStdlib) getSubfolder(nameOrPath string) (nFolder, error) {
	return &nFolderStdlib{stdlib: d.stdlib, pathInsideStdlib: path.Join(d.pathInsideStdlib, nameOrPath)}, nil
}

// inputFileStdlib is a file of a standard library.
type inputFileStdlib struct {
	stdlib     *stdlibLibrary
	stdlibPath string
}

func (f *inputFileStdlib) newInputStream() ([]byte, error) {
	name := strings.ReplaceAll(strings.ToLower(f.stdlibPath), ".puml", "")
	data := f.stdlib.loadPumlResource(name)
	if data == nil {
		throwRuntime("NullPointerException", "")
	}
	return data, nil
}

// getParentFolder returns the folder named like the file itself, as
// upstream does.
func (f *inputFileStdlib) getParentFolder() (nFolder, error) {
	return &nFolderStdlib{stdlib: f.stdlib, pathInsideStdlib: path.Clean(f.stdlibPath)}, nil
}

func (f *inputFileStdlib) String() string { return f.stdlib.name + "!" + f.stdlibPath }

// PathSystem resolves the files named by includes against the current
// directory, which changes while an included file is executed.
type PathSystem struct {
	fa            *fileAccess
	stdlib        *Stdlib
	currentFolder nFolder
}

func (p *PathSystem) withCurrentDir(parentFile nFolder) *PathSystem {
	return &PathSystem{fa: p.fa, stdlib: p.stdlib, currentFolder: parentFile}
}

func (p *PathSystem) getFile(filename string) (inputFile, error) { return p.getInputFile(filename) }

func (p *PathSystem) getInputFile(pth string) (inputFile, error) {
	if strings.HasPrefix(pth, "http://") || strings.HasPrefix(pth, "https://") {
		u, ok := createURL(pth)
		if !ok {
			return nil, errors.New("Cannot open URL " + pth)
		}
		return &inputFileURL{fa: p.fa, url: u}, nil
	}
	if strings.HasPrefix(pth, "<") && strings.HasSuffix(pth, ">") && len(pth) >= 2 {
		full := strings.ToLower(pth[1 : len(pth)-1])
		slash := strings.IndexByte(full, '/')
		if slash < 0 {
			throwRuntime("StringIndexOutOfBoundsException", "")
		}
		lib := p.stdlib.retrieve(full[:slash])
		return &inputFileStdlib{stdlib: lib, stdlibPath: path.Clean(full[slash+1:])}, nil
	}
	if rel, ok := strings.CutPrefix(pth, "::"); ok {
		// non-standard syntax: resolve from process launch directory
		rel = strings.TrimPrefix(rel, "/")
		target := filepath.Join(p.fa.launchDir(), rel)
		if p.fa.access.Config.FileOK(target) {
			return &sFile{fa: p.fa, path: target}, nil
		}
		return nil, nil
	}
	if rel, ok := strings.CutPrefix(pth, "~/"); ok {
		// Expand to the user's home directory
		if p.fa.access.FS != nil {
			return nil, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, nil
		}
		target := filepath.Join(home, rel)
		if p.fa.access.Config.FileOK(target) {
			return &sFile{fa: p.fa, path: target}, nil
		}
		return nil, nil
	}
	return p.currentFolder.getInputFile(pth)
}

// fileSystemGetFile is FileSystem.getFile, used by !import: the file
// relative to the directory of the source, then to the include path.
func (fa *fileAccess) fileSystemGetFile(currentDir, nameOrPath string) string {
	if filepath.IsAbs(nameOrPath) {
		return fa.canonical(nameOrPath)
	}
	filecurrent := filepath.Join(currentDir, nameOrPath)
	if fa.canRead(filecurrent) {
		return fa.canonical(filecurrent)
	}
	for _, dir := range fa.access.Config.IncludePaths {
		if file := filepath.Join(dir, nameOrPath); fa.canRead(file) {
			return fa.canonical(file)
		}
	}
	return filecurrent
}
