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

package migration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
)

// DownloadFile downloads a file and returns its contents
func DownloadFile(url string) (buf *bytes.Buffer, err error) {
	return DownloadFileWithHeaders(url, nil)
}

// DownloadFileWithHeaders downloads a file and allows you to pass in headers
func DownloadFileWithHeaders(url string, headers http.Header) (buf *bytes.Buffer, err error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for key, h := range headers {
		for _, hh := range h {
			req.Header.Add(key, hh)
		}
	}

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf = &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)

	return
}

// DoPost makes a form encoded post request
func DoPost(url string, form url.Values) (resp *http.Response, err error) {
	return DoPostWithHeaders(url, form, map[string]string{})
}

// DoGetWithHeaders makes an HTTP GET request with custom headers
func DoGetWithHeaders(urlStr string, headers map[string]string) (resp *http.Response, err error) {
	hc := http.Client{}

	err = utils.RetryWithBackoff("HTTP GET "+urlStr, func() error {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, urlStr, nil)
		if reqErr != nil {
			return reqErr
		}

		for key, value := range headers {
			req.Header.Add(key, value)
		}

		resp, err = hc.Do(req) //nolint:bodyclose // Caller is responsible for closing on success
		if err != nil {
			return err
		}

		// Log 4xx errors for debugging
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			// Re-create the body so the caller can still read it
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			log.Debugf("[Migration] HTTP GET %s returned %d: %s", urlStr, resp.StatusCode, string(bodyBytes))
			return nil // Don't retry on 4xx
		}

		// Retry on 5xx status codes, include response body in error
		if resp.StatusCode >= 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		return nil
	})

	return resp, err
}

// DoPostWithHeaders does an api request and allows to pass in arbitrary headers
func DoPostWithHeaders(urlStr string, form url.Values, headers map[string]string) (resp *http.Response, err error) {
	hc := http.Client{}

	err = utils.RetryWithBackoff("HTTP POST "+urlStr, func() error {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodPost, urlStr, strings.NewReader(form.Encode()))
		if reqErr != nil {
			return reqErr
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		for key, value := range headers {
			req.Header.Add(key, value)
		}

		resp, err = hc.Do(req) //nolint:bodyclose // Caller is responsible for closing on success
		if err != nil {
			return err
		}

		// Log 4xx errors for debugging
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			// Re-create the body so the caller can still read it
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			log.Debugf("[Migration] HTTP POST %s returned %d: %s", urlStr, resp.StatusCode, string(bodyBytes))
			return nil // Don't retry on 4xx
		}

		// Retry on 5xx status codes, include response body in error
		if resp.StatusCode >= 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		return nil
	})

	return resp, err
}
