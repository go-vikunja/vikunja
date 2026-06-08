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
type webhookListBody struct {
	Body Paginated[*models.Webhook]
}

// RegisterWebhookRoutes wires the nested project-webhook CRUD onto the Huma API.
// Project webhooks are gated by the webhooks.enabled config flag; the check runs
// here (not at init()) because RegisterAll fires after config is loaded. There is
// deliberately no ReadOne — webhooks carry secrets, so v1 never exposed a
// single-fetch route and v2 keeps that. Without a GET-one, AutoPatch synthesises
// no PATCH for this resource, so update is PUT only.
func RegisterWebhookRoutes(api huma.API) {
	if !config.WebhooksEnabled.GetBool() {
		return
	}

	tags := []string{"webhooks"}

	Register(api, huma.Operation{
		OperationID: "webhooks-list",
		Summary:     "List a project's webhooks",
		Description: "Returns the webhook targets configured for the given project, paginated. Requires read access to the project. Secret and basic-auth credentials are never included.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/webhooks",
		Tags:        tags,
	}, webhooksList)

	Register(api, huma.Operation{
		OperationID: "webhooks-create",
		Summary:     "Create a webhook target in a project",
		Description: "Creates a webhook target that receives POST requests about the subscribed events of the given project. The parent project is taken from the URL, not the body. Requires write access to the project. The secret and basic-auth credentials are write-only and not returned in the response.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/webhooks",
		Tags:        tags,
	}, webhooksCreate)

	Register(api, huma.Operation{
		OperationID: "webhooks-update",
		Summary:     "Update a webhook target's events",
		Description: "Changes the events a webhook target subscribes to. Only the events list can be changed; target_url, secret and auth are immutable after creation. The webhook must belong to the project in the path, and write access to that project is required.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/webhooks/{webhook}",
		Tags:        tags,
	}, webhooksUpdate)

	Register(api, huma.Operation{
		OperationID: "webhooks-delete",
		Summary:     "Delete a webhook target",
		Description: "Deletes a webhook target. The webhook must belong to the project in the path, and write access to that project is required.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/webhooks/{webhook}",
		Tags:        tags,
	}, webhooksDelete)
}

func init() { AddRouteRegistrar(RegisterWebhookRoutes) }

func webhooksList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*webhookListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Webhook{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.Webhook)
	if !ok {
		return nil, fmt.Errorf("webhooks.ReadAll returned unexpected type %T (expected []*models.Webhook)", result)
	}
	return &webhookListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func webhooksCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.Webhook
}) (*singleBody[models.Webhook], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Webhook]{Body: &in.Body}, nil
}

func webhooksUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"webhook"`
	Body      models.Webhook
}) (*singleBody[models.Webhook], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	// Only the events list is persisted (Webhook.Update writes Cols("events")),
	// but the id and parent must be set so the permission check resolves the
	// right webhook under the right project.
	in.Body.ID = in.ID
	in.Body.ProjectID = in.ProjectID
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Webhook]{Body: &in.Body}, nil
}

func webhooksDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"webhook"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Webhook{ID: in.ID, ProjectID: in.ProjectID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
