// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"time"
	"xorm.io/xorm"
)

type ProjectViewKind int

const (
	ProjectViewKindList ProjectViewKind = iota
	ProjectViewKindGantt
	ProjectViewKindTable
	ProjectViewKindKanban
)

type BucketConfigurationModeKind int

const (
	BucketConfigurationModeNone BucketConfigurationModeKind = iota
	BucketConfigurationModeManual
	BucketConfigurationModeFilter
)

type ProjectViewBucketConfiguration struct {
	Title  string
	Filter string
}

type ProjectView struct {
	// The unique numeric id of this view
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	// The title of this view
	Title string `xorm:"varchar(255) not null" json:"title" valid:"runelength(1|250)"`
	// The project this view belongs to
	ProjectID int64 `xorm:"not null index" json:"project_id" param:"project"`
	// The kind of this view. Can be `list`, `gantt`, `table` or `kanban`.
	ViewKind ProjectViewKind `xorm:"not null" json:"view_kind"`

	// The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation.
	Filter string `xorm:"text null default null" query:"filter" json:"filter"`
	// The position of this view in the list. The list of all views will be sorted by this parameter.
	Position float64 `xorm:"double null" json:"position"`

	// The bucket configuration mode. Can be `none`, `manual` or `filter`. `manual` allows to move tasks between buckets as you normally would. `filter` creates buckets based on a filter for each bucket.
	BucketConfigurationMode BucketConfigurationModeKind `xorm:"default 0" json:"bucket_configuration_mode"`
	// When the bucket configuration mode is not `manual`, this field holds the options of that configuration.
	BucketConfiguration []*ProjectViewBucketConfiguration `xorm:"json" json:"bucket_configuration"`

	// A timestamp when this view was updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`
	// A timestamp when this reaction was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

func (p *ProjectView) TableName() string {
	return "project_views"
}

// ReadAll gets all project views
// @Summary Get all project views for a project
// @Description Returns all project views for a sepcific project
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Success 200 {array} models.ProjectView "The project views"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views [get]
func (p *ProjectView) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	pp := &Project{ID: p.ProjectID}
	can, _, err := pp.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	projectViews := []*ProjectView{}
	err = s.
		Where("project_id = ?", p.ProjectID).
		Find(&projectViews)
	if err != nil {
		return
	}

	totalCount, err := s.
		Where("project_id = ?", p.ProjectID).
		Count(&ProjectView{})
	if err != nil {
		return
	}

	return projectViews, len(projectViews), totalCount, nil
}

// ReadOne implements the CRUD method to get one project view
// @Summary Get one project view
// @Description Returns a project view by its ID.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param id path int true "Project View ID"
// @Success 200 {object} models.ProjectView "The project view"
// @Failure 403 {object} web.HTTPError "The user does not have access to this project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{id} [get]
func (p *ProjectView) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	view, err := GetProjectViewByID(s, p.ID, p.ProjectID)
	if err != nil {
		return err
	}

	*p = *view
	return
}

// Delete removes the project view
// @Summary Delete a project view
// @Description Deletes a project view.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param id path int true "Project View ID"
// @Success 200 {object} models.Message "The project view was successfully deleted."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{id} [delete]
func (p *ProjectView) Delete(s *xorm.Session, a web.Auth) (err error) {
	_, err = s.
		Where("id = ? AND projec_id = ?", p.ID, p.ProjectID).
		Delete(&ProjectView{})
	return
}

// Create adds a new project view
// @Summary Create a project view
// @Description Create a project view in a specific project.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view body models.ProjectView true "The project view you want to create."
// @Success 200 {object} models.ProjectView "The created project view"
// @Failure 403 {object} web.HTTPError "The user does not have access to create a project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views [put]
func (p *ProjectView) Create(s *xorm.Session, a web.Auth) (err error) {
	_, err = s.Insert(p)
	return
}

// Update is the handler to update a project view
// @Summary Updates a project view
// @Description Updates a project view.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param id path int true "Project View ID"
// @Param view body models.ProjectView true "The project view with updated values you want to change."
// @Success 200 {object} models.ProjectView "The updated project view."
// @Failure 400 {object} web.HTTPError "Invalid project view object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{id} [post]
func (p *ProjectView) Update(s *xorm.Session, _ web.Auth) (err error) {
	// Check if the project view exists
	_, err = GetProjectViewByID(s, p.ID, p.ProjectID)
	if err != nil {
		return
	}

	_, err = s.ID(p.ID).Update(p)
	if err != nil {
		return
	}

	return
}

func GetProjectViewByID(s *xorm.Session, id, projectID int64) (view *ProjectView, err error) {
	exists, err := s.
		Where("id = ? AND project_id = ?", id, projectID).
		NoAutoCondition().
		Get(view)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &ErrProjectViewDoesNotExist{
			ProjectViewID: id,
		}
	}

	return
}

func CreateDefaultViewsForProject(s *xorm.Session, project *Project, a web.Auth) (err error) {
	list := &ProjectView{
		ProjectID: project.ID,
		Title:     "List",
		ViewKind:  ProjectViewKindList,
		Position:  100,
	}
	err = list.Create(s, a)
	if err != nil {
		return
	}

	gantt := &ProjectView{
		ProjectID: project.ID,
		Title:     "Gantt",
		ViewKind:  ProjectViewKindGantt,
		Position:  200,
	}
	err = gantt.Create(s, a)
	if err != nil {
		return
	}

	table := &ProjectView{
		ProjectID: project.ID,
		Title:     "Table",
		ViewKind:  ProjectViewKindTable,
		Position:  300,
	}
	err = table.Create(s, a)
	if err != nil {
		return
	}

	kanban := &ProjectView{
		ProjectID: project.ID,
		Title:     "Kanban",
		ViewKind:  ProjectViewKindKanban,
		Position:  400,
	}
	err = kanban.Create(s, a)
	return
}
