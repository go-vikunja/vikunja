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

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceContentIsRenderedOnce(t *testing.T) {
	task := &models.Task{UID: "uid-rendered-once", Title: "first"}
	rr := &VikunjaProjectResourceAdapter{task: task}

	content := rr.GetContent()
	require.Contains(t, content, "SUMMARY:first")

	// A second render would pick up the new title.
	task.Title = "second"
	assert.Equal(t, content, rr.GetContent())
	assert.Equal(t, int64(len(content)), rr.GetContentSize())
}
