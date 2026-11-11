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
	"slices"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type apiTokens20261020024804 struct {
	ID          int64               `xorm:"bigint autoincr not null unique pk"`
	Permissions map[string][]string `xorm:"json not null"`
}

func (apiTokens20261020024804) TableName() string {
	return "api_tokens"
}

// Old keys map to their v1 meaning, which is what the token UI offered (GHSA-j5v7-qcwp-5g32).
var renamedAPITokenPermissions20261020024804 = map[string]map[string][]string{
	"admin": {
		"users_post":        {"users_create"},
		"invite_links_post": {"invite_links_create"},
	},
	"notifications": {
		// v1 filed marking all as read under update, v2 marking one.
		"update": {"update", "mark_all_as_read"},
		"create": {"mark_all_as_read"},
	},
	"projects": {
		"background_get":          {"background"},
		"views_buckets_put":       {"views_buckets_create"},
		"views_buckets_post":      {"views_buckets_update"},
		"views_buckets_tasks":     {"views_buckets_tasks_update"},
		"views_buckets_tasks_get": {"views_buckets_tasks"},
	},
}

func renameAPITokenPermissions20261020024804(permissions map[string][]string) (renamed map[string][]string, changed bool) {
	renamed = make(map[string][]string, len(permissions))
	for group, perms := range permissions {
		renames := renamedAPITokenPermissions20261020024804[group]
		next := make([]string, 0, len(perms))
		for _, perm := range perms {
			replacements, has := renames[perm]
			if !has {
				replacements = []string{perm}
			} else {
				changed = true
			}
			for _, r := range replacements {
				if !slices.Contains(next, r) {
					next = append(next, r)
				}
			}
		}
		renamed[group] = next
	}
	return renamed, changed
}

func migrateAPITokenPermissions20261020024804(tx *xorm.Engine) error {
	sess := tx.NewSession()
	defer sess.Close()

	// A partial run can't be retried safely: views_buckets_tasks_get would be renamed twice.
	if err := sess.Begin(); err != nil {
		return fmt.Errorf("could not start transaction to rename api token permissions: %w", err)
	}

	tokens := []*apiTokens20261020024804{}
	if err := sess.Find(&tokens); err != nil {
		return fmt.Errorf("could not load api tokens: %w", err)
	}
	for _, token := range tokens {
		renamed, changed := renameAPITokenPermissions20261020024804(token.Permissions)
		if !changed {
			continue
		}
		token.Permissions = renamed
		if _, err := sess.ID(token.ID).Cols("permissions").Update(token); err != nil {
			return fmt.Errorf("could not update permissions of api token %d: %w", token.ID, err)
		}
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("could not commit api token permission rename: %w", err)
	}
	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261020024804",
		Description: "Rename API token permission keys so each grants the same action on v1 and v2",
		Migrate:     migrateAPITokenPermissions20261020024804,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
