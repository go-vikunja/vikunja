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
	"net/http"
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

// ApplyBasicClientAuth reads req's client_id from Basic auth (RFC 6749 §2.3.1); call it before ExchangeToken so a rejection doesn't consume the code.
func ApplyBasicClientAuth(req *TokenRequest, r *http.Request) (basic bool, err error) {
	// BasicAuth can't tell a malformed Basic header from a missing one.
	scheme, _, _ := strings.Cut(r.Header.Get("Authorization"), " ")
	if !strings.EqualFold(scheme, "Basic") {
		return false, nil
	}
	clientID, secret, err := parseBasicClientCredentials(r)
	if err != nil {
		return true, asTokenError(err)
	}
	if secret != "" {
		return true, asTokenError(&models.ErrOAuthClientSecretUnsupported{})
	}
	if req.ClientID != "" && req.ClientID != clientID {
		// RFC 6749 §5.2: "includes multiple credentials".
		return true, asTokenError(&models.ErrOAuthMalformedClientCredentials{Reason: "the client_id parameter does not match the Authorization header"})
	}
	req.ClientID = clientID
	return true, nil
}

func parseBasicClientCredentials(r *http.Request) (clientID, secret string, err error) {
	malformed := &models.ErrOAuthMalformedClientCredentials{Reason: "the Authorization header is not valid Basic credentials"}
	user, password, ok := r.BasicAuth()
	if !ok {
		return "", "", malformed
	}
	// Both are form-urlencoded before Basic encoding (RFC 6749 §2.3.1).
	clientID, err = url.QueryUnescape(user)
	if err != nil || clientID == "" {
		return "", "", malformed
	}
	secret, err = url.QueryUnescape(password)
	if err != nil {
		return "", "", malformed
	}
	return clientID, secret, nil
}
