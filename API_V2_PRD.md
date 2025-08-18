# Vikunja API v2 - Product Requirements Document

## 1. Introduction

This document outlines the requirements for the new Vikunja API v2. The goal of this project is to modernize and standardize the Vikunja API, making it more robust, easier to use, and more suitable for consumption by other services, including AI agents.

## 2. Goals

*   **Standardization:** The new API will adhere to modern RESTful principles, including consistent endpoint naming, use of appropriate HTTP verbs, and standardized request and response formats.
*   **Versioning:** The new API will be versioned to allow for future changes without breaking existing integrations.
*   **Discoverability:** The API will be self-discoverable through the use of HATEOAS links.
*   **Usability:** The API will be designed to be easy to use and understand, with clear and consistent documentation.
*   **Feature Parity:** The v2 API will, at a minimum, provide all the functionality of the v1 API.

## 3. Non-Goals

*   **New Features:** This project is focused on modernizing the existing API, not adding new features.
*   **Deprecation of v1:** The v1 API will continue to be supported for the foreseeable future.

## 4. Requirements

### 4.1. API Versioning

*   The new API will be available under the `/api/v2/` path.
*   The existing API will remain available under the `/api/v1/` path.

### 4.2. RESTful Principles

*   **Endpoint Naming:** Endpoints will use plural nouns for resources (e.g., `/api/v2/projects`).
*   **HTTP Verbs:** The API will use the following HTTP verbs:
    *   `GET`: Retrieve a resource or a collection of resources.
    *   `POST`: Create a new resource.
    *   `PUT`: Update an existing resource.
    *   `DELETE`: Delete a resource.
*   **Status Codes:** The API will use standard HTTP status codes to indicate the success or failure of a request.

### 4.3. HATEOAS

*   API responses will include `_links` to related resources.
*   This will allow clients to discover the API without prior knowledge of the endpoint structure.

### 4.4. Standardized Models

*   All API models will have consistent naming conventions.
*   Error responses will be standardized to include a machine-readable error code and a human-readable error message.

## 5. Endpoint Definitions

This section provides a detailed overview of the planned v2 API endpoints, ensuring feature parity with the v1 API.

### 5.1. Projects

| Method | Path                               | Description                                     | Corresponding v1 Endpoint(s)                                                                                                                                                             |
| :----- | :--------------------------------- | :---------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET`  | `/api/v2/projects`                 | Retrieve all projects for the current user.     | `GET /projects`                                                                                                                                                                          |
| `POST` | `/api/v2/projects`                 | Create a new project.                           | `PUT /projects`                                                                                                                                                                          |
| `GET`  | `/api/v2/projects/{id}`            | Retrieve a single project by its ID.            | `GET /projects/:project`                                                                                                                                                                 |
| `PUT`  | `/api/v2/projects/{id}`            | Update a project.                               | `POST /projects/:project`                                                                                                                                                                |
| `DELETE` | `/api/v2/projects/{id}`            | Delete a project.                               | `DELETE /projects/:project`                                                                                                                                                              |
| `POST` | `/api/v2/projects/{id}/duplicate`  | Duplicate a project.                            | `PUT /projects/:projectid/duplicate`                                                                                                                                                     |
| `GET`  | `/api/v2/projects/{id}/users`      | Get all users associated with a project.        | `GET /projects/:project/projectusers`, `GET /projects/:project/users`                                                                                                                    |
| `POST` | `/api/v2/projects/{id}/users`      | Add a user to a project.                        | `PUT /projects/:project/users`                                                                                                                                                           |
| `PUT`  | `/api/v2/projects/{id}/users/{userId}` | Update a user's role in a project.              | `POST /projects/:project/users/:user`                                                                                                                                                    |
| `DELETE` | `/api/v2/projects/{id}/users/{userId}` | Remove a user from a project.                   | `DELETE /projects/:project/users/:user`                                                                                                                                                  |
| `GET`  | `/api/v2/projects/{id}/teams`      | Get all teams associated with a project.        | `GET /projects/:project/teams`                                                                                                                                                           |
| `POST` | `/api/v2/projects/{id}/teams`      | Add a team to a project.                        | `PUT /projects/:project/teams`                                                                                                                                                           |
| `PUT`  | `/api/v2/projects/{id}/teams/{teamId}` | Update a team's role in a project.              | `POST /projects/:project/teams/:team`                                                                                                                                                    |
| `DELETE` | `/api/v2/projects/{id}/teams/{teamId}` | Remove a team from a project.                   | `DELETE /projects/:project/teams/:team`                                                                                                                                                  |

### 5.2. Tasks

| Method | Path                               | Description                                     | Corresponding v1 Endpoint(s)                                                                                                                                                             |
| :----- | :--------------------------------- | :---------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET`  | `/api/v2/tasks`                    | Retrieve all tasks for the current user.        | `GET /tasks/all`                                                                                                                                                                         |
| `GET`  | `/api/v2/projects/{id}/tasks`      | Retrieve all tasks for a project.               | `GET /projects/:project/tasks`, `GET /projects/:project/views/:view/tasks`                                                                                                               |
| `POST` | `/api/v2/projects/{id}/tasks`      | Create a new task in a project.                 | `PUT /projects/:project/tasks`                                                                                                                                                           |
| `GET`  | `/api/v2/tasks/{id}`               | Retrieve a single task by its ID.               | `GET /tasks/:projecttask`                                                                                                                                                                |
| `PUT`  | `/api/v2/tasks/{id}`               | Update a task.                                  | `POST /tasks/:projecttask`                                                                                                                                                               |
| `DELETE` | `/api/v2/tasks/{id}`               | Delete a task.                                  | `DELETE /tasks/:projecttask`                                                                                                                                                             |
| `PUT`  | `/api/v2/tasks/{id}/position`      | Update the position of a task.                  | `POST /tasks/:task/position`                                                                                                                                                             |
| `POST` | `/api/v2/tasks/bulk`               | Bulk update tasks.                              | `POST /tasks/bulk`                                                                                                                                                                       |
| `GET`  | `/api/v2/tasks/{id}/assignees`     | Get all assignees for a task.                   | `GET /tasks/:projecttask/assignees`                                                                                                                                                      |
| `POST` | `/api/v2/tasks/{id}/assignees`     | Add an assignee to a task.                      | `PUT /tasks/:projecttask/assignees`                                                                                                                                                      |
| `DELETE` | `/api/v2/tasks/{id}/assignees/{userId}` | Remove an assignee from a task.                 | `DELETE /tasks/:projecttask/assignees/:user`                                                                                                                                             |
| `POST` | `/api/v2/tasks/{id}/assignees/bulk` | Bulk add assignees to a task.                   | `POST /tasks/:projecttask/assignees/bulk`                                                                                                                                                |
| `GET`  | `/api/v2/tasks/{id}/labels`        | Get all labels for a task.                      | `GET /tasks/:projecttask/labels`                                                                                                                                                         |
| `POST` | `/api/v2/tasks/{id}/labels`        | Add a label to a task.                          | `PUT /tasks/:projecttask/labels`                                                                                                                                                         |
| `DELETE` | `/api/v2/tasks/{id}/labels/{labelId}` | Remove a label from a task.                     | `DELETE /tasks/:projecttask/labels/:label`                                                                                                                                               |
| `POST` | `/api/v2/tasks/{id}/labels/bulk`   | Bulk add labels to a task.                      | `POST /tasks/:projecttask/labels/bulk`                                                                                                                                                   |
| `GET`  | `/api/v2/tasks/{id}/relations`     | Get all relations for a task.                   | (New endpoint, based on `models.TaskRelation`)                                                                                                                                          |
| `POST` | `/api/v2/tasks/{id}/relations`     | Add a relation to a task.                       | `PUT /tasks/:task/relations`                                                                                                                                                             |
| `DELETE` | `/api/v2/tasks/{id}/relations/{relationId}` | Remove a relation from a task.                  | `DELETE /tasks/:task/relations/:relationKind/:otherTask`                                                                                                                                 |

### 5.3. Labels

| Method | Path                      | Description                               | Corresponding v1 Endpoint(s) |
| :----- | :------------------------ | :---------------------------------------- | :--------------------------- |
| `GET`  | `/api/v2/labels`          | Retrieve all labels for the current user. | `GET /labels`                |
| `POST` | `/api/v2/labels`          | Create a new label.                       | `PUT /labels`                |
| `GET`  | `/api/v2/labels/{id}`     | Retrieve a single label by its ID.        | `GET /labels/:label`         |
| `PUT`  | `/api/v2/labels/{id}`     | Update a label.                           | `POST /labels/:label`        |
| `DELETE` | `/api/v2/labels/{id}`     | Delete a label.                           | `DELETE /labels/:label`      |

### 5.4. Teams

| Method | Path                               | Description                               | Corresponding v1 Endpoint(s)                                                                                                                                                             |
| :----- | :--------------------------------- | :---------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET`  | `/api/v2/teams`                    | Retrieve all teams for the current user.  | `GET /teams`                                                                                                                                                                             |
| `POST` | `/api/v2/teams`                    | Create a new team.                        | `PUT /teams`                                                                                                                                                                             |
| `GET`  | `/api/v2/teams/{id}`               | Retrieve a single team by its ID.         | `GET /teams/:team`                                                                                                                                                                       |
| `PUT`  | `/api/v2/teams/{id}`               | Update a team.                            | `POST /teams/:team`                                                                                                                                                                      |
| `DELETE` | `/api/v2/teams/{id}`               | Delete a team.                            | `DELETE /teams/:team`                                                                                                                                                                    |
| `GET`  | `/api/v2/teams/{id}/members`       | Get all members of a team.                | (No direct equivalent, but based on `models.TeamMember`)                                                                                                                                 |
| `POST` | `/api/v2/teams/{id}/members`       | Add a member to a team.                   | `PUT /teams/:team/members`                                                                                                                                                               |
| `PUT`  | `/api/v2/teams/{id}/members/{userId}` | Update a member's role in a team.         | `POST /teams/:team/members/:user/admin`                                                                                                                                                  |
| `DELETE` | `/api/v2/teams/{id}/members/{userId}` | Remove a member from a team.              | `DELETE /teams/:team/members/:user`                                                                                                                                                      |

### 5.5. Other Resources

This section covers resources that are not core to the project management but are essential for feature parity.

| Resource         | Method | Path                                              | Description                                      | Corresponding v1 Endpoint(s)                                                                                                                                                           |
| :--------------- | :----- | :------------------------------------------------ | :----------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Notifications**| `GET`  | `/api/v2/notifications`                           | Get all notifications for the current user.      | `GET /notifications`                                                                                                                                                                   |
|                  | `PUT`  | `/api/v2/notifications/{id}`                      | Mark a notification as read/unread.              | `POST /notifications/:notificationid`                                                                                                                                                  |
|                  | `PUT`  | `/api/v2/notifications/read`                      | Mark all notifications as read.                  | `POST /notifications`                                                                                                                                                                  |
| **Link Sharing** | `GET`  | `/api/v2/projects/{id}/shares`                    | Get all shared links for a project.              | `GET /projects/:project/shares`                                                                                                                                                        |
|                  | `POST` | `/api/v2/projects/{id}/shares`                    | Create a new shared link for a project.          | `PUT /projects/:project/shares`                                                                                                                                                        |
|                  | `GET`  | `/api/v2/shares/{token}`                          | Get a shared link by its token.                  | `GET /projects/:project/shares/:share`                                                                                                                                                 |
|                  | `DELETE` | `/api/v2/shares/{token}`                          | Delete a shared link.                            | `DELETE /projects/:project/shares/:share`                                                                                                                                              |
| **Attachments**  | `GET`  | `/api/v2/tasks/{id}/attachments`                  | Get all attachments for a task.                  | `GET /tasks/:task/attachments`                                                                                                                                                         |
|                  | `POST` | `/api/v2/tasks/{id}/attachments`                  | Add an attachment to a task.                     | `PUT /tasks/:task/attachments`                                                                                                                                                         |
|                  | `GET`  | `/api/v2/attachments/{id}`                        | Get a single attachment.                         | `GET /tasks/:task/attachments/:attachment`                                                                                                                                             |
|                  | `DELETE` | `/api/v2/attachments/{id}`                        | Delete an attachment.                            | `DELETE /tasks/:task/attachments/:attachment`                                                                                                                                          |
| **Comments**     | `GET`  | `/api/v2/tasks/{id}/comments`                     | Get all comments for a task.                     | `GET /tasks/:task/comments`                                                                                                                                                            |
|                  | `POST` | `/api/v2/tasks/{id}/comments`                     | Add a comment to a task.                         | `PUT /tasks/:task/comments`                                                                                                                                                            |
|                  | `GET`  | `/api/v2/comments/{id}`                           | Get a single comment.                            | `GET /tasks/:task/comments/:commentid`                                                                                                                                                 |
|                  | `PUT`  | `/api/v2/comments/{id}`                           | Update a comment.                                | `POST /tasks/:task/comments/:commentid`                                                                                                                                                |
|                  | `DELETE` | `/api/v2/comments/{id}`                           | Delete a comment.                                | `DELETE /tasks/:task/comments/:commentid`                                                                                                                                              |
| **Saved Filters**| `GET`  | `/api/v2/filters`                                 | Get all saved filters for the current user.      | (No direct equivalent, but based on `models.SavedFilter`)                                                                                                                              |
|                  | `POST` | `/api/v2/filters`                                 | Create a new saved filter.                       | `PUT /filters`                                                                                                                                                                         |
|                  | `GET`  | `/api/v2/filters/{id}`                            | Get a single saved filter.                       | `GET /filters/:filter`                                                                                                                                                                 |
|                  | `PUT`  | `/api/v2/filters/{id}`                            | Update a saved filter.                           | `POST /filters/:filter`                                                                                                                                                                |
|                  | `DELETE` | `/api/v2/filters/{id}`                            | Delete a saved filter.                           | `DELETE /filters/:filter`                                                                                                                                                              |
| **Webhooks**     | `GET`  | `/api/v2/projects/{id}/webhooks`                  | Get all webhooks for a project.                  | `GET /projects/:project/webhooks`                                                                                                                                                      |
|                  | `POST` | `/api/v2/projects/{id}/webhooks`                  | Create a new webhook for a project.              | `PUT /projects/:project/webhooks`                                                                                                                                                      |
|                  | `PUT`  | `/api/v2/webhooks/{id}`                           | Update a webhook.                                | `POST /projects/:project/webhooks/:webhook`                                                                                                                                            |
|                  | `DELETE` | `/api/v2/webhooks/{id}`                           | Delete a webhook.                                | `DELETE /projects/:project/webhooks/:webhook`                                                                                                                                          |
|                  | `GET`  | `/api/v2/webhooks/events`                         | Get all available webhook events.                | `GET /webhooks/events`                                                                                                                                                                 |
| **Reactions**    | `GET`  | `/api/v2/{entityType}/{entityId}/reactions`       | Get all reactions for an entity.                 | `GET /:entitykind/:entityid/reactions`                                                                                                                                                 |
|                  | `POST` | `/api/v2/{entityType}/{entityId}/reactions`       | Add a reaction to an entity.                     | `PUT /:entitykind/:entityid/reactions`                                                                                                                                                 |
|                  | `DELETE` | `/api/v2/{entityType}/{entityId}/reactions/{reactionId}` | Remove a reaction from an entity.              | `POST /:entitykind/:entityid/reactions/delete` (Note: v1 uses POST for deletion)                                                                                                       |
| **Project Views**| `GET`  | `/api/v2/projects/{id}/views`                     | Get all views for a project.                     | `GET /projects/:project/views`                                                                                                                                                         |
|                  | `POST` | `/api/v2/projects/{id}/views`                     | Create a new view for a project.                 | `PUT /projects/:project/views`                                                                                                                                                         |
|                  | `GET`  | `/api/v2/views/{id}`                              | Get a single view.                               | `GET /projects/:project/views/:view`                                                                                                                                                   |
|                  | `PUT`  | `/api/v2/views/{id}`                              | Update a view.                                   | `POST /projects/:project/views/:view`                                                                                                                                                  |
|                  | `DELETE` | `/api/v2/views/{id}`                              | Delete a view.                                   | `DELETE /projects/:project/views/:view`                                                                                                                                                |
| **Kanban Buckets** | `GET`  | `/api/v2/views/{id}/buckets`                      | Get all buckets for a Kanban view.               | `GET /projects/:project/views/:view/buckets`                                                                                                                                           |
|                  | `POST` | `/api/v2/views/{id}/buckets`                      | Create a new bucket in a Kanban view.            | `PUT /projects/:project/views/:view/buckets`                                                                                                                                           |
|                  | `PUT`  | `/api/v2/buckets/{id}`                            | Update a bucket.                                 | `POST /projects/:project/views/:view/buckets/:bucket`                                                                                                                                  |
|                  | `DELETE` | `/api/v2/buckets/{id}`                            | Delete a bucket.                                 | `DELETE /projects/:project/views/:view/buckets/:bucket`                                                                                                                                |
|                  | `POST` | `/api/v2/buckets/{id}/tasks`                      | Add a task to a bucket.                          | `POST /projects/:project/views/:view/buckets/:bucket/tasks`                                                                                                                            |

## 6. Success Metrics

*   **Feature Parity:** The v2 API successfully implements all the functionality of the v1 API.
*   **Code Quality:** The new API is well-structured, easy to maintain, and follows best practices.
*   **Test Coverage:** The new API has a high level of test coverage, including unit and integration tests.
*   **User Adoption:** The new API is adopted by developers and used in new integrations.
