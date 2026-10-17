// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/picoweb/PicoWebServer.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/Pipe.java
// Provenance: plantuml-mit 1.2026.8 net/sourceforge/plantuml/StdrptPipe0.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ekalinin/gopuml"
	"github.com/ekalinin/gopuml/internal/code"
)

// infoBody is the JSON of PicoWebServer.handleInfo, with the line end its
// write() adds.
const infoBody = `{"version":"` + UpstreamVersion + `","PicoWebServer":true,"formats":["png","svg","txt"]}` + "\r\n"

// serveInfo is PicoWebServer.handleInfo (GET /serverinfo).
func (h *Handler) serveInfo(w http.ResponseWriter) {
	hdr := w.Header()
	hdr.Set("Cache-Control", cacheNoCache)
	hdr.Set("Server", serverHeader)
	writeText(w, http.StatusOK, "application/json", infoBody)
}

// serveLanguage is PicoWebServer.handleLanguage (GET /language).
func (h *Handler) serveLanguage(w http.ResponseWriter) {
	var buf bytes.Buffer
	if err := h.language(&buf); err != nil {
		h.log.Error("writing the language description failed", "error", err)
		writeText(w, http.StatusInternalServerError, textPlain, err.Error())
		return
	}
	hdr := w.Header()
	hdr.Set("Cache-Control", cacheNoCache)
	hdr.Set("Server", serverHeader)
	writeText(w, http.StatusOK, "text/text", buf.String())
}

// serveEncode answers POST /coder, a plantuml-server route: the body is
// diagram source, the answer its encoded text.
func (h *Handler) serveEncode(w http.ResponseWriter, r *http.Request) {
	body, ok := h.readBody(w, r)
	if !ok {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeText(w, http.StatusOK, textUTF8, code.Encode(string(body)))
}

// serveDecode answers GET /coder/<encoded>, a plantuml-server route: the
// decoded source, with @startuml and @enduml added when it has no @start
// line.
func (h *Handler) serveDecode(w http.ResponseWriter, encoded string) {
	source, err := code.Decode(encoded)
	if err != nil {
		writeText(w, http.StatusBadRequest, textPlain, err.Error())
		return
	}
	if int64(len(source)) > h.maxBody {
		writeText(w, http.StatusRequestEntityTooLarge, textPlain, "Decoded source too large")
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeText(w, http.StatusOK, textUTF8, source)
}

// serveCheck answers GET /check/<encoded>, a plantuml-server route, with
// the syntax check of the first diagram.
//
// No reference records plantuml-server's answer, so the text is upstream's
// own syntax check report, the one of "plantuml -syntax" (Pipe and
// StdrptPipe0):
//
//	SEQUENCE              ERROR
//	(2 participants)      2
//	                      Syntax Error? (Assumed diagram type: sequence)
//
// The type and description of a valid diagram, or "ERROR", the 0-based line
// and the message of a syntax error. A diagram of another kind prints
// "OTHER" and its description; an unsupported diagram prints nothing.
func (h *Handler) serveCheck(w http.ResponseWriter, r *http.Request, encoded string) {
	source, err := code.Decode(encoded)
	if err != nil {
		writeText(w, http.StatusBadRequest, textPlain, err.Error())
		return
	}
	if int64(len(source)) > h.maxBody {
		writeText(w, http.StatusRequestEntityTooLarge, textPlain, "Decoded source too large")
		return
	}
	ctx := r.Context()
	opts := gopuml.Options{SecurityProfile: h.profile}
	blocks, err := call(ctx, func(ctx context.Context) ([]*gopuml.Block, error) {
		return h.inspect(ctx, source, &opts)
	})
	switch {
	case err == nil:
	case context.Cause(ctx) == errTimeout:
		writeText(w, http.StatusServiceUnavailable, textPlain, "Rendering time limit exceeded")
		return
	case ctx.Err() != nil:
		return
	case errors.Is(err, gopuml.ErrNoDiagram):
		blocks = nil
	default:
		h.log.Error("syntax check failed", "path", r.URL.Path, "error", err)
		writeText(w, http.StatusServiceUnavailable, textPlain, err.Error())
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeText(w, http.StatusOK, textUTF8, checkReport(blocks))
}

// checkReport is Pipe.syntaxCheckDiagram for the first block.
func checkReport(blocks []*gopuml.Block) string {
	if len(blocks) == 0 || blocks[0] == nil {
		return ""
	}
	b := blocks[0]
	var syntax *gopuml.SyntaxError
	var unsupported *gopuml.UnsupportedError
	var sb strings.Builder
	switch {
	case b.Err == nil && b.Type != "":
		fmt.Fprintf(&sb, "%s\n%s\n", b.Type, b.Description)
	case b.Err == nil:
		fmt.Fprintf(&sb, "OTHER\n%s\n", b.Description)
	case errors.As(b.Err, &syntax):
		// StdrptPipe0.printInfo prints the 0-based position.
		sb.WriteString("ERROR\n" + strconv.Itoa(syntax.Line-1) + "\n" + syntax.Message + "\n")
	case errors.As(b.Err, &unsupported):
		sb.WriteString("ERROR\n" + strconv.Itoa(unsupported.Line-1) + "\n" + unsupported.Message + "\n")
	}
	return sb.String()
}
