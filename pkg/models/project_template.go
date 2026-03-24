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

package models

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// ProjectTemplate represents the action of promoting a project to a template
type ProjectTemplate struct {
	// The project id of the project to save as template
	ProjectID int64 `json:"-" param:"projectid"`

	// The resulting template project
	Project *Project `json:"project,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if a user has the permission to create a template from a project
func (pt *ProjectTemplate) CanCreate(s *xorm.Session, a web.Auth) (canCreate bool, err error) {
	p := &Project{ID: pt.ProjectID}
	canCreate, _, err = p.CanRead(s, a)
	return canCreate, err
}

// Create duplicates a project and marks the copy as a template
// @Summary Save a project as a template
// @Description Creates a template by duplicating the project structure (tasks, views, buckets, backgrounds) without permissions, shares, assignees, or comments.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projectid path int true "The project ID to save as template"
// @Success 201 {object} models.ProjectTemplate "The created template"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectid}/template [put]
func (pt *ProjectTemplate) Create(s *xorm.Session, doer web.Auth) (err error) {
	log.Debugf("Creating template from project %d", pt.ProjectID)

	pd := &ProjectDuplicate{
		ProjectID:  pt.ProjectID,
		IsTemplate: true,
	}

	// Read the source project so the duplicate has something to work with
	pd.Project = &Project{ID: pt.ProjectID}
	err = pd.Project.ReadOne(s, doer)
	if err != nil {
		return err
	}

	err = pd.Create(s, doer)
	if err != nil {
		return err
	}

	pt.Project = pd.Project
	return
}
