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
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"

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
		processedAvatar, err := cropAvatarTo1x1(raw)
		if err != nil {
			log.Debugf("Error processing LDAP avatar: %v", err)
			// Continue without avatar if processing fails
		} else {
			err = upload.StoreAvatarFile(s, u, bytes.NewReader(processedAvatar))
			if err != nil {
				return nil, err
			}
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

// cropAvatarTo1x1 crops the avatar image to a 1:1 aspect ratio, centered on the image
func cropAvatarTo1x1(imageData []byte) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, errors.New("empty image data")
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If already square, return original
	if width == height {
		return imageData, nil
	}

	// Determine the crop size (use the smaller dimension)
	size := width
	if height < width {
		size = height
	}

	// Calculate crop coordinates to center the image
	x0 := (width - size) / 2
	y0 := (height - size) / 2
	x1 := x0 + size
	y1 := y0 + size

	// Create the cropping rectangle
	cropRect := image.Rect(x0, y0, x1, y1)

	// Create a new RGBA image
	croppedImg := image.NewRGBA(image.Rect(0, 0, size, size))

	// Copy the cropped portion
	draw.Draw(croppedImg, croppedImg.Bounds(), img, cropRect.Min, draw.Src)

	// Encode the result
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, croppedImg, nil)
	case "png":
		err = png.Encode(&buf, croppedImg)
	default:
		// Default to PNG if format is unknown
		err = png.Encode(&buf, croppedImg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode cropped image: %w", err)
	}

	return buf.Bytes(), nil
}
