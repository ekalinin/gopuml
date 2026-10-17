// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/PicoWebServer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/BadRequest400.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/FileFormat.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ekalinin/gopuml"
	"github.com/ekalinin/gopuml/internal/code"
	"github.com/ekalinin/gopuml/internal/security"
)

// Defaults of Config and of the listen address.
const (
	// DefaultPort is the port of --http-server without a port.
	DefaultPort = 8080
	// DefaultMaxBodySize limits request bodies and decoded sources.
	DefaultMaxBodySize int64 = 1 << 20
	// DefaultTimeout limits the time spent on one request.
	DefaultTimeout = 60 * time.Second
	// DefaultSecurityProfile is used when neither Config.SecurityProfile nor
	// PLANTUML_SECURITY_PROFILE is set.
	DefaultSecurityProfile = "INTERNET"
	// UpstreamVersion is the PlantUML version this server follows. It
	// appears in the Server header and in /serverinfo, as upstream's
	// Version.versionString() does.
	UpstreamVersion = "1.2026.8"
)

// RenderFunc renders a source; it has the signature of gopuml.Render.
type RenderFunc func(ctx context.Context, source string, opts *gopuml.Options) (*gopuml.Result, error)

// InspectFunc lists the blocks of a source; it has the signature of
// gopuml.Inspect.
type InspectFunc func(ctx context.Context, source string, opts *gopuml.Options) ([]*gopuml.Block, error)

// Config configures a Handler. The zero value is valid.
type Config struct {
	// Render renders diagrams; nil means gopuml.Render.
	Render RenderFunc
	// Inspect checks diagrams for GET /check; nil means gopuml.Inspect.
	Inspect InspectFunc
	// SecurityProfile is passed to every render. Empty means the
	// PLANTUML_SECURITY_PROFILE environment variable, or INTERNET when it is
	// unset or unknown.
	SecurityProfile string
	// MaxBodySize limits request bodies and decoded sources, in bytes.
	// Larger requests get status 413. Zero means DefaultMaxBodySize.
	MaxBodySize int64
	// Timeout limits each request. Slower requests get status 503 and their
	// render context is cancelled, which also stops dot. Zero means
	// DefaultTimeout.
	Timeout time.Duration
	// Language writes the body of GET /language. Nil means the built-in
	// description, see WriteLanguage.
	Language func(w io.Writer) error
	// Logger receives rendering failures. Nil discards them.
	Logger *slog.Logger
}

// Handler serves the routes of upstream's built-in server (--http-server)
// and of plantuml-server.
type Handler struct {
	render   RenderFunc
	inspect  InspectFunc
	profile  string
	maxBody  int64
	timeout  time.Duration
	language func(io.Writer) error
	log      *slog.Logger
}

// New returns a handler for cfg. It fails when cfg.SecurityProfile is not
// a known profile name.
func New(cfg Config) (*Handler, error) {
	h := &Handler{
		render:   cfg.Render,
		inspect:  cfg.Inspect,
		profile:  cfg.SecurityProfile,
		maxBody:  cfg.MaxBodySize,
		timeout:  cfg.Timeout,
		language: cfg.Language,
		log:      cfg.Logger,
	}
	if h.render == nil {
		h.render = gopuml.Render
	}
	if h.inspect == nil {
		h.inspect = gopuml.Inspect
	}
	if h.profile == "" {
		h.profile = DefaultSecurityProfile
		if p, ok := security.ParseProfile(security.Getenv(os.Getenv, security.EnvProfile)); ok {
			h.profile = p.String()
		}
	} else if p, ok := security.ParseProfile(h.profile); ok {
		h.profile = p.String()
	} else {
		return nil, fmt.Errorf("server: unknown security profile %q", cfg.SecurityProfile)
	}
	if h.maxBody <= 0 {
		h.maxBody = DefaultMaxBodySize
	}
	if h.timeout <= 0 {
		h.timeout = DefaultTimeout
	}
	if h.language == nil {
		h.language = WriteLanguage
	}
	if h.log == nil {
		h.log = slog.New(slog.DiscardHandler)
	}
	return h, nil
}

// SecurityProfile returns the profile passed to every render.
func (h *Handler) SecurityProfile() string { return h.profile }

// mimeTypes are upstream's FileFormat.getMimeType values.
var mimeTypes = map[gopuml.Format]string{
	gopuml.PNG:  "image/png",
	gopuml.SVG:  "image/svg+xml",
	gopuml.TXT:  "text/plain",
	gopuml.UTXT: "text/plain;charset=UTF-8",
}

// Header values shared by several routes.
const (
	serverHeader = "PlantUML PicoWebServer " + UpstreamVersion
	// Upstream's built-in server sends every rendering and info response
	// with this Cache-Control.
	cacheNoCache = "no-cache"
	// plantuml-server's Cache-Control for GET rendering, also on error
	// images (recorded by the dark-mode-url and
	// syntax-error-by-url-plantuml-server references).
	cachePublic = "public, max-age=432000"
	// plantuml-server's type for text answers (recorded by the
	// encode-through-coder reference).
	textUTF8 = "text/plain;charset=utf-8"
	// Upstream's type for BadRequest400 and other plain-text errors.
	textPlain = "text/plain"
	// defaultLocation is where upstream's built-in server redirects every
	// request it does not handle.
	defaultLocation = "/plantuml/png/oqbDJyrBuGh8ISmh2VNrKGZ8JCuFJqqAJYqgIotY0aefG5G00000"
	// noValidStart is upstream's message for a POST /render source without
	// a block.
	noValidStart = "No valid @start/@end found, please check the version"
)

// style describes how a rendering route answers, following the upstream
// server that defines the route.
type style struct {
	// cache is the Cache-Control value.
	cache string
	// picoweb adds the Server header of upstream's built-in server.
	picoweb bool
	// errorStatus is the status of an error image.
	errorStatus int
	// redirectNoDiagram answers a source without a block with upstream's
	// redirect instead of status 400.
	redirectNoDiagram bool
}

var (
	// GET routes of upstream's built-in server (PicoWebServer.handleGET).
	builtinGET = style{cache: cacheNoCache, picoweb: true, errorStatus: http.StatusBadRequest, redirectNoDiagram: true}
	// POST /render of upstream's built-in server: handleRenderRequest
	// always answers 200, error images included.
	builtinRender = style{cache: cacheNoCache, picoweb: true, errorStatus: http.StatusOK}
	// GET routes that only plantuml-server defines.
	serverGET = style{cache: cachePublic, errorStatus: http.StatusBadRequest, redirectNoDiagram: true}
	// Raw-body POST routes of plantuml-server. No reference records their
	// Cache-Control; a POST answer is not cached by URL, so no-cache is used.
	serverPOST = style{cache: cacheNoCache, errorStatus: http.StatusBadRequest}
)

// getRoute is a GET rendering route: the path prefix before the encoded
// text, with and without the /plantuml prefix.
type getRoute struct {
	prefix string
	format gopuml.Format
	dark   bool
	style  style
}

var getRoutes = []getRoute{
	// Upstream's built-in server.
	{"/png/", gopuml.PNG, false, builtinGET},
	{"/svg/", gopuml.SVG, false, builtinGET},
	{"/txt/", gopuml.TXT, false, builtinGET},
	{"/utxt/", gopuml.UTXT, false, builtinGET},
	// plantuml-server.
	{"/img/", gopuml.PNG, false, serverGET},
	{"/dpng/", gopuml.PNG, true, serverGET},
	{"/dimg/", gopuml.PNG, true, serverGET},
	{"/dsvg/", gopuml.SVG, true, serverGET},
}

// Raw-body POST routes of plantuml-server.
var postRoutes = map[string]gopuml.Format{
	"/png": gopuml.PNG,
	"/svg": gopuml.SVG,
	"/txt": gopuml.TXT,
}

// ServeHTTP dispatches a request like PicoWebServer.run, extended with the
// plantuml-server routes. Every route also answers with the /plantuml
// prefix. Requests that match no route get upstream's redirect.
//
// The time limit starts here and covers reading the body and rendering.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeoutCause(r.Context(), h.timeout, errTimeout)
	defer cancel()
	r = r.WithContext(ctx)
	p := r.URL.Path
	if p == "/plantuml" || strings.HasPrefix(p, "/plantuml/") {
		p = p[len("/plantuml"):]
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		if h.serveGET(w, r, p) {
			return
		}
	case http.MethodPost:
		if h.servePOST(w, r, p) {
			return
		}
	}
	w.Header().Set("Location", defaultLocation)
	w.WriteHeader(http.StatusFound)
}

func (h *Handler) serveGET(w http.ResponseWriter, r *http.Request, p string) bool {
	for _, rt := range getRoutes {
		if strings.HasPrefix(p, rt.prefix) {
			h.serveEncoded(w, r, p, rt)
			return true
		}
	}
	switch {
	case strings.HasPrefix(p, "/serverinfo"):
		h.serveInfo(w)
	case strings.HasPrefix(p, "/language"):
		h.serveLanguage(w)
	case strings.HasPrefix(p, "/check/"):
		h.serveCheck(w, r, lastSegment(p))
	case strings.HasPrefix(p, "/coder/"):
		h.serveDecode(w, lastSegment(p))
	default:
		return false
	}
	return true
}

func (h *Handler) servePOST(w http.ResponseWriter, r *http.Request, p string) bool {
	p = strings.TrimSuffix(p, "/")
	if format, ok := postRoutes[p]; ok {
		h.serveRaw(w, r, format)
		return true
	}
	switch p {
	case "/render":
		h.serveRenderRequest(w, r)
	case "/coder":
		h.serveEncode(w, r)
	default:
		return false
	}
	return true
}

// lastSegment is the text after the last '/', where upstream's handleGET
// reads the encoded diagram.
func lastSegment(p string) string { return p[strings.LastIndexByte(p, '/')+1:] }

// serveEncoded renders the encoded text of a GET route. Text that is not
// encoded diagram source becomes upstream's diagram about a bad URL
// (TranscoderSmartProtected), rendered with status 200.
func (h *Handler) serveEncoded(w http.ResponseWriter, r *http.Request, p string, rt getRoute) {
	source, err := code.DecodeProtected(lastSegment(p))
	if err != nil {
		// Only the ~h, ~g and ~zip~ forms fail here. Upstream answers with a
		// Java stack trace and status 500; this is a client error.
		writeText(w, http.StatusBadRequest, textPlain, err.Error())
		return
	}
	if int64(len(source)) > h.maxBody {
		writeText(w, http.StatusRequestEntityTooLarge, textPlain, "Decoded source too large")
		return
	}
	h.serveRender(w, r, rt.style, source, gopuml.Options{Format: rt.format, DarkMode: rt.dark})
}

// serveRaw renders the request body of a raw POST route.
func (h *Handler) serveRaw(w http.ResponseWriter, r *http.Request, format gopuml.Format) {
	body, ok := h.readBody(w, r)
	if !ok {
		return
	}
	h.serveRender(w, r, serverPOST, wrapSource(string(body)), gopuml.Options{Format: format})
}

// serveRenderRequest is PicoWebServer.handleRenderRequest: a JSON body
// with the source and CLI options.
func (h *Handler) serveRenderRequest(w http.ResponseWriter, r *http.Request) {
	body, ok := h.readBody(w, r)
	if !ok {
		return
	}
	req, err := parseRenderRequest(body)
	if err != nil {
		writeText(w, http.StatusBadRequest, textPlain, err.Error())
		return
	}
	opts, err := parseOptions(req.options)
	if err != nil {
		writeText(w, http.StatusBadRequest, textPlain, err.Error())
		return
	}
	h.serveRender(w, r, builtinRender, wrapSource(req.source), opts)
}

// wrapSource adds @startuml and @enduml lines around a source that does
// not start with @start, as handleRenderRequest does.
func wrapSource(s string) string {
	if strings.HasPrefix(s, "@start") {
		return s
	}
	return "@startuml\n" + s + "\n@enduml"
}

// readBody reads a request body within the size and time limits. On
// failure it answers the request and returns false.
func (h *Handler) readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	// The time limit also covers the upload. The read deadline is cleared
	// once the body is read, so that it cannot break the connection while
	// the diagram is rendered.
	if deadline, ok := r.Context().Deadline(); ok {
		rc := http.NewResponseController(w)
		if rc.SetReadDeadline(deadline) == nil {
			defer rc.SetReadDeadline(time.Time{})
		}
	}
	body := http.MaxBytesReader(w, r.Body, h.maxBody)
	data, err := io.ReadAll(body)
	if err == nil {
		return data, true
	}
	// Do not let net/http drain the rest of a body that failed, and do not
	// reuse the connection.
	body.Close()
	w.Header().Set("Connection", "close")
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		writeText(w, http.StatusRequestEntityTooLarge, textPlain, "Request body too large")
	case errors.Is(err, os.ErrDeadlineExceeded):
		writeText(w, http.StatusServiceUnavailable, textPlain, "Request time limit exceeded")
	default:
		writeText(w, http.StatusBadRequest, textPlain, "Error reading request body: "+err.Error())
	}
	return nil, false
}

// errTimeout is the cancellation cause of a request that ran out of time.
var errTimeout = errors.New("server: request time limit exceeded")

// errPanic reports a renderer that panicked.
var errPanic = errors.New("server: renderer panicked")

// call runs f with ctx in its own goroutine, so that a renderer that does
// not watch ctx cannot hold the answer past the time limit. A panic in f
// becomes an error matching errPanic.
func call[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	type result struct {
		v   T
		err error
	}
	done := make(chan result, 1)
	go func() {
		var res result
		defer func() {
			if v := recover(); v != nil {
				res = result{err: fmt.Errorf("%w: %v", errPanic, v)}
			}
			done <- res
		}()
		res.v, res.err = f(ctx)
	}()
	select {
	case res := <-done:
		return res.v, res.err
	case <-ctx.Done():
		var zero T
		return zero, context.Cause(ctx)
	}
}

// serveRender renders source and answers like PicoWebServer.sendDiagram.
func (h *Handler) serveRender(w http.ResponseWriter, r *http.Request, st style, source string, opts gopuml.Options) {
	opts.SecurityProfile = h.profile
	ctx := r.Context()
	res, err := call(ctx, func(ctx context.Context) (*gopuml.Result, error) {
		return h.render(ctx, source, &opts)
	})
	if err == nil && res == nil {
		err = errors.New("server: renderer returned no result")
	}
	var syntax *gopuml.SyntaxError
	switch {
	case err == nil:
		h.writeDiagram(w, st, http.StatusOK, opts.Format, res, nil)
	case context.Cause(ctx) == errTimeout:
		h.log.Warn("rendering timed out", "path", r.URL.Path, "timeout", h.timeout)
		writeText(w, http.StatusServiceUnavailable, textPlain, "Rendering time limit exceeded")
	case ctx.Err() != nil:
		// The client went away; nobody reads the answer.
	case errors.As(err, &syntax), errors.Is(err, gopuml.ErrUnsupportedDiagram), errors.Is(err, gopuml.ErrUnsupportedFormat):
		h.writeDiagram(w, st, st.errorStatus, opts.Format, res, err)
	case errors.Is(err, gopuml.ErrNoDiagram) && st.redirectNoDiagram:
		// handleGET returns false without a block, and run() redirects.
		w.Header().Set("Location", defaultLocation)
		w.WriteHeader(http.StatusFound)
	case errors.Is(err, gopuml.ErrNoDiagram):
		writeText(w, http.StatusBadRequest, textPlain, noValidStart)
	default:
		// A rendering failure, such as a dot crash. Upstream draws a crash
		// image and answers 503.
		h.log.Error("rendering failed", "path", r.URL.Path, "error", err)
		if res != nil && len(res.Data) > 0 {
			h.writeDiagram(w, st, http.StatusServiceUnavailable, opts.Format, res, nil)
			return
		}
		writeText(w, http.StatusServiceUnavailable, textPlain, err.Error())
	}
}

// writeDiagram is PicoWebServer.sendDiagram. err is the diagram error of
// an error image, or nil.
func (h *Handler) writeDiagram(w http.ResponseWriter, st style, status int, format gopuml.Format, res *gopuml.Result, err error) {
	hdr := w.Header()
	hdr.Set("Cache-Control", st.cache)
	if st.picoweb {
		hdr.Set("Server", serverHeader)
	}
	hdr.Set("Access-Control-Allow-Origin", "*")
	hdr.Set("Content-Type", mimeTypes[format])
	var data []byte
	description := "(Error)" // PSystemError's description
	if res != nil {
		data = res.Data
		description = res.Description
		setRaw(hdr, "X-PlantUML-Diagram-Width", strconv.Itoa(res.Width))
		setRaw(hdr, "X-PlantUML-Diagram-Height", strconv.Itoa(res.Height))
	}
	hdr.Set("Content-Length", strconv.Itoa(len(data)))
	setRaw(hdr, "X-PlantUML-Diagram-Description", description)
	var syntax *gopuml.SyntaxError
	var unsupported *gopuml.UnsupportedError
	switch {
	case errors.As(err, &syntax):
		// The line of upstream's built-in server: 1 + the 0-based position.
		setRaw(hdr, "X-PlantUML-Diagram-Error", syntax.Message)
		setRaw(hdr, "X-PlantUML-Diagram-Error-Line", strconv.Itoa(syntax.Line))
	case errors.As(err, &unsupported):
		setRaw(hdr, "X-PlantUML-Diagram-Error", unsupported.Message)
		setRaw(hdr, "X-PlantUML-Diagram-Error-Line", strconv.Itoa(unsupported.Line))
	case err != nil:
		setRaw(hdr, "X-PlantUML-Diagram-Error", errorMessage(err))
	}
	w.WriteHeader(status)
	w.Write(data)
}

// errorMessage strips the package prefix from a diagram error.
func errorMessage(err error) string {
	msg := err.Error()
	for _, e := range []error{gopuml.ErrUnsupportedDiagram, gopuml.ErrUnsupportedFormat} {
		if errors.Is(err, e) {
			msg = strings.TrimPrefix(msg, e.Error()+": ")
		}
	}
	return msg
}

// setRaw sets a header with upstream's spelling, such as
// X-PlantUML-Diagram-Width, instead of the canonical X-Plantuml-....
// Header names are case-insensitive; the spelling only helps clients that
// compare them exactly.
func setRaw(h http.Header, key, value string) {
	h.Del(key)
	h[key] = []string{value}
}

// writeText answers with a plain-text body, like PicoWebServer.sendError
// for BadRequest400.
func writeText(w http.ResponseWriter, status int, contentType, body string) {
	hdr := w.Header()
	hdr.Set("Content-Type", contentType)
	hdr.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	io.WriteString(w, body)
}
