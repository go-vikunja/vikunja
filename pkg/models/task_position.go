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
	"fmt"
	"math"
	"sort"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"
)

// MinPositionSpacing is the smallest gap we allow between positions.
const MinPositionSpacing = 0.01

// getRootProjectID returns the root project ID for a given project by traversing
// up the parent chain iteratively.
func getRootProjectID(s *xorm.Session, projectID int64) (int64, error) {
	currentID := projectID
	for {
		project := &Project{}
		exists, err := s.ID(currentID).Cols("parent_project_id").Get(project)
		if err != nil {
			return 0, err
		}
		if !exists || project.ParentProjectID == 0 {
			return currentID, nil
		}
		currentID = project.ParentProjectID
	}
}

// getRootProjectViewID returns the corresponding view ID in the root project
// that matches the ViewKind of the given view. This is used to store positions
// at the root level for hierarchical task display.
//
// IMPORTANT: All task position storage must use root view IDs to ensure consistent
// ordering across project hierarchies. Functions that store positions should either:
// - Call getRootProjectViewID directly, or
// - Use resolveToRootView to get the full view object
func getRootProjectViewID(s *xorm.Session, viewID int64) (int64, error) {
	// Get the current view
	view, err := GetProjectViewByID(s, viewID)
	if err != nil {
		return 0, err
	}

	// Get the root project ID
	rootProjectID, err := getRootProjectID(s, view.ProjectID)
	if err != nil {
		return 0, err
	}

	// If already at root, return the current view ID
	if rootProjectID == view.ProjectID {
		return viewID, nil
	}

	// Find a matching view (same ViewKind) in the root project
	rootView := &ProjectView{}
	exists, err := s.Where("project_id = ? AND view_kind = ?", rootProjectID, view.ViewKind).Get(rootView)
	if err != nil {
		return 0, err
	}

	// If no matching view exists in root project, fall back to current view
	if !exists {
		return viewID, nil
	}

	return rootView.ID, nil
}

// resolveToRootView returns the corresponding view in the root project that matches
// the ViewKind of the given view. If the view is already in a root project, it returns
// the same view. This should be called at the start of any function that stores positions.
func resolveToRootView(s *xorm.Session, view *ProjectView) (*ProjectView, error) {
	rootViewID, err := getRootProjectViewID(s, view.ID)
	if err != nil {
		return nil, err
	}
	if rootViewID == view.ID {
		return view, nil
	}
	return GetProjectViewByID(s, rootViewID)
}

// StoreTaskPosition is the single entry point for storing a task position.
// It resolves the viewID to the root project view and upserts the position.
// All code that needs to store positions MUST use this function.
func StoreTaskPosition(s *xorm.Session, taskID, viewID int64, position float64) error {
	rootViewID, err := getRootProjectViewID(s, viewID)
	if err != nil {
		return err
	}

	exists, err := s.
		Where("task_id = ? AND project_view_id = ?", taskID, rootViewID).
		Exist(&TaskPosition{})
	if err != nil {
		return err
	}

	tp := &TaskPosition{
		TaskID:        taskID,
		ProjectViewID: rootViewID,
		Position:      position,
	}

	if !exists {
		_, err = s.Insert(tp)
	} else {
		_, err = s.
			Where("task_id = ? AND project_view_id = ?", taskID, rootViewID).
			Cols("position").
			Update(tp)
	}
	return err
}

// ReplaceAllPositionsForView removes all positions for a view and bulk inserts new ones.
// It resolves the view to the root project view first. Used for recalculation operations.
// All code that needs to replace positions in bulk MUST use this function.
func ReplaceAllPositionsForView(s *xorm.Session, view *ProjectView, newPositions []*TaskPosition) error {
	rootView, err := resolveToRootView(s, view)
	if err != nil {
		return err
	}

	// Delete all existing positions for this view
	_, err = s.Where("project_view_id = ?", rootView.ID).Delete(&TaskPosition{})
	if err != nil {
		return err
	}

	if len(newPositions) == 0 {
		return nil
	}

	// Update all positions to use root view ID
	for _, p := range newPositions {
		p.ProjectViewID = rootView.ID
	}

	_, err = s.Insert(newPositions)
	return err
}

type TaskPosition struct {
	// The ID of the task this position is for
	TaskID int64 `xorm:"bigint not null index" json:"task_id" param:"task"`
	// The project view this task is related to
	ProjectViewID int64 `xorm:"bigint not null index" json:"project_view_id"`
	// The position of the task - any task project can be sorted as usual by this parameter.
	// When accessing tasks via kanban buckets, this is primarily used to sort them based on a range
	// We're using a float64 here to make it possible to put any task within any two other tasks (by changing the number).
	// You would calculate the new position between two tasks with something like task3.position = (task2.position - task1.position) / 2.
	// A 64-Bit float leaves plenty of room to initially give tasks a position with 2^16 difference to the previous task
	// which also leaves a lot of room for rearranging and sorting later.
	// Positions are always saved per view. They will automatically be set if you request the tasks through a view
	// endpoint, otherwise they will always be 0. To update them, take a look at the Task Position endpoint.
	Position float64 `xorm:"double not null" json:"position"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (tp *TaskPosition) TableName() string {
	return "task_positions"
}

func (tp *TaskPosition) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	t := &Task{ID: tp.TaskID}
	return t.CanUpdate(s, a)
}

func (tp *TaskPosition) refresh(s *xorm.Session) (err error) {
	updatedPosition := &TaskPosition{}
	_, err = s.Where("task_id = ? AND project_view_id = ?", tp.TaskID, tp.ProjectViewID).Get(updatedPosition)
	if err != nil {
		return err
	}

	tp.Position = updatedPosition.Position
	return nil
}

// updateTaskPosition is the internal function that performs the task position update logic
// without dispatching events. This is used by moveTaskToDoneBuckets to avoid duplicate events.
// It delegates to StoreTaskPosition for the actual storage, then handles position spacing
// and conflict resolution.
func updateTaskPosition(s *xorm.Session, a web.Auth, tp *TaskPosition) (err error) {
	// Resolve to root view - we need to know it for the post-storage checks
	rootViewID, err := getRootProjectViewID(s, tp.ProjectViewID)
	if err != nil {
		return err
	}
	tp.ProjectViewID = rootViewID

	// Use the central storage function
	err = StoreTaskPosition(s, tp.TaskID, tp.ProjectViewID, tp.Position)
	if err != nil {
		return err
	}

	if tp.Position < MinPositionSpacing {
		view, err := GetProjectViewByID(s, tp.ProjectViewID)
		if err != nil {
			return err
		}
		err = RecalculateTaskPositions(s, view, a)
		if err != nil {
			return err
		}

		return tp.refresh(s)
	}

	// Check for and resolve position conflicts
	conflicts, err := findPositionConflicts(s, tp.ProjectViewID, tp.Position)
	if err != nil {
		return err
	}

	if len(conflicts) > 1 {
		err = resolveTaskPositionConflicts(s, tp.ProjectViewID, conflicts)
		if err != nil {
			return err
		}

		return tp.refresh(s)
	}

	return nil
}

// Update is the handler to update a task position
// @Summary Updates a task position
// @Description Updates a task position.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Param view body models.TaskPosition true "The task position with updated values you want to change."
// @Success 200 {object} models.TaskPosition "The updated task position."
// @Failure 400 {object} web.HTTPError "Invalid task position object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id}/position [post]
func (tp *TaskPosition) Update(s *xorm.Session, a web.Auth) (err error) {
	err = updateTaskPosition(s, a, tp)
	if err != nil {
		return err
	}

	return triggerTaskUpdatedEventForTaskID(s, a, tp.TaskID)
}

func RecalculateTaskPositions(s *xorm.Session, view *ProjectView, a web.Auth) (err error) {
	// Resolve to root view for hierarchical position storage
	view, err = resolveToRootView(s, view)
	if err != nil {
		return err
	}

	log.Debugf("Recalculating task positions for view %d", view.ID)

	opts := &taskSearchOptions{
		projectViewID: view.ID,
		sortby: []*sortParam{
			{
				projectViewID: view.ID,
				sortBy:        taskPropertyPosition,
				orderBy:       orderAscending,
			},
			{
				sortBy:  taskPropertyID,
				orderBy: orderAscending,
			},
		},
	}

	// Using the collection so that we get all tasks, even in cases where we're dealing with a saved filter underneath
	tc := &TaskCollection{
		ProjectID: view.ProjectID,
	}
	if view.ProjectID < -1 {
		tc.ProjectID = 0

		sf, err := GetSavedFilterSimpleByID(s, GetSavedFilterIDFromProjectID(view.ProjectID))
		if err != nil {
			return err
		}

		opts.filterIncludeNulls = sf.Filters.FilterIncludeNulls
		opts.filterTimezone = sf.Filters.FilterTimezone
		opts.filter = sf.Filters.Filter
		opts.parsedFilters, err = getTaskFiltersFromFilterString(opts.filter, opts.filterTimezone)
		if err != nil {
			return err
		}
	}

	projects, err := getRelevantProjectsFromCollection(s, a, tc, view)
	if err != nil {
		return err
	}

	for _, p := range projects {
		opts.projectIDs = append(opts.projectIDs, p.ID)
	}

	dbSearcher := &dbTaskSearcher{
		s: s,
		a: a,
	}

	allTasks, _, err := dbSearcher.Search(opts)
	if err != nil {
		return
	}
	if len(allTasks) == 0 {
		return
	}

	maxPosition := math.Pow(2, 32)
	newPositions := make([]*TaskPosition, 0, len(allTasks))

	for i, task := range allTasks {

		currentPosition := maxPosition / float64(len(allTasks)) * (float64(i + 1))

		newPositions = append(newPositions, &TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			Position:      currentPosition,
		})
	}

	// Use the central function for bulk position replacement
	err = ReplaceAllPositionsForView(s, view, newPositions)
	if err != nil {
		return err
	}

	log.Debugf("Inserted %d new positions for %d total tasks in view %d", len(newPositions), len(allTasks), view.ID)

	events.DispatchOnCommit(s, &TaskPositionsRecalculatedEvent{
		NewTaskPositions: newPositions,
	})
	return nil
}

func getPositionsForView(s *xorm.Session, view *ProjectView) (positions []*TaskPosition, err error) {
	positions = []*TaskPosition{}
	err = s.
		Where("project_view_id = ?", view.ID).
		Find(&positions)
	return
}

// recalculateTaskPositionsForRepair recalculates positions for all tasks in a view
// without requiring auth. Used by CLI repair when localized repair fails.
// Unlike RecalculateTaskPositions, this only operates on tasks that already have
// positions in the view (doesn't discover new tasks).
func recalculateTaskPositionsForRepair(s *xorm.Session, view *ProjectView) error {
	// Resolve to root view for hierarchical position storage
	var err error
	view, err = resolveToRootView(s, view)
	if err != nil {
		return err
	}

	log.Debugf("Recalculating task positions for view %d (repair mode)", view.ID)

	// Get all existing positions for this view, ordered by current position then task ID
	var existingPositions []*TaskPosition
	err = s.Where("project_view_id = ?", view.ID).
		OrderBy("position ASC, task_id ASC").
		Find(&existingPositions)
	if err != nil {
		return err
	}

	if len(existingPositions) == 0 {
		return nil
	}

	// Reassign evenly spaced positions
	maxPosition := math.Pow(2, 32)
	newPositions := make([]*TaskPosition, 0, len(existingPositions))

	for i, pos := range existingPositions {
		currentPosition := maxPosition / float64(len(existingPositions)) * float64(i+1)
		newPositions = append(newPositions, &TaskPosition{
			TaskID:        pos.TaskID,
			ProjectViewID: view.ID,
			Position:      currentPosition,
		})
	}

	// Use the central function for bulk position replacement
	err = ReplaceAllPositionsForView(s, view, newPositions)
	if err != nil {
		return err
	}

	log.Debugf("Repair: inserted %d new positions for view %d", len(newPositions), view.ID)

	events.DispatchOnCommit(s, &TaskPositionsRecalculatedEvent{
		NewTaskPositions: newPositions,
	})
	return nil
}

func calculateNewPositionForTask(s *xorm.Session, a web.Auth, t *Task, view *ProjectView) (*TaskPosition, error) {
	// Resolve to root view for hierarchical position storage
	rootViewID, err := getRootProjectViewID(s, view.ID)
	if err != nil {
		return nil, err
	}

	position := t.Position
	if position == 0 {
		lowestPosition := &TaskPosition{}
		exists, err := s.Where("project_view_id = ?", rootViewID).
			OrderBy("position asc").
			Get(lowestPosition)
		if err != nil {
			return nil, err
		}
		if exists {
			if lowestPosition.Position < MinPositionSpacing {
				rootView, err := GetProjectViewByID(s, rootViewID)
				if err != nil {
					return nil, err
				}
				err = RecalculateTaskPositions(s, rootView, a)
				if err != nil {
					return nil, err
				}

				lowestPosition = &TaskPosition{}
				_, err = s.Where("project_view_id = ?", rootViewID).
					OrderBy("position asc").
					Get(lowestPosition)
				if err != nil {
					return nil, err
				}
			}

			position = lowestPosition.Position / 2
		}
	}

	return &TaskPosition{
		TaskID:        t.ID,
		ProjectViewID: rootViewID,
		Position:      calculateDefaultPosition(t.Index, position),
	}, nil
}

// DeleteOrphanedTaskPositions removes task position records that reference
// tasks or project views that no longer exist.
// If dryRun is true, it counts the orphaned records without deleting them.
func DeleteOrphanedTaskPositions(s *xorm.Session, dryRun bool) (count int64, err error) {
	whereClause := "task_id not in (select id from tasks) OR project_view_id not in (select id from project_views)"

	if dryRun {
		return s.Where(whereClause).Count(&TaskPosition{})
	}

	return s.Where(whereClause).Delete(&TaskPosition{})
}

// createPositionsForTasksInView creates position records for tasks that don't have them.
// Used as a safety net during task fetching for saved filter views.
func createPositionsForTasksInView(s *xorm.Session, tasks []*Task, view *ProjectView, a web.Auth) error {
	if len(tasks) == 0 {
		return nil
	}

	// Resolve to root view for hierarchical position storage
	var err error
	view, err = resolveToRootView(s, view)
	if err != nil {
		return err
	}

	// Get the current lowest position to place new tasks at the top
	lowestPosition := &TaskPosition{}
	has, err := s.
		Where("project_view_id = ?", view.ID).
		OrderBy("position asc").
		Get(lowestPosition)
	if err != nil {
		return err
	}

	var basePosition float64
	if !has || lowestPosition.Position < MinPositionSpacing {
		return RecalculateTaskPositions(s, view, a)
	}

	// Place new tasks before the lowest position, evenly spaced
	basePosition = lowestPosition.Position
	spacing := basePosition / float64(len(tasks)+1)

	// Use the central function to store each position
	for i, task := range tasks {
		position := spacing * float64(i+1)
		err = StoreTaskPosition(s, task.ID, view.ID, position)
		if err != nil {
			return err
		}
	}

	return nil
}

// findPositionConflicts returns all task positions that share the same position value
// within a given project view. Returns an empty slice if no conflicts exist.
func findPositionConflicts(s *xorm.Session, projectViewID int64, position float64) (conflicts []*TaskPosition, err error) {
	conflicts = []*TaskPosition{}
	err = s.
		Where("project_view_id = ? AND position = ?", projectViewID, position).
		Find(&conflicts)
	if err != nil {
		return nil, err
	}
	return conflicts, nil
}

// RepairResult contains the summary of a repair operation.
type RepairResult struct {
	ViewsScanned    int
	ViewsRepaired   int
	TasksAffected   int
	FullRecalcViews int      // Views that needed full recalculation
	Errors          []string // Views that couldn't be repaired
}

// RepairTaskPositions scans all project views for duplicate task positions
// and repairs them using localized conflict resolution or full recalculation.
// If dryRun is true, it reports what would be fixed without making changes.
func RepairTaskPositions(s *xorm.Session, dryRun bool) (*RepairResult, error) {
	result := &RepairResult{}

	// Get all task positions in a single query
	var allPositions []*TaskPosition
	err := s.OrderBy("project_view_id ASC, position ASC").Find(&allPositions)
	if err != nil {
		return nil, err
	}

	if len(allPositions) == 0 {
		return result, nil
	}

	// Group positions by view ID
	positionsByView := make(map[int64][]*TaskPosition)
	for _, pos := range allPositions {
		positionsByView[pos.ProjectViewID] = append(positionsByView[pos.ProjectViewID], pos)
	}

	viewIDs := []int64{}
	for viewID := range positionsByView {
		viewIDs = append(viewIDs, viewID)
	}

	viewsByID := make(map[int64]*ProjectView)
	err = s.In("id", viewIDs).Find(&viewsByID)
	if err != nil {
		return nil, err
	}

	// Process each view
	for viewID, positions := range positionsByView {
		result.ViewsScanned++

		// Find duplicate positions within this view's positions
		duplicates := findDuplicatesInPositions(positions)
		if len(duplicates) == 0 {
			continue
		}

		if dryRun {
			// Count affected tasks without making changes
			for _, dup := range duplicates {
				result.TasksAffected += len(dup)
			}
			result.ViewsRepaired++
			log.Infof("[dry-run] Would repair %d position conflicts in view %d", len(duplicates), viewID)
			continue
		}

		view, has := viewsByID[viewID]
		if !has {
			continue
		}

		viewRepaired := false
		for _, conflicts := range duplicates {
			result.TasksAffected += len(conflicts)

			err = resolveTaskPositionConflicts(s, viewID, conflicts)
			if IsErrNeedsFullRecalculation(err) {
				// Fall back to full recalculation for this view
				err = recalculateTaskPositionsForRepair(s, view)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("view %d: recalculation failed: %v", viewID, err))
					continue
				}
				result.FullRecalcViews++
				viewRepaired = true
				// After full recalculation, no need to process more duplicates in this view
				break
			} else if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("view %d: %v", viewID, err))
				continue
			}
			viewRepaired = true
		}

		if viewRepaired {
			result.ViewsRepaired++
		}
	}

	return result, nil
}

// findDuplicatesInPositions finds groups of positions that share the same position value.
func findDuplicatesInPositions(positions []*TaskPosition) [][]*TaskPosition {
	// Group by position
	positionGroups := make(map[float64][]*TaskPosition)
	for _, tp := range positions {
		positionGroups[tp.Position] = append(positionGroups[tp.Position], tp)
	}

	// Filter to only duplicates (groups with more than 1 task)
	var duplicates [][]*TaskPosition
	for _, group := range positionGroups {
		if len(group) > 1 {
			duplicates = append(duplicates, group)
		}
	}

	return duplicates
}

// resolveTaskPositionConflicts redistributes conflicting task positions within the
// available gap between their neighbors. Returns ErrNeedsFullRecalculation if there
// is insufficient spacing to assign unique positions.
func resolveTaskPositionConflicts(s *xorm.Session, projectViewID int64, conflicts []*TaskPosition) error {
	if len(conflicts) <= 1 {
		return nil // No conflict to resolve
	}

	conflictPosition := conflicts[0].Position

	// Find the nearest distinct neighbor positions
	var leftNeighbor, rightNeighbor *TaskPosition
	var lowerBound, upperBound float64

	// Get the position immediately before the conflict position
	leftNeighbor = &TaskPosition{}
	hasLeft, err := s.
		Where("project_view_id = ? AND position < ?", projectViewID, conflictPosition).
		OrderBy("position DESC").
		Get(leftNeighbor)
	if err != nil {
		return err
	}
	if hasLeft {
		lowerBound = leftNeighbor.Position
	}

	// Get the position immediately after the conflict position
	rightNeighbor = &TaskPosition{}
	hasRight, err := s.
		Where("project_view_id = ? AND position > ?", projectViewID, conflictPosition).
		OrderBy("position ASC").
		Get(rightNeighbor)
	if err != nil {
		return err
	}
	if hasRight {
		upperBound = rightNeighbor.Position
	} else {
		upperBound = math.Pow(2, 32)
	}

	// Calculate spacing needed
	availableGap := upperBound - lowerBound
	spacing := availableGap / float64(len(conflicts)+1)

	// Check if we have enough spacing
	if spacing < MinPositionSpacing {
		return &ErrNeedsFullRecalculation{ProjectViewID: projectViewID}
	}

	// Sort conflicts by task ID for deterministic ordering
	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].TaskID < conflicts[j].TaskID
	})

	// Assign new positions
	for i, tp := range conflicts {
		newPosition := lowerBound + spacing*float64(i+1)
		_, err = s.
			Where("task_id = ? AND project_view_id = ?", tp.TaskID, projectViewID).
			Cols("position").
			Update(&TaskPosition{Position: newPosition})
		if err != nil {
			return err
		}
	}

	log.Debugf("Repaired position conflict in view %d: %d tasks respaced from position %.6f", projectViewID, len(conflicts), conflictPosition)

	return nil
}

// resolvePositionConflictsAfterInsert checks a batch of newly inserted task positions
// for conflicts (duplicate position values within the same view) and resolves them.
// This is called after bulk-inserting positions during task creation.
// If resolveTaskPositionConflicts returns ErrNeedsFullRecalculation for a view,
// it falls back to a full recalculation of all positions in that view.
func resolvePositionConflictsAfterInsert(s *xorm.Session, positions []*TaskPosition) error {
	// Track which (viewID, position) pairs we've already checked to avoid
	// resolving the same conflict group twice.
	checked := make(map[int64]map[float64]bool)
	// Track views that have already been fully recalculated so we skip
	// further conflict checks for them.
	recalculated := make(map[int64]bool)

	for _, pos := range positions {
		if recalculated[pos.ProjectViewID] {
			continue
		}
		if checked[pos.ProjectViewID] != nil && checked[pos.ProjectViewID][pos.Position] {
			continue
		}
		if checked[pos.ProjectViewID] == nil {
			checked[pos.ProjectViewID] = make(map[float64]bool)
		}
		checked[pos.ProjectViewID][pos.Position] = true

		conflicts, err := findPositionConflicts(s, pos.ProjectViewID, pos.Position)
		if err != nil {
			return err
		}

		if len(conflicts) <= 1 {
			continue
		}

		err = resolveTaskPositionConflicts(s, pos.ProjectViewID, conflicts)
		if IsErrNeedsFullRecalculation(err) {
			view := &ProjectView{ID: pos.ProjectViewID}
			err = recalculateTaskPositionsForRepair(s, view)
			if err != nil {
				return err
			}
			recalculated[pos.ProjectViewID] = true
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}
