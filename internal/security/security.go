// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/SecurityProfile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/SecurityUtils.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/SFile.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/SURL.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/security/URLCheck.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package security

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Profile limits which files and URLs diagrams may read.
type Profile int

// Profiles, from the most to the least restrictive.
const (
	Sandbox Profile = iota
	Allowlist
	Internet
	InternetWithDotSVG
	Legacy
	Insecure
)

var profileNames = [...]string{"SANDBOX", "ALLOWLIST", "INTERNET", "INTERNET_WITH_DOTSVG", "LEGACY", "INSECURE"}

func (p Profile) String() string { return profileNames[p] }

// ParseProfile reads a PLANTUML_SECURITY_PROFILE value. "INSECURE" and
// its misspelling "UNSECURE" select Insecure, as upstream does. ok is false
// for empty or unknown values.
func ParseProfile(s string) (p Profile, ok bool) {
	switch strings.ToUpper(s) {
	case "SANDBOX":
		return Sandbox, true
	case "ALLOWLIST":
		return Allowlist, true
	case "INTERNET":
		return Internet, true
	case "INTERNET_WITH_DOTSVG":
		return InternetWithDotSVG, true
	case "LEGACY":
		return Legacy, true
	case "INSECURE", "UNSECURE":
		return Insecure, true
	}
	return Legacy, false
}

// Environment names read by upstream, as system properties or variables.
const (
	EnvProfile        = "PLANTUML_SECURITY_PROFILE"
	EnvIncludePath    = "plantuml.include.path"
	EnvAllowlistPath  = "plantuml.allowlist.path"
	EnvAllowlistURL   = "plantuml.allowlist.url"
	EnvSecurityFolder = "plantuml.security.credentials.path"
)

// Getenv looks name up like upstream's SecurityUtils.getenv: as is, then
// with dots replaced by underscores in upper case.
func Getenv(lookup func(string) string, name string) string {
	if v := lookup(name); v != "" {
		return v
	}
	return lookup(strings.ToUpper(strings.ReplaceAll(name, ".", "_")))
}

// Config is the security setting of one render.
type Config struct {
	Profile Profile
	// IncludePaths and AllowlistPaths are directories whose files are always
	// readable (plantuml.include.path and plantuml.allowlist.path).
	IncludePaths, AllowlistPaths []string
	// AllowlistURLs are URL prefixes that are always readable.
	AllowlistURLs []string
	// SecurityFolder holds credentials and is never readable.
	SecurityFolder string
	// LookupIP resolves host names for the internal address check; nil uses
	// the system resolver.
	LookupIP func(host string) ([]net.IP, error)
	// HostFiles marks files that come from a file system the caller
	// supplied (Access.FS), such as the files a web page passes to the
	// WASM build: they are readable in every profile.
	HostFiles bool
}

// ConfigFromEnv builds the configuration from the environment, using
// def when no profile is set.
func ConfigFromEnv(lookup func(string) string, def Profile) Config {
	c := Config{Profile: def}
	if p, ok := ParseProfile(Getenv(lookup, EnvProfile)); ok {
		c.Profile = p
	}
	c.IncludePaths = directories(Getenv(lookup, EnvIncludePath))
	c.AllowlistPaths = directories(Getenv(lookup, EnvAllowlistPath))
	if s := Getenv(lookup, EnvAllowlistURL); s != "" {
		c.AllowlistURLs = strings.Split(s, ";")
	}
	if dirs := directories(Getenv(lookup, EnvSecurityFolder)); len(dirs) > 0 {
		c.SecurityFolder = dirs[0]
	}
	return c
}

// directories splits a path list and keeps the existing directories, made
// absolute and free of symbolic links.
func directories(list string) []string {
	list = strings.TrimSuffix(strings.TrimPrefix(list, `"`), `"`)
	var out []string
	for _, p := range filepath.SplitList(list) {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if real, err := filepath.EvalSymlinks(abs); err == nil {
			abs = real
		}
		if st, err := os.Stat(abs); err == nil && st.IsDir() {
			out = append(out, abs)
		}
	}
	return out
}

// cleanPath is upstream's getCleanPathSecure for an absolute path.
func cleanPath(p string) string {
	p = strings.ReplaceAll(p, "\x00", "")
	return strings.ReplaceAll(p, `\\`, "/")
}

// FileOK reports whether the file at the absolute path absPath may be read.
func (c *Config) FileOK(absPath string) bool {
	if c.HostFiles {
		return true
	}
	if c.Profile == Sandbox {
		return false
	}
	path := cleanPath(absPath)
	if c.SecurityFolder != "" && strings.HasPrefix(filepath.Clean(path), filepath.Clean(c.SecurityFolder)) {
		return false
	}
	for _, list := range [][]string{c.IncludePaths, c.AllowlistPaths} {
		for _, dir := range list {
			if strings.HasPrefix(path, cleanPath(dir)) {
				return true
			}
		}
	}
	switch c.Profile {
	case Internet, InternetWithDotSVG, Allowlist:
		return false
	case Insecure:
		return true
	}
	for _, denied := range []string{"/etc/", "/dev/", "/boot/", "/proc/", "/sys/", "//"} {
		if strings.HasPrefix(path, denied) {
			return false
		}
	}
	return true
}

var userInfo = regexp.MustCompile(`^(https?://)([-_0-9a-zA-Z]+@)([^@]*)$`)

// cleanURL is upstream's SURL.cleanPath: user information removed, lower
// case, default ports removed.
func cleanURL(s string) string {
	s = userInfo.ReplaceAllString(s, "$1$3")
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ":80/", "")
	return strings.ReplaceAll(s, ":443/", "")
}

// URLOK reports whether the URL may be read.
func (c *Config) URLOK(raw string) bool {
	if c.Profile == Sandbox {
		return false
	}
	full := cleanURL(raw)
	if !strings.Contains(full, "@") {
		for _, allow := range c.AllowlistURLs {
			if strings.HasPrefix(full, cleanURL(allow)) {
				return true
			}
		}
	}
	switch c.Profile {
	case Legacy:
		return !c.urlForbidden(full)
	case Insecure:
		return true
	case Internet, InternetWithDotSVG:
		if c.urlForbidden(full) {
			return false
		}
		u, err := url.Parse(raw)
		if err != nil {
			return false
		}
		port := u.Port()
		return port == "" || port == "80" || port == "443"
	}
	return false
}

var (
	ipHostURL  = regexp.MustCompile(`^https?://[-#.0-9:\[\]+]+/.*`)
	noDotPath  = regexp.MustCompile(`^https?://[^.]+/.*`)
	noDotURL   = regexp.MustCompile(`^https?://[^.]+$`)
	ipLikeHost = regexp.MustCompile(`^[-#.0-9:\[\]+]+$`)
)

// urlForbidden is upstream's URLCheck.isURLforbidden.
func (c *Config) urlForbidden(full string) bool {
	switch {
	case strings.Contains(full, "@"),
		!strings.HasPrefix(full, "https://") && !strings.HasPrefix(full, "http://"),
		ipHostURL.MatchString(full), noDotPath.MatchString(full), noDotURL.MatchString(full):
		return true
	}
	u, err := url.Parse(full)
	if err != nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return true
	}
	host := u.Hostname()
	if host == "" || !strings.Contains(host, ".") {
		return true
	}
	if c.Profile != Insecure {
		if ipLikeHost.MatchString(host) {
			return true
		}
		lookup := c.LookupIP
		if lookup == nil {
			lookup = net.LookupIP
		}
		ips, err := lookup(host)
		if err != nil || len(ips) == 0 {
			return true
		}
		if ip := ips[0]; ip.IsUnspecified() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() {
			return true
		}
	}
	if decoded, err := url.QueryUnescape(host); err != nil || decoded != host {
		return true
	}
	return false
}
