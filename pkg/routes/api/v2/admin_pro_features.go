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
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type proFeatureListBody struct {
	Body []*models.ProFeatureState
}

type userProFeatureListBody struct {
	Body []*models.UserProFeatureState
}

// Permissions are enforced by the gateV2AdminRoutes path middleware, not per-handler.
func RegisterAdminProFeatureRoutes(api huma.API) {
	tags := []string{"admin"}

	Register(api, huma.Operation{
		OperationID: "admin-pro-features-list",
		Summary:     "List pro features (admin)",
		Description: "Returns every pro feature with its license state and, for per-user toggleable features, the instance-wide default and where it comes from. Restricted to instance admins on a licensed instance; unlicensed or non-admin callers get a 404.",
		Method:      http.MethodGet,
		Path:        "/admin/pro-features",
		Tags:        tags,
	}, adminProFeaturesList)

	Register(api, huma.Operation{
		OperationID: "admin-pro-features-set-default",
		Summary:     "Set a pro feature's instance default (admin)",
		Description: "Sets the instance-wide default for a per-user toggleable feature. Users without a per-user override follow this default. Only per-user toggleable features accept a default; others yield a 422.",
		Method:      http.MethodPut,
		Path:        "/admin/pro-features/{feature}",
		Tags:        tags,
	}, adminProFeaturesSetDefault)

	Register(api, huma.Operation{
		OperationID: "admin-pro-features-reset-default",
		Summary:     "Reset a pro feature's instance default (admin)",
		Description: "Removes the admin-set instance default so the built-in code default applies again.",
		Method:      http.MethodDelete,
		Path:        "/admin/pro-features/{feature}",
		Tags:        tags,
	}, adminProFeaturesResetDefault)

	Register(api, huma.Operation{
		OperationID: "admin-user-pro-features-list",
		Summary:     "List a user's pro feature toggles (admin)",
		Description: "Returns the per-user toggleable pro features for the given user, with the admin-set override (null when the instance default applies) and the effective state including the license.",
		Method:      http.MethodGet,
		Path:        "/admin/users/{id}/pro-features",
		Tags:        tags,
	}, adminUserProFeaturesList)

	Register(api, huma.Operation{
		OperationID: "admin-user-pro-features-set",
		Summary:     "Set a user's pro feature override (admin)",
		Description: "Grants or revokes a per-user toggleable feature for the given user, overriding the instance default. Only per-user toggleable features accept an override; others yield a 422.",
		Method:      http.MethodPut,
		Path:        "/admin/users/{id}/pro-features/{feature}",
		Tags:        tags,
	}, adminUserProFeaturesSet)

	Register(api, huma.Operation{
		OperationID: "admin-user-pro-features-clear",
		Summary:     "Clear a user's pro feature override (admin)",
		Description: "Removes the per-user override so the instance default applies to this user again. Returns the user's refreshed toggle list.",
		Method:      http.MethodDelete,
		Path:        "/admin/users/{id}/pro-features/{feature}",
		// Override the wrapper's DELETE→204: the refreshed toggle list is returned.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, adminUserProFeaturesClear)
}

func init() { AddRouteRegistrar(RegisterAdminProFeatureRoutes) }

// perUserToggleableFeatureFromPath parses the {feature} path param and rejects
// features that cannot be managed per user.
func perUserToggleableFeatureFromPath(featureKey string) (license.Feature, error) {
	f, ok := license.FeatureFromString(featureKey)
	if !ok {
		return license.FeatureUnknown, huma.Error404NotFound("unknown feature " + featureKey)
	}
	if !license.IsPerUserToggleable(f) {
		return license.FeatureUnknown, huma.Error422UnprocessableEntity("feature " + featureKey + " is instance-wide and cannot be toggled per user")
	}
	return f, nil
}

func adminProFeaturesList(_ context.Context, _ *struct{}) (*proFeatureListBody, error) {
	s := db.NewSession()
	defer s.Close()

	states, err := models.GetProFeatureStates(s)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &proFeatureListBody{Body: states}, nil
}

func adminProFeaturesSetDefault(_ context.Context, in *struct {
	Feature string `path:"feature" doc:"The feature key, e.g. time_tracking."`
	Body    struct {
		DefaultEnabled bool `json:"default_enabled" doc:"Whether the feature should be enabled for users without a per-user override."`
	}
}) (*proFeatureListBody, error) {
	f, err := perUserToggleableFeatureFromPath(in.Feature)
	if err != nil {
		return nil, err
	}

	return commitProFeatureChange(func(s *xorm.Session) error {
		return models.SetProFeatureInstanceDefault(s, f, in.Body.DefaultEnabled)
	})
}

func adminProFeaturesResetDefault(_ context.Context, in *struct {
	Feature string `path:"feature" doc:"The feature key, e.g. time_tracking."`
}) (*emptyBody, error) {
	f, err := perUserToggleableFeatureFromPath(in.Feature)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	if err := models.ResetProFeatureInstanceDefault(s, f); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// commitProFeatureChange runs a write and returns the fresh feature list, the
// response shape both default-changing endpoints share.
func commitProFeatureChange(change func(s *xorm.Session) error) (*proFeatureListBody, error) {
	s := db.NewSession()
	defer s.Close()

	if err := change(s); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	states, err := models.GetProFeatureStates(s)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &proFeatureListBody{Body: states}, nil
}

func adminUserProFeaturesList(_ context.Context, in *struct {
	ID int64 `path:"id" doc:"The user id."`
}) (*userProFeatureListBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserByID(s, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	states, err := models.GetUserProFeatureStates(s, u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &userProFeatureListBody{Body: states}, nil
}

func adminUserProFeaturesSet(_ context.Context, in *struct {
	ID      int64  `path:"id" doc:"The user id."`
	Feature string `path:"feature" doc:"The feature key, e.g. time_tracking."`
	Body    struct {
		Enabled bool `json:"enabled" doc:"Whether the feature should be enabled for this user, regardless of the instance default."`
	}
}) (*userProFeatureListBody, error) {
	enabled := in.Body.Enabled
	return updateUserProFeatureOverride(in.ID, in.Feature, &enabled)
}

func adminUserProFeaturesClear(_ context.Context, in *struct {
	ID      int64  `path:"id" doc:"The user id."`
	Feature string `path:"feature" doc:"The feature key, e.g. time_tracking."`
}) (*userProFeatureListBody, error) {
	return updateUserProFeatureOverride(in.ID, in.Feature, nil)
}

func updateUserProFeatureOverride(userID int64, featureKey string, enabled *bool) (*userProFeatureListBody, error) {
	f, err := perUserToggleableFeatureFromPath(featureKey)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserByID(s, userID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err := models.SetUserProFeatureOverride(s, u, f, enabled); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	states, err := models.GetUserProFeatureStates(s, u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &userProFeatureListBody{Body: states}, nil
}
