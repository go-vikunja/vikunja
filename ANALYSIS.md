# Subscription Bug Analysis

We reproduced the behavior described in [issue #1181](https://github.com/go-vikunja/vikunja/issues/1181) using the local API.

## Reproduction steps

1. Built and started Vikunja (`mage build && ./vikunja`).
2. Created two users `tom` and `alex` via `/api/v1/register`.
3. Logged in both users and created a project as `tom`.
4. Added `alex` to the project.
5. As `tom`, subscribed to the project via `PUT /api/v1/subscriptions/project/{id}`.
6. Created a task in the project.
7. Fetched the task as `alex` via `/api/v1/tasks/{id}`.

`alex` receives the following subscription information even though they never
subscribed:

```json
{
  "id": 1,
  "entity": "project",
  "entity_id": 3,
  "created": "2025-07-27T10:43:00Z"
}
```

This matches the issue description: a subscription created by `tom` is shown to `alex` and cannot be removed.

## Root cause

The recursive query in `getSubscriptionsForEntitiesAndUser` does not filter project subscriptions by the current user when resolving task subscriptions. From `pkg/models/subscription.go`:

```go
    -- Check for project subscriptions (including parent projects)
    SELECT
        s.id,
        s.entity_type,
        s.entity_id,
        s.created,
        s.user_id,
        ph.level + 2 AS priority,
        ph.task_id
    FROM subscriptions s
        INNER JOIN project_hierarchy ph ON s.entity_id = ph.id
    WHERE s.entity_type = ?
```

Because `sUserCond` is missing in this part of the query, subscriptions for any user are considered. As a result, tasks inherit subscriptions from other users' project subscriptions. Projects themselves are queried with the user condition and therefore behave correctly.

Adding the missing user condition to the project subscription part of the query should fix the bug.
