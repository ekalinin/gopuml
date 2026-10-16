// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/tim/TContext.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package tim

import (
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/json"
	"github.com/ekalinin/plantuml-in-go/internal/text"
)

// maxCallDepth bounds the nesting of function and procedure calls. Upstream
// has no limit and dies with StackOverflowError on infinite recursion; a
// Go stack overflow cannot be recovered, so the recursion is stopped first,
// beyond the depth that upstream reaches (about 600 with the default stack).
const maxCallDepth = 1000

// TContext executes the lines of one block: it holds the result lines, the
// functions, the subs and the current include directory.
type TContext struct {
	resultList []text.StringLocated
	debug      []text.StringLocated

	functionsSet *FunctionsSet

	subs                 map[string]*Sub
	definitionsContainer DefinitionsContainer

	// filesUsedCurrent is never filled, as in upstream v1.2026.8: the
	// "already included" checks of !include and !include_once never match.
	filesUsedCurrent map[string]bool

	preprocessingArtifact *PreprocessingArtifact
	pathSystem            *PathSystem

	pendingAdd    *string
	themeMetadata *json.Object
	defines       *Defines

	// Go additions: the file access of the render, the bundled stdlib and
	// themes, the directory of the source (for !import), a cache of line
	// types, the logs and the included files, and the settings of the
	// builtins (see Options).
	fa         *fileAccess
	stdlib     *Stdlib
	themes     ThemeSource
	currentDir string
	lineTypes  map[string]TLineType
	userLog    []string
	info       []string
	included   []string
	depth      int
	now        func() time.Time
	location   *time.Location
	getenv     func(name string) (string, bool)
	randomSeed *int64
}

func newTContext(pathSystem *PathSystem, defines *Defines, definitionsContainer DefinitionsContainer) *TContext {
	c := &TContext{
		functionsSet:          newFunctionsSet(),
		subs:                  map[string]*Sub{},
		definitionsContainer:  definitionsContainer,
		filesUsedCurrent:      map[string]bool{},
		preprocessingArtifact: &PreprocessingArtifact{},
		pathSystem:            pathSystem,
		themeMetadata:         json.NewObject(),
		defines:               defines,
		fa:                    pathSystem.fa,
		stdlib:                pathSystem.stdlib,
		lineTypes:             map[string]TLineType{},
	}
	c.addStandardFunctions(defines)
	return c
}

// lineType is StringLocated.getType, cached by line text.
func (c *TContext) lineType(s text.StringLocated) TLineType {
	if t, ok := c.lineTypes[s.S]; ok {
		return t
	}
	t := getTLineType(s.S)
	c.lineTypes[s.S] = t
	return t
}

func (c *TContext) logInfo(s string) { c.info = append(c.info, s) }

func (c *TContext) addIncluded(name string) {
	if !slices.Contains(c.included, name) {
		c.included = append(c.included, name)
	}
}

// enterCall counts a function or procedure call; the returned function
// must be called when the call ends.
func (c *TContext) enterCall(location text.StringLocated) (func(), error) {
	if c.depth >= maxCallDepth {
		return nil, newEaterException("Too many nested calls (infinite recursion?)", location)
	}
	c.depth++
	return func() { c.depth-- }, nil
}

// knowledge is TContext.asKnowledge.
type knowledge struct {
	context  *TContext
	memory   TMemory
	location *text.LineLocation
}

func (k knowledge) GetVariable(name string) (*TValue, error) {
	if strings.Contains(name, ".") || strings.Contains(name, "[") {
		return k.context.fromJSON(k.memory, name, k.location)
	}
	return k.memory.GetVariable(name), nil
}

func (k knowledge) GetFunction(name TFunctionSignature) TFunction {
	return k.context.functionsSet.getFunctionSmart(name)
}

func (c *TContext) asKnowledge(memory TMemory, location *text.LineLocation) Knowledge {
	return knowledge{context: c, memory: memory, location: location}
}

func (c *TContext) fromJSON(memory TMemory, name string, location *text.LineLocation) (*TValue, error) {
	result, err := c.applyFunctionsAndVariables(memory, text.NewStringLocated(name, location))
	if err != nil {
		return nil, err
	}
	if result == nil {
		throwRuntime("NullPointerException", "")
	}
	j, err := json.Parse(*result)
	if err != nil {
		return TValueFromString(*result), nil
	}
	return TValueFromJSON(j), nil
}

func (c *TContext) buildCodeIterator(memory TMemory, body []text.StringLocated) codeIterator {
	logs := &c.debug
	var it codeIterator = newCodeIteratorImpl(body)
	it = &codeIteratorLongComment{abstractCodeIterator{it}, c, logs}
	it = &codeIteratorShortComment{abstractCodeIterator{it}, c, logs}
	it = &codeIteratorInnerComment{abstractCodeIterator{it}}
	it = &codeIteratorSub{abstractCodeIterator: abstractCodeIterator{it}, subs: c.subs, context: c, memory: memory}
	it = &codeIteratorReturnFunction{abstractCodeIterator{it}, c.functionsSet, c, memory, logs}
	it = &codeIteratorProcedure{abstractCodeIterator{it}, c.functionsSet, c, memory, logs}
	it = &codeIteratorIf{abstractCodeIterator{it}, c, memory, logs}
	it = &codeIteratorLegacyDefine{abstractCodeIterator{it}, c.functionsSet, c, memory, logs}
	it = &codeIteratorWhile{abstractCodeIterator{it}, c, memory, logs}
	it = &codeIteratorForeach{abstractCodeIterator{it}, c, memory, logs}
	it = &codeIteratorAffectation{abstractCodeIterator{it}, c, memory, logs}
	return it
}

// executeLines runs the lines; in a function body (modeSpecial) it returns
// the value of the first !return executed.
func (c *TContext) executeLines(memory TMemory, body []text.StringLocated, ftype TFunctionType, modeSpecial bool) (*TValue, error) {
	it := c.buildCodeIterator(memory, body)
	for {
		s, err := it.peek()
		if err != nil || s == nil {
			return nil, err
		}
		result, err := c.executeOneLineSafe(memory, *s, ftype, modeSpecial)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		if err := it.next(); err != nil {
			return nil, err
		}
	}
}

func (c *TContext) executeLinesInternal(memory TMemory, body []text.StringLocated, ftype TFunctionType) error {
	it := c.buildCodeIterator(memory, body)
	for {
		s, err := it.peek()
		if err != nil || s == nil {
			return err
		}
		if _, err := c.executeOneLineSafe(memory, *s, ftype, false); err != nil {
			return err
		}
		if err := it.next(); err != nil {
			return err
		}
	}
}

// executeOneLineSafe runs one line; a runtime exception becomes the error
// "Fatal parsing error" at the line.
func (c *TContext) executeOneLineSafe(memory TMemory, s text.StringLocated, ftype TFunctionType, modeSpecial bool) (result *TValue, err error) {
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, newEaterException("Fatal parsing error", s)
		}
	}()
	c.debug = append(c.debug, s)
	return c.executeOneLineNotSafe(memory, s, ftype, modeSpecial)
}

var onlyWhitespaceNonEmpty = regexp.MustCompile(`^[\t\n\x0B\f\r ]+$`)

func (c *TContext) executeOneLineNotSafe(memory TMemory, s text.StringLocated, ftype TFunctionType, modeSpecial bool) (*TValue, error) {
	typ := c.lineType(s)
	switch {
	case typ == lineIncludesub:
		return nil, c.executeIncludesub(memory, s)
	case typ == lineTheme:
		return nil, c.executeTheme(memory, s)
	case typ == lineInclude:
		return nil, c.executeInclude(memory, s)
	case typ == lineIncludeDef:
		return nil, c.executeIncludeDef(memory, s)
	case typ == lineImport:
		return nil, c.executeImport(memory, s)
	case typ == lineDumpMemory:
		return nil, newEaterDumpMemory(s.Trimmed()).analyze(c, memory)
	case typ == lineAssert:
		return nil, newEaterAssert(s.Trimmed()).analyze(c, memory)
	case typ == lineOption:
		return nil, newEaterOption(s.Trimmed()).analyze(c, memory)
	case typ == lineUndef:
		return nil, newEaterUndef(s).analyze(c, memory)
	case ftype != functionReturn && typ == linePlain:
		return nil, c.addPlain(memory, s)
	case ftype == functionReturn && typ == lineReturn:
		if modeSpecial {
			eaterReturn := newEaterReturn(s)
			if err := eaterReturn.analyze(c, memory); err != nil {
				return nil, err
			}
			return eaterReturn.getValue2(), nil
		}
		// Actually, ignore because we are in a if
		return nil, nil
	case ftype == functionReturn && typ == linePlain:
		_, err := c.applyFunctionsAndVariablesInternal(memory, s)
		return nil, err
	case typ == lineAffectationDefine:
		return nil, newEaterAffectationDefine(s).analyze(c, memory)
	case ftype == functionNone && typ == lineEndFunction:
		// CommandExecutionResult.error("error endfunc"), ignored
		return nil, nil
	case typ == lineLog:
		return nil, newEaterLog(s.Trimmed()).analyze(c, memory)
	case onlyWhitespaceNonEmpty.MatchString(s.S):
		return nil, nil
	}
	return nil, newEaterException("Compile Error "+ftype.String()+" "+typ.String(), s)
}

func (c *TContext) addPlain(memory TMemory, s text.StringLocated) error {
	tmp, err := c.applyFunctionsAndVariablesInternal(memory, s)
	if err != nil {
		return err
	}
	if tmp != nil {
		if c.pendingAdd != nil {
			tmp[0] = text.NewStringLocated(*c.pendingAdd+tmp[0].S, tmp[0].Location)
			c.pendingAdd = nil
		}
		c.resultList = append(c.resultList, tmp...)
	}
	return nil
}

func (c *TContext) applyFunctionsAndVariablesInternal(memory TMemory, located text.StringLocated) ([]text.StringLocated, error) {
	if memory.IsEmpty() && c.functionsSet.size() == 0 {
		return []text.StringLocated{located}, nil
	}
	result, err := c.applyFunctionsAndVariables(memory, located)
	if err != nil || result == nil {
		return nil, err
	}
	// Pragma.legacyReplaceBackslashNByNewline() is always true upstream.
	splited := javaSplitNewline(*result)
	tab := make([]text.StringLocated, len(splited))
	for i, part := range splited {
		tab[i] = text.NewStringLocated(part, located.Location)
	}
	return tab, nil
}

// applyFunctionsAndVariables replaces the function calls and the variables
// of a line. It returns nil (Java's null) when the line calls a procedure,
// whose output has then been added to the result lines.
func (c *TContext) applyFunctionsAndVariables(memory TMemory, str text.StringLocated) (*string, error) {
	if memory.IsEmpty() && c.functionsSet.size() == 0 {
		s := str.S
		return &s, nil
	}
	s := toUTF16(str.S)
	vm := variableManager{context: c, memory: memory, location: str}
	var result []uint16
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if presentFunction := c.getFunctionNameAt(s, i); presentFunction != nil {
			name := fromUTF16(presentFunction)
			sub := fromUTF16(s[i:])
			call := newEaterFunctionCall(text.NewStringLocated(sub, str.Location), c.isLegacyDefine(name), c.isUnquoted(name))
			if err := call.analyze(c, memory); err != nil {
				return nil, err
			}
			signature := TFunctionSignature{FunctionName: name, NbArg: len(call.getValues()), NamedArguments: call.namedArgumentNames()}
			function := c.functionsSet.getFunctionSmart(signature)
			if function == nil {
				return nil, newEaterException("Function not found "+name, str)
			}
			switch function.FunctionType() {
			case functionProcedure:
				pending := fromUTF16(result)
				c.pendingAdd = &pending
				if err := c.executeVoid3(str, memory, function, call); err != nil {
					return nil, err
				}
				i += call.getCurrentPosition()
				if remaining := s[i:]; len(remaining) > 0 {
					c.appendToLastResult(fromUTF16(remaining))
				}
				return nil, nil
			case functionLegacyDefineLong:
				pending := fromUTF16(s[:i])
				c.pendingAdd = &pending
				if err := c.executeVoid3(str, memory, function, call); err != nil {
					return nil, err
				}
				return nil, nil
			}
			functionReturn, err := function.ExecuteReturnFunction(c, memory, str, call.getValues(), call.getNamedArguments())
			if err != nil {
				return nil, err
			}
			result = append(result, toUTF16(functionReturn.String())...)
			i += call.getCurrentPosition() - 1
		} else if vm.getVarnameAt(s, i) != nil {
			next, err := vm.replaceVariables(s, i, &result)
			if err != nil {
				return nil, err
			}
			i = next
		} else {
			result = append(result, ch)
		}
	}
	out := fromUTF16(result)
	return &out, nil
}

func (c *TContext) appendToLastResult(remaining string) {
	last := c.resultList[len(c.resultList)-1]
	last.S += remaining
	c.resultList[len(c.resultList)-1] = last
}

func (c *TContext) executeVoid3(location text.StringLocated, memory TMemory, function TFunction, call *EaterFunctionCall) error {
	return function.ExecuteProcedureInternal(c, memory, location, call.getValues(), call.getNamedArguments())
}

func (c *TContext) executeImport(memory TMemory, s text.StringLocated) error {
	imp := newEaterImport(s.Trimmed())
	if err := imp.analyze(c, memory); err != nil {
		return err
	}
	what, err := c.applyFunctionsAndVariables(memory, text.NewStringLocated(imp.getWhat(), s.Location))
	if err != nil {
		return err
	}
	if what == nil {
		throwRuntime("NullPointerException", "")
	}
	// The file is only checked: PathSystem.addImportFile does nothing in
	// upstream v1.2026.8, so the files of an imported archive cannot be
	// included.
	if file := c.fa.fileSystemGetFile(c.currentDir, *what); c.fa.isFile(file) {
		return nil
	}
	return newEaterException("Cannot import", s)
}

func (c *TContext) executeIncludesub(memory TMemory, s text.StringLocated) (err error) {
	var saveImportedFiles *PathSystem
	defer func() {
		if saveImportedFiles != nil {
			c.pathSystem = saveImportedFiles
		}
	}()
	include := newEaterIncludesub(s.Trimmed())
	if err := include.analyze(c, memory); err != nil {
		return err
	}
	what := include.getWhat()
	var sub *Sub
	if filename, blocname, ok := strings.Cut(what, "!"); ok {
		f2, err := c.pathSystem.getFile(filename)
		if err != nil {
			return newEaterException("cannot include "+what, s)
		}
		if f2 != nil {
			saveImportedFiles = c.pathSystem
			parent, err := f2.getParentFolder()
			if err != nil {
				return newEaterException("cannot include "+what, s)
			}
			c.pathSystem = c.pathSystem.withCurrentDir(parent)
			data, err := f2.newInputStream()
			if err != nil {
				return newEaterException("cannot include "+what, s)
			}
			if data == nil {
				throwRuntime("NullPointerException", "")
			}
			c.addIncluded(f2.String())
			readerline := core.NewMergeLinesReader(core.NewUncommentReader(core.NewLineReader(string(data), what, s.Location)))
			if sub, err = subFromFile(readerline, blocname, c, memory); err != nil {
				return err
			}
		}
	}
	if sub == nil {
		sub = c.subs[what]
	}
	if sub == nil {
		return newEaterException("cannot include "+what, s)
	}
	return c.executeLinesInternal(memory, sub.lines(), functionNone)
}

func (c *TContext) executeIncludeDef(memory TMemory, s text.StringLocated) error {
	include := newEaterIncludeDef(s.Trimmed())
	if err := include.analyze(c, memory); err != nil {
		return err
	}
	var definition []string
	if c.definitionsContainer != nil {
		definition = c.definitionsContainer.GetDefinition(include.getLocation())
	}
	body := readAll(&readLineList{lines: definition, location: s.Location})
	return c.executeLinesInternal(memory, body, functionNone)
}

// ThemeMetadata returns the YAML header of the last theme loaded by !theme.
func (c *TContext) ThemeMetadata() *json.Object { return c.themeMetadata }

func (c *TContext) executeTheme(memory TMemory, s text.StringLocated) error {
	eater := newEaterTheme(s.Trimmed(), c.pathSystem)
	if err := eater.analyze(c, memory); err != nil {
		return err
	}
	theme, err := eater.getTheme(c)
	if err != nil {
		return err
	}
	if theme == nil {
		return newEaterException("No such theme "+eater.getName(), s)
	}
	saveImportedFiles := c.pathSystem
	c.pathSystem = eater.getNewImportedFiles()
	defer func() {
		c.themeMetadata = theme.getMetadata()
		c.pathSystem = saveImportedFiles
	}()
	var body []text.StringLocated
	for {
		sl, ok := theme.readLine()
		if !ok {
			break
		}
		body = append(body, sl)
	}
	_, err = c.executeLines(memory, body, functionNone, false)
	return err
}

func (c *TContext) executeInclude(memory TMemory, s text.StringLocated) error {
	include := newEaterInclude(s.Trimmed())
	if err := include.analyze(c, memory); err != nil {
		return err
	}
	what := include.getWhat()
	strategy := include.getPreprocessorIncludeStrategy()
	var suf *string
	if idx := strings.LastIndexByte(what, '!'); idx != -1 {
		tmp := what[idx+1:]
		suf = &tmp
		what = what[:idx]
	}

	var reader core.ReadLine
	var saveImportedFiles *PathSystem
	defer func() {
		if saveImportedFiles != nil {
			c.pathSystem = saveImportedFiles
		}
	}()
	switch {
	case strings.HasPrefix(what, "<") && strings.HasSuffix(what, ">") && len(what) >= 2:
		stdlibPath := what[1 : len(what)-1]
		if slash := strings.IndexByte(stdlibPath, '/'); slash >= 0 && !c.stdlib.hasLibrary(strings.ToLower(stdlibPath[:slash])) {
			return newEaterException(notBundledMessage(what, stdlibPath[:slash]), s)
		}
		saveImportedFiles = c.pathSystem
		tmp, err := c.pathSystem.getInputFile(what)
		if err != nil {
			return newEaterException("cannot include "+err.Error(), s)
		}
		parent, err := tmp.getParentFolder()
		if err != nil {
			return newEaterException("cannot include "+err.Error(), s)
		}
		c.pathSystem = c.pathSystem.withCurrentDir(parent)
		reader = getReaderStdlibInclude(c.stdlib, stdlibPath)
	case strings.HasPrefix(what, "http://") || strings.HasPrefix(what, "https://"):
		url, ok := createURL(what)
		if !ok {
			return newEaterException("Cannot open URL", s)
		}
		var err error
		if reader, err = c.getReaderIncludeURL(url, s, suf); err != nil {
			return err
		}
		c.addIncluded(url)
	case strings.HasPrefix(what, "[") && strings.HasSuffix(what, "]"):
		return newEaterException("cannot include java.io.IOException: To be finished", s)
	default:
		f2, err := c.pathSystem.getInputFile(what)
		if err != nil {
			return newEaterException("cannot include "+err.Error(), s)
		}
		if f2 != nil {
			if strategy == includeDefault && c.filesUsedCurrent[f2.String()] {
				return nil
			}
			if strategy == includeOnce && c.filesUsedCurrent[f2.String()] {
				return newEaterException("This file has already been included", s)
			}
			data, err := f2.newInputStream()
			if err != nil {
				return newEaterException("cannot include "+err.Error(), s)
			}
			if data == nil {
				throwRuntime("NullPointerException", "")
			}
			reader = extractFromData(data, "desc2", nil)
			if reader == nil {
				reader = core.NewLineReader(string(data), what, s.Location)
			}
			saveImportedFiles = c.pathSystem
			parent, err := f2.getParentFolder()
			if err != nil {
				return newEaterException("cannot include "+err.Error(), s)
			}
			c.pathSystem = c.pathSystem.withCurrentDir(parent)
			c.addIncluded(f2.String())
		}
	}
	if reader != nil {
		body := readAll(newReadLineWithYamlHeader(reader))
		_, err := c.executeLines(memory, body, functionNone, false)
		return err
	}
	return newEaterException("cannot include "+what, s)
}

func (c *TContext) isLegacyDefine(functionName string) bool {
	for _, fn := range c.functionsSet.getFunctionsByName(functionName) {
		if fn.FunctionType().isLegacy() {
			return true
		}
	}
	return false
}

func (c *TContext) isUnquoted(functionName string) bool {
	for _, fn := range c.functionsSet.getFunctionsByName(functionName) {
		if fn.IsUnquoted() {
			return true
		}
	}
	return false
}

func (c *TContext) doesFunctionExist(functionName string) bool {
	return c.functionsSet.doesFunctionExist(functionName)
}

// getFunctionNameAt returns the name of the function called at s[pos:], or
// nil.
func (c *TContext) getFunctionNameAt(s []uint16, pos int) []uint16 {
	justAfterALetter := pos > 0 && isLetterOrEmojiOrUnderscoreOrDigit(s[pos-1]) && !justAfterBackslashN(s, pos)
	if justAfterALetter && s[pos] != '%' && s[pos] != '$' {
		return nil
	}
	n := c.functionsSet.getLonguestMatchStartingIn(s, pos)
	if n == 0 {
		return nil
	}
	return s[pos : pos+n-1]
}

// extractFromResultList removes the result lines from index n1 and returns
// them joined with the newline marker.
func (c *TContext) extractFromResultList(n1 int) string {
	var sb strings.Builder
	for len(c.resultList) > n1 {
		sb.WriteString(c.resultList[n1].S)
		c.resultList = slices.Delete(c.resultList, n1, n1+1)
		if len(c.resultList) > n1 {
			sb.WriteRune(jawsNewline)
		}
	}
	return sb.String()
}

func (c *TContext) getFunctionSmart(signature TFunctionSignature) TFunction {
	return c.functionsSet.getFunctionSmart(signature)
}

// getXargs retrieves the data given after @startuml. Like upstream, it
// reads the Java toString of the first line, "(SL) @startuml ...".
func (c *TContext) getXargs() (string, bool) {
	if len(c.resultList) == 0 {
		return "", false
	}
	first := "(SL) " + c.resultList[0].S
	if c.resultList[0].S == "" {
		first = "<<<EMPTY STRING>>>"
	}
	idx := strings.IndexByte(first, ' ')
	if idx == -1 {
		return "", false
	}
	return text.Trin(first[idx+1:]), true
}
