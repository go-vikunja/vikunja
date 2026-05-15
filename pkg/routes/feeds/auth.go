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

package feeds

import (
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"

	"github.com/labstack/echo/v5"
)

func checkAPIToken(s *xorm.Session, username, token string) (*user.User, error) {
	apiToken, u, err := models.ValidateTokenAndGetOwner(s, token)
	if err != nil {
		return nil, err
	}
	if apiToken == nil || u == nil {
		return nil, nil
	}

	if !apiToken.HasFeedsAccess() {
		log.Debugf("[feeds auth] API token %d does not have feeds access permission", apiToken.ID)
		return nil, nil
	}

	if u.Username != username {
		log.Debugf("[feeds auth] API token %d owner %s does not match provided username %s", apiToken.ID, u.Username, username)
		return nil, nil
	}

	return u, nil
}

// BasicAuth authenticates feed requests. Only API tokens are accepted —
// password and LDAP credentials are rejected outright because feed URLs are
// commonly exported, shared, or cached by feed readers.
func BasicAuth(c *echo.Context, username, password string) (bool, error) {
	if !strings.HasPrefix(password, models.APITokenPrefix) {
		return false, nil
	}
	// GetTokenFromTokenString slices password[len-8:] without a length check,
	// so a stray "tk_" or other short prefix-only string would panic before
	// the credentials could be rejected. Real tokens are far longer than
	// prefix+8, so anything shorter is invalid by construction.
	if len(password) < len(models.APITokenPrefix)+8 {
		return false, nil
	}

	s := db.NewSession()
	defer s.Close()

	u, err := checkAPIToken(s, username, password)
	if err != nil {
		log.Errorf("Error during API token auth for feeds: %v", err)
		return false, nil
	}
	if u == nil {
		return false, nil
	}
	if u.IsBot() {
		log.Warningf("Feed auth rejected for bot user %d", u.ID)
		return false, nil
	}

	c.Set("userBasicAuth", u)
	return true, nil
}
