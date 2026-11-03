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

package openid

import (
	"context"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

type LinkedIdentity struct {
	user.Identity
	ProviderKey  string `json:"provider_key" readOnly:"true" doc:"The key of the configured provider this identity belongs to. Empty when no configured provider matches the issuer anymore."`
	ProviderName string `json:"provider_name" readOnly:"true" doc:"The display name of the configured provider. Empty when no configured provider matches the issuer anymore."`
}

type LinkIdentityRequest struct {
	Provider     string `json:"provider" doc:"The provider key as returned by /info."`
	Code         string `json:"code" doc:"The authorization code the provider redirected back with."`
	RedirectURL  string `json:"redirect_url" doc:"The redirect URL the authorization request was started with."`
	Password     string `json:"password" doc:"The current user's password."`
	TOTPPasscode string `json:"totp_passcode,omitempty" doc:"A current totp passcode. Required when the user has totp enabled."`
}

// An identity already linked to u is returned unchanged.
func LinkIdentity(ctx context.Context, s *xorm.Session, u *user.User, req *LinkIdentityRequest) (*LinkedIdentity, error) {
	// Before the exchange: the code is single-use, so a typo must not burn it.
	if err := user.Reauthenticate(ctx, s, u, req.Password, req.TOTPPasscode); err != nil {
		return nil, err
	}

	// Don't hold the transaction across the IdP round trips.
	if err := s.Rollback(); err != nil {
		return nil, err
	}

	provider, _, idToken, _, err := exchangeOidcTokens(ctx, &Callback{
		Code:        req.Code,
		RedirectURL: req.RedirectURL,
	}, req.Provider)
	if err != nil {
		return nil, err
	}

	if err := s.Begin(); err != nil {
		return nil, err
	}

	existing, found, err := user.GetIdentityByIssuerAndSubject(s, idToken.Issuer, idToken.Subject)
	if err != nil {
		return nil, err
	}
	if found {
		if existing.UserID != u.ID {
			return nil, &models.ErrUserIdentityLinkedToOtherUser{Issuer: idToken.Issuer}
		}
		return withProvider(existing, provider), nil
	}

	ownedByOIDCUser, err := user.OIDCUserExists(s, idToken.Issuer, idToken.Subject)
	if err != nil {
		return nil, err
	}
	if ownedByOIDCUser {
		return nil, &models.ErrUserIdentityLinkedToOtherUser{Issuer: idToken.Issuer}
	}

	hasIssuer, err := user.HasIdentityForIssuer(s, u.ID, idToken.Issuer)
	if err != nil {
		return nil, err
	}
	if hasIssuer {
		return nil, &models.ErrUserIdentityProviderAlreadyLinked{
			UserID: u.ID,
			Issuer: idToken.Issuer,
		}
	}

	newIdentity, err := user.CreateIdentity(s, u.ID, idToken.Issuer, idToken.Subject)
	if err != nil {
		return nil, err
	}

	return withProvider(newIdentity, provider), nil
}

// Another user's identity is reported as nonexistent.
func UnlinkIdentity(ctx context.Context, s *xorm.Session, u *user.User, identityID int64, password, passcode string) error {
	if err := user.Reauthenticate(ctx, s, u, password, passcode); err != nil {
		return err
	}

	deleted, err := user.DeleteIdentity(s, u.ID, identityID)
	if err != nil {
		return err
	}
	if !deleted {
		return &models.ErrUserIdentityDoesNotExist{ID: identityID}
	}
	return nil
}

func ListIdentities(s *xorm.Session, u *user.User) ([]*LinkedIdentity, error) {
	identities, err := user.GetIdentitiesByUserID(s, u.ID)
	if err != nil {
		return nil, err
	}

	if len(identities) == 0 {
		return []*LinkedIdentity{}, nil
	}

	providers, err := GetAllProviders()
	if err != nil {
		return nil, err
	}

	linked := make([]*LinkedIdentity, 0, len(identities))
	for _, identity := range identities {
		linked = append(linked, withProvider(identity, FindProviderByIssuer(providers, identity.Issuer)))
	}
	return linked, nil
}

func withProvider(identity *user.Identity, provider *Provider) *LinkedIdentity {
	linked := &LinkedIdentity{Identity: *identity}
	if provider != nil {
		linked.ProviderKey = provider.Key
		linked.ProviderName = provider.Name
	}
	return linked
}
