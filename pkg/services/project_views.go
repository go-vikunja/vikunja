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

package services

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// InitProjectViewService sets up dependency injection for project view-related model functions.
// This function must be called during initialization to enable service layer delegation.
func InitProjectViewService() {
	// Set up permission delegation (T-PERM-011)
	models.ProjectViewCanReadFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error) {
		pvs := NewProjectViewService(s.Engine())
		return pvs.CanRead(s, projectID, a)
	}
	models.ProjectViewCanCreateFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pvs := NewProjectViewService(s.Engine())
		return pvs.CanCreate(s, projectID, a)
	}
	models.ProjectViewCanUpdateFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pvs := NewProjectViewService(s.Engine())
		return pvs.CanUpdate(s, projectID, a)
	}
	models.ProjectViewCanDeleteFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pvs := NewProjectViewService(s.Engine())
		return pvs.CanDelete(s, projectID, a)
	}

	// Set up helper function delegation (T-PERM-014A Phase 2)
	models.GetProjectViewByIDFunc = func(s *xorm.Session, id int64) (*models.ProjectView, error) {
		pvs := NewProjectViewService(s.Engine())
		return pvs.GetByID(s, id)
	}
	models.GetProjectViewByIDAndProjectFunc = func(s *xorm.Session, viewID int64, projectID int64) (*models.ProjectView, error) {
		pvs := NewProjectViewService(s.Engine())
		return pvs.GetByIDAndProject(s, viewID, projectID)
	}
}

// ProjectViewService represents a service for managing project views.
type ProjectViewService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewProjectViewService creates a new ProjectViewService.
// Deprecated: Use ServiceRegistry.ProjectViews() instead.
func NewProjectViewService(db *xorm.Engine) *ProjectViewService {
	registry := NewServiceRegistry(db)
	return registry.ProjectViews()
}

// Create adds a new project view.
func (pvs *ProjectViewService) Create(s *xorm.Session, pv *models.ProjectView, a web.Auth, createBacklogBucket bool, addExistingTasksToView bool) error {
	if pv.Filter != nil && pv.Filter.Filter != "" {
		_, err := models.GetTaskFiltersFromFilterString(pv.Filter.Filter, pv.Filter.FilterTimezone)
		if err != nil {
			return err
		}
	}

	if pv.BucketConfigurationMode == models.BucketConfigurationModeFilter {
		for _, configuration := range pv.BucketConfiguration {
			if configuration.Filter != nil && configuration.Filter.Filter != "" {
				_, err := models.GetTaskFiltersFromFilterString(configuration.Filter.Filter, configuration.Filter.FilterTimezone)
				if err != nil {
					return err
				}
			}
		}
	}

	pv.ID = 0
	_, err := s.Insert(pv)
	if err != nil {
		return err
	}

	if pv.ViewKind == models.ProjectViewKindKanban && createBacklogBucket && pv.BucketConfigurationMode == models.BucketConfigurationModeManual {
		// Create default buckets for kanban view
		backlog := &models.Bucket{
			ProjectViewID: pv.ID,
			ProjectID:     pv.ProjectID,
			Title:         "To-Do",
			Position:      100,
		}
		// Use direct database insert to avoid service layer validation during view creation
		backlog.CreatedBy, err = models.GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}
		backlog.CreatedByID = backlog.CreatedBy.ID
		_, err = s.Insert(backlog)
		if err != nil {
			return err
		}
		backlog.Position = models.CalculateDefaultPosition(backlog.ID, backlog.Position)
		_, err = s.Where("id = ?", backlog.ID).Update(backlog)
		if err != nil {
			return err
		}

		doing := &models.Bucket{
			ProjectViewID: pv.ID,
			ProjectID:     pv.ProjectID,
			Title:         "Doing",
			Position:      200,
		}
		// Use direct database insert to avoid service layer validation during view creation
		doing.CreatedBy, err = models.GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}
		doing.CreatedByID = doing.CreatedBy.ID
		_, err = s.Insert(doing)
		if err != nil {
			return err
		}
		doing.Position = models.CalculateDefaultPosition(doing.ID, doing.Position)
		_, err = s.Where("id = ?", doing.ID).Update(doing)
		if err != nil {
			return err
		}

		done := &models.Bucket{
			ProjectViewID: pv.ID,
			ProjectID:     pv.ProjectID,
			Title:         "Done",
			Position:      300,
		}
		// Use direct database insert to avoid service layer validation during view creation
		done.CreatedBy, err = models.GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}
		done.CreatedByID = done.CreatedBy.ID
		_, err = s.Insert(done)
		if err != nil {
			return err
		}
		done.Position = models.CalculateDefaultPosition(done.ID, done.Position)
		_, err = s.Where("id = ?", done.ID).Update(done)
		if err != nil {
			return err
		}

		// Set Backlog as default bucket and Done as done bucket
		pv.DefaultBucketID = backlog.ID
		pv.DoneBucketID = done.ID
		_, err = s.ID(pv.ID).Cols("default_bucket_id", "done_bucket_id").Update(pv)
		if err != nil {
			return err
		}

		// Move all tasks into the new bucket when the project already has tasks
		if addExistingTasksToView {
			err = pvs.addTasksToView(s, a, pv, backlog)
			if err != nil {
				return err
			}
		}
	}

	if addExistingTasksToView {
		return models.RecalculateTaskPositions(s, pv, a)
	}

	return nil
}

// addTasksToView is a helper that adds all existing tasks in a project to a view's bucket.
func (pvs *ProjectViewService) addTasksToView(s *xorm.Session, a web.Auth, pv *models.ProjectView, b *models.Bucket) (err error) {
	c := &models.TaskCollection{
		ProjectID: pv.ProjectID,
	}
	ts, _, _, err := c.ReadAll(s, a, "", 0, -1)
	if err != nil {
		return err
	}
	tasks := ts.([]*models.Task)

	if len(tasks) == 0 {
		return nil
	}

	taskBuckets := []*models.TaskBucket{}
	for _, task := range tasks {
		taskBuckets = append(taskBuckets, &models.TaskBucket{
			TaskID:        task.ID,
			BucketID:      b.ID,
			ProjectViewID: pv.ID,
		})
	}

	_, err = s.Insert(&taskBuckets)
	return err
}

// Update updates a project view.
func (pvs *ProjectViewService) Update(s *xorm.Session, pv *models.ProjectView) error {
	if pv.Filter != nil && pv.Filter.Filter != "" {
		_, err := models.GetTaskFiltersFromFilterString(pv.Filter.Filter, pv.Filter.FilterTimezone)
		if err != nil {
			return err
		}
	}

	// Check if the project view exists
	_, err := pvs.GetByIDAndProject(s, pv.ID, pv.ProjectID)
	if err != nil {
		return err
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
	return err
}

// Delete removes a project view and all associated buckets and task positions.
func (pvs *ProjectViewService) Delete(s *xorm.Session, viewID int64, projectID int64) error {
	_, err := s.
		Where("id = ? AND project_id = ?", viewID, projectID).
		Delete(&models.ProjectView{})
	if err != nil {
		return err
	}

	_, err = s.Where("project_view_id = ?", viewID).Delete(&models.TaskBucket{})
	if err != nil {
		return err
	}

	_, err = s.Where("project_view_id = ?", viewID).Delete(&models.TaskPosition{})
	return err
}

// GetAll retrieves all project views for a specific project.
func (pvs *ProjectViewService) GetAll(s *xorm.Session, projectID int64, a web.Auth) (views []*models.ProjectView, totalCount int64, err error) {
	// Check permissions
	pp := &models.Project{ID: projectID}
	can, _, err := pp.CanRead(s, a)
	if err != nil {
		return nil, 0, err
	}
	if !can {
		return nil, 0, models.ErrGenericForbidden{}
	}

	views, err = pvs.getViewsForProject(s, projectID)
	if err != nil {
		return nil, 0, err
	}

	totalCount, err = s.
		Where("project_id = ?", projectID).
		Count(&models.ProjectView{})
	if err != nil {
		return nil, 0, err
	}

	return views, totalCount, nil
}

// getViewsForProject is a helper that retrieves all views for a project.
func (pvs *ProjectViewService) getViewsForProject(s *xorm.Session, projectID int64) (views []*models.ProjectView, err error) {
	views = []*models.ProjectView{}
	err = s.
		Where("project_id = ?", projectID).
		OrderBy("position asc").
		Find(&views)
	return
}

// GetByIDAndProject retrieves a project view by ID and project ID without permission checks
// This is a simple lookup helper used by permission methods
// MIGRATION: Exposed in T-PERM-004 (migrated from models.GetProjectViewByIDAndProject)
func (pvs *ProjectViewService) GetByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *models.ProjectView, err error) {
	if projectID == models.FavoritesPseudoProjectID && viewID < 0 {
		for _, v := range models.FavoritesPseudoProject.Views {
			if v.ID == viewID {
				return v, nil
			}
		}

		return nil, &models.ErrProjectViewDoesNotExist{
			ProjectViewID: viewID,
		}
	}

	view = &models.ProjectView{}
	exists, err := s.
		Where("id = ? AND project_id = ?", viewID, projectID).
		NoAutoCondition().
		Get(view)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &models.ErrProjectViewDoesNotExist{
			ProjectViewID: viewID,
		}
	}

	return
}

// GetByID retrieves a project view by ID without permission checks
// This is a simple lookup helper used by permission methods
// MIGRATION: Exposed in T-PERM-004 (migrated from models.GetProjectViewByID)
func (pvs *ProjectViewService) GetByID(s *xorm.Session, id int64) (view *models.ProjectView, err error) {
	view = &models.ProjectView{}
	exists, err := s.
		Where("id = ?", id).
		NoAutoCondition().
		Get(view)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &models.ErrProjectViewDoesNotExist{
			ProjectViewID: id,
		}
	}

	return
}

// CreateDefaultViewsForProject creates the default views (List, Gantt, Table, Kanban) for a new project.
func (pvs *ProjectViewService) CreateDefaultViewsForProject(s *xorm.Session, project *models.Project, a web.Auth, createBacklogBucket bool, createDefaultListFilter bool) error {
	list := &models.ProjectView{
		ProjectID: project.ID,
		Title:     "List",
		ViewKind:  models.ProjectViewKindList,
		Position:  100,
	}
	if createDefaultListFilter {
		list.Filter = &models.TaskCollection{
			Filter: "done = false",
		}
	}
	err := pvs.Create(s, list, a, createBacklogBucket, true)
	if err != nil {
		return err
	}

	gantt := &models.ProjectView{
		ProjectID: project.ID,
		Title:     "Gantt",
		ViewKind:  models.ProjectViewKindGantt,
		Position:  200,
	}
	err = pvs.Create(s, gantt, a, createBacklogBucket, true)
	if err != nil {
		return err
	}

	table := &models.ProjectView{
		ProjectID: project.ID,
		Title:     "Table",
		ViewKind:  models.ProjectViewKindTable,
		Position:  300,
	}
	err = pvs.Create(s, table, a, createBacklogBucket, true)
	if err != nil {
		return err
	}

	kanban := &models.ProjectView{
		ProjectID:               project.ID,
		Title:                   "Kanban",
		ViewKind:                models.ProjectViewKindKanban,
		Position:                400,
		BucketConfigurationMode: models.BucketConfigurationModeManual,
	}
	err = pvs.Create(s, kanban, a, createBacklogBucket, true)
	if err != nil {
		return err
	}

	project.Views = []*models.ProjectView{
		list,
		gantt,
		table,
		kanban,
	}

	return nil
}

// Permission Methods (T-PERM-011)

// CanRead checks if the user can read a project view.
// For saved filters, delegates to the saved filter's permission check.
// Otherwise, checks if user can read the project.
// MIGRATION: Migrated from models.ProjectView.CanRead
func (pvs *ProjectViewService) CanRead(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error) {
	// Handle saved filters
	filterID := models.GetSavedFilterIDFromProjectID(projectID)
	if filterID > 0 {
		sf := &models.SavedFilter{ID: filterID}
		return sf.CanRead(s, a)
	}

	// Check project read permission
	return pvs.Registry.Project().CanRead(s, projectID, a)
}

// CanCreate checks if the user can create a project view.
// For saved filters, requires update permission on the saved filter.
// Otherwise, requires admin permission on the project.
// MIGRATION: Migrated from models.ProjectView.CanCreate
func (pvs *ProjectViewService) CanCreate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Handle saved filters
	filterID := models.GetSavedFilterIDFromProjectID(projectID)
	if filterID > 0 {
		sf := &models.SavedFilter{ID: filterID}
		return sf.CanUpdate(s, a)
	}

	// Require admin permission on project
	return pvs.Registry.Project().IsAdmin(s, projectID, a)
}

// CanUpdate checks if the user can update a project view.
// For saved filters, requires update permission on the saved filter.
// Otherwise, requires admin permission on the project.
// MIGRATION: Migrated from models.ProjectView.CanUpdate
func (pvs *ProjectViewService) CanUpdate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Handle saved filters
	filterID := models.GetSavedFilterIDFromProjectID(projectID)
	if filterID > 0 {
		sf := &models.SavedFilter{ID: filterID}
		return sf.CanUpdate(s, a)
	}

	// Require admin permission on project
	return pvs.Registry.Project().IsAdmin(s, projectID, a)
}

// CanDelete checks if the user can delete a project view.
// For saved filters, requires delete permission on the saved filter.
// Otherwise, requires admin permission on the project.
// MIGRATION: Migrated from models.ProjectView.CanDelete
func (pvs *ProjectViewService) CanDelete(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Handle saved filters
	filterID := models.GetSavedFilterIDFromProjectID(projectID)
	if filterID > 0 {
		sf := &models.SavedFilter{ID: filterID}
		return sf.CanDelete(s, a)
	}

	// Require admin permission on project
	return pvs.Registry.Project().IsAdmin(s, projectID, a)
}
