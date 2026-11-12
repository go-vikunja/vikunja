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
package webtests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func remindersJSON(count int) string {
	reminders := make([]string, count)
	for i := range reminders {
		at := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Minute)
		reminders[i] = fmt.Sprintf(`{"reminder":%q}`, at.Format(time.RFC3339))
	}
	return "[" + strings.Join(reminders, ",") + "]"
}

func TestTaskLimits(t *testing.T) {
	tooManyReminders := fmt.Sprintf(`{"title":"x","reminders":%s}`, remindersJSON(models.MaxRemindersPerTask+1))
	tooLongTitle := fmt.Sprintf(`{"title":%q}`, strings.Repeat("a", models.MaxTaskTitleLength+1))

	t.Run("v1", func(t *testing.T) {
		testHandler := webHandlerTest{
			user:    &testuser1,
			strFunc: func() handler.CObject { return &models.Task{} },
			t:       t,
		}

		t.Run("create", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "1"}, tooManyReminders)
			var verr models.ValidationHTTPError
			require.ErrorAs(t, err, &verr)
			assert.Equal(t, http.StatusPreconditionFailed, verr.HTTPCode)
			assert.Equal(t, models.ErrCodeInvalidData, verr.Code)
		})
		t.Run("update", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, tooLongTitle)
			var verr models.ValidationHTTPError
			require.ErrorAs(t, err, &verr)
			assert.Equal(t, http.StatusPreconditionFailed, verr.HTTPCode)
			assert.Equal(t, models.ErrCodeInvalidData, verr.Code)
		})
	})

	t.Run("v2", func(t *testing.T) {
		h := webHandlerTestV2{user: &testuser1, t: t}
		require.NoError(t, h.ensureEnv())
		token := humaTokenFor(t, &testuser1)
		errCode := fmt.Sprintf(`"code":%d`, models.ErrCodeInvalidData)

		t.Run("create", func(t *testing.T) {
			rec := humaRequest(t, h.e, http.MethodPost, "/api/v2/projects/1/tasks", tooManyReminders, token, "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), errCode)
		})
		t.Run("update", func(t *testing.T) {
			rec := humaRequest(t, h.e, http.MethodPut, "/api/v2/tasks/1", tooLongTitle, token, "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
			assert.Contains(t, rec.Body.String(), errCode)
		})
	})
}
