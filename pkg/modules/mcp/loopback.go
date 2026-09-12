// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

var (
	errToolNotFound = errors.New("mcp: tool not found")
	errScopeDenied  = errors.New("mcp: tool not authorized for this token")
	errNoCaller     = errors.New("mcp: no caller request in context")
)

func callTool(ctx context.Context, name string, rawArgs json.RawMessage) (any, error) {
	t, ok := findTool(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errToolNotFound, name)
	}
	// Read the caller before dispatching: the loopback re-enters the group middleware, which stashes its own echo context.
	ec := humabridge.EchoContextFrom(ctx)
	if ec == nil {
		return nil, errNoCaller
	}
	token := tokenFrom(ec)
	if !t.authorized(token) {
		return nil, fmt.Errorf("%w: %s", errScopeDenied, name)
	}
	args, err := decodeArgs(t.spec, rawArgs)
	if err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", name, err)
	}
	req, err := t.newRequest(ctx, ec, args)
	if err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", name, err)
	}
	rec := httptest.NewRecorder()
	currentAPI().Adapter().ServeHTTP(rec, req)
	recordTokenUsage(ctx, token)
	return parseResponse(rec)
}

// The loopback runs as an internal dispatch, for which the auth middleware skips the usage event.
func recordTokenUsage(ctx context.Context, token *models.APIToken) {
	if token == nil || !config.AuditEnabled.GetBool() {
		return
	}
	err := events.DispatchWithContext(ctx, &models.APITokenUsedEvent{
		TokenID: token.ID,
		OwnerID: token.OwnerID,
	})
	if err != nil {
		log.Errorf("[mcp] could not dispatch api token used event: %s", err)
	}
}
func (t *tool) newRequest(ctx context.Context, ec *echo.Context, args map[string]json.RawMessage) (*http.Request, error) {
	caller := ec.Request()
	path := t.op.Path
	query := url.Values{}
	body := map[string]json.RawMessage{}
	for name, raw := range args {
		p, isParam := t.spec.params[name]
		switch {
		case !isParam:
			body[name] = raw
		// A null body field clears it through merge-patch; a null parameter is simply unset.
		case isJSONNull(raw):
		case p.In == "path":
			v, err := scalarString(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid value for %q: %w", name, err)
			}
			path = strings.ReplaceAll(path, "{"+name+"}", url.PathEscape(v))
		default:
			values, err := queryValues(p, raw)
			if err != nil {
				return nil, fmt.Errorf("invalid value for %q: %w", name, err)
			}
			for _, v := range values {
				query.Add(name, v)
			}
		}
	}
	var reader io.Reader
	if t.spec.hasBody {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	target := path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, t.op.Method, target, reader)
	if err != nil {
		return nil, err
	}
	// The transport authorised one of possibly several Authorization values; the loopback must not authenticate as a different one.
	if auth, ok := models.APITokenAuthorization(caller.Header); ok {
		req.Header.Set("Authorization", auth)
	}
	req.Header.Set("Accept", "application/json")
	if reader != nil {
		req.Header.Set("Content-Type", t.contentType)
	}
	// Keep the public origin so generated links and the rate limiter see the real client.
	req.Host = caller.Host
	req.TLS = caller.TLS
	// AutoPatch's inner request carries a path-only URL, so huma falls back to this header for the schema host.
	if forwardedHost := caller.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		req.Header.Set("X-Forwarded-Host", forwardedHost)
	} else {
		req.Header.Set("X-Forwarded-Host", caller.Host)
	}
	for _, h := range []string{
		"X-Forwarded-For",
		"X-Real-Ip",
		"X-Request-Id",
		"User-Agent",
	} {
		if vs := caller.Header.Values(h); len(vs) > 0 {
			req.Header[http.CanonicalHeaderKey(h)] = slices.Clone(vs)
		}
	}
	if id := requestID(ec); id != "" {
		req.Header.Set(echo.HeaderXRequestID, id)
	}
	req.RemoteAddr = caller.RemoteAddr
	return req, nil
}

// echo's RequestID middleware writes the id it generated to the response only.
func requestID(ec *echo.Context) string {
	if id := ec.Response().Header().Get(echo.HeaderXRequestID); id != "" {
		return id
	}
	return ec.Request().Header.Get(echo.HeaderXRequestID)
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
func scalarString(raw json.RawMessage) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	switch x := v.(type) {
	case string:
		return x, nil
	case json.Number:
		return x.String(), nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	}
	return "", errors.New("must be a string, number or boolean")
}
func queryValues(p *huma.Param, raw json.RawMessage) ([]string, error) {
	if !bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
		s, err := scalarString(raw)
		return []string{s}, err
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		s, err := scalarString(it)
		if err != nil {
			return nil, err
		}
		parts = append(parts, s)
	}
	if p.Explode != nil && *p.Explode {
		return parts, nil
	}
	return []string{strings.Join(parts, ",")}, nil
}
func parseResponse(rec *httptest.ResponseRecorder) (any, error) {
	// AutoPatch answers a patch that changes nothing with an empty 304.
	if rec.Code == http.StatusNotModified {
		return map[string]any{
			"ok":        true,
			"unchanged": true,
		}, nil
	}
	if rec.Code >= 300 {
		return nil, errors.New(errorText(rec))
	}
	if rec.Body.Len() == 0 {
		return map[string]any{"ok": true}, nil
	}
	dec := json.NewDecoder(rec.Body)
	dec.UseNumber()
	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("mcp: decode v2 response: %w", err)
	}
	return out, nil
}

const maxErrorTextRunes = 2000

const scopeHint = "the API token lacks a scope required by this call (check route and expand scopes)"

func errorText(rec *httptest.ResponseRecorder) string {
	text := truncateRunes(problemText(rec), maxErrorTextRunes)
	// The transport already authenticated the token, so a loopback 401 is a missing scope.
	if rec.Code == http.StatusUnauthorized {
		text += " — " + scopeHint
	}
	return text
}
func problemText(rec *httptest.ResponseRecorder) string {
	var problem struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Errors []struct {
			Location string `json:"location"`
			Message  string `json:"message"`
		} `json:"errors"`
	}
	body := rec.Body.Bytes()
	if err := json.Unmarshal(body, &problem); err != nil || problem.Title == "" {
		return fmt.Sprintf("%d: %s", rec.Code, strings.TrimSpace(string(body)))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d %s", rec.Code, problem.Title)
	if problem.Detail != "" {
		b.WriteString(": " + problem.Detail)
	}
	for i, e := range problem.Errors {
		if i == 0 {
			b.WriteString(" (")
		} else {
			b.WriteString("; ")
		}
		if e.Location != "" {
			b.WriteString(e.Location + ": ")
		}
		b.WriteString(e.Message)
		if i == len(problem.Errors)-1 {
			b.WriteString(")")
		}
	}
	return b.String()
}
func truncateRunes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "…"
}
