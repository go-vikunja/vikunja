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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

func TestCreateParentNotNullIndex20260906010000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(&models.Project{}))

	require.NoError(t, createParentNotNullIndex20260906010000(x))
	// Idempotent.
	require.NoError(t, createParentNotNullIndex20260906010000(x))

	if x.Dialect().URI().DBType == schemas.MYSQL {
		return
	}
	var n int64
	switch x.Dialect().URI().DBType {
	case schemas.POSTGRES:
		_, err = x.SQL("SELECT count(*) FROM pg_indexes WHERE indexname = ?", parentNotNullIndex20260906010000).Get(&n)
	default:
		_, err = x.SQL("SELECT count(*) FROM sqlite_master WHERE type = 'index' AND name = ?", parentNotNullIndex20260906010000).Get(&n)
	}
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
}
