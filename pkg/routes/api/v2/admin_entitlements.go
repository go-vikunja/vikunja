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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/entitlement"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

type adminEntitlementsBody struct {
	Body struct {
		Entitlements map[entitlement.Feature]int64 `json:"entitlements" doc:"Raw per-user rows keyed by feature name. Flags are 0/1, limits are the maximum. A missing feature means no restriction."`
	}
}

// Permissions are enforced by the gateV2AdminRoutes path middleware, not per-handler.
func RegisterAdminEntitlementRoutes(api huma.API) {
	tags := []string{"admin"}

	Register(api, huma.Operation{
		OperationID: "admin-users-entitlements-read",
		Summary:     "Get a user's entitlements (admin)",
		Description: "Returns the raw entitlement rows stored for the user, not the resolved view the user sees on GET /user. A feature missing from the map has no restriction. Restricted to instance admins on a licensed instance.",
		Method:      http.MethodGet,
		Path:        "/admin/users/{id}/entitlements",
		Tags:        tags,
	}, adminUsersEntitlementsRead)

	Register(api, huma.Operation{
		OperationID: "admin-users-entitlements-replace",
		Summary:     "Replace a user's entitlements (admin)",
		Description: "Replaces the full set of entitlement rows for the user: features missing from the body are deleted, so sending an empty map lifts every restriction. Idempotent. Unknown feature names and instance-wide features (admin_panel, audit_logs) are refused with 400. Restricted to instance admins on a licensed instance.",
		Method:      http.MethodPut,
		Path:        "/admin/users/{id}/entitlements",
		Tags:        tags,
	}, adminUsersEntitlementsReplace)
}

func init() { AddRouteRegistrar(RegisterAdminEntitlementRoutes) }

func adminUsersEntitlementsRead(_ context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric ID of the user."`
}) (*adminEntitlementsBody, error) {
	s := db.NewSession()
	defer s.Close()

	if _, err := user.GetUserByID(s, in.ID); err != nil {
		return nil, translateDomainError(err)
	}
	rows, err := entitlement.Rows(s, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &adminEntitlementsBody{}
	out.Body.Entitlements = rows
	return out, nil
}

func adminUsersEntitlementsReplace(_ context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the user."`
	Body struct {
		Entitlements map[entitlement.Feature]int64 `json:"entitlements" required:"true" doc:"The complete set of rows for this user, keyed by feature name. Flags take 0/1, limits take the maximum."`
	}
}) (*adminEntitlementsBody, error) {
	s := db.NewSession()
	defer s.Close()

	if _, err := user.GetUserByID(s, in.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := entitlement.Replace(s, in.ID, in.Body.Entitlements); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	out := &adminEntitlementsBody{}
	out.Body.Entitlements = in.Body.Entitlements
	return out, nil
}
