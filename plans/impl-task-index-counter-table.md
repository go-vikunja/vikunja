# Task Index Counter Table Implementation Plan

**Goal:** Eliminate concurrent duplicate task indexes by using a PostgreSQL counter table with row-level locking to assign project-scoped task indexes atomically.

**Architecture:** On PostgreSQL, a new `project_task_counters` table and trigger atomically assign task indexes on INSERT. On MySQL/SQLite, nothing changes — the existing application-level `MAX(index)+1` logic remains as-is with no new tables or schema changes. The frontend stops pre-calculating indexes during bulk insert — the backend always owns index assignment.

**Tech Stack:** Go (XORM ORM), PostgreSQL triggers/plpgsql, Vue 3 + TypeScript

---

### Task 1: Create the database migration file and update initSchema

**Files:**
- Create: `pkg/migration/20260130120000.go`
- Modify: `pkg/migration/migration.go:265-273`

**Why two files:** The migration handles existing installs that are upgrading. But on brand-new installations, `initSchema` runs instead (via xormigrate's `InitSchema` callback), creates all tables from `GetTables()`, and marks every migration as already applied — so the migration never executes. Since `ProjectTaskCounter` is deliberately not in `GetTables()`, we must also set up the PostgreSQL counter table and trigger inside `initSchema`.

**Step 1: Generate migration skeleton**

Run: `mage dev:make-migration ProjectTaskCounter`

This creates a timestamped file in `pkg/migration/`. The actual timestamp will differ — use whatever `mage` generates. The instructions below reference the generated filename.

**Step 2: Implement the migration**

Replace the generated migration content with:

```go
package migration

import (
	"fmt"

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type ProjectTaskCounter20260130120000 struct {
	ProjectID int64 `xorm:"bigint not null pk"`
	LastIndex int64 `xorm:"bigint not null default 0"`
}

func (ProjectTaskCounter20260130120000) TableName() string {
	return "project_task_counters"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260130120000",
		Description: "Add project task counter table and PostgreSQL trigger for atomic index assignment",
		Migrate: func(tx *xorm.Engine) error {
			// Only PostgreSQL gets the counter table and trigger.
			// MySQL/SQLite continue using the existing application-level logic unchanged.
			if db.Type() != schemas.POSTGRES {
				return nil
			}
			return setupPostgresTaskIndexCounter(tx)
		},
		Rollback: func(tx *xorm.Engine) error {
			if db.Type() != schemas.POSTGRES {
				return nil
			}
			if _, err := tx.Exec(`DROP TRIGGER IF EXISTS task_set_index ON tasks`); err != nil {
				return err
			}
			if _, err := tx.Exec(`DROP FUNCTION IF EXISTS set_task_index()`); err != nil {
				return err
			}
			return tx.DropTables(ProjectTaskCounter20260130120000{})
		},
	})
}

// setupPostgresTaskIndexCounter creates the counter table, trigger function,
// and trigger for atomic task index assignment on PostgreSQL.
// Called from both the migration (existing installs) and initSchema (fresh installs).
func setupPostgresTaskIndexCounter(tx *xorm.Engine) error {
	// 1. Create the counter table
	if err := tx.Sync(ProjectTaskCounter20260130120000{}); err != nil {
		return fmt.Errorf("create project_task_counters table: %w", err)
	}

	// 2. Seed the counter table from existing task data (no-op on fresh installs)
	_, err := tx.Exec(`
		INSERT INTO project_task_counters (project_id, last_index)
		SELECT project_id, COALESCE(MAX("index"), 0)
		FROM tasks
		GROUP BY project_id
		ON CONFLICT (project_id) DO UPDATE
			SET last_index = EXCLUDED.last_index
	`)
	if err != nil {
		return fmt.Errorf("seed project_task_counters: %w", err)
	}

	// 3. Create the trigger function
	_, err = tx.Exec(`
		CREATE OR REPLACE FUNCTION set_task_index()
		RETURNS TRIGGER AS $$
		BEGIN
			INSERT INTO project_task_counters (project_id, last_index)
			VALUES (NEW.project_id, 1)
			ON CONFLICT (project_id) DO UPDATE
				SET last_index = project_task_counters.last_index + 1
			RETURNING last_index INTO NEW."index";
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	if err != nil {
		return fmt.Errorf("create set_task_index function: %w", err)
	}

	// 4. Create the trigger on the tasks table
	_, err = tx.Exec(`
		CREATE TRIGGER task_set_index
		BEFORE INSERT ON tasks
		FOR EACH ROW EXECUTE FUNCTION set_task_index()
	`)
	if err != nil {
		return fmt.Errorf("create task_set_index trigger: %w", err)
	}

	return nil
}
```

Note: The function is named `setupPostgresTaskIndexCounter` (exported-style within the package) so `initSchema` can call it too.

**Step 3: Update `initSchema` to call the PostgreSQL setup on fresh installs**

In `pkg/migration/migration.go`, modify the `initSchema` function (around line 265):

```go
func initSchema(tx *xorm.Engine) error {
	schemeBeans := []interface{}{}
	schemeBeans = append(schemeBeans, models.GetTables()...)
	schemeBeans = append(schemeBeans, files.GetTables()...)
	schemeBeans = append(schemeBeans, migration.GetTables()...)
	schemeBeans = append(schemeBeans, user.GetTables()...)
	schemeBeans = append(schemeBeans, notifications.GetTables()...)
	if err := tx.Sync2(schemeBeans...); err != nil {
		return err
	}

	// Set up PostgreSQL-specific counter table and trigger for atomic task index
	// assignment. This table is intentionally not in GetTables() to avoid creating
	// it on MySQL/SQLite.
	if db.Type() == schemas.POSTGRES {
		if err := setupPostgresTaskIndexCounter(tx); err != nil {
			return err
		}
	}

	return nil
}
```

You'll need to add `"xorm.io/xorm/schemas"` to the imports of `migration.go` if not already present.

**Step 4: Verify it compiles**

Run: `mage build`

Expected: Successful build with no errors.

**Step 5: Commit**

```bash
git add pkg/migration/<generated-timestamp>.go pkg/migration/migration.go
git commit -m "feat(db): add migration and initSchema setup for postgres task index counter"
```

---

### Task 2: Create the counter model struct (PostgreSQL only)

**Files:**
- Create: `pkg/models/project_task_counter.go`

The model struct is needed so Go code can reference the table (for deletion cleanup, counter reads, etc.), but it must **not** be registered in `GetTables()`. Registering it there would cause XORM's `Sync2` to auto-create the table on all databases, including MySQL/SQLite where we don't want it.

The table is created exclusively by the migration (Task 1) on PostgreSQL only.

**Step 1: Create the model struct**

Create `pkg/models/project_task_counter.go`:

```go
package models

// ProjectTaskCounter tracks the last assigned task index per project.
// This table only exists on PostgreSQL, where it is created by a migration
// and maintained by a database trigger. It is NOT registered in GetTables()
// to avoid XORM auto-syncing it on MySQL/SQLite.
type ProjectTaskCounter struct {
	ProjectID int64 `xorm:"bigint not null pk" json:"-"`
	LastIndex int64 `xorm:"bigint not null default 0" json:"-"`
}

func (ProjectTaskCounter) TableName() string {
	return "project_task_counters"
}
```

**Step 2: Verify it compiles**

Run: `mage build`

**Step 3: Commit**

```bash
git add pkg/models/project_task_counter.go
git commit -m "feat(models): add ProjectTaskCounter struct for postgres counter table"
```

---

### Task 3: Write failing tests for the new index logic

**Files:**
- Modify: `pkg/models/tasks_test.go`

We need tests that verify:
1. On PostgreSQL, the trigger assigns the index (application code skips assignment)
2. On non-PostgreSQL, the existing logic still works
3. Creating multiple tasks in the same project yields sequential indexes
4. Moving a task between projects still gets a correct index

**Step 1: Write the failing test for concurrent-safe index assignment**

Add to `pkg/models/tasks_test.go`, inside or after the existing `TestTask_Create` function:

```go
func TestTask_Create_IndexAssignment(t *testing.T) {
	usr := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	t.Run("sequential index assignment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 1 has tasks with indexes up to 17 in fixtures.
		// Creating a new task should get index 18.
		task1 := &Task{
			Title:     "First new task",
			ProjectID: 1,
		}
		err := task1.Create(s, usr)
		require.NoError(t, err)
		assert.Equal(t, int64(18), task1.Index)

		// Creating another task in the same session should get 19.
		task2 := &Task{
			Title:     "Second new task",
			ProjectID: 1,
		}
		err = task2.Create(s, usr)
		require.NoError(t, err)
		assert.Equal(t, int64(19), task2.Index)

		err = s.Commit()
		require.NoError(t, err)
	})

	t.Run("index assignment in empty project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 25 has no tasks. First task should get index 1.
		task := &Task{
			Title:     "First task in empty project",
			ProjectID: 25,
		}
		err := task.Create(s, usr)
		require.NoError(t, err)
		assert.Equal(t, int64(1), task.Index)

		err = s.Commit()
		require.NoError(t, err)
	})

	t.Run("provided index is ignored on postgres", func(t *testing.T) {
		if db.Type() != schemas.POSTGRES {
			t.Skip("trigger-based index assignment only on PostgreSQL")
		}

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Even if the client provides an index, the trigger should override it.
		task := &Task{
			Title:     "Task with provided index",
			ProjectID: 1,
			Index:     999,
		}
		err := task.Create(s, usr)
		require.NoError(t, err)
		// The trigger always assigns the next sequential index.
		assert.Equal(t, int64(18), task.Index)

		err = s.Commit()
		require.NoError(t, err)
	})
}
```

Note: You'll also need to add `"xorm.io/xorm/schemas"` to the import block of this test file (if not already present) for the `schemas.POSTGRES` reference.

**Step 2: Run the tests to verify they fail (or pass on the existing logic)**

Run: `mage test:filter TestTask_Create_IndexAssignment`

Expected: The first two tests should pass (they verify existing behavior). The third test (`provided index is ignored on postgres`) will only run on PostgreSQL and will fail until we modify the application code, because the current code respects provided indexes.

**Step 3: Commit**

```bash
git add pkg/models/tasks_test.go
git commit -m "test(tasks): add tests for sequential and trigger-based index assignment"
```

---

### Task 4: Modify backend index logic to be database-aware

**Files:**
- Modify: `pkg/models/tasks.go:835-869` (the `calculateNextTaskIndex` and `setNewTaskIndex` functions)

**Step 1: Modify `setNewTaskIndex` to skip on PostgreSQL**

On PostgreSQL, the trigger handles index assignment. The application code should set `Index = 0` so XORM doesn't include it in the INSERT column list and the trigger's `NEW.index` assignment takes effect.

However — there's a subtlety. XORM's `Insert` will include `index` in the INSERT statement if the struct field is non-zero. On PostgreSQL, we need XORM to either:
- Not include `index` in the INSERT (so the trigger sets it), or
- Include it but the trigger overwrites it regardless.

Looking at the trigger: it **always** runs `RETURNING last_index INTO NEW."index"` which overwrites whatever value was in `NEW.index`. So the trigger always wins, regardless of what XORM sends. This means the application code change is simpler: on PostgreSQL, we can skip the `setNewTaskIndex` call entirely, but we still need to read back the trigger-assigned index after INSERT.

Replace the functions at `pkg/models/tasks.go:835-869`:

```go
func calculateNextTaskIndex(s *xorm.Session, projectID int64) (nextIndex int64, err error) {
	latestTask := &Task{}
	_, err = s.
		Where("project_id = ?", projectID).
		OrderBy("`index` desc").
		Get(latestTask)
	if err != nil {
		return 0, err
	}

	return latestTask.Index + 1, nil
}

func setNewTaskIndex(s *xorm.Session, t *Task) (err error) {
	// On PostgreSQL, the database trigger handles index assignment atomically.
	// We set index to 0 so that after INSERT we can read back the trigger-assigned value.
	if db.Type() == schemas.POSTGRES {
		t.Index = 0
		return nil
	}

	// For MySQL/SQLite: keep the existing application-level logic.
	if t.Index == 0 {
		t.Index, err = calculateNextTaskIndex(s, t.ProjectID)
		return
	}

	exists, err := s.Where("project_id = ? AND `index` = ?", t.ProjectID, t.Index).Exist(&Task{})
	if err != nil {
		return err
	}
	if exists {
		t.Index, err = calculateNextTaskIndex(s, t.ProjectID)
		if err != nil {
			return err
		}
	}

	return
}
```

You also need to add the imports to `tasks.go` if not already present:

```go
"code.vikunja.io/api/pkg/db"
"xorm.io/xorm/schemas"
```

**Step 2: Read back the trigger-assigned index after INSERT (PostgreSQL only)**

In the `createTask` function at `pkg/models/tasks.go:889-991`, after `s.Insert(t)` (line 922), add a read-back for PostgreSQL:

Find this block (around lines 922-925):

```go
	_, err = s.Insert(t)
	if err != nil {
		return err
	}
```

Replace with:

```go
	_, err = s.Insert(t)
	if err != nil {
		return err
	}

	// On PostgreSQL, the trigger assigns the index. Read it back.
	if db.Type() == schemas.POSTGRES {
		has, err := s.ID(t.ID).Cols("index").Get(t)
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("task %d not found after insert", t.ID)
		}
	}
```

Note: Make sure `"fmt"` is in the imports.

**Step 3: Handle the project-move case**

In `updateSingleTask` around line 1188-1196, the code calls `calculateNextTaskIndex` when moving between projects. On PostgreSQL, the trigger only fires on INSERT, not UPDATE. So the application code still needs to calculate the next index for project moves. But we should use the counter table on PostgreSQL for this too.

Add a new helper function after `calculateNextTaskIndex`:

```go
// calculateNextTaskIndexFromCounter reads the counter table and returns the next index,
// atomically incrementing it. Used for project moves on PostgreSQL.
func calculateNextTaskIndexFromCounter(s *xorm.Session, projectID int64) (nextIndex int64, err error) {
	var lastIndex int64
	_, err = s.SQL(`
		INSERT INTO project_task_counters (project_id, last_index)
		VALUES (?, 1)
		ON CONFLICT (project_id) DO UPDATE
			SET last_index = project_task_counters.last_index + 1
		RETURNING last_index
	`, projectID).Get(&lastIndex)
	if err != nil {
		return 0, err
	}
	return lastIndex, nil
}
```

Then modify the project-move block in `updateSingleTask` (around line 1188-1196):

```go
	// If the task is being moved between projects, make sure to move the bucket + index as well
	if t.ProjectID != 0 && ot.ProjectID != t.ProjectID {
		if db.Type() == schemas.POSTGRES {
			t.Index, err = calculateNextTaskIndexFromCounter(s, t.ProjectID)
		} else {
			t.Index, err = calculateNextTaskIndex(s, t.ProjectID)
		}
		if err != nil {
			return err
		}
		t.BucketID = 0
		colsToUpdate = append(colsToUpdate, "index")
	}
```

**Step 4: Verify it compiles**

Run: `mage build`

**Step 5: Run all task tests**

Run: `mage test:filter TestTask_Create`

Expected: All tests pass.

Also run: `mage test:filter TestTask_Create_IndexAssignment`

Expected: All tests pass (including the PostgreSQL-specific test if running against PostgreSQL).

**Step 6: Commit**

```bash
git add pkg/models/tasks.go
git commit -m "feat(tasks): use postgres trigger for atomic index assignment, keep app logic for mysql/sqlite"
```

---

### Task 5: Clean up counter rows on project deletion (PostgreSQL only)

**Files:**
- Modify: `pkg/models/project.go:1164-1277`

Since the counter table only exists on PostgreSQL, the cleanup must be guarded by a `db.Type()` check.

**Step 1: Write a failing test**

Add to the existing project deletion tests (find where `TestProject_Delete` or similar exists, or add to `pkg/models/project_test.go`):

```go
t.Run("deleting a project should clean up its task counter", func(t *testing.T) {
	if db.Type() != schemas.POSTGRES {
		t.Skip("Counter table only exists on PostgreSQL")
	}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// First create a task to ensure a counter row exists
	task := &Task{
		Title:     "Counter test task",
		ProjectID: 36, // Use a project the test user can delete
	}
	err := task.Create(s, usr)
	require.NoError(t, err)

	// Verify counter row exists
	counter := &ProjectTaskCounter{ProjectID: 36}
	has, err := s.Get(counter)
	require.NoError(t, err)
	assert.True(t, has)

	// Delete the project
	p := &Project{ID: 36}
	err = p.Delete(s, usr)
	require.NoError(t, err)

	// Counter row should be gone
	has, err = s.Get(&ProjectTaskCounter{ProjectID: 36})
	require.NoError(t, err)
	assert.False(t, has)

	err = s.Commit()
	require.NoError(t, err)
})
```

Note: You'll need to pick a suitable project ID from the fixtures that the test user owns and can delete. Adjust the project ID based on the available fixtures.

**Step 2: Run the test to verify it fails**

Run: `mage test:filter "deleting a project should clean up its task counter"`

Expected: FAIL — counter row is not cleaned up yet.

**Step 3: Add cleanup to project Delete**

In `pkg/models/project.go`, in the `Delete` method, add the counter cleanup. Find the section that deletes related project entities (around line 1234) and add before the project itself is deleted. Guard it with a PostgreSQL check:

```go
	if db.Type() == schemas.POSTGRES {
		_, err = s.Where("project_id = ?", p.ID).Delete(&ProjectTaskCounter{})
		if err != nil {
			return
		}
	}
```

Add it just before the line `// Delete the project` (around line 1249).

You'll need to add imports for `"code.vikunja.io/api/pkg/db"` and `"xorm.io/xorm/schemas"` if not already present in `project.go`.

**Step 4: Run the test to verify it passes**

Run: `mage test:filter "deleting a project should clean up its task counter"`

Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/models/project.go pkg/models/project_test.go
git commit -m "fix(projects): clean up project_task_counters row on project deletion on postgres"
```

---

### Task 6: Remove frontend bulk index pre-calculation

**Files:**
- Modify: `frontend/src/components/tasks/AddTask.vue:136-196`

**Step 1: Remove the index pre-calculation logic**

The frontend currently queries the newest task per project and pre-calculates indexes for bulk task creation. This is no longer needed — the backend always assigns the correct index (atomically on PostgreSQL, sequentially on other DBs).

In `frontend/src/components/tasks/AddTask.vue`, simplify the `addTask()` function.

Remove the `projectIndices` Map and the loop that fetches the newest task per project (lines ~144-169):

```typescript
// DELETE THIS BLOCK:
const taskCollectionService = new TaskService()
const projectIndices = new Map<number, number>()

let currentProjectId = authStore.settings.defaultProjectId
if (typeof router.currentRoute.value.params.projectId !== 'undefined') {
	currentProjectId = Number(router.currentRoute.value.params.projectId)
}

// Create a map of project indices before creating tasks
if (tasksToCreate.length > 1) {
	for (const {project} of tasksToCreate) {
		const projectId = project !== null
			? await taskStore.findProjectId({project, projectId: 0})
			: currentProjectId

		if (!projectIndices.has(projectId)) {
			const newestTask = await taskCollectionService.getAll(new TaskModel({}), {
				sort_by: ['id'],
				order_by: ['desc'],
				per_page: 1,
				filter: `project_id = ${projectId}`,
			})
			projectIndices.set(projectId, newestTask[0]?.index || 0)
		}
	}
}
```

Keep the `currentProjectId` variable (it's still used below) — just move it up above the deleted block. Also remove the `taskCollectionService` usage here (but check if it's used elsewhere in the function — if not, remove its instantiation).

Then simplify the task creation map (lines ~171-196). Remove the index calculation:

```typescript
// REPLACE the newTasks map with this simplified version:
const newTasks = tasksToCreate.map(async ({title, project}, index) => {
	if (title === '') {
		return
	}

	const projectId = project !== null
		? await taskStore.findProjectId({project, projectId: 0})
		: currentProjectId

	const task = await taskStore.createNewTask({
		title,
		projectId: projectId || authStore.settings.defaultProjectId,
		position: props.defaultPosition,
	})
	createdTasks[title] = task
	return task
})
```

The key change: we no longer pass `index: taskIndex` to `createNewTask`. The backend assigns the index.

Also remove the `TaskService` import at the top of the `<script>` if it's no longer used anywhere else in this file:

```typescript
// Check if TaskService is still used. If not, remove:
import TaskService from '@/services/task'
```

And remove the `TaskModel` import if it was only used for the index query:

```typescript
// Check if TaskModel is still used. If not, remove:
import TaskModel from '@/models/task'
```

**Step 2: Verify the frontend builds**

Run: `cd frontend && pnpm build`

Expected: Successful build with no errors.

**Step 3: Verify lint passes**

Run: `cd frontend && pnpm lint`

Expected: No lint errors.

**Step 4: Commit**

```bash
git add frontend/src/components/tasks/AddTask.vue
git commit -m "refactor(frontend): remove client-side task index pre-calculation

The backend now always assigns task indexes atomically (via
PostgreSQL trigger or application-level logic for other DBs).
The frontend no longer needs to pre-calculate indexes during
bulk task creation."
```

---

### Task 7: Update the task store to stop passing index on create

**Files:**
- Modify: `frontend/src/stores/tasks.ts`

**Step 1: Remove `index` from `createNewTask` signature and usage**

In `frontend/src/stores/tasks.ts`, the `createNewTask` function accepts an `index` parameter. Since the backend now always assigns the index, the frontend should not pass it.

Find the function (around line 422) and remove the `index` parameter:

```typescript
async function createNewTask({
	title,
	bucketId,
	projectId,
	position,
} :
	Partial<ITask>,
) {
```

Remove `index` from the `TaskModel` constructor calls inside this function (around lines 443 and 482):

```typescript
// Change from:
return taskService.create(new TaskModel({
	title,
	projectId,
	bucketId,
	position,
	index,
}))

// To:
return taskService.create(new TaskModel({
	title,
	projectId,
	bucketId,
	position,
}))
```

Do this for both occurrences in the function (the quick-add path and the normal path).

**Step 2: Verify the frontend builds**

Run: `cd frontend && pnpm build`

**Step 3: Verify lint passes**

Run: `cd frontend && pnpm lint`

**Step 4: Commit**

```bash
git add frontend/src/stores/tasks.ts
git commit -m "refactor(frontend): remove index parameter from createNewTask"
```

---

### Task 8: Add integration tests for concurrent index safety

**Files:**
- Modify: `pkg/models/tasks_test.go`

These tests verify the core guarantee: concurrent inserts get unique sequential indexes.

**Step 1: Write the concurrency test**

Add to `pkg/models/tasks_test.go`:

```go
func TestTask_Create_ConcurrentIndexSafety(t *testing.T) {
	if db.Type() != schemas.POSTGRES {
		t.Skip("Concurrency test only meaningful on PostgreSQL with trigger")
	}

	db.LoadAndAssertFixtures(t)

	usr := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	const numGoroutines = 10
	results := make(chan int64, numGoroutines)
	errs := make(chan error, numGoroutines)

	// Launch concurrent task creations
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			s := db.NewSession()
			defer s.Close()

			task := &Task{
				Title:     fmt.Sprintf("Concurrent task %d", i),
				ProjectID: 1,
			}
			err := task.Create(s, usr)
			if err != nil {
				errs <- err
				return
			}
			err = s.Commit()
			if err != nil {
				errs <- err
				return
			}
			results <- task.Index
		}(i)
	}

	// Collect results
	indexes := make(map[int64]bool)
	for i := 0; i < numGoroutines; i++ {
		select {
		case idx := <-results:
			assert.False(t, indexes[idx], "duplicate index %d detected", idx)
			indexes[idx] = true
		case err := <-errs:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// All indexes should be unique
	assert.Len(t, indexes, numGoroutines)

	// Verify they are sequential (starting from 18, since project 1 has indexes up to 17)
	for i := int64(18); i < int64(18+numGoroutines); i++ {
		assert.True(t, indexes[i], "expected index %d to be present", i)
	}
}
```

Note: Add `"fmt"` to the imports if not already present.

**Step 2: Run the test**

Run: `mage test:filter TestTask_Create_ConcurrentIndexSafety`

Expected: PASS on PostgreSQL. Skipped on MySQL/SQLite.

**Step 3: Commit**

```bash
git add pkg/models/tasks_test.go
git commit -m "test(tasks): add concurrent index safety test for postgres trigger"
```

---

### Task 9: Add test for project move index assignment via counter table

**Files:**
- Modify: `pkg/models/tasks_test.go`

**Step 1: Add/extend the existing project move test**

The existing test at `tasks_test.go:392` already tests moving a task between projects. We need to verify it works correctly with both code paths. The existing test should continue to pass. Add a PostgreSQL-specific variant:

```go
t.Run("moving a task between projects on postgres uses counter table", func(t *testing.T) {
	if db.Type() != schemas.POSTGRES {
		t.Skip("Counter table path only on PostgreSQL")
	}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	task := &Task{
		ID:        12,
		ProjectID: 2, // Moving from project 1 to project 2
	}
	err := task.Update(s, u)
	require.NoError(t, err)
	err = s.Commit()
	require.NoError(t, err)
	assert.Equal(t, int64(3), task.Index)

	// Verify the counter table was updated
	counter := &ProjectTaskCounter{ProjectID: 2}
	has, err := s.Get(counter)
	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, int64(3), counter.LastIndex)
})
```

**Step 2: Run the test**

Run: `mage test:filter "moving a task between projects"`

Expected: PASS.

**Step 3: Commit**

```bash
git add pkg/models/tasks_test.go
git commit -m "test(tasks): verify project move updates counter table on postgres"
```

---

### Task 10: Run the full test suite and fix any regressions

**Step 1: Run backend tests**

Run: `mage test:feature`

Expected: All tests pass.

**Step 2: Run web tests**

Run: `mage test:web`

Expected: All tests pass. Pay attention to:
- `TestTask` Create subtests in `pkg/webtests/task_test.go`
- Any test that creates tasks and asserts on the response body

If any web tests fail because the returned `index` value differs from expectations, update those test assertions. The web tests use SQLite by default, so the existing logic should still work. If tests use PostgreSQL, the trigger will assign indexes.

**Step 3: Run backend lint**

Run: `mage lint`

Expected: No lint errors. Fix any issues.

**Step 4: Run frontend checks**

Run: `cd frontend && pnpm lint && pnpm typecheck && pnpm build`

Expected: All pass.

**Step 5: Commit any fixes**

```bash
git add -u
git commit -m "fix: address test regressions from task index counter table changes"
```

---

### Task 11: Manual verification

**Step 1: Start the dev server with PostgreSQL**

Make sure `config.yml` is configured to use PostgreSQL. Start the backend:

Run: `mage build && ./vikunja`

Verify in the logs that migrations ran successfully.

**Step 2: Verify the trigger exists**

Connect to the PostgreSQL database and run:

```sql
SELECT tgname, tgrelid::regclass, tgenabled
FROM pg_trigger
WHERE tgname = 'task_set_index';
```

Expected: One row showing the trigger on the `tasks` table.

**Step 3: Verify the counter table**

```sql
SELECT * FROM project_task_counters ORDER BY project_id;
```

Expected: One row per project that has tasks, with `last_index` matching the highest task index in that project.

**Step 4: Create tasks via the UI and verify indexes**

1. Create a single task — verify it gets the next sequential index
2. Create multiple tasks via bulk add (paste multiple lines) — verify each gets a unique sequential index
3. Move a task to another project — verify it gets the correct index in the new project

**Step 5: Verify MySQL/SQLite still work**

Switch `config.yml` to SQLite, rebuild, and repeat the task creation tests. The behavior should be identical to before this change — application-level index assignment.

---

## Summary of all changed files

| File | Action | Purpose |
|------|--------|---------|
| `pkg/migration/<timestamp>.go` | Create | Migration + `setupPostgresTaskIndexCounter` (PG only, no-op on MySQL/SQLite) |
| `pkg/migration/migration.go` | Modify | Call `setupPostgresTaskIndexCounter` from `initSchema` for fresh PG installs |
| `pkg/models/project_task_counter.go` | Create | Model struct for counter table (not registered in GetTables — PG only) |
| `pkg/models/tasks.go` | Modify | Skip app-level index on PG, read back trigger value, use counter for moves |
| `pkg/models/project.go` | Modify | Delete counter row on project deletion (PG only) |
| `pkg/models/tasks_test.go` | Modify | Add sequential, PG-specific, and concurrency tests |
| `frontend/src/components/tasks/AddTask.vue` | Modify | Remove client-side index pre-calculation |
| `frontend/src/stores/tasks.ts` | Modify | Remove index parameter from createNewTask |
