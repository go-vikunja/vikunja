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
	var tokenErr *TokenError
	if errors.As(err, &tokenErr) {
		return err
	}
	for _, m := range tokenErrorCodes {
		if m.is(err) {
			return &TokenError{Code: m.code, Err: err}
		}
	}
	return err
}

// The meaning is the token endpoint's: a redirect or client mismatch there is a grant issued to someone else.
var tokenErrorCodes = []struct {
	is   func(error) bool
	code TokenErrorCode
}{
	{models.IsErrOAuthCodeInvalid, TokenErrorInvalidGrant},
	{models.IsErrOAuthCodeExpired, TokenErrorInvalidGrant},
	{models.IsErrOAuthPKCEVerifyFailed, TokenErrorInvalidGrant},
	{models.IsErrOAuthInvalidRedirectURI, TokenErrorInvalidGrant},
	{models.IsErrOAuthClientNotFound, TokenErrorInvalidGrant},
	{models.IsErrInvalidRefreshToken, TokenErrorInvalidGrant},
	{models.IsErrSessionExpired, TokenErrorInvalidGrant},
	{models.IsErrRefreshTokenAlreadyUsed, TokenErrorInvalidGrant},
	// The user the grant was issued to was deleted, disabled or locked since.
	{user.IsErrUserDoesNotExist, TokenErrorInvalidGrant},
	{user.IsErrUserCannotAuthenticate, TokenErrorInvalidGrant},
	{models.IsErrOAuthInvalidClientID, TokenErrorInvalidClient},
	{models.IsErrOAuthClientIDIsServer, TokenErrorInvalidClient},
	{models.IsErrOAuthClientMetadataUnavailable, TokenErrorInvalidClient},
	{models.IsErrOAuthInvalidClientMetadata, TokenErrorInvalidClient},
	{models.IsErrOAuthClientSecretUnsupported, TokenErrorInvalidClient},
	{models.IsErrOAuthMalformedClientCredentials, TokenErrorInvalidRequest},
	{models.IsErrOAuthInvalidGrantType, TokenErrorUnsupportedGrantType},
	{models.IsErrOAuthInvalidScope, TokenErrorInvalidScope},
	{models.IsErrOAuthInvalidTarget, TokenErrorInvalidTarget},
}
