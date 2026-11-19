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
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// setUserZone stores a user's time zone inside the test session. Fixture users have none.
func setUserZone(t *testing.T, s *xorm.Session, userID int64, zone string) {
	t.Helper()
	_, err := s.ID(userID).Cols("timezone").Update(&user.User{Timezone: zone})
	require.NoError(t, err)
}

func setProjectOwner(t *testing.T, s *xorm.Session, projectID, ownerID int64) {
	t.Helper()
	_, err := s.ID(projectID).Cols("owner_id").Update(&Project{OwnerID: ownerID})
	require.NoError(t, err)
}

// TestResolveTaskZone_Order: the completing user's zone, then the zone of the owner of the task's
// own project, then the server's; unusable names and failed lookups are skipped.
// Fixtures: project 2 is owned by user 3, link share 2 is on project 2, user 17 is disabled.
func TestResolveTaskZone_Order(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(t *testing.T, s *xorm.Session)
		auth      web.Auth
		projectID int64
		want      string
	}{
		{
			name: "a JWT-shaped doer is re-fetched for its zone",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 1, "America/New_York")
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &user.User{ID: 1, Username: "user1"},
			projectID: 2,
			want:      "America/New_York",
		},
		{
			name: "a doer without a zone falls through to the project owner",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 1, "")
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &user.User{ID: 1, Username: "user1"},
			projectID: 2,
			want:      "Europe/Berlin",
		},
		{
			name: "an invalid doer zone falls through to the project owner",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 1, "Mars/Olympus")
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &user.User{ID: 1, Username: "user1"},
			projectID: 2,
			want:      "Europe/Berlin",
		},
		{
			name: "the process zone name Local is not a user zone",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 1, "Local")
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &user.User{ID: 1, Username: "user1"},
			projectID: 2,
			want:      "Europe/Berlin",
		},
		{
			name: "a link share falls through to the project owner",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &LinkSharing{ID: 2},
			projectID: 2,
			want:      "Europe/Berlin",
		},
		{
			name: "a missing owner falls through to the server",
			setup: func(t *testing.T, s *xorm.Session) {
				setProjectOwner(t, s, 2, 999999)
			},
			auth:      &LinkSharing{ID: 2},
			projectID: 2,
			want:      config.GetTimeZone().String(),
		},
		{
			name: "a missing project falls through to the server",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &LinkSharing{ID: 2},
			projectID: 999999,
			want:      config.GetTimeZone().String(),
		},
		{
			name: "a disabled owner keeps its zone",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 17, "Asia/Kolkata")
				setProjectOwner(t, s, 2, 17)
			},
			auth:      &user.User{ID: 1, Username: "user1"},
			projectID: 2,
			want:      "Asia/Kolkata",
		},
		{
			name: "a doer that cannot be loaded does not lend the auth object's zone",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &user.User{ID: 999999, Username: "ghost", Timezone: "America/New_York"},
			projectID: 2,
			want:      "Europe/Berlin",
		},
		{
			name: "a disabled doer keeps its zone",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 17, "Pacific/Auckland")
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      &user.User{ID: 17},
			projectID: 2,
			want:      "Pacific/Auckland",
		},
		{
			name: "no auth falls through to the project owner",
			setup: func(t *testing.T, s *xorm.Session) {
				setUserZone(t, s, 3, "Europe/Berlin")
			},
			auth:      nil,
			projectID: 2,
			want:      "Europe/Berlin",
		},
		{
			name:      "nothing usable falls through to the server",
			setup:     func(*testing.T, *xorm.Session) {},
			auth:      &user.User{ID: 1, Username: "user1"},
			projectID: 2,
			want:      config.GetTimeZone().String(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// Zones are written before the first read, so no memoised row is stale.
			tc.setup(t, s)
			loc := resolveTaskZone(s, tc.auth, tc.projectID)
			require.NotNil(t, loc)
			assert.Equal(t, tc.want, loc.String())
		})
	}
}

// TestResolveTaskZone_LoadZoneCache: a zone name is loaded once, and an unusable one is refused.
func TestResolveTaskZone_LoadZoneCache(t *testing.T) {
	first, ok := loadZone("Europe/Berlin")
	require.True(t, ok)
	second, ok := loadZone("Europe/Berlin")
	require.True(t, ok)
	assert.Same(t, first, second)

	for _, name := range []string{"", "Local", "Mars/Olympus"} {
		loc, ok := loadZone(name)
		assert.False(t, ok, "%q", name)
		assert.Nil(t, loc, "%q", name)
	}
}
