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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/coreos/go-oidc/v3/oidc"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"xorm.io/xorm"
)

// Callback contains the callback after an auth request was made and redirected
type Callback struct {
	Code        string `query:"code" json:"code"`
	Scope       string `query:"scop" json:"scope"`
	RedirectURL string `json:"redirect_url"`
}

// Provider is the structure of an OpenID Connect provider
type Provider struct {
	Name             string `json:"name"`
	Key              string `json:"key"`
	OriginalAuthURL  string `json:"-"`
	AuthURL          string `json:"auth_url"`
	LogoutURL        string `json:"logout_url"`
	ClientID         string `json:"client_id"`
	Scope            string `json:"scope"`
	EmailFallback    bool   `json:"email_fallback"`
	UsernameFallback bool   `json:"username_fallback"`
	ForceUserInfo    bool   `json:"force_user_info"`
	ClientSecret     string `json:"-"`
	openIDProvider   *oidc.Provider
	Oauth2Config     *oauth2.Config `json:"-"`
}

type claims struct {
	Email              string                   `json:"email"`
	Name               string                   `json:"name"`
	PreferredUsername  string                   `json:"preferred_username"`
	Nickname           string                   `json:"nickname"`
	VikunjaGroups      []map[string]interface{} `json:"vikunja_groups"`
	Picture            string                   `json:"picture"`
	ExtraSettingsLinks map[string]any           `json:"extra_settings_links"`
}

func init() {
	petname.NonDeterministicMode()
}

func (p *Provider) setOicdProvider() (err error) {
	p.openIDProvider, err = oidc.NewProvider(context.Background(), p.OriginalAuthURL)
	return err
}

func (p *Provider) Issuer() (issuerURL string, err error) {
	type Issuer struct {
		Issuer string `json:"issuer"`
	}

	if p.openIDProvider == nil {
		err = p.setOicdProvider()
		if err != nil {
			return "", err
		}
	}

	iss := &Issuer{}
	err = p.openIDProvider.Claims(iss)
	if err != nil {
		return "", err
	}
	return iss.Issuer, nil
}

// HandleCallback handles the auth request callback after redirecting from the provider with an auth code
// @Summary Authenticate a user with OpenID Connect
// @Description After a redirect from the OpenID Connect provider to the frontend has been made with the authentication `code`, this endpoint can be used to obtain a jwt token for that user and thus log them in.
// @ID get-token-openid
// @tags auth
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param callback body openid.Callback true "The openid callback"
// @Param provider path int true "The OpenID Connect provider key as returned by the /info endpoint"
// @Success 200 {object} auth.Token
// @Failure 500 {object} models.Message "Internal error"
// @Router /auth/openid/{provider}/callback [post]
func HandleCallback(c echo.Context) error {

	provider, oauthToken, idToken, err := getProviderAndOidcTokens(c)
	if err != nil {
		var detailedErr *models.ErrOpenIDBadRequestWithDetails
		if errors.As(err, &detailedErr) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": detailedErr.Message,
				"details": detailedErr.Details,
			})
		}
		return handler.HandleHTTPError(err)
	}

	cl, err := getClaims(provider, oauthToken, idToken)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	s := db.NewSession()
	defer s.Close()

	// Check if we have seen this user before
	u, err := getOrCreateUser(s, cl, provider, idToken)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error creating new user for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err)
	}

	teamData := getTeamDataFromToken(cl.VikunjaGroups, provider)

	err = models.SyncExternalTeamsForUser(s, u, teamData, idToken.Issuer, "OIDC")
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error creating new team for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err)
	}

	// Create token
	return auth.NewUserAuthTokenResponse(u, c, false)
}

func getTeamDataFromToken(groups []map[string]interface{}, provider *Provider) (teamData []*models.Team) {
	teamData = []*models.Team{}
	for _, t := range groups {
		var name string
		var description string
		var oidcID string
		var isPublic bool

		// Read name
		_, exists := t["name"]
		if exists {
			name = t["name"].(string)
		}

		// Read description
		_, exists = t["description"]
		if exists {
			description = t["description"].(string)
		}

		// Read isPublic flag
		_, exists = t["isPublic"]
		if exists {
			isPublic = t["isPublic"].(bool)
		}

		// Read oidcID
		_, exists = t["oidcID"]
		if exists {
			switch id := t["oidcID"].(type) {
			case string:
				oidcID = id
			case int64:
				oidcID = strconv.FormatInt(id, 10)
			case float64:
				oidcID = strconv.FormatFloat(id, 'f', -1, 64)
			default:
				log.Errorf("No oidcID assigned for %v or type %v not supported", t, t)
			}
		}
		if name == "" || oidcID == "" {
			log.Errorf("Claim of your custom scope does not hold name or oidcID for automatic group assignment through oidc provider. Please check %s", provider.Name)
			continue
		}
		teamData = append(teamData, &models.Team{
			Name:        name,
			ExternalID:  oidcID,
			Description: description,
			IsPublic:    isPublic,
		})
	}

	return teamData
}

// Download and store a user's avatar from an OpenID provider
func syncUserAvatarFromOpenID(s *xorm.Session, u *user.User, pictureURL string) (err error) {
	// Don't sync avatar if no picture URL is provided
	if pictureURL == "" {
		return fmt.Errorf("no picture URL provided")
	}

	log.Debugf("Found avatar URL for user %s: %s", u.Username, pictureURL)

	// Download avatar
	avatarData, err := utils.DownloadImage(pictureURL)
	if err != nil {
		return fmt.Errorf("error downloading avatar: %w", err)
	}

	// Process avatar, ensure 1:1 ratio
	processedAvatar, err := utils.CropAvatarTo1x1(avatarData)
	if err != nil {
		return fmt.Errorf("error processing avatar: %w", err)
	}

	// Set avatar provider to openid
	u.AvatarProvider = "openid"

	// Store avatar and update user
	err = upload.StoreAvatarFile(s, u, bytes.NewReader(processedAvatar))
	if err != nil {
		return fmt.Errorf("error storing avatar: %w", err)
	}

	avatar.FlushAllCaches(u)

	return nil
}

func getOrCreateUser(s *xorm.Session, cl *claims, provider *Provider, idToken *oidc.IDToken) (u *user.User, err error) {

	// set defaults
	fallbackMatchFound := false
	alreadyCreatedFromIssuer := false

	// first check if the user already signed up using the provider

	u, err = user.GetUserWithEmail(s, &user.User{
		Issuer:  idToken.Issuer,
		Subject: idToken.Subject,
	})
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		return nil, err
	}
	alreadyCreatedFromIssuer = err == nil // found if no error, not found if we reach it here despite an error

	if !alreadyCreatedFromIssuer && (provider.EmailFallback || provider.UsernameFallback) {

		// try finding the user on fallback mappingproperties

		searchUser := &user.User{
			Issuer: user.IssuerLocal,
		}
		if provider.UsernameFallback {
			// Match oidc subject on username as each is unique identifier in its own referential
			// Discouraged if multiple account providers are used.
			searchUser.Username = idToken.Subject
		}
		if provider.EmailFallback {
			// Used alone, allow for someone to connect from various provider to the same account
			// Discouraged for untrusted provider where someone can set email without verification
			// Note : mapping on email prevent from auto-updating user email
			searchUser.Email = cl.Email
		}

		// Check if the user exists for the given fallback matching options
		u, err = user.GetUserWithEmail(s, searchUser)
		if err != nil && !user.IsErrUserDoesNotExist(err) {
			return nil, err
		}
		fallbackMatchFound = err == nil // found if no error, not found if we reach it here despite an error
	}

	if !alreadyCreatedFromIssuer && !fallbackMatchFound {

		// If no user exists, create one with the preferred username if it is not already taken
		uu := &user.User{
			Username:           strings.ReplaceAll(cl.PreferredUsername, " ", "-"),
			Email:              cl.Email,
			Name:               cl.Name,
			Status:             user.StatusActive,
			Issuer:             idToken.Issuer,
			Subject:            idToken.Subject,
			ExtraSettingsLinks: cl.ExtraSettingsLinks,
		}

		u, err = auth.CreateUserWithRandomUsername(s, uu)
		if err != nil {
			return nil, err
		}
	} else if alreadyCreatedFromIssuer {

		// try updating user.Name and/or user.Email if necessary
		if cl.Email != u.Email {
			u.Email = cl.Email
		}
		if cl.Name != u.Name {
			u.Name = cl.Name
		}

		u.ExtraSettingsLinks = cl.ExtraSettingsLinks

		u, err = user.UpdateUser(s, u, false)
		if err != nil {
			return nil, err
		}
	}

	// Try sync avatar if available
	err = syncUserAvatarFromOpenID(s, u, cl.Picture)
	if err != nil {
		log.Errorf("Error syncing avatar for user %s: %v", u.Username, err)
	}

	return u, nil
}

// mergeClaims combines claims from token and userinfo based on the ForceUserInfo setting
// cl represents the claims from the token, cl2 represents the claims from userinfo
func mergeClaims(cl *claims, cl2 *claims, forceUserInfo bool) error {
	if (forceUserInfo && cl2.Email != "") || cl.Email == "" {
		cl.Email = cl2.Email
	}

	if (forceUserInfo && cl2.Name != "") || cl.Name == "" {
		cl.Name = cl2.Name
	}

	if (forceUserInfo && cl2.PreferredUsername != "") || cl.PreferredUsername == "" {
		cl.PreferredUsername = cl2.PreferredUsername
	}

	if cl.PreferredUsername == "" && cl2.Nickname != "" {
		cl.PreferredUsername = cl2.Nickname
	}

	if (forceUserInfo && cl2.Picture != "") || cl.Picture == "" {
		cl.Picture = cl2.Picture
	}

	if cl.Email == "" {
		return &user.ErrNoOpenIDEmailProvided{}
	}

	return nil
}

func getClaims(provider *Provider, oauth2Token *oauth2.Token, idToken *oidc.IDToken) (*claims, error) {

	cl := &claims{}
	err := idToken.Claims(cl)
	if err != nil {
		log.Errorf("Error getting token claims for provider %s: %v", provider.Name, err)
		return nil, err
	}

	if provider.ForceUserInfo || cl.Email == "" || cl.Name == "" || cl.PreferredUsername == "" || cl.Picture == "" {
		info, err := provider.openIDProvider.UserInfo(context.Background(), provider.Oauth2Config.TokenSource(context.Background(), oauth2Token))
		if err != nil {
			log.Errorf("Error getting userinfo for provider %s: %v", provider.Name, err)
			return nil, err
		}

		cl2 := &claims{}
		err = info.Claims(cl2)
		if err != nil {
			log.Errorf("Error parsing userinfo claims for provider %s: %v", provider.Name, err)
			return nil, err
		}

		err = mergeClaims(cl, cl2, provider.ForceUserInfo)
		if err != nil {
			if user.IsErrNoEmailProvided(err) {
				log.Errorf("Claim does not contain an email address for provider %s", provider.Name)
			}

			return nil, err
		}
	}
	return cl, nil
}

func getProviderAndOidcTokens(c echo.Context) (*Provider, *oauth2.Token, *oidc.IDToken, error) {

	cb := &Callback{}
	if err := c.Bind(cb); err != nil {
		return nil, nil, nil, &models.ErrOpenIDBadRequest{Message: "Bad data"}
	}

	// Check if the provider exists
	providerKey := c.Param("provider")
	provider, err := GetProvider(providerKey)
	if err != nil {
		return nil, nil, nil, err
	}
	if provider == nil {
		return nil, nil, nil, &models.ErrOpenIDBadRequest{Message: "Provider does not exist"}
	}

	log.Debugf("Trying to authenticate user using provider: %s", provider.Key)

	provider.Oauth2Config.RedirectURL = cb.RedirectURL
	// Parse the access & ID token
	oauth2Token, err := provider.Oauth2Config.Exchange(context.Background(), cb.Code)
	if err != nil {
		var rerr *oauth2.RetrieveError
		if errors.As(err, &rerr) {

			details := make(map[string]interface{})
			if err := json.Unmarshal(rerr.Body, &details); err != nil {
				log.Errorf("Error unmarshalling token for provider %s: %v", provider.Name, err)
				log.Debugf("Raw token value is %s", rerr.Body)
				return nil, nil, nil, err
			}

			log.Errorf("Error retrieving token: %s", err)
			log.Debugf("Raw token value is %s", rerr.Body)
			return nil, nil, nil, &models.ErrOpenIDBadRequestWithDetails{
				Message: "Could not authenticate against third party.",
				Details: details,
			}
		}

		return nil, nil, nil, err
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Debugf("Could not get id_token, raw token is %v", oauth2Token)
		return nil, nil, nil, &models.ErrOpenIDBadRequest{Message: "Missing token"}
	}

	verifier := provider.openIDProvider.Verifier(&oidc.Config{ClientID: provider.ClientID})

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		log.Errorf("Error verifying token for provider %s: %v", provider.Name, err)
		return nil, nil, nil, err
	}

	return provider, oauth2Token, idToken, nil
}
