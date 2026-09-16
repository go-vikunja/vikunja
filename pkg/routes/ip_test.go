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

package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestNewIPExtractor(t *testing.T) {
	log.InitLogger()

	// RemoteAddr Go sets for a request coming in through a unix socket
	const unixSocketRemoteAddr = "@"

	tests := []struct {
		name           string
		method         string
		trustedProxies string
		remoteAddr     string
		headers        map[string]string
		want           string
	}{
		{
			name:       "direct over tcp",
			method:     "direct",
			remoteAddr: "203.0.113.5:44231",
			want:       "203.0.113.5",
		},
		{
			name:       "direct ignores forwarding headers",
			method:     "direct",
			remoteAddr: "127.0.0.1:44231",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.5"},
			want:       "127.0.0.1",
		},
		{
			name:           "xff over tcp",
			method:         "xff",
			trustedProxies: "127.0.0.1/32",
			remoteAddr:     "127.0.0.1:44231",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.5"},
			want:           "203.0.113.5",
		},
		{
			name:           "xff over unix socket",
			method:         "xff",
			trustedProxies: "127.0.0.1/32",
			remoteAddr:     unixSocketRemoteAddr,
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.5"},
			want:           "203.0.113.5",
		},
		{
			name:           "xff over unix socket with a proxy chain",
			method:         "xff",
			trustedProxies: "127.0.0.1/32",
			remoteAddr:     unixSocketRemoteAddr,
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.5, 10.0.0.2, 127.0.0.1"},
			want:           "203.0.113.5",
		},
		{
			// echo only honours X-Real-IP when the header value itself is a
			// trusted IP, hence the private address here
			name:       "realip over unix socket",
			method:     "realip",
			remoteAddr: unixSocketRemoteAddr,
			headers:    map[string]string{"X-Real-Ip": "10.0.0.5"},
			want:       "10.0.0.5",
		},
		{
			name:       "direct over unix socket falls back to loopback",
			method:     "direct",
			remoteAddr: unixSocketRemoteAddr,
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.5"},
			want:       "127.0.0.1",
		},
		{
			name:       "xff over unix socket without any header",
			method:     "xff",
			remoteAddr: unixSocketRemoteAddr,
			want:       "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/login", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			extractor := newIPExtractor(tt.method, tt.trustedProxies)

			assert.Equal(t, tt.want, extractor(req))
			assert.Equal(t, tt.remoteAddr, req.RemoteAddr, "the request must not be modified")
		})
	}
}
