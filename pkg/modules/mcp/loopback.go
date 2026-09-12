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
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

var (
	ErrToolNotFound = errors.New("mcp: tool not found")
	ErrScopeDenied  = errors.New("mcp: tool not authorized for this token")
	ErrNoCaller     = errors.New("mcp: no caller request in context")
)

type apiError struct {
	status int
	text   string
}

func (e *apiError) Error() string { return e.text }
func callTool(ctx context.Context, name string, rawArgs json.RawMessage) (any, error) {
	t, ok := findTool(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}
	if !t.authorized(TokenFromContext(ctx)) {
		return nil, fmt.Errorf("%w: %s", ErrScopeDenied, name)
	}
	caller := CallerFromContext(ctx)
	if caller == nil {
		return nil, ErrNoCaller
	}
	args, err := decodeArgs(t.spec, rawArgs)
	if err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", name, err)
	}
	req, err := t.newRequest(ctx, caller, args)
	if err != nil {
		return nil, fmt.Errorf("mcp: invalid arguments for %s: %w", name, err)
	}
	rec := httptest.NewRecorder()
	currentAPI().Adapter().ServeHTTP(rec, req)
	return parseResponse(rec)
}
func (t *tool) newRequest(ctx context.Context, caller *http.Request, args map[string]json.RawMessage) (*http.Request, error) {
	path := t.op.Path
	query := url.Values{}
	body := map[string]json.RawMessage{}
	for name, raw := range args {
		p, isParam := t.spec.params[name]
		switch {
		case !isParam:
			body[name] = raw
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
	if p, ok := t.spec.params["format"]; ok && p.In == "query" && !query.Has("format") {
		query.Set("format", "markdown")
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
	req.Header.Set("Authorization", caller.Header.Get("Authorization"))
	req.Header.Set("Accept", "application/json")
	if reader != nil {
		req.Header.Set("Content-Type", t.contentType)
	}
	// Preserve the client identity used by the rate limiter and access logs.
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip", "Accept-Language", "User-Agent"} {
		if v := caller.Header.Get(h); v != "" {
			req.Header.Set(h, v)
		}
	}
	req.RemoteAddr = caller.RemoteAddr
	return req, nil
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
	if rec.Code >= 300 {
		return nil, &apiError{status: rec.Code, text: errorText(rec)}
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

func errorText(rec *httptest.ResponseRecorder) string {
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
		return truncateRunes(fmt.Sprintf("%d: %s", rec.Code, strings.TrimSpace(string(body))), maxErrorTextRunes)
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
	return truncateRunes(b.String(), maxErrorTextRunes)
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
