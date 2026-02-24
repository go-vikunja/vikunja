# Patch File Manifest

Every file in this patch directory and where it goes in the Vikunja source tree.
Run `patch-phase2.ps1` to apply all files and build the Docker image.

## Backend — Go Models (pkg/models/)

| Patch File | Target | Description |
|-----------|--------|-------------|
| task_chain.go | pkg/models/ | Chain template model, step struct with time units |
| task_from_chain.go | pkg/models/ | Create tasks from chain with unit-aware date math |
| auto_task_template.go | pkg/models/ | Auto-task template model, log struct, CRUD, permissions |
| auto_task_create.go | pkg/models/ | Auto-creation logic, trigger, completion handler |

## Backend — Go Handlers (pkg/routes/api/v1/)

| Patch File | Target | Description |
|-----------|--------|-------------|
| chain_step_attachment.go | pkg/routes/api/v1/ | Upload/delete chain step attachments (echo v5) |
| auto_task_handler.go | pkg/routes/api/v1/ | Auto-task trigger and check endpoints (echo v5) |

## Backend — Routes (pkg/routes/)

| Patch File | Target | Description |
|-----------|--------|-------------|
| routes.go | pkg/routes/ | All API endpoint registrations |

## Migrations (pkg/migration/)

| Patch File | ID | Description |
|-----------|-----|-------------|
| 20260224050000.go | 20260224050000 | task_chain_step_attachments table |
| 20260224060000.go | 20260224060000 | offset_unit, duration_unit on task_chain_steps |
| 20260224070000.go | 20260224070000 | auto_task_templates, auto_task_log, auto_task_template_attachments; tasks.auto_template_id |

## Frontend — Services (frontend/src/services/)

| Patch File | Target | Description |
|-----------|--------|-------------|
| taskChainApi.ts | services/ | Chain API client + TimeUnit types + conversions |
| autoTaskApi.ts | services/ | Auto-task API client |

## Frontend — Components (frontend/src/components/)

| Patch File | Target | Description |
|-----------|--------|-------------|
| ChainEditor.vue | tasks/partials/ | Chain editor with time unit dropdowns, default unit pref |
| CreateFromChainModal.vue | tasks/partials/ | Modal for creating tasks from chain template |
| AutoTaskEditor.vue | tasks/partials/ | Auto-task template editor and card list |
| SingleTaskInProject.vue | tasks/partials/ | Task row: title left, project right |
| SubprojectFilter.vue | project/partials/ | Subproject filter fix |
| GanttDependencyArrows.vue | gantt/ | Dependency arrows between gantt bars |
| GanttChart.vue | gantt/ | Gantt chart with arrow integration |
| GanttRowBars.vue | gantt/ | Gantt bars with tooltips |

## Frontend — Views (frontend/src/views/)

| Patch File | Target | Description |
|-----------|--------|-------------|
| ListTemplates.vue | templates/ | 3-tab template manager (Templates, Chains, Auto-Generated) |
| ListLabels.vue | labels/ | Modernized layout |
| ListTeams.vue | teams/ | Modernized layout |
| ListProjects.vue | project/ | Modernized layout |
| ShowTasks.vue | tasks/ | Filters, checkbox persistence, assigned-to-me |
| Home.vue | . | Tasks above last viewed, auto-task check on mount |

## Frontend — Stores & Composables

| Patch File | Target | Description |
|-----------|--------|-------------|
| tasks.ts | stores/ | Task store with cascade updates |
| useGanttTaskList.ts | views/project/helpers/ | Gantt task list helper |
| useDragReorder.ts | composables/ | Drag-to-reorder composable |

## Frontend — i18n

| Patch File | Target | Description |
|-----------|--------|-------------|
| en.json | i18n/lang/ | All translation keys |

## Documentation

| File | Description |
|------|-------------|
| CHANGELOG.md | Full changelog for all phases |
| docs/AUTO_TASKS.md | Auto-generated tasks technical architecture |
| docs/PATCH_MANIFEST.md | This file |
