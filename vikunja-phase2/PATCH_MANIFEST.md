# Patch File Manifest

Every file in this patch directory and where it goes in the Vikunja source tree.
Run `patch-phase2.ps1` to apply all files automatically.

## Backend — Go (pkg/)

| Patch File | Target Path | Description |
|-----------|-------------|-------------|
| task_chain.go | pkg/models/task_chain.go | Chain template model, step struct with time units |
| task_from_chain.go | pkg/models/task_from_chain.go | Create tasks from chain with unit-aware date math |
| auto_task_template.go | pkg/models/auto_task_template.go | Auto-task template model, log, CRUD, permissions |
| auto_task_create.go | pkg/models/auto_task_create.go | Auto-creation logic, trigger, completion handler |
| chain_step_attachment.go | pkg/routes/api/v1/chain_step_attachment.go | Upload/delete chain step attachments |
| auto_task_handler.go | pkg/routes/api/v1/auto_task_handler.go | Auto-task trigger and check HTTP handlers |
| routes.go | pkg/routes/routes.go | All API endpoint registrations |

## Migrations (pkg/migration/)

| Patch File | Migration ID | Description |
|-----------|-------------|-------------|
| 20260224050000.go | 20260224050000 | task_chain_step_attachments table |
| 20260224060000.go | 20260224060000 | offset_unit, duration_unit columns on task_chain_steps |
| 20260224070000.go | 20260224070000 | auto_task_templates, auto_task_log, auto_task_template_attachments tables; tasks.auto_template_id column |

## Frontend — Services (frontend/src/services/)

| Patch File | Target Path | Description |
|-----------|-------------|-------------|
| taskChainApi.ts | frontend/src/services/taskChainApi.ts | Chain API client + TimeUnit type + conversion helpers |
| autoTaskApi.ts | frontend/src/services/autoTaskApi.ts | Auto-task API client |

## Frontend — Components (frontend/src/components/)

| Patch File | Target Path | Description |
|-----------|-------------|-------------|
| ChainEditor.vue | .../tasks/partials/ChainEditor.vue | Chain editor with time unit dropdowns, default unit pref |
| CreateFromChainModal.vue | .../tasks/partials/CreateFromChainModal.vue | Modal for creating tasks from chain |
| AutoTaskEditor.vue | .../tasks/partials/AutoTaskEditor.vue | Auto-task template editor and card list |
| SingleTaskInProject.vue | .../tasks/partials/SingleTaskInProject.vue | Task row: title left, project right |
| SubprojectFilter.vue | .../project/partials/SubprojectFilter.vue | Subproject filter fix |
| GanttDependencyArrows.vue | .../gantt/GanttDependencyArrows.vue | Dependency arrows between gantt bars |
| GanttChart.vue | .../gantt/GanttChart.vue | Gantt chart with arrow integration |
| GanttRowBars.vue | .../gantt/GanttRowBars.vue | Gantt bars with tooltips |

## Frontend — Views (frontend/src/views/)

| Patch File | Target Path | Description |
|-----------|-------------|-------------|
| ListTemplates.vue | .../views/templates/ListTemplates.vue | 3-tab template manager (Templates, Chains, Auto-Generated) |
| ListLabels.vue | .../views/labels/ListLabels.vue | Modernized layout |
| ListTeams.vue | .../views/teams/ListTeams.vue | Modernized layout |
| ListProjects.vue | .../views/project/ListProjects.vue | Modernized layout |
| ShowTasks.vue | .../views/tasks/ShowTasks.vue | Filters, checkbox persistence, assigned-to-me |
| Home.vue | .../views/Home.vue | Tasks above last viewed, auto-task check on mount |

## Frontend — Stores & Composables

| Patch File | Target Path | Description |
|-----------|-------------|-------------|
| tasks.ts | frontend/src/stores/tasks.ts | Task store with cascade updates |
| useGanttTaskList.ts | .../views/project/helpers/useGanttTaskList.ts | Gantt task list helper |
| useDragReorder.ts | frontend/src/composables/useDragReorder.ts | Drag-to-reorder composable |

## Frontend — i18n

| Patch File | Target Path | Description |
|-----------|-------------|-------------|
| en.json | frontend/src/i18n/lang/en.json | All translation keys |

## Documentation

| File | Description |
|------|-------------|
| CHANGELOG.md | Full changelog for all phases |
| docs/AUTO_TASKS.md | Auto-generated tasks technical architecture |
| docs/PATCH_MANIFEST.md | This file |
