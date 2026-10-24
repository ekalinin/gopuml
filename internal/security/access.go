// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/security/SecurityProfile.java
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/security/SURL.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package security

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// ErrDenied reports a file or URL the security profile does not allow.
var ErrDenied = errors.New("access denied by the security profile")

// Timeout is how long the profile waits for a URL, as upstream.
func (p Profile) Timeout() time.Duration {
	switch p {
	case Sandbox:
		return time.Second
	case Internet, InternetWithDotSVG:
		return 10 * time.Second
	case Legacy:
		return time.Minute
	}
	return 5 * time.Minute
}

// Access reads files and URLs for one render, applying the configuration.
type Access struct {
	Config Config
	// FS holds the files; paths are absolute slash paths inside it. Nil
	// reads from the operating system.
	FS fs.FS
	// HTTP fetches URLs; nil uses http.DefaultClient.
	HTTP *http.Client
}

// ReadFile reads the file at an absolute path.
func (a *Access) ReadFile(absPath string) ([]byte, error) {
	if !a.Config.FileOK(absPath) {
		return nil, fmt.Errorf("%w: %s", ErrDenied, absPath)
	}
	if a.FS == nil {
		return os.ReadFile(absPath)
	}
	return fs.ReadFile(a.FS, strings.TrimPrefix(path.Clean("/"+absPath), "/"))
}

// ReadURL fetches a URL within the profile's timeout.
func (a *Access) ReadURL(ctx context.Context, url string) ([]byte, error) {
	if !a.Config.URLOK(url) {
		return nil, fmt.Errorf("%w: %s", ErrDenied, url)
	}
	ctx, cancel := context.WithTimeout(ctx, a.Config.Profile.Timeout())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := a.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
