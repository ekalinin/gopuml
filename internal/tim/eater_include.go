// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterInclude.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterIncludeDef.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterIncludesub.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/EaterImport.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import "github.com/ekalinin/gopuml/internal/text"

// EaterInclude is "!include", "!includeurl", "!include_many" or
// "!include_once" followed by what to include.
type EaterInclude struct {
	Eater
	what     string
	strategy PreprocessorIncludeStrategy
}

func newEaterInclude(sl text.StringLocated) *EaterInclude {
	return &EaterInclude{Eater: newEater(sl), strategy: includeDefault}
}

func (e *EaterInclude) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!include"); err != nil {
		return err
	}
	switch e.peekChar() {
	case 'u':
		if err := e.checkAndEatString("url"); err != nil {
			return err
		}
	case '_':
		if err := e.checkAndEatString("_"); err != nil {
			return err
		}
		if e.peekChar() == 'm' {
			if err := e.checkAndEatString("many"); err != nil {
				return err
			}
			e.strategy = includeMany
		} else {
			if err := e.checkAndEatString("once"); err != nil {
				return err
			}
			e.strategy = includeOnce
		}
	}
	e.skipSpaces()
	var err error
	e.what, err = applyToRest(&e.Eater, context, memory)
	return err
}

// applyToRest applies variables and functions to the rest of the line.
// A procedure call gives Java's null; the callers then fail with
// NullPointerException.
func applyToRest(e *Eater, context *TContext, memory TMemory) (string, error) {
	what, err := context.applyFunctionsAndVariables(memory, text.NewStringLocated(e.eatAllToEnd(), e.getLineLocation()))
	if err != nil {
		return "", err
	}
	if what == nil {
		throwRuntime("NullPointerException", "")
	}
	return *what, nil
}

func (e *EaterInclude) getWhat() string { return e.what }

func (e *EaterInclude) getPreprocessorIncludeStrategy() PreprocessorIncludeStrategy {
	return e.strategy
}

// EaterIncludeDef is "!includedef name".
type EaterIncludeDef struct {
	Eater
	location string
}

func newEaterIncludeDef(sl text.StringLocated) *EaterIncludeDef {
	return &EaterIncludeDef{Eater: newEater(sl)}
}

func (e *EaterIncludeDef) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!includedef"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.location, err = applyToRest(&e.Eater, context, memory)
	return err
}

func (e *EaterIncludeDef) getLocation() string { return e.location }

// EaterIncludesub is "!includesub [file!]name".
type EaterIncludesub struct {
	Eater
	what string
}

func newEaterIncludesub(sl text.StringLocated) *EaterIncludesub {
	return &EaterIncludesub{Eater: newEater(sl)}
}

func (e *EaterIncludesub) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!includesub"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.what, err = applyToRest(&e.Eater, context, memory)
	return err
}

func (e *EaterIncludesub) getWhat() string { return e.what }

// EaterImport is "!import file".
type EaterImport struct {
	Eater
	what string
}

func newEaterImport(sl text.StringLocated) *EaterImport { return &EaterImport{Eater: newEater(sl)} }

func (e *EaterImport) analyze(context *TContext, memory TMemory) error {
	e.skipSpaces()
	if err := e.checkAndEatString("!import"); err != nil {
		return err
	}
	e.skipSpaces()
	var err error
	e.what, err = applyToRest(&e.Eater, context, memory)
	return err
}

func (e *EaterImport) getWhat() string { return e.what }
