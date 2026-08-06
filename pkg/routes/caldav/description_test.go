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

package caldav

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDescriptionFromMarkdown(t *testing.T) {
	t.Run("unchanged round trip keeps stored html verbatim", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const stored = `<p>Hello <strong>world</strong></p>`
		vTask := &models.Task{Description: "Hello **world**"}

		require.NoError(t, applyDescriptionFromMarkdown(s, vTask, stored))
		assert.Equal(t, stored, vTask.Description)
	})

	t.Run("edited markdown is converted to html", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const stored = `<p>Hello <strong>world</strong></p>`
		vTask := &models.Task{Description: "Hello **mars**"}

		require.NoError(t, applyDescriptionFromMarkdown(s, vTask, stored))
		assert.Equal(t, "<p>Hello <strong>mars</strong></p>", vTask.Description)
	})

	t.Run("mention is rebuilt from markdown", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		vTask := &models.Task{Description: "ping @user1"}

		require.NoError(t, applyDescriptionFromMarkdown(s, vTask, ""))
		assert.Contains(t, vTask.Description, `<mention-user data-id="user1"`)
	})

	t.Run("new task markdown description becomes html", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		vTask := &models.Task{Description: "- [x] done"}

		require.NoError(t, applyDescriptionFromMarkdown(s, vTask, ""))
		assert.Contains(t, vTask.Description, `data-type="taskList"`)
		assert.Contains(t, vTask.Description, `data-checked="true"`)
		assert.Contains(t, vTask.Description, "<p>done</p>")
	})

	t.Run("emptying a description is honoured", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		vTask := &models.Task{Description: ""}

		require.NoError(t, applyDescriptionFromMarkdown(s, vTask, "<p>was here</p>"))
		assert.Empty(t, vTask.Description)
	})
}
