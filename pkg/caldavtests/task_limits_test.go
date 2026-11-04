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
package caldavtests

import (
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestPUTRejectsTooManyAlarms(t *testing.T) {
	e := setupTestEnv(t)

	builder := NewVTodo("too-many-alarms", "Too many alarms")
	for i := 0; i <= models.MaxRemindersPerTask; i++ {
		builder.AlarmAbsolute(time.Date(2030, 1, 1, 0, i, 0, 0, time.UTC))
	}

	rec := caldavPUT(t, e, "/dav/projects/36/too-many-alarms.ics", builder.Build())
	assert.GreaterOrEqual(t, rec.Code, http.StatusBadRequest, "body: %s", rec.Body.String())

	rec = caldavRequest(t, e, http.MethodGet, "/dav/projects/36/too-many-alarms.ics", "", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
