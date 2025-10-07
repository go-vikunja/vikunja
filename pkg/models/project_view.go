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

// ProjectViewServiceProvider is the interface that must be implemented by a service providing project view functionality.
// This enables dependency injection and allows the model layer to delegate to the service layer.
type ProjectViewServiceProvider interface {
	Create(s *xorm.Session, pv *ProjectView, a web.Auth, createBacklogBucket bool, addExistingTasksToView bool) error
	Update(s *xorm.Session, pv *ProjectView) error
	Delete(s *xorm.Session, viewID int64, projectID int64) error
	GetAll(s *xorm.Session, projectID int64, a web.Auth) (views []*ProjectView, totalCount int64, err error)
	GetByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *ProjectView, err error)
	GetByID(s *xorm.Session, id int64) (view *ProjectView, err error)
	CreateDefaultViewsForProject(s *xorm.Session, project *Project, a web.Auth, createBacklogBucket bool, createDefaultListFilter bool) error
}

var projectViewService ProjectViewServiceProvider

// RegisterProjectViewService registers the service implementation for project views.
// This should be called during application initialization.
func RegisterProjectViewService(service ProjectViewServiceProvider) {
	projectViewService = service
}

// getProjectViewService retrieves the registered project view service.
// Panics if no service has been registered (indicates missing initialization).
func getProjectViewService() ProjectViewServiceProvider {
	if projectViewService == nil {
		panic("ProjectViewService not registered. Call RegisterProjectViewService during initialization.")
	}
	return projectViewService
}

// getViewsForProject retrieves all views for a project.
// This is a simple database query helper used by other models (Project, Task).
// Note: This is NOT business logic - it's a pure database query with no validation or processing.
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
// @Deprecated Use ProjectViewService.GetAll instead. This method only exists for backward compatibility.
func (pv *ProjectView) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	service := getProjectViewService()
	views, totalCount, err := service.GetAll(s, pv.ProjectID, a)
	if err != nil {
		return nil, 0, 0, err
	}
	return views, len(views), totalCount, nil
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
// @Deprecated Use ProjectViewService.GetByIDAndProject instead. This method only exists for backward compatibility.
func (pv *ProjectView) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	service := getProjectViewService()
	view, err := service.GetByIDAndProject(s, pv.ID, pv.ProjectID)
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
// @Deprecated Use ProjectViewService.Delete instead. This method only exists for backward compatibility.
func (pv *ProjectView) Delete(s *xorm.Session, _ web.Auth) (err error) {
	service := getProjectViewService()
	return service.Delete(s, pv.ID, pv.ProjectID)
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
// @Deprecated Use ProjectViewService.Create instead. This method only exists for backward compatibility.
func (pv *ProjectView) Create(s *xorm.Session, a web.Auth) (err error) {
	service := getProjectViewService()
	return service.Create(s, pv, a, true, true)
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
// @Deprecated Use ProjectViewService.Update instead. This method only exists for backward compatibility.
func (pv *ProjectView) Update(s *xorm.Session, _ web.Auth) (err error) {
	service := getProjectViewService()
	return service.Update(s, pv)
}

// GetProjectViewByIDAndProject retrieves a project view by its ID and project ID.
// @Deprecated Use ProjectViewService.GetByIDAndProject instead. This function only exists for backward compatibility.
func GetProjectViewByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *ProjectView, err error) {
	service := getProjectViewService()
	return service.GetByIDAndProject(s, viewID, projectID)
}

// GetProjectViewByID retrieves a project view by its ID.
// @Deprecated Use ProjectViewService.GetByID instead. This function only exists for backward compatibility.
func GetProjectViewByID(s *xorm.Session, id int64) (view *ProjectView, err error) {
	service := getProjectViewService()
	return service.GetByID(s, id)
}

// CreateDefaultViewsForProject creates the default views for a project.
// @Deprecated Use ProjectViewService.CreateDefaultViewsForProject instead. This function only exists for backward compatibility.
func CreateDefaultViewsForProject(s *xorm.Session, project *Project, a web.Auth, createBacklogBucket bool, createDefaultListFilter bool) (err error) {
	service := getProjectViewService()
	return service.CreateDefaultViewsForProject(s, project, a, createBacklogBucket, createDefaultListFilter)
}

// createProjectView is a deprecated helper function.
// @Deprecated Use ProjectViewService.Create instead. This function only exists for backward compatibility.
func createProjectView(s *xorm.Session, p *ProjectView, a web.Auth, createBacklogBucket bool, addExistingTasksToView bool) (err error) {
	service := getProjectViewService()
	return service.Create(s, p, a, createBacklogBucket, addExistingTasksToView)
}
