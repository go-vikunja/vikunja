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
)

type sessionListBody struct {
	Body Paginated[*models.Session]
}

// RegisterSessionRoutes wires the session list/delete operations onto the Huma API.
// Sessions are created by the login flow, not by CRUD, so there is no create,
// read-one, or update — hence no max_permission or AutoPatch round trip.
func RegisterSessionRoutes(api huma.API) {
	tags := []string{"sessions"}

	Register(api, huma.Operation{
		OperationID: "sessions-list",
		Summary:     "List sessions",
		Description: "Returns the authenticated user's own active sessions, most recently active first. Never lists other users' sessions; link share tokens are forbidden.",
		Method:      http.MethodGet,
		Path:        "/user/sessions",
		Tags:        tags,
	}, sessionsList)

	Register(api, huma.Operation{
		OperationID: "sessions-delete",
		Summary:     "Delete a session",
		Description: "Revokes a session by its UUID. Only the owning user may delete it; deleting another user's session is forbidden.",
		Method:      http.MethodDelete,
		Path:        "/user/sessions/{session}",
		Tags:        tags,
	}, sessionsDelete)
}

func init() { AddRouteRegistrar(RegisterSessionRoutes) }

func sessionsList(ctx context.Context, in *ListParams) (*sessionListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Session{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.Session)
	if !ok {
		return nil, fmt.Errorf("sessions.ReadAll returned unexpected type %T (expected []*models.Session)", result)
	}
	return &sessionListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

// The session path param is a string UUID, not an int64 id.
func sessionsDelete(ctx context.Context, in *struct {
	Session string `path:"session" doc:"The UUID of the session to delete."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Session{ID: in.Session}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
