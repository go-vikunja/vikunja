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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// models.Webhook.ReadAll returns []*models.Webhook, so that's the element type.
type userWebhookListBody struct {
	Body Paginated[*models.Webhook]
}

type userWebhookEventsBody struct {
	Body []string
}

// RegisterUserWebhookRoutes wires the per-user webhook CRUD onto the Huma API.
// User webhooks are the project-less sibling of the project webhooks (see
// webhooks.go): they fire across all of a user's projects and are owned by the
// user, not a project. Both resources share the webhooks.enabled gate; the check
// runs here (not at init()) because RegisterAll fires after config is loaded.
// Like project webhooks there is deliberately no ReadOne — webhooks carry
// credentials — so AutoPatch synthesises no PATCH and update is PUT only.
func RegisterUserWebhookRoutes(api huma.API) {
	if !config.WebhooksEnabled.GetBool() {
		return
	}

	tags := []string{"webhooks"}

	Register(api, huma.Operation{
		OperationID: "user-webhooks-list",
		Summary:     "List the current user's webhooks",
		Description: "Returns the webhook targets the authenticated user has configured for themselves (not project webhooks), paginated. Secret and basic-auth credentials are never included.",
		Method:      http.MethodGet,
		Path:        "/user/settings/webhooks",
		Tags:        tags,
	}, userWebhooksList)

	Register(api, huma.Operation{
		OperationID: "user-webhooks-events",
		Summary:     "List available user-directed webhook events",
		Description: "Returns the webhook event names a user-level webhook may subscribe to. This is a subset of the project webhook events — only events that target a single user.",
		Method:      http.MethodGet,
		Path:        "/user/settings/webhooks/events",
		Tags:        tags,
	}, userWebhooksEvents)

	Register(api, huma.Operation{
		OperationID: "user-webhooks-create",
		Summary:     "Create a webhook for the current user",
		Description: "Creates a webhook target owned by the authenticated user that receives POST requests across all of their projects. The owning user is taken from the token, not the body. May only subscribe to user-directed events (see the events route). The secret and basic-auth credentials are write-only and not returned in the response.",
		Method:      http.MethodPost,
		Path:        "/user/settings/webhooks",
		Tags:        tags,
	}, userWebhooksCreate)

	Register(api, huma.Operation{
		OperationID: "user-webhooks-update",
		Summary:     "Update a user webhook's events",
		Description: "Changes the events a user webhook subscribes to. Only the events list can be changed; target_url, secret and auth are immutable after creation. Only the owning user may update it.",
		Method:      http.MethodPut,
		Path:        "/user/settings/webhooks/{webhook}",
		Tags:        tags,
	}, userWebhooksUpdate)

	Register(api, huma.Operation{
		OperationID: "user-webhooks-delete",
		Summary:     "Delete a user webhook",
		Description: "Deletes a webhook owned by the authenticated user. Only the owning user may delete it.",
		Method:      http.MethodDelete,
		Path:        "/user/settings/webhooks/{webhook}",
		Tags:        tags,
	}, userWebhooksDelete)
}

func init() { AddRouteRegistrar(RegisterUserWebhookRoutes) }

func userWebhooksList(ctx context.Context, in *ListParams) (*userWebhookListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Webhook{UserID: a.GetID()}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.Webhook)
	if !ok {
		return nil, fmt.Errorf("webhooks.ReadAll returned unexpected type %T (expected []*models.Webhook)", result)
	}
	return &userWebhookListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func userWebhooksEvents(_ context.Context, _ *struct{}) (*userWebhookEventsBody, error) {
	return &userWebhookEventsBody{Body: models.GetUserDirectedWebhookEvents()}, nil
}

func userWebhooksCreate(ctx context.Context, in *struct {
	Body models.Webhook
}) (*singleBody[models.Webhook], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	// Force user ownership: a user webhook is keyed on the user, never a project.
	in.Body.UserID = a.GetID()
	in.Body.ProjectID = 0
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Webhook]{Body: &in.Body}, nil
}

func userWebhooksUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"webhook"`
	Body models.Webhook
}) (*singleBody[models.Webhook], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	// canDoWebhook resolves the owner from the stored row, so only the id is
	// needed to gate the update; the rest of the body's ownership fields are
	// ignored. Update persists only the events list.
	in.Body.ID = in.ID
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Webhook]{Body: &in.Body}, nil
}

func userWebhooksDelete(ctx context.Context, in *struct {
	ID int64 `path:"webhook"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Webhook{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
