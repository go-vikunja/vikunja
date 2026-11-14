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
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestLastModifiedIsUTC(t *testing.T) {
	updated := time.Date(2024, 6, 15, 23, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	task := &taskResource{task: &models.Task{Updated: updated}}

	value, found := taskProps.resolve(propContext{}, task, propGetLastModified)
	assert.True(t, found)
	assert.Equal(t, textValue("Sat, 15 Jun 2024 21:30:00 GMT"), value)
}
