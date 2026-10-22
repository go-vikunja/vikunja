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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"github.com/danielgtaylor/huma/v2"
)

type inviteLinkListBody struct {
	Body Paginated[*models.UserInviteLink]
}

func init() { AddRouteRegistrar(RegisterAdminInviteLinkRoutes) }

func RegisterAdminInviteLinkRoutes(api huma.API) {
	Register(api, huma.Operation{OperationID: "admin-invite-links-list", Summary: "List invite links", Description: "Requires an instance admin and the admin_panel and user_invites features. Secret tokens are never returned.", Method: http.MethodGet, Path: "/admin/invite-links", Tags: []string{"admin"}}, adminInviteLinksList)
	Register(api, huma.Operation{OperationID: "admin-invite-links-create", Summary: "Create an invite link", Description: "Requires an instance admin and both licensed features. Returns the secret token once. Invitees can register even when public registration is disabled.", Method: http.MethodPost, Path: "/admin/invite-links", Tags: []string{"admin"}}, adminInviteLinksCreate)
	Register(api, huma.Operation{OperationID: "admin-invite-links-delete", Summary: "Delete an invite link", Description: "Requires an instance admin and both licensed features. Prevents further use without affecting users who already registered.", Method: http.MethodDelete, Path: "/admin/invite-links/{id}", Tags: []string{"admin"}}, adminInviteLinksDelete)
}

func adminInviteLinksList(_ context.Context, in *ListParams) (*inviteLinkListBody, error) {
	s := db.NewSession()
	defer s.Close()
	links, total, err := models.ListInviteLinksAsAdmin(s, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &inviteLinkListBody{Body: NewPaginated(links, total, in.Page, in.PerPage)}, nil
}

func adminInviteLinksCreate(ctx context.Context, in *struct{ Body models.CreateInviteLinkBody }) (*singleBody[models.UserInviteLink], error) {
	doer, err := adminDoerFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	link, err := models.CreateInviteLinkAsAdmin(s, doer, &in.Body)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &singleBody[models.UserInviteLink]{Body: link}, nil
}

func adminInviteLinksDelete(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"Numeric invite link ID."`
}) (*emptyBody, error) {
	doer, err := adminDoerFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	if err := models.DeleteInviteLinkAsAdmin(s, doer, in.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &emptyBody{}, nil
}
