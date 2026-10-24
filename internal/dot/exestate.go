// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/dot/ExeState.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package dot

import (
	"fmt"
	"path/filepath"
)

// ExeState tells whether a dot executable path is usable.
type ExeState int

// States of a dot executable path, as upstream's ExeState enum.
const (
	NullUndefined ExeState = iota
	OK
	DoesNotExist
	IsADirectory
	NotAFile
	CannotBeRead
)

// CheckFile returns the state of the dot executable at dotExe. An empty
// dotExe stands for upstream's null File.
func CheckFile(dotExe string, h Host) ExeState {
	if dotExe == "" {
		return NullUndefined
	}
	fi, err := h.stat(dotExe)
	switch {
	case err != nil:
		return DoesNotExist
	case fi.IsDir():
		return IsADirectory
	case !fi.Mode().IsRegular():
		return NotAFile
	case !h.canRead(fi):
		return CannotBeRead
	}
	return OK
}

// TextMessage returns the description of the state without the file name.
func (s ExeState) TextMessage() string {
	switch s {
	case OK:
		return "Dot executable OK"
	case NullUndefined:
		return "No dot executable found"
	case DoesNotExist:
		return "Dot executable does not exist"
	case IsADirectory:
		return "Dot executable should be an executable, not a directory"
	case NotAFile:
		return "Dot executable is not a valid file"
	case CannotBeRead:
		return "Dot executable cannot be read"
	}
	panic(fmt.Sprintf("dot: unknown ExeState %d", int(s)))
}

// TextMessageFor returns the description of the state naming the absolute
// path of exe (upstream's getTextMessage(File)).
func (s ExeState) TextMessageFor(exe string) string {
	abs := absolutePath(exe)
	switch s {
	case OK:
		return "File " + abs + " OK"
	case NullUndefined:
		return NullUndefined.TextMessage()
	case DoesNotExist:
		return "File " + abs + " does not exist"
	case IsADirectory:
		return "File " + abs + " should be an executable, not a directory"
	case NotAFile:
		return "File " + abs + " is not a valid file"
	case CannotBeRead:
		return "File " + abs + " cannot be read"
	}
	panic(fmt.Sprintf("dot: unknown ExeState %d", int(s)))
}

// String returns the upstream enum name.
func (s ExeState) String() string {
	switch s {
	case NullUndefined:
		return "NULL_UNDEFINED"
	case OK:
		return "OK"
	case DoesNotExist:
		return "DOES_NOT_EXIST"
	case IsADirectory:
		return "IS_A_DIRECTORY"
	case NotAFile:
		return "NOT_A_FILE"
	case CannotBeRead:
		return "CANNOT_BE_READ"
	}
	return fmt.Sprintf("ExeState(%d)", int(s))
}

// absolutePath mirrors java.io.File.getAbsolutePath: relative paths are
// resolved against the current directory.
func absolutePath(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

// javaFileString mirrors string concatenation with a java.io.File that may be
// null: the path itself, or "null".
func javaFileString(p string) string {
	if p == "" {
		return "null"
	}
	return p
}
