// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/TContext.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

// standardFunction is one builtin of TContext.addStandardFunctions.
type standardFunction struct {
	// signature is the SIGNATURE constant of the upstream class.
	signature TFunctionSignature
	// build creates the builtin.
	build func(defines *Defines) TFunction
}

func sig(name string, nbArg int) TFunctionSignature { return NewTFunctionSignature(name, nbArg) }

// standardFunctions lists the 74 builtins of upstream (JVM build), in the
// order of TContext.addStandardFunctions; the order decides the iteration
// order of the function table (see FunctionsSet).
var standardFunctions = []standardFunction{
	{sig("%false", 0), func(*Defines) TFunction { return alwaysFalse{} }},
	{sig("%true", 0), func(*Defines) TFunction { return alwaysTrue{} }},
	{sig("%backslash", 0), newCharFunction("%backslash", jawsRealBackslash)},
	{sig("%boolval", 1), func(*Defines) TFunction { return boolVal{} }},
	{sig("%breakline", 0), newCharFunction("%breakline", jawsBreakline)},
	{sig("%call_user_func", 1), func(*Defines) TFunction { return callUserFunction{} }},
	{sig("%chr", 1), func(*Defines) TFunction { return chr{} }},
	{sig("%darken", 2), func(*Defines) TFunction { return darken{} }},
	{sig("%date", 3), func(*Defines) TFunction { return dateFunction{} }},
	{sig("%dec2hex", 1), func(*Defines) TFunction { return dec2hex{} }},
	{sig("%dirpath", 0), newEnvironmentFunction("%dirpath", "dirpath")},
	{sig("%dollar", 0), newCharFunction("%dollar", '$')},
	{sig("%eval", 1), func(*Defines) TFunction { return eval{} }},
	{sig("%feature", 1), func(*Defines) TFunction { return feature{} }},
	{sig("%filedate", 0), newEnvironmentFunction("%filedate", "filedate")},
	{sig("%file_exists", 1), func(*Defines) TFunction { return fileExists{} }},
	{sig("%filename", 0), newEnvironmentFunction("%filename", "filename")},
	{sig("%filename_no_extension", 0), newEnvironmentFunction("%filename_no_extension", "filenameNoExtension")},
	{sig("%function_exists", 1), func(*Defines) TFunction { return functionExists{} }},
	{sig("%get_all_stdlib", 1), func(*Defines) TFunction { return getAllStdlib{} }},
	{sig("%get_all_theme", 0), func(*Defines) TFunction { return getAllTheme{} }},
	{sig("%get_current_theme", 0), func(*Defines) TFunction { return getCurrentTheme{} }},
	{sig("%get_json_keys", 1), func(*Defines) TFunction { return getJSONKey{} }},
	{sig("%get_json_type", 1), func(*Defines) TFunction { return getJSONType{} }},
	{sig("%get_stdlib", 1), func(*Defines) TFunction { return getStdlib{} }},
	{sig("%get_variable_value", 1), func(*Defines) TFunction { return getVariableValue{} }},
	{sig("%version", 0), func(*Defines) TFunction { return getVersion{} }},
	{sig("%getenv", 1), func(*Defines) TFunction { return getenv{} }},
	{sig("%hex2dec", 1), func(*Defines) TFunction { return hex2dec{} }},
	{sig("%hsl_color", 3), func(*Defines) TFunction { return hslColor{} }},
	{sig("%intval", 1), func(*Defines) TFunction { return intVal{} }},
	{sig("%invoke_procedure", 1), func(*Defines) TFunction { return invokeProcedure{} }},
	{sig("%is_dark", 1), func(*Defines) TFunction { return isDark{} }},
	{sig("%is_light", 1), func(*Defines) TFunction { return isLight{} }},
	{sig("%json_add", 3), func(*Defines) TFunction { return jsonAdd{} }},
	{sig("%json_key_exists", 1), func(*Defines) TFunction { return jsonKeyExists{} }},
	{sig("%json_merge", 2), func(*Defines) TFunction { return jsonMerge{} }},
	{sig("%json_remove", 2), func(*Defines) TFunction { return jsonRemove{} }},
	{sig("%json_set", 3), func(*Defines) TFunction { return jsonSet{} }},
	{sig("%left_align", 0), newCharFunction("%left_align", jawsNewlineLeftAlign)},
	{sig("%lighten", 2), func(*Defines) TFunction { return lighten{} }},
	{sig("%load_json", 3), func(*Defines) TFunction { return loadJSON{} }},
	{sig("%and", 2), newLogicalFunction("%and", logicalAnd)},
	{sig("%nand", 2), newLogicalFunction("%nand", logicalNand)},
	{sig("%nor", 2), newLogicalFunction("%nor", logicalNor)},
	{sig("%not", 1), func(*Defines) TFunction { return logicalNot{} }},
	{sig("%nxor", 2), newLogicalFunction("%nxor", logicalNxor)},
	{sig("%or", 2), newLogicalFunction("%or", logicalOr)},
	{sig("%xor", 2), newLogicalFunction("%xor", logicalXor)},
	{sig("%lower", 1), func(*Defines) TFunction { return lower{} }},
	{sig("%mod", 2), func(*Defines) TFunction { return modulo{} }},
	{sig("%newline", 0), newCharFunction("%newline", jawsNewline)},
	{sig("%n", 0), newCharFunction("%n", jawsNewline)},
	{sig("%now", 0), func(*Defines) TFunction { return now{} }},
	{sig("%ord", 1), func(*Defines) TFunction { return ord{} }},
	{sig("%percent", 0), newCharFunction("%percent", '%')},
	{sig("%random", 2), func(*Defines) TFunction { return &randomFunction{} }},
	{sig("%retrieve_procedure", 1), func(*Defines) TFunction { return retrieveProcedure{} }},
	{sig("%reverse_color", 1), func(*Defines) TFunction { return reverseColor{} }},
	{sig("%reverse_hsluv_color", 1), func(*Defines) TFunction { return reverseHsluvColor{} }},
	{sig("%right_align", 0), newCharFunction("%right_align", jawsNewlineRightAlign)},
	{sig("%set_variable_value", 2), func(*Defines) TFunction { return setVariableValue{} }},
	{sig("%size", 1), func(*Defines) TFunction { return size{} }},
	{sig("%splitstr", 3), func(*Defines) TFunction { return splitStr{} }},
	{sig("%splitstr_regex", 2), func(*Defines) TFunction { return splitStrRegex{} }},
	{sig("%str2json", 1), func(*Defines) TFunction { return str2JSON{} }},
	{sig("%string", 1), func(*Defines) TFunction { return stringFunction{} }},
	{sig("%strlen", 1), func(*Defines) TFunction { return strlen{} }},
	{sig("%strpos", 2), func(*Defines) TFunction { return strpos{} }},
	{sig("%substr", 3), func(*Defines) TFunction { return substr{} }},
	{sig("%tab", 0), newCharFunction("%tab", jawsRealTabulation)},
	{sig("%upper", 1), func(*Defines) TFunction { return upper{} }},
	{sig("%variable_exists", 1), func(*Defines) TFunction { return variableExists{} }},
	{sig("%xargs", 0), func(*Defines) TFunction { return xargs{} }},
}

// addStandardFunctions is TContext.addStandardFunctions.
func (c *TContext) addStandardFunctions(defines *Defines) {
	for _, f := range standardFunctions {
		c.functionsSet.addFunction(f.build(defines))
	}
}
