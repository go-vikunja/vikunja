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

package handler

import (
	"context"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDoCreate_HappyPath creates a label through the framework-agnostic
// core and proves the row lands in the DB.
func TestDoCreate_HappyPath(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	u := &user.User{ID: 1}

	label := &models.Label{Title: "spike-label"}
	err := DoCreate(context.Background(), label, u)
	require.NoError(t, err)
	assert.NotZero(t, label.ID)
	assert.Equal(t, int64(1), label.CreatedByID)
}
