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
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadImageIsBounded(t *testing.T) {
	// httptest servers live on 127.0.0.1, which the SSRF-safe client refuses
	// unless non-routable destinations are explicitly allowed.
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set("false") })

	maxBytes := 20 * 1024 * 1024

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(bytes.Repeat([]byte{0x42}, maxBytes+1024))
	}))
	t.Cleanup(srv.Close)

	_, err := DownloadImage(t.Context(), srv.URL)
	require.Error(t, err, "a response larger than the max file size must not be buffered in full")

	picture := &bytes.Buffer{}
	require.NoError(t, png.Encode(picture, image.NewNRGBA(image.Rect(0, 0, 4, 4))))
	small := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(picture.Bytes())
	}))
	t.Cleanup(small.Close)

	data, err := DownloadImage(t.Context(), small.URL)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}
