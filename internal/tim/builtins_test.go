// SPDX-License-Identifier: MIT
// Provenance: original

package tim

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ekalinin/plantuml-in-go/internal/core"
	"github.com/ekalinin/plantuml-in-go/internal/security"
)

// The builtins are compared with the reference jar by the differential
// cases of testdata/upstream/cases/builtins_*.txt. These tests cover what
// those cases cannot: seeded random numbers, a fixed clock and time zone,
// bundled themes and libraries, the file environment and the process
// environment.

// preprocessBuiltins preprocesses "@startuml", the lines and "@enduml" and
// returns the lines without the start and end lines, or the error.
func preprocessBuiltins(t *testing.T, access *security.Access, opts Options, defines *Defines, lines ...string) ([]string, error) {
	t.Helper()
	if access == nil {
		access = &security.Access{Config: security.Config{Profile: security.Legacy}, FS: fstest.MapFS{}}
	}
	opts.Defines = defines
	ctx := core.NewRenderContext()
	ctx.CurrentDir = "/"
	src := "@startuml\n" + strings.Join(lines, "\n") + "\n@enduml\n"
	blocks := core.BuildBlocks(src, "string", nil)
	res, err := PreprocessWith(ctx, access, blocks[0].Lines, opts)
	out := res.PreprocLines()
	if err != nil {
		return out, err
	}
	return out[1 : len(out)-1], nil
}

func checkLines(t *testing.T, got []string, err error, want ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("error %v; lines %q", err, got)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestJavaRandom(t *testing.T) {
	// Values of java.util.Random in the reference JVM for the bounds 2, 10,
	// 16, 2^30+1, 2^31-1, 3, 2^30+1, 2^30+1 and 2^30.
	bounds := []int32{2, 10, 16, 1<<30 + 1, 1<<31 - 1, 3, 1<<30 + 1, 1<<30 + 1, 1 << 30}
	tests := []struct {
		seed int64
		want []int32
	}{
		{42, []int32{1, 3, 10, 102948884, 662969970, 1, 595021505, 196118093, 969988591}},
		{0, []int32{1, 8, 3, 663681053, 1182054491, 0, 715581077, 542832677, 413593736}},
		{-1, []int32{0, 5, 0, 894294477, 857465478, 1, 121412731, 665220512, 60977158}},
		{123456789012345, []int32{1, 4, 13, 185864260, 1992532989, 1, 849765284, 146958768, 300431744}},
	}
	for _, tt := range tests {
		r := newJavaRandom(tt.seed)
		var got []int32
		for _, b := range bounds {
			v, ok := r.nextInt(b)
			if !ok {
				t.Fatalf("seed %d: nextInt(%d) failed", tt.seed, b)
			}
			got = append(got, v)
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("seed %d: %v, want %v", tt.seed, got, tt.want)
		}
	}
	if _, ok := newJavaRandom(1).nextInt(0); ok {
		t.Error("nextInt(0) succeeded")
	}
}

func TestRandomSeed(t *testing.T) {
	seed := int64(42)
	got, err := preprocessBuiltins(t, nil, Options{RandomSeed: &seed}, nil,
		"A : %random(2) %random(10) %random(16)", "!$r = %random(10, 20)", "A : $r")
	// 10 + the fourth value of the generator below 10
	r := newJavaRandom(42)
	for _, b := range []int32{2, 10, 16} {
		r.nextInt(b)
	}
	v, _ := r.nextInt(10)
	checkLines(t, got, err, "A : 1 3 10", "A : "+itoa(10+v))
}

func itoa(v int32) string { return TValueFromInt(v).String() }

func TestDateWithClockAndZone(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 16, 10, 34, 56, 7e6, time.UTC)
	opts := Options{Now: func() time.Time { return now }, Location: paris}
	got, err := preprocessBuiltins(t, nil, opts, nil,
		"A : %date()",
		"A : %date(\"yyyy-MM-dd HH:mm:ss.SSS S z zzzz Z XXX\")",
		"A : %date(\"HH:mm z\", 0)",
		"A : %now()",
		"A : %date(\"HH:mm z\", %now(), \"Asia/Tokyo\")")
	checkLines(t, got, err,
		"A : Fri Oct 16 12:34:56 CEST 2026",
		"A : 2026-10-16 12:34:56.007 7 CEST Central European Summer Time +0200 +02:00",
		"A : 01:00 CET",
		"A : 1792146896",
		"A : 19:34 JST")

	// Zones without Java names
	for _, tt := range []struct {
		loc  *time.Location
		want string
	}{
		{time.FixedZone("XYZ", 3600), "A : 11:34 XYZ XYZ +0100"},
		{time.FixedZone("", 5400), "A : 12:04 GMT+01:30 GMT+01:30 +0130"},
		{time.FixedZone("+03", -3*3600-1800), "A : 07:04 GMT-03:30 GMT-03:30 -0330"},
	} {
		opts.Location = tt.loc
		got, err := preprocessBuiltins(t, nil, opts, nil, "A : %date(\"HH:mm z zzzz Z\")")
		checkLines(t, got, err, tt.want)
	}

	// %now wraps like Java's (int) cast after 2038
	opts.Now = func() time.Time { return time.Unix(1<<31+5, 0) }
	got, err = preprocessBuiltins(t, nil, opts, nil, "A : %now()")
	checkLines(t, got, err, "A : -2147483643")
}

func TestJavaDateFormatErrors(t *testing.T) {
	for _, p := range []string{"'", "yyyy'", "q", "XXXX", "Xxxxx", "yyyy-MM-dd t"} {
		if _, err := newJavaDateFormat(p); err == nil {
			t.Errorf("%q: no error", p)
		}
	}
	for _, p := range []string{"", "''", "XXX", "GyMdkHmsSEDFwWahKzZYuXL", "'q'", "é!"} {
		if _, err := newJavaDateFormat(p); err != nil {
			t.Errorf("%q: %v", p, err)
		}
	}
}

func TestSystemVZone(t *testing.T) {
	z, ok := lookupJavaZone("SystemV/EST5EDT")
	if !ok {
		t.Fatal("SystemV/EST5EDT not found")
	}
	f, _ := newJavaDateFormat("yyyy-MM-dd HH:mm z")
	for _, tt := range []struct {
		t    time.Time
		want string
	}{
		// 2:00 EST on the last Sunday of April, 2:00 EDT on the last Sunday of October
		{time.Date(2020, 4, 26, 6, 59, 0, 0, time.UTC), "2020-04-26 01:59 EST"},
		{time.Date(2020, 4, 26, 7, 0, 0, 0, time.UTC), "2020-04-26 03:00 EDT"},
		{time.Date(2020, 10, 25, 5, 59, 0, 0, time.UTC), "2020-10-25 01:59 EDT"},
		{time.Date(2020, 10, 25, 6, 0, 0, 0, time.UTC), "2020-10-25 01:00 EST"},
		{time.Date(1974, 1, 6, 7, 0, 0, 0, time.UTC), "1974-01-06 03:00 EDT"},
		{time.Date(1975, 2, 23, 6, 59, 0, 0, time.UTC), "1975-02-23 01:59 EST"},
	} {
		if got := f.format(tt.t, z); got != tt.want {
			t.Errorf("%v: %s, want %s", tt.t, got, tt.want)
		}
	}
	if _, ok := lookupJavaZone("Mars/Base"); ok {
		t.Error("unknown zone found")
	}
}

// themeList is a ThemeSource that can list its themes.
type themeList []string

func (l themeList) Theme(name string) ([]byte, bool) {
	if slices.Contains(l, name) {
		return []byte("---\nname: " + name + "\n---\n"), true
	}
	return nil, false
}

func (l themeList) ThemeNames() []string { return l }

func TestThemeBuiltins(t *testing.T) {
	got, err := preprocessBuiltins(t, nil, Options{Themes: themeList{"b", "a", "B", "_x"}}, nil,
		"A : %get_all_theme()", "!theme a", "A : %get_current_theme()")
	checkLines(t, got, err, `A : ["B","_x","a","b"]`, `A : {"name":"a"}`)
}

func TestStdlibBuiltins(t *testing.T) {
	stdlib := NewStdlib(map[string]map[string][]byte{
		"lib1": {
			"info.spm":  []byte("VERSION=1.0\nsource=src\r\nName=Lib\nversion=v2\n=empty\nbad\nx=y=z\na=\n"),
			"data.json": []byte(`{"k": [1]}`),
			"bad.json":  []byte(`{`),
			"f.puml":    []byte("A -> B\n"),
		},
		"lib0": {"info.spm": []byte("version=2")},
	})
	opts := Options{Stdlib: stdlib}
	got, err := preprocessBuiltins(t, nil, opts, nil,
		"A : %get_all_stdlib()",
		"A : %get_all_stdlib(1)",
		"A : %get_stdlib()",
		"A : %get_stdlib(\"lib1\")",
		"A : [%get_stdlib(\"lib1\", \"name\")] %get_stdlib(\"lib1\", \"Version\") %get_stdlib(\"lib1\", \"source\") [%get_stdlib(\"lib1\", \"\")]",
		"A : %load_json(\"<lib1/data.json>\") %load_json(\"<lib1>\")",
		"!include <lib1/f>")
	checkLines(t, got, err,
		`A : ["lib0","lib1"]`,
		`A : {"lib0":{"name":"lib0","version":"2","source":null},"lib1":{"name":"lib1","version":"1.0","source":"src"}}`,
		`A : {"lib0":{"version":"2"},"lib1":{"":"empty","version":"1.0","source":"src","version":"v2","name":"Lib"}}`,
		`A : {"":"empty","version":"1.0","source":"src","version":"v2","name":"Lib"}`,
		`A : [] v2 src [empty]`,
		`A : {"k":[1]} {}`,
		"A -> B")

	for _, tt := range []struct{ line, err string }{
		{`A : %load_json("<lib1/missing.json>")`, "Fatal parsing error"},
		{`A : %load_json("<LIB1/data.json>")`, "Fatal parsing error"},
		{`A : %load_json("<lib1/bad.json>")`, "JSON parse issue in source <lib1/bad.json> on location 1:2"},
		{`A : %get_stdlib("LIB1")`, "Fatal parsing error"},
	} {
		_, err := preprocessBuiltins(t, nil, opts, nil, tt.line)
		if err == nil || err.Error() != tt.err {
			t.Errorf("%s: error %v, want %s", tt.line, err, tt.err)
		}
	}
}

func TestFileEnvironmentBuiltins(t *testing.T) {
	modTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	defines := DefinesWithFileName("my.diagram.puml", modTime, `C:\dir\sub`)
	got, err := preprocessBuiltins(t, nil, Options{}, defines,
		"A : %filename() %filename_no_extension() %dirpath() %filedate()")
	checkLines(t, got, err, "A : my.diagram.puml my.diagram C:/dir/sub Fri Jan 02 03:04:05 UTC 2026")
}

func TestGetenvDefault(t *testing.T) {
	t.Setenv("PLANTUML_TIM_TEST", "v1")
	t.Setenv("TIM_TEST_VARIABLE", "v2")
	legacy := &security.Access{Config: security.Config{Profile: security.Legacy}, FS: fstest.MapFS{}}
	got, err := preprocessBuiltins(t, legacy, Options{}, nil,
		`A : %getenv("PLANTUML_TIM_TEST") [%getenv("TIM_TEST_VARIABLE")] %strlen(%getenv("line.separator")) [%getenv("file.separator")]`)
	checkLines(t, got, err, "A : v1 [] 1 []")

	insecure := &security.Access{Config: security.Config{Profile: security.Insecure}, FS: fstest.MapFS{}}
	got, err = preprocessBuiltins(t, insecure, Options{}, nil,
		`A : %getenv("TIM_TEST_VARIABLE") %getenv("file.separator") %getenv("path.separator")`)
	checkLines(t, got, err, "A : v2 "+string(os.PathSeparator)+" "+string(os.PathListSeparator))

	lookup := func(name string) (string, bool) { return strings.ToUpper(name), name != "none" }
	got, err = preprocessBuiltins(t, insecure, Options{Getenv: lookup}, nil, `A : %getenv("abc") [%getenv("none")]`)
	checkLines(t, got, err, "A : ABC []")
}

func TestFileExists(t *testing.T) {
	fsys := fstest.MapFS{"data/x.txt": {Data: []byte("x")}, "etc/passwd": {Data: []byte("x")}}
	legacy := &security.Access{Config: security.Config{Profile: security.Legacy}, FS: fsys}
	got, err := preprocessBuiltins(t, legacy, Options{}, nil,
		`A : %file_exists("/data/x.txt") %file_exists("data/x.txt") %file_exists("/data") %file_exists("/data//x.txt/")`,
		`A : %file_exists("/etc/passwd") %file_exists("/data/../etc/passwd") %file_exists("/data/y.txt")`)
	checkLines(t, got, err, "A : 1 1 1 1", "A : 0 1 0")

	// Without an FS, a relative path is resolved against the working
	// directory of the process, the package directory in tests.
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	osAccess := &security.Access{Config: security.Config{Profile: security.Legacy}}
	got, err = preprocessBuiltins(t, osAccess, Options{}, nil,
		`A : %file_exists("`+filepath.ToSlash(file)+`") %file_exists("builtins_test.go") %file_exists("nofile.go")`)
	checkLines(t, got, err, "A : 1 1 0")
}

func TestLoadJSONURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/empty" {
			return
		}
		_, _ = w.Write([]byte(`{"u": 1}`))
	}))
	defer server.Close()
	insecure := &security.Access{Config: security.Config{Profile: security.Insecure}, FS: fstest.MapFS{}, HTTP: server.Client()}
	got, err := preprocessBuiltins(t, insecure, Options{}, nil,
		`A : %load_json("`+server.URL+`/x") %load_json("`+server.URL+`/empty", "[1]")`)
	checkLines(t, got, err, `A : {"u":1} [1]`)

	// The INTERNET profile denies local addresses.
	internet := &security.Access{Config: security.Config{Profile: security.Internet}, FS: fstest.MapFS{}, HTTP: server.Client()}
	got, err = preprocessBuiltins(t, internet, Options{}, nil, `A : %load_json("`+server.URL+`/x", "[2]")`)
	checkLines(t, got, err, `A : [2]`)
}

func TestSplitStrRegexTimeout(t *testing.T) {
	re, err := compileJavaRegex(`(a+)+$`)
	if err != nil {
		t.Fatal(err)
	}
	re.MatchTimeout = 50 * time.Millisecond
	start := time.Now()
	if _, err := javaSplit(strings.Repeat("a", 40)+"b", re); err != errRegexTimeout {
		t.Errorf("error %v, want a timeout", err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("the timeout took %v", d)
	}
}

func TestTranslateJavaRegex(t *testing.T) {
	tests := []struct{ java, net string }{
		{`a\d+`, `a[0-9]+`},
		{`a*+b`, `(?>a*)b`},
		{`(a|b){2,3}+`, `(?>(a|b){2,3})`},
		{`\Q.*\E`, `\.\*`},
		{`(?i)ab`, `[aA][bB]`},
		{`[a-c&&[^b]]`, `(?:(?=[^b])[a-c])`},
		{`(?<x>a)\k<x>`, `(a)(?:\1)`},
		{`\2(a)`, `(?!)(a)`},
		{`.`, `[^\n\r\u0085\u2028\u2029]`},
		{`(?s).`, `[\s\S]`},
		{`\x{1F600}`, "\U0001F600"},
		{`{2}a`, `(?:){2}a`},
		{`\p{IsLatin}`, `[\p{Latin}]`},
		{`\p{Alpha}`, `[A-Za-z]`},
		{`_`, `_`},
	}
	for _, tt := range tests {
		got, err := translateJavaRegex(tt.java)
		if err != nil || got != tt.net {
			t.Errorf("%q: %q, %v; want %q", tt.java, got, err, tt.net)
		}
	}
	for _, p := range []string{`*`, `a{`, `(`, `)`, `[a`, `\`, `\y`, `[z-a]`, `\p{InGreek}`, `a{3,2}`, `(?<1>a)`, `\k<n>`} {
		if got, err := translateJavaRegex(p); err == nil {
			t.Errorf("%q: no error, got %q", p, got)
		}
	}
}

func TestJavaStringHelpers(t *testing.T) {
	for _, tt := range []struct {
		s     string
		radix int
		v     int32
		ok    bool
	}{
		{"ff", 16, 255, true}, {"-80000000", 16, -1 << 31, true}, {"80000000", 16, 0, false},
		{"+7fffffff", 16, 1<<31 - 1, true}, {"", 16, 0, false}, {"-", 16, 0, false}, {"ｆ", 16, 15, true},
		{"z", 36, 35, true}, {"g", 16, 0, false}, {"٩", 10, 9, true},
	} {
		v, ok := javaParseIntRadix(tt.s, tt.radix)
		if v != tt.v || ok != tt.ok {
			t.Errorf("javaParseIntRadix(%q, %d) = %d, %v; want %d, %v", tt.s, tt.radix, v, ok, tt.v, tt.ok)
		}
	}
	for in, want := range map[string]string{"ABC": "abc", "İ": "i̇", "ΣΑΣ": "σας", "Σ": "σ", "ΑΣ Β": "ας β"} {
		if got := javaToLowerCase(in); got != want {
			t.Errorf("javaToLowerCase(%q) = %q, want %q", in, got, want)
		}
	}
	if javaIndexOf("😀ab", "b") != 3 || javaIndexOf("abc", "x") != -1 {
		t.Error("javaIndexOf")
	}
	if !javaEqualsIgnoreCase("ſtyle", "STYLE") || javaEqualsIgnoreCase("style", "styles") {
		t.Error("javaEqualsIgnoreCase")
	}
}
