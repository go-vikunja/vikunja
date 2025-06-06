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
	"time"

	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type projects20221228112131 struct {
	// This is the one new property
	ParentProjectID int64 `xorm:"bigint INDEX null" json:"parent_project_id"`

	// Those only exist to make the migration independent of future changes
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	Title       string    `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	Description string    `xorm:"longtext null" json:"description"`
	HexColor    string    `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`
	OwnerID     int64     `xorm:"bigint INDEX not null" json:"-"`
	IsArchived  bool      `xorm:"not null default false" json:"is_archived" query:"is_archived"`
	Created     time.Time `xorm:"created not null" json:"created"`
	Updated     time.Time `xorm:"updated not null" json:"updated"`
	NamespaceID int64     `xorm:"bigint INDEX not null" json:"namespace_id" param:"namespace"`
}

func (projects20221228112131) TableName() string {
	return "projects"
}

type namespace20221228112131 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id" param:"namespace"`
	Title       string    `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	Description string    `xorm:"longtext null" json:"description"`
	OwnerID     int64     `xorm:"bigint not null INDEX" json:"-"`
	HexColor    string    `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`
	IsArchived  bool      `xorm:"not null default false" json:"is_archived" query:"is_archived"`
	Created     time.Time `xorm:"created not null" json:"created"`
	Updated     time.Time `xorm:"updated not null" json:"updated"`
}

func (namespace20221228112131) TableName() string {
	return "namespaces"
}

type teamNamespace20221228112131 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TeamID      int64     `xorm:"bigint not null INDEX" json:"team_id" param:"team"`
	NamespaceID int64     `xorm:"bigint not null INDEX" json:"-" param:"namespace"`
	Right       int       `xorm:"bigint INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`
	Created     time.Time `xorm:"created not null" json:"created"`
	Updated     time.Time `xorm:"updated not null" json:"updated"`
}

func (teamNamespace20221228112131) TableName() string {
	return "team_namespaces"
}

type teamProject20221228112131 struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TeamID    int64     `xorm:"bigint not null INDEX" json:"team_id" param:"team"`
	ProjectID int64     `xorm:"bigint not null INDEX" json:"-" param:"project"`
	Right     int       `xorm:"bigint INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`
	Created   time.Time `xorm:"created not null" json:"created"`
	Updated   time.Time `xorm:"updated not null" json:"updated"`
}

func (teamProject20221228112131) TableName() string {
	return "team_projects"
}

type namespaceUser20221228112131 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id" param:"namespace"`
	UserID      int64     `xorm:"bigint not null INDEX" json:"-"`
	NamespaceID int64     `xorm:"bigint not null INDEX" json:"-" param:"namespace"`
	Right       int       `xorm:"bigint INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`
	Created     time.Time `xorm:"created not null" json:"created"`
	Updated     time.Time `xorm:"updated not null" json:"updated"`
}

func (namespaceUser20221228112131) TableName() string {
	return "users_namespaces"
}

type projectUser20221228112131 struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk" json:"id" param:"namespace"`
	UserID    int64     `xorm:"bigint not null INDEX" json:"-"`
	ProjectID int64     `xorm:"bigint not null INDEX" json:"-" param:"project"`
	Right     int       `xorm:"bigint INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`
	Created   time.Time `xorm:"created not null" json:"created"`
	Updated   time.Time `xorm:"updated not null" json:"updated"`
}

func (projectUser20221228112131) TableName() string {
	return "users_projects"
}

const sqliteRemoveNamespaceColumn20221228112131 = `
create table projects_dg_tmp

(
    id                   INTEGER           not null
        primary key autoincrement,
    title                TEXT              not null,
    description          TEXT,
    identifier           TEXT,
    hex_color            TEXT,
    owner_id             INTEGER           not null,
    is_archived          INTEGER default 0 not null,
    background_file_id   INTEGER,
    background_blur_hash TEXT,
    position             REAL,
    created              DATETIME          not null,
    updated              DATETIME          not null,
    parent_project_id    INTEGER
);

insert into projects_dg_tmp(id, title, description, identifier, hex_color, owner_id, is_archived, background_file_id,
                            background_blur_hash, position, created, updated, parent_project_id)
select id,
       title,
       description,
       identifier,
       hex_color,
       owner_id,
       is_archived,
       background_file_id,
       background_blur_hash,
       position,
       created,
       updated,
       parent_project_id
from projects;

drop table projects;

alter table projects_dg_tmp
    rename to projects;

create index IDX_lists_owner_id
    on projects (owner_id);

create index IDX_projects_parent_project_id
    on projects (parent_project_id);

create unique index UQE_lists_id
    on projects (id);
`

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20221228112131",
		Description: "make projects nestable",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(projects20221228112131{})
			if err != nil {
				return err
			}

			allNamespaces := []*namespace20221228112131{}
			err = tx.Find(&allNamespaces)
			if err != nil {
				return err
			}

			// namespace id is the key
			namespacesToProjects := make(map[int64]*projects20221228112131)

			for _, n := range allNamespaces {
				p := &projects20221228112131{
					Title:       n.Title,
					Description: n.Description,
					OwnerID:     n.OwnerID,
					HexColor:    n.HexColor,
					IsArchived:  n.IsArchived,
					Created:     n.Created,
					Updated:     n.Updated,
				}

				_, err = tx.Insert(p)
				if err != nil {
					return err
				}
				namespacesToProjects[n.ID] = p
			}

			err = setParentProject(tx, namespacesToProjects)
			if err != nil {
				return err
			}

			err = setTeamNamespacesShare(tx, namespacesToProjects)
			if err != nil {
				return err
			}

			err = setUserNamespacesShare(tx, namespacesToProjects)
			if err != nil {
				return err
			}

			return removeNamespaceLeftovers(tx)
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

func setParentProject(tx *xorm.Engine, namespacesToProjects map[int64]*projects20221228112131) error {
	for namespaceID, project := range namespacesToProjects {
		_, err := tx.Where("namespace_id = ?", namespaceID).
			Update(&projects20221228112131{
				ParentProjectID: project.ID,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func setTeamNamespacesShare(tx *xorm.Engine, namespacesToProjects map[int64]*projects20221228112131) error {
	teamNamespaces := []*teamNamespace20221228112131{}
	err := tx.Find(&teamNamespaces)
	if err != nil {
		return err
	}

	for _, tn := range teamNamespaces {
		if _, exists := namespacesToProjects[tn.NamespaceID]; !exists {
			log.Warningf("Namespace %d does not exist but is shared with team %d - this is probably caused by an old share which was not properly deleted.", tn.NamespaceID, tn.TeamID)
			continue
		}

		_, err = tx.Insert(&teamProject20221228112131{
			TeamID:    tn.TeamID,
			Right:     tn.Right,
			Created:   tn.Created,
			Updated:   tn.Updated,
			ProjectID: namespacesToProjects[tn.NamespaceID].ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func setUserNamespacesShare(tx *xorm.Engine, namespacesToProjects map[int64]*projects20221228112131) error {
	userNamespace := []*namespaceUser20221228112131{}
	err := tx.Find(&userNamespace)
	if err != nil {
		return err
	}

	for _, un := range userNamespace {
		if _, exists := namespacesToProjects[un.NamespaceID]; !exists {
			log.Warningf("Namespace %d does not exist but is shared with user %d - this is probably caused by an old share which was not properly deleted.", un.NamespaceID, un.UserID)
			continue
		}

		_, err = tx.Insert(&projectUser20221228112131{
			UserID:    un.UserID,
			Right:     un.Right,
			Created:   un.Created,
			Updated:   un.Updated,
			ProjectID: namespacesToProjects[un.NamespaceID].ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func removeNamespaceLeftovers(tx *xorm.Engine) error {
	err := tx.DropTables("namespaces", "team_namespaces", "users_namespaces")
	if err != nil {
		return err
	}

	if tx.Dialect().URI().DBType == schemas.SQLITE {
		_, err := tx.Exec(sqliteRemoveNamespaceColumn20221228112131)
		return err
	}

	return dropTableColum(tx, "projects", "namespace_id")
}
