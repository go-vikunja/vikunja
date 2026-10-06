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
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/config"
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
	if config.AuthLdapUserFilter.GetString() == "" {
		log.Fatal("LDAP user filter is not configured")
	}

	err := utils.RetryWithBackoff("LDAP server", func() error {
		l, connErr := ConnectAndBindToLDAPDirectory()
		if connErr == nil {
			_ = l.Close()
		}
		return connErr
	})

	if err != nil {
		log.Fatalf("Could not connect to LDAP server: %s", err)
	}
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
		return nil, fmt.Errorf("could not connect to LDAP server: %w", err)
	}

	bindDN := config.AuthLdapBindDN.GetString()
	bindPassword := config.AuthLdapBindPassword.GetString()

	if bindDN != "" && bindPassword != "" {
		// Standard authentication
		err = l.Bind(bindDN, bindPassword)
	} else {
		// Anonymous bind attempt (depending on the server, this call is explicit or automatic)
		log.Info("No LDAP bind DN or password configured, attempting anonymous bind")
		err = l.UnauthenticatedBind("")
	}
	return
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
	escapedUsername := ldap.EscapeFilter(username)

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

	attributes := []string{
		"dn",
		config.AuthLdapAttributeUsername.GetString(),
		config.AuthLdapAttributeEmail.GetString(),
		config.AuthLdapAttributeDisplayname.GetString(),
		"jpegPhoto",
	}
	if idAttribute := config.AuthLdapAttributeUserID.GetString(); idAttribute != "" {
		attributes = append(attributes, idAttribute)
	}

	searchRequest := ldap.NewSearchRequest(
		config.AuthLdapBaseDN.GetString(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		userFilter,
		attributes,
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

	// After verifying the user's password above the connection is bound as the
	// end user. Many directories restrict group searches to service accounts, so
	// re-bind as the service account before enumerating groups when configured.
	if config.AuthLdapGroupSyncUseServiceAccount.GetBool() {
		bindDN := config.AuthLdapBindDN.GetString()
		bindPassword := config.AuthLdapBindPassword.GetString()
		if bindDN != "" && bindPassword != "" {
			if err = l.Bind(bindDN, bindPassword); err != nil {
				return nil, fmt.Errorf("could not re-bind service account for group sync: %w", err)
			}
		} else {
			if err = l.UnauthenticatedBind(""); err != nil {
				return nil, fmt.Errorf("could not re-bind anonymously for group sync: %w", err)
			}
		}
	}

	ldapUsername := sr.Entries[0].GetAttributeValue(config.AuthLdapAttributeUsername.GetString())
	err = syncUserGroups(s, l, u, userdn, ldapUsername)

	return u, err
}

func getOrCreateLdapUser(s *xorm.Session, entry *ldap.Entry) (u *user.User, err error) {
	username := entry.GetAttributeValue(config.AuthLdapAttributeUsername.GetString())
	email := entry.GetAttributeValue(config.AuthLdapAttributeEmail.GetString())
	name := entry.GetAttributeValue(config.AuthLdapAttributeDisplayname.GetString())

	subject := username
	if idAttribute := config.AuthLdapAttributeUserID.GetString(); idAttribute != "" {
		subject, err = directoryID(entry, idAttribute)
		if err != nil {
			return nil, err
		}
	}
	// An empty subject would match any LDAP user.
	if subject == "" {
		return nil, errors.New("ldap user has no value for the configured username or id attribute")
	}

	u, err = findOrMigrateLdapUser(s, subject, username)
	if err != nil && !user.IsErrUserDoesNotExist(err) && !user.IsErrUserStatusError(err) {
		return nil, err
	}

	// If the user exists but is disabled/locked, return early without updating profile
	if user.IsErrUserStatusError(err) {
		return u, nil
	}

	// If no user exists, create one with the preferred username if it is not already taken
	if user.IsErrUserDoesNotExist(err) {
		uu := &user.User{
			Username: strings.ReplaceAll(username, " ", "-"),
			Email:    email,
			Name:     name,
			Status:   user.StatusActive,
			Issuer:   user.IssuerLDAP,
			Subject:  subject,
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

func findOrMigrateLdapUser(s *xorm.Session, subject, username string) (*user.User, error) {
	u, err := user.GetUserWithEmail(s, &user.User{
		Issuer:  user.IssuerLDAP,
		Subject: subject,
	})
	if subject == username || username == "" || !user.IsErrUserDoesNotExist(err) {
		return u, err
	}

	u, err = user.GetUserWithEmail(s, &user.User{
		Issuer:  user.IssuerLDAP,
		Subject: username,
	})
	if err != nil && !user.IsErrUserStatusError(err) {
		return u, err
	}

	log.Debugf("Migrating subject of LDAP user %d from %s to %s", u.ID, username, subject)
	_, updateErr := s.
		Where("id = ?", u.ID).
		Cols("subject").
		Update(&user.User{Subject: subject})
	if updateErr != nil {
		return nil, updateErr
	}
	u.Subject = subject

	return u, err
}

// objectGUID is formatted the way AD tools show it.
func directoryID(entry *ldap.Entry, attribute string) (string, error) {
	raw := entry.GetEqualFoldRawAttributeValue(attribute)
	if len(raw) == 0 {
		return "", fmt.Errorf("ldap entry %s has no value for attribute %s", entry.DN, attribute)
	}

	if strings.EqualFold(attribute, "objectGUID") {
		return formatObjectGUID(raw)
	}

	if !utf8.Valid(raw) || bytes.ContainsFunc(raw, unicode.IsControl) {
		return "", fmt.Errorf("ldap entry %s has a binary value for attribute %s, only objectGUID is supported as a binary id", entry.DN, attribute)
	}

	return string(raw), nil
}

// formatObjectGUID decodes AD's mixed-endian layout: the first three groups
// are little-endian, the rest is in byte order.
func formatObjectGUID(b []byte) (string, error) {
	if len(b) != 16 {
		return "", fmt.Errorf("objectGUID must be 16 bytes, got %d", len(b))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%x-%x",
		binary.LittleEndian.Uint32(b[0:4]),
		binary.LittleEndian.Uint16(b[4:6]),
		binary.LittleEndian.Uint16(b[6:8]),
		b[8:10],
		b[10:16],
	), nil
}

const (
	groupSyncFilterUserDN   = "{userdn}"
	groupSyncFilterUsername = "{username}"
)

// Below AD's default MaxPageSize of 1000.
const groupSearchPageSize = 500

// buildGroupSyncFilter returns perUser = true when the template references the
// user, in which case every group matching the filter is a membership.
func buildGroupSyncFilter(template, userDN, username string) (filter string, perUser bool, err error) {
	hasUserDN := strings.Contains(template, groupSyncFilterUserDN)
	hasUsername := strings.Contains(template, groupSyncFilterUsername)
	if !hasUserDN && !hasUsername {
		return template, false, nil
	}

	// An empty value would turn e.g. (memberUid={username}*) into (memberUid=*) and grant every group.
	if hasUserDN && userDN == "" {
		return "", false, fmt.Errorf("group sync filter uses %s but the user DN is empty", groupSyncFilterUserDN)
	}
	if hasUsername && username == "" {
		return "", false, fmt.Errorf("group sync filter uses %s but the user has no %s attribute", groupSyncFilterUsername, config.AuthLdapAttributeUsername.GetString())
	}

	return strings.NewReplacer(
		groupSyncFilterUserDN, ldap.EscapeFilter(userDN),
		groupSyncFilterUsername, ldap.EscapeFilter(username),
	).Replace(template), true, nil
}

func syncUserGroups(s *xorm.Session, l *ldap.Conn, u *user.User, userdn, ldapUsername string) (err error) {
	filter, perUser, err := buildGroupSyncFilter(config.AuthLdapGroupSyncFilter.GetString(), userdn, ldapUsername)
	if err != nil {
		return err
	}
	memberAttribute := config.AuthLdapAttributeMemberID.GetString()
	idAttribute := config.AuthLdapAttributeGroupID.GetString()

	attributes := []string{"cn", "description"}
	if !perUser {
		attributes = append(attributes, memberAttribute)
	}
	if idAttribute != "" {
		attributes = append(attributes, idAttribute)
	}

	searchRequest := ldap.NewSearchRequest(
		config.AuthLdapBaseDN.GetString(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, groupSearchPageSize)
	if err != nil {
		log.Errorf("Error searching for LDAP groups: %v", err)
		return err
	}

	log.Debugf("Found %d LDAP groups for user %s with filter %s", len(sr.Entries), userdn, filter)

	var teams []*models.Team
	newIDsByDN := map[string]string{}

	for _, group := range sr.Entries {
		if !perUser && !slices.ContainsFunc(group.GetAttributeValues(memberAttribute), func(member string) bool {
			return member == userdn || member == ldapUsername
		}) {
			continue
		}

		externalID := group.DN
		if idAttribute != "" {
			externalID, err = directoryID(group, idAttribute)
			if err != nil {
				return err
			}
			newIDsByDN[group.DN] = externalID
		}

		teams = append(teams, &models.Team{
			Name:        group.GetAttributeValue("cn"),
			ExternalID:  externalID,
			Description: group.GetAttributeValue("description"),
		})
	}

	err = models.MigrateExternalTeamIDs(s, user.IssuerLDAP, newIDsByDN)
	if err != nil {
		return err
	}

	err = models.SyncExternalTeamsForUser(s, u, teams, user.IssuerLDAP, "LDAP")
	return
}
