# Vikunja Custom Build — Changelog

All notable changes to this custom Vikunja build are documented here.

## Phase 2g: Auto-Generated Tasks + Build Fixes (2025-02-24)

### New Feature: Auto-Generated Task Templates
Task templates that automatically create task instances when they become due,
without cluttering the board with future tasks.

**Core behavior:**
- Only ONE open (undone) instance per template at any time — no pile-up
- If the previous task isn't completed, it simply goes overdue naturally
- Next due date recalculates from COMPLETION time, not creation time
- Pause/resume individual templates without deleting them
- Manual "Send to project now" button for immediate creation
- Generation log tracks every creation (system vs manual trigger)

**Backend files:**
- `auto_task_template.go` — Model, CRUD, permissions (owner-only)
- `auto_task_create.go` — Check logic, trigger, completion handler
- `auto_task_handler.go` — API handlers for trigger and check endpoints
- Migration `20260224070000.go` — 3 tables + `tasks.auto_template_id` column

**API Endpoints:**
- `GET    /api/v1/autotasks` — List user's templates (with generation log)
- `GET    /api/v1/autotasks/:id` — Get one template
- `PUT    /api/v1/autotasks` — Create template
- `POST   /api/v1/autotasks/:id` — Update template
- `DELETE /api/v1/autotasks/:id` — Delete template + log
- `POST   /api/v1/autotasks/:id/trigger` — Manually create task NOW
- `POST   /api/v1/autotasks/check` — Check and create due tasks

**Frontend:**
- `AutoTaskEditor.vue` — Card-based editor with pause/resume, send now, generation log
- `autoTaskApi.ts` — HTTP client
- Third "Auto-Generated" tab in Templates page (robot icon)
- Auto-check trigger on Home page load
- 41 new i18n keys

### Build Compatibility Fixes
- All handler files updated to echo v5 (`github.com/labstack/echo/v5`)
- Handler function signatures use `c *echo.Context` (pointer, v5 style)
- `echo.NewHTTPError` always called with two args `(statusCode, message)`
- Auth retrieved via `auth2.GetAuthFromClaims(c)` from `pkg/modules/auth`
- `auto_template_id` set via raw SQL (not added to Task struct)
- `createTask` called with correct 5-arg signature `(s, task, auth, false, false)`
- `TaskAssginee` uses Vikunja's original spelling (typo in upstream)
- `ReadAll` returns `(interface{}, int, int64, error)` matching CObject interface
- `PermissionAdmin` used instead of non-existent `web.RightAdmin`

### Default Duration Unit Setting
- ChainEditor now uses `useStorage('chainDefaultTimeUnit', 'days')`
- New chain steps default to the user's preferred unit
- Persists in localStorage

### Known TODOs
- Backend cron goroutine for auto-task reliability without frontend trigger
- Hook `OnAutoTaskCompleted` into task update path when `done` changes
- Attachment copying from template to generated task
- Auto-gen indicator icon on task list items

---

## Phase 2f: Time Units + Filters (2025-02-24)

### Selectable Time Units for Chain Steps
- Chain step offset and duration support hours, days, weeks, months
- Dropdown selectors replace hardcoded "(days)" labels
- Backend converts units to `time.Duration` for date calculation
- Migration `20260224060000.go` adds `offset_unit`, `duration_unit` columns

### Task Row Layout
- Task title left-aligned, project name pushed to right side

### Assigned-to-Me Filter
- "Assigned to me" checkbox on both Overview and Upcoming pages
- Filter bar visible on Overview (previously hidden)
- Persists in localStorage

### Templates Page Tabs Spacing
- Reduced gap between description and tabs

---

## Phase 2e: Layout Consistency (2025-02-24)

### Consistent Page Layouts
All management pages match the Templates page pattern:
- `content-widescreen` wrapper (900px centered)
- `<h2>` heading + grey description paragraph, standardized padding

**Pages updated:** Labels, Teams, Projects, Upcoming, Templates

### Home Page
- Current Tasks section renders above Last Viewed

### Upcoming Page Improvements
- Checkbox state persists in localStorage

---

## Phase 2d: Drag-to-Reorder (2025-02-24)

### Chain Step Reordering
- Drag handles with visual reordering via `useDragReorder.ts` composable

---

## Phase 2c: Chain Enhancements (2025-02-24)

### Step Descriptions & Attachments
- Rich text description per chain step (collapsible)
- File attachments per step with upload/delete

### Gantt Improvements
- Dependency arrows between related tasks
- Bar tooltips with task details

---

## Phase 2b: Task Chains (2025-02-24)

### Chain Workflow System
- Define sequences of tasks with relative timing
- Create all tasks at once from anchor date
- Tasks linked via precedes/follows relations

---

## Phase 1: Task Templates & Duplication (2025-02-23)

### Task Templates
- Save any task as a reusable template
- Create tasks from templates with project selection
- Template management page at `/templates`

### Task Duplication
- Duplicate tasks within the same project
