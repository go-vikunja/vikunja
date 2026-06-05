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

package caldav

import (
	"errors"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

func checkAPIToken(s *xorm.Session, username, token string) (*user.User, error) {
	apiToken, u, err := models.ValidateTokenAndGetOwner(s, token)
	if err != nil {
		return nil, err
	}
	if apiToken == nil || u == nil {
		return nil, nil
	}

	if !apiToken.HasCaldavAccess() {
		log.Debugf("[caldav auth] API token %d does not have caldav access permission", apiToken.ID)
		return nil, nil
	}

	if u.Username != username {
		log.Debugf("[caldav auth] API token %d owner %s does not match provided username %s", apiToken.ID, u.Username, username)
		return nil, nil
	}

	return u, nil
}

func BasicAuth(c *echo.Context, username, password string) (bool, error) {
	s := db.NewSession()
	defer s.Close()

	// If the password looks like an API token, validate it as one.
	// Don't fall through to other auth methods — tk_ prefix is unambiguous.
	if strings.HasPrefix(password, models.APITokenPrefix) {
		u, err := checkAPIToken(s, username, password)
		if err != nil {
			log.Errorf("Error during API token auth for caldav: %v", err)
			return false, nil
		}
		if u != nil {
			if u.IsBot() {
				log.Warningf("CalDAV auth rejected for bot user %d", u.ID)
				return false, nil
			}
			c.Set("userBasicAuth", u)
			return true, nil
		}
		return false, nil
	}

	credentials := &user.Login{
		Username: username,
		Password: password,
	}
	var err error
	u, err := checkUserCaldavTokens(s, credentials)
	if user.IsErrUserDoesNotExist(err) {
		return false, nil
	}
	if user.IsErrUserStatusError(err) {
		return false, nil
	}
	if u == nil {
		u, err = user.CheckUserCredentials(s, credentials)
		if err != nil {
			log.Errorf("Error during basic auth for caldav: %v", err)
			return false, nil
		}

		// If the user has TOTP enabled, reject password-based basic auth.
		// They must use a CalDAV token instead.
		totpEnabled, err := user.TOTPEnabledForUser(s, u)
		if err != nil {
			log.Errorf("Error checking TOTP status for caldav basic auth: %v", err)
			return false, nil
		}
		if totpEnabled {
			log.Warningf("CalDAV basic auth rejected for user %d: TOTP is enabled, a CalDAV token is required", u.ID)
			return false, nil
		}
	}
	if u != nil && err == nil {
		if u.IsBot() {
			log.Warningf("CalDAV basic auth rejected for bot user %d", u.ID)
			return false, nil
		}
		c.Set("userBasicAuth", u)
		return true, nil
	}
	return false, nil
}

func checkUserCaldavTokens(s *xorm.Session, login *user.Login) (*user.User, error) {
	usr, err := user.GetUserByUsername(s, login.Username)
	if err != nil || usr == nil {
		log.Warningf("Error while retrieving users from database: %v", err)
		return nil, err
	}
	tokens, err := user.GetCaldavTokensWithSession(s, usr)
	if err != nil {
		log.Errorf("Error while getting tokens for caldav auth: %v", err)
		return nil, err
	}
	// Looping over all tokens until we find one that matches
	for _, token := range tokens {
		err = bcrypt.CompareHashAndPassword([]byte(token.Token), []byte(login.Password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				continue
			}
			log.Errorf("Error while verifying tokens for caldav auth: %v", err)
			return nil, nil
		}
		return usr, nil
	}
	return nil, nil
}
