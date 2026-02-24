# Auto-Generated Tasks — Architecture

## Concept
Task templates that auto-materialize when they're due, without cluttering the board
with future instances. Only one active instance exists at a time.

## Data Model

### AutoTaskTemplate (new table: `auto_task_templates`)
```
id              bigint PK autoincr
owner_id        bigint NOT NULL INDEX (FK users)
project_id      bigint NULL (FK projects, NULL = user's default project)
title           varchar(250) NOT NULL
description     longtext NULL
priority        bigint NULL
hex_color       varchar(7) NULL
label_ids       json NULL
assignee_ids    json NULL

-- Scheduling
interval_value  int NOT NULL default 1
interval_unit   varchar(10) NOT NULL default 'days'  -- hours/days/weeks/months
start_date      datetime NOT NULL                     -- first occurrence
end_date        datetime NULL                         -- optional: stop generating after this
active          boolean NOT NULL default true          -- pause/resume toggle

-- Tracking
last_created_at datetime NULL   -- when the last task instance was created
next_due_at     datetime NULL   -- pre-computed: when the next instance should exist

-- Metadata
created         datetime NOT NULL
updated         datetime NOT NULL
```

### AutoTaskTemplateAttachment (new table: `auto_task_template_attachments`)
```
id              bigint PK autoincr
template_id     bigint NOT NULL INDEX (FK auto_task_templates)
file_id         bigint NOT NULL (FK files)
file_name       varchar(250) NOT NULL
created_by_id   bigint NOT NULL
created         datetime NOT NULL
```

### Task linkage
When an auto-task is created, the task gets a special field:
- `auto_template_id bigint NULL` on the `tasks` table
This lets us check "does an open instance already exist?" and link back.

## API Endpoints

### CRUD
- `GET    /api/v1/autotasks`           — list user's templates
- `POST   /api/v1/autotasks`           — create template
- `GET    /api/v1/autotasks/:id`       — get one
- `PUT    /api/v1/autotasks/:id`       — update
- `DELETE /api/v1/autotasks/:id`       — delete
- `POST   /api/v1/autotasks/:id/trigger` — manually create a task NOW

### Attachments
- `PUT    /api/v1/autotasks/:id/attachments` — upload
- `DELETE /api/v1/autotasks/:id/attachments/:attachmentId` — remove

### Auto-creation trigger
- `POST   /api/v1/autotasks/check`     — frontend calls on page load
  Checks all active templates for the current user, creates any that are due.
  Returns list of newly created tasks.

## Auto-Creation Logic

```
For each active template where next_due_at <= now():
  1. Check: is there already an open (done=false) task with this auto_template_id?
     → YES: skip (user hasn't completed the previous one yet)
     → NO: continue
  2. Create task in target project (or user's default)
     - Set all properties from template
     - Set due_date = next_due_at
     - Set auto_template_id = template.id
     - Copy attachments from template to task
  3. Update template:
     - last_created_at = now()
     - next_due_at = calculate_next(next_due_at, interval_value, interval_unit)
```

### next_due_at calculation
```
next = current_due + interval
If next is still in the past (missed multiple intervals), fast-forward:
  while next < now(): next += interval
```

## Backend Cron
A goroutine runs every 15 minutes:
1. Query all active templates where next_due_at <= now()
2. For each, run the auto-creation logic above
3. Log created tasks

## Frontend Trigger
On Home.vue and ShowTasks.vue mount:
1. Call `POST /api/v1/autotasks/check`
2. If tasks were created, refresh the task list
3. Show a subtle toast: "Auto-created X tasks"

## Frontend UI

### Templates Tab (in ListTemplates.vue)
Add a third tab: "Auto-Generated" alongside Templates and Chains.

### AutoTaskEditor component
- Title, description (rich text), priority, color, labels, assignees
- Project selector (default: inbox)
- Interval: value + unit (hours/days/weeks/months)
- Start date picker
- Optional end date
- Active toggle
- Attachments section
- "Send to Project Now" button (calls /trigger)
- Preview: shows next_due_at, last_created_at

### Task list indicator
Tasks created by auto-gen show a small ⟳ icon to indicate they're auto-generated.
Clicking it links to the template.

## File layout
Backend:
- pkg/models/auto_task_template.go       — model + CRUD
- pkg/models/auto_task_create.go         — auto-creation logic
- pkg/models/auto_task_cron.go           — cron runner
- pkg/migration/20260224070000.go        — tables + task column

Frontend:
- frontend/src/services/autoTaskApi.ts   — API client
- frontend/src/components/tasks/partials/AutoTaskEditor.vue
- Update ListTemplates.vue with third tab
- Update Home.vue / ShowTasks.vue with check trigger
