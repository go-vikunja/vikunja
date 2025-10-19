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

package services

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUserService_NewUserProxyFromLinkShare(t *testing.T) {
	us := &UserService{}

	t.Run("should create a proxy user from a link share without a name", func(t *testing.T) {
		now := time.Now()
		share := &models.LinkSharing{
			ID:      1,
			Created: now,
			Updated: now,
		}

		user := us.NewUserProxyFromLinkShare(share)
		assert.NotNil(t, user)
		assert.Equal(t, int64(-1), user.ID)
		assert.Equal(t, "link-share-1", user.Username)
		assert.Equal(t, "Link Share", user.Name)
		assert.Equal(t, now, user.Created)
		assert.Equal(t, now, user.Updated)
	})

	t.Run("should create a proxy user from a link share with a name", func(t *testing.T) {
		now := time.Now()
		share := &models.LinkSharing{
			ID:      2,
			Name:    "My Share",
			Created: now,
			Updated: now,
		}

		user := us.NewUserProxyFromLinkShare(share)
		assert.NotNil(t, user)
		assert.Equal(t, int64(-2), user.ID)
		assert.Equal(t, "link-share-2", user.Username)
		assert.Equal(t, "My Share (Link Share)", user.Name)
		assert.Equal(t, now, user.Created)
		assert.Equal(t, now, user.Updated)
	})
}
