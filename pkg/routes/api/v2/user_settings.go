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

package apiv2

import (
	"context"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/api/shared"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkuchiki/go-timezone"
)

// userInfoBody is the GET /user response: the public user fields plus the
// computed account facts v1 returned alongside the user object.
type userInfoBody struct {
	user.User
	Settings            *models.UserGeneralSettings `json:"settings" readOnly:"true" doc:"The current user's settings."`
	DeletionScheduledAt time.Time                   `json:"deletion_scheduled_at" readOnly:"true" doc:"When the account is scheduled for deletion, if a deletion was requested."`
	IsLocalUser         bool                        `json:"is_local_user" readOnly:"true" doc:"True if the user authenticates locally (not via LDAP or OpenID)."`
	AuthProvider        string                      `json:"auth_provider" readOnly:"true" doc:"The name of the source the user authenticated with: 'local', 'ldap', or the configured OpenID provider name."`
	IsAdmin             bool                        `json:"is_admin" readOnly:"true" doc:"True if the user is an instance administrator."`
}

// userAvatarProviderBody is the get/set body for the user's avatar provider.
type userAvatarProviderBody struct {
	AvatarProvider string `json:"avatar_provider" doc:"The avatar provider. One of: gravatar (uses the user email), upload, initials, marble (random per user), ldap (synced from LDAP), openid (synced from OpenID), default."`
}

type userActionMessageBody struct {
	Message string `json:"message" readOnly:"true" doc:"A confirmation message."`
}

// RegisterUserSettingsRoutes wires the current-user account & settings
// endpoints onto the Huma API. These are not CRUDable resources: each operates
// on the authenticated user pulled from the request context.
func RegisterUserSettingsRoutes(api huma.API) {
	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID: "user-show",
		Summary:     "Get the current user",
		Description: "Returns the authenticated user together with their settings and computed account facts (auth_provider, is_local_user, is_admin, deletion_scheduled_at).",
		Method:      http.MethodGet,
		Path:        "/user",
		Tags:        tags,
	}, userShow)

	Register(api, huma.Operation{
		OperationID: "user-change-password",
		Summary:     "Change the current user's password",
		Description: "Changes the authenticated user's password after verifying the old one. All of the user's existing sessions are invalidated.",
		Method:      http.MethodPost,
		Path:        "/user/password",
		// Changes a password, it creates nothing — keep 200 over the wrapper's POST→201.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, userChangePassword)

	Register(api, huma.Operation{
		OperationID: "user-update-email",
		Summary:     "Update the current user's email address",
		Description: "Sets a new email address for the authenticated user after verifying their password. If the mailer is enabled the change is pending until the user confirms it via a link sent to the new address; otherwise it takes effect immediately.",
		Method:      http.MethodPut,
		Path:        "/user/settings/email",
		Tags:        tags,
	}, userUpdateEmail)

	Register(api, huma.Operation{
		OperationID: "user-update-settings",
		Summary:     "Update the current user's general settings",
		Description: "Replaces the authenticated user's general settings (name, reminders, discoverability, default project, week start, language, timezone, frontend settings).",
		Method:      http.MethodPut,
		Path:        "/user/settings/general",
		Tags:        tags,
	}, userUpdateSettings)

	// Path differs from v1's /user/settings/avatar: on v2 that path is the
	// binary avatar upload (PUT), so the provider get/set live on a sub-path.
	Register(api, huma.Operation{
		OperationID: "user-get-avatar-provider",
		Summary:     "Get the current user's avatar provider",
		Description: "Returns the avatar provider configured for the authenticated user.",
		Method:      http.MethodGet,
		Path:        "/user/settings/avatar/provider",
		Tags:        tags,
	}, userGetAvatarProvider)

	Register(api, huma.Operation{
		OperationID: "user-set-avatar-provider",
		Summary:     "Set the current user's avatar provider",
		Description: "Changes the avatar provider for the authenticated user. Valid values: gravatar, upload, initials, marble, ldap, openid, default.",
		Method:      http.MethodPut,
		Path:        "/user/settings/avatar/provider",
		Tags:        tags,
	}, userSetAvatarProvider)

	Register(api, huma.Operation{
		OperationID: "user-timezones",
		Summary:     "List available time zones",
		Description: "Returns every time zone this Vikunja instance can handle. The list depends on the host system and is unsorted; sort it client-side.",
		Method:      http.MethodGet,
		Path:        "/user/timezones",
		Tags:        tags,
	}, userTimezones)
}

func init() { AddRouteRegistrar(RegisterUserSettingsRoutes) }

func userShow(ctx context.Context, _ *struct{}) (*singleBody[userInfoBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	u, err := models.GetUserOrLinkShareUser(s, a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	info := &userInfoBody{
		User:                *u,
		Settings:            models.NewUserGeneralSettings(u),
		DeletionScheduledAt: u.DeletionScheduledAt,
		IsLocalUser:         u.Issuer == user.IssuerLocal,
		IsAdmin:             u.IsAdmin,
	}

	// nolint:contextcheck // openid.GetAllProviders/Issuer (called via shared) take
	// no context; threading one would change those signatures across both APIs.
	info.AuthProvider, err = shared.GetAuthProviderName(u)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[userInfoBody]{Body: info}, nil
}

func userChangePassword(ctx context.Context, in *struct {
	Body struct {
		OldPassword string `json:"old_password" doc:"The current password, for confirmation."`
		NewPassword string `json:"new_password" valid:"bcrypt_password" minLength:"8" maxLength:"72" doc:"The new password. Max 72 bytes (a bcrypt limit), which may be fewer than 72 characters."`
	}
}) (*singleBody[userActionMessageBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	if err := models.ChangeUserPassword(s, doer, in.Body.OldPassword, in.Body.NewPassword); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[userActionMessageBody]{Body: &userActionMessageBody{Message: "The password was updated successfully."}}, nil
}

func userUpdateEmail(ctx context.Context, in *struct {
	Body struct {
		NewEmail string `json:"new_email" valid:"email,length(0|250),required" maxLength:"250" doc:"The new email address."`
		Password string `json:"password" doc:"The current password, for confirmation."`
	}
}) (*singleBody[userActionMessageBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	if err := user.ChangeUserEmail(s, doer, in.Body.Password, in.Body.NewEmail); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[userActionMessageBody]{Body: &userActionMessageBody{Message: "We sent you email with a link to confirm your email address."}}, nil
}

func userUpdateSettings(ctx context.Context, in *struct {
	Body models.UserGeneralSettings
}) (*singleBody[userActionMessageBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserWithEmail(s, &user.User{ID: doer.ID})
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := models.UpdateUserGeneralSettings(s, u, &in.Body); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[userActionMessageBody]{Body: &userActionMessageBody{Message: "The settings were updated successfully."}}, nil
}

func userGetAvatarProvider(ctx context.Context, _ *struct{}) (*singleBody[userAvatarProviderBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserWithEmail(s, &user.User{ID: doer.ID})
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[userAvatarProviderBody]{Body: &userAvatarProviderBody{AvatarProvider: u.AvatarProvider}}, nil
}

func userSetAvatarProvider(ctx context.Context, in *struct {
	Body userAvatarProviderBody
}) (*singleBody[userAvatarProviderBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserWithEmail(s, &user.User{ID: doer.ID})
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := models.UpdateUserAvatarProvider(s, u, in.Body.AvatarProvider); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[userAvatarProviderBody]{Body: &userAvatarProviderBody{AvatarProvider: u.AvatarProvider}}, nil
}

type timezonesBody struct {
	Body []string
}

func userTimezones(ctx context.Context, _ *struct{}) (*timezonesBody, error) {
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}

	timezoneMap := make(map[string]bool) // de-dupe across the per-abbreviation groups
	for _, group := range timezone.New().Timezones() {
		for _, t := range group {
			timezoneMap[t] = true
		}
	}

	ts := make([]string, 0, len(timezoneMap))
	for t := range timezoneMap {
		ts = append(ts, t)
	}

	return &timezonesBody{Body: ts}, nil
}
