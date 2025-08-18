# Vikunja API v2 - Task List

This file contains a list of tasks to be completed for the Vikunja API v2 project.

See API_V2_PRD.md for design rules.

## I. Core Resource Implementation

### 1. Projects (`/api/v2/projects`)

- [ ] `GET /api/v2/projects`: Implement pagination, filtering by `is_archived`, and search by title/description.
- [ ] `POST /api/v2/projects`: Implement project creation, including handling of parent projects and default views.
- [ ] `GET /api/v2/projects/{id}`: Implement retrieval of a single project, including its owner and views.
- [ ] `PUT /api/v2/projects/{id}`: Implement project updates, including archiving and background image handling.
- [ ] `DELETE /api/v2/projects/{id}`: Implement project deletion, including all associated tasks and other resources.

### 2. Tasks (`/api/v2/tasks`)

- [ ] `GET /api/v2/projects/{id}/tasks`: Implement retrieval of all tasks for a project, with pagination and filtering.
- [ ] `POST /api/v2/projects/{id}/tasks`: Implement task creation, including setting assignees, labels, and reminders.
- [ ] `GET /api/v2/tasks/{id}`: Implement retrieval of a single task, including all its details.
- [ ] `PUT /api/v2/tasks/{id}`: Implement task updates, including marking as done, repeating tasks, and updating assignees/labels.
- [ ] `DELETE /api/v2/tasks/{id}`: Implement task deletion.

### 3. Labels (`/api/v2/labels`)

- [ ] `GET /api/v2/labels`: Implement retrieval of all labels for the current user.
- [ ] `POST /api/v2/labels`: Implement label creation.
- [ ] `GET /api/v2/labels/{id}`: Implement retrieval of a single label.
- [ ] `PUT /api/v2/labels/{id}`: Implement label updates.
- [ ] `DELETE /api/v2/labels/{id}`: Implement label deletion.

### 4. Teams (`/api/v2/teams`)

- [ ] `GET /api/v2/teams`: Implement retrieval of all teams for the current user.
- [ ] `POST /api/v2/teams`: Implement team creation.
- [ ] `GET /api/v2/teams/{id}`: Implement retrieval of a single team, including its members.
- [ ] `PUT /api/v2/teams/{id}`: Implement team updates.
- [ ] `DELETE /api/v2/teams/{id}`: Implement team deletion.
- [ ] `GET /api/v2/teams/{id}/members`: Implement retrieval of all members for a team.
- [ ] `POST /api/v2/teams/{id}/members`: Add a member to a team.
- [ ] `DELETE /api/v2/teams/{id}/members/{userId}`: Remove a member from a team.

## II. Remaining Endpoint Analysis and Implementation

- [ ] **Notifications:** Analyze and implement v2 endpoints for notifications.
- [ ] **Link Sharing:** Analyze and implement v2 endpoints for link sharing.
- [ ] **Task Attachments:** Analyze and implement v2 endpoints for task attachments.
- [ ] **Task Comments:** Analyze and implement v2 endpoints for task comments.
- [ ] **Task Relations:** Analyze and implement v2 endpoints for task relations.
- [ ] **Saved Filters:** Analyze and implement v2 endpoints for saved filters.
- [ ] **Webhooks:** Analyze and implement v2 endpoints for webhooks.
- [ ] **Reactions:** Analyze and implement v2 endpoints for reactions.
- [ ] **Project Views:** Analyze and implement v2 endpoints for project views.
- [ ] **Kanban Buckets:** Analyze and implement v2 endpoints for Kanban buckets.

## III. Testing

- [ ] **Unit Tests:** Write unit tests for all new v2 handlers and models.
- [ ] **Integration Tests:** Write integration tests to ensure the v2 API works correctly as a whole.
- [ ] **Regression Testing:** Run the full test suite to catch any regressions in the v1 API.
