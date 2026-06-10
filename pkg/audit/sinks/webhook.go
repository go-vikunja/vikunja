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

package sinks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"code.vikunja.io/api/pkg/utils"
)

// Webhook POSTs each entry as a JSON body to a fixed URL.
type Webhook struct {
	url     string
	headers map[string]string
	client  *http.Client
}

func NewWebhook(url string, headers map[string]string) (*Webhook, error) {
	if url == "" {
		return nil, fmt.Errorf("webhook forwarder requires a url")
	}
	return &Webhook{
		url:     url,
		headers: headers,
		client:  utils.NewSSRFSafeHTTPClient(),
	}, nil
}

func (w *Webhook) Write(line []byte) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, w.url, bytes.NewReader(line))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Vikunja/audit")
	for key, value := range w.headers {
		req.Header.Set(key, value)
	}

	resp, err := w.client.Do(req) // #nosec G704 -- URL is the operator-configured sink target; the SSRF-safe client enforces IP restrictions
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("audit webhook %s returned status %d", w.url, resp.StatusCode)
	}
	return nil
}
