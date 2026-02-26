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

// FlutterClientID is the hard-coded client ID for the official Flutter app.
const FlutterClientID = "vikunja-flutter"

// allowedRedirectURIs lists the redirect URIs allowed for the Flutter client.
// These are custom protocol URIs that the mobile app registers to handle.
var allowedRedirectURIs = []string{
	"vikunja://callback",
}

// ValidateClient checks that the client_id is recognized.
func ValidateClient(clientID string) bool {
	return clientID == FlutterClientID
}

// ValidateRedirectURI checks that the redirect_uri is allowed for the given client.
func ValidateRedirectURI(clientID, redirectURI string) bool {
	if clientID != FlutterClientID {
		return false
	}
	for _, uri := range allowedRedirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}
