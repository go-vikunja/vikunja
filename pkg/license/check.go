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

package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/version"
)

// License server URLs — hardcoded to prevent bypass.
var licenseServers = []string{
	"https://license1.vikunja.io/api/v1/check",
	"https://license2.vikunja.io/api/v1/check",
}

const (
	maxRetries     = 3
	requestTimeout = 10 * time.Second
)

// CheckRequest is the payload sent to the license server.
type CheckRequest struct {
	LicenseKey   string     `json:"license_key"`
	InstanceID   string     `json:"instance_id"`
	Version      string     `json:"version"`
	DatabaseType string     `json:"database_type"`
	UserCounts   UserCounts `json:"user_counts"`
	HostOS       string     `json:"host_os"`
	IsContainer  bool       `json:"is_container"`
}

// UserCounts holds user counts by status.
type UserCounts struct {
	Active                   int64 `json:"active"`
	Disabled                 int64 `json:"disabled"`
	EmailConfirmationPending int64 `json:"email_confirmation_pending"`
}

// Response is the response from the license server.
type Response struct {
	Valid     bool      `json:"valid"`
	Message   string    `json:"message,omitempty"`
	Features  []string  `json:"features"`
	MaxUsers  int64     `json:"max_users"`
	ExpiresAt time.Time `json:"expires_at"`
}

func checkLicense(key string) (*Response, error) {
	payload, err := buildPayload(key)
	if err != nil {
		return nil, fmt.Errorf("building license check payload: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling license check payload: %w", err)
	}

	for _, server := range licenseServers {
		resp, err := tryServer(server, body)
		if err != nil {
			log.Debugf("License server %s unreachable: %s", server, err)
			continue
		}
		return resp, nil
	}

	return nil, fmt.Errorf("all license servers unreachable")
}

func tryServer(serverURL string, body []byte) (*Response, error) {
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			baseDelay := time.Duration(1) * time.Second
			for range attempt {
				baseDelay *= 3
			}
			// Add ±30% jitter
			jitter := 1.0 + (rand.Float64()*0.6 - 0.3) // #nosec G404 - jitter does not need cryptographic randomness
			delay := time.Duration(float64(baseDelay) * jitter)
			time.Sleep(delay)
		}

		resp, err := doRequest(serverURL, body)
		if err != nil {
			lastErr = err
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

func doRequest(serverURL string, body []byte) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var licenseResp Response
	if err := json.NewDecoder(resp.Body).Decode(&licenseResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &licenseResp, nil
}

func buildPayload(key string) (*CheckRequest, error) {
	userCounts, err := getUserCounts()
	if err != nil {
		return nil, fmt.Errorf("getting user counts: %w", err)
	}

	return &CheckRequest{
		LicenseKey:   key,
		InstanceID:   instanceID,
		Version:      version.Version,
		DatabaseType: config.DatabaseType.GetString(),
		UserCounts:   userCounts,
		HostOS:       runtime.GOOS,
		IsContainer:  detectContainer(),
	}, nil
}

func getUserCounts() (UserCounts, error) {
	s := db.NewSession()
	defer s.Close()

	var counts UserCounts

	active, err := s.Table("users").Where("status = ?", user.StatusActive).Count()
	if err != nil {
		return counts, err
	}
	counts.Active = active

	disabled, err := s.Table("users").Where("status = ?", user.StatusDisabled).Count()
	if err != nil {
		return counts, err
	}
	counts.Disabled = disabled

	pending, err := s.Table("users").Where("status = ?", user.StatusEmailConfirmationRequired).Count()
	if err != nil {
		return counts, err
	}
	counts.EmailConfirmationPending = pending

	return counts, nil
}

func detectContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return true
	}
	return false
}

func parseResponse(raw string) (*Response, error) {
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("parsing cached license response: %w", err)
	}
	return &resp, nil
}

func serializeResponse(resp *Response) (string, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("serializing license response: %w", err)
	}
	return string(data), nil
}
