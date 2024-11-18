// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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
	"errors"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
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
	Name            string `json:"name"`
	Key             string `json:"key"`
	OriginalAuthURL string `json:"-"`
	AuthURL         string `json:"auth_url"`
	LogoutURL       string `json:"logout_url"`
	ClientID        string `json:"client_id"`
	Scope           string `json:"scope"`
	ClientSecret    string `json:"-"`
	openIDProvider  *oidc.Provider
	Oauth2Config    *oauth2.Config `json:"-"`
}
type claims struct {
	Email             string                   `json:"email"`
	Name              string                   `json:"name"`
	PreferredUsername string                   `json:"preferred_username"`
	Nickname          string                   `json:"nickname"`
	VikunjaGroups     []map[string]interface{} `json:"vikunja_groups"`
}

func init() {
	petname.NonDeterministicMode()
}

func (p *Provider) setOicdProvider() (err error) {
	p.openIDProvider, err = oidc.NewProvider(context.Background(), p.OriginalAuthURL)
	return err
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
	cb := &Callback{}
	if err := c.Bind(cb); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Bad data"})
	}

	// Check if the provider exists
	providerKey := c.Param("provider")
	provider, err := GetProvider(providerKey)
	log.Debugf("Provider: %v", provider)
	if err != nil {
		return handler.HandleHTTPError(err)
	}
	if provider == nil {
		return c.JSON(http.StatusBadRequest, models.Message{Message: "Provider does not exist"})
	}

	provider.Oauth2Config.RedirectURL = cb.RedirectURL

	// Parse the access & ID token
	oauth2Token, err := provider.Oauth2Config.Exchange(context.Background(), cb.Code)
	if err != nil {
		var rerr *oauth2.RetrieveError
		if errors.As(err, &rerr) {

			details := make(map[string]interface{})
			if err := json.Unmarshal(rerr.Body, &details); err != nil {
				log.Errorf("Error unmarshalling token for provider %s: %v", provider.Name, err)
				return handler.HandleHTTPError(err)
			}

			log.Error(err)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Could not authenticate against third party.",
				"details": details,
			})
		}

		return handler.HandleHTTPError(err)
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
		return handler.HandleHTTPError(err)
	}

	// Extract custom claims
	cl := &claims{}

	err = idToken.Claims(cl)
	if err != nil {
		log.Errorf("Error getting token claims for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err)
	}

	if cl.Email == "" || cl.Name == "" || cl.PreferredUsername == "" {
		info, err := provider.openIDProvider.UserInfo(context.Background(), provider.Oauth2Config.TokenSource(context.Background(), oauth2Token))
		if err != nil {
			log.Errorf("Error getting userinfo for provider %s: %v", provider.Name, err)
			return handler.HandleHTTPError(err)
		}

		cl2 := &claims{}
		err = info.Claims(cl2)
		if err != nil {
			log.Errorf("Error parsing userinfo claims for provider %s: %v", provider.Name, err)
			return handler.HandleHTTPError(err)
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
			return handler.HandleHTTPError(&user.ErrNoOpenIDEmailProvided{})
		}
	}

	s := db.NewSession()
	defer s.Close()

	// Check if we have seen this user before
	u, err := getOrCreateUser(s, cl, idToken.Issuer, idToken.Subject)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error creating new user for provider %s: %v", provider.Name, err)
		return handler.HandleHTTPError(err)
	}

	// does the oidc token contain well formed "vikunja_groups" through vikunja_scope
	log.Debugf("Checking for vikunja_groups in token %v", cl.VikunjaGroups)
	teamData, errs := getTeamDataFromToken(cl.VikunjaGroups, provider)
	if len(teamData) > 0 {
		for _, err := range errs {
			log.Errorf("Error creating teams for user and vikunja groups %s: %v", cl.VikunjaGroups, err)
		}

		// find old teams for user through oidc
		oldOidcTeams, err := models.FindAllOidcTeamIDsForUser(s, u.ID)
		if err != nil {
			log.Debugf("No oidc teams found for user %v", err)
		}
		oidcTeams, err := AssignOrCreateUserToTeams(s, u, teamData, idToken.Issuer)
		if err != nil {
			log.Errorf("Could not proceed with group routine %v", err)
		}
		teamIDsToLeave := utils.NotIn(oldOidcTeams, oidcTeams)
		err = RemoveUserFromTeamsByIDs(s, u, teamIDsToLeave)
		if err != nil {
			log.Errorf("Error while leaving teams %v", err)
		}
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

func AssignOrCreateUserToTeams(s *xorm.Session, u *user.User, teamData []*models.OIDCTeam, issuer string) (oidcTeams []int64, err error) {
	if len(teamData) == 0 {
		return
	}
	// check if we have seen these teams before.
	// find or create Teams and assign user as teammember.
	teams, err := GetOrCreateTeamsByOIDC(s, teamData, u, issuer)
	if err != nil {
		log.Errorf("Error verifying team for %v, got %v. Error: %v", u.Name, teams, err)
		return nil, err
	}
	for _, team := range teams {
		tm := models.TeamMember{TeamID: team.ID, UserID: u.ID, Username: u.Username}
		exists, _ := tm.MembershipExists(s)
		if !exists {
			err = tm.Create(s, u)
			if err != nil {
				log.Errorf("Could not assign user %s to team %s: %v", u.Username, team.Name, err)
			}
		}
		oidcTeams = append(oidcTeams, team.ID)
	}
	return oidcTeams, err
}

func RemoveUserFromTeamsByIDs(s *xorm.Session, u *user.User, teamIDs []int64) (err error) {

	if len(teamIDs) < 1 {
		return nil
	}

	log.Debugf("Removing team_member with user_id %v from team_ids %v", u.ID, teamIDs)
	_, err = s.In("team_id", teamIDs).And("user_id = ?", u.ID).Delete(&models.TeamMember{})
	return err
}

func getTeamDataFromToken(groups []map[string]interface{}, provider *Provider) (teamData []*models.OIDCTeam, errs []error) {
	teamData = []*models.OIDCTeam{}
	errs = []error{}
	for _, team := range groups {
		var name string
		var description string
		var oidcID string
		var IsPublic bool

		// Read name
		_, exists := team["name"]
		if exists {
			name = team["name"].(string)
		}

		// Read description
		_, exists = team["description"]
		if exists {
			description = team["description"].(string)
		}

		// Read isPublic flag
		_, exists = team["isPublic"]
		if exists {
			IsPublic = team["isPublic"].(bool)
		}

		// Read oidcID
		_, exists = team["oidcID"]
		if exists {
			switch t := team["oidcID"].(type) {
			case string:
				oidcID = team["oidcID"].(string)
			case int64:
				oidcID = strconv.FormatInt(team["oidcID"].(int64), 10)
			case float64:
				oidcID = strconv.FormatFloat(team["oidcID"].(float64), 'f', -1, 64)
			default:
				log.Errorf("No oidcID assigned for %v or type %v not supported", team, t)
			}
		}
		if name == "" || oidcID == "" {
			log.Errorf("Claim of your custom scope does not hold name or oidcID for automatic group assignment through oidc provider. Please check %s", provider.Name)
			errs = append(errs, &user.ErrOpenIDCustomScopeMalformed{})
			continue
		}
		teamData = append(teamData, &models.OIDCTeam{Name: name, OidcID: oidcID, Description: description, IsPublic: IsPublic})
	}
	return teamData, errs
}

func getOIDCTeamName(name string) string {
	return name + " (OIDC)"
}

func CreateOIDCTeam(s *xorm.Session, teamData *models.OIDCTeam, u *user.User, issuer string) (team *models.Team, err error) {
	team = &models.Team{
		Name:        getOIDCTeamName(teamData.Name),
		Description: teamData.Description,
		OidcID:      teamData.OidcID,
		Issuer:      issuer,
		IsPublic:    teamData.IsPublic,
	}
	err = team.CreateNewTeam(s, u, false)
	return team, err
}

// GetOrCreateTeamsByOIDC returns a slice of teams which were generated from the oidc data. If a team did not exist previously it is automatically created.
func GetOrCreateTeamsByOIDC(s *xorm.Session, teamData []*models.OIDCTeam, u *user.User, issuer string) (te []*models.Team, err error) {
	te = []*models.Team{}
	// Procedure can only be successful if oidcID is set
	for _, oidcTeam := range teamData {
		team, err := models.GetTeamByOidcIDAndIssuer(s, oidcTeam.OidcID, issuer)
		if err != nil && !models.IsErrOIDCTeamDoesNotExist(err) {
			return nil, err
		}
		if err != nil && models.IsErrOIDCTeamDoesNotExist(err) {
			log.Debugf("Team with oidc_id %v and name %v does not exist. Creating team… ", oidcTeam.OidcID, oidcTeam.Name)
			newTeam, err := CreateOIDCTeam(s, oidcTeam, u, issuer)
			if err != nil {
				return te, err
			}
			te = append(te, newTeam)
			continue
		}

		// Compare the name and update if it changed
		if team.Name != getOIDCTeamName(oidcTeam.Name) {
			team.Name = getOIDCTeamName(oidcTeam.Name)
		}

		// Compare the description and update if it changed
		if team.Description != oidcTeam.Description {
			team.Description = oidcTeam.Description
		}

		// Compare the isPublic flag and update if it changed
		if team.IsPublic != oidcTeam.IsPublic {
			team.IsPublic = oidcTeam.IsPublic
		}

		err = team.Update(s, u)
		if err != nil {
			return nil, err
		}

		log.Debugf("Team with oidc_id %v and name %v already exists.", team.OidcID, team.Name)
		te = append(te, team)
	}
	return te, err
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
			Username: strings.ReplaceAll(cl.PreferredUsername, " ", "-"),
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

		// If their preferred username is already taken, generate a random one
		if user.IsErrUsernameExists(err) {
			uu.Username = petname.Generate(3, "-")
			u, err = user.CreateUser(s, uu)
			if err != nil {
				return nil, err
			}
		}

		// And create their project
		err = models.CreateNewProjectForUser(s, u)
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

		u, err = user.UpdateUser(s, u, false)
		if err != nil {
			return nil, err
		}
	}

	return
}
