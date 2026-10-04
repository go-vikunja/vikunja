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
	"context"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// ---- helpers ----------------------------------------------------------------------------------------

func newRisk(t *testing.T, as web.Auth, projectID int64, risk *Risk) (*Risk, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()

	risk.ProjectID = projectID
	can, err := risk.CanCreate(s, as)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !can {
		_ = s.Rollback()
		return nil, ErrGenericForbidden{}
	}
	err = risk.Create(s, as)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return risk, nil
}

func mustRisk(t *testing.T, as web.Auth, projectID int64, risk *Risk) *Risk {
	t.Helper()
	created, err := newRisk(t, as, projectID, risk)
	require.NoError(t, err)
	return created
}

func riskStatus(t *testing.T, as web.Auth, id int64, status, note string) (*Risk, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	risk, err := ChangeRiskStatus(s, as, id, status, note)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return risk, nil
}

func listRisks(t *testing.T, as web.Auth, filter *Risk, search string) ([]*Risk, int64) {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	res, _, total, err := filter.ReadAll(s, as, search, 1, 100)
	require.NoError(t, err)
	risks, ok := res.([]*Risk)
	require.True(t, ok)
	return risks, total
}

func riskTitles(risks []*Risk) []string {
	out := make([]string, 0, len(risks))
	for _, r := range risks {
		out = append(out, r.Title)
	}
	return out
}

func updateRisk(t *testing.T, as web.Auth, risk *Risk) (*Risk, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	can, err := risk.CanUpdate(s, as)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !can {
		_ = s.Rollback()
		return nil, ErrGenericForbidden{}
	}
	err = risk.Update(s, as)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return risk, nil
}

func riskHistory(t *testing.T, as web.Auth, id int64) []*RiskStatusHistory {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	entries, _, err := ListRiskHistory(s, as, id, 1, 100)
	require.NoError(t, err)
	return entries
}

func archiveProject(t *testing.T, id int64) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(id).Cols("is_archived").Update(&Project{IsArchived: true})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// ---- rating -------------------------------------------------------------------------------------------

func TestRiskRating(t *testing.T) {
	// Every boundary of the bands. The frontend has the same table in helpers/riskRating.test.ts.
	cases := map[int]string{
		1: RiskRatingLow, 4: RiskRatingLow,
		5: RiskRatingMedium, 9: RiskRatingMedium,
		10: RiskRatingHigh, 16: RiskRatingHigh,
		17: RiskRatingCritical, 25: RiskRatingCritical,
		0: "", 26: "", -3: "",
	}
	for score, want := range cases {
		assert.Equal(t, want, RiskRatingForScore(score), "score %d", score)
	}

	assert.Equal(t, 20, RiskScore(4, 5))
	assert.Equal(t, 1, RiskScore(1, 1))
	assert.Equal(t, 25, RiskScore(5, 5))

	// Every score from 1 to 25 has exactly one rating, and the bands do not overlap or leave a gap.
	for score := 1; score <= 25; score++ {
		matches := 0
		for _, band := range riskRatingBands {
			if score >= band.Min && score <= band.Max {
				matches++
			}
		}
		assert.Equal(t, 1, matches, "score %d", score)
	}
}

func TestRiskStatusValidity(t *testing.T) {
	for _, status := range []string{"open", "mitigating", "accepted", "closed"} {
		assert.True(t, IsValidRiskStatus(status), status)
	}
	for _, status := range []string{"", "Open", "done", "resolved", "closed "} {
		assert.False(t, IsValidRiskStatus(status), status)
	}
}

// ---- permissions ----------------------------------------------------------------------------------------

func TestRiskPermissions(t *testing.T) {
	type people struct {
		owner, writer, reader, stranger *user.User
		project                          *Project
		risk                             *Risk
	}
	setup := func(t *testing.T) people {
		db.LoadAndAssertFixtures(t)
		p := people{
			owner:    registerPerson(t, "owner", "owner@corp.com"),
			writer:   registerPerson(t, "writer", "writer@corp.com"),
			reader:   registerPerson(t, "reader", "reader@corp.com"),
			stranger: registerPerson(t, "stranger", "stranger@corp.com"),
		}
		p.project = createProjectFor(t, p.owner, "Plan")
		shareProject(t, p.project.ID, p.writer.ID, PermissionWrite)
		shareProject(t, p.project.ID, p.reader.ID, PermissionRead)
		p.risk = mustRisk(t, p.owner, p.project.ID, &Risk{Title: "Supplier delay"})
		return p
	}

	t.Run("the project owner and a user with write access can do everything", func(t *testing.T) {
		p := setup(t)
		for _, who := range []*user.User{p.owner, p.writer} {
			s := db.NewSession()
			read, _, err := (&Risk{ID: p.risk.ID}).CanRead(s, who)
			require.NoError(t, err)
			assert.True(t, read, who.Username)
			create, err := (&Risk{ProjectID: p.project.ID}).CanCreate(s, who)
			require.NoError(t, err)
			assert.True(t, create, who.Username)
			update, err := (&Risk{ID: p.risk.ID}).CanUpdate(s, who)
			require.NoError(t, err)
			assert.True(t, update, who.Username)
			del, err := (&Risk{ID: p.risk.ID}).CanDelete(s, who)
			require.NoError(t, err)
			assert.True(t, del, who.Username)
			s.Close()
		}
	})
	t.Run("read access reads and nothing else", func(t *testing.T) {
		p := setup(t)
		s := db.NewSession()
		defer s.Close()

		read, maxPermission, err := (&Risk{ID: p.risk.ID}).CanRead(s, p.reader)
		require.NoError(t, err)
		assert.True(t, read)
		assert.Equal(t, int(PermissionRead), maxPermission)

		create, err := (&Risk{ProjectID: p.project.ID}).CanCreate(s, p.reader)
		require.NoError(t, err)
		assert.False(t, create)
		update, err := (&Risk{ID: p.risk.ID}).CanUpdate(s, p.reader)
		require.NoError(t, err)
		assert.False(t, update)
		del, err := (&Risk{ID: p.risk.ID}).CanDelete(s, p.reader)
		require.NoError(t, err)
		assert.False(t, del)

		_, err = riskStatus(t, p.reader, p.risk.ID, RiskStatusClosed, "")
		require.Error(t, err)
		assert.IsType(t, ErrGenericForbidden{}, err)
	})
	t.Run("a stranger can do nothing, and the risk stays as it was", func(t *testing.T) {
		p := setup(t)
		s := db.NewSession()
		defer s.Close()

		read, _, err := (&Risk{ID: p.risk.ID}).CanRead(s, p.stranger)
		require.NoError(t, err)
		assert.False(t, read)
		create, err := (&Risk{ProjectID: p.project.ID}).CanCreate(s, p.stranger)
		require.NoError(t, err)
		assert.False(t, create)
		update, err := (&Risk{ID: p.risk.ID}).CanUpdate(s, p.stranger)
		require.NoError(t, err)
		assert.False(t, update)
		del, err := (&Risk{ID: p.risk.ID}).CanDelete(s, p.stranger)
		require.NoError(t, err)
		assert.False(t, del)

		_, err = riskStatus(t, p.stranger, p.risk.ID, RiskStatusClosed, "")
		require.Error(t, err)
		_, err = newRisk(t, p.stranger, p.project.ID, &Risk{Title: "Planted"})
		require.Error(t, err)
		assert.Zero(t, countRows(t, &Risk{}, builder.Eq{"title": "Planted"}))
		db.AssertExists(t, "risks", map[string]interface{}{"id": p.risk.ID, "status": RiskStatusOpen}, false)
	})
	t.Run("access is inherited from a parent project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		owner := registerPerson(t, "owner", "owner@corp.com")
		reader := registerPerson(t, "reader", "reader@corp.com")
		parent := createProjectFor(t, owner, "Parent")
		s := db.NewSession()
		child := &Project{Title: "Child", ParentProjectID: parent.ID}
		require.NoError(t, child.Create(s, owner))
		require.NoError(t, s.Commit())
		s.Close()
		shareProject(t, parent.ID, reader.ID, PermissionRead)
		risk := mustRisk(t, owner, child.ID, &Risk{Title: "In the child"})

		s = db.NewSession()
		defer s.Close()
		read, _, err := (&Risk{ID: risk.ID}).CanRead(s, reader)
		require.NoError(t, err)
		assert.True(t, read)
		assert.Contains(t, riskTitles(firstOf(listRisks(t, reader, &Risk{}, ""))), "In the child")
	})
	t.Run("a link share has no access to risks", func(t *testing.T) {
		p := setup(t)
		share := &LinkSharing{ProjectID: p.project.ID, Permission: PermissionRead}
		s := db.NewSession()
		defer s.Close()

		read, _, err := (&Risk{ID: p.risk.ID}).CanRead(s, share)
		require.NoError(t, err)
		assert.False(t, read)
		write := &LinkSharing{ProjectID: p.project.ID, Permission: PermissionWrite}
		create, err := (&Risk{ProjectID: p.project.ID}).CanCreate(s, write)
		require.NoError(t, err)
		assert.False(t, create)
		update, err := (&Risk{ID: p.risk.ID}).CanUpdate(s, write)
		require.NoError(t, err)
		assert.False(t, update)

		risks, total := listRisks(t, share, &Risk{ProjectID: p.project.ID}, "")
		assert.Empty(t, risks)
		assert.Zero(t, total)

		_, err = riskStatus(t, write, p.risk.ID, RiskStatusClosed, "")
		require.Error(t, err)
		_, _, err = ListRiskHistory(s, share, p.risk.ID, 1, 10)
		require.Error(t, err)
	})
	t.Run("an archived project can be read, not changed", func(t *testing.T) {
		p := setup(t)
		archiveProject(t, p.project.ID)
		s := db.NewSession()
		defer s.Close()

		read, _, err := (&Risk{ID: p.risk.ID}).CanRead(s, p.owner)
		require.NoError(t, err)
		assert.True(t, read)

		_, err = (&Risk{ProjectID: p.project.ID}).CanCreate(s, p.owner)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err), "%v", err)
		_, err = (&Risk{ID: p.risk.ID}).CanUpdate(s, p.owner)
		require.Error(t, err)
		_, err = riskStatus(t, p.owner, p.risk.ID, RiskStatusClosed, "")
		require.Error(t, err)
		db.AssertExists(t, "risks", map[string]interface{}{"id": p.risk.ID, "status": RiskStatusOpen}, false)
	})
	t.Run("a risk or project that does not exist is not found", func(t *testing.T) {
		p := setup(t)
		s := db.NewSession()
		defer s.Close()

		_, _, err := (&Risk{ID: 99999}).CanRead(s, p.owner)
		require.Error(t, err)
		assert.True(t, IsErrRiskDoesNotExist(err))
		_, err = (&Risk{ProjectID: 99999}).CanCreate(s, p.owner)
		require.Error(t, err)
		_, err = riskStatus(t, p.owner, 99999, RiskStatusClosed, "")
		require.Error(t, err)
		assert.True(t, IsErrRiskDoesNotExist(err))
	})
}

func firstOf(risks []*Risk, _ int64) []*Risk { return risks }

// ---- create -----------------------------------------------------------------------------------------------

func TestRiskCreate(t *testing.T) {
	t.Run("defaults, the creator, the first history entry and the computed fields", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		events.ClearDispatchedEvents()

		risk := mustRisk(t, jane, project.ID, &Risk{Title: "  Supplier delay  ", Category: " Schedule "})

		assert.Equal(t, "Supplier delay", risk.Title, "trimmed")
		assert.Equal(t, "Schedule", risk.Category)
		assert.Equal(t, 3, risk.Probability)
		assert.Equal(t, 3, risk.Impact)
		assert.Equal(t, 9, risk.Score)
		assert.Equal(t, RiskRatingMedium, risk.Rating)
		assert.Equal(t, RiskStatusOpen, risk.Status)
		assert.Equal(t, jane.ID, risk.CreatedBy.ID)
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "project_id": project.ID, "created_by_id": jane.ID, "status": "open"}, false)

		history := riskHistory(t, jane, risk.ID)
		require.Len(t, history, 1)
		assert.Empty(t, history[0].FromStatus)
		assert.Equal(t, RiskStatusOpen, history[0].ToStatus)
		assert.Equal(t, jane.ID, history[0].ChangedBy.ID)

		assert.Equal(t, 1, events.CountDispatchedEvents((&RiskCreatedEvent{}).Name()))
	})
	t.Run("the status and closing fields of the request are not taken", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		closed := time.Now()

		risk := mustRisk(t, jane, project.ID, &Risk{
			Title: "Sneaky", Status: RiskStatusClosed, ClosedAt: &closed, ClosedByID: 5, Resolution: "done", ID: 4711,
		})

		assert.Equal(t, RiskStatusOpen, risk.Status)
		assert.NotEqual(t, int64(4711), risk.ID)
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "status": "open", "closed_by_id": 0, "resolution": ""}, false)
		assert.Nil(t, risk.ClosedAt)
	})
	t.Run("validation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		before := countRows(t, &Risk{}, builder.Expr("1 = 1"))
		later := time.Now().Add(24 * time.Hour)
		earlier := time.Now().Add(-24 * time.Hour)

		cases := map[string]*Risk{
			"empty title":        {Title: ""},
			"blank title":        {Title: "   "},
			"long title":         {Title: strings.Repeat("a", 251)},
			"long category":      {Title: "T", Category: strings.Repeat("c", 101)},
			"long description":   {Title: "T", Description: strings.Repeat("d", maxRiskTextLength+1)},
			"probability low":    {Title: "T", Probability: -1, Impact: 3},
			"probability high":   {Title: "T", Probability: 6, Impact: 3},
			"impact high":        {Title: "T", Probability: 3, Impact: 6},
			"negative owner":     {Title: "T", OwnerID: -4},
			"due before found":   {Title: "T", IdentifiedDate: &later, DueDate: &earlier},
		}
		for name, risk := range cases {
			_, err := newRisk(t, jane, project.ID, risk)
			require.Error(t, err, name)
		}
		assert.Equal(t, before, countRows(t, &Risk{}, builder.Expr("1 = 1")), "nothing was written")

		// The edges that are fine.
		_, err := newRisk(t, jane, project.ID, &Risk{Title: strings.Repeat("a", 250), Probability: 1, Impact: 5})
		require.NoError(t, err)
		_, err = newRisk(t, jane, project.ID, &Risk{Title: "Same day", IdentifiedDate: &later, DueDate: &later})
		require.NoError(t, err)
	})
	t.Run("the owner has to be an active person who can read the project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		member := registerPerson(t, "member", "member@corp.com")
		outsider := registerPerson(t, "outsider", "outsider@corp.com")
		gone := registerPerson(t, "gone", "gone@corp.com")
		project := createProjectFor(t, jane, "Plan")
		shareProject(t, project.ID, member.ID, PermissionRead)
		shareProject(t, project.ID, gone.ID, PermissionRead)
		s := db.NewSession()
		_, err := s.ID(gone.ID).Cols("status").Update(&user.User{Status: user.StatusDisabled})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		ok := mustRisk(t, jane, project.ID, &Risk{Title: "Owned", OwnerID: member.ID})
		assert.Equal(t, member.ID, ok.Owner.ID)
		_, err = newRisk(t, jane, project.ID, &Risk{Title: "Mine", OwnerID: jane.ID})
		require.NoError(t, err)

		for name, owner := range map[string]int64{"outsider": outsider.ID, "disabled": gone.ID, "bot": 23, "missing": 99999} {
			_, err = newRisk(t, jane, project.ID, &Risk{Title: "No " + name, OwnerID: owner})
			require.Error(t, err, name)
			assert.True(t, IsErrRiskOwnerHasNoAccess(err), "%s: %v", name, err)
		}
	})
}

// ---- update -----------------------------------------------------------------------------------------------

func TestRiskUpdate(t *testing.T) {
	t.Run("changes the editable fields and keeps the project, status and closing data", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		other := createProjectFor(t, jane, "Other")
		risk := mustRisk(t, jane, project.ID, &Risk{Title: "Old"})
		_, err := riskStatus(t, jane, risk.ID, RiskStatusClosed, "Fixed")
		require.NoError(t, err)
		due := time.Now().Add(72 * time.Hour)
		events.ClearDispatchedEvents()

		updated, err := updateRisk(t, jane, &Risk{
			ID: risk.ID, Title: " New ", Description: "d", Category: "Budget", Probability: 5, Impact: 4,
			Mitigation: "m", Contingency: "c", DueDate: &due,
			// None of these may take effect.
			ProjectID: other.ID, Status: RiskStatusOpen, Resolution: "hacked", ClosedByID: 99,
		})
		require.NoError(t, err)

		assert.Equal(t, "New", updated.Title)
		assert.Equal(t, 20, updated.Score)
		assert.Equal(t, RiskRatingCritical, updated.Rating)
		assert.Equal(t, project.ID, updated.ProjectID)
		assert.Equal(t, RiskStatusClosed, updated.Status)
		assert.Equal(t, "Fixed", updated.Resolution)
		require.NotNil(t, updated.DueDate)
		db.AssertExists(t, "risks", map[string]interface{}{
			"id": risk.ID, "title": "New", "category": "Budget", "probability": 5, "impact": 4, "mitigation": "m",
			"contingency": "c", "project_id": project.ID, "status": "closed", "resolution": "Fixed", "closed_by_id": jane.ID,
		}, false)
		assert.Equal(t, 1, events.CountDispatchedEvents((&RiskUpdatedEvent{}).Name()))
		assert.Len(t, riskHistory(t, jane, risk.ID), 2, "editing adds no history entry")
	})
	t.Run("a due date can be cleared", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		due := time.Now().Add(48 * time.Hour)
		risk := mustRisk(t, jane, project.ID, &Risk{Title: "Dated", DueDate: &due})
		require.NotNil(t, risk.DueDate)

		updated, err := updateRisk(t, jane, &Risk{ID: risk.ID, Title: "Dated", Probability: 3, Impact: 3})

		require.NoError(t, err)
		assert.Nil(t, updated.DueDate)
	})
	t.Run("validation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		risk := mustRisk(t, jane, project.ID, &Risk{Title: "Keep"})

		for name, in := range map[string]*Risk{
			"empty title":      {ID: risk.ID, Title: "", Probability: 3, Impact: 3},
			"zero probability": {ID: risk.ID, Title: "x", Probability: 0, Impact: 3},
			"impact too high":  {ID: risk.ID, Title: "x", Probability: 3, Impact: 9},
		} {
			_, err := updateRisk(t, jane, in)
			require.Error(t, err, name)
		}
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "title": "Keep"}, false)
	})
	t.Run("a new owner has to qualify, an owner who lost access is kept", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		member := registerPerson(t, "member", "member@corp.com")
		outsider := registerPerson(t, "outsider", "outsider@corp.com")
		project := createProjectFor(t, jane, "Plan")
		shareProject(t, project.ID, member.ID, PermissionRead)
		risk := mustRisk(t, jane, project.ID, &Risk{Title: "Owned", OwnerID: member.ID})

		_, err := updateRisk(t, jane, &Risk{ID: risk.ID, Title: "Owned", Probability: 3, Impact: 3, OwnerID: outsider.ID})
		require.Error(t, err)
		assert.True(t, IsErrRiskOwnerHasNoAccess(err))

		// The member loses access. Editing something else leaves the owner alone.
		s := db.NewSession()
		_, err = s.Where("project_id = ? AND user_id = ?", project.ID, member.ID).Delete(&ProjectUser{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		updated, err := updateRisk(t, jane, &Risk{ID: risk.ID, Title: "Renamed", Probability: 3, Impact: 3, OwnerID: member.ID})
		require.NoError(t, err)
		assert.Equal(t, member.ID, updated.OwnerID)

		// Clearing the owner always works.
		cleared, err := updateRisk(t, jane, &Risk{ID: risk.ID, Title: "Renamed", Probability: 3, Impact: 3})
		require.NoError(t, err)
		assert.Zero(t, cleared.OwnerID)
		assert.Nil(t, cleared.Owner)
	})
}

// ---- status -----------------------------------------------------------------------------------------------

func TestRiskStatusChanges(t *testing.T) {
	setup := func(t *testing.T) (*user.User, *user.User, *Risk) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		project := createProjectFor(t, jane, "Plan")
		shareProject(t, project.ID, joe.ID, PermissionWrite)
		return jane, joe, mustRisk(t, jane, project.ID, &Risk{Title: "Supplier delay"})
	}

	t.Run("closing records who, when and why", func(t *testing.T) {
		_, joe, risk := setup(t)
		events.ClearDispatchedEvents()
		before := time.Now().Add(-time.Second)

		closed, err := riskStatus(t, joe, risk.ID, RiskStatusClosed, "  Supplier delivered  ")

		require.NoError(t, err)
		assert.Equal(t, RiskStatusClosed, closed.Status)
		require.NotNil(t, closed.ClosedAt)
		assert.True(t, closed.ClosedAt.After(before))
		assert.Equal(t, joe.ID, closed.ClosedBy.ID)
		assert.Equal(t, "Supplier delivered", closed.Resolution)
		assert.Equal(t, 1, events.CountDispatchedEvents((&RiskStatusChangedEvent{}).Name()))
	})
	t.Run("a closed risk can be reopened, to open or to any other status", func(t *testing.T) {
		jane, _, risk := setup(t)
		_, err := riskStatus(t, jane, risk.ID, RiskStatusClosed, "Fixed")
		require.NoError(t, err)

		reopened, err := riskStatus(t, jane, risk.ID, RiskStatusOpen, "It came back")

		require.NoError(t, err)
		assert.Equal(t, RiskStatusOpen, reopened.Status)
		assert.Nil(t, reopened.ClosedAt)
		assert.Nil(t, reopened.ClosedBy)
		assert.Empty(t, reopened.Resolution, "the resolution of the closing is cleared")
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "status": "open", "closed_by_id": 0, "resolution": ""}, false)

		for _, target := range []string{RiskStatusMitigating, RiskStatusAccepted} {
			_, err = riskStatus(t, jane, risk.ID, RiskStatusClosed, "again")
			require.NoError(t, err)
			again, err := riskStatus(t, jane, risk.ID, target, "")
			require.NoError(t, err)
			assert.Equal(t, target, again.Status)
			assert.Nil(t, again.ClosedAt)
		}
	})
	t.Run("every status can follow every other", func(t *testing.T) {
		jane, _, risk := setup(t)
		for _, from := range riskStatuses {
			for _, to := range riskStatuses {
				_, err := riskStatus(t, jane, risk.ID, from, "")
				require.NoError(t, err, "to %s", from)
				got, err := riskStatus(t, jane, risk.ID, to, "")
				require.NoError(t, err, "%s to %s", from, to)
				assert.Equal(t, to, got.Status, "%s to %s", from, to)
				if to == RiskStatusClosed {
					assert.NotNil(t, got.ClosedAt)
				} else {
					assert.Nil(t, got.ClosedAt, "%s to %s", from, to)
				}
			}
		}
	})
	t.Run("the history lists every change, newest first, with the note", func(t *testing.T) {
		jane, joe, risk := setup(t)
		_, err := riskStatus(t, jane, risk.ID, RiskStatusMitigating, "Plan made")
		require.NoError(t, err)
		_, err = riskStatus(t, joe, risk.ID, RiskStatusClosed, "Fixed")
		require.NoError(t, err)
		_, err = riskStatus(t, jane, risk.ID, RiskStatusOpen, "Back")
		require.NoError(t, err)

		history := riskHistory(t, jane, risk.ID)

		require.Len(t, history, 4)
		assert.Equal(t, []string{"open", "closed", "mitigating", "open"}, []string{history[0].ToStatus, history[1].ToStatus, history[2].ToStatus, history[3].ToStatus})
		assert.Equal(t, "closed", history[0].FromStatus)
		assert.Equal(t, "Back", history[0].Note)
		assert.Equal(t, joe.ID, history[1].ChangedBy.ID)
		assert.Equal(t, "Fixed", history[1].Note)
		assert.Empty(t, history[3].FromStatus, "the creation")
	})
	t.Run("the same status again changes nothing and is not recorded", func(t *testing.T) {
		jane, _, risk := setup(t)
		events.ClearDispatchedEvents()

		same, err := riskStatus(t, jane, risk.ID, RiskStatusOpen, "noop")

		require.NoError(t, err)
		assert.Equal(t, RiskStatusOpen, same.Status)
		assert.Len(t, riskHistory(t, jane, risk.ID), 1)
		assert.Zero(t, events.CountDispatchedEvents((&RiskStatusChangedEvent{}).Name()))
	})
	t.Run("an unknown status is refused and nothing changes", func(t *testing.T) {
		jane, _, risk := setup(t)
		for _, status := range []string{"", "done", "Closed", "reopened"} {
			_, err := riskStatus(t, jane, risk.ID, status, "")
			require.Error(t, err, status)
			assert.True(t, IsErrInvalidRiskStatus(err), "%q: %v", status, err)
		}
		_, err := riskStatus(t, jane, risk.ID, RiskStatusClosed, strings.Repeat("n", maxRiskTextLength+1))
		require.Error(t, err)
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "status": "open"}, false)
		assert.Len(t, riskHistory(t, jane, risk.ID), 1)
	})
	t.Run("the status cannot be changed through an update, and only the model writes it", func(t *testing.T) {
		jane, _, risk := setup(t)
		_, err := updateRisk(t, jane, &Risk{ID: risk.ID, Title: "x", Probability: 3, Impact: 3, Status: RiskStatusClosed})
		require.NoError(t, err)
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "status": "open"}, false)
		assert.Len(t, riskHistory(t, jane, risk.ID), 1)
	})
	t.Run("a stale status is a conflict, not a second history entry", func(t *testing.T) {
		jane, _, risk := setup(t)
		// The risk was read as open. Somebody else closes it before the change is written.
		_, err := riskStatus(t, jane, risk.ID, RiskStatusClosed, "first")
		require.NoError(t, err)

		s := db.NewSession()
		defer s.Close()
		affected, err := s.
			Where("id = ? AND status = ?", risk.ID, RiskStatusOpen).
			Cols("status").
			Update(&Risk{Status: RiskStatusMitigating})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		assert.Zero(t, affected, "the compare-and-set finds a different status and writes nothing")
		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "status": "closed"}, false)
		assert.Len(t, riskHistory(t, jane, risk.ID), 2)
	})
}

// ---- list -------------------------------------------------------------------------------------------------

func TestRiskList(t *testing.T) {
	type world struct {
		jane, joe, ann *user.User
		a, b, hidden   *Project
	}
	setup := func(t *testing.T) world {
		db.LoadAndAssertFixtures(t)
		w := world{
			jane: registerPerson(t, "jane", "jane@corp.com"),
			joe:  registerPerson(t, "joe", "joe@corp.com"),
			ann:  registerPerson(t, "ann", "ann@corp.com"),
		}
		w.a = createProjectFor(t, w.jane, "Alpha")
		w.b = createProjectFor(t, w.jane, "Beta")
		w.hidden = createProjectFor(t, w.joe, "Joes secret")
		shareProject(t, w.a.ID, w.joe.ID, PermissionWrite)

		overdue := time.Now().Add(-48 * time.Hour)
		soon := time.Now().Add(48 * time.Hour)
		mustRisk(t, w.jane, w.a.ID, &Risk{Title: "Delay", Category: "Schedule", Probability: 5, Impact: 5, DueDate: &overdue, OwnerID: w.joe.ID, Description: "supplier is late"})
		mustRisk(t, w.jane, w.a.ID, &Risk{Title: "Budget cut", Category: "budget", Probability: 2, Impact: 2, DueDate: &soon})
		mustRisk(t, w.jane, w.b.ID, &Risk{Title: "Staff leaves", Category: "People", Probability: 3, Impact: 4, Mitigation: "cross training"})
		closed := mustRisk(t, w.jane, w.b.ID, &Risk{Title: "Old news", Category: "Schedule", Probability: 1, Impact: 1, DueDate: &overdue})
		_, err := riskStatus(t, w.jane, closed.ID, RiskStatusClosed, "")
		require.NoError(t, err)
		mustRisk(t, w.joe, w.hidden.ID, &Risk{Title: "Secret", Probability: 5, Impact: 5})
		return w
	}

	t.Run("a list over all projects holds only what the user can read, worst first", func(t *testing.T) {
		w := setup(t)

		risks, total := listRisks(t, w.jane, &Risk{}, "")
		assert.Equal(t, []string{"Delay", "Staff leaves", "Budget cut", "Old news"}, riskTitles(risks))
		assert.Equal(t, int64(4), total)

		joes, _ := listRisks(t, w.joe, &Risk{}, "")
		assert.ElementsMatch(t, []string{"Delay", "Budget cut", "Secret"}, riskTitles(joes), "alpha is shared, beta is not")

		anns, total := listRisks(t, w.ann, &Risk{}, "")
		assert.Empty(t, anns)
		assert.Zero(t, total)
	})
	t.Run("one project", func(t *testing.T) {
		w := setup(t)
		risks, total := listRisks(t, w.jane, &Risk{ProjectID: w.b.ID}, "")
		assert.Equal(t, []string{"Staff leaves", "Old news"}, riskTitles(risks))
		assert.Equal(t, int64(2), total)

		none, _ := listRisks(t, w.ann, &Risk{ProjectID: w.b.ID}, "")
		assert.Empty(t, none, "no access, no risks, no error")
		missing, _ := listRisks(t, w.jane, &Risk{ProjectID: 99999}, "")
		assert.Empty(t, missing)
	})
	t.Run("project_ids narrows and can never widen access", func(t *testing.T) {
		w := setup(t)
		risks, _ := listRisks(t, w.jane, &Risk{ProjectIDs: []int64{w.a.ID}}, "")
		assert.ElementsMatch(t, []string{"Delay", "Budget cut"}, riskTitles(risks))

		// Asking for a project jane cannot read adds nothing.
		risks, _ = listRisks(t, w.jane, &Risk{ProjectIDs: []int64{w.a.ID, w.hidden.ID}}, "")
		assert.ElementsMatch(t, []string{"Delay", "Budget cut"}, riskTitles(risks))
		risks, total := listRisks(t, w.jane, &Risk{ProjectIDs: []int64{w.hidden.ID}}, "")
		assert.Empty(t, risks)
		assert.Zero(t, total)
	})
	t.Run("status", func(t *testing.T) {
		w := setup(t)
		open, _ := listRisks(t, w.jane, &Risk{Statuses: []string{RiskStatusOpen}}, "")
		assert.ElementsMatch(t, []string{"Delay", "Budget cut", "Staff leaves"}, riskTitles(open))
		closed, _ := listRisks(t, w.jane, &Risk{Statuses: []string{RiskStatusClosed}}, "")
		assert.Equal(t, []string{"Old news"}, riskTitles(closed))
		both, _ := listRisks(t, w.jane, &Risk{Statuses: []string{RiskStatusClosed, RiskStatusOpen}}, "")
		assert.Len(t, both, 4)

		s := db.NewReadSession()
		defer s.Close()
		_, _, _, err := (&Risk{Statuses: []string{"nonsense"}}).ReadAll(s, w.jane, "", 1, 10)
		require.Error(t, err, "an unknown value is an error, not an empty list")
		assert.True(t, IsErrInvalidRiskStatus(err))
	})
	t.Run("rating", func(t *testing.T) {
		w := setup(t)
		for rating, want := range map[string][]string{
			RiskRatingCritical: {"Delay"},
			RiskRatingHigh:     {"Staff leaves"},
			RiskRatingLow:      {"Budget cut", "Old news"},
			RiskRatingMedium:   {},
		} {
			risks, _ := listRisks(t, w.jane, &Risk{Ratings: []string{rating}}, "")
			assert.ElementsMatch(t, want, riskTitles(risks), rating)
		}
		two, _ := listRisks(t, w.jane, &Risk{Ratings: []string{RiskRatingCritical, RiskRatingHigh}}, "")
		assert.ElementsMatch(t, []string{"Delay", "Staff leaves"}, riskTitles(two))

		s := db.NewReadSession()
		defer s.Close()
		_, _, _, err := (&Risk{Ratings: []string{"huge"}}).ReadAll(s, w.jane, "", 1, 10)
		require.Error(t, err)
	})
	t.Run("owner, including unowned", func(t *testing.T) {
		w := setup(t)
		owned, _ := listRisks(t, w.jane, &Risk{OwnerIDs: []int64{w.joe.ID}}, "")
		assert.Equal(t, []string{"Delay"}, riskTitles(owned))
		unowned, _ := listRisks(t, w.jane, &Risk{OwnerIDs: []int64{0}}, "")
		assert.ElementsMatch(t, []string{"Budget cut", "Staff leaves", "Old news"}, riskTitles(unowned))
		either, _ := listRisks(t, w.jane, &Risk{OwnerIDs: []int64{0, w.joe.ID}}, "")
		assert.Len(t, either, 4)
	})
	t.Run("category without regard to case", func(t *testing.T) {
		w := setup(t)
		risks, _ := listRisks(t, w.jane, &Risk{Categories: []string{"SCHEDULE"}}, "")
		assert.ElementsMatch(t, []string{"Delay", "Old news"}, riskTitles(risks))
		risks, _ = listRisks(t, w.jane, &Risk{Categories: []string{"budget", "people"}}, "")
		assert.ElementsMatch(t, []string{"Budget cut", "Staff leaves"}, riskTitles(risks))
	})
	t.Run("overdue leaves out closed and future risks", func(t *testing.T) {
		w := setup(t)
		risks, _ := listRisks(t, w.jane, &Risk{Overdue: true}, "")
		assert.Equal(t, []string{"Delay"}, riskTitles(risks), "Old news is overdue too but closed")
	})
	t.Run("search looks at the title and the text fields", func(t *testing.T) {
		w := setup(t)
		for search, want := range map[string][]string{
			"delay":    {"Delay"},
			"SUPPLIER": {"Delay"},
			"cross":    {"Staff leaves"},
			"budget":   {"Budget cut"},
			"nothing":  {},
		} {
			risks, _ := listRisks(t, w.jane, &Risk{}, search)
			assert.ElementsMatch(t, want, riskTitles(risks), search)
		}
	})
	t.Run("filters combine", func(t *testing.T) {
		w := setup(t)
		risks, _ := listRisks(t, w.jane, &Risk{Categories: []string{"Schedule"}, Statuses: []string{RiskStatusOpen}, Ratings: []string{RiskRatingCritical}}, "")
		assert.Equal(t, []string{"Delay"}, riskTitles(risks))
		risks, _ = listRisks(t, w.jane, &Risk{ProjectID: w.b.ID, Statuses: []string{RiskStatusOpen}}, "")
		assert.Equal(t, []string{"Staff leaves"}, riskTitles(risks))
	})
	t.Run("sorting and paging", func(t *testing.T) {
		w := setup(t)
		byTitle, _ := listRisks(t, w.jane, &Risk{SortBy: []string{"title"}, OrderBy: []string{"asc"}}, "")
		assert.Equal(t, []string{"Budget cut", "Delay", "Old news", "Staff leaves"}, riskTitles(byTitle))
		byTitleDesc, _ := listRisks(t, w.jane, &Risk{SortBy: []string{"title"}, OrderBy: []string{"desc"}}, "")
		assert.Equal(t, []string{"Staff leaves", "Old news", "Delay", "Budget cut"}, riskTitles(byTitleDesc))
		byImpact, _ := listRisks(t, w.jane, &Risk{SortBy: []string{"impact", "title"}, OrderBy: []string{"desc", "asc"}}, "")
		assert.Equal(t, "Delay", byImpact[0].Title)

		s := db.NewReadSession()
		defer s.Close()
		res, count, total, err := (&Risk{}).ReadAll(s, w.jane, "", 2, 3)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, int64(4), total, "the total ignores the page")
		assert.Len(t, res.([]*Risk), 1)

		_, _, _, err = (&Risk{SortBy: []string{"id; DROP TABLE risks"}}).ReadAll(s, w.jane, "", 1, 10)
		require.Error(t, err, "only known columns can be sorted by")
		_, _, _, err = (&Risk{SortBy: []string{"title"}, OrderBy: []string{"sideways"}}).ReadAll(s, w.jane, "", 1, 10)
		require.Error(t, err)
	})
	t.Run("archived projects only with include_archived, or when chosen", func(t *testing.T) {
		w := setup(t)
		archiveProject(t, w.b.ID)

		hidden, _ := listRisks(t, w.jane, &Risk{}, "")
		assert.ElementsMatch(t, []string{"Delay", "Budget cut"}, riskTitles(hidden))
		with, _ := listRisks(t, w.jane, &Risk{IncludeArchived: true}, "")
		assert.Len(t, with, 4)
		chosen, _ := listRisks(t, w.jane, &Risk{ProjectIDs: []int64{w.b.ID}}, "")
		assert.Len(t, chosen, 2, "a project that was chosen is shown")
		single, _ := listRisks(t, w.jane, &Risk{ProjectID: w.b.ID}, "")
		assert.Len(t, single, 2)
	})
	t.Run("a child of an archived project counts as archived", func(t *testing.T) {
		w := setup(t)
		s := db.NewSession()
		child := &Project{Title: "Child of Beta", ParentProjectID: w.b.ID}
		require.NoError(t, child.Create(s, w.jane))
		require.NoError(t, s.Commit())
		s.Close()
		mustRisk(t, w.jane, child.ID, &Risk{Title: "In the child"})
		archiveProject(t, w.b.ID)

		risks, _ := listRisks(t, w.jane, &Risk{}, "")
		assert.NotContains(t, riskTitles(risks), "In the child")
		with, _ := listRisks(t, w.jane, &Risk{IncludeArchived: true}, "")
		assert.Contains(t, riskTitles(with), "In the child")
	})
	t.Run("the people on a risk are loaded and carry no email", func(t *testing.T) {
		w := setup(t)
		risks, _ := listRisks(t, w.jane, &Risk{OwnerIDs: []int64{w.joe.ID}}, "")
		require.Len(t, risks, 1)
		require.NotNil(t, risks[0].Owner)
		assert.Equal(t, w.joe.ID, risks[0].Owner.ID)
		assert.Empty(t, risks[0].Owner.Email)
		assert.Equal(t, w.jane.ID, risks[0].CreatedBy.ID)
		assert.Equal(t, 25, risks[0].Score)
		assert.Equal(t, RiskRatingCritical, risks[0].Rating)
	})
}

// ---- delete and cleanup ---------------------------------------------------------------------------------------

func TestRiskDelete(t *testing.T) {
	t.Run("removes the risk and its history", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, jane, "Plan")
		risk := mustRisk(t, jane, project.ID, &Risk{Title: "Gone"})
		_, err := riskStatus(t, jane, risk.ID, RiskStatusClosed, "")
		require.NoError(t, err)
		events.ClearDispatchedEvents()

		s := db.NewSession()
		target := &Risk{ID: risk.ID}
		can, err := target.CanDelete(s, jane)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, target.Delete(s, jane))
		require.NoError(t, s.Commit())
		events.DispatchPending(context.Background(), s)
		s.Close()
		assert.Equal(t, 1, events.CountDispatchedEvents((&RiskDeletedEvent{}).Name()))

		assert.Zero(t, countRows(t, &Risk{}, builder.Eq{"id": risk.ID}))
		assert.Zero(t, countRows(t, &RiskStatusHistory{}, builder.Eq{"risk_id": risk.ID}))
	})
	t.Run("deleting a project removes its risks and history, and nobody else's", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		doomed := createProjectFor(t, jane, "Doomed")
		kept := createProjectFor(t, jane, "Kept")
		child := &Project{Title: "Doomed child", ParentProjectID: doomed.ID}
		s := db.NewSession()
		require.NoError(t, child.Create(s, jane))
		require.NoError(t, s.Commit())
		s.Close()
		a := mustRisk(t, jane, doomed.ID, &Risk{Title: "In the project"})
		b := mustRisk(t, jane, child.ID, &Risk{Title: "In the child"})
		c := mustRisk(t, jane, kept.ID, &Risk{Title: "Survives"})

		s = db.NewSession()
		require.NoError(t, (&Project{ID: doomed.ID}).Delete(s, jane))
		require.NoError(t, s.Commit())
		s.Close()

		assert.Zero(t, countRows(t, &Risk{}, builder.In("id", a.ID, b.ID)))
		assert.Zero(t, countRows(t, &RiskStatusHistory{}, builder.In("risk_id", a.ID, b.ID)))
		assert.Equal(t, int64(1), countRows(t, &Risk{}, builder.Eq{"id": c.ID}))
		assert.Equal(t, int64(1), countRows(t, &RiskStatusHistory{}, builder.Eq{"risk_id": c.ID}))
	})
	t.Run("deleting a user clears the risks they owned and keeps the risks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		project := createProjectFor(t, jane, "Plan")
		shareProject(t, project.ID, joe.ID, PermissionWrite)
		risk := mustRisk(t, joe, project.ID, &Risk{Title: "Joe owns it", OwnerID: joe.ID})
		_, err := riskStatus(t, joe, risk.ID, RiskStatusClosed, "done")
		require.NoError(t, err)

		s := db.NewSession()
		require.NoError(t, DeleteUser(s, joe))
		require.NoError(t, s.Commit())
		s.Close()

		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "owner_id": 0, "created_by_id": joe.ID}, false)
		risks, _ := listRisks(t, jane, &Risk{ProjectID: project.ID}, "")
		require.Len(t, risks, 1)
		assert.Nil(t, risks[0].Owner)
		assert.Nil(t, risks[0].CreatedBy, "shown as a user that does not exist any more")
		history := riskHistory(t, jane, risk.ID)
		require.Len(t, history, 2)
		assert.Nil(t, history[0].ChangedBy)
	})
}
