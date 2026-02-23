# Task Templates Feature — Vikunja

Adds reusable task templates that can be saved from existing tasks and used to create new tasks in any project.

## Overview

**Two ways to create templates:**
1. **Save as Template** — From any existing task (kanban card menu or task detail sidebar)
2. **Create from scratch** — Via the API (frontend template editor can be added later)

**Three ways to use templates:**
1. **Kanban board** — "From Template" button in the header
2. **Task detail view** — "Save as Template" in the sidebar actions
3. **API** — `PUT /tasktemplates/:template/tasks` endpoint

## What's included in a template

- Title, description, priority, color, percent done
- Repeat settings (repeat after, repeat mode)
- Label IDs (applied when creating a task from the template)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `PUT` | `/tasktemplates` | Create a template |
| `GET` | `/tasktemplates` | List all user's templates |
| `GET` | `/tasktemplates/:id` | Get one template |
| `POST` | `/tasktemplates/:id` | Update a template |
| `DELETE` | `/tasktemplates/:id` | Delete a template |
| `PUT` | `/tasktemplates/:id/tasks` | Create a task from a template |

### Create task from template request:
```json
{
  "target_project_id": 123,
  "title": "Optional override title"
}
```

## Files

### New Files (11)

| File | Description |
|------|-------------|
| `pkg/models/task_template.go` | Backend TaskTemplate model with full CRUD |
| `pkg/models/task_from_template.go` | Backend TaskFromTemplate - creates tasks from templates |
| `pkg/migration/20260223120000.go` | Database migration for task_templates table |
| `frontend/src/modelTypes/ITaskTemplate.ts` | TypeScript interface |
| `frontend/src/modelTypes/ITaskFromTemplate.ts` | TypeScript interface |
| `frontend/src/models/taskTemplate.ts` | Frontend model |
| `frontend/src/models/taskFromTemplate.ts` | Frontend model |
| `frontend/src/services/taskTemplateService.ts` | CRUD service |
| `frontend/src/services/taskFromTemplateService.ts` | Create-from-template service |
| `frontend/src/components/tasks/partials/CreateFromTemplateModal.vue` | Template picker + task creation modal |
| `frontend/src/components/tasks/partials/SaveAsTemplateModal.vue` | Save task as template modal |

### Modified Files (5) — same files as task-duplicate, updated

| File | Changes |
|------|---------|
| `pkg/routes/routes.go` | Added 6 template CRUD + create-from-template routes |
| `KanbanCard.vue` | Added "Save as Template" to card dropdown |
| `ProjectKanban.vue` | Added "From Template" header button, both template modals |
| `TaskDetailView.vue` | Added "Save as Template" sidebar button and modal |
| `en.json` | Added `task.template.*` and `task.detail.actions.saveAsTemplate` keys |

## Database

A new `task_templates` table is created automatically via migration:

| Column | Type | Description |
|--------|------|-------------|
| id | bigint PK | Auto-increment |
| title | varchar(250) | Template name |
| description | longtext | Template description |
| priority | bigint | Task priority |
| hex_color | varchar(6) | Color hex |
| percent_done | double | Progress |
| repeat_after | bigint | Repeat interval in seconds |
| repeat_mode | int | Repeat mode |
| label_ids | json | Array of label IDs to apply |
| owner_id | bigint INDEX | User who owns the template |
| created | timestamp | Created at |
| updated | timestamp | Updated at |

## Deployment

Same process as the task-duplicate feature:
1. Place `new-files` in your vikunja source directory
2. Run `patch-templates.ps1`
3. Export image and load on server
4. Update Portainer stack

**Note:** The modified files (routes.go, KanbanCard, etc.) include BOTH the task-duplicate and task-template changes. They replace the previous versions.

**Database migration runs automatically** on first startup — no manual steps needed.
