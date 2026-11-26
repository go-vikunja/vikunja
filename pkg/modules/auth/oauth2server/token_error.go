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
	"errors"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
)

// TokenErrorCode is the error of a token error response (RFC 6749 §5.2, RFC 8707 §2).
type TokenErrorCode string

const (
	TokenErrorInvalidRequest       TokenErrorCode = "invalid_request"
	TokenErrorInvalidClient        TokenErrorCode = "invalid_client"
	TokenErrorInvalidGrant         TokenErrorCode = "invalid_grant"
	TokenErrorUnsupportedGrantType TokenErrorCode = "unsupported_grant_type"
	TokenErrorInvalidScope         TokenErrorCode = "invalid_scope"
	TokenErrorInvalidTarget        TokenErrorCode = "invalid_target"
)

// TokenError is a rejected token request. Err is the domain error it was rejected with.
type TokenError struct {
	Code TokenErrorCode
	Err  error
}

func (e *TokenError) Error() string { return e.Err.Error() }

func (e *TokenError) Unwrap() error { return e.Err }

// asTokenError leaves errors without an OAuth meaning, internal ones, as they are.
func asTokenError(err error) error {
	code, ok := tokenErrorCode(err)
	if !ok {
		return err
	}
	return &TokenError{Code: code, Err: err}
}

// The meaning is the token endpoint's: a redirect or client mismatch there is a grant issued to someone else.
func tokenErrorCode(err error) (TokenErrorCode, bool) {
	{
		var errCase0 *models.ErrOAuthCodeInvalid
		var errCase1 *models.ErrOAuthCodeExpired
		var errCase2 *models.ErrOAuthPKCEVerifyFailed
		var errCase3 *models.ErrOAuthInvalidRedirectURI
		var errCase4 *models.ErrOAuthClientNotFound
		var errCase5 *models.ErrInvalidRefreshToken
		var errCase6 *models.ErrSessionExpired
		var errCase7 *models.ErrRefreshTokenAlreadyUsed
		var errCase8 *models.ErrOAuthInvalidClientID
		var errCase9 *models.ErrOAuthClientIDIsServer
		var errCase10 *models.ErrOAuthClientMetadataUnavailable
		var errCase11 *models.ErrOAuthInvalidClientMetadata
		var errCase12 *models.ErrOAuthInvalidGrantType
		var errCase13 *models.ErrOAuthInvalidScope
		var errCase14 *models.ErrOAuthInvalidTarget
		switch {
		case errors.As(err, &errCase0), errors.As(err, &errCase1), errors.As(err, &errCase2), errors.As(err, &errCase3), errors.As(err, &errCase4), errors.As(err, &errCase5), errors.As(err, &errCase6), errors.As(err, &errCase7):
			return TokenErrorInvalidGrant, true
		case errors.As(err, &errCase8), errors.As(err, &errCase9), errors.As(err, &errCase10), errors.As(err, &errCase11):
			return TokenErrorInvalidClient, true
		case errors.As(err, &errCase12):
			return TokenErrorUnsupportedGrantType, true
		case errors.As(err, &errCase13):
			return TokenErrorInvalidScope, true
		case errors.As(err, &errCase14):
			return TokenErrorInvalidTarget, true
		}
	}
	// The user the grant was issued to was deleted, disabled or locked since.
	if user.IsErrUserDoesNotExist(err) || user.IsErrUserCannotAuthenticate(err) {
		return TokenErrorInvalidGrant, true
	}
	return "", false
}
