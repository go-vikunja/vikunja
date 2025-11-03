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
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"
)

type Provider struct{}

func (p *Provider) GetAvatar(user *user.User, size int64) (avatar []byte, mimeType string, err error) {
	up := upload.Provider{}

	return up.GetAvatar(user, size)
}

func (p *Provider) FlushCache(u *user.User) error {
	up := upload.Provider{}
	return up.FlushCache(u)
}
