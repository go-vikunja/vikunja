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

### 4.5. Endpoint Implementation

The following endpoints will be implemented in the v2 API:

**I. Core Resources**

*   **Projects (`/api/v2/projects`)**
    *   `GET /api/v2/projects`: Retrieve all projects.
    *   `POST /api/v2/projects`: Create a new project.
    *   `GET /api/v2/projects/{id}`: Retrieve a single project.
    *   `PUT /api/v2/projects/{id}`: Update a project.
    *   `DELETE /api/v2/projects/{id}`: Delete a project.
*   **Tasks (`/api/v2/tasks`)**
    *   `GET /api/v2/projects/{id}/tasks`: Retrieve all tasks for a project.
    *   `POST /api/v2/projects/{id}/tasks`: Create a new task.
    *   `GET /api/v2/tasks/{id}`: Retrieve a single task.
    *   `PUT /api/v2/tasks/{id}`: Update a task.
    *   `DELETE /api/v2/tasks/{id}`: Delete a task.
*   **Labels (`/api/v2/labels`)**
    *   `GET /api/v2/labels`: Retrieve all labels.
    *   `POST /api/v2/labels`: Create a new label.
    *   `GET /api/v2/labels/{id}`: Retrieve a single label.
    *   `PUT /api/v2/labels/{id}`: Update a label.
    *   `DELETE /api/v2/labels/{id}`: Delete a label.
*   **Teams (`/api/v2/teams`)**
    *   `GET /api/v2/teams`: Retrieve all teams.
    *   `POST /api/v2/teams`: Create a new team.
    *   `GET /api/v2/teams/{id}`: Retrieve a single team.
    *   `PUT /api/v2/teams/{id}`: Update a team.
    *   `DELETE /api/v2/teams/{id}`: Delete a team.
    *   `GET /api/v2/teams/{id}/members`: Retrieve all members for a team.
    *   `POST /api/v2/teams/{id}/members`: Add a member to a team.
    *   `DELETE /api/v2/teams/{id}/members/{userId}`: Remove a member from a team.

**II. Other Resources**

*   **Notifications:** Endpoints for managing notifications.
*   **Link Sharing:** Endpoints for managing shared links.
*   **Task Attachments:** Endpoints for managing task attachments.
*   **Task Comments:** Endpoints for managing task comments.
*   **Task Relations:** Endpoints for managing task relations.
*   **Saved Filters:** Endpoints for managing saved filters.
*   **Webhooks:** Endpoints for managing webhooks.
*   **Reactions:** Endpoints for managing reactions.
*   **Project Views:** Endpoints for managing project views.
*   **Kanban Buckets:** Endpoints for managing Kanban buckets.

## 5. Success Metrics

*   **Feature Parity:** The v2 API successfully implements all the functionality of the v1 API.
*   **Code Quality:** The new API is well-structured, easy to maintain, and follows best practices.
*   **Test Coverage:** The new API has a high level of test coverage, including unit and integration tests.
*   **User Adoption:** The new API is adopted by developers and used in new integrations.
