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

package avatar

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/ldap"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/modules/avatar/openid"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"
)

// Provider defines the avatar provider interface
type Provider interface {
	// GetAvatar is the method used to get an actual avatar for a user
	GetAvatar(user *user.User, size int64) (avatar []byte, mimeType string, err error)
	// FlushCache removes cached avatar data for the user
	FlushCache(u *user.User) error
}

// FlushAllCaches removes cached avatars for the given user for all providers
func FlushAllCaches(u *user.User) {
	providers := []Provider{
		&upload.Provider{},
		&gravatar.Provider{},
		&initials.Provider{},
		&ldap.Provider{},
		&openid.Provider{},
		&marble.Provider{},
		&empty.Provider{},
	}
	for _, p := range providers {
		if err := p.FlushCache(u); err != nil {
			log.Errorf("Error flushing avatar cache: %v", err)
		}
	}
}
