// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/tim/EaterTheme.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"errors"
	"strings"

	"github.com/ekalinin/gopuml/internal/text"
)

// EaterTheme is "!theme name [from location]".
type EaterTheme struct {
	Eater
	realName   string
	name       string
	from       *string
	pathSystem *PathSystem
}

func newEaterTheme(sl text.StringLocated, pathSystem *PathSystem) *EaterTheme {
	return &EaterTheme{Eater: newEater(sl), pathSystem: pathSystem}
}

func (e *EaterTheme) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!theme"); err != nil {
		return err
	}
	e.skipSpaces()
	e.name = e.eatAllToEnd()
	if x := strings.Index(lowerASCII(e.name), " from "); x != -1 {
		fromTmp := text.JavaTrim(e.name[x+len(" from "):])
		from, err := context.applyFunctionsAndVariables(memory, text.NewStringLocated(fromTmp, e.getLineLocation()))
		if err != nil {
			return err
		}
		if from == nil {
			throwRuntime("NullPointerException", "")
		}
		e.from = from
		e.name = text.JavaTrim(e.name[:x])
	}
	realName, err := context.applyFunctionsAndVariables(memory, text.NewStringLocated(e.name, e.getLineLocation()))
	if err != nil {
		return err
	}
	if realName == nil {
		throwRuntime("NullPointerException", "")
	}
	e.realName = *realName
	return nil
}

func (e *EaterTheme) getTheme(context *TContext) (*theme, error) {
	t, err := context.loadTheme(e.pathSystem, e.realName, e.from, e.getStringLocated())
	if err != nil {
		if _, ok := errors.AsType[*EaterException](err); ok {
			return nil, err
		}
		return nil, newEaterException("Cannot load theme "+e.realName, e.getStringLocated())
	}
	if t == nil {
		location := ""
		if e.from != nil {
			location = " in " + *e.from
		}
		return nil, newEaterException("Cannot load theme "+e.realName+location, e.getStringLocated())
	}
	return t, nil
}

func (e *EaterTheme) getName() string { return e.name }

func (e *EaterTheme) getNewImportedFiles() *PathSystem { return e.pathSystem }
