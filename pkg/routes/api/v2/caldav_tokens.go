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

	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

// CalDAV tokens are scoped to the authenticated user, not a CRUDable resource:
// there is no per-token Can* method, so these handlers own their own user lookup
// (user.GetFromAuth refuses link shares) and session/commit lives in the user package.

type caldavTokenListBody struct {
	Body Paginated[*user.Token]
}

type caldavTokenBody struct {
	Body *user.Token
}

// RegisterCalDAVTokenRoutes wires the current user's CalDAV token operations onto the Huma API.
func RegisterCalDAVTokenRoutes(api huma.API) {
	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID: "caldav-tokens-create",
		Summary:     "Generate a CalDAV token",
		Description: "Generates a CalDAV token for the authenticated user. The clear-text token is returned only in this response and can never be retrieved again. Link shares cannot have CalDAV tokens.",
		Method:      http.MethodPost,
		Path:        "/user/settings/token/caldav",
		Tags:        tags,
	}, caldavTokensCreate)

	Register(api, huma.Operation{
		OperationID: "caldav-tokens-list",
		Summary:     "List CalDAV tokens",
		Description: "Returns the authenticated user's CalDAV tokens. Only the id and creation date are returned — never the token value, which is shown once on creation.",
		Method:      http.MethodGet,
		Path:        "/user/settings/token/caldav",
		Tags:        tags,
	}, caldavTokensList)

	Register(api, huma.Operation{
		OperationID: "caldav-tokens-delete",
		Summary:     "Delete a CalDAV token",
		Description: "Deletes one of the authenticated user's CalDAV tokens by id. Tokens of other users are out of scope and cannot be deleted.",
		Method:      http.MethodDelete,
		Path:        "/user/settings/token/caldav/{id}",
		Tags:        tags,
	}, caldavTokensDelete)
}

func init() { AddRouteRegistrar(RegisterCalDAVTokenRoutes) }

func caldavTokensCreate(ctx context.Context, _ *struct{}) (*caldavTokenBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	token, err := user.GenerateNewCaldavToken(u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &caldavTokenBody{Body: token}, nil
}

func caldavTokensList(ctx context.Context, in *ListParams) (*caldavTokenListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	tokens, err := user.GetCaldavTokens(u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &caldavTokenListBody{Body: NewPaginated(tokens, int64(len(tokens)), in.Page, in.PerPage)}, nil
}

func caldavTokensDelete(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the CalDAV token to delete."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err := user.DeleteCaldavTokenByID(u, in.ID); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
