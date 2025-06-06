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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func BasicAuth(username, password string, c echo.Context) (bool, error) {
	s := db.NewSession()
	defer s.Close()

	credentials := &user.Login{
		Username: username,
		Password: password,
	}
	var err error
	u, err := checkUserCaldavTokens(s, credentials)
	if user.IsErrUserDoesNotExist(err) {
		return false, nil
	}
	if u == nil {
		u, err = user.CheckUserCredentials(s, credentials)
		if err != nil {
			log.Errorf("Error during basic auth for caldav: %v", err)
			return false, nil
		}
	}
	if u != nil && err == nil {
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
	tokens, err := user.GetCaldavTokens(usr)
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
