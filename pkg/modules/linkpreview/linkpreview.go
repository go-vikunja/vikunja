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

// Package linkpreview fetches an external http(s) URL and extracts its
// OpenGraph/meta preview (title, description, image, site name, favicon).
// Fetching goes through the SSRF-safe HTTP client; parsing is delegated to
// github.com/otiai10/opengraph.
package linkpreview

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/utils"

	"github.com/otiai10/opengraph/v2"
)

const (
	// The <head> metadata sits at the top of the document, so a couple of MiB
	// is plenty and caps memory for hostile responses.
	maxBodyBytes = 2 << 20
	fetchTimeout = 8 * time.Second
	maxTitle     = 300
	maxDesc      = 500
)

// Errors the caller can map onto transport-specific status codes.
var (
	ErrInvalidURL  = errors.New("linkpreview: only absolute http(s) urls are supported")
	ErrFetchFailed = errors.New("linkpreview: could not fetch url")
	ErrTimeout     = errors.New("linkpreview: upstream timed out")
)

// Preview is the OpenGraph/meta summary of an external URL.
type Preview struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	Favicon     string `json:"favicon,omitempty"`
}

// GetPreview fetches rawURL and returns its preview. The SSRF-safe client
// re-validates the resolved IP on every dial (including redirect hops), so
// private/loopback/link-local targets are refused before any connection.
func GetPreview(ctx context.Context, rawURL string) (*Preview, error) {
	target, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return nil, ErrInvalidURL
	}

	reqCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, ErrInvalidURL
	}
	req.Header.Set("User-Agent", "Vikunja-LinkPreview/1.0 (+https://vikunja.io)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := utils.NewSSRFSafeHTTPClient().Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		if reqCtx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, errors.Join(ErrFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, ErrFetchFailed
	}

	// Resolve relative image/favicon URLs against the final (post-redirect) URL.
	finalURL := resp.Request.URL
	preview := &Preview{URL: finalURL.String()}

	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.Contains(strings.ToLower(ct), "html") {
		// Nothing to unfurl from non-HTML responses; return the bare URL.
		return preview, nil
	}

	og := opengraph.New(finalURL.String())
	if err := og.Parse(io.LimitReader(resp.Body, maxBodyBytes)); err != nil {
		return preview, nil
	}

	fillPreview(preview, og, finalURL)
	return preview, nil
}

func fillPreview(preview *Preview, og *opengraph.OpenGraph, base *url.URL) {
	preview.Title = truncate(og.Title, maxTitle)
	preview.Description = truncate(og.Description, maxDesc)
	preview.SiteName = truncate(og.SiteName, maxTitle)
	if len(og.Image) > 0 {
		preview.Image = resolveURL(base, og.Image[0].URL)
	}
	preview.Favicon = resolveURL(base, og.Favicon.URL)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimSpace(string(r[:max])) + "…"
}

// resolveURL turns a possibly-relative asset reference into an absolute http(s)
// URL, dropping anything with a non-http(s) scheme (data:, javascript:, …).
func resolveURL(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	return resolved.String()
}
