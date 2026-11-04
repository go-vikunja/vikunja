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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

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

type IdentityPreview struct {
	Account string `json:"account" doc:"The provider account being linked, from the verified ID token."`
}

type preparedIdentity struct {
	Issuer       string
	Subject      string
	ProviderKey  string
	ProviderName string
	Account      string
}

func identityPreviewKey(userID int64, req *LinkIdentityRequest) string {
	return "openid_identity_preview_" + utils.Sha256Hex(fmt.Sprintf("%d:%q:%q:%q", userID, req.Provider, req.RedirectURL, req.Code))
}

func cachedIdentityPreview(userID int64, req *LinkIdentityRequest) (prepared preparedIdentity, found bool, err error) {
	found, err = keyvalue.GetWithValue(identityPreviewKey(userID, req), &prepared)
	if err != nil {
		return prepared, false, err
	}
	if found && rawProviderConfigs()[req.Provider] == nil {
		return prepared, false, &models.ErrOpenIDBadRequest{Message: "Provider does not exist"}
	}
	return prepared, found, nil
}

func PreviewIdentity(ctx context.Context, u *user.User, req *LinkIdentityRequest) (*IdentityPreview, error) {
	prepared, found, err := cachedIdentityPreview(u.ID, req)
	if err != nil {
		return nil, err
	}
	if found {
		return previewIdentity(u, prepared)
	}

	provider, _, idToken, _, err := exchangeOidcTokens(ctx, &Callback{
		Code:        req.Code,
		RedirectURL: req.RedirectURL,
	}, req.Provider)
	if err != nil {
		return nil, err
	}
	var cl struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	if err := idToken.Claims(&cl); err != nil {
		return nil, err
	}
	account := idToken.Subject
	for _, label := range []string{cl.Email, cl.PreferredUsername, cl.Name} {
		if label != "" {
			account = label
			break
		}
	}
	prepared = preparedIdentity{
		Issuer:       idToken.Issuer,
		Subject:      idToken.Subject,
		ProviderKey:  provider.Key,
		ProviderName: provider.Name,
		Account:      account,
	}
	preview, err := previewIdentity(u, prepared)
	if err != nil {
		return nil, err
	}
	ttl := min(5*time.Minute, time.Until(idToken.Expiry))
	if err := keyvalue.PutWithTTL(identityPreviewKey(u.ID, req), prepared, ttl); err != nil {
		return nil, err
	}
	return preview, nil
}

func previewIdentity(u *user.User, prepared preparedIdentity) (*IdentityPreview, error) {
	s := db.NewSession()
	defer s.Close()
	if _, err := identityToLink(s, u.ID, prepared); err != nil {
		return nil, err
	}
	return &IdentityPreview{Account: prepared.Account}, nil
}

// An identity already linked to u is returned unchanged.
func LinkIdentity(ctx context.Context, s *xorm.Session, u *user.User, req *LinkIdentityRequest) (*LinkedIdentity, error) {
	// A typo must preserve both the authorization code and a prepared identity.
	if err := user.Reauthenticate(ctx, s, u, req.Password, req.TOTPPasscode); err != nil {
		return nil, err
	}

	// Don't hold the transaction across the IdP round trips.
	if err := s.Rollback(); err != nil {
		return nil, err
	}

	prepared, found, err := cachedIdentityPreview(u.ID, req)
	if err != nil {
		return nil, err
	}
	if found {
		// Consume before writing, so a later unlink cannot replay the preview.
		if err := keyvalue.Del(identityPreviewKey(u.ID, req)); err != nil {
			return nil, err
		}
	} else {
		provider, _, idToken, _, err := exchangeOidcTokens(ctx, &Callback{
			Code:        req.Code,
			RedirectURL: req.RedirectURL,
		}, req.Provider)
		if err != nil {
			return nil, err
		}
		prepared = preparedIdentity{
			Issuer:       idToken.Issuer,
			Subject:      idToken.Subject,
			ProviderKey:  provider.Key,
			ProviderName: provider.Name,
		}
	}
	provider := &Provider{
		Key:  prepared.ProviderKey,
		Name: prepared.ProviderName,
	}

	if err := s.Begin(); err != nil {
		return nil, err
	}

	existing, err := identityToLink(s, u.ID, prepared)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return withProvider(existing, provider), nil
	}

	newIdentity, err := user.CreateIdentity(s, u, prepared.Issuer, prepared.Subject, provider.Name)
	if err != nil {
		return nil, err
	}

	return withProvider(newIdentity, provider), nil
}

func identityToLink(s *xorm.Session, userID int64, prepared preparedIdentity) (*user.Identity, error) {
	existing, found, err := user.GetIdentityByIssuerAndSubject(s, prepared.Issuer, prepared.Subject)
	if err != nil {
		return nil, err
	}
	if found {
		if existing.UserID != userID {
			return nil, &models.ErrUserIdentityLinkedToOtherUser{Issuer: prepared.Issuer}
		}
		return existing, nil
	}

	ownedByOIDCUser, err := user.OIDCUserExists(s, prepared.Issuer, prepared.Subject)
	if err != nil {
		return nil, err
	}
	if ownedByOIDCUser {
		return nil, &models.ErrUserIdentityLinkedToOtherUser{Issuer: prepared.Issuer}
	}

	hasIssuer, err := user.HasIdentityForIssuer(s, userID, prepared.Issuer)
	if err != nil {
		return nil, err
	}
	if hasIssuer {
		return nil, &models.ErrUserIdentityProviderAlreadyLinked{
			UserID: userID,
			Issuer: prepared.Issuer,
		}
	}

	return nil, nil
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
