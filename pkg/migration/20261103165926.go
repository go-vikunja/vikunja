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
	"strings"

	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type taskRelations20261103165926 struct {
	OtherTaskID  int64  `xorm:"bigint not null index(other_task_kind)"`
	RelationKind string `xorm:"varchar(50) not null index(other_task_kind)"`
}

func (taskRelations20261103165926) TableName() string {
	return "task_relations"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261103165926",
		Description: "Add index on task_relations (other_task_id, relation_kind) for the relation cycle check",
		Migrate:     addTaskRelationsOtherTaskKindIndex20261103165926,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

func addTaskRelationsOtherTaskKindIndex20261103165926(tx *xorm.Engine) error {
	if tx.Dialect().URI().DBType == schemas.MYSQL {
		rows, err := tx.QueryString(`SELECT DATA_TYPE AS data_type FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'task_relations' AND COLUMN_NAME = 'relation_kind'`)
		if err != nil {
			return fmt.Errorf("could not read the type of task_relations.relation_kind: %w", err)
		}
		// DBs converted from sqlite carry TEXT here, which mysql can't index
		// without a key length (#4374). The index is only a speedup.
		if len(rows) > 0 && mysqlUnindexableTypes20260720120000[strings.ToUpper(rows[0]["data_type"])] {
			log.Warningf("Could not create the index on task_relations (other_task_id, relation_kind): mysql cannot index the text column relation_kind without a key length. Change it to VARCHAR(50), then run this manually: CREATE INDEX IDX_task_relations_other_task_kind ON task_relations (other_task_id, relation_kind)")
			return nil
		}
	}
	return partialSync(tx, taskRelations20261103165926{})
}
