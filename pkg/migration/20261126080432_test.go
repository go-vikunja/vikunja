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
	"slices"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type apiTokensBefore20261126080432 struct {
	ID             int64     `xorm:"bigint autoincr not null unique pk"`
	Title          string    `xorm:"not null"`
	TokenSalt      string    `xorm:"null"`
	TokenHash      string    `xorm:"null unique"`
	TokenLastEight string    `xorm:"null index varchar(8)"`
	TokenSha256    string    `xorm:"varchar(64) null unique"`
	Permissions    string    `xorm:"json not null"`
	ExpiresAt      time.Time `xorm:"not null"`
	Created        time.Time `xorm:"created not null"`
	OwnerID        int64     `xorm:"bigint not null"`
}

func (apiTokensBefore20261126080432) TableName() string { return "api_tokens" }

type oauthCodesBefore20261126080432 struct {
	ID                  int64     `xorm:"autoincr not null unique pk"`
	UserID              int64     `xorm:"bigint not null"`
	Code                string    `xorm:"varchar(128) not null unique index"`
	ExpiresAt           time.Time `xorm:"not null"`
	ClientID            string    `xorm:"varchar(255) not null"`
	RedirectURI         string    `xorm:"text not null"`
	CodeChallenge       string    `xorm:"varchar(128) not null"`
	CodeChallengeMethod string    `xorm:"varchar(10) not null"`
	Created             time.Time `xorm:"created not null"`
}

func (oauthCodesBefore20261126080432) TableName() string { return "oauth_codes" }

func TestAddOAuthGrantColumns20261126080432(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []any{
		apiTokensBefore20261126080432{},
		oauthCodesBefore20261126080432{},
	}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	_, err = x.Insert(&apiTokensBefore20261126080432{
		ID:          1,
		Title:       "existing",
		TokenSha256: "abc",
		Permissions: `{"tasks":["read_all"]}`,
		ExpiresAt:   time.Now().Add(time.Hour),
		OwnerID:     1,
	})
	require.NoError(t, err)

	before := map[string]*schemas.Table{
		"api_tokens":  tableMeta20261126080432(t, x, "api_tokens"),
		"oauth_codes": tableMeta20261126080432(t, x, "oauth_codes"),
	}

	idx := slices.IndexFunc(migrations, func(m *xormigrate.Migration) bool { return m.ID == "20261126080432" })
	require.NotEqual(t, -1, idx)
	require.NoError(t, migrations[idx].Migrate(x))

	added := map[string][]string{
		"api_tokens": {
			"oauth_client_id",
			"oauth_resource",
		},
		"oauth_codes": {
			"scope",
			"resource",
		},
	}
	for name, table := range before {
		after := tableMeta20261126080432(t, x, name)
		for _, column := range added[name] {
			col := after.GetColumn(column)
			require.NotNilf(t, col, "%s.%s missing", name, column)
			require.Truef(t, col.Nullable, "%s.%s must be nullable", name, column)
		}
		for _, column := range table.ColumnsSeq() {
			require.NotNilf(t, after.GetColumn(column), "migration dropped column %s.%s", name, column)
		}
		for indexName, index := range table.Indexes {
			preserved, found := after.Indexes[indexName]
			require.Truef(t, found, "migration dropped index %s on %s", indexName, name)
			require.Equal(t, index.Type, preserved.Type)
			require.Equal(t, index.Cols, preserved.Cols)
		}
	}

	untouched, err := x.Table("api_tokens").Where(builder.IsNull{"oauth_client_id"}).Count()
	require.NoError(t, err)
	require.Equal(t, int64(1), untouched)
}

func tableMeta20261126080432(t *testing.T, x *xorm.Engine, name string) *schemas.Table {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == name {
			return table
		}
	}
	t.Fatalf("%s table not found", name)
	return nil
}
