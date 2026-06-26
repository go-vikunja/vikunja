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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/veans/internal/output"
)

// Client is a thin JSON wrapper around the Vikunja REST API. It holds the
// server base URL and a bearer token (either a JWT from POST /login or an
// API token minted via POST /tokens). Every method in this package is a thin
// shim over Do.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// apiBasePath is the version prefix every request is mounted under. veans
// targets the Huma-backed /api/v2 exclusively — v1 is frozen and the bucket
// CRUD endpoints veans needs only exist on v2.
const apiBasePath = "/api/v2"

// contentTypeJSON / contentTypeMergePatch are the request body content types
// Do and DoMerge send. Merge-patch (RFC 7396) is how v2 does partial updates:
// only the fields present in the body are written, the rest are left intact.
const (
	contentTypeJSON       = "application/json"
	contentTypeMergePatch = "application/merge-patch+json"
)

// UserAgent is the value sent in the User-Agent header on every request.
// main sets this at startup with the linker-injected version + the
// runtime os/arch (e.g. "veans/0.3.1 (linux/amd64)"). Tests get the
// default "veans/dev". Vikunja admins see this in their access logs.
var UserAgent = "veans/dev"

// defaultHTTPTimeout is the timeout applied to the HTTP client returned by
// New. Callers that need a different value (e.g. the runtime loader honoring
// `http_timeout` from .veans.yml) can overwrite HTTPClient.Timeout after
// construction.
const defaultHTTPTimeout = 30 * time.Second

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// vikunjaError matches the RFC 9457 problem+json body /api/v2 returns
// (huma.ErrorModel augmented with Vikunja's numeric domain `code`). The
// human-readable message lives in `detail`; `title` is the status text
// fallback. `message` is v1's legacy field, kept only as a fallback so a
// stray legacy/proxy error body still yields a readable message instead of
// raw JSON. The HTTP status used for output.Code mapping comes from the
// response status line, not this body.
type vikunjaError struct {
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Do performs a single JSON request against /api/v2<path>. body, if non-nil,
// is JSON-marshalled. out, if non-nil, is JSON-unmarshalled. query is appended
// as URL-encoded params.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	return c.do(ctx, method, path, query, body, out, contentTypeJSON)
}

// DoMerge is like Do but sends the body as a JSON Merge Patch
// (application/merge-patch+json). Used for PATCH updates so only the fields
// present in `body` are written server-side — see UpdateTask.
func (c *Client) DoMerge(ctx context.Context, method, path string, query url.Values, body, out any) error {
	return c.do(ctx, method, path, query, body, out, contentTypeMergePatch)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any, contentType string) error {
	full := c.BaseURL + apiBasePath + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, full, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return output.Wrap(output.CodeUnknown, err, "%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return mapHTTPError(method, path, resp.StatusCode, respBody,
			parseRetryAfter(resp.Header.Get("Retry-After")))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode %s %s: %w", method, path, err)
		}
	}
	return nil
}

// Paginated mirrors the standard /api/v2 list envelope. Every v2 list
// operation returns this shape (v1 returned a bare array plus an
// x-pagination-total-pages header, which is gone). Single-object responses
// stay unwrapped.
type Paginated[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// doList GETs `path` and decodes the standard v2 list envelope, returning the
// items plus the server's total page count so a caller can page until
// page >= totalPages. Generic so each list endpoint reuses it without a
// per-type wrapper struct.
func doList[T any](ctx context.Context, c *Client, path string, query url.Values) (items []T, totalPages int, err error) {
	var env Paginated[T]
	if err := c.Do(ctx, "GET", path, query, nil, &env); err != nil {
		return nil, 0, err
	}
	return env.Items, env.TotalPages, nil
}

// doListAll pages through a v2 list endpoint, accumulating every item until
// page >= total_pages.
//
// Use it ONLY for endpoints whose model honours page/per_page — the
// server-paginated lists (tasks, projects, labels, comments, bots). For the
// endpoints whose ReadAll ignores pagination and returns every row in a single
// page (buckets, views), call doList instead: looping those re-fetches the full
// set on every page and duplicates it.
func doListAll[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	var all []T
	page := 1
	for {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", "50")
		batch, totalPages, err := doList[T](ctx, c, path, q)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if page >= totalPages {
			return all, nil
		}
		page++
	}
}

// DoRaw is the escape hatch used by `veans api`. It returns the raw response
// body, status, and the parsed Retry-After (if any). Auth + base URL handling
// matches Do. The caller is responsible for deciding whether to surface the
// body to stdout — non-2xx bodies should NOT be written there (the contract is
// "stdout is for the success payload; errors go through the envelope on
// stderr"); see commands/api.go for the canonical handling.
func (c *Client) DoRaw(ctx context.Context, method, path string, query url.Values, body []byte) (status int, respBody []byte, retryAfter time.Duration, err error) {
	full := c.BaseURL + apiBasePath + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	var br io.Reader
	if len(body) > 0 {
		br = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, full, br)
	if err != nil {
		return 0, nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err = io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	return resp.StatusCode, respBody, parseRetryAfter(resp.Header.Get("Retry-After")), err
}

// maxBodyBytes caps the size of any response body we'll read into memory.
// Vikunja JSON payloads are far smaller; the cap exists so a misbehaving
// proxy can't OOM the CLI by streaming an unbounded body.
const maxBodyBytes = 32 * 1024 * 1024 // 32 MiB

// parseRetryAfter parses an HTTP Retry-After header value. Supports both
// the delta-seconds form and the HTTP-date form; returns 0 on unparseable
// or empty input.
func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func mapHTTPError(method, path string, status int, body []byte, retryAfter time.Duration) error {
	var ve vikunjaError
	_ = json.Unmarshal(body, &ve)
	// v2's problem+json carries the human-readable text in `detail`; fall back
	// to `title`, then v1's legacy `message`, then the raw body, then the
	// status text.
	msg := strings.TrimSpace(ve.Detail)
	if msg == "" {
		msg = strings.TrimSpace(ve.Title)
	}
	if msg == "" {
		msg = strings.TrimSpace(ve.Message)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(status)
		}
	}
	// Truncate so an HTML error page (e.g. from a reverse proxy) doesn't
	// dump several KB into the agent's stderr envelope.
	if len(msg) > maxErrorMessageBytes {
		msg = msg[:maxErrorMessageBytes] + "…(truncated)"
	}

	var code output.Code
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		code = output.CodeAuth
	case status == http.StatusNotFound:
		code = output.CodeNotFound
	case status == http.StatusConflict:
		code = output.CodeConflict
	case status == http.StatusTooManyRequests:
		code = output.CodeRateLimited
	case status >= 400 && status < 500:
		code = output.CodeValidation
	default:
		code = output.CodeUnknown
	}

	formatted := fmt.Sprintf("%s %s: %d %s", method, path, status, msg)
	if retryAfter > 0 {
		formatted += fmt.Sprintf(" (retry-after %s)", retryAfter)
	}
	return &output.Error{
		Code:    code,
		Message: formatted,
	}
}

// maxErrorMessageBytes caps how much upstream-error text we embed in the
// envelope's `error` field. Anything longer is almost always an HTML page
// from a proxy and useless for the agent to read.
const maxErrorMessageBytes = 512
