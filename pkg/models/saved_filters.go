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
	"slices"
	"time"

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
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"filter" readOnly:"true" doc:"The unique, numeric id of this saved filter."`
	// The actual filters this filter contains
	Filters *TaskCollection `xorm:"JSON not null" json:"filters" valid:"required" doc:"The task filter query and collection options this saved filter wraps."`
	// The title of the filter.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The title of the filter."`
	// The description of the filter
	Description string `xorm:"longtext null" json:"description" doc:"The description of the filter."`
	OwnerID     int64  `xorm:"bigint not null INDEX" json:"-"`

	// The user who owns this filter
	Owner *user.User `xorm:"-" json:"owner" valid:"-" readOnly:"true" doc:"The user who owns this filter; set by the server."`

	// True if the filter is a favorite. Favorite filters show up in a separate parent project together with favorite projects.
	IsFavorite bool `xorm:"default false" json:"is_favorite" doc:"If true, the filter shows up in the Favorites pseudo-project alongside favorite projects."`

	// A timestamp when this filter was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this filter was created. You cannot change this value."`
	// A timestamp when this filter was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this filter was last updated. You cannot change this value."`

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
	if share, is := auth.(*LinkSharing); is {
		return nil, ErrSavedFilterNotAvailableForLinkShare{LinkShareID: share.ID}
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
		ID:              getProjectIDFromSavedFilterID(sf.ID),
		Title:           sf.Title,
		Description:     sf.Description,
		IsFavorite:      sf.IsFavorite,
		Created:         sf.Created,
		Updated:         sf.Updated,
		Owner:           sf.Owner,
		ParentProjectID: noParentProjectID(),
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

// ReadAll shadows the embedded web.CRUDable method (filters are listed via the
// pseudo-project). An unshadowed promoted method breaks Huma's $schema wrapper (go#15924).
func (sf *SavedFilter) ReadAll(_ *xorm.Session, _ web.Auth, _ string, _ int, _ int) (interface{}, int, int64, error) {
	return nil, 0, 0, ErrGenericForbidden{}
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

	sf.OwnerID = origFilter.OwnerID

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

	err = lockViewsForPositionUpdate(s, kanbanFilterViews)
	if err != nil {
		return err
	}

	for _, view := range kanbanFilterViews {
		taskIDs, err := filteredTasksWithoutBucketInView(s, view.ID, sf)
		if err != nil {
			return err
		}
		if len(taskIDs) == 0 {
			continue
		}

		bucketID, err := getDefaultBucketID(s, view)
		if err != nil {
			return err
		}

		if err = insertTaskBuckets(s, view.ID, bucketID, taskIDs); err != nil {
			return err
		}

		// New tasks have no position in this view yet
		if err = RecalculateTaskPositions(s, view, &user.User{ID: sf.OwnerID}); err != nil {
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

// dropFiltersWithInactiveOwners removes filters owned by a disabled, locked or deleted
// user. Evaluating a filter needs its owner's project list, which getRawProjectsForUser
// refuses to return for those, so one such filter would fail the whole batch for everyone.
func dropFiltersWithInactiveOwners(s *xorm.Session, filters map[int64]*SavedFilter) (timezoneByOwner map[int64]string, err error) {
	timezoneByOwner = map[int64]string{}
	if len(filters) == 0 {
		return timezoneByOwner, nil
	}

	seenOwners := map[int64]bool{}
	ownerIDs := make([]int64, 0, len(filters))
	for _, filter := range filters {
		if seenOwners[filter.OwnerID] {
			continue
		}
		seenOwners[filter.OwnerID] = true
		ownerIDs = append(ownerIDs, filter.OwnerID)
	}

	for chunk := range slices.Chunk(ownerIDs, idChunkSize) {
		activeOwners := []*user.User{}
		err = s.
			In("id", chunk).
			NotIn("status", user.StatusDisabled, user.StatusAccountLocked).
			Cols("id", "timezone").
			Find(&activeOwners)
		if err != nil {
			return nil, err
		}

		for _, owner := range activeOwners {
			timezoneByOwner[owner.ID] = owner.Timezone
		}
	}

	for id, filter := range filters {
		if _, isActive := timezoneByOwner[filter.OwnerID]; !isActive {
			log.Debugf("Skipping filter %d, owner %d is disabled, locked or deleted", filter.ID, filter.OwnerID)
			delete(filters, id)
		}
	}

	return timezoneByOwner, nil
}

func parseFilterCond(filter, timezone string, includeNulls bool) (cond builder.Cond, joinTaskBuckets bool, err error) {
	parsedFilters, err := getTaskFiltersFromFilterString(filter, timezone)
	if err != nil {
		return nil, false, err
	}

	// convertFiltersToDBFilterCond rewrites the field names in place, so this has
	// to be answered before it runs.
	joinTaskBuckets = hasBucketIDInParsedFilter(parsedFilters)

	cond, err = convertFiltersToDBFilterCond(parsedFilters, includeNulls)
	return cond, joinTaskBuckets, err
}

type filterView struct {
	view   *ProjectView
	filter *SavedFilter
}

type viewTask struct {
	viewID int64
	taskID int64
}

// Existing bucket and position rows are preloaded so the per-task loop issues no
// existence queries; default bucket ids are memoized on first use.
type filterViewState struct {
	hasBucket        map[viewTask]bool
	hasPosition      map[viewTask]bool
	defaultBucketIDs map[int64]int64
}

func (state *filterViewState) defaultBucketID(s *xorm.Session, view *ProjectView) (bucketID int64, err error) {
	bucketID, has := state.defaultBucketIDs[view.ID]
	if has {
		return bucketID, nil
	}

	bucketID, err = getDefaultBucketID(s, view)
	if err != nil {
		return 0, err
	}
	state.defaultBucketIDs[view.ID] = bucketID
	return bucketID, nil
}

func preloadFilterViewState(s *xorm.Session, viewsByTask map[int64][]filterView) (state *filterViewState, err error) {
	state = &filterViewState{
		hasBucket:        map[viewTask]bool{},
		hasPosition:      map[viewTask]bool{},
		defaultBucketIDs: map[int64]int64{},
	}

	taskIDs := make([]int64, 0, len(viewsByTask))
	viewIDs := []int64{}
	seenViews := map[int64]bool{}
	for taskID, views := range viewsByTask {
		if len(views) == 0 {
			continue
		}
		taskIDs = append(taskIDs, taskID)
		for _, fv := range views {
			if !seenViews[fv.view.ID] {
				seenViews[fv.view.ID] = true
				viewIDs = append(viewIDs, fv.view.ID)
			}
		}
	}
	if len(taskIDs) == 0 || len(viewIDs) == 0 {
		return state, nil
	}

	cond := builder.And(
		builder.In("task_id", taskIDs),
		builder.In("project_view_id", viewIDs),
	)

	taskBuckets := []*TaskBucket{}
	err = s.Where(cond).Find(&taskBuckets)
	if err != nil {
		return nil, err
	}
	for _, tb := range taskBuckets {
		state.hasBucket[viewTask{viewID: tb.ProjectViewID, taskID: tb.TaskID}] = true
	}

	taskPositions := []*TaskPosition{}
	err = s.Where(cond).Find(&taskPositions)
	if err != nil {
		return nil, err
	}
	for _, tp := range taskPositions {
		state.hasPosition[viewTask{viewID: tp.ProjectViewID, taskID: tp.TaskID}] = true
	}

	return state, nil
}

func addTaskToFilterView(s *xorm.Session, filter *SavedFilter, view *ProjectView, task *Task, state *filterViewState) (taskBucket *TaskBucket, taskPosition *TaskPosition, err error) {
	if !state.hasBucket[viewTask{viewID: view.ID, taskID: task.ID}] {
		bucketID, err := state.defaultBucketID(s, view)
		if err != nil {
			return nil, nil, err
		}

		taskBucket = &TaskBucket{
			BucketID:      bucketID,
			TaskID:        task.ID,
			ProjectViewID: view.ID,
		}
	}

	if !state.hasPosition[viewTask{viewID: view.ID, taskID: task.ID}] {
		taskPosition, err = calculateNewPositionForTask(s, &user.User{ID: filter.OwnerID}, task, view)
		if err != nil {
			return nil, nil, err
		}
	}

	return
}

// Keeps IN clauses well below the parameter limits of all supported databases.
const idChunkSize = 100

func getActiveSavedFiltersOwnedBy(s *xorm.Session, ownerIDs []int64) (filters map[int64]*SavedFilter, timezoneByOwner map[int64]string, err error) {
	filters = map[int64]*SavedFilter{}
	for chunk := range slices.Chunk(ownerIDs, idChunkSize) {
		batch := map[int64]*SavedFilter{}
		err = s.In("owner_id", chunk).Find(&batch)
		if err != nil {
			return nil, nil, err
		}
		for id, filter := range batch {
			filters[id] = filter
		}
	}

	timezoneByOwner, err = dropFiltersWithInactiveOwners(s, filters)
	if err != nil {
		return nil, nil, err
	}
	return filters, timezoneByOwner, nil
}

func getKanbanFilterViewsForFilters(s *xorm.Session, filters map[int64]*SavedFilter) (views []*ProjectView, err error) {
	if len(filters) == 0 {
		return nil, nil
	}

	filterProjectIDs := make([]int64, 0, len(filters))
	for _, filter := range filters {
		filterProjectIDs = append(filterProjectIDs, getProjectIDFromSavedFilterID(filter.ID))
	}

	views = []*ProjectView{}
	for chunk := range slices.Chunk(filterProjectIDs, idChunkSize) {
		chunkViews := []*ProjectView{}
		err = s.And(
			builder.Eq{"view_kind": ProjectViewKindKanban},
			builder.Eq{"bucket_configuration_mode": BucketConfigurationModeManual},
			builder.In("project_id", chunk),
		).Find(&chunkViews)
		if err != nil {
			return nil, err
		}
		views = append(views, chunkViews...)
	}
	return views, nil
}

func matchTasksToFilterViews(s *xorm.Session, tasks []*Task, filters map[int64]*SavedFilter, views []*ProjectView, accessByProject map[int64]map[int64]bool, timezoneByOwner map[int64]string) (viewsByTask map[int64][]filterView, err error) {
	viewsByTask = map[int64][]filterView{}

	filterIDs := []int64{}
	viewsByFilter := map[int64][]*ProjectView{}
	for _, view := range views {
		filterID := GetSavedFilterIDFromProjectID(view.ProjectID)
		if _, has := viewsByFilter[filterID]; !has {
			filterIDs = append(filterIDs, filterID)
		}
		viewsByFilter[filterID] = append(viewsByFilter[filterID], view)
	}

	for _, filterID := range filterIDs {
		filter := filters[filterID]
		matched, err := matchTasksToViewsOfFilter(s, tasks, filter, viewsByFilter[filterID], accessByProject, timezoneByOwner[filter.OwnerID])
		if err != nil {
			return nil, err
		}
		for taskID, matchedViews := range matched {
			viewsByTask[taskID] = append(viewsByTask[taskID], matchedViews...)
		}
	}

	return viewsByTask, nil
}

func matchTasksToViewsOfFilter(s *xorm.Session, tasks []*Task, filter *SavedFilter, views []*ProjectView, accessByProject map[int64]map[int64]bool, fallbackTimezone string) (viewsByTask map[int64][]filterView, err error) {
	viewsByTask = map[int64][]filterView{}

	if filter.Filters == nil {
		log.Warningf("Skipping filter %d, it has no filters", filter.ID)
		return viewsByTask, nil
	}

	timezone := filter.Filters.FilterTimezone
	if timezone == "" {
		timezone = fallbackTimezone
	}

	cond, joinTaskBuckets, err := parseFilterCond(filter.Filters.Filter, timezone, filter.Filters.FilterIncludeNulls)
	if err != nil {
		if !isErrInvalidFilter(err) {
			return nil, err
		}
		log.Warningf("Skipping filter %d, it cannot be parsed: %v", filter.ID, err)
		return viewsByTask, nil
	}

	candidateIDs := []int64{}
	for _, task := range tasks {
		if accessByProject[task.ProjectID][filter.OwnerID] {
			candidateIDs = append(candidateIDs, task.ID)
		}
	}
	if len(candidateIDs) == 0 {
		return viewsByTask, nil
	}

	matchCond := builder.And(cond, builder.In("tasks.id", candidateIDs))

	// A filter on bucket_id resolves against task_buckets, which only has meaning per view.
	if joinTaskBuckets {
		for _, view := range views {
			matchingIDs, err := taskIDsMatchingInView(s, view.ID, matchCond)
			if err != nil {
				return nil, err
			}
			for _, id := range matchingIDs {
				viewsByTask[id] = append(viewsByTask[id], filterView{view: view, filter: filter})
			}
		}
		return viewsByTask, nil
	}

	matchingIDs, err := taskIDsMatching(s, matchCond)
	if err != nil {
		return nil, err
	}
	for _, id := range matchingIDs {
		for _, view := range views {
			viewsByTask[id] = append(viewsByTask[id], filterView{view: view, filter: filter})
		}
	}

	return viewsByTask, nil
}

// Insert per task so a mid-loop recalculation sees earlier members' rows.
func addTaskToFilterViews(s *xorm.Session, task *Task, views []filterView, state *filterViewState) (err error) {
	taskPositions := []*TaskPosition{}

	for _, fv := range views {
		taskBucket, taskPosition, err := addTaskToFilterView(s, fv.filter, fv.view, task, state)
		if err != nil {
			return err
		}

		if taskBucket != nil {
			if err := insertTaskBuckets(s, taskBucket.ProjectViewID, taskBucket.BucketID, []int64{task.ID}); err != nil {
				return err
			}
		}
		if taskPosition != nil {
			taskPositions = append(taskPositions, taskPosition)
		}
	}

	if len(taskPositions) == 0 {
		return nil
	}

	// The cron writes the same (task_id, project_view_id) key every minute, so skip rows it already created.
	return bulkInsertTaskPositions(s, taskPositions, false)
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

		_, err = dropFiltersWithInactiveOwners(s, filters)
		if err != nil {
			log.Errorf("%sError checking filter owners: %s", logPrefix, err)
			return
		}

		if len(filters) == 0 {
			return
		}

		kanbanFilterViews, err := getKanbanFilterViewsForFilters(s, filters)
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

		viewsToRecalc := map[int64]struct {
			view    *ProjectView
			ownerID int64
		}{}
		staleTaskIDsByView := map[int64][]int64{}
		staleViews := []*ProjectView{}
		for _, view := range kanbanFilterViews {
			filterID := GetSavedFilterIDFromProjectID(view.ProjectID)
			filter := filters[filterID]

			// currently saved
			tasks, has := filterTasksCache[filterID]
			if !has {
				tc := &TaskCollection{
					ProjectID: view.ProjectID,
				}
				resultTasks, _, _, err := tc.ReadAll(s, &user.User{ID: filter.OwnerID}, "", 1, -1)
				if err != nil {
					log.Errorf("%sError fetching tasks for filter %d: %s", logPrefix, filterID, err)
					continue
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
					// Mark view for recalculation - RecalculateTaskPositions will create
					// positions for all tasks including new ones
					if _, ok := viewsToRecalc[view.ID]; !ok {
						viewsToRecalc[view.ID] = struct {
							view    *ProjectView
							ownerID int64
						}{view: view, ownerID: filter.OwnerID}
					}
				}
			}

			// Remove tasks that should not be there
			if staleIDs := staleFilterTaskIDs(savedTaskBucketMap, tasks); len(staleIDs) > 0 {
				staleTaskIDsByView[view.ID] = staleIDs
				staleViews = append(staleViews, view)
			}
		}

		// The loop above reads for seconds; only lock the views actually written, right before writing them.
		viewsToLock := staleViews
		for _, data := range viewsToRecalc {
			viewsToLock = append(viewsToLock, data.view)
		}
		if len(viewsToLock) > 0 {
			if err := lockViewsForPositionUpdate(s, viewsToLock); err != nil {
				log.Errorf("%sError locking kanban filter views: %s", logPrefix, err)
				return
			}
		}

		for viewID, staleIDs := range staleTaskIDsByView {
			deleteStaleFilterTasks(s, logPrefix, viewID, staleIDs)
		}

		upsertRelatedTaskProperties(s, logPrefix, newTaskBuckets, newTaskPositions)

		for _, data := range viewsToRecalc {
			if err := RecalculateTaskPositions(s, data.view, &user.User{ID: data.ownerID}); err != nil {
				log.Errorf("%sError recalculating task positions for view %d: %s", logPrefix, data.view.ID, err)
			}
		}

		if err := s.Commit(); err != nil {
			log.Errorf("%sError committing: %s", logPrefix, err)
		}
	})
	if err != nil {
		log.Fatalf("Could register add task to filter view cron: %s", err)
	}
}

func upsertRelatedTaskProperties(s *xorm.Session, logPrefix string, newTaskBuckets []*TaskBucket, newTaskPositions []*TaskPosition) {
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
}

func staleFilterTaskIDs(savedTaskBucketMap map[int64]*TaskBucket, tasks []*Task) (taskIDs []int64) {
	for taskID := range savedTaskBucketMap {
		found := false
		for _, task := range tasks {
			if task.ID == taskID {
				found = true
				break
			}
		}
		if !found {
			taskIDs = append(taskIDs, taskID)
		}
	}
	return taskIDs
}

func deleteStaleFilterTasks(s *xorm.Session, logPrefix string, viewID int64, taskIDs []int64) {
	if len(taskIDs) == 0 {
		return
	}

	_, err := s.Where(builder.Eq{"project_view_id": viewID}).
		And(builder.In("task_id", taskIDs)).
		Delete(&TaskBucket{})
	if err != nil {
		log.Errorf("%sError deleting task buckets: %s", logPrefix, err)
	}
	_, err = s.Where(builder.Eq{"project_view_id": viewID}).
		And(builder.In("task_id", taskIDs)).
		Delete(&TaskPosition{})
	if err != nil {
		log.Errorf("%sError deleting task positions: %s", logPrefix, err)
	}
}
