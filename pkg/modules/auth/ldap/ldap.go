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

package ldap

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"github.com/go-ldap/ldap/v3"
	"xorm.io/xorm"
)

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

// escapeLDAPFilterValue escapes special characters in LDAP filter values according to RFC 4515.
// This prevents LDAP injection attacks by properly escaping all special characters.
func escapeLDAPFilterValue(value string) string {
	var buf strings.Builder
	buf.Grow(len(value) * 2) // Pre-allocate to avoid reallocations

	for _, r := range value {
		switch r {
		case 0x00: // NULL
			buf.WriteString(`\00`)
		case '(':
			buf.WriteString(`\28`)
		case ')':
			buf.WriteString(`\29`)
		case '*':
			buf.WriteString(`\2a`)
		case '\\':
			buf.WriteString(`\5c`)
		case '&':
			buf.WriteString(`\26`)
		case '|':
			buf.WriteString(`\7c`)
		case '=':
			buf.WriteString(`\3d`)
		case '<':
			buf.WriteString(`\3c`)
		case '>':
			buf.WriteString(`\3e`)
		case '~':
			buf.WriteString(`\7e`)
		default:
			buf.WriteRune(r)
		}
	}

	return buf.String()
}

// Adjusted from https://github.com/go-gitea/gitea/blob/6ca91f555ab9778310ac46cbbe33849c59286793/services/auth/source/ldap/source_search.go#L34
func sanitizedUserQuery(username string) (string, bool) {
	// Validate username is not empty and doesn't contain control characters
	if username == "" {
		log.Debugf("Empty username provided. Aborting.")
		return "", false
	}

	// Check for control characters that shouldn't be in usernames
	for _, r := range username {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, LF, CR but block other control chars
			log.Debugf("Username contains control character 0x%02x. Aborting.", r)
			return "", false
		}
	}

	// Escape the username according to RFC 4515 to prevent LDAP injection
	escapedUsername := escapeLDAPFilterValue(username)

	return fmt.Sprintf(config.AuthLdapUserFilter.GetString(), escapedUsername), true
}

func AuthenticateUserInLDAP(s *xorm.Session, username, password string, syncGroups bool, avatarSyncAttribute string) (u *user.User, err error) {
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
			"jpegPhoto",
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
	if err != nil {
		return nil, err
	}

	if avatarSyncAttribute != "" {
		raw := sr.Entries[0].GetRawAttributeValue(avatarSyncAttribute)
		u.AvatarProvider = "ldap"

		// Process the avatar image to ensure 1:1 aspect ratio
		processedAvatar, err := utils.CropAvatarTo1x1(raw)
		if err != nil {
			log.Debugf("Error processing LDAP avatar: %v", err)
			// Continue without avatar if processing fails
		} else {
			err = upload.StoreAvatarFile(s, u, bytes.NewReader(processedAvatar))
			if err != nil {
				return nil, err
			}
			avatar.FlushAllCaches(u)
		}
	}

	if !syncGroups {
		return
	}

	err = syncUserGroups(l, u, userdn)

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

	// Check if user information has changed and update if necessary
	needsUpdate := false

	if u.Email != email && email != "" {
		u.Email = email
		needsUpdate = true
	}

	if u.Name != name && name != "" {
		u.Name = name
		needsUpdate = true
	}

	if needsUpdate {
		log.Debugf("Updating LDAP user information for %s", username)
		_, err = s.Where("id = ?", u.ID).
			Cols("email", "name").
			Update(u)
		if err != nil {
			log.Errorf("Failed to update user information: %v", err)
			return nil, err
		}
	}

	return
}

func syncUserGroups(l *ldap.Conn, u *user.User, userdn string) (err error) {
	s := db.NewSession()
	defer s.Close()

	searchRequest := ldap.NewSearchRequest(
		config.AuthLdapBaseDN.GetString(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		config.AuthLdapGroupSyncFilter.GetString(),
		[]string{
			"dn",
			"cn",
			config.AuthLdapAttributeMemberID.GetString(),
			"description",
		},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Errorf("Error searching for LDAP groups: %v", err)
		return err
	}

	var teams []*models.Team

	for _, group := range sr.Entries {
		groupName := group.GetAttributeValue("cn")
		members := group.GetAttributeValues(config.AuthLdapAttributeMemberID.GetString())
		description := group.GetAttributeValue("description")

		log.Debugf("Group %s has %d members", groupName, len(members))

		for _, member := range members {
			if member == userdn || member == u.Username {
				teams = append(teams, &models.Team{
					Name:        groupName,
					ExternalID:  group.DN,
					Description: description,
				})
			}
		}
	}

	err = models.SyncExternalTeamsForUser(s, u, teams, user.IssuerLDAP, "LDAP")
	if err != nil {
		return
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
	}

	return
}
