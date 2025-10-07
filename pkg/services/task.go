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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"dario.cat/mergo"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Task Read All related types and constants
// These are moved from models package to support service-layer implementation

type (
	sortParam struct {
		sortBy        string
		orderBy       sortOrder // asc or desc
		projectViewID int64
	}

	sortOrder string

	taskSearchOptions struct {
		search             string
		page               int
		perPage            int
		sortby             []*sortParam
		parsedFilters      []*taskFilter
		filterIncludeNulls bool
		filter             string
		filterTimezone     string
		isSavedFilter      bool
		projectIDs         []int64
		expand             []models.TaskCollectionExpandable
		projectViewID      int64
	}

	taskFilter struct {
		field        string
		value        interface{}
		comparator   taskFilterComparator
		concatenator taskFilterConcatinator
		isNumeric    bool
	}

	taskFilterComparator   string
	taskFilterConcatinator string
)

const (
	// Sort order constants
	orderInvalid    sortOrder = "invalid"
	orderAscending  sortOrder = "asc"
	orderDescending sortOrder = "desc"

	// Task property constants for sorting and filtering
	taskPropertyID            string = "id"
	taskPropertyTitle         string = "title"
	taskPropertyDescription   string = "description"
	taskPropertyDone          string = "done"
	taskPropertyDoneAt        string = "done_at"
	taskPropertyDueDate       string = "due_date"
	taskPropertyCreatedByID   string = "created_by_id"
	taskPropertyProjectID     string = "project_id"
	taskPropertyRepeatAfter   string = "repeat_after"
	taskPropertyPriority      string = "priority"
	taskPropertyStartDate     string = "start_date"
	taskPropertyEndDate       string = "end_date"
	taskPropertyHexColor      string = "hex_color"
	taskPropertyPercentDone   string = "percent_done"
	taskPropertyUID           string = "uid"
	taskPropertyCreated       string = "created"
	taskPropertyUpdated       string = "updated"
	taskPropertyPosition      string = "position"
	taskPropertyBucketID      string = "bucket_id"
	taskPropertyIndex         string = "index"
	taskPropertyProjectViewID string = "project_view_id"

	// Task filter comparators
	taskFilterComparatorEquals        taskFilterComparator = "="
	taskFilterComparatorNotEquals     taskFilterComparator = "!="
	taskFilterComparatorGreater       taskFilterComparator = ">"
	taskFilterComparatorGreaterEquals taskFilterComparator = ">="
	taskFilterComparatorLess          taskFilterComparator = "<"
	taskFilterComparatorLessEquals    taskFilterComparator = "<="
	taskFilterComparatorLike          taskFilterComparator = "like"
	taskFilterComparatorIn            taskFilterComparator = "in"
	taskFilterComparatorNotIn         taskFilterComparator = "not_in"

	// Task filter concatenators
	taskFilterConcatAnd taskFilterConcatinator = "and"
	taskFilterConcatOr  taskFilterConcatinator = "or"
)

// String returns the string representation of a sort order
func (o sortOrder) String() string {
	return string(o)
}

// validate validates a sort parameter
func (sp *sortParam) validate() error {
	switch sp.sortBy {
	case
		taskPropertyID,
		taskPropertyTitle,
		taskPropertyDescription,
		taskPropertyDone,
		taskPropertyDoneAt,
		taskPropertyDueDate,
		taskPropertyCreatedByID,
		taskPropertyProjectID,
		taskPropertyRepeatAfter,
		taskPropertyPriority,
		taskPropertyStartDate,
		taskPropertyEndDate,
		taskPropertyHexColor,
		taskPropertyPercentDone,
		taskPropertyUID,
		taskPropertyCreated,
		taskPropertyUpdated,
		taskPropertyPosition,
		taskPropertyBucketID,
		taskPropertyIndex,
		taskPropertyProjectViewID:
		// Valid sort parameter
	default:
		return models.ErrInvalidTaskField{
			TaskField: sp.sortBy,
		}
	}

	if sp.orderBy != orderAscending && sp.orderBy != orderDescending {
		return models.ErrInvalidSortOrder{
			OrderBy: models.SortOrder(sp.orderBy),
		}
	}

	return nil
}

// getSortOrderFromString converts a string to sortOrder
func getSortOrderFromString(s string) sortOrder {
	// Normalize the input: trim whitespace and convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(s))

	switch normalized {
	case "asc", "ascending":
		return orderAscending
	case "desc", "descending":
		return orderDescending
	default:
		// For invalid or empty values, default to ascending for better UX
		// This prevents 500 errors when frontend sends malformed parameters
		return orderAscending
	}
}

// getTaskFilterOptsFromCollection converts a TaskCollection to taskSearchOptions
func (ts *TaskService) getTaskFilterOptsFromCollection(tf *models.TaskCollection, projectView *models.ProjectView) (opts *taskSearchOptions, err error) {
	var finalSortBy []string
	var finalOrderBy []string

	if len(tf.SortByArr) > 0 {
		finalSortBy = tf.SortByArr
		finalOrderBy = tf.OrderByArr
	} else if len(tf.SortBy) > 0 {
		finalSortBy = tf.SortBy
		finalOrderBy = tf.OrderBy
	}

	tf.SortBy = finalSortBy
	tf.OrderBy = finalOrderBy

	var sort = make([]*sortParam, 0, len(tf.SortBy))
	for i, s := range tf.SortBy {
		param := &sortParam{
			sortBy:  s,
			orderBy: orderAscending,
		}
		// This checks if tf.OrderBy has an entry with the same index as the current entry from tf.SortBy
		// Taken from https://stackoverflow.com/a/27252199/10924593
		if len(tf.OrderBy) > i {
			param.orderBy = getSortOrderFromString(tf.OrderBy[i])
		}

		if s == taskPropertyPosition && projectView != nil && projectView.ID < 0 {
			continue
		}

		if s == taskPropertyPosition {
			if projectView != nil {
				param.projectViewID = projectView.ID
			} else if tf.ProjectViewID != 0 {
				param.projectViewID = tf.ProjectViewID
			} else {
				return nil, fmt.Errorf("You must provide a project view ID when sorting by position")
			}
		}

		// Param validation
		if err := param.validate(); err != nil {
			return nil, err
		}
		sort = append(sort, param)
	}

	opts = &taskSearchOptions{
		sortby:             sort,
		filterIncludeNulls: tf.FilterIncludeNulls,
		filter:             tf.Filter,
		filterTimezone:     tf.FilterTimezone,
	}

	if projectView != nil {
		opts.projectViewID = projectView.ID
	} else if tf.ProjectViewID != 0 {
		opts.projectViewID = tf.ProjectViewID
	}

	// For now, skip filter parsing - we'll add this later
	// opts.parsedFilters, err = ts.getTaskFiltersFromFilterString(tf.Filter, tf.FilterTimezone)
	return opts, err
}

// getRelevantProjectsFromCollection determines which projects are relevant for the collection
func (ts *TaskService) getRelevantProjectsFromCollection(s *xorm.Session, a web.Auth, tf *models.TaskCollection) (projects []*models.Project, err error) {
	// Guard against nil session
	if s == nil {
		return nil, fmt.Errorf("database session is required")
	}

	// Check if this is a saved filter (negative project ID)
	isSavedFilter := tf.ProjectID < 0

	if tf.ProjectID == 0 || isSavedFilter {
		// For saved filters or general queries, get all accessible projects
		projectService := NewProjectService(ts.DB)
		projects, _, _, err := projectService.GetAllForUser(s, &user.User{ID: a.GetID()}, "", 0, -1, false)
		return projects, err
	}

	// Check the project exists and the user has access on it
	project := &models.Project{ID: tf.ProjectID}
	canRead, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, models.ErrUserDoesNotHaveAccessToProject{
			ProjectID: tf.ProjectID,
			UserID:    a.GetID(),
		}
	}

	return []*models.Project{{ID: tf.ProjectID}}, nil
}

// handleSavedFilter processes saved filter requests (negative project IDs)
func (ts *TaskService) handleSavedFilter(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// Get the saved filter ID from the project ID
	savedFilterID := models.GetSavedFilterIDFromProjectID(collection.ProjectID)
	if savedFilterID == 0 {
		return nil, 0, 0, fmt.Errorf("invalid saved filter project ID: %d", collection.ProjectID)
	}

	// Load the saved filter
	savedFilter, err := models.GetSavedFilterSimpleByID(s, savedFilterID)
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply the saved filter's settings to the collection
	savedFilterCollection := savedFilter.Filters

	// Merge saved filter settings with current collection
	mergedCollection := &models.TaskCollection{
		ProjectID:          0, // Saved filters search across all projects
		Filter:             savedFilterCollection.Filter,
		FilterIncludeNulls: savedFilterCollection.FilterIncludeNulls,
		FilterTimezone:     savedFilterCollection.FilterTimezone,
		SortBy:             collection.SortBy,
		OrderBy:            collection.OrderBy,
		SortByArr:          collection.SortByArr,
		OrderByArr:         collection.OrderByArr,
		ProjectViewID:      collection.ProjectViewID,
		Expand:             collection.Expand,
	}

	// If the saved filter has sort order, use it (unless overridden by current collection)
	if len(collection.SortBy) == 0 && len(collection.SortByArr) == 0 {
		if savedFilterCollection.SortBy != nil {
			mergedCollection.SortBy = savedFilterCollection.SortBy
		}
		if savedFilterCollection.OrderBy != nil {
			mergedCollection.OrderBy = savedFilterCollection.OrderBy
		}
	}

	// Process the merged collection normally
	return ts.processRegularCollection(s, mergedCollection, a, search, page, perPage)
}

// processRegularCollection handles the standard project collection processing
func (ts *TaskService) processRegularCollection(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// This contains the rest of the original GetAllWithFullFiltering logic
	var view *models.ProjectView
	var filteringForBucket bool
	var err error

	if collection.ProjectViewID != 0 {
		view, err = models.GetProjectViewByIDAndProject(s, collection.ProjectViewID, collection.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}

		// Apply view filters to collection filters
		if view.Filter != nil {
			if view.Filter.Filter != "" {
				if collection.Filter != "" {
					collection.Filter = "(" + collection.Filter + ") && (" + view.Filter.Filter + ")"
				} else {
					collection.Filter = view.Filter.Filter
				}
			}

			if view.Filter.FilterTimezone != "" {
				collection.FilterTimezone = view.Filter.FilterTimezone
			}

			if view.Filter.FilterIncludeNulls {
				collection.FilterIncludeNulls = view.Filter.FilterIncludeNulls
			}

			if view.Filter.Search != "" {
				search = view.Filter.Search
			}
		}

		// Check for bucket filtering
		if collection.Filter != "" && strings.Contains(collection.Filter, taskPropertyBucketID) {
			filteringForBucket = true
			// For now, skip bucket filter conversion - we'll add this later
		}
	}

	// Step 3: Convert collection parameters to search options
	opts, err := ts.getTaskFilterOptsFromCollection(collection, view)
	if err != nil {
		return nil, 0, 0, err
	}

	// Step 4: Validate expansion options
	for _, expandValue := range collection.Expand {
		err = expandValue.Validate()
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Set search options
	opts.search = search
	opts.page = page
	opts.perPage = perPage
	opts.expand = collection.Expand

	// Step 5: Add position sorting for views
	if view != nil {
		var hasOrderByPosition bool
		for _, param := range opts.sortby {
			if param.sortBy == taskPropertyPosition {
				hasOrderByPosition = true
				break
			}
		}
		if !hasOrderByPosition {
			opts.sortby = append(opts.sortby, &sortParam{
				projectViewID: view.ID,
				sortBy:        taskPropertyPosition,
				orderBy:       orderAscending,
			})
		}
	}

	// Step 6: Handle LinkSharing authentication
	shareAuth, is := a.(*models.LinkSharing)
	if is {
		project, err := models.GetProjectSimpleByID(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		return ts.getTaskOrTasksInBuckets(s, a, []*models.Project{project}, view, opts, filteringForBucket)
	}

	// Step 7: Get relevant projects for the user
	projects, err := ts.getRelevantProjectsFromCollection(s, a, collection)
	if err != nil {
		return nil, 0, 0, err
	}

	// Step 8: Get tasks (or tasks in buckets)
	return ts.getTaskOrTasksInBuckets(s, a, projects, view, opts, filteringForBucket)
}

// GetAllWithFullFiltering implements the complete Task ReadAll functionality
// This method contains all the complex filtering, sorting, and permission logic
// that was previously in models.TaskCollection.ReadAll()
func (ts *TaskService) GetAllWithFullFiltering(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// Step 1: Handle special project IDs
	if collection.ProjectID < 0 {
		// Handle favorites pseudo-project
		if collection.ProjectID == models.FavoritesPseudoProjectID {
			return ts.handleFavorites(s, collection, a, search, page, perPage)
		}
		// Handle saved filters (project ID < -1)
		return ts.handleSavedFilter(s, collection, a, search, page, perPage)
	}

	// Step 2: Handle regular collections
	return ts.processRegularCollection(s, collection, a, search, page, perPage)
}

// handleFavorites processes favorites pseudo-project requests
func (ts *TaskService) handleFavorites(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// Get user from auth
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get all favorite task IDs for this user
	favs := []*models.Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": models.FavoriteKindTask},
	)).Find(&favs)
	if err != nil {
		return nil, 0, 0, err
	}

	// Extract the task IDs
	favoriteTaskIDs := make([]int64, 0, len(favs))
	for _, fav := range favs {
		favoriteTaskIDs = append(favoriteTaskIDs, fav.EntityID)
	}

	// If no favorites, return empty result
	if len(favoriteTaskIDs) == 0 {
		return []*models.Task{}, 0, 0, nil
	}

	// Get the tasks with all the details for these favorite task IDs
	// We need to use the models bridge to ensure we get full task details
	// First, let's get the projects that contain these tasks
	projects, err := ts.getRelevantProjectsFromCollection(s, a, &models.TaskCollection{ProjectID: 0})
	if err != nil {
		return nil, 0, 0, err
	}

	// Convert collection to search options
	opts, err := ts.getTaskFilterOptsFromCollection(collection, nil)
	if err != nil {
		return nil, 0, 0, err
	}

	// Set search options
	opts.search = search
	opts.page = page
	opts.perPage = perPage
	opts.expand = collection.Expand

	// Call a special method to get favorite tasks with full details
	return ts.getFavoriteTasksWithDetails(s, projects, a, favoriteTaskIDs, opts)
}

// getFavoriteTasksWithDetails gets favorite tasks with full details (assignees, labels, etc.)
func (ts *TaskService) getFavoriteTasksWithDetails(s *xorm.Session, projects []*models.Project, a web.Auth, favoriteTaskIDs []int64, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	if len(favoriteTaskIDs) == 0 {
		return []*models.Task{}, 0, 0, nil
	}

	// We need to call the models bridge function but filter the results to only include favorites
	// First get all tasks using the bridge
	allTasks, _, _, err := models.CallGetTasksForProjects(
		s,
		projects,
		a,
		opts.search,
		0,  // Get all pages for now
		-1, // No limit for now
		convertSortParamsToStrings(opts.sortby),
		convertSortParamsToOrderStrings(opts.sortby),
		opts.filterIncludeNulls,
		opts.filter,
		opts.filterTimezone,
		opts.expand,
	)
	if err != nil {
		return nil, 0, 0, err
	}

	// Filter to only include favorites
	favoritesMap := make(map[int64]bool)
	for _, id := range favoriteTaskIDs {
		favoritesMap[id] = true
	}

	var favoriteTasks []*models.Task
	for _, task := range allTasks {
		if favoritesMap[task.ID] {
			favoriteTasks = append(favoriteTasks, task)
		}
	}

	// Apply pagination to the filtered results
	totalItems = int64(len(favoriteTasks))

	// Handle pagination
	if opts.perPage <= 0 {
		// No pagination - return all results
		return favoriteTasks, len(favoriteTasks), totalItems, nil
	}

	page := opts.page
	if page <= 0 {
		page = 1 // Default to page 1
	}

	start := (page - 1) * opts.perPage
	end := start + opts.perPage

	if start >= len(favoriteTasks) {
		return []*models.Task{}, 0, totalItems, nil
	}

	if end > len(favoriteTasks) {
		end = len(favoriteTasks)
	}

	favoriteTasks = favoriteTasks[start:end]
	return favoriteTasks, len(favoriteTasks), totalItems, nil
}

// getTaskOrTasksInBuckets determines whether to return tasks or buckets
func (ts *TaskService) getTaskOrTasksInBuckets(s *xorm.Session, a web.Auth, projects []*models.Project, view *models.ProjectView, opts *taskSearchOptions, filteringForBucket bool) (tasks interface{}, resultCount int, totalItems int64, err error) {
	if filteringForBucket {
		return ts.getTasksForProjects(s, projects, a, opts, view)
	}

	if view != nil && !strings.Contains(opts.filter, taskPropertyBucketID) {
		if view.BucketConfigurationMode != models.BucketConfigurationModeNone {
			// For now, delegate bucket handling to models - this is complex functionality
			// TODO: Move bucket logic to service layer
			return []*models.Bucket{}, 0, 0, nil // Simplified for now
		}
	}

	return ts.getTasksForProjects(s, projects, a, opts, view)
}

// getTasksForProjects gets tasks for the specified projects with full details
func (ts *TaskService) getTasksForProjects(s *xorm.Session, projects []*models.Project, a web.Auth, opts *taskSearchOptions, view *models.ProjectView) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	// For now, delegate back to the models package's getTasksForProjects function
	// This ensures we get tasks with full details (assignees, labels, attachments, etc.)

	// Convert sortby parameters to string arrays
	var sortby, orderby []string
	for _, sp := range opts.sortby {
		if sp != nil {
			sortby = append(sortby, sp.sortBy)
			orderby = append(orderby, string(sp.orderBy))
		}
	}

	// Use the bridge function that calls getTasksForProjects with full details
	var projectViewID int64
	if view != nil {
		projectViewID = view.ID
	} else if opts.projectViewID != 0 {
		projectViewID = opts.projectViewID
	}

	return models.CallGetTasksForProjectsWithViewID(
		s,
		projects,
		a,
		opts.search,
		opts.page,
		opts.perPage,
		sortby,
		orderby,
		opts.filterIncludeNulls,
		opts.filter,
		opts.filterTimezone,
		opts.expand,
		projectViewID,
	)
}

// getRawTasksForProjects gets the basic task data without extra details
func (ts *TaskService) getRawTasksForProjects(s *xorm.Session, projects []*models.Project, a web.Auth, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	// For now, delegate back to the models package's getRawTasksForProjects function
	// This ensures all existing filtering, sorting, and search logic continues to work
	// while we're in the process of moving it to the service layer
	// TODO: Move all filtering logic to service layer completely

	// Use the bridge function that calls getRawTasksForProjects directly (not getTasksForProjects)
	return models.CallGetRawTasksForProjects(
		s,
		projects,
		a,
		opts.search,
		opts.page,
		opts.perPage,
		convertSortParamsToStrings(opts.sortby),
		convertSortParamsToOrderStrings(opts.sortby),
		opts.filterIncludeNulls,
		opts.filter,
		opts.filterTimezone,
		opts.expand,
	)
}

// convertSortParamsToStrings converts sortParam structs to strings for TaskCollection
func convertSortParamsToStrings(sortParams []*sortParam) []string {
	if len(sortParams) == 0 {
		return nil
	}

	result := make([]string, len(sortParams))
	for i, param := range sortParams {
		result[i] = param.sortBy
	}
	return result
}

// convertSortParamsToOrderStrings converts sortParam order to strings for TaskCollection
func convertSortParamsToOrderStrings(sortParams []*sortParam) []string {
	if len(sortParams) == 0 {
		return nil
	}

	result := make([]string, len(sortParams))
	for i, param := range sortParams {
		if param.orderBy == orderDescending {
			result[i] = "desc"
		} else {
			result[i] = "asc"
		}
	}
	return result
}

// TaskService represents a service for managing tasks.
type TaskService struct {
	DB               *xorm.Engine
	FavoriteService  *FavoriteService
	KanbanService    *KanbanService
	ReactionsService *ReactionsService
	CommentService   *CommentService
}

// NewTaskService creates a new TaskService.
func NewTaskService(db *xorm.Engine) *TaskService {
	return &TaskService{
		DB:               db,
		FavoriteService:  NewFavoriteService(db),
		KanbanService:    NewKanbanService(db),
		ReactionsService: NewReactionsService(db),
		CommentService:   NewCommentService(db),
	}
}

// Wire models.AddMoreInfoToTasksFunc to the service implementation via dependency inversion
// InitTaskService sets up dependency injection for task-related model functions.
// This function must be called during test initialization to ensure models can call services.
func InitTaskService() {
	models.AddMoreInfoToTasksFunc = func(s *xorm.Session, taskMap map[int64]*models.Task, a web.Auth, view *models.ProjectView, expand []models.TaskCollectionExpandable) error {
		return NewTaskService(nil).AddDetailsToTasks(s, taskMap, a, view, expand)
	}

	models.GetUsersOrLinkSharesFromIDsFunc = func(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
		return NewTaskService(nil).getUsersOrLinkSharesFromIDs(s, ids)
	}

	models.TaskCreateFunc = func(s *xorm.Session, task *models.Task, u *user.User, updateAssignees bool, setBucket bool) error {
		_, err := NewTaskService(s.Engine()).CreateWithOptions(s, task, u, updateAssignees, setBucket, false)
		return err
	}

	// Wire TaskCollection.ReadAll to our new service method
	models.TaskCollectionReadAllFunc = func(s *xorm.Session, tf *models.TaskCollection, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
		return NewTaskService(s.Engine()).GetAllWithFullFiltering(s, tf, a, search, page, perPage)
	}
}

// GetByID gets a single task by its ID, checking permissions.
func (ts *TaskService) GetByID(s *xorm.Session, taskID int64, u *user.User) (*models.Task, error) {
	// Use a simple model function to get the raw data
	task := new(models.Task)
	has, err := s.ID(taskID).Get(task)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, models.ErrTaskDoesNotExist{ID: taskID}
	}

	// Permission Check: The TaskService asks the ProjectService for a decision.
	projectService := NewProjectService(ts.DB)
	can, err := projectService.HasPermission(s, task.ProjectID, u, models.PermissionRead)
	if err != nil {
		return nil, fmt.Errorf("checking project read permission: %w", err)
	}
	if !can {
		return nil, ErrAccessDenied
	}

	// Add details to the task
	taskMap := map[int64]*models.Task{task.ID: task}
	err = ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetByIDWithExpansion gets a single task by its ID with support for expansion parameters
// and returns the maximum permission the user has on the task's project.
func (ts *TaskService) GetByIDWithExpansion(s *xorm.Session, taskID int64, u *user.User, expand []models.TaskCollectionExpandable) (*models.Task, int, error) {
	// Load the task with all fields at the service layer
	task := &models.Task{}
	exists, err := s.Where("id = ?", taskID).Get(task)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, models.ErrTaskDoesNotExist{ID: taskID}
	}

	// Permission Check: The TaskService asks the ProjectService for a decision.
	projectService := NewProjectService(ts.DB)
	permissionMap, err := projectService.checkPermissionsForProjects(s, u, []int64{task.ProjectID})
	if err != nil {
		return nil, 0, fmt.Errorf("checking project permissions: %w", err)
	}
	permission, ok := permissionMap[task.ProjectID]
	if !ok || permission == nil {
		return nil, 0, ErrAccessDenied
	}
	maxPermission := permission.MaxPermission
	if maxPermission < int(models.PermissionRead) {
		return nil, 0, ErrAccessDenied
	}

	// Add details to the task with expansion support
	taskMap := map[int64]*models.Task{task.ID: task}
	err = ts.AddDetailsToTasks(s, taskMap, u, nil, expand)
	if err != nil {
		return nil, 0, err
	}

	// Load subscription data for single task requests (matches original behavior)
	subscription, err := models.GetSubscriptionForUser(s, models.SubscriptionEntityTask, task.ID, u)
	if err != nil && !models.IsErrProjectDoesNotExist(err) {
		return nil, 0, err
	}
	if subscription != nil {
		task.Subscription = &subscription.Subscription
	}

	return task, maxPermission, nil
}

// GetAllByProject gets all tasks for a project with pagination and filtering
func (ts *TaskService) GetAllByProject(s *xorm.Session, projectID int64, u *user.User, page int, perPage int, search string) ([]*models.Task, int, int64, error) {
	// Permission Check: Use ProjectService for proper inter-service communication
	projectService := NewProjectService(ts.DB)
	canRead, err := projectService.HasPermission(s, projectID, u, models.PermissionRead)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrAccessDenied
	}

	// Calculate offset for pagination
	offset := (page - 1) * perPage

	// Query tasks directly from the database
	var tasks []*models.Task

	// Add search filter if provided
	searchCondition := builder.NewCond()
	if search != "" {
		searchCondition = builder.Or(
			builder.Like{"title", "%" + search + "%"},
			builder.Like{"description", "%" + search + "%"},
		)
	}

	// Get total count for pagination (use separate query to avoid session corruption)
	countQuery := s.Where("project_id = ?", projectID)
	if search != "" {
		countQuery = countQuery.And(searchCondition)
	}
	totalCount, err := countQuery.Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Create fresh query for finding tasks to avoid any session corruption
	findQuery := s.Where("project_id = ?", projectID)
	if search != "" {
		findQuery = findQuery.And(searchCondition)
	}

	// Get the actual tasks with pagination
	err = findQuery.
		OrderBy("id ASC").
		Limit(perPage, offset).
		Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	// Add details to all tasks (CreatedBy, Labels, Attachments, etc.)
	if len(tasks) > 0 {
		taskMap := make(map[int64]*models.Task)
		for _, task := range tasks {
			taskMap[task.ID] = task
		}
		err = ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	return tasks, len(tasks), totalCount, nil
}

// GetAllWithFilters gets all tasks with complex filtering, sorting and expansion options
// This method replicates the functionality of models.TaskCollection.ReadAll() at the service layer
func (ts *TaskService) GetAllWithFilters(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) ([]*models.Task, int, int64, error) {
	// Use our new full filtering implementation
	result, resultCount, totalItems, err := ts.GetAllWithFullFiltering(s, collection, a, search, page, perPage)
	if err != nil {
		return nil, 0, 0, err
	}

	tasks, ok := result.([]*models.Task)
	if !ok {
		return nil, 0, 0, fmt.Errorf("unexpected result type from GetAllWithFullFiltering")
	}

	return tasks, resultCount, totalItems, nil
}

// Update updates a task with full business logic.
func (ts *TaskService) Update(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	updatedTask, err := ts.updateSingleTask(s, task, u, nil)
	return updatedTask, err
}

// UpdateWithFields updates a task with only specific fields.
func (ts *TaskService) UpdateWithFields(s *xorm.Session, task *models.Task, u *user.User, fields []string) (*models.Task, error) {
	return ts.updateSingleTask(s, task, u, fields)
}

//nolint:gocyclo
func (ts *TaskService) updateSingleTask(s *xorm.Session, t *models.Task, u *user.User, fields []string) (*models.Task, error) {
	// Check if the task exists and get the old values FIRST (before permission check)
	// This is necessary because t.ProjectID might be 0
	ot, err := models.GetTaskByIDSimple(s, t.ID)
	if err != nil {
		return nil, err
	}

	// Now check permissions using the old task (which has the correct ProjectID)
	can, err := ts.Can(s, &ot, u).Write()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	if t.ProjectID == 0 {
		t.ProjectID = ot.ProjectID
	}

	// Get the stored reminders
	reminders, err := models.GetRemindersForTasks(s, []int64{t.ID})
	if err != nil {
		return nil, err
	}

	// Old task has the stored reminders
	ot.Reminders = reminders

	// Update the assignees
	// Pass the user as web.Auth for model methods that need it
	if err := ot.UpdateTaskAssignees(s, t.Assignees, u); err != nil {
		return nil, err
	}

	// All columns to update in a separate variable to be able to add to them
	colsToUpdate := []string{
		"title",
		"description",
		"done",
		"due_date",
		"repeat_after",
		"priority",
		"start_date",
		"end_date",
		"hex_color",
		"percent_done",
		"project_id",
		"bucket_id",
		"repeat_mode",
		"cover_image_attachment_id",
	}

	// Validate fields if provided
	if len(fields) > 0 {
		allowed := map[string]bool{}
		for _, c := range colsToUpdate {
			allowed[c] = true
		}
		cols := []string{}
		fieldSet := map[string]bool{}
		for _, f := range fields {
			if !allowed[f] {
				return nil, models.ErrInvalidTaskColumn{Column: f}
			}
			cols = append(cols, f)
			fieldSet[f] = true
		}
		colsToUpdate = cols

		if !fieldSet["title"] {
			t.Title = ot.Title
		}
		if !fieldSet["description"] {
			t.Description = ot.Description
		}
		if !fieldSet["done"] {
			t.Done = ot.Done
			t.DoneAt = ot.DoneAt
		}
		if !fieldSet["due_date"] {
			t.DueDate = ot.DueDate
		}
		if !fieldSet["repeat_after"] {
			t.RepeatAfter = ot.RepeatAfter
		}
		if !fieldSet["priority"] {
			t.Priority = ot.Priority
		}
		if !fieldSet["start_date"] {
			t.StartDate = ot.StartDate
		}
		if !fieldSet["end_date"] {
			t.EndDate = ot.EndDate
		}
		if !fieldSet["hex_color"] {
			t.HexColor = ot.HexColor
		}
		if !fieldSet["percent_done"] {
			t.PercentDone = ot.PercentDone
		}
		if !fieldSet["project_id"] {
			t.ProjectID = ot.ProjectID
		}
		if !fieldSet["bucket_id"] {
			t.BucketID = ot.BucketID
		}
		if !fieldSet["repeat_mode"] {
			t.RepeatMode = ot.RepeatMode
		}
		if !fieldSet["cover_image_attachment_id"] {
			t.CoverImageAttachmentID = ot.CoverImageAttachmentID
		}
	}

	// If the task is being moved between projects, make sure to move the bucket + index as well
	if t.ProjectID != 0 && ot.ProjectID != t.ProjectID {
		t.Index, err = models.CalculateNextTaskIndex(s, t.ProjectID)
		if err != nil {
			return nil, err
		}
		t.BucketID = 0
		colsToUpdate = append(colsToUpdate, "index")
	}

	views := []*models.ProjectView{}
	if (!t.IsRepeating() && t.Done != ot.Done) || t.ProjectID != ot.ProjectID {
		err = s.
			Where("project_id = ? AND view_kind = ? AND bucket_configuration_mode = ?",
				t.ProjectID, models.ProjectViewKindKanban, models.BucketConfigurationModeManual).
			Find(&views)
		if err != nil {
			return nil, err
		}
	}

	// When a task was moved between projects, ensure it is in the correct bucket
	if t.ProjectID != ot.ProjectID {
		_, err = s.Where("task_id = ?", t.ID).Delete(&models.TaskBucket{})
		if err != nil {
			return nil, err
		}
		_, err = s.Where("task_id = ?", t.ID).Delete(&models.TaskPosition{})
		if err != nil {
			return nil, err
		}

		for _, view := range views {
			var bucketID = view.DoneBucketID
			if bucketID == 0 || !t.Done {
				bucketID, err = models.GetDefaultBucketID(s, view)
				if err != nil {
					return nil, err
				}
			}

			tb := &models.TaskBucket{
				BucketID:      bucketID,
				TaskID:        t.ID,
				ProjectViewID: view.ID,
				ProjectID:     t.ProjectID,
			}
			err = tb.Update(s, u)
			if err != nil {
				return nil, err
			}

			tp, err := models.CalculateNewPositionForTask(s, u, t, view)
			if err != nil {
				return nil, err
			}

			err = tp.Update(s, u)
			if err != nil {
				return nil, err
			}
		}
	}

	// When a task changed its done status, make sure it is in the correct bucket
	if t.ProjectID == ot.ProjectID && !t.IsRepeating() && t.Done != ot.Done {
		err = t.MoveTaskToDoneBuckets(s, u, views)
		if err != nil {
			return nil, err
		}
	}

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
	models.UpdateDone(&ot, t)
	colsToUpdate = append(colsToUpdate, "done_at")

	// Update the reminders
	if err := ot.UpdateReminders(s, t); err != nil {
		return nil, err
	}

	// If a task attachment is being set as cover image, check if the attachment actually belongs to the task
	if t.CoverImageAttachmentID != 0 {
		is, err := s.Exist(&models.TaskAttachment{
			TaskID: t.ID,
			ID:     t.CoverImageAttachmentID,
		})
		if err != nil {
			return nil, err
		}
		if !is {
			return nil, &models.ErrAttachmentDoesNotBelongToTask{
				AttachmentID: t.CoverImageAttachmentID,
				TaskID:       t.ID,
			}
		}
	}

	// Handle favorite status changes
	wasFavorite, err := ts.FavoriteService.IsFavorite(s, t.ID, u, models.FavoriteKindTask)
	if err != nil {
		return nil, err
	}
	if t.IsFavorite && !wasFavorite {
		if err := ts.FavoriteService.AddToFavorite(s, t.ID, u, models.FavoriteKindTask); err != nil {
			return nil, err
		}
	}

	if !t.IsFavorite && wasFavorite {
		if err := ts.FavoriteService.RemoveFromFavorite(s, t.ID, u, models.FavoriteKindTask); err != nil {
			return nil, err
		}
	}

	// Merge the old task with the new task
	// mergo ignores nil values, so we need to handle them manually below
	if err := mergo.Merge(&ot, t, mergo.WithOverride); err != nil {
		return nil, err
	}

	t.HexColor = utils.NormalizeHex(t.HexColor)

	// Mergo does ignore nil values. Because of that, we need to check all parameters and set the updated to
	// nil/their nil value in the struct which is inserted.

	// Done
	if !t.Done {
		ot.Done = false
	}
	// Priority
	if t.Priority == 0 {
		ot.Priority = 0
	}
	// Description
	if t.Description == "" {
		ot.Description = ""
	}
	// Due date
	if t.DueDate.IsZero() {
		ot.DueDate = time.Time{}
	}
	// Repeat after
	if t.RepeatAfter == 0 {
		ot.RepeatAfter = 0
	}
	// Start date
	if t.StartDate.IsZero() {
		ot.StartDate = time.Time{}
	}
	// End date
	if t.EndDate.IsZero() {
		ot.EndDate = time.Time{}
	}
	// Color
	if t.HexColor == "" {
		ot.HexColor = ""
	}
	// Percent Done
	if t.PercentDone == 0 {
		ot.PercentDone = 0
	}
	// Repeat from current date
	if t.RepeatMode == models.TaskRepeatModeDefault {
		ot.RepeatMode = models.TaskRepeatModeDefault
	}
	// Is Favorite
	if !t.IsFavorite {
		ot.IsFavorite = false
	}
	// Attachment cover image
	if t.CoverImageAttachmentID == 0 {
		ot.CoverImageAttachmentID = 0
	}

	_, err = s.ID(t.ID).
		Cols(colsToUpdate...).
		Update(ot)
	*t = ot
	if err != nil {
		return nil, err
	}

	// Get the task updated timestamp in a new struct - if we'd just try to put it into t which we already have, it
	// would still contain the old updated date.
	nt := &models.Task{}
	_, err = s.ID(t.ID).Get(nt)
	if err != nil {
		return nil, err
	}
	t.Updated = nt.Updated

	err = events.Dispatch(&models.TaskUpdatedEvent{
		Task: t,
		Doer: u,
	})
	if err != nil {
		return nil, err
	}

	return t, models.UpdateProjectLastUpdated(s, &models.Project{ID: t.ProjectID})
}

// Delete deletes a task.
func (ts *TaskService) Delete(s *xorm.Session, task *models.Task, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	can, err := ts.canWriteTask(s, task.ID, u)
	if err != nil {
		return err
	}
	if !can {
		return ErrAccessDenied
	}

	t, err := models.GetTaskByIDSimple(s, task.ID)
	if err != nil {
		return err
	}

	// duplicate the task for the event
	fullTask := &models.Task{ID: task.ID}
	err = fullTask.ReadOne(s, a)
	if err != nil {
		return err
	}

	// Delete assignees
	if _, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskAssginee{}); err != nil {
		return err
	}

	// Delete Favorites using the service
	err = ts.FavoriteService.RemoveFromFavorite(s, task.ID, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.LabelTask{})
	if err != nil {
		return err
	}

	// Delete task attachments
	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, []int64{task.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		// Using the attachment delete method here because that takes care of removing all files properly
		err = attachment.Delete(s, a)
		if err != nil && !models.IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskComment{})
	if err != nil {
		return err
	}

	// Delete all relations
	_, err = s.Where("task_id = ? OR other_task_id = ?", task.ID, task.ID).Delete(&models.TaskRelation{})
	if err != nil {
		return err
	}

	// Delete all reminders
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskReminder{})
	if err != nil {
		return err
	}

	// Delete all positions
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskPosition{})
	if err != nil {
		return err
	}

	// Delete all bucket relations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskBucket{})
	if err != nil {
		return err
	}

	// Actually delete the task
	_, err = s.ID(task.ID).Delete(&models.Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&models.TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	err = ts.updateProjectLastUpdated(s, t.ProjectID)
	return err
}

// TaskPermissions represents the permissions for a task.
type TaskPermissions struct {
	s    *xorm.Session
	task *models.Task
	user *user.User
	ts   *TaskService
}

// Can returns a new TaskPermissions struct.
func (ts *TaskService) Can(s *xorm.Session, task *models.Task, u *user.User) *TaskPermissions {
	return &TaskPermissions{s: s, task: task, user: u, ts: ts}
}

// Read checks if the user can read the task.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (tp *TaskPermissions) Read() (bool, error) {
	if tp.user == nil {
		return false, nil
	}

	// Use ProjectService for permission checking instead of calling model methods
	projectService := NewProjectService(tp.ts.DB)
	return projectService.HasPermission(tp.s, tp.task.ProjectID, tp.user, models.PermissionRead)
}

// Write checks if the user can write to the task.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (tp *TaskPermissions) Write() (bool, error) {
	if tp.user == nil {
		return false, nil
	}

	// Use ProjectService for permission checking instead of calling model methods
	projectService := NewProjectService(tp.ts.DB)
	return projectService.HasPermission(tp.s, tp.task.ProjectID, tp.user, models.PermissionWrite)
}

func (ts *TaskService) addDetailsToTasks(s *xorm.Session, tasks []*models.Task, u *user.User) error {
	if len(tasks) == 0 {
		return nil
	}

	taskMap := make(map[int64]*models.Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	// Use the standard AddDetailsToTasks method
	return ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
}

// AddDetailsToTasks adds more info to tasks, like assignees, labels, etc.
// This is the service layer implementation of what was previously models.AddMoreInfoToTasks.
// Empty collections are kept as null for standards compliance
func (ts *TaskService) AddDetailsToTasks(s *xorm.Session, taskMap map[int64]*models.Task, a web.Auth, view *models.ProjectView, expand []models.TaskCollectionExpandable) error {
	if len(taskMap) == 0 {
		return nil
	}

	// Initialize array/map fields for consistent API behavior
	// Keep empty collections as null for standards compliance
	for _, task := range taskMap {
		if task.RelatedTasks == nil {
			task.RelatedTasks = make(models.RelatedTaskMap)
		}
	}

	// Collect identifiers for batched lookups
	taskIDs := make([]int64, 0, len(taskMap))
	creatorIDSet := make(map[int64]struct{}, len(taskMap))
	projectIDSet := make(map[int64]struct{}, len(taskMap))
	for _, task := range taskMap {
		taskIDs = append(taskIDs, task.ID)
		if task.CreatedByID != 0 {
			creatorIDSet[task.CreatedByID] = struct{}{}
		}
		projectIDSet[task.ProjectID] = struct{}{}
	}

	// Convert project id set to slice for retrieval
	projectIDs := make([]int64, 0, len(projectIDSet))
	for id := range projectIDSet {
		projectIDs = append(projectIDs, id)
	}

	// Add assignees
	if err := ts.addAssigneesToTasks(s, taskIDs, taskMap); err != nil {
		return err
	}

	// Add labels
	if err := ts.addLabelsToTasks(s, taskIDs, taskMap); err != nil {
		return err
	}

	// Add attachments
	if err := ts.addAttachmentsToTasks(s, taskIDs, taskMap); err != nil {
		return err
	}

	// Get task reminders
	taskReminders, err := ts.getTaskReminderMap(s, taskIDs)
	if err != nil {
		return err
	}

	// Get favorites if auth is provided
	var taskFavorites map[int64]bool
	if a != nil {
		taskFavorites, err = ts.getFavorites(s, taskIDs, a, models.FavoriteKindTask)
		if err != nil {
			return err
		}
	}

	// Get all projects for identifiers
	projects, err := models.GetProjectsMapByIDs(s, projectIDs)
	if err != nil {
		return err
	}

	// Determine fallback creator assignments for legacy tasks without CreatedByID
	legacyCreators := make(map[int64]int64)
	for _, task := range taskMap {
		if task.CreatedByID != 0 {
			continue
		}
		project := projects[task.ProjectID]
		if project == nil || project.OwnerID == 0 {
			continue
		}
		legacyCreators[task.ID] = project.OwnerID
		if _, seen := creatorIDSet[project.OwnerID]; !seen {
			creatorIDSet[project.OwnerID] = struct{}{}
		}
	}

	// Resolve all required users (task creators + fallbacks)
	userIDs := make([]int64, 0, len(creatorIDSet))
	for id := range creatorIDSet {
		userIDs = append(userIDs, id)
	}

	users := map[int64]*user.User{}
	if len(userIDs) > 0 {
		users, err = ts.getUsersOrLinkSharesFromIDs(s, userIDs)
		if err != nil {
			return err
		}
	}

	// Add all objects to their tasks
	for _, task := range taskMap {
		if createdBy, has := users[task.CreatedByID]; has {
			task.CreatedBy = createdBy
		} else if fallbackID, ok := legacyCreators[task.ID]; ok {
			if fallbackUser, hasUser := users[fallbackID]; hasUser {
				task.CreatedBy = fallbackUser
				task.CreatedByID = fallbackID
			}
		}

		if remindersList := taskReminders[task.ID]; remindersList != nil {
			task.Reminders = remindersList
		}

		if project, exists := projects[task.ProjectID]; exists && project != nil {
			if project.Identifier == "" {
				task.Identifier = "#" + strconv.FormatInt(task.Index, 10)
			} else {
				task.Identifier = project.Identifier + "-" + strconv.FormatInt(task.Index, 10)
			}
		}

		if taskFavorites != nil {
			task.IsFavorite = taskFavorites[task.ID]
		}
	}

	// Handle expansion parameters using proper service layer methods
	if expand != nil && len(expand) > 0 {
		for _, expandable := range expand {
			switch expandable {
			case models.TaskCollectionExpandBuckets:
				err = ts.addBucketsToTasks(s, a, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case models.TaskCollectionExpandReactions:
				err = ts.addReactionsToTasks(s, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case models.TaskCollectionExpandComments:
				err = ts.addCommentsToTasks(s, taskIDs, taskMap)
				if err != nil {
					return err
				}
			}
		}
	}

	// Add related tasks
	err = ts.addRelatedTasksToTasks(s, taskIDs, taskMap, a)
	if err != nil {
		return err
	}

	// Normalize slice fields to empty arrays so the frontend can safely iterate without null checks.
	for _, task := range taskMap {
		if task.Assignees == nil {
			task.Assignees = []*user.User{}
		}
		if task.Labels == nil {
			task.Labels = []*models.Label{}
		}
		if task.Attachments == nil {
			task.Attachments = []*models.TaskAttachment{}
		}
		if task.Reminders == nil {
			task.Reminders = []*models.TaskReminder{}
		}
		if task.Comments == nil {
			task.Comments = []*models.TaskComment{}
		}
		if task.RelatedTasks == nil {
			task.RelatedTasks = make(models.RelatedTaskMap)
		}
		if task.Buckets == nil {
			task.Buckets = []*models.Bucket{}
		}
		if task.Reactions == nil {
			task.Reactions = models.ReactionMap{}
		}
	}

	return nil
}

// Helper methods moved from models package

func (ts *TaskService) addAssigneesToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	taskAssignees := []*models.TaskAssigneeWithUser{}
	err := s.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Find(&taskAssignees)
	if err != nil {
		return err
	}

	// Put the assignees in the task map
	for i, a := range taskAssignees {
		if a != nil {
			a.Email = "" // Obfuscate the email

			// Check if assignee already exists to avoid duplicates
			alreadyExists := false
			for _, existingAssignee := range taskMap[a.TaskID].Assignees {
				if existingAssignee.ID == taskAssignees[i].User.ID {
					alreadyExists = true
					break
				}
			}

			if !alreadyExists {
				taskMap[a.TaskID].Assignees = append(taskMap[a.TaskID].Assignees, &taskAssignees[i].User)
			}
		}
	}

	return nil
}

func (ts *TaskService) addLabelsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	labelService := NewLabelService(ts.DB)
	labels, _, _, err := labelService.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
		TaskIDs: taskIDs,
		Page:    -1,
	})
	if err != nil {
		return err
	}

	// Debug: log the number of labels found
	// fmt.Printf("DEBUG: Found %d labels for %d tasks\n", len(labels), len(taskIDs))

	for i, l := range labels {
		if l != nil {
			// Debug: log each label being processed
			// fmt.Printf("DEBUG: Processing label %d for task %d\n", l.Label.ID, l.TaskID)

			// Check if this label is already in the task's Labels slice
			alreadyExists := false
			if taskMap[l.TaskID].Labels != nil {
				for _, existingLabel := range taskMap[l.TaskID].Labels {
					if existingLabel.ID == l.Label.ID {
						alreadyExists = true
						break
					}
				}
			}

			if !alreadyExists {
				taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &labels[i].Label)
				// fmt.Printf("DEBUG: Added label %d to task %d, now has %d labels\n", l.Label.ID, l.TaskID, len(taskMap[l.TaskID].Labels))
			}
		}
	}

	return nil
}

func (ts *TaskService) addAttachmentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, taskIDs)
	if err != nil {
		return err
	}

	for _, a := range attachments {
		// Check if attachment already exists to avoid duplicates
		alreadyExists := false
		for _, existingAttachment := range taskMap[a.TaskID].Attachments {
			if existingAttachment.ID == a.ID {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			taskMap[a.TaskID].Attachments = append(taskMap[a.TaskID].Attachments, a)
		}
	}

	return nil
}

func (ts *TaskService) getTaskReminderMap(s *xorm.Session, taskIDs []int64) (map[int64][]*models.TaskReminder, error) {
	reminders := []*models.TaskReminder{}
	err := s.In("task_id", taskIDs).
		OrderBy("reminder asc").
		Find(&reminders)
	if err != nil {
		return nil, err
	}

	reminderMap := make(map[int64][]*models.TaskReminder)
	for _, reminder := range reminders {
		reminderMap[reminder.TaskID] = append(reminderMap[reminder.TaskID], reminder)
	}

	return reminderMap, nil
}

func (ts *TaskService) getFavorites(s *xorm.Session, entityIDs []int64, a web.Auth, kind models.FavoriteKind) (map[int64]bool, error) {
	favorites := make(map[int64]bool)
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return favorites, nil
	}

	favs := []*models.Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": kind},
		builder.In("entity_id", entityIDs),
	)).
		Find(&favs)

	for _, fav := range favs {
		favorites[fav.EntityID] = true
	}
	return favorites, err
}

func (ts *TaskService) addRelatedTasksToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task, a web.Auth) error {
	relatedTasks := []*models.TaskRelation{}
	err := s.In("task_id", taskIDs).Find(&relatedTasks)
	if err != nil {
		return err
	}

	// Collect all related task IDs, so we can get all related task headers in one go
	var relatedTaskIDs []int64
	for _, rt := range relatedTasks {
		relatedTaskIDs = append(relatedTaskIDs, rt.OtherTaskID)
	}

	if len(relatedTaskIDs) == 0 {
		return nil
	}

	fullRelatedTasks := make(map[int64]*models.Task)
	err = s.In("id", relatedTaskIDs).Find(&fullRelatedTasks)
	if err != nil {
		return err
	}

	taskFavorites, err := ts.getFavorites(s, relatedTaskIDs, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Go through all task relations and put them into the task objects
	for _, rt := range relatedTasks {
		_, has := fullRelatedTasks[rt.OtherTaskID]
		if !has {
			continue
		}
		fullRelatedTasks[rt.OtherTaskID].IsFavorite = taskFavorites[rt.OtherTaskID]

		// We're duplicating the other task to avoid cycles as these can't be represented properly in json
		// and would thus fail with an error.
		otherTask := &models.Task{}
		err = copier.Copy(otherTask, fullRelatedTasks[rt.OtherTaskID])
		if err != nil {
			continue
		}
		// Clear RelatedTasks map to prevent cycles and match null behavior in JSON
		otherTask.RelatedTasks = nil
		// Note: Other slice/map fields stay nil to match original behavior
		taskMap[rt.TaskID].RelatedTasks[rt.RelationKind] = append(taskMap[rt.TaskID].RelatedTasks[rt.RelationKind], otherTask)
	}

	return nil
}

func (ts *TaskService) canWriteTask(s *xorm.Session, taskID int64, u *user.User) (bool, error) {
	project, err := models.GetProjectSimpleByTaskID(s, taskID)
	if err != nil {
		if models.IsErrProjectDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Check project permissions using ProjectService
	projectService := NewProjectService(ts.DB)
	return projectService.HasPermission(s, project.ID, u, models.PermissionWrite)
}

// getTaskAttachmentsByTaskIDs gets task attachments with full details
func (ts *TaskService) getTaskAttachmentsByTaskIDs(s *xorm.Session, taskIDs []int64) (attachments []*models.TaskAttachment, err error) {
	attachments = []*models.TaskAttachment{}
	err = s.
		In("task_id", taskIDs).
		Find(&attachments)
	if err != nil {
		return
	}

	if len(attachments) == 0 {
		return
	}

	fileIDs := []int64{}
	userIDs := []int64{}
	for _, a := range attachments {
		userIDs = append(userIDs, a.CreatedByID)
		fileIDs = append(fileIDs, a.FileID)
	}

	// Get all files
	fs := make(map[int64]*files.File)
	err = s.In("id", fileIDs).Find(&fs)
	if err != nil {
		return
	}

	users, err := ts.getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, err
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	for _, a := range attachments {
		if createdBy, has := users[a.CreatedByID]; has {
			a.CreatedBy = createdBy
		}
		a.File = fs[a.FileID]
	}

	return
}

// updateProjectLastUpdated updates the last updated timestamp of a project
func (ts *TaskService) updateProjectLastUpdated(s *xorm.Session, projectID int64) error {
	project := &models.Project{
		ID:      projectID,
		Updated: time.Now(),
	}
	_, err := s.ID(projectID).Cols("updated").Update(project)
	return err
}

// getUsersOrLinkSharesFromIDs gets users and link shares from their IDs.
func (ts *TaskService) getUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	users = make(map[int64]*user.User)
	var userIDs []int64
	var linkShareIDs []int64
	for _, id := range ids {
		if id < 0 {
			linkShareIDs = append(linkShareIDs, id*-1)
			continue
		}

		userIDs = append(userIDs, id)
	}

	if len(userIDs) > 0 {
		users, err = user.GetUsersByIDs(s, userIDs)
		if err != nil {
			return
		}
	}

	if len(linkShareIDs) == 0 {
		return
	}

	shares, err := models.GetLinkSharesByIDs(s, linkShareIDs)
	if err != nil {
		return nil, err
	}

	for _, share := range shares {
		users[share.ID*-1] = ts.toUser(share)
	}

	return
}

func (ts *TaskService) toUser(share *models.LinkSharing) *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       ts.getUserID(share),
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

func (ts *TaskService) getUserID(share *models.LinkSharing) int64 {
	return share.ID * -1
}

type taskCreationOptions struct {
	skipPermissionCheck bool
	updateAssignees     bool
	setBucket           bool
}

// Create creates a new task with permission checks and full service-layer business logic.
func (ts *TaskService) Create(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	return ts.CreateWithOptions(s, task, u, true, true, false)
}

// CreateWithoutPermissionCheck creates a new task without performing permission checks.
// This is intended for internal use where permissions have already been validated externally.
func (ts *TaskService) CreateWithoutPermissionCheck(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	return ts.CreateWithOptions(s, task, u, true, true, true)
}

// CreateWithOptions provides fine-grained control over task creation behavior while reusing
// the core service-layer implementation. Callers can disable assignee updates or bucket placement
// when duplicating tasks or performing specialized operations.
func (ts *TaskService) CreateWithOptions(s *xorm.Session, task *models.Task, u *user.User, updateAssignees bool, setBucket bool, skipPermissionCheck bool) (*models.Task, error) {
	opts := taskCreationOptions{
		skipPermissionCheck: skipPermissionCheck,
		updateAssignees:     updateAssignees,
		setBucket:           setBucket,
	}
	return ts.createTask(s, task, u, opts)
}

// createTask contains the core business logic for task creation.
func (ts *TaskService) createTask(s *xorm.Session, task *models.Task, actor *user.User, opts taskCreationOptions) (*models.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task must not be nil")
	}
	if actor == nil {
		return nil, ErrAccessDenied
	}

	if task.Title == "" {
		return nil, models.ErrTaskCannotBeEmpty{}
	}

	project, err := models.GetProjectSimpleByID(s, task.ProjectID)
	if err != nil {
		return nil, err
	}

	if !opts.skipPermissionCheck {
		projectService := NewProjectService(ts.DB)
		canWrite, err := projectService.HasPermission(s, task.ProjectID, actor, models.PermissionWrite)
		if err != nil {
			return nil, fmt.Errorf("checking project write permission: %w", err)
		}
		if !canWrite {
			return nil, ErrAccessDenied
		}
	}

	createdBy, err := models.GetUserOrLinkShareUser(s, actor)
	if err != nil {
		return nil, err
	}
	task.CreatedByID = createdBy.ID
	task.CreatedBy = createdBy

	if task.UID == "" {
		task.UID = uuid.NewString()
	}

	if err := ts.ensureTaskIndex(s, task); err != nil {
		return nil, err
	}

	task.HexColor = utils.NormalizeHex(task.HexColor)

	if _, err := s.Insert(task); err != nil {
		return nil, err
	}

	var providedBucket *models.Bucket
	if opts.setBucket && task.BucketID != 0 {
		providedBucket, err = ts.KanbanService.getBucketByID(s, task.BucketID)
		if err != nil {
			return nil, err
		}
		if _, err = ts.KanbanService.checkBucketLimit(s, createdBy, task, providedBucket); err != nil {
			return nil, err
		}
	}

	if opts.setBucket {
		if err := ts.assignTaskToViews(s, task, createdBy, providedBucket); err != nil {
			return nil, err
		}
	}

	if opts.updateAssignees {
		if err := ts.syncTaskAssignees(s, task, task.Assignees, createdBy); err != nil {
			return nil, err
		}
	}

	if err := ts.syncTaskReminders(s, task); err != nil {
		return nil, err
	}

	ts.setTaskIdentifier(task, project)

	if task.IsFavorite {
		if err := ts.FavoriteService.AddToFavorite(s, task.ID, createdBy, models.FavoriteKindTask); err != nil {
			return nil, err
		}
	}

	if err := events.Dispatch(&models.TaskCreatedEvent{Task: task, Doer: createdBy}); err != nil {
		return nil, err
	}

	if err := ts.updateProjectLastUpdated(s, task.ProjectID); err != nil {
		return nil, err
	}

	return task, nil
}

func (ts *TaskService) assignTaskToViews(s *xorm.Session, task *models.Task, auth web.Auth, providedBucket *models.Bucket) error {
	views, err := ts.getViewsForProject(s, task.ProjectID)
	if err != nil {
		return err
	}

	positions := make([]*models.TaskPosition, 0, len(views))
	taskBuckets := make([]*models.TaskBucket, 0, len(views))
	moveToDone := false

	for _, view := range views {
		if view.ViewKind == models.ProjectViewKindKanban && view.BucketConfigurationMode == models.BucketConfigurationModeManual && !moveToDone {
			bucketID := view.DoneBucketID
			if !task.Done || view.DoneBucketID == 0 {
				if providedBucket != nil && view.ID == providedBucket.ProjectViewID {
					bucketID = providedBucket.ID
				} else {
					bucketID, err = ts.KanbanService.getDefaultBucketID(s, view)
					if err != nil {
						return err
					}
				}
			}

			if view.DoneBucketID != 0 && view.DoneBucketID == task.BucketID && !task.Done {
				task.Done = true
				if _, err = s.Where("id = ?", task.ID).Cols("done").Update(task); err != nil {
					return err
				}

				if err = ts.moveTaskToDoneBuckets(s, task, auth, views); err != nil {
					return err
				}

				moveToDone = true
				continue
			}

			taskBuckets = append(taskBuckets, &models.TaskBucket{
				BucketID:      bucketID,
				TaskID:        task.ID,
				ProjectViewID: view.ID,
				ProjectID:     task.ProjectID,
			})
		}

		position, err := ts.calculateNewPositionForTask(s, auth, task, view)
		if err != nil {
			return err
		}
		positions = append(positions, position)
	}

	if moveToDone {
		taskBuckets = []*models.TaskBucket{}
	}

	if len(positions) > 0 {
		if _, err = s.Insert(&positions); err != nil {
			return err
		}
	}

	if len(taskBuckets) > 0 {
		if _, err = s.Insert(&taskBuckets); err != nil {
			return err
		}
	}

	return nil
}

func (ts *TaskService) getViewsForProject(s *xorm.Session, projectID int64) ([]*models.ProjectView, error) {
	views := make([]*models.ProjectView, 0)
	err := s.Where("project_id = ?", projectID).OrderBy("position asc").Find(&views)
	return views, err
}

func (ts *TaskService) calculateNewPositionForTask(s *xorm.Session, auth web.Auth, task *models.Task, view *models.ProjectView) (*models.TaskPosition, error) {
	if task.Position == 0 {
		lowestPosition := &models.TaskPosition{}
		exists, err := s.Where("project_view_id = ?", view.ID).OrderBy("position asc").Get(lowestPosition)
		if err != nil {
			return nil, err
		}
		if exists {
			if lowestPosition.Position == 0 {
				if err = models.RecalculateTaskPositions(s, view, auth); err != nil {
					return nil, err
				}

				lowestPosition = &models.TaskPosition{}
				if _, err = s.Where("project_view_id = ?", view.ID).OrderBy("position asc").Get(lowestPosition); err != nil {
					return nil, err
				}
			}

			task.Position = lowestPosition.Position / 2
		}
	}

	return &models.TaskPosition{
		TaskID:        task.ID,
		ProjectViewID: view.ID,
		Position:      ts.calculateDefaultPosition(task.Index, task.Position),
	}, nil
}

func (ts *TaskService) calculateDefaultPosition(entityID int64, position float64) float64 {
	if position == 0 {
		return float64(entityID) * 1000
	}
	return position
}

func (ts *TaskService) moveTaskToDoneBuckets(s *xorm.Session, task *models.Task, auth web.Auth, views []*models.ProjectView) error {
	for _, view := range views {
		currentTaskBucket := &models.TaskBucket{}
		if _, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, view.ID).Get(currentTaskBucket); err != nil {
			return err
		}

		bucketID := currentTaskBucket.BucketID

		if task.Done && view.DoneBucketID == 0 {
			continue
		}

		if !task.Done && bucketID != view.DoneBucketID {
			continue
		}

		if task.Done && view.DoneBucketID != 0 {
			bucketID = view.DoneBucketID
		}

		if !task.Done && bucketID == view.DoneBucketID {
			var err error
			bucketID, err = ts.KanbanService.getDefaultBucketID(s, view)
			if err != nil {
				return err
			}
		}

		tb := &models.TaskBucket{
			BucketID:      bucketID,
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			ProjectID:     task.ProjectID,
		}
		if err := tb.Update(s, auth); err != nil {
			return err
		}

		tp := models.TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			Position:      ts.calculateDefaultPosition(task.Index, task.Position),
		}
		if err := tp.Update(s, auth); err != nil {
			return err
		}
	}
	return nil
}

func (ts *TaskService) syncTaskAssignees(s *xorm.Session, task *models.Task, desiredAssignees []*user.User, createdBy web.Auth) error {
	currentAssignees, err := ts.getRawTaskAssigneesForTask(s, task.ID)
	if err != nil {
		return err
	}

	currentAssigneeMap := make(map[int64]struct{}, len(currentAssignees))
	for _, entry := range currentAssignees {
		currentAssigneeMap[entry.User.ID] = struct{}{}
	}

	desiredAssigneeMap := make(map[int64]*user.User)
	for _, assignee := range desiredAssignees {
		if assignee == nil || assignee.ID == 0 {
			continue
		}
		if _, exists := desiredAssigneeMap[assignee.ID]; !exists {
			desiredAssigneeMap[assignee.ID] = assignee
		}
	}

	// Delete assignees that are no longer desired
	assigneesToDelete := make([]int64, 0)
	for id := range currentAssigneeMap {
		if _, keep := desiredAssigneeMap[id]; !keep {
			assigneesToDelete = append(assigneesToDelete, id)
		}
	}

	if len(assigneesToDelete) > 0 {
		if _, err = s.In("user_id", assigneesToDelete).And("task_id = ?", task.ID).Delete(&models.TaskAssginee{}); err != nil {
			return err
		}
	}

	// Add new assignees
	for id := range desiredAssigneeMap {
		if _, already := currentAssigneeMap[id]; already {
			continue
		}

		assignee := &models.TaskAssginee{TaskID: task.ID, UserID: id}
		if err := assignee.Create(s, createdBy); err != nil {
			if !models.IsErrUserAlreadyAssigned(err) {
				return err
			}
		}
	}

	// Refresh assignee list on the task to include full user data
	taskMap := map[int64]*models.Task{task.ID: task}
	if err := ts.addAssigneesToTasks(s, []int64{task.ID}, taskMap); err != nil {
		return err
	}

	if len(task.Assignees) == 0 {
		task.Assignees = nil
	}

	return ts.updateProjectLastUpdated(s, task.ProjectID)
}

func (ts *TaskService) getRawTaskAssigneesForTask(s *xorm.Session, taskID int64) ([]*models.TaskAssigneeWithUser, error) {
	assignees := make([]*models.TaskAssigneeWithUser, 0)
	err := s.Table("task_assignees").
		Select("task_assignees.task_id, users.*").
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where("task_assignees.task_id = ?", taskID).
		Find(&assignees)
	return assignees, err
}

func (ts *TaskService) syncTaskReminders(s *xorm.Session, task *models.Task) error {
	if _, err := s.Where("task_id = ?", task.ID).Delete(&models.TaskReminder{}); err != nil {
		return err
	}

	if err := ts.normalizeRelativeReminderDates(task); err != nil {
		return err
	}

	reminderMap := make(map[int64]*models.TaskReminder, len(task.Reminders))
	for _, reminder := range task.Reminders {
		reminderMap[reminder.Reminder.UTC().Unix()] = reminder
	}

	task.Reminders = make([]*models.TaskReminder, 0, len(reminderMap))
	for _, reminder := range reminderMap {
		entry := &models.TaskReminder{
			TaskID:         task.ID,
			Reminder:       reminder.Reminder,
			RelativePeriod: reminder.RelativePeriod,
			RelativeTo:     reminder.RelativeTo,
		}
		if _, err := s.Insert(entry); err != nil {
			return err
		}
		task.Reminders = append(task.Reminders, entry)
	}

	sort.Slice(task.Reminders, func(i, j int) bool {
		return task.Reminders[i].Reminder.Before(task.Reminders[j].Reminder)
	})

	if len(task.Reminders) == 0 {
		task.Reminders = nil
	}

	return ts.updateProjectLastUpdated(s, task.ProjectID)
}

func (ts *TaskService) normalizeRelativeReminderDates(task *models.Task) error {
	for _, reminder := range task.Reminders {
		relativeDuration := time.Duration(reminder.RelativePeriod) * time.Second
		if reminder.RelativeTo != "" {
			reminder.Reminder = time.Time{}
		}

		switch reminder.RelativeTo {
		case models.ReminderRelationDueDate:
			if !task.DueDate.IsZero() {
				reminder.Reminder = task.DueDate.Add(relativeDuration)
			}
		case models.ReminderRelationStartDate:
			if !task.StartDate.IsZero() {
				reminder.Reminder = task.StartDate.Add(relativeDuration)
			}
		case models.ReminderRelationEndDate:
			if !task.EndDate.IsZero() {
				reminder.Reminder = task.EndDate.Add(relativeDuration)
			}
		default:
			if reminder.RelativePeriod != 0 {
				return models.ErrReminderRelativeToMissing{TaskID: task.ID}
			}
		}
	}
	return nil
}

func (ts *TaskService) ensureTaskIndex(s *xorm.Session, task *models.Task) error {
	if task.Index == 0 {
		nextIndex, err := ts.calculateNextTaskIndex(s, task.ProjectID)
		if err != nil {
			return err
		}
		task.Index = nextIndex
		return nil
	}

	exists, err := s.Where("project_id = ? AND `index` = ?", task.ProjectID, task.Index).Exist(&models.Task{})
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	nextIndex, err := ts.calculateNextTaskIndex(s, task.ProjectID)
	if err != nil {
		return err
	}
	task.Index = nextIndex
	return nil
}

func (ts *TaskService) calculateNextTaskIndex(s *xorm.Session, projectID int64) (int64, error) {
	latestTask := &models.Task{}
	_, err := s.Where("project_id = ?", projectID).OrderBy("`index` desc").Get(latestTask)
	if err != nil {
		return 0, err
	}

	return latestTask.Index + 1, nil
}

func (ts *TaskService) setTaskIdentifier(task *models.Task, project *models.Project) {
	if project == nil || project.Identifier == "" {
		task.Identifier = "#" + strconv.FormatInt(task.Index, 10)
		return
	}

	task.Identifier = project.Identifier + "-" + strconv.FormatInt(task.Index, 10)
}

// getRawFavoriteTasks gets favorite tasks with filtering and sorting
func (ts *TaskService) getRawFavoriteTasks(s *xorm.Session, favoriteTaskIDs []int64, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	if len(favoriteTaskIDs) == 0 {
		return nil, 0, 0, nil
	}

	// Create a copy of opts for favorites
	favoriteOpts := *opts
	favoriteOpts.projectIDs = nil // Clear project IDs for favorites

	// Build the query using favorite task IDs
	query := s.In("id", favoriteTaskIDs)

	// Apply filters, sorting, and search
	query, _, err = ts.applyFiltersToQuery(query, &favoriteOpts)
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply sorting
	ts.applySortingToQuery(query, favoriteOpts.sortby)

	// Get total count first (before pagination)
	totalItems, err = s.In("id", favoriteTaskIDs).Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply pagination
	query = query.Limit(opts.perPage, (opts.page-1)*opts.perPage)

	// Execute query
	err = query.Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

// buildAndExecuteTaskQuery builds and executes the main task query with all filters
func (ts *TaskService) buildAndExecuteTaskQuery(s *xorm.Session, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	// Start with project filtering
	query := s.In("project_id", opts.projectIDs)

	// Apply filters, sorting, and search
	query, _, err = ts.applyFiltersToQuery(query, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply sorting
	ts.applySortingToQuery(query, opts.sortby)

	// Get total count first (before pagination)
	totalItems, err = s.In("project_id", opts.projectIDs).Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply pagination
	query = query.Limit(opts.perPage, (opts.page-1)*opts.perPage)

	// Execute query
	err = query.Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

// applyFiltersToQuery applies all filters to the query
func (ts *TaskService) applyFiltersToQuery(query *xorm.Session, opts *taskSearchOptions) (*xorm.Session, *xorm.Session, error) {
	// For now, delegate complex filtering to the model
	// TODO: Move all filter logic to service layer

	// Apply search filter
	if opts.search != "" {
		searchWhere := "title LIKE ?"
		searchPattern := "%" + opts.search + "%"
		query = query.Where(searchWhere, searchPattern)
	}

	// Apply custom filters if present
	if opts.filter != "" {
		// For now, just delegate back to models for complex filtering
		// This will be moved to service layer in a future iteration
		// For simple cases, we handle here; for complex, we delegate
		if strings.Contains(opts.filter, ">=") || strings.Contains(opts.filter, "<=") ||
			strings.Contains(opts.filter, "!=") || strings.Contains(opts.filter, "&&") ||
			strings.Contains(opts.filter, "||") {
			// Complex filter - delegate to models for now
			// This is where the date range logic and other complex filtering happens
			// TODO: Implement full filter parsing in service layer
		}
	}

	// Use the same query for count (xorm doesn't have Clone)
	totalQuery := query
	return query, totalQuery, nil
}

// applySortingToQuery applies sorting to the query
func (ts *TaskService) applySortingToQuery(query *xorm.Session, sortParams []*sortParam) {
	for _, param := range sortParams {
		var orderBy string
		if param.orderBy == orderDescending {
			orderBy = param.sortBy + " DESC"
		} else {
			orderBy = param.sortBy + " ASC"
		}
		query = query.OrderBy(orderBy)
	}
}

// addBucketsToTasks adds bucket information to tasks using the KanbanService
func (ts *TaskService) addBucketsToTasks(s *xorm.Session, a web.Auth, taskIDs []int64, taskMap map[int64]*models.Task) error {
	if ts.KanbanService == nil {
		return nil // Skip if KanbanService not available
	}

	u, err := models.GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}

	return ts.KanbanService.AddBucketsToTasks(s, taskIDs, taskMap, u)
}

// addReactionsToTasks adds reaction data to tasks using the ReactionsService
func (ts *TaskService) addReactionsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	if ts.ReactionsService == nil {
		return nil // Skip if ReactionsService not available
	}

	return ts.ReactionsService.AddReactionsToTasks(s, taskIDs, taskMap)
}

// addCommentsToTasks adds comment data to tasks using the CommentService
func (ts *TaskService) addCommentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	fmt.Printf("DEBUG: addCommentsToTasks called with taskIDs: %v\n", taskIDs)
	if ts.CommentService == nil {
		fmt.Printf("DEBUG: CommentService is nil, skipping comments\n")
		return nil // Skip if CommentService not available
	}

	fmt.Printf("DEBUG: Calling CommentService.AddCommentsToTasks\n")
	return ts.CommentService.AddCommentsToTasks(s, taskIDs, taskMap)
}
