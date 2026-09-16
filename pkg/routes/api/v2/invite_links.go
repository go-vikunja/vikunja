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
	"github.com/danielgtaylor/huma/v2"
)

type publicInviteLink struct {
	Name             string                  `json:"name" doc:"Name of this invitation."`
	Teams            []models.InviteLinkTeam `json:"teams" doc:"Teams the new account will join."`
	SkipEmailConfirm bool                    `json:"skip_email_confirm" doc:"Whether the link activates accounts without email confirmation."`
}

type inviteLinkCheckBody struct {
	Token string `json:"token" doc:"Secret invitation token."`
}

func init() { AddRouteRegistrar(RegisterPublicInviteLinkRoutes) }

func RegisterPublicInviteLinkRoutes(api huma.API) {
	if !config.AuthLocalEnabled.GetBool() {
		return
	}
	Register(api, huma.Operation{OperationID: "invite-links-check", Summary: "Check an invitation", Description: "Available without authentication, even when public registration is disabled. Invalid, expired, exhausted and unlicensed links all return 404.", Method: http.MethodPost, Path: "/invite-links/check", DefaultStatus: http.StatusOK, Tags: []string{"auth"}, Security: publicSecurity}, inviteLinksCheck)
}

func inviteLinksCheck(_ context.Context, in *struct {
	Body inviteLinkCheckBody
}) (*singleBody[publicInviteLink], error) {
	s := db.NewSession()
	defer s.Close()
	link, err := models.GetInviteLinkByToken(s, in.Body.Token)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[publicInviteLink]{Body: &publicInviteLink{Name: link.Name, Teams: link.Teams, SkipEmailConfirm: link.SkipEmailConfirm}}, nil
}
