# Permission Dependencies Analysis

**Status**: COMPLETE  
**Created**: Based on T-PERM-001 task requirements  
**Purpose**: Map cross-entity permission checks to plan refactor order  
**Source**: Analysis of `/home/aron/projects/vikunja/pkg/models/*_permissions.go` files  

---

## Executive Summary

This document provides a complete dependency graph of all permission methods across the Vikunja codebase. The analysis reveals a **clean hierarchical structure with ZERO circular dependencies**, making migration straightforward following a bottom-up approach (foundation entities first, dependent entities later).

**Key Findings**:
- ✅ **No circular dependencies** - permission checks flow in one direction
- 📊 **20 permission files** with ~70+ permission methods
- 🏗️ **3 dependency levels** - clear migration order
- 🔑 **2 foundation entities** - Project and Label (no dependencies)
- 📦 **~15 helper functions** to migrate alongside permissions

---

## Dependency Hierarchy

### Level 0: Core Infrastructure (No Dependencies)

#### User (auth layer)
- **Location**: `pkg/user/` and auth layer
- **Permission Methods**: None (authentication, not authorization)
- **Notes**: Used BY all permission checks, not part of migration

---

### Level 1: Foundation Entities (Depend only on User)

These entities have NO dependencies on other entities for permission checks. They check:
- Owner/creator relationship
- Direct user permissions tables (ProjectUser, TeamMember)
- Team permissions tables (ProjectTeam)
- Link share auth type

#### 1.1 Project
**File**: `pkg/models/project_permissions.go`  
**Permission Methods**:
- `CanWrite(s, a)` - checks owner, ProjectUser table, ProjectTeam table, LinkSharing auth
- `CanRead(s, a)` - checks owner, ProjectUser table, ProjectTeam table, LinkSharing auth, SavedFilter special case
- `CanUpdate(s, a)` - calls `CanWrite`, plus checks parent project permissions if moving
- `CanDelete(s, a)` - calls `IsAdmin` (checks owner + ProjectUser/ProjectTeam admin permission)
- `CanCreate(s, a)` - checks parent project `CanWrite` if creating sub-project

**Helper Functions Used**:
- `GetProjectSimpleByID(s, projectID)` - loads project data
- `checkPermission(s, u, ...perms)` - queries ProjectUser and ProjectTeam tables
- `isOwner(u)` - checks `p.OwnerID == u.ID`
- `IsAdmin(s, a)` - checks owner or admin permission

**Database Tables Accessed**:
- `projects` (project data)
- `users_projects` (user permissions)
- `team_projects` (team permissions)
- `team_members` (team membership)

**Special Cases**:
- Favorites pseudo-project (ID = -1)
- Saved filter projects (negative IDs < -1)
- Archived project checks
- Parent project permission checks (sub-projects)

#### 1.2 Label
**File**: `pkg/models/label_permissions.go`  
**Permission Methods**:
- `CanUpdate(s, a)` - checks if user is label creator
- `CanDelete(s, a)` - checks if user is label creator  
- `CanRead(s, a)` - calls `hasAccessToLabel` (checks creator OR user has tasks with this label)
- `CanCreate(s, a)` - always true for authenticated users

**Helper Functions Used**:
- `getLabelByIDSimple(s, labelID)` - loads label data
- `hasAccessToLabel(s, a)` - checks creator OR task label access (performs JOIN query)

**Database Tables Accessed**:
- `labels` (label data)
- `label_tasks` (which tasks have this label)
- `tasks` (task data for permission checks)

**Notes**: 
- Labels have a unique permission model (creator-based + task-based access)
- `hasAccessToLabel` performs a complex JOIN to check if user has ANY task with this label

#### 1.3 Team
**File**: `pkg/models/teams_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - always true for authenticated users
- `CanUpdate(s, a)` - checks if user is team admin (via TeamMember table)
- `CanDelete(s, a)` - checks if user is team admin
- `CanRead(s, a)` - checks if user is team member (any permission level)

**Helper Functions Used**:
- `GetTeamByID(s, teamID)` - loads team data

**Database Tables Accessed**:
- `teams` (team data)
- `team_members` (membership and permissions)

#### 1.4 SavedFilter
**File**: `pkg/models/saved_filters_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - checks if user owns the saved filter
- `CanDelete(s, a)` - checks if user owns the saved filter
- `CanUpdate(s, a)` - checks if user owns the saved filter
- `CanCreate(s, a)` - always true for authenticated users

**Helper Functions Used**:
- `GetSavedFilterSimpleByID(s, filterID)` - loads saved filter data

**Database Tables Accessed**:
- `saved_filters` (filter data)

**Notes**: Simple ownership model (only creator can read/update/delete)

#### 1.5 APIToken
**File**: `pkg/models/api_tokens_permissions.go`  
**Permission Methods**:
- `CanDelete(s, a)` - checks if token belongs to user
- `CanCreate(s, a)` - always true for authenticated users

**Helper Functions Used**:
- `GetAPITokenByID(s, tokenID)` - loads token data
- `GetTokenFromTokenString(s, tokenString)` - loads token by string

**Database Tables Accessed**:
- `api_tokens` (token data)

---

### Level 2: Project-Dependent Entities

These entities check Project permissions as their primary authorization.

#### 2.1 Task
**File**: `pkg/models/tasks_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - loads task, then checks `Project.CanRead`
- `CanWrite(s, a)` - calls `canDoTask` (checks project permissions)
- `CanUpdate(s, a)` - calls `canDoTask`, plus checks new project permissions if moving
- `CanDelete(s, a)` - calls `canDoTask`
- `CanCreate(s, a)` - checks `Project.CanWrite`

**Helper Functions Used**:
- `GetTaskByIDSimple(s, taskID)` - loads task data
- `canDoTask(s, a)` - loads task, then checks `Project.CanWrite`

**Dependencies**:
- **→ Project**: ALL permission methods delegate to project permissions

**Database Tables Accessed**:
- `tasks` (task data)
- Then delegates to Project tables

**Special Cases**:
- Moving tasks between projects (checks permissions on BOTH projects)

#### 2.2 LinkSharing
**File**: `pkg/models/link_sharing_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - loads project, checks `Project.CanRead`
- `CanDelete(s, a)` - calls `canDoLinkShare`
- `CanUpdate(s, a)` - calls `canDoLinkShare`
- `CanCreate(s, a)` - calls `canDoLinkShare`

**Helper Functions Used**:
- `GetProjectSimpleByID(s, projectID)` - loads project data
- `GetProjectByShareHash(s, hash)` - loads project by share hash
- `canDoLinkShare(s, a)` - checks `Project.CanWrite` or `Project.IsAdmin` (if creating admin link)

**Dependencies**:
- **→ Project**: ALL methods check project permissions

**Database Tables Accessed**:
- `link_shares` (share data)
- Then delegates to Project tables

**Special Cases**:
- Link share users (negative user IDs) cannot create link shares
- Admin-level link shares require project admin permission

#### 2.3 ProjectView
**File**: `pkg/models/project_view_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - loads view, checks `Project.CanRead`
- `CanDelete(s, a)` - loads view, checks `Project.CanWrite`
- `CanUpdate(s, a)` - loads view, checks `Project.CanWrite`
- `CanCreate(s, a)` - checks `Project.CanWrite`

**Helper Functions Used**:
- `GetProjectViewByID(s, viewID)` - loads view data
- `GetProjectViewByIDAndProject(s, viewID, projectID)` - loads view for specific project

**Dependencies**:
- **→ Project**: ALL methods check project permissions

**Database Tables Accessed**:
- `project_views` (view configuration)
- Then delegates to Project tables

#### 2.4 ProjectUser
**File**: `pkg/models/project_users_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Project.CanWrite`
- `CanDelete(s, a)` - checks `Project.CanWrite`
- `CanUpdate(s, a)` - checks `Project.CanWrite`

**Dependencies**:
- **→ Project**: ALL methods check project permissions

**Database Tables Accessed**:
- `users_projects` (user permission assignments)
- Then delegates to Project tables

#### 2.5 ProjectTeam (TeamProject)
**File**: `pkg/models/project_team_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Project.CanWrite`
- `CanDelete(s, a)` - checks `Project.CanWrite`
- `CanUpdate(s, a)` - checks `Project.CanWrite`

**Dependencies**:
- **→ Project**: ALL methods check project permissions

**Database Tables Accessed**:
- `team_projects` (team permission assignments)
- Then delegates to Project tables

#### 2.6 Bucket (Kanban)
**File**: `pkg/models/kanban_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - loads bucket, checks `Project.CanWrite`
- `CanUpdate(s, a)` - loads bucket, checks `Project.CanWrite`
- `CanDelete(s, a)` - loads bucket, checks `Project.CanWrite`

**Helper Functions Used**:
- `getBucketByID(s, bucketID)` - loads bucket data

**Dependencies**:
- **→ Project**: ALL methods check project permissions

**Database Tables Accessed**:
- `buckets` (bucket/column data)
- Then delegates to Project tables

#### 2.7 TaskBucket
**File**: `pkg/models/kanban_task_bucket.go`  
**Permission Methods**:
- `CanUpdate(s, a)` - loads task, checks `Task.CanUpdate` (which checks project)

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- `tasks` (task data)
- Then delegates to Task/Project tables

#### 2.8 Webhook
**File**: `pkg/models/webhooks_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - checks `Project.CanRead`
- `CanDelete(s, a)` - checks `Project.IsAdmin`
- `CanUpdate(s, a)` - checks `Project.IsAdmin`
- `CanCreate(s, a)` - checks `Project.IsAdmin`

**Dependencies**:
- **→ Project**: ALL methods check project permissions (admin required for write operations)

**Database Tables Accessed**:
- `webhooks` (webhook configuration)
- Then delegates to Project tables

#### 2.9 Subscription (Project & Task types)
**File**: `pkg/models/subscription_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Project.CanRead` OR `Task.CanRead` (based on entity type)
- `CanDelete(s, a)` - checks if subscription belongs to user

**Dependencies**:
- **→ Project** (for project subscriptions)
- **→ Task** → Project (for task subscriptions)

**Database Tables Accessed**:
- `subscriptions` (subscription data)
- Then delegates based on entity type

**Special Cases**:
- Supports both project and task subscriptions
- Link share users cannot subscribe

#### 2.10 ProjectDuplicate
**File**: `pkg/models/project_duplicate.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Project.CanRead` on source project

**Dependencies**:
- **→ Project**: Checks read permission on project being duplicated

**Database Tables Accessed**:
- Then delegates to Project tables

#### 2.11 BulkTask
**File**: `pkg/models/bulk_task.go`  
**Permission Methods**:
- `CanUpdate(s, a)` - loads ALL tasks, checks `Project.CanWrite` for each task's project

**Dependencies**:
- **→ Task** → Project (for each task in bulk operation)

**Database Tables Accessed**:
- `tasks` (loads all task IDs)
- Then delegates to Project tables for each unique project

**Special Cases**:
- Must check permissions for ALL tasks in the bulk operation
- Different tasks may be in different projects

---

### Level 3: Task-Dependent Entities

These entities check Task permissions (which in turn check Project permissions).

#### 3.1 TaskComment
**File**: `pkg/models/task_comment_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - checks `Task.CanRead`
- `CanCreate(s, a)` - checks `Task.CanWrite`
- `CanUpdate(s, a)` - checks `Task.CanWrite` AND user is comment author
- `CanDelete(s, a)` - checks `Task.CanWrite` AND user is comment author

**Helper Functions Used**:
- `getTaskCommentSimple(s, tc)` - loads comment data

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- `task_comments` (comment data)
- Then delegates to Task tables

**Special Cases**:
- Update/Delete require BOTH task write permission AND being the comment author

#### 3.2 TaskAttachment
**File**: `pkg/models/task_attachment_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - checks `Task.CanRead`
- `CanCreate(s, a)` - checks `Task.CanWrite`
- `CanDelete(s, a)` - checks `Task.CanWrite`

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- `task_attachments` (attachment metadata)
- Then delegates to Task tables

#### 3.3 TaskRelation
**File**: `pkg/models/task_relation_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Task.CanWrite` on BOTH tasks (source and related)
- `CanDelete(s, a)` - checks `Task.CanWrite`

**Dependencies**:
- **→ Task** → Project (must check permissions on multiple tasks)

**Database Tables Accessed**:
- `task_relations` (relation data)
- Then delegates to Task tables

**Special Cases**:
- Creating relations requires write permission on BOTH tasks
- Tasks may be in different projects

#### 3.4 TaskAssignee
**File**: `pkg/models/task_assignees_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Task.CanWrite`
- `CanDelete(s, a)` - checks `Task.CanWrite`

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- `task_assignees` (assignee data)
- Then delegates to Task tables

#### 3.5 BulkAssignees
**File**: `pkg/models/task_assignees_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Task.CanWrite`

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- Same as TaskAssignee

#### 3.6 LabelTask
**File**: `pkg/models/label_task_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Label.hasAccessToLabel` AND `canDoLabelTask` (task write permission)
- `CanDelete(s, a)` - checks `canDoLabelTask` (task write permission)

**Helper Functions Used**:
- `getLabelByIDSimple(s, labelID)` - loads label data
- `canDoLabelTask(s, taskID, a)` - checks `Task.CanWrite`

**Dependencies**:
- **→ Label** (for CanCreate only - checks label access)
- **→ Task** → Project (for task write permission)

**Database Tables Accessed**:
- `label_tasks` (label-task associations)
- Then delegates to Label and Task tables

**Special Cases**:
- CanCreate requires access to BOTH label AND task
- CanDelete only requires task write permission

#### 3.7 LabelTaskBulk
**File**: `pkg/models/label_task_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Task.CanWrite`

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- Same as LabelTask

#### 3.8 TaskPosition
**File**: `pkg/models/task_position.go`  
**Permission Methods**:
- `CanUpdate(s, a)` - checks `Task.CanUpdate`

**Dependencies**:
- **→ Task** → Project

**Database Tables Accessed**:
- `tasks` (task position data)
- Then delegates to Task tables

#### 3.9 Reaction
**File**: `pkg/models/reaction_permissions.go`  
**Permission Methods**:
- `CanRead(s, a)` - loads entity, checks `Task.CanRead` or `TaskComment.CanRead`
- `CanCreate(s, a)` - loads entity, checks `Task.CanRead` or `TaskComment.CanRead`
- `CanDelete(s, a)` - checks user is reaction author

**Dependencies**:
- **→ Task** → Project (for task reactions)
- **→ TaskComment** → Task → Project (for comment reactions)

**Database Tables Accessed**:
- `reactions` (reaction data)
- Then delegates based on entity type

**Special Cases**:
- Supports reactions on both tasks and comments
- Entity type determines which permission check to use

---

### Level 4: Team-Dependent Entities

#### 4.1 TeamMember
**File**: `pkg/models/team_members_permissions.go`  
**Permission Methods**:
- `CanCreate(s, a)` - checks `Team.IsAdmin` (user is team admin)
- `CanDelete(s, a)` - checks `Team.IsAdmin` (user is team admin)
- `CanUpdate(s, a)` - checks `Team.IsAdmin` (user is team admin)

**Dependencies**:
- **→ Team**

**Database Tables Accessed**:
- `team_members` (membership data)
- Then delegates to Team tables

---

### Level 5: Special Cases

#### 5.1 DatabaseNotifications
**File**: `pkg/models/notifications_database.go`  
**Permission Methods**:
- `CanUpdate(s, a)` - checks if notification belongs to user

**Dependencies**: None (simple ownership check)

**Database Tables Accessed**:
- `notifications` (notification data)

**Notes**: Not part of core permission migration (notification-specific logic)

---

## Helper Functions Dependency Analysis

These helper functions are used by permission methods and also need migration to services.

### Project Helpers
**File**: `pkg/models/project.go`

1. **GetProjectSimpleByID(s, projectID)** → Service: `ProjectService.GetByID`
   - Used by: Project, Task, LinkSharing, ProjectView, Webhook, Subscription, ProjectDuplicate, BulkTask
   - Performs: Single project lookup by ID
   - Database: `projects` table

2. **GetProjectsMapByIDs(s, projectIDs)** → Service: `ProjectService.GetMapByIDs`
   - Used by: Various batch operations
   - Performs: Batch project lookup, returns map[id]*Project
   - Database: `projects` table

3. **GetProjectsByIDs(s, projectIDs)** → Service: `ProjectService.GetByIDs`
   - Used by: Various batch operations
   - Performs: Batch project lookup, returns []*Project
   - Database: `projects` table

4. **GetProjectSimpleByTaskID(s, taskID)** → Service: `TaskService.GetProject` or `ProjectService.GetByTaskID`
   - Used by: Task-related operations
   - Performs: Lookup project for a task
   - Database: JOIN `tasks` and `projects`

### Task Helpers
**File**: `pkg/models/tasks.go`

5. **GetTaskByIDSimple(s, taskID)** → Service: `TaskService.GetByID`
   - Used by: Task, TaskComment, TaskAttachment, TaskRelation, TaskAssignee, LabelTask, TaskPosition, Reaction
   - Performs: Single task lookup by ID
   - Database: `tasks` table

6. **GetTasksSimpleByIDs(s, taskIDs)** → Service: `TaskService.GetByIDs`
   - Used by: Bulk operations
   - Performs: Batch task lookup
   - Database: `tasks` table

### Label Helpers
**File**: `pkg/models/label.go`

7. **getLabelByIDSimple(s, labelID)** → Service: `LabelService.GetByID`
   - Used by: Label, LabelTask
   - Performs: Single label lookup by ID
   - Database: `labels` table

### LinkSharing Helpers
**File**: `pkg/models/link_sharing.go`

8. **GetLinkShareByID(s, linkShareID)** → Service: `LinkShareService.GetByID`
   - Used by: LinkSharing operations
   - Performs: Single link share lookup by ID
   - Database: `link_shares` table

9. **GetLinkSharesByIDs(s, linkShareIDs)** → Service: `LinkShareService.GetByIDs`
   - Used by: Batch operations
   - Performs: Batch link share lookup
   - Database: `link_shares` table

### Team Helpers
**File**: `pkg/models/teams.go`

10. **GetTeamByID(s, teamID)** → Service: `TeamService.GetByID`
    - Used by: Team, TeamMember operations
    - Performs: Single team lookup by ID
    - Database: `teams` table

### SavedFilter Helpers
**File**: `pkg/models/saved_filters.go`

11. **GetSavedFilterSimpleByID(s, filterID)** → Service: `SavedFilterService.GetByID`
    - Used by: Project.CanUpdate (for saved filter projects)
    - Performs: Single saved filter lookup by ID
    - Database: `saved_filters` table

### Bucket/Kanban Helpers
**File**: `pkg/models/kanban.go`

12. **getBucketByID(s, bucketID)** → Service: `KanbanService.GetBucketByID`
    - Used by: Bucket permission checks
    - Performs: Single bucket lookup by ID
    - Database: `buckets` table

### ProjectView Helpers
**File**: `pkg/models/project_view.go`

13. **GetProjectViewByID(s, viewID)** → Service: `ProjectViewService.GetByID`
    - Used by: ProjectView permission checks
    - Performs: Single view lookup by ID
    - Database: `project_views` table

14. **GetProjectViewByIDAndProject(s, viewID, projectID)** → Service: `ProjectViewService.GetByIDAndProject`
    - Used by: ProjectView permission checks
    - Performs: View lookup with project validation
    - Database: `project_views` table

### APIToken Helpers
**File**: `pkg/models/api_tokens.go`

15. **GetAPITokenByID(s, tokenID)** → Service: `APITokenService.GetByID`
    - Used by: APIToken permission checks
    - Performs: Single token lookup by ID
    - Database: `api_tokens` table

16. **GetTokenFromTokenString(s, tokenString)** → Service: `APITokenService.GetByTokenString`
    - Used by: Token authentication
    - Performs: Token lookup by string value
    - Database: `api_tokens` table

---

## Circular Dependency Analysis

**Result**: ✅ **ZERO CIRCULAR DEPENDENCIES FOUND**

Permission checks flow in a **strictly hierarchical direction**:
- Task → Project (never reverse)
- TaskComment → Task → Project (linear chain)
- LinkSharing → Project (one-way)
- LabelTask → Label + Task (two independent checks, no cycles)
- Subscription → Project OR Task (conditional, no cycles)

**Why No Cycles?**:
- Projects never check task permissions
- Tasks never check comment/attachment permissions
- Labels never check project/task permissions (except for access checks)
- All relationships are "child checks parent", never "parent checks child"

This makes migration straightforward: migrate foundation entities first, then entities that depend on them.

---

## Recommended Migration Order

Based on dependency analysis, tasks should be executed in this order:

### ✅ Phase A: Foundation (No Dependencies)
**Tasks**: T-PERM-006  
**Order**: Can run in parallel
1. Project permissions
2. Label permissions (independent)
3. Team permissions (independent)
4. SavedFilter permissions (independent)
5. APIToken permissions (independent)

**Rationale**: These have no cross-entity dependencies, only database lookups

---

### ✅ Phase B: Project-Dependent (Depend on Project Only)
**Tasks**: T-PERM-009, T-PERM-011  
**Prerequisites**: Project permissions migrated  
**Order**: Can run in parallel after Project is complete
1. LinkSharing
2. ProjectView
3. ProjectUser
4. ProjectTeam
5. Webhook
6. Bucket (Kanban)
7. ProjectDuplicate
8. Subscription (project subscriptions)

**Rationale**: All depend ONLY on Project permissions

---

### ✅ Phase C: Task (Depends on Project)
**Tasks**: T-PERM-007  
**Prerequisites**: Project permissions migrated  
**Order**: Must complete before Phase D
1. Task permissions

**Rationale**: Task depends on Project, and many entities depend on Task

---

### ✅ Phase D: Task-Dependent (Depend on Task → Project)
**Tasks**: T-PERM-010  
**Prerequisites**: Task permissions migrated  
**Order**: Can run in parallel after Task is complete
1. TaskComment
2. TaskAttachment
3. TaskRelation
4. TaskAssignee
5. BulkAssignees
6. LabelTask (depends on Label + Task)
7. TaskBucket
8. TaskPosition
9. Subscription (task subscriptions)
10. Reaction (task & comment reactions)
11. BulkTask

**Rationale**: All depend on Task permissions (which depend on Project)

---

### ✅ Phase E: Team-Dependent
**Tasks**: T-PERM-012  
**Prerequisites**: Team permissions migrated  
**Order**: After Team is complete
1. TeamMember

**Rationale**: Depends on Team permissions

---

## Special Case Handlers

### Pseudo-Projects
**Location**: `pkg/models/project_permissions.go`
- **Favorites Pseudo-Project** (ID = -1): Special read-only handling
- **Saved Filter Projects** (ID < -1): Delegate to SavedFilter permissions

**Migration Strategy**: Service layer should handle these special cases before normal permission checks

### Link Share Authentication
**Detection**: `_, ok := a.(*LinkSharing)`
**Behavior**: Many entities prevent link share users from creating/modifying entities
**Migration Strategy**: Service layer should check auth type early in permission methods

### Archived Projects
**Check**: `ErrProjectIsArchived{}`
**Behavior**: Archived projects block most write operations (except un-archiving)
**Migration Strategy**: Service layer should perform archived check in write methods

### Moving Tasks Between Projects
**Check**: `t.ProjectID != 0 && t.ProjectID != ot.ProjectID`
**Behavior**: Requires write permission on BOTH old and new projects
**Migration Strategy**: Service layer should check permissions on both projects during update

### Bulk Operations
**Entities**: BulkTask, BulkAssignees
**Behavior**: Must verify permissions for ALL entities in the batch
**Migration Strategy**: Service layer should collect all unique projects and verify permissions once per project

---

## Migration Validation Checklist

For each permission method migrated, verify:

- [ ] **Baseline test passes** - new service method produces identical results to old model method
- [ ] **All dependencies migrated** - helper functions and dependent permissions already in services
- [ ] **Special cases handled** - pseudo-projects, link shares, archived checks, etc.
- [ ] **Database tables accessed** - service has access to all required tables
- [ ] **Error handling preserved** - same error types returned
- [ ] **Performance maintained** - no additional database queries introduced
- [ ] **Cross-entity checks** - lazy initialization prevents circular dependencies

---

## Entity Permission Method Catalog

Complete list of all permission methods across 24 model files:

| Entity | File | Methods | Dependencies |
|--------|------|---------|--------------|
| **Project** | `project_permissions.go` | CanRead, CanWrite, CanUpdate, CanDelete, CanCreate | None |
| **Task** | `tasks_permissions.go` | CanRead, CanWrite, CanUpdate, CanDelete, CanCreate | Project |
| **Label** | `label_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | None |
| **Team** | `teams_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | None |
| **SavedFilter** | `saved_filters_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | None |
| **APIToken** | `api_tokens_permissions.go` | CanDelete, CanCreate | None |
| **LinkSharing** | `link_sharing_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | Project |
| **ProjectView** | `project_view_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | Project |
| **ProjectUser** | `project_users_permissions.go` | CanCreate, CanUpdate, CanDelete | Project |
| **ProjectTeam** | `project_team_permissions.go` | CanCreate, CanUpdate, CanDelete | Project |
| **Bucket** | `kanban_permissions.go` | CanCreate, CanUpdate, CanDelete | Project |
| **TaskBucket** | `kanban_task_bucket.go` | CanUpdate | Task |
| **Webhook** | `webhooks_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | Project |
| **Subscription** | `subscription_permissions.go` | CanCreate, CanDelete | Project OR Task |
| **ProjectDuplicate** | `project_duplicate.go` | CanCreate | Project |
| **BulkTask** | `bulk_task.go` | CanUpdate | Task |
| **TaskComment** | `task_comment_permissions.go` | CanRead, CanUpdate, CanDelete, CanCreate | Task |
| **TaskAttachment** | `task_attachment_permissions.go` | CanRead, CanDelete, CanCreate | Task |
| **TaskRelation** | `task_relation_permissions.go` | CanCreate, CanDelete | Task |
| **TaskAssignee** | `task_assignees_permissions.go` | CanCreate, CanDelete | Task |
| **BulkAssignees** | `task_assignees_permissions.go` | CanCreate | Task |
| **LabelTask** | `label_task_permissions.go` | CanCreate, CanDelete | Label + Task |
| **TaskPosition** | `task_position.go` | CanUpdate | Task |
| **Reaction** | `reaction_permissions.go` | CanRead, CanCreate, CanDelete | Task OR TaskComment |
| **TeamMember** | `team_members_permissions.go` | CanCreate, CanUpdate, CanDelete | Team |

**Total**: 24 entities, 70+ permission methods

---

## Conclusion

The permission system in Vikunja has a **clean, hierarchical architecture** that makes migration straightforward:

1. ✅ **No circular dependencies** - strictly tree-structured
2. ✅ **Clear levels** - 5 distinct dependency levels
3. ✅ **Simple patterns** - most entities delegate to Project or Task
4. ✅ **Bottom-up migration** - foundation first, dependents later
5. ✅ **Parallelizable** - many entities can be migrated simultaneously within each level

**Migration Complexity**: MODERATE (not HIGH)
- Clean architecture reduces risk
- Baseline tests provide safety net
- Dependency graph ensures correct order
- No circular dependencies to resolve

**Estimated Time per Entity**:
- Simple entities (owner-only checks): 0.5 days
- Medium entities (Project/Task delegates): 1 day
- Complex entities (multi-entity checks): 1.5 days

**Total Estimated Time**: 8-12 days (as per T-PERMISSIONS-PLAN.md)
