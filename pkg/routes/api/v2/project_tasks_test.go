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

package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectTask(t *testing.T) {
	e, auth := web.NewTestContext(t)

	p := &models.Project{
		OwnerID: auth.User.ID,
		Title:   "Test Project",
	}
	err := p.Create(e.DB, auth)
	assert.NoError(t, err)

	task := &models.Task{
		Title: "Test Task",
	}
	body, err := json.Marshal(task)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/projects/{id}/tasks", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.GetIDString())

	err = CreateProjectTask(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var createdTask models.Task
	err = json.Unmarshal(rec.Body.Bytes(), &createdTask)
	assert.NoError(t, err)
	assert.Equal(t, task.Title, createdTask.Title)
	assert.Equal(t, p.ID, createdTask.ProjectID)
}
