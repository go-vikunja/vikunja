// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package openid

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"code.vikunja.io/web/handler"

	"code.vikunja.io/api/pkg/db"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"github.com/coreos/go-oidc/v3/oidc"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// Callback contains the callback after an auth request was made and redirected
type Callback struct {
	Code  string `query:"code" json:"code"`
	Scope string `query:"scop" json:"scope"`
}

// Provider is the structure of an OpenID Connect provider
type Provider struct {
	Name            string `json:"name"`
	Key             string `json:"key"`
	OriginalAuthURL string `json:"-"`
	AuthURL         string `json:"auth_url"`
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"-"`
	openIDProvider  *oidc.Provider
	Oauth2Config    *oauth2.Config `json:"-"`
}

type claims struct {
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Nickname          string `json:"nickname"`
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (p *Provider) setOicdProvider() (err error) {
	p.openIDProvider, err = oidc.NewProvider(context.Background(), p.OriginalAuthURL)
	return err
}

// HandleCallback handles the auth request callback after redirecting from the provider with an auth code
// @Summary Authenticate a user with OpenID Connect
// @Description After a redirect from the OpenID Connect provider to the frontend has been made with the authentication `code`, this endpoint can be used to obtain a jwt token for that user and thus log them in.
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
	cb := &Callback{}
	if err := c.Bind(cb); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Bad data"})
	}

	// Check if the provider exists
	providerKey := c.Param("provider")
	provider, err := GetProvider(providerKey)
	if err != nil {
		log.Error(err)
		return handler.HandleHTTPError(err, c)
	}
	if provider == nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Provider does not exist"})
	}

	// Parse the access & ID token
	oauth2Token, err := provider.Oauth2Config.Exchange(context.Background(), cb.Code)
	if err != nil {
		if rerr, is := err.(*oauth2.RetrieveError); is {
			log.Error(err)

			details := make(map[string]interface{})
			if err := json.Unmarshal(rerr.Body, &details); err != nil {
				log.Errorf("Error unmarshaling token for provider %s: %v", provider.Name, err)
				return handler.HandleHTTPError(err, c)
			}

			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Could not authenticate against third party.",
				"details": details,
			})
		}

		return handler.HandleHTTPError(err, c)
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Missing token"})
	}

	verifier := provider.openIDProvider.Verifier(&oidc.Config{ClientID: provider.ClientID})

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		log.Errorf("Error verifying token for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err, c)
	}

	// Extract custom claims
	cl := &claims{}
	err = idToken.Claims(cl)
	if err != nil {
		log.Errorf("Error getting token claims for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err, c)
	}

	if cl.Email == "" || cl.Name == "" || cl.PreferredUsername == "" {
		info, err := provider.openIDProvider.UserInfo(context.Background(), provider.Oauth2Config.TokenSource(context.Background(), oauth2Token))
		if err != nil {
			log.Errorf("Error getting userinfo for provider %s: %v", provider.Name, err)
			return handler.HandleHTTPError(err, c)
		}

		cl2 := &claims{}
		err = info.Claims(cl2)
		if err != nil {
			log.Errorf("Error parsing userinfo claims for provider %s: %v", provider.Name, err)
			return handler.HandleHTTPError(err, c)
		}

		if cl.Email == "" {
			cl.Email = cl2.Email
		}

		if cl.Name == "" {
			cl.Name = cl2.Name
		}

		if cl.PreferredUsername == "" {
			cl.PreferredUsername = cl2.PreferredUsername
		}

		if cl.PreferredUsername == "" && cl2.Nickname != "" {
			cl.PreferredUsername = cl2.Nickname
		}

		if cl.Email == "" {
			log.Errorf("Claim does not contain an email address for provider %s", provider.Name)
			return handler.HandleHTTPError(&user.ErrNoOpenIDEmailProvided{}, c)
		}
	}

	s := db.NewSession()
	defer s.Close()

	// Check if we have seen this user before
	u, err := getOrCreateUser(s, cl, idToken.Issuer, idToken.Subject)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error creating new user for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err, c)
	}

	err = s.Commit()
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	// Create token
	return auth.NewUserAuthTokenResponse(u, c)
}

func getOrCreateUser(s *xorm.Session, cl *claims, issuer, subject string) (u *user.User, err error) {
	// Check if the user exists for that issuer and subject
	u, err = user.GetUserWithEmail(s, &user.User{
		Issuer:  issuer,
		Subject: subject,
	})
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		return nil, err
	}

	// If no user exists, create one with the preferred username if it is not already taken
	if user.IsErrUserDoesNotExist(err) {
		uu := &user.User{
			Username: cl.PreferredUsername,
			Email:    cl.Email,
			Name:     cl.Name,
			Status:   user.StatusActive,
			Issuer:   issuer,
			Subject:  subject,
		}

		// Check if we actually have a preferred username and generate a random one right away if we don't
		if uu.Username == "" {
			uu.Username = petname.Generate(3, "-")
		}

		u, err = user.CreateUser(s, uu)
		if err != nil && !user.IsErrUsernameExists(err) {
			return nil, err
		}

		// If their preferred username is already taken, create some random one from the email and subject
		if user.IsErrUsernameExists(err) {
			uu.Username = petname.Generate(3, "-")
			u, err = user.CreateUser(s, uu)
			if err != nil {
				return nil, err
			}
		}

		// And create its namespace
		err = models.CreateNewNamespaceForUser(s, u)
		if err != nil {
			return nil, err
		}

		return
	}

	// If it exists, check if the email address changed and change it if not
	if cl.Email != u.Email || cl.Name != u.Name {
		if cl.Email != u.Email {
			u.Email = cl.Email
		}
		if cl.Name != u.Name {
			u.Name = cl.Name
		}
		u, err = user.UpdateUser(s, &user.User{
			ID:      u.ID,
			Email:   u.Email,
			Name:    u.Name,
			Issuer:  issuer,
			Subject: subject,
		})
		if err != nil {
			return nil, err
		}
	}

	return
}
