# Auto-Generated Tasks — Technical Architecture

## Overview

Auto-Generated Tasks are recurring task templates that materialize task instances
only when they become due. Unlike Vikunja's built-in repeating tasks, they don't
populate the board with future instances.

## Core Rules

1. **One instance at a time** — Only one open (undone) task per template can exist.
   If the previous task isn't completed, no new one is created — it goes overdue.

2. **Completion-based scheduling** — `next_due_at` is recalculated from when the
   user COMPLETES the task, not when it was created. This prevents cascading
   backlog if a user falls behind.

3. **Pause without destroy** — Templates can be paused (`active = false`) to stop
   generation without losing configuration. Resume at any time.

4. **Dual trigger** — Tasks are created by:
   - Frontend auto-check on page load (instant feel)
   - Backend cron job every 15 minutes (reliability) [TODO]

## Data Model

### auto_task_templates
| Column | Type | Description |
|--------|------|-------------|
| id | BIGINT PK | Auto-increment |
| owner_id | BIGINT NOT NULL INDEX | FK to users |
| project_id | BIGINT NULL | FK to projects (NULL = default) |
| title | VARCHAR(250) NOT NULL | Task title |
| description | LONGTEXT NULL | Task description |
| priority | BIGINT NULL | 0-5 |
| hex_color | VARCHAR(7) NULL | Color |
| label_ids | JSON NULL | e.g. [1, 5, 12] |
| assignee_ids | JSON NULL | e.g. [3, 7] |
| interval_value | INT NOT NULL DEFAULT 1 | Repeat every N... |
| interval_unit | VARCHAR(10) NOT NULL DEFAULT 'days' | hours/days/weeks/months |
| start_date | DATETIME NOT NULL | First occurrence |
| end_date | DATETIME NULL | Stop after (optional) |
| active | BOOLEAN NOT NULL DEFAULT TRUE | FALSE = paused |
| last_created_at | DATETIME NULL | Last task creation time |
| last_completed_at | DATETIME NULL | Last task completion time |
| next_due_at | DATETIME NULL | When next task should appear |

### auto_task_log
| Column | Type | Description |
|--------|------|-------------|
| id | BIGINT PK | Auto-increment |
| template_id | BIGINT NOT NULL INDEX | FK to auto_task_templates |
| task_id | BIGINT NOT NULL | FK to tasks (created task) |
| trigger_type | VARCHAR(20) NOT NULL | 'system' / 'manual' / 'cron' |
| triggered_by_id | BIGINT NULL | FK to users (NULL for system) |
| created | DATETIME NOT NULL | Timestamp |

### tasks (modified)
| Column | Type | Description |
|--------|------|-------------|
| auto_template_id | BIGINT NULL INDEX | FK to auto_task_templates |

**Note:** `auto_template_id` is added as a DB column via migration but is NOT
added to the Go `Task` struct. It is read/written via raw SQL to avoid modifying
the large upstream `tasks.go` file.

## Auto-Creation Flow

```
CheckAndCreateAutoTasks(session, user):
    templates = SELECT * FROM auto_task_templates
                WHERE owner_id = user.id
                AND active = true
                AND next_due_at <= NOW()

    for each template:
        if end_date is set and NOW() > end_date:
            → deactivate template, skip

        open_count = SELECT COUNT(*) FROM tasks
                     WHERE auto_template_id = template.id
                     AND done = false

        if open_count > 0:
            → skip (previous task still open, goes overdue naturally)

        task = createTask(template properties)
        UPDATE tasks SET auto_template_id = template.id WHERE id = task.id
        INSERT INTO auto_task_log (template_id, task_id, trigger_type)
        UPDATE template SET last_created_at = NOW()
```

## Completion Hook

```
OnAutoTaskCompleted(task):
    auto_template_id = SQL: SELECT auto_template_id FROM tasks WHERE id = task.id
    if auto_template_id == 0: return

    template = load(auto_template_id)
    template.last_completed_at = NOW()
    template.next_due_at = NOW() + interval   ← key: from completion, not creation
    save(template)
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/autotasks | List templates (with last 10 log entries) |
| GET | /api/v1/autotasks/:id | Get one (with last 20 log entries) |
| PUT | /api/v1/autotasks | Create template |
| POST | /api/v1/autotasks/:id | Update template |
| DELETE | /api/v1/autotasks/:id | Delete template + log + attachments |
| POST | /api/v1/autotasks/:id/trigger | Manual "send now" |
| POST | /api/v1/autotasks/check | Auto-check all due templates |

## Build Compatibility Notes

The Vikunja codebase has specific patterns that must be followed:

- **Echo v5** — `github.com/labstack/echo/v5`, NOT v4
- **Handler signatures** — `func Name(c *echo.Context) error` (pointer receiver)
- **NewHTTPError** — Always two args: `echo.NewHTTPError(http.StatusXxx, "message")`
- **Auth** — `auth2.GetAuthFromClaims(c)` from `code.vikunja.io/api/pkg/modules/auth`
- **CObject interface** — `ReadAll` returns `(interface{}, int, int64, error)` (3rd is int64)
- **createTask** — 5 args: `(session, task, auth, updateAssignees, setBucket)`
- **TaskAssginee** — Note the typo; this is the upstream spelling
- **Permissions** — Use `PermissionAdmin` from models package, not `web.RightAdmin`

## File Manifest

| File | Location | Purpose |
|------|----------|---------|
| auto_task_template.go | pkg/models/ | Model, CRUD, permissions, errors |
| auto_task_create.go | pkg/models/ | Check, trigger, completion, scheduling |
| auto_task_handler.go | pkg/routes/api/v1/ | Echo HTTP handlers |
| 20260224070000.go | pkg/migration/ | DB migration |
| routes.go | pkg/routes/ | Endpoint registration |
| autoTaskApi.ts | frontend/src/services/ | HTTP client |
| AutoTaskEditor.vue | frontend/src/components/tasks/partials/ | Editor UI |
| ListTemplates.vue | frontend/src/views/templates/ | Tab integration |
| Home.vue | frontend/src/views/ | Auto-check trigger |
| en.json | frontend/src/i18n/lang/ | 41 i18n keys |

## TODO

- [ ] Backend cron goroutine (15-min interval)
- [ ] Hook OnAutoTaskCompleted into task update path
- [ ] Copy attachments from template to generated task
- [ ] Auto-gen indicator (⟳) on SingleTaskInProject for linked tasks
- [ ] Bulk pause/resume
- [ ] Log pagination for templates with long history
