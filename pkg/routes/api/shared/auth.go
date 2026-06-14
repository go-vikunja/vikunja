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
	"code.vikunja.io/api/pkg/user"
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
