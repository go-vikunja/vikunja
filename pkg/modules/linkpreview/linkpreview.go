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

package linkpreview

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/version"

	"github.com/otiai10/opengraph/v2"
	"golang.org/x/net/html/charset"
)

type Preview struct {
	URL         string `json:"url" doc:"Final HTTP(S) destination after redirects."`
	Title       string `json:"title" doc:"Open Graph title, falling back to the HTML title."`
	Description string `json:"description" doc:"Open Graph description, falling back to the page description."`
	Image       string `json:"image" doc:"Absolute HTTP(S) image URL, or empty when unavailable."`
	SiteName    string `json:"site_name" doc:"Name of the linked website."`
	Favicon     string `json:"favicon" doc:"Absolute HTTP(S) icon URL, or empty when unavailable."`
}

const previewCacheTTL = 24 * time.Hour

func GetPreview(ctx context.Context, rawURL string) (*Preview, error) {
	u, err := previewURL(rawURL)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, &ErrPreviewUnavailable{Err: err}
	}
	cacheKey := "link_preview_" + utils.Sha256Hex(u.String())
	var cached Preview
	exists, err := keyvalue.GetWithValue(cacheKey, &cached)
	if err != nil {
		log.Errorf("Error retrieving link preview from keyvalue store: %s", err)
	} else if exists {
		if _, err := previewURL(cached.URL); err != nil {
			return nil, err
		}
		return &cached, nil
	}

	preview, err := fetchPreview(ctx, u)
	if err != nil {
		return nil, err
	}
	if err := keyvalue.PutWithTTL(cacheKey, *preview, previewCacheTTL); err != nil {
		log.Errorf("Error storing link preview in keyvalue store: %s", err)
	}
	return preview, nil
}

func fetchPreview(ctx context.Context, u *url.URL) (*Preview, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	client := utils.NewSSRFSafeHTTPClient()
	defer client.CloseIdleConnections()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		_, err := previewURL(req.URL.String())
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, &ErrInvalidURL{}
	}
	req.Header.Set("User-Agent", "Vikunja-LinkPreview/"+version.Version)
	req.Header.Set("Accept", "text/html, application/xhtml+xml")
	resp, err := client.Do(req) //nolint:gosec // The shared client guards resolved IPs, including redirects.
	if err != nil {
		return nil, &ErrPreviewUnavailable{Err: err}
	}
	defer resp.Body.Close()
	contentType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices ||
		(contentType != "text/html" && contentType != "application/xhtml+xml") {
		return nil, &ErrPreviewUnavailable{Err: fmt.Errorf("upstream did not return an HTML page")}
	}
	const maxBodySize = 2 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize+1))
	if err != nil {
		return nil, &ErrPreviewUnavailable{Err: err}
	}
	if len(body) > maxBodySize {
		return nil, &ErrPreviewUnavailable{Err: fmt.Errorf("HTML page exceeds size limit")}
	}
	reader, err := charset.NewReader(bytes.NewReader(body), resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, &ErrPreviewUnavailable{Err: err}
	}
	metadata := opengraph.New(resp.Request.URL.String())
	if err := metadata.Parse(reader); err != nil {
		return nil, &ErrPreviewUnavailable{Err: err}
	}
	preview := &Preview{
		URL:         resp.Request.URL.String(),
		Title:       strings.TrimSpace(metadata.Title),
		Description: strings.TrimSpace(metadata.Description),
		SiteName:    strings.TrimSpace(metadata.SiteName),
		Favicon:     assetURL(resp.Request.URL, metadata.Favicon.URL),
	}
	for _, image := range metadata.Image {
		preview.Image = assetURL(resp.Request.URL, image.SecureURL)
		if preview.Image == "" {
			preview.Image = assetURL(resp.Request.URL, image.URL)
		}
		if preview.Image != "" {
			break
		}
	}
	return preview, nil
}

func previewURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, &ErrInvalidURL{}
	}
	base, err := url.Parse(config.ServicePublicURL.GetString())
	if err == nil && strings.EqualFold(u.Hostname(), base.Hostname()) &&
		u.Scheme == base.Scheme && httpPort(u) == httpPort(base) {
		path := strings.TrimSuffix(base.Path, "/")
		if u.Path == path || strings.HasPrefix(u.Path, path+"/") {
			return nil, &ErrInvalidURL{}
		}
	}
	u.Fragment = ""
	return u, nil
}

func httpPort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
}

func assetURL(base *url.URL, raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	u = base.ResolveReference(u)
	if u.User != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return u.String()
}
