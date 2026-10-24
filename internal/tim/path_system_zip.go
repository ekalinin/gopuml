// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/nio/NFolderZip.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/nio/InputFileZip.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"path"
	"path/filepath"
	"strings"
)

// nFolderZip is a location inside a zip archive: the root, or a directory
// of its entries. The archive is read through the file access, so the
// security profile applies to it.
type nFolderZip struct {
	fa      *fileAccess
	zipFile string
	// pathInsideZip is the current path inside the archive, relative,
	// without leading slash, "" for the root.
	pathInsideZip string
}

func newNFolderZip(fa *fileAccess, zipFile string) *nFolderZip {
	return &nFolderZip{fa: fa, zipFile: zipFile}
}

func (d *nFolderZip) getInputFile(nameOrPath string) (inputFile, error) {
	return &inputFileZip{fa: d.fa, zipFile: d.zipFile, entryName: d.toZipEntryName(nameOrPath)}, nil
}

// contains reports whether nameOrPath is a file (not a directory) of the
// archive. Unlike getInputFile, whose file fails when read, it checks that
// the entry exists.
func (d *nFolderZip) contains(nameOrPath string) bool {
	entry := d.toZipEntryName(nameOrPath)
	if entry == "" {
		return false
	}
	ze := d.fa.zipEntry(d.zipFile, entry)
	return ze != nil && !ze.FileInfo().IsDir()
}

func (d *nFolderZip) getSubfolder(nameOrPath string) (nFolder, error) {
	return &nFolderZip{fa: d.fa, zipFile: d.zipFile, pathInsideZip: d.toZipEntryName(nameOrPath)}, nil
}

func (d *nFolderZip) String() string { return filepath.Base(d.zipFile) + "!" + d.pathInsideZip }

// toZipEntryName resolves a path inside the archive: a slash entry name
// without leading slash.
func (d *nFolderZip) toZipEntryName(child string) string {
	resolved := child
	if !path.IsAbs(child) {
		resolved = path.Join(d.pathInsideZip, child)
	}
	resolved = path.Clean(resolved)
	if resolved == "." {
		resolved = ""
	}
	return strings.TrimLeft(strings.ReplaceAll(resolved, `\`, "/"), "/")
}

// zipEntry returns the entry of the archive named entry (or "entry/", as
// ZipFile.getEntry looks for), nil when the archive cannot be read or has
// no such entry.
func (fa *fileAccess) zipEntry(zipFile, entry string) *zip.File {
	data := fa.openFile(zipFile)
	if data == nil {
		return nil
	}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		// Not a valid archive
		return nil
	}
	var dir *zip.File
	for _, f := range r.File {
		if f.Name == entry {
			return f
		}
		if f.Name == entry+"/" {
			dir = f
		}
	}
	return dir
}

// inputFileZip is an entry of a zip archive.
type inputFileZip struct {
	fa        *fileAccess
	zipFile   string
	entryName string
}

func (f *inputFileZip) newInputStream() ([]byte, error) {
	ze := f.fa.zipEntry(f.zipFile, f.entryName)
	if ze == nil {
		return nil, errors.New("Entry not found in ZIP: " + f.entryName + " (" + f.zipFile + ")")
	}
	rc, err := ze.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (f *inputFileZip) String() string { return f.zipFile + "!" + f.entryName }

// getParentFolder returns the directory of the entry inside the archive.
func (f *inputFileZip) getParentFolder() (nFolder, error) {
	return newNFolderZip(f.fa, f.zipFile).getSubfolder(zipParentFrom(f.entryName))
}

// zipParentFrom returns the directory of an entry: "a/b" for "a/b/c.txt",
// "" for "a.txt".
func zipParentFrom(entryName string) string {
	idx := strings.LastIndexByte(strings.ReplaceAll(entryName, `\`, "/"), '/')
	if idx <= 0 {
		return ""
	}
	return entryName[:idx]
}
