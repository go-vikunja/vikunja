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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// {entity} stays a string: Can{Create,Delete} derive the numeric EntityType
// from it and reject unknown kinds (412). The enum tag makes Huma reject
// anything else with a 422 before the handler runs.
type subscriptionPathParams struct {
	Entity   string `path:"entity" enum:"project,task" doc:"The kind of entity to (un)subscribe from. Either project or task."`
	EntityID int64  `path:"entityID" doc:"The numeric id of the entity to (un)subscribe from."`
}

func RegisterSubscriptionRoutes(api huma.API) {
	tags := []string{"subscriptions"}

	Register(api, huma.Operation{
		OperationID: "subscriptions-create",
		Summary:     "Subscribe to an entity",
		Description: "Subscribes the authenticated user to a project or task so they receive its notifications. The user needs read access to the entity. Fails if a subscription already exists.",
		Method:      http.MethodPost,
		Path:        "/subscriptions/{entity}/{entityID}",
		Tags:        tags,
	}, subscriptionsCreate)

	Register(api, huma.Operation{
		OperationID: "subscriptions-delete",
		Summary:     "Unsubscribe from an entity",
		Description: "Removes the authenticated user's own subscription to a project or task. Only affects the caller's subscription, not other users'.",
		Method:      http.MethodDelete,
		Path:        "/subscriptions/{entity}/{entityID}",
		Tags:        tags,
	}, subscriptionsDelete)
}

func init() { AddRouteRegistrar(RegisterSubscriptionRoutes) }

func subscriptionsCreate(ctx context.Context, in *subscriptionPathParams) (*singleBody[models.Subscription], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	sb := &models.Subscription{Entity: in.Entity, EntityID: in.EntityID}
	if err := handler.DoCreate(ctx, sb, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Subscription]{Body: sb}, nil
}

func subscriptionsDelete(ctx context.Context, in *subscriptionPathParams) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	sb := &models.Subscription{Entity: in.Entity, EntityID: in.EntityID}
	if err := handler.DoDelete(ctx, sb, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
