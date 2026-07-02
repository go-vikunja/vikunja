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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/routes/api/shared"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type adminOverviewBody struct {
	Body *models.Overview
}

type adminUserBody struct {
	Body *shared.AdminUser
}

// adminIsAdminPatchBody uses a pointer so an omitted is_admin leaves the flag unchanged
// instead of silently demoting.
type adminIsAdminPatchBody struct {
	IsAdmin *bool `json:"is_admin" doc:"New admin flag. Omitting it leaves the current value unchanged."`
}

// adminStatusPatchBody uses a pointer so an omitted status leaves the account unchanged
// instead of silently reactivating.
type adminStatusPatchBody struct {
	Status *user.Status `json:"status" doc:"New account status (0=active, 1=email-confirmation required, 2=disabled, 3=locked). Omitting it leaves the current value unchanged."`
}

type adminSetPasswordBody struct {
	NewPassword string `json:"new_password" valid:"bcrypt_password" minLength:"8" maxLength:"72" doc:"The new password. Max 72 bytes (a bcrypt limit), which may be fewer than 72 characters."`
}

// Permissions are enforced by the gateV2AdminRoutes path middleware, not per-handler.
func RegisterAdminUserRoutes(api huma.API) {
	tags := []string{"admin"}

	Register(api, huma.Operation{
		OperationID: "admin-overview",
		Summary:     "Admin overview",
		Description: "Returns per-instance counts (users, projects, tasks, teams, shares) plus the current license snapshot. Restricted to instance admins on a licensed instance; unlicensed or non-admin callers get a 404, making the endpoint indistinguishable from one that is not registered.",
		Method:      http.MethodGet,
		Path:        "/admin/overview",
		Tags:        tags,
	}, adminOverview)

	Register(api, huma.Operation{
		OperationID: "admin-users-create",
		Summary:     "Create a user (admin)",
		Description: "Creates a local user account, bypassing the public-registration toggle. Honours the admin-only is_admin and skip_email_confirm fields. Restricted to instance admins on a licensed instance.",
		Method:      http.MethodPost,
		Path:        "/admin/users",
		Tags:        tags,
	}, adminUsersCreate)

	Register(api, huma.Operation{
		OperationID: "admin-users-patch-admin",
		Summary:     "Promote or demote a user (admin)",
		Description: "Sets a user's instance-admin flag. The body field is a pointer: omitting is_admin leaves the flag unchanged. Demoting the last remaining admin is refused with 400.",
		Method:      http.MethodPatch,
		Path:        "/admin/users/{id}/admin",
		Tags:        tags,
	}, adminUsersPatchAdmin)

	Register(api, huma.Operation{
		OperationID: "admin-users-patch-status",
		Summary:     "Set a user's status (admin)",
		Description: "Changes a user's account status without requiring them to log in. The body field is a pointer: omitting status leaves it unchanged. Moving the last remaining admin out of Active is refused with 400.",
		Method:      http.MethodPatch,
		Path:        "/admin/users/{id}/status",
		Tags:        tags,
	}, adminUsersPatchStatus)

	Register(api, huma.Operation{
		OperationID: "admin-users-set-password",
		Summary:     "Set a user's password (admin)",
		Description: "Sets a new password for a local account without requiring the current one, then invalidates all of the user's sessions. Accounts managed by a third-party authentication provider are refused with 412.",
		Method:      http.MethodPatch,
		Path:        "/admin/users/{id}/password",
		Tags:        tags,
	}, adminUsersSetPassword)

	Register(api, huma.Operation{
		OperationID:   "admin-users-password-reset-email",
		Summary:       "Send a password-reset email (admin)",
		Description:   "Triggers the self-service password-reset email for a local account. Refused with 412 when no mailer is configured or when the account is managed by a third-party authentication provider.",
		Method:        http.MethodPost,
		Path:          "/admin/users/{id}/password-reset-email",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, adminUsersPasswordResetEmail)

	Register(api, huma.Operation{
		OperationID: "admin-users-delete",
		Summary:     "Delete a user (admin)",
		Description: "Deletes a user. With mode=now the user is removed immediately. With mode=scheduled (the default) the user is scheduled for deletion through the email-confirmation self-deletion flow. Deleting the last remaining admin is refused with 400.",
		Method:      http.MethodDelete,
		Path:        "/admin/users/{id}",
		Tags:        tags,
	}, adminUsersDelete)
}

func init() { AddRouteRegistrar(RegisterAdminUserRoutes) }

func adminOverview(_ context.Context, _ *struct{}) (*adminOverviewBody, error) {
	s := db.NewSession()
	defer s.Close()

	overview, err := models.BuildOverview(s)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &adminOverviewBody{Body: overview}, nil
}

func adminUsersCreate(_ context.Context, in *struct{ Body models.CreateUserBody }) (*adminUserBody, error) {
	s := db.NewSession()
	defer s.Close()

	newUser, err := models.CreateUserAsAdmin(s, &in.Body)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	providers, err := openid.GetAllProviders() //nolint:contextcheck // GetAllProviders reads a cached map; it takes no context, like the v1 admin handlers.
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &adminUserBody{Body: shared.NewAdminUser(newUser, providers)}, nil //nolint:contextcheck // OIDC provider init deliberately uses a background context — provider lifetime exceeds the request
}

func adminUsersPatchAdmin(_ context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body adminIsAdminPatchBody
}) (*adminUserBody, error) {
	if in.Body.IsAdmin == nil {
		return nil, translateDomainError(models.ErrInvalidData{Message: "is_admin is required"})
	}
	return adminCommitUser(func(s *xorm.Session) (*user.User, error) { //nolint:contextcheck // see adminCommitUser.
		return models.SetUserAdminFlag(s, in.ID, *in.Body.IsAdmin)
	})
}

func adminUsersPatchStatus(_ context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body adminStatusPatchBody
}) (*adminUserBody, error) {
	if in.Body.Status == nil {
		return nil, translateDomainError(models.ErrInvalidData{Message: "status is required"})
	}
	newStatus := *in.Body.Status
	if newStatus < user.StatusActive || newStatus > user.StatusAccountLocked {
		return nil, translateDomainError(models.ErrInvalidData{Message: "invalid status"})
	}
	return adminCommitUser(func(s *xorm.Session) (*user.User, error) { //nolint:contextcheck // see adminCommitUser.
		return models.SetUserStatusAsAdmin(s, in.ID, newStatus)
	})
}

func adminUsersSetPassword(_ context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body adminSetPasswordBody
}) (*adminUserBody, error) {
	return adminCommitUser(func(s *xorm.Session) (*user.User, error) { //nolint:contextcheck // see adminCommitUser.
		return models.SetUserPasswordAsAdmin(s, in.ID, in.Body.NewPassword)
	})
}

func adminUsersPasswordResetEmail(_ context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric ID of the user."`
}) (*messageBody, error) {
	// Checked here, not in the model action: RequestUserPasswordResetToken
	// silently skips the email when the mailer is off, which would read as success.
	if !config.MailerEnabled.GetBool() {
		return nil, huma.Error412PreconditionFailed("No mailer is configured on this instance, so no password-reset email can be sent.")
	}

	s := db.NewSession()
	defer s.Close()
	if err := models.RequestPasswordResetAsAdmin(s, in.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	out := &messageBody{}
	out.Body.Message = "A password-reset email was sent."
	return out, nil
}

func adminUsersDelete(_ context.Context, in *struct {
	ID   int64  `path:"id" doc:"The numeric ID of the user."`
	Mode string `query:"mode" doc:"'now' deletes immediately; 'scheduled' (the default) triggers the email-confirmation self-deletion flow."`
}) (*emptyBody, error) {
	mode := in.Mode
	if mode == "" {
		mode = "scheduled"
	}
	if mode != "now" && mode != "scheduled" {
		return nil, translateDomainError(models.ErrInvalidData{Message: "invalid mode, expected 'now' or 'scheduled'"})
	}

	s := db.NewSession()
	defer s.Close()
	if err := models.DeleteUserAsAdmin(s, in.ID, mode); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// adminCommitUser runs a user-returning admin action in its own transaction and
// renders the admin user view. The action does the load/guard/mutate against the
// session (shared with v1 via the models layer); this owns the commit and response.
func adminCommitUser(action func(s *xorm.Session) (*user.User, error)) (*adminUserBody, error) {
	s := db.NewSession()
	defer s.Close()

	target, err := action(s)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	providers, err := openid.GetAllProviders() //nolint:contextcheck // GetAllProviders reads a cached map; it takes no context, like the v1 admin handlers.
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &adminUserBody{Body: shared.NewAdminUser(target, providers)}, nil
}
