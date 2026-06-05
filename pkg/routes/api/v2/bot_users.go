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
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

type botUserListBody struct {
	Body Paginated[*models.BotUser]
}

// RegisterBotUserRoutes wires bot-user CRUD onto the Huma API.
func RegisterBotUserRoutes(api huma.API) {
	tags := []string{"bots"}

	Register(api, huma.Operation{
		OperationID: "bots-list",
		Summary:     "List bot users",
		Description: "Returns only the bot users owned by the authenticated user. Bots owned by anyone else are never listed.",
		Method:      http.MethodGet,
		Path:        "/user/bots",
		Tags:        tags,
	}, botUsersList)

	Register(api, huma.Operation{
		OperationID: "bots-read",
		Summary:     "Get a bot user",
		Description: "Returns a single bot user. Only the owner may read it; otherwise the request is refused. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/user/bots/{bot}",
		Tags:        tags,
	}, botUsersRead)

	Register(api, huma.Operation{
		OperationID: "bots-create",
		Summary:     "Create a bot user",
		Description: "Creates a bot user owned by the authenticated user. The username must start with the 'bot-' prefix. Bots have no email or password and cannot create further bots. Requires a real user account — link shares cannot create bots.",
		Method:      http.MethodPost,
		Path:        "/user/bots",
		Tags:        tags,
	}, botUsersCreate)

	Register(api, huma.Operation{
		OperationID: "bots-update",
		Summary:     "Update a bot user",
		Description: "Updates an owned bot user's name, status, and username. Only the owner may update it. Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/user/bots/{bot}",
		Tags:        tags,
	}, botUsersUpdate)

	Register(api, huma.Operation{
		OperationID: "bots-delete",
		Summary:     "Delete a bot user",
		Description: "Permanently deletes an owned bot user and all data associated with it. Only the owner may delete it.",
		Method:      http.MethodDelete,
		Path:        "/user/bots/{bot}",
		Tags:        tags,
	}, botUsersDelete)
}

func init() { AddRouteRegistrar(RegisterBotUserRoutes) }

func botUsersList(ctx context.Context, in *ListParams) (*botUserListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.BotUser{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.BotUser)
	if !ok {
		return nil, fmt.Errorf("bots.ReadAll returned unexpected type %T (expected []*models.BotUser)", result)
	}
	return &botUserListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

type botUserReadBody struct {
	models.BotUser
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this bot user (0=read, 1=read/write, 2=admin)."`
}

func botUsersRead(ctx context.Context, in *struct {
	ID int64 `path:"bot"`
	conditional.Params
}) (*singleReadBody[botUserReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	bot := &models.BotUser{}
	bot.ID = in.ID
	maxPermission, err := handler.DoReadOne(ctx, bot, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &botUserReadBody{BotUser: *bot, MaxPermission: models.Permission(maxPermission)}
	return conditionalReadResponse(&in.Params, body, bot.Updated, maxPermission)
}

func botUsersCreate(ctx context.Context, in *struct {
	Body models.BotUser
}) (*singleBody[models.BotUser], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.BotUser]{Body: &in.Body}, nil
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func botUsersUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"bot"`
	Body botUserReadBody
}) (*singleBody[models.BotUser], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	bot := &in.Body.BotUser
	bot.ID = in.ID // URL wins over body
	if err := handler.DoUpdate(ctx, bot, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.BotUser]{Body: bot}, nil
}

func botUsersDelete(ctx context.Context, in *struct {
	ID int64 `path:"bot"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	bot := &models.BotUser{}
	bot.ID = in.ID
	if err := handler.DoDelete(ctx, bot, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
