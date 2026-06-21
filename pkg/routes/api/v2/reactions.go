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

// {entitykind} stays a string: the model derives the numeric EntityKind from
// it and rejects unknown kinds. The enum tag (repeated on the create/delete
// inputs) makes Huma reject anything else with a 422 before the handler runs;
// keep the values in sync with models.Reaction.setEntityKindFromString.
type reactionPathParams struct {
	EntityKind string `path:"entitykind" enum:"tasks,comments" doc:"The kind of entity being reacted to. Either tasks or comments (task comments)."`
	EntityID   int64  `path:"entityid" doc:"The numeric id of the entity being reacted to."`
}

// Reactions list as a map keyed by reaction value, not a slice, so it does not
// fit the Paginated envelope.
type reactionListBody struct {
	Body models.ReactionMap
}

func RegisterReactionRoutes(api huma.API) {
	tags := []string{"reactions"}

	Register(api, huma.Operation{
		OperationID: "reactions-list",
		Summary:     "List reactions for an entity",
		Description: "Returns every reaction on the entity, grouped as a map keyed by reaction value; each value maps to the users who reacted with it. Requires read access to the entity. Not paginated.",
		Method:      http.MethodGet,
		Path:        "/{entitykind}/{entityid}/reactions",
		Tags:        tags,
	}, reactionsList)

	Register(api, huma.Operation{
		OperationID: "reactions-create",
		Summary:     "React to an entity",
		Description: "Adds the authenticated user's reaction to the entity. Requires write access. No-op if the same reaction already exists.",
		Method:      http.MethodPost,
		Path:        "/{entitykind}/{entityid}/reactions",
		Tags:        tags,
	}, reactionsCreate)

	Register(api, huma.Operation{
		OperationID:   "reactions-delete",
		Summary:       "Remove a reaction from an entity",
		Description:   "Removes the authenticated user's own reaction from the entity. The reaction to remove is named in the body (there is no per-reaction id), so this is a POST with a body rather than a DELETE. Requires write access.",
		Method:        http.MethodPost,
		Path:          "/{entitykind}/{entityid}/reactions/delete",
		Tags:          tags,
		DefaultStatus: http.StatusOK,
	}, reactionsDelete)
}

func init() { AddRouteRegistrar(RegisterReactionRoutes) }

func reactionsList(ctx context.Context, in *reactionPathParams) (*reactionListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	r := &models.Reaction{EntityID: in.EntityID, EntityKindString: in.EntityKind}
	result, _, _, err := handler.DoReadAll(ctx, r, a, "", 1, -1)
	if err != nil {
		return nil, translateDomainError(err)
	}
	reactions, ok := result.(models.ReactionMap)
	if !ok {
		return nil, fmt.Errorf("reactions.ReadAll returned unexpected type %T (expected models.ReactionMap)", result)
	}
	if reactions == nil {
		reactions = models.ReactionMap{}
	}
	return &reactionListBody{Body: reactions}, nil
}

// Path params are flattened (not via the embedded reactionPathParams) because
// Huma fails to bind an embedded path-param struct when the input also has a Body.
func reactionsCreate(ctx context.Context, in *struct {
	EntityKind string `path:"entitykind" enum:"tasks,comments" doc:"The kind of entity being reacted to. Either tasks or comments (task comments)."`
	EntityID   int64  `path:"entityid" doc:"The numeric id of the entity being reacted to."`
	Body       models.Reaction
}) (*singleBody[models.Reaction], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	r := &in.Body
	r.EntityID = in.EntityID
	r.EntityKindString = in.EntityKind
	if err := handler.DoCreate(ctx, r, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Reaction]{Body: r}, nil
}

func reactionsDelete(ctx context.Context, in *struct {
	EntityKind string `path:"entitykind" enum:"tasks,comments" doc:"The kind of entity being reacted to. Either tasks or comments (task comments)."`
	EntityID   int64  `path:"entityid" doc:"The numeric id of the entity being reacted to."`
	Body       models.Reaction
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	r := &in.Body
	r.EntityID = in.EntityID
	r.EntityKindString = in.EntityKind
	if err := handler.DoDelete(ctx, r, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
