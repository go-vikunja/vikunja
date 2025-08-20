# Vikunja API v2 - Task List

This document outlines the tasks required to implement the Vikunja API v2. The goal is to create a modern, RESTful API with full feature parity with v1. For a detailed overview of the API design and endpoint definitions, please refer to the [API_V2_PRD.md](API_V2_PRD.md).

## API Versioning Strategy

All new v2 endpoints should be implemented under the `/api/v2` path. When a new set of v2 endpoints is implemented, the frontend client should also be updated to use these new endpoints. This ensures a clear separation between v1 and v2 of the API and allows for a gradual migration.

## How to Work on This Task List

Each task in this list corresponds to a single API endpoint or a group of related endpoints. When working on a task, please follow these steps:

1.  **Implement the Endpoint:** Create the necessary route, handler, and any required business logic in the `pkg/` directory.
2.  **Update the Frontend:** Update the frontend client in the `frontend/` directory to use the new v2 endpoint.
3.  **Write Tests:** Add unit and integration tests for the new endpoint. Ensure that all tests pass.
4.  **Ensure Feature Parity:** Refer to the corresponding v1 endpoint(s) to ensure that all functionality is replicated.
5.  **Update this Document:** Once the endpoint is fully implemented and tested, mark the corresponding task as completed by checking the box.

## I. Projects

### 1.1. `GET /api/v2/projects`

*   **Description:** Retrieve all projects for the current user.
*   **V1 Equivalent:** `GET /projects`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Implement pagination.
    *   Allow filtering by `is_archived`.
    *   Allow searching by title and description.
*   **Tests:**
    *   Unit test for the handler.
    *   Integration test to verify the endpoint works as expected.
    *   Integration test to verify filtering and pagination.

### 1.2. `POST /api/v2/projects`

*   **Description:** Create a new project.
*   **V1 Equivalent:** `PUT /projects`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Handle creation of parent projects.
    *   Set default views upon creation.
*   **Tests:**
    *   Unit test for the handler.
    *   Integration test to verify project creation.

### 1.3. `GET /api/v2/projects/{id}`

*   **Description:** Retrieve a single project by its ID.
*   **V1 Equivalent:** `GET /projects/:project`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Include the project owner and views in the response.
*   **Tests:**
    *   Unit test for the handler.
    *   Integration test to verify retrieval of a single project.

### 1.4. `PUT /api/v2/projects/{id}`

*   **Description:** Update a project.
*   **V1 Equivalent:** `POST /projects/:project`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Support archiving and unarchiving.
    *   Handle updates to the background image.
*   **Tests:**
    *   Unit test for the handler.
    *   Integration test to verify project updates.

### 1.5. `DELETE /api/v2/projects/{id}`

*   **Description:** Delete a project.
*   **V1 Equivalent:** `DELETE /projects/:project`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Ensure all associated tasks and other resources are deleted.
*   **Tests:**
    *   Unit test for the handler.
    *   Integration test to verify project deletion.

### 1.6. `POST /api/v2/projects/{id}/duplicate`

*   **Description:** Duplicate a project.
*   **V1 Equivalent:** `PUT /projects/:projectid/duplicate`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Tests:**
    *   Unit test for the handler.
    *   Integration test to verify project duplication.

### 1.7. Project Users & Teams

*   **Description:** Manage user and team associations for a project.
*   **V1 Equivalents:** `GET /projects/:project/projectusers`, `PUT /projects/:project/users`, etc.
*   **Tasks:**
    *   [x] `GET /api/v2/projects/{id}/users`
    *   [x] `POST /api/v2/projects/{id}/users`
    *   [x] `PUT /api/v2/projects/{id}/users/{userId}`
    *   [x] `DELETE /api/v2/projects/{id}/users/{userId}`
    *   [x] Update the frontend client to use these endpoints.
    *   [x] `GET /api/v2/projects/{id}/teams`
    *   [x] `POST /api/v2/projects/{id}/teams`
    *   [x] `PUT /api/v2/projects/{id}/teams/{teamId}`
    *   [x] `DELETE /api/v2/projects/{id}/teams/{teamId}`
    *   [x] Update the frontend client to use these endpoints.
*   **Tests:**
    *   Unit and integration tests for each endpoint.

## II. Tasks

### 2.1. `GET /api/v2/tasks`

*   **Description:** Retrieve all tasks for the current user.
*   **V1 Equivalent:** `GET /tasks/all`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Implement pagination and filtering.
*   **Tests:**
    *   Unit and integration tests.

### 2.2. `GET /api/v2/projects/{id}/tasks`

*   **Description:** Retrieve all tasks for a project.
*   **V1 Equivalent:** `GET /projects/:project/tasks`, `GET /projects/:project/views/:view/tasks`
*   **Tasks:**
    *   [x] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Implement pagination and filtering.
*   **Tests:**
    *   Unit and integration tests.

### 2.3. `POST /api/v2/projects/{id}/tasks`

*   **Description:** Create a new task in a project.
*   **V1 Equivalent:** `PUT /projects/:project/tasks`
*   **Tasks:**
    *   [ ] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Handle setting assignees, labels, and reminders.
*   **Tests:**
    *   Unit and integration tests.

### 2.4. `GET /api/v2/tasks/{id}`

*   **Description:** Retrieve a single task by its ID.
*   **V1 Equivalent:** `GET /tasks/:projecttask`
*   **Tasks:**
    *   [ ] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Tests:**
    *   Unit and integration tests.

### 2.5. `PUT /api/v2/tasks/{id}`

*   **Description:** Update a task.
*   *   **V1 Equivalent:** `POST /tasks/:projecttask`
*   **Tasks:**
    *   [ ] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Requirements:**
    *   Support marking as done, repeating tasks, and updating assignees/labels.
*   **Tests:**
    *   Unit and integration tests.

### 2.6. `DELETE /api/v2/tasks/{id}`

*   **Description:** Delete a task.
*   **V1 Equivalent:** `DELETE /tasks/:projecttask`
*   **Tasks:**
    *   [ ] Implement the backend endpoint.
    *   [ ] Update the frontend client to use this endpoint.
*   **Tests:**
    *   Unit and integration tests.

### 2.7. Task Position & Bulk Operations

*   **Description:** Manage task positions and perform bulk updates.
*   **V1 Equivalents:** `POST /tasks/:task/position`, `POST /tasks/bulk`
*   **Tasks:**
    *   [ ] `PUT /api/v2/tasks/{id}/position`
    *   [ ] `POST /api/v2/tasks/bulk`
    *   [ ] Update the frontend client to use these endpoints.
*   **Tests:**
    *   Unit and integration tests for each endpoint.

### 2.8. Task Assignees, Labels, & Relations

*   **Description:** Manage assignees, labels, and relations for a task.
*   **Tasks:**
    *   [ ] `GET /api/v2/tasks/{id}/assignees`
    *   [ ] `POST /api/v2/tasks/{id}/assignees`
    *   [ ] `DELETE /api/v2/tasks/{id}/assignees/{userId}`
    *   [ ] `POST /api/v2/tasks/{id}/assignees/bulk`
    *   [ ] `GET /api/v2/tasks/{id}/labels`
    *   [ ] `POST /api/v2/tasks/{id}/labels`
    *   [ ] `DELETE /api/v2/tasks/{id}/labels/{labelId}`
    *   [ ] `POST /api/v2/tasks/{id}/labels/bulk`
    *   [ ] `GET /api/v2/tasks/{id}/relations`
    *   [ ] `POST /api/v2/tasks/{id}/relations`
    *   [ ] `DELETE /api/v2/tasks/{id}/relations/{relationId}`
    *   [ ] Update the frontend client to use these endpoints.
*   **Tests:**
    *   Unit and integration tests for each endpoint.

## III. Other Resources

This section covers the remaining resources that need to be implemented for full feature parity. For each resource, a separate set of tasks should be created, similar to the ones for Projects and Tasks above.

*   **[ ] Labels:**
    *   [ ] Implement `GET /api/v2/labels`
    *   [ ] Implement `POST /api/v2/labels`
    *   [ ] Implement `GET /api/v2/labels/{id}`
    *   [ ] Implement `PUT /api/v2/labels/{id}`
    *   [ ] Implement `DELETE /api/v2/labels/{id}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Teams:**
    *   [ ] Implement `GET /api/v2/teams`
    *   [ ] Implement `POST /api/v2/teams`
    *   [ ] Implement `GET /api/v2/teams/{id}`
    *   [ ] Implement `PUT /api/v2/teams/{id}`
    *   [ ] Implement `DELETE /api/v2/teams/{id}`
    *   [ ] Implement `GET /api/v2/teams/{id}/members`
    *   [ ] Implement `POST /api/v2/teams/{id}/members`
    *   [ ] Implement `PUT /api/v2/teams/{id}/members/{userId}`
    *   [ ] Implement `DELETE /api/v2/teams/{id}/members/{userId}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Notifications:**
    *   [ ] Implement `GET /api/v2/notifications`
    *   [ ] Implement `PUT /api/v2/notifications/{id}`
    *   [ ] Implement `PUT /api/v2/notifications/read`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Link Sharing:**
    *   [ ] Implement `GET /api/v2/projects/{id}/shares`
    *   [ ] Implement `POST /api/v2/projects/{id}/shares`
    *   [ ] Implement `GET /api/v2/shares/{token}`
    *   [ ] Implement `DELETE /api/v2/shares/{token}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Attachments:**
    *   [ ] Implement `GET /api/v2/tasks/{id}/attachments`
    *   [ ] Implement `POST /api/v2/tasks/{id}/attachments`
    *   [ ] Implement `GET /api/v2/attachments/{id}`
    *   [ ] Implement `DELETE /api/v2/attachments/{id}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Comments:**
    *   [ ] Implement `GET /api/v2/tasks/{id}/comments`
    *   [ ] Implement `POST /api/v2/tasks/{id}/comments`
    *   [ ] Implement `GET /api/v2/comments/{id}`
    *   [ ] Implement `PUT /api/v2/comments/{id}`
    *   [ ] Implement `DELETE /api/v2/comments/{id}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Saved Filters:**
    *   [ ] Implement `GET /api/v2/filters`
    *   [ ] Implement `POST /api/v2/filters`
    *   [ ] Implement `GET /api/v2/filters/{id}`
    *   [ ] Implement `PUT /api/v2/filters/{id}`
    *   [ ] Implement `DELETE /api/v2/filters/{id}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Webhooks:**
    *   [ ] Implement `GET /api/v2/projects/{id}/webhooks`
    *   [ ] Implement `POST /api/v2/projects/{id}/webhooks`
    *   [ ] Implement `PUT /api/v2/webhooks/{id}`
    *   [ ] Implement `DELETE /api/v2/webhooks/{id}`
    *   [ ] Implement `GET /api/v2/webhooks/events`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Reactions:**
    *   [ ] Implement `GET /api/v2/{entityType}/{entityId}/reactions`
    *   [ ] Implement `POST /api/v2/{entityType}/{entityId}/reactions`
    *   [ ] Implement `DELETE /api/v2/{entityType}/{entityId}/reactions/{reactionId}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Project Views:**
    *   [ ] Implement `GET /api/v2/projects/{id}/views`
    *   [ ] Implement `POST /api/v2/projects/{id}/views`
    *   [ ] Implement `GET /api/v2/views/{id}`
    *   [ ] Implement `PUT /api/v2/views/{id}`
    *   [ ] Implement `DELETE /api/v2/views/{id}`
    *   [ ] Update the frontend client to use these endpoints.
*   **[ ] Kanban Buckets:**
    *   [ ] Implement `GET /api/v2/views/{id}/buckets`
    *   [ ] Implement `POST /api/v2/views/{id}/buckets`
    *   [ ] Implement `PUT /api/v2/buckets/{id}`
    *   [ ] Implement `DELETE /api/v2/buckets/{id}`
    *   [ ] Implement `POST /api/v2/buckets/{id}/tasks`
    *   [ ] Update the frontend client to use these endpoints.

## IV. Testing

*   **[ ] Unit Tests:** Ensure all new v2 handlers and models have comprehensive unit tests.
*   **[ ] Integration Tests:** Write integration tests to ensure the v2 API works correctly as a whole.
*   **[ ] Regression Testing:** Run the full test suite to catch any regressions in the v1 API.

## V. User & Info

*   **[ ] Info:**
    *   [ ] Implement `GET /api/v2/info`
    *   [ ] Update the frontend client to use this endpoint.
*   **[ ] Timezones:**
    *   [ ] Implement `GET /api/v2/timezones`
    *   [ ] Update the frontend client to use this endpoint.
*   **[ ] Avatars:**
    *   [ ] Implement `GET /api/v2/users/{username}/avatar`
    *   [ ] Update the frontend client to use this endpoint.

## VI. Subscriptions

*   **[ ] Subscriptions:**
    *   [ ] Implement `POST /api/v2/{entityType}/{entityId}/subscription`
    *   [ ] Implement `DELETE /api/v2/{entityType}/{entityId}/subscription`
    *   [ ] Update the frontend client to use these endpoints.
