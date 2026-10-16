// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/PicoWebServer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliOptions.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

// readHeaderTimeout limits the time a client takes to send the request
// line and headers.
const readHeaderTimeout = 10 * time.Second

// ListenAddr returns the listen address for the values of the CLI flag
// --http-server[:port[:address]], the parts after the colons, as
// CliOptions.getPicowebPort and getPicowebBindAddress read them. Without a
// port the port is DefaultPort; without an address the server listens on
// every interface. Port 0 selects a free port.
func ListenAddr(values ...string) (string, error) {
	port := strconv.Itoa(DefaultPort)
	if len(values) > 0 {
		n, err := strconv.Atoi(values[0])
		if err != nil || n < 0 || n > 65535 {
			return "", fmt.Errorf("server: invalid port %q", values[0])
		}
		port = strconv.Itoa(n)
	}
	host := ""
	if len(values) > 1 {
		host = values[1]
	}
	return net.JoinHostPort(host, port), nil
}

// ListenAndServe listens on addr and serves h until ctx is done, like
// PicoWebServer.startServer. An empty addr means ListenAddr(), port 8080
// on every interface. errLog receives the errors of the HTTP server itself;
// nil discards them.
func ListenAndServe(ctx context.Context, addr string, h http.Handler, errLog *slog.Logger) error {
	if addr == "" {
		addr, _ = ListenAddr()
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return Serve(ctx, ln, h, errLog)
}

// Serve serves h on ln until ctx is done, then stops accepting
// connections, waits for the active requests and returns nil. It closes
// ln. errLog receives the errors of the HTTP server itself; nil discards
// them.
func Serve(ctx context.Context, ln net.Listener, h http.Handler, errLog *slog.Logger) error {
	if errLog == nil {
		errLog = slog.New(slog.DiscardHandler)
	}
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ErrorLog:          slog.NewLogLogger(errLog.Handler(), slog.LevelError),
	}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()
	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	// Requests end within the handler's time limit.
	err := srv.Shutdown(context.Background())
	if serveErr := <-served; !errors.Is(serveErr, http.ErrServerClosed) {
		err = errors.Join(err, serveErr)
	}
	return err
}
