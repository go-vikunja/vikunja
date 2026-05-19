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
	"fmt"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestExtractQuotedCommentIDs(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  []int64
	}{
		{"empty", "", nil},
		{"no blockquote", `<p>hello</p>`, []int64{}},
		{"plain blockquote without attr", `<blockquote>hi</blockquote>`, []int64{}},
		{"single attributed quote", `<blockquote data-comment-id="42">hi</blockquote>`, []int64{42}},
		{
			"nested inside paragraph",
			`<p><blockquote data-comment-id="7">hi</blockquote></p>`,
			[]int64{7},
		},
		{
			"two quotes - deduped order preserved",
			`<blockquote data-comment-id="3">a</blockquote><blockquote data-comment-id="5">b</blockquote><blockquote data-comment-id="3">a again</blockquote>`,
			[]int64{3, 5},
		},
		{"malformed - non-numeric", `<blockquote data-comment-id="abc">hi</blockquote>`, []int64{}},
		{"malformed - negative", `<blockquote data-comment-id="-1">hi</blockquote>`, []int64{}},
		{"malformed - zero", `<blockquote data-comment-id="0">hi</blockquote>`, []int64{}},
		{"malformed - empty", `<blockquote data-comment-id="">hi</blockquote>`, []int64{}},
		{
			"nested blockquote inside blockquote",
			`<blockquote data-comment-id="9"><blockquote data-comment-id="8">inner</blockquote></blockquote>`,
			[]int64{9, 8},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractQuotedCommentIDs(c.in)
			if c.out == nil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, c.out, got)
		})
	}
}

// notifQuery builds a where clause matching a TaskCommentNotification for a
// given subject + recipient.
func notifQuery(subjectID, userID int64) builder.Cond {
	return builder.And(
		builder.Eq{"subject_id": subjectID},
		builder.Eq{"notifiable_id": userID},
		builder.Eq{"name": (&TaskCommentNotification{}).Name()},
	)
}

func TestTaskComment_CommentReplies_Notifications(t *testing.T) {
	doer := &user.User{ID: 1}

	// task 32 is owned by user 1 (the doer) on project 3.
	// user 2 has access to project 3 (this is exercised by the existing
	// "should send notifications for comment mentions" test).

	t.Run("blockquote pointing at a same-task comment authored by another user notifies that user once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// existing comment by user 2 on task 32
		parent := &TaskComment{Comment: "original", TaskID: 32, AuthorID: 2}
		_, err := s.Insert(parent)
		require.NoError(t, err)

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: fmt.Sprintf(`<blockquote data-comment-id="%d">original</blockquote><p>thanks!</p>`, parent.ID),
			TaskID:  32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertCount(t, "notifications", notifQuery(tc.ID, 2), 1)
	})

	t.Run("blockquote and @mention referring to the same user still result in exactly one notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := &TaskComment{Comment: "original", TaskID: 32, AuthorID: 2}
		_, err := s.Insert(parent)
		require.NoError(t, err)

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: fmt.Sprintf(
				`<p><mention-user data-id="user2">@user2</mention-user></p><blockquote data-comment-id="%d">original</blockquote>`,
				parent.ID,
			),
			TaskID: 32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertCount(t, "notifications", notifQuery(tc.ID, 2), 1)
	})

	t.Run("two blockquotes pointing at comments by two different users each notify their author once", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parentB := &TaskComment{Comment: "by B", TaskID: 32, AuthorID: 2}
		_, err := s.Insert(parentB)
		require.NoError(t, err)
		parentC := &TaskComment{Comment: "by C", TaskID: 32, AuthorID: 3}
		_, err = s.Insert(parentC)
		require.NoError(t, err)

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: fmt.Sprintf(
				`<blockquote data-comment-id="%d">by B</blockquote><blockquote data-comment-id="%d">by C</blockquote>`,
				parentB.ID, parentC.ID,
			),
			TaskID: 32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertCount(t, "notifications", notifQuery(tc.ID, 2), 1)
		db.AssertCount(t, "notifications", notifQuery(tc.ID, 3), 1)
	})

	t.Run("blockquote pointing at a comment on a different task contributes nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user 2 authored a comment on a different task (task 1, owned by user 1).
		offTask := &TaskComment{Comment: "elsewhere", TaskID: 1, AuthorID: 2}
		_, err := s.Insert(offTask)
		require.NoError(t, err)

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: fmt.Sprintf(`<blockquote data-comment-id="%d">elsewhere</blockquote>`, offTask.ID),
			TaskID:  32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 2,
			"name":          (&TaskCommentNotification{}).Name(),
		})
	})

	t.Run("blockquote pointing at a missing comment is silently ignored", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: `<blockquote data-comment-id="99999">missing</blockquote>`,
			TaskID:  32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id": tc.ID,
			"name":       (&TaskCommentNotification{}).Name(),
		})
	})

	t.Run("blockquote pointing at the replier's own comment does not self-notify", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		mine := &TaskComment{Comment: "by me", TaskID: 32, AuthorID: 1}
		_, err := s.Insert(mine)
		require.NoError(t, err)

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: fmt.Sprintf(`<blockquote data-comment-id="%d">by me</blockquote>`, mine.ID),
			TaskID:  32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 1,
			"name":          (&TaskCommentNotification{}).Name(),
		})
	})

	t.Run("blockquote with non-integer data-comment-id triggers no DB lookup and no notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: `<blockquote data-comment-id="abc">malformed</blockquote>`,
			TaskID:  32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id": tc.ID,
			"name":       (&TaskCommentNotification{}).Name(),
		})
	})

	t.Run("blockquote nested deeper in the document is still counted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := &TaskComment{Comment: "original", TaskID: 32, AuthorID: 2}
		_, err := s.Insert(parent)
		require.NoError(t, err)

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)

		tc := &TaskComment{
			Comment: fmt.Sprintf(
				`<div><section><blockquote data-comment-id="%d">deep</blockquote></section></div>`,
				parent.ID,
			),
			TaskID: 32,
		}
		err = tc.Create(s, doer)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		events.TestListener(t, &TaskCommentCreatedEvent{
			Task:    &task,
			Doer:    doer,
			Comment: tc,
		}, &SendTaskCommentNotification{})

		db.AssertCount(t, "notifications", notifQuery(tc.ID, 2), 1)
	})
}
