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

package caldav

import (
	"errors"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func BasicAuth(username, password string, c echo.Context) (bool, error) {
	creds := &user.Login{
		Username: username,
		Password: password,
	}
	s := db.NewSession()
	defer s.Close()
	u, err := user.CheckUserCredentials(s, creds)
	if err != nil && !user.IsErrWrongUsernameOrPassword(err) {
		log.Errorf("Error during basic auth for caldav: %v", err)
		return false, nil
	}

	if err == nil {
		c.Set("userBasicAuth", u)
		return true, nil
	}

	tokens, err := user.GetCaldavTokens(u)
	if err != nil {
		log.Errorf("Error while getting tokens for caldav auth: %v", err)
		return false, nil
	}

	// Looping over all tokens until we find one that matches
	for _, token := range tokens {
		err = bcrypt.CompareHashAndPassword([]byte(token.Token), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				continue
			}
			log.Errorf("Error while verifying tokens for caldav auth: %v", err)
			return false, nil
		}

		c.Set("userBasicAuth", u)
		return true, nil
	}

	return false, nil
}
