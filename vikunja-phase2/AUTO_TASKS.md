# Auto-Generated Tasks — Technical Architecture

## Overview

Auto-Generated Tasks are recurring task templates that materialize task instances
only when they become due. Unlike Vikunja's built-in repeating tasks, they don't
populate the board with future instances.

## Core Rules

1. **One instance at a time** — Only one open (undone) task per template can exist.
   If the previous task isn't completed, no new one is created.

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
```sql
id                BIGINT PK AUTO_INCREMENT
owner_id          BIGINT NOT NULL INDEX     -- FK to users
project_id        BIGINT NULL               -- FK to projects (NULL = default)
title             VARCHAR(250) NOT NULL
description       LONGTEXT NULL
priority          BIGINT NULL
hex_color         VARCHAR(7) NULL
label_ids         JSON NULL                 -- [1, 5, 12]
assignee_ids      JSON NULL                 -- [3, 7]
interval_value    INT NOT NULL DEFAULT 1
interval_unit     VARCHAR(10) NOT NULL DEFAULT 'days'  -- hours|days|weeks|months
start_date        DATETIME NOT NULL
end_date          DATETIME NULL
active            BOOLEAN NOT NULL DEFAULT TRUE
last_created_at   DATETIME NULL
last_completed_at DATETIME NULL
next_due_at       DATETIME NULL
created           DATETIME NOT NULL
updated           DATETIME NOT NULL
```

### auto_task_log
```sql
id                BIGINT PK AUTO_INCREMENT
template_id       BIGINT NOT NULL INDEX     -- FK to auto_task_templates
task_id           BIGINT NOT NULL           -- FK to tasks (the created task)
trigger_type      VARCHAR(20) NOT NULL      -- 'system' | 'manual' | 'cron'
triggered_by_id   BIGINT NULL               -- FK to users (NULL for system)
created           DATETIME NOT NULL
```

### tasks (modified)
```sql
auto_template_id  BIGINT NULL INDEX         -- FK to auto_task_templates
```

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

        open_count = COUNT(*) FROM tasks
                     WHERE auto_template_id = template.id
                     AND done = false

        if open_count > 0:
            → skip (previous task still open, goes overdue naturally)

        task = create_task(template properties)
        INSERT INTO auto_task_log (template_id, task_id, trigger_type)

        UPDATE template SET last_created_at = NOW()
```

## Completion Hook

```
OnAutoTaskCompleted(task):
    if task.auto_template_id == 0: return

    template = load(task.auto_template_id)
    template.last_completed_at = NOW()
    template.next_due_at = NOW() + interval
    save(template)
```

Key: next_due_at is calculated from NOW (completion time), not from the
original due date. If a task was due Monday and completed Wednesday,
and interval is 1 day, next due = Thursday (not Tuesday).

## Manual Trigger

```
TriggerAutoTask(template_id, user):
    if open task exists for template:
        → error "complete the previous task first"

    task = create_task(template properties, due_date = NOW())
    log(trigger_type = 'manual', triggered_by = user)
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/autotasks | List user's templates (with last 10 log entries) |
| GET | /api/v1/autotasks/:id | Get one template (with last 20 log entries) |
| PUT | /api/v1/autotasks | Create template |
| POST | /api/v1/autotasks/:id | Update template |
| DELETE | /api/v1/autotasks/:id | Delete template + log + attachments |
| POST | /api/v1/autotasks/:id/trigger | Manual "send now" |
| POST | /api/v1/autotasks/check | Auto-check all due templates |

## Frontend Components

### AutoTaskEditor.vue
- Card list showing all templates with status dot (green=active, grey=paused)
- Inline actions: pause/resume, send now, edit, delete
- Metadata row: interval, project, next due date (red if overdue)
- Collapsible generation log per template
- Create/edit modal with full task properties

### Integration Points
- `ListTemplates.vue` — Third tab "Auto-Generated" with robot icon
- `Home.vue` — Calls `/autotasks/check` on mount, refreshes task list if new tasks created
- `autoTaskApi.ts` — Frontend HTTP client

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
