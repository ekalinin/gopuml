// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/builtin/Getenv.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/SecurityProfile.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"os"
	"runtime"
	"strings"

	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/text"
)

// getenv is %getenv(name): the Java system property or environment
// variable, or "" when it is not set or the security profile does not allow
// reading it.
type getenv struct{ simpleReturnFunction }

func (getenv) Signature() TFunctionSignature { return NewTFunctionSignature("%getenv", 1) }

func (getenv) CanCover(nbArg int, _ []string) bool { return nbArg == 1 }

func (getenv) ExecuteReturnFunction(context *TContext, _ TMemory, _ text.StringLocated, values []*TValue,
	_ map[string]*TValue) (*TValue, error) {
	name := values[0].String()
	profile := context.fa.access.Config.Profile
	if !canWeReadThisEnvironmentVariable(profile, name) {
		return TValueFromString(""), nil
	}
	if name == "" {
		// System.getProperty("") throws IllegalArgumentException
		throwRuntime("IllegalArgumentException", "key can't be empty")
	}
	if value, ok := context.lookupEnv(name); ok {
		return TValueFromString(value), nil
	}
	return TValueFromString(""), nil
}

// canWeReadThisEnvironmentVariable is
// SecurityProfile.canWeReadThisEnvironmentVariable: names starting with
// "plantuml" (except "plantuml.security") and the line and path separators
// are readable in every profile, anything in the UNSECURE profile.
func canWeReadThisEnvironmentVariable(profile security.Profile, name string) bool {
	lname := javaToLowerCase(name)
	if strings.HasPrefix(lname, "plantuml.security") {
		return false
	}
	if strings.HasPrefix(lname, "plantuml") {
		return true
	}
	if lname == "path.separator" || lname == "line.separator" {
		return true
	}
	return profile == security.Insecure
}

// lookupEnv looks a name up with Options.Getenv, or with defaultGetenv.
func (c *TContext) lookupEnv(name string) (string, bool) {
	if c.getenv != nil {
		return c.getenv(name)
	}
	return defaultGetenv(name)
}

// defaultGetenv answers the Java system properties line.separator,
// path.separator and file.separator, then reads the environment.
func defaultGetenv(name string) (string, bool) {
	switch name {
	case "line.separator":
		if runtime.GOOS == "windows" {
			return "\r\n", true
		}
		return "\n", true
	case "path.separator":
		return string(os.PathListSeparator), true
	case "file.separator":
		return string(os.PathSeparator), true
	}
	return os.LookupEnv(name)
}
