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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/mail"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"xorm.io/builder"
	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

func inviteLinkSetup(t *testing.T) (*xorm.Session, *user.User) {
	t.Helper()
	adminActionsSetup(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureUserInvites})
	t.Cleanup(license.ResetForTests)
	s := db.NewSession()
	t.Cleanup(func() { events.CleanupPending(s); _ = s.Close() })
	_, err := s.ID(1).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)
	return s, &user.User{ID: 1}
}

func TestInviteLinkAdminCreate(t *testing.T) {
	s, admin := inviteLinkSetup(t)
	var queries bytes.Buffer
	logger := s.Engine().Logger()
	t.Cleanup(func() { s.Engine().SetLogger(logger) })
	queryLogger := xormlog.NewSimpleLogger(&queries)
	queryLogger.ShowSQL(true)
	s.Engine().SetLogger(queryLogger)
	link, err := CreateInviteLinkAsAdmin(s, admin, &CreateInviteLinkBody{Name: "  Welcome  ", TeamIDs: []int64{8, 1, 8}, SkipEmailConfirm: true})
	require.NoError(t, err)
	teamTable := s.Engine().Quote(s.Engine().TableName(&Team{}, true))
	require.Equal(t, 1, strings.Count(queries.String(), "FROM "+teamTable+" "), "invite teams should load in one query")
	require.Equal(t, "Welcome", link.Name)
	require.Len(t, link.ClearTextToken, 64)
	require.Equal(t, utils.Sha256Hex(link.ClearTextToken), link.TokenHash)
	require.Equal(t, []InviteLinkTeam{{ID: 8, Name: "testteam8"}, {ID: 1, Name: "testteam1"}}, link.Teams)
	require.NotNil(t, link.CreatedBy)
	require.Equal(t, "user1", link.CreatedBy.Username)
	require.Empty(t, link.CreatedBy.Email)
	stored := &UserInviteLink{}
	found, err := s.ID(link.ID).Get(stored)
	require.NoError(t, err)
	require.True(t, found)
	require.Empty(t, stored.ClearTextToken)
	require.Equal(t, link.TokenHash, stored.TokenHash)
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	require.Equal(t, link.ID, singleDispatchedEvent[*AdminInviteLinkCreatedEvent](t).Link.ID)
}

func TestInviteLinkAdminValidation(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	for _, tc := range []struct {
		name  string
		body  CreateInviteLinkBody
		want  error
		field string
	}{
		{"blank name", CreateInviteLinkBody{Name: "  "}, nil, "name"},
		{"past expiry", CreateInviteLinkBody{Name: "test", ExpiresAt: &past}, nil, "expires_at"},
		{"unknown team", CreateInviteLinkBody{Name: "test", TeamIDs: []int64{1, 999}}, ErrTeamDoesNotExist{TeamID: 999}, ""},
		{"unknown team order", CreateInviteLinkBody{Name: "test", TeamIDs: []int64{998, 999}}, ErrTeamDoesNotExist{TeamID: 998}, ""},
		{"external team", CreateInviteLinkBody{Name: "test", TeamIDs: []int64{1, 14}}, ErrInviteLinkExternalTeam{}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, admin := inviteLinkSetup(t)
			_, err := CreateInviteLinkAsAdmin(s, admin, &tc.body)
			if tc.field != "" {
				var validation ValidationHTTPError
				require.ErrorAs(t, err, &validation)
				require.Equal(t, ErrCodeInvalidData, validation.Code)
				require.Len(t, validation.InvalidFields, 1)
				require.True(t, strings.HasPrefix(validation.InvalidFields[0], tc.field))
			} else {
				require.ErrorIs(t, err, tc.want)
			}
			count, err := s.Count(&UserInviteLink{})
			require.NoError(t, err)
			require.EqualValues(t, 4, count)
		})
	}
}

func TestInviteLinkAdminListDelete(t *testing.T) {
	s, admin := inviteLinkSetup(t)
	links, total, err := ListInviteLinksAsAdmin(s, 1, 2)
	require.NoError(t, err)
	require.EqualValues(t, 4, total)
	require.Len(t, links, 2)
	require.NotEmpty(t, links[0].Teams)
	data, err := json.Marshal(links)
	require.NoError(t, err)
	require.NotContains(t, string(data), "token")
	require.NotContains(t, string(data), "created_by_id")
	var response []struct {
		CreatedBy *user.User `json:"created_by"`
	}
	require.NoError(t, json.Unmarshal(data, &response))
	for _, item := range response {
		require.NotNil(t, item.CreatedBy)
		require.EqualValues(t, 1, item.CreatedBy.ID)
		require.Equal(t, "user1", item.CreatedBy.Username)
		require.Empty(t, item.CreatedBy.Email)
	}

	require.NoError(t, DeleteInviteLinkAsAdmin(s, admin, 1))
	n, err := s.Where(builder.Eq{"invite_link_id": 1}).Count(&UserInviteLinkTeam{})
	require.NoError(t, err)
	require.Zero(t, n)
	require.ErrorIs(t, DeleteInviteLinkAsAdmin(s, admin, 999), ErrInviteLinkDoesNotExist{})
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	require.EqualValues(t, 1, singleDispatchedEvent[*AdminInviteLinkDeletedEvent](t).Link.ID)
}

func TestInviteLinkAdminListTeams(t *testing.T) {
	s, _ := inviteLinkSetup(t)
	_, err := s.Insert(&UserInviteLinkTeam{InviteLinkID: 3, TeamID: 8}, &UserInviteLinkTeam{InviteLinkID: 3, TeamID: 1})
	require.NoError(t, err)
	_, err = s.Where(builder.Eq{"invite_link_id": 4}).Delete(&UserInviteLinkTeam{})
	require.NoError(t, err)

	var queries bytes.Buffer
	logger := s.Engine().Logger()
	t.Cleanup(func() { s.Engine().SetLogger(logger) })
	queryLogger := xormlog.NewSimpleLogger(&queries)
	queryLogger.ShowSQL(true)
	s.Engine().SetLogger(queryLogger)

	links, total, err := ListInviteLinksAsAdmin(s, 1, 50)
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(queries.String(), "INNER JOIN"), "team associations should load in one query")
	require.EqualValues(t, 4, total)
	require.Len(t, links, 4)
	require.Equal(t, []InviteLinkTeam{}, links[0].Teams)
	require.Equal(t, []InviteLinkTeam{{ID: 1, Name: "testteam1"}, {ID: 3, Name: "testteam3_write_on_project7"}, {ID: 8, Name: "testteam8"}}, links[1].Teams)
	require.Equal(t, []InviteLinkTeam{{ID: 2, Name: "testteam2_read_only_on_project6"}}, links[2].Teams)
	require.Equal(t, []InviteLinkTeam{{ID: 1, Name: "testteam1"}}, links[3].Teams)

	links, total, err = ListInviteLinksAsAdmin(s, 3, 2)
	require.NoError(t, err)
	require.EqualValues(t, 4, total)
	require.Equal(t, []*UserInviteLink{}, links)
}

func TestInviteLinkAdminMissingCreator(t *testing.T) {
	s, _ := inviteLinkSetup(t)
	_, err := s.ID(4).Cols("created_by_id").Update(&UserInviteLink{CreatedByID: 999})
	require.NoError(t, err)
	links, total, err := ListInviteLinksAsAdmin(s, 1, 50)
	require.NoError(t, err)
	require.EqualValues(t, 4, total)
	require.EqualValues(t, 4, links[0].ID)
	require.Nil(t, links[0].CreatedBy)
	require.NotNil(t, links[1].CreatedBy)
}

func TestInviteLinkAdminTeams(t *testing.T) {
	s, admin := inviteLinkSetup(t)
	teams, total, err := ListTeamsAsAdmin(s, "testteam", 1, 50)
	require.NoError(t, err)
	require.Greater(t, total, int64(2))
	ids := []int64{}
	for _, team := range teams {
		ids = append(ids, team.ID)
	}
	require.Contains(t, ids, int64(8))
	require.NotContains(t, ids, int64(14))
	require.NotContains(t, ids, int64(15))
	teams, _, err = ListTeamsAsAdmin(s, "testteam8", 1, 1)
	require.NoError(t, err)
	require.Len(t, teams, 1)
	require.EqualValues(t, 8, teams[0].ID)
	teams, total, err = ListTeamsAsAdmin(s, "TeStTeAm8", 1, 50)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, []InviteLinkTeam{{ID: 8, Name: "testteam8"}}, teams)
	require.NoError(t, (&Team{ID: 1}).Delete(s, admin))
	n, err := s.Where(builder.Eq{"team_id": 1}).Count(&UserInviteLinkTeam{})
	require.NoError(t, err)
	require.Zero(t, n)
}

func TestInviteLinkRegistration(t *testing.T) {
	s, _ := inviteLinkSetup(t)
	link, err := GetInviteLinkByToken(s, "unlimited")
	require.NoError(t, err)
	require.Len(t, link.Teams, 1)
	u, err := RegisterUserViaInviteLink(s, "unlimited", &user.User{Username: "invitee", Email: "invitee@example.com", Password: "12345678"})
	require.NoError(t, err)
	require.Equal(t, user.StatusActive, u.Status)
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	events.AssertDispatched(t, &user.CreatedEvent{})
	event := singleDispatchedEvent[*TeamMemberAddedEvent](t)
	require.Equal(t, u.ID, event.Doer.ID)
	require.Equal(t, u.ID, event.Member.ID)
	db.AssertExists(t, "team_members", map[string]interface{}{"team_id": 1, "user_id": u.ID}, false)
	db.AssertExists(t, "user_invite_links", map[string]interface{}{"id": 1, "uses": 1}, false)
}

func TestInviteLinkUnavailable(t *testing.T) {
	for _, token := range []string{"unknown", "expired", "exhausted", "feature-off"} {
		t.Run(token, func(t *testing.T) {
			s, _ := inviteLinkSetup(t)
			if token == "feature-off" {
				license.ResetForTests()
				token = "unlimited"
			}
			_, err := GetInviteLinkByToken(s, token)
			require.ErrorIs(t, err, ErrInviteLinkInvalid{})
			_, err = RegisterUserViaInviteLink(s, token, &user.User{Username: "blocked", Email: "blocked@example.com", Password: "12345678"})
			require.ErrorIs(t, err, ErrInviteLinkInvalid{})
			exists, err := s.Where("username = ?", "blocked").Exist(&user.User{})
			require.NoError(t, err)
			require.False(t, exists)
		})
	}
}

func TestInviteLinkLastUse(t *testing.T) {
	s, _ := inviteLinkSetup(t)
	_, err := RegisterUserViaInviteLink(s, "last-slot", &user.User{Username: "last-one", Email: "last-one@example.com", Password: "12345678"})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	rs := db.NewSession()
	defer rs.Close()
	defer events.CleanupPending(rs)
	_, err = RegisterUserViaInviteLink(rs, "last-slot", &user.User{Username: "too-late", Email: "too-late@example.com", Password: "12345678"})
	require.ErrorIs(t, err, ErrInviteLinkInvalid{})
	require.NoError(t, rs.Rollback())
	db.AssertExists(t, "user_invite_links", map[string]interface{}{"id": 2, "uses": 2}, false)
}

func TestInviteLinkRollback(t *testing.T) {
	s, _ := inviteLinkSetup(t)
	existing, err := user.GetUserByID(s, 2)
	require.NoError(t, err)
	_, err = RegisterUserViaInviteLink(s, "unlimited", &user.User{Username: "rolled-back", Email: existing.Email, Password: "12345678"})
	require.Error(t, err)
	require.NoError(t, s.Rollback())
	events.CleanupPending(s)
	events.DispatchPending(context.Background(), s)
	require.Zero(t, events.CountDispatchedEvents((&user.CreatedEvent{}).Name()))
	require.Zero(t, events.CountDispatchedEvents((&TeamMemberAddedEvent{}).Name()))
	db.AssertExists(t, "user_invite_links", map[string]interface{}{"id": 1, "uses": 0}, false)
	db.AssertMissing(t, "users", map[string]interface{}{"username": "rolled-back"})
}

func TestInviteLinkConcurrentClaim(t *testing.T) {
	s, _ := inviteLinkSetup(t)
	require.NoError(t, s.Commit())
	start := make(chan struct{})
	results := make(chan error, 2)
	for i := range 2 {
		go func() {
			session := db.NewSession()
			defer session.Close()
			defer events.CleanupPending(session)
			<-start
			name := fmt.Sprintf("concurrent-invite-%d", i)
			_, err := RegisterUserViaInviteLink(session, "last-slot", &user.User{Username: name, Email: name + "@example.com", Password: "12345678"})
			if err == nil {
				err = session.Commit()
			}
			if err != nil {
				_ = session.Rollback()
			}
			results <- err
		}()
	}
	close(start)
	successes := 0
	for range 2 {
		if <-results == nil {
			successes++
		}
	}
	require.Equal(t, 1, successes)
	db.AssertExists(t, "user_invite_links", map[string]interface{}{"id": 2, "uses": 2}, false)
	rs := db.NewSession()
	defer rs.Close()
	n, err := rs.Where("username LIKE ?", "concurrent-invite-%").Count(&user.User{})
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
}

func TestInviteLinkConfirmation(t *testing.T) {
	for _, skip := range []bool{true, false} {
		t.Run(fmt.Sprint(skip), func(t *testing.T) {
			s, admin := inviteLinkSetup(t)
			oldMailer := config.MailerEnabled.GetBool()
			config.MailerEnabled.Set(true)
			t.Cleanup(func() { config.MailerEnabled.Set(oldMailer) })
			mail.Fake()
			t.Cleanup(mail.ResetSent)
			link, err := CreateInviteLinkAsAdmin(s, admin, &CreateInviteLinkBody{Name: "confirm", SkipEmailConfirm: skip})
			require.NoError(t, err)
			created, err := RegisterUserViaInviteLink(s, link.ClearTextToken, &user.User{Username: "confirm-invite", Email: "confirm-invite@example.com", Password: "12345678"})
			require.NoError(t, err)
			require.Nil(t, mail.LastSent())
			require.NoError(t, s.Commit())
			if skip {
				require.Equal(t, user.StatusActive, created.Status)
				require.Empty(t, mail.SentMails())
			} else {
				require.Equal(t, user.StatusEmailConfirmationRequired, created.Status)
				require.Len(t, mail.SentMails(), 1, "confirmation must be queued before Commit returns")
				require.Equal(t, "confirm-invite@example.com", mail.LastSent().To)
			}
		})
	}
}
