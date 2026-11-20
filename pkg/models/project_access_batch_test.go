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
	"bytes"
	"regexp"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	xormlog "xorm.io/xorm/log"
)

func TestAccessLossBatchQueryCount(t *testing.T) {
	f := setupAccessLoss(t)
	other, otherWebhook := f.otherUserWithWebhook(t)
	var queries bytes.Buffer
	readQueries := regexp.MustCompile(`\[SQL\]\s*(?:SELECT|WITH)\b`)
	logger := f.s.Engine().Logger()
	t.Cleanup(func() { f.s.Engine().SetLogger(logger) })
	queryLogger := xormlog.NewSimpleLogger(&queries)
	queryLogger.ShowSQL(true)
	f.s.Engine().SetLogger(queryLogger)

	require.NoError(t, cleanupAfterProjectAccessLossForUsers(f.s, f.owner, []int64{f.member.ID}, []int64{f.parent.ID}, AccessLossCleanup{}))
	single := len(readQueries.FindAllString(queries.String(), -1))
	queries.Reset()
	require.NoError(t, cleanupAfterProjectAccessLossForUsers(f.s, f.owner, []int64{
		f.member.ID,
		other.ID,
		4,
	}, []int64{f.parent.ID}, AccessLossCleanup{}))
	batch := len(readQueries.FindAllString(queries.String(), -1))
	require.Positive(t, single)
	require.Equal(t, single, batch, "permission queries must not grow with candidate users")
	require.NoError(t, f.s.Commit())
	db.AssertMissing(t, "webhooks", map[string]interface{}{"id": otherWebhook})
}

func TestProjectMovePreservesPreviouslyInaccessibleData(t *testing.T) {
	f := setupAccessLoss(t)
	f.shareDirectly(t, PermissionWrite)
	other, oldWebhook := f.otherUserWithWebhook(t)
	oldShare := &LinkSharing{
		Hash:       "previously-inaccessible",
		ProjectID:  f.child.ID,
		SharedByID: other.ID,
		Permission: PermissionRead,
	}
	_, err := f.s.Insert(oldShare)
	require.NoError(t, err)
	require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{
		KeepWebhooks:     true,
		RevokeLinkShares: true,
	}, false))
	require.NoError(t, f.s.Commit())
	db.AssertExists(t, "webhooks", map[string]interface{}{
		"id":            oldWebhook,
		"created_by_id": other.ID,
	}, false)
	db.AssertExists(t, "link_shares", map[string]interface{}{
		"id":           oldShare.ID,
		"shared_by_id": other.ID,
	}, false)
	db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.shareC})
}

func TestProjectMoveDoesNotAskAboutPreviouslyInaccessibleData(t *testing.T) {
	f := setupAccessLoss(t)
	require.NoError(t, f.moveChildToTopLevel(t, AccessLossCleanup{}, true))
	require.NoError(t, f.s.Commit())
	db.AssertExists(t, "webhooks", map[string]interface{}{
		"id":            f.webhookC,
		"created_by_id": f.member.ID,
	}, false)
	db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.shareC}, false)
	db.AssertExists(t, "projects", map[string]interface{}{
		"id":                f.child.ID,
		"parent_project_id": nil,
	}, false)
}

func TestProjectReadAccessBatch(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	projects, err := GetProjectsMapByIDs(s, []int64{
		1,
		2,
		6,
		9,
		10,
		11,
		39,
		40,
		42,
		43,
	})
	require.NoError(t, err)
	access, err := getProjectReadAccessForUsers(s, []int64{
		1,
		3,
		14,
		23,
	}, projects)
	require.NoError(t, err)
	require.Equal(t, projectReadAccess{
		1: {
			1:  true,
			6:  true,
			9:  true,
			10: true,
			11: true,
			39: true,
			40: true,
			43: true,
		},
		3: {
			2:  true,
			40: true,
		},
		14: {42: true},
		23: {},
	}, access)
}

func TestProjectReadAccessBatchAdminLicense(t *testing.T) {
	for _, tt := range []struct {
		name     string
		licensed bool
		status   user.Status
		readable bool
	}{
		{
			name:     "licensed admin",
			licensed: true,
			readable: true,
		},
		{name: "unlicensed admin"},
		{
			name:     "disabled admin",
			licensed: true,
			status:   user.StatusDisabled,
		},
		{
			name:     "locked admin",
			licensed: true,
			status:   user.StatusAccountLocked,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := setupAccessLoss(t)
			license.SetForTests(nil)
			if tt.licensed {
				license.SetForTests([]license.Feature{license.FeatureAdminPanel})
			}
			t.Cleanup(license.ResetForTests)
			_, err := f.s.ID(f.member.ID).Cols("is_admin", "status").Update(&user.User{
				IsAdmin: true,
				Status:  tt.status,
			})
			require.NoError(t, err)
			access, err := getProjectReadAccessForUsers(f.s, []int64{
				f.member.ID,
				3,
			}, map[int64]*Project{f.child.ID: f.child})
			require.NoError(t, err)
			require.Equal(t, tt.readable, access[f.member.ID][f.child.ID])
			require.False(t, access[3][f.child.ID])
		})
	}
}
