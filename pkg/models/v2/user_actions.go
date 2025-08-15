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

// EmailConfirm represents the request body for confirming a user's email.
type EmailConfirm struct {
	Token string `json:"token"`
}

// PasswordReset represents the request body for resetting a user's password.
type PasswordReset struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// PasswordTokenRequest represents the request body for requesting a password reset token.
type PasswordTokenRequest struct {
	Username string `json:"username"`
}
