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

package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
)

// DownloadImage downloads an image from a URL and returns the image data
func DownloadImage(ctx context.Context, url string) ([]byte, error) {
	// 3 seconds is enough for downloading an avatar
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := NewSSRFSafeHTTPClient().Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	// Bound the remote stream before buffering it.
	// #nosec G115 - the configured size is far below int64 max.
	maxBytes := int64(config.GetMaxFileSizeInMBytes()) * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("downloaded image exceeds the maximum file size of %d bytes", maxBytes)
	}

	return data, nil
}
