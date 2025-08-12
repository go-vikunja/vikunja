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
	"encoding/json"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

type ProjectViewKind int

func (p *ProjectViewKind) MarshalJSON() ([]byte, error) {
	switch *p {
	case ProjectViewKindList:
		return []byte(`"list"`), nil
	case ProjectViewKindGantt:
		return []byte(`"gantt"`), nil
	case ProjectViewKindTable:
		return []byte(`"table"`), nil
	case ProjectViewKindKanban:
		return []byte(`"kanban"`), nil
	}

	return []byte(`null`), nil
}

func (p *ProjectViewKind) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "list":
		*p = ProjectViewKindList
	case "gantt":
		*p = ProjectViewKindGantt
	case "table":
		*p = ProjectViewKindTable
	case "kanban":
		*p = ProjectViewKindKanban
	default:
		return fmt.Errorf("unknown project view kind: %s", value)
	}

	return nil
}

// NOTE: When adding or changing enum values for ProjectViewKind,
// make sure to update the corresponding `enums` tag in the ProjectView struct
// to keep the OpenAPI documentation in sync.

const (
	ProjectViewKindList ProjectViewKind = iota
	ProjectViewKindGantt
	ProjectViewKindTable
	ProjectViewKindKanban
)

type BucketConfigurationModeKind int

// NOTE: When adding or changing enum values for BucketConfigurationModeKind,
// make sure to update the corresponding `enums` tag in the ProjectView struct
// to keep the OpenAPI documentation in sync.

const (
	BucketConfigurationModeNone BucketConfigurationModeKind = iota
	BucketConfigurationModeManual
	BucketConfigurationModeFilter
)

func (p *BucketConfigurationModeKind) MarshalJSON() ([]byte, error) {
	switch *p {
	case BucketConfigurationModeNone:
		return []byte(`"none"`), nil
	case BucketConfigurationModeManual:
		return []byte(`"manual"`), nil
	case BucketConfigurationModeFilter:
		return []byte(`"filter"`), nil
	}

	return []byte(`null`), nil
}

func (p *BucketConfigurationModeKind) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "none":
		*p = BucketConfigurationModeNone
	case "manual":
		*p = BucketConfigurationModeManual
	case "filter":
		*p = BucketConfigurationModeFilter
	default:
		return fmt.Errorf("unknown bucket configuration mode kind: %s", value)
	}

	return nil
}

type ProjectViewBucketConfiguration struct {
	Title  string          `json:"title"`
	Filter *TaskCollection `json:"filter"`
}

type ProjectView struct {
	// The unique numeric id of this view
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	// The title of this view
	Title string `xorm:"varchar(255) not null" json:"title" valid:"required,runelength(1|250)"`
	// The project this view belongs to
	ProjectID int64 `xorm:"not null index" json:"project_id" param:"project"`
	// The kind of this view. Can be `list`, `gantt`, `table` or `kanban`.
	ViewKind ProjectViewKind `xorm:"not null" json:"view_kind" swaggertype:"string" enums:"list,gantt,table,kanban"`

	// The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation.
	Filter *TaskCollection `xorm:"json null default null" query:"filter" json:"filter"`
	// The position of this view in the list. The list of all views will be sorted by this parameter.
	Position float64 `xorm:"double null" json:"position"`

	// The bucket configuration mode. Can be `none`, `manual` or `filter`. `manual` allows to move tasks between buckets as you normally would. `filter` creates buckets based on a filter for each bucket.
	BucketConfigurationMode BucketConfigurationModeKind `xorm:"default 0" json:"bucket_configuration_mode" swaggertype:"string" enums:"none,manual,filter,manual"`
	// When the bucket configuration mode is not `manual`, this field holds the options of that configuration.
	BucketConfiguration []*ProjectViewBucketConfiguration `xorm:"json" json:"bucket_configuration"`
	// The ID of the bucket where new tasks without a bucket are added to. By default, this is the leftmost bucket in a view.
	DefaultBucketID int64 `xorm:"bigint INDEX null" json:"default_bucket_id"`
	// If tasks are moved to the done bucket, they are marked as done. If they are marked as done individually, they are moved into the done bucket.
	DoneBucketID int64 `xorm:"bigint INDEX null" json:"done_bucket_id"`

	// A timestamp when this view was updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`
	// A timestamp when this reaction was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (pv *ProjectView) TableName() string {
	return "project_views"
}

func getViewsForProject(s *xorm.Session, projectID int64) (views []*ProjectView, err error) {
	views = []*ProjectView{}
	err = s.
		Where("project_id = ?", projectID).
		OrderBy("position asc").
		Find(&views)
	return
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
func (pv *ProjectView) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	pp := &Project{ID: pv.ProjectID}
	can, _, err := pp.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	projectViews, err := getViewsForProject(s, pv.ProjectID)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := s.
		Where("project_id = ?", pv.ProjectID).
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
func (pv *ProjectView) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	view, err := GetProjectViewByIDAndProject(s, pv.ID, pv.ProjectID)
	if err != nil {
		return err
	}

	*pv = *view
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
func (pv *ProjectView) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.
		Where("id = ? AND project_id = ?", pv.ID, pv.ProjectID).
		Delete(&ProjectView{})
	if err != nil {
		return
	}

	_, err = s.Where("project_view_id = ?", pv.ID).Delete(&TaskBucket{})
	if err != nil {
		return
	}

	_, err = s.Where("project_view_id = ?", pv.ID).Delete(&TaskPosition{})
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
func (pv *ProjectView) Create(s *xorm.Session, a web.Auth) (err error) {
	return createProjectView(s, pv, a, true, true)
}

func createProjectView(s *xorm.Session, p *ProjectView, a web.Auth, createBacklogBucket bool, addExistingTasksToView bool) (err error) {
	if p.Filter != nil && p.Filter.Filter != "" {
		_, err = getTaskFiltersFromFilterString(p.Filter.Filter, p.Filter.FilterTimezone)
		if err != nil {
			return
		}
	}

	if p.BucketConfigurationMode == BucketConfigurationModeFilter {
		for _, configuration := range p.BucketConfiguration {
			if configuration.Filter != nil && configuration.Filter.Filter != "" {
				_, err = getTaskFiltersFromFilterString(configuration.Filter.Filter, configuration.Filter.FilterTimezone)
				if err != nil {
					return
				}
			}
		}
	}

	p.ID = 0
	_, err = s.Insert(p)
	if err != nil {
		return
	}

	if p.ViewKind == ProjectViewKindKanban && createBacklogBucket && p.BucketConfigurationMode == BucketConfigurationModeManual {
		// Create default buckets for kanban view
		backlog := &Bucket{
			ProjectViewID: p.ID,
			Title:         "To-Do",
			Position:      100,
		}
		err = backlog.Create(s, a)
		if err != nil {
			return
		}

		doing := &Bucket{
			ProjectViewID: p.ID,
			Title:         "Doing",
			Position:      200,
		}
		err = doing.Create(s, a)
		if err != nil {
			return
		}

		done := &Bucket{
			ProjectViewID: p.ID,
			Title:         "Done",
			Position:      300,
		}
		err = done.Create(s, a)
		if err != nil {
			return
		}

		// Set Backlog as default bucket and Done as done bucket
		p.DefaultBucketID = backlog.ID
		p.DoneBucketID = done.ID
		_, err = s.ID(p.ID).Cols("default_bucket_id", "done_bucket_id").Update(p)
		if err != nil {
			return
		}

		// Move all tasks into the new bucket when the project already has tasks
		if addExistingTasksToView {
			err = addTasksToView(s, a, p, backlog)
			if err != nil {
				return
			}
		}
	}

	if addExistingTasksToView {
		return RecalculateTaskPositions(s, p, a)
	}

	return
}

func addTasksToView(s *xorm.Session, a web.Auth, pv *ProjectView, b *Bucket) (err error) {
	c := &TaskCollection{
		ProjectID: pv.ProjectID,
	}
	ts, _, _, err := c.ReadAll(s, a, "", 0, -1)
	if err != nil {
		return err
	}
	tasks := ts.([]*Task)

	if len(tasks) == 0 {
		return nil
	}

	taskBuckets := []*TaskBucket{}
	for _, task := range tasks {
		taskBuckets = append(taskBuckets, &TaskBucket{
			TaskID:        task.ID,
			BucketID:      b.ID,
			ProjectViewID: pv.ID,
		})
	}

	_, err = s.Insert(&taskBuckets)
	return err
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
func (pv *ProjectView) Update(s *xorm.Session, _ web.Auth) (err error) {
	if pv.Filter != nil && pv.Filter.Filter != "" {
		_, err = getTaskFiltersFromFilterString(pv.Filter.Filter, pv.Filter.FilterTimezone)
		if err != nil {
			return
		}
	}

	// Check if the project view exists
	_, err = GetProjectViewByIDAndProject(s, pv.ID, pv.ProjectID)
	if err != nil {
		return
	}

	_, err = s.
		ID(pv.ID).
		Cols(
			"title",
			"view_kind",
			"filter",
			"position",
			"bucket_configuration_mode",
			"bucket_configuration",
			"default_bucket_id",
			"done_bucket_id",
		).
		Update(pv)
	return
}

func GetProjectViewByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *ProjectView, err error) {
	if projectID == FavoritesPseudoProjectID && viewID < 0 {
		for _, v := range FavoritesPseudoProject.Views {
			if v.ID == viewID {
				return v, nil
			}
		}

		return nil, &ErrProjectViewDoesNotExist{
			ProjectViewID: viewID,
		}
	}

	view = &ProjectView{}
	exists, err := s.
		Where("id = ? AND project_id = ?", viewID, projectID).
		NoAutoCondition().
		Get(view)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &ErrProjectViewDoesNotExist{
			ProjectViewID: viewID,
		}
	}

	return
}

func GetProjectViewByID(s *xorm.Session, id int64) (view *ProjectView, err error) {
	view = &ProjectView{}
	exists, err := s.
		Where("id = ?", id).
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

func CreateDefaultViewsForProject(s *xorm.Session, project *Project, a web.Auth, createBacklogBucket bool, createDefaultListFilter bool) (err error) {
	list := &ProjectView{
		ProjectID: project.ID,
		Title:     "List",
		ViewKind:  ProjectViewKindList,
		Position:  100,
	}
	if createDefaultListFilter {
		list.Filter = &TaskCollection{
			Filter: "done = false",
		}
	}
	err = createProjectView(s, list, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	gantt := &ProjectView{
		ProjectID: project.ID,
		Title:     "Gantt",
		ViewKind:  ProjectViewKindGantt,
		Position:  200,
	}
	err = createProjectView(s, gantt, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	table := &ProjectView{
		ProjectID: project.ID,
		Title:     "Table",
		ViewKind:  ProjectViewKindTable,
		Position:  300,
	}
	err = createProjectView(s, table, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	kanban := &ProjectView{
		ProjectID:               project.ID,
		Title:                   "Kanban",
		ViewKind:                ProjectViewKindKanban,
		Position:                400,
		BucketConfigurationMode: BucketConfigurationModeManual,
	}
	err = createProjectView(s, kanban, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	project.Views = []*ProjectView{
		list,
		gantt,
		table,
		kanban,
	}

	return
}
