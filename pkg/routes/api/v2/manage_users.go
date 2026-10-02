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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

// The people area. Access is decided by the gateV2ManageRoutes path middleware (instance admin
// flag only, no license, no API tokens), not per handler.

type manageUserBody struct {
	Body *models.ManagedUser
}

type manageUserListBody struct {
	Body Paginated[*models.ManagedUser]
}

type manageProjectsBody struct {
	Body struct {
		Items []*models.TransferableProject `json:"items" doc:"The projects the user owns, archived ones included."`
	}
}

type manageTransferBody struct {
	Body struct {
		Transferred int `json:"transferred" doc:"The number of projects that were handed over."`
	}
}

// manageProfileBody is the editable part of a local user's profile. The username, job title and
// department are not part of it. There is no password here either: use the password endpoint.
type manageProfileBody struct {
	Name     string `json:"name" maxLength:"250" doc:"The display name."`
	Email    string `json:"email" minLength:"3" maxLength:"250" doc:"The email address. Takes effect at once, and replaces a change the user left unconfirmed."`
	Language string `json:"language,omitempty" doc:"IETF BCP 47 language code that exists in this installation. Omit to keep the current one."`
}

type manageCreateUserBody struct {
	models.CreateUserBody
	RequireChange *bool `json:"require_change,omitempty" doc:"Make the user choose a new password at the first login. Defaults to true."`
}

type manageSetPasswordBody struct {
	NewPassword   string `json:"new_password" valid:"bcrypt_password" minLength:"8" maxLength:"72" doc:"The new password. Max 72 bytes (a bcrypt limit), which may be fewer than 72 characters."`
	RequireChange *bool  `json:"require_change,omitempty" doc:"Make the user choose a new password at the next login. Defaults to true, since somebody else chose this one."`
}

type manageTransferRequest struct {
	NewOwnerID int64   `json:"new_owner_id" minimum:"1" doc:"The user who becomes the owner."`
	ProjectIDs []int64 `json:"project_ids,omitempty" doc:"The projects to hand over. Omit to hand over every project the user owns."`
}

type manageUsersListParams struct {
	Page    int    `query:"page" default:"1" minimum:"1" doc:"1-based page number."`
	PerPage int    `query:"per_page" minimum:"1" doc:"Items per page. Defaults to and is capped at the instance's configured maximum."`
	Q       string `query:"q" doc:"Matches username, name and email."`
	Status  *int   `query:"status" minimum:"0" maximum:"3" doc:"Only users with this status (0=active, 1=email-confirmation required, 2=disabled, 3=locked)."`
	Source  string `query:"source" enum:"local,entra,import,ldap,other" doc:"Only users from this source. 'import' are people the user list import created who have not logged in yet."`
}

func boolOrDefault(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func init() { AddRouteRegistrar(RegisterManageUserRoutes) }

// RegisterManageUserRoutes registers the user operations of the people area.
func RegisterManageUserRoutes(api huma.API) {
	tags := []string{"manage"}

	Register(api, huma.Operation{
		OperationID: "manage-users-list",
		Summary:     "List users (people area)",
		Description: "Lists every user except bots, paginated, with q matched against username, name and email. Each row says where the user comes from (local, entra, import for people who have not logged in yet, ldap, other) and whether their profile can be edited. Available to instance admins without an admin panel license; everyone else gets a 404. API tokens are refused with 404.",
		Method:      http.MethodGet,
		Path:        "/manage/users",
		Tags:        tags,
	}, manageUsersList)

	Register(api, huma.Operation{
		OperationID: "manage-users-read",
		Summary:     "Get a user (people area)",
		Description: "Returns one user with their login source and import id.",
		Method:      http.MethodGet,
		Path:        "/manage/users/{id}",
		Tags:        tags,
	}, manageUsersRead)

	Register(api, huma.Operation{
		OperationID: "manage-users-create",
		Summary:     "Create a local user (people area)",
		Description: "Creates a local account, bypassing the public-registration toggle. By default the user has to choose a new password at the first login (require_change).",
		Method:      http.MethodPost,
		Path:        "/manage/users",
		Tags:        tags,
	}, manageUsersCreate)

	Register(api, huma.Operation{
		OperationID: "manage-users-update",
		Summary:     "Update a user's profile (people area)",
		Description: "Replaces name, email and language of a LOCAL user and changes nothing else. Users whose profile is managed by Microsoft Entra ID, LDAP, another login provider or the user list import are refused with 403 (error code 1042). The username can never be changed, because rich text refers to users by username. An email change takes effect immediately.",
		Method:      http.MethodPut,
		Path:        "/manage/users/{id}",
		Tags:        tags,
	}, manageUsersUpdate)

	Register(api, huma.Operation{
		OperationID: "manage-users-patch-admin",
		Summary:     "Promote or demote a user (people area)",
		Description: "Sets the instance-admin flag. Omitting is_admin is refused. Demoting the last remaining admin is refused with 400.",
		Method:      http.MethodPatch,
		Path:        "/manage/users/{id}/admin",
		Tags:        tags,
	}, manageUsersPatchAdmin)

	Register(api, huma.Operation{
		OperationID: "manage-users-patch-status",
		Summary:     "Set a user's status (people area)",
		Description: "Activates, disables or locks an account. Moving the last remaining admin out of Active is refused with 400.",
		Method:      http.MethodPatch,
		Path:        "/manage/users/{id}/status",
		Tags:        tags,
	}, manageUsersPatchStatus)

	Register(api, huma.Operation{
		OperationID: "manage-users-set-password",
		Summary:     "Set a user's password (people area)",
		Description: "Sets a new password for a LOCAL account and ends all of the user's sessions. By default the user has to choose a new password at the next login. Accounts managed by a third-party provider are refused with 412.",
		Method:      http.MethodPatch,
		Path:        "/manage/users/{id}/password",
		Tags:        tags,
	}, manageUsersSetPassword)

	Register(api, huma.Operation{
		OperationID: "manage-users-delete",
		Summary:     "Delete a user (people area)",
		Description: "Deletes a user. With mode=now the user is removed immediately together with every project they own, so transfer the projects first. With mode=scheduled (the default) the email-confirmation self-deletion flow starts. You cannot delete yourself, and the last remaining admin cannot be deleted.",
		Method:      http.MethodDelete,
		Path:        "/manage/users/{id}",
		Tags:        tags,
	}, manageUsersDelete)

	Register(api, huma.Operation{
		OperationID: "manage-users-projects",
		Summary:     "List the projects a user owns (people area)",
		Description: "Returns every project the user owns, archived included. Use it to decide what to transfer before deleting a leaver.",
		Method:      http.MethodGet,
		Path:        "/manage/users/{id}/projects",
		Tags:        tags,
	}, manageUsersProjects)

	Register(api, huma.Operation{
		OperationID:   "manage-users-transfer-projects",
		Summary:       "Hand over a user's projects (people area)",
		Description:   "Makes another active user the owner of all, or of the selected, projects of this user, in one transaction: either every project moves or none does. Only projects change owner; saved filters, API tokens, webhooks and bots of the user stay with them. The new owner must be an active, non-bot user who is not scheduled for deletion.",
		Method:        http.MethodPost,
		Path:          "/manage/users/{id}/transfer-projects",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, manageUsersTransferProjects)
}

// manageDoer resolves the acting admin. The gate guarantees a user principal.
func manageDoer(ctx context.Context) (*user.User, error) {
	return adminDoerFromCtx(ctx)
}

// manageCommit runs fn in its own transaction, commits and dispatches the queued events, and rolls
// everything back when fn fails.
func manageCommit(ctx context.Context, fn func(s *xorm.Session, doer *user.User) error) error {
	doer, err := manageDoer(ctx)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	err = fn(s, doer)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return translateDomainError(err)
	}
	err = s.Commit()
	if err != nil {
		events.CleanupPending(s)
		return translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return nil
}

func manageUsersList(ctx context.Context, in *manageUsersListParams) (*manageUserListBody, error) {
	doer, err := manageDoer(ctx)
	if err != nil {
		return nil, err
	}

	perPage := in.PerPage
	if maxItems := config.ServiceMaxItemsPerPage.GetInt(); perPage <= 0 || perPage > maxItems {
		perPage = maxItems
	}
	status := -1
	if in.Status != nil {
		status = *in.Status
	}

	s := db.NewSession()
	defer s.Close()

	users, total, err := models.ListManagedUsers(s, doer, models.ManageUserFilter{Search: in.Q, Status: status, Source: in.Source}, in.Page, perPage)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	err = s.Commit()
	if err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)

	out := make([]*models.ManagedUser, 0, len(users))
	for _, u := range users {
		out = append(out, models.NewManagedUser(u))
	}
	return &manageUserListBody{Body: NewPaginated(out, total, in.Page, perPage)}, nil
}

func manageUsersRead(_ context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric ID of the user."`
}) (*manageUserBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := models.GetManagedUser(s, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &manageUserBody{Body: models.NewManagedUser(u)}, nil
}

func manageUsersCreate(ctx context.Context, in *struct{ Body manageCreateUserBody }) (*manageUserBody, error) {
	doer, err := manageDoer(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	// The forced change is written in the transaction that creates the account.
	created, err := models.CreateUserAsAdminWithOptions(s, doer, &in.Body.CreateUserBody, boolOrDefault(in.Body.RequireChange, true))
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	// CreateUserAsAdminWithOptions committed; the events it queued still have to go out.
	events.DispatchPending(ctx, s)
	// The reloaded user comes back without the email, the caller is an admin and sent it anyway.
	created.Email = in.Body.Email

	return &manageUserBody{Body: models.NewManagedUser(created)}, nil
}

func manageUsersUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body manageProfileBody
}) (*manageUserBody, error) {
	var out *user.User
	err := manageCommit(ctx, func(s *xorm.Session, doer *user.User) error {
		var err error
		out, err = models.UpdateUserProfileAsAdmin(s, doer, in.ID, models.ManageProfileUpdate{
			Name:     in.Body.Name,
			Email:    in.Body.Email,
			Language: in.Body.Language,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &manageUserBody{Body: models.NewManagedUser(out)}, nil
}

func manageUsersPatchAdmin(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body adminIsAdminPatchBody
}) (*manageUserBody, error) {
	if in.Body.IsAdmin == nil {
		return nil, translateDomainError(models.ErrInvalidData{Message: "is_admin is required"})
	}
	var out *user.User
	err := manageCommit(ctx, func(s *xorm.Session, doer *user.User) error {
		var err error
		out, err = models.SetUserAdminFlag(s, doer, in.ID, *in.Body.IsAdmin)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &manageUserBody{Body: models.NewManagedUser(out)}, nil
}

func manageUsersPatchStatus(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body adminStatusPatchBody
}) (*manageUserBody, error) {
	if in.Body.Status == nil {
		return nil, translateDomainError(models.ErrInvalidData{Message: "status is required"})
	}
	newStatus := *in.Body.Status
	if newStatus < user.StatusActive || newStatus > user.StatusAccountLocked {
		return nil, translateDomainError(models.ErrInvalidData{Message: "invalid status"})
	}
	var out *user.User
	err := manageCommit(ctx, func(s *xorm.Session, doer *user.User) error {
		var err error
		out, err = models.SetUserStatusAsAdmin(s, doer, in.ID, newStatus)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &manageUserBody{Body: models.NewManagedUser(out)}, nil
}

func manageUsersSetPassword(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body manageSetPasswordBody
}) (*manageUserBody, error) {
	var out *user.User
	err := manageCommit(ctx, func(s *xorm.Session, doer *user.User) error {
		var err error
		out, err = models.SetUserPasswordAsAdmin(s, doer, in.ID, in.Body.NewPassword)
		if err != nil {
			return err
		}
		// The password write clears the flag, so a forced change is set after it. Setting a
		// password ended every session, so the flag is effective at once.
		return user.SetMustChangePassword(s, out, boolOrDefault(in.Body.RequireChange, true))
	})
	if err != nil {
		return nil, err
	}
	return &manageUserBody{Body: models.NewManagedUser(out)}, nil
}

func manageUsersDelete(ctx context.Context, in *struct {
	ID   int64  `path:"id" doc:"The numeric ID of the user."`
	Mode string `query:"mode" enum:"now,scheduled" doc:"'now' deletes immediately, together with the user's projects. 'scheduled' (the default) triggers the email-confirmation self-deletion flow."`
}) (*emptyBody, error) {
	mode := in.Mode
	if mode == "" {
		mode = "scheduled"
	}

	err := manageCommit(ctx, func(s *xorm.Session, doer *user.User) error {
		if doer.ID == in.ID {
			return models.ErrInvalidData{Message: "you cannot delete your own account"}
		}
		return models.DeleteUserAsAdmin(s, doer, in.ID, mode)
	})
	if err != nil {
		return nil, err
	}
	return &emptyBody{}, nil
}

func manageUsersProjects(_ context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric ID of the user."`
}) (*manageProjectsBody, error) {
	s := db.NewSession()
	defer s.Close()

	projects, err := models.ListOwnedProjects(s, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &manageProjectsBody{}
	out.Body.Items = projects
	return out, nil
}

func manageUsersTransferProjects(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user whose projects are handed over."`
	Body manageTransferRequest
}) (*manageTransferBody, error) {
	moved := 0
	err := manageCommit(ctx, func(s *xorm.Session, doer *user.User) error {
		var err error
		moved, err = models.TransferProjectsAsAdmin(s, doer, in.ID, in.Body.NewOwnerID, in.Body.ProjectIDs)
		return err
	})
	if err != nil {
		return nil, err
	}
	out := &manageTransferBody{}
	out.Body.Transferred = moved
	return out, nil
}
