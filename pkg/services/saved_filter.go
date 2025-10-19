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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// InitSavedFilterService initializes the saved filter service dependency injection.
func InitSavedFilterService() {
	// Set up dependency injection for models to use service layer functions
	models.CreateSavedFilterFunc = func(s *xorm.Session, sf *models.SavedFilter, u *user.User) error {
		sfs := NewSavedFilterService(s.Engine())
		return sfs.Create(s, sf, u)
	}
	models.UpdateSavedFilterFunc = func(s *xorm.Session, sf *models.SavedFilter, u *user.User) error {
		sfs := NewSavedFilterService(s.Engine())
		return sfs.Update(s, sf, u)
	}
	models.DeleteSavedFilterFunc = func(s *xorm.Session, filterID int64, u *user.User) error {
		sfs := NewSavedFilterService(s.Engine())
		return sfs.Delete(s, filterID, u)
	}
	models.GetSavedFilterByIDFunc = func(s *xorm.Session, id int64) (*models.SavedFilter, error) {
		sfs := NewSavedFilterService(s.Engine())
		return sfs.GetByIDSimple(s, id)
	}
	// Register permission delegation
	models.RegisterSavedFilterService(&savedFilterServiceDelegator{})
}

// savedFilterServiceDelegator implements SavedFilterServiceProvider for dependency injection
type savedFilterServiceDelegator struct{}

func (d *savedFilterServiceDelegator) CanRead(s *xorm.Session, filterID int64, a web.Auth) (bool, int, error) {
	sfs := NewSavedFilterService(s.Engine())
	return sfs.CanRead(s, filterID, a)
}

func (d *savedFilterServiceDelegator) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	sfs := NewSavedFilterService(s.Engine())
	return sfs.CanCreate(s, a)
}

func (d *savedFilterServiceDelegator) CanUpdate(s *xorm.Session, filterID int64, a web.Auth) (bool, error) {
	sfs := NewSavedFilterService(s.Engine())
	return sfs.CanUpdate(s, filterID, a)
}

func (d *savedFilterServiceDelegator) CanDelete(s *xorm.Session, filterID int64, a web.Auth) (bool, error) {
	sfs := NewSavedFilterService(s.Engine())
	return sfs.CanDelete(s, filterID, a)
}

// SavedFilterService represents a service for managing saved filters.
type SavedFilterService struct {
	DB *xorm.Engine
}

// NewSavedFilterService creates a new SavedFilterService.
func NewSavedFilterService(db *xorm.Engine) *SavedFilterService {
	return &SavedFilterService{
		DB: db,
	}
}

// Get returns a saved filter by its ID and checks if the user has permission to access it.
func (sfs *SavedFilterService) Get(s *xorm.Session, filterID int64, u *user.User) (*models.SavedFilter, error) {
	sf := &models.SavedFilter{}
	exists, err := s.
		Where("id = ?", filterID).
		Get(sf)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, models.ErrSavedFilterDoesNotExist{SavedFilterID: filterID}
	}

	// Permission check: Only the owner can access the filter.
	if sf.OwnerID != u.ID {
		return nil, ErrAccessDenied
	}

	sf.Owner = u
	return sf, nil
}

// GetByIDSimple gets a saved filter by its ID without permission checks.
// MIGRATION: This is a simple lookup helper migrated from models layer (T-PERM-004).
// No permission checks are performed - use Get() for permission-aware retrieval.
func (sfs *SavedFilterService) GetByIDSimple(s *xorm.Session, id int64) (*models.SavedFilter, error) {
	sf := &models.SavedFilter{}
	exists, err := s.
		Where("id = ?", id).
		Get(sf)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, models.ErrSavedFilterDoesNotExist{SavedFilterID: id}
	}
	return sf, nil
}

// GetAllForUser returns all saved filters for a user.
func (sfs *SavedFilterService) GetAllForUser(s *xorm.Session, u *user.User, search string) ([]*models.SavedFilter, error) {
	// Link shares can't view or modify saved filters, therefore we can error out right away
	// This check is implicit by using a *user.User, but we'll keep it for clarity.
	if u == nil {
		return nil, ErrAccessDenied
	}

	filters := make([]*models.SavedFilter, 0)
	query := s.Where("owner_id = ?", u.ID)
	if search != "" {
		query = query.And("title LIKE ?", "%"+search+"%")
	}
	err := query.Find(&filters)
	if err != nil {
		return nil, err
	}

	for _, sf := range filters {
		sf.Owner = u
	}

	return filters, nil
}

// Create creates a new saved filter.
func (sfs *SavedFilterService) Create(s *xorm.Session, sf *models.SavedFilter, u *user.User) error {
	// Validate filter string
	_, err := models.GetTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return err
	}

	sf.OwnerID = u.ID
	sf.ID = 0
	_, err = s.Insert(sf)
	if err != nil {
		return err
	}

	// Create default views for this saved filter's pseudo-project
	err = models.CreateDefaultViewsForProject(s, &models.Project{ID: models.GetProjectIDFromSavedFilterID(sf.ID)}, u, true, false)
	return err
}

// Update updates a saved filter.
func (sfs *SavedFilterService) Update(s *xorm.Session, sf *models.SavedFilter, u *user.User) error {
	// Permission check
	origFilter, err := sfs.Get(s, sf.ID, u)
	if err != nil {
		return err
	}

	// If filters are not provided in update, preserve original
	if sf.Filters == nil {
		sf.Filters = origFilter.Filters
	}

	// Validate filter string
	_, err = models.GetTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return err
	}

	// Update the saved filter record
	_, err = s.
		Where("id = ?", sf.ID).
		Cols(
			"title",
			"description",
			"filters",
			"is_favorite",
		).
		Update(sf)
	if err != nil {
		return err
	}

	// Synchronize kanban views: Add all tasks which are not already in a bucket to the default bucket
	kanbanFilterViews := []*models.ProjectView{}
	err = s.Where(
		"project_id = ? and view_kind = ? and bucket_configuration_mode = ?",
		models.GetProjectIDFromSavedFilterID(sf.ID),
		models.ProjectViewKindKanban,
		models.BucketConfigurationModeManual,
	).
		Find(&kanbanFilterViews)
	if err != nil || len(kanbanFilterViews) == 0 {
		return err
	}

	parsedFilters, err := models.GetTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return err
	}

	filterCond, err := models.ConvertFiltersToDBFilterCond(parsedFilters, sf.Filters.FilterIncludeNulls)
	if err != nil {
		return err
	}

	taskBuckets := []*models.TaskBucket{}
	taskPositions := []*models.TaskPosition{}

	for _, view := range kanbanFilterViews {
		// Fetch all tasks in the filter but not in task_bucket
		// select * from tasks where id not in (select task_id from task_buckets where project_view_id = ?) and FILTER_COND
		tasksToAdd := []*models.Task{}
		err = s.Where(builder.And(
			builder.NotIn("id",
				builder.
					Select("task_id").
					From("task_buckets").
					Where(builder.Eq{"project_view_id": view.ID})),
			filterCond,
		)).
			Find(&tasksToAdd)
		if err != nil {
			return err
		}

		bucketID, err := models.GetDefaultBucketID(s, view)
		if err != nil {
			return err
		}

		for _, task := range tasksToAdd {
			taskBuckets = append(taskBuckets, &models.TaskBucket{
				TaskID:        task.ID,
				BucketID:      bucketID,
				ProjectViewID: view.ID,
			})

			taskPositions = append(taskPositions, &models.TaskPosition{
				TaskID:        task.ID,
				ProjectViewID: view.ID,
				Position:      0,
			})
		}
	}

	if len(taskBuckets) > 0 && len(taskPositions) > 0 {
		_, err = s.Insert(taskBuckets)
		if err != nil {
			return err
		}
		_, err = s.Insert(taskPositions)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes a saved filter.
func (sfs *SavedFilterService) Delete(s *xorm.Session, filterID int64, u *user.User) error {
	// Permission check
	_, err := sfs.Get(s, filterID, u)
	if err != nil {
		return err
	}

	_, err = s.
		Where("id = ?", filterID).
		Delete(&models.SavedFilter{})
	return err
}

// CanRead checks if a user has permission to read a saved filter
func (sfs *SavedFilterService) CanRead(s *xorm.Session, filterID int64, a web.Auth) (bool, int, error) {
	can, err := sfs.canDoFilter(s, filterID, a)
	return can, int(models.PermissionAdmin), err
}

// CanCreate checks if a user has permission to create a saved filter
func (sfs *SavedFilterService) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	// Link shares can't create saved filters
	if _, is := a.(*models.LinkSharing); is {
		return false, nil
	}
	return true, nil
}

// CanUpdate checks if a user has permission to update a saved filter
func (sfs *SavedFilterService) CanUpdate(s *xorm.Session, filterID int64, a web.Auth) (bool, error) {
	return sfs.canDoFilter(s, filterID, a)
}

// CanDelete checks if a user has permission to delete a saved filter
func (sfs *SavedFilterService) CanDelete(s *xorm.Session, filterID int64, a web.Auth) (bool, error) {
	return sfs.canDoFilter(s, filterID, a)
}

// canDoFilter is a helper function to check saved filter permissions
func (sfs *SavedFilterService) canDoFilter(s *xorm.Session, filterID int64, a web.Auth) (bool, error) {
	// Link shares can't view or modify saved filters
	if _, is := a.(*models.LinkSharing); is {
		return false, nil
	}

	// Get the saved filter
	sf := &models.SavedFilter{}
	exists, err := s.Where("id = ?", filterID).Get(sf)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	// Only owners are allowed to do something with a saved filter
	return sf.OwnerID == a.GetID(), nil
}
