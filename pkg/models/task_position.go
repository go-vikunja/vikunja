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
func updateTaskPosition(s *xorm.Session, a web.Auth, tp *TaskPosition) (err error) {
	exists, err := s.
		Where("task_id = ? AND project_view_id = ?", tp.TaskID, tp.ProjectViewID).
		Exist(&TaskPosition{})
	if err != nil {
		return err
	}

	if !exists {
		_, err = s.Insert(tp)
		if err != nil {
			return
		}
	} else {
		_, err = s.
			Where("task_id = ? AND project_view_id = ?", tp.TaskID, tp.ProjectViewID).
			Cols("project_view_id", "position").
			Update(tp)
		if err != nil {
			return
		}
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

	projects, err := getRelevantProjectsFromCollection(s, a, tc)
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

	// We're directly using the db here, even if Typesense is configured, because in some edge cases Typesense
	// does not know about all tasks. These tasks then won't have their position recalculated, which means they will
	// seemingly jump around after reloading their project.
	// The real fix here is of course to make sure all tasks are indexed in Typesense, but until that's fixed,
	// this solves the issue of task positions not being saved.
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

	_, err = s.
		Where("project_view_id = ?", view.ID).
		Delete(&TaskPosition{})
	if err != nil {
		return
	}

	count, err := s.Insert(newPositions)
	if err != nil {
		return
	}

	log.Debugf("Inserted %d new positions for %d total tasks in view %d", count, len(allTasks), view.ID)

	return events.Dispatch(&TaskPositionsRecalculatedEvent{
		NewTaskPositions: newPositions,
	})
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
	log.Debugf("Recalculating task positions for view %d (repair mode)", view.ID)

	// Get all existing positions for this view, ordered by current position then task ID
	var existingPositions []*TaskPosition
	err := s.Where("project_view_id = ?", view.ID).
		OrderBy("position ASC, task_id ASC").
		Find(&existingPositions)
	if err != nil {
		return err
	}

	if len(existingPositions) == 0 {
		return nil
	}

	// Delete all existing positions
	_, err = s.Where("project_view_id = ?", view.ID).Delete(&TaskPosition{})
	if err != nil {
		return err
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

	count, err := s.Insert(newPositions)
	if err != nil {
		return err
	}

	log.Debugf("Repair: inserted %d new positions for view %d", count, view.ID)

	return events.Dispatch(&TaskPositionsRecalculatedEvent{
		NewTaskPositions: newPositions,
	})
}

func calculateNewPositionForTask(s *xorm.Session, a web.Auth, t *Task, view *ProjectView) (*TaskPosition, error) {
	if t.Position == 0 {
		lowestPosition := &TaskPosition{}
		exists, err := s.Where("project_view_id = ?", view.ID).
			OrderBy("position asc").
			Get(lowestPosition)
		if err != nil {
			return nil, err
		}
		if exists {
			if lowestPosition.Position == 0 {
				err = RecalculateTaskPositions(s, view, a)
				if err != nil {
					return nil, err
				}

				lowestPosition = &TaskPosition{}
				_, err = s.Where("project_view_id = ?", view.ID).
					OrderBy("position asc").
					Get(lowestPosition)
				if err != nil {
					return nil, err
				}
			}

			t.Position = lowestPosition.Position / 2
		}
	}

	return &TaskPosition{
		TaskID:        t.ID,
		ProjectViewID: view.ID,
		Position:      calculateDefaultPosition(t.Index, t.Position),
	}, nil
}

func DeleteOrphanedTaskPositions(s *xorm.Session) (count int64, err error) {
	return s.
		Where("task_id not in (select id from tasks) OR project_view_id not in (select id from project_views)").
		Delete(&TaskPosition{})
}

// createPositionsForTasksInView creates position records for tasks that don't have them.
// Used as a safety net during task fetching for saved filter views.
func createPositionsForTasksInView(s *xorm.Session, tasks []*Task, view *ProjectView, a web.Auth) error {
	if len(tasks) == 0 {
		return nil
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

	newPositions := make([]*TaskPosition, 0, len(tasks))
	for i, task := range tasks {
		newPositions = append(newPositions, &TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			Position:      spacing * float64(i+1),
		})
	}

	_, err = s.Insert(&newPositions)
	return err
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

	// Group positions by view ID
	positionsByView := make(map[int64][]*TaskPosition)
	viewIDs := []int64{}
	for _, pos := range allPositions {
		positionsByView[pos.ProjectViewID] = append(positionsByView[pos.ProjectViewID], pos)
		viewIDs = append(viewIDs, pos.ProjectViewID)
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
