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

	"code.vikunja.io/api/pkg/health"
	"code.vikunja.io/api/pkg/modules/auth/openid"

	"github.com/danielgtaylor/huma/v2"
)

type healthBody struct {
	Body struct {
		Status          string                  `json:"status" enum:"OK,degraded" doc:"\"OK\" when the service and its dependencies are reachable, \"degraded\" when the service itself is healthy but at least one configured OpenID Connect provider is not available." example:"OK"`
		OpenIDProviders []openid.ProviderStatus `json:"openid_providers,omitempty" doc:"Availability of each configured OpenID Connect provider. Omitted when OpenID Connect authentication is not configured."`
	}
}

// RegisterHealthRoutes wires the public healthcheck endpoint onto the Huma API.
func RegisterHealthRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "health",
		Summary:     "Healthcheck",
		Description: "Reports whether the service and its dependencies (database, Redis if enabled) are reachable. Returns 200 with status \"OK\" when healthy, 500 otherwise. When OpenID Connect providers are configured, each provider's availability is reported too; an unavailable provider (typically because it was unreachable while Vikunja started) degrades the status but never fails the check, since initialization is retried every minute and a restart would not help. Public — no authentication required.",
		Method:      http.MethodGet,
		Path:        "/health",
		Tags:        []string{"service"},
		// Public: opt out of the globally-applied auth. The path is also listed
		// in unauthenticatedAPIPaths so the token middleware lets it through.
		Security: []map[string][]string{},
	}, healthcheck)
}

func init() { AddRouteRegistrar(RegisterHealthRoutes) }

//nolint:contextcheck // health.Check and openid.GetAllProviders are shared v1/v2 code; they take no context and use background contexts for their own pings.
func healthcheck(_ context.Context, _ *struct{}) (*healthBody, error) {
	if err := health.Check(); err != nil {
		// Mirror v1: a failed check is an internal error; the cause is logged,
		// not leaked to the client.
		return nil, huma.Error500InternalServerError("Internal server error", err)
	}
	out := &healthBody{}
	out.Body.Status = "OK"
	out.Body.OpenIDProviders = openid.GetProvidersStatus()
	for _, p := range out.Body.OpenIDProviders {
		if !p.Available {
			out.Body.Status = "degraded"
			break
		}
	}
	return out, nil
}
