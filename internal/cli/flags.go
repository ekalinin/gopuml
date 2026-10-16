// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliFlag.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/Arity.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

// Package cli is upstream's command line interface: flags, parsing and
// help, compatible with java -jar plantuml.jar.
package cli

import "strings"

// Arity is how a flag takes its value.
type Arity int

// Arities.
const (
	UnaryBoolean             Arity = iota // --svg
	UnaryImmediateAction                  // --version: runs and exits
	UnaryOptionalColon                    // --http-server[:port]
	UnaryInlineKeyOrKeyValue              // -DKEY=VALUE or -D KEY=VALUE
	BinaryNextArgumentValue               // --output-dir DIR
)

// Flag identifies an upstream command line flag. Flags are matched in
// declaration order, which matters: "-SVG" is the -S (skinparam) flag.
type Flag int

// FlagInfo describes a flag as upstream declares it.
type FlagInfo struct {
	Name    string
	Flag    string
	Aliases []string
	Arity   Arity
	// Format is the output format a format flag selects, upstream's name.
	Format string
	// Doc, Usage, Level and NewGroup feed the help text. A documented flag
	// without an explicit level has level 999, as upstream's annotation.
	Doc      string
	Usage    string
	Level    int
	NewGroup string
	// Default is the value when the flag is absent.
	Default string
}

// Info returns the description of a flag.
func (f Flag) Info() *FlagInfo { return &flags[f] }

func (f Flag) String() string { return flags[f].Flag }

// match reports whether an argument selects the flag.
func (f Flag) match(arg string) bool {
	info := &flags[f]
	switch info.Arity {
	case UnaryInlineKeyOrKeyValue:
		for _, a := range info.Aliases {
			if strings.HasPrefix(arg, a) {
				return true
			}
		}
		return strings.HasPrefix(arg, info.Flag)
	case UnaryOptionalColon:
		for _, a := range info.Aliases {
			if hasPrefixFold(arg, a) {
				return true
			}
		}
		return hasPrefixFold(arg, info.Flag)
	}
	for _, a := range info.Aliases {
		if strings.EqualFold(arg, a) {
			return true
		}
	}
	return strings.EqualFold(arg, info.Flag)
}

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

const (
	Help Flag = iota
	HelpMore
	Version
	Author
	Gui
	DarkMode
	Verbose
	Duration
	Progress
	Splash
	TestDot
	Picoweb
	Exclude
	Pipe
	Pipemap
	Pipedelimitor
	Pipenostderr
	PipeImageIndex
	Define
	DefineLong
	Include
	IncludeLong
	Pragma
	PragmaLong
	Skinparam
	SkinparamLong
	Theme
	Config
	Charset
	CheckOnly
	FailFast
	FailFast2
	NoErrorImage
	Timeout
	IgnoreStartumlFilename
	NbThread
	Loop
	RetrieveMetadata
	NoMetadata
	CheckMetadata
	EncodeSprite
	ComputeUrl
	DecodeUrl
	Language
	GraphvizDot
	Ftp
	OutputDir
	Overwrite
	Clipboard
	Clipboardloop
	DebugSvek
	FileDir
	Filename
	Headless
	PrintFonts
	StdLib
	Stdrpt
	Syntax
	License
	Word
	UseSeparatorMinus
	Format
	TEps
	TEpsText
	THtml
	TLatex
	TLatexNopreamble
	Obfuscate
	TPdf
	TPng
	Preprocess
	TScxml
	TSvg
	TNull
	TTxt
	TUtxt
	TVdx
	TXmi
	TXmiArgo
	TXmiCustom
	TXmiScript
	TXmiStar
	TBase64
	TBraille
	DisableStats
	EnableStats
	Dumphtmlstats
	Dumpstats
	HtmlStats
	XmlStats
	RealtimeStats
	LoopStats
	flagCount
)

var flags = [flagCount]FlagInfo{
	Help:                   {Name: "HELP", Flag: "--help", Arity: UnaryImmediateAction, Aliases: []string{"-h", "-?", "-help"}, Doc: "Show help and usage information", Level: 0},
	HelpMore:               {Name: "HELP_MORE", Flag: "--help-more", Arity: UnaryImmediateAction, Aliases: []string{"-h:more", "-help:more"}, Doc: "Show extended help (advanced options)", Level: 0},
	Version:                {Name: "VERSION", Flag: "--version", Arity: UnaryImmediateAction, Aliases: []string{"-version"}, Doc: "Show PlantUML and Java version", Level: 0},
	Author:                 {Name: "AUTHOR", Flag: "--author", Arity: UnaryImmediateAction, Aliases: []string{"--about", "-authors", "-about"}, Doc: "Show information about PlantUML authors", Level: 0},
	Gui:                    {Name: "GUI", Flag: "--gui", Arity: UnaryBoolean, Aliases: []string{"-gui"}, Doc: "Launch the graphical user interface", Level: 0},
	DarkMode:               {Name: "DARK_MODE", Flag: "--dark-mode", Arity: UnaryBoolean, Aliases: []string{"-darkmode"}, Doc: "Render diagrams in dark mode", Level: 0},
	Verbose:                {Name: "VERBOSE", Flag: "--verbose", Arity: UnaryBoolean, Aliases: []string{"-v", "-verbose"}, Doc: "Enable verbose logging", Level: 0},
	Duration:               {Name: "DURATION", Flag: "--duration", Arity: UnaryBoolean, Aliases: []string{"-duration"}, Doc: "Print total processing time", Level: 0},
	Progress:               {Name: "PROGRESS", Flag: "--progress-bar", Arity: UnaryBoolean, Aliases: []string{"-progress"}, Doc: "Show a textual progress bar", Level: 0},
	Splash:                 {Name: "SPLASH", Flag: "--splash-screen", Arity: UnaryBoolean, Aliases: []string{"-splash"}, Doc: "Show splash screen with progress bar", Level: 0},
	TestDot:                {Name: "TEST_DOT", Flag: "--check-graphviz", Arity: UnaryImmediateAction, Aliases: []string{"-testdot"}, Doc: "Check Graphviz installation", Level: 0},
	Picoweb:                {Name: "PICOWEB", Flag: "--http-server", Arity: UnaryOptionalColon, Aliases: []string{"-picoweb"}, Doc: "Start internal HTTP server for rendering (default port : 8080)", Usage: "--http-server[:<port>]", Level: 0},
	Exclude:                {Name: "EXCLUDE", Flag: "--exclude", Arity: BinaryNextArgumentValue, Aliases: []string{"-x", "-exclude"}, Doc: "Exclude input files matching the given pattern", Usage: "--exclude <pattern>", Level: 1, NewGroup: "Input & preprocessing"},
	Pipe:                   {Name: "PIPE", Flag: "--pipe", Arity: UnaryBoolean, Aliases: []string{"-p", "-pipe"}, Doc: "Read source from stdin, write result to stdout", Level: 0},
	Pipemap:                {Name: "PIPEMAP", Flag: "-pipemap", Arity: UnaryBoolean, Level: -1},
	Pipedelimitor:          {Name: "PIPEDELIMITOR", Flag: "-pipedelimitor", Arity: BinaryNextArgumentValue, Level: -1},
	Pipenostderr:           {Name: "PIPENOSTDERR", Flag: "-pipenostderr", Arity: UnaryBoolean, Level: -1},
	PipeImageIndex:         {Name: "PIPE_IMAGE_INDEX", Flag: "--pipe-image-index", Arity: BinaryNextArgumentValue, Aliases: []string{"-pipeimageindex"}, Doc: "Generate the Nth image with pipe option", Level: 999, Default: "0"},
	Define:                 {Name: "DEFINE", Flag: "-D", Arity: UnaryInlineKeyOrKeyValue, Doc: "Define a preprocessing variable (equivalent to '!define <var> <value>')", Usage: "-d, --define <VAR>=<value>", Level: 0},
	DefineLong:             {Name: "DEFINE_LONG", Flag: "--define", Arity: BinaryNextArgumentValue, Level: -1},
	Include:                {Name: "INCLUDE", Flag: "-I", Arity: UnaryInlineKeyOrKeyValue, Doc: "Include external file (as with '!include <file>')", Usage: "-I, --include <file>", Level: 1},
	IncludeLong:            {Name: "INCLUDE_LONG", Flag: "--include", Arity: BinaryNextArgumentValue, Level: -1},
	Pragma:                 {Name: "PRAGMA", Flag: "-P", Arity: UnaryInlineKeyOrKeyValue, Doc: "Set pragma (equivalent to '!pragma <key> <value>')", Usage: "-P, --pragma <key>=<value>", Level: 1},
	PragmaLong:             {Name: "PRAGMA_LONG", Flag: "--pragma", Arity: BinaryNextArgumentValue, Level: -1},
	Skinparam:              {Name: "SKINPARAM", Flag: "-S", Arity: UnaryInlineKeyOrKeyValue, Doc: "Set skin parameter (equivalent to 'skinparam <key> <value>')", Usage: "--skinparam <key>=<value>", Level: 1},
	SkinparamLong:          {Name: "SKINPARAM_LONG", Flag: "--skinparam", Arity: BinaryNextArgumentValue, Level: -1},
	Theme:                  {Name: "THEME", Flag: "--theme", Arity: BinaryNextArgumentValue, Aliases: []string{"-theme"}, Doc: "Apply a theme", Usage: "--theme <name>", Level: 1},
	Config:                 {Name: "CONFIG", Flag: "--config", Arity: BinaryNextArgumentValue, Aliases: []string{"-config"}, Doc: "Specify configuration file", Usage: "--config <file>", Level: 1},
	Charset:                {Name: "CHARSET", Flag: "--charset", Arity: BinaryNextArgumentValue, Aliases: []string{"-charset"}, Doc: "Use a specific input charset", Usage: "--charset <name>", Level: 1, Default: "UTF-8"},
	CheckOnly:              {Name: "CHECK_ONLY", Flag: "--check-syntax", Arity: UnaryBoolean, Aliases: []string{"--syntax-check", "-checkonly"}, Doc: "Check diagram syntax without generating images", Level: 0, NewGroup: "Execution control"},
	FailFast:               {Name: "FAIL_FAST", Flag: "--stop-on-error", Arity: UnaryBoolean, Aliases: []string{"-failfast"}, Doc: "Stop at the first syntax error", Level: 0},
	FailFast2:              {Name: "FAIL_FAST2", Flag: "--check-before-run", Arity: UnaryBoolean, Aliases: []string{"-failfast2"}, Doc: "Pre-check syntax of all inputs and stop faster on error", Level: 0},
	NoErrorImage:           {Name: "NO_ERROR_IMAGE", Flag: "--no-error-image", Arity: UnaryBoolean, Aliases: []string{"-noerror"}, Doc: "Do not generate error images for diagrams with syntax errors", Level: 0},
	Timeout:                {Name: "TIMEOUT", Flag: "--graphviz-timeout", Arity: BinaryNextArgumentValue, Aliases: []string{"-timeout"}, Doc: "Set Graphviz processing timeout (in seconds)", Usage: "--graphviz-timeout <seconds>", Level: 1},
	IgnoreStartumlFilename: {Name: "IGNORE_STARTUML_FILENAME", Flag: "--ignore-startuml-filename", Arity: UnaryBoolean, Doc: "Ignore '@startuml <name>' and always derive output filenames from input files", Level: 1},
	NbThread:               {Name: "NB_THREAD", Flag: "--threads", Arity: BinaryNextArgumentValue, Aliases: []string{"-nbthread"}, Doc: "Use <n> threads for processing  (auto = available processors)", Usage: "--threads <n|auto>", Level: 1},
	Loop:                   {Name: "LOOP", Flag: "--loop", Arity: BinaryNextArgumentValue, Doc: "Runs the generation process <n> times, useful for performance testing", Usage: "--loop <n>", Level: 1, Default: "1"},
	RetrieveMetadata:       {Name: "RETRIEVE_METADATA", Flag: "--extract-source", Arity: UnaryBoolean, Aliases: []string{"-metadata"}, Doc: "Extract embedded PlantUML source from PNG or SVG metadata", Level: 0, NewGroup: "Metadata & assets"},
	NoMetadata:             {Name: "NO_METADATA", Flag: "--disable-metadata", Arity: UnaryBoolean, Aliases: []string{"-nometadata"}, Doc: "Do not include metadata in generated files", Level: 1},
	CheckMetadata:          {Name: "CHECK_METADATA", Flag: "--skip-fresh", Arity: UnaryBoolean, Aliases: []string{"-checkmetadata"}, Doc: "Skip PNG/SVG files that are already up-to-date (using metadata)", Level: 0},
	EncodeSprite:           {Name: "ENCODE_SPRITE", Flag: "--encode-sprite", Arity: UnaryBoolean, Aliases: []string{"-sprite", "-encodesprite"}, Doc: "Encode a sprite definition from an image file", Usage: "--sprite <4|8|16[z]> <file>", Level: 0},
	ComputeUrl:             {Name: "COMPUTE_URL", Flag: "--encode-url", Arity: UnaryBoolean, Aliases: []string{"--compute-url", "-computeurl", "-encodeurl"}, Doc: "Generate an encoded PlantUML URL from a source file", Level: 1},
	DecodeUrl:              {Name: "DECODE_URL", Flag: "--decode-url", Arity: UnaryBoolean, Aliases: []string{"-decodeurl"}, Doc: "Decode a PlantUML encoded URL back to its source", Usage: "--decode-url <string>", Level: 1},
	Language:               {Name: "LANGUAGE", Flag: "--list-keywords", Arity: UnaryImmediateAction, Aliases: []string{"-language"}, Doc: "Print the list of PlantUML language keywords", Level: 1},
	GraphvizDot:            {Name: "GRAPHVIZ_DOT", Flag: "--dot-path", Arity: BinaryNextArgumentValue, Aliases: []string{"-graphvizdot", "-graphviz_dot"}, Doc: "Specify the path to the Graphviz 'dot' executable", Usage: "--dot-path <path-to-dot-exe>", Level: 1},
	Ftp:                    {Name: "FTP", Flag: "--ftp-server", Arity: UnaryOptionalColon, Aliases: []string{"-ftp"}, Doc: "Start a local FTP server for diagram rendering (rarely used)", Level: 1},
	OutputDir:              {Name: "OUTPUT_DIR", Flag: "--output-dir", Arity: BinaryNextArgumentValue, Aliases: []string{"-o", "output_dir", "-output", "-odir"}, Doc: "Generate output files in the specified directory", Usage: "--output-dir <dir>", Level: 1, NewGroup: "Output control"},
	Overwrite:              {Name: "OVERWRITE", Flag: "--overwrite", Arity: UnaryBoolean, Aliases: []string{"--force-overwrite", "-overwrite"}, Doc: "Allow overwriting of read-only output files", Level: 1},
	Clipboard:              {Name: "CLIPBOARD", Flag: "--clipboard", Arity: UnaryImmediateAction, Level: -1},
	Clipboardloop:          {Name: "CLIPBOARDLOOP", Flag: "--clipboardloop", Arity: UnaryImmediateAction, Level: -1},
	DebugSvek:              {Name: "DEBUG_SVEK", Flag: "-debugsvek", Arity: UnaryBoolean, Aliases: []string{"-debug_svek"}, Doc: "Generate intermediate Svek files", Level: 999},
	FileDir:                {Name: "FILE_DIR", Flag: "-filedir", Arity: BinaryNextArgumentValue, Doc: "Pretend input files are located in given directory", Level: 999},
	Filename:               {Name: "FILENAME", Flag: "-filename", Arity: BinaryNextArgumentValue, Doc: "Override %filename% variable", Level: 999},
	Headless:               {Name: "HEADLESS", Flag: "-headless", Arity: UnaryBoolean, Level: -1},
	PrintFonts:             {Name: "PRINT_FONTS", Flag: "-printfonts", Arity: UnaryImmediateAction, Doc: "List fonts available on your system", Level: 999},
	StdLib:                 {Name: "STD_LIB", Flag: "-stdlib", Arity: UnaryImmediateAction, Doc: "Print standard library information", Level: 999},
	Stdrpt:                 {Name: "STDRPT", Flag: "-stdrpt", Arity: UnaryOptionalColon, Level: -1},
	Syntax:                 {Name: "SYNTAX", Flag: "-syntax", Arity: UnaryBoolean, Doc: "Report syntax errors from stdin without generating images", Level: 999},
	License:                {Name: "LICENSE", Flag: "-license", Arity: UnaryImmediateAction, Aliases: []string{"-licence"}, Level: -1},
	Word:                   {Name: "WORD", Flag: "-word", Arity: UnaryBoolean, Level: -1},
	UseSeparatorMinus:      {Name: "USE_SEPARATOR_MINUS", Flag: "-useseparatorminus", Arity: UnaryBoolean, Level: -1},
	Format:                 {Name: "FORMAT", Flag: "--format", Arity: BinaryNextArgumentValue, Aliases: []string{"-f"}, Doc: "Set the output format for generated diagrams\n(e.g. png, svg, pdf, eps, latex, txt, utxt, obfuscate, preproc...)", Usage: "-f, --format <name>", Level: 0, NewGroup: "Output format (choose one)"},
	TEps:                   {Name: "T_EPS", Flag: "--eps", Arity: UnaryBoolean, Aliases: []string{"-teps", "-eps"}, Format: "EPS", Doc: "Generate images in EPS format", Level: 0, NewGroup: "Available formats"},
	TEpsText:               {Name: "T_EPS_TEXT", Flag: "--teps:text", Arity: UnaryBoolean, Aliases: []string{"-teps:text", "-eps:text"}, Format: "EPS_TEXT", Level: -1},
	THtml:                  {Name: "T_HTML", Flag: "--html", Arity: UnaryBoolean, Aliases: []string{"-thtml", "-html"}, Format: "HTML", Doc: "Generate HTML files for class diagrams", Level: 1},
	TLatex:                 {Name: "T_LATEX", Flag: "--latex", Arity: UnaryBoolean, Aliases: []string{"-tlatex", "-latex"}, Format: "LATEX", Doc: "Generate LaTeX/TikZ output", Level: 0},
	TLatexNopreamble:       {Name: "T_LATEX_NOPREAMBLE", Flag: "--latex-nopreamble", Arity: UnaryBoolean, Aliases: []string{"-tlatex:nopreamble", "-latex:nopreamble"}, Format: "LATEX_NO_PREAMBLE", Doc: "Generate LaTeX/TikZ output without preamble", Level: 1},
	Obfuscate:              {Name: "OBFUSCATE", Flag: "--obfuscate", Arity: UnaryBoolean, Aliases: []string{"-cypher"}, Format: "OBFUSCATE", Doc: "Replace text in diagrams with obfuscated strings to share diagrams safely", Level: 0},
	TPdf:                   {Name: "T_PDF", Flag: "--pdf", Arity: UnaryBoolean, Aliases: []string{"-tpdf", "-pdf"}, Format: "PDF", Doc: "Generate PDF images", Level: 1},
	TPng:                   {Name: "T_PNG", Flag: "--png", Arity: UnaryBoolean, Aliases: []string{"-tpng", "-png"}, Format: "PNG", Doc: "Generate PNG images (default)", Level: 0},
	Preprocess:             {Name: "PREPROCESS", Flag: "--preproc", Arity: UnaryBoolean, Aliases: []string{"-preproc"}, Format: "PREPROC", Doc: "Generate the preprocessed source after applying !include, !define... (no rendering)", Level: 0},
	TScxml:                 {Name: "T_SCXML", Flag: "--scxml", Arity: UnaryBoolean, Aliases: []string{"-tscxml"}, Format: "SCXML", Doc: "Generate SCXML files for state diagrams", Level: 1},
	TSvg:                   {Name: "T_SVG", Flag: "--svg", Arity: UnaryBoolean, Aliases: []string{"-tsvg", "-svg"}, Format: "SVG", Doc: "Generate SVG images", Level: 0},
	TNull:                  {Name: "T_NULL", Flag: "--null", Arity: UnaryBoolean, Format: "NULL", Level: -1},
	TTxt:                   {Name: "T_TXT", Flag: "--txt", Arity: UnaryBoolean, Aliases: []string{"-ttxt", "-txt"}, Format: "ATXT", Doc: "Generate ASCII art diagrams", Level: 0},
	TUtxt:                  {Name: "T_UTXT", Flag: "--utxt", Arity: UnaryBoolean, Aliases: []string{"-tutxt", "-utxt"}, Format: "UTXT", Doc: "Generate ASCII art diagrams using Unicode characters", Level: 0},
	TVdx:                   {Name: "T_VDX", Flag: "--vdx", Arity: UnaryBoolean, Aliases: []string{"-tvdx", "-vdx"}, Format: "VDX", Doc: "Generate VDX files", Level: 1},
	TXmi:                   {Name: "T_XMI", Flag: "--xmi", Arity: UnaryBoolean, Aliases: []string{"-txmi", "-xmi"}, Format: "XMI_STANDARD", Doc: "Generate XMI files for class diagrams", Level: 1},
	TXmiArgo:               {Name: "T_XMI_ARGO", Flag: "--xmi:argo", Arity: UnaryBoolean, Aliases: []string{"-xmi:argo"}, Format: "XMI_ARGO", Level: -1},
	TXmiCustom:             {Name: "T_XMI_CUSTOM", Flag: "--xmi:custom", Arity: UnaryBoolean, Aliases: []string{"-xmi:custom"}, Format: "XMI_CUSTOM", Level: -1},
	TXmiScript:             {Name: "T_XMI_SCRIPT", Flag: "--xmi:script", Arity: UnaryBoolean, Aliases: []string{"-xmi:script"}, Format: "XMI_SCRIPT", Level: -1},
	TXmiStar:               {Name: "T_XMI_STAR", Flag: "--xmi:star", Arity: UnaryBoolean, Aliases: []string{"-xmi:star"}, Format: "XMI_STAR", Level: -1},
	TBase64:                {Name: "T_BASE64", Flag: "--base64", Arity: UnaryBoolean, Format: "BASE64", Level: -1},
	TBraille:               {Name: "T_BRAILLE", Flag: "--braille", Arity: UnaryBoolean, Format: "BRAILLE_PNG", Level: -1},
	DisableStats:           {Name: "DISABLE_STATS", Flag: "--disable-stats", Arity: UnaryBoolean, Aliases: []string{"-disablestats"}, Doc: "Disable statistics collection (default behavior)", Level: 1, NewGroup: "Statistics"},
	EnableStats:            {Name: "ENABLE_STATS", Flag: "--enable-stats", Arity: UnaryBoolean, Aliases: []string{"-enablestats"}, Doc: "Enable statistics collection", Level: 1},
	Dumphtmlstats:          {Name: "DUMPHTMLSTATS", Flag: "--export-stats-html", Arity: UnaryImmediateAction, Doc: "Export collected statistics to an HTML report and exit", Level: 1},
	Dumpstats:              {Name: "DUMPSTATS", Flag: "--export-stats", Arity: UnaryImmediateAction, Doc: "Export collected statistics to a text report and exit", Level: 1},
	HtmlStats:              {Name: "HTML_STATS", Flag: "--html-stats", Arity: UnaryBoolean, Aliases: []string{"-htmlstats"}, Doc: "Output general statistics in HTML format", Level: 1},
	XmlStats:               {Name: "XML_STATS", Flag: "--xml-stats", Arity: UnaryBoolean, Aliases: []string{"-xmlstats"}, Doc: "Output general statistics in XML format", Level: 1},
	RealtimeStats:          {Name: "REALTIME_STATS", Flag: "--realtime-stats", Arity: UnaryBoolean, Aliases: []string{"-realtimestats"}, Doc: "Generate statistics in real time during processing", Level: 1},
	LoopStats:              {Name: "LOOP_STATS", Flag: "--loop-stats", Arity: UnaryImmediateAction, Aliases: []string{"-loopstats"}, Doc: "Continuously print usage statistics during execution", Level: 1},
}
