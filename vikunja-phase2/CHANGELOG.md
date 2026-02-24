# Vikunja Custom Build — Changelog

All notable changes to this custom Vikunja build are documented here.

## Phase 2f: Auto-Generated Tasks (2025-02-24)

### New Feature: Auto-Generated Task Templates
Task templates that automatically create task instances when they become due,
without cluttering the board with future tasks.

**How it works:**
- Create a template with title, interval, project, priority, labels, assignees
- System checks on page load and creates tasks that are due
- Only ONE open (undone) instance per template at any time
- If previous task isn't completed, it simply goes overdue — no pile-up
- Next due date recalculates from COMPLETION time, not creation time
- Pause/resume individual templates without deleting them
- Manual "Send to project now" button for immediate creation
- Generation log tracks every creation (system vs manual trigger)

**Backend:**
- `auto_task_template.go` — Model, CRUD, permissions (owner-only)
- `auto_task_create.go` — Check logic, trigger, completion handler
- `auto_task_handler.go` — API handlers for trigger and check endpoints
- Migration `20260224070000.go` — Tables: `auto_task_templates`, `auto_task_template_attachments`, `auto_task_log`; Column: `tasks.auto_template_id`

**API Endpoints:**
- `GET /api/v1/autotasks` — List user's templates
- `GET /api/v1/autotasks/:id` — Get one template (with log)
- `PUT /api/v1/autotasks` — Create template
- `POST /api/v1/autotasks/:id` — Update template
- `DELETE /api/v1/autotasks/:id` — Delete template
- `POST /api/v1/autotasks/:id/trigger` — Manually create task NOW
- `POST /api/v1/autotasks/check` — Check all templates and create due tasks

**Frontend:**
- `AutoTaskEditor.vue` — Full editor with card list, pause/resume, send now, generation log
- `autoTaskApi.ts` — HTTP client
- Third "Auto-Generated" tab in Templates page
- Auto-check trigger on Home page load
- 41 new i18n keys

**Known TODOs:**
- Backend cron goroutine for reliability without frontend trigger
- Hook `OnAutoTaskCompleted` into task update path when `done` changes
- Attachment copying from template to generated task
- Auto-gen indicator icon on task list items

---

## Phase 2e: Layout Consistency + Filters (2025-02-24)

### Consistent Page Layouts
All management pages now match the Templates page pattern:
- `content-widescreen` wrapper (900px centered)
- `<h2>` heading + grey description paragraph
- Horizontal separator line
- Action buttons below separator
- Standardized padding: `1.5rem 1rem`

**Pages updated:** Labels, Teams, Projects, Upcoming, Templates

### Task Row Layout
- Task title now left-aligned, project name pushed to right side
- Component: `SingleTaskInProject.vue`

### Home Page
- Current Tasks section now renders above Last Viewed
- Auto-task check on page mount

### Upcoming Page Improvements
- Checkbox state persists in localStorage across navigation
- "Assigned to me" filter on both Overview and Upcoming pages
- Filter bar visible on Overview (previously hidden)

---

## Phase 2d: Drag-to-Reorder (2025-02-24)

### Chain Step Reordering
- Drag handles on chain steps for visual reordering
- `useDragReorder.ts` composable with grab cursor, drop zones, animations
- Sequences auto-renumber on drop

---

## Phase 2c: Chain Enhancements (2025-02-24)

### Step Descriptions & Attachments
- Rich text description per chain step (collapsible)
- File attachments per step with upload/delete
- Backend: `TaskChainStepAttachment` model, file handler, migration

### Selectable Time Units
- Chain step offset and duration support hours, days, weeks, months
- Dropdown selectors replace hardcoded "(days)" labels
- Backend converts units to `time.Duration` for date calculation
- Default time unit stored in localStorage

### Gantt Improvements
- Dependency arrows between related tasks
- Bar tooltips with task details
- Cascade date updates on drag

---

## Phase 2b: Task Chains (2025-02-24)

### Chain Workflow System
- Define sequences of tasks with relative timing
- Create all tasks at once with calculated dates from anchor date
- Tasks linked via precedes/follows relations
- Title prefix support (e.g. "Batch #42 - ")
- Preview before creation

### Templates Page
- Tabbed interface: Templates | Chains | Auto-Generated
- Chain editor with step timeline preview
- Create-from-chain modal with project selector

---

## Phase 1: Task Templates & Duplication (2025-02-23)

### Task Templates
- Save any task as a reusable template
- Create tasks from templates with project selection
- Template management page at `/templates`
- Subproject inclusion toggle

### Task Duplication
- Duplicate tasks within the same project
- Copies all properties including labels and assignees

### Navigation
- Templates link in sidebar navigation
