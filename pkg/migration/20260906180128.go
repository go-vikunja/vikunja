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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// Nullable: old rows get token_sha256 backfilled lazily, new rows never write the legacy columns.
type apiTokenSha25620260906180128 struct {
	ID             int64               `xorm:"bigint autoincr not null unique pk"`
	Title          string              `xorm:"not null"`
	TokenSalt      string              `xorm:"null"`
	TokenHash      string              `xorm:"null unique"`
	TokenLastEight string              `xorm:"null index varchar(8)"`
	TokenSha256    string              `xorm:"varchar(64) null unique"`
	Permissions    map[string][]string `xorm:"json not null"`
	ExpiresAt      time.Time           `xorm:"not null"`
	Created        time.Time           `xorm:"created not null"`
	OwnerID        int64               `xorm:"bigint not null"`
}

func (apiTokenSha25620260906180128) TableName() string {
	return "api_tokens"
}

func addAPITokenSha25620260906180128(tx *xorm.Engine) error {
	if db.Type() == schemas.SQLITE {
		return rebuildAPITokensSQLite20260906180128(tx)
	}

	if err := partialSync(tx, apiTokenSha25620260906180128{}); err != nil {
		return err
	}

	// partialSync skips unique constraints. Built through the dialect so the name is quoted like xorm's own Sync
	// would (Postgres folds unquoted names to lowercase), which keeps a later sync from creating a duplicate.
	index := &schemas.Index{Name: "token_sha256", Type: schemas.UniqueType, Cols: []string{"token_sha256"}}
	checkSQL, args := tx.Dialect().IndexCheckSQL("api_tokens", index.XName("api_tokens"))
	existing, err := tx.Query(append([]any{checkSQL}, args...)...)
	if err != nil {
		return fmt.Errorf("could not check for unique index on api_tokens.token_sha256: %w", err)
	}
	if len(existing) == 0 {
		if _, err := tx.Exec(tx.Dialect().CreateIndexSQL("api_tokens", index)); err != nil {
			return fmt.Errorf("could not create unique index on api_tokens.token_sha256: %w", err)
		}
	}

	var queries []string
	if db.Type() == schemas.MYSQL {
		queries = []string{
			"ALTER TABLE api_tokens MODIFY COLUMN token_salt varchar(255) NULL",
			"ALTER TABLE api_tokens MODIFY COLUMN token_hash varchar(255) NULL",
			"ALTER TABLE api_tokens MODIFY COLUMN token_last_eight varchar(8) NULL",
		}
	} else {
		queries = []string{
			"ALTER TABLE api_tokens ALTER COLUMN token_salt DROP NOT NULL, ALTER COLUMN token_hash DROP NOT NULL, ALTER COLUMN token_last_eight DROP NOT NULL",
		}
	}
	for _, q := range queries {
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("could not drop not null on api_tokens legacy columns: %w", err)
		}
	}

	return nil
}

// SQLite cannot drop NOT NULL, so the table is rebuilt; its DDL is transactional, so a crash leaves the old table intact.
func rebuildAPITokensSQLite20260906180128(tx *xorm.Engine) error {
	// A re-run would copy rows without token_sha256, nulling every value written since the upgrade.
	exists, err := columnExists(tx, "api_tokens", "token_sha256")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	sess := tx.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("could not start transaction to rebuild api_tokens: %w", err)
	}

	// Indexes keep their names when the table is renamed and would clash with the ones Sync creates.
	for _, idx := range []string{
		"UQE_api_tokens_id",
		"UQE_api_tokens_token_hash",
		"IDX_api_tokens_token_last_eight",
		"UQE_api_tokens_token_sha256",
	} {
		if _, err := sess.Exec("DROP INDEX IF EXISTS " + idx); err != nil {
			return fmt.Errorf("could not drop index %s: %w", idx, err)
		}
	}

	if _, err := sess.Exec("ALTER TABLE api_tokens RENAME TO api_tokens_old"); err != nil {
		return fmt.Errorf("could not rename api_tokens: %w", err)
	}

	// Table was just renamed away, so Sync only creates; nothing for it to drop.
	if err := sess.Sync(apiTokenSha25620260906180128{}); err != nil {
		return fmt.Errorf("could not create new api_tokens table: %w", err)
	}

	cols := "id, title, token_salt, token_hash, token_last_eight, permissions, expires_at, created, owner_id"
	if _, err := sess.Exec("INSERT INTO api_tokens (" + cols + ") SELECT " + cols + " FROM api_tokens_old"); err != nil {
		return fmt.Errorf("could not copy api_tokens rows: %w", err)
	}

	if err := sess.DropTable("api_tokens_old"); err != nil {
		return fmt.Errorf("could not drop api_tokens_old: %w", err)
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("could not commit api_tokens rebuild: %w", err)
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260906180128",
		Description: "Add nullable token_sha256 column with unique index to api_tokens and make legacy hash columns nullable",
		Migrate:     addAPITokenSha25620260906180128,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
