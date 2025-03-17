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

package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"github.com/go-ldap/ldap/v3"
	"xorm.io/xorm"
)

type team struct {
	Name        string
	DN          string
	Description string
}

func InitializeLDAPConnection() {
	if !config.AuthLdapEnabled.GetBool() {
		return
	}

	if config.AuthLdapHost.GetString() == "" {
		log.Fatal("LDAP host is not configured")
	}
	if config.AuthLdapPort.GetInt() == 0 {
		log.Fatal("LDAP port is not configured")
	}
	if config.AuthLdapBaseDN.GetString() == "" {
		log.Fatal("LDAP base DN is not configured")
	}
	if config.AuthLdapBindDN.GetString() == "" {
		log.Fatal("LDAP bind DN is not configured")
	}
	if config.AuthLdapBindPassword.GetString() == "" {
		log.Fatal("LDAP bind password is not configured")
	}
	if config.AuthLdapUserFilter.GetString() == "" {
		log.Fatal("LDAP user filter is not configured")
	}

	l, err := ConnectAndBindToLDAPDirectory()
	if err != nil {
		log.Fatalf("Could not bind to LDAP server: %s", err)
	}
	_ = l.Close()
}

func ConnectAndBindToLDAPDirectory() (l *ldap.Conn, err error) {
	var protocol = "ldap"
	if config.AuthLdapUseTLS.GetBool() {
		protocol = "ldaps"
	}
	url := fmt.Sprintf(
		"%s://%s:%d",
		protocol,
		config.AuthLdapHost.GetString(),
		config.AuthLdapPort.GetInt(),
	)

	opts := []ldap.DialOpt{}
	if config.AuthLdapUseTLS.GetBool() {
		// #nosec G402
		opts = append(opts, ldap.DialWithTLSConfig(&tls.Config{
			InsecureSkipVerify: !config.AuthLdapVerifyTLS.GetBool(),
		}))
	}

	l, err = ldap.DialURL(url, opts...)
	if err != nil {
		log.Fatalf("Could not connect to LDAP server: %s", err)
	}

	err = l.Bind(
		config.AuthLdapBindDN.GetString(),
		config.AuthLdapBindPassword.GetString(),
	)
	return
}

// Adjusted from https://github.com/go-gitea/gitea/blob/6ca91f555ab9778310ac46cbbe33849c59286793/services/auth/source/ldap/source_search.go#L34
func sanitizedUserQuery(username string) (string, bool) {
	// See http://tools.ietf.org/search/rfc4515
	badCharacters := "\x00()*\\"
	if strings.ContainsAny(username, badCharacters) {
		log.Debugf("'%s' contains invalid query characters. Aborting.", username)
		return "", false
	}

	return fmt.Sprintf(config.AuthLdapUserFilter.GetString(), username), true
}

func AuthenticateUserInLDAP(s *xorm.Session, username, password string) (u *user.User, err error) {
	if password == "" || username == "" {
		return nil, user.ErrNoUsernamePassword{}
	}

	l, err := ConnectAndBindToLDAPDirectory()
	if err != nil {
		log.Errorf("Could not bind to LDAP server: %s", err)
		return
	}
	defer l.Close()

	log.Debugf("Connected to LDAP server")

	userFilter, ok := sanitizedUserQuery(username)
	if !ok {
		log.Debugf("Could not sanitize username %s", username)
		return nil, user.ErrWrongUsernameOrPassword{}
	}

	searchRequest := ldap.NewSearchRequest(
		config.AuthLdapBaseDN.GetString(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		userFilter,
		[]string{
			"dn",
			config.AuthLdapAttributeUsername.GetString(),
			config.AuthLdapAttributeEmail.GetString(),
			config.AuthLdapAttributeDisplayname.GetString(),
		},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return
	}

	if len(sr.Entries) > 1 || len(sr.Entries) == 0 {
		log.Debugf("Found %d entries for username %s", len(sr.Entries), username)
		return nil, user.ErrWrongUsernameOrPassword{}
	}

	userdn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind(userdn, password)
	if err != nil {
		var lerr *ldap.Error
		if errors.As(err, &lerr) && lerr.ResultCode == ldap.LDAPResultInvalidCredentials {
			return nil, user.ErrWrongUsernameOrPassword{}
		}

		return
	}

	u, err = getOrCreateLdapUser(s, sr.Entries[0])

	// TODO this should be unified with openid
	syncUserGroups(l, u, userdn)

	return u, err
}

func getOrCreateLdapUser(s *xorm.Session, entry *ldap.Entry) (u *user.User, err error) {
	username := entry.GetAttributeValue(config.AuthLdapAttributeUsername.GetString())
	email := entry.GetAttributeValue(config.AuthLdapAttributeEmail.GetString())
	name := entry.GetAttributeValue(config.AuthLdapAttributeDisplayname.GetString())

	u, err = user.GetUserWithEmail(s, &user.User{
		Issuer:  user.IssuerLDAP,
		Subject: username,
	})
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		return nil, err
	}

	// If no user exists, create one with the preferred username if it is not already taken
	if user.IsErrUserDoesNotExist(err) {
		uu := &user.User{
			Username: strings.ReplaceAll(username, " ", "-"),
			Email:    email,
			Name:     name,
			Status:   user.StatusActive,
			Issuer:   user.IssuerLDAP,
			Subject:  username,
		}

		return auth.CreateUserWithRandomUsername(s, uu)
	}

	return
}

func syncUserGroups(l *ldap.Conn, u *user.User, userdn string) {
	s := db.NewSession()
	defer s.Close()

	searchRequest := ldap.NewSearchRequest(
		config.AuthLdapBaseDN.GetString(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectclass=*)(|(objectclass=group)(objectclass=groupOfNames)))",
		[]string{
			"dn",
			"cn",
			"member",
			"description",
		},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Errorf("Error searching for LDAP groups: %v", err)
		return
	}

	var teams []*team

	for _, group := range sr.Entries {
		groupName := group.GetAttributeValue("cn")
		members := group.GetAttributeValues("member")
		description := group.GetAttributeValue("description")

		log.Debugf("Group %s has %d members", groupName, len(members))

		for _, member := range members {
			if member == userdn {
				teams = append(teams, &team{
					Name:        groupName,
					DN:          group.DN,
					Description: description,
				})
			}
		}
	}

	if len(teams) > 0 {
		// Find old teams for user through LDAP
		oldLdapTeams, err := models.FindAllExternalTeamIDsForUser(s, u.ID)
		if err != nil {
			log.Errorf("Error retrieving external team ids for user: %v", err)
			return
		}

		// Assign or create teams for the user
		ldapTeamIDs, err := assignOrCreateUserToTeams(s, u, teams)
		if err != nil {
			log.Errorf("Could not assign or create user to teams: %v", err)
			return
		}

		// Remove user from teams they're no longer a member of
		teamIDsToLeave := utils.NotIn(oldLdapTeams, ldapTeamIDs)
		err = RemoveUserFromTeamsByIDs(s, u, teamIDsToLeave)
		if err != nil {
			log.Errorf("Error while removing user from teams: %v", err)
			return
		}

		err = s.Commit()
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error committing LDAP team changes: %v", err)
		}
	}
}

func assignOrCreateUserToTeams(s *xorm.Session, u *user.User, teamData []*team) (ldapTeamIDs []int64, err error) {
	if len(teamData) == 0 {
		return
	}

	// Check if we have seen these teams before.
	// Find or create Teams and assign user as teammember.
	teams, err := GetOrCreateTeamsByLDAP(s, teamData, u)
	if err != nil {
		log.Errorf("Error verifying team for %v, got %v. Error: %v", u.Name, teams, err)
		return nil, err
	}

	for _, team := range teams {
		tm := models.TeamMember{
			TeamID:   team.ID,
			UserID:   u.ID,
			Username: u.Username,
		}
		exists, _ := tm.MembershipExists(s)
		if !exists {
			err = tm.Create(s, u)
			if err != nil {
				log.Errorf("Could not assign user %s to team %s: %v", u.Username, team.Name, err)
			}
		}
		ldapTeamIDs = append(ldapTeamIDs, team.ID)
	}

	return ldapTeamIDs, err
}

func RemoveUserFromTeamsByIDs(s *xorm.Session, u *user.User, teamIDs []int64) (err error) {
	if len(teamIDs) < 1 {
		return nil
	}

	log.Debugf("Removing team_member with user_id %v from team_ids %v", u.ID, teamIDs)
	_, err = s.
		In("team_id", teamIDs).
		And("user_id = ?", u.ID).
		Delete(&models.TeamMember{})
	return err
}

func getLDAPTeamName(name string) string {
	return name + " (LDAP)"
}

func createLDAPTeam(s *xorm.Session, teamData *team, u *user.User) (team *models.Team, err error) {
	team = &models.Team{
		Name:        getLDAPTeamName(teamData.Name),
		Description: teamData.Description,
		ExternalID:  teamData.DN,
		Issuer:      user.IssuerLDAP,
	}
	err = team.CreateNewTeam(s, u, false)
	return team, err
}

// GetOrCreateTeamsByLDAP returns a slice of teams which were generated from the LDAP data.
// If a team did not exist previously it is automatically created.
func GetOrCreateTeamsByLDAP(s *xorm.Session, teamData []*team, u *user.User) (teams []*models.Team, err error) {
	teams = []*models.Team{}

	for _, ldapTeam := range teamData {
		t, err := models.GetTeamByExternalIDAndIssuer(s, ldapTeam.DN, user.IssuerLDAP)
		if err != nil && !models.IsErrExternalTeamDoesNotExist(err) {
			return nil, err
		}

		if err != nil && models.IsErrExternalTeamDoesNotExist(err) {
			log.Debugf("Team with LDAP DN %v and name %v does not exist. Creating team...", ldapTeam.DN, ldapTeam.Name)
			newTeam, err := createLDAPTeam(s, ldapTeam, u)
			if err != nil {
				return teams, err
			}
			teams = append(teams, newTeam)
			continue
		}

		// Compare the name and update if it changed
		if t.Name != getLDAPTeamName(ldapTeam.Name) {
			t.Name = getLDAPTeamName(ldapTeam.Name)
		}

		// Compare the description and update if it changed
		if t.Description != ldapTeam.Description {
			t.Description = ldapTeam.Description
		}

		err = t.Update(s, u)
		if err != nil {
			return nil, err
		}

		log.Debugf("Team with LDAP DN %v and name %v already exists.", ldapTeam.DN, t.Name)
		teams = append(teams, t)
	}

	return teams, err
}
