// SPDX-License-Identifier: MIT
// Provenance: original

package themes_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ekalinin/gopuml"
	"github.com/ekalinin/gopuml/internal/core"
	"github.com/ekalinin/gopuml/internal/security"
	"github.com/ekalinin/gopuml/internal/themes"
	"github.com/ekalinin/gopuml/internal/tim"
)

var (
	_ tim.ThemeSource = themes.Bundled
	_ tim.ThemeLister = themes.Bundled
)

// bundled lists the themes of upstream v1.2026.9 that are bundled, in
// Java's sort order.
var bundled = []string{
	"aws-orange", "black-knight", "bluegray", "cerulean", "cerulean-outline", "cloudscape-design",
	"cyborg", "cyborg-outline", "hacker", "lightgray", "mars", "materia", "materia-outline", "metal",
	"minty", "reddress-darkblue", "reddress-darkgreen", "reddress-darkorange", "reddress-darkred",
	"reddress-lightblue", "reddress-lightgreen", "reddress-lightorange", "reddress-lightred",
	"sandstone", "silver", "sketchy", "sketchy-outline", "spacelab", "spacelab-white", "superhero",
	"superhero-outline", "toy", "united", "vibrant",
}

// excluded lists the themes of upstream v1.2026.9 that are not bundled,
// for the license reasons given in NOTICE.
var excluded = []string{
	"_none_", "amiga", "blueprint", "carbon-gray", "crt-amber", "crt-green", "mimeograph", "mono",
	"plain", "sunlust",
}

func TestThemeNames(t *testing.T) {
	if got := themes.Bundled.ThemeNames(); !slices.Equal(got, bundled) {
		t.Errorf("ThemeNames() = %q, want %q", got, bundled)
	}
}

// TestUpstreamThemes checks the bundled files against upstream's themes/
// folder, when the unpacked sources jar is available: every upstream
// theme is either bundled unchanged or excluded.
func TestUpstreamThemes(t *testing.T) {
	dir := filepath.Join("..", "..", ".cache", "upstream", "src", "themes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("upstream sources not available: %v", err)
	}
	var names []string
	for _, e := range entries {
		name, ok := strings.CutPrefix(e.Name(), "puml-theme-")
		if !ok {
			continue
		}
		name, ok = strings.CutSuffix(name, ".puml")
		if !ok {
			continue
		}
		names = append(names, name)
		want, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		got, ok := themes.Bundled.Theme(name)
		switch {
		case slices.Contains(excluded, name):
			if ok {
				t.Errorf("excluded theme %s is bundled", name)
			}
		case !ok:
			t.Errorf("theme %s is neither bundled nor excluded", name)
		case !bytes.Equal(got, want):
			t.Errorf("theme %s differs from upstream", name)
		}
	}
	all := slices.Concat(bundled, excluded)
	slices.Sort(all)
	slices.Sort(names)
	if !slices.Equal(names, all) {
		t.Errorf("upstream themes %q, want %q", names, all)
	}
}

func TestThemeLookup(t *testing.T) {
	got, ok := themes.Bundled.Theme("cerulean")
	if !ok || !bytes.HasPrefix(got, []byte("---\nname: cerulean\n")) {
		t.Fatalf("Theme(cerulean) = %.40q, %v", got, ok)
	}
	// The name must be the exact middle part of the file name, as for
	// upstream's jar entries.
	for _, name := range []string{
		"", "Cerulean", "CERULEAN", "puml-theme-cerulean", "cerulean.puml", " cerulean", "cerulean ",
		"./cerulean", "x/../cerulean", "../themes/puml-theme-cerulean", "*", "cerulean*", "sunlust",
		"amiga", "_none_", "does-not-exist",
	} {
		if _, ok := themes.Bundled.Theme(name); ok {
			t.Errorf("Theme(%q) found a theme", name)
		}
	}
	// Each call returns its own copy.
	got[0] = 'X'
	if again, _ := themes.Bundled.Theme("cerulean"); again[0] != '-' {
		t.Error("Theme returned shared data")
	}
}

// preprocess runs the preprocessor with the bundled themes on src.
func preprocess(t *testing.T, src string, access *security.Access) (*tim.Result, error) {
	t.Helper()
	blocks := core.BuildBlocks(src, "string", nil)
	if len(blocks) == 0 {
		t.Fatal("no block")
	}
	return tim.PreprocessWith(core.NewRenderContext(), access, blocks[0].Lines, tim.Options{Themes: themes.Bundled})
}

func memAccess(files map[string]string) *security.Access {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return &security.Access{Config: security.Config{Profile: security.Legacy}, FS: fsys}
}

func wantError(t *testing.T, err error, message string, position int) {
	t.Helper()
	e, ok := errors.AsType[*tim.EaterException](err)
	if !ok {
		t.Fatalf("got %v, want the error %q", err, message)
	}
	if e.Message != message {
		t.Errorf("message %q, want %q", e.Message, message)
	}
	if e.Location.Location == nil || e.Location.Location.Position != position {
		t.Errorf("error at %+v, want position %d", e.Location.Location, position)
	}
}

func TestBundledTheme(t *testing.T) {
	// Spec scenario "Bundled theme".
	res, err := preprocess(t, "@startuml\n!theme cerulean\nA : $THEME\nAlice -> Bob : hello\n@enduml\n", memAccess(nil))
	if err != nil {
		t.Fatal(err)
	}
	lines := res.PreprocLines()
	if !slices.Contains(lines, "A : cerulean") || !slices.Contains(lines, "<style>") {
		t.Errorf("lines %q", lines)
	}
	if got := res.ThemeMetadata.String(); !strings.Contains(got, `"name":"cerulean"`) || !strings.Contains(got, `"license":"MIT"`) {
		t.Errorf("metadata %s", got)
	}
	// Every bundled theme loads without error.
	for _, name := range bundled {
		if _, err := preprocess(t, "@startuml\n!theme "+name+"\nAlice -> Bob : hello\n@enduml\n", memAccess(nil)); err != nil {
			t.Errorf("theme %s: %v", name, err)
		}
	}
}

func TestUnknownTheme(t *testing.T) {
	// Spec scenario "Unknown theme": the error names the theme, at the
	// directive line. The excluded themes give the same error.
	for _, name := range append([]string{"does-not-exist", "Cerulean"}, excluded...) {
		_, err := preprocess(t, "@startuml\nAlice -> Bob : hello\n!theme "+name+"\n@enduml\n", memAccess(nil))
		wantError(t, err, "Cannot load theme "+name, 2)
	}
}

func TestLocalTheme(t *testing.T) {
	files := map[string]string{
		"puml-theme-mine.puml":     "skinparam backgroundColor red\n",
		"puml-theme-cerulean.puml": "skinparam backgroundColor local\n",
		"puml-theme-sunlust.puml":  "skinparam backgroundColor sun\n",
		"dir/puml-theme-mine.puml": "skinparam backgroundColor fromdir\n",
		"dir/puml-theme-cerulean.puml": "---\nname: dircerulean\n---\n" +
			"skinparam backgroundColor dircerulean\n",
	}
	tests := []struct{ src, want string }{
		// Local files are found when no bundled theme has the name.
		{"!theme mine", "skinparam backgroundColor red"},
		{"!theme sunlust", "skinparam backgroundColor sun"},
		// "from" reads the directory, not the bundled theme.
		{"!theme mine from dir", "skinparam backgroundColor fromdir"},
		{"!theme cerulean from dir", "skinparam backgroundColor dircerulean"},
		{"!theme cerulean from dir/", "skinparam backgroundColor dircerulean"},
		{"!theme cerulean FROM dir", "skinparam backgroundColor dircerulean"},
		{"!$d = \"dir\"\n!theme mine from $d", "skinparam backgroundColor fromdir"},
	}
	for _, tt := range tests {
		res, err := preprocess(t, "@startuml\n"+tt.src+"\n@enduml\n", memAccess(files))
		if err != nil {
			t.Errorf("%q: %v", tt.src, err)
			continue
		}
		if lines := res.PreprocLines(); !slices.Equal(lines, []string{"@startuml", tt.want, "@enduml"}) {
			t.Errorf("%q: lines %q", tt.src, lines)
		}
	}
	// Bundled themes come before local files with the same name.
	res, err := preprocess(t, "@startuml\n!theme cerulean\n@enduml\n", memAccess(files))
	if err != nil {
		t.Fatal(err)
	}
	if lines := res.PreprocLines(); slices.Contains(lines, "skinparam backgroundColor local") {
		t.Errorf("local file used instead of the bundled theme: %q", lines)
	}
	// A theme that is missing in a "from" location fails as upstream does,
	// even when a bundled theme has the name: a library reports a runtime
	// error, a directory names the theme.
	_, err = preprocess(t, "@startuml\n!theme cerulean from <lib>\n@enduml\n", memAccess(files))
	wantError(t, err, "Fatal parsing error", 1)
	for _, c := range []struct{ src, want string }{
		{"!theme cerulean from nowhere", "Cannot load theme cerulean in nowhere"},
		{"!theme other from dir", "Cannot load theme other in dir"},
	} {
		_, err = preprocess(t, "@startuml\n"+c.src+"\n@enduml\n", memAccess(files))
		wantError(t, err, c.want, 1)
	}
}

// fakeResolver makes every host resolve to a public address, so that the
// URL checks of the security profile accept the test server.
func fakeResolver(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }

func TestURLTheme(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path == "/themes/puml-theme-cerulean.puml" {
			w.Write([]byte("skinparam backgroundColor remote\n"))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	addr := server.Listener.Addr().String()
	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}}
	access := func(p security.Profile) *security.Access {
		return &security.Access{Config: security.Config{Profile: p, LookupIP: fakeResolver}, HTTP: client}
	}

	res, err := preprocess(t, "@startuml\n!theme cerulean from http://example.com/themes\n@enduml\n", access(security.Internet))
	if err != nil {
		t.Fatal(err)
	}
	if lines := res.PreprocLines(); !slices.Equal(lines, []string{"@startuml", "skinparam backgroundColor remote", "@enduml"}) {
		t.Errorf("lines %q", lines)
	}
	if !slices.Equal(paths, []string{"/themes/puml-theme-cerulean.puml"}) {
		t.Errorf("requested %q", paths)
	}

	_, err = preprocess(t, "@startuml\n!theme mars from http://example.com/themes/\n@enduml\n", access(security.Internet))
	wantError(t, err, "Cannot open URL", 1)

	// The security profile applies to theme URLs.
	paths = nil
	_, err = preprocess(t, "@startuml\n!theme cerulean from http://example.com/themes\n@enduml\n", access(security.Sandbox))
	wantError(t, err, "Cannot open URL", 1)
	if len(paths) != 0 {
		t.Errorf("sandbox requested %q", paths)
	}
}

// sourceLine matches the source line attribute of a group.
var sourceLine = regexp.MustCompile(` data-source-line="\d+"`)

// TestLibraryTheme checks gopuml.Options.Theme: it acts as a !theme line
// placed before the Config lines, as upstream's --theme comes before
// --config, -P and -S.
func TestLibraryTheme(t *testing.T) {
	render := func(src string, opts gopuml.Options) []byte {
		t.Helper()
		opts.NoMetadata = true
		res, err := gopuml.Render(context.Background(), src, &opts)
		if err != nil {
			t.Fatal(err)
		}
		// The !theme and skinparam lines shift the source lines of the groups.
		return sourceLine.ReplaceAll(res.Data, nil)
	}
	const body = "Alice -> Bob : hello\n@enduml\n"
	want := render("@startuml\n!theme cerulean\nskinparam backgroundColor red\n"+body, gopuml.Options{})
	got := render("@startuml\n"+body, gopuml.Options{Theme: "cerulean", Config: []string{"skinparam backgroundColor red"}})
	if !bytes.Equal(got, want) {
		t.Errorf("Theme and Config render\n%s\nwant\n%s", got, want)
	}
	if plain := render("@startuml\n"+body, gopuml.Options{}); bytes.Equal(plain, want) {
		t.Error("the theme changed nothing")
	}

	// An excluded theme is an unknown theme.
	opts := &gopuml.Options{Theme: "sunlust", NoErrorImage: true}
	_, err := gopuml.Render(context.Background(), "@startuml\n"+body, opts)
	if e, ok := errors.AsType[*gopuml.SyntaxError](err); !ok || e.Message != "Cannot load theme sunlust" {
		t.Errorf("got %v, want the unknown theme error", err)
	}
}

func TestGetAllTheme(t *testing.T) {
	res, err := preprocess(t, "@startuml\nA : %get_all_theme()\nA : %size(%get_all_theme())\n@enduml\n", memAccess(nil))
	if err != nil {
		t.Fatal(err)
	}
	want := `A : ["` + strings.Join(bundled, `","`) + `"]`
	if lines := res.PreprocLines(); len(lines) != 4 || lines[1] != want || lines[2] != "A : 34" {
		t.Errorf("lines %q, want %q", lines, want)
	}
}
