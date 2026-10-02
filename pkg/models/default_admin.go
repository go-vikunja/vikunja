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

package models

import (
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// DefaultAdminUsername is the username of the account EnsureDefaultAdmin creates.
const DefaultAdminUsername = "admin"

// SeedDefaultAdmin opens its own transaction and creates the default admin when needed. See
// EnsureDefaultAdmin. Failures are logged, they must not stop the server from starting.
func SeedDefaultAdmin() {
	s := db.NewSession()
	defer s.Close()

	created, err := EnsureDefaultAdmin(s)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Could not create the default admin user: %s", err)
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("Could not create the default admin user: %s", err)
		return
	}
	if created {
		log.Warningf("Created the default admin user %q. Its password must be changed at the first login.", DefaultAdminUsername)
	}

	warnIfDefaultAdminPasswordUnchanged()
}

// warnIfDefaultAdminPasswordUnchanged logs on every start while the default admin still has to
// change its password, so the open door does not go unnoticed.
func warnIfDefaultAdminPasswordUnchanged() {
	s := db.NewReadSession()
	defer s.Close()

	pending, err := s.Where("is_admin = ? AND must_change_password = ? AND issuer = ?", true, true, user.IssuerLocal).
		Exist(&user.User{})
	if err != nil {
		log.Errorf("Could not check for admin users with an unchanged password: %s", err)
		return
	}
	if pending {
		log.Warningf("An admin user still has the initial password. Sign in and change it, and do not expose this instance before that.")
	}
}

// EnsureDefaultAdmin creates the local admin user `admin` on a brand-new installation. It does
// nothing when
//   - service.createdefaultadmin is off, or local login is disabled (a known password would be
//     pointless and a hidden account),
//   - there is any user at all. Existing instances are never changed, not even when none of their
//     users is an admin: the account has a published password, so it must not appear by surprise.
//     Create the first admin with `vikunja user create` and `user set-admin` there.
//
// The account has to change its password at the first login. It does not commit; the caller owns
// the transaction. It reports whether a user was created.
func EnsureDefaultAdmin(s *xorm.Session) (bool, error) {
	if !config.ServiceCreateDefaultAdmin.GetBool() || !config.AuthLocalEnabled.GetBool() {
		return false, nil
	}

	// Only a brand-new installation gets the account. An existing instance without an admin (for
	// example one where everybody signs in through Entra and nobody was ever promoted) must not
	// suddenly gain an admin with a published password just by being upgraded or restarted.
	hasUsers, err := s.Exist(&user.User{})
	if err != nil {
		return false, err
	}
	if hasUsers {
		return false, nil
	}

	email := strings.TrimSpace(config.ServiceDefaultAdminEmail.GetString())
	password := config.ServiceDefaultAdminPassword.GetString()

	// CreateUser hashes whatever it gets, the length rules only run in the API layer.
	if len(password) < 8 || len(password) > 72 {
		log.Warningf("Not creating the default admin user: service.defaultadminpassword must be between 8 and 72 bytes long.")
		return false, nil
	}
	if email == "" {
		log.Warningf("Not creating the default admin user: service.defaultadminemail is empty.")
		return false, nil
	}

	u, err := RegisterUser(s, &user.User{
		Username: DefaultAdminUsername,
		Email:    email,
		Password: password,
		Name:     "Administrator",
		IsAdmin:  true,
	}, user.CreateUserOptions{SkipEmailConfirm: true})
	if err != nil {
		return false, err
	}

	// CreateUser is also used for ordinary registrations, so the two flags are written
	// explicitly instead of relying on what it copies from the struct.
	_, err = s.
		Where("id = ?", u.ID).
		Cols("is_admin", "must_change_password").
		Update(&user.User{IsAdmin: true, MustChangePassword: true})
	if err != nil {
		return false, err
	}

	return true, nil
}
