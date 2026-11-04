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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/auth/openid"

	"github.com/danielgtaylor/huma/v2"
)

type userIdentityListBody struct {
	Body Paginated[*openid.LinkedIdentity]
}

func init() { AddRouteRegistrar(RegisterUserIdentityRoutes) }

func RegisterUserIdentityRoutes(api huma.API) {
	if !config.AuthOpenIDEnabled.GetBool() {
		return
	}

	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID: "user-identities-list",
		Summary:     "List linked identities",
		Description: "Returns the OpenID Connect identities linked to the authenticated user. The provider is resolved from the identity's issuer; provider_key and provider_name are empty when no configured provider matches it anymore. Not paginated: page and per_page always describe a single page.",
		Method:      http.MethodGet,
		Path:        "/user/settings/identities",
		Tags:        tags,
	}, userIdentitiesList)

	Register(api, huma.Operation{
		OperationID: "user-identities-link",
		Summary:     "Link an identity",
		Description: "Links the OpenID Connect identity behind an authorization code to the authenticated user, so they can log in through that provider. Requires the current password, plus a totp passcode when totp is enabled; failures count toward the account lockout and are throttled per user (429). These are checked before the code is exchanged, so a rejected attempt can be retried with the same code. Relinking an identity already linked to the user returns it unchanged. Fails with 409 if the identity belongs to another user or the user already linked this provider. Local accounts only.",
		Method:      http.MethodPost,
		Path:        "/user/settings/identities",
		Tags:        tags,
	}, userIdentitiesLink)

	Register(api, huma.Operation{
		OperationID:   "user-identities-unlink",
		Summary:       "Unlink an identity",
		Description:   "Removes a linked identity. Requires the current password, plus a totp passcode when totp is enabled; failures count toward the account lockout and are throttled per user (429). Another user's identity is a 404. Local accounts only.",
		Method:        http.MethodPost,
		Path:          "/user/settings/identities/{identity}/unlink",
		DefaultStatus: http.StatusNoContent,
		Tags:          tags,
	}, userIdentitiesUnlink)
}

func userIdentitiesList(ctx context.Context, _ *struct{}) (*userIdentityListBody, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	identities, err := openid.ListIdentities(s, u) //nolint:contextcheck // provider discovery runs on its own background context.
	if err != nil {
		return nil, translateDomainError(err)
	}
	total := int64(len(identities))
	return &userIdentityListBody{Body: NewPaginated(identities, total, 1, max(len(identities), 1))}, nil
}

func userIdentitiesLink(ctx context.Context, in *struct {
	Body openid.LinkIdentityRequest
}) (*singleBody[openid.LinkedIdentity], error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	defer events.CleanupPending(s)

	identity, err := openid.LinkIdentity(ctx, s, u, &in.Body)
	if err != nil {
		_ = s.Rollback()
		return nil, translateOpenIDError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &singleBody[openid.LinkedIdentity]{Body: identity}, nil
}

func userIdentitiesUnlink(ctx context.Context, in *struct {
	Identity int64 `path:"identity" doc:"The id of the linked identity."`
	Body     struct {
		Password     string `json:"password" doc:"The current user's password."`
		TOTPPasscode string `json:"totp_passcode,omitempty" doc:"A current totp passcode. Required when the user has totp enabled."`
	}
}) (*emptyBody, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	if err := openid.UnlinkIdentity(ctx, s, u, in.Identity, in.Body.Password, in.Body.TOTPPasscode); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
