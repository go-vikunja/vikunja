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

package shared

import (
	"context"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/ldap"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// UserRegister carries the fields accepted by the public registration endpoint:
// username, password and email (from APIUserPassword) plus the new user's
// preferred language.
type UserRegister struct {
	// The language of the new user. Must be a valid IETF BCP 47 language code and exist in Vikunja.
	Language string `json:"language" valid:"language" doc:"The language of the new user as an IETF BCP 47 code (e.g. en, de-DE)."`
	user.APIUserPassword
}

// RegisterUser creates a new local user account from the registration input and
// busts the cached user-count metric so the registration shows up immediately.
// The caller is responsible for the registration-enabled gate and input
// validation; both v1 and v2 share this body.
func RegisterUser(ctx context.Context, in *UserRegister) (*user.User, error) {
	s := db.NewSession()
	defer s.Close()
	// Discards events queued during a rolled-back transaction; a no-op once
	// DispatchPending has run.
	defer events.CleanupPending(s)

	newUser, err := models.RegisterUser(s, &user.User{
		Username: in.Username,
		Password: in.Password,
		Email:    in.Email,
		Language: in.Language,
	})
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	events.DispatchPending(ctx, s)

	// Bust the cached user count so the new registration shows up in metrics
	// immediately instead of after the regular cache expiry.
	if config.MetricsEnabled.GetBool() {
		if err := metrics.InvalidateCount(metrics.UserCountKey); err != nil {
			log.Errorf("Could not invalidate user count metric: %s", err)
		}
	}

	return newUser, nil
}

// AuthenticateUserCredentials verifies a login against local (and, if configured,
// LDAP) credentials and enforces the account-status and TOTP gates, returning the
// authenticated user on success. It is the transport-agnostic core of the login
// flow shared by v1 and v2; the caller issues the token and sets the cookie. The
// returned errors carry their own HTTP semantics (wrong credentials, disabled
// account, missing/invalid TOTP) so both APIs surface them identically.
func AuthenticateUserCredentials(ctx context.Context, login *user.Login) (*user.User, error) {
	s := db.NewSession()
	defer s.Close()
	// Discards events queued during a rolled-back transaction (e.g. LDAP user
	// creation); a no-op once DispatchPending has run.
	defer events.CleanupPending(s)

	u, err := resolveLoginUser(ctx, s, login)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if u.Status == user.StatusDisabled {
		_ = s.Rollback()
		return nil, &user.ErrAccountDisabled{UserID: u.ID}
	}
	if u.Status == user.StatusAccountLocked {
		_ = s.Rollback()
		return nil, &user.ErrAccountLocked{UserID: u.ID}
	}

	if err := enforceLoginTOTP(s, u, login.TOTPPasscode); err != nil {
		return nil, err
	}

	if err := keyvalue.Del(u.GetFailedTOTPAttemptsKey()); err != nil {
		return nil, err
	}
	if err := keyvalue.Del(u.GetFailedPasswordAttemptsKey()); err != nil {
		return nil, err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	events.DispatchPending(ctx, s)

	return u, nil
}

// resolveLoginUser authenticates the credentials against LDAP (when enabled) and
// then against local accounts, mirroring v1's order so local users keep working
// alongside LDAP. Bots are rejected before bcrypt runs because they have no
// password hash.
func resolveLoginUser(ctx context.Context, s *xorm.Session, login *user.Login) (*user.User, error) {
	if config.AuthLdapEnabled.GetBool() {
		u, err := ldap.AuthenticateUserInLDAP(s, login.Username, login.Password, config.AuthLdapGroupSyncEnabled.GetBool(), config.AuthLdapAvatarSyncAttribute.GetString())
		if err != nil && !user.IsErrWrongUsernameOrPassword(err) {
			return nil, err
		}
		if u != nil {
			return u, nil
		}
	}

	existingUser, lookupErr := user.GetUserByUsername(s, login.Username)
	if lookupErr == nil && existingUser.IsBot() {
		return nil, &user.ErrAccountIsBot{UserID: existingUser.ID}
	}

	return user.CheckUserCredentials(ctx, s, login)
}

// enforceLoginTOTP runs the TOTP gate for users who have it enabled, mirroring
// v1: a missing passcode is rejected, and a wrong one trips the failed-attempt
// lockout via HandleFailedTOTPAuth. The session is rolled back before
// HandleFailedTOTPAuth so its dedicated session can acquire a write lock on
// SQLite shared-cache (the lockout write is decoupled from this transaction —
// see GHSA-fgfv-pv97-6cmj).
func enforceLoginTOTP(s *xorm.Session, u *user.User, passcode string) error {
	totpEnabled, err := user.TOTPEnabledForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if !totpEnabled {
		return nil
	}

	if passcode == "" {
		_ = s.Rollback()
		return user.ErrInvalidTOTPPasscode{}
	}

	_, err = user.ValidateTOTPPasscode(s, &user.TOTPPasscode{User: u, Passcode: passcode})
	if err != nil {
		_ = s.Rollback()
		if user.IsErrInvalidTOTPPasscode(err) {
			user.HandleFailedTOTPAuth(u)
		}
		return err
	}

	return nil
}

// DeleteSession removes the session with the given id, logging the user out
// server-side. An empty sid is a no-op (the token carried no session, e.g. an
// API token or a link share), matching v1. Shared by v1 and v2; the caller is
// responsible for clearing the refresh cookie.
func DeleteSession(sid string) error {
	_, err := LogoutSession(sid)
	return err
}

// LogoutSession reads the session, builds an OpenID Connect RP-Initiated Logout
// URL when the session was created via OIDC, then deletes the session. It
// returns the end-session URL (empty for non-OIDC sessions or when no logout
// endpoint is configured) so the frontend can redirect the user agent to the
// identity provider's end_session_endpoint with id_token_hint and
// post_logout_redirect_uri. An empty sid is a no-op. The caller clears the
// refresh cookie.
func LogoutSession(sid string) (endSessionURL string, err error) {
	if sid == "" {
		return "", nil
	}

	s := db.NewSession()
	defer s.Close()

	// Read the session before deleting so the stored id_token can be replayed as
	// id_token_hint. A missing session just means there is nothing to log out.
	session, err := models.GetSessionByID(s, sid)
	if err != nil && !models.IsErrSessionNotFound(err) {
		_ = s.Rollback()
		return "", err
	}
	if session != nil && session.OIDCProviderKey != "" {
		url, buildErr := openid.BuildEndSessionURL(session.OIDCProviderKey, &models.SessionOIDCData{
			IDToken:     session.OIDCIDToken,
			ProviderKey: session.OIDCProviderKey,
		})
		if buildErr != nil {
			// Don't fail logout just because the logout URL could not be built;
			// the session is still destroyed server-side below.
			log.Errorf("Could not build OIDC end-session URL for session %s: %s", sid, buildErr)
		} else {
			endSessionURL = url
		}
	}

	if _, err := s.Where("id = ?", sid).Delete(&models.Session{}); err != nil {
		_ = s.Rollback()
		return "", err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return "", err
	}

	return endSessionURL, nil
}

// ResetPassword resets a user's password from a previously issued reset token
// and invalidates all of that user's sessions, so a leaked password cannot be
// used after a reset. Shared by v1 and v2.
func ResetPassword(reset *user.PasswordReset) error {
	s := db.NewSession()
	defer s.Close()

	userID, err := user.ResetPassword(s, reset)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := models.DeleteAllUserSessions(s, userID); err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}

// RequestPasswordResetToken issues a password-reset token for the account with
// the given email and sends it via email. Shared by v1 and v2.
func RequestPasswordResetToken(req *user.PasswordTokenRequest) error {
	s := db.NewSession()
	defer s.Close()

	if err := user.RequestUserPasswordResetTokenByEmail(s, req); err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}

// ConfirmEmail confirms a newly registered user's email from the token sent to
// them. Shared by v1 and v2.
func ConfirmEmail(confirm *user.EmailConfirm) error {
	s := db.NewSession()
	defer s.Close()

	if err := user.ConfirmEmail(s, confirm); err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}

// LinkShareToken is the response for the link-share auth endpoint. It embeds the
// authenticated share alongside the issued JWT and re-exposes the project id
// (which LinkSharing hides with json:"-"). The embedded share's write-only
// Password is blanked by AuthenticateLinkShare before this is returned.
type LinkShareToken struct {
	auth.Token
	*models.LinkSharing
	ProjectID int64 `json:"project_id" readOnly:"true" doc:"The id of the project this share grants access to."`
}

// AuthenticateLinkShare resolves a link share by its public hash, verifies the
// password for password-protected shares, and issues a JWT auth token for it.
// The returned token's embedded share has its password blanked. Shared by v1
// and v2.
func AuthenticateLinkShare(hash, password string) (*LinkShareToken, error) {
	s := db.NewSession()
	defer s.Close()

	share, err := models.GetLinkShareByHash(s, hash)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if share.SharingType == models.SharingTypeWithPassword {
		if err := models.VerifyLinkSharePassword(share, password); err != nil {
			_ = s.Rollback()
			return nil, err
		}
	}

	t, err := auth.NewLinkShareJWTAuthtoken(share)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	share.Password = ""

	return &LinkShareToken{
		Token:       auth.Token{Token: t},
		LinkSharing: share,
		ProjectID:   share.ProjectID,
	}, nil
}
