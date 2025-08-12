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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// SavedFilter represents a saved bunch of filters
type SavedFilter struct {
	// The unique numeric id of this saved filter
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"filter"`
	// The actual filters this filter contains
	Filters *TaskCollection `xorm:"JSON not null" json:"filters" valid:"required"`
	// The title of the filter.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the filter
	Description string `xorm:"longtext null" json:"description"`
	OwnerID     int64  `xorm:"bigint not null INDEX" json:"-"`

	// The user who owns this filter
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// True if the filter is a favorite. Favorite filters show up in a separate parent project together with favorite projects.
	IsFavorite bool `xorm:"default false" json:"is_favorite"`

	// A timestamp when this filter was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this filter was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName returns a better table name for saved filters
func (sf *SavedFilter) TableName() string {
	return "saved_filters"
}

func (sf *SavedFilter) getTaskCollection() *TaskCollection {
	// We're resetting the projectID to return tasks from all projects
	sf.Filters.ProjectID = 0
	return sf.Filters
}

// GetSavedFilterIDFromProjectID returns the saved filter ID from a project ID. Will not check if the filter actually exists.
// If the returned ID is zero, means that it is probably invalid.
func GetSavedFilterIDFromProjectID(projectID int64) (filterID int64) {
	// We get the id of the saved filter by multiplying the ProjectID with -1 and subtracting one
	filterID = projectID*-1 - 1
	// FilterIDs from projectIDs are always positive
	if filterID < 0 {
		filterID = 0
	}
	return
}

func getProjectIDFromSavedFilterID(filterID int64) (projectID int64) {
	projectID = filterID*-1 - 1
	// ProjectIDs from saved filters are always negative
	if projectID > 0 {
		projectID = 0
	}
	return
}

func getSavedFiltersForUser(s *xorm.Session, auth web.Auth, search string) (filters []*SavedFilter, err error) {
	// Link shares can't view or modify saved filters, therefore we can error out right away
	if _, is := auth.(*LinkSharing); is {
		return nil, ErrSavedFilterNotAvailableForLinkShare{LinkShareID: auth.GetID()}
	}

	query := s.Where("owner_id = ?", auth.GetID())
	if search != "" {
		query = query.And("title LIKE ?", "%"+search+"%")
	}
	err = query.Find(&filters)
	return
}

func (sf *SavedFilter) ToProject() *Project {
	return &Project{
		ID:          getProjectIDFromSavedFilterID(sf.ID),
		Title:       sf.Title,
		Description: sf.Description,
		IsFavorite:  sf.IsFavorite,
		Created:     sf.Created,
		Updated:     sf.Updated,
		Owner:       sf.Owner,
	}
}

// Create creates a new saved filter
// @Summary Creates a new saved filter
// @Description Creates a new saved filter
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 201 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters [put]
func (sf *SavedFilter) Create(s *xorm.Session, auth web.Auth) (err error) {
	_, err = getTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return
	}

	sf.OwnerID = auth.GetID()
	sf.ID = 0
	_, err = s.Insert(sf)
	if err != nil {
		return
	}

	err = CreateDefaultViewsForProject(s, &Project{ID: getProjectIDFromSavedFilterID(sf.ID)}, auth, true, false)
	return err
}

func GetSavedFilterSimpleByID(s *xorm.Session, id int64) (sf *SavedFilter, err error) {
	sf = &SavedFilter{}
	exists, err := s.
		Where("id = ?", id).
		Get(sf)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSavedFilterDoesNotExist{SavedFilterID: id}
	}
	return
}

// ReadOne returns one saved filter
// @Summary Gets one saved filter
// @Description Returns a saved filter by its ID.
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Filter ID"
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters/{id} [get]
func (sf *SavedFilter) ReadOne(s *xorm.Session, _ web.Auth) error {
	// s already contains almost the full saved filter from the permissions check, we only need to add the user
	u, err := user.GetUserByID(s, sf.OwnerID)
	sf.Owner = u
	return err
}

// Update updates an existing filter
// @Summary Updates a saved filter
// @Description Updates a saved filter by its ID.
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Filter ID"
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 404 {object} web.HTTPError "The saved filter does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters/{id} [post]
func (sf *SavedFilter) Update(s *xorm.Session, _ web.Auth) error {
	origFilter, err := GetSavedFilterSimpleByID(s, sf.ID)
	if err != nil {
		return err
	}

	if sf.Filters == nil {
		sf.Filters = origFilter.Filters
	}

	_, err = getTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return err
	}

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

	// Add all tasks which are not already in a bucket to the default bucket
	kanbanFilterViews := []*ProjectView{}
	err = s.Where(
		"project_id = ? and view_kind = ? and bucket_configuration_mode = ?",
		getProjectIDFromSavedFilterID(sf.ID),
		ProjectViewKindKanban,
		BucketConfigurationModeManual,
	).
		Find(&kanbanFilterViews)
	if err != nil || len(kanbanFilterViews) == 0 {
		return err
	}

	parsedFilters, err := getTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return err
	}

	filterCond, err := convertFiltersToDBFilterCond(parsedFilters, sf.Filters.FilterIncludeNulls)
	if err != nil {
		return err
	}

	taskBuckets := []*TaskBucket{}
	taskPositions := []*TaskPosition{}

	for _, view := range kanbanFilterViews {
		// Fetch all tasks in the filter but not in task_bucket
		// select * from tasks where id not in (select task_id from task_buckets where project_view_id = ?) and FILTER_COND
		tasksToAdd := []*Task{}
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

		bucketID, err := getDefaultBucketID(s, view)
		if err != nil {
			return err
		}

		for _, task := range tasksToAdd {
			taskBuckets = append(taskBuckets, &TaskBucket{
				TaskID:        task.ID,
				BucketID:      bucketID,
				ProjectViewID: view.ID,
			})

			taskPositions = append(taskPositions, &TaskPosition{
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

// Delete removes a saved filter
// @Summary Removes a saved filter
// @Description Removes a saved filter by its ID.
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Filter ID"
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 404 {object} web.HTTPError "The saved filter does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters/{id} [delete]
func (sf *SavedFilter) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.
		Where("id = ?", sf.ID).
		Delete(sf)
	return err
}

func addTaskToFilter(s *xorm.Session, filter *SavedFilter, view *ProjectView, fallbackTimezone string, task *Task) (taskBucket *TaskBucket, taskPosition *TaskPosition, err error) {

	filterString := filter.Filters.Filter

	if filter.Filters.FilterTimezone == "" {
		filter.Filters.FilterTimezone = fallbackTimezone
	}

	parsedFilters, err := getTaskFiltersFromFilterString(filterString, filter.Filters.FilterTimezone)
	if err != nil {
		log.Errorf("Could not parse filter string '%s' from view %d and saved filter %d: %v", filterString, view.ID, filter.ID, err)
		return
	}

	filterCond, err := convertFiltersToDBFilterCond(parsedFilters, filter.Filters.FilterIncludeNulls)
	if err != nil {
		log.Errorf("Could not convert filter string '%s' from view %d and saved filter %d to db conditions: %v", filterString, view.ID, filter.ID, err)
		return
	}

	taskIsInCurrentFilterAndView, err := s.Where(builder.And(
		filterCond,
		builder.Eq{"id": task.ID},
	)).Exist(&Task{})
	if !taskIsInCurrentFilterAndView {
		return
	}
	if err != nil {
		return nil, nil, err
	}

	taskHasBucketInView, err := s.Where(builder.And(
		builder.Eq{"task_id": task.ID},
		builder.Eq{"project_view_id": view.ID},
	)).Exist(&TaskBucket{})
	if err != nil {
		return nil, nil, err
	}
	if !taskHasBucketInView {
		bucketID, err := getDefaultBucketID(s, view)
		if err != nil {
			return nil, nil, err
		}

		taskBucket = &TaskBucket{
			BucketID:      bucketID,
			TaskID:        task.ID,
			ProjectViewID: view.ID,
		}
	}

	// Check if a position for this task already exists
	existingTaskPosition, err := s.Where(builder.And(
		builder.Eq{"task_id": task.ID},
		builder.Eq{"project_view_id": view.ID},
	)).Exist(&TaskPosition{})
	if err != nil {
		return nil, nil, err
	}
	if !existingTaskPosition {
		taskPosition = &TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			Position:      calculateDefaultPosition(task.Index, task.Position),
		}
	}

	return
}

func RegisterAddTaskToFilterViewCron() {
	const logPrefix = "[Add Task To Filter View Cron] "

	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		// Get all filters with a date clause and a manual kanban view
		where := "filters LIKE '%_date%'"
		if db.GetDialect() == builder.POSTGRES {
			where = "filters::jsonb ?| array['due_date', 'start_date', 'end_date']"
		}

		filters := map[int64]*SavedFilter{}
		err := s.Where(where).Find(&filters)
		if err != nil {
			log.Errorf("%sError fetching filters: %s", logPrefix, err)
			return
		}

		if len(filters) == 0 {
			return
		}

		filterProjectIDs := []int64{}
		for _, f := range filters {
			filterProjectIDs = append(filterProjectIDs, getProjectIDFromSavedFilterID(f.ID))
		}

		kanbanFilterViews := []*ProjectView{}
		err = s.And(
			builder.Eq{"view_kind": ProjectViewKindKanban},
			builder.Eq{"bucket_configuration_mode": BucketConfigurationModeManual},
			builder.In("project_id", filterProjectIDs),
		).
			Find(&kanbanFilterViews)
		if err != nil {
			log.Errorf("%sError fetching kanban filter views: %s", logPrefix, err)
			return
		}

		if len(kanbanFilterViews) == 0 {
			return
		}

		log.Debugf("%sFound %d kanban filter views with dates", logPrefix, len(kanbanFilterViews))

		filterTasksCache := make(map[int64][]*Task)
		newTaskBuckets := []*TaskBucket{}
		newTaskPositions := []*TaskPosition{}
		deleteCond := []builder.Cond{}
		taskIDsToRemove := []int64{}
		for _, view := range kanbanFilterViews {
			filterID := GetSavedFilterIDFromProjectID(view.ProjectID)
			filter, exists := filters[filterID]
			if !exists {
				log.Debugf("%sDid not find filter for view %d", logPrefix, view.ID)
				continue
			}

			// currently saved
			tasks, has := filterTasksCache[filterID]
			if !has {
				tc := &TaskCollection{
					ProjectID: view.ProjectID,
				}
				resultTasks, _, _, err := tc.ReadAll(s, &user.User{ID: filter.OwnerID}, "", 1, -1)
				if err != nil {
					log.Errorf("%sError fetching tasks for filter %d: %s", logPrefix, filterID, err)
					return
				}
				tasks = resultTasks.([]*Task)
			}

			// Get saved tasks in task_buckets and task_positions
			savedTaskBuckets := []*TaskBucket{}
			err = s.Where("project_view_id = ?", view.ID).Find(&savedTaskBuckets)
			if err != nil {
				log.Errorf("%sError fetching saved task buckets: %s", logPrefix, err)
				continue
			}
			savedTaskBucketMap := make(map[int64]*TaskBucket)
			for _, tb := range savedTaskBuckets {
				savedTaskBucketMap[tb.TaskID] = tb
			}

			savedTaskPositions := []*TaskPosition{}
			err = s.Where("project_view_id = ?", view.ID).Find(&savedTaskPositions)
			if err != nil {
				log.Errorf("%sError fetching saved task positions: %s", logPrefix, err)
				continue
			}
			savedTaskPositionMap := make(map[int64]*TaskPosition)
			for _, tp := range savedTaskPositions {
				savedTaskPositionMap[tp.TaskID] = tp
			}

			// Collect new tasks to task_buckets and task_positions
			for _, task := range tasks {
				if _, exists := savedTaskBucketMap[task.ID]; !exists {
					view.DefaultBucketID, err = getDefaultBucketID(s, view)
					if err != nil {
						log.Errorf("%sError fetching default bucket for view %d: %s", logPrefix, view.ID, err)
						continue
					}
					tb := &TaskBucket{
						TaskID:        task.ID,
						ProjectViewID: view.ID,
						BucketID:      view.DefaultBucketID,
					}
					newTaskBuckets = append(newTaskBuckets, tb)
				}
				if _, exists := savedTaskPositionMap[task.ID]; !exists {
					tp := &TaskPosition{
						TaskID:        task.ID,
						ProjectViewID: view.ID,
						Position:      task.Position,
					}
					newTaskPositions = append(newTaskPositions, tp)
				}
			}

			// Remove tasks that should not be there
			for taskID := range savedTaskBucketMap {
				found := false
				for _, task := range tasks {
					if task.ID == taskID {
						found = true
						break
					}
				}
				if !found {
					deleteCond = append(deleteCond, builder.And(
						builder.Eq{"task_id": taskID},
						builder.Eq{"project_view_id": view.ID},
					))
					taskIDsToRemove = append(taskIDsToRemove, taskID)
				}
			}
		}

		upsertRelatedTaskProperties(s, logPrefix, newTaskBuckets, newTaskPositions, deleteCond, taskIDsToRemove)
	})
	if err != nil {
		log.Fatalf("Could register add task to filter view cron: %s", err)
	}
}

func upsertRelatedTaskProperties(s *xorm.Session, logPrefix string, newTaskBuckets []*TaskBucket, newTaskPositions []*TaskPosition, deleteCond []builder.Cond, taskIDsToRemove []int64) {
	var err error
	if len(newTaskBuckets) > 0 {
		_, err = s.Insert(newTaskBuckets)
		if err != nil {
			log.Errorf("%sError inserting task buckets: %s", logPrefix, err)
		}
	}
	if len(newTaskPositions) > 0 {
		_, err = s.Insert(newTaskPositions)
		if err != nil {
			log.Errorf("%sError inserting task positions: %s", logPrefix, err)
		}
	}

	if len(deleteCond) > 0 {
		_, err = s.Where(builder.Or(deleteCond...)).Delete(&TaskBucket{})
		if err != nil {
			log.Errorf("%sError deleting task buckets: %s", logPrefix, err)
		}
		_, err = s.Where(builder.Or(deleteCond...)).Delete(&TaskPosition{})
		if err != nil {
			log.Errorf("%sError deleting task positions: %s", logPrefix, err)
		}
	}

	if config.TypesenseEnabled.GetBool() && (len(newTaskPositions) > 0 || len(taskIDsToRemove) > 0) {
		taskIDs := []int64{}
		for _, position := range newTaskPositions {
			taskIDs = append(taskIDs, position.TaskID)
		}
		taskIDs = append(taskIDs, taskIDsToRemove...)
		tasks, err := GetTasksSimpleByIDs(s, taskIDs)
		if err != nil {
			log.Errorf("%sError fetching tasks: %s", logPrefix, err)
			return
		}
		taskMap := make(map[int64]*Task)
		for _, t := range tasks {
			taskMap[t.ID] = t
		}

		err = reindexTasksInTypesense(s, taskMap)
		if err != nil {
			log.Errorf("%sError reindexing tasks into Typesense: %s", logPrefix, err)
			return
		}
	}
}
