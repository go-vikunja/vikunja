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

package richtext

import (
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownToHTMLWithMentions(t *testing.T) {
	t.Run("known mention is rebuilt", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "hi @user1")
		require.NoError(t, err)
		assert.Equal(t, `<p>hi <mention-user data-id="user1" data-label="user1">@user1</mention-user></p>`, got)
	})

	t.Run("unknown mention stays literal text", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "hi @nosuchuser")
		require.NoError(t, err)
		assert.Equal(t, "<p>hi @nosuchuser</p>", got)
	})

	t.Run("mention next to punctuation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "cc @user1, please review")
		require.NoError(t, err)
		assert.Equal(t, `<p>cc <mention-user data-id="user1" data-label="user1">@user1</mention-user>, please review</p>`, got)
	})

	t.Run("multiple mentions resolve in one pass", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "ping @user1 and @user2")
		require.NoError(t, err)
		assert.Contains(t, got, `<mention-user data-id="user1" data-label="user1">@user1</mention-user>`)
		assert.Contains(t, got, `<mention-user data-id="user2" data-label="user2">@user2</mention-user>`)
	})

	t.Run("email is not a mention", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "reach me at user1@example.com please")
		require.NoError(t, err)
		assert.NotContains(t, got, "mention-user")
	})

	t.Run("mention inside code span is ignored", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "use `@user1` literally")
		require.NoError(t, err)
		assert.NotContains(t, got, "mention-user")
		assert.Contains(t, got, "<code>@user1</code>")
	})

	t.Run("mention inside task list item", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		got, err := MarkdownToHTMLWithMentions(s, "- [ ] ping @user1")
		require.NoError(t, err)
		assert.Contains(t, got, `data-type="taskItem"`)
		assert.Contains(t, got, `<mention-user data-id="user1" data-label="user1">@user1</mention-user>`)
	})

	t.Run("no session leaves mention as text", func(t *testing.T) {
		got, err := MarkdownToHTML("hi @user1")
		require.NoError(t, err)
		assert.Equal(t, "<p>hi @user1</p>", got)
	})
}
