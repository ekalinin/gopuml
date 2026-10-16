// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/TContext.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

// standardFunction is one builtin of TContext.addStandardFunctions.
type standardFunction struct {
	// signature is the SIGNATURE constant of the upstream class.
	signature TFunctionSignature
	// build creates the builtin; nil for a builtin that is not ported yet.
	build func(defines *Defines) TFunction
}

func sig(name string, nbArg int) TFunctionSignature { return NewTFunctionSignature(name, nbArg) }

// standardFunctions lists the 74 builtins of upstream (JVM build), in the
// order of TContext.addStandardFunctions; the order decides the iteration
// order of the function table (see FunctionsSet). To register a builtin,
// set its build function, for example:
//
//	{sig("%chr", 1), func(*Defines) TFunction { return chr{} }},
//
// Not ported yet (build is nil): %backslash, %boolval, %breakline, %chr,
// %darken, %date, %dec2hex, %dirpath, %dollar, %feature, %filedate,
// %file_exists, %filename, %filename_no_extension, %get_all_stdlib,
// %get_all_theme, %get_current_theme, %get_json_keys, %get_json_type,
// %get_stdlib, %version, %getenv, %hex2dec, %hsl_color, %is_dark,
// %is_light, %json_add, %json_key_exists, %json_merge, %json_remove,
// %json_set, %left_align, %lighten, %load_json, %and, %nand, %nor, %nxor,
// %or, %xor, %lower, %mod, %newline, %n, %now, %ord, %percent, %random,
// %reverse_color, %reverse_hsluv_color, %right_align, %size, %splitstr,
// %splitstr_regex, %str2json, %strpos, %substr, %tab.
var standardFunctions = []standardFunction{
	{sig("%false", 0), func(*Defines) TFunction { return alwaysFalse{} }},
	{sig("%true", 0), func(*Defines) TFunction { return alwaysTrue{} }},
	{sig("%backslash", 0), nil},
	{sig("%boolval", 1), nil},
	{sig("%breakline", 0), nil},
	{sig("%call_user_func", 1), func(*Defines) TFunction { return callUserFunction{} }},
	{sig("%chr", 1), nil},
	{sig("%darken", 2), nil},
	{sig("%date", 3), nil},
	{sig("%dec2hex", 1), nil},
	{sig("%dirpath", 0), nil},
	{sig("%dollar", 0), nil},
	{sig("%eval", 1), func(*Defines) TFunction { return eval{} }},
	{sig("%feature", 1), nil},
	{sig("%filedate", 0), nil},
	{sig("%file_exists", 1), nil},
	{sig("%filename", 0), nil},
	{sig("%filename_no_extension", 0), nil},
	{sig("%function_exists", 1), func(*Defines) TFunction { return functionExists{} }},
	{sig("%get_all_stdlib", 1), nil},
	{sig("%get_all_theme", 0), nil},
	{sig("%get_current_theme", 0), nil},
	{sig("%get_json_keys", 1), nil},
	{sig("%get_json_type", 1), nil},
	{sig("%get_stdlib", 1), nil},
	{sig("%get_variable_value", 1), func(*Defines) TFunction { return getVariableValue{} }},
	{sig("%version", 0), nil},
	{sig("%getenv", 1), nil},
	{sig("%hex2dec", 1), nil},
	{sig("%hsl_color", 3), nil},
	{sig("%intval", 1), func(*Defines) TFunction { return intVal{} }},
	{sig("%invoke_procedure", 1), func(*Defines) TFunction { return invokeProcedure{} }},
	{sig("%is_dark", 1), nil},
	{sig("%is_light", 1), nil},
	{sig("%json_add", 3), nil},
	{sig("%json_key_exists", 1), nil},
	{sig("%json_merge", 2), nil},
	{sig("%json_remove", 2), nil},
	{sig("%json_set", 3), nil},
	{sig("%left_align", 0), nil},
	{sig("%lighten", 2), nil},
	{sig("%load_json", 3), nil},
	{sig("%and", 2), nil},
	{sig("%nand", 2), nil},
	{sig("%nor", 2), nil},
	{sig("%not", 1), func(*Defines) TFunction { return logicalNot{} }},
	{sig("%nxor", 2), nil},
	{sig("%or", 2), nil},
	{sig("%xor", 2), nil},
	{sig("%lower", 1), nil},
	{sig("%mod", 2), nil},
	{sig("%newline", 0), nil},
	{sig("%n", 0), nil},
	{sig("%now", 0), nil},
	{sig("%ord", 1), nil},
	{sig("%percent", 0), nil},
	{sig("%random", 2), nil},
	{sig("%retrieve_procedure", 1), func(*Defines) TFunction { return retrieveProcedure{} }},
	{sig("%reverse_color", 1), nil},
	{sig("%reverse_hsluv_color", 1), nil},
	{sig("%right_align", 0), nil},
	{sig("%set_variable_value", 2), func(*Defines) TFunction { return setVariableValue{} }},
	{sig("%size", 1), nil},
	{sig("%splitstr", 3), nil},
	{sig("%splitstr_regex", 2), nil},
	{sig("%str2json", 1), nil},
	{sig("%string", 1), func(*Defines) TFunction { return stringFunction{} }},
	{sig("%strlen", 1), func(*Defines) TFunction { return strlen{} }},
	{sig("%strpos", 2), nil},
	{sig("%substr", 3), nil},
	{sig("%tab", 0), nil},
	{sig("%upper", 1), func(*Defines) TFunction { return upper{} }},
	{sig("%variable_exists", 1), func(*Defines) TFunction { return variableExists{} }},
	{sig("%xargs", 0), func(*Defines) TFunction { return xargs{} }},
}

// addStandardFunctions is TContext.addStandardFunctions.
func (c *TContext) addStandardFunctions(defines *Defines) {
	for _, f := range standardFunctions {
		if f.build == nil {
			c.functionsSet.addPhantom(f.signature)
			continue
		}
		c.functionsSet.addFunction(f.build(defines))
	}
}
