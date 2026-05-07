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

package oauth2server

import (
	"net"
	"net/url"
	"strings"
)

// ValidateRedirectURI checks that the redirect_uri is either a Vikunja native
// app scheme (e.g. vikunja-flutter://callback) or a loopback http URL as
// recommended by RFC 8252 for native apps that cannot register a custom
// scheme. Any address in 127.0.0.0/8, the IPv6 loopback (::1, in any
// notation), and the literal hostname "localhost" are accepted; dangerous
// schemes like javascript:, data:, https://, or non-loopback http:// targets
// are rejected.
func ValidateRedirectURI(redirectURI string) bool {
	u, err := url.Parse(redirectURI)
	if err != nil || u.Scheme == "" {
		return false
	}

	if strings.HasPrefix(u.Scheme, "vikunja-") {
		return true
	}

	if u.Scheme == "http" {
		host := u.Hostname()
		if strings.EqualFold(host, "localhost") {
			return true
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return true
		}
	}

	return false
}
