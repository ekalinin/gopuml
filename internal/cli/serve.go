// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/PicoWebServer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/cli/CliOptions.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ekalinin/plantuml-in-go/internal/server"
)

// serve is --http-server[:port[:address]] (PicoWebServer.startServer): it
// prints the port and address to stderr and serves until interrupted.
func (r *run) serve() int {
	var values []string
	for _, v := range r.p.List(Picoweb) {
		values = append(values, v.S)
	}
	addr, err := server.ListenAddr(values...)
	if err != nil {
		fmt.Fprintln(r.Stderr, "plantuml:", err)
		return ExitUnsupported
	}
	h, err := server.New(server.Config{Logger: slog.New(slog.NewTextHandler(r.Stderr, nil))})
	if err != nil {
		fmt.Fprintln(r.Stderr, "plantuml:", err)
		return ExitUnsupported
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintln(r.Stderr, "plantuml:", err)
		return ExitUnsupported
	}
	bind := "0.0.0.0"
	if len(values) > 1 {
		bind = values[1]
	}
	fmt.Fprintf(r.Stderr, "webPort=%d\n", ln.Addr().(*net.TCPAddr).Port)
	fmt.Fprintf(r.Stderr, "webAddress=%s\n", bind)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Serve(ctx, ln, h, nil); err != nil {
		fmt.Fprintln(r.Stderr, "plantuml:", err)
		return ExitUnsupported
	}
	return ExitOK
}
