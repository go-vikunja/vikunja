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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoRequestRefusesRedirects(t *testing.T) {
	config.InitDefaultConfig()
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
	defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"valid":true}`))
	}))
	defer target.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	resp, err := doRequest(redirector.URL, []byte(`{}`))
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, errRedirectNotFollowed)
}

func TestDoRequestBlocksNonRoutableTarget(t *testing.T) {
	config.InitDefaultConfig()
	config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"valid":true}`))
	}))
	defer server.Close()

	resp, err := doRequest(server.URL, []byte(`{}`))
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.NotErrorIs(t, err, errRedirectNotFollowed)
}

func TestDoRequestSucceeds(t *testing.T) {
	config.InitDefaultConfig()
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
	defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"valid":true,"max_users":5}`))
	}))
	defer server.Close()

	resp, err := doRequest(server.URL, []byte(`{"license_key":"abc"}`))
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, int64(5), resp.MaxUsers)
	assert.JSONEq(t, `{"license_key":"abc"}`, string(gotBody))
}
