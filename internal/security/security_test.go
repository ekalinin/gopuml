// SPDX-License-Identifier: MIT
// Provenance: original

package security

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestParseProfile(t *testing.T) {
	for in, want := range map[string]Profile{"sandbox": Sandbox, "ALLOWLIST": Allowlist, "Internet": Internet, "INTERNET_WITH_DOTSVG": InternetWithDotSVG, "legacy": Legacy, "UNSECURE": Insecure} {
		if got, ok := ParseProfile(in); !ok || got != want {
			t.Errorf("ParseProfile(%q) = %v, %v", in, got, ok)
		}
	}
	for _, bad := range []string{"", "INSECURE", "strict"} {
		if _, ok := ParseProfile(bad); ok {
			t.Errorf("ParseProfile(%q) must be unknown", bad)
		}
	}
}

func TestConfigFromEnv(t *testing.T) {
	inc := t.TempDir()
	env := map[string]string{"PLANTUML_SECURITY_PROFILE": "sandbox", "PLANTUML_INCLUDE_PATH": inc + string(filepath.ListSeparator) + "/does/not/exist"}
	c := ConfigFromEnv(func(k string) string { return env[k] }, Legacy)
	real, _ := filepath.EvalSymlinks(inc)
	if c.Profile != Sandbox || len(c.IncludePaths) != 1 || c.IncludePaths[0] != real {
		t.Errorf("config = %+v", c)
	}
	if c := ConfigFromEnv(func(string) string { return "" }, Internet); c.Profile != Internet {
		t.Errorf("default profile = %v", c.Profile)
	}
}

func TestFileOK(t *testing.T) {
	tests := []struct {
		profile Profile
		path    string
		want    bool
	}{
		// Spec scenarios.
		{Sandbox, "/work/common.iuml", false}, // Sandbox blocks local files
		{Legacy, "/work/common.iuml", true},   // Legacy allows local files
		{Legacy, "/etc/passwd", false},        // System paths are denied by default
		{Insecure, "/etc/passwd", true},
		{Internet, "/work/common.iuml", false},
		{Allowlist, "/work/common.iuml", false},
		{Legacy, "/proc/self/environ", false},
		{Legacy, "//server/share/x.iuml", false},
		{Legacy, "/work/etc/passwd", true},
		{Legacy, "/work/a\x00b", true},
	}
	for _, tt := range tests {
		c := Config{Profile: tt.profile}
		if got := c.FileOK(tt.path); got != tt.want {
			t.Errorf("%v FileOK(%q) = %v, want %v", tt.profile, tt.path, got, tt.want)
		}
	}
	c := Config{Profile: Internet, IncludePaths: []string{"/lib/puml"}, SecurityFolder: "/secrets"}
	if !c.FileOK("/lib/puml/style.iuml") {
		t.Error("include path files are allowed in every profile but sandbox")
	}
	c.Profile, c.AllowlistPaths = Legacy, []string{"/secrets"}
	if c.FileOK("/secrets/token.json") {
		t.Error("the security folder is never readable")
	}
}

func publicIP(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }

func TestURLOK(t *testing.T) {
	tests := []struct {
		profile Profile
		url     string
		want    bool
	}{
		{Internet, "https://example.com/style.iuml", true}, // URL include under the internet profile
		{Internet, "https://example.com:8443/style.iuml", false},
		{Legacy, "https://example.com:8443/style.iuml", true},
		{Sandbox, "https://example.com/style.iuml", false},
		{Allowlist, "https://example.com/style.iuml", false},
		{Legacy, "https://user@example.com/x", true}, // user info names stored credentials
		{Legacy, "https://example.com/a@b", false},
		{Legacy, "ftp://example.com/x", false},
		{Legacy, "https://10.0.0.1/x", false},
		{Legacy, "https://localhost/x", false},
		{Insecure, "http://localhost:8080/x", true},
	}
	for _, tt := range tests {
		c := Config{Profile: tt.profile, LookupIP: publicIP}
		if got := c.URLOK(tt.url); got != tt.want {
			t.Errorf("%v URLOK(%q) = %v, want %v", tt.profile, tt.url, got, tt.want)
		}
	}
	private := Config{Profile: Legacy, LookupIP: func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("192.168.1.5")}, nil }}
	if private.URLOK("https://intranet.example.com/x") {
		t.Error("hosts resolving to private addresses are forbidden")
	}
	allow := Config{Profile: Allowlist, AllowlistURLs: []string{"https://cdn.example.com/"}, LookupIP: publicIP}
	if !allow.URLOK("https://CDN.example.com/lib.iuml") || allow.URLOK("https://other.example.com/lib.iuml") {
		t.Error("allowlist prefixes")
	}
}

func TestAccess(t *testing.T) {
	fsys := fstest.MapFS{"work/common.iuml": {Data: []byte("x")}}
	a := &Access{Config: Config{Profile: Legacy}, FS: fsys}
	if data, err := a.ReadFile("/work/common.iuml"); err != nil || string(data) != "x" {
		t.Errorf("ReadFile = %q, %v", data, err)
	}
	if _, err := a.ReadFile("/etc/passwd"); !errors.Is(err, ErrDenied) {
		t.Errorf("ReadFile(/etc/passwd) error = %v", err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.iuml"), []byte("os"), 0o644)
	osAccess := &Access{Config: Config{Profile: Insecure}}
	if data, err := osAccess.ReadFile(filepath.Join(dir, "a.iuml")); err != nil || string(data) != "os" {
		t.Errorf("OS ReadFile = %q, %v", data, err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("remote")) }))
	defer srv.Close()
	web := &Access{Config: Config{Profile: Insecure}, HTTP: srv.Client()}
	if data, err := web.ReadURL(t.Context(), srv.URL+"/x"); err != nil || string(data) != "remote" {
		t.Errorf("ReadURL = %q, %v", data, err)
	}
	web.Config.Profile = Sandbox
	if _, err := web.ReadURL(context.Background(), srv.URL+"/x"); !errors.Is(err, ErrDenied) {
		t.Errorf("sandbox ReadURL error = %v", err)
	}
}
