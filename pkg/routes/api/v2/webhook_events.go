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
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type webhookEventsBody struct {
	Body []string `json:"events" doc:"The events a webhook target can subscribe to."`
}

// RegisterWebhookEventRoutes wires the available-webhook-events listing onto the
// Huma API. Like v1, the whole endpoint only exists when webhooks are enabled.
func RegisterWebhookEventRoutes(api huma.API) {
	if !config.WebhooksEnabled.GetBool() {
		return
	}

	Register(api, huma.Operation{
		OperationID: "webhooks-events-list",
		Summary:     "List available webhook events",
		Description: "Returns every event a webhook target can subscribe to. Use these values when creating or updating a webhook.",
		Method:      http.MethodGet,
		Path:        "/webhooks/events",
		Tags:        []string{"webhooks"},
	}, webhookEventsList)
}

func init() { AddRouteRegistrar(RegisterWebhookEventRoutes) }

func webhookEventsList(_ context.Context, _ *struct{}) (*webhookEventsBody, error) {
	return &webhookEventsBody{Body: models.GetAvailableWebhookEvents()}, nil
}
