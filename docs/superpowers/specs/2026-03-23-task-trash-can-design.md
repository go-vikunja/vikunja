# Task Trash Can (Soft-Delete)

**Date:** 2026-03-23
**Status:** Approved

## Summary

Add trash can functionality to Chaku so that deleted tasks are soft-deleted rather than permanently removed. Trashed tasks retain all relationships (assignees, labels, reminders, comments, attachments, relations, positions, buckets) and can be restored within 30 days. After 30 days, a cron job permanently deletes them using the existing hard-delete cascade.

## Motivation

Currently, `Task.Delete()` is a destructive cascade that removes the task and all 9 related table entries immediately. There is no undo. Users who accidentally delete a task lose all associated data permanently. A trash can provides a safety net consistent with how every major application handles deletion.

## Design

### Approach: `deleted_at` Timestamp

A nullable `deleted_at` column on the `tasks` table serves dual purpose:

1. **Trash indicator** — `deleted_at IS NOT NULL` means the task is in the trash
2. **Purge clock** — the timestamp tells the cron job when to permanently delete

This is preferred over a boolean `is_trashed` (which would require a separate timestamp for purge timing) and over a separate `trashed_tasks` table (which would duplicate the schema and complicate restore).

### Database

#### Migration

Add to `tasks` table:

```sql
ALTER TABLE tasks ADD COLUMN deleted_at DATETIME NULL;
CREATE INDEX idx_tasks_deleted_at ON tasks (deleted_at);
```

The index supports both the filter (`WHERE deleted_at IS NULL`) applied to all task queries and the purge job (`WHERE deleted_at < ?`).

### Backend Changes

#### Task Struct

```go
// In pkg/models/tasks.go, add to Task struct:
DeletedAt *time.Time `xorm:"datetime null index 'deleted_at'" json:"deleted_at,omitempty"`
```

The field is a pointer so it serializes as `null` (omitted) rather than zero-time when not trashed.

#### Soft-Delete (Trash)

`Task.Delete()` changes from hard-delete cascade to:

```go
func (t *Task) Delete(s *xorm.Session, a web.Auth) error {
    // Read full task for the event (existing logic)
    fullTask := &Task{ID: t.ID}
    err := fullTask.ReadOne(s, a)
    if err != nil {
        return err
    }

    now := time.Now()
    fullTask.DeletedAt = &now
    _, err = s.ID(t.ID).Cols("deleted_at").Update(fullTask)
    if err != nil {
        return err
    }

    doer, _ := user.GetFromAuth(a)
    events.DispatchOnCommit(s, &TaskTrashedEvent{
        Task: fullTask,
        Doer: doer,
    })
    return nil
}
```

No relationships are modified. The task row stays, all FK references remain intact.

#### Hard-Delete

The current cascade logic in `Task.Delete()` moves to a new `Task.HardDelete()` method. This is called by:

- The purge cron job
- The "permanently delete" API endpoint
- The "empty trash" API endpoint

```go
func (t *Task) HardDelete(s *xorm.Session, a web.Auth) error {
    // ... existing cascade: assignees, favorites, labels, attachments,
    //     comments, unread statuses, relations, reminders, positions,
    //     buckets, then the task row itself
    // ... fire TaskDeletedEvent (existing)
}
```

#### Query Filtering

All task queries must exclude trashed tasks. The primary locations:

1. **`TaskCollection.ReadAll()`** (`pkg/models/task_collection.go`) — add `AND tasks.deleted_at IS NULL` to the WHERE clause
2. **`GetTaskByID()` / `GetTaskSimple()`** (`pkg/models/tasks.go`) — add `.And("tasks.deleted_at IS NULL")`
3. **`ReadOne()`** — add the filter
4. **Kanban bucket queries** — tasks in buckets must exclude trashed
5. **Gantt view queries** — same
6. **CalDAV task listing** (`pkg/routes/caldav/listStorageProvider.go`) — exclude trashed
7. **Task search** (`pkg/models/task_search.go`) — exclude trashed
8. **Task count metrics** — exclude trashed from counts
9. **Relation lookups** (`pkg/models/task_relation.go`) — when loading related tasks, exclude trashed ones
10. **Reminder queries** — trashed tasks should not trigger reminders

The trash view endpoints bypass this filter to show only trashed tasks (`WHERE deleted_at IS NOT NULL`).

#### New API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/trash` | List trashed tasks (paginated). Supports `?project_id=X` filter. Returns tasks with project info. Only returns tasks the authenticated user has read access to. |
| `POST` | `/api/v1/trash/{taskID}/restore` | Restore a trashed task. Sets `deleted_at = NULL`. Requires `CanUpdate` permission on the task's project. |
| `DELETE` | `/api/v1/trash/{taskID}` | Permanently delete a single trashed task. Calls `HardDelete`. Requires `CanDelete` permission. |
| `DELETE` | `/api/v1/trash` | Empty trash. Hard-deletes all trashed tasks the user has `CanDelete` permission for. |

The `GET /api/v1/trash` response includes the task's `project_id` and project title so the frontend can group by project.

#### Permissions

- **Trashing** (soft-delete): existing `CanDelete` permission — same as current delete
- **Viewing trash**: existing `CanRead` permission on the task's project
- **Restoring**: existing `CanUpdate` permission on the task's project (the task is being moved back, which is an update)
- **Permanent delete**: existing `CanDelete` permission
- **Empty trash**: iterates all trashed tasks, hard-deletes those where user has `CanDelete`

No new permission types needed.

#### Events

Two new events:

```go
// TaskTrashedEvent fires when a task is moved to trash
type TaskTrashedEvent struct {
    Task *Task
    Doer *user.User
}

func (t *TaskTrashedEvent) Name() string { return "task.trashed" }

// TaskRestoredEvent fires when a task is restored from trash
type TaskRestoredEvent struct {
    Task *Task
    Doer *user.User
}

func (t *TaskRestoredEvent) Name() string { return "task.restored" }
```

The existing `TaskDeletedEvent` continues to fire only on hard-delete (permanent deletion). This distinction matters for webhooks — consumers should know whether a task was trashed (recoverable) or permanently deleted.

Register both events for webhook dispatch if webhooks are enabled.

**Notifications:** Register a `SendTaskTrashedNotification` listener on `TaskTrashedEvent` that notifies task subscribers that the task was moved to trash (not permanently deleted). The existing `SendTaskDeletedNotification` stays on `TaskDeletedEvent` for permanent deletions. Similarly, register a notification listener on `TaskRestoredEvent`.

**Metrics:** `TaskTrashedEvent` should fire `DecreaseTaskCounter` (trashed tasks should not count as active). `TaskRestoredEvent` should fire `IncreaseTaskCounter`.

#### Cron Job

A daily cron job purges expired trash:

```go
func RegisterTrashPurgeJob() {
    cron.Schedule("@daily", func() {
        s := db.NewSession()
        defer s.Close()

        cutoff := time.Now().Add(-30 * 24 * time.Hour)

        var tasks []*Task
        err := s.Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).Find(&tasks)
        // For each task, call HardDelete with a system auth
        // Log the count of purged tasks
    })
}
```

Registered during server startup alongside existing cron jobs.

#### CalDAV

CalDAV `DELETE` operations map to soft-delete (trash). CalDAV `REPORT`/`PROPFIND` queries exclude trashed tasks. This means a CalDAV client that deletes a task will see it disappear, but it can be restored via the web UI within 30 days.

### Frontend Changes

#### Trash Navigation Item

Add a trash can entry at the bottom of the left sidebar (`frontend/src/components/home/Navigation.vue`):

- Icon: trash can (FontAwesome `fa-trash-can`)
- Label: "Trash"
- Badge: count of trashed tasks (fetched from `GET /api/v1/trash` with a lightweight count-only parameter, or derived from the response)
- Position: absolute bottom of the sidebar, visually separated from the project list

#### Trash View

New view at `/trash` (`frontend/src/views/trash/TrashView.vue`):

- **Header:** "Trash" with subtitle "Items are permanently deleted after 30 days"
- **Empty Trash button:** Top-right, with confirmation dialog ("This will permanently delete all items in the trash. This cannot be undone.")
- **Task list:** Grouped by project. Each task shows:
  - Task title
  - Project name (with link/badge)
  - "Deleted X days ago" relative timestamp
  - "Y days remaining" until auto-purge
  - Action buttons: "Restore" and "Delete Permanently"
- **Empty state:** "Trash is empty" message
- **Delete Permanently confirmation:** Per-task confirmation dialog

#### Delete Action Change

The current "Delete" action on tasks throughout the UI becomes "Move to Trash":

- No confirmation dialog needed (the action is reversible)
- Brief toast notification: "Task moved to trash" with an inline "Undo" link that calls the restore endpoint
- The new `TaskTrashedEvent`-driven notification email informs subscribers the task was trashed (not permanently deleted)

#### Router

Add route:

```typescript
{
    path: '/trash',
    name: 'trash',
    component: () => import('@/views/trash/TrashView.vue'),
    meta: { requiresAuth: true },
}
```

### What Does NOT Change

- Task relationships (assignees, labels, reminders, comments, attachments, relations, positions, buckets) are untouched on trash
- Permission model — no new permission types
- Existing `TaskDeletedEvent` — still fires on hard-delete only
- Database schema for all related tables — no FK changes needed since the task row persists

### Edge Cases

1. **Trashed task's project is deleted:** When a project is deleted, `Project.Delete()` must call `HardDelete()` (not the new soft-delete `Delete()`) on all its tasks. This must include both active and already-trashed tasks — the query in `Project.Delete()` must bypass the `deleted_at IS NULL` filter to avoid orphaning trashed tasks that would never be purged.
2. **Restoring a task whose project was deleted:** Not possible — the task would have been hard-deleted with the project.
3. **Duplicate task in trash:** A user creates a task, trashes it, creates another with the same title. Both exist independently. No conflict.
4. **Task relations pointing to trashed tasks:** When rendering relations for a visible task, trashed related tasks are excluded from the response. On restore, relations reappear naturally since the FK rows were never removed.
5. **CalDAV sync after trash:** The task disappears from CalDAV clients. If restored, it reappears on next sync.
6. **Recurring tasks:** If a recurring task is trashed, it stays trashed. The recurrence engine should skip trashed tasks (filter on `deleted_at IS NULL`).

### Testing

- Unit tests for `Task.Delete()` verifying soft-delete behavior (sets `deleted_at`, does not remove relationships)
- Unit tests for `Task.HardDelete()` verifying cascade behavior (existing tests, relocated)
- Unit tests for `Task.Restore()` verifying `deleted_at` is cleared
- Integration tests for all four trash endpoints
- Test that trashed tasks are excluded from: task collections, search, Kanban, Gantt, CalDAV, relations, reminders, metrics
- Test the purge cron job with tasks at various ages
- Test permission enforcement on trash operations
- Frontend: E2E test for trash → restore → verify task reappears
