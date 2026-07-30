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
	"cmp"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"
)

// MinPositionSpacing is the smallest gap we allow between positions.
const MinPositionSpacing = 0.01

type TaskPosition struct {
	// The ID of the task this position is for
	TaskID int64 `xorm:"bigint not null index unique(task_view)" json:"task_id" param:"task" readOnly:"true" doc:"The numeric id of the task this position belongs to. Taken from the URL; ignored in the request body."`
	// The project view this task is related to
	ProjectViewID int64 `xorm:"bigint not null index unique(task_view) index(view_position)" json:"project_view_id" doc:"The id of the project view this position applies to. Positions are stored per view, so the same task has an independent position in each of its project's views."`
	// The position of the task - any task project can be sorted as usual by this parameter.
	// When accessing tasks via kanban buckets, this is primarily used to sort them based on a range
	// We're using a float64 here to make it possible to put any task within any two other tasks (by changing the number).
	// You would calculate the new position between two tasks with something like task3.position = (task2.position - task1.position) / 2.
	// A 64-Bit float leaves plenty of room to initially give tasks a position with 2^16 difference to the previous task
	// which also leaves a lot of room for rearranging and sorting later.
	// Positions are always saved per view. They will automatically be set if you request the tasks through a view
	// endpoint, otherwise they will always be 0. To update them, take a look at the Task Position endpoint.
	Position float64 `xorm:"double not null index(view_position)" json:"position" doc:"The task's sort position within the view, as a float so a task can be placed between any two others. To drop a task between two neighbours, set this to their midpoint. Values below the minimum spacing trigger a server-side recalculation of all positions in the view, so the stored value may differ from what you sent."`

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

// upsertTaskPosition atomically inserts or updates the position row for the
// given task and view. A separate exists-check followed by an insert races
// with concurrent requests doing the same and violates the unique index on
// (task_id, project_view_id).
func upsertTaskPosition(s *xorm.Session, tp *TaskPosition) (err error) {
	onConflict := "ON CONFLICT (task_id, project_view_id) DO UPDATE SET position = excluded.position"
	if db.Type() == schemas.MYSQL {
		onConflict = "ON DUPLICATE KEY UPDATE position = VALUES(position)"
	}

	// Raw SQL bypasses xorm's bean-based table-name handling, so qualify the
	// table ourselves to honor a configured postgres schema (database.schema).
	table := s.Engine().TableName(tp, true)
	_, err = s.Exec(
		"INSERT INTO "+table+" (task_id, project_view_id, position) VALUES (?, ?, ?) "+onConflict,
		tp.TaskID, tp.ProjectViewID, tp.Position)
	return
}

// bulkInsertTaskPositions inserts position rows in batches, tolerating rows
// whose (task_id, project_view_id) pair already exists because a concurrent
// transaction created them in the meantime. The overwrite flag decides who
// wins such a conflict:
//
// Callers which only add rows for tasks that had none (saved filter healing,
// the task-created listener) pass false: if another transaction beat them to
// it, the existing row is the correct state — e.g. a position the user set
// via drag & drop in between — and the caller's value is stale, so the row is
// kept as is.
//
// The recalculation paths pass true: they rewrite every position in the view,
// so their values are authoritative. They hold the view lock, but writers of
// the first kind deliberately don't take it and can slip a row in between the
// recalculation's delete and reinsert (invisible to the delete's snapshot
// under READ COMMITTED). A plain insert would then fail on the unique index;
// overwriting resolves the race with the recalculated value instead.
func bulkInsertTaskPositions(s *xorm.Session, positions []*TaskPosition, overwrite bool) (err error) {
	// Keep statements well below the parameter limits of all supported databases.
	const batchSize = 100

	var onConflict string
	switch {
	case db.Type() == schemas.MYSQL && overwrite:
		onConflict = "ON DUPLICATE KEY UPDATE position = VALUES(position)"
	case db.Type() == schemas.MYSQL:
		// The self-assignment is a no-op update which only ignores duplicate
		// keys, unlike INSERT IGNORE which would swallow unrelated errors too.
		onConflict = "ON DUPLICATE KEY UPDATE position = position"
	case overwrite:
		onConflict = "ON CONFLICT (task_id, project_view_id) DO UPDATE SET position = excluded.position"
	default:
		onConflict = "ON CONFLICT (task_id, project_view_id) DO NOTHING"
	}

	// Raw SQL bypasses xorm's bean-based table-name handling, so qualify the
	// table ourselves to honor a configured postgres schema (database.schema).
	table := s.Engine().TableName(&TaskPosition{}, true)

	for start := 0; start < len(positions); start += batchSize {
		batch := positions[start:min(start+batchSize, len(positions))]

		placeholders := make([]string, 0, len(batch))
		args := make([]interface{}, 1, len(batch)*3+1)
		for _, p := range batch {
			placeholders = append(placeholders, "(?, ?, ?)")
			args = append(args, p.TaskID, p.ProjectViewID, p.Position)
		}

		args[0] = "INSERT INTO " + table + " (task_id, project_view_id, position) VALUES " + strings.Join(placeholders, ", ") + " " + onConflict

		_, err = s.Exec(args...)
		if err != nil {
			return err
		}
	}

	return nil
}

// lockPositionsForViewUpdate takes a row lock on the view, serializing
// concurrent transactions which rewrite the positions of the same view.
// Without it, two transactions can both delete the old position rows and then
// insert overlapping new ones, violating the unique index on
// (task_id, project_view_id). SQLite allows only a single writer at a time and
// does not support FOR UPDATE, so no explicit lock is needed there.
func lockPositionsForViewUpdate(s *xorm.Session, viewID int64) (err error) {
	if db.Type() == schemas.SQLITE {
		return nil
	}

	table := s.Engine().TableName(&ProjectView{}, true)
	_, err = s.Exec("SELECT id FROM "+table+" WHERE id = ? FOR UPDATE", viewID)
	return
}

// updateTaskPosition is the internal function that performs the task position update logic
// without dispatching events. This is used by moveTaskToDoneBuckets to avoid duplicate events.
func updateTaskPosition(s *xorm.Session, a web.Auth, tp *TaskPosition) (err error) {
	if tp.Position < MinPositionSpacing {
		// The recalculation below will need the view lock anyway. Taking it
		// before the upsert keeps the lock order consistent with concurrent
		// recalculations (view lock first, then position rows), avoiding a
		// lock-order deadlock.
		err = lockPositionsForViewUpdate(s, tp.ProjectViewID)
		if err != nil {
			return err
		}
	}

	err = upsertTaskPosition(s, tp)
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

	log.Debugf("Recalculating task positions for view %d", view.ID)

	err = lockPositionsForViewUpdate(s, view.ID)
	if err != nil {
		return err
	}

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

	allTasks, _, err := dbSearcher.Search(opts)
	if err != nil {
		return
	}
	if len(allTasks) == 0 {
		return
	}

	maxPosition := math.Pow(2, 32)
	newPositions := make([]*TaskPosition, 0, len(allTasks))
	seenTasks := make(map[int64]bool, len(allTasks))

	for i, task := range allTasks {
		// The search can return a task more than once, but the unique index
		// on (task_id, project_view_id) allows only one row per task and view.
		if seenTasks[task.ID] {
			continue
		}
		seenTasks[task.ID] = true

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

	err = bulkInsertTaskPositions(s, newPositions, true)
	if err != nil {
		return
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
	log.Debugf("Recalculating task positions for view %d (repair mode)", view.ID)

	err := lockPositionsForViewUpdate(s, view.ID)
	if err != nil {
		return err
	}

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

	err = bulkInsertTaskPositions(s, newPositions, true)
	if err != nil {
		return err
	}

	log.Debugf("Repair: inserted %d new positions for view %d", len(newPositions), view.ID)

	events.DispatchOnCommit(s, &TaskPositionsRecalculatedEvent{
		NewTaskPositions: newPositions,
	})
	return nil
}

// getLowestPositionInView returns the smallest position in a view. has is false when no
// task in the view has a position at all.
func getLowestPositionInView(s *xorm.Session, viewID int64) (position float64, has bool, err error) {
	lowest := &TaskPosition{}
	has, err = s.Where("project_view_id = ?", viewID).
		OrderBy("position asc").
		Get(lowest)
	return lowest.Position, has, err
}

func calculateNewPositionForTask(s *xorm.Session, a web.Auth, t *Task, view *ProjectView) (*TaskPosition, error) {
	position := t.Position
	if position == 0 {
		lowestPosition, exists, err := getLowestPositionInView(s, view.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			if lowestPosition < MinPositionSpacing {
				err = RecalculateTaskPositions(s, view, a)
				if err != nil {
					return nil, err
				}

				lowestPosition, _, err = getLowestPositionInView(s, view.ID)
				if err != nil {
					return nil, err
				}
			}

			position = lowestPosition / 2
		}
	}

	return &TaskPosition{
		TaskID:        t.ID,
		ProjectViewID: view.ID,
		Position:      calculateDefaultPosition(t.Index, position),
	}, nil
}

// makeRoomAtTopOfView recalculates a view's positions when the gap above its first task is
// too small to fit count more tasks, and returns the lowest position afterwards. It returns
// 0 both when count is 0 and when the view has no positioned tasks - in either case there is
// nothing for the caller to insert above.
//
// This has to run before the new tasks are inserted: a recalculation orders by position,
// which they do not have yet, so they would land in an arbitrary order.
func makeRoomAtTopOfView(s *xorm.Session, view *ProjectView, count int, a web.Auth) (lowest float64, err error) {
	// Nothing to fit, so never recalculate: lowest/1 would fail the check below for any
	// view whose first task sits under the minimum spacing. Checked before the lock so an
	// empty batch does not serialize against real ones.
	if count == 0 {
		return 0, nil
	}

	// Without the lock two concurrent batches read the same lowest position and compute
	// byte-identical slots for their tasks. Callers must take the locks in a stable view
	// order to avoid deadlocking against each other.
	err = lockPositionsForViewUpdate(s, view.ID)
	if err != nil {
		return 0, err
	}

	lowest, has, err := getLowestPositionInView(s, view.ID)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil
	}

	if lowest/float64(count+1) >= MinPositionSpacing {
		return lowest, nil
	}

	err = RecalculateTaskPositions(s, view, a)
	if err != nil {
		return 0, err
	}

	lowest, _, err = getLowestPositionInView(s, view.ID)
	return lowest, err
}

// spreadTasksAtTopOfView places tasks in the gap above the view's first task, in slice
// order, so a batch of tasks placed together keeps the order it was passed in.
//
// lowest is what makeRoomAtTopOfView returned for this view. 0 means there is nothing to
// insert above, so the tasks start at the default spacing instead. Counting them off by
// their slice position rather than their task index keeps them apart in a view spanning
// several projects, where indexes repeat.
func spreadTasksAtTopOfView(tasks []*Task, view *ProjectView, lowest float64) []*TaskPosition {
	spacing := lowest / float64(len(tasks)+1)
	if lowest == 0 {
		spacing = math.Pow(2, 16)
	}

	positions := make([]*TaskPosition, 0, len(tasks))
	for i, t := range tasks {
		positions = append(positions, &TaskPosition{
			TaskID:        t.ID,
			ProjectViewID: view.ID,
			Position:      spacing * float64(i+1),
		})
	}

	return positions
}

// createTasksAtTopOfViews places a batch of tasks above everything the views of their
// projects already contain, keeping the slice order the tasks have per project.
//
// insert has to create exactly the tasks it is handed, without positions, and is called from
// here because the room must exist before it runs: making room can recalculate a view, which
// orders by position - tasks already inserted but not positioned yet would land in an
// arbitrary order. The state it gets is the one used for the view lookups here, so both share
// the same fetch.
//
// makeRoomAtTopOfView locks each view it touches, so projects and views are walked in a
// stable id order: two batches sharing a project would otherwise be free to grab the same
// locks in opposite orders and deadlock.
func createTasksAtTopOfViews(s *xorm.Session, a web.Auth, tasks []*Task, insert func(tasks []*Task, state *taskCreateState) error) (err error) {
	state := &taskCreateState{}

	tasksByProject := map[int64][]*Task{}
	for _, t := range tasks {
		tasksByProject[t.ProjectID] = append(tasksByProject[t.ProjectID], t)
	}

	projectIDs := make([]int64, 0, len(tasksByProject))
	for projectID := range tasksByProject {
		projectIDs = append(projectIDs, projectID)
	}
	slices.Sort(projectIDs)

	lowestByView := map[int64]float64{}
	for _, projectID := range projectIDs {
		views, err := state.viewsFor(s, projectID)
		if err != nil {
			return err
		}

		// Sorting a copy keeps the shared slice in the position order everything else
		// relies on - the lock order is only needed here.
		byID := slices.Clone(views)
		slices.SortFunc(byID, func(a, b *ProjectView) int {
			return cmp.Compare(a.ID, b.ID)
		})

		for _, view := range byID {
			lowest, err := makeRoomAtTopOfView(s, view, len(tasksByProject[projectID]), a)
			if err != nil {
				return err
			}
			lowestByView[view.ID] = lowest
		}
	}

	err = insert(tasks, state)
	if err != nil {
		return err
	}

	positions := []*TaskPosition{}
	for _, projectID := range projectIDs {
		views, err := state.viewsFor(s, projectID)
		if err != nil {
			return err
		}
		for _, view := range views {
			positions = append(positions, spreadTasksAtTopOfView(tasksByProject[projectID], view, lowestByView[view.ID])...)
		}
	}

	err = bulkInsertTaskPositions(s, positions, false)
	if err != nil {
		return err
	}

	return resolvePositionConflictsAfterInsert(s, positions)
}

type taskPositionKey struct {
	taskID int64
	viewID int64
}

// filterNewTaskPositions returns the positions whose (task_id, project_view_id)
// row does not exist yet, also deduplicating within the slice. Position creation
// during task creation can trigger a full recalculation (calculateNewPositionForTask
// or moveTaskToDoneBuckets) that already persists rows for the new task, so inserting
// the queued positions unconditionally would violate the unique index on
// (task_id, project_view_id).
func filterNewTaskPositions(s *xorm.Session, positions []*TaskPosition) ([]*TaskPosition, error) {
	if len(positions) == 0 {
		return positions, nil
	}

	taskIDs := make([]int64, 0, len(positions))
	seenTask := make(map[int64]bool, len(positions))
	for _, p := range positions {
		if seenTask[p.TaskID] {
			continue
		}
		seenTask[p.TaskID] = true
		taskIDs = append(taskIDs, p.TaskID)
	}

	// Fetch all existing rows for the involved tasks in one query so this stays
	// cheap when createTask runs in a loop (bulk import, project duplication).
	existing := []*TaskPosition{}
	err := s.In("task_id", taskIDs).Find(&existing)
	if err != nil {
		return nil, err
	}

	seen := make(map[taskPositionKey]bool, len(positions)+len(existing))
	for _, e := range existing {
		seen[taskPositionKey{taskID: e.TaskID, viewID: e.ProjectViewID}] = true
	}

	filtered := make([]*TaskPosition, 0, len(positions))
	for _, p := range positions {
		key := taskPositionKey{taskID: p.TaskID, viewID: p.ProjectViewID}
		if seen[key] {
			continue
		}
		seen[key] = true
		filtered = append(filtered, p)
	}
	return filtered, nil
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

	lowest, err := makeRoomAtTopOfView(s, view, len(tasks), a)
	if err != nil {
		return err
	}

	return bulkInsertTaskPositions(s, spreadTasksAtTopOfView(tasks, view, lowest), false)
}

// ensureTaskPositionsForSavedFilterView creates position rows for all tasks matching a saved
// filter which don't have one in its view yet. This must run before the fetch query: healing
// afterwards means the query sorts fresh matches last (NULL position) and they jump to the top
// on the next fetch once the position row exists.
func ensureTaskPositionsForSavedFilterView(s *xorm.Session, a web.Auth, projects []*Project, view *ProjectView, opts *taskSearchOptions) (err error) {
	if len(projects) == 0 {
		return nil
	}

	// Parse a fresh copy of the filters because convertFiltersToDBFilterCond mutates the
	// field names in place — reusing opts.parsedFilters would double-prefix them for the
	// subsequent fetch query.
	parsedFilters, err := getTaskFiltersFromFilterString(opts.filter, opts.filterTimezone)
	if err != nil {
		return err
	}

	// Check before converting: the conversion renames the field to task_buckets.bucket_id in place.
	joinTaskBuckets := hasBucketIDInParsedFilter(parsedFilters)

	filterCond, err := convertFiltersToDBFilterCond(parsedFilters, opts.filterIncludeNulls)
	if err != nil {
		return err
	}

	projectIDs, _ := getProjectIDsFromProjects(projects)
	if len(projectIDs) == 0 {
		return nil
	}

	query := s.
		Select("tasks.*").
		Join("LEFT", "task_positions", "task_positions.task_id = tasks.id AND task_positions.project_view_id = ?", view.ID).
		Where(builder.And(builder.In("tasks.project_id", projectIDs), filterCond)).
		And("task_positions.task_id IS NULL")

	if joinTaskBuckets {
		query = query.Join("LEFT", "task_buckets", "task_buckets.task_id = tasks.id AND task_buckets.project_view_id = ?", view.ID)
	}

	tasks := []*Task{}
	// Ordering by id keeps position assignment deterministic when multiple tasks are healed at once.
	err = query.
		OrderBy("tasks.id ASC").
		Find(&tasks)
	if err != nil {
		sql, vals := query.LastSQL()
		return fmt.Errorf("could not fetch unpositioned tasks, error was '%w', sql: '%v', values: %v", err, sql, vals)
	}

	return createPositionsForTasksInView(s, tasks, view, a)
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

	// Resolving a group respaces its tasks across the whole gap between its neighbours, so
	// members land both below and above where they were and the gaps left for the remaining
	// groups change - the order groups come back in decides whether a view gets respaced
	// locally or falls back to a full recalculation. Map iteration would make that differ
	// per run.
	slices.SortFunc(duplicates, func(a, b []*TaskPosition) int {
		return cmp.Compare(a[0].Position, b[0].Position)
	})

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
// This is called after bulk-inserting positions during task creation. A concurrent single
// create computes lowest/2, which can land on a slot of the batch.
// If resolveTaskPositionConflicts returns ErrNeedsFullRecalculation for a view,
// it falls back to a full recalculation of all positions in that view.
func resolvePositionConflictsAfterInsert(s *xorm.Session, positions []*TaskPosition) error {
	viewIDs := []int64{}
	insertedByView := map[int64][]float64{}
	seen := map[int64]map[float64]bool{}

	for _, pos := range positions {
		if seen[pos.ProjectViewID] == nil {
			seen[pos.ProjectViewID] = map[float64]bool{}
			viewIDs = append(viewIDs, pos.ProjectViewID)
		}
		if seen[pos.ProjectViewID][pos.Position] {
			continue
		}
		seen[pos.ProjectViewID][pos.Position] = true
		insertedByView[pos.ProjectViewID] = append(insertedByView[pos.ProjectViewID], pos.Position)
	}

	for _, viewID := range viewIDs {
		conflicts, err := findPositionConflictsAmong(s, viewID, insertedByView[viewID])
		if err != nil {
			return err
		}

		for _, group := range conflicts {
			err = resolveTaskPositionConflicts(s, viewID, group)
			if IsErrNeedsFullRecalculation(err) {
				err = recalculateTaskPositionsForRepair(s, &ProjectView{ID: viewID})
				if err != nil {
					return err
				}
				// The recalculation gave every task in the view a unique position, so the
				// remaining groups are gone with it.
				break
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// findPositionConflictsAmong returns the groups of task positions in a view which share a
// position value, looking only at the given positions. One query per view instead of one per
// position keeps a large batch from turning into a query per task and view.
func findPositionConflictsAmong(s *xorm.Session, projectViewID int64, positions []float64) (conflicts [][]*TaskPosition, err error) {
	// Keep statements well below the parameter limits of all supported databases.
	const batchSize = 100

	all := []*TaskPosition{}
	for start := 0; start < len(positions); start += batchSize {
		batch := positions[start:min(start+batchSize, len(positions))]

		found := []*TaskPosition{}
		err = s.
			Where("project_view_id = ?", projectViewID).
			In("position", batch).
			Find(&found)
		if err != nil {
			return nil, err
		}

		all = append(all, found...)
	}

	return findDuplicatesInPositions(all), nil
}
