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

package v2

// TOTP represents a TOTP secret for a user.
type TOTP struct {
	Secret   string   `json:"secret,omitempty"`
	URL      string   `json:"url,omitempty"`
	Recovery []string `json:"recovery,omitempty"`
	Links    *TOTPLinks `json:"_links"`
}

// TOTPLinks represents the links for a TOTP secret.
type TOTPLinks struct {
	Self   *Link `json:"self"`
	QRCode *Link `json:"qrcode"`
}

// TOTPPasscode represents the request body for enabling TOTP.
type TOTPPasscode struct {
	Passcode string `json:"passcode"`
}
